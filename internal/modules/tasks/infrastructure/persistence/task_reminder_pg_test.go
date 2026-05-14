package persistence

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/entities"
)

func newReminderRepoMock(t *testing.T) (*TaskReminderRepositoryPG, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)
	repo := NewTaskReminderRepositoryPG(db)
	return repo, mock, func() { _ = db.Close() }
}

// TestCreate_InsertsAndAssignsID pins the INSERT shape — all 5
// user-supplied columns are bound via WithArgs (mutation-resistance
// per feedback_sqlmock_withargs_for_mutation_resistance). The
// RETURNING id flows back to the entity through hydration.
func TestCreate_InsertsAndAssignsID(t *testing.T) {
	repo, mock, cleanup := newReminderRepoMock(t)
	defer cleanup()

	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	reminder, err := entities.NewTaskReminder(42, 7, entities.ReminderTypeTelegram, 15, now)
	require.NoError(t, err)

	rows := sqlmock.NewRows([]string{"id", "created_at"}).AddRow(int64(101), now)
	mock.ExpectQuery(`INSERT INTO task_reminders (task_id, user_id, reminder_type, minutes_before, is_sent) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`).
		WithArgs(int64(42), int64(7), "telegram", 15, false).
		WillReturnRows(rows)

	require.NoError(t, repo.Create(context.Background(), reminder))
	assert.Equal(t, int64(101), reminder.ID(), "INSERT RETURNING id must hydrate back onto the entity")
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestDelete_HappyPath verifies the delete-by-id DELETE shape and
// that a successful RowsAffected==1 returns nil.
func TestDelete_HappyPath(t *testing.T) {
	repo, mock, cleanup := newReminderRepoMock(t)
	defer cleanup()

	mock.ExpectExec(`DELETE FROM task_reminders WHERE id = $1`).
		WithArgs(int64(101)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.Delete(context.Background(), 101))
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestDelete_NotFound surfaces ErrTaskReminderNotFound when no row
// matches — handler maps it к 404.
func TestDelete_NotFound(t *testing.T) {
	repo, mock, cleanup := newReminderRepoMock(t)
	defer cleanup()

	mock.ExpectExec(`DELETE FROM task_reminders WHERE id = $1`).
		WithArgs(int64(999)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Delete(context.Background(), 999)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrTaskReminderNotFound),
		"missing row → ErrTaskReminderNotFound, got %v", err)
}

// TestGetByID_ReturnsHydratedRow pins SELECT column order via
// HydrateFromPersistence (every column flows to its private field).
func TestGetByID_ReturnsHydratedRow(t *testing.T) {
	repo, mock, cleanup := newReminderRepoMock(t)
	defer cleanup()

	created := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{"id", "task_id", "user_id", "reminder_type", "minutes_before", "is_sent", "sent_at", "created_at"}).
		AddRow(int64(101), int64(42), int64(7), "telegram", 15, false, nil, created)
	mock.ExpectQuery(`SELECT id, task_id, user_id, reminder_type, minutes_before, is_sent, sent_at, created_at FROM task_reminders WHERE id = $1`).
		WithArgs(int64(101)).
		WillReturnRows(rows)

	r, err := repo.GetByID(context.Background(), 101)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, int64(101), r.ID())
	assert.Equal(t, int64(42), r.TaskID())
	assert.Equal(t, int64(7), r.UserID())
	assert.Equal(t, entities.ReminderTypeTelegram, r.ReminderType())
	assert.Equal(t, 15, r.MinutesBefore())
	assert.False(t, r.IsSent())
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestGetByID_NotFound — sql.ErrNoRows → ErrTaskReminderNotFound.
func TestGetByID_NotFound(t *testing.T) {
	repo, mock, cleanup := newReminderRepoMock(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, task_id, user_id, reminder_type, minutes_before, is_sent, sent_at, created_at FROM task_reminders WHERE id = $1`).
		WithArgs(int64(999)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "task_id", "user_id", "reminder_type", "minutes_before", "is_sent", "sent_at", "created_at"}))

	_, err := repo.GetByID(context.Background(), 999)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrTaskReminderNotFound),
		"missing row → ErrTaskReminderNotFound, got %v", err)
}

// TestListByTaskAndUser_ReturnsRows pins the SELECT shape +
// filtering by composite key. Two rows returned, ordered by
// created_at ASC.
func TestListByTaskAndUser_ReturnsRows(t *testing.T) {
	repo, mock, cleanup := newReminderRepoMock(t)
	defer cleanup()

	created1 := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	created2 := created1.Add(time.Minute)
	rows := sqlmock.NewRows([]string{"id", "task_id", "user_id", "reminder_type", "minutes_before", "is_sent", "sent_at", "created_at"}).
		AddRow(int64(101), int64(42), int64(7), "telegram", 15, false, nil, created1).
		AddRow(int64(102), int64(42), int64(7), "email", 60, false, nil, created2)
	mock.ExpectQuery(`SELECT id, task_id, user_id, reminder_type, minutes_before, is_sent, sent_at, created_at FROM task_reminders WHERE task_id = $1 AND user_id = $2 ORDER BY created_at ASC`).
		WithArgs(int64(42), int64(7)).
		WillReturnRows(rows)

	out, err := repo.ListByTaskAndUser(context.Background(), 42, 7)
	require.NoError(t, err)
	require.Len(t, out, 2)
	assert.Equal(t, entities.ReminderTypeTelegram, out[0].ReminderType())
	assert.Equal(t, entities.ReminderTypeEmail, out[1].ReminderType())
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestListByTaskAndUser_Empty returns an empty (non-nil) slice on
// no matches — handler renders [] not null.
func TestListByTaskAndUser_Empty(t *testing.T) {
	repo, mock, cleanup := newReminderRepoMock(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "task_id", "user_id", "reminder_type", "minutes_before", "is_sent", "sent_at", "created_at"})
	mock.ExpectQuery(`SELECT id, task_id, user_id, reminder_type, minutes_before, is_sent, sent_at, created_at FROM task_reminders WHERE task_id = $1 AND user_id = $2 ORDER BY created_at ASC`).
		WithArgs(int64(42), int64(7)).
		WillReturnRows(rows)

	out, err := repo.ListByTaskAndUser(context.Background(), 42, 7)
	require.NoError(t, err)
	require.NotNil(t, out, "empty slice not nil")
	assert.Len(t, out, 0)
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestGetPendingReminders_JoinsTasksAndFiltersByTriggerTime pins
// the scheduler's read query. The JOIN + INTERVAL math lives at
// SQL level — caller never reasons about timestamp arithmetic.
func TestGetPendingReminders_JoinsTasksAndFiltersByTriggerTime(t *testing.T) {
	repo, mock, cleanup := newReminderRepoMock(t)
	defer cleanup()

	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	created := now.Add(-time.Hour)
	rows := sqlmock.NewRows([]string{"id", "task_id", "user_id", "reminder_type", "minutes_before", "is_sent", "sent_at", "created_at"}).
		AddRow(int64(101), int64(42), int64(7), "telegram", 15, false, nil, created)
	mock.ExpectQuery(`SELECT r.id, r.task_id, r.user_id, r.reminder_type, r.minutes_before, r.is_sent, r.sent_at, r.created_at FROM task_reminders r JOIN tasks t ON r.task_id = t.id WHERE r.is_sent = FALSE AND t.due_date IS NOT NULL AND t.due_date - (r.minutes_before * INTERVAL '1 minute') <= $1 ORDER BY r.id ASC LIMIT 100`).
		WithArgs(now).
		WillReturnRows(rows)

	pending, err := repo.GetPendingReminders(context.Background(), now)
	require.NoError(t, err)
	require.Len(t, pending, 1)
	assert.Equal(t, int64(101), pending[0].ID())
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestMarkSentBatch_BulkUpdate pins the bulk dispatch flip.
// is_sent → true и sent_at → now for every id in the slice.
func TestMarkSentBatch_BulkUpdate(t *testing.T) {
	repo, mock, cleanup := newReminderRepoMock(t)
	defer cleanup()

	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	mock.ExpectExec(`UPDATE task_reminders SET is_sent = TRUE, sent_at = $1 WHERE id = ANY($2)`).
		WithArgs(now, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 2))

	require.NoError(t, repo.MarkSentBatch(context.Background(), []int64{101, 102}, now))
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestMarkSentBatch_Empty is a no-op — no SQL emitted, no error.
func TestMarkSentBatch_Empty(t *testing.T) {
	repo, _, cleanup := newReminderRepoMock(t)
	defer cleanup()

	require.NoError(t, repo.MarkSentBatch(context.Background(), nil, time.Now()))
}
