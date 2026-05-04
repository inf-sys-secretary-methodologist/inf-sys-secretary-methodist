// Package usecases provides application use cases for the assignments
// module — the academic Tasks Context.
package usecases

import (
	"context"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

// SaveGradeNotifier is a narrow port through which the SaveGrade use case
// emits the "submission graded" signal. Concrete implementations adapt
// this to the platform NotificationUseCase at the DI seam — keeping the
// use case free of cross-module imports.
type SaveGradeNotifier interface {
	NotifyGraded(ctx context.Context, studentID, assignmentID int64, score, maxScore int) error
}

// SaveGradeInput is the use-case input contract.
type SaveGradeInput struct {
	AssignmentID int64
	StudentID    int64
	Value        int
	Feedback     string
}

// SaveGradeUseCase records a teacher's grade on a student's submission
// for a given assignment. It is the only entry point that mutates
// submission state in the academic Tasks Context.
type SaveGradeUseCase struct {
	assignmentRepo repositories.AssignmentRepository
	submissionRepo repositories.SubmissionRepository
	notifier       SaveGradeNotifier
	auditLogger    *logging.AuditLogger
	clock          func() time.Time
}

// NewSaveGradeUseCase wires the use case. clock defaults to time.Now when
// nil so production callers do not have to supply one.
func NewSaveGradeUseCase(
	assignmentRepo repositories.AssignmentRepository,
	submissionRepo repositories.SubmissionRepository,
	notifier SaveGradeNotifier,
	auditLogger *logging.AuditLogger,
	clock func() time.Time,
) *SaveGradeUseCase {
	if clock == nil {
		clock = time.Now
	}
	return &SaveGradeUseCase{
		assignmentRepo: assignmentRepo,
		submissionRepo: submissionRepo,
		notifier:       notifier,
		auditLogger:    auditLogger,
		clock:          clock,
	}
}

// Execute fetches the assignment, authorises the caller as grader,
// validates the score against the assignment's max, applies (or creates
// then applies) the grade transition on the submission, persists, and
// notifies the student. Notification failure is logged but does not
// abort the grading — recording the grade is the system of record.
//
// Errors surface domain sentinels (errors.Is-friendly):
//   - repositories.ErrAssignmentNotFound      → 404
//   - entities.ErrAssignmentScopeForbidden    → 403
//   - entities.ErrInvalidScore                → 422
//   - entities.ErrAlreadyGraded               → 409
func (uc *SaveGradeUseCase) Execute(ctx context.Context, teacherID int64, in SaveGradeInput) error {
	// stub for RED — no-op
	_ = ctx
	_ = teacherID
	_ = in
	return nil
}
