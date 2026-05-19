package persistence

// v0.153.7 Phase 6 backfill — closes uncovered branches in
// TaskReminderRepositoryPG: transport errors after QueryContext/ExecContext,
// RowsAffected inspection failure, mid-iteration scan error, rows.Err
// propagation, и the sentAt.Valid populated branch inside scanReminderRow.
//
// Mirrors existing per-file conventions (sqlmock.QueryMatcherEqual exact
// matching + WithArgs pinning). No production change.

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/entities"
)

func TestCreate_InsertError(t *testing.T) {
	repo, mock, cleanup := newReminderRepoMock(t)
	defer cleanup()
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	reminder, err := entities.NewTaskReminder(42, 7, entities.ReminderTypeTelegram, 15, now)
	require.NoError(t, err)

	mock.ExpectQuery(`INSERT INTO task_reminders (task_id, user_id, reminder_type, minutes_before, is_sent) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`).
		WithArgs(int64(42), int64(7), "telegram", 15, false).
		WillReturnError(fmt.Errorf("conn refused"))

	err = repo.Create(context.Background(), reminder)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "insert failed")
}

func TestDelete_ExecError(t *testing.T) {
	repo, mock, cleanup := newReminderRepoMock(t)
	defer cleanup()

	mock.ExpectExec(`DELETE FROM task_reminders WHERE id = $1`).
		WithArgs(int64(101)).
		WillReturnError(fmt.Errorf("conn refused"))

	err := repo.Delete(context.Background(), 101)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "delete failed")
}

func TestDelete_RowsAffectedError(t *testing.T) {
	repo, mock, cleanup := newReminderRepoMock(t)
	defer cleanup()

	mock.ExpectExec(`DELETE FROM task_reminders WHERE id = $1`).
		WithArgs(int64(101)).
		WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("rows affected error")))

	err := repo.Delete(context.Background(), 101)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "inspect delete result")
}

func TestGetByID_TransportError(t *testing.T) {
	// Non-NoRows error path: wraps with "get failed", does NOT map to
	// ErrTaskReminderNotFound sentinel (unlike sql.ErrNoRows).
	repo, mock, cleanup := newReminderRepoMock(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, task_id, user_id, reminder_type, minutes_before, is_sent, sent_at, created_at FROM task_reminders WHERE id = $1`).
		WithArgs(int64(101)).
		WillReturnError(fmt.Errorf("conn refused"))

	_, err := repo.GetByID(context.Background(), 101)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "get failed")
}

func TestGetByID_SentAtPopulated(t *testing.T) {
	// Covers `if sentAt.Valid` branch in scanReminderRow (line 174-177).
	repo, mock, cleanup := newReminderRepoMock(t)
	defer cleanup()
	created := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	sent := created.Add(15 * time.Minute)

	rows := sqlmock.NewRows([]string{"id", "task_id", "user_id", "reminder_type", "minutes_before", "is_sent", "sent_at", "created_at"}).
		AddRow(int64(101), int64(42), int64(7), "telegram", 15, true, sent, created)
	mock.ExpectQuery(`SELECT id, task_id, user_id, reminder_type, minutes_before, is_sent, sent_at, created_at FROM task_reminders WHERE id = $1`).
		WithArgs(int64(101)).
		WillReturnRows(rows)

	r, err := repo.GetByID(context.Background(), 101)
	require.NoError(t, err)
	require.NotNil(t, r.SentAt())
	assert.True(t, r.SentAt().Equal(sent))
	assert.True(t, r.IsSent())
}

func TestListByTaskAndUser_QueryError(t *testing.T) {
	repo, mock, cleanup := newReminderRepoMock(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, task_id, user_id, reminder_type, minutes_before, is_sent, sent_at, created_at FROM task_reminders WHERE task_id = $1 AND user_id = $2 ORDER BY created_at ASC`).
		WithArgs(int64(42), int64(7)).
		WillReturnError(fmt.Errorf("conn refused"))

	_, err := repo.ListByTaskAndUser(context.Background(), 42, 7)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "list failed")
}

func TestListByTaskAndUser_ScanError(t *testing.T) {
	// scanReminderRows propagates rows.Scan errors via "scan failed".
	repo, mock, cleanup := newReminderRepoMock(t)
	defer cleanup()

	// Wrong column count triggers scan error inside the iteration loop.
	rows := sqlmock.NewRows([]string{"id"}).AddRow(int64(1))
	mock.ExpectQuery(`SELECT id, task_id, user_id, reminder_type, minutes_before, is_sent, sent_at, created_at FROM task_reminders WHERE task_id = $1 AND user_id = $2 ORDER BY created_at ASC`).
		WithArgs(int64(42), int64(7)).
		WillReturnRows(rows)

	_, err := repo.ListByTaskAndUser(context.Background(), 42, 7)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "scan failed")
}

func TestListByTaskAndUser_RowsErrPropagates(t *testing.T) {
	// scanReminderRows propagates iteration errors via "rows iteration failed".
	repo, mock, cleanup := newReminderRepoMock(t)
	defer cleanup()
	created := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)

	rows := sqlmock.NewRows([]string{"id", "task_id", "user_id", "reminder_type", "minutes_before", "is_sent", "sent_at", "created_at"}).
		AddRow(int64(101), int64(42), int64(7), "telegram", 15, false, nil, created).
		RowError(0, fmt.Errorf("connection reset"))
	mock.ExpectQuery(`SELECT id, task_id, user_id, reminder_type, minutes_before, is_sent, sent_at, created_at FROM task_reminders WHERE task_id = $1 AND user_id = $2 ORDER BY created_at ASC`).
		WithArgs(int64(42), int64(7)).
		WillReturnRows(rows)

	_, err := repo.ListByTaskAndUser(context.Background(), 42, 7)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rows iteration failed")
}

func TestGetPendingReminders_QueryError(t *testing.T) {
	repo, mock, cleanup := newReminderRepoMock(t)
	defer cleanup()
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)

	mock.ExpectQuery(`SELECT r.id, r.task_id, r.user_id, r.reminder_type, r.minutes_before, r.is_sent, r.sent_at, r.created_at FROM task_reminders r JOIN tasks t ON r.task_id = t.id WHERE r.is_sent = FALSE AND t.due_date IS NOT NULL AND t.due_date - (r.minutes_before * INTERVAL '1 minute') <= $1 ORDER BY r.id ASC LIMIT 100`).
		WithArgs(now).
		WillReturnError(fmt.Errorf("conn refused"))

	_, err := repo.GetPendingReminders(context.Background(), now)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pending lookup failed")
}

func TestMarkSentBatch_ExecError(t *testing.T) {
	repo, mock, cleanup := newReminderRepoMock(t)
	defer cleanup()
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)

	// pq.Array argument is opaque; substring match on the UPDATE shape.
	mock.ExpectExec(`UPDATE task_reminders SET is_sent = TRUE, sent_at = $1 WHERE id = ANY($2)`).
		WillReturnError(fmt.Errorf("conn refused"))

	err := repo.MarkSentBatch(context.Background(), []int64{1, 2, 3}, now)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mark sent batch failed")
}
