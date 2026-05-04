package repositories

import (
	"context"
	"errors"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/entities"
)

// ErrSubmissionNotFound signals that no submission exists for the
// (assignment_id, student_id) pair. The SaveGrade use case treats this
// as "first grading" and creates a fresh Submission rather than failing.
var ErrSubmissionNotFound = errors.New("assignments: submission not found")

// SubmissionRepository is the persistence port for per-student grade
// records. Save is upsert-style: it inserts a new submission if ID==0
// and updates an existing row otherwise.
type SubmissionRepository interface {
	// GetByAssignmentAndStudent returns the submission for the given pair
	// or ErrSubmissionNotFound when no row exists.
	GetByAssignmentAndStudent(ctx context.Context, assignmentID, studentID int64) (*entities.Submission, error)

	// Save persists the submission (insert when ID==0, update otherwise).
	// On insert, Save sets the assigned ID on the entity in-place.
	Save(ctx context.Context, s *entities.Submission) error
}
