package repositories

import (
	"context"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/entities"
)

// TaskReminderRepository defines persistence for task reminders.
// The interface is intentionally narrow:
//
//   - Create / Delete for the user-facing CRUD surface (POST + DELETE).
//   - ListByTaskAndUser для GET /api/tasks/:id/reminders filtered by
//     caller user_id (privacy boundary — per-user reminders).
//   - GetPendingReminders для the scheduler: returns reminders whose
//     trigger time (`tasks.due_date - minutes_before`) has elapsed
//     and which have NOT been dispatched yet. The repo encapsulates
//     the SQL JOIN + interval arithmetic so callers never reason
//     about timestamp math.
//   - MarkSentBatch is the bulk dispatch-flag flip after batch
//     processing (mirror к EventReminderRepository.MarkMultipleAsSent
//     для consistency).
type TaskReminderRepository interface {
	Create(ctx context.Context, reminder *entities.TaskReminder) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*entities.TaskReminder, error)
	ListByTaskAndUser(ctx context.Context, taskID, userID int64) ([]*entities.TaskReminder, error)
	GetPendingReminders(ctx context.Context, now time.Time) ([]*entities.TaskReminder, error)
	MarkSentBatch(ctx context.Context, ids []int64, now time.Time) error
}
