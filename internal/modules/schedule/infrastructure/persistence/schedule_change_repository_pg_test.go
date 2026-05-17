package persistence

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
)

// v0.153.1 #196 coverage push: schedule_change_repository_pg.go was at
// 0% coverage. Tests pin Create + 2 read methods + scanChanges error
// branch using sqlmock.

func newScheduleChangeRepoMock(t *testing.T) (*ScheduleChangeRepositoryPG, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewScheduleChangeRepositoryPG(db), mock
}

func sampleScheduleChange() *entities.ScheduleChange {
	now := time.Date(2026, 5, 17, 10, 0, 0, 0, time.UTC)
	return entities.NewScheduleChange(int64(42), domain.ChangeTypeMoved, now, int64(7), now)
}

// --- Create ---

func TestScheduleChangeRepoCreate_HappyPath(t *testing.T) {
	repo, mock := newScheduleChangeRepoMock(t)
	change := sampleScheduleChange()

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO schedule_changes")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(99)))

	err := repo.Create(context.Background(), change)
	require.NoError(t, err)
	assert.Equal(t, int64(99), change.ID, "Create must populate ID from RETURNING")
}

func TestScheduleChangeRepoCreate_DBError(t *testing.T) {
	repo, mock := newScheduleChangeRepoMock(t)
	change := sampleScheduleChange()

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO schedule_changes")).
		WillReturnError(errors.New("unique violation"))

	err := repo.Create(context.Background(), change)
	require.Error(t, err)
	assert.Equal(t, int64(0), change.ID, "Create must not mutate ID on error")
}

// --- GetByLessonID ---

func scheduleChangeRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "lesson_id", "change_type", "original_date", "new_date",
		"new_classroom_id", "new_teacher_id", "reason", "created_by", "created_at",
	})
}

func TestScheduleChangeRepoGetByLessonID_HappyPath(t *testing.T) {
	repo, mock := newScheduleChangeRepoMock(t)
	now := time.Now()
	reason := "Преподаватель в командировке"
	rows := scheduleChangeRows().
		AddRow(int64(1), int64(42), "canceled", now, nil, nil, nil, &reason, int64(7), now).
		AddRow(int64(2), int64(42), "moved", now, &now, nil, nil, nil, int64(7), now)

	mock.ExpectQuery(regexp.QuoteMeta("FROM schedule_changes WHERE lesson_id")).
		WithArgs(int64(42)).
		WillReturnRows(rows)

	changes, err := repo.GetByLessonID(context.Background(), int64(42))
	require.NoError(t, err)
	require.Len(t, changes, 2)
	assert.Equal(t, domain.ChangeType("canceled"), changes[0].ChangeType)
	require.NotNil(t, changes[0].Reason)
	assert.Equal(t, reason, *changes[0].Reason)
	assert.Nil(t, changes[1].Reason, "nullable reason NULL must scan to nil")
}

func TestScheduleChangeRepoGetByLessonID_QueryError(t *testing.T) {
	repo, mock := newScheduleChangeRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM schedule_changes WHERE lesson_id")).
		WithArgs(int64(42)).
		WillReturnError(errors.New("query failed"))

	changes, err := repo.GetByLessonID(context.Background(), int64(42))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get schedule changes")
	assert.Nil(t, changes)
}

func TestScheduleChangeRepoGetByLessonID_ScanError(t *testing.T) {
	repo, mock := newScheduleChangeRepoMock(t)
	// Row column count mismatch — Scan returns error.
	rows := sqlmock.NewRows([]string{"id", "lesson_id"}).AddRow(int64(1), int64(42))
	mock.ExpectQuery(regexp.QuoteMeta("FROM schedule_changes WHERE lesson_id")).
		WithArgs(int64(42)).
		WillReturnRows(rows)

	changes, err := repo.GetByLessonID(context.Background(), int64(42))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to scan schedule change")
	assert.Nil(t, changes)
}

// --- GetByDateRange ---

func TestScheduleChangeRepoGetByDateRange_HappyPath(t *testing.T) {
	repo, mock := newScheduleChangeRepoMock(t)
	start := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 5, 31, 23, 59, 59, 0, time.UTC)
	rows := scheduleChangeRows().
		AddRow(int64(1), int64(42), "canceled", start.AddDate(0, 0, 5), nil, nil, nil, nil, int64(7), start)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE original_date >= $1 AND original_date <= $2")).
		WithArgs(start, end).
		WillReturnRows(rows)

	changes, err := repo.GetByDateRange(context.Background(), start, end)
	require.NoError(t, err)
	require.Len(t, changes, 1)
	assert.Equal(t, int64(1), changes[0].ID)
}

func TestScheduleChangeRepoGetByDateRange_QueryError(t *testing.T) {
	repo, mock := newScheduleChangeRepoMock(t)
	start := time.Now()
	end := start.Add(48 * time.Hour)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE original_date >= $1 AND original_date <= $2")).
		WithArgs(start, end).
		WillReturnError(errors.New("transient db error"))

	changes, err := repo.GetByDateRange(context.Background(), start, end)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get schedule changes by date range")
	assert.Nil(t, changes)
}

// --- Constructor ---

func TestNewScheduleChangeRepositoryPG_StoresHandle(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := NewScheduleChangeRepositoryPG(db)
	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}
