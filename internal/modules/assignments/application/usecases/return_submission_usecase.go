package usecases

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/entities"
)

// ReturnSubmissionNotifier is the narrow port through which the use case
// emits the "submission returned for revision" signal. Concrete adapters
// live at the DI seam (main.go), keeping the use case free of cross-module
// imports — same pattern as SaveGradeNotifier.
type ReturnSubmissionNotifier interface {
	NotifyReturned(ctx context.Context, studentID, assignmentID int64, reason string) error
}

// ReturnSubmissionInput is the use-case input contract.
type ReturnSubmissionInput struct {
	AssignmentID int64
	StudentID    int64
	Reason       string
}

// ReturnSubmissionUseCase records that a teacher / methodist /
// academic_secretary / system_admin has returned a student's submission
// for revision. Mirrors SaveGradeUseCase in shape: load, authorize,
// transition the entity, persist, notify (best-effort), audit.
//
// State transitions allowed: pending → returned, graded → returned.
// Already-returned is rejected by the entity (ErrAlreadyReturned → 409).
type ReturnSubmissionUseCase struct {
	assignmentRepo AssignmentRepository
	submissionRepo SubmissionRepository
	notifier       ReturnSubmissionNotifier
	auditSink      AuditSink
	clock          func() time.Time
}

// NewReturnSubmissionUseCase wires the use case. clock defaults to
// time.Now when nil so production callers do not have to supply one.
// auditSink takes the narrow AuditSink port — *logging.AuditLogger
// satisfies it structurally so production wiring stays unchanged.
func NewReturnSubmissionUseCase(
	assignmentRepo AssignmentRepository,
	submissionRepo SubmissionRepository,
	notifier ReturnSubmissionNotifier,
	auditSink AuditSink,
	clock func() time.Time,
) *ReturnSubmissionUseCase {
	if clock == nil {
		clock = time.Now
	}
	return &ReturnSubmissionUseCase{
		assignmentRepo: assignmentRepo,
		submissionRepo: submissionRepo,
		notifier:       notifier,
		auditSink:      auditSink,
		clock:          clock,
	}
}

// Execute fetches the assignment, authorizes the caller via
// AuthorizeGrader (same permission predicate as grading — anyone who can
// grade can also return), loads the submission (or creates a fresh one
// if none yet exists), applies Return, persists, and notifies the
// student. Notification failure is logged but does not abort.
//
// Errors surface domain sentinels (errors.Is-friendly):
//   - ErrAssignmentNotFound    → 404
//   - entities.ErrAssignmentScopeForbidden  → 403
//   - entities.ErrInvalidReturn             → 422
//   - entities.ErrAlreadyReturned           → 409
func (uc *ReturnSubmissionUseCase) Execute(ctx context.Context, actorID int64, in ReturnSubmissionInput) error {
	assignment, err := uc.assignmentRepo.GetByID(ctx, in.AssignmentID)
	if err != nil {
		return fmt.Errorf("return submission: load assignment: %w", err)
	}

	if err := assignment.AuthorizeGrader(actorID); err != nil {
		// Audit a denied attempt explicitly. Returning is a security-
		// relevant write; forensic trail must include refused attempts.
		emitAudit(uc.auditSink, ctx, actorID, "assignment.return_denied", map[string]any{
			"assignment_id": in.AssignmentID,
			"student_id":    in.StudentID,
			"reason":        "not_author",
		})
		return err
	}

	submission, err := uc.submissionRepo.GetByAssignmentAndStudent(ctx, in.AssignmentID, in.StudentID)
	switch {
	case errors.Is(err, ErrSubmissionNotFound):
		// First-touch return: methodist returns an upload that was never
		// graded. Mirror SaveGrade's first-grade-on-not-found pattern.
		submission = entities.NewSubmission(in.AssignmentID, in.StudentID, uc.clock())
	case err != nil:
		return fmt.Errorf("return submission: load submission: %w", err)
	}

	// Capture prior-grade state BEFORE Return clears it. The audit event
	// emitted at the end of the flow needs the values that Return is
	// about to wipe — forensic compliance: a reviewer reading the audit
	// log post-incident must see what was undone.
	prevGrade := submission.GradeValue()
	prevFeedback := submission.Feedback()

	if err := submission.Return(in.Reason, actorID, uc.clock()); err != nil {
		return err
	}

	if err := uc.submissionRepo.Save(ctx, submission); err != nil {
		return fmt.Errorf("return submission: persist: %w", err)
	}

	// Best-effort notification: a delivery failure must not roll back the
	// state transition. The grade has been cleared and the row persisted —
	// the student must eventually learn, but a transient SMTP outage
	// should not block the methodist's workflow. Failures are audited
	// separately so on-call can spot persistent outages.
	if uc.notifier != nil {
		if notifyErr := uc.notifier.NotifyReturned(ctx, in.StudentID, in.AssignmentID, in.Reason); notifyErr != nil {
			emitAudit(uc.auditSink, ctx, actorID, "assignment.return_notify_failed", map[string]any{
				"assignment_id": in.AssignmentID,
				"student_id":    in.StudentID,
				"error":         notifyErr.Error(),
			})
		}
	}

	returnedFields := map[string]any{
		"assignment_id": in.AssignmentID,
		"student_id":    in.StudentID,
		"reason":        in.Reason,
	}
	// Include previous_grade / previous_feedback only when there was a
	// prior grade to clear. Pending → Returned has no prior grade —
	// emitting empty/null keys would add noise to the audit stream.
	if prevGrade != nil {
		returnedFields["previous_grade"] = *prevGrade
		returnedFields["previous_feedback"] = prevFeedback
	}
	emitAudit(uc.auditSink, ctx, actorID, "assignment.returned", returnedFields)

	return nil
}
