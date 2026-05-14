package usecases

import (
	"context"
	"errors"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/repositories"
)

// ListTaskRemindersInput pins the read scope: caller sees only
// their own reminders for the supplied task (per-user privacy).
type ListTaskRemindersInput struct {
	TaskID      int64
	ActorUserID int64
}

// ListTaskRemindersUseCase returns the caller's reminders для a
// given task.
type ListTaskRemindersUseCase struct {
	repo repositories.TaskReminderRepository
}

// NewListTaskRemindersUseCase wires the use case. Panics on nil
// repo.
func NewListTaskRemindersUseCase(repo repositories.TaskReminderRepository) *ListTaskRemindersUseCase {
	if repo == nil {
		panic("tasks: NewListTaskRemindersUseCase requires non-nil repo")
	}
	return &ListTaskRemindersUseCase{repo: repo}
}

// Execute returns the caller's reminders для the supplied task.
//
// Stub for RED — GREEN replaces the body with the repo call.
func (uc *ListTaskRemindersUseCase) Execute(ctx context.Context, in ListTaskRemindersInput) ([]*entities.TaskReminder, error) {
	_ = ctx
	_ = in
	return nil, errors.New("list_task_reminders: not implemented yet")
}
