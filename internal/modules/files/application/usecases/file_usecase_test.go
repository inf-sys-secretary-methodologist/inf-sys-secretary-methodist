// Package usecases содержит тесты бизнес-логики модуля files.
package usecases

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/files/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/files/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/storage"
)

// --- Mock implementations ---

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

// MockStorageClient — мок клиента хранилища.
type MockStorageClient struct {
	mock.Mock
}

func (m *MockStorageClient) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) (*storage.FileInfo, error) {
	args := m.Called(ctx, key, reader, size, contentType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*storage.FileInfo), args.Error(1)
}

func (m *MockStorageClient) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockStorageClient) GetPresignedURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	args := m.Called(ctx, key, expires)
	return args.String(0), args.Error(1)
}

// MockFileNameValidator — мок валидатора имён файлов.
type MockFileNameValidator struct {
	mock.Mock
}

func (m *MockFileNameValidator) ValidateFileName(fileName string) (string, error) {
	args := m.Called(fileName)
	return args.String(0), args.Error(1)
}

// MockAuditLogger — мок аудит-логгера.
type MockAuditLogger struct {
	mock.Mock
}

func (m *MockAuditLogger) LogAuditEvent(ctx context.Context, action string, resource string, fields map[string]interface{}) {
	m.Called(ctx, action, resource, fields)
}

// --- Helper to create FileUseCase with interfaces for testing ---

func newTestFileUseCase(
	fileRepo *MockFileMetadataRepository,
	versionRepo *MockFileVersionRepository,
	storageClient StorageClient,
	validator FileNameValidator,
	auditLogger AuditEventLogger,
) *FileUseCase {
	return &FileUseCase{
		fileRepo:       fileRepo,
		versionRepo:    versionRepo,
		storageClient:  storageClient,
		fileValidator:  validator,
		auditLogger:    auditLogger,
		tempExpiration: 24 * time.Hour,
	}
}

// --- Tests ---

func TestNewFileUseCase(t *testing.T) {
	uc := NewFileUseCase(nil, nil, nil, nil, nil)
	require.NotNil(t, uc)
	assert.Equal(t, 24*time.Hour, uc.tempExpiration)
	assert.Nil(t, uc.storageClient)
	assert.Nil(t, uc.fileValidator)
	assert.Nil(t, uc.auditLogger)
}

// --- UploadFile ---

func TestFileUseCase_UploadFile_Success(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockStorage := new(MockStorageClient)
	mockAudit := new(MockAuditLogger)

	uc := newTestFileUseCase(mockFileRepo, nil, mockStorage, nil, mockAudit)
	ctx := context.Background()

	content := []byte("test file content")
	reader := bytes.NewReader(content)

	input := &dto.UploadFileInput{
		OriginalName: "report.pdf",
		MimeType:     "application/pdf",
		Size:         int64(len(content)),
		UserID:       42,
	}

	mockStorage.On("Upload", ctx, mock.AnythingOfType("string"), mock.Anything, input.Size, input.MimeType).
		Return(&storage.FileInfo{Size: int64(len(content))}, nil).Once()
	mockFileRepo.On("Create", ctx, mock.AnythingOfType("*entities.FileMetadata")).Return(nil).Once()
	mockAudit.On("LogAuditEvent", ctx, "upload", "file", mock.Anything).Once()

	resp, err := uc.UploadFile(ctx, reader, input)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "report.pdf", resp.OriginalName)
	assert.Equal(t, int64(len(content)), resp.Size)
	assert.Equal(t, "application/pdf", resp.MimeType)
	assert.NotEmpty(t, resp.Checksum)
	mockStorage.AssertExpectations(t)
	mockFileRepo.AssertExpectations(t)
	mockAudit.AssertExpectations(t)
}

func TestFileUseCase_UploadFile_ChecksumCorrectness(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockStorage := new(MockStorageClient)

	uc := newTestFileUseCase(mockFileRepo, nil, mockStorage, nil, nil)
	ctx := context.Background()

	content := []byte("deterministic content")
	reader := bytes.NewReader(content)

	expectedHash := sha256.Sum256(content)
	expectedChecksum := hex.EncodeToString(expectedHash[:])

	input := &dto.UploadFileInput{
		OriginalName: "file.txt",
		MimeType:     "text/plain",
		Size:         int64(len(content)),
		UserID:       1,
	}

	// The mock Upload must read from the reader so that hashingReader computes the hash
	mockStorage.On("Upload", ctx, mock.AnythingOfType("string"), mock.Anything, input.Size, input.MimeType).
		Run(func(args mock.Arguments) {
			r := args.Get(2).(io.Reader)
			_, _ = io.ReadAll(r)
		}).
		Return(&storage.FileInfo{Size: int64(len(content))}, nil).Once()
	mockFileRepo.On("Create", ctx, mock.AnythingOfType("*entities.FileMetadata")).Return(nil).Once()

	resp, err := uc.UploadFile(ctx, reader, input)

	require.NoError(t, err)
	assert.Equal(t, expectedChecksum, resp.Checksum)
}

func TestFileUseCase_UploadFile_ValidationError(t *testing.T) {
	mockValidator := new(MockFileNameValidator)

	uc := newTestFileUseCase(nil, nil, nil, mockValidator, nil)
	ctx := context.Background()

	input := &dto.UploadFileInput{
		OriginalName: "",
		MimeType:     "text/plain",
		Size:         100,
		UserID:       1,
	}

	mockValidator.On("ValidateFileName", "").Return("", errors.New("имя файла не может быть пустым")).Once()

	resp, err := uc.UploadFile(ctx, bytes.NewReader([]byte("x")), input)

	assert.Nil(t, resp)
	require.Error(t, err)
	var valErr *ValidationError
	assert.True(t, errors.As(err, &valErr))
	assert.Contains(t, valErr.Message, "имя файла не может быть пустым")
	mockValidator.AssertExpectations(t)
}

func TestFileUseCase_UploadFile_StorageUploadError(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockStorage := new(MockStorageClient)

	uc := newTestFileUseCase(mockFileRepo, nil, mockStorage, nil, nil)
	ctx := context.Background()

	input := &dto.UploadFileInput{
		OriginalName: "file.pdf",
		MimeType:     "application/pdf",
		Size:         100,
		UserID:       1,
	}

	mockStorage.On("Upload", ctx, mock.AnythingOfType("string"), mock.Anything, int64(100), "application/pdf").
		Return(nil, errors.New("storage unavailable")).Once()

	resp, err := uc.UploadFile(ctx, bytes.NewReader([]byte("x")), input)

	assert.Nil(t, resp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка загрузки файла в хранилище")
	mockStorage.AssertExpectations(t)
}

func TestFileUseCase_UploadFile_RepoCreateError_CleansUpStorage(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockStorage := new(MockStorageClient)

	uc := newTestFileUseCase(mockFileRepo, nil, mockStorage, nil, nil)
	ctx := context.Background()

	input := &dto.UploadFileInput{
		OriginalName: "file.pdf",
		MimeType:     "application/pdf",
		Size:         5,
		UserID:       1,
	}

	mockStorage.On("Upload", ctx, mock.AnythingOfType("string"), mock.Anything, int64(5), "application/pdf").
		Return(&storage.FileInfo{Size: 5}, nil).Once()
	mockFileRepo.On("Create", ctx, mock.AnythingOfType("*entities.FileMetadata")).
		Return(errors.New("db error")).Once()
	mockStorage.On("Delete", ctx, mock.AnythingOfType("string")).Return(nil).Once()

	resp, err := uc.UploadFile(ctx, bytes.NewReader([]byte("hello")), input)

	assert.Nil(t, resp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка сохранения метаданных файла")
	mockStorage.AssertExpectations(t)
	mockFileRepo.AssertExpectations(t)
}

func TestFileUseCase_UploadFile_WithoutAuditLogger(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockStorage := new(MockStorageClient)

	uc := newTestFileUseCase(mockFileRepo, nil, mockStorage, nil, nil)
	ctx := context.Background()

	input := &dto.UploadFileInput{
		OriginalName: "file.txt",
		MimeType:     "text/plain",
		Size:         4,
		UserID:       1,
	}

	mockStorage.On("Upload", ctx, mock.AnythingOfType("string"), mock.Anything, int64(4), "text/plain").
		Return(&storage.FileInfo{Size: 4}, nil).Once()
	mockFileRepo.On("Create", ctx, mock.AnythingOfType("*entities.FileMetadata")).Return(nil).Once()

	resp, err := uc.UploadFile(ctx, bytes.NewReader([]byte("data")), input)

	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestFileUseCase_UploadFile_WithValidatorSuccess(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockStorage := new(MockStorageClient)
	mockValidator := new(MockFileNameValidator)

	uc := newTestFileUseCase(mockFileRepo, nil, mockStorage, mockValidator, nil)
	ctx := context.Background()

	input := &dto.UploadFileInput{
		OriginalName: "valid_file.pdf",
		MimeType:     "application/pdf",
		Size:         4,
		UserID:       1,
	}

	mockValidator.On("ValidateFileName", "valid_file.pdf").Return("valid_file.pdf", nil).Once()
	mockStorage.On("Upload", ctx, mock.AnythingOfType("string"), mock.Anything, int64(4), "application/pdf").
		Return(&storage.FileInfo{Size: 4}, nil).Once()
	mockFileRepo.On("Create", ctx, mock.AnythingOfType("*entities.FileMetadata")).Return(nil).Once()

	resp, err := uc.UploadFile(ctx, bytes.NewReader([]byte("data")), input)

	require.NoError(t, err)
	require.NotNil(t, resp)
	mockValidator.AssertExpectations(t)
}

// --- GetFile ---

func TestFileUseCase_GetFile_Success(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)

	uc := NewFileUseCase(mockFileRepo, nil, nil, nil, nil)
	ctx := context.Background()

	now := time.Now()
	expectedFile := &entities.FileMetadata{
		ID:           1,
		OriginalName: "test.pdf",
		StorageKey:   "temp/123/test.pdf",
		Size:         1024,
		MimeType:     "application/pdf",
		Checksum:     "abc123",
		UploadedBy:   1,
		IsTemporary:  true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	mockFileRepo.On("GetByID", ctx, int64(1)).Return(expectedFile, nil).Once()

	result, err := uc.GetFile(ctx, 1)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, expectedFile.ID, result.ID)
	assert.Equal(t, expectedFile.OriginalName, result.OriginalName)
	assert.Equal(t, expectedFile.Size, result.Size)
	assert.Equal(t, expectedFile.MimeType, result.MimeType)
	assert.Equal(t, expectedFile.Checksum, result.Checksum)
	assert.Equal(t, expectedFile.UploadedBy, result.UploadedBy)
	assert.Equal(t, expectedFile.IsTemporary, result.IsTemporary)
	assert.Equal(t, expectedFile.CreatedAt, result.CreatedAt)
	assert.Equal(t, expectedFile.UpdatedAt, result.UpdatedAt)
	mockFileRepo.AssertExpectations(t)
}

func TestFileUseCase_GetFile_NotFound(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)

	uc := NewFileUseCase(mockFileRepo, nil, nil, nil, nil)
	ctx := context.Background()

	mockFileRepo.On("GetByID", ctx, int64(999)).Return(nil, errors.New("not found")).Once()

	result, err := uc.GetFile(ctx, 999)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockFileRepo.AssertExpectations(t)
}

// --- GetFileWithDownloadURL ---

func TestFileUseCase_GetFileWithDownloadURL_Success(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockStorage := new(MockStorageClient)

	uc := newTestFileUseCase(mockFileRepo, nil, mockStorage, nil, nil)
	ctx := context.Background()

	file := &entities.FileMetadata{
		ID:           1,
		OriginalName: "test.pdf",
		StorageKey:   "temp/1/test.pdf",
		Size:         1024,
		MimeType:     "application/pdf",
		Checksum:     "abc",
		UploadedBy:   1,
	}

	mockFileRepo.On("GetByID", ctx, int64(1)).Return(file, nil).Once()
	mockStorage.On("GetPresignedURL", ctx, "temp/1/test.pdf", 5*time.Minute).
		Return("https://storage.example.com/presigned/test.pdf", nil).Once()

	result, err := uc.GetFileWithDownloadURL(ctx, 1, 5*time.Minute)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, int64(1), result.ID)
	assert.Equal(t, "https://storage.example.com/presigned/test.pdf", result.DownloadURL)
	mockFileRepo.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}

func TestFileUseCase_GetFileWithDownloadURL_FileNotFound(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)

	uc := newTestFileUseCase(mockFileRepo, nil, nil, nil, nil)
	ctx := context.Background()

	mockFileRepo.On("GetByID", ctx, int64(1)).Return(nil, errors.New("not found")).Once()

	result, err := uc.GetFileWithDownloadURL(ctx, 1, 5*time.Minute)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestFileUseCase_GetFileWithDownloadURL_PresignError(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockStorage := new(MockStorageClient)

	uc := newTestFileUseCase(mockFileRepo, nil, mockStorage, nil, nil)
	ctx := context.Background()

	file := &entities.FileMetadata{
		ID:         1,
		StorageKey: "temp/1/test.pdf",
	}

	mockFileRepo.On("GetByID", ctx, int64(1)).Return(file, nil).Once()
	mockStorage.On("GetPresignedURL", ctx, "temp/1/test.pdf", 5*time.Minute).
		Return("", errors.New("presign failed")).Once()

	result, err := uc.GetFileWithDownloadURL(ctx, 1, 5*time.Minute)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "ошибка генерации URL для скачивания")
}

// --- DownloadFile ---

func TestFileUseCase_DownloadFile_Success(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockStorage := new(MockStorageClient)

	uc := newTestFileUseCase(mockFileRepo, nil, mockStorage, nil, nil)
	ctx := context.Background()

	file := &entities.FileMetadata{
		ID:           1,
		OriginalName: "doc.pdf",
		StorageKey:   "temp/1/doc.pdf",
		MimeType:     "application/pdf",
		Size:         2048,
	}

	mockFileRepo.On("GetByID", ctx, int64(1)).Return(file, nil).Once()
	mockStorage.On("GetPresignedURL", ctx, "temp/1/doc.pdf", time.Hour).
		Return("https://example.com/presigned", nil).Once()

	result, err := uc.DownloadFile(ctx, 1)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "https://example.com/presigned", result.PresignedURL)
	assert.Equal(t, "doc.pdf", result.FileName)
	assert.Equal(t, "application/pdf", result.MimeType)
	assert.Equal(t, int64(2048), result.Size)
}

func TestFileUseCase_DownloadFile_FileNotFound(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)

	uc := newTestFileUseCase(mockFileRepo, nil, nil, nil, nil)
	ctx := context.Background()

	mockFileRepo.On("GetByID", ctx, int64(1)).Return(nil, errors.New("not found")).Once()

	result, err := uc.DownloadFile(ctx, 1)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestFileUseCase_DownloadFile_PresignError(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockStorage := new(MockStorageClient)

	uc := newTestFileUseCase(mockFileRepo, nil, mockStorage, nil, nil)
	ctx := context.Background()

	file := &entities.FileMetadata{
		ID:         1,
		StorageKey: "key",
	}

	mockFileRepo.On("GetByID", ctx, int64(1)).Return(file, nil).Once()
	mockStorage.On("GetPresignedURL", ctx, "key", time.Hour).
		Return("", errors.New("s3 error")).Once()

	result, err := uc.DownloadFile(ctx, 1)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "ошибка генерации URL для скачивания")
}

// --- AttachFile ---

func TestFileUseCase_AttachFile_ToDocument(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockAudit := new(MockAuditLogger)

	uc := newTestFileUseCase(mockFileRepo, nil, nil, nil, mockAudit)
	ctx := context.Background()

	tempFile := &entities.FileMetadata{
		ID:          1,
		IsTemporary: true,
		UploadedBy:  1,
	}

	mockFileRepo.On("GetByID", ctx, int64(1)).Return(tempFile, nil).Once()
	mockFileRepo.On("Update", ctx, mock.AnythingOfType("*entities.FileMetadata")).Return(nil).Once()
	mockAudit.On("LogAuditEvent", ctx, "attach", "file", mock.Anything).Once()

	docID := int64(100)
	input := &dto.AttachFileInput{
		FileID:     1,
		DocumentID: &docID,
	}

	err := uc.AttachFile(ctx, input)

	require.NoError(t, err)
	mockFileRepo.AssertExpectations(t)
	mockAudit.AssertExpectations(t)
}

func TestFileUseCase_AttachFile_ToTask(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)

	uc := newTestFileUseCase(mockFileRepo, nil, nil, nil, nil)
	ctx := context.Background()

	tempFile := &entities.FileMetadata{
		ID:          1,
		IsTemporary: true,
	}

	mockFileRepo.On("GetByID", ctx, int64(1)).Return(tempFile, nil).Once()
	mockFileRepo.On("Update", ctx, mock.AnythingOfType("*entities.FileMetadata")).Return(nil).Once()

	taskID := int64(200)
	input := &dto.AttachFileInput{
		FileID: 1,
		TaskID: &taskID,
	}

	err := uc.AttachFile(ctx, input)

	require.NoError(t, err)
	mockFileRepo.AssertExpectations(t)
}

func TestFileUseCase_AttachFile_ToAnnouncement(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)

	uc := newTestFileUseCase(mockFileRepo, nil, nil, nil, nil)
	ctx := context.Background()

	tempFile := &entities.FileMetadata{
		ID:          1,
		IsTemporary: true,
	}

	mockFileRepo.On("GetByID", ctx, int64(1)).Return(tempFile, nil).Once()
	mockFileRepo.On("Update", ctx, mock.AnythingOfType("*entities.FileMetadata")).Return(nil).Once()

	annID := int64(300)
	input := &dto.AttachFileInput{
		FileID:         1,
		AnnouncementID: &annID,
	}

	err := uc.AttachFile(ctx, input)

	require.NoError(t, err)
	mockFileRepo.AssertExpectations(t)
}

func TestFileUseCase_AttachFile_AlreadyAttached(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)

	uc := newTestFileUseCase(mockFileRepo, nil, nil, nil, nil)
	ctx := context.Background()

	docID := int64(100)
	file := &entities.FileMetadata{
		ID:          1,
		IsTemporary: false,
		DocumentID:  &docID,
	}

	mockFileRepo.On("GetByID", ctx, int64(1)).Return(file, nil).Once()

	input := &dto.AttachFileInput{
		FileID:     1,
		DocumentID: &docID,
	}

	err := uc.AttachFile(ctx, input)

	require.Error(t, err)
	var valErr *ValidationError
	assert.True(t, errors.As(err, &valErr))
	assert.Contains(t, valErr.Message, "файл уже прикреплён")
}

func TestFileUseCase_AttachFile_NoEntitySpecified(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)

	uc := newTestFileUseCase(mockFileRepo, nil, nil, nil, nil)
	ctx := context.Background()

	tempFile := &entities.FileMetadata{
		ID:          1,
		IsTemporary: true,
	}

	mockFileRepo.On("GetByID", ctx, int64(1)).Return(tempFile, nil).Once()

	input := &dto.AttachFileInput{FileID: 1}

	err := uc.AttachFile(ctx, input)

	require.Error(t, err)
	var valErr *ValidationError
	assert.True(t, errors.As(err, &valErr))
	assert.Contains(t, valErr.Message, "необходимо указать")
}

func TestFileUseCase_AttachFile_GetByIDError(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)

	uc := newTestFileUseCase(mockFileRepo, nil, nil, nil, nil)
	ctx := context.Background()

	mockFileRepo.On("GetByID", ctx, int64(1)).Return(nil, errors.New("db error")).Once()

	docID := int64(100)
	input := &dto.AttachFileInput{
		FileID:     1,
		DocumentID: &docID,
	}

	err := uc.AttachFile(ctx, input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

func TestFileUseCase_AttachFile_UpdateError(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)

	uc := newTestFileUseCase(mockFileRepo, nil, nil, nil, nil)
	ctx := context.Background()

	tempFile := &entities.FileMetadata{
		ID:          1,
		IsTemporary: true,
	}

	mockFileRepo.On("GetByID", ctx, int64(1)).Return(tempFile, nil).Once()
	mockFileRepo.On("Update", ctx, mock.AnythingOfType("*entities.FileMetadata")).
		Return(errors.New("update failed")).Once()

	docID := int64(100)
	input := &dto.AttachFileInput{
		FileID:     1,
		DocumentID: &docID,
	}

	err := uc.AttachFile(ctx, input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка обновления метаданных файла")
}

// --- DeleteFile ---

func TestFileUseCase_DeleteFile_Success(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockAudit := new(MockAuditLogger)

	uc := newTestFileUseCase(mockFileRepo, nil, nil, nil, mockAudit)
	ctx := context.Background()

	file := &entities.FileMetadata{
		ID:           1,
		OriginalName: "test.pdf",
		UploadedBy:   1,
	}

	mockFileRepo.On("GetByID", ctx, int64(1)).Return(file, nil).Once()
	mockFileRepo.On("Delete", ctx, int64(1)).Return(nil).Once()
	mockAudit.On("LogAuditEvent", ctx, "delete", "file", mock.Anything).Once()

	err := uc.DeleteFile(ctx, 1, 1)

	require.NoError(t, err)
	mockFileRepo.AssertExpectations(t)
	mockAudit.AssertExpectations(t)
}

func TestFileUseCase_DeleteFile_NoPermission(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)

	uc := newTestFileUseCase(mockFileRepo, nil, nil, nil, nil)
	ctx := context.Background()

	file := &entities.FileMetadata{
		ID:         1,
		UploadedBy: 1,
	}

	mockFileRepo.On("GetByID", ctx, int64(1)).Return(file, nil).Once()

	err := uc.DeleteFile(ctx, 1, 2)

	require.Error(t, err)
	var permErr *PermissionError
	assert.True(t, errors.As(err, &permErr))
}

func TestFileUseCase_DeleteFile_GetByIDError(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)

	uc := newTestFileUseCase(mockFileRepo, nil, nil, nil, nil)
	ctx := context.Background()

	mockFileRepo.On("GetByID", ctx, int64(1)).Return(nil, errors.New("not found")).Once()

	err := uc.DeleteFile(ctx, 1, 1)

	require.Error(t, err)
}

func TestFileUseCase_DeleteFile_RepoDeleteError(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)

	uc := newTestFileUseCase(mockFileRepo, nil, nil, nil, nil)
	ctx := context.Background()

	file := &entities.FileMetadata{
		ID:         1,
		UploadedBy: 1,
	}

	mockFileRepo.On("GetByID", ctx, int64(1)).Return(file, nil).Once()
	mockFileRepo.On("Delete", ctx, int64(1)).Return(errors.New("db error")).Once()

	err := uc.DeleteFile(ctx, 1, 1)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

func TestFileUseCase_DeleteFile_WithoutAuditLogger(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)

	uc := newTestFileUseCase(mockFileRepo, nil, nil, nil, nil)
	ctx := context.Background()

	file := &entities.FileMetadata{
		ID:         1,
		UploadedBy: 1,
	}

	mockFileRepo.On("GetByID", ctx, int64(1)).Return(file, nil).Once()
	mockFileRepo.On("Delete", ctx, int64(1)).Return(nil).Once()

	err := uc.DeleteFile(ctx, 1, 1)

	require.NoError(t, err)
}

// --- ListFiles ---

func TestFileUseCase_ListFiles_Success(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)

	uc := NewFileUseCase(mockFileRepo, nil, nil, nil, nil)
	ctx := context.Background()

	files := []*entities.FileMetadata{
		{ID: 1, OriginalName: "file1.pdf", Size: 1024},
		{ID: 2, OriginalName: "file2.docx", Size: 2048},
	}

	mockFileRepo.On("List", ctx, 10, 0).Return(files, nil).Once()
	mockFileRepo.On("Count", ctx).Return(int64(2), nil).Once()

	result, err := uc.ListFiles(ctx, 1, 10)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Files, 2)
	assert.Equal(t, int64(2), result.Total)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 10, result.Limit)
	assert.Equal(t, 1, result.TotalPages)
}

func TestFileUseCase_ListFiles_Pagination(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)

	uc := NewFileUseCase(mockFileRepo, nil, nil, nil, nil)
	ctx := context.Background()

	files := []*entities.FileMetadata{
		{ID: 3, OriginalName: "file3.pdf"},
	}

	// page 2, limit 2 => offset 2
	mockFileRepo.On("List", ctx, 2, 2).Return(files, nil).Once()
	mockFileRepo.On("Count", ctx).Return(int64(5), nil).Once()

	result, err := uc.ListFiles(ctx, 2, 2)

	require.NoError(t, err)
	assert.Equal(t, 2, result.Page)
	assert.Equal(t, 2, result.Limit)
	assert.Equal(t, 3, result.TotalPages) // ceil(5/2) = 3
}

func TestFileUseCase_ListFiles_DefaultsForInvalidInput(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)

	uc := NewFileUseCase(mockFileRepo, nil, nil, nil, nil)
	ctx := context.Background()

	// page < 1 => page = 1, limit <= 0 => limit = 10
	mockFileRepo.On("List", ctx, 10, 0).Return([]*entities.FileMetadata{}, nil).Once()
	mockFileRepo.On("Count", ctx).Return(int64(0), nil).Once()

	result, err := uc.ListFiles(ctx, 0, -5)

	require.NoError(t, err)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 10, result.Limit)
}

func TestFileUseCase_ListFiles_LimitCappedAt100(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)

	uc := NewFileUseCase(mockFileRepo, nil, nil, nil, nil)
	ctx := context.Background()

	mockFileRepo.On("List", ctx, 100, 0).Return([]*entities.FileMetadata{}, nil).Once()
	mockFileRepo.On("Count", ctx).Return(int64(0), nil).Once()

	result, err := uc.ListFiles(ctx, 1, 500)

	require.NoError(t, err)
	assert.Equal(t, 100, result.Limit)
}

func TestFileUseCase_ListFiles_ListError(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)

	uc := NewFileUseCase(mockFileRepo, nil, nil, nil, nil)
	ctx := context.Background()

	mockFileRepo.On("List", ctx, 10, 0).Return(nil, errors.New("db error")).Once()

	result, err := uc.ListFiles(ctx, 1, 10)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestFileUseCase_ListFiles_CountError(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)

	uc := NewFileUseCase(mockFileRepo, nil, nil, nil, nil)
	ctx := context.Background()

	mockFileRepo.On("List", ctx, 10, 0).Return([]*entities.FileMetadata{}, nil).Once()
	mockFileRepo.On("Count", ctx).Return(int64(0), errors.New("count error")).Once()

	result, err := uc.ListFiles(ctx, 1, 10)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestFileUseCase_ListFiles_TotalPagesExact(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)

	uc := NewFileUseCase(mockFileRepo, nil, nil, nil, nil)
	ctx := context.Background()

	mockFileRepo.On("List", ctx, 5, 0).Return([]*entities.FileMetadata{}, nil).Once()
	mockFileRepo.On("Count", ctx).Return(int64(10), nil).Once()

	result, err := uc.ListFiles(ctx, 1, 5)

	require.NoError(t, err)
	assert.Equal(t, 2, result.TotalPages) // 10/5 = 2, no remainder
}

// --- GetFilesByDocument ---

func TestFileUseCase_GetFilesByDocument_Success(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)

	uc := NewFileUseCase(mockFileRepo, nil, nil, nil, nil)
	ctx := context.Background()

	docID := int64(1)
	files := []*entities.FileMetadata{
		{ID: 1, OriginalName: "att1.pdf", DocumentID: &docID},
		{ID: 2, OriginalName: "att2.pdf", DocumentID: &docID},
	}

	mockFileRepo.On("GetByDocumentID", ctx, int64(1)).Return(files, nil).Once()

	result, err := uc.GetFilesByDocument(ctx, 1)

	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestFileUseCase_GetFilesByDocument_Error(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)

	uc := NewFileUseCase(mockFileRepo, nil, nil, nil, nil)
	ctx := context.Background()

	mockFileRepo.On("GetByDocumentID", ctx, int64(1)).Return(nil, errors.New("db error")).Once()

	result, err := uc.GetFilesByDocument(ctx, 1)

	assert.Error(t, err)
	assert.Nil(t, result)
}

// --- GetFilesByTask ---

func TestFileUseCase_GetFilesByTask_Success(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)

	uc := NewFileUseCase(mockFileRepo, nil, nil, nil, nil)
	ctx := context.Background()

	taskID := int64(1)
	files := []*entities.FileMetadata{
		{ID: 1, OriginalName: "task_file.pdf", TaskID: &taskID},
	}

	mockFileRepo.On("GetByTaskID", ctx, int64(1)).Return(files, nil).Once()

	result, err := uc.GetFilesByTask(ctx, 1)

	require.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestFileUseCase_GetFilesByTask_Error(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)

	uc := NewFileUseCase(mockFileRepo, nil, nil, nil, nil)
	ctx := context.Background()

	mockFileRepo.On("GetByTaskID", ctx, int64(1)).Return(nil, errors.New("db error")).Once()

	result, err := uc.GetFilesByTask(ctx, 1)

	assert.Error(t, err)
	assert.Nil(t, result)
}

// --- GetFilesByAnnouncement ---

func TestFileUseCase_GetFilesByAnnouncement_Success(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)

	uc := NewFileUseCase(mockFileRepo, nil, nil, nil, nil)
	ctx := context.Background()

	annID := int64(1)
	files := []*entities.FileMetadata{
		{ID: 1, OriginalName: "image.png", AnnouncementID: &annID},
	}

	mockFileRepo.On("GetByAnnouncementID", ctx, int64(1)).Return(files, nil).Once()

	result, err := uc.GetFilesByAnnouncement(ctx, 1)

	require.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestFileUseCase_GetFilesByAnnouncement_Error(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)

	uc := NewFileUseCase(mockFileRepo, nil, nil, nil, nil)
	ctx := context.Background()

	mockFileRepo.On("GetByAnnouncementID", ctx, int64(1)).Return(nil, errors.New("db error")).Once()

	result, err := uc.GetFilesByAnnouncement(ctx, 1)

	assert.Error(t, err)
	assert.Nil(t, result)
}

// --- CleanupExpiredFiles ---

func TestFileUseCase_CleanupExpiredFiles_NoExpiredFiles(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockStorage := new(MockStorageClient)

	uc := newTestFileUseCase(mockFileRepo, nil, mockStorage, nil, nil)
	ctx := context.Background()

	mockFileRepo.On("GetExpiredTemporaryFiles", ctx, 100).Return([]*entities.FileMetadata{}, nil).Once()
	mockFileRepo.On("CleanupExpired", ctx).Return(int64(0), nil).Once()

	count, err := uc.CleanupExpiredFiles(ctx)

	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestFileUseCase_CleanupExpiredFiles_WithExpiredFiles(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockStorage := new(MockStorageClient)
	mockAudit := new(MockAuditLogger)

	uc := newTestFileUseCase(mockFileRepo, nil, mockStorage, nil, mockAudit)
	ctx := context.Background()

	expiredFiles := []*entities.FileMetadata{
		{ID: 1, StorageKey: "temp/1/file1.pdf"},
		{ID: 2, StorageKey: "temp/2/file2.pdf"},
	}

	mockFileRepo.On("GetExpiredTemporaryFiles", ctx, 100).Return(expiredFiles, nil).Once()
	mockStorage.On("Delete", ctx, "temp/1/file1.pdf").Return(nil).Once()
	mockStorage.On("Delete", ctx, "temp/2/file2.pdf").Return(nil).Once()
	mockFileRepo.On("CleanupExpired", ctx).Return(int64(2), nil).Once()
	mockAudit.On("LogAuditEvent", ctx, "cleanup", "files", mock.Anything).Once()

	count, err := uc.CleanupExpiredFiles(ctx)

	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
	mockStorage.AssertExpectations(t)
	mockAudit.AssertExpectations(t)
}

func TestFileUseCase_CleanupExpiredFiles_GetExpiredError(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)

	uc := newTestFileUseCase(mockFileRepo, nil, nil, nil, nil)
	ctx := context.Background()

	mockFileRepo.On("GetExpiredTemporaryFiles", ctx, 100).Return(nil, errors.New("db error")).Once()

	count, err := uc.CleanupExpiredFiles(ctx)

	assert.Error(t, err)
	assert.Equal(t, int64(0), count)
}

func TestFileUseCase_CleanupExpiredFiles_CleanupExpiredError(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockStorage := new(MockStorageClient)

	uc := newTestFileUseCase(mockFileRepo, nil, mockStorage, nil, nil)
	ctx := context.Background()

	mockFileRepo.On("GetExpiredTemporaryFiles", ctx, 100).Return([]*entities.FileMetadata{}, nil).Once()
	mockFileRepo.On("CleanupExpired", ctx).Return(int64(0), errors.New("cleanup error")).Once()

	count, err := uc.CleanupExpiredFiles(ctx)

	assert.Error(t, err)
	assert.Equal(t, int64(0), count)
}

func TestFileUseCase_CleanupExpiredFiles_StorageDeleteErrorIgnored(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockStorage := new(MockStorageClient)

	uc := newTestFileUseCase(mockFileRepo, nil, mockStorage, nil, nil)
	ctx := context.Background()

	expiredFiles := []*entities.FileMetadata{
		{ID: 1, StorageKey: "temp/1/file1.pdf"},
	}

	mockFileRepo.On("GetExpiredTemporaryFiles", ctx, 100).Return(expiredFiles, nil).Once()
	// S3 delete fails, but error is ignored (code uses _ = uc.storageClient.Delete(...))
	mockStorage.On("Delete", ctx, "temp/1/file1.pdf").Return(errors.New("s3 error")).Once()
	mockFileRepo.On("CleanupExpired", ctx).Return(int64(1), nil).Once()

	count, err := uc.CleanupExpiredFiles(ctx)

	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestFileUseCase_CleanupExpiredFiles_ZeroCountNoAudit(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockStorage := new(MockStorageClient)
	mockAudit := new(MockAuditLogger)

	uc := newTestFileUseCase(mockFileRepo, nil, mockStorage, nil, mockAudit)
	ctx := context.Background()

	mockFileRepo.On("GetExpiredTemporaryFiles", ctx, 100).Return([]*entities.FileMetadata{}, nil).Once()
	mockFileRepo.On("CleanupExpired", ctx).Return(int64(0), nil).Once()
	// audit should NOT be called when count == 0

	count, err := uc.CleanupExpiredFiles(ctx)

	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
	mockAudit.AssertNotCalled(t, "LogAuditEvent")
}

// --- toFileResponse ---

func TestFileUseCase_toFileResponse(t *testing.T) {
	uc := NewFileUseCase(nil, nil, nil, nil, nil)

	now := time.Now()
	docID := int64(10)
	taskID := int64(20)
	annID := int64(30)
	file := &entities.FileMetadata{
		ID:             1,
		OriginalName:   "test.pdf",
		Size:           2048,
		MimeType:       "application/pdf",
		Checksum:       "sha256hash",
		UploadedBy:     42,
		DocumentID:     &docID,
		TaskID:         &taskID,
		AnnouncementID: &annID,
		IsTemporary:    false,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	resp := uc.toFileResponse(file)

	assert.Equal(t, int64(1), resp.ID)
	assert.Equal(t, "test.pdf", resp.OriginalName)
	assert.Equal(t, int64(2048), resp.Size)
	assert.Equal(t, "application/pdf", resp.MimeType)
	assert.Equal(t, "sha256hash", resp.Checksum)
	assert.Equal(t, int64(42), resp.UploadedBy)
	assert.Equal(t, &docID, resp.DocumentID)
	assert.Equal(t, &taskID, resp.TaskID)
	assert.Equal(t, &annID, resp.AnnouncementID)
	assert.False(t, resp.IsTemporary)
	assert.Equal(t, now, resp.CreatedAt)
	assert.Equal(t, now, resp.UpdatedAt)
}

// --- hashingReader ---

func TestHashingReader_Read(t *testing.T) {
	content := []byte("test content for hashing")
	reader := bytes.NewReader(content)
	hasher := sha256.New()

	hr := &hashingReader{reader: reader, hasher: hasher}

	buf := make([]byte, 1024)
	n, err := hr.Read(buf)

	require.NoError(t, err)
	assert.Equal(t, len(content), n)

	// The hasher should have been updated
	expectedHash := sha256.Sum256(content)
	assert.Equal(t, expectedHash[:], hr.hasher.Sum(nil))
}

func TestHashingReader_ReadChunked(t *testing.T) {
	content := []byte("hello world")
	reader := bytes.NewReader(content)
	hasher := sha256.New()

	hr := &hashingReader{reader: reader, hasher: hasher}

	// Read in small chunks
	buf := make([]byte, 5)
	n1, err1 := hr.Read(buf)
	assert.NoError(t, err1)
	assert.Equal(t, 5, n1)

	n2, err2 := hr.Read(buf)
	assert.NoError(t, err2)
	assert.Equal(t, 5, n2)

	n3, err3 := hr.Read(buf)
	assert.NoError(t, err3)
	assert.Equal(t, 1, n3)

	// Next read should be EOF
	_, err4 := hr.Read(buf)
	assert.Equal(t, io.EOF, err4)

	expectedHash := sha256.Sum256(content)
	assert.Equal(t, expectedHash[:], hr.hasher.Sum(nil))
}

func TestHashingReader_ReadEOF(t *testing.T) {
	// Empty reader
	reader := bytes.NewReader([]byte{})
	hasher := sha256.New()

	hr := &hashingReader{reader: reader, hasher: hasher}

	buf := make([]byte, 10)
	n, err := hr.Read(buf)

	assert.Equal(t, 0, n)
	assert.Equal(t, io.EOF, err)
}

// --- Error types ---

func TestValidationError(t *testing.T) {
	err := &ValidationError{Message: "тестовая ошибка валидации"}
	assert.Equal(t, "тестовая ошибка валидации", err.Error())
	assert.Implements(t, (*error)(nil), err)
}

func TestPermissionError(t *testing.T) {
	err := &PermissionError{Message: "тестовая ошибка доступа"}
	assert.Equal(t, "тестовая ошибка доступа", err.Error())
	assert.Implements(t, (*error)(nil), err)
}
