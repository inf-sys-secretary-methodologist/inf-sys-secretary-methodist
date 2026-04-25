package usecases

import (
	"context"
	"errors"
	"io"
	"time"

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
)
