package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/entities"
)

// MockWebPushRepository is a mock implementation of WebPushRepository
type MockWebPushRepository struct {
	mock.Mock
}

func (m *MockWebPushRepository) Create(ctx context.Context, sub *entities.WebPushSubscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

func (m *MockWebPushRepository) GetByID(ctx context.Context, id int64) (*entities.WebPushSubscription, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.WebPushSubscription), args.Error(1)
}

func (m *MockWebPushRepository) GetByEndpoint(ctx context.Context, endpoint string) (*entities.WebPushSubscription, error) {
	args := m.Called(ctx, endpoint)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.WebPushSubscription), args.Error(1)
}

func (m *MockWebPushRepository) GetByUserID(ctx context.Context, userID int64) ([]*entities.WebPushSubscription, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*entities.WebPushSubscription), args.Error(1)
}

func (m *MockWebPushRepository) GetActiveByUserID(ctx context.Context, userID int64) ([]*entities.WebPushSubscription, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*entities.WebPushSubscription), args.Error(1)
}

func (m *MockWebPushRepository) Update(ctx context.Context, sub *entities.WebPushSubscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

func (m *MockWebPushRepository) UpdateLastUsed(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockWebPushRepository) Deactivate(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockWebPushRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockWebPushRepository) DeleteByEndpoint(ctx context.Context, endpoint string) error {
	args := m.Called(ctx, endpoint)
	return args.Error(0)
}

func (m *MockWebPushRepository) DeleteByUserID(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockWebPushRepository) CountByUserID(ctx context.Context, userID int64) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

// TestNewWebPushService tests service creation
func TestNewWebPushService(t *testing.T) {
	mockRepo := new(MockWebPushRepository)

	service := NewWebPushService(
		mockRepo,
		"test-public-key",
		"test-private-key",
		"mailto:test@example.com",
		nil,
	)

	assert.NotNil(t, service)
}

// TestIsConfigured tests the IsConfigured method
func TestIsConfigured(t *testing.T) {
	mockRepo := new(MockWebPushRepository)

	tests := []struct {
		name        string
		publicKey   string
		privateKey  string
		subject     string
		expected    bool
	}{
		{
			name:       "fully configured",
			publicKey:  "public-key",
			privateKey: "private-key",
			subject:    "mailto:test@example.com",
			expected:   true,
		},
		{
			name:       "missing public key",
			publicKey:  "",
			privateKey: "private-key",
			subject:    "mailto:test@example.com",
			expected:   false,
		},
		{
			name:       "missing private key",
			publicKey:  "public-key",
			privateKey: "",
			subject:    "mailto:test@example.com",
			expected:   false,
		},
		{
			name:       "missing subject",
			publicKey:  "public-key",
			privateKey: "private-key",
			subject:    "",
			expected:   false,
		},
		{
			name:       "all empty",
			publicKey:  "",
			privateKey: "",
			subject:    "",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewWebPushService(
				mockRepo,
				tt.publicKey,
				tt.privateKey,
				tt.subject,
				nil,
			)

			result := service.IsConfigured()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetVAPIDPublicKey tests the GetVAPIDPublicKey method
func TestGetVAPIDPublicKey(t *testing.T) {
	mockRepo := new(MockWebPushRepository)
	expectedKey := "BNpqXjYJ5NQ3zCFmXHc0GDL7KWJ5EBHdXxMOjT7OMjSWMw"

	service := NewWebPushService(
		mockRepo,
		expectedKey,
		"private-key",
		"mailto:test@example.com",
		nil,
	)

	result := service.GetVAPIDPublicKey()
	assert.Equal(t, expectedKey, result)
}

// TestSendNotification_NotConfigured tests sending when not configured
func TestSendNotification_NotConfigured(t *testing.T) {
	mockRepo := new(MockWebPushRepository)

	service := NewWebPushService(
		mockRepo,
		"", // Empty public key
		"",
		"",
		nil,
	)

	sub := entities.NewWebPushSubscription(1, "https://push.example.com", "p256dh", "auth")
	payload := entities.NewWebPushPayload("Test", "Message")

	err := service.SendNotification(context.Background(), sub, payload)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")
}

// TestSendToUser_NotConfigured tests sending to user when not configured
func TestSendToUser_NotConfigured(t *testing.T) {
	mockRepo := new(MockWebPushRepository)

	service := NewWebPushService(
		mockRepo,
		"", // Empty public key
		"",
		"",
		nil,
	)

	payload := entities.NewWebPushPayload("Test", "Message")

	err := service.SendToUser(context.Background(), 1, payload)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")
}

// TestSendToUser_NoSubscriptions tests sending when user has no subscriptions
func TestSendToUser_NoSubscriptions(t *testing.T) {
	mockRepo := new(MockWebPushRepository)

	// Return empty list
	mockRepo.On("GetActiveByUserID", mock.Anything, int64(1)).
		Return([]*entities.WebPushSubscription{}, nil)

	service := NewWebPushService(
		mockRepo,
		"public-key",
		"private-key",
		"mailto:test@example.com",
		nil,
	)

	payload := entities.NewWebPushPayload("Test", "Message")

	err := service.SendToUser(context.Background(), 1, payload)
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

// TestSendToUser_GetSubscriptionsError tests error handling when getting subscriptions fails
func TestSendToUser_GetSubscriptionsError(t *testing.T) {
	mockRepo := new(MockWebPushRepository)

	// Return error
	mockRepo.On("GetActiveByUserID", mock.Anything, int64(1)).
		Return([]*entities.WebPushSubscription{}, assert.AnError)

	service := NewWebPushService(
		mockRepo,
		"public-key",
		"private-key",
		"mailto:test@example.com",
		nil,
	)

	payload := entities.NewWebPushPayload("Test", "Message")

	err := service.SendToUser(context.Background(), 1, payload)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get user subscriptions")

	mockRepo.AssertExpectations(t)
}

// TestWebPushPayload tests payload creation and methods
func TestWebPushPayload(t *testing.T) {
	payload := entities.NewWebPushPayload("Test Title", "Test Body")

	assert.Equal(t, "Test Title", payload.Title)
	assert.Equal(t, "Test Body", payload.Body)
	assert.NotEmpty(t, payload.Icon)
	assert.NotEmpty(t, payload.Badge)

	// Test builder methods
	payload.WithURL("/test-url")
	assert.Equal(t, "/test-url", payload.URL)

	payload.WithTag("test-tag")
	assert.Equal(t, "test-tag", payload.Tag)

	payload.WithRequireInteraction(true)
	assert.True(t, payload.RequireInteraction)

	payload.WithData(map[string]any{"key": "value"})
	assert.Equal(t, "value", payload.Data["key"])

	payload.AddAction("action1", "Action 1")
	assert.Len(t, payload.Actions, 1)
	assert.Equal(t, "action1", payload.Actions[0].Action)
}

// TestWebPushPayloadFromNotification tests creating payload from notification
func TestWebPushPayloadFromNotification(t *testing.T) {
	notification := &entities.Notification{
		ID:       1,
		UserID:   100,
		Type:     entities.NotificationTypeTask,
		Priority: entities.PriorityHigh,
		Title:    "Task Notification",
		Message:  "You have a new task",
		Link:     "/tasks/123",
	}

	payload := entities.WebPushPayloadFromNotification(notification)

	assert.Equal(t, notification.Title, payload.Title)
	assert.Equal(t, notification.Message, payload.Body)
	assert.Equal(t, notification.Link, payload.URL)
	assert.Equal(t, string(notification.Type), payload.Tag)
	assert.True(t, payload.RequireInteraction) // High priority
	assert.Equal(t, notification.ID, payload.Data["notification_id"])
	assert.Equal(t, string(notification.Type), payload.Data["type"])
	assert.Equal(t, string(notification.Priority), payload.Data["priority"])
}
