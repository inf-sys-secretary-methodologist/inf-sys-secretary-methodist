package usecases

import (
	"context"
	"errors"

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
//     ErrReminderNotFoundForTask),
//   - the caller is the owner (else 403 ErrReminderOwnerOnly).
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
// audit.
//
// Stub for RED — GREEN replaces the body with the real composition.
func (uc *DeleteReminderUseCase) Execute(ctx context.Context, in DeleteReminderInput) error {
	_ = ctx
	_ = in
	return errors.New("delete_reminder: not implemented yet")
}
