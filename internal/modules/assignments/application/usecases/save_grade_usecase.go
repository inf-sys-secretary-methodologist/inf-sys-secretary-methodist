// Package usecases provides application use cases for the assignments
// module — the academic Tasks Context.
package usecases

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/entities"
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
	assignment, err := uc.assignmentRepo.GetByID(ctx, in.AssignmentID)
	if err != nil {
		return fmt.Errorf("save grade: load assignment: %w", err)
	}

	if err := assignment.AuthorizeGrader(teacherID); err != nil {
		// Audit a denied grading attempt explicitly. A grading flow is
		// security-relevant; forensic trail must include refused
		// attempts, not only successes. Failure-closed bias from
		// v0.108.3 carried through.
		uc.logAudit(ctx, teacherID, "assignment.grade_denied", map[string]any{
			"assignment_id": in.AssignmentID,
			"student_id":    in.StudentID,
			"reason":        "not_author",
		})
		return err
	}

	score, err := assignment.NewSubmissionScore(in.Value)
	if err != nil {
		return err
	}

	submission, err := uc.submissionRepo.GetByAssignmentAndStudent(ctx, in.AssignmentID, in.StudentID)
	switch {
	case errors.Is(err, repositories.ErrSubmissionNotFound):
		submission = entities.NewSubmission(in.AssignmentID, in.StudentID, uc.clock())
	case err != nil:
		return fmt.Errorf("save grade: load submission: %w", err)
	}

	if err := submission.Grade(score, in.Feedback, teacherID, uc.clock()); err != nil {
		return err
	}

	if err := uc.submissionRepo.Save(ctx, submission); err != nil {
		return fmt.Errorf("save grade: persist submission: %w", err)
	}

	// Notification is best-effort: a delivery failure must not roll back
	// the grade — the grade is the system of record. Failures are audited
	// separately so on-call can notice persistent outages.
	if uc.notifier != nil {
		if notifyErr := uc.notifier.NotifyGraded(ctx, in.StudentID, in.AssignmentID, in.Value, assignment.MaxScore()); notifyErr != nil {
			uc.logAudit(ctx, teacherID, "assignment.grade_notify_failed", map[string]any{
				"assignment_id": in.AssignmentID,
				"student_id":    in.StudentID,
				"error":         notifyErr.Error(),
			})
		}
	}

	uc.logAudit(ctx, teacherID, "assignment.graded", map[string]any{
		"assignment_id": in.AssignmentID,
		"student_id":    in.StudentID,
		"score":         in.Value,
		"max_score":     assignment.MaxScore(),
	})

	return nil
}

func (uc *SaveGradeUseCase) logAudit(ctx context.Context, teacherID int64, action string, fields map[string]any) {
	if uc.auditLogger == nil {
		return
	}
	enriched := map[string]any{"actor_user_id": teacherID}
	for k, v := range fields {
		enriched[k] = v
	}
	uc.auditLogger.LogAuditEvent(ctx, action, "assignment", enriched)
}
