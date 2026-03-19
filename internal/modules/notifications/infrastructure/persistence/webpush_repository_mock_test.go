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

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/entities"
)

func newWPRepoMock(t *testing.T) (*WebPushRepositoryPG, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return &WebPushRepositoryPG{db: db}, mock
}

var wpCols = []string{
	"id", "user_id", "endpoint", "p256dh_key", "auth_key",
	"user_agent", "device_name", "is_active", "last_used_at",
	"created_at", "updated_at",
}

func TestNewWebPushRepositoryPG_Mock(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer func() { _ = db.Close() }()
	repo := NewWebPushRepositoryPG(db)
	assert.NotNil(t, repo)
}

// --- Create ---

func TestWPRepo_Create_Success(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	now := time.Now()
	sub := &entities.WebPushSubscription{
		UserID: 1, Endpoint: "https://push.example.com/sub1",
		P256dhKey: "key1", AuthKey: "auth1",
		UserAgent: "Mozilla", DeviceName: "Test",
		IsActive: true, CreatedAt: now, UpdatedAt: now,
	}
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO webpush_subscriptions")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
	err := repo.Create(context.Background(), sub)
	require.NoError(t, err)
	assert.Equal(t, int64(1), sub.ID)
}

func TestWPRepo_Create_Error(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	sub := &entities.WebPushSubscription{
		UserID: 1, Endpoint: "https://push.example.com/sub1",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO webpush_subscriptions")).
		WillReturnError(fmt.Errorf("db err"))
	err := repo.Create(context.Background(), sub)
	assert.Error(t, err)
}

// --- GetByID ---

func TestWPRepo_GetByID_Success(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows(wpCols).AddRow(
		int64(1), int64(1), "https://push.example.com", "key", "auth",
		sql.NullString{String: "Mozilla", Valid: true}, sql.NullString{String: "Device", Valid: true},
		true, sql.NullTime{Time: now, Valid: true}, now, now,
	)
	mock.ExpectQuery(regexp.QuoteMeta("FROM webpush_subscriptions")).
		WithArgs(int64(1)).WillReturnRows(rows)
	sub, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.NotNil(t, sub)
	assert.Equal(t, "Mozilla", sub.UserAgent)
	assert.NotNil(t, sub.LastUsedAt)
}

func TestWPRepo_GetByID_NotFound(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM webpush_subscriptions")).
		WithArgs(int64(999)).WillReturnError(sql.ErrNoRows)
	sub, err := repo.GetByID(context.Background(), 999)
	require.NoError(t, err)
	assert.Nil(t, sub)
}

func TestWPRepo_GetByID_DBError(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM webpush_subscriptions")).
		WithArgs(int64(1)).WillReturnError(fmt.Errorf("conn"))
	_, err := repo.GetByID(context.Background(), 1)
	assert.Error(t, err)
}

func TestWPRepo_GetByID_NoLastUsed(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows(wpCols).AddRow(
		int64(1), int64(1), "https://push.example.com", "key", "auth",
		sql.NullString{}, sql.NullString{},
		true, sql.NullTime{}, now, now,
	)
	mock.ExpectQuery(regexp.QuoteMeta("FROM webpush_subscriptions")).
		WithArgs(int64(1)).WillReturnRows(rows)
	sub, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Nil(t, sub.LastUsedAt)
}

// --- GetByEndpoint ---

func TestWPRepo_GetByEndpoint_Success(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows(wpCols).AddRow(
		int64(1), int64(1), "https://push.example.com", "key", "auth",
		sql.NullString{}, sql.NullString{}, true, sql.NullTime{}, now, now,
	)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE endpoint = $1")).
		WithArgs("https://push.example.com").WillReturnRows(rows)
	sub, err := repo.GetByEndpoint(context.Background(), "https://push.example.com")
	require.NoError(t, err)
	assert.NotNil(t, sub)
}

func TestWPRepo_GetByEndpoint_NotFound(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE endpoint = $1")).
		WithArgs("nonexistent").WillReturnError(sql.ErrNoRows)
	sub, err := repo.GetByEndpoint(context.Background(), "nonexistent")
	require.NoError(t, err)
	assert.Nil(t, sub)
}

func TestWPRepo_GetByEndpoint_DBError(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE endpoint = $1")).
		WithArgs("ep").WillReturnError(fmt.Errorf("conn"))
	_, err := repo.GetByEndpoint(context.Background(), "ep")
	assert.Error(t, err)
}

// --- GetByUserID ---

func TestWPRepo_GetByUserID_Success(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows(wpCols).
		AddRow(int64(1), int64(1), "ep1", "k1", "a1", sql.NullString{}, sql.NullString{}, true, sql.NullTime{}, now, now).
		AddRow(int64(2), int64(1), "ep2", "k2", "a2", sql.NullString{}, sql.NullString{}, true, sql.NullTime{}, now, now)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE user_id = $1")).
		WithArgs(int64(1)).WillReturnRows(rows)
	subs, err := repo.GetByUserID(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, subs, 2)
}

func TestWPRepo_GetByUserID_DBError(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE user_id = $1")).
		WithArgs(int64(1)).WillReturnError(fmt.Errorf("err"))
	_, err := repo.GetByUserID(context.Background(), 1)
	assert.Error(t, err)
}

// --- GetActiveByUserID ---

func TestWPRepo_GetActiveByUserID_Success(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows(wpCols).
		AddRow(int64(1), int64(1), "ep1", "k1", "a1", sql.NullString{}, sql.NullString{}, true, sql.NullTime{}, now, now)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE user_id = $1 AND is_active = true")).
		WithArgs(int64(1)).WillReturnRows(rows)
	subs, err := repo.GetActiveByUserID(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, subs, 1)
}

func TestWPRepo_GetActiveByUserID_DBError(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE user_id = $1 AND is_active = true")).
		WithArgs(int64(1)).WillReturnError(fmt.Errorf("err"))
	_, err := repo.GetActiveByUserID(context.Background(), 1)
	assert.Error(t, err)
}

// --- Update ---

func TestWPRepo_Update_Success(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	sub := &entities.WebPushSubscription{ID: 1, P256dhKey: "newkey", AuthKey: "newauth", IsActive: true}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE webpush_subscriptions SET")).
		WillReturnResult(sqlmock.NewResult(0, 1))
	err := repo.Update(context.Background(), sub)
	require.NoError(t, err)
}

func TestWPRepo_Update_NotFound(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	sub := &entities.WebPushSubscription{ID: 999}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE webpush_subscriptions SET")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	err := repo.Update(context.Background(), sub)
	assert.Error(t, err)
}

func TestWPRepo_Update_DBError(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	sub := &entities.WebPushSubscription{ID: 1}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE webpush_subscriptions SET")).
		WillReturnError(fmt.Errorf("db err"))
	err := repo.Update(context.Background(), sub)
	assert.Error(t, err)
}

// --- UpdateLastUsed ---

func TestWPRepo_UpdateLastUsed_Success(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE webpush_subscriptions SET last_used_at")).
		WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	err := repo.UpdateLastUsed(context.Background(), 1)
	require.NoError(t, err)
}

func TestWPRepo_UpdateLastUsed_DBError(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE webpush_subscriptions SET last_used_at")).
		WithArgs(int64(1)).WillReturnError(fmt.Errorf("err"))
	err := repo.UpdateLastUsed(context.Background(), 1)
	assert.Error(t, err)
}

// --- Deactivate ---

func TestWPRepo_Deactivate_Success(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE webpush_subscriptions SET is_active = false")).
		WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	err := repo.Deactivate(context.Background(), 1)
	require.NoError(t, err)
}

func TestWPRepo_Deactivate_NotFound(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE webpush_subscriptions SET is_active = false")).
		WithArgs(int64(999)).WillReturnResult(sqlmock.NewResult(0, 0))
	err := repo.Deactivate(context.Background(), 999)
	assert.Error(t, err)
}

func TestWPRepo_Deactivate_DBError(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE webpush_subscriptions SET is_active = false")).
		WithArgs(int64(1)).WillReturnError(fmt.Errorf("err"))
	err := repo.Deactivate(context.Background(), 1)
	assert.Error(t, err)
}

// --- Delete ---

func TestWPRepo_Delete_Success(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM webpush_subscriptions WHERE id = $1")).
		WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	err := repo.Delete(context.Background(), 1)
	require.NoError(t, err)
}

func TestWPRepo_Delete_NotFound(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM webpush_subscriptions WHERE id = $1")).
		WithArgs(int64(999)).WillReturnResult(sqlmock.NewResult(0, 0))
	err := repo.Delete(context.Background(), 999)
	assert.Error(t, err)
}

func TestWPRepo_Delete_DBError(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM webpush_subscriptions WHERE id = $1")).
		WithArgs(int64(1)).WillReturnError(fmt.Errorf("err"))
	err := repo.Delete(context.Background(), 1)
	assert.Error(t, err)
}

// --- DeleteByEndpoint ---

func TestWPRepo_DeleteByEndpoint_Success(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM webpush_subscriptions WHERE endpoint = $1")).
		WithArgs("ep1").WillReturnResult(sqlmock.NewResult(0, 1))
	err := repo.DeleteByEndpoint(context.Background(), "ep1")
	require.NoError(t, err)
}

func TestWPRepo_DeleteByEndpoint_DBError(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM webpush_subscriptions WHERE endpoint = $1")).
		WithArgs("ep1").WillReturnError(fmt.Errorf("err"))
	err := repo.DeleteByEndpoint(context.Background(), "ep1")
	assert.Error(t, err)
}

// --- DeleteByUserID ---

func TestWPRepo_DeleteByUserID_Success(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM webpush_subscriptions WHERE user_id = $1")).
		WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 2))
	err := repo.DeleteByUserID(context.Background(), 1)
	require.NoError(t, err)
}

func TestWPRepo_DeleteByUserID_DBError(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM webpush_subscriptions WHERE user_id = $1")).
		WithArgs(int64(1)).WillReturnError(fmt.Errorf("err"))
	err := repo.DeleteByUserID(context.Background(), 1)
	assert.Error(t, err)
}

// --- CountByUserID ---

func TestWPRepo_CountByUserID_Success(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(3)))
	count, err := repo.CountByUserID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestWPRepo_CountByUserID_DBError(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WithArgs(int64(1)).WillReturnError(fmt.Errorf("err"))
	_, err := repo.CountByUserID(context.Background(), 1)
	assert.Error(t, err)
}

// --- scanSubscriptions with data ---

func TestWPRepo_GetByUserID_WithLastUsedAt(t *testing.T) {
	repo, mock := newWPRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows(wpCols).AddRow(
		int64(1), int64(1), "ep", "k", "a",
		sql.NullString{String: "UA", Valid: true}, sql.NullString{String: "Dev", Valid: true},
		true, sql.NullTime{Time: now, Valid: true}, now, now,
	)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE user_id = $1")).
		WithArgs(int64(1)).WillReturnRows(rows)
	subs, err := repo.GetByUserID(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, subs, 1)
	assert.Equal(t, "UA", subs[0].UserAgent)
	assert.Equal(t, "Dev", subs[0].DeviceName)
	assert.NotNil(t, subs[0].LastUsedAt)
}
