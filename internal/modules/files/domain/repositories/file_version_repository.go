// Package repositories определяет интерфейсы репозиториев модуля files.
package repositories

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/files/domain/entities"
)

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
