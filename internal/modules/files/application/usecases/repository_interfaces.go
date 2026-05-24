// Package usecases — repository ports for the files module.
//
// Repository interfaces live в the consuming package (usecases) per
// Dependency Inversion Principle. Domain layer stays free of infrastructure
// abstractions; entities are the only thing imported here. Mirror v0.157.1
// curriculum + v0.160.1 users + v0.162.1 messaging.
package usecases

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

	// CountByUploadedBy возвращает количество неудалённых файлов,
	// загруженных конкретным пользователем. Используется ListFiles
	// для пагинации non-admin actor'ов после #290 ADR-1 fix-cycle:
	// non-admin actor видит только свои файлы и нуждается в
	// точном total count для UI.
	CountByUploadedBy(ctx context.Context, userID int64) (int64, error)

	// GetExpiredTemporaryFiles получает временные файлы с истёкшим сроком.
	GetExpiredTemporaryFiles(ctx context.Context, limit int) ([]*entities.FileMetadata, error)

	// CleanupExpired удаляет временные файлы с истёкшим сроком.
	CleanupExpired(ctx context.Context) (int64, error)
}

// FileVersionRepository определяет интерфейс доступа к версиям файлов.
type FileVersionRepository interface {
	// Create сохраняет новую запись версии файла.
	Create(ctx context.Context, version *entities.FileVersion) error

	// GetByID получает версию файла по ID.
	GetByID(ctx context.Context, id int64) (*entities.FileVersion, error)

	// GetByFileMetadataID получает все версии файла.
	GetByFileMetadataID(ctx context.Context, fileMetadataID int64) ([]*entities.FileVersion, error)

	// GetLatestVersion получает последнюю версию файла.
	GetLatestVersion(ctx context.Context, fileMetadataID int64) (*entities.FileVersion, error)

	// GetByVersionNumber получает конкретную версию по номеру.
	GetByVersionNumber(ctx context.Context, fileMetadataID int64, versionNumber int) (*entities.FileVersion, error)

	// Delete удаляет запись версии файла.
	Delete(ctx context.Context, id int64) error

	// DeleteByFileMetadataID удаляет все версии файла.
	DeleteByFileMetadataID(ctx context.Context, fileMetadataID int64) error

	// CountByFileMetadataID возвращает количество версий файла.
	CountByFileMetadataID(ctx context.Context, fileMetadataID int64) (int64, error)

	// GetNextVersionNumber возвращает следующий номер версии для файла.
	GetNextVersionNumber(ctx context.Context, fileMetadataID int64) (int, error)
}
