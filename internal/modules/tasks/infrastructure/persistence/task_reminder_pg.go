// Package persistence — TaskReminderRepositoryPG implements the
// TaskReminderRepository port using PostgreSQL. Greenfield в v0.138.0.
package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/entities"
)

// ErrTaskReminderNotFound is returned by GetByID + Delete when the
// requested row is absent — disambiguates a 404 in the handler.
var ErrTaskReminderNotFound = errors.New("task_reminder: not found")

// reminderSelectColumns — canonical column order. All SELECTs use
// this constant so a column shape change is one edit.
const reminderSelectColumns = `id, task_id, user_id, reminder_type, minutes_before, is_sent, sent_at, created_at`

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
var _ usecases.TaskReminderRepository = (*TaskReminderRepositoryPG)(nil)

// Create inserts a new reminder, returning the assigned id and
// created_at by scanning the RETURNING clause back onto the entity
// via hydrate (preserves domain encapsulation — no public setter).
func (r *TaskReminderRepositoryPG) Create(ctx context.Context, reminder *entities.TaskReminder) error {
	const query = `INSERT INTO task_reminders (task_id, user_id, reminder_type, minutes_before, is_sent) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`
	var (
		id        int64
		createdAt time.Time
	)
	err := r.db.QueryRowContext(ctx, query,
		reminder.TaskID(),
		reminder.UserID(),
		string(reminder.ReminderType()),
		reminder.MinutesBefore(),
		reminder.IsSent(),
	).Scan(&id, &createdAt)
	if err != nil {
		return fmt.Errorf("task_reminder: insert failed: %w", err)
	}
	*reminder = *entities.HydrateFromPersistence(
		id,
		reminder.TaskID(),
		reminder.UserID(),
		reminder.ReminderType(),
		reminder.MinutesBefore(),
		reminder.IsSent(),
		reminder.SentAt(),
		createdAt,
	)
	return nil
}

// Delete removes the row. Returns ErrTaskReminderNotFound if no
// row matches (RowsAffected == 0).
func (r *TaskReminderRepositoryPG) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM task_reminders WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("task_reminder: delete failed: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("task_reminder: failed to inspect delete result: %w", err)
	}
	if rows == 0 {
		return ErrTaskReminderNotFound
	}
	return nil
}

// GetByID returns the row by primary key. Returns
// ErrTaskReminderNotFound if absent.
func (r *TaskReminderRepositoryPG) GetByID(ctx context.Context, id int64) (*entities.TaskReminder, error) {
	query := `SELECT ` + reminderSelectColumns + ` FROM task_reminders WHERE id = $1`
	reminder, err := scanReminderRow(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTaskReminderNotFound
		}
		return nil, fmt.Errorf("task_reminder: get failed: %w", err)
	}
	return reminder, nil
}

// ListByTaskAndUser returns all reminders для (taskID, userID).
// Empty (non-nil) slice on no matches. Ordered by created_at ASC
// so consumers render в creation order.
func (r *TaskReminderRepositoryPG) ListByTaskAndUser(ctx context.Context, taskID, userID int64) ([]*entities.TaskReminder, error) {
	query := `SELECT ` + reminderSelectColumns + ` FROM task_reminders WHERE task_id = $1 AND user_id = $2 ORDER BY created_at ASC`
	rows, err := r.db.QueryContext(ctx, query, taskID, userID)
	if err != nil {
		return nil, fmt.Errorf("task_reminder: list failed: %w", err)
	}
	defer func() { _ = rows.Close() }()
	return scanReminderRows(rows)
}

// GetPendingReminders returns reminders whose trigger time has
// elapsed and which have NOT been dispatched. Trigger time is
// computed at SQL level: tasks.due_date - r.minutes_before *
// INTERVAL '1 minute'. Reminders whose parent task has NULL
// due_date are excluded (no trigger time computable).
//
// LIMIT 100 caps batch size to match the existing
// ReminderScheduler.batchSize default — prevents the scheduler
// from holding a multi-thousand-row result set during dispatch.
func (r *TaskReminderRepositoryPG) GetPendingReminders(ctx context.Context, now time.Time) ([]*entities.TaskReminder, error) {
	const query = `SELECT r.id, r.task_id, r.user_id, r.reminder_type, r.minutes_before, r.is_sent, r.sent_at, r.created_at FROM task_reminders r JOIN tasks t ON r.task_id = t.id WHERE r.is_sent = FALSE AND t.due_date IS NOT NULL AND t.due_date - (r.minutes_before * INTERVAL '1 minute') <= $1 ORDER BY r.id ASC LIMIT 100`
	rows, err := r.db.QueryContext(ctx, query, now)
	if err != nil {
		return nil, fmt.Errorf("task_reminder: pending lookup failed: %w", err)
	}
	defer func() { _ = rows.Close() }()
	return scanReminderRows(rows)
}

// MarkSentBatch flips is_sent + sent_at for the supplied ids в one
// statement. Empty ids slice is a no-op (no SQL emitted).
func (r *TaskReminderRepositoryPG) MarkSentBatch(ctx context.Context, ids []int64, now time.Time) error {
	if len(ids) == 0 {
		return nil
	}
	const query = `UPDATE task_reminders SET is_sent = TRUE, sent_at = $1 WHERE id = ANY($2)`
	if _, err := r.db.ExecContext(ctx, query, now, pq.Array(ids)); err != nil {
		return fmt.Errorf("task_reminder: mark sent batch failed: %w", err)
	}
	return nil
}

// rowScanner is the minimal interface both *sql.Row and *sql.Rows
// satisfy via Scan. Lets scanReminderRow service both
// QueryRowContext and Next-iteration без duplication.
type rowScanner interface {
	Scan(dest ...any) error
}

// scanReminderRow scans the canonical 8 columns into a fresh
// TaskReminder via HydrateFromPersistence so domain encapsulation
// stays intact.
func scanReminderRow(scanner rowScanner) (*entities.TaskReminder, error) {
	var (
		id            int64
		taskID        int64
		userID        int64
		reminderType  string
		minutesBefore int
		isSent        bool
		sentAt        sql.NullTime
		createdAt     time.Time
	)
	if err := scanner.Scan(&id, &taskID, &userID, &reminderType, &minutesBefore, &isSent, &sentAt, &createdAt); err != nil {
		return nil, err
	}
	var sentAtPtr *time.Time
	if sentAt.Valid {
		t := sentAt.Time
		sentAtPtr = &t
	}
	return entities.HydrateFromPersistence(id, taskID, userID, entities.ReminderType(reminderType), minutesBefore, isSent, sentAtPtr, createdAt), nil
}

// scanReminderRows drains a *sql.Rows iterator into a slice. Empty
// iterator returns an empty (non-nil) slice so JSON renders [] not
// null.
func scanReminderRows(rows *sql.Rows) ([]*entities.TaskReminder, error) {
	out := []*entities.TaskReminder{}
	for rows.Next() {
		reminder, err := scanReminderRow(rows)
		if err != nil {
			return nil, fmt.Errorf("task_reminder: scan failed: %w", err)
		}
		out = append(out, reminder)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("task_reminder: rows iteration failed: %w", err)
	}
	return out, nil
}
