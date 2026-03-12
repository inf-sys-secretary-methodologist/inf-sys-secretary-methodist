// Package persistence реализует интерфейсы репозиториев модуля files.
package persistence

import (
	"context"
	"database/sql"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/files/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/files/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/database"
)

// FileVersionRepositoryPG реализует PostgreSQL репозиторий версий файлов.
type FileVersionRepositoryPG struct {
	db *sql.DB
}

// NewFileVersionRepositoryPG создаёт новый PostgreSQL репозиторий версий файлов.
func NewFileVersionRepositoryPG(db *sql.DB) repositories.FileVersionRepository {
	return &FileVersionRepositoryPG{db: db}
}

// Create сохраняет новую запись версии файла.
func (r *FileVersionRepositoryPG) Create(ctx context.Context, version *entities.FileVersion) error {
	query := `
		INSERT INTO file_versions (file_metadata_id, version_number, storage_key, size, checksum, comment, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`
	err := r.db.QueryRowContext(ctx, query,
		version.FileMetadataID,
		version.VersionNumber,
		version.StorageKey,
		version.Size,
		version.Checksum,
		version.Comment,
		version.CreatedBy,
		version.CreatedAt,
	).Scan(&version.ID)

	if err != nil {
		return database.MapPostgresError(err)
	}
	return nil
}

// GetByID получает версию файла по ID.
func (r *FileVersionRepositoryPG) GetByID(ctx context.Context, id int64) (*entities.FileVersion, error) {
	version := &entities.FileVersion{}
	query := `
		SELECT id, file_metadata_id, version_number, storage_key, size, checksum, comment, created_by, created_at
		FROM file_versions
		WHERE id = $1
	`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&version.ID,
		&version.FileMetadataID,
		&version.VersionNumber,
		&version.StorageKey,
		&version.Size,
		&version.Checksum,
		&version.Comment,
		&version.CreatedBy,
		&version.CreatedAt,
	)

	if err != nil {
		return nil, database.MapPostgresError(err)
	}
	return version, nil
}

// GetByFileMetadataID получает все версии файла.
func (r *FileVersionRepositoryPG) GetByFileMetadataID(ctx context.Context, fileMetadataID int64) ([]*entities.FileVersion, error) {
	query := `
		SELECT id, file_metadata_id, version_number, storage_key, size, checksum, comment, created_by, created_at
		FROM file_versions
		WHERE file_metadata_id = $1
		ORDER BY version_number DESC
	`
	return r.queryVersions(ctx, query, fileMetadataID)
}

// GetLatestVersion получает последнюю версию файла.
func (r *FileVersionRepositoryPG) GetLatestVersion(ctx context.Context, fileMetadataID int64) (*entities.FileVersion, error) {
	version := &entities.FileVersion{}
	query := `
		SELECT id, file_metadata_id, version_number, storage_key, size, checksum, comment, created_by, created_at
		FROM file_versions
		WHERE file_metadata_id = $1
		ORDER BY version_number DESC
		LIMIT 1
	`
	err := r.db.QueryRowContext(ctx, query, fileMetadataID).Scan(
		&version.ID,
		&version.FileMetadataID,
		&version.VersionNumber,
		&version.StorageKey,
		&version.Size,
		&version.Checksum,
		&version.Comment,
		&version.CreatedBy,
		&version.CreatedAt,
	)

	if err != nil {
		return nil, database.MapPostgresError(err)
	}
	return version, nil
}

// GetByVersionNumber получает конкретную версию по номеру.
func (r *FileVersionRepositoryPG) GetByVersionNumber(ctx context.Context, fileMetadataID int64, versionNumber int) (*entities.FileVersion, error) {
	version := &entities.FileVersion{}
	query := `
		SELECT id, file_metadata_id, version_number, storage_key, size, checksum, comment, created_by, created_at
		FROM file_versions
		WHERE file_metadata_id = $1 AND version_number = $2
	`
	err := r.db.QueryRowContext(ctx, query, fileMetadataID, versionNumber).Scan(
		&version.ID,
		&version.FileMetadataID,
		&version.VersionNumber,
		&version.StorageKey,
		&version.Size,
		&version.Checksum,
		&version.Comment,
		&version.CreatedBy,
		&version.CreatedAt,
	)

	if err != nil {
		return nil, database.MapPostgresError(err)
	}
	return version, nil
}

// Delete удаляет запись версии файла.
func (r *FileVersionRepositoryPG) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM file_versions WHERE id = $1`, id)
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

// DeleteByFileMetadataID удаляет все версии файла.
func (r *FileVersionRepositoryPG) DeleteByFileMetadataID(ctx context.Context, fileMetadataID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM file_versions WHERE file_metadata_id = $1`, fileMetadataID)
	if err != nil {
		return database.MapPostgresError(err)
	}
	return nil
}

// CountByFileMetadataID возвращает количество версий файла.
func (r *FileVersionRepositoryPG) CountByFileMetadataID(ctx context.Context, fileMetadataID int64) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM file_versions WHERE file_metadata_id = $1`, fileMetadataID).Scan(&count)
	if err != nil {
		return 0, database.MapPostgresError(err)
	}
	return count, nil
}

// GetNextVersionNumber возвращает следующий номер версии для файла.
func (r *FileVersionRepositoryPG) GetNextVersionNumber(ctx context.Context, fileMetadataID int64) (int, error) {
	var maxVersion sql.NullInt64
	err := r.db.QueryRowContext(ctx,
		`SELECT MAX(version_number) FROM file_versions WHERE file_metadata_id = $1`,
		fileMetadataID,
	).Scan(&maxVersion)
	if err != nil {
		return 0, database.MapPostgresError(err)
	}

	if !maxVersion.Valid {
		return 1, nil
	}
	return int(maxVersion.Int64) + 1, nil
}

// queryVersions - вспомогательный метод для выполнения запросов к версиям.
func (r *FileVersionRepositoryPG) queryVersions(ctx context.Context, query string, args ...interface{}) ([]*entities.FileVersion, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, database.MapPostgresError(err)
	}
	defer func() { _ = rows.Close() }()

	versions := []*entities.FileVersion{}
	for rows.Next() {
		version := &entities.FileVersion{}
		if err := rows.Scan(
			&version.ID,
			&version.FileMetadataID,
			&version.VersionNumber,
			&version.StorageKey,
			&version.Size,
			&version.Checksum,
			&version.Comment,
			&version.CreatedBy,
			&version.CreatedAt,
		); err != nil {
			return nil, database.MapPostgresError(err)
		}
		versions = append(versions, version)
	}

	if err := rows.Err(); err != nil {
		return nil, database.MapPostgresError(err)
	}

	return versions, nil
}
