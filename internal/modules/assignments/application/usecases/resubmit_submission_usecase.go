package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/repositories"
)

// ResubmitSubmissionNotifier is the narrow port through which the use
// case emits the "submission resubmitted" signal — the teacher
// (assignment author) needs to know a fresh attempt awaits grading.
// Concrete adapters live at the DI seam (cmd/server/main.go), keeping
// this package free of cross-module Go imports — same pattern as
// ReturnSubmissionNotifier and SaveGradeNotifier.
type ResubmitSubmissionNotifier interface {
	NotifyResubmitted(ctx context.Context, teacherID, assignmentID int64) error
}

// ResubmitSubmissionInput is the use-case input contract. StudentID is
// passed in alongside actorID so the use case can load the right
// submission row; AuthorizeResubmitter then enforces actorID==StudentID
// at the domain layer (defense-in-depth — a handler that mismatches
// them is rejected at the entity boundary, not silently honored).
type ResubmitSubmissionInput struct {
	AssignmentID int64
	StudentID    int64
}

// ResubmitSubmissionUseCase records that a student has resubmitted
// their own returned work for a fresh grading cycle. Mirror of
// ReturnSubmissionUseCase in shape: load, authorise, transition the
// entity, persist, notify (best-effort), audit.
//
// State transition allowed: returned → pending. Other states reject
// at the entity boundary (ErrNotReturned → 409). Authorisation is
// own-submission only (Submission.AuthorizeResubmitter): a student
// may resubmit only their own work — anyone else is rejected with
// ErrSubmissionOwnerOnly → 403.
type ResubmitSubmissionUseCase struct {
	assignmentRepo repositories.AssignmentRepository
	submissionRepo repositories.SubmissionRepository
	notifier       ResubmitSubmissionNotifier
	auditSink      AuditSink
	clock          func() time.Time
}

// NewResubmitSubmissionUseCase wires the use case. clock defaults to
// time.Now when nil so production callers do not have to supply one.
// auditSink takes the narrow AuditSink port; *logging.AuditLogger
// satisfies it structurally so production wiring stays unchanged.
func NewResubmitSubmissionUseCase(
	assignmentRepo repositories.AssignmentRepository,
	submissionRepo repositories.SubmissionRepository,
	notifier ResubmitSubmissionNotifier,
	auditSink AuditSink,
	clock func() time.Time,
) *ResubmitSubmissionUseCase {
	if clock == nil {
		clock = time.Now
	}
	return &ResubmitSubmissionUseCase{
		assignmentRepo: assignmentRepo,
		submissionRepo: submissionRepo,
		notifier:       notifier,
		auditSink:      auditSink,
		clock:          clock,
	}
}

// Execute fetches the assignment (for the teacher id used by the
// notifier), loads the submission, authorises the caller as the
// submission's owning student, applies Resubmit, persists, and
// notifies the teacher. Notification failure is logged but does not
// abort the resubmit — the state transition is the system of record.
//
// Errors surface domain sentinels (errors.Is-friendly):
//   - repositories.ErrAssignmentNotFound   → 404
//   - repositories.ErrSubmissionNotFound   → 404
//   - entities.ErrSubmissionOwnerOnly      → 403
//   - entities.ErrNotReturned              → 409
func (uc *ResubmitSubmissionUseCase) Execute(ctx context.Context, actorID int64, in ResubmitSubmissionInput) error {
	assignment, err := uc.assignmentRepo.GetByID(ctx, in.AssignmentID)
	if err != nil {
		return fmt.Errorf("resubmit submission: load assignment: %w", err)
	}

	submission, err := uc.submissionRepo.GetByAssignmentAndStudent(ctx, in.AssignmentID, in.StudentID)
	if err != nil {
		return fmt.Errorf("resubmit submission: load submission: %w", err)
	}

	if err := submission.AuthorizeResubmitter(actorID); err != nil {
		// Audit a denied resubmit explicitly. Resubmit is a security-
		// relevant write (it lets the student replace the teacher's
		// verdict cycle); forensic trail must include refused attempts,
		// not only successes. Mirrors return_denied / grade_denied.
		emitAudit(uc.auditSink, ctx, actorID, "assignment.resubmit_denied", map[string]any{
			"assignment_id": in.AssignmentID,
			"student_id":    in.StudentID,
			"reason":        "not_owner",
		})
		return err
	}

	// Capture the prior return_reason BEFORE Resubmit clears it. The
	// success audit event needs what the transition undid — symmetric
	// to the previous_grade / previous_feedback capture on the Return
	// side. A reviewer reading the audit log post-incident must see
	// why the work had originally been sent back.
	prevReason := submission.ReturnReason()

	if err := submission.Resubmit(uc.clock()); err != nil {
		return err
	}

	if err := uc.submissionRepo.Save(ctx, submission); err != nil {
		return fmt.Errorf("resubmit submission: persist: %w", err)
	}

	// Best-effort notification: a delivery failure must not roll back
	// the resubmit — the state transition is the system of record. The
	// teacher must eventually learn a fresh attempt awaits grading, but
	// a transient SMTP outage cannot block the student's workflow.
	// Failures are audited separately so on-call can spot persistent
	// outages.
	if uc.notifier != nil {
		if notifyErr := uc.notifier.NotifyResubmitted(ctx, assignment.TeacherID(), in.AssignmentID); notifyErr != nil {
			emitAudit(uc.auditSink, ctx, actorID, "assignment.resubmit_notify_failed", map[string]any{
				"assignment_id": in.AssignmentID,
				"student_id":    in.StudentID,
				"error":         notifyErr.Error(),
			})
		}
	}

	emitAudit(uc.auditSink, ctx, actorID, "assignment.resubmitted", map[string]any{
		"assignment_id":          in.AssignmentID,
		"student_id":             in.StudentID,
		"previous_return_reason": prevReason,
	})

	return nil
}
