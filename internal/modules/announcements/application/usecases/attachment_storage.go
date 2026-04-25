package usecases

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path"
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
)

// attachmentStorageKey computes the object-storage key for an attachment.
//
// Single source of truth for the announcements-module keying scheme:
//
//	announcements/{announcement_id}/{uuid}-{file_name}
//
// Centralising this here keeps the scheme out of business logic. If the
// storage layout ever changes (e.g. partitioning by date, multi-tenant
// prefix), this is the only place to edit. Strict DDD would push this
// further into an infrastructure adapter, but for now it lives next to
// the storage interface where it is clearly an infrastructure concern.
func attachmentStorageKey(announcementID int64, fileName string) string {
	return path.Join(
		"announcements",
		fmt.Sprintf("%d", announcementID),
		uuid.NewString()+"-"+fileName,
	)
}
