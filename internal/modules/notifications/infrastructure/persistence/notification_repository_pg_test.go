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

func newNotifRepoMock(t *testing.T) (*NotificationRepositoryPG, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return &NotificationRepositoryPG{db: db}, mock
}

func TestNewNotificationRepositoryPG(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer func() { _ = db.Close() }()
	repo := NewNotificationRepositoryPG(db)
	assert.NotNil(t, repo)
}

func TestNullString(t *testing.T) {
	ns := nullString("")
	assert.False(t, ns.Valid)
	ns = nullString("hello")
	assert.True(t, ns.Valid)
	assert.Equal(t, "hello", ns.String)
}

func TestNotifRepo_Create_Success(t *testing.T) {
	repo, mock := newNotifRepoMock(t)
	n := &entities.Notification{
		UserID: 1, Type: "info", Priority: "normal",
		Title: "Test", Message: "msg",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO notifications")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
	require.NoError(t, repo.Create(context.Background(), n))
	assert.Equal(t, int64(1), n.ID)
}

func TestNotifRepo_Create_Error(t *testing.T) {
	repo, mock := newNotifRepoMock(t)
	n := &entities.Notification{UserID: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO notifications")).WillReturnError(fmt.Errorf("db"))
	assert.Error(t, repo.Create(context.Background(), n))
}

func TestNotifRepo_Update_Success(t *testing.T) {
	repo, mock := newNotifRepoMock(t)
	n := &entities.Notification{ID: 1, Type: "info", Priority: "normal", Title: "Up"}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE notifications SET")).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.Update(context.Background(), n))
}

func TestNotifRepo_Update_NotFound(t *testing.T) {
	repo, mock := newNotifRepoMock(t)
	n := &entities.Notification{ID: 999}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE notifications SET")).WillReturnResult(sqlmock.NewResult(0, 0))
	assert.Error(t, repo.Update(context.Background(), n))
}

func TestNotifRepo_Delete_Success(t *testing.T) {
	repo, mock := newNotifRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM notifications WHERE id")).
		WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.Delete(context.Background(), 1))
}

func TestNotifRepo_Delete_NotFound(t *testing.T) {
	repo, mock := newNotifRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM notifications WHERE id")).
		WithArgs(int64(999)).WillReturnResult(sqlmock.NewResult(0, 0))
	assert.Error(t, repo.Delete(context.Background(), 999))
}

var notifCols = []string{
	"id", "user_id", "type", "priority", "title", "message", "link", "image_url",
	"is_read", "read_at", "expires_at", "metadata", "created_at", "updated_at",
}

func TestNotifRepo_GetByID_Success(t *testing.T) {
	repo, mock := newNotifRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows(notifCols).AddRow(
		1, int64(1), "info", "normal", "Test", "msg",
		sql.NullString{}, sql.NullString{},
		false, nil, nil, nil, now, now,
	)
	mock.ExpectQuery(regexp.QuoteMeta("FROM notifications")).WithArgs(int64(1)).WillReturnRows(rows)
	n, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.NotNil(t, n)
	assert.Equal(t, "Test", n.Title)
}

func TestNotifRepo_GetByID_NotFound(t *testing.T) {
	repo, mock := newNotifRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM notifications")).WithArgs(int64(999)).WillReturnError(sql.ErrNoRows)
	n, err := repo.GetByID(context.Background(), 999)
	require.NoError(t, err)
	assert.Nil(t, n)
}

func TestNotifRepo_List(t *testing.T) {
	repo, mock := newNotifRepoMock(t)
	isRead := false
	filter := &entities.NotificationFilter{
		UserID: 1, Type: "info", Priority: "normal",
		IsRead: &isRead, Limit: 10, Offset: 0,
	}
	rows := sqlmock.NewRows(notifCols)
	mock.ExpectQuery(regexp.QuoteMeta("FROM notifications")).WillReturnRows(rows)
	notifs, err := repo.List(context.Background(), filter)
	require.NoError(t, err)
	assert.Empty(t, notifs)
}

func TestNotifRepo_List_NoFilters(t *testing.T) {
	repo, mock := newNotifRepoMock(t)
	filter := &entities.NotificationFilter{Limit: 0}
	rows := sqlmock.NewRows(notifCols)
	mock.ExpectQuery(regexp.QuoteMeta("FROM notifications")).WillReturnRows(rows)
	_, err := repo.List(context.Background(), filter)
	require.NoError(t, err)
}

func TestNotifRepo_GetByUserID(t *testing.T) {
	repo, mock := newNotifRepoMock(t)
	rows := sqlmock.NewRows(notifCols)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE user_id = $1")).WithArgs(int64(1), 50, 0).WillReturnRows(rows)
	notifs, err := repo.GetByUserID(context.Background(), 1, 0, 0)
	require.NoError(t, err)
	assert.Empty(t, notifs)
}

func TestNotifRepo_GetUnreadByUserID(t *testing.T) {
	repo, mock := newNotifRepoMock(t)
	rows := sqlmock.NewRows(notifCols)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE user_id = $1 AND is_read = false")).WithArgs(int64(1)).WillReturnRows(rows)
	notifs, err := repo.GetUnreadByUserID(context.Background(), 1)
	require.NoError(t, err)
	assert.Empty(t, notifs)
}

func TestNotifRepo_MarkAsRead_Success(t *testing.T) {
	repo, mock := newNotifRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE notifications SET is_read = true")).
		WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.MarkAsRead(context.Background(), 1))
}

func TestNotifRepo_MarkAsRead_NotFound(t *testing.T) {
	repo, mock := newNotifRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE notifications SET is_read = true")).
		WithArgs(int64(999)).WillReturnResult(sqlmock.NewResult(0, 0))
	assert.Error(t, repo.MarkAsRead(context.Background(), 999))
}

func TestNotifRepo_MarkAllAsRead(t *testing.T) {
	repo, mock := newNotifRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE notifications SET is_read = true")).
		WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 5))
	require.NoError(t, repo.MarkAllAsRead(context.Background(), 1))
}

func TestNotifRepo_DeleteByUserID(t *testing.T) {
	repo, mock := newNotifRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM notifications WHERE user_id")).
		WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 3))
	require.NoError(t, repo.DeleteByUserID(context.Background(), 1))
}

func TestNotifRepo_DeleteExpired(t *testing.T) {
	repo, mock := newNotifRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM notifications WHERE expires_at")).
		WillReturnResult(sqlmock.NewResult(0, 2))
	count, err := repo.DeleteExpired(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestNotifRepo_GetUnreadCount(t *testing.T) {
	repo, mock := newNotifRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WithArgs(int64(1)).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(5)))
	count, err := repo.GetUnreadCount(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(5), count)
}

func TestNotifRepo_GetStats(t *testing.T) {
	repo, mock := newNotifRepoMock(t)
	rows := sqlmock.NewRows([]string{"total", "unread", "today", "urgent", "expired"}).
		AddRow(int64(100), int64(10), int64(5), int64(2), int64(1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WithArgs(int64(1)).WillReturnRows(rows)
	stats, err := repo.GetStats(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(100), stats.TotalCount)
	assert.Equal(t, int64(10), stats.UnreadCount)
}

func TestNotifRepo_CreateBulk_Empty(t *testing.T) {
	repo, _ := newNotifRepoMock(t)
	require.NoError(t, repo.CreateBulk(context.Background(), nil))
}

func TestNotifRepo_CreateBulk_Success(t *testing.T) {
	repo, mock := newNotifRepoMock(t)
	now := time.Now()
	notifs := []*entities.Notification{
		{UserID: 1, Type: "info", Priority: "normal", Title: "T1", Message: "M1", CreatedAt: now, UpdatedAt: now},
	}
	mock.ExpectBegin()
	mock.ExpectPrepare(regexp.QuoteMeta("INSERT INTO notifications"))
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO notifications")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
	mock.ExpectCommit()
	require.NoError(t, repo.CreateBulk(context.Background(), notifs))
}

func TestNotifRepo_CreateBulk_BeginError(t *testing.T) {
	repo, mock := newNotifRepoMock(t)
	mock.ExpectBegin().WillReturnError(fmt.Errorf("tx error"))
	assert.Error(t, repo.CreateBulk(context.Background(), []*entities.Notification{{}}))
}

func TestNotifRepo_CreateBulk_PrepareError(t *testing.T) {
	repo, mock := newNotifRepoMock(t)
	mock.ExpectBegin()
	mock.ExpectPrepare(regexp.QuoteMeta("INSERT INTO notifications")).WillReturnError(fmt.Errorf("prepare error"))
	mock.ExpectRollback()
	assert.Error(t, repo.CreateBulk(context.Background(), []*entities.Notification{{}}))
}

func TestNotifRepo_Update_Error(t *testing.T) {
	repo, mock := newNotifRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE notifications SET")).WillReturnError(fmt.Errorf("db"))
	assert.Error(t, repo.Update(context.Background(), &entities.Notification{ID: 1}))
}

func TestNotifRepo_Delete_Error(t *testing.T) {
	repo, mock := newNotifRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM notifications")).WillReturnError(fmt.Errorf("db"))
	assert.Error(t, repo.Delete(context.Background(), 1))
}

func TestNotifRepo_MarkAsRead_Error(t *testing.T) {
	repo, mock := newNotifRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE notifications SET is_read")).WillReturnError(fmt.Errorf("db"))
	assert.Error(t, repo.MarkAsRead(context.Background(), 1))
}

func TestNotifRepo_MarkAllAsRead_Error(t *testing.T) {
	repo, mock := newNotifRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE notifications SET is_read")).WillReturnError(fmt.Errorf("db"))
	assert.Error(t, repo.MarkAllAsRead(context.Background(), 1))
}

func TestNotifRepo_DeleteByUserID_Error(t *testing.T) {
	repo, mock := newNotifRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM notifications WHERE user_id")).WillReturnError(fmt.Errorf("db"))
	assert.Error(t, repo.DeleteByUserID(context.Background(), 1))
}

func TestNotifRepo_DeleteExpired_Error(t *testing.T) {
	repo, mock := newNotifRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM notifications WHERE expires_at")).WillReturnError(fmt.Errorf("db"))
	_, err := repo.DeleteExpired(context.Background())
	assert.Error(t, err)
}

func TestNotifRepo_GetUnreadCount_Error(t *testing.T) {
	repo, mock := newNotifRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).WillReturnError(fmt.Errorf("db"))
	_, err := repo.GetUnreadCount(context.Background(), 1)
	assert.Error(t, err)
}

func TestNotifRepo_GetStats_Error(t *testing.T) {
	repo, mock := newNotifRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("COUNT(*)")).WillReturnError(fmt.Errorf("db"))
	_, err := repo.GetStats(context.Background(), 1)
	assert.Error(t, err)
}

func TestNotifRepo_GetByID_Error(t *testing.T) {
	repo, mock := newNotifRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM notifications WHERE id")).WillReturnError(fmt.Errorf("db"))
	_, err := repo.GetByID(context.Background(), 1)
	assert.Error(t, err)
}
