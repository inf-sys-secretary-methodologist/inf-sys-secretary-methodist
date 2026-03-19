// Package usecases содержит тесты бизнес-логики версионирования файлов.
package usecases

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
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

// --- Helper to create VersionUseCase with interfaces for testing ---

func newTestVersionUseCase(
	fileRepo *MockFileMetadataRepository,
	versionRepo *MockFileVersionRepository,
	storageClient StorageClient,
	auditLogger AuditEventLogger,
) *VersionUseCase {
	return &VersionUseCase{
		fileRepo:      fileRepo,
		versionRepo:   versionRepo,
		storageClient: storageClient,
		auditLogger:   auditLogger,
	}
}

// --- NewVersionUseCase ---

func TestNewVersionUseCase(t *testing.T) {
	uc := NewVersionUseCase(nil, nil, nil, nil)
	require.NotNil(t, uc)
	assert.Nil(t, uc.storageClient)
	assert.Nil(t, uc.auditLogger)
}

// --- CreateVersion ---

func TestVersionUseCase_CreateVersion_Success(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockVersionRepo := new(MockFileVersionRepository)
	mockStorage := new(MockStorageClient)
	mockAudit := new(MockAuditLogger)

	uc := newTestVersionUseCase(mockFileRepo, mockVersionRepo, mockStorage, mockAudit)
	ctx := context.Background()

	content := []byte("new version content")
	reader := bytes.NewReader(content)

	file := &entities.FileMetadata{
		ID:          1,
		StorageKey:  "documents/1/file.pdf",
		MimeType:    "application/pdf",
		IsTemporary: false,
		UploadedBy:  1,
	}

	input := &dto.CreateVersionInput{
		FileID:  1,
		Comment: "Updated version",
		UserID:  1,
	}

	expectedStorageKey := fmt.Sprintf("%s/v%d", file.StorageKey, 2)

	mockFileRepo.On("GetByID", ctx, int64(1)).Return(file, nil).Once()
	mockVersionRepo.On("GetNextVersionNumber", ctx, int64(1)).Return(2, nil).Once()
	mockStorage.On("Upload", ctx, expectedStorageKey, mock.Anything, int64(len(content)), "application/pdf").
		Run(func(args mock.Arguments) {
			r := args.Get(2).(io.Reader)
			_, _ = io.ReadAll(r)
		}).
		Return(&storage.FileInfo{Size: int64(len(content))}, nil).Once()
	mockVersionRepo.On("Create", ctx, mock.AnythingOfType("*entities.FileVersion")).Return(nil).Once()
	mockAudit.On("LogAuditEvent", ctx, "create_version", "file", mock.Anything).Once()

	resp, err := uc.CreateVersion(ctx, reader, int64(len(content)), input)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 2, resp.VersionNumber)
	assert.Equal(t, int64(len(content)), resp.Size)
	assert.Equal(t, "Updated version", resp.Comment)
	assert.Equal(t, int64(1), resp.CreatedBy)

	// Verify checksum
	expectedHash := sha256.Sum256(content)
	assert.Equal(t, hex.EncodeToString(expectedHash[:]), resp.Checksum)

	mockFileRepo.AssertExpectations(t)
	mockVersionRepo.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
	mockAudit.AssertExpectations(t)
}

func TestVersionUseCase_CreateVersion_FileNotFound(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockVersionRepo := new(MockFileVersionRepository)

	uc := newTestVersionUseCase(mockFileRepo, mockVersionRepo, nil, nil)
	ctx := context.Background()

	mockFileRepo.On("GetByID", ctx, int64(1)).Return(nil, errors.New("not found")).Once()

	input := &dto.CreateVersionInput{FileID: 1, UserID: 1}
	resp, err := uc.CreateVersion(ctx, bytes.NewReader([]byte("x")), 1, input)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestVersionUseCase_CreateVersion_TemporaryFileError(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockVersionRepo := new(MockFileVersionRepository)

	uc := newTestVersionUseCase(mockFileRepo, mockVersionRepo, nil, nil)
	ctx := context.Background()

	file := &entities.FileMetadata{
		ID:          1,
		IsTemporary: true,
	}

	mockFileRepo.On("GetByID", ctx, int64(1)).Return(file, nil).Once()

	input := &dto.CreateVersionInput{FileID: 1, UserID: 1}
	resp, err := uc.CreateVersion(ctx, bytes.NewReader([]byte("x")), 1, input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	var valErr *ValidationError
	assert.True(t, errors.As(err, &valErr))
	assert.Contains(t, valErr.Message, "нельзя создать версию временного файла")
}

func TestVersionUseCase_CreateVersion_GetNextVersionError(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockVersionRepo := new(MockFileVersionRepository)

	uc := newTestVersionUseCase(mockFileRepo, mockVersionRepo, nil, nil)
	ctx := context.Background()

	file := &entities.FileMetadata{
		ID:          1,
		IsTemporary: false,
	}

	mockFileRepo.On("GetByID", ctx, int64(1)).Return(file, nil).Once()
	mockVersionRepo.On("GetNextVersionNumber", ctx, int64(1)).Return(0, errors.New("db error")).Once()

	input := &dto.CreateVersionInput{FileID: 1, UserID: 1}
	resp, err := uc.CreateVersion(ctx, bytes.NewReader([]byte("x")), 1, input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "ошибка получения номера версии")
}

func TestVersionUseCase_CreateVersion_UploadError(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockVersionRepo := new(MockFileVersionRepository)
	mockStorage := new(MockStorageClient)

	uc := newTestVersionUseCase(mockFileRepo, mockVersionRepo, mockStorage, nil)
	ctx := context.Background()

	file := &entities.FileMetadata{
		ID:          1,
		StorageKey:  "documents/1/file.pdf",
		MimeType:    "application/pdf",
		IsTemporary: false,
	}

	mockFileRepo.On("GetByID", ctx, int64(1)).Return(file, nil).Once()
	mockVersionRepo.On("GetNextVersionNumber", ctx, int64(1)).Return(3, nil).Once()
	mockStorage.On("Upload", ctx, "documents/1/file.pdf/v3", mock.Anything, int64(1), "application/pdf").
		Return(nil, errors.New("upload failed")).Once()

	input := &dto.CreateVersionInput{FileID: 1, UserID: 1}
	resp, err := uc.CreateVersion(ctx, bytes.NewReader([]byte("x")), 1, input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "ошибка загрузки версии в хранилище")
}

func TestVersionUseCase_CreateVersion_RepoCreateError_CleansUpStorage(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockVersionRepo := new(MockFileVersionRepository)
	mockStorage := new(MockStorageClient)

	uc := newTestVersionUseCase(mockFileRepo, mockVersionRepo, mockStorage, nil)
	ctx := context.Background()

	file := &entities.FileMetadata{
		ID:          1,
		StorageKey:  "documents/1/file.pdf",
		MimeType:    "application/pdf",
		IsTemporary: false,
	}

	storageKey := "documents/1/file.pdf/v2"

	mockFileRepo.On("GetByID", ctx, int64(1)).Return(file, nil).Once()
	mockVersionRepo.On("GetNextVersionNumber", ctx, int64(1)).Return(2, nil).Once()
	mockStorage.On("Upload", ctx, storageKey, mock.Anything, int64(4), "application/pdf").
		Return(&storage.FileInfo{Size: 4}, nil).Once()
	mockVersionRepo.On("Create", ctx, mock.AnythingOfType("*entities.FileVersion")).
		Return(errors.New("db error")).Once()
	mockStorage.On("Delete", ctx, storageKey).Return(nil).Once()

	input := &dto.CreateVersionInput{FileID: 1, UserID: 1}
	resp, err := uc.CreateVersion(ctx, bytes.NewReader([]byte("data")), 4, input)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "ошибка сохранения версии")
	mockStorage.AssertExpectations(t)
}

func TestVersionUseCase_CreateVersion_WithoutAuditLogger(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockVersionRepo := new(MockFileVersionRepository)
	mockStorage := new(MockStorageClient)

	uc := newTestVersionUseCase(mockFileRepo, mockVersionRepo, mockStorage, nil)
	ctx := context.Background()

	file := &entities.FileMetadata{
		ID:          1,
		StorageKey:  "documents/1/file.pdf",
		MimeType:    "application/pdf",
		IsTemporary: false,
	}

	mockFileRepo.On("GetByID", ctx, int64(1)).Return(file, nil).Once()
	mockVersionRepo.On("GetNextVersionNumber", ctx, int64(1)).Return(1, nil).Once()
	mockStorage.On("Upload", ctx, "documents/1/file.pdf/v1", mock.Anything, int64(4), "application/pdf").
		Return(&storage.FileInfo{Size: 4}, nil).Once()
	mockVersionRepo.On("Create", ctx, mock.AnythingOfType("*entities.FileVersion")).Return(nil).Once()

	input := &dto.CreateVersionInput{FileID: 1, Comment: "first", UserID: 1}
	resp, err := uc.CreateVersion(ctx, bytes.NewReader([]byte("data")), 4, input)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 1, resp.VersionNumber)
}

// --- GetVersions ---

func TestVersionUseCase_GetVersions_Success(t *testing.T) {
	mockVersionRepo := new(MockFileVersionRepository)

	uc := newTestVersionUseCase(nil, mockVersionRepo, nil, nil)
	ctx := context.Background()

	versions := []*entities.FileVersion{
		{ID: 1, FileMetadataID: 1, VersionNumber: 1, Size: 1024, Checksum: "abc"},
		{ID: 2, FileMetadataID: 1, VersionNumber: 2, Size: 2048, Checksum: "def"},
	}

	mockVersionRepo.On("GetByFileMetadataID", ctx, int64(1)).Return(versions, nil).Once()

	result, err := uc.GetVersions(ctx, 1)

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, 1, result[0].VersionNumber)
	assert.Equal(t, 2, result[1].VersionNumber)
}

func TestVersionUseCase_GetVersions_Empty(t *testing.T) {
	mockVersionRepo := new(MockFileVersionRepository)

	uc := newTestVersionUseCase(nil, mockVersionRepo, nil, nil)
	ctx := context.Background()

	mockVersionRepo.On("GetByFileMetadataID", ctx, int64(999)).Return([]*entities.FileVersion{}, nil).Once()

	result, err := uc.GetVersions(ctx, 999)

	require.NoError(t, err)
	assert.Len(t, result, 0)
}

func TestVersionUseCase_GetVersions_Error(t *testing.T) {
	mockVersionRepo := new(MockFileVersionRepository)

	uc := newTestVersionUseCase(nil, mockVersionRepo, nil, nil)
	ctx := context.Background()

	mockVersionRepo.On("GetByFileMetadataID", ctx, int64(1)).Return(nil, errors.New("db error")).Once()

	result, err := uc.GetVersions(ctx, 1)

	assert.Error(t, err)
	assert.Nil(t, result)
}

// --- GetVersion ---

func TestVersionUseCase_GetVersion_Success(t *testing.T) {
	mockVersionRepo := new(MockFileVersionRepository)

	uc := newTestVersionUseCase(nil, mockVersionRepo, nil, nil)
	ctx := context.Background()

	version := &entities.FileVersion{
		ID:             1,
		FileMetadataID: 1,
		VersionNumber:  1,
		Size:           1024,
		Checksum:       "abc123",
		Comment:        "First version",
		CreatedBy:      1,
		CreatedAt:      time.Now(),
	}

	mockVersionRepo.On("GetByVersionNumber", ctx, int64(1), 1).Return(version, nil).Once()

	result, err := uc.GetVersion(ctx, 1, 1)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 1, result.VersionNumber)
	assert.Equal(t, "abc123", result.Checksum)
}

func TestVersionUseCase_GetVersion_NotFound(t *testing.T) {
	mockVersionRepo := new(MockFileVersionRepository)

	uc := newTestVersionUseCase(nil, mockVersionRepo, nil, nil)
	ctx := context.Background()

	mockVersionRepo.On("GetByVersionNumber", ctx, int64(1), 999).Return(nil, errors.New("not found")).Once()

	result, err := uc.GetVersion(ctx, 1, 999)

	assert.Error(t, err)
	assert.Nil(t, result)
}

// --- GetLatestVersion ---

func TestVersionUseCase_GetLatestVersion_Success(t *testing.T) {
	mockVersionRepo := new(MockFileVersionRepository)

	uc := newTestVersionUseCase(nil, mockVersionRepo, nil, nil)
	ctx := context.Background()

	version := &entities.FileVersion{
		ID:            3,
		VersionNumber: 3,
		Comment:       "Latest",
		CreatedBy:     1,
	}

	mockVersionRepo.On("GetLatestVersion", ctx, int64(1)).Return(version, nil).Once()

	result, err := uc.GetLatestVersion(ctx, 1)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 3, result.VersionNumber)
}

func TestVersionUseCase_GetLatestVersion_NoVersions(t *testing.T) {
	mockVersionRepo := new(MockFileVersionRepository)

	uc := newTestVersionUseCase(nil, mockVersionRepo, nil, nil)
	ctx := context.Background()

	mockVersionRepo.On("GetLatestVersion", ctx, int64(999)).Return(nil, errors.New("no versions")).Once()

	result, err := uc.GetLatestVersion(ctx, 999)

	assert.Error(t, err)
	assert.Nil(t, result)
}

// --- DownloadVersion ---

func TestVersionUseCase_DownloadVersion_Success(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockVersionRepo := new(MockFileVersionRepository)
	mockStorage := new(MockStorageClient)

	uc := newTestVersionUseCase(mockFileRepo, mockVersionRepo, mockStorage, nil)
	ctx := context.Background()

	file := &entities.FileMetadata{
		ID:           1,
		OriginalName: "document.pdf",
		MimeType:     "application/pdf",
	}

	version := &entities.FileVersion{
		ID:            1,
		VersionNumber: 2,
		StorageKey:    "documents/1/file.pdf/v2",
		Size:          3072,
	}

	mockFileRepo.On("GetByID", ctx, int64(1)).Return(file, nil).Once()
	mockVersionRepo.On("GetByVersionNumber", ctx, int64(1), 2).Return(version, nil).Once()
	mockStorage.On("GetPresignedURL", ctx, "documents/1/file.pdf/v2", time.Hour).
		Return("https://example.com/presigned/v2", nil).Once()

	result, err := uc.DownloadVersion(ctx, 1, 2)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "https://example.com/presigned/v2", result.PresignedURL)
	assert.Equal(t, "document.pdf_v2", result.FileName)
	assert.Equal(t, "application/pdf", result.MimeType)
	assert.Equal(t, int64(3072), result.Size)
}

func TestVersionUseCase_DownloadVersion_FileNotFound(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockVersionRepo := new(MockFileVersionRepository)

	uc := newTestVersionUseCase(mockFileRepo, mockVersionRepo, nil, nil)
	ctx := context.Background()

	mockFileRepo.On("GetByID", ctx, int64(999)).Return(nil, errors.New("not found")).Once()

	result, err := uc.DownloadVersion(ctx, 999, 1)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestVersionUseCase_DownloadVersion_VersionNotFound(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockVersionRepo := new(MockFileVersionRepository)

	uc := newTestVersionUseCase(mockFileRepo, mockVersionRepo, nil, nil)
	ctx := context.Background()

	file := &entities.FileMetadata{
		ID:           1,
		OriginalName: "document.pdf",
		MimeType:     "application/pdf",
	}

	mockFileRepo.On("GetByID", ctx, int64(1)).Return(file, nil).Once()
	mockVersionRepo.On("GetByVersionNumber", ctx, int64(1), 999).Return(nil, errors.New("not found")).Once()

	result, err := uc.DownloadVersion(ctx, 1, 999)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestVersionUseCase_DownloadVersion_PresignError(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockVersionRepo := new(MockFileVersionRepository)
	mockStorage := new(MockStorageClient)

	uc := newTestVersionUseCase(mockFileRepo, mockVersionRepo, mockStorage, nil)
	ctx := context.Background()

	file := &entities.FileMetadata{
		ID:       1,
		MimeType: "application/pdf",
	}
	version := &entities.FileVersion{
		ID:         1,
		StorageKey: "key/v1",
	}

	mockFileRepo.On("GetByID", ctx, int64(1)).Return(file, nil).Once()
	mockVersionRepo.On("GetByVersionNumber", ctx, int64(1), 1).Return(version, nil).Once()
	mockStorage.On("GetPresignedURL", ctx, "key/v1", time.Hour).
		Return("", errors.New("presign error")).Once()

	result, err := uc.DownloadVersion(ctx, 1, 1)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "ошибка генерации URL для скачивания")
}

// --- DeleteVersion ---

func TestVersionUseCase_DeleteVersion_SuccessAsFileOwner(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockVersionRepo := new(MockFileVersionRepository)
	mockStorage := new(MockStorageClient)
	mockAudit := new(MockAuditLogger)

	uc := newTestVersionUseCase(mockFileRepo, mockVersionRepo, mockStorage, mockAudit)
	ctx := context.Background()

	version := &entities.FileVersion{
		ID:             1,
		FileMetadataID: 1,
		VersionNumber:  2,
		StorageKey:     "documents/1/file.pdf/v2",
		CreatedBy:      2, // Different from file owner
	}

	file := &entities.FileMetadata{
		ID:         1,
		UploadedBy: 10, // File owner
	}

	mockVersionRepo.On("GetByID", ctx, int64(1)).Return(version, nil).Once()
	mockFileRepo.On("GetByID", ctx, int64(1)).Return(file, nil).Once()
	mockStorage.On("Delete", ctx, "documents/1/file.pdf/v2").Return(nil).Once()
	mockVersionRepo.On("Delete", ctx, int64(1)).Return(nil).Once()
	mockAudit.On("LogAuditEvent", ctx, "delete_version", "file", mock.Anything).Once()

	err := uc.DeleteVersion(ctx, 1, 10) // user is file owner

	require.NoError(t, err)
	mockVersionRepo.AssertExpectations(t)
	mockFileRepo.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
	mockAudit.AssertExpectations(t)
}

func TestVersionUseCase_DeleteVersion_SuccessAsVersionCreator(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockVersionRepo := new(MockFileVersionRepository)
	mockStorage := new(MockStorageClient)

	uc := newTestVersionUseCase(mockFileRepo, mockVersionRepo, mockStorage, nil)
	ctx := context.Background()

	version := &entities.FileVersion{
		ID:             1,
		FileMetadataID: 1,
		VersionNumber:  2,
		StorageKey:     "documents/1/file.pdf/v2",
		CreatedBy:      5, // Version creator
	}

	file := &entities.FileMetadata{
		ID:         1,
		UploadedBy: 10, // Different from version creator
	}

	mockVersionRepo.On("GetByID", ctx, int64(1)).Return(version, nil).Once()
	mockFileRepo.On("GetByID", ctx, int64(1)).Return(file, nil).Once()
	mockStorage.On("Delete", ctx, "documents/1/file.pdf/v2").Return(nil).Once()
	mockVersionRepo.On("Delete", ctx, int64(1)).Return(nil).Once()

	err := uc.DeleteVersion(ctx, 1, 5) // user is version creator

	require.NoError(t, err)
}

func TestVersionUseCase_DeleteVersion_NoPermission(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockVersionRepo := new(MockFileVersionRepository)

	uc := newTestVersionUseCase(mockFileRepo, mockVersionRepo, nil, nil)
	ctx := context.Background()

	version := &entities.FileVersion{
		ID:             1,
		FileMetadataID: 1,
		CreatedBy:      2,
	}
	file := &entities.FileMetadata{
		ID:         1,
		UploadedBy: 3,
	}

	mockVersionRepo.On("GetByID", ctx, int64(1)).Return(version, nil).Once()
	mockFileRepo.On("GetByID", ctx, int64(1)).Return(file, nil).Once()

	err := uc.DeleteVersion(ctx, 1, 99) // user is neither owner nor creator

	require.Error(t, err)
	var permErr *PermissionError
	assert.True(t, errors.As(err, &permErr))
}

func TestVersionUseCase_DeleteVersion_VersionNotFound(t *testing.T) {
	mockVersionRepo := new(MockFileVersionRepository)

	uc := newTestVersionUseCase(nil, mockVersionRepo, nil, nil)
	ctx := context.Background()

	mockVersionRepo.On("GetByID", ctx, int64(999)).Return(nil, errors.New("not found")).Once()

	err := uc.DeleteVersion(ctx, 999, 1)

	require.Error(t, err)
}

func TestVersionUseCase_DeleteVersion_FileGetByIDError(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockVersionRepo := new(MockFileVersionRepository)

	uc := newTestVersionUseCase(mockFileRepo, mockVersionRepo, nil, nil)
	ctx := context.Background()

	version := &entities.FileVersion{
		ID:             1,
		FileMetadataID: 1,
		CreatedBy:      1,
	}

	mockVersionRepo.On("GetByID", ctx, int64(1)).Return(version, nil).Once()
	mockFileRepo.On("GetByID", ctx, int64(1)).Return(nil, errors.New("file not found")).Once()

	err := uc.DeleteVersion(ctx, 1, 1)

	require.Error(t, err)
}

func TestVersionUseCase_DeleteVersion_StorageDeleteError(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockVersionRepo := new(MockFileVersionRepository)
	mockStorage := new(MockStorageClient)

	uc := newTestVersionUseCase(mockFileRepo, mockVersionRepo, mockStorage, nil)
	ctx := context.Background()

	version := &entities.FileVersion{
		ID:             1,
		FileMetadataID: 1,
		StorageKey:     "key/v1",
		CreatedBy:      1,
	}
	file := &entities.FileMetadata{
		ID:         1,
		UploadedBy: 1,
	}

	mockVersionRepo.On("GetByID", ctx, int64(1)).Return(version, nil).Once()
	mockFileRepo.On("GetByID", ctx, int64(1)).Return(file, nil).Once()
	mockStorage.On("Delete", ctx, "key/v1").Return(errors.New("s3 error")).Once()

	err := uc.DeleteVersion(ctx, 1, 1)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка удаления версии из хранилища")
}

func TestVersionUseCase_DeleteVersion_RepoDeleteError(t *testing.T) {
	mockFileRepo := new(MockFileMetadataRepository)
	mockVersionRepo := new(MockFileVersionRepository)
	mockStorage := new(MockStorageClient)

	uc := newTestVersionUseCase(mockFileRepo, mockVersionRepo, mockStorage, nil)
	ctx := context.Background()

	version := &entities.FileVersion{
		ID:             1,
		FileMetadataID: 1,
		StorageKey:     "key/v1",
		CreatedBy:      1,
	}
	file := &entities.FileMetadata{
		ID:         1,
		UploadedBy: 1,
	}

	mockVersionRepo.On("GetByID", ctx, int64(1)).Return(version, nil).Once()
	mockFileRepo.On("GetByID", ctx, int64(1)).Return(file, nil).Once()
	mockStorage.On("Delete", ctx, "key/v1").Return(nil).Once()
	mockVersionRepo.On("Delete", ctx, int64(1)).Return(errors.New("db error")).Once()

	err := uc.DeleteVersion(ctx, 1, 1)

	require.Error(t, err)
}

// --- GetVersionCount ---

func TestVersionUseCase_GetVersionCount_Success(t *testing.T) {
	mockVersionRepo := new(MockFileVersionRepository)

	uc := newTestVersionUseCase(nil, mockVersionRepo, nil, nil)
	ctx := context.Background()

	mockVersionRepo.On("CountByFileMetadataID", ctx, int64(1)).Return(int64(5), nil).Once()

	count, err := uc.GetVersionCount(ctx, 1)

	require.NoError(t, err)
	assert.Equal(t, int64(5), count)
}

func TestVersionUseCase_GetVersionCount_Zero(t *testing.T) {
	mockVersionRepo := new(MockFileVersionRepository)

	uc := newTestVersionUseCase(nil, mockVersionRepo, nil, nil)
	ctx := context.Background()

	mockVersionRepo.On("CountByFileMetadataID", ctx, int64(999)).Return(int64(0), nil).Once()

	count, err := uc.GetVersionCount(ctx, 999)

	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestVersionUseCase_GetVersionCount_Error(t *testing.T) {
	mockVersionRepo := new(MockFileVersionRepository)

	uc := newTestVersionUseCase(nil, mockVersionRepo, nil, nil)
	ctx := context.Background()

	mockVersionRepo.On("CountByFileMetadataID", ctx, int64(1)).Return(int64(0), errors.New("db error")).Once()

	count, err := uc.GetVersionCount(ctx, 1)

	assert.Error(t, err)
	assert.Equal(t, int64(0), count)
}

// --- toVersionResponse ---

func TestVersionUseCase_toVersionResponse(t *testing.T) {
	uc := NewVersionUseCase(nil, nil, nil, nil)

	now := time.Now()
	version := &entities.FileVersion{
		ID:             1,
		FileMetadataID: 1,
		VersionNumber:  3,
		Size:           4096,
		Checksum:       "xyz789",
		Comment:        "Final version",
		CreatedBy:      42,
		CreatedAt:      now,
	}

	response := uc.toVersionResponse(version)

	assert.Equal(t, int64(1), response.ID)
	assert.Equal(t, 3, response.VersionNumber)
	assert.Equal(t, int64(4096), response.Size)
	assert.Equal(t, "xyz789", response.Checksum)
	assert.Equal(t, "Final version", response.Comment)
	assert.Equal(t, int64(42), response.CreatedBy)
	assert.Equal(t, now, response.CreatedAt)
}
