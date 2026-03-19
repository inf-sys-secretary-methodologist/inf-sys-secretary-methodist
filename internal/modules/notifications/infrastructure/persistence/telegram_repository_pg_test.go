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

func newTGRepoMock(t *testing.T) (*TelegramRepositoryPG, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return &TelegramRepositoryPG{db: db}, mock
}

func TestNewTelegramRepositoryPG(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer func() { _ = db.Close() }()
	assert.NotNil(t, NewTelegramRepositoryPG(db))
}

func TestTGRepo_CreateVerificationCode(t *testing.T) {
	repo, mock := newTGRepoMock(t)
	code := &entities.TelegramVerificationCode{
		UserID: 1, Code: "123456", ExpiresAt: time.Now().Add(time.Hour), CreatedAt: time.Now(),
	}
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO telegram_verification_codes")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
	require.NoError(t, repo.CreateVerificationCode(context.Background(), code))
	assert.Equal(t, int64(1), code.ID)
}

func TestTGRepo_CreateVerificationCode_Error(t *testing.T) {
	repo, mock := newTGRepoMock(t)
	code := &entities.TelegramVerificationCode{UserID: 1, Code: "123456", ExpiresAt: time.Now(), CreatedAt: time.Now()}
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO telegram_verification_codes")).WillReturnError(fmt.Errorf("err"))
	assert.Error(t, repo.CreateVerificationCode(context.Background(), code))
}

func TestTGRepo_GetVerificationCodeByCode_Success(t *testing.T) {
	repo, mock := newTGRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "user_id", "code", "expires_at", "used_at", "created_at"}).
		AddRow(1, int64(1), "123456", now.Add(time.Hour), nil, now)
	mock.ExpectQuery(regexp.QuoteMeta("FROM telegram_verification_codes")).WithArgs("123456").WillReturnRows(rows)
	vc, err := repo.GetVerificationCodeByCode(context.Background(), "123456")
	require.NoError(t, err)
	assert.NotNil(t, vc)
	assert.Equal(t, "123456", vc.Code)
}

func TestTGRepo_GetVerificationCodeByCode_NotFound(t *testing.T) {
	repo, mock := newTGRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM telegram_verification_codes")).WithArgs("xxx").WillReturnError(sql.ErrNoRows)
	vc, err := repo.GetVerificationCodeByCode(context.Background(), "xxx")
	require.NoError(t, err)
	assert.Nil(t, vc)
}

func TestTGRepo_GetActiveVerificationCodeByUserID(t *testing.T) {
	repo, mock := newTGRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "user_id", "code", "expires_at", "used_at", "created_at"}).
		AddRow(1, int64(1), "abc", now.Add(time.Hour), nil, now)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE user_id = $1 AND used_at IS NULL")).WithArgs(int64(1)).WillReturnRows(rows)
	vc, err := repo.GetActiveVerificationCodeByUserID(context.Background(), 1)
	require.NoError(t, err)
	assert.NotNil(t, vc)
}

func TestTGRepo_MarkCodeAsUsed_Success(t *testing.T) {
	repo, mock := newTGRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE telegram_verification_codes SET used_at")).
		WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.MarkCodeAsUsed(context.Background(), 1))
}

func TestTGRepo_MarkCodeAsUsed_NotFound(t *testing.T) {
	repo, mock := newTGRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE telegram_verification_codes SET used_at")).
		WithArgs(int64(999)).WillReturnResult(sqlmock.NewResult(0, 0))
	assert.Error(t, repo.MarkCodeAsUsed(context.Background(), 999))
}

func TestTGRepo_DeleteExpiredCodes(t *testing.T) {
	repo, mock := newTGRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM telegram_verification_codes")).
		WillReturnResult(sqlmock.NewResult(0, 3))
	require.NoError(t, repo.DeleteExpiredCodes(context.Background()))
}

func TestTGRepo_CreateConnection(t *testing.T) {
	repo, mock := newTGRepoMock(t)
	conn := &entities.TelegramConnection{
		UserID: 1, TelegramChatID: 12345, IsActive: true,
		ConnectedAt: time.Now(), UpdatedAt: time.Now(),
	}
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO user_telegram_connections")).
		WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.CreateConnection(context.Background(), conn))
}

func TestTGRepo_GetConnectionByUserID_Success(t *testing.T) {
	repo, mock := newTGRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows([]string{"user_id", "telegram_chat_id", "telegram_username", "telegram_first_name", "is_active", "connected_at", "updated_at"}).
		AddRow(int64(1), int64(12345), sql.NullString{String: "user", Valid: true}, sql.NullString{}, true, now, now)
	mock.ExpectQuery(regexp.QuoteMeta("FROM user_telegram_connections")).WithArgs(int64(1)).WillReturnRows(rows)
	conn, err := repo.GetConnectionByUserID(context.Background(), 1)
	require.NoError(t, err)
	assert.NotNil(t, conn)
	assert.Equal(t, int64(12345), conn.TelegramChatID)
}

func TestTGRepo_GetConnectionByUserID_NotFound(t *testing.T) {
	repo, mock := newTGRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM user_telegram_connections")).WithArgs(int64(999)).WillReturnError(sql.ErrNoRows)
	conn, err := repo.GetConnectionByUserID(context.Background(), 999)
	require.NoError(t, err)
	assert.Nil(t, conn)
}

func TestTGRepo_GetConnectionByChatID(t *testing.T) {
	repo, mock := newTGRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows([]string{"user_id", "telegram_chat_id", "telegram_username", "telegram_first_name", "is_active", "connected_at", "updated_at"}).
		AddRow(int64(1), int64(12345), sql.NullString{}, sql.NullString{}, true, now, now)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE telegram_chat_id = $1")).WithArgs(int64(12345)).WillReturnRows(rows)
	conn, err := repo.GetConnectionByChatID(context.Background(), 12345)
	require.NoError(t, err)
	assert.NotNil(t, conn)
}

func TestTGRepo_GetActiveConnections(t *testing.T) {
	repo, mock := newTGRepoMock(t)
	rows := sqlmock.NewRows([]string{"user_id", "telegram_chat_id", "telegram_username", "telegram_first_name", "is_active", "connected_at", "updated_at"})
	mock.ExpectQuery(regexp.QuoteMeta("WHERE is_active = true")).WillReturnRows(rows)
	conns, err := repo.GetActiveConnections(context.Background())
	require.NoError(t, err)
	assert.Empty(t, conns)
}

func TestTGRepo_UpdateConnection_Success(t *testing.T) {
	repo, mock := newTGRepoMock(t)
	conn := &entities.TelegramConnection{UserID: 1, TelegramChatID: 12345, IsActive: true}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE user_telegram_connections SET")).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.UpdateConnection(context.Background(), conn))
}

func TestTGRepo_UpdateConnection_NotFound(t *testing.T) {
	repo, mock := newTGRepoMock(t)
	conn := &entities.TelegramConnection{UserID: 999}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE user_telegram_connections SET")).WillReturnResult(sqlmock.NewResult(0, 0))
	assert.Error(t, repo.UpdateConnection(context.Background(), conn))
}

func TestTGRepo_DeleteConnection_Success(t *testing.T) {
	repo, mock := newTGRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM user_telegram_connections")).
		WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.DeleteConnection(context.Background(), 1))
}

func TestTGRepo_DeleteConnection_NotFound(t *testing.T) {
	repo, mock := newTGRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM user_telegram_connections")).
		WithArgs(int64(999)).WillReturnResult(sqlmock.NewResult(0, 0))
	assert.Error(t, repo.DeleteConnection(context.Background(), 999))
}

func TestTGRepo_GetActiveConnections_Success(t *testing.T) {
	repo, mock := newTGRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows([]string{"user_id", "telegram_chat_id", "telegram_username", "telegram_first_name", "is_active", "connected_at", "updated_at"}).
		AddRow(int64(1), int64(12345), sql.NullString{String: "user1", Valid: true}, sql.NullString{String: "First", Valid: true}, true, now, now)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE is_active = true")).WillReturnRows(rows)
	conns, err := repo.GetActiveConnections(context.Background())
	require.NoError(t, err)
	assert.Len(t, conns, 1)
	assert.Equal(t, "user1", conns[0].TelegramUsername)
}

func TestTGRepo_GetActiveConnections_Error(t *testing.T) {
	repo, mock := newTGRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE is_active = true")).WillReturnError(fmt.Errorf("db"))
	_, err := repo.GetActiveConnections(context.Background())
	assert.Error(t, err)
}

func TestTGRepo_GetActiveVerificationCodeByUserID_NotFound(t *testing.T) {
	repo, mock := newTGRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE user_id = $1 AND used_at IS NULL")).WillReturnError(sql.ErrNoRows)
	vc, err := repo.GetActiveVerificationCodeByUserID(context.Background(), 1)
	require.NoError(t, err)
	assert.Nil(t, vc)
}

func TestTGRepo_GetConnectionByChatID_Error(t *testing.T) {
	repo, mock := newTGRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE telegram_chat_id")).WillReturnError(fmt.Errorf("db"))
	_, err := repo.GetConnectionByChatID(context.Background(), 123)
	assert.Error(t, err)
}

func TestTGRepo_CreateConnection_Error(t *testing.T) {
	repo, mock := newTGRepoMock(t)
	conn := &entities.TelegramConnection{UserID: 1, TelegramChatID: 123}
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO user_telegram_connections")).WillReturnError(fmt.Errorf("db"))
	assert.Error(t, repo.CreateConnection(context.Background(), conn))
}

func TestTGRepo_DeleteExpiredCodes_Error(t *testing.T) {
	repo, mock := newTGRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM telegram_verification_codes")).WillReturnError(fmt.Errorf("db"))
	assert.Error(t, repo.DeleteExpiredCodes(context.Background()))
}
