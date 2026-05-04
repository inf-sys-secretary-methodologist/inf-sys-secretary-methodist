// Package persistence provides PostgreSQL implementations of the
// assignments module's repository ports.
package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/repositories"
)

// AssignmentRepositoryPG is the SQL implementation of AssignmentRepository.
type AssignmentRepositoryPG struct {
	db *sql.DB
}

// NewAssignmentRepositoryPG constructs the repository.
func NewAssignmentRepositoryPG(db *sql.DB) *AssignmentRepositoryPG {
	return &AssignmentRepositoryPG{db: db}
}

const assignmentSelectColumns = `id, title, description, teacher_id, group_name, subject, max_score, due_date, created_at, updated_at`

// GetByID returns the assignment with the given id.
func (r *AssignmentRepositoryPG) GetByID(ctx context.Context, id int64) (*entities.Assignment, error) {
	query := `SELECT ` + assignmentSelectColumns + ` FROM assignments WHERE id = $1`

	var (
		idv         int64
		title       string
		description sql.NullString
		teacherID   int64
		groupName   string
		subject     string
		maxScore    int
		dueDate     sql.NullTime
		createdAt   time.Time
		updatedAt   time.Time
	)
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&idv, &title, &description, &teacherID, &groupName, &subject,
		&maxScore, &dueDate, &createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repositories.ErrAssignmentNotFound
		}
		return nil, fmt.Errorf("assignments: get by id: %w", err)
	}

	var due *time.Time
	if dueDate.Valid {
		due = &dueDate.Time
	}
	return entities.ReconstituteAssignment(
		idv, title, description.String, teacherID, groupName, subject,
		maxScore, due, createdAt, updatedAt,
	), nil
}
