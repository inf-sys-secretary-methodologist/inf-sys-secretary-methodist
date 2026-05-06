package entities

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrAlreadyGraded indicates that Grade was called on a Submission whose
// status is already graded. Re-grading must go through an explicit
// "returned" transition (issued for revisions in a later release) so that
// audit history is preserved. Handlers map this sentinel to HTTP 409.
var ErrAlreadyGraded = errors.New("assignments: submission already graded")

// ErrAlreadyReturned indicates Return was called on a submission whose
// status is already returned. The student-side resubmit flow (returned
// → pending) is a separate use case. Handlers map this sentinel to 409.
var ErrAlreadyReturned = errors.New("assignments: submission already returned")

// ErrInvalidReturn signals violation of a Return invariant — empty
// reason, reason exceeding 4096 chars, or non-positive returnedBy.
// Handlers map this sentinel to 422 via a single errors.Is dispatch.
var ErrInvalidReturn = errors.New("assignments: invalid return")

// ErrNotReturned indicates Resubmit was called on a submission whose
// status is not "returned". Resubmit is the student-side counterpart of
// Return: it transitions back to pending only from a returned state.
// Handlers map this sentinel to HTTP 409.
var ErrNotReturned = errors.New("assignments: submission not in returned state")

// ErrSubmissionOwnerOnly indicates that an actor without ownership of
// this submission attempted a write that the domain restricts to its
// student — currently Resubmit. The sentinel is distinct from
// ErrAssignmentScopeForbidden (which is the teacher-side rule on the
// Assignment aggregate); a student is never the assignment author, so
// reusing that name would be misleading. Handlers map to HTTP 403.
var ErrSubmissionOwnerOnly = errors.New("assignments: caller is not the submission owner")

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
// the assignment is published, with status=Pending and no grade.
// Teachers transition it into Graded by calling Grade; teachers,
// methodists, academic_secretaries and system_admins can transition it
// (back) into Returned for revision by calling Return.
type Submission struct {
	ID           int64
	AssignmentID int64
	StudentID    int64

	gradeValue   *int
	feedback     string
	gradedBy     *int64
	gradedAt     *time.Time
	returnReason string
	returnedBy   *int64
	returnedAt   *time.Time
	status       SubmissionStatus
	createdAt    time.Time
	updatedAt    time.Time
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
	returnReason string, returnedBy *int64, returnedAt *time.Time,
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
		returnReason: returnReason,
		returnedBy:   returnedBy,
		returnedAt:   returnedAt,
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

// ReturnReason returns the explanation captured at return time.
func (s *Submission) ReturnReason() string { return s.returnReason }

// ReturnedBy returns the user id of the actor who returned the submission.
func (s *Submission) ReturnedBy() *int64 { return s.returnedBy }

// ReturnedAt returns when the submission was returned for revision.
func (s *Submission) ReturnedAt() *time.Time { return s.returnedAt }

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

// Return marks the submission as returned for revision. Allowed from
// Pending and Graded; rejected with ErrAlreadyReturned when already
// returned. On success, any prior grade is cleared on the entity —
// the audit log preserves the historical value.
//
// Invariants enforced here (each violation wraps ErrInvalidReturn):
//   - reason trimmed-non-empty
//   - reason ≤ 4096 chars after trim
//   - returnedBy > 0
func (s *Submission) Return(reason string, returnedBy int64, now time.Time) error {
	if s.status == StatusReturned {
		return ErrAlreadyReturned
	}
	trimmed := strings.TrimSpace(reason)
	if trimmed == "" {
		return fmt.Errorf("%w: reason must not be empty", ErrInvalidReturn)
	}
	if len(trimmed) > 4096 {
		return fmt.Errorf("%w: reason exceeds 4096 chars", ErrInvalidReturn)
	}
	if returnedBy <= 0 {
		return fmt.Errorf("%w: returnedBy must be positive", ErrInvalidReturn)
	}
	// Clear prior grade — the audit log preserves history.
	s.gradeValue = nil
	s.feedback = ""
	s.gradedBy = nil
	s.gradedAt = nil
	s.returnReason = trimmed
	s.returnedBy = &returnedBy
	s.returnedAt = &now
	s.status = StatusReturned
	s.updatedAt = now
	return nil
}

// AuthorizeResubmitter returns nil if actorID is the student who owns
// this submission, otherwise ErrSubmissionOwnerOnly. The actor must be
// strictly positive — a zero or negative actor id signals missing JWT
// context, which must never accidentally satisfy ownership even if a
// student record were ever stored with id 0.
//
// Centralised on the entity so the Resubmit use case does not duplicate
// the predicate inline; mirrors Assignment.AuthorizeGrader on the
// teacher side.
func (s *Submission) AuthorizeResubmitter(actorID int64) error {
	if actorID <= 0 || actorID != s.StudentID {
		return fmt.Errorf("%w: user %d is not the owner (%d)",
			ErrSubmissionOwnerOnly, actorID, s.StudentID)
	}
	return nil
}

// AuthorizeReader returns nil if actorID is the student who owns this
// submission, otherwise ErrSubmissionOwnerOnly. Same predicate as
// AuthorizeResubmitter but kept as a separate method so call sites read
// the verb that matches the operation (read vs mutation). If a third
// owner-only operation lands, fold both into a private helper.
func (s *Submission) AuthorizeReader(actorID int64) error {
	if actorID <= 0 || actorID != s.StudentID {
		return fmt.Errorf("%w: user %d is not the owner (%d)",
			ErrSubmissionOwnerOnly, actorID, s.StudentID)
	}
	return nil
}

// Resubmit transitions a returned submission back to pending so that the
// student can supply revisions for a fresh grading cycle. The return
// triple (return_reason / returned_by / returned_at) is cleared on the
// entity — the audit log preserves the history of why the work was sent
// back. The grade triple is left untouched: Return already cleared it,
// and Resubmit must not resurrect a stale grade.
//
// Allowed only from Returned. Other states reject with ErrNotReturned;
// handlers map that to HTTP 409.
func (s *Submission) Resubmit(now time.Time) error {
	if s.status != StatusReturned {
		return ErrNotReturned
	}
	s.returnReason = ""
	s.returnedBy = nil
	s.returnedAt = nil
	s.status = StatusPending
	s.updatedAt = now
	return nil
}
