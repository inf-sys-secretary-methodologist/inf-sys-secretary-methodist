package usecases

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/messaging/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/messaging/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/messaging/infrastructure/websocket"
)

// recordingAuditSink captures every LogAuditEvent invocation so the
// emission contract (action / resource / fields shape) can be pinned
// independently of the platform AuditLogger side effects. Mirror к
// the assignments package fakeAuditSink pattern.
type recordingAuditSink struct {
	events []recordedAuditEvent
}

type recordedAuditEvent struct {
	action   string
	resource string
	fields   map[string]any
}

func (r *recordingAuditSink) LogAuditEvent(_ context.Context, action, resource string, fields map[string]any) {
	r.events = append(r.events, recordedAuditEvent{action: action, resource: resource, fields: fields})
}

// newUseCaseWithAudit builds the SUT wired with mock repos + a fresh
// recordingAuditSink so each test can assert against the captured
// events. Keeps every test setup uniform.
func newUseCaseWithAudit(t *testing.T) (*MessagingUseCase, *MockConversationRepository, *MockMessageRepository, *recordingAuditSink) {
	t.Helper()
	mockConvRepo := new(MockConversationRepository)
	mockMsgRepo := new(MockMessageRepository)
	logger := createTestLogger()
	hub := websocket.NewHub(logger)
	sink := &recordingAuditSink{}
	uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil).WithAuditSink(sink)
	return uc, mockConvRepo, mockMsgRepo, sink
}

func TestMessagingUseCase_AuditEmission_CreateDirectConversation(t *testing.T) {
	ctx := context.Background()

	t.Run("emits conversation.created on successful create", func(t *testing.T) {
		uc, convRepo, _, sink := newUseCaseWithAudit(t)

		convRepo.On("GetDirectConversation", mock.Anything, int64(7), int64(42)).Return(nil, nil)
		convRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Conversation")).Return(nil)

		_, err := uc.CreateDirectConversation(ctx, 7, dto.CreateDirectConversationInput{RecipientID: 42})
		assert.NoError(t, err)

		require.Len(t, sink.events, 1, "expected exactly one audit emission")
		assert.Equal(t, "conversation.created", sink.events[0].action)
		assert.Equal(t, "conversation", sink.events[0].resource)
		assert.Equal(t, int64(7), sink.events[0].fields["actor_user_id"])
		assert.Equal(t, "direct", sink.events[0].fields["type"])
		assert.Equal(t, int64(42), sink.events[0].fields["recipient_id"])
	})

	t.Run("does not emit when returning existing direct conversation", func(t *testing.T) {
		uc, convRepo, _, sink := newUseCaseWithAudit(t)

		existing := &entities.Conversation{ID: 99, Type: entities.ConversationTypeDirect}
		convRepo.On("GetDirectConversation", mock.Anything, int64(7), int64(42)).Return(existing, nil)

		_, err := uc.CreateDirectConversation(ctx, 7, dto.CreateDirectConversationInput{RecipientID: 42})
		assert.NoError(t, err)

		assert.Empty(t, sink.events, "no audit event must fire when conversation already existed")
	})
}

func TestMessagingUseCase_AuditEmission_CreateGroupConversation(t *testing.T) {
	ctx := context.Background()

	t.Run("emits conversation.created with type=group and participants count", func(t *testing.T) {
		uc, convRepo, _, sink := newUseCaseWithAudit(t)

		convRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Conversation")).Return(nil)
		// AddParticipant invocations for each participant (including creator):
		convRepo.On("AddParticipant", mock.Anything, mock.AnythingOfType("*entities.Participant")).Return(nil).Maybe()
		// Welcome system message may or may not be created depending on impl path:
		msgRepo := uc.messageRepo.(*MockMessageRepository)
		msgRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Message")).Return(nil).Maybe()

		input := dto.CreateGroupConversationInput{
			Title:          testGroupTitle,
			ParticipantIDs: []int64{42, 99},
		}
		_, err := uc.CreateGroupConversation(ctx, 7, input)
		assert.NoError(t, err)

		require.Len(t, sink.events, 1)
		ev := sink.events[0]
		assert.Equal(t, "conversation.created", ev.action)
		assert.Equal(t, "conversation", ev.resource)
		assert.Equal(t, int64(7), ev.fields["actor_user_id"])
		assert.Equal(t, "group", ev.fields["type"])
		// participants_count includes creator + provided participants
		assert.Equal(t, 3, ev.fields["participants_count"])
	})
}

func TestMessagingUseCase_AuditEmission_UpdateConversation(t *testing.T) {
	ctx := context.Background()

	t.Run("emits conversation.updated when admin edits group conversation", func(t *testing.T) {
		uc, convRepo, _, sink := newUseCaseWithAudit(t)

		conv := &entities.Conversation{
			ID:   1,
			Type: entities.ConversationTypeGroup,
			Participants: []entities.Participant{
				{UserID: 7, Role: entities.ParticipantRoleAdmin},
			},
		}
		convRepo.On("GetByID", mock.Anything, int64(1)).Return(conv, nil)
		convRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.Conversation")).Return(nil)

		newTitle := updatedGroupName
		_, err := uc.UpdateConversation(ctx, 7, 1, dto.UpdateConversationInput{Title: &newTitle})
		assert.NoError(t, err)

		require.Len(t, sink.events, 1)
		ev := sink.events[0]
		assert.Equal(t, "conversation.updated", ev.action)
		assert.Equal(t, "conversation", ev.resource)
		assert.Equal(t, int64(7), ev.fields["actor_user_id"])
		assert.Equal(t, int64(1), ev.fields["conversation_id"])
	})

	t.Run("does not emit when update is denied for non-admin", func(t *testing.T) {
		uc, convRepo, _, sink := newUseCaseWithAudit(t)

		conv := &entities.Conversation{
			ID:   1,
			Type: entities.ConversationTypeGroup,
			Participants: []entities.Participant{
				{UserID: 7, Role: entities.ParticipantRoleMember},
			},
		}
		convRepo.On("GetByID", mock.Anything, int64(1)).Return(conv, nil)

		newTitle := updatedGroupName
		_, err := uc.UpdateConversation(ctx, 7, 1, dto.UpdateConversationInput{Title: &newTitle})
		assert.Error(t, err)
		assert.Empty(t, sink.events, "denied update must not emit audit event")
	})
}

func TestMessagingUseCase_AuditEmission_SendMessage(t *testing.T) {
	ctx := context.Background()

	t.Run("emits message.sent on successful send", func(t *testing.T) {
		uc, convRepo, msgRepo, sink := newUseCaseWithAudit(t)

		participant := &entities.Participant{
			ConversationID: 5,
			UserID:         7,
			Role:           entities.ParticipantRoleMember,
		}
		convRepo.On("GetParticipant", mock.Anything, int64(5), int64(7)).Return(participant, nil)
		msgRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Message")).Return(nil)

		_, err := uc.SendMessage(ctx, 7, 5, dto.SendMessageInput{Content: "Hello"})
		assert.NoError(t, err)

		require.Len(t, sink.events, 1)
		ev := sink.events[0]
		assert.Equal(t, "message.sent", ev.action)
		assert.Equal(t, "message", ev.resource)
		assert.Equal(t, int64(7), ev.fields["actor_user_id"])
		assert.Equal(t, int64(5), ev.fields["conversation_id"])
	})

	t.Run("does not emit when send is denied (not a participant)", func(t *testing.T) {
		uc, convRepo, _, sink := newUseCaseWithAudit(t)

		convRepo.On("GetParticipant", mock.Anything, int64(5), int64(99)).
			Return((*entities.Participant)(nil), entities.ErrNotParticipant)

		_, err := uc.SendMessage(ctx, 99, 5, dto.SendMessageInput{Content: "Hello"})
		assert.Error(t, err)
		assert.Empty(t, sink.events, "denied send must not emit audit event")
	})
}

func TestMessagingUseCase_AuditEmission_DeleteMessage(t *testing.T) {
	ctx := context.Background()

	t.Run("emits message.deleted on successful delete by author", func(t *testing.T) {
		uc, convRepo, msgRepo, sink := newUseCaseWithAudit(t)

		msg := &entities.Message{
			ID:             10,
			ConversationID: 5,
			SenderID:       7,
		}
		conv := &entities.Conversation{
			ID: 5,
			Participants: []entities.Participant{
				{UserID: 7, Role: entities.ParticipantRoleMember},
			},
		}
		msgRepo.On("GetByID", mock.Anything, int64(10)).Return(msg, nil)
		convRepo.On("GetByID", mock.Anything, int64(5)).Return(conv, nil)
		msgRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.Message")).Return(nil)

		err := uc.DeleteMessage(ctx, 7, 10)
		assert.NoError(t, err)

		require.Len(t, sink.events, 1)
		ev := sink.events[0]
		assert.Equal(t, "message.deleted", ev.action)
		assert.Equal(t, "message", ev.resource)
		assert.Equal(t, int64(7), ev.fields["actor_user_id"])
		assert.Equal(t, int64(10), ev.fields["message_id"])
		assert.Equal(t, int64(5), ev.fields["conversation_id"])
	})

	t.Run("does not emit when delete is denied (not author / not admin)", func(t *testing.T) {
		uc, convRepo, msgRepo, sink := newUseCaseWithAudit(t)

		msg := &entities.Message{
			ID:             10,
			ConversationID: 5,
			SenderID:       7,
		}
		conv := &entities.Conversation{
			ID: 5,
			Participants: []entities.Participant{
				{UserID: 99, Role: entities.ParticipantRoleMember},
			},
		}
		msgRepo.On("GetByID", mock.Anything, int64(10)).Return(msg, nil)
		convRepo.On("GetByID", mock.Anything, int64(5)).Return(conv, nil)

		err := uc.DeleteMessage(ctx, 99, 10)
		assert.Error(t, err)
		assert.Empty(t, sink.events, "denied delete must not emit audit event")
	})
}

func TestMessagingUseCase_AuditEmission_NilSinkSilent(t *testing.T) {
	ctx := context.Background()

	t.Run("CreateDirectConversation with nil sink does not panic and silently no-ops", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockMsgRepo := new(MockMessageRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)
		// No WithAuditSink call — sink stays nil (default).
		uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

		mockConvRepo.On("GetDirectConversation", mock.Anything, int64(1), int64(2)).Return(nil, nil)
		mockConvRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Conversation")).Return(nil)

		_, err := uc.CreateDirectConversation(ctx, 1, dto.CreateDirectConversationInput{RecipientID: 2})
		assert.NoError(t, err)
	})
}
