package usecases

import (
	"context"
	"errors"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/repositories"
)

// ErrReminderOwnerOnly is returned when the caller is not the user
// who owns the reminder. Handler maps к 403.
var ErrReminderOwnerOnly = errors.New("task_reminder: caller is not reminder owner")

// ErrReminderNotFoundForTask is returned when the reminder exists
// но belongs to a different task than the URL path addresses.
// Surface as 404 in the handler so we do not leak the row's actual
// task_id to a probing caller.
var ErrReminderNotFoundForTask = errors.New("task_reminder: not found under supplied task")

// DeleteReminderInput captures the deletion target + caller.
// TaskID is included so the use case can refuse a request that
// addresses a reminder under the wrong task path (defense in
// depth against URL-tampering).
type DeleteReminderInput struct {
	ReminderID  int64
	TaskID      int64
	ActorUserID int64
}

// DeleteReminderUseCase removes a reminder after verifying:
//   - the row exists (else 404 from underlying repo not-found),
//   - the row belongs to the supplied task (else 404
//     ErrReminderNotFoundForTask — URL-path vs row mismatch is
//     indistinguishable from missing row from the caller's
//     perspective; we don't leak ownership existence),
//   - the caller is the owner (else 403 ErrReminderOwnerOnly).
//
// Audit emitted on success.
type DeleteReminderUseCase struct {
	repo  repositories.TaskReminderRepository
	audit AuditSink
}

// NewDeleteReminderUseCase wires the use case. Panics on nil repo.
// audit may be nil — emission skipped (test-friendly).
func NewDeleteReminderUseCase(repo repositories.TaskReminderRepository, audit AuditSink) *DeleteReminderUseCase {
	if repo == nil {
		panic("tasks: NewDeleteReminderUseCase requires non-nil repo")
	}
	return &DeleteReminderUseCase{repo: repo, audit: audit}
}

// Execute loads → task-scope check → ownership check → deletes →
// audit. Order pinned: existence first, then task scope, then
// ownership, then mutation. Audit emit only on successful delete.
func (uc *DeleteReminderUseCase) Execute(ctx context.Context, in DeleteReminderInput) error {
	reminder, err := uc.repo.GetByID(ctx, in.ReminderID)
	if err != nil {
		return err
	}
	if reminder.TaskID() != in.TaskID {
		return ErrReminderNotFoundForTask
	}
	if reminder.UserID() != in.ActorUserID {
		return ErrReminderOwnerOnly
	}
	if err := uc.repo.Delete(ctx, in.ReminderID); err != nil {
		return err
	}
	uc.emitAudit(ctx, reminder)
	return nil
}

// emitAudit logs a task_reminder.deleted forensic event. Nil-safe.
func (uc *DeleteReminderUseCase) emitAudit(ctx context.Context, reminder *entities.TaskReminder) {
	if uc.audit == nil {
		return
	}
	uc.audit.LogAuditEvent(ctx, "task_reminder.deleted", "task_reminder", map[string]any{
		"reminder_id":    reminder.ID(),
		"task_id":        reminder.TaskID(),
		"user_id":        reminder.UserID(),
		"reminder_type":  string(reminder.ReminderType()),
		"minutes_before": reminder.MinutesBefore(),
	})
}
