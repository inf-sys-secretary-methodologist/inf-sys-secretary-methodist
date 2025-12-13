// Package usecases содержит бизнес-логику модуля files.
package usecases

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/files/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/files/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/files/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/storage"
)

// FileUseCase обрабатывает бизнес-логику управления файлами.
type FileUseCase struct {
	fileRepo       repositories.FileMetadataRepository
	versionRepo    repositories.FileVersionRepository
	s3Client       *storage.S3Client
	fileValidator  *storage.FileValidator
	auditLogger    *logging.AuditLogger
	tempExpiration time.Duration // Срок жизни временных файлов
}

// NewFileUseCase создаёт новый use case для файлов.
func NewFileUseCase(
	fileRepo repositories.FileMetadataRepository,
	versionRepo repositories.FileVersionRepository,
	s3Client *storage.S3Client,
	fileValidator *storage.FileValidator,
	auditLogger *logging.AuditLogger,
) *FileUseCase {
	return &FileUseCase{
		fileRepo:       fileRepo,
		versionRepo:    versionRepo,
		s3Client:       s3Client,
		fileValidator:  fileValidator,
		auditLogger:    auditLogger,
		tempExpiration: 24 * time.Hour, // Временные файлы живут 24 часа
	}
}

// UploadFile загружает файл в хранилище.
func (uc *FileUseCase) UploadFile(ctx context.Context, reader io.Reader, input *dto.UploadFileInput) (*dto.UploadResponse, error) {
	// Валидируем имя файла
	if uc.fileValidator != nil {
		_, err := uc.fileValidator.ValidateFileName(input.OriginalName)
		if err != nil {
			return nil, &ValidationError{Message: err.Error()}
		}
	}

	// Генерируем ключ для хранилища
	storageKey := storage.GenerateTempKey(input.UserID, input.OriginalName)

	// Создаём обёртку для подсчёта хеша
	hashReader := &hashingReader{reader: reader, hasher: sha256.New()}

	// Загружаем файл в S3/MinIO
	fileInfo, err := uc.s3Client.Upload(ctx, storageKey, hashReader, input.Size, input.MimeType)
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки файла в хранилище: %w", err)
	}

	checksum := hex.EncodeToString(hashReader.hasher.Sum(nil))

	// Создаём метаданные файла
	expires := time.Now().Add(uc.tempExpiration)
	fileMeta := entities.NewFileMetadata(
		input.OriginalName,
		storageKey,
		input.MimeType,
		checksum,
		fileInfo.Size,
		input.UserID,
	)
	fileMeta.ExpiresAt = &expires

	// Сохраняем метаданные в БД
	err = uc.fileRepo.Create(ctx, fileMeta)
	if err != nil {
		// Пытаемся удалить файл из хранилища при ошибке
		_ = uc.s3Client.Delete(ctx, storageKey)
		return nil, fmt.Errorf("ошибка сохранения метаданных файла: %w", err)
	}

	if uc.auditLogger != nil {
		uc.auditLogger.LogAuditEvent(ctx, "upload", "file", map[string]interface{}{
			"file_id":       fileMeta.ID,
			"original_name": fileMeta.OriginalName,
			"size":          fileMeta.Size,
			"uploaded_by":   input.UserID,
		})
	}

	return &dto.UploadResponse{
		FileID:       fileMeta.ID,
		OriginalName: fileMeta.OriginalName,
		Size:         fileMeta.Size,
		MimeType:     fileMeta.MimeType,
		Checksum:     fileMeta.Checksum,
	}, nil
}

// GetFile получает информацию о файле по ID.
func (uc *FileUseCase) GetFile(ctx context.Context, id int64) (*dto.FileResponse, error) {
	file, err := uc.fileRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return uc.toFileResponse(file), nil
}

// GetFileWithDownloadURL получает информацию о файле с URL для скачивания.
func (uc *FileUseCase) GetFileWithDownloadURL(ctx context.Context, id int64, urlExpiration time.Duration) (*dto.FileResponse, error) {
	file, err := uc.fileRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	response := uc.toFileResponse(file)

	// Генерируем presigned URL для скачивания
	downloadURL, err := uc.s3Client.GetPresignedURL(ctx, file.StorageKey, urlExpiration)
	if err != nil {
		return nil, fmt.Errorf("ошибка генерации URL для скачивания: %w", err)
	}
	response.DownloadURL = downloadURL

	return response, nil
}

// DownloadFile возвращает данные для скачивания файла.
func (uc *FileUseCase) DownloadFile(ctx context.Context, id int64) (*dto.DownloadResponse, error) {
	file, err := uc.fileRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Генерируем presigned URL (действует 1 час)
	presignedURL, err := uc.s3Client.GetPresignedURL(ctx, file.StorageKey, time.Hour)
	if err != nil {
		return nil, fmt.Errorf("ошибка генерации URL для скачивания: %w", err)
	}

	return &dto.DownloadResponse{
		PresignedURL: presignedURL,
		FileName:     file.OriginalName,
		MimeType:     file.MimeType,
		Size:         file.Size,
	}, nil
}

// AttachFile прикрепляет файл к документу, задаче или объявлению.
func (uc *FileUseCase) AttachFile(ctx context.Context, input *dto.AttachFileInput) error {
	file, err := uc.fileRepo.GetByID(ctx, input.FileID)
	if err != nil {
		return err
	}

	// Проверяем, что файл временный
	if !file.IsTemporary {
		return &ValidationError{Message: "файл уже прикреплён"}
	}

	// Прикрепляем к соответствующей сущности
	if input.DocumentID != nil {
		file.AttachToDocument(*input.DocumentID)
	} else if input.TaskID != nil {
		file.AttachToTask(*input.TaskID)
	} else if input.AnnouncementID != nil {
		file.AttachToAnnouncement(*input.AnnouncementID)
	} else {
		return &ValidationError{Message: "необходимо указать document_id, task_id или announcement_id"}
	}

	err = uc.fileRepo.Update(ctx, file)
	if err != nil {
		return fmt.Errorf("ошибка обновления метаданных файла: %w", err)
	}

	if uc.auditLogger != nil {
		uc.auditLogger.LogAuditEvent(ctx, "attach", "file", map[string]interface{}{
			"file_id":         input.FileID,
			"document_id":     input.DocumentID,
			"task_id":         input.TaskID,
			"announcement_id": input.AnnouncementID,
		})
	}

	return nil
}

// DeleteFile удаляет файл (мягкое удаление).
func (uc *FileUseCase) DeleteFile(ctx context.Context, id int64, userID int64) error {
	file, err := uc.fileRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Проверяем права (только загрузивший может удалить)
	if file.UploadedBy != userID {
		return &PermissionError{Message: "нет прав на удаление файла"}
	}

	err = uc.fileRepo.Delete(ctx, id)
	if err != nil {
		return err
	}

	if uc.auditLogger != nil {
		uc.auditLogger.LogAuditEvent(ctx, "delete", "file", map[string]interface{}{
			"file_id":       id,
			"original_name": file.OriginalName,
			"deleted_by":    userID,
		})
	}

	return nil
}

// ListFiles получает список файлов с пагинацией.
func (uc *FileUseCase) ListFiles(ctx context.Context, page, limit int) (*dto.FileListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	files, err := uc.fileRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	total, err := uc.fileRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	responses := make([]*dto.FileResponse, len(files))
	for i, file := range files {
		responses[i] = uc.toFileResponse(file)
	}

	return &dto.FileListResponse{
		Files:      responses,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// GetFilesByDocument получает все файлы документа.
func (uc *FileUseCase) GetFilesByDocument(ctx context.Context, documentID int64) ([]*dto.FileResponse, error) {
	files, err := uc.fileRepo.GetByDocumentID(ctx, documentID)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.FileResponse, len(files))
	for i, file := range files {
		responses[i] = uc.toFileResponse(file)
	}

	return responses, nil
}

// GetFilesByTask получает все файлы задачи.
func (uc *FileUseCase) GetFilesByTask(ctx context.Context, taskID int64) ([]*dto.FileResponse, error) {
	files, err := uc.fileRepo.GetByTaskID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.FileResponse, len(files))
	for i, file := range files {
		responses[i] = uc.toFileResponse(file)
	}

	return responses, nil
}

// GetFilesByAnnouncement получает все файлы объявления.
func (uc *FileUseCase) GetFilesByAnnouncement(ctx context.Context, announcementID int64) ([]*dto.FileResponse, error) {
	files, err := uc.fileRepo.GetByAnnouncementID(ctx, announcementID)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.FileResponse, len(files))
	for i, file := range files {
		responses[i] = uc.toFileResponse(file)
	}

	return responses, nil
}

// CleanupExpiredFiles удаляет временные файлы с истёкшим сроком.
func (uc *FileUseCase) CleanupExpiredFiles(ctx context.Context) (int64, error) {
	// Получаем список истёкших файлов для удаления из хранилища
	expiredFiles, err := uc.fileRepo.GetExpiredTemporaryFiles(ctx, 100)
	if err != nil {
		return 0, err
	}

	// Удаляем файлы из хранилища
	for _, file := range expiredFiles {
		_ = uc.s3Client.Delete(ctx, file.StorageKey)
	}

	// Помечаем как удалённые в БД
	count, err := uc.fileRepo.CleanupExpired(ctx)
	if err != nil {
		return 0, err
	}

	if uc.auditLogger != nil && count > 0 {
		uc.auditLogger.LogAuditEvent(ctx, "cleanup", "files", map[string]interface{}{
			"deleted_count": count,
		})
	}

	return count, nil
}

// toFileResponse конвертирует entity в DTO.
func (uc *FileUseCase) toFileResponse(file *entities.FileMetadata) *dto.FileResponse {
	return &dto.FileResponse{
		ID:             file.ID,
		OriginalName:   file.OriginalName,
		Size:           file.Size,
		MimeType:       file.MimeType,
		Checksum:       file.Checksum,
		UploadedBy:     file.UploadedBy,
		DocumentID:     file.DocumentID,
		TaskID:         file.TaskID,
		AnnouncementID: file.AnnouncementID,
		IsTemporary:    file.IsTemporary,
		CreatedAt:      file.CreatedAt,
		UpdatedAt:      file.UpdatedAt,
	}
}

// hashingReader - обёртка для подсчёта хеша при чтении.
type hashingReader struct {
	reader io.Reader
	hasher hash.Hash
}

func (hr *hashingReader) Read(p []byte) (n int, err error) {
	n, err = hr.reader.Read(p)
	if n > 0 {
		hr.hasher.Write(p[:n])
	}
	return
}

// ValidationError ошибка валидации.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// PermissionError ошибка доступа.
type PermissionError struct {
	Message string
}

func (e *PermissionError) Error() string {
	return e.Message
}
