// Package usecases содержит тесты бизнес-логики версионирования файлов.
package usecases

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/files/domain/entities"
)

func TestVersionUseCase_GetVersions(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockVersionRepo := new(MockFileVersionRepository)

	usecase := NewVersionUseCase(mockFileRepo, mockVersionRepo, nil, nil)

	ctx := context.Background()

	t.Run("получение всех версий файла", func(t *testing.T) {
		versions := []*entities.FileVersion{
			{
				ID:             1,
				FileMetadataID: 1,
				VersionNumber:  1,
				Size:           1024,
				Checksum:       "abc123",
				Comment:        "Первая версия",
				CreatedBy:      1,
				CreatedAt:      time.Now(),
			},
			{
				ID:             2,
				FileMetadataID: 1,
				VersionNumber:  2,
				Size:           2048,
				Checksum:       "def456",
				Comment:        "Исправления",
				CreatedBy:      1,
				CreatedAt:      time.Now(),
			},
		}

		mockVersionRepo.On("GetByFileMetadataID", ctx, int64(1)).Return(versions, nil).Once()

		result, err := usecase.GetVersions(ctx, 1)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 2)
		assert.Equal(t, 1, result[0].VersionNumber)
		assert.Equal(t, 2, result[1].VersionNumber)

		mockVersionRepo.AssertExpectations(t)
	})

	t.Run("пустой список версий", func(t *testing.T) {
		mockVersionRepo.On("GetByFileMetadataID", ctx, int64(999)).Return([]*entities.FileVersion{}, nil).Once()

		result, err := usecase.GetVersions(ctx, 999)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 0)

		mockVersionRepo.AssertExpectations(t)
	})
}

func TestVersionUseCase_GetVersion(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockVersionRepo := new(MockFileVersionRepository)

	usecase := NewVersionUseCase(mockFileRepo, mockVersionRepo, nil, nil)

	ctx := context.Background()

	t.Run("получение конкретной версии", func(t *testing.T) {
		expectedVersion := &entities.FileVersion{
			ID:             1,
			FileMetadataID: 1,
			VersionNumber:  1,
			Size:           1024,
			Checksum:       "abc123",
			Comment:        "Первая версия",
			CreatedBy:      1,
			CreatedAt:      time.Now(),
		}

		mockVersionRepo.On("GetByVersionNumber", ctx, int64(1), 1).Return(expectedVersion, nil).Once()

		result, err := usecase.GetVersion(ctx, 1, 1)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.VersionNumber)
		assert.Equal(t, "abc123", result.Checksum)

		mockVersionRepo.AssertExpectations(t)
	})

	t.Run("несуществующая версия", func(t *testing.T) {
		mockVersionRepo.On("GetByVersionNumber", ctx, int64(1), 999).Return(nil, assert.AnError).Once()

		result, err := usecase.GetVersion(ctx, 1, 999)

		assert.Error(t, err)
		assert.Nil(t, result)

		mockVersionRepo.AssertExpectations(t)
	})
}

func TestVersionUseCase_GetLatestVersion(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockVersionRepo := new(MockFileVersionRepository)

	usecase := NewVersionUseCase(mockFileRepo, mockVersionRepo, nil, nil)

	ctx := context.Background()

	t.Run("получение последней версии", func(t *testing.T) {
		latestVersion := &entities.FileVersion{
			ID:             3,
			FileMetadataID: 1,
			VersionNumber:  3,
			Size:           3072,
			Checksum:       "ghi789",
			Comment:        "Финальная версия",
			CreatedBy:      1,
			CreatedAt:      time.Now(),
		}

		mockVersionRepo.On("GetLatestVersion", ctx, int64(1)).Return(latestVersion, nil).Once()

		result, err := usecase.GetLatestVersion(ctx, 1)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 3, result.VersionNumber)
		assert.Equal(t, "Финальная версия", result.Comment)

		mockVersionRepo.AssertExpectations(t)
	})

	t.Run("файл без версий", func(t *testing.T) {
		mockVersionRepo.On("GetLatestVersion", ctx, int64(999)).Return(nil, assert.AnError).Once()

		result, err := usecase.GetLatestVersion(ctx, 999)

		assert.Error(t, err)
		assert.Nil(t, result)

		mockVersionRepo.AssertExpectations(t)
	})
}

func TestVersionUseCase_GetVersionCount(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockVersionRepo := new(MockFileVersionRepository)

	usecase := NewVersionUseCase(mockFileRepo, mockVersionRepo, nil, nil)

	ctx := context.Background()

	t.Run("подсчёт версий файла", func(t *testing.T) {
		mockVersionRepo.On("CountByFileMetadataID", ctx, int64(1)).Return(int64(5), nil).Once()

		count, err := usecase.GetVersionCount(ctx, 1)

		assert.NoError(t, err)
		assert.Equal(t, int64(5), count)

		mockVersionRepo.AssertExpectations(t)
	})

	t.Run("файл без версий", func(t *testing.T) {
		mockVersionRepo.On("CountByFileMetadataID", ctx, int64(999)).Return(int64(0), nil).Once()

		count, err := usecase.GetVersionCount(ctx, 999)

		assert.NoError(t, err)
		assert.Equal(t, int64(0), count)

		mockVersionRepo.AssertExpectations(t)
	})
}

func TestVersionUseCase_DeleteVersion(t *testing.T) {
	// Примечание: полноценное тестирование DeleteVersion требует
	// мока S3Client, но S3Client не является интерфейсом.
	// Тестируем только проверку прав доступа.

	t.Run("ошибка - нет прав на удаление", func(t *testing.T) {
		mockFileRepo := new(MockFileMetadataRepository)
		mockVersionRepo := new(MockFileVersionRepository)

		usecase := NewVersionUseCase(mockFileRepo, mockVersionRepo, nil, nil)

		ctx := context.Background()

		version := &entities.FileVersion{
			ID:             1,
			FileMetadataID: 1,
			VersionNumber:  2,
			StorageKey:     "files/1/v2",
			CreatedBy:      2,
		}

		file := &entities.FileMetadata{
			ID:         1,
			UploadedBy: 3,
		}

		mockVersionRepo.On("GetByID", ctx, int64(1)).Return(version, nil).Once()
		mockFileRepo.On("GetByID", ctx, int64(1)).Return(file, nil).Once()

		err := usecase.DeleteVersion(ctx, 1, 1) // пользователь без прав

		assert.Error(t, err)
		assert.IsType(t, &PermissionError{}, err)
		mockVersionRepo.AssertExpectations(t)
		mockFileRepo.AssertExpectations(t)
	})

	t.Run("несуществующая версия", func(t *testing.T) {
		mockFileRepo := new(MockFileMetadataRepository)
		mockVersionRepo := new(MockFileVersionRepository)

		usecase := NewVersionUseCase(mockFileRepo, mockVersionRepo, nil, nil)

		ctx := context.Background()

		mockVersionRepo.On("GetByID", ctx, int64(999)).Return(nil, assert.AnError).Once()

		err := usecase.DeleteVersion(ctx, 999, 1)

		assert.Error(t, err)
		mockVersionRepo.AssertExpectations(t)
	})
}

func TestVersionUseCase_DownloadVersion(t *testing.T) {
	// Примечание: полноценное тестирование DownloadVersion требует
	// мока S3Client для генерации presigned URL.
	// Тестируем только получение метаданных файла и версии.

	t.Run("ошибка при несуществующем файле", func(t *testing.T) {
		mockFileRepo := new(MockFileMetadataRepository)
		mockVersionRepo := new(MockFileVersionRepository)

		usecase := NewVersionUseCase(mockFileRepo, mockVersionRepo, nil, nil)

		ctx := context.Background()

		mockFileRepo.On("GetByID", ctx, int64(999)).Return(nil, assert.AnError).Once()

		_, err := usecase.DownloadVersion(ctx, 999, 1)

		assert.Error(t, err)
		mockFileRepo.AssertExpectations(t)
	})

	t.Run("ошибка при несуществующей версии", func(t *testing.T) {
		mockFileRepo := new(MockFileMetadataRepository)
		mockVersionRepo := new(MockFileVersionRepository)

		usecase := NewVersionUseCase(mockFileRepo, mockVersionRepo, nil, nil)

		ctx := context.Background()

		file := &entities.FileMetadata{
			ID:           1,
			OriginalName: "document.pdf",
			MimeType:     "application/pdf",
		}

		mockFileRepo.On("GetByID", ctx, int64(1)).Return(file, nil).Once()
		mockVersionRepo.On("GetByVersionNumber", ctx, int64(1), 999).Return(nil, assert.AnError).Once()

		_, err := usecase.DownloadVersion(ctx, 1, 999)

		assert.Error(t, err)
		mockFileRepo.AssertExpectations(t)
		mockVersionRepo.AssertExpectations(t)
	})
}

func TestVersionUseCase_CreateVersion_Validation(t *testing.T) {
	// Этот тест проверяет корректность создания use case.
	// Полноценное тестирование CreateVersion требует мока S3Client.

	t.Run("проверка создания use case", func(t *testing.T) {
		mockFileRepo := new(MockFileMetadataRepository)
		mockVersionRepo := new(MockFileVersionRepository)

		usecase := NewVersionUseCase(mockFileRepo, mockVersionRepo, nil, nil)

		assert.NotNil(t, usecase)
	})
}

func TestVersionResponse(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockVersionRepo := new(MockFileVersionRepository)

	usecase := NewVersionUseCase(mockFileRepo, mockVersionRepo, nil, nil)

	t.Run("конвертация entity в response", func(t *testing.T) {
		now := time.Now()
		version := &entities.FileVersion{
			ID:             1,
			FileMetadataID: 1,
			VersionNumber:  1,
			Size:           1024,
			Checksum:       "abc123",
			Comment:        "Тестовый комментарий",
			CreatedBy:      1,
			CreatedAt:      now,
		}

		response := usecase.toVersionResponse(version)

		assert.Equal(t, int64(1), response.ID)
		assert.Equal(t, 1, response.VersionNumber)
		assert.Equal(t, int64(1024), response.Size)
		assert.Equal(t, "abc123", response.Checksum)
		assert.Equal(t, "Тестовый комментарий", response.Comment)
		assert.Equal(t, int64(1), response.CreatedBy)
		assert.Equal(t, now, response.CreatedAt)
	})
}
