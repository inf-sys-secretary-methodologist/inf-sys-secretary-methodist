// Package repositories определяет интерфейсы репозиториев модуля files.
package repositories

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/files/domain/entities"
)

// FileMetadataRepository определяет интерфейс доступа к метаданным файлов.
type FileMetadataRepository interface {
	// Create сохраняет новую запись метаданных файла.
	Create(ctx context.Context, file *entities.FileMetadata) error

	// GetByID получает метаданные файла по ID.
	GetByID(ctx context.Context, id int64) (*entities.FileMetadata, error)

	// GetByStorageKey получает метаданные файла по ключу хранилища.
	GetByStorageKey(ctx context.Context, storageKey string) (*entities.FileMetadata, error)

	// Update обновляет существующую запись метаданных файла.
	Update(ctx context.Context, file *entities.FileMetadata) error

	// Delete выполняет мягкое удаление метаданных файла.
	Delete(ctx context.Context, id int64) error

	// HardDelete полностью удаляет запись метаданных файла.
	HardDelete(ctx context.Context, id int64) error

	// List получает список файлов с пагинацией.
	List(ctx context.Context, limit, offset int) ([]*entities.FileMetadata, error)

	// Count возвращает общее количество неудалённых файлов.
	Count(ctx context.Context) (int64, error)

	// GetByDocumentID получает все файлы, прикреплённые к документу.
	GetByDocumentID(ctx context.Context, documentID int64) ([]*entities.FileMetadata, error)

	// GetByTaskID получает все файлы, прикреплённые к задаче.
	GetByTaskID(ctx context.Context, taskID int64) ([]*entities.FileMetadata, error)

	// GetByAnnouncementID получает все файлы, прикреплённые к объявлению.
	GetByAnnouncementID(ctx context.Context, announcementID int64) ([]*entities.FileMetadata, error)

	// GetByUploadedBy получает все файлы, загруженные пользователем.
	GetByUploadedBy(ctx context.Context, userID int64, limit, offset int) ([]*entities.FileMetadata, error)

	// GetExpiredTemporaryFiles получает временные файлы с истёкшим сроком.
	GetExpiredTemporaryFiles(ctx context.Context, limit int) ([]*entities.FileMetadata, error)

	// CleanupExpired удаляет временные файлы с истёкшим сроком.
	CleanupExpired(ctx context.Context) (int64, error)
}
