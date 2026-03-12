// Package persistence provides database implementations for document repositories.
package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
)

// DocumentTagRepositoryPG implements DocumentTagRepository using PostgreSQL
type DocumentTagRepositoryPG struct {
	db *sql.DB
}

// NewDocumentTagRepositoryPG creates a new PostgreSQL document tag repository
func NewDocumentTagRepositoryPG(db *sql.DB) *DocumentTagRepositoryPG {
	return &DocumentTagRepositoryPG{db: db}
}

// Create creates a new document tag
func (r *DocumentTagRepositoryPG) Create(ctx context.Context, tag *entities.DocumentTag) error {
	query := `
		INSERT INTO document_tags (name, color, created_at)
		VALUES ($1, $2, NOW())
		RETURNING id, created_at`

	err := r.db.QueryRowContext(ctx, query, tag.Name, tag.Color).
		Scan(&tag.ID, &tag.CreatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return fmt.Errorf("тег с таким именем уже существует")
		}
		return fmt.Errorf("failed to create tag: %w", err)
	}
	return nil
}

// Update updates an existing document tag
func (r *DocumentTagRepositoryPG) Update(ctx context.Context, tag *entities.DocumentTag) error {
	query := `
		UPDATE document_tags
		SET name = $1, color = $2
		WHERE id = $3`

	result, err := r.db.ExecContext(ctx, query, tag.Name, tag.Color, tag.ID)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return fmt.Errorf("тег с таким именем уже существует")
		}
		return fmt.Errorf("failed to update tag: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("тег не найден")
	}
	return nil
}

// Delete deletes a document tag
func (r *DocumentTagRepositoryPG) Delete(ctx context.Context, id int64) error {
	// Relations will be deleted by CASCADE
	result, err := r.db.ExecContext(ctx, `DELETE FROM document_tags WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("тег не найден")
	}
	return nil
}

// GetByID retrieves a document tag by ID
func (r *DocumentTagRepositoryPG) GetByID(ctx context.Context, id int64) (*entities.DocumentTag, error) {
	query := `SELECT id, name, color, created_at FROM document_tags WHERE id = $1`

	tag := &entities.DocumentTag{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(&tag.ID, &tag.Name, &tag.Color, &tag.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("тег не найден")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}
	return tag, nil
}

// GetByName retrieves a document tag by name
func (r *DocumentTagRepositoryPG) GetByName(ctx context.Context, name string) (*entities.DocumentTag, error) {
	query := `SELECT id, name, color, created_at FROM document_tags WHERE name = $1`

	tag := &entities.DocumentTag{}
	err := r.db.QueryRowContext(ctx, query, name).Scan(&tag.ID, &tag.Name, &tag.Color, &tag.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("тег не найден")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}
	return tag, nil
}

// GetAll retrieves all document tags
func (r *DocumentTagRepositoryPG) GetAll(ctx context.Context) ([]*entities.DocumentTag, error) {
	query := `SELECT id, name, color, created_at FROM document_tags ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var tags []*entities.DocumentTag
	for rows.Next() {
		tag := &entities.DocumentTag{}
		err := rows.Scan(&tag.ID, &tag.Name, &tag.Color, &tag.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

// Search searches for tags by name prefix
func (r *DocumentTagRepositoryPG) Search(ctx context.Context, query string, limit int) ([]*entities.DocumentTag, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	sqlQuery := `
		SELECT id, name, color, created_at
		FROM document_tags
		WHERE name ILIKE $1
		ORDER BY name
		LIMIT $2`

	rows, err := r.db.QueryContext(ctx, sqlQuery, query+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search tags: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var tags []*entities.DocumentTag
	for rows.Next() {
		tag := &entities.DocumentTag{}
		err := rows.Scan(&tag.ID, &tag.Name, &tag.Color, &tag.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

// AddTagToDocument adds a tag to a document
func (r *DocumentTagRepositoryPG) AddTagToDocument(ctx context.Context, documentID, tagID int64) error {
	query := `
		INSERT INTO document_tag_relations (document_id, tag_id, created_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (document_id, tag_id) DO NOTHING`

	_, err := r.db.ExecContext(ctx, query, documentID, tagID)
	if err != nil {
		return fmt.Errorf("failed to add tag to document: %w", err)
	}
	return nil
}

// RemoveTagFromDocument removes a tag from a document
func (r *DocumentTagRepositoryPG) RemoveTagFromDocument(ctx context.Context, documentID, tagID int64) error {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM document_tag_relations WHERE document_id = $1 AND tag_id = $2`,
		documentID, tagID)
	if err != nil {
		return fmt.Errorf("failed to remove tag from document: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("связь тега с документом не найдена")
	}
	return nil
}

// GetTagsByDocumentID retrieves all tags for a document
func (r *DocumentTagRepositoryPG) GetTagsByDocumentID(ctx context.Context, documentID int64) ([]*entities.DocumentTag, error) {
	query := `
		SELECT t.id, t.name, t.color, t.created_at
		FROM document_tags t
		INNER JOIN document_tag_relations r ON t.id = r.tag_id
		WHERE r.document_id = $1
		ORDER BY t.name`

	rows, err := r.db.QueryContext(ctx, query, documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get document tags: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var tags []*entities.DocumentTag
	for rows.Next() {
		tag := &entities.DocumentTag{}
		err := rows.Scan(&tag.ID, &tag.Name, &tag.Color, &tag.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

// GetDocumentsByTagID retrieves document IDs that have a specific tag
func (r *DocumentTagRepositoryPG) GetDocumentsByTagID(ctx context.Context, tagID int64, limit, offset int) ([]int64, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// Get total count
	var total int64
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM document_tag_relations WHERE tag_id = $1`, tagID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count documents: %w", err)
	}

	// Get document IDs
	query := `
		SELECT r.document_id
		FROM document_tag_relations r
		INNER JOIN documents d ON r.document_id = d.id
		WHERE r.tag_id = $1 AND d.deleted_at IS NULL
		ORDER BY r.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, tagID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get documents by tag: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var documentIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, 0, fmt.Errorf("failed to scan document ID: %w", err)
		}
		documentIDs = append(documentIDs, id)
	}
	return documentIDs, total, nil
}

// SetDocumentTags replaces all tags for a document
func (r *DocumentTagRepositoryPG) SetDocumentTags(ctx context.Context, documentID int64, tagIDs []int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Remove all existing tags
	_, err = tx.ExecContext(ctx, `DELETE FROM document_tag_relations WHERE document_id = $1`, documentID)
	if err != nil {
		return fmt.Errorf("failed to remove existing tags: %w", err)
	}

	// Add new tags
	if len(tagIDs) > 0 {
		stmt, err := tx.PrepareContext(ctx,
			`INSERT INTO document_tag_relations (document_id, tag_id, created_at) VALUES ($1, $2, NOW())`)
		if err != nil {
			return fmt.Errorf("failed to prepare statement: %w", err)
		}
		defer func() { _ = stmt.Close() }()

		for _, tagID := range tagIDs {
			_, err = stmt.ExecContext(ctx, documentID, tagID)
			if err != nil {
				return fmt.Errorf("failed to add tag %d: %w", tagID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// GetTagUsageCount returns the number of documents using a tag
func (r *DocumentTagRepositoryPG) GetTagUsageCount(ctx context.Context, tagID int64) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM document_tag_relations r
		 INNER JOIN documents d ON r.document_id = d.id
		 WHERE r.tag_id = $1 AND d.deleted_at IS NULL`, tagID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count tag usage: %w", err)
	}
	return count, nil
}
