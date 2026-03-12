// Package persistence provides database implementations for document repositories.
package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/repositories"
	"github.com/lib/pq"
)

// DocumentRepositoryPG implements DocumentRepository using PostgreSQL
type DocumentRepositoryPG struct {
	db *sql.DB
}

// NewDocumentRepositoryPG creates a new PostgreSQL document repository
func NewDocumentRepositoryPG(db *sql.DB) *DocumentRepositoryPG {
	return &DocumentRepositoryPG{db: db}
}

// Create inserts a new document
func (r *DocumentRepositoryPG) Create(ctx context.Context, doc *entities.Document) error {
	var metadataJSON interface{} = nil // Use interface{} to properly pass NULL to jsonb
	var err error
	if doc.Metadata != nil {
		metadataJSON, err = json.Marshal(doc.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	query := `
		INSERT INTO documents (
			document_type_id, category_id, registration_number, registration_date,
			title, subject, content, author_id, author_department, author_position,
			recipient_id, recipient_department, recipient_position, recipient_external,
			status, file_name, file_path, file_size, mime_type, version,
			parent_document_id, deadline, execution_date, metadata, is_public, importance,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28
		) RETURNING id`

	err = r.db.QueryRowContext(ctx, query,
		doc.DocumentTypeID, doc.CategoryID, doc.RegistrationNumber, doc.RegistrationDate,
		doc.Title, doc.Subject, doc.Content, doc.AuthorID, doc.AuthorDepartment, doc.AuthorPosition,
		doc.RecipientID, doc.RecipientDepartment, doc.RecipientPosition, doc.RecipientExternal,
		doc.Status, doc.FileName, doc.FilePath, doc.FileSize, doc.MimeType, doc.Version,
		doc.ParentDocumentID, doc.Deadline, doc.ExecutionDate, metadataJSON, doc.IsPublic, doc.Importance,
		doc.CreatedAt, doc.UpdatedAt,
	).Scan(&doc.ID)

	if err != nil {
		return fmt.Errorf("failed to create document: %w", err)
	}
	return nil
}

// Update updates an existing document
func (r *DocumentRepositoryPG) Update(ctx context.Context, doc *entities.Document) error {
	doc.UpdatedAt = time.Now()

	var metadataJSON interface{} = nil // Use interface{} to properly pass NULL to jsonb
	var err error
	if doc.Metadata != nil {
		metadataJSON, err = json.Marshal(doc.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	query := `
		UPDATE documents SET
			document_type_id = $1, category_id = $2, registration_number = $3, registration_date = $4,
			title = $5, subject = $6, content = $7, author_department = $8, author_position = $9,
			recipient_id = $10, recipient_department = $11, recipient_position = $12, recipient_external = $13,
			status = $14, file_name = $15, file_path = $16, file_size = $17, mime_type = $18, version = $19,
			deadline = $20, execution_date = $21, metadata = $22, is_public = $23, importance = $24,
			updated_at = $25
		WHERE id = $26 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query,
		doc.DocumentTypeID, doc.CategoryID, doc.RegistrationNumber, doc.RegistrationDate,
		doc.Title, doc.Subject, doc.Content, doc.AuthorDepartment, doc.AuthorPosition,
		doc.RecipientID, doc.RecipientDepartment, doc.RecipientPosition, doc.RecipientExternal,
		doc.Status, doc.FileName, doc.FilePath, doc.FileSize, doc.MimeType, doc.Version,
		doc.Deadline, doc.ExecutionDate, metadataJSON, doc.IsPublic, doc.Importance,
		doc.UpdatedAt, doc.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("document not found")
	}
	return nil
}

// GetByID retrieves a document by ID
func (r *DocumentRepositoryPG) GetByID(ctx context.Context, id int64) (*entities.Document, error) {
	query := `
		SELECT d.id, d.document_type_id, d.category_id, d.registration_number, d.registration_date,
			d.title, d.subject, d.content, d.author_id, d.author_department, d.author_position,
			d.recipient_id, d.recipient_department, d.recipient_position, d.recipient_external,
			d.status, d.file_name, d.file_path, d.file_size, d.mime_type, d.version,
			d.parent_document_id, d.deadline, d.execution_date, d.metadata, d.is_public, d.importance,
			d.created_at, d.updated_at, d.deleted_at,
			author.name as author_name, recipient.name as recipient_name
		FROM documents d
		LEFT JOIN users author ON d.author_id = author.id
		LEFT JOIN users recipient ON d.recipient_id = recipient.id
		WHERE d.id = $1 AND d.deleted_at IS NULL`

	doc := &entities.Document{}
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&doc.ID, &doc.DocumentTypeID, &doc.CategoryID, &doc.RegistrationNumber, &doc.RegistrationDate,
		&doc.Title, &doc.Subject, &doc.Content, &doc.AuthorID, &doc.AuthorDepartment, &doc.AuthorPosition,
		&doc.RecipientID, &doc.RecipientDepartment, &doc.RecipientPosition, &doc.RecipientExternal,
		&doc.Status, &doc.FileName, &doc.FilePath, &doc.FileSize, &doc.MimeType, &doc.Version,
		&doc.ParentDocumentID, &doc.Deadline, &doc.ExecutionDate, &metadataJSON, &doc.IsPublic, &doc.Importance,
		&doc.CreatedAt, &doc.UpdatedAt, &doc.DeletedAt,
		&doc.AuthorName, &doc.RecipientName,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("document not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	if metadataJSON != nil {
		if err := json.Unmarshal(metadataJSON, &doc.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return doc, nil
}

// Delete hard deletes a document
func (r *DocumentRepositoryPG) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM documents WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// SoftDelete marks a document as deleted
func (r *DocumentRepositoryPG) SoftDelete(ctx context.Context, id int64) error {
	query := `UPDATE documents SET deleted_at = $1, updated_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to soft delete document: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("document not found")
	}
	return nil
}

// List retrieves documents with filters and access control
func (r *DocumentRepositoryPG) List(ctx context.Context, filter repositories.DocumentFilter) ([]*entities.Document, int64, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	if !filter.IncludeDeleted {
		conditions = append(conditions, "d.deleted_at IS NULL")
	}

	// Access control: user can see documents if:
	// 1. They are the author
	// 2. Document is public
	// 3. They have explicit permission via document_permissions
	// 4. They are admin (can see all)
	if filter.CurrentUserID > 0 && filter.CurrentUserRole != "admin" {
		accessCondition := fmt.Sprintf(`(
			d.author_id = $%d
			OR d.is_public = true
			OR EXISTS (
				SELECT 1 FROM document_permissions dp
				WHERE dp.document_id = d.id
				AND (dp.expires_at IS NULL OR dp.expires_at > NOW())
				AND (dp.user_id = $%d OR dp.role = $%d)
			)
		)`, argIndex, argIndex, argIndex+1)
		conditions = append(conditions, accessCondition)
		args = append(args, filter.CurrentUserID, filter.CurrentUserRole)
		argIndex += 2
	}

	if filter.AuthorID != nil {
		conditions = append(conditions, fmt.Sprintf("d.author_id = $%d", argIndex))
		args = append(args, *filter.AuthorID)
		argIndex++
	}
	if filter.RecipientID != nil {
		conditions = append(conditions, fmt.Sprintf("d.recipient_id = $%d", argIndex))
		args = append(args, *filter.RecipientID)
		argIndex++
	}
	if filter.DocumentTypeID != nil {
		conditions = append(conditions, fmt.Sprintf("d.document_type_id = $%d", argIndex))
		args = append(args, *filter.DocumentTypeID)
		argIndex++
	}
	if filter.CategoryID != nil {
		conditions = append(conditions, fmt.Sprintf("d.category_id = $%d", argIndex))
		args = append(args, *filter.CategoryID)
		argIndex++
	}
	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("d.status = $%d", argIndex))
		args = append(args, *filter.Status)
		argIndex++
	}
	if filter.Importance != nil {
		conditions = append(conditions, fmt.Sprintf("d.importance = $%d", argIndex))
		args = append(args, *filter.Importance)
		argIndex++
	}
	if filter.IsPublic != nil {
		conditions = append(conditions, fmt.Sprintf("d.is_public = $%d", argIndex))
		args = append(args, *filter.IsPublic)
		argIndex++
	}
	if filter.SearchQuery != nil && *filter.SearchQuery != "" {
		conditions = append(conditions, fmt.Sprintf("(d.title ILIKE $%d OR d.subject ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+*filter.SearchQuery+"%")
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM documents d %s", whereClause)
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count documents: %w", err)
	}

	// Get documents
	orderBy := "created_at DESC"
	if filter.OrderBy != "" {
		orderBy = filter.OrderBy
	}

	query := fmt.Sprintf(`
		SELECT d.id, d.document_type_id, d.category_id, d.registration_number, d.registration_date,
			d.title, d.subject, d.content, d.author_id, d.author_department, d.author_position,
			d.recipient_id, d.recipient_department, d.recipient_position, d.recipient_external,
			d.status, d.file_name, d.file_path, d.file_size, d.mime_type, d.version,
			d.parent_document_id, d.deadline, d.execution_date, d.metadata, d.is_public, d.importance,
			d.created_at, d.updated_at, d.deleted_at,
			author.name as author_name, recipient.name as recipient_name
		FROM documents d
		LEFT JOIN users author ON d.author_id = author.id
		LEFT JOIN users recipient ON d.recipient_id = recipient.id
		%s ORDER BY d.%s LIMIT $%d OFFSET $%d`,
		whereClause, orderBy, argIndex, argIndex+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list documents: %w", err)
	}
	defer rows.Close()

	var docs []*entities.Document
	for rows.Next() {
		doc := &entities.Document{}
		var metadataJSON []byte

		err := rows.Scan(
			&doc.ID, &doc.DocumentTypeID, &doc.CategoryID, &doc.RegistrationNumber, &doc.RegistrationDate,
			&doc.Title, &doc.Subject, &doc.Content, &doc.AuthorID, &doc.AuthorDepartment, &doc.AuthorPosition,
			&doc.RecipientID, &doc.RecipientDepartment, &doc.RecipientPosition, &doc.RecipientExternal,
			&doc.Status, &doc.FileName, &doc.FilePath, &doc.FileSize, &doc.MimeType, &doc.Version,
			&doc.ParentDocumentID, &doc.Deadline, &doc.ExecutionDate, &metadataJSON, &doc.IsPublic, &doc.Importance,
			&doc.CreatedAt, &doc.UpdatedAt, &doc.DeletedAt,
			&doc.AuthorName, &doc.RecipientName,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan document: %w", err)
		}

		if metadataJSON != nil {
			if err := json.Unmarshal(metadataJSON, &doc.Metadata); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		docs = append(docs, doc)
	}

	return docs, total, nil
}

// GetByAuthorID retrieves documents by author
func (r *DocumentRepositoryPG) GetByAuthorID(ctx context.Context, authorID int64, limit, offset int) ([]*entities.Document, error) {
	docs, _, err := r.List(ctx, repositories.DocumentFilter{
		AuthorID: &authorID,
		Limit:    limit,
		Offset:   offset,
	})
	return docs, err
}

// GetByStatus retrieves documents by status
func (r *DocumentRepositoryPG) GetByStatus(ctx context.Context, status entities.DocumentStatus, limit, offset int) ([]*entities.Document, error) {
	docs, _, err := r.List(ctx, repositories.DocumentFilter{
		Status: &status,
		Limit:  limit,
		Offset: offset,
	})
	return docs, err
}

// CreateVersion creates a new document version with full snapshot
func (r *DocumentRepositoryPG) CreateVersion(ctx context.Context, version *entities.DocumentVersion) error {
	var metadataJSON interface{} = nil
	var err error
	if version.Metadata != nil {
		metadataJSON, err = json.Marshal(version.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	query := `
		INSERT INTO document_versions (
			document_id, version, title, subject, content, status,
			file_name, file_path, file_size, mime_type, storage_key,
			metadata, changed_by, change_description, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id`

	return r.db.QueryRowContext(ctx, query,
		version.DocumentID, version.Version, version.Title, version.Subject, version.Content, version.Status,
		version.FileName, version.FilePath, version.FileSize, version.MimeType, version.StorageKey,
		metadataJSON, version.ChangedBy, version.ChangeDescription, time.Now(),
	).Scan(&version.ID)
}

// GetVersions retrieves all versions of a document with user names
func (r *DocumentRepositoryPG) GetVersions(ctx context.Context, documentID int64) ([]*entities.DocumentVersion, error) {
	query := `
		SELECT dv.id, dv.document_id, dv.version, dv.title, dv.subject, dv.content, dv.status,
			dv.file_name, dv.file_path, dv.file_size, dv.mime_type, dv.storage_key,
			dv.metadata, dv.changed_by, dv.change_description, dv.created_at,
			u.name as changed_by_name
		FROM document_versions dv
		LEFT JOIN users u ON dv.changed_by = u.id
		WHERE dv.document_id = $1
		ORDER BY dv.version DESC`

	rows, err := r.db.QueryContext(ctx, query, documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get versions: %w", err)
	}
	defer rows.Close()

	var versions []*entities.DocumentVersion
	for rows.Next() {
		v := &entities.DocumentVersion{}
		var metadataJSON []byte
		if err := rows.Scan(
			&v.ID, &v.DocumentID, &v.Version, &v.Title, &v.Subject, &v.Content, &v.Status,
			&v.FileName, &v.FilePath, &v.FileSize, &v.MimeType, &v.StorageKey,
			&metadataJSON, &v.ChangedBy, &v.ChangeDescription, &v.CreatedAt,
			&v.ChangedByName,
		); err != nil {
			return nil, fmt.Errorf("failed to scan version: %w", err)
		}
		if metadataJSON != nil {
			if err := json.Unmarshal(metadataJSON, &v.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}
		versions = append(versions, v)
	}
	return versions, nil
}

// GetVersion retrieves a specific version of a document
func (r *DocumentRepositoryPG) GetVersion(ctx context.Context, documentID int64, version int) (*entities.DocumentVersion, error) {
	query := `
		SELECT dv.id, dv.document_id, dv.version, dv.title, dv.subject, dv.content, dv.status,
			dv.file_name, dv.file_path, dv.file_size, dv.mime_type, dv.storage_key,
			dv.metadata, dv.changed_by, dv.change_description, dv.created_at,
			u.name as changed_by_name
		FROM document_versions dv
		LEFT JOIN users u ON dv.changed_by = u.id
		WHERE dv.document_id = $1 AND dv.version = $2`

	v := &entities.DocumentVersion{}
	var metadataJSON []byte
	err := r.db.QueryRowContext(ctx, query, documentID, version).Scan(
		&v.ID, &v.DocumentID, &v.Version, &v.Title, &v.Subject, &v.Content, &v.Status,
		&v.FileName, &v.FilePath, &v.FileSize, &v.MimeType, &v.StorageKey,
		&metadataJSON, &v.ChangedBy, &v.ChangeDescription, &v.CreatedAt,
		&v.ChangedByName,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("version not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get version: %w", err)
	}
	if metadataJSON != nil {
		if err := json.Unmarshal(metadataJSON, &v.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}
	return v, nil
}

// GetLatestVersion retrieves the most recent version of a document
func (r *DocumentRepositoryPG) GetLatestVersion(ctx context.Context, documentID int64) (*entities.DocumentVersion, error) {
	query := `
		SELECT dv.id, dv.document_id, dv.version, dv.title, dv.subject, dv.content, dv.status,
			dv.file_name, dv.file_path, dv.file_size, dv.mime_type, dv.storage_key,
			dv.metadata, dv.changed_by, dv.change_description, dv.created_at,
			u.name as changed_by_name
		FROM document_versions dv
		LEFT JOIN users u ON dv.changed_by = u.id
		WHERE dv.document_id = $1
		ORDER BY dv.version DESC
		LIMIT 1`

	v := &entities.DocumentVersion{}
	var metadataJSON []byte
	err := r.db.QueryRowContext(ctx, query, documentID).Scan(
		&v.ID, &v.DocumentID, &v.Version, &v.Title, &v.Subject, &v.Content, &v.Status,
		&v.FileName, &v.FilePath, &v.FileSize, &v.MimeType, &v.StorageKey,
		&metadataJSON, &v.ChangedBy, &v.ChangeDescription, &v.CreatedAt,
		&v.ChangedByName,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no versions found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest version: %w", err)
	}
	if metadataJSON != nil {
		if err := json.Unmarshal(metadataJSON, &v.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}
	return v, nil
}

// RestoreVersion restores a document to a specific version
func (r *DocumentRepositoryPG) RestoreVersion(ctx context.Context, documentID int64, version int, userID int64) error {
	// Get the version to restore
	v, err := r.GetVersion(ctx, documentID, version)
	if err != nil {
		return fmt.Errorf("failed to get version to restore: %w", err)
	}

	// Get current document to create a backup version
	doc, err := r.GetByID(ctx, documentID)
	if err != nil {
		return fmt.Errorf("failed to get current document: %w", err)
	}

	// Create a backup version of current state
	backupDescription := fmt.Sprintf("Резервная копия перед восстановлением версии %d", version)
	backupVersion := entities.NewDocumentVersion(doc, userID, backupDescription)
	if err := r.CreateVersion(ctx, backupVersion); err != nil {
		return fmt.Errorf("failed to create backup version: %w", err)
	}

	// Update document with version data
	var metadataJSON interface{} = nil
	if v.Metadata != nil {
		metadataJSON, err = json.Marshal(v.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	query := `
		UPDATE documents SET
			title = COALESCE($1, title),
			subject = $2,
			content = $3,
			file_name = $4,
			file_path = $5,
			file_size = $6,
			mime_type = $7,
			metadata = $8,
			version = version + 1,
			updated_at = NOW()
		WHERE id = $9 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query,
		v.Title, v.Subject, v.Content, v.FileName, v.FilePath, v.FileSize, v.MimeType,
		metadataJSON, documentID,
	)
	if err != nil {
		return fmt.Errorf("failed to restore version: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("document not found")
	}

	return nil
}

// DeleteVersion deletes a specific version (cannot delete current version)
func (r *DocumentRepositoryPG) DeleteVersion(ctx context.Context, documentID int64, version int) error {
	// Get current document version
	doc, err := r.GetByID(ctx, documentID)
	if err != nil {
		return fmt.Errorf("failed to get document: %w", err)
	}

	// Cannot delete current version
	if doc.Version == version {
		return fmt.Errorf("cannot delete current version")
	}

	query := `DELETE FROM document_versions WHERE document_id = $1 AND version = $2`
	result, err := r.db.ExecContext(ctx, query, documentID, version)
	if err != nil {
		return fmt.Errorf("failed to delete version: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("version not found")
	}

	return nil
}

// CreateVersionDiff creates a cached diff between two versions
func (r *DocumentRepositoryPG) CreateVersionDiff(ctx context.Context, diff *entities.DocumentVersionDiff) error {
	var diffDataJSON interface{} = nil
	var err error
	if diff.DiffData != nil {
		diffDataJSON, err = json.Marshal(diff.DiffData)
		if err != nil {
			return fmt.Errorf("failed to marshal diff data: %w", err)
		}
	}

	query := `
		INSERT INTO document_version_diffs (document_id, from_version, to_version, changed_fields, diff_data, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (document_id, from_version, to_version) DO UPDATE SET
			changed_fields = EXCLUDED.changed_fields,
			diff_data = EXCLUDED.diff_data,
			created_at = NOW()
		RETURNING id`

	return r.db.QueryRowContext(ctx, query,
		diff.DocumentID, diff.FromVersion, diff.ToVersion, pq.Array(diff.ChangedFields), diffDataJSON,
	).Scan(&diff.ID)
}

// GetVersionDiff retrieves a cached diff between two versions
func (r *DocumentRepositoryPG) GetVersionDiff(ctx context.Context, documentID int64, fromVersion, toVersion int) (*entities.DocumentVersionDiff, error) {
	query := `
		SELECT id, document_id, from_version, to_version, changed_fields, diff_data, created_at
		FROM document_version_diffs
		WHERE document_id = $1 AND from_version = $2 AND to_version = $3`

	diff := &entities.DocumentVersionDiff{}
	var diffDataJSON []byte
	err := r.db.QueryRowContext(ctx, query, documentID, fromVersion, toVersion).Scan(
		&diff.ID, &diff.DocumentID, &diff.FromVersion, &diff.ToVersion,
		pq.Array(&diff.ChangedFields), &diffDataJSON, &diff.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil // Not found, will need to compute
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get version diff: %w", err)
	}
	if diffDataJSON != nil {
		if err := json.Unmarshal(diffDataJSON, &diff.DiffData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal diff data: %w", err)
		}
	}
	return diff, nil
}

// CompareVersions computes the difference between two versions
func (r *DocumentRepositoryPG) CompareVersions(ctx context.Context, documentID int64, fromVersion, toVersion int) (*entities.DocumentVersionDiff, error) {
	// Check if we have a cached diff
	cachedDiff, err := r.GetVersionDiff(ctx, documentID, fromVersion, toVersion)
	if err != nil {
		return nil, err
	}
	if cachedDiff != nil {
		return cachedDiff, nil
	}

	// Get both versions
	v1, err := r.GetVersion(ctx, documentID, fromVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to get from version: %w", err)
	}
	v2, err := r.GetVersion(ctx, documentID, toVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to get to version: %w", err)
	}

	// Compute differences
	changedFields := []string{}
	diffData := make(map[string]interface{})

	// Compare title
	if !strPtrEqual(v1.Title, v2.Title) {
		changedFields = append(changedFields, "title")
		diffData["title"] = map[string]interface{}{"from": ptrToStr(v1.Title), "to": ptrToStr(v2.Title)}
	}

	// Compare subject
	if !strPtrEqual(v1.Subject, v2.Subject) {
		changedFields = append(changedFields, "subject")
		diffData["subject"] = map[string]interface{}{"from": ptrToStr(v1.Subject), "to": ptrToStr(v2.Subject)}
	}

	// Compare content
	if !strPtrEqual(v1.Content, v2.Content) {
		changedFields = append(changedFields, "content")
		diffData["content"] = map[string]interface{}{"from": ptrToStr(v1.Content), "to": ptrToStr(v2.Content)}
	}

	// Compare status
	if !strPtrEqual(v1.Status, v2.Status) {
		changedFields = append(changedFields, "status")
		diffData["status"] = map[string]interface{}{"from": ptrToStr(v1.Status), "to": ptrToStr(v2.Status)}
	}

	// Compare file_name
	if !strPtrEqual(v1.FileName, v2.FileName) {
		changedFields = append(changedFields, "file_name")
		diffData["file_name"] = map[string]interface{}{"from": ptrToStr(v1.FileName), "to": ptrToStr(v2.FileName)}
	}

	// Compare file_path
	if !strPtrEqual(v1.FilePath, v2.FilePath) {
		changedFields = append(changedFields, "file_path")
		diffData["file_path"] = map[string]interface{}{"from": ptrToStr(v1.FilePath), "to": ptrToStr(v2.FilePath)}
	}

	// Compare file_size
	if !int64PtrEqual(v1.FileSize, v2.FileSize) {
		changedFields = append(changedFields, "file_size")
		diffData["file_size"] = map[string]interface{}{"from": ptrToInt64(v1.FileSize), "to": ptrToInt64(v2.FileSize)}
	}

	diff := &entities.DocumentVersionDiff{
		DocumentID:    documentID,
		FromVersion:   fromVersion,
		ToVersion:     toVersion,
		ChangedFields: changedFields,
		DiffData:      diffData,
		CreatedAt:     time.Now(),
	}

	// Cache the diff for future use
	if err := r.CreateVersionDiff(ctx, diff); err != nil {
		// Log but don't fail - caching is optional
		fmt.Printf("Warning: failed to cache version diff: %v\n", err)
	}

	return diff, nil
}

// Helper functions for comparison
func strPtrEqual(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func int64PtrEqual(a, b *int64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func ptrToStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func ptrToInt64(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

// AddHistory adds a history entry
func (r *DocumentRepositoryPG) AddHistory(ctx context.Context, history *entities.DocumentHistory) error {
	var detailsJSON []byte
	var err error
	if history.Details != nil {
		detailsJSON, err = json.Marshal(history.Details)
		if err != nil {
			return fmt.Errorf("failed to marshal details: %w", err)
		}
	}

	query := `
		INSERT INTO document_history (document_id, user_id, action, details, ip_address, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`

	return r.db.QueryRowContext(ctx, query,
		history.DocumentID, history.UserID, history.Action, detailsJSON,
		history.IPAddress, history.UserAgent, time.Now(),
	).Scan(&history.ID)
}

// GetHistory retrieves history for a document
func (r *DocumentRepositoryPG) GetHistory(ctx context.Context, documentID int64) ([]*entities.DocumentHistory, error) {
	query := `
		SELECT id, document_id, user_id, action, details, ip_address, user_agent, created_at
		FROM document_history WHERE document_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, documentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*entities.DocumentHistory
	for rows.Next() {
		h := &entities.DocumentHistory{}
		var detailsJSON []byte
		if err := rows.Scan(&h.ID, &h.DocumentID, &h.UserID, &h.Action, &detailsJSON, &h.IPAddress, &h.UserAgent, &h.CreatedAt); err != nil {
			return nil, err
		}
		if detailsJSON != nil {
			if err := json.Unmarshal(detailsJSON, &h.Details); err != nil {
				return nil, err
			}
		}
		history = append(history, h)
	}
	return history, nil
}

// sanitizeForTsquery sanitizes user input and converts it to a prefix-enabled tsquery format.
// For example: "со" -> "со:*", "собака документ" -> "собака:* & документ:*"
// This enables autocomplete-style prefix matching in PostgreSQL full-text search.
func sanitizeForTsquery(query string) string {
	// First, replace all special tsquery characters with spaces
	// This handles cases like "test&query" -> "test query"
	replacer := strings.NewReplacer(
		"&", " ",
		"|", " ",
		"!", " ",
		"(", " ",
		")", " ",
		":", " ",
		"'", " ",
		"*", " ",
		"\\", " ",
		"<", " ",
		">", " ",
	)
	sanitized := replacer.Replace(query)

	// Now split into words
	words := strings.Fields(sanitized)
	if len(words) == 0 {
		return ""
	}

	// Add prefix operator to each word
	var parts []string
	for _, word := range words {
		if word != "" {
			parts = append(parts, word+":*")
		}
	}

	if len(parts) == 0 {
		return ""
	}

	// Join with AND operator for multi-word search
	return strings.Join(parts, " & ")
}

// Search performs full-text search on documents with ranking and highlighting
func (r *DocumentRepositoryPG) Search(ctx context.Context, filter repositories.SearchFilter) ([]*repositories.SearchResult, int64, error) {
	if filter.Query == "" {
		return nil, 0, fmt.Errorf("search query cannot be empty")
	}

	// Convert user query to prefix-enabled tsquery format
	tsqueryStr := sanitizeForTsquery(filter.Query)
	if tsqueryStr == "" {
		return []*repositories.SearchResult{}, 0, nil
	}

	var conditions []string
	var args []interface{}
	argIndex := 1

	// Add the sanitized search query as first argument (for to_tsquery)
	args = append(args, tsqueryStr)
	// Use to_tsquery instead of plainto_tsquery to support prefix matching with :*
	searchCondition := fmt.Sprintf("d.search_vector @@ to_tsquery('russian', $%d)", argIndex)
	conditions = append(conditions, searchCondition)
	argIndex++

	// Add deleted filter
	if !filter.IncludeDeleted {
		conditions = append(conditions, "d.deleted_at IS NULL")
	}

	// Access control: user can see documents if:
	// 1. They are the author
	// 2. Document is public
	// 3. They have explicit permission via document_permissions
	// 4. They are admin (can see all)
	if filter.CurrentUserID > 0 && filter.CurrentUserRole != "admin" {
		accessCondition := fmt.Sprintf(`(
			d.author_id = $%d
			OR d.is_public = true
			OR EXISTS (
				SELECT 1 FROM document_permissions dp
				WHERE dp.document_id = d.id
				AND (dp.expires_at IS NULL OR dp.expires_at > NOW())
				AND (dp.user_id = $%d OR dp.role = $%d)
			)
		)`, argIndex, argIndex, argIndex+1)
		conditions = append(conditions, accessCondition)
		args = append(args, filter.CurrentUserID, filter.CurrentUserRole)
		argIndex += 2
	}

	// Add optional filters
	if filter.DocumentTypeID != nil {
		conditions = append(conditions, fmt.Sprintf("d.document_type_id = $%d", argIndex))
		args = append(args, *filter.DocumentTypeID)
		argIndex++
	}
	if filter.CategoryID != nil {
		conditions = append(conditions, fmt.Sprintf("d.category_id = $%d", argIndex))
		args = append(args, *filter.CategoryID)
		argIndex++
	}
	if filter.AuthorID != nil {
		conditions = append(conditions, fmt.Sprintf("d.author_id = $%d", argIndex))
		args = append(args, *filter.AuthorID)
		argIndex++
	}
	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("d.status = $%d", argIndex))
		args = append(args, *filter.Status)
		argIndex++
	}
	if filter.Importance != nil {
		conditions = append(conditions, fmt.Sprintf("d.importance = $%d", argIndex))
		args = append(args, *filter.Importance)
		argIndex++
	}
	if filter.FromDate != nil {
		conditions = append(conditions, fmt.Sprintf("d.created_at >= $%d", argIndex))
		args = append(args, *filter.FromDate)
		argIndex++
	}
	if filter.ToDate != nil {
		conditions = append(conditions, fmt.Sprintf("d.created_at <= $%d", argIndex))
		args = append(args, *filter.ToDate)
		argIndex++
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	// Count total results
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM documents d %s", whereClause) // #nosec G201 -- dynamic WHERE from parameterized conditions, not user input
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count search results: %w", err)
	}

	if total == 0 {
		return []*repositories.SearchResult{}, 0, nil
	}

	// Get search results with ranking and highlighting
	// We need to pass the query again for ts_headline functions
	// Using to_tsquery to support prefix matching with :*
	query := fmt.Sprintf(`
		SELECT
			d.id, d.document_type_id, d.category_id, d.registration_number, d.registration_date,
			d.title, d.subject, d.content, d.author_id, d.author_department, d.author_position,
			d.recipient_id, d.recipient_department, d.recipient_position, d.recipient_external,
			d.status, d.file_name, d.file_path, d.file_size, d.mime_type, d.version,
			d.parent_document_id, d.deadline, d.execution_date, d.metadata, d.is_public, d.importance,
			d.created_at, d.updated_at, d.deleted_at,
			author.name as author_name, recipient.name as recipient_name,
			ts_rank(d.search_vector, to_tsquery('russian', $1)) as rank,
			ts_headline('russian', coalesce(d.title, ''), to_tsquery('russian', $1),
				'StartSel=<mark>, StopSel=</mark>, MaxWords=50, MinWords=10, MaxFragments=1') as highlighted_title,
			ts_headline('russian', coalesce(d.subject, ''), to_tsquery('russian', $1),
				'StartSel=<mark>, StopSel=</mark>, MaxWords=50, MinWords=10, MaxFragments=1') as highlighted_subject,
			ts_headline('russian', coalesce(d.content, ''), to_tsquery('russian', $1),
				'StartSel=<mark>, StopSel=</mark>, MaxWords=100, MinWords=25, MaxFragments=3') as highlighted_content
		FROM documents d
		LEFT JOIN users author ON d.author_id = author.id
		LEFT JOIN users recipient ON d.recipient_id = recipient.id
		%s
		ORDER BY rank DESC, d.created_at DESC
		LIMIT $%d OFFSET $%d`,
		whereClause, argIndex, argIndex+1) // #nosec G201 -- dynamic WHERE from parameterized conditions, not user input

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search documents: %w", err)
	}
	defer rows.Close()

	var results []*repositories.SearchResult
	for rows.Next() {
		doc := &entities.Document{}
		result := &repositories.SearchResult{Document: doc}
		var metadataJSON []byte

		err := rows.Scan(
			&doc.ID, &doc.DocumentTypeID, &doc.CategoryID, &doc.RegistrationNumber, &doc.RegistrationDate,
			&doc.Title, &doc.Subject, &doc.Content, &doc.AuthorID, &doc.AuthorDepartment, &doc.AuthorPosition,
			&doc.RecipientID, &doc.RecipientDepartment, &doc.RecipientPosition, &doc.RecipientExternal,
			&doc.Status, &doc.FileName, &doc.FilePath, &doc.FileSize, &doc.MimeType, &doc.Version,
			&doc.ParentDocumentID, &doc.Deadline, &doc.ExecutionDate, &metadataJSON, &doc.IsPublic, &doc.Importance,
			&doc.CreatedAt, &doc.UpdatedAt, &doc.DeletedAt,
			&doc.AuthorName, &doc.RecipientName,
			&result.Rank,
			&result.HighlightedTitle, &result.HighlightedSubject, &result.HighlightedContent,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan search result: %w", err)
		}

		if metadataJSON != nil {
			if err := json.Unmarshal(metadataJSON, &doc.Metadata); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		results = append(results, result)
	}

	return results, total, nil
}
