package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/messaging/domain/entities"
	messagingHttp "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/messaging/interfaces/http"

	messagingUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/messaging/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/messaging/infrastructure/websocket"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/validation"
)

// --- Mock Repositories ---

type MockConversationRepo struct {
	mock.Mock
}

func (m *MockConversationRepo) Create(ctx context.Context, conv *entities.Conversation) error {
	args := m.Called(ctx, conv)
	if args.Error(0) == nil {
		conv.ID = 1
	}
	return args.Error(0)
}

func (m *MockConversationRepo) Update(ctx context.Context, conv *entities.Conversation) error {
	args := m.Called(ctx, conv)
	return args.Error(0)
}

func (m *MockConversationRepo) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockConversationRepo) GetByID(ctx context.Context, id int64) (*entities.Conversation, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Conversation), args.Error(1)
}

func (m *MockConversationRepo) List(ctx context.Context, filter entities.ConversationFilter) ([]*entities.Conversation, int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*entities.Conversation), args.Get(1).(int64), args.Error(2)
}

func (m *MockConversationRepo) GetDirectConversation(ctx context.Context, userID1, userID2 int64) (*entities.Conversation, error) {
	args := m.Called(ctx, userID1, userID2)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Conversation), args.Error(1)
}

func (m *MockConversationRepo) AddParticipant(ctx context.Context, participant *entities.Participant) error {
	args := m.Called(ctx, participant)
	return args.Error(0)
}

func (m *MockConversationRepo) RemoveParticipant(ctx context.Context, conversationID, userID int64) error {
	args := m.Called(ctx, conversationID, userID)
	return args.Error(0)
}

func (m *MockConversationRepo) UpdateParticipant(ctx context.Context, participant *entities.Participant) error {
	args := m.Called(ctx, participant)
	return args.Error(0)
}

func (m *MockConversationRepo) GetParticipants(ctx context.Context, conversationID int64) ([]entities.Participant, error) {
	args := m.Called(ctx, conversationID)
	return args.Get(0).([]entities.Participant), args.Error(1)
}

func (m *MockConversationRepo) GetParticipant(ctx context.Context, conversationID, userID int64) (*entities.Participant, error) {
	args := m.Called(ctx, conversationID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Participant), args.Error(1)
}

func (m *MockConversationRepo) UpdateLastRead(ctx context.Context, conversationID, userID, messageID int64) error {
	args := m.Called(ctx, conversationID, userID, messageID)
	return args.Error(0)
}

func (m *MockConversationRepo) GetUnreadCount(ctx context.Context, conversationID, userID int64) (int, error) {
	args := m.Called(ctx, conversationID, userID)
	return args.Get(0).(int), args.Error(1)
}

type MockMessageRepo struct {
	mock.Mock
}

func (m *MockMessageRepo) Create(ctx context.Context, msg *entities.Message) error {
	args := m.Called(ctx, msg)
	if args.Error(0) == nil {
		msg.ID = 1
	}
	return args.Error(0)
}

func (m *MockMessageRepo) GetByID(ctx context.Context, id int64) (*entities.Message, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Message), args.Error(1)
}

func (m *MockMessageRepo) List(ctx context.Context, filter entities.MessageFilter) ([]*entities.Message, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*entities.Message), args.Error(1)
}

func (m *MockMessageRepo) Update(ctx context.Context, msg *entities.Message) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

func (m *MockMessageRepo) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockMessageRepo) GetLastMessage(ctx context.Context, conversationID int64) (*entities.Message, error) {
	args := m.Called(ctx, conversationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Message), args.Error(1)
}

func (m *MockMessageRepo) CountUnread(ctx context.Context, conversationID, userID int64, lastReadAt *int64) (int, error) {
	args := m.Called(ctx, conversationID, userID, lastReadAt)
	return args.Get(0).(int), args.Error(1)
}

func (m *MockMessageRepo) CreateAttachment(ctx context.Context, attachment *entities.Attachment) error {
	args := m.Called(ctx, attachment)
	return args.Error(0)
}

func (m *MockMessageRepo) GetAttachments(ctx context.Context, messageID int64) ([]entities.Attachment, error) {
	args := m.Called(ctx, messageID)
	return args.Get(0).([]entities.Attachment), args.Error(1)
}

func (m *MockMessageRepo) Search(ctx context.Context, conversationID int64, query string, limit, offset int) ([]*entities.Message, int64, error) {
	args := m.Called(ctx, conversationID, query, limit, offset)
	return args.Get(0).([]*entities.Message), args.Get(1).(int64), args.Error(2)
}

// --- Test Helpers ---

func setupTestHandler(t *testing.T) (*messagingHttp.MessagingHandler, *MockConversationRepo, *MockMessageRepo) {
	t.Helper()
	convRepo := new(MockConversationRepo)
	msgRepo := new(MockMessageRepo)
	logger := logging.NewLogger("error")
	hub := websocket.NewHub(logger)
	validator := validation.NewValidator()
	useCase := messagingUsecases.NewMessagingUseCase(convRepo, msgRepo, hub, logger, nil, nil)
	handler := messagingHttp.NewMessagingHandler(useCase, hub, logger, validator)
	return handler, convRepo, msgRepo
}

func setupRouter(handler *messagingHttp.MessagingHandler, userID int64) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	authMiddleware := func(c *gin.Context) {
		if userID > 0 {
			c.Set("user_id", userID)
		}
		c.Next()
	}
	handler.RegisterRoutes(router.Group(""), authMiddleware)
	return router
}

func setupRouterNoAuth(handler *messagingHttp.MessagingHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	authMiddleware := func(c *gin.Context) {
		// Do not set user_id
		c.Next()
	}
	handler.RegisterRoutes(router.Group(""), authMiddleware)
	return router
}

func makeGroupConversation(creatorID int64) *entities.Conversation {
	title := "Test Group"
	return &entities.Conversation{
		ID:        1,
		Type:      entities.ConversationTypeGroup,
		Title:     &title,
		CreatedBy: creatorID,
		Participants: []entities.Participant{
			{ID: 1, UserID: creatorID, Role: entities.ParticipantRoleAdmin, JoinedAt: time.Now()},
			{ID: 2, UserID: 2, Role: entities.ParticipantRoleMember, JoinedAt: time.Now()},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}


func makeMessage(id, convID, senderID int64) *entities.Message {
	return &entities.Message{
		ID:             id,
		ConversationID: convID,
		SenderID:       senderID,
		Type:           entities.MessageTypeText,
		Content:        "Hello",
		CreatedAt:      time.Now(),
	}
}

// --- Tests ---

func TestCreateDirectConversation_Success(t *testing.T) {
	handler, convRepo, _ := setupTestHandler(t)
	router := setupRouter(handler, 1)

	convRepo.On("GetDirectConversation", mock.Anything, int64(1), int64(2)).Return(nil, nil)
	convRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Conversation")).Return(nil)

	body, _ := json.Marshal(map[string]any{"recipient_id": 2})
	req := httptest.NewRequest(http.MethodPost, "/conversations/direct", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	convRepo.AssertExpectations(t)
}

func TestCreateDirectConversation_Unauthorized(t *testing.T) {
	handler, _, _ := setupTestHandler(t)
	router := setupRouterNoAuth(handler)

	body, _ := json.Marshal(map[string]any{"recipient_id": 2})
	req := httptest.NewRequest(http.MethodPost, "/conversations/direct", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCreateDirectConversation_InvalidBody(t *testing.T) {
	handler, _, _ := setupTestHandler(t)
	router := setupRouter(handler, 1)

	req := httptest.NewRequest(http.MethodPost, "/conversations/direct", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateDirectConversation_ValidationError(t *testing.T) {
	handler, _, _ := setupTestHandler(t)
	router := setupRouter(handler, 1)

	// recipient_id = 0 should fail validation (gt=0)
	body, _ := json.Marshal(map[string]any{"recipient_id": 0})
	req := httptest.NewRequest(http.MethodPost, "/conversations/direct", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateDirectConversation_UseCaseError(t *testing.T) {
	handler, convRepo, _ := setupTestHandler(t)
	router := setupRouter(handler, 1)

	convRepo.On("GetDirectConversation", mock.Anything, int64(1), int64(2)).Return(nil, errors.New("db error"))

	body, _ := json.Marshal(map[string]any{"recipient_id": 2})
	req := httptest.NewRequest(http.MethodPost, "/conversations/direct", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCreateGroupConversation_Success(t *testing.T) {
	handler, convRepo, msgRepo := setupTestHandler(t)
	router := setupRouter(handler, 1)

	convRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Conversation")).Return(nil)
	msgRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Message")).Return(nil)

	body, _ := json.Marshal(map[string]any{
		"title":           "My Group",
		"participant_ids": []int64{2, 3},
	})
	req := httptest.NewRequest(http.MethodPost, "/conversations/group", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCreateGroupConversation_InvalidBody(t *testing.T) {
	handler, _, _ := setupTestHandler(t)
	router := setupRouter(handler, 1)

	req := httptest.NewRequest(http.MethodPost, "/conversations/group", bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Validation error: title is required
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateGroupConversation_UseCaseError(t *testing.T) {
	handler, convRepo, _ := setupTestHandler(t)
	router := setupRouter(handler, 1)

	convRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Conversation")).Return(errors.New("db error"))

	body, _ := json.Marshal(map[string]any{
		"title":           "My Group",
		"participant_ids": []int64{2, 3},
	})
	req := httptest.NewRequest(http.MethodPost, "/conversations/group", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestListConversations_Success(t *testing.T) {
	handler, convRepo, _ := setupTestHandler(t)
	router := setupRouter(handler, 1)

	convRepo.On("List", mock.Anything, mock.AnythingOfType("entities.ConversationFilter")).
		Return([]*entities.Conversation{}, int64(0), nil)

	req := httptest.NewRequest(http.MethodGet, "/conversations", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestListConversations_WithQueryParams(t *testing.T) {
	handler, convRepo, _ := setupTestHandler(t)
	router := setupRouter(handler, 1)

	convRepo.On("List", mock.Anything, mock.AnythingOfType("entities.ConversationFilter")).
		Return([]*entities.Conversation{}, int64(0), nil)

	req := httptest.NewRequest(http.MethodGet, "/conversations?type=direct&limit=10&offset=5", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestListConversations_UseCaseError(t *testing.T) {
	handler, convRepo, _ := setupTestHandler(t)
	router := setupRouter(handler, 1)

	convRepo.On("List", mock.Anything, mock.AnythingOfType("entities.ConversationFilter")).
		Return([]*entities.Conversation{}, int64(0), errors.New("db error"))

	req := httptest.NewRequest(http.MethodGet, "/conversations", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestListConversations_Unauthorized(t *testing.T) {
	handler, _, _ := setupTestHandler(t)
	router := setupRouterNoAuth(handler)

	req := httptest.NewRequest(http.MethodGet, "/conversations", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetConversation_Success(t *testing.T) {
	handler, convRepo, msgRepo := setupTestHandler(t)
	router := setupRouter(handler, 1)

	conv := makeGroupConversation(1)
	convRepo.On("GetByID", mock.Anything, int64(1)).Return(conv, nil)
	convRepo.On("GetUnreadCount", mock.Anything, int64(1), int64(1)).Return(0, nil)
	msgRepo.On("GetLastMessage", mock.Anything, int64(1)).Return(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/conversations/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetConversation_InvalidID(t *testing.T) {
	handler, _, _ := setupTestHandler(t)
	router := setupRouter(handler, 1)

	req := httptest.NewRequest(http.MethodGet, "/conversations/invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetConversation_NotFound(t *testing.T) {
	handler, convRepo, _ := setupTestHandler(t)
	router := setupRouter(handler, 1)

	convRepo.On("GetByID", mock.Anything, int64(999)).Return(nil, errors.New("not found"))

	req := httptest.NewRequest(http.MethodGet, "/conversations/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateConversation_Success(t *testing.T) {
	handler, convRepo, _ := setupTestHandler(t)
	router := setupRouter(handler, 1)

	conv := makeGroupConversation(1)
	convRepo.On("GetByID", mock.Anything, int64(1)).Return(conv, nil)
	convRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.Conversation")).Return(nil)

	body, _ := json.Marshal(map[string]any{"title": "Updated Title"})
	req := httptest.NewRequest(http.MethodPatch, "/conversations/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpdateConversation_InvalidID(t *testing.T) {
	handler, _, _ := setupTestHandler(t)
	router := setupRouter(handler, 1)

	body, _ := json.Marshal(map[string]any{"title": "Updated Title"})
	req := httptest.NewRequest(http.MethodPatch, "/conversations/invalid", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateConversation_InvalidBody(t *testing.T) {
	handler, _, _ := setupTestHandler(t)
	router := setupRouter(handler, 1)

	req := httptest.NewRequest(http.MethodPatch, "/conversations/1", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateConversation_UseCaseError(t *testing.T) {
	handler, convRepo, _ := setupTestHandler(t)
	router := setupRouter(handler, 1)

	convRepo.On("GetByID", mock.Anything, int64(1)).Return(nil, errors.New("error"))

	body, _ := json.Marshal(map[string]any{"title": "Updated"})
	req := httptest.NewRequest(http.MethodPatch, "/conversations/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAddParticipants_Success(t *testing.T) {
	handler, convRepo, msgRepo := setupTestHandler(t)
	router := setupRouter(handler, 1)

	conv := makeGroupConversation(1)
	convRepo.On("GetByID", mock.Anything, int64(1)).Return(conv, nil)
	convRepo.On("AddParticipant", mock.Anything, mock.AnythingOfType("*entities.Participant")).Return(nil)
	msgRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Message")).Return(nil)

	body, _ := json.Marshal(map[string]any{"user_ids": []int64{3}})
	req := httptest.NewRequest(http.MethodPost, "/conversations/1/participants", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAddParticipants_InvalidID(t *testing.T) {
	handler, _, _ := setupTestHandler(t)
	router := setupRouter(handler, 1)

	body, _ := json.Marshal(map[string]any{"user_ids": []int64{3}})
	req := httptest.NewRequest(http.MethodPost, "/conversations/invalid/participants", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAddParticipants_InvalidBody(t *testing.T) {
	handler, _, _ := setupTestHandler(t)
	router := setupRouter(handler, 1)

	req := httptest.NewRequest(http.MethodPost, "/conversations/1/participants", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAddParticipants_UseCaseError(t *testing.T) {
	handler, convRepo, _ := setupTestHandler(t)
	router := setupRouter(handler, 1)

	convRepo.On("GetByID", mock.Anything, int64(1)).Return(nil, errors.New("not found"))

	body, _ := json.Marshal(map[string]any{"user_ids": []int64{3}})
	req := httptest.NewRequest(http.MethodPost, "/conversations/1/participants", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLeaveConversation_Success(t *testing.T) {
	handler, convRepo, msgRepo := setupTestHandler(t)
	router := setupRouter(handler, 2)

	conv := makeGroupConversation(1)
	// Add user 2 and a second admin so user 2 can leave
	convRepo.On("GetByID", mock.Anything, int64(1)).Return(conv, nil)
	convRepo.On("RemoveParticipant", mock.Anything, int64(1), int64(2)).Return(nil)
	msgRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Message")).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/conversations/1/leave", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLeaveConversation_InvalidID(t *testing.T) {
	handler, _, _ := setupTestHandler(t)
	router := setupRouter(handler, 1)

	req := httptest.NewRequest(http.MethodPost, "/conversations/invalid/leave", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLeaveConversation_UseCaseError(t *testing.T) {
	handler, convRepo, _ := setupTestHandler(t)
	router := setupRouter(handler, 1)

	convRepo.On("GetByID", mock.Anything, int64(1)).Return(nil, errors.New("error"))

	req := httptest.NewRequest(http.MethodPost, "/conversations/1/leave", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSendMessage_Success(t *testing.T) {
	handler, convRepo, msgRepo := setupTestHandler(t)
	router := setupRouter(handler, 1)

	participant := &entities.Participant{
		ID:       1,
		UserID:   1,
		UserName: "Test User",
	}
	convRepo.On("GetParticipant", mock.Anything, int64(1), int64(1)).Return(participant, nil)
	msgRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.Message")).Return(nil)

	body, _ := json.Marshal(map[string]any{"content": "Hello World"})
	req := httptest.NewRequest(http.MethodPost, "/conversations/1/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestSendMessage_InvalidConversationID(t *testing.T) {
	handler, _, _ := setupTestHandler(t)
	router := setupRouter(handler, 1)

	body, _ := json.Marshal(map[string]any{"content": "Hello"})
	req := httptest.NewRequest(http.MethodPost, "/conversations/invalid/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSendMessage_InvalidBody(t *testing.T) {
	handler, _, _ := setupTestHandler(t)
	router := setupRouter(handler, 1)

	req := httptest.NewRequest(http.MethodPost, "/conversations/1/messages", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSendMessage_EmptyContent(t *testing.T) {
	handler, _, _ := setupTestHandler(t)
	router := setupRouter(handler, 1)

	body, _ := json.Marshal(map[string]any{"content": ""})
	req := httptest.NewRequest(http.MethodPost, "/conversations/1/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSendMessage_UseCaseError(t *testing.T) {
	handler, convRepo, _ := setupTestHandler(t)
	router := setupRouter(handler, 1)

	convRepo.On("GetParticipant", mock.Anything, int64(1), int64(1)).Return(nil, errors.New("not participant"))

	body, _ := json.Marshal(map[string]any{"content": "Hello"})
	req := httptest.NewRequest(http.MethodPost, "/conversations/1/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetMessages_Success(t *testing.T) {
	handler, convRepo, msgRepo := setupTestHandler(t)
	router := setupRouter(handler, 1)

	participant := &entities.Participant{ID: 1, UserID: 1}
	convRepo.On("GetParticipant", mock.Anything, int64(1), int64(1)).Return(participant, nil)
	msgRepo.On("List", mock.Anything, mock.AnythingOfType("entities.MessageFilter")).
		Return([]*entities.Message{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/conversations/1/messages", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetMessages_InvalidID(t *testing.T) {
	handler, _, _ := setupTestHandler(t)
	router := setupRouter(handler, 1)

	req := httptest.NewRequest(http.MethodGet, "/conversations/invalid/messages", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetMessages_UseCaseError(t *testing.T) {
	handler, convRepo, _ := setupTestHandler(t)
	router := setupRouter(handler, 1)

	convRepo.On("GetParticipant", mock.Anything, int64(1), int64(1)).Return(nil, errors.New("error"))

	req := httptest.NewRequest(http.MethodGet, "/conversations/1/messages", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEditMessage_Success(t *testing.T) {
	handler, convRepo, msgRepo := setupTestHandler(t)
	router := setupRouter(handler, 1)

	msg := makeMessage(1, 1, 1)
	msgRepo.On("GetByID", mock.Anything, int64(1)).Return(msg, nil)
	msgRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.Message")).Return(nil)
	// EditMessage broadcasts to conversation, ignore unused
	_ = convRepo

	body, _ := json.Marshal(map[string]any{"content": "Updated message"})
	req := httptest.NewRequest(http.MethodPatch, "/conversations/1/messages/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEditMessage_InvalidMessageID(t *testing.T) {
	handler, _, _ := setupTestHandler(t)
	router := setupRouter(handler, 1)

	body, _ := json.Marshal(map[string]any{"content": "Updated"})
	req := httptest.NewRequest(http.MethodPatch, "/conversations/1/messages/invalid", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEditMessage_InvalidBody(t *testing.T) {
	handler, _, _ := setupTestHandler(t)
	router := setupRouter(handler, 1)

	req := httptest.NewRequest(http.MethodPatch, "/conversations/1/messages/1", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEditMessage_UseCaseError(t *testing.T) {
	handler, _, msgRepo := setupTestHandler(t)
	router := setupRouter(handler, 1)

	msgRepo.On("GetByID", mock.Anything, int64(1)).Return(nil, errors.New("not found"))

	body, _ := json.Marshal(map[string]any{"content": "Updated"})
	req := httptest.NewRequest(http.MethodPatch, "/conversations/1/messages/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteMessage_Success(t *testing.T) {
	handler, convRepo, msgRepo := setupTestHandler(t)
	router := setupRouter(handler, 1)

	msg := makeMessage(1, 1, 1)
	conv := makeGroupConversation(1)
	msgRepo.On("GetByID", mock.Anything, int64(1)).Return(msg, nil)
	convRepo.On("GetByID", mock.Anything, int64(1)).Return(conv, nil)
	msgRepo.On("Update", mock.Anything, mock.AnythingOfType("*entities.Message")).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/conversations/1/messages/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeleteMessage_InvalidMessageID(t *testing.T) {
	handler, _, _ := setupTestHandler(t)
	router := setupRouter(handler, 1)

	req := httptest.NewRequest(http.MethodDelete, "/conversations/1/messages/invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteMessage_UseCaseError(t *testing.T) {
	handler, _, msgRepo := setupTestHandler(t)
	router := setupRouter(handler, 1)

	msgRepo.On("GetByID", mock.Anything, int64(1)).Return(nil, errors.New("error"))

	req := httptest.NewRequest(http.MethodDelete, "/conversations/1/messages/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMarkAsRead_Success(t *testing.T) {
	handler, convRepo, _ := setupTestHandler(t)
	router := setupRouter(handler, 1)

	participant := &entities.Participant{ID: 1, UserID: 1}
	convRepo.On("GetParticipant", mock.Anything, int64(1), int64(1)).Return(participant, nil)
	convRepo.On("UpdateLastRead", mock.Anything, int64(1), int64(1), int64(10)).Return(nil)

	body, _ := json.Marshal(map[string]any{"message_id": 10})
	req := httptest.NewRequest(http.MethodPost, "/conversations/1/read", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMarkAsRead_InvalidConversationID(t *testing.T) {
	handler, _, _ := setupTestHandler(t)
	router := setupRouter(handler, 1)

	body, _ := json.Marshal(map[string]any{"message_id": 10})
	req := httptest.NewRequest(http.MethodPost, "/conversations/invalid/read", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMarkAsRead_InvalidBody(t *testing.T) {
	handler, _, _ := setupTestHandler(t)
	router := setupRouter(handler, 1)

	req := httptest.NewRequest(http.MethodPost, "/conversations/1/read", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMarkAsRead_UseCaseError(t *testing.T) {
	handler, convRepo, _ := setupTestHandler(t)
	router := setupRouter(handler, 1)

	convRepo.On("GetParticipant", mock.Anything, int64(1), int64(1)).Return(nil, errors.New("error"))

	body, _ := json.Marshal(map[string]any{"message_id": 10})
	req := httptest.NewRequest(http.MethodPost, "/conversations/1/read", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSearchMessages_Success(t *testing.T) {
	handler, convRepo, msgRepo := setupTestHandler(t)
	router := setupRouter(handler, 1)

	participant := &entities.Participant{ID: 1, UserID: 1}
	convRepo.On("GetParticipant", mock.Anything, int64(1), int64(1)).Return(participant, nil)
	msgRepo.On("Search", mock.Anything, int64(1), "hello", 20, 0).
		Return([]*entities.Message{}, int64(0), nil)

	req := httptest.NewRequest(http.MethodGet, "/conversations/1/messages/search?q=hello", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSearchMessages_WithPagination(t *testing.T) {
	handler, convRepo, msgRepo := setupTestHandler(t)
	router := setupRouter(handler, 1)

	participant := &entities.Participant{ID: 1, UserID: 1}
	convRepo.On("GetParticipant", mock.Anything, int64(1), int64(1)).Return(participant, nil)
	msgRepo.On("Search", mock.Anything, int64(1), "hello", 10, 5).
		Return([]*entities.Message{}, int64(0), nil)

	req := httptest.NewRequest(http.MethodGet, "/conversations/1/messages/search?q=hello&limit=10&offset=5", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSearchMessages_InvalidConversationID(t *testing.T) {
	handler, _, _ := setupTestHandler(t)
	router := setupRouter(handler, 1)

	req := httptest.NewRequest(http.MethodGet, "/conversations/invalid/messages/search?q=hello", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSearchMessages_MissingQuery(t *testing.T) {
	handler, _, _ := setupTestHandler(t)
	router := setupRouter(handler, 1)

	req := httptest.NewRequest(http.MethodGet, "/conversations/1/messages/search", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSearchMessages_UseCaseError(t *testing.T) {
	handler, convRepo, _ := setupTestHandler(t)
	router := setupRouter(handler, 1)

	convRepo.On("GetParticipant", mock.Anything, int64(1), int64(1)).Return(nil, errors.New("error"))

	req := httptest.NewRequest(http.MethodGet, "/conversations/1/messages/search?q=hello", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetConversation_Unauthorized(t *testing.T) {
	handler, _, _ := setupTestHandler(t)
	router := setupRouterNoAuth(handler)

	req := httptest.NewRequest(http.MethodGet, "/conversations/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSendMessage_Unauthorized(t *testing.T) {
	handler, _, _ := setupTestHandler(t)
	router := setupRouterNoAuth(handler)

	body, _ := json.Marshal(map[string]any{"content": "Hello"})
	req := httptest.NewRequest(http.MethodPost, "/conversations/1/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUserIDAsWrongType(t *testing.T) {
	handler, _, _ := setupTestHandler(t)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	authMiddleware := func(c *gin.Context) {
		c.Set("user_id", "not_an_int") // wrong type
		c.Next()
	}
	handler.RegisterRoutes(router.Group(""), authMiddleware)

	req := httptest.NewRequest(http.MethodGet, "/conversations", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRegisterRoutes_AllEndpointsExist(t *testing.T) {
	handler, convRepo, msgRepo := setupTestHandler(t)
	router := setupRouter(handler, 1)

	// Prepare mocks for endpoints that need them
	conv := makeGroupConversation(1)
	participant := &entities.Participant{ID: 1, UserID: 1, UserName: "Test"}

	tests := []struct {
		method string
		path   string
		status int
	}{
		{http.MethodGet, "/conversations", 0},                 // Will work if mocked
		{http.MethodPost, "/conversations/direct", 0},         // Needs body
		{http.MethodPost, "/conversations/group", 0},          // Needs body
		{http.MethodGet, "/conversations/1", 0},               // Needs mock
		{http.MethodPatch, "/conversations/1", 0},             // Needs body
		{http.MethodPost, "/conversations/1/participants", 0}, // Needs body
		{http.MethodPost, "/conversations/1/leave", 0},        // Needs mock
		{http.MethodPost, "/conversations/1/messages", 0},     // Needs body
		{http.MethodGet, "/conversations/1/messages", 0},      // Needs mock
	}

	// Setup broad mocks
	convRepo.On("GetByID", mock.Anything, mock.Anything).Return(conv, nil).Maybe()
	convRepo.On("GetDirectConversation", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil).Maybe()
	convRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()
	convRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Maybe()
	convRepo.On("RemoveParticipant", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	convRepo.On("GetParticipant", mock.Anything, mock.Anything, mock.Anything).Return(participant, nil).Maybe()
	convRepo.On("UpdateLastRead", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	convRepo.On("GetUnreadCount", mock.Anything, mock.Anything, mock.Anything).Return(0, nil).Maybe()
	convRepo.On("List", mock.Anything, mock.Anything).Return([]*entities.Conversation{}, int64(0), nil).Maybe()
	convRepo.On("AddParticipant", mock.Anything, mock.Anything).Return(nil).Maybe()
	msgRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()
	msgRepo.On("GetLastMessage", mock.Anything, mock.Anything).Return(nil, nil).Maybe()
	msgRepo.On("GetByID", mock.Anything, mock.Anything).Return(makeMessage(1, 1, 1), nil).Maybe()
	msgRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Maybe()
	msgRepo.On("List", mock.Anything, mock.Anything).Return([]*entities.Message{}, nil).Maybe()
	msgRepo.On("Search", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.Message{}, int64(0), nil).Maybe()

	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		// All endpoints should not return 404 (they are registered)
		require.NotEqual(t, http.StatusNotFound, w.Code, "Endpoint %s %s returned 404", tt.method, tt.path)
	}
}

func TestGetMessages_WithMessages(t *testing.T) {
	handler, convRepo, msgRepo := setupTestHandler(t)
	router := setupRouter(handler, 1)

	participant := &entities.Participant{ID: 1, UserID: 1}
	msgs := []*entities.Message{makeMessage(1, 1, 1), makeMessage(2, 1, 2)}
	convRepo.On("GetParticipant", mock.Anything, int64(1), int64(1)).Return(participant, nil)
	msgRepo.On("List", mock.Anything, mock.AnythingOfType("entities.MessageFilter")).Return(msgs, nil)

	req := httptest.NewRequest(http.MethodGet, "/conversations/1/messages", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
}

func TestListConversations_WithConversations(t *testing.T) {
	handler, convRepo, _ := setupTestHandler(t)
	router := setupRouter(handler, 1)

	convs := []*entities.Conversation{makeGroupConversation(1)}
	convRepo.On("List", mock.Anything, mock.AnythingOfType("entities.ConversationFilter")).
		Return(convs, int64(1), nil)

	req := httptest.NewRequest(http.MethodGet, "/conversations", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
}

func TestSearchMessages_WithResults(t *testing.T) {
	handler, convRepo, msgRepo := setupTestHandler(t)
	router := setupRouter(handler, 1)

	participant := &entities.Participant{ID: 1, UserID: 1}
	msgs := []*entities.Message{makeMessage(1, 1, 1)}
	convRepo.On("GetParticipant", mock.Anything, int64(1), int64(1)).Return(participant, nil)
	msgRepo.On("Search", mock.Anything, int64(1), "test", 20, 0).
		Return(msgs, int64(1), nil)

	req := httptest.NewRequest(http.MethodGet, "/conversations/1/messages/search?q=test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
}

func TestCreateGroupConversation_Unauthorized(t *testing.T) {
	handler, _, _ := setupTestHandler(t)
	router := setupRouterNoAuth(handler)

	body, _ := json.Marshal(map[string]any{
		"title":           "Group",
		"participant_ids": []int64{2},
	})
	req := httptest.NewRequest(http.MethodPost, "/conversations/group", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUpdateConversation_Unauthorized(t *testing.T) {
	handler, _, _ := setupTestHandler(t)
	router := setupRouterNoAuth(handler)

	body, _ := json.Marshal(map[string]any{"title": "Updated"})
	req := httptest.NewRequest(http.MethodPatch, "/conversations/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAddParticipants_Unauthorized(t *testing.T) {
	handler, _, _ := setupTestHandler(t)
	router := setupRouterNoAuth(handler)

	body, _ := json.Marshal(map[string]any{"user_ids": []int64{3}})
	req := httptest.NewRequest(http.MethodPost, "/conversations/1/participants", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLeaveConversation_Unauthorized(t *testing.T) {
	handler, _, _ := setupTestHandler(t)
	router := setupRouterNoAuth(handler)

	req := httptest.NewRequest(http.MethodPost, "/conversations/1/leave", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestEditMessage_Unauthorized(t *testing.T) {
	handler, _, _ := setupTestHandler(t)
	router := setupRouterNoAuth(handler)

	body, _ := json.Marshal(map[string]any{"content": "Updated"})
	req := httptest.NewRequest(http.MethodPatch, "/conversations/1/messages/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestDeleteMessage_Unauthorized(t *testing.T) {
	handler, _, _ := setupTestHandler(t)
	router := setupRouterNoAuth(handler)

	req := httptest.NewRequest(http.MethodDelete, "/conversations/1/messages/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMarkAsRead_Unauthorized(t *testing.T) {
	handler, _, _ := setupTestHandler(t)
	router := setupRouterNoAuth(handler)

	body, _ := json.Marshal(map[string]any{"message_id": 10})
	req := httptest.NewRequest(http.MethodPost, "/conversations/1/read", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSearchMessages_Unauthorized(t *testing.T) {
	handler, _, _ := setupTestHandler(t)
	router := setupRouterNoAuth(handler)

	req := httptest.NewRequest(http.MethodGet, "/conversations/1/messages/search?q=hello", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleWebSocket_Unauthorized(t *testing.T) {
	handler, _, _ := setupTestHandler(t)
	router := setupRouterNoAuth(handler)

	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
