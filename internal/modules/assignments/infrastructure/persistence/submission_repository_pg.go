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

// SubmissionRepositoryPG is the SQL implementation of SubmissionRepository.
type SubmissionRepositoryPG struct {
	db *sql.DB
}

// NewSubmissionRepositoryPG constructs the repository.
func NewSubmissionRepositoryPG(db *sql.DB) *SubmissionRepositoryPG {
	return &SubmissionRepositoryPG{db: db}
}

const submissionSelectColumns = `id, assignment_id, student_id, grade_value, feedback, graded_by, graded_at, status, created_at, updated_at`

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
		createdAt    time.Time
		updatedAt    time.Time
	)
	err := r.db.QueryRowContext(ctx, query, assignmentID, studentID).Scan(
		&id, &aid, &sid, &gradeValue, &feedback, &gradedBy, &gradedAt,
		&status, &createdAt, &updatedAt,
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
		ga = &gradedAt.Time
	}
	return entities.ReconstituteSubmission(
		id, aid, sid, gv, feedback.String, gb, ga,
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
			assignment_id, student_id, grade_value, feedback, graded_by, graded_at, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (assignment_id, student_id) DO UPDATE SET
			grade_value = EXCLUDED.grade_value,
			feedback    = EXCLUDED.feedback,
			graded_by   = EXCLUDED.graded_by,
			graded_at   = EXCLUDED.graded_at,
			status      = EXCLUDED.status,
			updated_at  = EXCLUDED.updated_at
		RETURNING id`

	gradeValue := nullableInt(s.GradeValue())
	gradedBy := nullableInt64(s.GradedBy())
	gradedAt := nullableTime(s.GradedAt())

	var newID int64
	err := r.db.QueryRowContext(ctx, query,
		s.AssignmentID, s.StudentID, gradeValue, s.Feedback(),
		gradedBy, gradedAt, string(s.Status()),
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
