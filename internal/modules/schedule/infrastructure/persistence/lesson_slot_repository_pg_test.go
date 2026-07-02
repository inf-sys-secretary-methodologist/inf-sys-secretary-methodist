package persistence

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
)

func newLessonSlotRepoMock(t *testing.T) (*LessonSlotRepositoryPG, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewLessonSlotRepositoryPG(db), mock
}

func sampleSlot(now time.Time) *entities.LessonSlot {
	return &entities.LessonSlot{Number: 1, TimeStart: "08:30", TimeEnd: "10:00", CreatedAt: now, UpdatedAt: now}
}

func TestLessonSlotRepositoryPG_Create_Success(t *testing.T) {
	repo, mock := newLessonSlotRepoMock(t)
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	slot := sampleSlot(now)

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO lesson_slots")).
		WithArgs(1, "08:30", "10:00", now, now).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(5)))

	err := repo.Create(context.Background(), slot)
	require.NoError(t, err)
	assert.Equal(t, int64(5), slot.ID, "Create must populate the generated id")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLessonSlotRepositoryPG_Create_DuplicateNumber(t *testing.T) {
	repo, mock := newLessonSlotRepoMock(t)
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	slot := sampleSlot(now)

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO lesson_slots")).
		WithArgs(1, "08:30", "10:00", now, now).
		WillReturnError(&pq.Error{Code: "23505"})

	err := repo.Create(context.Background(), slot)
	assert.ErrorIs(t, err, entities.ErrLessonSlotNumberTaken)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLessonSlotRepositoryPG_Update_Success(t *testing.T) {
	repo, mock := newLessonSlotRepoMock(t)
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	slot := sampleSlot(now)
	slot.ID = 5

	mock.ExpectExec(regexp.QuoteMeta("UPDATE lesson_slots")).
		WithArgs(1, "08:30", "10:00", now, int64(5)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), slot)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLessonSlotRepositoryPG_Update_NotFound(t *testing.T) {
	repo, mock := newLessonSlotRepoMock(t)
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	slot := sampleSlot(now)
	slot.ID = 99

	mock.ExpectExec(regexp.QuoteMeta("UPDATE lesson_slots")).
		WithArgs(1, "08:30", "10:00", now, int64(99)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Update(context.Background(), slot)
	assert.ErrorIs(t, err, entities.ErrLessonSlotNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLessonSlotRepositoryPG_Update_DuplicateNumber(t *testing.T) {
	repo, mock := newLessonSlotRepoMock(t)
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	slot := sampleSlot(now)
	slot.ID = 5

	mock.ExpectExec(regexp.QuoteMeta("UPDATE lesson_slots")).
		WithArgs(1, "08:30", "10:00", now, int64(5)).
		WillReturnError(&pq.Error{Code: "23505"})

	err := repo.Update(context.Background(), slot)
	assert.ErrorIs(t, err, entities.ErrLessonSlotNumberTaken)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLessonSlotRepositoryPG_Delete_Success(t *testing.T) {
	repo, mock := newLessonSlotRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM lesson_slots WHERE id = $1")).
		WithArgs(int64(5)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), 5)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLessonSlotRepositoryPG_Delete_NotFound(t *testing.T) {
	repo, mock := newLessonSlotRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM lesson_slots WHERE id = $1")).
		WithArgs(int64(99)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Delete(context.Background(), 99)
	assert.ErrorIs(t, err, entities.ErrLessonSlotNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLessonSlotRepositoryPG_GetByID_Found(t *testing.T) {
	repo, mock := newLessonSlotRepoMock(t)
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	rows := sqlmock.NewRows([]string{"id", "number", "time_start", "time_end", "created_at", "updated_at"}).
		AddRow(int64(5), 1, "08:30", "10:00", now, now)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE id = $1")).
		WithArgs(int64(5)).
		WillReturnRows(rows)

	got, err := repo.GetByID(context.Background(), 5)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, int64(5), got.ID)
	assert.Equal(t, 1, got.Number)
	assert.Equal(t, "08:30", got.TimeStart)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLessonSlotRepositoryPG_GetByID_NotFound(t *testing.T) {
	repo, mock := newLessonSlotRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE id = $1")).
		WithArgs(int64(99)).
		WillReturnError(sql.ErrNoRows)

	got, err := repo.GetByID(context.Background(), 99)
	assert.Nil(t, got)
	assert.ErrorIs(t, err, entities.ErrLessonSlotNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLessonSlotRepositoryPG_List_OrderedByNumber(t *testing.T) {
	repo, mock := newLessonSlotRepoMock(t)
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	rows := sqlmock.NewRows([]string{"id", "number", "time_start", "time_end", "created_at", "updated_at"}).
		AddRow(int64(1), 1, "08:30", "10:00", now, now).
		AddRow(int64(2), 2, "10:10", "11:40", now, now)
	mock.ExpectQuery(regexp.QuoteMeta("ORDER BY number")).
		WillReturnRows(rows)

	got, err := repo.List(context.Background())
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, 1, got[0].Number)
	assert.Equal(t, 2, got[1].Number)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLessonSlotRepositoryPG_List_DBError(t *testing.T) {
	repo, mock := newLessonSlotRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("ORDER BY number")).
		WillReturnError(errors.New("connection reset"))

	got, err := repo.List(context.Background())
	assert.Nil(t, got)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
