package persistence

import (
	"context"
	"database/sql"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/database"
)

// UserRepositoryPG implements PostgreSQL user repository
type UserRepositoryPG struct {
	db *sql.DB
}

// NewUserRepositoryPG creates a new PostgreSQL user repository
func NewUserRepositoryPG(db *sql.DB) repositories.UserRepository {
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

// Save updates an existing user
func (r *UserRepositoryPG) Save(ctx context.Context, user *entities.User) error {
	query := `
		UPDATE users
		SET email = $1, password = $2, name = $3, role = $4, status = $5, updated_at = $6
		WHERE id = $7
	`
	result, err := r.db.ExecContext(ctx, query,
		user.Email,
		user.Password,
		user.Name,
		user.Role,
		user.Status,
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
	user := &entities.User{}
	query := `
		SELECT id, email, password, name, role, status, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Name,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, database.MapPostgresError(err)
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepositoryPG) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	user := &entities.User{}
	query := `
		SELECT id, email, password, name, role, status, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Name,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, database.MapPostgresError(err)
	}

	return user, nil
}

// GetByEmailForAuth retrieves a user by email for authentication
// In PG implementation, this is the same as GetByEmail (no caching at this level)
func (r *UserRepositoryPG) GetByEmailForAuth(ctx context.Context, email string) (*entities.User, error) {
	return r.GetByEmail(ctx, email)
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
		SELECT id, email, password, name, role, status, created_at, updated_at
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
		if err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Password,
			&user.Name,
			&user.Role,
			&user.Status,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, database.MapPostgresError(err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, database.MapPostgresError(err)
	}

	return users, nil
}
