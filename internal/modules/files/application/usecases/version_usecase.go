// Package usecases содержит бизнес-логику модуля files.
package usecases

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	authDomain "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/files/application/dto"
	filesDomain "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/files/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/files/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/sanitization"
)

// VersionUseCase обрабатывает бизнес-логику версионирования файлов.
type VersionUseCase struct {
	fileRepo      FileMetadataRepository
	versionRepo   FileVersionRepository
	storageClient StorageClient
	auditLogger   AuditEventLogger
}

// NewVersionUseCase создаёт новый use case для версий файлов.
//
// Mirror NewFileUseCase — narrow interfaces accepted instead of concrete
// infrastructure types. Existing main.go wiring satisfies the ports
// structurally; nil values fall back to no-op behavior.
func NewVersionUseCase(
	fileRepo FileMetadataRepository,
	versionRepo FileVersionRepository,
	storageClient StorageClient,
	auditLogger AuditEventLogger,
) *VersionUseCase {
	uc := &VersionUseCase{
		fileRepo:    fileRepo,
		versionRepo: versionRepo,
	}
	if storageClient != nil {
		uc.storageClient = storageClient
	}
	if auditLogger != nil {
		uc.auditLogger = auditLogger
	}
	return uc
}

// CreateVersion создаёт новую версию файла.
//
// Closes #290 ADR-2 (CreateVersion ownership hijack): only the
// original uploader may push a new version. No admin write-override.
// Comment is sanitized — closes Tier 1 stored XSS finding (was
// persisted raw to version history + audit log).
func (uc *VersionUseCase) CreateVersion(ctx context.Context, reader io.Reader, size int64, input *dto.CreateVersionInput) (*dto.FileVersionResponse, error) {
	// Получаем метаданные файла
	file, err := uc.fileRepo.GetByID(ctx, input.FileID)
	if err != nil {
		return nil, err
	}

	// Ownership gate (#290 ADR-2)
	if err := filesDomain.AuthorizeFileAccess(input.UserID, input.UserRole, file, filesDomain.FileActionCreateVersion); err != nil {
		emitAccessDenied(ctx, uc.auditLogger, input.UserID, file.ID, filesDomain.FileActionCreateVersion, "create_version")
		return nil, err
	}

	// Файл должен быть прикреплён
	if file.IsTemporary {
		return nil, &ValidationError{Message: "нельзя создать версию временного файла"}
	}

	// Sanitize free-text comment — prevents stored XSS in version
	// history rendering + audit log surfacing (#290 Tier 1 §8).
	// SanitizeHTML strips <script>/<style>/<iframe>/event-handlers AND
	// escapes any remaining HTML; SanitizeString only handles whitespace.
	safeComment := sanitization.SanitizeHTML(input.Comment)

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
		safeComment,
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
			"comment":        safeComment,
		})
	}

	return uc.toVersionResponse(version), nil
}

// GetVersions получает все версии файла.
//
// Closes #290 reviewer T0-1 round 1: requires actor ownership of the
// parent file (or system_admin read override). Без этого гейта ответ
// leaked version metadata + comment trail чужого файла.
func (uc *VersionUseCase) GetVersions(ctx context.Context, fileID int64, actorID int64, actorRole authDomain.RoleType) ([]*dto.FileVersionResponse, error) {
	file, err := uc.fileRepo.GetByID(ctx, fileID)
	if err != nil {
		return nil, err
	}

	if err := filesDomain.AuthorizeFileAccess(actorID, actorRole, file, filesDomain.FileActionRead); err != nil {
		emitAccessDenied(ctx, uc.auditLogger, actorID, file.ID, filesDomain.FileActionRead, "list_versions")
		return nil, err
	}

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
//
// Closes #290 ADR-1 leak: previously took only (fileID, versionNumber)
// so any authenticated user could enumerate version URLs across files.
// Now gates through AuthorizeFileAccess(FileActionRead) — uploader или
// system_admin only.
func (uc *VersionUseCase) DownloadVersion(ctx context.Context, fileID int64, versionNumber int, actorID int64, actorRole authDomain.RoleType) (*dto.DownloadResponse, error) {
	// Получаем метаданные файла
	file, err := uc.fileRepo.GetByID(ctx, fileID)
	if err != nil {
		return nil, err
	}

	if err := filesDomain.AuthorizeFileAccess(actorID, actorRole, file, filesDomain.FileActionRead); err != nil {
		emitAccessDenied(ctx, uc.auditLogger, actorID, file.ID, filesDomain.FileActionRead, "download_version")
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
//
// Two-path authorisation: either the file's owner (mirrors broader
// file-write gate) OR the version's own creator (legitimate
// "remove my contribution" path — kept for backwards compatibility
// with previous behavior). Both paths funnel denial through the
// sentinel + audit emit pattern so the surface stays uniform.
// Closes #290 ADR-2 audit-emit gap (was silently returning struct
// PermissionError before).
func (uc *VersionUseCase) DeleteVersion(ctx context.Context, versionID int64, actorID int64, actorRole authDomain.RoleType) error {
	version, err := uc.versionRepo.GetByID(ctx, versionID)
	if err != nil {
		return err
	}

	// Получаем метаданные файла для проверки прав
	file, err := uc.fileRepo.GetByID(ctx, version.FileMetadataID)
	if err != nil {
		return err
	}

	fileOwnerOk := filesDomain.AuthorizeFileAccess(actorID, actorRole, file, filesDomain.FileActionDelete) == nil
	versionAuthorOk := version.CreatedBy == actorID
	if !fileOwnerOk && !versionAuthorOk {
		emitAccessDenied(ctx, uc.auditLogger, actorID, file.ID, filesDomain.FileActionDelete, "delete_version")
		return filesDomain.ErrFileAccessDenied
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
			"deleted_by":     actorID,
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
