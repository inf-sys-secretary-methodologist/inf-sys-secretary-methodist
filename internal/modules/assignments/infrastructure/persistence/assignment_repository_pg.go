// Package persistence provides PostgreSQL implementations of the
// assignments module's repository ports.
package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/entities"
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
			return nil, usecases.ErrAssignmentNotFound
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

// List returns a page of assignments matching the filter and the
// total of matching rows. The total is computed by a separate COUNT
// query rather than a window-function so an empty page past the end
// of the result set still reports the correct dataset size — the UI
// pagination needs that to render meaningful page controls.
//
// Filters use a sentinel-value pattern: NULL for TeacherID disables
// the teacher predicate, empty string for subject / group_name
// disables those. PostgreSQL accepts the cast-style "$1::bigint IS
// NULL" check uniformly with sql.NullInt64.
func (r *AssignmentRepositoryPG) List(ctx context.Context, filter usecases.AssignmentListFilter) (usecases.AssignmentListResult, error) {
	var tid sql.NullInt64
	if filter.TeacherID != nil {
		tid = sql.NullInt64{Int64: *filter.TeacherID, Valid: true}
	}

	const filterClause = `WHERE ($1::bigint IS NULL OR teacher_id = $1::bigint)
		AND ($2 = '' OR subject = $2)
		AND ($3 = '' OR group_name = $3)`

	countQuery := `SELECT COUNT(*) FROM assignments ` + filterClause
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, tid, filter.Subject, filter.GroupName).Scan(&total); err != nil {
		return usecases.AssignmentListResult{}, fmt.Errorf("assignments: count: %w", err)
	}

	listQuery := `SELECT ` + assignmentSelectColumns + ` FROM assignments ` + filterClause + `
		ORDER BY created_at DESC, id DESC
		LIMIT $4 OFFSET $5`

	rows, err := r.db.QueryContext(ctx, listQuery, tid, filter.Subject, filter.GroupName, filter.Limit, filter.Offset)
	if err != nil {
		return usecases.AssignmentListResult{}, fmt.Errorf("assignments: list: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var items []*entities.Assignment
	for rows.Next() {
		var (
			id          int64
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
		if err := rows.Scan(&id, &title, &description, &teacherID, &groupName, &subject,
			&maxScore, &dueDate, &createdAt, &updatedAt); err != nil {
			return usecases.AssignmentListResult{}, fmt.Errorf("assignments: list scan: %w", err)
		}
		var due *time.Time
		if dueDate.Valid {
			d := dueDate.Time
			due = &d
		}
		items = append(items, entities.ReconstituteAssignment(
			id, title, description.String, teacherID, groupName, subject,
			maxScore, due, createdAt, updatedAt,
		))
	}
	if err := rows.Err(); err != nil {
		return usecases.AssignmentListResult{}, fmt.Errorf("assignments: list iter: %w", err)
	}
	return usecases.AssignmentListResult{Items: items, Total: total}, nil
}

// AggregateGradeDistribution counts submissions grouped by
// (assignment.subject, submission.status) for submissions whose
// created_at lies in the half-open [from, to) range. Empty result
// is not an error.
func (r *AssignmentRepositoryPG) AggregateGradeDistribution(ctx context.Context, from, to time.Time) ([]usecases.AssignmentGradeDistributionAgg, error) {
	const query = `SELECT a.subject, s.status, COUNT(*) FROM submissions s
		JOIN assignments a ON a.id = s.assignment_id
		WHERE s.created_at >= $1 AND s.created_at < $2
		GROUP BY a.subject, s.status
		ORDER BY a.subject, s.status`

	rows, err := r.db.QueryContext(ctx, query, from, to)
	if err != nil {
		return nil, fmt.Errorf("assignments: aggregate grade distribution: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []usecases.AssignmentGradeDistributionAgg
	for rows.Next() {
		var (
			subject   string
			statusStr string
			count     int
		)
		if err := rows.Scan(&subject, &statusStr, &count); err != nil {
			return nil, fmt.Errorf("assignments: aggregate grade scan: %w", err)
		}
		out = append(out, usecases.AssignmentGradeDistributionAgg{
			Subject: subject,
			Status:  entities.SubmissionStatus(statusStr),
			Count:   count,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("assignments: aggregate grade rows: %w", err)
	}
	return out, nil
}
