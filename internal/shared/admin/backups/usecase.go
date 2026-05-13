package backups

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// fileLister is the consumer-side projection of FileReader.List.
// Declared here (use-case package) rather than in the infrastructure
// adapter so DIP holds — the use case depends on the abstraction,
// the concrete FileReader satisfies it structurally.
type fileLister interface {
	List(ctx context.Context) ([]BackupFile, error)
}

// metricsScraper mirrors MetricsReader.Read for the same reason.
type metricsScraper interface {
	Read(ctx context.Context) (*BackupMetrics, error)
}

// fileOpener is the narrow port the use case uses to open a vetted
// backup file for streaming download. The default implementation is
// osFileOpener (os.Open + stat); tests substitute fakes when needed.
type fileOpener interface {
	Open(absPath string) (io.ReadCloser, int64, error)
}

// AuditSink mirrors the platform-wide narrow port that audit
// emissions flow through. nil-safe — when the use case is wired
// without WithAuditSink we skip emission rather than panic so the
// happy path keeps working in tests that don't care.
type AuditSink interface {
	LogAuditEvent(ctx context.Context, action, resource string, fields map[string]any)
}

// AdminBackupUseCase orchestrates the read-only admin observability
// surface over the /backup sidecar's data + metrics volumes. It owns
// the filename-grammar + path-traversal hardening and is the single
// audit-emission point for downloads.
type AdminBackupUseCase struct {
	files    fileLister
	metrics  metricsScraper
	opener   fileOpener
	audit    AuditSink
	filesDir string
}

// CombinedSnapshot is the read-only DTO returned to the admin UI —
// the file list and metrics block side-by-side so the page renders
// from one round-trip.
type CombinedSnapshot struct {
	Files   []BackupFile
	Metrics *BackupMetrics
}

// DownloadResult carries everything the handler needs to stream a
// vetted backup file back to the admin client. The caller MUST
// Close() the Reader when streaming is complete.
type DownloadResult struct {
	Reader      io.ReadCloser
	Size        int64
	ContentType string
	Filename    string
}

// ErrInvalidBackupName signals that the requested :type/:name pair
// failed validation (unknown type, filename does not match the
// sidecar's grammar, or path-traversal attempt). Handler maps to 400.
var ErrInvalidBackupName = errors.New("backups: invalid backup name")

// ErrBackupNotFound signals a vetted name that does not resolve to
// an existing file (race with retention GC, sidecar volume cleared,
// etc.). Handler maps to 404.
var ErrBackupNotFound = errors.New("backups: backup not found")

// NewAdminBackupUseCase wires the use case against the file lister,
// metrics scraper, and the resolved backup_data volume root. The
// default file opener is wired here; WithAuditSink layers on the
// audit emission after construction (mirrors v0.131.1 setter pattern).
func NewAdminBackupUseCase(files fileLister, metrics metricsScraper, filesDir string) *AdminBackupUseCase {
	if files == nil {
		panic("backups: nil fileLister")
	}
	if metrics == nil {
		panic("backups: nil metricsScraper")
	}
	if filesDir == "" {
		panic("backups: empty filesDir")
	}
	return &AdminBackupUseCase{
		files:    files,
		metrics:  metrics,
		opener:   &osFileOpener{},
		filesDir: filesDir,
	}
}

// WithAuditSink wires an audit emitter into the use case. Returns
// the receiver so the call chains at construction time:
//
//	backups.NewAdminBackupUseCase(...).WithAuditSink(auditLogger)
func (uc *AdminBackupUseCase) WithAuditSink(sink AuditSink) *AdminBackupUseCase {
	uc.audit = sink
	return uc
}

// ListWithMetrics returns the combined files+metrics snapshot. Errors
// from either the file lister or the metrics scraper bubble — the
// admin UI cannot meaningfully render a partial result.
func (uc *AdminBackupUseCase) ListWithMetrics(ctx context.Context) (CombinedSnapshot, error) {
	files, err := uc.files.List(ctx)
	if err != nil {
		return CombinedSnapshot{}, err
	}
	metrics, err := uc.metrics.Read(ctx)
	if err != nil {
		return CombinedSnapshot{}, err
	}
	return CombinedSnapshot{Files: files, Metrics: metrics}, nil
}

// Download validates the requested artifact, opens the file, and
// returns the streaming reader. Emits the `backup.downloaded` audit
// event on success — denied paths emit nothing (forensic
// false-positive guard, mirrors v0.131.1 messaging/integration
// patterns).
func (uc *AdminBackupUseCase) Download(ctx context.Context, actorID int64, backupType BackupType, name string) (*DownloadResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var grammar *regexp.Regexp
	switch backupType {
	case BackupTypePostgres:
		grammar = postgresBackupRe
	case BackupTypeMinIO:
		grammar = minioBackupRe
	default:
		return nil, ErrInvalidBackupName
	}
	if !grammar.MatchString(name) {
		return nil, ErrInvalidBackupName
	}

	// Path resolution + defense-in-depth root containment check.
	// The anchored regex already disallows '/' and '..' so this is
	// belt-and-braces — keeps the invariant explicit at the layer
	// that touches the filesystem in case the grammar is loosened
	// later.
	cleanRoot := filepath.Clean(uc.filesDir)
	absPath := filepath.Clean(filepath.Join(cleanRoot, string(backupType), name))
	if !strings.HasPrefix(absPath, cleanRoot+string(filepath.Separator)) {
		return nil, ErrInvalidBackupName
	}

	reader, size, err := uc.opener.Open(absPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, ErrBackupNotFound
		}
		return nil, err
	}

	if uc.audit != nil {
		uc.audit.LogAuditEvent(ctx, "backup.downloaded", "backup_admin", map[string]any{
			"filename":        name,
			"file_size_bytes": size,
			"backup_type":     string(backupType),
			"actor_user_id":   actorID,
		})
	}

	return &DownloadResult{
		Reader:      reader,
		Size:        size,
		ContentType: detectContentType(name),
		Filename:    name,
	}, nil
}

// detectContentType picks a Content-Type for the download response.
// We deliberately keep this minimal — gzip dumps and tarballs are
// the common case; encrypted suffixes always fall back to the
// generic octet-stream so browsers do not attempt inline decoding.
func detectContentType(name string) string {
	switch {
	case strings.HasSuffix(name, ".age"), strings.HasSuffix(name, ".gpg"):
		return "application/octet-stream"
	case strings.HasSuffix(name, ".gz"):
		return "application/gzip"
	default:
		return "application/octet-stream"
	}
}

// osFileOpener is the production fileOpener — opens the path via
// os.Open and stats it for the Content-Length header.
type osFileOpener struct{}

// Open implements fileOpener.
func (osFileOpener) Open(absPath string) (io.ReadCloser, int64, error) {
	f, err := os.Open(absPath) // #nosec G304 — path is whitelist-validated + root-contained upstream
	if err != nil {
		return nil, 0, err
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, 0, err
	}
	return f, info.Size(), nil
}
