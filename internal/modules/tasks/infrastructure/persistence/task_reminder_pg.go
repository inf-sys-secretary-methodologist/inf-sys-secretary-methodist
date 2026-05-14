// Package persistence — TaskReminderRepositoryPG implements the
// TaskReminderRepository port using PostgreSQL. Greenfield в v0.138.0.
package persistence

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/repositories"
)

// ErrTaskReminderNotFound is returned by GetByID + Delete when the
// requested row is absent — disambiguates a 404 in the handler.
var ErrTaskReminderNotFound = errors.New("task_reminder: not found")

// TaskReminderRepositoryPG persists TaskReminder rows in PostgreSQL.
type TaskReminderRepositoryPG struct {
	db *sql.DB
}

// NewTaskReminderRepositoryPG builds the repository against the
// given DB handle. Caller owns the lifecycle of *sql.DB.
func NewTaskReminderRepositoryPG(db *sql.DB) *TaskReminderRepositoryPG {
	return &TaskReminderRepositoryPG{db: db}
}

// Compile-time assertion that the concrete type satisfies the port.
var _ repositories.TaskReminderRepository = (*TaskReminderRepositoryPG)(nil)

// Create inserts a new reminder, returning the assigned id by
// scanning the RETURNING clause back onto the entity via hydrate.
func (r *TaskReminderRepositoryPG) Create(ctx context.Context, reminder *entities.TaskReminder) error {
	return errors.New("task_reminder_pg: Create not implemented yet")
}

// Delete removes the row. Returns ErrTaskReminderNotFound if no
// row matches (RowsAffected == 0).
func (r *TaskReminderRepositoryPG) Delete(ctx context.Context, id int64) error {
	return errors.New("task_reminder_pg: Delete not implemented yet")
}

// GetByID returns the row by primary key. Returns
// ErrTaskReminderNotFound if absent.
func (r *TaskReminderRepositoryPG) GetByID(ctx context.Context, id int64) (*entities.TaskReminder, error) {
	return nil, errors.New("task_reminder_pg: GetByID not implemented yet")
}

// ListByTaskAndUser returns all reminders для (taskID, userID).
// Empty slice (not nil) on no matches. ordered by created_at ASC.
func (r *TaskReminderRepositoryPG) ListByTaskAndUser(ctx context.Context, taskID, userID int64) ([]*entities.TaskReminder, error) {
	return nil, errors.New("task_reminder_pg: ListByTaskAndUser not implemented yet")
}

// GetPendingReminders returns reminders whose trigger time has
// elapsed and which have NOT been dispatched. Trigger time is
// computed at SQL level: tasks.due_date - r.minutes_before *
// INTERVAL '1 minute'. Reminders whose parent task has NULL
// due_date are excluded (no trigger time computable).
func (r *TaskReminderRepositoryPG) GetPendingReminders(ctx context.Context, now time.Time) ([]*entities.TaskReminder, error) {
	return nil, errors.New("task_reminder_pg: GetPendingReminders not implemented yet")
}

// MarkSentBatch flips is_sent + sent_at for the supplied ids in one
// statement. Empty ids slice is a no-op (no SQL emitted).
func (r *TaskReminderRepositoryPG) MarkSentBatch(ctx context.Context, ids []int64, now time.Time) error {
	if len(ids) == 0 {
		return nil
	}
	return errors.New("task_reminder_pg: MarkSentBatch not implemented yet")
}
