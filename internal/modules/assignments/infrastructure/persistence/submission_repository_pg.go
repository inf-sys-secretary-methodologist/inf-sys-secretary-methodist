package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/views"
)

// SubmissionRepositoryPG is the SQL implementation of SubmissionRepository.
type SubmissionRepositoryPG struct {
	db *sql.DB
}

// NewSubmissionRepositoryPG constructs the repository.
func NewSubmissionRepositoryPG(db *sql.DB) *SubmissionRepositoryPG {
	return &SubmissionRepositoryPG{db: db}
}

const submissionSelectColumns = `id, assignment_id, student_id, grade_value, feedback, graded_by, graded_at, status, return_reason, returned_by, returned_at, created_at, updated_at`

// GetByAssignmentAndStudent returns the (assignment, student) submission row.
func (r *SubmissionRepositoryPG) GetByAssignmentAndStudent(ctx context.Context, assignmentID, studentID int64) (*entities.Submission, error) {
	query := `SELECT ` + submissionSelectColumns + ` FROM submissions WHERE assignment_id = $1 AND student_id = $2`

	var (
		id           int64
		aid          int64
		sid          int64
		gradeValue   sql.NullInt64
		feedback     sql.NullString
		gradedBy     sql.NullInt64
		gradedAt     sql.NullTime
		status       string
		returnReason sql.NullString
		returnedBy   sql.NullInt64
		returnedAt   sql.NullTime
		createdAt    time.Time
		updatedAt    time.Time
	)
	err := r.db.QueryRowContext(ctx, query, assignmentID, studentID).Scan(
		&id, &aid, &sid, &gradeValue, &feedback, &gradedBy, &gradedAt,
		&status, &returnReason, &returnedBy, &returnedAt,
		&createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repositories.ErrSubmissionNotFound
		}
		return nil, fmt.Errorf("submissions: get by pair: %w", err)
	}

	var gv *int
	if gradeValue.Valid {
		v := int(gradeValue.Int64)
		gv = &v
	}
	var gb *int64
	if gradedBy.Valid {
		gb = &gradedBy.Int64
	}
	var ga *time.Time
	if gradedAt.Valid {
		t := gradedAt.Time
		ga = &t
	}
	var rb *int64
	if returnedBy.Valid {
		rb = &returnedBy.Int64
	}
	var rat *time.Time
	if returnedAt.Valid {
		t := returnedAt.Time
		rat = &t
	}
	return entities.ReconstituteSubmission(
		id, aid, sid, gv, feedback.String, gb, ga,
		returnReason.String, rb, rat,
		entities.SubmissionStatus(status), createdAt, updatedAt,
	), nil
}

// Save upserts the submission keyed on (assignment_id, student_id).
// Insert when ID==0, update otherwise. The unique constraint
// uq_submissions_assignment_student guarantees that a concurrent
// "first grading" race materialises at most one row, and the upsert
// merges into it deterministically.
func (r *SubmissionRepositoryPG) Save(ctx context.Context, s *entities.Submission) error {
	query := `
		INSERT INTO submissions (
			assignment_id, student_id, grade_value, feedback, graded_by, graded_at, status,
			return_reason, returned_by, returned_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (assignment_id, student_id) DO UPDATE SET
			grade_value   = EXCLUDED.grade_value,
			feedback      = EXCLUDED.feedback,
			graded_by     = EXCLUDED.graded_by,
			graded_at     = EXCLUDED.graded_at,
			status        = EXCLUDED.status,
			return_reason = EXCLUDED.return_reason,
			returned_by   = EXCLUDED.returned_by,
			returned_at   = EXCLUDED.returned_at,
			updated_at    = EXCLUDED.updated_at
		RETURNING id`

	gradeValue := nullableInt(s.GradeValue())
	gradedBy := nullableInt64(s.GradedBy())
	gradedAt := nullableTime(s.GradedAt())
	returnReason := nullableString(s.ReturnReason())
	returnedBy := nullableInt64(s.ReturnedBy())
	returnedAt := nullableTime(s.ReturnedAt())

	var newID int64
	err := r.db.QueryRowContext(ctx, query,
		s.AssignmentID, s.StudentID, gradeValue, s.Feedback(),
		gradedBy, gradedAt, string(s.Status()),
		returnReason, returnedBy, returnedAt,
		s.CreatedAt(), s.UpdatedAt(),
	).Scan(&newID)
	if err != nil {
		return fmt.Errorf("submissions: upsert: %w", err)
	}
	s.ID = newID
	return nil
}

func nullableInt(p *int) sql.NullInt64 {
	if p == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(*p), Valid: true}
}

func nullableInt64(p *int64) sql.NullInt64 {
	if p == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *p, Valid: true}
}

func nullableTime(p *time.Time) sql.NullTime {
	if p == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *p, Valid: true}
}

// nullableString maps an empty Go string to a SQL NULL. Required for
// columns whose CHECK constraint forbids non-NULL values outside a
// specific status (e.g. submissions.return_reason: NULL when status is
// not 'returned').
func nullableString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

// ListByAssignment returns the read-side projection of submission rows
// for the given assignment, joined with users.name so the grading UI
// renders student names without a second round-trip. A nil status
// pointer means "any status"; the empty-string sentinel inside SQL
// disables the predicate.
func (r *SubmissionRepositoryPG) ListByAssignment(ctx context.Context, assignmentID int64, status *entities.SubmissionStatus) ([]views.SubmissionView, error) {
	statusFilter := ""
	if status != nil {
		statusFilter = string(*status)
	}

	query := `
		SELECT s.id, s.assignment_id, s.student_id, COALESCE(u.name, ''),
		       s.grade_value, s.feedback, s.graded_by, s.graded_at,
		       s.status, s.created_at, s.updated_at
		FROM submissions s
		JOIN users u ON u.id = s.student_id
		WHERE s.assignment_id = $1
		  AND ($2 = '' OR s.status = $2)
		ORDER BY u.name, s.id`

	rows, err := r.db.QueryContext(ctx, query, assignmentID, statusFilter)
	if err != nil {
		return nil, fmt.Errorf("submissions: list by assignment: %w", err)
	}
	defer rows.Close()

	var out []views.SubmissionView
	for rows.Next() {
		var (
			id, aid, sid int64
			studentName  string
			gradeValue   sql.NullInt64
			feedback     sql.NullString
			gradedBy     sql.NullInt64
			gradedAt     sql.NullTime
			statusStr    string
			createdAt    time.Time
			updatedAt    time.Time
		)
		if err := rows.Scan(&id, &aid, &sid, &studentName,
			&gradeValue, &feedback, &gradedBy, &gradedAt,
			&statusStr, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("submissions: list by assignment scan: %w", err)
		}

		var gv *int
		if gradeValue.Valid {
			v := int(gradeValue.Int64)
			gv = &v
		}
		var gb *int64
		if gradedBy.Valid {
			gb = &gradedBy.Int64
		}
		var ga *time.Time
		if gradedAt.Valid {
			t := gradedAt.Time
			ga = &t
		}

		out = append(out, views.SubmissionView{
			ID:           id,
			AssignmentID: aid,
			StudentID:    sid,
			StudentName:  studentName,
			GradeValue:   gv,
			Feedback:     feedback.String,
			GradedBy:     gb,
			GradedAt:     ga,
			Status:       entities.SubmissionStatus(statusStr),
			CreatedAt:    createdAt,
			UpdatedAt:    updatedAt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("submissions: list by assignment iter: %w", err)
	}
	return out, nil
}
