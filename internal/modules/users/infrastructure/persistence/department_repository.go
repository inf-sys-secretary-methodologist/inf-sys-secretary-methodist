// Package persistence implements repository interfaces for the users module.
package persistence

import (
	"context"
	"database/sql"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/database"
)

// DepartmentRepositoryPG implements PostgreSQL department repository.
type DepartmentRepositoryPG struct {
	db *sql.DB
}

// NewDepartmentRepositoryPG creates a new PostgreSQL department repository.
func NewDepartmentRepositoryPG(db *sql.DB) repositories.DepartmentRepository {
	return &DepartmentRepositoryPG{db: db}
}

// Create creates a new department in the database.
func (r *DepartmentRepositoryPG) Create(ctx context.Context, department *entities.Department) error {
	query := `
		INSERT INTO org_departments (name, code, description, parent_id, head_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`
	err := r.db.QueryRowContext(ctx, query,
		department.Name,
		department.Code,
		department.Description,
		department.ParentID,
		department.HeadID,
		department.IsActive,
		department.CreatedAt,
		department.UpdatedAt,
	).Scan(&department.ID)

	if err != nil {
		return database.MapPostgresError(err)
	}
	return nil
}

// GetByID retrieves a department by ID.
func (r *DepartmentRepositoryPG) GetByID(ctx context.Context, id int64) (*entities.Department, error) {
	department := &entities.Department{}
	query := `
		SELECT id, name, code, description, parent_id, head_id, is_active, created_at, updated_at
		FROM org_departments
		WHERE id = $1
	`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&department.ID,
		&department.Name,
		&department.Code,
		&department.Description,
		&department.ParentID,
		&department.HeadID,
		&department.IsActive,
		&department.CreatedAt,
		&department.UpdatedAt,
	)

	if err != nil {
		return nil, database.MapPostgresError(err)
	}
	return department, nil
}

// GetByCode retrieves a department by code.
func (r *DepartmentRepositoryPG) GetByCode(ctx context.Context, code string) (*entities.Department, error) {
	department := &entities.Department{}
	query := `
		SELECT id, name, code, description, parent_id, head_id, is_active, created_at, updated_at
		FROM org_departments
		WHERE code = $1
	`
	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&department.ID,
		&department.Name,
		&department.Code,
		&department.Description,
		&department.ParentID,
		&department.HeadID,
		&department.IsActive,
		&department.CreatedAt,
		&department.UpdatedAt,
	)

	if err != nil {
		return nil, database.MapPostgresError(err)
	}
	return department, nil
}

// Update updates an existing department.
func (r *DepartmentRepositoryPG) Update(ctx context.Context, department *entities.Department) error {
	query := `
		UPDATE org_departments
		SET name = $1, code = $2, description = $3, parent_id = $4, head_id = $5, is_active = $6, updated_at = $7
		WHERE id = $8
	`
	result, err := r.db.ExecContext(ctx, query,
		department.Name,
		department.Code,
		department.Description,
		department.ParentID,
		department.HeadID,
		department.IsActive,
		department.UpdatedAt,
		department.ID,
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

// Delete removes a department by ID.
func (r *DepartmentRepositoryPG) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM org_departments WHERE id = $1`, id)
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

// List retrieves a paginated list of departments.
func (r *DepartmentRepositoryPG) List(ctx context.Context, limit, offset int, activeOnly bool) ([]*entities.Department, error) {
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
		SELECT id, name, code, description, parent_id, head_id, is_active, created_at, updated_at
		FROM org_departments
	`
	args := []interface{}{}

	if activeOnly {
		query += ` WHERE is_active = true`
	}

	query += ` ORDER BY name ASC LIMIT $1 OFFSET $2`
	if activeOnly {
		args = append(args, limit, offset)
	} else {
		args = append(args, limit, offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, database.MapPostgresError(err)
	}
	defer rows.Close()

	departments := []*entities.Department{}
	for rows.Next() {
		department := &entities.Department{}
		if err := rows.Scan(
			&department.ID,
			&department.Name,
			&department.Code,
			&department.Description,
			&department.ParentID,
			&department.HeadID,
			&department.IsActive,
			&department.CreatedAt,
			&department.UpdatedAt,
		); err != nil {
			return nil, database.MapPostgresError(err)
		}
		departments = append(departments, department)
	}

	if err := rows.Err(); err != nil {
		return nil, database.MapPostgresError(err)
	}

	return departments, nil
}

// Count returns total count of departments.
func (r *DepartmentRepositoryPG) Count(ctx context.Context, activeOnly bool) (int64, error) {
	query := `SELECT COUNT(*) FROM org_departments`
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

// GetChildren retrieves all child departments of a parent.
func (r *DepartmentRepositoryPG) GetChildren(ctx context.Context, parentID int64) ([]*entities.Department, error) {
	query := `
		SELECT id, name, code, description, parent_id, head_id, is_active, created_at, updated_at
		FROM org_departments
		WHERE parent_id = $1
		ORDER BY name ASC
	`
	rows, err := r.db.QueryContext(ctx, query, parentID)
	if err != nil {
		return nil, database.MapPostgresError(err)
	}
	defer rows.Close()

	departments := []*entities.Department{}
	for rows.Next() {
		department := &entities.Department{}
		if err := rows.Scan(
			&department.ID,
			&department.Name,
			&department.Code,
			&department.Description,
			&department.ParentID,
			&department.HeadID,
			&department.IsActive,
			&department.CreatedAt,
			&department.UpdatedAt,
		); err != nil {
			return nil, database.MapPostgresError(err)
		}
		departments = append(departments, department)
	}

	if err := rows.Err(); err != nil {
		return nil, database.MapPostgresError(err)
	}

	return departments, nil
}
