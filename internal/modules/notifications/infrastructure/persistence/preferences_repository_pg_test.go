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

func newPrefsRepoMock(t *testing.T) (*PreferencesRepositoryPG, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return &PreferencesRepositoryPG{db: db}, mock
}

func TestNewPreferencesRepositoryPG(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer func() { _ = db.Close() }()
	assert.NotNil(t, NewPreferencesRepositoryPG(db))
}

func TestPrefsRepo_Create_Success(t *testing.T) {
	repo, mock := newPrefsRepoMock(t)
	p := entities.NewUserNotificationPreferences(1)
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO notification_preferences")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
	require.NoError(t, repo.Create(context.Background(), p))
	assert.Equal(t, int64(1), p.ID)
}

func TestPrefsRepo_Create_Error(t *testing.T) {
	repo, mock := newPrefsRepoMock(t)
	p := entities.NewUserNotificationPreferences(1)
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO notification_preferences")).WillReturnError(fmt.Errorf("db"))
	assert.Error(t, repo.Create(context.Background(), p))
}

func TestPrefsRepo_Update_Success(t *testing.T) {
	repo, mock := newPrefsRepoMock(t)
	p := entities.NewUserNotificationPreferences(1)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE notification_preferences SET")).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.Update(context.Background(), p))
}

func TestPrefsRepo_Update_NotFound(t *testing.T) {
	repo, mock := newPrefsRepoMock(t)
	p := entities.NewUserNotificationPreferences(999)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE notification_preferences SET")).WillReturnResult(sqlmock.NewResult(0, 0))
	assert.Error(t, repo.Update(context.Background(), p))
}

func TestPrefsRepo_Delete_Success(t *testing.T) {
	repo, mock := newPrefsRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM notification_preferences")).
		WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.Delete(context.Background(), 1))
}

func TestPrefsRepo_Delete_NotFound(t *testing.T) {
	repo, mock := newPrefsRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM notification_preferences")).
		WithArgs(int64(999)).WillReturnResult(sqlmock.NewResult(0, 0))
	assert.Error(t, repo.Delete(context.Background(), 999))
}

func TestPrefsRepo_GetByUserID_Success(t *testing.T) {
	repo, mock := newPrefsRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "user_id", "email_enabled", "push_enabled", "in_app_enabled",
		"telegram_enabled", "slack_enabled", "quiet_hours_enabled",
		"quiet_hours_start", "quiet_hours_end", "timezone",
		"digest_enabled", "digest_frequency", "digest_time",
		"type_preferences", "created_at", "updated_at",
	}).AddRow(
		1, int64(1), true, true, true, false, false, false,
		"22:00", "08:00", "UTC",
		false, "daily", "09:00",
		nil, now, now,
	)
	mock.ExpectQuery(regexp.QuoteMeta("FROM notification_preferences")).WithArgs(int64(1)).WillReturnRows(rows)
	p, err := repo.GetByUserID(context.Background(), 1)
	require.NoError(t, err)
	assert.NotNil(t, p)
	assert.True(t, p.EmailEnabled)
}

func TestPrefsRepo_GetByUserID_NotFound(t *testing.T) {
	repo, mock := newPrefsRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM notification_preferences")).WithArgs(int64(999)).WillReturnError(sql.ErrNoRows)
	p, err := repo.GetByUserID(context.Background(), 999)
	require.NoError(t, err)
	assert.Nil(t, p)
}

func TestPrefsRepo_GetOrCreate_Existing(t *testing.T) {
	repo, mock := newPrefsRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "user_id", "email_enabled", "push_enabled", "in_app_enabled",
		"telegram_enabled", "slack_enabled", "quiet_hours_enabled",
		"quiet_hours_start", "quiet_hours_end", "timezone",
		"digest_enabled", "digest_frequency", "digest_time",
		"type_preferences", "created_at", "updated_at",
	}).AddRow(1, int64(1), true, true, true, false, false, false, "22:00", "08:00", "UTC", false, "daily", "09:00", nil, now, now)
	mock.ExpectQuery(regexp.QuoteMeta("FROM notification_preferences")).WithArgs(int64(1)).WillReturnRows(rows)
	p, err := repo.GetOrCreate(context.Background(), 1)
	require.NoError(t, err)
	assert.NotNil(t, p)
}

func TestPrefsRepo_GetOrCreate_New(t *testing.T) {
	repo, mock := newPrefsRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM notification_preferences")).WithArgs(int64(1)).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO notification_preferences")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
	p, err := repo.GetOrCreate(context.Background(), 1)
	require.NoError(t, err)
	assert.NotNil(t, p)
}

func TestPrefsRepo_UpdateChannelEnabled_Success(t *testing.T) {
	repo, mock := newPrefsRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE notification_preferences SET")).
		WithArgs(int64(1), true).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.UpdateChannelEnabled(context.Background(), 1, entities.ChannelEmail, true))
}

func TestPrefsRepo_UpdateChannelEnabled_AllChannels(t *testing.T) {
	channels := []entities.NotificationChannel{
		entities.ChannelEmail, entities.ChannelPush, entities.ChannelInApp,
		entities.ChannelTelegram, entities.ChannelSlack,
	}
	for _, ch := range channels {
		repo, mock := newPrefsRepoMock(t)
		mock.ExpectExec(regexp.QuoteMeta("UPDATE notification_preferences SET")).
			WithArgs(int64(1), true).WillReturnResult(sqlmock.NewResult(0, 1))
		require.NoError(t, repo.UpdateChannelEnabled(context.Background(), 1, ch, true))
	}
}

func TestPrefsRepo_UpdateChannelEnabled_UnknownChannel(t *testing.T) {
	repo, _ := newPrefsRepoMock(t)
	err := repo.UpdateChannelEnabled(context.Background(), 1, "unknown", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown channel")
}

func TestPrefsRepo_UpdateChannelEnabled_CreateIfNotExists(t *testing.T) {
	repo, mock := newPrefsRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE notification_preferences SET")).
		WithArgs(int64(1), true).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO notification_preferences")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
	require.NoError(t, repo.UpdateChannelEnabled(context.Background(), 1, entities.ChannelEmail, true))
}

func TestPrefsRepo_UpdateQuietHours_Success(t *testing.T) {
	repo, mock := newPrefsRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE notification_preferences SET")).
		WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.UpdateQuietHours(context.Background(), 1, true, "22:00", "08:00", "UTC"))
}

func TestPrefsRepo_UpdateQuietHours_CreateIfNotExists(t *testing.T) {
	repo, mock := newPrefsRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE notification_preferences SET")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO notification_preferences")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
	require.NoError(t, repo.UpdateQuietHours(context.Background(), 1, true, "22:00", "08:00", "UTC"))
}
