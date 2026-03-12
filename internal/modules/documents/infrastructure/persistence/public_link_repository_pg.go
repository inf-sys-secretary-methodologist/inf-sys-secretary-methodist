package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
	domainErrors "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/domain/errors"
)

// PublicLinkRepositoryPG is a PostgreSQL implementation of PublicLinkRepository
type PublicLinkRepositoryPG struct {
	db *sql.DB
}

// NewPublicLinkRepositoryPG creates a new PublicLinkRepositoryPG
func NewPublicLinkRepositoryPG(db *sql.DB) *PublicLinkRepositoryPG {
	return &PublicLinkRepositoryPG{db: db}
}

// Create creates a new public link
func (r *PublicLinkRepositoryPG) Create(ctx context.Context, link *entities.PublicLink) error {
	query := `
		INSERT INTO document_public_links (
			document_id, token, permission, created_by, expires_at,
			max_uses, use_count, password_hash, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id`

	now := time.Now()
	link.CreatedAt = now
	link.UpdatedAt = now

	err := r.db.QueryRowContext(ctx, query,
		link.DocumentID,
		link.Token,
		link.Permission,
		link.CreatedBy,
		link.ExpiresAt,
		link.MaxUses,
		link.UseCount,
		link.PasswordHash,
		link.IsActive,
		link.CreatedAt,
		link.UpdatedAt,
	).Scan(&link.ID)

	if err != nil {
		return fmt.Errorf("failed to create public link: %w", err)
	}
	return nil
}

// Update updates an existing public link
func (r *PublicLinkRepositoryPG) Update(ctx context.Context, link *entities.PublicLink) error {
	query := `
		UPDATE document_public_links
		SET permission = $1, expires_at = $2, max_uses = $3, password_hash = $4,
		    is_active = $5, updated_at = $6
		WHERE id = $7`

	link.UpdatedAt = time.Now()
	result, err := r.db.ExecContext(ctx, query,
		link.Permission,
		link.ExpiresAt,
		link.MaxUses,
		link.PasswordHash,
		link.IsActive,
		link.UpdatedAt,
		link.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update public link: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domainErrors.ErrNotFound
	}
	return nil
}

// Delete deletes a public link by ID
func (r *PublicLinkRepositoryPG) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM document_public_links WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete public link: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domainErrors.ErrNotFound
	}
	return nil
}

// GetByID retrieves a public link by ID
func (r *PublicLinkRepositoryPG) GetByID(ctx context.Context, id int64) (*entities.PublicLink, error) {
	query := `
		SELECT pl.id, pl.document_id, pl.token, pl.permission, pl.created_by,
		       pl.expires_at, pl.max_uses, pl.use_count, pl.password_hash,
		       pl.is_active, pl.created_at, pl.updated_at,
		       d.title as document_title, u.name as created_by_name
		FROM document_public_links pl
		JOIN documents d ON pl.document_id = d.id
		JOIN users u ON pl.created_by = u.id
		WHERE pl.id = $1`

	var link entities.PublicLink
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&link.ID, &link.DocumentID, &link.Token, &link.Permission, &link.CreatedBy,
		&link.ExpiresAt, &link.MaxUses, &link.UseCount, &link.PasswordHash,
		&link.IsActive, &link.CreatedAt, &link.UpdatedAt,
		&link.DocumentTitle, &link.CreatedByName,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainErrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get public link: %w", err)
	}
	return &link, nil
}

// GetByToken retrieves a public link by token
func (r *PublicLinkRepositoryPG) GetByToken(ctx context.Context, token string) (*entities.PublicLink, error) {
	query := `
		SELECT pl.id, pl.document_id, pl.token, pl.permission, pl.created_by,
		       pl.expires_at, pl.max_uses, pl.use_count, pl.password_hash,
		       pl.is_active, pl.created_at, pl.updated_at,
		       d.title as document_title, u.name as created_by_name
		FROM document_public_links pl
		JOIN documents d ON pl.document_id = d.id
		JOIN users u ON pl.created_by = u.id
		WHERE pl.token = $1`

	var link entities.PublicLink
	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&link.ID, &link.DocumentID, &link.Token, &link.Permission, &link.CreatedBy,
		&link.ExpiresAt, &link.MaxUses, &link.UseCount, &link.PasswordHash,
		&link.IsActive, &link.CreatedAt, &link.UpdatedAt,
		&link.DocumentTitle, &link.CreatedByName,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainErrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get public link: %w", err)
	}
	return &link, nil
}

// GetByDocumentID retrieves all public links for a document
func (r *PublicLinkRepositoryPG) GetByDocumentID(ctx context.Context, documentID int64) ([]*entities.PublicLink, error) {
	query := `
		SELECT pl.id, pl.document_id, pl.token, pl.permission, pl.created_by,
		       pl.expires_at, pl.max_uses, pl.use_count, pl.password_hash,
		       pl.is_active, pl.created_at, pl.updated_at,
		       d.title as document_title, u.name as created_by_name
		FROM document_public_links pl
		JOIN documents d ON pl.document_id = d.id
		JOIN users u ON pl.created_by = u.id
		WHERE pl.document_id = $1
		ORDER BY pl.created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query public links: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var links []*entities.PublicLink
	for rows.Next() {
		var link entities.PublicLink
		if err := rows.Scan(
			&link.ID, &link.DocumentID, &link.Token, &link.Permission, &link.CreatedBy,
			&link.ExpiresAt, &link.MaxUses, &link.UseCount, &link.PasswordHash,
			&link.IsActive, &link.CreatedAt, &link.UpdatedAt,
			&link.DocumentTitle, &link.CreatedByName,
		); err != nil {
			return nil, fmt.Errorf("failed to scan public link: %w", err)
		}
		links = append(links, &link)
	}
	return links, nil
}

// GetByCreatedBy retrieves all public links created by a user
func (r *PublicLinkRepositoryPG) GetByCreatedBy(ctx context.Context, userID int64) ([]*entities.PublicLink, error) {
	query := `
		SELECT pl.id, pl.document_id, pl.token, pl.permission, pl.created_by,
		       pl.expires_at, pl.max_uses, pl.use_count, pl.password_hash,
		       pl.is_active, pl.created_at, pl.updated_at,
		       d.title as document_title, u.name as created_by_name
		FROM document_public_links pl
		JOIN documents d ON pl.document_id = d.id
		JOIN users u ON pl.created_by = u.id
		WHERE pl.created_by = $1
		ORDER BY pl.created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query public links: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var links []*entities.PublicLink
	for rows.Next() {
		var link entities.PublicLink
		if err := rows.Scan(
			&link.ID, &link.DocumentID, &link.Token, &link.Permission, &link.CreatedBy,
			&link.ExpiresAt, &link.MaxUses, &link.UseCount, &link.PasswordHash,
			&link.IsActive, &link.CreatedAt, &link.UpdatedAt,
			&link.DocumentTitle, &link.CreatedByName,
		); err != nil {
			return nil, fmt.Errorf("failed to scan public link: %w", err)
		}
		links = append(links, &link)
	}
	return links, nil
}

// GetActiveByDocumentID retrieves all active public links for a document
func (r *PublicLinkRepositoryPG) GetActiveByDocumentID(ctx context.Context, documentID int64) ([]*entities.PublicLink, error) {
	query := `
		SELECT pl.id, pl.document_id, pl.token, pl.permission, pl.created_by,
		       pl.expires_at, pl.max_uses, pl.use_count, pl.password_hash,
		       pl.is_active, pl.created_at, pl.updated_at,
		       d.title as document_title, u.name as created_by_name
		FROM document_public_links pl
		JOIN documents d ON pl.document_id = d.id
		JOIN users u ON pl.created_by = u.id
		WHERE pl.document_id = $1
		AND pl.is_active = true
		AND (pl.expires_at IS NULL OR pl.expires_at > NOW())
		AND (pl.max_uses IS NULL OR pl.use_count < pl.max_uses)
		ORDER BY pl.created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query public links: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var links []*entities.PublicLink
	for rows.Next() {
		var link entities.PublicLink
		if err := rows.Scan(
			&link.ID, &link.DocumentID, &link.Token, &link.Permission, &link.CreatedBy,
			&link.ExpiresAt, &link.MaxUses, &link.UseCount, &link.PasswordHash,
			&link.IsActive, &link.CreatedAt, &link.UpdatedAt,
			&link.DocumentTitle, &link.CreatedByName,
		); err != nil {
			return nil, fmt.Errorf("failed to scan public link: %w", err)
		}
		links = append(links, &link)
	}
	return links, nil
}

// IncrementUseCount increments the use count of a public link
func (r *PublicLinkRepositoryPG) IncrementUseCount(ctx context.Context, id int64) error {
	query := `UPDATE document_public_links SET use_count = use_count + 1, updated_at = NOW() WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to increment use count: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domainErrors.ErrNotFound
	}
	return nil
}

// Deactivate deactivates a public link
func (r *PublicLinkRepositoryPG) Deactivate(ctx context.Context, id int64) error {
	query := `UPDATE document_public_links SET is_active = false, updated_at = NOW() WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to deactivate public link: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domainErrors.ErrNotFound
	}
	return nil
}

// Activate activates a public link
func (r *PublicLinkRepositoryPG) Activate(ctx context.Context, id int64) error {
	query := `UPDATE document_public_links SET is_active = true, updated_at = NOW() WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to activate public link: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domainErrors.ErrNotFound
	}
	return nil
}

// DeleteByDocumentID deletes all public links for a document
func (r *PublicLinkRepositoryPG) DeleteByDocumentID(ctx context.Context, documentID int64) error {
	query := `DELETE FROM document_public_links WHERE document_id = $1`
	_, err := r.db.ExecContext(ctx, query, documentID)
	if err != nil {
		return fmt.Errorf("failed to delete public links: %w", err)
	}
	return nil
}

// DeactivateExpired deactivates all expired public links and returns the count
func (r *PublicLinkRepositoryPG) DeactivateExpired(ctx context.Context) (int64, error) {
	query := `
		UPDATE document_public_links
		SET is_active = false, updated_at = NOW()
		WHERE is_active = true
		AND (
			(expires_at IS NOT NULL AND expires_at < NOW())
			OR (max_uses IS NOT NULL AND use_count >= max_uses)
		)`

	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to deactivate expired links: %w", err)
	}
	return result.RowsAffected()
}
