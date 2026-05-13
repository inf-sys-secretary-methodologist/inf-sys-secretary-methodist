// Package backups hosts the admin-only read-API for the database +
// MinIO backup files produced by the /backup sidecar container. The
// sidecar (see /backup/Dockerfile + /backup/scripts) is the source of
// truth for backup lifecycle (cron schedule, encryption, retention,
// offsite sync). This package reads the shared backup_data volume +
// Prometheus textfile metrics, exposes a thin use case + HTTP handler
// pair gated by RequireRole(system_admin), and does NOT trigger or
// mutate backups itself.
//
// "shared/admin" houses cross-cutting administrative features (see
// auditlog for the canonical precedent).
package backups

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// BackupType discriminates the two artifact families produced by the
// sidecar — full PostgreSQL dump (custom-format pg_dump) versus MinIO
// bucket tarball.
type BackupType string

// BackupTypePostgres / BackupTypeMinIO are the two artifact families
// the sidecar writes under <rootDir>/postgres/ and <rootDir>/minio/
// respectively.
const (
	BackupTypePostgres BackupType = "postgres"
	BackupTypeMinIO    BackupType = "minio"
)

// EncryptionScheme reflects the optional encryption applied by the
// sidecar at write-time. The admin UI surfaces this so an operator
// knows whether the downloaded file is directly usable or requires
// an age / GPG key first.
type EncryptionScheme string

// EncryptionNone is the default scheme — file extension lacks a
// known suffix and the artifact is directly usable. EncryptionAge /
// EncryptionGPG correspond to .age and .gpg suffixes appended by the
// sidecar's encryption step.
const (
	EncryptionNone EncryptionScheme = ""
	EncryptionAge  EncryptionScheme = "age"
	EncryptionGPG  EncryptionScheme = "gpg"
)

// BackupFile is the read-only projection of one artifact in the
// backup_data volume.
type BackupFile struct {
	Name       string
	Type       BackupType
	Size       int64
	ModifiedAt int64
	Encryption EncryptionScheme
}

// ErrInvalidRoot signals an empty / non-existent backup root passed to
// the constructor. We fail fast at DI time rather than per-request so
// misconfiguration surfaces at startup.
var ErrInvalidRoot = errors.New("backups: invalid root directory")

// FileReader lists artifacts under <rootDir>/{postgres,minio}/ that
// match the sidecar's filename grammar.
type FileReader struct {
	rootDir string
}

// NewFileReader constructs a reader rooted at the backup_data mount
// path (typically /var/backups inside the backend container).
func NewFileReader(rootDir string) (*FileReader, error) {
	if rootDir == "" {
		return nil, ErrInvalidRoot
	}
	return &FileReader{rootDir: rootDir}, nil
}

// Filename grammar from /backup/scripts/backup-postgres.sh +
// backup-minio.sh: <kind>_<YYYYMMDD>_<HHMMSS>.<ext>[<.age|.gpg>].
// Anchored on both ends so partial matches inside longer filenames
// cannot smuggle through.
var (
	postgresBackupRe = regexp.MustCompile(`^postgres_\d{8}_\d{6}\.sql\.gz(\.age|\.gpg)?$`)
	minioBackupRe    = regexp.MustCompile(`^minio_\d{8}_\d{6}\.tar\.gz(\.age|\.gpg)?$`)
)

// List returns every artifact matching the sidecar's known filename
// grammar under <rootDir>/postgres/ and <rootDir>/minio/. Entries are
// sorted newest-first by ModifiedAt. Missing subdirectories are
// tolerated (the sidecar may not yet have produced artifacts on a
// fresh volume) and yield an empty slice rather than an error.
func (r *FileReader) List(ctx context.Context) ([]BackupFile, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	type subdir struct {
		name string
		kind BackupType
		re   *regexp.Regexp
	}
	subs := []subdir{
		{name: "postgres", kind: BackupTypePostgres, re: postgresBackupRe},
		{name: "minio", kind: BackupTypeMinIO, re: minioBackupRe},
	}

	out := make([]BackupFile, 0)
	for _, s := range subs {
		absDir := filepath.Join(r.rootDir, s.name)
		entries, err := os.ReadDir(absDir)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return nil, err
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if !s.re.MatchString(name) {
				continue
			}
			info, err := e.Info()
			if err != nil {
				continue
			}
			out = append(out, BackupFile{
				Name:       name,
				Type:       s.kind,
				Size:       info.Size(),
				ModifiedAt: info.ModTime().Unix(),
				Encryption: classifyEncryption(name),
			})
		}
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].ModifiedAt > out[j].ModifiedAt
	})
	return out, nil
}

func classifyEncryption(name string) EncryptionScheme {
	switch {
	case strings.HasSuffix(name, ".age"):
		return EncryptionAge
	case strings.HasSuffix(name, ".gpg"):
		return EncryptionGPG
	default:
		return EncryptionNone
	}
}
