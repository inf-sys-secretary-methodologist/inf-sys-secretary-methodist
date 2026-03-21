package persistence

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/messaging/domain/entities"
)

func newMsgRepoMock(t *testing.T) (*MessageRepositoryPG, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	repo := NewMessageRepositoryPG(db)
	return repo.(*MessageRepositoryPG), mock
}

var msgFullCols = []string{
	"id", "conversation_id", "sender_id", "type", "content", "reply_to_id",
	"is_edited", "edited_at", "is_deleted", "deleted_at", "created_at",
	"name", "avatar",
}

var msgBasicCols = []string{
	"id", "conversation_id", "sender_id", "type", "content",
	"is_edited", "is_deleted", "created_at",
	"name", "avatar",
}

var attachCols = []string{
	"id", "message_id", "file_id", "file_name", "file_size", "mime_type", "url", "created_at",
}

// --- Create ---

func TestMessageRepositoryPG_Create_Success(t *testing.T) {
	repo, mock := newMsgRepoMock(t)
	msg, _ := entities.NewTextMessage(1, 2, "hello")

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO messages")).
		WithArgs(msg.ConversationID, msg.SenderID, msg.Type, msg.Content, msg.ReplyToID, msg.CreatedAt).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	err := repo.Create(context.Background(), msg)
	require.NoError(t, err)
	assert.Equal(t, int64(1), msg.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepositoryPG_Create_Error(t *testing.T) {
	repo, mock := newMsgRepoMock(t)
	msg, _ := entities.NewTextMessage(1, 2, "hello")

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO messages")).
		WillReturnError(sql.ErrConnDone)

	err := repo.Create(context.Background(), msg)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetByID ---

func TestMessageRepositoryPG_GetByID_Success(t *testing.T) {
	repo, mock := newMsgRepoMock(t)
	now := time.Now()
	avatar := "/av.jpg"

	mock.ExpectQuery(regexp.QuoteMeta("SELECT m.id, m.conversation_id")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(msgFullCols).AddRow(
			int64(1), int64(1), int64(2), "text", "hello", nil,
			false, nil, false, nil, now,
			"User", &avatar,
		))

	// GetAttachments
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, message_id")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(attachCols))

	msg, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	require.NotNil(t, msg)
	assert.Equal(t, "hello", msg.Content)
	assert.Equal(t, "User", msg.SenderName)
	assert.NotNil(t, msg.SenderAvatarURL)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepositoryPG_GetByID_WithReply(t *testing.T) {
	repo, mock := newMsgRepoMock(t)
	now := time.Now()
	replyToID := int64(50)

	// GetByID for message itself
	mock.ExpectQuery(regexp.QuoteMeta("SELECT m.id, m.conversation_id")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(msgFullCols).AddRow(
			int64(1), int64(1), int64(2), "text", "reply", &replyToID,
			false, nil, false, nil, now,
			"User", nil,
		))

	// GetByID recursively for reply
	mock.ExpectQuery(regexp.QuoteMeta("SELECT m.id, m.conversation_id")).
		WithArgs(int64(50)).
		WillReturnRows(sqlmock.NewRows(msgFullCols).AddRow(
			int64(50), int64(1), int64(3), "text", "original", nil,
			false, nil, false, nil, now,
			"Other", nil,
		))
	// GetAttachments for reply
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, message_id")).
		WithArgs(int64(50)).
		WillReturnRows(sqlmock.NewRows(attachCols))
	// GetAttachments for message
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, message_id")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(attachCols))

	msg, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	require.NotNil(t, msg)
	require.NotNil(t, msg.ReplyTo)
	assert.Equal(t, "original", msg.ReplyTo.Content)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepositoryPG_GetByID_NotFound(t *testing.T) {
	repo, mock := newMsgRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT m.id")).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	msg, err := repo.GetByID(context.Background(), 999)
	require.ErrorIs(t, err, entities.ErrMessageNotFound)
	assert.Nil(t, msg)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepositoryPG_GetByID_Error(t *testing.T) {
	repo, mock := newMsgRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT m.id")).
		WithArgs(int64(1)).
		WillReturnError(sql.ErrConnDone)

	msg, err := repo.GetByID(context.Background(), 1)
	require.Error(t, err)
	assert.Nil(t, msg)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepositoryPG_GetByID_AttachmentError(t *testing.T) {
	repo, mock := newMsgRepoMock(t)
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT m.id, m.conversation_id")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(msgFullCols).AddRow(
			int64(1), int64(1), int64(2), "text", "hello", nil,
			false, nil, false, nil, now, nil, nil,
		))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, message_id")).
		WithArgs(int64(1)).
		WillReturnError(sql.ErrConnDone)

	msg, err := repo.GetByID(context.Background(), 1)
	require.Error(t, err)
	assert.Nil(t, msg)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- Update ---

func TestMessageRepositoryPG_Update_Success(t *testing.T) {
	repo, mock := newMsgRepoMock(t)
	msg := &entities.Message{ID: 1, Content: "updated", IsEdited: true}
	now := time.Now()
	msg.EditedAt = &now

	mock.ExpectExec(regexp.QuoteMeta("UPDATE messages")).
		WithArgs(msg.ID, msg.Content, msg.IsEdited, msg.EditedAt, msg.IsDeleted, msg.DeletedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), msg)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepositoryPG_Update_NotFound(t *testing.T) {
	repo, mock := newMsgRepoMock(t)
	msg := &entities.Message{ID: 999}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE messages")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Update(context.Background(), msg)
	require.ErrorIs(t, err, entities.ErrMessageNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepositoryPG_Update_Error(t *testing.T) {
	repo, mock := newMsgRepoMock(t)
	msg := &entities.Message{ID: 1}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE messages")).
		WillReturnError(sql.ErrConnDone)

	err := repo.Update(context.Background(), msg)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- Delete ---

func TestMessageRepositoryPG_Delete_Success(t *testing.T) {
	repo, mock := newMsgRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM messages WHERE id = $1")).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), 1)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepositoryPG_Delete_NotFound(t *testing.T) {
	repo, mock := newMsgRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM messages")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Delete(context.Background(), 999)
	require.ErrorIs(t, err, entities.ErrMessageNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepositoryPG_Delete_Error(t *testing.T) {
	repo, mock := newMsgRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM messages")).
		WillReturnError(sql.ErrConnDone)

	err := repo.Delete(context.Background(), 1)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetLastMessage ---

func TestMessageRepositoryPG_GetLastMessage_Success(t *testing.T) {
	repo, mock := newMsgRepoMock(t)
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT m.id, m.conversation_id")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(msgFullCols).AddRow(
			int64(10), int64(1), int64(2), "text", "last msg", nil,
			false, nil, false, nil, now, "User", nil,
		))

	msg, err := repo.GetLastMessage(context.Background(), 1)
	require.NoError(t, err)
	require.NotNil(t, msg)
	assert.Equal(t, "last msg", msg.Content)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepositoryPG_GetLastMessage_NoMessages(t *testing.T) {
	repo, mock := newMsgRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT m.id")).
		WithArgs(int64(1)).
		WillReturnError(sql.ErrNoRows)

	msg, err := repo.GetLastMessage(context.Background(), 1)
	require.NoError(t, err)
	assert.Nil(t, msg)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepositoryPG_GetLastMessage_Error(t *testing.T) {
	repo, mock := newMsgRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT m.id")).
		WillReturnError(sql.ErrConnDone)

	_, err := repo.GetLastMessage(context.Background(), 1)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- CountUnread ---

func TestMessageRepositoryPG_CountUnread_WithLastReadID(t *testing.T) {
	repo, mock := newMsgRepoMock(t)
	lastID := int64(50)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WithArgs(int64(1), lastID, int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	count, err := repo.CountUnread(context.Background(), 1, 2, &lastID)
	require.NoError(t, err)
	assert.Equal(t, 3, count)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepositoryPG_CountUnread_NoLastReadID(t *testing.T) {
	repo, mock := newMsgRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WithArgs(int64(1), int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))

	count, err := repo.CountUnread(context.Background(), 1, 2, nil)
	require.NoError(t, err)
	assert.Equal(t, 10, count)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepositoryPG_CountUnread_Error(t *testing.T) {
	repo, mock := newMsgRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnError(sql.ErrConnDone)

	_, err := repo.CountUnread(context.Background(), 1, 2, nil)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- CreateAttachment ---

func TestMessageRepositoryPG_CreateAttachment_Success(t *testing.T) {
	repo, mock := newMsgRepoMock(t)
	att := &entities.Attachment{MessageID: 1, FileID: 10, FileName: "f.txt", FileSize: 100, MimeType: "text/plain", URL: "/url", CreatedAt: time.Now()}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO message_attachments")).
		WithArgs(att.MessageID, att.FileID, att.FileName, att.FileSize, att.MimeType, att.URL, att.CreatedAt).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	err := repo.CreateAttachment(context.Background(), att)
	require.NoError(t, err)
	assert.Equal(t, int64(1), att.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepositoryPG_CreateAttachment_Error(t *testing.T) {
	repo, mock := newMsgRepoMock(t)
	att := &entities.Attachment{MessageID: 1, CreatedAt: time.Now()}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO message_attachments")).
		WillReturnError(sql.ErrConnDone)

	err := repo.CreateAttachment(context.Background(), att)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetAttachments ---

func TestMessageRepositoryPG_GetAttachments_Success(t *testing.T) {
	repo, mock := newMsgRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows(attachCols).
		AddRow(int64(1), int64(1), int64(10), "f.txt", int64(100), "text/plain", "/url", now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, message_id")).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	attachments, err := repo.GetAttachments(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, attachments, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepositoryPG_GetAttachments_Error(t *testing.T) {
	repo, mock := newMsgRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, message_id")).
		WillReturnError(sql.ErrConnDone)

	_, err := repo.GetAttachments(context.Background(), 1)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepositoryPG_GetAttachments_ScanError(t *testing.T) {
	repo, mock := newMsgRepoMock(t)
	rows := sqlmock.NewRows(attachCols).
		AddRow("bad", int64(1), int64(10), "f.txt", int64(100), "text/plain", "/url", time.Now())

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, message_id")).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	_, err := repo.GetAttachments(context.Background(), 1)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- List ---

func TestMessageRepositoryPG_List_Success(t *testing.T) {
	repo, mock := newMsgRepoMock(t)
	now := time.Now()
	filter := entities.MessageFilter{ConversationID: 1, Limit: 10}

	rows := sqlmock.NewRows(msgFullCols).
		AddRow(int64(1), int64(1), int64(2), "text", "msg", nil,
			false, nil, false, nil, now, "User", nil)

	mock.ExpectQuery("SELECT m.id, m.conversation_id").WillReturnRows(rows)

	// GetAttachments for each message
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, message_id")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(attachCols))

	msgs, err := repo.List(context.Background(), filter)
	require.NoError(t, err)
	assert.Len(t, msgs, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepositoryPG_List_WithFilters(t *testing.T) {
	repo, mock := newMsgRepoMock(t)
	now := time.Now()
	beforeID := int64(100)
	afterID := int64(1)
	senderID := int64(2)
	msgType := entities.MessageTypeText
	filter := entities.MessageFilter{
		ConversationID: 1,
		BeforeID:       &beforeID,
		AfterID:        &afterID,
		SenderID:       &senderID,
		Type:           &msgType,
		Limit:          5,
	}

	rows := sqlmock.NewRows(msgFullCols).
		AddRow(int64(50), int64(1), int64(2), "text", "filtered", nil,
			false, nil, false, nil, now, "User", nil)

	mock.ExpectQuery("SELECT m.id, m.conversation_id").WillReturnRows(rows)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, message_id")).
		WithArgs(int64(50)).
		WillReturnRows(sqlmock.NewRows(attachCols))

	msgs, err := repo.List(context.Background(), filter)
	require.NoError(t, err)
	assert.Len(t, msgs, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepositoryPG_List_DefaultLimit(t *testing.T) {
	repo, mock := newMsgRepoMock(t)
	filter := entities.MessageFilter{ConversationID: 1, Limit: 0}

	mock.ExpectQuery("SELECT m.id, m.conversation_id").
		WillReturnRows(sqlmock.NewRows(msgFullCols))

	msgs, err := repo.List(context.Background(), filter)
	require.NoError(t, err)
	assert.Empty(t, msgs)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepositoryPG_List_Error(t *testing.T) {
	repo, mock := newMsgRepoMock(t)
	filter := entities.MessageFilter{ConversationID: 1, Limit: 10}

	mock.ExpectQuery("SELECT m.id").
		WillReturnError(sql.ErrConnDone)

	_, err := repo.List(context.Background(), filter)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepositoryPG_List_ScanError(t *testing.T) {
	repo, mock := newMsgRepoMock(t)
	filter := entities.MessageFilter{ConversationID: 1, Limit: 10}

	rows := sqlmock.NewRows(msgFullCols).
		AddRow("bad", int64(1), int64(2), "text", "msg", nil,
			false, nil, false, nil, time.Now(), nil, nil)

	mock.ExpectQuery("SELECT m.id").WillReturnRows(rows)

	_, err := repo.List(context.Background(), filter)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepositoryPG_List_WithReply(t *testing.T) {
	repo, mock := newMsgRepoMock(t)
	now := time.Now()
	replyToID := int64(50)
	filter := entities.MessageFilter{ConversationID: 1, Limit: 10}

	rows := sqlmock.NewRows(msgFullCols).
		AddRow(int64(1), int64(1), int64(2), "text", "reply msg", &replyToID,
			false, nil, false, nil, now, "User", nil)

	mock.ExpectQuery("SELECT m.id, m.conversation_id").WillReturnRows(rows)

	// getMessageBasic for the reply
	mock.ExpectQuery(regexp.QuoteMeta("SELECT m.id, m.conversation_id, m.sender_id, m.type, m.content")).
		WithArgs(int64(50)).
		WillReturnRows(sqlmock.NewRows(msgBasicCols).AddRow(
			int64(50), int64(1), int64(3), "text", "original",
			false, false, now, "Other", nil,
		))

	// GetAttachments
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, message_id")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(attachCols))

	msgs, err := repo.List(context.Background(), filter)
	require.NoError(t, err)
	require.Len(t, msgs, 1)
	require.NotNil(t, msgs[0].ReplyTo)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- Search ---

func TestMessageRepositoryPG_Search_Success(t *testing.T) {
	repo, mock := newMsgRepoMock(t)
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WithArgs(int64(1), "test").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	searchCols := []string{
		"id", "conversation_id", "sender_id", "type", "content",
		"is_edited", "edited_at", "is_deleted", "deleted_at", "created_at",
		"name", "avatar", "rank",
	}
	rows := sqlmock.NewRows(searchCols).AddRow(
		int64(1), int64(1), int64(2), "text", "test msg",
		false, nil, false, nil, now,
		"User", nil, float64(0.5),
	)

	mock.ExpectQuery("SELECT m.id, m.conversation_id").
		WithArgs(int64(1), "test", 10, 0).
		WillReturnRows(rows)

	msgs, total, err := repo.Search(context.Background(), 1, "test", 10, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, msgs, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepositoryPG_Search_CountError(t *testing.T) {
	repo, mock := newMsgRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnError(sql.ErrConnDone)

	_, _, err := repo.Search(context.Background(), 1, "test", 10, 0)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepositoryPG_Search_QueryError(t *testing.T) {
	repo, mock := newMsgRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	mock.ExpectQuery("SELECT m.id").
		WillReturnError(sql.ErrConnDone)

	_, _, err := repo.Search(context.Background(), 1, "test", 10, 0)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepositoryPG_Search_ScanError(t *testing.T) {
	repo, mock := newMsgRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	searchCols := []string{
		"id", "conversation_id", "sender_id", "type", "content",
		"is_edited", "edited_at", "is_deleted", "deleted_at", "created_at",
		"name", "avatar", "rank",
	}
	rows := sqlmock.NewRows(searchCols).AddRow(
		"bad", int64(1), int64(2), "text", "test",
		false, nil, false, nil, time.Now(),
		nil, nil, float64(0.5),
	)
	mock.ExpectQuery("SELECT m.id").WillReturnRows(rows)

	_, _, err := repo.Search(context.Background(), 1, "test", 10, 0)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- getMessageBasic ---

func TestMessageRepositoryPG_getMessageBasic_Success(t *testing.T) {
	repo, mock := newMsgRepoMock(t)
	now := time.Now()
	avatar := "/av.jpg"

	mock.ExpectQuery(regexp.QuoteMeta("SELECT m.id, m.conversation_id, m.sender_id, m.type, m.content")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(msgBasicCols).AddRow(
			int64(1), int64(1), int64(2), "text", "basic",
			false, false, now, "User", &avatar,
		))

	msg, err := repo.getMessageBasic(context.Background(), 1)
	require.NoError(t, err)
	require.NotNil(t, msg)
	assert.Equal(t, "basic", msg.Content)
	assert.Equal(t, "User", msg.SenderName)
	assert.NotNil(t, msg.SenderAvatarURL)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepositoryPG_getMessageBasic_Error(t *testing.T) {
	repo, mock := newMsgRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT m.id")).
		WithArgs(int64(1)).
		WillReturnError(sql.ErrConnDone)

	msg, err := repo.getMessageBasic(context.Background(), 1)
	require.Error(t, err)
	assert.Nil(t, msg)
	require.NoError(t, mock.ExpectationsWereMet())
}
