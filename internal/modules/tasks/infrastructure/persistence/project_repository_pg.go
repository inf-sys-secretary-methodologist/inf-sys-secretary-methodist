package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/repositories"
)

// ProjectRepositoryPG implements ProjectRepository using PostgreSQL.
type ProjectRepositoryPG struct {
	db *sql.DB
}

// NewProjectRepositoryPG creates a new ProjectRepositoryPG.
func NewProjectRepositoryPG(db *sql.DB) *ProjectRepositoryPG {
	return &ProjectRepositoryPG{db: db}
}

// Create creates a new project.
func (r *ProjectRepositoryPG) Create(ctx context.Context, project *entities.Project) error {
	query := `
		INSERT INTO projects (name, description, owner_id, status, start_date, end_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`

	return r.db.QueryRowContext(ctx, query,
		project.Name, project.Description, project.OwnerID, project.Status,
		project.StartDate, project.EndDate, project.CreatedAt, project.UpdatedAt,
	).Scan(&project.ID)
}

// Save updates an existing project.
func (r *ProjectRepositoryPG) Save(ctx context.Context, project *entities.Project) error {
	query := `
		UPDATE projects SET
			name = $1, description = $2, status = $3, start_date = $4,
			end_date = $5, updated_at = $6
		WHERE id = $7`

	_, err := r.db.ExecContext(ctx, query,
		project.Name, project.Description, project.Status,
		project.StartDate, project.EndDate, project.UpdatedAt, project.ID,
	)
	return err
}

// GetByID retrieves a project by ID.
func (r *ProjectRepositoryPG) GetByID(ctx context.Context, id int64) (*entities.Project, error) {
	query := `
		SELECT id, name, description, owner_id, status, start_date, end_date, created_at, updated_at
		FROM projects WHERE id = $1`

	project := &entities.Project{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&project.ID, &project.Name, &project.Description, &project.OwnerID,
		&project.Status, &project.StartDate, &project.EndDate,
		&project.CreatedAt, &project.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return project, nil
}

// Delete deletes a project.
func (r *ProjectRepositoryPG) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM projects WHERE id = $1", id)
	return err
}

// List lists projects with filters.
func (r *ProjectRepositoryPG) List(ctx context.Context, filter repositories.ProjectFilter, limit, offset int) ([]*entities.Project, error) {
	query, args := r.buildListQuery(filter, limit, offset, false)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return r.scanProjects(rows)
}

// Count counts projects with filters.
func (r *ProjectRepositoryPG) Count(ctx context.Context, filter repositories.ProjectFilter) (int64, error) {
	query, args := r.buildListQuery(filter, 0, 0, true)

	var count int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

func (r *ProjectRepositoryPG) buildListQuery(filter repositories.ProjectFilter, limit, offset int, countOnly bool) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argNum := 1

	if filter.OwnerID != nil {
		conditions = append(conditions, fmt.Sprintf("owner_id = $%d", argNum))
		args = append(args, *filter.OwnerID)
		argNum++
	}

	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argNum))
		args = append(args, *filter.Status)
		argNum++
	}

	if filter.Search != nil && *filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d)", argNum, argNum))
		args = append(args, "%"+*filter.Search+"%")
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	if countOnly {
		return "SELECT COUNT(*) FROM projects" + whereClause, args
	}

	query := `
		SELECT id, name, description, owner_id, status, start_date, end_date, created_at, updated_at
		FROM projects` + whereClause + ` ORDER BY created_at DESC`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)
	}

	return query, args
}

func (r *ProjectRepositoryPG) scanProjects(rows *sql.Rows) ([]*entities.Project, error) {
	var projects []*entities.Project

	for rows.Next() {
		project := &entities.Project{}
		err := rows.Scan(
			&project.ID, &project.Name, &project.Description, &project.OwnerID,
			&project.Status, &project.StartDate, &project.EndDate,
			&project.CreatedAt, &project.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}

	return projects, rows.Err()
}

// GetByOwner retrieves projects by owner ID.
func (r *ProjectRepositoryPG) GetByOwner(ctx context.Context, ownerID int64, limit, offset int) ([]*entities.Project, error) {
	filter := repositories.ProjectFilter{OwnerID: &ownerID}
	return r.List(ctx, filter, limit, offset)
}

// GetByStatus retrieves projects by status.
func (r *ProjectRepositoryPG) GetByStatus(ctx context.Context, status domain.ProjectStatus, limit, offset int) ([]*entities.Project, error) {
	filter := repositories.ProjectFilter{Status: &status}
	return r.List(ctx, filter, limit, offset)
}
