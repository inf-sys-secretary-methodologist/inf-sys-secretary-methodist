// Package usecases содержит бизнес-логику модуля files.
package usecases

import (
	"context"
	"io"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/storage"
)

// StorageClient определяет интерфейс для работы с объектным хранилищем.
// *storage.S3Client реализует этот интерфейс.
type StorageClient interface {
	Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) (*storage.FileInfo, error)
	Delete(ctx context.Context, key string) error
	GetPresignedURL(ctx context.Context, key string, expires time.Duration) (string, error)
}

// FileNameValidator определяет интерфейс для валидации имён файлов.
// *storage.FileValidator реализует этот интерфейс.
type FileNameValidator interface {
	ValidateFileName(fileName string) (string, error)
}

// AuditEventLogger определяет интерфейс для логирования аудит-событий.
// *logging.AuditLogger реализует этот интерфейс.
type AuditEventLogger interface {
	LogAuditEvent(ctx context.Context, action string, resource string, fields map[string]interface{})
}
