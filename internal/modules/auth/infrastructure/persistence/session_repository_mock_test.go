package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"
)

func newSessionRepoMock(t *testing.T) (*SessionRepositoryPG, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return &SessionRepositoryPG{db: db}, mock
}

var sessionCols = []string{"id", "user_id", "refresh_token", "user_agent", "ip_address", "expires_at", "created_at", "updated_at"}

func TestNewSessionRepositoryPG(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer func() { _ = db.Close() }()
	repo := NewSessionRepositoryPG(db)
	assert.NotNil(t, repo)
}

// --- Create ---

func TestSessionRepo_Create_Success(t *testing.T) {
	repo, mock := newSessionRepoMock(t)
	now := time.Now()
	session := &entities.Session{
		UserID: 1, RefreshToken: "token123", UserAgent: "Mozilla/5.0",
		IPAddress: "127.0.0.1", ExpiresAt: now.Add(24 * time.Hour),
		CreatedAt: now, UpdatedAt: now,
	}
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO sessions")).
		WithArgs(session.UserID, session.RefreshToken, session.UserAgent, session.IPAddress, session.ExpiresAt, session.CreatedAt, session.UpdatedAt).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
	err := repo.Create(context.Background(), session)
	require.NoError(t, err)
	assert.Equal(t, int64(1), session.ID)
}

func TestSessionRepo_Create_Error(t *testing.T) {
	repo, mock := newSessionRepoMock(t)
	now := time.Now()
	session := &entities.Session{
		UserID: 1, RefreshToken: "token123",
		CreatedAt: now, UpdatedAt: now, ExpiresAt: now.Add(time.Hour),
	}
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO sessions")).
		WillReturnError(fmt.Errorf("db error"))
	err := repo.Create(context.Background(), session)
	assert.Error(t, err)
}

// --- GetByRefreshToken ---

func TestSessionRepo_GetByRefreshToken_Success(t *testing.T) {
	repo, mock := newSessionRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows(sessionCols).AddRow(
		int64(1), int64(1), "token123", "Mozilla/5.0", "127.0.0.1",
		now.Add(24*time.Hour), now, now,
	)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, refresh_token, user_agent, ip_address, expires_at, created_at, updated_at")).
		WithArgs("token123").WillReturnRows(rows)
	session, err := repo.GetByRefreshToken(context.Background(), "token123")
	require.NoError(t, err)
	assert.Equal(t, "token123", session.RefreshToken)
	assert.Equal(t, int64(1), session.UserID)
}

func TestSessionRepo_GetByRefreshToken_NotFound(t *testing.T) {
	repo, mock := newSessionRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, refresh_token, user_agent, ip_address, expires_at, created_at, updated_at")).
		WithArgs("nonexistent").WillReturnError(sql.ErrNoRows)
	_, err := repo.GetByRefreshToken(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestSessionRepo_GetByRefreshToken_DBError(t *testing.T) {
	repo, mock := newSessionRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, refresh_token, user_agent, ip_address, expires_at, created_at, updated_at")).
		WithArgs("token").WillReturnError(fmt.Errorf("connection error"))
	_, err := repo.GetByRefreshToken(context.Background(), "token")
	assert.Error(t, err)
}

// --- Delete ---

func TestSessionRepo_Delete_Success(t *testing.T) {
	repo, mock := newSessionRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM sessions WHERE refresh_token = $1")).
		WithArgs("token123").WillReturnResult(sqlmock.NewResult(0, 1))
	err := repo.Delete(context.Background(), "token123")
	require.NoError(t, err)
}

func TestSessionRepo_Delete_NotFound(t *testing.T) {
	repo, mock := newSessionRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM sessions WHERE refresh_token = $1")).
		WithArgs("nonexistent").WillReturnResult(sqlmock.NewResult(0, 0))
	err := repo.Delete(context.Background(), "nonexistent")
	assert.Error(t, err)
}

func TestSessionRepo_Delete_DBError(t *testing.T) {
	repo, mock := newSessionRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM sessions WHERE refresh_token = $1")).
		WithArgs("token").WillReturnError(fmt.Errorf("db error"))
	err := repo.Delete(context.Background(), "token")
	assert.Error(t, err)
}

func TestSessionRepo_Delete_RowsAffectedError(t *testing.T) {
	repo, mock := newSessionRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM sessions WHERE refresh_token = $1")).
		WithArgs("token").WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("rows error")))
	err := repo.Delete(context.Background(), "token")
	assert.Error(t, err)
}

// --- DeleteByUserID ---

func TestSessionRepo_DeleteByUserID_Success(t *testing.T) {
	repo, mock := newSessionRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM sessions WHERE user_id = $1")).
		WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 3))
	err := repo.DeleteByUserID(context.Background(), 1)
	require.NoError(t, err)
}

func TestSessionRepo_DeleteByUserID_NoRows(t *testing.T) {
	repo, mock := newSessionRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM sessions WHERE user_id = $1")).
		WithArgs(int64(999)).WillReturnResult(sqlmock.NewResult(0, 0))
	// DeleteByUserID does not return error if no rows affected
	err := repo.DeleteByUserID(context.Background(), 999)
	require.NoError(t, err)
}

func TestSessionRepo_DeleteByUserID_DBError(t *testing.T) {
	repo, mock := newSessionRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM sessions WHERE user_id = $1")).
		WithArgs(int64(1)).WillReturnError(fmt.Errorf("db error"))
	err := repo.DeleteByUserID(context.Background(), 1)
	assert.Error(t, err)
}

// --- DeleteExpired ---

func TestSessionRepo_DeleteExpired_Success(t *testing.T) {
	repo, mock := newSessionRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM sessions WHERE expires_at < $1")).
		WillReturnResult(sqlmock.NewResult(0, 5))
	err := repo.DeleteExpired(context.Background())
	require.NoError(t, err)
}

func TestSessionRepo_DeleteExpired_DBError(t *testing.T) {
	repo, mock := newSessionRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM sessions WHERE expires_at < $1")).
		WillReturnError(fmt.Errorf("db error"))
	err := repo.DeleteExpired(context.Background())
	assert.Error(t, err)
}

// --- GetActiveByUserID ---

func TestSessionRepo_GetActiveByUserID_Success(t *testing.T) {
	repo, mock := newSessionRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows(sessionCols).
		AddRow(int64(1), int64(1), "token1", "Mozilla/5.0", "127.0.0.1", now.Add(24*time.Hour), now, now).
		AddRow(int64(2), int64(1), "token2", "Chrome", "192.168.1.1", now.Add(48*time.Hour), now, now)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, refresh_token, user_agent, ip_address, expires_at, created_at, updated_at")).
		WillReturnRows(rows)
	sessions, err := repo.GetActiveByUserID(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, sessions, 2)
}

func TestSessionRepo_GetActiveByUserID_Empty(t *testing.T) {
	repo, mock := newSessionRepoMock(t)
	rows := sqlmock.NewRows(sessionCols)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, refresh_token, user_agent, ip_address, expires_at, created_at, updated_at")).
		WillReturnRows(rows)
	sessions, err := repo.GetActiveByUserID(context.Background(), 1)
	require.NoError(t, err)
	assert.Empty(t, sessions)
}

func TestSessionRepo_GetActiveByUserID_DBError(t *testing.T) {
	repo, mock := newSessionRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, refresh_token, user_agent, ip_address, expires_at, created_at, updated_at")).
		WillReturnError(fmt.Errorf("db error"))
	_, err := repo.GetActiveByUserID(context.Background(), 1)
	assert.Error(t, err)
}

func TestSessionRepo_GetActiveByUserID_ScanError(t *testing.T) {
	repo, mock := newSessionRepoMock(t)
	// Wrong column count to trigger scan error
	rows := sqlmock.NewRows([]string{"id"}).AddRow(int64(1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, refresh_token, user_agent, ip_address, expires_at, created_at, updated_at")).
		WillReturnRows(rows)
	_, err := repo.GetActiveByUserID(context.Background(), 1)
	assert.Error(t, err)
}
