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

// List retrieves documents with filters
func (r *DocumentRepositoryPG) List(ctx context.Context, filter repositories.DocumentFilter) ([]*entities.Document, int64, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	if !filter.IncludeDeleted {
		conditions = append(conditions, "d.deleted_at IS NULL")
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

// CreateVersion creates a new document version
func (r *DocumentRepositoryPG) CreateVersion(ctx context.Context, version *entities.DocumentVersion) error {
	query := `
		INSERT INTO document_versions (document_id, version, content, file_name, file_path, file_size, changed_by, change_description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`

	return r.db.QueryRowContext(ctx, query,
		version.DocumentID, version.Version, version.Content, version.FileName,
		version.FilePath, version.FileSize, version.ChangedBy, version.ChangeDescription, time.Now(),
	).Scan(&version.ID)
}

// GetVersions retrieves all versions of a document
func (r *DocumentRepositoryPG) GetVersions(ctx context.Context, documentID int64) ([]*entities.DocumentVersion, error) {
	query := `
		SELECT id, document_id, version, content, file_name, file_path, file_size, changed_by, change_description, created_at
		FROM document_versions WHERE document_id = $1 ORDER BY version DESC`

	rows, err := r.db.QueryContext(ctx, query, documentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []*entities.DocumentVersion
	for rows.Next() {
		v := &entities.DocumentVersion{}
		if err := rows.Scan(&v.ID, &v.DocumentID, &v.Version, &v.Content, &v.FileName, &v.FilePath, &v.FileSize, &v.ChangedBy, &v.ChangeDescription, &v.CreatedAt); err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	return versions, nil
}

// GetVersion retrieves a specific version of a document
func (r *DocumentRepositoryPG) GetVersion(ctx context.Context, documentID int64, version int) (*entities.DocumentVersion, error) {
	query := `
		SELECT id, document_id, version, content, file_name, file_path, file_size, changed_by, change_description, created_at
		FROM document_versions WHERE document_id = $1 AND version = $2`

	v := &entities.DocumentVersion{}
	err := r.db.QueryRowContext(ctx, query, documentID, version).Scan(
		&v.ID, &v.DocumentID, &v.Version, &v.Content, &v.FileName, &v.FilePath, &v.FileSize, &v.ChangedBy, &v.ChangeDescription, &v.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("version not found")
	}
	return v, err
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
