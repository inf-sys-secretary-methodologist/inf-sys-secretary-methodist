package entities

import (
	"errors"
	"time"
)

// ErrAlreadyGraded indicates that Grade was called on a Submission whose
// status is already graded. Re-grading must go through an explicit
// "returned" transition (issued for revisions in a later release) so that
// audit history is preserved. Handlers map this sentinel to HTTP 409.
var ErrAlreadyGraded = errors.New("assignments: submission already graded")

// SubmissionStatus is the typed enum mirroring the SQL CHECK on
// submissions.status. It exists so domain code can never pass a "magic
// string" through a Submission.
type SubmissionStatus string

const (
	StatusPending  SubmissionStatus = "pending"
	StatusGraded   SubmissionStatus = "graded"
	StatusReturned SubmissionStatus = "returned"
)

// IsValid reports whether s is one of the recognised statuses. Repository
// reads use this when reconstituting a Submission from a row.
func (s SubmissionStatus) IsValid() bool {
	switch s {
	case StatusPending, StatusGraded, StatusReturned:
		return true
	default:
		return false
	}
}

// Submission represents a single student's record against an assignment.
// One Submission is created per (assignment, student) pair as soon as
// the assignment is published, with status=Pending and no grade. The
// teacher transitions it into Graded by calling Grade.
type Submission struct {
	ID           int64
	AssignmentID int64
	StudentID    int64

	gradeValue *int
	feedback   string
	gradedBy   *int64
	gradedAt   *time.Time
	status     SubmissionStatus
	createdAt  time.Time
	updatedAt  time.Time
}

// NewSubmission creates a fresh Submission in Pending state. The clock is
// injected so tests stay deterministic.
func NewSubmission(assignmentID, studentID int64, now time.Time) *Submission {
	return &Submission{
		AssignmentID: assignmentID,
		StudentID:    studentID,
		status:       StatusPending,
		createdAt:    now,
		updatedAt:    now,
	}
}

// ReconstituteSubmission rebuilds a Submission from a storage row.
// Bypasses transition rules because the persisted state is already the
// outcome of those rules. Used exclusively by repository implementations.
func ReconstituteSubmission(
	id, assignmentID, studentID int64,
	gradeValue *int, feedback string, gradedBy *int64, gradedAt *time.Time,
	status SubmissionStatus, createdAt, updatedAt time.Time,
) *Submission {
	return &Submission{
		ID:           id,
		AssignmentID: assignmentID,
		StudentID:    studentID,
		gradeValue:   gradeValue,
		feedback:     feedback,
		gradedBy:     gradedBy,
		gradedAt:     gradedAt,
		status:       status,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
}

// Status exposes the current status (read-only).
func (s *Submission) Status() SubmissionStatus { return s.status }

// IsGraded is sugar for Status() == StatusGraded.
func (s *Submission) IsGraded() bool { return s.status == StatusGraded }

// GradeValue returns the awarded grade if any. Nil for Pending or Returned.
func (s *Submission) GradeValue() *int { return s.gradeValue }

// Feedback returns the grader's free-text feedback.
func (s *Submission) Feedback() string { return s.feedback }

// GradedBy returns the user ID of the grader.
func (s *Submission) GradedBy() *int64 { return s.gradedBy }

// GradedAt returns when the grade was awarded.
func (s *Submission) GradedAt() *time.Time { return s.gradedAt }

// CreatedAt returns the creation timestamp.
func (s *Submission) CreatedAt() time.Time { return s.createdAt }

// UpdatedAt returns the last-update timestamp.
func (s *Submission) UpdatedAt() time.Time { return s.updatedAt }

// Grade records score, feedback, grader and timestamp on the submission.
// Returns ErrAlreadyGraded if the submission is already in Graded state;
// callers must transition through an explicit "returned" first to grade
// twice.
func (s *Submission) Grade(score Score, feedback string, gradedBy int64, now time.Time) error {
	if s.status == StatusGraded {
		return ErrAlreadyGraded
	}
	v := score.Value()
	s.gradeValue = &v
	s.feedback = feedback
	s.gradedBy = &gradedBy
	s.gradedAt = &now
	s.status = StatusGraded
	s.updatedAt = now
	return nil
}
