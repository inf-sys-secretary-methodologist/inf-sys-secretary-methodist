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

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/entities"
)

func newConvRepoMock(t *testing.T) (*ConversationRepositoryPg, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	repo := NewConversationRepositoryPg(db)
	return repo.(*ConversationRepositoryPg), mock
}

func newMsgRepoMock(t *testing.T) (*MessageRepositoryPg, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	repo := NewMessageRepositoryPg(db)
	return repo.(*MessageRepositoryPg), mock
}

// ---- Conversation tests ----

func TestConversationCreate_Success(t *testing.T) {
	repo, mock := newConvRepoMock(t)
	conv := &entities.Conversation{UserID: 1, Title: "Test", Model: "gpt-4", MessageCount: 0}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO ai_conversations")).
		WithArgs(conv.UserID, conv.Title, conv.Model, conv.MessageCount, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(42)))

	err := repo.Create(context.Background(), conv)
	require.NoError(t, err)
	assert.Equal(t, int64(42), conv.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationCreate_Error(t *testing.T) {
	repo, mock := newConvRepoMock(t)
	conv := &entities.Conversation{UserID: 1, Title: "Test", Model: "gpt-4"}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO ai_conversations")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("db error"))

	err := repo.Create(context.Background(), conv)
	assert.Error(t, err)
}

func TestConversationGetByID_Success(t *testing.T) {
	repo, mock := newConvRepoMock(t)
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, title, model, message_count, last_message_at, created_at, updated_at")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "title", "model", "message_count", "last_message_at", "created_at", "updated_at"}).
			AddRow(int64(1), int64(10), "Test", "gpt-4", 5, &now, now, now))

	conv, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), conv.ID)
	assert.Equal(t, "Test", conv.Title)
}

func TestConversationGetByID_NotFound(t *testing.T) {
	repo, mock := newConvRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, title, model")).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	conv, err := repo.GetByID(context.Background(), 999)
	assert.Nil(t, conv)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "conversation not found")
}

func TestConversationGetByID_DBError(t *testing.T) {
	repo, mock := newConvRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, title, model")).
		WithArgs(int64(1)).
		WillReturnError(fmt.Errorf("connection refused"))

	conv, err := repo.GetByID(context.Background(), 1)
	assert.Nil(t, conv)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get conversation")
}

func TestConversationGetByUserID_Success(t *testing.T) {
	repo, mock := newConvRepoMock(t)
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	cols := []string{"id", "user_id", "title", "model", "message_count", "last_message_at", "created_at", "updated_at"}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, title, model, message_count, last_message_at, created_at, updated_at")).
		WithArgs(int64(1), 10, 0).
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow(int64(1), int64(1), "Conv1", "gpt-4", 3, &now, now, now).
			AddRow(int64(2), int64(1), "Conv2", "gpt-4", 1, nil, now, now))

	convs, total, err := repo.GetByUserID(context.Background(), 1, 10, 0)
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, convs, 2)
}

func TestConversationGetByUserID_CountError(t *testing.T) {
	repo, mock := newConvRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WithArgs(int64(1)).
		WillReturnError(fmt.Errorf("db error"))

	_, _, err := repo.GetByUserID(context.Background(), 1, 10, 0)
	assert.Error(t, err)
}

func TestConversationGetByUserID_QueryError(t *testing.T) {
	repo, mock := newConvRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id")).
		WithArgs(int64(1), 10, 0).
		WillReturnError(fmt.Errorf("query error"))

	_, _, err := repo.GetByUserID(context.Background(), 1, 10, 0)
	assert.Error(t, err)
}

func TestConversationGetByUserID_ScanError(t *testing.T) {
	repo, mock := newConvRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id")).
		WithArgs(int64(1), 10, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("invalid"))

	_, _, err := repo.GetByUserID(context.Background(), 1, 10, 0)
	assert.Error(t, err)
}

func TestConversationUpdate_Success(t *testing.T) {
	repo, mock := newConvRepoMock(t)
	conv := &entities.Conversation{ID: 1, Title: "Updated", Model: "gpt-4"}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE ai_conversations")).
		WithArgs(conv.Title, conv.Model, sqlmock.AnyArg(), conv.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), conv)
	require.NoError(t, err)
}

func TestConversationUpdate_Error(t *testing.T) {
	repo, mock := newConvRepoMock(t)
	conv := &entities.Conversation{ID: 1, Title: "Updated", Model: "gpt-4"}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE ai_conversations")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("update error"))

	err := repo.Update(context.Background(), conv)
	assert.Error(t, err)
}

func TestConversationDelete_Success(t *testing.T) {
	repo, mock := newConvRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM ai_conversations")).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), 1)
	require.NoError(t, err)
}

func TestConversationDelete_Error(t *testing.T) {
	repo, mock := newConvRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM ai_conversations")).
		WithArgs(int64(1)).
		WillReturnError(fmt.Errorf("delete error"))

	err := repo.Delete(context.Background(), 1)
	assert.Error(t, err)
}

func TestConversationSearch_Success(t *testing.T) {
	repo, mock := newConvRepoMock(t)
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WithArgs(int64(1), "test").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	cols := []string{"id", "user_id", "title", "model", "message_count", "last_message_at", "created_at", "updated_at"}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, title, model")).
		WithArgs(int64(1), "test", 10, 0).
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow(int64(1), int64(1), "Test Conv", "gpt-4", 2, nil, now, now))

	convs, total, err := repo.Search(context.Background(), 1, "test", 10, 0)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, convs, 1)
}

func TestConversationSearch_CountError(t *testing.T) {
	repo, mock := newConvRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WithArgs(int64(1), "test").
		WillReturnError(fmt.Errorf("count error"))

	_, _, err := repo.Search(context.Background(), 1, "test", 10, 0)
	assert.Error(t, err)
}

func TestConversationSearch_QueryError(t *testing.T) {
	repo, mock := newConvRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WithArgs(int64(1), "test").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, title, model")).
		WithArgs(int64(1), "test", 10, 0).
		WillReturnError(fmt.Errorf("query error"))

	_, _, err := repo.Search(context.Background(), 1, "test", 10, 0)
	assert.Error(t, err)
}

func TestConversationSearch_ScanError(t *testing.T) {
	repo, mock := newConvRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WithArgs(int64(1), "test").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, user_id, title, model")).
		WithArgs(int64(1), "test", 10, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("invalid"))

	_, _, err := repo.Search(context.Background(), 1, "test", 10, 0)
	assert.Error(t, err)
}

// ---- Message tests ----

func TestMessageCreate_Success(t *testing.T) {
	repo, mock := newMsgRepoMock(t)
	msg := &entities.Message{ConversationID: 1, Role: "user", Content: "Hello"}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO ai_messages")).
		WithArgs(msg.ConversationID, msg.Role, msg.Content, msg.TokensUsed, msg.Model, msg.ErrorMessage, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(10)))

	err := repo.Create(context.Background(), msg)
	require.NoError(t, err)
	assert.Equal(t, int64(10), msg.ID)
}

func TestMessageCreate_Error(t *testing.T) {
	repo, mock := newMsgRepoMock(t)
	msg := &entities.Message{ConversationID: 1, Role: "user", Content: "Hello"}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO ai_messages")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("insert error"))

	err := repo.Create(context.Background(), msg)
	assert.Error(t, err)
}

func TestMessageGetByConversationID_NilBeforeID(t *testing.T) {
	repo, mock := newMsgRepoMock(t)
	now := time.Now()

	cols := []string{"id", "conversation_id", "role", "content", "tokens_used", "model", "error_message", "created_at"}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, conversation_id, role, content")).
		WithArgs(int64(1), 11).
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow(int64(2), int64(1), "user", "msg2", nil, nil, nil, now).
			AddRow(int64(1), int64(1), "assistant", "msg1", nil, nil, nil, now))

	msgs, hasMore, err := repo.GetByConversationID(context.Background(), 1, 10, nil)
	require.NoError(t, err)
	assert.False(t, hasMore)
	assert.Len(t, msgs, 2)
	// Should be reversed (chronological order)
	assert.Equal(t, int64(1), msgs[0].ID)
	assert.Equal(t, int64(2), msgs[1].ID)
}

func TestMessageGetByConversationID_WithBeforeID(t *testing.T) {
	repo, mock := newMsgRepoMock(t)
	now := time.Now()
	beforeID := int64(50)

	cols := []string{"id", "conversation_id", "role", "content", "tokens_used", "model", "error_message", "created_at"}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, conversation_id, role, content")).
		WithArgs(int64(1), beforeID, 11).
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow(int64(2), int64(1), "user", "msg", nil, nil, nil, now))

	msgs, hasMore, err := repo.GetByConversationID(context.Background(), 1, 10, &beforeID)
	require.NoError(t, err)
	assert.False(t, hasMore)
	assert.Len(t, msgs, 1)
}

func TestMessageGetByConversationID_HasMore(t *testing.T) {
	repo, mock := newMsgRepoMock(t)
	now := time.Now()
	limit := 2

	cols := []string{"id", "conversation_id", "role", "content", "tokens_used", "model", "error_message", "created_at"}
	rows := sqlmock.NewRows(cols)
	for i := 3; i >= 1; i-- {
		rows.AddRow(int64(i), int64(1), "user", fmt.Sprintf("msg%d", i), nil, nil, nil, now)
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, conversation_id, role, content")).
		WithArgs(int64(1), limit+1).
		WillReturnRows(rows)

	msgs, hasMore, err := repo.GetByConversationID(context.Background(), 1, limit, nil)
	require.NoError(t, err)
	assert.True(t, hasMore)
	assert.Len(t, msgs, limit)
}

func TestMessageGetByConversationID_QueryError(t *testing.T) {
	repo, mock := newMsgRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, conversation_id")).
		WithArgs(int64(1), 11).
		WillReturnError(fmt.Errorf("query error"))

	_, _, err := repo.GetByConversationID(context.Background(), 1, 10, nil)
	assert.Error(t, err)
}

func TestMessageGetByConversationID_ScanError(t *testing.T) {
	repo, mock := newMsgRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, conversation_id")).
		WithArgs(int64(1), 11).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("invalid"))

	_, _, err := repo.GetByConversationID(context.Background(), 1, 10, nil)
	assert.Error(t, err)
}

func TestMessageGetByID_Success(t *testing.T) {
	repo, mock := newMsgRepoMock(t)
	now := time.Now()

	cols := []string{"id", "conversation_id", "role", "content", "tokens_used", "model", "error_message", "created_at"}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, conversation_id, role, content")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(cols).AddRow(int64(1), int64(10), "user", "hello", nil, nil, nil, now))

	msg, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), msg.ID)
}

func TestMessageGetByID_NotFound(t *testing.T) {
	repo, mock := newMsgRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, conversation_id")).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	msg, err := repo.GetByID(context.Background(), 999)
	assert.Nil(t, msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "message not found")
}

func TestMessageGetByID_DBError(t *testing.T) {
	repo, mock := newMsgRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, conversation_id")).
		WithArgs(int64(1)).
		WillReturnError(fmt.Errorf("db error"))

	msg, err := repo.GetByID(context.Background(), 1)
	assert.Nil(t, msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get message")
}

func TestMessageCreateMessageSource_Success(t *testing.T) {
	repo, mock := newMsgRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO ai_message_sources")).
		WithArgs(int64(1), int64(2), 0.95, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.CreateMessageSource(context.Background(), 1, 2, 0.95)
	require.NoError(t, err)
}

func TestMessageCreateMessageSource_Error(t *testing.T) {
	repo, mock := newMsgRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO ai_message_sources")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("insert error"))

	err := repo.CreateMessageSource(context.Background(), 1, 2, 0.95)
	assert.Error(t, err)
}

func TestMessageGetMessageSources_Success(t *testing.T) {
	repo, mock := newMsgRepoMock(t)

	cols := []string{"id", "message_id", "chunk_id", "document_id", "document_title", "chunk_text", "similarity_score", "page_number"}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow(int64(1), int64(1), int64(2), int64(3), "Doc", "text", 0.9, nil))

	sources, err := repo.GetMessageSources(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, sources, 1)
	assert.Equal(t, "Doc", sources[0].DocumentTitle)
}

func TestMessageGetMessageSources_QueryError(t *testing.T) {
	repo, mock := newMsgRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
		WithArgs(int64(1)).
		WillReturnError(fmt.Errorf("query error"))

	_, err := repo.GetMessageSources(context.Background(), 1)
	assert.Error(t, err)
}

func TestMessageGetMessageSources_ScanError(t *testing.T) {
	repo, mock := newMsgRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("invalid"))

	_, err := repo.GetMessageSources(context.Background(), 1)
	assert.Error(t, err)
}

func TestMessageDeleteByConversationID_Success(t *testing.T) {
	repo, mock := newMsgRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM ai_message_sources")).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 5))

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM ai_messages WHERE conversation_id")).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 10))

	err := repo.DeleteByConversationID(context.Background(), 1)
	require.NoError(t, err)
}

func TestMessageDeleteByConversationID_SourceDeleteError(t *testing.T) {
	repo, mock := newMsgRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM ai_message_sources")).
		WithArgs(int64(1)).
		WillReturnError(fmt.Errorf("delete error"))

	err := repo.DeleteByConversationID(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete message sources")
}

func TestMessageDeleteByConversationID_MessageDeleteError(t *testing.T) {
	repo, mock := newMsgRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM ai_message_sources")).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM ai_messages WHERE conversation_id")).
		WithArgs(int64(1)).
		WillReturnError(fmt.Errorf("delete error"))

	err := repo.DeleteByConversationID(context.Background(), 1)
	assert.Error(t, err)
}
