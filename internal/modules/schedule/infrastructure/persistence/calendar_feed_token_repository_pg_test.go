package persistence

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
)

func newFeedTokenRepoMock(t *testing.T) (*CalendarFeedTokenRepositoryPG, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewCalendarFeedTokenRepositoryPG(db), mock
}

func TestCalendarFeedTokenRepositoryPG_Save_Upsert(t *testing.T) {
	repo, mock := newFeedTokenRepoMock(t)
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	tok := &entities.CalendarFeedToken{UserID: 42, Token: "deadbeef", CreatedAt: now}

	// The upsert/rotation contract is the headline behavior of this slice:
	// require the ON CONFLICT (user_id) DO UPDATE and RETURNING id clauses so
	// a regression to a plain INSERT (which would fail on the second call with
	// a unique violation) is caught.
	mock.ExpectQuery(`INSERT INTO calendar_feed_tokens[\s\S]+ON CONFLICT \(user_id\) DO UPDATE[\s\S]+RETURNING id`).
		WithArgs(int64(42), "deadbeef", now).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(7)))

	err := repo.Save(context.Background(), tok)
	require.NoError(t, err)
	assert.Equal(t, int64(7), tok.ID, "Save must populate the generated id")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCalendarFeedTokenRepositoryPG_Save_DBError(t *testing.T) {
	repo, mock := newFeedTokenRepoMock(t)
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	tok := &entities.CalendarFeedToken{UserID: 42, Token: "deadbeef", CreatedAt: now}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO calendar_feed_tokens")).
		WithArgs(int64(42), "deadbeef", now).
		WillReturnError(errors.New("connection reset"))

	err := repo.Save(context.Background(), tok)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCalendarFeedTokenRepositoryPG_GetByUserID_Found(t *testing.T) {
	repo, mock := newFeedTokenRepoMock(t)
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	rows := sqlmock.NewRows([]string{"id", "user_id", "token", "created_at"}).
		AddRow(int64(7), int64(42), "deadbeef", now)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE user_id = $1")).
		WithArgs(int64(42)).
		WillReturnRows(rows)

	got, err := repo.GetByUserID(context.Background(), 42)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, int64(7), got.ID)
	assert.Equal(t, int64(42), got.UserID)
	assert.Equal(t, "deadbeef", got.Token)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCalendarFeedTokenRepositoryPG_GetByUserID_NotFound(t *testing.T) {
	repo, mock := newFeedTokenRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE user_id = $1")).
		WithArgs(int64(99)).
		WillReturnError(sql.ErrNoRows)

	got, err := repo.GetByUserID(context.Background(), 99)
	assert.Nil(t, got)
	assert.ErrorIs(t, err, entities.ErrCalendarFeedTokenNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCalendarFeedTokenRepositoryPG_GetByToken_Found(t *testing.T) {
	repo, mock := newFeedTokenRepoMock(t)
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	rows := sqlmock.NewRows([]string{"id", "user_id", "token", "created_at"}).
		AddRow(int64(7), int64(42), "deadbeef", now)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE token = $1")).
		WithArgs("deadbeef").
		WillReturnRows(rows)

	got, err := repo.GetByToken(context.Background(), "deadbeef")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, int64(42), got.UserID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCalendarFeedTokenRepositoryPG_GetByToken_NotFound(t *testing.T) {
	repo, mock := newFeedTokenRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE token = $1")).
		WithArgs("missing").
		WillReturnError(sql.ErrNoRows)

	got, err := repo.GetByToken(context.Background(), "missing")
	assert.Nil(t, got)
	assert.ErrorIs(t, err, entities.ErrCalendarFeedTokenNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCalendarFeedTokenRepositoryPG_DeleteByUserID(t *testing.T) {
	repo, mock := newFeedTokenRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM calendar_feed_tokens WHERE user_id = $1")).
		WithArgs(int64(42)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.DeleteByUserID(context.Background(), 42)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCalendarFeedTokenRepositoryPG_DeleteByUserID_MissingIsNoOp(t *testing.T) {
	repo, mock := newFeedTokenRepoMock(t)

	// Zero rows affected: deleting a non-existent token is not an error.
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM calendar_feed_tokens WHERE user_id = $1")).
		WithArgs(int64(7)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.DeleteByUserID(context.Background(), 7)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCalendarFeedTokenRepositoryPG_DeleteByUserID_DBError(t *testing.T) {
	repo, mock := newFeedTokenRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM calendar_feed_tokens WHERE user_id = $1")).
		WithArgs(int64(42)).
		WillReturnError(errors.New("connection reset"))

	err := repo.DeleteByUserID(context.Background(), 42)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
