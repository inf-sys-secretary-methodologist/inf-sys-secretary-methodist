// Package persistence реализует интерфейсы репозиториев модуля files.
package persistence

import (
	"context"
	"database/sql"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/files/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/files/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/database"
)

// FileMetadataRepositoryPG реализует PostgreSQL репозиторий метаданных файлов.
type FileMetadataRepositoryPG struct {
	db *sql.DB
}

// NewFileMetadataRepositoryPG создаёт новый PostgreSQL репозиторий метаданных файлов.
func NewFileMetadataRepositoryPG(db *sql.DB) repositories.FileMetadataRepository {
	return &FileMetadataRepositoryPG{db: db}
}

// Create сохраняет новую запись метаданных файла.
func (r *FileMetadataRepositoryPG) Create(ctx context.Context, file *entities.FileMetadata) error {
	query := `
		INSERT INTO file_metadata (
			original_name, storage_key, size, mime_type, checksum, uploaded_by,
			document_id, task_id, announcement_id, is_temporary, expires_at,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id
	`
	err := r.db.QueryRowContext(ctx, query,
		file.OriginalName,
		file.StorageKey,
		file.Size,
		file.MimeType,
		file.Checksum,
		file.UploadedBy,
		file.DocumentID,
		file.TaskID,
		file.AnnouncementID,
		file.IsTemporary,
		file.ExpiresAt,
		file.CreatedAt,
		file.UpdatedAt,
	).Scan(&file.ID)

	if err != nil {
		return database.MapPostgresError(err)
	}
	return nil
}

// GetByID получает метаданные файла по ID.
func (r *FileMetadataRepositoryPG) GetByID(ctx context.Context, id int64) (*entities.FileMetadata, error) {
	file := &entities.FileMetadata{}
	query := `
		SELECT id, original_name, storage_key, size, mime_type, checksum, uploaded_by,
			document_id, task_id, announcement_id, is_temporary, expires_at,
			created_at, updated_at, deleted_at
		FROM file_metadata
		WHERE id = $1 AND deleted_at IS NULL
	`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&file.ID,
		&file.OriginalName,
		&file.StorageKey,
		&file.Size,
		&file.MimeType,
		&file.Checksum,
		&file.UploadedBy,
		&file.DocumentID,
		&file.TaskID,
		&file.AnnouncementID,
		&file.IsTemporary,
		&file.ExpiresAt,
		&file.CreatedAt,
		&file.UpdatedAt,
		&file.DeletedAt,
	)

	if err != nil {
		return nil, database.MapPostgresError(err)
	}
	return file, nil
}

// GetByStorageKey получает метаданные файла по ключу хранилища.
func (r *FileMetadataRepositoryPG) GetByStorageKey(ctx context.Context, storageKey string) (*entities.FileMetadata, error) {
	file := &entities.FileMetadata{}
	query := `
		SELECT id, original_name, storage_key, size, mime_type, checksum, uploaded_by,
			document_id, task_id, announcement_id, is_temporary, expires_at,
			created_at, updated_at, deleted_at
		FROM file_metadata
		WHERE storage_key = $1 AND deleted_at IS NULL
	`
	err := r.db.QueryRowContext(ctx, query, storageKey).Scan(
		&file.ID,
		&file.OriginalName,
		&file.StorageKey,
		&file.Size,
		&file.MimeType,
		&file.Checksum,
		&file.UploadedBy,
		&file.DocumentID,
		&file.TaskID,
		&file.AnnouncementID,
		&file.IsTemporary,
		&file.ExpiresAt,
		&file.CreatedAt,
		&file.UpdatedAt,
		&file.DeletedAt,
	)

	if err != nil {
		return nil, database.MapPostgresError(err)
	}
	return file, nil
}

// Update обновляет существующую запись метаданных файла.
func (r *FileMetadataRepositoryPG) Update(ctx context.Context, file *entities.FileMetadata) error {
	query := `
		UPDATE file_metadata
		SET original_name = $1, storage_key = $2, size = $3, mime_type = $4, checksum = $5,
			document_id = $6, task_id = $7, announcement_id = $8, is_temporary = $9,
			expires_at = $10, updated_at = $11
		WHERE id = $12 AND deleted_at IS NULL
	`
	result, err := r.db.ExecContext(ctx, query,
		file.OriginalName,
		file.StorageKey,
		file.Size,
		file.MimeType,
		file.Checksum,
		file.DocumentID,
		file.TaskID,
		file.AnnouncementID,
		file.IsTemporary,
		file.ExpiresAt,
		file.UpdatedAt,
		file.ID,
	)
	if err != nil {
		return database.MapPostgresError(err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return database.MapPostgresError(err)
	}

	if rows == 0 {
		return database.MapPostgresError(sql.ErrNoRows)
	}

	return nil
}

// Delete выполняет мягкое удаление метаданных файла.
func (r *FileMetadataRepositoryPG) Delete(ctx context.Context, id int64) error {
	query := `UPDATE file_metadata SET deleted_at = $1, updated_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return database.MapPostgresError(err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return database.MapPostgresError(err)
	}

	if rows == 0 {
		return database.MapPostgresError(sql.ErrNoRows)
	}

	return nil
}

// HardDelete полностью удаляет запись метаданных файла.
func (r *FileMetadataRepositoryPG) HardDelete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM file_metadata WHERE id = $1`, id)
	if err != nil {
		return database.MapPostgresError(err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return database.MapPostgresError(err)
	}

	if rows == 0 {
		return database.MapPostgresError(sql.ErrNoRows)
	}

	return nil
}

// List получает список файлов с пагинацией.
func (r *FileMetadataRepositoryPG) List(ctx context.Context, limit, offset int) ([]*entities.FileMetadata, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	if limit > 100 {
		limit = 100
	}

	query := `
		SELECT id, original_name, storage_key, size, mime_type, checksum, uploaded_by,
			document_id, task_id, announcement_id, is_temporary, expires_at,
			created_at, updated_at, deleted_at
		FROM file_metadata
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	return r.queryFiles(ctx, query, limit, offset)
}

// Count возвращает общее количество неудалённых файлов.
func (r *FileMetadataRepositoryPG) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM file_metadata WHERE deleted_at IS NULL`).Scan(&count)
	if err != nil {
		return 0, database.MapPostgresError(err)
	}
	return count, nil
}

// GetByDocumentID получает все файлы, прикреплённые к документу.
func (r *FileMetadataRepositoryPG) GetByDocumentID(ctx context.Context, documentID int64) ([]*entities.FileMetadata, error) {
	query := `
		SELECT id, original_name, storage_key, size, mime_type, checksum, uploaded_by,
			document_id, task_id, announcement_id, is_temporary, expires_at,
			created_at, updated_at, deleted_at
		FROM file_metadata
		WHERE document_id = $1 AND deleted_at IS NULL
		ORDER BY created_at ASC
	`
	return r.queryFilesSingleArg(ctx, query, documentID)
}

// GetByTaskID получает все файлы, прикреплённые к задаче.
func (r *FileMetadataRepositoryPG) GetByTaskID(ctx context.Context, taskID int64) ([]*entities.FileMetadata, error) {
	query := `
		SELECT id, original_name, storage_key, size, mime_type, checksum, uploaded_by,
			document_id, task_id, announcement_id, is_temporary, expires_at,
			created_at, updated_at, deleted_at
		FROM file_metadata
		WHERE task_id = $1 AND deleted_at IS NULL
		ORDER BY created_at ASC
	`
	return r.queryFilesSingleArg(ctx, query, taskID)
}

// GetByAnnouncementID получает все файлы, прикреплённые к объявлению.
func (r *FileMetadataRepositoryPG) GetByAnnouncementID(ctx context.Context, announcementID int64) ([]*entities.FileMetadata, error) {
	query := `
		SELECT id, original_name, storage_key, size, mime_type, checksum, uploaded_by,
			document_id, task_id, announcement_id, is_temporary, expires_at,
			created_at, updated_at, deleted_at
		FROM file_metadata
		WHERE announcement_id = $1 AND deleted_at IS NULL
		ORDER BY created_at ASC
	`
	return r.queryFilesSingleArg(ctx, query, announcementID)
}

// GetByUploadedBy получает все файлы, загруженные пользователем.
func (r *FileMetadataRepositoryPG) GetByUploadedBy(ctx context.Context, userID int64, limit, offset int) ([]*entities.FileMetadata, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	if limit > 100 {
		limit = 100
	}

	query := `
		SELECT id, original_name, storage_key, size, mime_type, checksum, uploaded_by,
			document_id, task_id, announcement_id, is_temporary, expires_at,
			created_at, updated_at, deleted_at
		FROM file_metadata
		WHERE uploaded_by = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	return r.queryFiles(ctx, query, userID, limit, offset)
}

// GetExpiredTemporaryFiles получает временные файлы с истёкшим сроком.
func (r *FileMetadataRepositoryPG) GetExpiredTemporaryFiles(ctx context.Context, limit int) ([]*entities.FileMetadata, error) {
	if limit <= 0 {
		limit = 100
	}

	query := `
		SELECT id, original_name, storage_key, size, mime_type, checksum, uploaded_by,
			document_id, task_id, announcement_id, is_temporary, expires_at,
			created_at, updated_at, deleted_at
		FROM file_metadata
		WHERE is_temporary = true AND expires_at < $1 AND deleted_at IS NULL
		ORDER BY expires_at ASC
		LIMIT $2
	`
	return r.queryFiles(ctx, query, time.Now(), limit)
}

// CleanupExpired удаляет временные файлы с истёкшим сроком.
func (r *FileMetadataRepositoryPG) CleanupExpired(ctx context.Context) (int64, error) {
	query := `
		UPDATE file_metadata
		SET deleted_at = $1, updated_at = $1
		WHERE is_temporary = true AND expires_at < $1 AND deleted_at IS NULL
	`
	result, err := r.db.ExecContext(ctx, query, time.Now())
	if err != nil {
		return 0, database.MapPostgresError(err)
	}

	return result.RowsAffected()
}

// queryFiles - вспомогательный метод для выполнения запросов к файлам.
func (r *FileMetadataRepositoryPG) queryFiles(ctx context.Context, query string, args ...interface{}) ([]*entities.FileMetadata, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, database.MapPostgresError(err)
	}
	defer rows.Close()

	files := []*entities.FileMetadata{}
	for rows.Next() {
		file := &entities.FileMetadata{}
		if err := rows.Scan(
			&file.ID,
			&file.OriginalName,
			&file.StorageKey,
			&file.Size,
			&file.MimeType,
			&file.Checksum,
			&file.UploadedBy,
			&file.DocumentID,
			&file.TaskID,
			&file.AnnouncementID,
			&file.IsTemporary,
			&file.ExpiresAt,
			&file.CreatedAt,
			&file.UpdatedAt,
			&file.DeletedAt,
		); err != nil {
			return nil, database.MapPostgresError(err)
		}
		files = append(files, file)
	}

	if err := rows.Err(); err != nil {
		return nil, database.MapPostgresError(err)
	}

	return files, nil
}

// queryFilesSingleArg - вспомогательный метод для запросов с одним аргументом.
func (r *FileMetadataRepositoryPG) queryFilesSingleArg(ctx context.Context, query string, arg interface{}) ([]*entities.FileMetadata, error) {
	return r.queryFiles(ctx, query, arg)
}
