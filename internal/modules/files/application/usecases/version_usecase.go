// Package usecases содержит бизнес-логику модуля files.
package usecases

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/files/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/files/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/files/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/storage"
)

// VersionUseCase обрабатывает бизнес-логику версионирования файлов.
type VersionUseCase struct {
	fileRepo      repositories.FileMetadataRepository
	versionRepo   repositories.FileVersionRepository
	storageClient StorageClient
	auditLogger   AuditEventLogger
}

// NewVersionUseCase создаёт новый use case для версий файлов.
func NewVersionUseCase(
	fileRepo repositories.FileMetadataRepository,
	versionRepo repositories.FileVersionRepository,
	s3Client *storage.S3Client,
	auditLogger *logging.AuditLogger,
) *VersionUseCase {
	uc := &VersionUseCase{
		fileRepo:    fileRepo,
		versionRepo: versionRepo,
	}
	if s3Client != nil {
		uc.storageClient = s3Client
	}
	if auditLogger != nil {
		uc.auditLogger = auditLogger
	}
	return uc
}

// CreateVersion создаёт новую версию файла.
func (uc *VersionUseCase) CreateVersion(ctx context.Context, reader io.Reader, size int64, input *dto.CreateVersionInput) (*dto.FileVersionResponse, error) {
	// Получаем метаданные файла
	file, err := uc.fileRepo.GetByID(ctx, input.FileID)
	if err != nil {
		return nil, err
	}

	// Файл должен быть прикреплён
	if file.IsTemporary {
		return nil, &ValidationError{Message: "нельзя создать версию временного файла"}
	}

	// Получаем следующий номер версии
	nextVersion, err := uc.versionRepo.GetNextVersionNumber(ctx, input.FileID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения номера версии: %w", err)
	}

	// Генерируем ключ для версии
	storageKey := fmt.Sprintf("%s/v%d", file.StorageKey, nextVersion)

	// Создаём обёртку для подсчёта хеша
	hashReader := &hashingReader{reader: reader, hasher: sha256.New()}

	// Загружаем версию в S3/MinIO
	fileInfo, err := uc.storageClient.Upload(ctx, storageKey, hashReader, size, file.MimeType)
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки версии в хранилище: %w", err)
	}

	checksum := hex.EncodeToString(hashReader.hasher.Sum(nil))

	// Создаём запись о версии
	version := entities.NewFileVersion(
		input.FileID,
		nextVersion,
		storageKey,
		checksum,
		input.Comment,
		fileInfo.Size,
		input.UserID,
	)

	err = uc.versionRepo.Create(ctx, version)
	if err != nil {
		// Пытаемся удалить файл из хранилища при ошибке
		_ = uc.storageClient.Delete(ctx, storageKey)
		return nil, fmt.Errorf("ошибка сохранения версии: %w", err)
	}

	if uc.auditLogger != nil {
		uc.auditLogger.LogAuditEvent(ctx, "create_version", "file", map[string]interface{}{
			"file_id":        input.FileID,
			"version_number": nextVersion,
			"created_by":     input.UserID,
			"comment":        input.Comment,
		})
	}

	return uc.toVersionResponse(version), nil
}

// GetVersions получает все версии файла.
func (uc *VersionUseCase) GetVersions(ctx context.Context, fileID int64) ([]*dto.FileVersionResponse, error) {
	versions, err := uc.versionRepo.GetByFileMetadataID(ctx, fileID)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.FileVersionResponse, len(versions))
	for i, version := range versions {
		responses[i] = uc.toVersionResponse(version)
	}

	return responses, nil
}

// GetVersion получает конкретную версию файла.
func (uc *VersionUseCase) GetVersion(ctx context.Context, fileID int64, versionNumber int) (*dto.FileVersionResponse, error) {
	version, err := uc.versionRepo.GetByVersionNumber(ctx, fileID, versionNumber)
	if err != nil {
		return nil, err
	}

	return uc.toVersionResponse(version), nil
}

// GetLatestVersion получает последнюю версию файла.
func (uc *VersionUseCase) GetLatestVersion(ctx context.Context, fileID int64) (*dto.FileVersionResponse, error) {
	version, err := uc.versionRepo.GetLatestVersion(ctx, fileID)
	if err != nil {
		return nil, err
	}

	return uc.toVersionResponse(version), nil
}

// DownloadVersion возвращает данные для скачивания версии файла.
func (uc *VersionUseCase) DownloadVersion(ctx context.Context, fileID int64, versionNumber int) (*dto.DownloadResponse, error) {
	// Получаем метаданные файла
	file, err := uc.fileRepo.GetByID(ctx, fileID)
	if err != nil {
		return nil, err
	}

	// Получаем версию
	version, err := uc.versionRepo.GetByVersionNumber(ctx, fileID, versionNumber)
	if err != nil {
		return nil, err
	}

	// Генерируем presigned URL (действует 1 час)
	presignedURL, err := uc.storageClient.GetPresignedURL(ctx, version.StorageKey, time.Hour)
	if err != nil {
		return nil, fmt.Errorf("ошибка генерации URL для скачивания: %w", err)
	}

	return &dto.DownloadResponse{
		PresignedURL: presignedURL,
		FileName:     fmt.Sprintf("%s_v%d", file.OriginalName, versionNumber),
		MimeType:     file.MimeType,
		Size:         version.Size,
	}, nil
}

// DeleteVersion удаляет версию файла.
func (uc *VersionUseCase) DeleteVersion(ctx context.Context, versionID int64, userID int64) error {
	version, err := uc.versionRepo.GetByID(ctx, versionID)
	if err != nil {
		return err
	}

	// Получаем метаданные файла для проверки прав
	file, err := uc.fileRepo.GetByID(ctx, version.FileMetadataID)
	if err != nil {
		return err
	}

	// Проверяем права (только загрузивший может удалить)
	if file.UploadedBy != userID && version.CreatedBy != userID {
		return &PermissionError{Message: "нет прав на удаление версии"}
	}

	// Удаляем файл из хранилища
	err = uc.storageClient.Delete(ctx, version.StorageKey)
	if err != nil {
		return fmt.Errorf("ошибка удаления версии из хранилища: %w", err)
	}

	// Удаляем запись из БД
	err = uc.versionRepo.Delete(ctx, versionID)
	if err != nil {
		return err
	}

	if uc.auditLogger != nil {
		uc.auditLogger.LogAuditEvent(ctx, "delete_version", "file", map[string]interface{}{
			"version_id":     versionID,
			"file_id":        version.FileMetadataID,
			"version_number": version.VersionNumber,
			"deleted_by":     userID,
		})
	}

	return nil
}

// GetVersionCount возвращает количество версий файла.
func (uc *VersionUseCase) GetVersionCount(ctx context.Context, fileID int64) (int64, error) {
	return uc.versionRepo.CountByFileMetadataID(ctx, fileID)
}

// toVersionResponse конвертирует entity в DTO.
func (uc *VersionUseCase) toVersionResponse(version *entities.FileVersion) *dto.FileVersionResponse {
	return &dto.FileVersionResponse{
		ID:            version.ID,
		VersionNumber: version.VersionNumber,
		Size:          version.Size,
		Checksum:      version.Checksum,
		Comment:       version.Comment,
		CreatedBy:     version.CreatedBy,
		CreatedAt:     version.CreatedAt,
	}
}
