package repositories

import (
	"context"
	"errors"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/views"
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

	// ListByAssignment returns the read-side projection of submissions
	// belonging to the given assignment, optionally filtered by status.
	// The student first/last name come from a JOIN with the users table
	// so the grading UI can render rows without a second round-trip.
	// A nil status means "any". An empty result is not an error.
	ListByAssignment(ctx context.Context, assignmentID int64, status *entities.SubmissionStatus) ([]views.SubmissionView, error)
}
