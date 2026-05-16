package persistence

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/database"
)

// UserRepositoryPG implements PostgreSQL user repository
type UserRepositoryPG struct {
	db *sql.DB
}

// NewUserRepositoryPG creates a new PostgreSQL user repository
func NewUserRepositoryPG(db *sql.DB) usecases.UserRepository {
	return &UserRepositoryPG{db: db}
}

// Create creates a new user in the database
func (r *UserRepositoryPG) Create(ctx context.Context, user *entities.User) error {
	query := `
		INSERT INTO users (email, password, name, role, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	err := r.db.QueryRowContext(ctx, query,
		user.Email,
		user.Password,
		user.Name,
		user.Role,
		user.Status,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID)

	if err != nil {
		return database.MapPostgresError(err)
	}
	return nil
}

// Save updates an existing user, including MFA enrollment fields.
func (r *UserRepositoryPG) Save(ctx context.Context, user *entities.User) error {
	query := `
		UPDATE users
		SET email = $1, password = $2, name = $3, role = $4, status = $5,
		    mfa_secret = $6, mfa_enabled = $7, updated_at = $8
		WHERE id = $9
	`
	var mfaSecret sql.NullString
	if user.MFASecret != nil {
		mfaSecret = sql.NullString{String: user.MFASecret.String(), Valid: true}
	}
	result, err := r.db.ExecContext(ctx, query,
		user.Email,
		user.Password,
		user.Name,
		user.Role,
		user.Status,
		mfaSecret,
		user.MFAEnabled,
		user.UpdatedAt,
		user.ID,
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

// GetByID retrieves a user by ID
func (r *UserRepositoryPG) GetByID(ctx context.Context, userID int64) (*entities.User, error) {
	return r.scanUserByQuery(ctx, `
		SELECT id, email, password, name, role, status,
		       mfa_secret, mfa_enabled, created_at, updated_at
		FROM users
		WHERE id = $1
	`, userID)
}

// GetByEmail retrieves a user by email
func (r *UserRepositoryPG) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	return r.scanUserByQuery(ctx, `
		SELECT id, email, password, name, role, status,
		       mfa_secret, mfa_enabled, created_at, updated_at
		FROM users
		WHERE email = $1
	`, email)
}

// GetByEmailForAuth retrieves a user by email for authentication
// In PG implementation, this is the same as GetByEmail (no caching at this level)
func (r *UserRepositoryPG) GetByEmailForAuth(ctx context.Context, email string) (*entities.User, error) {
	return r.GetByEmail(ctx, email)
}

// GetByIDForAuth retrieves a user by ID bypassing cache; PG layer is already
// cache-free, so this delegates to GetByID.
func (r *UserRepositoryPG) GetByIDForAuth(ctx context.Context, id int64) (*entities.User, error) {
	return r.GetByID(ctx, id)
}

// scanUserByQuery executes a single-row SELECT and decodes the MFA secret
// into the typed VO; centralized so all read paths share the same parsing.
func (r *UserRepositoryPG) scanUserByQuery(ctx context.Context, query string, arg any) (*entities.User, error) {
	user := &entities.User{}
	var mfaSecret sql.NullString
	err := r.db.QueryRowContext(ctx, query, arg).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Name,
		&user.Role,
		&user.Status,
		&mfaSecret,
		&user.MFAEnabled,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, database.MapPostgresError(err)
	}
	if mfaSecret.Valid && mfaSecret.String != "" {
		secret, err := entities.NewMFASecret(mfaSecret.String)
		if err != nil {
			return nil, fmt.Errorf("repository: invalid persisted MFA secret for user %d: %w", user.ID, err)
		}
		user.MFASecret = &secret
	}
	return user, nil
}

// Delete removes a user by ID
func (r *UserRepositoryPG) Delete(ctx context.Context, userID int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, userID)
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

// List retrieves a paginated list of users
func (r *UserRepositoryPG) List(ctx context.Context, limit, offset int) ([]*entities.User, error) {
	// Validate pagination parameters
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	if limit > 100 {
		limit = 100 // Max limit
	}

	query := `
		SELECT id, email, password, name, role, status,
		       mfa_secret, mfa_enabled, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, database.MapPostgresError(err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			// Log error but don't fail the operation
			_ = err
		}
	}()

	users := []*entities.User{}
	for rows.Next() {
		user := &entities.User{}
		var mfaSecret sql.NullString
		if err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Password,
			&user.Name,
			&user.Role,
			&user.Status,
			&mfaSecret,
			&user.MFAEnabled,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, database.MapPostgresError(err)
		}
		if mfaSecret.Valid && mfaSecret.String != "" {
			secret, err := entities.NewMFASecret(mfaSecret.String)
			if err != nil {
				return nil, fmt.Errorf("repository: invalid persisted MFA secret for user %d: %w", user.ID, err)
			}
			user.MFASecret = &secret
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, database.MapPostgresError(err)
	}

	return users, nil
}
