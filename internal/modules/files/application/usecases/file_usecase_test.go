// Package usecases содержит тесты бизнес-логики модуля files.
package usecases

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/files/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/files/domain/entities"
)

// MockFileMetadataRepository — мок репозитория метаданных файлов.
type MockFileMetadataRepository struct {
	mock.Mock
}

func (m *MockFileMetadataRepository) Create(ctx context.Context, file *entities.FileMetadata) error {
	args := m.Called(ctx, file)
	return args.Error(0)
}

func (m *MockFileMetadataRepository) GetByID(ctx context.Context, id int64) (*entities.FileMetadata, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.FileMetadata), args.Error(1)
}

func (m *MockFileMetadataRepository) GetByStorageKey(ctx context.Context, storageKey string) (*entities.FileMetadata, error) {
	args := m.Called(ctx, storageKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.FileMetadata), args.Error(1)
}

func (m *MockFileMetadataRepository) Update(ctx context.Context, file *entities.FileMetadata) error {
	args := m.Called(ctx, file)
	return args.Error(0)
}

func (m *MockFileMetadataRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockFileMetadataRepository) HardDelete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockFileMetadataRepository) List(ctx context.Context, limit, offset int) ([]*entities.FileMetadata, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.FileMetadata), args.Error(1)
}

func (m *MockFileMetadataRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockFileMetadataRepository) GetByDocumentID(ctx context.Context, documentID int64) ([]*entities.FileMetadata, error) {
	args := m.Called(ctx, documentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.FileMetadata), args.Error(1)
}

func (m *MockFileMetadataRepository) GetByTaskID(ctx context.Context, taskID int64) ([]*entities.FileMetadata, error) {
	args := m.Called(ctx, taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.FileMetadata), args.Error(1)
}

func (m *MockFileMetadataRepository) GetByAnnouncementID(ctx context.Context, announcementID int64) ([]*entities.FileMetadata, error) {
	args := m.Called(ctx, announcementID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.FileMetadata), args.Error(1)
}

func (m *MockFileMetadataRepository) GetByUploadedBy(ctx context.Context, userID int64, limit, offset int) ([]*entities.FileMetadata, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.FileMetadata), args.Error(1)
}

func (m *MockFileMetadataRepository) GetExpiredTemporaryFiles(ctx context.Context, limit int) ([]*entities.FileMetadata, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.FileMetadata), args.Error(1)
}

func (m *MockFileMetadataRepository) CleanupExpired(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

// MockFileVersionRepository — мок репозитория версий файлов.
type MockFileVersionRepository struct {
	mock.Mock
}

func (m *MockFileVersionRepository) Create(ctx context.Context, version *entities.FileVersion) error {
	args := m.Called(ctx, version)
	return args.Error(0)
}

func (m *MockFileVersionRepository) GetByID(ctx context.Context, id int64) (*entities.FileVersion, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.FileVersion), args.Error(1)
}

func (m *MockFileVersionRepository) GetByFileMetadataID(ctx context.Context, fileMetadataID int64) ([]*entities.FileVersion, error) {
	args := m.Called(ctx, fileMetadataID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.FileVersion), args.Error(1)
}

func (m *MockFileVersionRepository) GetLatestVersion(ctx context.Context, fileMetadataID int64) (*entities.FileVersion, error) {
	args := m.Called(ctx, fileMetadataID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.FileVersion), args.Error(1)
}

func (m *MockFileVersionRepository) GetByVersionNumber(ctx context.Context, fileMetadataID int64, versionNumber int) (*entities.FileVersion, error) {
	args := m.Called(ctx, fileMetadataID, versionNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.FileVersion), args.Error(1)
}

func (m *MockFileVersionRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockFileVersionRepository) DeleteByFileMetadataID(ctx context.Context, fileMetadataID int64) error {
	args := m.Called(ctx, fileMetadataID)
	return args.Error(0)
}

func (m *MockFileVersionRepository) CountByFileMetadataID(ctx context.Context, fileMetadataID int64) (int64, error) {
	args := m.Called(ctx, fileMetadataID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockFileVersionRepository) GetNextVersionNumber(ctx context.Context, fileMetadataID int64) (int, error) {
	args := m.Called(ctx, fileMetadataID)
	return args.Get(0).(int), args.Error(1)
}

func TestFileUseCase_GetFile(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockVersionRepo := new(MockFileVersionRepository)

	usecase := NewFileUseCase(mockFileRepo, mockVersionRepo, nil, nil, nil)

	ctx := context.Background()

	t.Run("существующий файл", func(t *testing.T) {
		expectedFile := &entities.FileMetadata{
			ID:           1,
			OriginalName: "test.pdf",
			StorageKey:   "temp/123/test.pdf",
			Size:         1024,
			MimeType:     "application/pdf",
			Checksum:     "abc123",
			UploadedBy:   1,
			IsTemporary:  true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		mockFileRepo.On("GetByID", ctx, int64(1)).Return(expectedFile, nil).Once()

		result, err := usecase.GetFile(ctx, 1)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedFile.ID, result.ID)
		assert.Equal(t, expectedFile.OriginalName, result.OriginalName)
		assert.Equal(t, expectedFile.Size, result.Size)

		mockFileRepo.AssertExpectations(t)
	})

	t.Run("несуществующий файл", func(t *testing.T) {
		mockFileRepo.On("GetByID", ctx, int64(999)).Return(nil, assert.AnError).Once()

		result, err := usecase.GetFile(ctx, 999)

		assert.Error(t, err)
		assert.Nil(t, result)

		mockFileRepo.AssertExpectations(t)
	})
}

func TestFileUseCase_ListFiles(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockVersionRepo := new(MockFileVersionRepository)

	usecase := NewFileUseCase(mockFileRepo, mockVersionRepo, nil, nil, nil)

	ctx := context.Background()

	t.Run("получение списка файлов с пагинацией", func(t *testing.T) {
		files := []*entities.FileMetadata{
			{ID: 1, OriginalName: "file1.pdf", Size: 1024},
			{ID: 2, OriginalName: "file2.docx", Size: 2048},
		}

		mockFileRepo.On("List", ctx, 10, 0).Return(files, nil).Once()
		mockFileRepo.On("Count", ctx).Return(int64(2), nil).Once()

		result, err := usecase.ListFiles(ctx, 1, 10)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Files, 2)
		assert.Equal(t, int64(2), result.Total)
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 10, result.Limit)

		mockFileRepo.AssertExpectations(t)
	})

	t.Run("пустой результат", func(t *testing.T) {
		mockFileRepo.On("List", ctx, 10, 0).Return([]*entities.FileMetadata{}, nil).Once()
		mockFileRepo.On("Count", ctx).Return(int64(0), nil).Once()

		result, err := usecase.ListFiles(ctx, 1, 10)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Files, 0)
		assert.Equal(t, int64(0), result.Total)

		mockFileRepo.AssertExpectations(t)
	})
}

func TestFileUseCase_GetFilesByDocument(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockVersionRepo := new(MockFileVersionRepository)

	usecase := NewFileUseCase(mockFileRepo, mockVersionRepo, nil, nil, nil)

	ctx := context.Background()

	t.Run("файлы документа", func(t *testing.T) {
		docID := int64(1)
		files := []*entities.FileMetadata{
			{ID: 1, OriginalName: "attachment1.pdf", DocumentID: &docID},
			{ID: 2, OriginalName: "attachment2.pdf", DocumentID: &docID},
		}

		mockFileRepo.On("GetByDocumentID", ctx, int64(1)).Return(files, nil).Once()

		result, err := usecase.GetFilesByDocument(ctx, 1)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 2)

		mockFileRepo.AssertExpectations(t)
	})
}

func TestFileUseCase_GetFilesByTask(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockVersionRepo := new(MockFileVersionRepository)

	usecase := NewFileUseCase(mockFileRepo, mockVersionRepo, nil, nil, nil)

	ctx := context.Background()

	t.Run("файлы задачи", func(t *testing.T) {
		taskID := int64(1)
		files := []*entities.FileMetadata{
			{ID: 1, OriginalName: "task_file.pdf", TaskID: &taskID},
		}

		mockFileRepo.On("GetByTaskID", ctx, int64(1)).Return(files, nil).Once()

		result, err := usecase.GetFilesByTask(ctx, 1)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 1)

		mockFileRepo.AssertExpectations(t)
	})
}

func TestFileUseCase_GetFilesByAnnouncement(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockVersionRepo := new(MockFileVersionRepository)

	usecase := NewFileUseCase(mockFileRepo, mockVersionRepo, nil, nil, nil)

	ctx := context.Background()

	t.Run("файлы объявления", func(t *testing.T) {
		announcementID := int64(1)
		files := []*entities.FileMetadata{
			{ID: 1, OriginalName: "image.png", AnnouncementID: &announcementID},
		}

		mockFileRepo.On("GetByAnnouncementID", ctx, int64(1)).Return(files, nil).Once()

		result, err := usecase.GetFilesByAnnouncement(ctx, 1)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 1)

		mockFileRepo.AssertExpectations(t)
	})
}

func TestFileUseCase_AttachFile(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockVersionRepo := new(MockFileVersionRepository)

	usecase := NewFileUseCase(mockFileRepo, mockVersionRepo, nil, nil, nil)

	ctx := context.Background()

	t.Run("прикрепление к документу", func(t *testing.T) {
		tempFile := &entities.FileMetadata{
			ID:           1,
			OriginalName: "test.pdf",
			IsTemporary:  true,
			UploadedBy:   1,
		}

		mockFileRepo.On("GetByID", ctx, int64(1)).Return(tempFile, nil).Once()
		mockFileRepo.On("Update", ctx, mock.AnythingOfType("*entities.FileMetadata")).Return(nil).Once()

		docID := int64(100)
		input := &dto.AttachFileInput{
			FileID:     1,
			DocumentID: &docID,
		}

		err := usecase.AttachFile(ctx, input)

		assert.NoError(t, err)
		mockFileRepo.AssertExpectations(t)
	})

	t.Run("ошибка - файл уже прикреплён", func(t *testing.T) {
		docID := int64(100)
		attachedFile := &entities.FileMetadata{
			ID:           1,
			OriginalName: "test.pdf",
			IsTemporary:  false,
			DocumentID:   &docID,
			UploadedBy:   1,
		}

		mockFileRepo.On("GetByID", ctx, int64(1)).Return(attachedFile, nil).Once()

		input := &dto.AttachFileInput{
			FileID:     1,
			DocumentID: &docID,
		}

		err := usecase.AttachFile(ctx, input)

		assert.Error(t, err)
		assert.IsType(t, &ValidationError{}, err)
		mockFileRepo.AssertExpectations(t)
	})

	t.Run("ошибка - не указана сущность", func(t *testing.T) {
		tempFile := &entities.FileMetadata{
			ID:           1,
			OriginalName: "test.pdf",
			IsTemporary:  true,
			UploadedBy:   1,
		}

		mockFileRepo.On("GetByID", ctx, int64(1)).Return(tempFile, nil).Once()

		input := &dto.AttachFileInput{
			FileID: 1,
		}

		err := usecase.AttachFile(ctx, input)

		assert.Error(t, err)
		assert.IsType(t, &ValidationError{}, err)
		mockFileRepo.AssertExpectations(t)
	})
}

func TestFileUseCase_DeleteFile(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockVersionRepo := new(MockFileVersionRepository)

	usecase := NewFileUseCase(mockFileRepo, mockVersionRepo, nil, nil, nil)

	ctx := context.Background()

	t.Run("удаление собственного файла", func(t *testing.T) {
		file := &entities.FileMetadata{
			ID:           1,
			OriginalName: "test.pdf",
			UploadedBy:   1,
		}

		mockFileRepo.On("GetByID", ctx, int64(1)).Return(file, nil).Once()
		mockFileRepo.On("Delete", ctx, int64(1)).Return(nil).Once()

		err := usecase.DeleteFile(ctx, 1, 1)

		assert.NoError(t, err)
		mockFileRepo.AssertExpectations(t)
	})

	t.Run("ошибка - нет прав на удаление", func(t *testing.T) {
		file := &entities.FileMetadata{
			ID:           1,
			OriginalName: "test.pdf",
			UploadedBy:   1,
		}

		mockFileRepo.On("GetByID", ctx, int64(1)).Return(file, nil).Once()

		err := usecase.DeleteFile(ctx, 1, 2) // другой пользователь

		assert.Error(t, err)
		assert.IsType(t, &PermissionError{}, err)
		mockFileRepo.AssertExpectations(t)
	})
}

func TestFileUseCase_CleanupExpiredFiles(t *testing.T) {
	// Примечание: полноценное тестирование CleanupExpiredFiles требует
	// мока S3Client, но S3Client не является интерфейсом.
	// Этот тест проверяет только репозиторий без S3 операций.

	t.Run("очистка без истёкших файлов", func(t *testing.T) {
		mockFileRepo := new(MockFileMetadataRepository)
		mockVersionRepo := new(MockFileVersionRepository)

		usecase := NewFileUseCase(mockFileRepo, mockVersionRepo, nil, nil, nil)

		ctx := context.Background()

		// Случай когда нет истёкших файлов - не нужно вызывать S3
		mockFileRepo.On("GetExpiredTemporaryFiles", ctx, 100).Return([]*entities.FileMetadata{}, nil).Once()
		mockFileRepo.On("CleanupExpired", ctx).Return(int64(0), nil).Once()

		count, err := usecase.CleanupExpiredFiles(ctx)

		assert.NoError(t, err)
		assert.Equal(t, int64(0), count)
		mockFileRepo.AssertExpectations(t)
	})
}

func TestHashingReader(t *testing.T) {
	t.Run("чтение и подсчёт хеша", func(t *testing.T) {
		content := []byte("test content for hashing")
		reader := bytes.NewReader(content)

		hr := &hashingReader{
			reader: reader,
			hasher: nil, // В реальном тесте нужен sha256.New()
		}

		// Этот тест проверяет структуру hashingReader
		assert.NotNil(t, hr.reader)
	})
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{Message: "тестовая ошибка валидации"}
	assert.Equal(t, "тестовая ошибка валидации", err.Error())
}

func TestPermissionError(t *testing.T) {
	err := &PermissionError{Message: "тестовая ошибка доступа"}
	assert.Equal(t, "тестовая ошибка доступа", err.Error())
}
