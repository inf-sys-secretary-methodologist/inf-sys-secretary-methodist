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

func newConvRepoMock(t *testing.T) (*ConversationRepositoryPG, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	repo := NewConversationRepositoryPG(db)
	return repo.(*ConversationRepositoryPG), mock
}

var convCols = []string{
	"id", "type", "title", "description", "avatar_url", "created_by", "created_at", "updated_at",
}

var participantCols = []string{
	"id", "conversation_id", "user_id", "role", "last_read_at", "is_muted", "joined_at", "left_at",
	"name", "avatar",
}

// --- Create ---

func TestConversationRepositoryPG_Create_Success(t *testing.T) {
	repo, mock := newConvRepoMock(t)
	conv := entities.NewDirectConversation(1, 2)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO conversations")).
		WithArgs(conv.Type, conv.Title, conv.Description, conv.AvatarURL,
			conv.CreatedBy, conv.CreatedAt, conv.UpdatedAt).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(10)))

	for i := range conv.Participants {
		mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO conversation_participants")).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(i + 1)))
	}
	mock.ExpectCommit()

	err := repo.Create(context.Background(), conv)
	require.NoError(t, err)
	assert.Equal(t, int64(10), conv.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationRepositoryPG_Create_BeginError(t *testing.T) {
	repo, mock := newConvRepoMock(t)
	conv := entities.NewDirectConversation(1, 2)

	mock.ExpectBegin().WillReturnError(sql.ErrConnDone)

	err := repo.Create(context.Background(), conv)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationRepositoryPG_Create_InsertConvError(t *testing.T) {
	repo, mock := newConvRepoMock(t)
	conv := entities.NewDirectConversation(1, 2)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO conversations")).
		WillReturnError(sql.ErrConnDone)
	mock.ExpectRollback()

	err := repo.Create(context.Background(), conv)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationRepositoryPG_Create_ParticipantError(t *testing.T) {
	repo, mock := newConvRepoMock(t)
	conv := entities.NewDirectConversation(1, 2)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO conversations")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(10)))
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO conversation_participants")).
		WillReturnError(sql.ErrConnDone)
	mock.ExpectRollback()

	err := repo.Create(context.Background(), conv)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetByID ---

func TestConversationRepositoryPG_GetByID_Success(t *testing.T) {
	repo, mock := newConvRepoMock(t)
	now := time.Now()
	title := "Test"

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, type, title")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(convCols).AddRow(
			int64(1), entities.ConversationTypeGroup, &title, nil, nil, int64(1), now, now,
		))

	// GetParticipants
	mock.ExpectQuery(regexp.QuoteMeta("SELECT cp.id, cp.conversation_id")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(participantCols).AddRow(
			int64(1), int64(1), int64(1), entities.ParticipantRoleAdmin, nil, false, now, nil, "User1", nil,
		))

	conv, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	require.NotNil(t, conv)
	assert.Equal(t, "Test", *conv.Title)
	assert.Len(t, conv.Participants, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationRepositoryPG_GetByID_NotFound(t *testing.T) {
	repo, mock := newConvRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, type, title")).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	conv, err := repo.GetByID(context.Background(), 999)
	require.Error(t, err)
	assert.ErrorIs(t, err, entities.ErrConversationNotFound)
	assert.Nil(t, conv)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationRepositoryPG_GetByID_ScanError(t *testing.T) {
	repo, mock := newConvRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, type, title")).
		WithArgs(int64(1)).
		WillReturnError(sql.ErrConnDone)

	conv, err := repo.GetByID(context.Background(), 1)
	require.Error(t, err)
	assert.Nil(t, conv)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationRepositoryPG_GetByID_ParticipantsError(t *testing.T) {
	repo, mock := newConvRepoMock(t)
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, type, title")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(convCols).AddRow(
			int64(1), entities.ConversationTypeDirect, nil, nil, nil, int64(1), now, now,
		))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT cp.id, cp.conversation_id")).
		WithArgs(int64(1)).
		WillReturnError(sql.ErrConnDone)

	conv, err := repo.GetByID(context.Background(), 1)
	require.Error(t, err)
	assert.Nil(t, conv)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetDirectConversation ---

func TestConversationRepositoryPG_GetDirectConversation_Found(t *testing.T) {
	repo, mock := newConvRepoMock(t)
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT c.id")).
		WithArgs(int64(1), int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(5)))

	// GetByID is called
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, type, title")).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows(convCols).AddRow(
			int64(5), entities.ConversationTypeDirect, nil, nil, nil, int64(1), now, now,
		))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT cp.id, cp.conversation_id")).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows(participantCols))

	conv, err := repo.GetDirectConversation(context.Background(), 1, 2)
	require.NoError(t, err)
	require.NotNil(t, conv)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationRepositoryPG_GetDirectConversation_NotFound(t *testing.T) {
	repo, mock := newConvRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT c.id")).
		WithArgs(int64(1), int64(2)).
		WillReturnError(sql.ErrNoRows)

	conv, err := repo.GetDirectConversation(context.Background(), 1, 2)
	require.NoError(t, err)
	assert.Nil(t, conv)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationRepositoryPG_GetDirectConversation_Error(t *testing.T) {
	repo, mock := newConvRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT c.id")).
		WithArgs(int64(1), int64(2)).
		WillReturnError(sql.ErrConnDone)

	conv, err := repo.GetDirectConversation(context.Background(), 1, 2)
	require.Error(t, err)
	assert.Nil(t, conv)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- Update ---

func TestConversationRepositoryPG_Update_Success(t *testing.T) {
	repo, mock := newConvRepoMock(t)
	title := "Updated"
	conv := &entities.Conversation{ID: 1, Title: &title}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE conversations")).
		WithArgs(conv.ID, conv.Title, conv.Description, conv.AvatarURL).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), conv)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationRepositoryPG_Update_NotFound(t *testing.T) {
	repo, mock := newConvRepoMock(t)
	conv := &entities.Conversation{ID: 999}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE conversations")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Update(context.Background(), conv)
	require.ErrorIs(t, err, entities.ErrConversationNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationRepositoryPG_Update_Error(t *testing.T) {
	repo, mock := newConvRepoMock(t)
	conv := &entities.Conversation{ID: 1}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE conversations")).
		WillReturnError(sql.ErrConnDone)

	err := repo.Update(context.Background(), conv)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- Delete ---

func TestConversationRepositoryPG_Delete_Success(t *testing.T) {
	repo, mock := newConvRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM conversations WHERE id = $1")).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), 1)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationRepositoryPG_Delete_NotFound(t *testing.T) {
	repo, mock := newConvRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM conversations")).
		WithArgs(int64(999)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Delete(context.Background(), 999)
	require.ErrorIs(t, err, entities.ErrConversationNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationRepositoryPG_Delete_Error(t *testing.T) {
	repo, mock := newConvRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM conversations")).
		WillReturnError(sql.ErrConnDone)

	err := repo.Delete(context.Background(), 1)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- AddParticipant ---

func TestConversationRepositoryPG_AddParticipant_Success(t *testing.T) {
	repo, mock := newConvRepoMock(t)
	p := &entities.Participant{ConversationID: 1, UserID: 5, Role: entities.ParticipantRoleMember, JoinedAt: time.Now()}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO conversation_participants")).
		WithArgs(p.ConversationID, p.UserID, p.Role, p.IsMuted, p.JoinedAt).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	err := repo.AddParticipant(context.Background(), p)
	require.NoError(t, err)
	assert.Equal(t, int64(1), p.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationRepositoryPG_AddParticipant_Error(t *testing.T) {
	repo, mock := newConvRepoMock(t)
	p := &entities.Participant{ConversationID: 1, UserID: 5, JoinedAt: time.Now()}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO conversation_participants")).
		WillReturnError(sql.ErrConnDone)

	err := repo.AddParticipant(context.Background(), p)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- RemoveParticipant ---

func TestConversationRepositoryPG_RemoveParticipant_Success(t *testing.T) {
	repo, mock := newConvRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE conversation_participants")).
		WithArgs(int64(1), int64(2)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.RemoveParticipant(context.Background(), 1, 2)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationRepositoryPG_RemoveParticipant_NotFound(t *testing.T) {
	repo, mock := newConvRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE conversation_participants")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.RemoveParticipant(context.Background(), 1, 999)
	require.ErrorIs(t, err, entities.ErrNotParticipant)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationRepositoryPG_RemoveParticipant_Error(t *testing.T) {
	repo, mock := newConvRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE conversation_participants")).
		WillReturnError(sql.ErrConnDone)

	err := repo.RemoveParticipant(context.Background(), 1, 2)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- UpdateParticipant ---

func TestConversationRepositoryPG_UpdateParticipant_Success(t *testing.T) {
	repo, mock := newConvRepoMock(t)
	p := &entities.Participant{ConversationID: 1, UserID: 2, Role: entities.ParticipantRoleAdmin, IsMuted: true}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE conversation_participants")).
		WithArgs(p.ConversationID, p.UserID, p.Role, p.IsMuted).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdateParticipant(context.Background(), p)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationRepositoryPG_UpdateParticipant_NotFound(t *testing.T) {
	repo, mock := newConvRepoMock(t)
	p := &entities.Participant{ConversationID: 1, UserID: 999}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE conversation_participants")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.UpdateParticipant(context.Background(), p)
	require.ErrorIs(t, err, entities.ErrNotParticipant)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationRepositoryPG_UpdateParticipant_Error(t *testing.T) {
	repo, mock := newConvRepoMock(t)
	p := &entities.Participant{ConversationID: 1, UserID: 2}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE conversation_participants")).
		WillReturnError(sql.ErrConnDone)

	err := repo.UpdateParticipant(context.Background(), p)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetParticipants ---

func TestConversationRepositoryPG_GetParticipants_Success(t *testing.T) {
	repo, mock := newConvRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows(participantCols).
		AddRow(int64(1), int64(1), int64(2), entities.ParticipantRoleMember, nil, false, now, nil, "User", nil)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT cp.id, cp.conversation_id")).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	participants, err := repo.GetParticipants(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, participants, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationRepositoryPG_GetParticipants_Error(t *testing.T) {
	repo, mock := newConvRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT cp.id")).
		WillReturnError(sql.ErrConnDone)

	_, err := repo.GetParticipants(context.Background(), 1)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationRepositoryPG_GetParticipants_ScanError(t *testing.T) {
	repo, mock := newConvRepoMock(t)
	rows := sqlmock.NewRows(participantCols).
		AddRow("bad", int64(1), int64(2), "member", nil, false, time.Now(), nil, "User", nil)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT cp.id")).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	_, err := repo.GetParticipants(context.Background(), 1)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetParticipant ---

func TestConversationRepositoryPG_GetParticipant_Success(t *testing.T) {
	repo, mock := newConvRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows(participantCols).
		AddRow(int64(1), int64(1), int64(2), entities.ParticipantRoleMember, nil, false, now, nil, "User", nil)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT cp.id, cp.conversation_id")).
		WithArgs(int64(1), int64(2)).
		WillReturnRows(rows)

	p, err := repo.GetParticipant(context.Background(), 1, 2)
	require.NoError(t, err)
	require.NotNil(t, p)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationRepositoryPG_GetParticipant_NotFound(t *testing.T) {
	repo, mock := newConvRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT cp.id")).
		WithArgs(int64(1), int64(999)).
		WillReturnError(sql.ErrNoRows)

	p, err := repo.GetParticipant(context.Background(), 1, 999)
	require.ErrorIs(t, err, entities.ErrNotParticipant)
	assert.Nil(t, p)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationRepositoryPG_GetParticipant_Error(t *testing.T) {
	repo, mock := newConvRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT cp.id")).
		WillReturnError(sql.ErrConnDone)

	_, err := repo.GetParticipant(context.Background(), 1, 2)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- UpdateLastRead ---

func TestConversationRepositoryPG_UpdateLastRead_Success(t *testing.T) {
	repo, mock := newConvRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE conversation_participants")).
		WithArgs(int64(1), int64(2), int64(100)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdateLastRead(context.Background(), 1, 2, 100)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationRepositoryPG_UpdateLastRead_Error(t *testing.T) {
	repo, mock := newConvRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE conversation_participants")).
		WillReturnError(sql.ErrConnDone)

	err := repo.UpdateLastRead(context.Background(), 1, 2, 100)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetUnreadCount ---

func TestConversationRepositoryPG_GetUnreadCount_Success(t *testing.T) {
	repo, mock := newConvRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WithArgs(int64(1), int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	count, err := repo.GetUnreadCount(context.Background(), 1, 2)
	require.NoError(t, err)
	assert.Equal(t, 5, count)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationRepositoryPG_GetUnreadCount_Error(t *testing.T) {
	repo, mock := newConvRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnError(sql.ErrConnDone)

	_, err := repo.GetUnreadCount(context.Background(), 1, 2)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- List ---

func TestConversationRepositoryPG_List_Success(t *testing.T) {
	repo, mock := newConvRepoMock(t)
	now := time.Now()
	filter := entities.ConversationFilter{UserID: 1, Limit: 10, Offset: 0}

	// Count
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	// List query
	listCols := []string{
		"id", "type", "title", "description", "avatar_url", "created_by", "created_at", "updated_at",
		"msg_id", "msg_sender_id", "msg_type", "msg_content", "msg_created_at",
		"sender_name", "sender_avatar",
	}
	rows := sqlmock.NewRows(listCols).AddRow(
		int64(1), entities.ConversationTypeDirect, nil, nil, nil, int64(1), now, now,
		nil, nil, nil, nil, nil,
		nil, nil,
	)
	mock.ExpectQuery("SELECT c.id, c.type").WillReturnRows(rows)

	// GetParticipants for each conv
	mock.ExpectQuery(regexp.QuoteMeta("SELECT cp.id")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(participantCols))

	// GetUnreadCount
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	convs, total, err := repo.List(context.Background(), filter)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, convs, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationRepositoryPG_List_WithLastMessage(t *testing.T) {
	repo, mock := newConvRepoMock(t)
	now := time.Now()
	filter := entities.ConversationFilter{UserID: 1, Limit: 10, Offset: 0}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	listCols := []string{
		"id", "type", "title", "description", "avatar_url", "created_by", "created_at", "updated_at",
		"msg_id", "msg_sender_id", "msg_type", "msg_content", "msg_created_at",
		"sender_name", "sender_avatar",
	}
	avatar := "/avatar.jpg"
	rows := sqlmock.NewRows(listCols).AddRow(
		int64(1), entities.ConversationTypeDirect, nil, nil, nil, int64(1), now, now,
		int64(100), int64(2), "text", "hello", now,
		"User2", &avatar,
	)
	mock.ExpectQuery("SELECT c.id, c.type").WillReturnRows(rows)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT cp.id")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(participantCols))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	convs, _, err := repo.List(context.Background(), filter)
	require.NoError(t, err)
	require.Len(t, convs, 1)
	require.NotNil(t, convs[0].LastMessage)
	assert.Equal(t, "hello", convs[0].LastMessage.Content)
	assert.Equal(t, "User2", convs[0].LastMessage.SenderName)
	assert.NotNil(t, convs[0].LastMessage.SenderAvatarURL)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationRepositoryPG_List_WithFilters(t *testing.T) {
	repo, mock := newConvRepoMock(t)
	convType := entities.ConversationTypeGroup
	search := "test"
	filter := entities.ConversationFilter{
		UserID: 1,
		Type:   &convType,
		Search: &search,
		Limit:  10,
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	listCols := []string{
		"id", "type", "title", "description", "avatar_url", "created_by", "created_at", "updated_at",
		"msg_id", "msg_sender_id", "msg_type", "msg_content", "msg_created_at",
		"sender_name", "sender_avatar",
	}
	mock.ExpectQuery("SELECT c.id, c.type").WillReturnRows(sqlmock.NewRows(listCols))

	convs, total, err := repo.List(context.Background(), filter)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, convs)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationRepositoryPG_List_CountError(t *testing.T) {
	repo, mock := newConvRepoMock(t)
	filter := entities.ConversationFilter{UserID: 1, Limit: 10}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnError(sql.ErrConnDone)

	_, _, err := repo.List(context.Background(), filter)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationRepositoryPG_List_QueryError(t *testing.T) {
	repo, mock := newConvRepoMock(t)
	filter := entities.ConversationFilter{UserID: 1, Limit: 10}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	mock.ExpectQuery("SELECT c.id, c.type").
		WillReturnError(sql.ErrConnDone)

	_, _, err := repo.List(context.Background(), filter)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationRepositoryPG_List_ScanError(t *testing.T) {
	repo, mock := newConvRepoMock(t)
	filter := entities.ConversationFilter{UserID: 1, Limit: 10}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	listCols := []string{
		"id", "type", "title", "description", "avatar_url", "created_by", "created_at", "updated_at",
		"msg_id", "msg_sender_id", "msg_type", "msg_content", "msg_created_at",
		"sender_name", "sender_avatar",
	}
	rows := sqlmock.NewRows(listCols).AddRow(
		"bad", "direct", nil, nil, nil, int64(1), time.Now(), time.Now(),
		nil, nil, nil, nil, nil, nil, nil,
	)
	mock.ExpectQuery("SELECT c.id, c.type").WillReturnRows(rows)

	_, _, err := repo.List(context.Background(), filter)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
