package usecases

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/storage"
)

// AttachmentStorage abstracts the object-storage operations the announcements
// module needs. *storage.S3Client satisfies this interface.
//
// Defining the interface here (consumer side) follows DIP: the announcements
// usecase depends on its own contract, not on a concrete S3 client. Tests
// substitute an in-memory mock without dragging in the real S3 SDK.
type AttachmentStorage interface {
	Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) (*storage.FileInfo, error)
	Delete(ctx context.Context, key string) error
	GetPresignedURL(ctx context.Context, key string, expires time.Duration) (string, error)
}

var (
	// ErrStorageNotConfigured is returned when an attachment operation is
	// attempted but the usecase has no AttachmentStorage wired in (e.g. local
	// dev mode without MinIO).
	ErrStorageNotConfigured = errors.New("attachment storage not configured")

	// ErrAttachmentNotFound is returned when an attachment lookup fails.
	ErrAttachmentNotFound = errors.New("attachment not found")

	// ErrAttachmentForbidden is returned when the caller is not allowed
	// to mutate an attachment (not the author, not an admin, OR URL
	// announcement_id does not match the attachment's). v0.163.0
	// ADR-3 (#303 TIER 0).
	ErrAttachmentForbidden = errors.New("attachment access forbidden")

	// ErrAttachmentTooLarge is returned when an upload exceeds the
	// configured size cap. v0.163.0 ADR-5 (#303 TIER 1).
	ErrAttachmentTooLarge = errors.New("attachment exceeds maximum allowed size")

	// ErrAttachmentMimeRejected is returned when an upload's content
	// type is not in the allowlist. v0.163.0 ADR-5 (#303 TIER 1).
	ErrAttachmentMimeRejected = errors.New("attachment content type is not allowed")
)

// Maximum attachment size in bytes (10 MiB). v0.163.0 ADR-5 default.
// Conservative cap suitable for academic announcements — heavier
// documents should live в the documents module which has its own
// 50 MiB ceiling.
const attachmentMaxSize int64 = 10 * 1024 * 1024

// allowedAttachmentMimeTypes is the announcement-attachment MIME
// allowlist. v0.163.0 ADR-5 (#303 TIER 1). Trim-set focused on the
// formats academic secretaries actually attach (PDFs, Office docs,
// images for visual context). NO executable / archive types — those
// are documents-module territory.
var allowedAttachmentMimeTypes = map[string]bool{
	"application/pdf":   true,
	"application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"application/vnd.ms-excel": true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true,
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
	"text/plain": true,
}

// attachmentStorageKey computes the object-storage key for an attachment.
//
// Single source of truth for the announcements-module keying scheme:
//
//	announcements/{announcement_id}/{uuid}-{file_name}
//
// Centralizing this here keeps the scheme out of business logic. If the
// storage layout ever changes (e.g. partitioning by date, multi-tenant
// prefix), this is the only place to edit. Strict DDD would push this
// further into an infrastructure adapter, but for now it lives next to
// the storage interface where it is clearly an infrastructure concern.
func attachmentStorageKey(announcementID int64, fileName string) string {
	return path.Join(
		"announcements",
		fmt.Sprintf("%d", announcementID),
		uuid.NewString()+"-"+sanitizeAttachmentFileName(fileName),
	)
}

// sanitizeAttachmentFileName strips directory components and parent-dir
// segments so the result is safe to embed in a storage key.
//
// v0.163.0 ADR-4 (#303 TIER 1): pre-fix path.Join called path.Clean
// which collapsed `..` and could escape the per-announcement prefix
// (`evil/../../escape.bin` → `escape.bin` placed outside the
// `announcements/{id}/` namespace).
func sanitizeAttachmentFileName(name string) string {
	// Collapse path separators (both POSIX and Windows) to underscore.
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	// Replace `..` substrings so path.Clean (in path.Join downstream)
	// cannot reinterpret them as parent-dir segments. Single dots are
	// fine — `report.pdf` must survive.
	name = strings.ReplaceAll(name, "..", "__")
	// Strip leading dots so `.foo` cannot resolve to hidden-file
	// semantics in some storage backends.
	name = strings.TrimLeft(name, ".")
	// Collapse remaining whitespace-only / empty input к a stable
	// default so the key is never `{uuid}-`.
	name = strings.TrimSpace(name)
	if name == "" {
		return "unnamed"
	}
	return name
}
