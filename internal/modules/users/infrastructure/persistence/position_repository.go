// Package persistence implements repository interfaces for the users module.
package persistence

import (
	"context"
	"database/sql"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/database"
)

// PositionRepositoryPG implements PostgreSQL position repository.
type PositionRepositoryPG struct {
	db *sql.DB
}

// NewPositionRepositoryPG creates a new PostgreSQL position repository.
func NewPositionRepositoryPG(db *sql.DB) repositories.PositionRepository {
	return &PositionRepositoryPG{db: db}
}

// Create creates a new position in the database.
func (r *PositionRepositoryPG) Create(ctx context.Context, position *entities.Position) error {
	query := `
		INSERT INTO positions (name, code, description, level, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	err := r.db.QueryRowContext(ctx, query,
		position.Name,
		position.Code,
		position.Description,
		position.Level,
		position.IsActive,
		position.CreatedAt,
		position.UpdatedAt,
	).Scan(&position.ID)

	if err != nil {
		return database.MapPostgresError(err)
	}
	return nil
}

// GetByID retrieves a position by ID.
func (r *PositionRepositoryPG) GetByID(ctx context.Context, id int64) (*entities.Position, error) {
	position := &entities.Position{}
	query := `
		SELECT id, name, code, description, level, is_active, created_at, updated_at
		FROM positions
		WHERE id = $1
	`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&position.ID,
		&position.Name,
		&position.Code,
		&position.Description,
		&position.Level,
		&position.IsActive,
		&position.CreatedAt,
		&position.UpdatedAt,
	)

	if err != nil {
		return nil, database.MapPostgresError(err)
	}
	return position, nil
}

// GetByCode retrieves a position by code.
func (r *PositionRepositoryPG) GetByCode(ctx context.Context, code string) (*entities.Position, error) {
	position := &entities.Position{}
	query := `
		SELECT id, name, code, description, level, is_active, created_at, updated_at
		FROM positions
		WHERE code = $1
	`
	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&position.ID,
		&position.Name,
		&position.Code,
		&position.Description,
		&position.Level,
		&position.IsActive,
		&position.CreatedAt,
		&position.UpdatedAt,
	)

	if err != nil {
		return nil, database.MapPostgresError(err)
	}
	return position, nil
}

// Update updates an existing position.
func (r *PositionRepositoryPG) Update(ctx context.Context, position *entities.Position) error {
	query := `
		UPDATE positions
		SET name = $1, code = $2, description = $3, level = $4, is_active = $5, updated_at = $6
		WHERE id = $7
	`
	result, err := r.db.ExecContext(ctx, query,
		position.Name,
		position.Code,
		position.Description,
		position.Level,
		position.IsActive,
		position.UpdatedAt,
		position.ID,
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

// Delete removes a position by ID.
func (r *PositionRepositoryPG) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM positions WHERE id = $1`, id)
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

// List retrieves a paginated list of positions.
func (r *PositionRepositoryPG) List(ctx context.Context, limit, offset int, activeOnly bool) ([]*entities.Position, error) {
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
		SELECT id, name, code, description, level, is_active, created_at, updated_at
		FROM positions
	`
	args := []interface{}{}

	if activeOnly {
		query += ` WHERE is_active = true`
	}

	query += ` ORDER BY level ASC, name ASC LIMIT $1 OFFSET $2`
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, database.MapPostgresError(err)
	}
	defer rows.Close()

	positions := []*entities.Position{}
	for rows.Next() {
		position := &entities.Position{}
		if err := rows.Scan(
			&position.ID,
			&position.Name,
			&position.Code,
			&position.Description,
			&position.Level,
			&position.IsActive,
			&position.CreatedAt,
			&position.UpdatedAt,
		); err != nil {
			return nil, database.MapPostgresError(err)
		}
		positions = append(positions, position)
	}

	if err := rows.Err(); err != nil {
		return nil, database.MapPostgresError(err)
	}

	return positions, nil
}

// Count returns total count of positions.
func (r *PositionRepositoryPG) Count(ctx context.Context, activeOnly bool) (int64, error) {
	query := `SELECT COUNT(*) FROM positions`
	if activeOnly {
		query += ` WHERE is_active = true`
	}

	var count int64
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, database.MapPostgresError(err)
	}
	return count, nil
}
