package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/messaging/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/messaging/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/messaging/infrastructure/websocket"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

const (
	testGroupTitle   = "Test Group"
	updatedGroupName = "Updated Group"
	testAvatarURL    = "https://example.com/avatar.png"
)

// MockConversationRepository is a mock implementation of ConversationRepository
type MockConversationRepository struct {
	mock.Mock
}

func (m *MockConversationRepository) Create(ctx context.Context, conv *entities.Conversation) error {
	args := m.Called(ctx, conv)
	if args.Get(0) == nil {
		conv.ID = 1
	}
	return args.Error(0)
}

func (m *MockConversationRepository) Update(ctx context.Context, conv *entities.Conversation) error {
	args := m.Called(ctx, conv)
	return args.Error(0)
}

func (m *MockConversationRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockConversationRepository) GetByID(ctx context.Context, id int64) (*entities.Conversation, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Conversation), args.Error(1)
}

func (m *MockConversationRepository) List(ctx context.Context, filter entities.ConversationFilter) ([]*entities.Conversation, int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*entities.Conversation), args.Get(1).(int64), args.Error(2)
}

func (m *MockConversationRepository) GetDirectConversation(ctx context.Context, userID1, userID2 int64) (*entities.Conversation, error) {
	args := m.Called(ctx, userID1, userID2)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Conversation), args.Error(1)
}

func (m *MockConversationRepository) AddParticipant(ctx context.Context, participant *entities.Participant) error {
	args := m.Called(ctx, participant)
	return args.Error(0)
}

func (m *MockConversationRepository) RemoveParticipant(ctx context.Context, conversationID, userID int64) error {
	args := m.Called(ctx, conversationID, userID)
	return args.Error(0)
}

func (m *MockConversationRepository) UpdateParticipant(ctx context.Context, participant *entities.Participant) error {
	args := m.Called(ctx, participant)
	return args.Error(0)
}

func (m *MockConversationRepository) GetParticipants(ctx context.Context, conversationID int64) ([]entities.Participant, error) {
	args := m.Called(ctx, conversationID)
	return args.Get(0).([]entities.Participant), args.Error(1)
}

func (m *MockConversationRepository) GetParticipant(ctx context.Context, conversationID, userID int64) (*entities.Participant, error) {
	args := m.Called(ctx, conversationID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Participant), args.Error(1)
}

func (m *MockConversationRepository) UpdateLastRead(ctx context.Context, conversationID, userID, messageID int64) error {
	args := m.Called(ctx, conversationID, userID, messageID)
	return args.Error(0)
}

func (m *MockConversationRepository) GetUnreadCount(ctx context.Context, conversationID, userID int64) (int, error) {
	args := m.Called(ctx, conversationID, userID)
	return args.Get(0).(int), args.Error(1)
}

// MockMessageRepository is a mock implementation of MessageRepository
type MockMessageRepository struct {
	mock.Mock
}

func (m *MockMessageRepository) Create(ctx context.Context, msg *entities.Message) error {
	args := m.Called(ctx, msg)
	if args.Get(0) == nil {
		msg.ID = 1
	}
	return args.Error(0)
}

func (m *MockMessageRepository) Update(ctx context.Context, msg *entities.Message) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

func (m *MockMessageRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockMessageRepository) GetByID(ctx context.Context, id int64) (*entities.Message, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Message), args.Error(1)
}

func (m *MockMessageRepository) List(ctx context.Context, filter entities.MessageFilter) ([]*entities.Message, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*entities.Message), args.Error(1)
}

func (m *MockMessageRepository) GetLastMessage(ctx context.Context, conversationID int64) (*entities.Message, error) {
	args := m.Called(ctx, conversationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Message), args.Error(1)
}

func (m *MockMessageRepository) CountUnread(ctx context.Context, conversationID, userID int64, lastReadAt *int64) (int, error) {
	args := m.Called(ctx, conversationID, userID, lastReadAt)
	return args.Get(0).(int), args.Error(1)
}

func (m *MockMessageRepository) CreateAttachment(ctx context.Context, attachment *entities.Attachment) error {
	args := m.Called(ctx, attachment)
	return args.Error(0)
}

func (m *MockMessageRepository) GetAttachments(ctx context.Context, messageID int64) ([]entities.Attachment, error) {
	args := m.Called(ctx, messageID)
	return args.Get(0).([]entities.Attachment), args.Error(1)
}

func (m *MockMessageRepository) Search(ctx context.Context, conversationID int64, query string, limit, offset int) ([]*entities.Message, int64, error) {
	args := m.Called(ctx, conversationID, query, limit, offset)
	return args.Get(0).([]*entities.Message), args.Get(1).(int64), args.Error(2)
}

func createTestLogger() *logging.Logger {
	return logging.NewLogger("debug")
}

func TestMessagingUseCase_CreateDirectConversation(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully creates direct conversation", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockMsgRepo := new(MockMessageRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

		input := dto.CreateDirectConversationInput{
			RecipientID: 2,
		}

		mockConvRepo.On("GetDirectConversation", ctx, int64(1), int64(2)).Return(nil, nil)
		mockConvRepo.On("Create", ctx, mock.AnythingOfType("*entities.Conversation")).Return(nil)

		conv, err := uc.CreateDirectConversation(ctx, 1, input)

		assert.NoError(t, err)
		assert.NotNil(t, conv)
		assert.Equal(t, entities.ConversationTypeDirect, conv.Type)
		mockConvRepo.AssertExpectations(t)
	})

	t.Run("returns existing conversation if it exists", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockMsgRepo := new(MockMessageRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

		existingConv := &entities.Conversation{
			ID:   1,
			Type: entities.ConversationTypeDirect,
		}

		input := dto.CreateDirectConversationInput{
			RecipientID: 2,
		}

		mockConvRepo.On("GetDirectConversation", ctx, int64(1), int64(2)).Return(existingConv, nil)

		conv, err := uc.CreateDirectConversation(ctx, 1, input)

		assert.NoError(t, err)
		assert.NotNil(t, conv)
		assert.Equal(t, int64(1), conv.ID)
		mockConvRepo.AssertExpectations(t)
	})
}

func TestMessagingUseCase_CreateGroupConversation(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully creates group conversation", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockMsgRepo := new(MockMessageRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

		description := "Test group description"
		input := dto.CreateGroupConversationInput{
			Title:          testGroupTitle,
			Description:    &description,
			ParticipantIDs: []int64{2, 3, 4},
		}

		mockConvRepo.On("Create", ctx, mock.AnythingOfType("*entities.Conversation")).Return(nil)
		mockMsgRepo.On("Create", ctx, mock.AnythingOfType("*entities.Message")).Return(nil)

		conv, err := uc.CreateGroupConversation(ctx, 1, input)

		assert.NoError(t, err)
		assert.NotNil(t, conv)
		assert.Equal(t, entities.ConversationTypeGroup, conv.Type)
		assert.Equal(t, testGroupTitle, *conv.Title)
		mockConvRepo.AssertExpectations(t)
		mockMsgRepo.AssertExpectations(t)
	})
}

func TestMessagingUseCase_GetConversation(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully gets conversation", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockMsgRepo := new(MockMessageRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

		now := time.Now()
		conv := &entities.Conversation{
			ID:   1,
			Type: entities.ConversationTypeDirect,
			Participants: []entities.Participant{
				{UserID: 1, Role: entities.ParticipantRoleMember},
				{UserID: 2, Role: entities.ParticipantRoleMember},
			},
			CreatedAt: now,
			UpdatedAt: now,
		}

		mockConvRepo.On("GetByID", ctx, int64(1)).Return(conv, nil)
		mockConvRepo.On("GetUnreadCount", ctx, int64(1), int64(1)).Return(5, nil)
		mockMsgRepo.On("GetLastMessage", ctx, int64(1)).Return(nil, nil)

		result, err := uc.GetConversation(ctx, 1, 1)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(1), result.ID)
		assert.Equal(t, 5, result.UnreadCount)
		mockConvRepo.AssertExpectations(t)
		mockMsgRepo.AssertExpectations(t)
	})

	t.Run("returns error when user is not a participant", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockMsgRepo := new(MockMessageRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

		conv := &entities.Conversation{
			ID:   1,
			Type: entities.ConversationTypeDirect,
			Participants: []entities.Participant{
				{UserID: 2, Role: entities.ParticipantRoleMember},
				{UserID: 3, Role: entities.ParticipantRoleMember},
			},
		}

		mockConvRepo.On("GetByID", ctx, int64(1)).Return(conv, nil)

		result, err := uc.GetConversation(ctx, 1, 1)

		assert.Error(t, err)
		assert.Equal(t, entities.ErrNotParticipant, err)
		assert.Nil(t, result)
		mockConvRepo.AssertExpectations(t)
	})
}

func TestMessagingUseCase_ListConversations(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully lists conversations", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockMsgRepo := new(MockMessageRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

		now := time.Now()
		conversations := []*entities.Conversation{
			{ID: 1, Type: entities.ConversationTypeDirect, CreatedAt: now, UpdatedAt: now},
			{ID: 2, Type: entities.ConversationTypeGroup, CreatedAt: now, UpdatedAt: now},
		}

		input := dto.ConversationFilterInput{
			Limit:  20,
			Offset: 0,
		}

		mockConvRepo.On("List", ctx, mock.AnythingOfType("entities.ConversationFilter")).Return(conversations, int64(2), nil)

		result, total, err := uc.ListConversations(ctx, 1, input)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, int64(2), total)
		mockConvRepo.AssertExpectations(t)
	})
}

func TestMessagingUseCase_SendMessage(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully sends message", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockMsgRepo := new(MockMessageRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

		participant := &entities.Participant{
			ConversationID: 1,
			UserID:         1,
			UserName:       "Test User",
			Role:           entities.ParticipantRoleMember,
		}

		input := dto.SendMessageInput{
			Content: "Hello, World!",
		}

		mockConvRepo.On("GetParticipant", ctx, int64(1), int64(1)).Return(participant, nil)
		mockMsgRepo.On("Create", ctx, mock.AnythingOfType("*entities.Message")).Return(nil)

		msg, err := uc.SendMessage(ctx, 1, 1, input)

		assert.NoError(t, err)
		assert.NotNil(t, msg)
		assert.Equal(t, "Hello, World!", msg.Content)
		assert.Equal(t, entities.MessageTypeText, msg.Type)
		mockConvRepo.AssertExpectations(t)
		mockMsgRepo.AssertExpectations(t)
	})

	t.Run("successfully sends reply message", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockMsgRepo := new(MockMessageRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

		participant := &entities.Participant{
			ConversationID: 1,
			UserID:         1,
			UserName:       "Test User",
			Role:           entities.ParticipantRoleMember,
		}

		replyToID := int64(10)
		replyToMsg := &entities.Message{
			ID:      10,
			Content: "Original message",
		}

		input := dto.SendMessageInput{
			Content:   "This is a reply",
			ReplyToID: &replyToID,
		}

		mockConvRepo.On("GetParticipant", ctx, int64(1), int64(1)).Return(participant, nil)
		mockMsgRepo.On("Create", ctx, mock.AnythingOfType("*entities.Message")).Return(nil)
		mockMsgRepo.On("GetByID", ctx, int64(10)).Return(replyToMsg, nil)

		msg, err := uc.SendMessage(ctx, 1, 1, input)

		assert.NoError(t, err)
		assert.NotNil(t, msg)
		assert.Equal(t, "This is a reply", msg.Content)
		assert.NotNil(t, msg.ReplyToID)
		assert.Equal(t, int64(10), *msg.ReplyToID)
		mockConvRepo.AssertExpectations(t)
		mockMsgRepo.AssertExpectations(t)
	})
}

func TestMessagingUseCase_GetMessages(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully gets messages", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockMsgRepo := new(MockMessageRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

		participant := &entities.Participant{
			ConversationID: 1,
			UserID:         1,
			Role:           entities.ParticipantRoleMember,
		}

		now := time.Now()
		messages := []*entities.Message{
			{ID: 1, Content: "Message 1", CreatedAt: now},
			{ID: 2, Content: "Message 2", CreatedAt: now},
		}

		input := dto.MessageFilterInput{
			Limit: 50,
		}

		mockConvRepo.On("GetParticipant", ctx, int64(1), int64(1)).Return(participant, nil)
		mockMsgRepo.On("List", ctx, mock.AnythingOfType("entities.MessageFilter")).Return(messages, nil)

		result, hasMore, err := uc.GetMessages(ctx, 1, 1, input)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.False(t, hasMore)
		mockConvRepo.AssertExpectations(t)
		mockMsgRepo.AssertExpectations(t)
	})

	t.Run("indicates more messages available", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockMsgRepo := new(MockMessageRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

		participant := &entities.Participant{
			ConversationID: 1,
			UserID:         1,
			Role:           entities.ParticipantRoleMember,
		}

		now := time.Now()
		messages := make([]*entities.Message, 3)
		for i := 0; i < 3; i++ {
			messages[i] = &entities.Message{ID: int64(i + 1), Content: "Message", CreatedAt: now}
		}

		input := dto.MessageFilterInput{
			Limit: 2,
		}

		mockConvRepo.On("GetParticipant", ctx, int64(1), int64(1)).Return(participant, nil)
		mockMsgRepo.On("List", ctx, mock.AnythingOfType("entities.MessageFilter")).Return(messages, nil)

		result, hasMore, err := uc.GetMessages(ctx, 1, 1, input)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.True(t, hasMore)
		mockConvRepo.AssertExpectations(t)
		mockMsgRepo.AssertExpectations(t)
	})
}

func TestMessagingUseCase_EditMessage(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully edits message", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockMsgRepo := new(MockMessageRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

		now := time.Now()
		msg := &entities.Message{
			ID:             1,
			ConversationID: 1,
			SenderID:       1,
			Content:        "Original message",
			Type:           entities.MessageTypeText,
			CreatedAt:      now,
		}

		input := dto.EditMessageInput{
			Content: "Edited message",
		}

		mockMsgRepo.On("GetByID", ctx, int64(1)).Return(msg, nil)
		mockMsgRepo.On("Update", ctx, mock.AnythingOfType("*entities.Message")).Return(nil)

		result, err := uc.EditMessage(ctx, 1, 1, input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Edited message", result.Content)
		assert.True(t, result.IsEdited)
		mockMsgRepo.AssertExpectations(t)
	})

	t.Run("fails when user is not the sender", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockMsgRepo := new(MockMessageRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

		now := time.Now()
		msg := &entities.Message{
			ID:             1,
			ConversationID: 1,
			SenderID:       2, // Different user
			Content:        "Original message",
			Type:           entities.MessageTypeText,
			CreatedAt:      now,
		}

		input := dto.EditMessageInput{
			Content: "Edited message",
		}

		mockMsgRepo.On("GetByID", ctx, int64(1)).Return(msg, nil)

		result, err := uc.EditMessage(ctx, 1, 1, input)

		assert.Error(t, err)
		assert.Equal(t, entities.ErrCannotEditMessage, err)
		assert.Nil(t, result)
		mockMsgRepo.AssertExpectations(t)
	})
}

func TestMessagingUseCase_DeleteMessage(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully deletes own message", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockMsgRepo := new(MockMessageRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

		now := time.Now()
		msg := &entities.Message{
			ID:             1,
			ConversationID: 1,
			SenderID:       1,
			Content:        "Message to delete",
			Type:           entities.MessageTypeText,
			CreatedAt:      now,
		}

		conv := &entities.Conversation{
			ID:   1,
			Type: entities.ConversationTypeDirect,
			Participants: []entities.Participant{
				{UserID: 1, Role: entities.ParticipantRoleMember},
				{UserID: 2, Role: entities.ParticipantRoleMember},
			},
		}

		mockMsgRepo.On("GetByID", ctx, int64(1)).Return(msg, nil)
		mockConvRepo.On("GetByID", ctx, int64(1)).Return(conv, nil)
		mockMsgRepo.On("Update", ctx, mock.AnythingOfType("*entities.Message")).Return(nil)

		err := uc.DeleteMessage(ctx, 1, 1)

		assert.NoError(t, err)
		mockMsgRepo.AssertExpectations(t)
		mockConvRepo.AssertExpectations(t)
	})

	t.Run("admin can delete other's message", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockMsgRepo := new(MockMessageRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

		now := time.Now()
		msg := &entities.Message{
			ID:             1,
			ConversationID: 1,
			SenderID:       2, // Different user
			Content:        "Message to delete",
			Type:           entities.MessageTypeText,
			CreatedAt:      now,
		}

		conv := &entities.Conversation{
			ID:   1,
			Type: entities.ConversationTypeGroup,
			Participants: []entities.Participant{
				{UserID: 1, Role: entities.ParticipantRoleAdmin}, // Admin
				{UserID: 2, Role: entities.ParticipantRoleMember},
			},
		}

		mockMsgRepo.On("GetByID", ctx, int64(1)).Return(msg, nil)
		mockConvRepo.On("GetByID", ctx, int64(1)).Return(conv, nil)
		mockMsgRepo.On("Update", ctx, mock.AnythingOfType("*entities.Message")).Return(nil)

		err := uc.DeleteMessage(ctx, 1, 1)

		assert.NoError(t, err)
		mockMsgRepo.AssertExpectations(t)
		mockConvRepo.AssertExpectations(t)
	})

	t.Run("non-admin cannot delete other's message", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockMsgRepo := new(MockMessageRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

		now := time.Now()
		msg := &entities.Message{
			ID:             1,
			ConversationID: 1,
			SenderID:       2, // Different user
			Content:        "Message to delete",
			Type:           entities.MessageTypeText,
			CreatedAt:      now,
		}

		conv := &entities.Conversation{
			ID:   1,
			Type: entities.ConversationTypeGroup,
			Participants: []entities.Participant{
				{UserID: 1, Role: entities.ParticipantRoleMember}, // Not admin
				{UserID: 2, Role: entities.ParticipantRoleMember},
			},
		}

		mockMsgRepo.On("GetByID", ctx, int64(1)).Return(msg, nil)
		mockConvRepo.On("GetByID", ctx, int64(1)).Return(conv, nil)

		err := uc.DeleteMessage(ctx, 1, 1)

		assert.Error(t, err)
		assert.Equal(t, entities.ErrCannotDeleteMessage, err)
		mockMsgRepo.AssertExpectations(t)
		mockConvRepo.AssertExpectations(t)
	})
}

func TestMessagingUseCase_MarkAsRead(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully marks as read", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockMsgRepo := new(MockMessageRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

		participant := &entities.Participant{
			ConversationID: 1,
			UserID:         1,
			Role:           entities.ParticipantRoleMember,
		}

		mockConvRepo.On("GetParticipant", ctx, int64(1), int64(1)).Return(participant, nil)
		mockConvRepo.On("UpdateLastRead", ctx, int64(1), int64(1), int64(10)).Return(nil)

		err := uc.MarkAsRead(ctx, 1, 1, 10)

		assert.NoError(t, err)
		mockConvRepo.AssertExpectations(t)
	})
}

func TestMessagingUseCase_SearchMessages(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully searches messages", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockMsgRepo := new(MockMessageRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

		participant := &entities.Participant{
			ConversationID: 1,
			UserID:         1,
			Role:           entities.ParticipantRoleMember,
		}

		now := time.Now()
		messages := []*entities.Message{
			{ID: 1, Content: "Hello world", CreatedAt: now},
			{ID: 2, Content: "Hello there", CreatedAt: now},
		}

		mockConvRepo.On("GetParticipant", ctx, int64(1), int64(1)).Return(participant, nil)
		mockMsgRepo.On("Search", ctx, int64(1), "Hello", 20, 0).Return(messages, int64(2), nil)

		result, total, err := uc.SearchMessages(ctx, 1, 1, "Hello", 20, 0)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, int64(2), total)
		mockConvRepo.AssertExpectations(t)
		mockMsgRepo.AssertExpectations(t)
	})
}

func TestMessagingUseCase_AddParticipants(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully adds participants to group", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockMsgRepo := new(MockMessageRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

		conv := &entities.Conversation{
			ID:   1,
			Type: entities.ConversationTypeGroup,
			Participants: []entities.Participant{
				{UserID: 1, Role: entities.ParticipantRoleAdmin},
			},
		}

		input := dto.AddParticipantsInput{
			UserIDs: []int64{2, 3},
		}

		mockConvRepo.On("GetByID", ctx, int64(1)).Return(conv, nil)
		mockConvRepo.On("AddParticipant", ctx, mock.AnythingOfType("*entities.Participant")).Return(nil).Times(2)
		mockMsgRepo.On("Create", ctx, mock.AnythingOfType("*entities.Message")).Return(nil).Times(2)

		err := uc.AddParticipants(ctx, 1, 1, input)

		assert.NoError(t, err)
		mockConvRepo.AssertExpectations(t)
		mockMsgRepo.AssertExpectations(t)
	})

	t.Run("fails for direct conversation", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockMsgRepo := new(MockMessageRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

		conv := &entities.Conversation{
			ID:   1,
			Type: entities.ConversationTypeDirect,
		}

		input := dto.AddParticipantsInput{
			UserIDs: []int64{3},
		}

		mockConvRepo.On("GetByID", ctx, int64(1)).Return(conv, nil)

		err := uc.AddParticipants(ctx, 1, 1, input)

		assert.Error(t, err)
		assert.Equal(t, entities.ErrCannotAddToDirectChat, err)
		mockConvRepo.AssertExpectations(t)
	})

	t.Run("fails when not admin", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockMsgRepo := new(MockMessageRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

		conv := &entities.Conversation{
			ID:   1,
			Type: entities.ConversationTypeGroup,
			Participants: []entities.Participant{
				{UserID: 1, Role: entities.ParticipantRoleMember}, // Not admin
			},
		}

		input := dto.AddParticipantsInput{
			UserIDs: []int64{3},
		}

		mockConvRepo.On("GetByID", ctx, int64(1)).Return(conv, nil)

		err := uc.AddParticipants(ctx, 1, 1, input)

		assert.Error(t, err)
		assert.Equal(t, entities.ErrNotParticipant, err)
		mockConvRepo.AssertExpectations(t)
	})
}

func TestMessagingUseCase_UpdateConversation(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully updates group conversation", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockMsgRepo := new(MockMessageRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

		title := testGroupTitle
		conv := &entities.Conversation{
			ID:    1,
			Type:  entities.ConversationTypeGroup,
			Title: &title,
			Participants: []entities.Participant{
				{UserID: 1, Role: entities.ParticipantRoleAdmin},
				{UserID: 2, Role: entities.ParticipantRoleMember},
			},
		}

		newTitle := updatedGroupName
		newDescription := "New description"
		input := dto.UpdateConversationInput{
			Title:       &newTitle,
			Description: &newDescription,
		}

		mockConvRepo.On("GetByID", ctx, int64(1)).Return(conv, nil)
		mockConvRepo.On("Update", ctx, mock.AnythingOfType("*entities.Conversation")).Return(nil)

		result, err := uc.UpdateConversation(ctx, 1, 1, input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, updatedGroupName, *result.Title)
		assert.Equal(t, "New description", *result.Description)
		mockConvRepo.AssertExpectations(t)
	})

	t.Run("non-admin cannot update group conversation", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockMsgRepo := new(MockMessageRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

		title := testGroupTitle
		conv := &entities.Conversation{
			ID:    1,
			Type:  entities.ConversationTypeGroup,
			Title: &title,
			Participants: []entities.Participant{
				{UserID: 1, Role: entities.ParticipantRoleMember}, // Not admin
				{UserID: 2, Role: entities.ParticipantRoleAdmin},
			},
		}

		newTitle := updatedGroupName
		input := dto.UpdateConversationInput{
			Title: &newTitle,
		}

		mockConvRepo.On("GetByID", ctx, int64(1)).Return(conv, nil)

		result, err := uc.UpdateConversation(ctx, 1, 1, input)

		assert.Error(t, err)
		assert.Equal(t, entities.ErrNotParticipant, err)
		assert.Nil(t, result)
		mockConvRepo.AssertExpectations(t)
	})

	t.Run("error when conversation not found", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockMsgRepo := new(MockMessageRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

		newTitle := updatedGroupName
		input := dto.UpdateConversationInput{
			Title: &newTitle,
		}

		mockConvRepo.On("GetByID", ctx, int64(1)).Return(nil, entities.ErrConversationNotFound)

		result, err := uc.UpdateConversation(ctx, 1, 1, input)

		assert.Error(t, err)
		assert.Equal(t, entities.ErrConversationNotFound, err)
		assert.Nil(t, result)
		mockConvRepo.AssertExpectations(t)
	})

	t.Run("updates avatar URL", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockMsgRepo := new(MockMessageRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

		title := testGroupTitle
		conv := &entities.Conversation{
			ID:    1,
			Type:  entities.ConversationTypeGroup,
			Title: &title,
			Participants: []entities.Participant{
				{UserID: 1, Role: entities.ParticipantRoleAdmin},
			},
		}

		newAvatarURL := testAvatarURL
		input := dto.UpdateConversationInput{
			AvatarURL: &newAvatarURL,
		}

		mockConvRepo.On("GetByID", ctx, int64(1)).Return(conv, nil)
		mockConvRepo.On("Update", ctx, mock.AnythingOfType("*entities.Conversation")).Return(nil)

		result, err := uc.UpdateConversation(ctx, 1, 1, input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, testAvatarURL, *result.AvatarURL)
		mockConvRepo.AssertExpectations(t)
	})
}

func TestMessagingUseCase_LeaveConversation(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully leaves group conversation", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockMsgRepo := new(MockMessageRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

		conv := &entities.Conversation{
			ID:   1,
			Type: entities.ConversationTypeGroup,
			Participants: []entities.Participant{
				{UserID: 1, Role: entities.ParticipantRoleMember},
				{UserID: 2, Role: entities.ParticipantRoleAdmin},
			},
		}

		mockConvRepo.On("GetByID", ctx, int64(1)).Return(conv, nil)
		mockConvRepo.On("RemoveParticipant", ctx, int64(1), int64(1)).Return(nil)
		mockMsgRepo.On("Create", ctx, mock.AnythingOfType("*entities.Message")).Return(nil)

		err := uc.LeaveConversation(ctx, 1, 1)

		assert.NoError(t, err)
		mockConvRepo.AssertExpectations(t)
		mockMsgRepo.AssertExpectations(t)
	})

	t.Run("fails for direct conversation", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockMsgRepo := new(MockMessageRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

		conv := &entities.Conversation{
			ID:   1,
			Type: entities.ConversationTypeDirect,
		}

		mockConvRepo.On("GetByID", ctx, int64(1)).Return(conv, nil)

		err := uc.LeaveConversation(ctx, 1, 1)

		assert.Error(t, err)
		assert.Equal(t, entities.ErrCannotLeaveDirectChat, err)
		mockConvRepo.AssertExpectations(t)
	})

	t.Run("fails when last admin tries to leave", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockMsgRepo := new(MockMessageRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

		conv := &entities.Conversation{
			ID:   1,
			Type: entities.ConversationTypeGroup,
			Participants: []entities.Participant{
				{UserID: 1, Role: entities.ParticipantRoleAdmin}, // Last admin
				{UserID: 2, Role: entities.ParticipantRoleMember},
			},
		}

		mockConvRepo.On("GetByID", ctx, int64(1)).Return(conv, nil)

		err := uc.LeaveConversation(ctx, 1, 1)

		assert.Error(t, err)
		assert.Equal(t, entities.ErrCannotRemoveLastAdmin, err)
		mockConvRepo.AssertExpectations(t)
	})

	t.Run("fails when user is not participant", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockMsgRepo := new(MockMessageRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

		conv := &entities.Conversation{
			ID:   1,
			Type: entities.ConversationTypeGroup,
			Participants: []entities.Participant{
				{UserID: 2, Role: entities.ParticipantRoleAdmin},
				{UserID: 3, Role: entities.ParticipantRoleMember},
			},
		}

		mockConvRepo.On("GetByID", ctx, int64(1)).Return(conv, nil)

		err := uc.LeaveConversation(ctx, 1, 1)

		assert.Error(t, err)
		assert.Equal(t, entities.ErrNotParticipant, err)
		mockConvRepo.AssertExpectations(t)
	})
}

func TestMessagingUseCase_SendMessage_ErrorCases(t *testing.T) {
	ctx := context.Background()

	t.Run("fails when user is not participant", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockMsgRepo := new(MockMessageRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

		input := dto.SendMessageInput{
			Content: "Hello, World!",
		}

		mockConvRepo.On("GetParticipant", ctx, int64(1), int64(1)).Return(nil, entities.ErrNotParticipant)

		msg, err := uc.SendMessage(ctx, 1, 1, input)

		assert.Error(t, err)
		assert.Nil(t, msg)
		mockConvRepo.AssertExpectations(t)
	})

	t.Run("sends message with attachments", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockMsgRepo := new(MockMessageRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

		participant := &entities.Participant{
			ConversationID: 1,
			UserID:         1,
			UserName:       "Test User",
			Role:           entities.ParticipantRoleMember,
		}

		input := dto.SendMessageInput{
			Content: "Check this image",
			Attachments: []dto.AttachmentInput{
				{
					FileID:   123,
					FileName: "image.png",
					FileSize: 1024,
					MimeType: "image/png",
					URL:      "https://example.com/image.png",
				},
			},
		}

		mockConvRepo.On("GetParticipant", ctx, int64(1), int64(1)).Return(participant, nil)
		mockMsgRepo.On("Create", ctx, mock.AnythingOfType("*entities.Message")).Return(nil)
		mockMsgRepo.On("CreateAttachment", ctx, mock.AnythingOfType("*entities.Attachment")).Return(nil)

		msg, err := uc.SendMessage(ctx, 1, 1, input)

		assert.NoError(t, err)
		assert.NotNil(t, msg)
		assert.Equal(t, entities.MessageTypeImage, msg.Type)
		assert.Len(t, msg.Attachments, 1)
		mockConvRepo.AssertExpectations(t)
		mockMsgRepo.AssertExpectations(t)
	})

	t.Run("sends message with file attachment", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockMsgRepo := new(MockMessageRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

		participant := &entities.Participant{
			ConversationID: 1,
			UserID:         1,
			UserName:       "Test User",
			Role:           entities.ParticipantRoleMember,
		}

		input := dto.SendMessageInput{
			Content: "Check this file",
			Attachments: []dto.AttachmentInput{
				{
					FileID:   456,
					FileName: "document.pdf",
					FileSize: 2048,
					MimeType: "application/pdf",
					URL:      "https://example.com/document.pdf",
				},
			},
		}

		mockConvRepo.On("GetParticipant", ctx, int64(1), int64(1)).Return(participant, nil)
		mockMsgRepo.On("Create", ctx, mock.AnythingOfType("*entities.Message")).Return(nil)
		mockMsgRepo.On("CreateAttachment", ctx, mock.AnythingOfType("*entities.Attachment")).Return(nil)

		msg, err := uc.SendMessage(ctx, 1, 1, input)

		assert.NoError(t, err)
		assert.NotNil(t, msg)
		assert.Equal(t, entities.MessageTypeFile, msg.Type)
		mockConvRepo.AssertExpectations(t)
		mockMsgRepo.AssertExpectations(t)
	})
}

func TestMessagingUseCase_GetMessages_ErrorCases(t *testing.T) {
	ctx := context.Background()

	t.Run("fails when user is not participant", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockMsgRepo := new(MockMessageRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

		input := dto.MessageFilterInput{
			Limit: 50,
		}

		mockConvRepo.On("GetParticipant", ctx, int64(1), int64(1)).Return(nil, entities.ErrNotParticipant)

		result, hasMore, err := uc.GetMessages(ctx, 1, 1, input)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.False(t, hasMore)
		mockConvRepo.AssertExpectations(t)
	})

	t.Run("uses default limit when not specified", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockMsgRepo := new(MockMessageRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

		participant := &entities.Participant{
			ConversationID: 1,
			UserID:         1,
			Role:           entities.ParticipantRoleMember,
		}

		input := dto.MessageFilterInput{
			Limit: 0, // Should default to 50
		}

		mockConvRepo.On("GetParticipant", ctx, int64(1), int64(1)).Return(participant, nil)
		mockMsgRepo.On("List", ctx, mock.AnythingOfType("entities.MessageFilter")).Return([]*entities.Message{}, nil)

		result, hasMore, err := uc.GetMessages(ctx, 1, 1, input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, hasMore)
		mockConvRepo.AssertExpectations(t)
		mockMsgRepo.AssertExpectations(t)
	})
}

func TestMessagingUseCase_SearchMessages_ErrorCases(t *testing.T) {
	ctx := context.Background()

	t.Run("fails when user is not participant", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockMsgRepo := new(MockMessageRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

		mockConvRepo.On("GetParticipant", ctx, int64(1), int64(1)).Return(nil, entities.ErrNotParticipant)

		result, total, err := uc.SearchMessages(ctx, 1, 1, "test", 20, 0)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, int64(0), total)
		mockConvRepo.AssertExpectations(t)
	})

	t.Run("uses default limit when not specified", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockMsgRepo := new(MockMessageRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

		participant := &entities.Participant{
			ConversationID: 1,
			UserID:         1,
			Role:           entities.ParticipantRoleMember,
		}

		mockConvRepo.On("GetParticipant", ctx, int64(1), int64(1)).Return(participant, nil)
		mockMsgRepo.On("Search", ctx, int64(1), "test", 20, 0).Return([]*entities.Message{}, int64(0), nil)

		result, total, err := uc.SearchMessages(ctx, 1, 1, "test", 0, 0) // Should use default 20

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(0), total)
		mockConvRepo.AssertExpectations(t)
		mockMsgRepo.AssertExpectations(t)
	})
}

func TestMessagingUseCase_MarkAsRead_ErrorCases(t *testing.T) {
	ctx := context.Background()

	t.Run("fails when user is not participant", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockMsgRepo := new(MockMessageRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

		mockConvRepo.On("GetParticipant", ctx, int64(1), int64(1)).Return(nil, entities.ErrNotParticipant)

		err := uc.MarkAsRead(ctx, 1, 1, 10)

		assert.Error(t, err)
		mockConvRepo.AssertExpectations(t)
	})
}

func TestMessagingUseCase_ListConversations_WithFilters(t *testing.T) {
	ctx := context.Background()

	t.Run("lists with type filter", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockMsgRepo := new(MockMessageRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

		now := time.Now()
		conversations := []*entities.Conversation{
			{ID: 1, Type: entities.ConversationTypeGroup, CreatedAt: now, UpdatedAt: now},
		}

		convType := string(entities.ConversationTypeGroup)
		input := dto.ConversationFilterInput{
			Type:   &convType,
			Limit:  20,
			Offset: 0,
		}

		mockConvRepo.On("List", ctx, mock.AnythingOfType("entities.ConversationFilter")).Return(conversations, int64(1), nil)

		result, total, err := uc.ListConversations(ctx, 1, input)

		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, int64(1), total)
		mockConvRepo.AssertExpectations(t)
	})

	t.Run("uses default limit when not specified", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockMsgRepo := new(MockMessageRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

		input := dto.ConversationFilterInput{
			Limit: 0, // Should default to 20
		}

		mockConvRepo.On("List", ctx, mock.AnythingOfType("entities.ConversationFilter")).Return([]*entities.Conversation{}, int64(0), nil)

		result, total, err := uc.ListConversations(ctx, 1, input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(0), total)
		mockConvRepo.AssertExpectations(t)
	})
}

func TestMessagingUseCase_EditMessage_ErrorCases(t *testing.T) {
	ctx := context.Background()

	t.Run("fails when message not found", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockMsgRepo := new(MockMessageRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

		input := dto.EditMessageInput{
			Content: "Edited message",
		}

		mockMsgRepo.On("GetByID", ctx, int64(1)).Return(nil, entities.ErrMessageNotFound)

		result, err := uc.EditMessage(ctx, 1, 1, input)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockMsgRepo.AssertExpectations(t)
	})
}

func TestMessagingUseCase_DeleteMessage_ErrorCases(t *testing.T) {
	ctx := context.Background()

	t.Run("fails when message not found", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockMsgRepo := new(MockMessageRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

		mockMsgRepo.On("GetByID", ctx, int64(1)).Return(nil, entities.ErrMessageNotFound)

		err := uc.DeleteMessage(ctx, 1, 1)

		assert.Error(t, err)
		mockMsgRepo.AssertExpectations(t)
	})
}

func TestMessagingUseCase_CreateDirectConversation_Errors(t *testing.T) {
	ctx := context.Background()

	t.Run("error when checking existing conversation", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockMsgRepo := new(MockMessageRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

		input := dto.CreateDirectConversationInput{
			RecipientID: 2,
		}

		mockConvRepo.On("GetDirectConversation", ctx, int64(1), int64(2)).Return(nil, assert.AnError)

		conv, err := uc.CreateDirectConversation(ctx, 1, input)

		assert.Error(t, err)
		assert.Nil(t, conv)
		assert.Contains(t, err.Error(), "failed to check existing conversation")
		mockConvRepo.AssertExpectations(t)
	})
}

// MockMessageNotifier is a mock implementation of MessageNotifier
type MockMessageNotifier struct {
	mock.Mock
}

func (m *MockMessageNotifier) NotifyNewMessage(ctx context.Context, userID int64, senderName, content string, conversationID, messageID int64) error {
	args := m.Called(ctx, userID, senderName, content, conversationID, messageID)
	return args.Error(0)
}

func TestResolveAvatarURL(t *testing.T) {
	logger := createTestLogger()
	hub := websocket.NewHub(logger)

	t.Run("nil path returns nil", func(t *testing.T) {
		uc := NewMessagingUseCase(nil, nil, hub, logger, nil, nil)
		result := uc.resolveAvatarURL(context.Background(), nil)
		assert.Nil(t, result)
	})

	t.Run("empty path returns nil", func(t *testing.T) {
		uc := NewMessagingUseCase(nil, nil, hub, logger, nil, nil)
		empty := ""
		result := uc.resolveAvatarURL(context.Background(), &empty)
		assert.Nil(t, result)
	})

	t.Run("http URL returned as-is", func(t *testing.T) {
		uc := NewMessagingUseCase(nil, nil, hub, logger, nil, nil)
		url := "http://example.com/avatar.png"
		result := uc.resolveAvatarURL(context.Background(), &url)
		assert.NotNil(t, result)
		assert.Equal(t, url, *result)
	})

	t.Run("https URL returned as-is", func(t *testing.T) {
		uc := NewMessagingUseCase(nil, nil, hub, logger, nil, nil)
		url := testAvatarURL
		result := uc.resolveAvatarURL(context.Background(), &url)
		assert.NotNil(t, result)
		assert.Equal(t, url, *result)
	})

	t.Run("storage path with nil s3Client returns nil", func(t *testing.T) {
		uc := NewMessagingUseCase(nil, nil, hub, logger, nil, nil)
		path := "avatars/user1.png"
		result := uc.resolveAvatarURL(context.Background(), &path)
		assert.Nil(t, result)
	})
}

func TestResolveConversationAvatars(t *testing.T) {
	logger := createTestLogger()
	hub := websocket.NewHub(logger)
	uc := NewMessagingUseCase(nil, nil, hub, logger, nil, nil)

	t.Run("nil conversation does not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			uc.resolveConversationAvatars(context.Background(), nil)
		})
	})

	t.Run("resolves participant avatars", func(t *testing.T) {
		httpURL := testAvatarURL
		conv := &entities.Conversation{
			AvatarURL: nil,
			Participants: []entities.Participant{
				{UserID: 1, UserAvatarURL: &httpURL},
				{UserID: 2, UserAvatarURL: nil},
			},
		}
		uc.resolveConversationAvatars(context.Background(), conv)
		assert.NotNil(t, conv.Participants[0].UserAvatarURL)
		assert.Equal(t, httpURL, *conv.Participants[0].UserAvatarURL)
		assert.Nil(t, conv.Participants[1].UserAvatarURL)
	})

	t.Run("resolves last message sender avatar", func(t *testing.T) {
		senderURL := "https://example.com/sender.png"
		conv := &entities.Conversation{
			LastMessage: &entities.Message{
				SenderAvatarURL: &senderURL,
			},
		}
		uc.resolveConversationAvatars(context.Background(), conv)
		assert.NotNil(t, conv.LastMessage.SenderAvatarURL)
		assert.Equal(t, senderURL, *conv.LastMessage.SenderAvatarURL)
	})

	t.Run("nil last message does not panic", func(t *testing.T) {
		conv := &entities.Conversation{
			LastMessage: nil,
		}
		assert.NotPanics(t, func() {
			uc.resolveConversationAvatars(context.Background(), conv)
		})
	})
}

func TestNotifyParticipants(t *testing.T) {
	ctx := context.Background()

	t.Run("sends notification to offline participants", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockNotifier := new(MockMessageNotifier)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, nil, hub, logger, mockNotifier, nil)

		mockConvRepo.On("GetParticipants", ctx, int64(1)).Return([]entities.Participant{
			{UserID: 10}, // sender, should be skipped
			{UserID: 20}, // offline participant
			{UserID: 30}, // offline participant
		}, nil)

		mockNotifier.On("NotifyNewMessage", ctx, int64(20), "Sender", "Hello", int64(1), int64(100)).Return(nil)
		mockNotifier.On("NotifyNewMessage", ctx, int64(30), "Sender", "Hello", int64(1), int64(100)).Return(nil)

		msg := &entities.Message{
			ID:         100,
			Content:    "Hello",
			SenderName: "Sender",
		}

		uc.notifyParticipants(ctx, 1, 10, msg)

		mockConvRepo.AssertExpectations(t)
		mockNotifier.AssertExpectations(t)
	})

	t.Run("skips left participants", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockNotifier := new(MockMessageNotifier)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, nil, hub, logger, mockNotifier, nil)

		leftAt := time.Now()
		mockConvRepo.On("GetParticipants", ctx, int64(1)).Return([]entities.Participant{
			{UserID: 10},                  // sender
			{UserID: 20, LeftAt: &leftAt}, // left participant, should be skipped
		}, nil)

		msg := &entities.Message{
			ID:         100,
			Content:    "Hello",
			SenderName: "Sender",
		}

		uc.notifyParticipants(ctx, 1, 10, msg)

		mockConvRepo.AssertExpectations(t)
		// No notification should be sent
		mockNotifier.AssertNotCalled(t, "NotifyNewMessage")
	})

	t.Run("handles GetParticipants error", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, nil, hub, logger, nil, nil)

		mockConvRepo.On("GetParticipants", ctx, int64(1)).Return([]entities.Participant{}, errors.New("db error"))

		msg := &entities.Message{ID: 1, Content: "test", SenderName: "S"}
		assert.NotPanics(t, func() {
			uc.notifyParticipants(ctx, 1, 10, msg)
		})
	})

	t.Run("truncates long content", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockNotifier := new(MockMessageNotifier)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, nil, hub, logger, mockNotifier, nil)

		longContent := ""
		for i := 0; i < 120; i++ {
			longContent += "x"
		}

		mockConvRepo.On("GetParticipants", ctx, int64(1)).Return([]entities.Participant{
			{UserID: 10}, // sender
			{UserID: 20}, // offline
		}, nil)

		truncated := longContent[:97] + "..."
		mockNotifier.On("NotifyNewMessage", ctx, int64(20), "Sender", truncated, int64(1), int64(100)).Return(nil)

		msg := &entities.Message{
			ID:         100,
			Content:    longContent,
			SenderName: "Sender",
		}

		uc.notifyParticipants(ctx, 1, 10, msg)

		mockConvRepo.AssertExpectations(t)
		mockNotifier.AssertExpectations(t)
	})

	t.Run("handles notification error", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockNotifier := new(MockMessageNotifier)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, nil, hub, logger, mockNotifier, nil)

		mockConvRepo.On("GetParticipants", ctx, int64(1)).Return([]entities.Participant{
			{UserID: 10},
			{UserID: 20},
		}, nil)

		mockNotifier.On("NotifyNewMessage", ctx, int64(20), "Sender", "Hi", int64(1), int64(1)).Return(errors.New("send failed"))

		msg := &entities.Message{ID: 1, Content: "Hi", SenderName: "Sender"}

		assert.NotPanics(t, func() {
			uc.notifyParticipants(ctx, 1, 10, msg)
		})

		mockNotifier.AssertExpectations(t)
	})
}

func TestSearchMessages(t *testing.T) {
	ctx := context.Background()

	t.Run("negative limit defaults to 50", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		mockMsgRepo := new(MockMessageRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)

		uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

		mockConvRepo.On("GetParticipant", ctx, int64(1), int64(10)).Return(&entities.Participant{UserID: 10}, nil)
		mockMsgRepo.On("Search", ctx, int64(1), "test", 20, 0).Return([]*entities.Message{}, int64(0), nil)

		messages, total, err := uc.SearchMessages(ctx, 10, 1, "test", -1, 0)
		assert.NoError(t, err)
		assert.NotNil(t, messages)
		assert.Equal(t, int64(0), total)
	})
}

func TestCreateGroupConversation_SystemMessageError(t *testing.T) {
	ctx := context.Background()
	mockConvRepo := new(MockConversationRepository)
	mockMsgRepo := new(MockMessageRepository)
	logger := createTestLogger()
	hub := websocket.NewHub(logger)

	uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

	mockConvRepo.On("Create", ctx, mock.AnythingOfType("*entities.Conversation")).Return(nil)
	mockMsgRepo.On("Create", ctx, mock.AnythingOfType("*entities.Message")).Return(errors.New("msg error"))

	conv, err := uc.CreateGroupConversation(ctx, 1, dto.CreateGroupConversationInput{
		Title:          "Test",
		ParticipantIDs: []int64{1, 2},
	})

	assert.NoError(t, err) // should not fail even if system message fails
	assert.NotNil(t, conv)
}

func TestGetConversation_Errors(t *testing.T) {
	ctx := context.Background()

	t.Run("GetByID error", func(t *testing.T) {
		mockConvRepo := new(MockConversationRepository)
		logger := createTestLogger()
		hub := websocket.NewHub(logger)
		uc := NewMessagingUseCase(mockConvRepo, nil, hub, logger, nil, nil)

		mockConvRepo.On("GetByID", ctx, int64(1)).Return(nil, errors.New("not found"))

		conv, err := uc.GetConversation(ctx, 10, 1)
		assert.Error(t, err)
		assert.Nil(t, conv)
	})
}

func TestMarkAsRead_Error(t *testing.T) {
	ctx := context.Background()
	mockConvRepo := new(MockConversationRepository)
	logger := createTestLogger()
	hub := websocket.NewHub(logger)
	uc := NewMessagingUseCase(mockConvRepo, nil, hub, logger, nil, nil)

	mockConvRepo.On("GetParticipant", ctx, int64(1), int64(10)).Return(&entities.Participant{UserID: 10}, nil)
	mockConvRepo.On("UpdateLastRead", ctx, int64(1), int64(10), int64(5)).Return(errors.New("db error"))

	err := uc.MarkAsRead(ctx, 10, 1, 5)
	assert.Error(t, err)
}

// --- Additional tests for uncovered branches ---

func TestCreateDirectConversation_CreateError(t *testing.T) {
	ctx := context.Background()
	mockConvRepo := new(MockConversationRepository)
	mockMsgRepo := new(MockMessageRepository)
	logger := createTestLogger()
	hub := websocket.NewHub(logger)

	uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

	input := dto.CreateDirectConversationInput{RecipientID: 2}

	mockConvRepo.On("GetDirectConversation", ctx, int64(1), int64(2)).Return(nil, nil)
	mockConvRepo.On("Create", ctx, mock.AnythingOfType("*entities.Conversation")).Return(errors.New("db error"))

	conv, err := uc.CreateDirectConversation(ctx, 1, input)

	assert.Error(t, err)
	assert.Nil(t, conv)
	assert.Contains(t, err.Error(), "failed to create conversation")
	mockConvRepo.AssertExpectations(t)
}

func TestCreateGroupConversation_CreateError(t *testing.T) {
	ctx := context.Background()
	mockConvRepo := new(MockConversationRepository)
	mockMsgRepo := new(MockMessageRepository)
	logger := createTestLogger()
	hub := websocket.NewHub(logger)

	uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

	input := dto.CreateGroupConversationInput{
		Title:          "Group",
		ParticipantIDs: []int64{2},
	}

	mockConvRepo.On("Create", ctx, mock.AnythingOfType("*entities.Conversation")).Return(errors.New("db error"))

	conv, err := uc.CreateGroupConversation(ctx, 1, input)

	assert.Error(t, err)
	assert.Nil(t, conv)
	assert.Contains(t, err.Error(), "failed to create group conversation")
	mockConvRepo.AssertExpectations(t)
}

func TestGetConversation_UnreadCountError(t *testing.T) {
	ctx := context.Background()
	mockConvRepo := new(MockConversationRepository)
	mockMsgRepo := new(MockMessageRepository)
	logger := createTestLogger()
	hub := websocket.NewHub(logger)

	uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

	conv := &entities.Conversation{
		ID:   1,
		Type: entities.ConversationTypeDirect,
		Participants: []entities.Participant{
			{UserID: 1, Role: entities.ParticipantRoleMember},
			{UserID: 2, Role: entities.ParticipantRoleMember},
		},
	}

	mockConvRepo.On("GetByID", ctx, int64(1)).Return(conv, nil)
	mockConvRepo.On("GetUnreadCount", ctx, int64(1), int64(1)).Return(0, errors.New("db error"))
	mockMsgRepo.On("GetLastMessage", ctx, int64(1)).Return(nil, nil)

	result, err := uc.GetConversation(ctx, 1, 1)

	assert.NoError(t, err) // should not fail, just log
	assert.NotNil(t, result)
	assert.Equal(t, 0, result.UnreadCount)
	mockConvRepo.AssertExpectations(t)
}

func TestGetConversation_LastMessageError(t *testing.T) {
	ctx := context.Background()
	mockConvRepo := new(MockConversationRepository)
	mockMsgRepo := new(MockMessageRepository)
	logger := createTestLogger()
	hub := websocket.NewHub(logger)

	uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

	conv := &entities.Conversation{
		ID:   1,
		Type: entities.ConversationTypeDirect,
		Participants: []entities.Participant{
			{UserID: 1, Role: entities.ParticipantRoleMember},
			{UserID: 2, Role: entities.ParticipantRoleMember},
		},
	}

	mockConvRepo.On("GetByID", ctx, int64(1)).Return(conv, nil)
	mockConvRepo.On("GetUnreadCount", ctx, int64(1), int64(1)).Return(3, nil)
	mockMsgRepo.On("GetLastMessage", ctx, int64(1)).Return(nil, errors.New("db error"))

	result, err := uc.GetConversation(ctx, 1, 1)

	assert.NoError(t, err) // should not fail, just log
	assert.NotNil(t, result)
	assert.Nil(t, result.LastMessage)
	mockConvRepo.AssertExpectations(t)
	mockMsgRepo.AssertExpectations(t)
}

func TestGetConversation_WithLastMessage(t *testing.T) {
	ctx := context.Background()
	mockConvRepo := new(MockConversationRepository)
	mockMsgRepo := new(MockMessageRepository)
	logger := createTestLogger()
	hub := websocket.NewHub(logger)

	uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

	conv := &entities.Conversation{
		ID:   1,
		Type: entities.ConversationTypeDirect,
		Participants: []entities.Participant{
			{UserID: 1, Role: entities.ParticipantRoleMember},
			{UserID: 2, Role: entities.ParticipantRoleMember},
		},
	}

	lastMsg := &entities.Message{
		ID:      10,
		Content: "Last message",
	}

	mockConvRepo.On("GetByID", ctx, int64(1)).Return(conv, nil)
	mockConvRepo.On("GetUnreadCount", ctx, int64(1), int64(1)).Return(0, nil)
	mockMsgRepo.On("GetLastMessage", ctx, int64(1)).Return(lastMsg, nil)

	result, err := uc.GetConversation(ctx, 1, 1)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.LastMessage)
	assert.Equal(t, int64(10), result.LastMessage.ID)
	mockConvRepo.AssertExpectations(t)
	mockMsgRepo.AssertExpectations(t)
}

func TestListConversations_ListError(t *testing.T) {
	ctx := context.Background()
	mockConvRepo := new(MockConversationRepository)
	mockMsgRepo := new(MockMessageRepository)
	logger := createTestLogger()
	hub := websocket.NewHub(logger)

	uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

	input := dto.ConversationFilterInput{Limit: 20}

	mockConvRepo.On("List", ctx, mock.AnythingOfType("entities.ConversationFilter")).Return([]*entities.Conversation{}, int64(0), errors.New("db error"))

	result, total, err := uc.ListConversations(ctx, 1, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, int64(0), total)
	mockConvRepo.AssertExpectations(t)
}

func TestUpdateConversation_UpdateRepoError(t *testing.T) {
	ctx := context.Background()
	mockConvRepo := new(MockConversationRepository)
	mockMsgRepo := new(MockMessageRepository)
	logger := createTestLogger()
	hub := websocket.NewHub(logger)

	uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

	title := testGroupTitle
	conv := &entities.Conversation{
		ID:    1,
		Type:  entities.ConversationTypeGroup,
		Title: &title,
		Participants: []entities.Participant{
			{UserID: 1, Role: entities.ParticipantRoleAdmin},
		},
	}

	newTitle := "New Title"
	input := dto.UpdateConversationInput{Title: &newTitle}

	mockConvRepo.On("GetByID", ctx, int64(1)).Return(conv, nil)
	mockConvRepo.On("Update", ctx, mock.AnythingOfType("*entities.Conversation")).Return(errors.New("db error"))

	result, err := uc.UpdateConversation(ctx, 1, 1, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to update conversation")
	mockConvRepo.AssertExpectations(t)
}

func TestUpdateConversation_DirectConversationNonAdmin(t *testing.T) {
	ctx := context.Background()
	mockConvRepo := new(MockConversationRepository)
	mockMsgRepo := new(MockMessageRepository)
	logger := createTestLogger()
	hub := websocket.NewHub(logger)

	uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

	// Direct conversations do not check admin, so update should succeed
	conv := &entities.Conversation{
		ID:   1,
		Type: entities.ConversationTypeDirect,
		Participants: []entities.Participant{
			{UserID: 1, Role: entities.ParticipantRoleMember},
			{UserID: 2, Role: entities.ParticipantRoleMember},
		},
	}

	newTitle := "DM Title"
	input := dto.UpdateConversationInput{Title: &newTitle}

	mockConvRepo.On("GetByID", ctx, int64(1)).Return(conv, nil)
	mockConvRepo.On("Update", ctx, mock.AnythingOfType("*entities.Conversation")).Return(nil)

	result, err := uc.UpdateConversation(ctx, 1, 1, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "DM Title", *result.Title)
	mockConvRepo.AssertExpectations(t)
}

func TestAddParticipants_SkipsExistingParticipant(t *testing.T) {
	ctx := context.Background()
	mockConvRepo := new(MockConversationRepository)
	mockMsgRepo := new(MockMessageRepository)
	logger := createTestLogger()
	hub := websocket.NewHub(logger)

	uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

	conv := &entities.Conversation{
		ID:   1,
		Type: entities.ConversationTypeGroup,
		Participants: []entities.Participant{
			{UserID: 1, Role: entities.ParticipantRoleAdmin},
			{UserID: 2, Role: entities.ParticipantRoleMember},
		},
	}

	// Try to add user 2 (already a participant) and user 3 (new)
	input := dto.AddParticipantsInput{UserIDs: []int64{2, 3}}

	mockConvRepo.On("GetByID", ctx, int64(1)).Return(conv, nil)
	// Only user 3 should be added
	mockConvRepo.On("AddParticipant", ctx, mock.AnythingOfType("*entities.Participant")).Return(nil).Once()
	mockMsgRepo.On("Create", ctx, mock.AnythingOfType("*entities.Message")).Return(nil).Once()

	err := uc.AddParticipants(ctx, 1, 1, input)

	assert.NoError(t, err)
	mockConvRepo.AssertExpectations(t)
	mockMsgRepo.AssertExpectations(t)
}

func TestAddParticipants_AddParticipantError(t *testing.T) {
	ctx := context.Background()
	mockConvRepo := new(MockConversationRepository)
	mockMsgRepo := new(MockMessageRepository)
	logger := createTestLogger()
	hub := websocket.NewHub(logger)

	uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

	conv := &entities.Conversation{
		ID:   1,
		Type: entities.ConversationTypeGroup,
		Participants: []entities.Participant{
			{UserID: 1, Role: entities.ParticipantRoleAdmin},
		},
	}

	input := dto.AddParticipantsInput{UserIDs: []int64{5}}

	mockConvRepo.On("GetByID", ctx, int64(1)).Return(conv, nil)
	mockConvRepo.On("AddParticipant", ctx, mock.AnythingOfType("*entities.Participant")).Return(errors.New("db error"))

	err := uc.AddParticipants(ctx, 1, 1, input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to add participant")
	mockConvRepo.AssertExpectations(t)
}

func TestAddParticipants_SystemMessageError(t *testing.T) {
	ctx := context.Background()
	mockConvRepo := new(MockConversationRepository)
	mockMsgRepo := new(MockMessageRepository)
	logger := createTestLogger()
	hub := websocket.NewHub(logger)

	uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

	conv := &entities.Conversation{
		ID:   1,
		Type: entities.ConversationTypeGroup,
		Participants: []entities.Participant{
			{UserID: 1, Role: entities.ParticipantRoleAdmin},
		},
	}

	input := dto.AddParticipantsInput{UserIDs: []int64{5}}

	mockConvRepo.On("GetByID", ctx, int64(1)).Return(conv, nil)
	mockConvRepo.On("AddParticipant", ctx, mock.AnythingOfType("*entities.Participant")).Return(nil)
	mockMsgRepo.On("Create", ctx, mock.AnythingOfType("*entities.Message")).Return(errors.New("msg error"))

	err := uc.AddParticipants(ctx, 1, 1, input)

	assert.NoError(t, err) // system message error is logged, not returned
	mockConvRepo.AssertExpectations(t)
	mockMsgRepo.AssertExpectations(t)
}

func TestAddParticipants_GetByIDError(t *testing.T) {
	ctx := context.Background()
	mockConvRepo := new(MockConversationRepository)
	mockMsgRepo := new(MockMessageRepository)
	logger := createTestLogger()
	hub := websocket.NewHub(logger)

	uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

	input := dto.AddParticipantsInput{UserIDs: []int64{5}}

	mockConvRepo.On("GetByID", ctx, int64(1)).Return(nil, errors.New("not found"))

	err := uc.AddParticipants(ctx, 1, 1, input)

	assert.Error(t, err)
	mockConvRepo.AssertExpectations(t)
}

func TestLeaveConversation_GetByIDError(t *testing.T) {
	ctx := context.Background()
	mockConvRepo := new(MockConversationRepository)
	mockMsgRepo := new(MockMessageRepository)
	logger := createTestLogger()
	hub := websocket.NewHub(logger)

	uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

	mockConvRepo.On("GetByID", ctx, int64(1)).Return(nil, errors.New("not found"))

	err := uc.LeaveConversation(ctx, 1, 1)

	assert.Error(t, err)
	mockConvRepo.AssertExpectations(t)
}

func TestLeaveConversation_RemoveParticipantError(t *testing.T) {
	ctx := context.Background()
	mockConvRepo := new(MockConversationRepository)
	mockMsgRepo := new(MockMessageRepository)
	logger := createTestLogger()
	hub := websocket.NewHub(logger)

	uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

	conv := &entities.Conversation{
		ID:   1,
		Type: entities.ConversationTypeGroup,
		Participants: []entities.Participant{
			{UserID: 1, Role: entities.ParticipantRoleMember},
			{UserID: 2, Role: entities.ParticipantRoleAdmin},
		},
	}

	mockConvRepo.On("GetByID", ctx, int64(1)).Return(conv, nil)
	mockConvRepo.On("RemoveParticipant", ctx, int64(1), int64(1)).Return(errors.New("db error"))

	err := uc.LeaveConversation(ctx, 1, 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to leave conversation")
	mockConvRepo.AssertExpectations(t)
}

func TestLeaveConversation_SystemMessageError(t *testing.T) {
	ctx := context.Background()
	mockConvRepo := new(MockConversationRepository)
	mockMsgRepo := new(MockMessageRepository)
	logger := createTestLogger()
	hub := websocket.NewHub(logger)

	uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

	conv := &entities.Conversation{
		ID:   1,
		Type: entities.ConversationTypeGroup,
		Participants: []entities.Participant{
			{UserID: 1, Role: entities.ParticipantRoleMember},
			{UserID: 2, Role: entities.ParticipantRoleAdmin},
		},
	}

	mockConvRepo.On("GetByID", ctx, int64(1)).Return(conv, nil)
	mockConvRepo.On("RemoveParticipant", ctx, int64(1), int64(1)).Return(nil)
	mockMsgRepo.On("Create", ctx, mock.AnythingOfType("*entities.Message")).Return(errors.New("msg error"))

	err := uc.LeaveConversation(ctx, 1, 1)

	assert.NoError(t, err) // system message error is logged, not returned
	mockConvRepo.AssertExpectations(t)
	mockMsgRepo.AssertExpectations(t)
}

func TestLeaveConversation_AdminWithMultipleAdmins(t *testing.T) {
	ctx := context.Background()
	mockConvRepo := new(MockConversationRepository)
	mockMsgRepo := new(MockMessageRepository)
	logger := createTestLogger()
	hub := websocket.NewHub(logger)

	uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

	conv := &entities.Conversation{
		ID:   1,
		Type: entities.ConversationTypeGroup,
		Participants: []entities.Participant{
			{UserID: 1, Role: entities.ParticipantRoleAdmin},
			{UserID: 2, Role: entities.ParticipantRoleAdmin}, // Another admin
			{UserID: 3, Role: entities.ParticipantRoleMember},
		},
	}

	mockConvRepo.On("GetByID", ctx, int64(1)).Return(conv, nil)
	mockConvRepo.On("RemoveParticipant", ctx, int64(1), int64(1)).Return(nil)
	mockMsgRepo.On("Create", ctx, mock.AnythingOfType("*entities.Message")).Return(nil)

	err := uc.LeaveConversation(ctx, 1, 1)

	assert.NoError(t, err) // should succeed since there's another admin
	mockConvRepo.AssertExpectations(t)
	mockMsgRepo.AssertExpectations(t)
}

func TestSendMessage_CreateError(t *testing.T) {
	ctx := context.Background()
	mockConvRepo := new(MockConversationRepository)
	mockMsgRepo := new(MockMessageRepository)
	logger := createTestLogger()
	hub := websocket.NewHub(logger)

	uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

	participant := &entities.Participant{
		ConversationID: 1,
		UserID:         1,
		UserName:       "Test User",
		Role:           entities.ParticipantRoleMember,
	}

	input := dto.SendMessageInput{Content: "Hello"}

	mockConvRepo.On("GetParticipant", ctx, int64(1), int64(1)).Return(participant, nil)
	mockMsgRepo.On("Create", ctx, mock.AnythingOfType("*entities.Message")).Return(errors.New("db error"))

	msg, err := uc.SendMessage(ctx, 1, 1, input)

	assert.Error(t, err)
	assert.Nil(t, msg)
	assert.Contains(t, err.Error(), "failed to create message")
	mockConvRepo.AssertExpectations(t)
	mockMsgRepo.AssertExpectations(t)
}

func TestSendMessage_EmptyContent(t *testing.T) {
	ctx := context.Background()
	mockConvRepo := new(MockConversationRepository)
	mockMsgRepo := new(MockMessageRepository)
	logger := createTestLogger()
	hub := websocket.NewHub(logger)

	uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

	participant := &entities.Participant{
		ConversationID: 1,
		UserID:         1,
		UserName:       "Test User",
		Role:           entities.ParticipantRoleMember,
	}

	input := dto.SendMessageInput{Content: ""} // empty content

	mockConvRepo.On("GetParticipant", ctx, int64(1), int64(1)).Return(participant, nil)

	msg, err := uc.SendMessage(ctx, 1, 1, input)

	assert.Error(t, err)
	assert.Nil(t, msg)
	assert.Equal(t, entities.ErrEmptyMessageContent, err)
	mockConvRepo.AssertExpectations(t)
}

func TestSendMessage_AttachmentCreateError(t *testing.T) {
	ctx := context.Background()
	mockConvRepo := new(MockConversationRepository)
	mockMsgRepo := new(MockMessageRepository)
	logger := createTestLogger()
	hub := websocket.NewHub(logger)

	uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

	participant := &entities.Participant{
		ConversationID: 1,
		UserID:         1,
		UserName:       "Test User",
		Role:           entities.ParticipantRoleMember,
	}

	input := dto.SendMessageInput{
		Content: "Check this",
		Attachments: []dto.AttachmentInput{
			{
				FileID:   123,
				FileName: "file.pdf",
				FileSize: 1024,
				MimeType: "application/pdf",
				URL:      "https://example.com/file.pdf",
			},
		},
	}

	mockConvRepo.On("GetParticipant", ctx, int64(1), int64(1)).Return(participant, nil)
	mockMsgRepo.On("Create", ctx, mock.AnythingOfType("*entities.Message")).Return(nil)
	mockMsgRepo.On("CreateAttachment", ctx, mock.AnythingOfType("*entities.Attachment")).Return(errors.New("attachment error"))

	msg, err := uc.SendMessage(ctx, 1, 1, input)

	assert.NoError(t, err) // attachment error is logged, not returned
	assert.NotNil(t, msg)
	assert.Empty(t, msg.Attachments)                    // no successful attachments
	assert.Equal(t, entities.MessageTypeText, msg.Type) // stays text since no attachments succeeded
	mockConvRepo.AssertExpectations(t)
	mockMsgRepo.AssertExpectations(t)
}

func TestSendMessage_ReplyError(t *testing.T) {
	ctx := context.Background()
	mockConvRepo := new(MockConversationRepository)
	mockMsgRepo := new(MockMessageRepository)
	logger := createTestLogger()
	hub := websocket.NewHub(logger)

	uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

	participant := &entities.Participant{
		ConversationID: 1,
		UserID:         1,
		UserName:       "Test User",
		Role:           entities.ParticipantRoleMember,
	}

	replyToID := int64(10)
	input := dto.SendMessageInput{
		Content:   "Reply",
		ReplyToID: &replyToID,
	}

	mockConvRepo.On("GetParticipant", ctx, int64(1), int64(1)).Return(participant, nil)
	mockMsgRepo.On("Create", ctx, mock.AnythingOfType("*entities.Message")).Return(nil)
	mockMsgRepo.On("GetByID", ctx, int64(10)).Return(nil, errors.New("not found"))

	msg, err := uc.SendMessage(ctx, 1, 1, input)

	assert.NoError(t, err) // reply load error is silently handled
	assert.NotNil(t, msg)
	assert.Nil(t, msg.ReplyTo) // reply not loaded but message still sent
	mockConvRepo.AssertExpectations(t)
	mockMsgRepo.AssertExpectations(t)
}

func TestEditMessage_EditValidationError(t *testing.T) {
	ctx := context.Background()
	mockConvRepo := new(MockConversationRepository)
	mockMsgRepo := new(MockMessageRepository)
	logger := createTestLogger()
	hub := websocket.NewHub(logger)

	uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

	now := time.Now()
	msg := &entities.Message{
		ID:             1,
		ConversationID: 1,
		SenderID:       1,
		Content:        "Original",
		Type:           entities.MessageTypeText,
		CreatedAt:      now,
	}

	// Empty content should fail Edit validation
	input := dto.EditMessageInput{Content: ""}

	mockMsgRepo.On("GetByID", ctx, int64(1)).Return(msg, nil)

	result, err := uc.EditMessage(ctx, 1, 1, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, entities.ErrEmptyMessageContent, err)
	mockMsgRepo.AssertExpectations(t)
}

func TestEditMessage_UpdateRepoError(t *testing.T) {
	ctx := context.Background()
	mockConvRepo := new(MockConversationRepository)
	mockMsgRepo := new(MockMessageRepository)
	logger := createTestLogger()
	hub := websocket.NewHub(logger)

	uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

	now := time.Now()
	msg := &entities.Message{
		ID:             1,
		ConversationID: 1,
		SenderID:       1,
		Content:        "Original",
		Type:           entities.MessageTypeText,
		CreatedAt:      now,
	}

	input := dto.EditMessageInput{Content: "Updated"}

	mockMsgRepo.On("GetByID", ctx, int64(1)).Return(msg, nil)
	mockMsgRepo.On("Update", ctx, mock.AnythingOfType("*entities.Message")).Return(errors.New("db error"))

	result, err := uc.EditMessage(ctx, 1, 1, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to update message")
	mockMsgRepo.AssertExpectations(t)
}

func TestDeleteMessage_ConvGetByIDError(t *testing.T) {
	ctx := context.Background()
	mockConvRepo := new(MockConversationRepository)
	mockMsgRepo := new(MockMessageRepository)
	logger := createTestLogger()
	hub := websocket.NewHub(logger)

	uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

	now := time.Now()
	msg := &entities.Message{
		ID:             1,
		ConversationID: 1,
		SenderID:       1,
		Content:        "Message",
		Type:           entities.MessageTypeText,
		CreatedAt:      now,
	}

	mockMsgRepo.On("GetByID", ctx, int64(1)).Return(msg, nil)
	mockConvRepo.On("GetByID", ctx, int64(1)).Return(nil, errors.New("conv not found"))

	err := uc.DeleteMessage(ctx, 1, 1)

	assert.Error(t, err)
	mockMsgRepo.AssertExpectations(t)
	mockConvRepo.AssertExpectations(t)
}

func TestDeleteMessage_AlreadyDeleted(t *testing.T) {
	ctx := context.Background()
	mockConvRepo := new(MockConversationRepository)
	mockMsgRepo := new(MockMessageRepository)
	logger := createTestLogger()
	hub := websocket.NewHub(logger)

	uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

	now := time.Now()
	msg := &entities.Message{
		ID:             1,
		ConversationID: 1,
		SenderID:       1,
		Content:        "",
		Type:           entities.MessageTypeText,
		IsDeleted:      true,
		DeletedAt:      &now,
		CreatedAt:      now,
	}

	conv := &entities.Conversation{
		ID:   1,
		Type: entities.ConversationTypeDirect,
		Participants: []entities.Participant{
			{UserID: 1, Role: entities.ParticipantRoleMember},
		},
	}

	mockMsgRepo.On("GetByID", ctx, int64(1)).Return(msg, nil)
	mockConvRepo.On("GetByID", ctx, int64(1)).Return(conv, nil)

	err := uc.DeleteMessage(ctx, 1, 1)

	assert.Error(t, err)
	assert.Equal(t, entities.ErrCannotDeleteMessage, err)
	mockMsgRepo.AssertExpectations(t)
	mockConvRepo.AssertExpectations(t)
}

func TestDeleteMessage_UpdateRepoError(t *testing.T) {
	ctx := context.Background()
	mockConvRepo := new(MockConversationRepository)
	mockMsgRepo := new(MockMessageRepository)
	logger := createTestLogger()
	hub := websocket.NewHub(logger)

	uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

	now := time.Now()
	msg := &entities.Message{
		ID:             1,
		ConversationID: 1,
		SenderID:       1,
		Content:        "Message",
		Type:           entities.MessageTypeText,
		CreatedAt:      now,
	}

	conv := &entities.Conversation{
		ID:   1,
		Type: entities.ConversationTypeDirect,
		Participants: []entities.Participant{
			{UserID: 1, Role: entities.ParticipantRoleMember},
		},
	}

	mockMsgRepo.On("GetByID", ctx, int64(1)).Return(msg, nil)
	mockConvRepo.On("GetByID", ctx, int64(1)).Return(conv, nil)
	mockMsgRepo.On("Update", ctx, mock.AnythingOfType("*entities.Message")).Return(errors.New("db error"))

	err := uc.DeleteMessage(ctx, 1, 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete message")
	mockMsgRepo.AssertExpectations(t)
	mockConvRepo.AssertExpectations(t)
}

func TestGetMessages_ListError(t *testing.T) {
	ctx := context.Background()
	mockConvRepo := new(MockConversationRepository)
	mockMsgRepo := new(MockMessageRepository)
	logger := createTestLogger()
	hub := websocket.NewHub(logger)

	uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

	participant := &entities.Participant{
		ConversationID: 1,
		UserID:         1,
		Role:           entities.ParticipantRoleMember,
	}

	input := dto.MessageFilterInput{Limit: 50}

	mockConvRepo.On("GetParticipant", ctx, int64(1), int64(1)).Return(participant, nil)
	mockMsgRepo.On("List", ctx, mock.AnythingOfType("entities.MessageFilter")).Return([]*entities.Message{}, errors.New("db error"))

	result, hasMore, err := uc.GetMessages(ctx, 1, 1, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.False(t, hasMore)
	assert.Contains(t, err.Error(), "failed to list messages")
	mockConvRepo.AssertExpectations(t)
	mockMsgRepo.AssertExpectations(t)
}

func TestGetMessages_WithAvatarURLs(t *testing.T) {
	ctx := context.Background()
	mockConvRepo := new(MockConversationRepository)
	mockMsgRepo := new(MockMessageRepository)
	logger := createTestLogger()
	hub := websocket.NewHub(logger)

	uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

	participant := &entities.Participant{
		ConversationID: 1,
		UserID:         1,
		Role:           entities.ParticipantRoleMember,
	}

	avatarURL := testAvatarURL
	messages := []*entities.Message{
		{ID: 1, Content: "Message 1", SenderAvatarURL: &avatarURL, CreatedAt: time.Now()},
	}

	input := dto.MessageFilterInput{Limit: 50}

	mockConvRepo.On("GetParticipant", ctx, int64(1), int64(1)).Return(participant, nil)
	mockMsgRepo.On("List", ctx, mock.AnythingOfType("entities.MessageFilter")).Return(messages, nil)

	result, _, err := uc.GetMessages(ctx, 1, 1, input)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.NotNil(t, result[0].SenderAvatarURL)
	assert.Equal(t, avatarURL, *result[0].SenderAvatarURL)
	mockConvRepo.AssertExpectations(t)
	mockMsgRepo.AssertExpectations(t)
}

func TestSearchMessages_SearchError(t *testing.T) {
	ctx := context.Background()
	mockConvRepo := new(MockConversationRepository)
	mockMsgRepo := new(MockMessageRepository)
	logger := createTestLogger()
	hub := websocket.NewHub(logger)

	uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

	participant := &entities.Participant{
		ConversationID: 1,
		UserID:         1,
		Role:           entities.ParticipantRoleMember,
	}

	mockConvRepo.On("GetParticipant", ctx, int64(1), int64(1)).Return(participant, nil)
	mockMsgRepo.On("Search", ctx, int64(1), "test", 20, 0).Return([]*entities.Message{}, int64(0), errors.New("search error"))

	_, total, err := uc.SearchMessages(ctx, 1, 1, "test", 20, 0)

	assert.Error(t, err)
	assert.Equal(t, int64(0), total)
	mockConvRepo.AssertExpectations(t)
	mockMsgRepo.AssertExpectations(t)
}

func TestListConversations_WithSearchFilter(t *testing.T) {
	ctx := context.Background()
	mockConvRepo := new(MockConversationRepository)
	mockMsgRepo := new(MockMessageRepository)
	logger := createTestLogger()
	hub := websocket.NewHub(logger)

	uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

	searchQuery := "test"
	input := dto.ConversationFilterInput{
		Search: &searchQuery,
		Limit:  10,
	}

	mockConvRepo.On("List", ctx, mock.AnythingOfType("entities.ConversationFilter")).Return([]*entities.Conversation{}, int64(0), nil)

	result, total, err := uc.ListConversations(ctx, 1, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(0), total)
	mockConvRepo.AssertExpectations(t)
}

func TestCreateGroupConversation_WithoutDescription(t *testing.T) {
	ctx := context.Background()
	mockConvRepo := new(MockConversationRepository)
	mockMsgRepo := new(MockMessageRepository)
	logger := createTestLogger()
	hub := websocket.NewHub(logger)

	uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

	input := dto.CreateGroupConversationInput{
		Title:          "No Desc Group",
		ParticipantIDs: []int64{2, 3},
	}

	mockConvRepo.On("Create", ctx, mock.AnythingOfType("*entities.Conversation")).Return(nil)
	mockMsgRepo.On("Create", ctx, mock.AnythingOfType("*entities.Message")).Return(nil)

	conv, err := uc.CreateGroupConversation(ctx, 1, input)

	assert.NoError(t, err)
	assert.NotNil(t, conv)
	assert.Nil(t, conv.Description) // no description set
	mockConvRepo.AssertExpectations(t)
	mockMsgRepo.AssertExpectations(t)
}

func TestSendMessage_WithNotifier(t *testing.T) {
	ctx := context.Background()
	mockConvRepo := new(MockConversationRepository)
	mockMsgRepo := new(MockMessageRepository)
	mockNotifier := new(MockMessageNotifier)
	logger := createTestLogger()
	hub := websocket.NewHub(logger)

	uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, mockNotifier, nil)

	participant := &entities.Participant{
		ConversationID: 1,
		UserID:         1,
		UserName:       "Test User",
		Role:           entities.ParticipantRoleMember,
	}

	input := dto.SendMessageInput{Content: "Hello"}

	mockConvRepo.On("GetParticipant", ctx, int64(1), int64(1)).Return(participant, nil)
	mockMsgRepo.On("Create", ctx, mock.AnythingOfType("*entities.Message")).Return(nil)

	// For the goroutine: notifyParticipants will use context.Background()
	mockConvRepo.On("GetParticipants", mock.Anything, int64(1)).Return([]entities.Participant{
		{UserID: 1}, // sender, skipped
	}, nil).Maybe()

	msg, err := uc.SendMessage(ctx, 1, 1, input)

	assert.NoError(t, err)
	assert.NotNil(t, msg)
	mockConvRepo.AssertExpectations(t)
	mockMsgRepo.AssertExpectations(t)
}

func TestGetMessages_WithFilterParams(t *testing.T) {
	ctx := context.Background()
	mockConvRepo := new(MockConversationRepository)
	mockMsgRepo := new(MockMessageRepository)
	logger := createTestLogger()
	hub := websocket.NewHub(logger)

	uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

	participant := &entities.Participant{
		ConversationID: 1,
		UserID:         1,
		Role:           entities.ParticipantRoleMember,
	}

	beforeID := int64(100)
	search := "hello"
	input := dto.MessageFilterInput{
		Limit:    10,
		BeforeID: &beforeID,
		Search:   &search,
	}

	mockConvRepo.On("GetParticipant", ctx, int64(1), int64(1)).Return(participant, nil)
	mockMsgRepo.On("List", ctx, mock.AnythingOfType("entities.MessageFilter")).Return([]*entities.Message{}, nil)

	result, hasMore, err := uc.GetMessages(ctx, 1, 1, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, hasMore)
	mockConvRepo.AssertExpectations(t)
	mockMsgRepo.AssertExpectations(t)
}

func TestLeaveConversation_AdminWithLeftAdmin(t *testing.T) {
	ctx := context.Background()
	mockConvRepo := new(MockConversationRepository)
	mockMsgRepo := new(MockMessageRepository)
	logger := createTestLogger()
	hub := websocket.NewHub(logger)

	uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

	leftAt := time.Now()
	conv := &entities.Conversation{
		ID:   1,
		Type: entities.ConversationTypeGroup,
		Participants: []entities.Participant{
			{UserID: 1, Role: entities.ParticipantRoleAdmin},                  // Active admin trying to leave
			{UserID: 2, Role: entities.ParticipantRoleAdmin, LeftAt: &leftAt}, // Left admin - should not count
			{UserID: 3, Role: entities.ParticipantRoleMember},
		},
	}

	mockConvRepo.On("GetByID", ctx, int64(1)).Return(conv, nil)

	err := uc.LeaveConversation(ctx, 1, 1)

	// Should fail because user 2 is a left admin, so user 1 is the last active admin
	assert.Error(t, err)
	assert.Equal(t, entities.ErrCannotRemoveLastAdmin, err)
	mockConvRepo.AssertExpectations(t)
}

func TestSendMessage_ShortMimeType(t *testing.T) {
	ctx := context.Background()
	mockConvRepo := new(MockConversationRepository)
	mockMsgRepo := new(MockMessageRepository)
	logger := createTestLogger()
	hub := websocket.NewHub(logger)

	uc := NewMessagingUseCase(mockConvRepo, mockMsgRepo, hub, logger, nil, nil)

	participant := &entities.Participant{
		ConversationID: 1,
		UserID:         1,
		UserName:       "Test User",
		Role:           entities.ParticipantRoleMember,
	}

	// Mime type shorter than 5 chars - should go to file type
	input := dto.SendMessageInput{
		Content: "Check this",
		Attachments: []dto.AttachmentInput{
			{
				FileID:   123,
				FileName: "data.bin",
				FileSize: 512,
				MimeType: "bin",
				URL:      "https://example.com/data.bin",
			},
		},
	}

	mockConvRepo.On("GetParticipant", ctx, int64(1), int64(1)).Return(participant, nil)
	mockMsgRepo.On("Create", ctx, mock.AnythingOfType("*entities.Message")).Return(nil)
	mockMsgRepo.On("CreateAttachment", ctx, mock.AnythingOfType("*entities.Attachment")).Return(nil)

	msg, err := uc.SendMessage(ctx, 1, 1, input)

	assert.NoError(t, err)
	assert.NotNil(t, msg)
	assert.Equal(t, entities.MessageTypeFile, msg.Type)
	mockConvRepo.AssertExpectations(t)
	mockMsgRepo.AssertExpectations(t)
}
