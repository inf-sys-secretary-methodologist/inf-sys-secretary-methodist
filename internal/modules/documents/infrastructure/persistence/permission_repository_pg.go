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

// PermissionRepositoryPG is a PostgreSQL implementation of PermissionRepository
type PermissionRepositoryPG struct {
	db *sql.DB
}

// NewPermissionRepositoryPG creates a new PermissionRepositoryPG
func NewPermissionRepositoryPG(db *sql.DB) *PermissionRepositoryPG {
	return &PermissionRepositoryPG{db: db}
}

// Create creates a new document permission
func (r *PermissionRepositoryPG) Create(ctx context.Context, permission *entities.DocumentPermission) error {
	query := `
		INSERT INTO document_permissions (
			document_id, user_id, role, permission, granted_by, expires_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`

	permission.CreatedAt = time.Now()
	err := r.db.QueryRowContext(ctx, query,
		permission.DocumentID,
		permission.UserID,
		permission.Role,
		permission.Permission,
		permission.GrantedBy,
		permission.ExpiresAt,
		permission.CreatedAt,
	).Scan(&permission.ID)

	if err != nil {
		return fmt.Errorf("failed to create permission: %w", err)
	}
	return nil
}

// Update updates an existing permission
func (r *PermissionRepositoryPG) Update(ctx context.Context, permission *entities.DocumentPermission) error {
	query := `
		UPDATE document_permissions
		SET permission = $1, expires_at = $2
		WHERE id = $3`

	result, err := r.db.ExecContext(ctx, query,
		permission.Permission,
		permission.ExpiresAt,
		permission.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update permission: %w", err)
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

// Delete deletes a permission by ID
func (r *PermissionRepositoryPG) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM document_permissions WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete permission: %w", err)
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

// GetByID retrieves a permission by ID
func (r *PermissionRepositoryPG) GetByID(ctx context.Context, id int64) (*entities.DocumentPermission, error) {
	query := `
		SELECT dp.id, dp.document_id, dp.user_id, dp.role, dp.permission,
		       dp.granted_by, dp.expires_at, dp.created_at,
		       u.name as user_name, u.email as user_email,
		       g.name as granted_by_name
		FROM document_permissions dp
		LEFT JOIN users u ON dp.user_id = u.id
		LEFT JOIN users g ON dp.granted_by = g.id
		WHERE dp.id = $1`

	var permission entities.DocumentPermission
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&permission.ID,
		&permission.DocumentID,
		&permission.UserID,
		&permission.Role,
		&permission.Permission,
		&permission.GrantedBy,
		&permission.ExpiresAt,
		&permission.CreatedAt,
		&permission.UserName,
		&permission.UserEmail,
		&permission.GrantedByName,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainErrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}
	return &permission, nil
}

// GetByDocumentID retrieves all permissions for a document
func (r *PermissionRepositoryPG) GetByDocumentID(ctx context.Context, documentID int64) ([]*entities.DocumentPermission, error) {
	query := `
		SELECT dp.id, dp.document_id, dp.user_id, dp.role, dp.permission,
		       dp.granted_by, dp.expires_at, dp.created_at,
		       u.name as user_name, u.email as user_email,
		       g.name as granted_by_name
		FROM document_permissions dp
		LEFT JOIN users u ON dp.user_id = u.id
		LEFT JOIN users g ON dp.granted_by = g.id
		WHERE dp.document_id = $1
		ORDER BY dp.created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query permissions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var permissions []*entities.DocumentPermission
	for rows.Next() {
		var p entities.DocumentPermission
		if err := rows.Scan(
			&p.ID, &p.DocumentID, &p.UserID, &p.Role, &p.Permission,
			&p.GrantedBy, &p.ExpiresAt, &p.CreatedAt,
			&p.UserName, &p.UserEmail, &p.GrantedByName,
		); err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		permissions = append(permissions, &p)
	}
	return permissions, nil
}

// GetByUserID retrieves all permissions for a user
func (r *PermissionRepositoryPG) GetByUserID(ctx context.Context, userID int64) ([]*entities.DocumentPermission, error) {
	query := `
		SELECT dp.id, dp.document_id, dp.user_id, dp.role, dp.permission,
		       dp.granted_by, dp.expires_at, dp.created_at,
		       u.name as user_name, u.email as user_email,
		       g.name as granted_by_name
		FROM document_permissions dp
		LEFT JOIN users u ON dp.user_id = u.id
		LEFT JOIN users g ON dp.granted_by = g.id
		WHERE dp.user_id = $1 AND (dp.expires_at IS NULL OR dp.expires_at > NOW())
		ORDER BY dp.created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query permissions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var permissions []*entities.DocumentPermission
	for rows.Next() {
		var p entities.DocumentPermission
		if err := rows.Scan(
			&p.ID, &p.DocumentID, &p.UserID, &p.Role, &p.Permission,
			&p.GrantedBy, &p.ExpiresAt, &p.CreatedAt,
			&p.UserName, &p.UserEmail, &p.GrantedByName,
		); err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		permissions = append(permissions, &p)
	}
	return permissions, nil
}

// GetByUserIDOrRole retrieves all permissions for a user either by direct user_id or by role
func (r *PermissionRepositoryPG) GetByUserIDOrRole(ctx context.Context, userID int64, role string) ([]*entities.DocumentPermission, error) {
	query := `
		SELECT dp.id, dp.document_id, dp.user_id, dp.role, dp.permission,
		       dp.granted_by, dp.expires_at, dp.created_at,
		       u.name as user_name, u.email as user_email,
		       g.name as granted_by_name
		FROM document_permissions dp
		LEFT JOIN users u ON dp.user_id = u.id
		LEFT JOIN users g ON dp.granted_by = g.id
		WHERE (dp.user_id = $1 OR dp.role = $2)
		AND (dp.expires_at IS NULL OR dp.expires_at > NOW())
		ORDER BY dp.created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID, role)
	if err != nil {
		return nil, fmt.Errorf("failed to query permissions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var permissions []*entities.DocumentPermission
	for rows.Next() {
		var p entities.DocumentPermission
		if err := rows.Scan(
			&p.ID, &p.DocumentID, &p.UserID, &p.Role, &p.Permission,
			&p.GrantedBy, &p.ExpiresAt, &p.CreatedAt,
			&p.UserName, &p.UserEmail, &p.GrantedByName,
		); err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		permissions = append(permissions, &p)
	}
	return permissions, nil
}

// GetByGrantedBy retrieves all permissions granted by a specific user
func (r *PermissionRepositoryPG) GetByGrantedBy(ctx context.Context, userID int64) ([]*entities.DocumentPermission, error) {
	query := `
		SELECT dp.id, dp.document_id, dp.user_id, dp.role, dp.permission,
		       dp.granted_by, dp.expires_at, dp.created_at,
		       u.name as user_name, u.email as user_email,
		       g.name as granted_by_name
		FROM document_permissions dp
		LEFT JOIN users u ON dp.user_id = u.id
		LEFT JOIN users g ON dp.granted_by = g.id
		WHERE dp.granted_by = $1 AND (dp.expires_at IS NULL OR dp.expires_at > NOW())
		ORDER BY dp.created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query permissions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var permissions []*entities.DocumentPermission
	for rows.Next() {
		var p entities.DocumentPermission
		if err := rows.Scan(
			&p.ID, &p.DocumentID, &p.UserID, &p.Role, &p.Permission,
			&p.GrantedBy, &p.ExpiresAt, &p.CreatedAt,
			&p.UserName, &p.UserEmail, &p.GrantedByName,
		); err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		permissions = append(permissions, &p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating permissions: %w", err)
	}
	return permissions, nil
}

// GetByDocumentAndUser retrieves permission for a specific document and user
func (r *PermissionRepositoryPG) GetByDocumentAndUser(ctx context.Context, documentID, userID int64) (*entities.DocumentPermission, error) {
	query := `
		SELECT id, document_id, user_id, role, permission, granted_by, expires_at, created_at
		FROM document_permissions
		WHERE document_id = $1 AND user_id = $2
		AND (expires_at IS NULL OR expires_at > NOW())`

	var p entities.DocumentPermission
	err := r.db.QueryRowContext(ctx, query, documentID, userID).Scan(
		&p.ID, &p.DocumentID, &p.UserID, &p.Role, &p.Permission,
		&p.GrantedBy, &p.ExpiresAt, &p.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainErrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}
	return &p, nil
}

// GetByDocumentAndRole retrieves permission for a specific document and role
func (r *PermissionRepositoryPG) GetByDocumentAndRole(ctx context.Context, documentID int64, role entities.UserRole) (*entities.DocumentPermission, error) {
	query := `
		SELECT id, document_id, user_id, role, permission, granted_by, expires_at, created_at
		FROM document_permissions
		WHERE document_id = $1 AND role = $2
		AND (expires_at IS NULL OR expires_at > NOW())`

	var p entities.DocumentPermission
	err := r.db.QueryRowContext(ctx, query, documentID, role).Scan(
		&p.ID, &p.DocumentID, &p.UserID, &p.Role, &p.Permission,
		&p.GrantedBy, &p.ExpiresAt, &p.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainErrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}
	return &p, nil
}

// HasPermission checks if a user has a specific permission level for a document
func (r *PermissionRepositoryPG) HasPermission(ctx context.Context, documentID, userID int64, permission entities.PermissionLevel) (bool, error) {
	// Check both direct user permission and role-based permission
	query := `
		SELECT EXISTS(
			SELECT 1 FROM document_permissions dp
			LEFT JOIN users u ON u.id = $2
			WHERE dp.document_id = $1
			AND (dp.expires_at IS NULL OR dp.expires_at > NOW())
			AND (
				dp.user_id = $2
				OR (dp.role IS NOT NULL AND dp.role = u.role)
			)
			AND (
				dp.permission = $3
				OR dp.permission = 'admin'
				OR ($3 = 'read' AND dp.permission IN ('write', 'delete'))
				OR ($3 = 'write' AND dp.permission = 'delete')
			)
		)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, documentID, userID, permission).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check permission: %w", err)
	}
	return exists, nil
}

// HasAnyPermission checks if a user has any permission for a document
func (r *PermissionRepositoryPG) HasAnyPermission(ctx context.Context, documentID, userID int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM document_permissions dp
			LEFT JOIN users u ON u.id = $2
			WHERE dp.document_id = $1
			AND (dp.expires_at IS NULL OR dp.expires_at > NOW())
			AND (dp.user_id = $2 OR (dp.role IS NOT NULL AND dp.role = u.role))
		)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, documentID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check permission: %w", err)
	}
	return exists, nil
}

// GetUserPermissionLevel returns the highest permission level for a user on a document
func (r *PermissionRepositoryPG) GetUserPermissionLevel(ctx context.Context, documentID, userID int64, userRole entities.UserRole) (*entities.PermissionLevel, error) {
	query := `
		SELECT dp.permission
		FROM document_permissions dp
		WHERE dp.document_id = $1
		AND (dp.expires_at IS NULL OR dp.expires_at > NOW())
		AND (dp.user_id = $2 OR dp.role = $3)
		ORDER BY
			CASE dp.permission
				WHEN 'admin' THEN 1
				WHEN 'delete' THEN 2
				WHEN 'write' THEN 3
				WHEN 'read' THEN 4
			END
		LIMIT 1`

	var permission entities.PermissionLevel
	err := r.db.QueryRowContext(ctx, query, documentID, userID, userRole).Scan(&permission)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get permission level: %w", err)
	}
	return &permission, nil
}

// DeleteByDocumentID deletes all permissions for a document
func (r *PermissionRepositoryPG) DeleteByDocumentID(ctx context.Context, documentID int64) error {
	query := `DELETE FROM document_permissions WHERE document_id = $1`
	_, err := r.db.ExecContext(ctx, query, documentID)
	if err != nil {
		return fmt.Errorf("failed to delete permissions: %w", err)
	}
	return nil
}

// DeleteByUserID deletes all permissions for a user
func (r *PermissionRepositoryPG) DeleteByUserID(ctx context.Context, userID int64) error {
	query := `DELETE FROM document_permissions WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete permissions: %w", err)
	}
	return nil
}

// DeleteExpired deletes all expired permissions and returns the count
func (r *PermissionRepositoryPG) DeleteExpired(ctx context.Context) (int64, error) {
	query := `DELETE FROM document_permissions WHERE expires_at IS NOT NULL AND expires_at < NOW()`
	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired permissions: %w", err)
	}
	return result.RowsAffected()
}
