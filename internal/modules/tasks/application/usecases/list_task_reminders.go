package usecases

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/entities"
)

// ListTaskRemindersInput pins the read scope: caller sees only
// their own reminders for the supplied task (per-user privacy).
type ListTaskRemindersInput struct {
	TaskID      int64
	ActorUserID int64
}

// ListTaskRemindersUseCase returns the caller's reminders для a
// given task. Returns empty (non-nil) slice on no matches so JSON
// renders [] not null.
type ListTaskRemindersUseCase struct {
	repo TaskReminderRepository
}

// NewListTaskRemindersUseCase wires the use case. Panics on nil
// repo.
func NewListTaskRemindersUseCase(repo TaskReminderRepository) *ListTaskRemindersUseCase {
	if repo == nil {
		panic("tasks: NewListTaskRemindersUseCase requires non-nil repo")
	}
	return &ListTaskRemindersUseCase{repo: repo}
}

// Execute returns the caller's reminders для the supplied task.
// Repo errors propagate as-is.
func (uc *ListTaskRemindersUseCase) Execute(ctx context.Context, in ListTaskRemindersInput) ([]*entities.TaskReminder, error) {
	return uc.repo.ListByTaskAndUser(ctx, in.TaskID, in.ActorUserID)
}
