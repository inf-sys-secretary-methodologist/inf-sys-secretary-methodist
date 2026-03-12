package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/entities"
	notifHttp "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/interfaces/http"
)

const invalidValue = "invalid"

// MockWebPushRepository is a mock implementation of WebPushRepository
type MockWebPushRepository struct {
	mock.Mock
}

func (m *MockWebPushRepository) Create(ctx context.Context, sub *entities.WebPushSubscription) error {
	args := m.Called(ctx, sub)
	if sub != nil && sub.ID == 0 {
		sub.ID = 1
		sub.CreatedAt = time.Now()
	}
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

// MockWebPushService is a mock implementation of WebPushService
type MockWebPushService struct {
	mock.Mock
}

func (m *MockWebPushService) SendNotification(ctx context.Context, sub *entities.WebPushSubscription, payload *entities.WebPushPayload) error {
	args := m.Called(ctx, sub, payload)
	return args.Error(0)
}

func (m *MockWebPushService) SendToUser(ctx context.Context, userID int64, payload *entities.WebPushPayload) error {
	args := m.Called(ctx, userID, payload)
	return args.Error(0)
}

func (m *MockWebPushService) GetVAPIDPublicKey() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockWebPushService) IsConfigured() bool {
	args := m.Called()
	return args.Bool(0)
}

// mockAuthMiddlewareWebPush creates a middleware that sets user_id for testing
func mockAuthMiddlewareWebPush(userID int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	}
}

// setupRouter creates a test router with WebPush handler
func setupRouter(repo *MockWebPushRepository, service *MockWebPushService, authenticated bool, userID int64) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	if authenticated {
		router.Use(mockAuthMiddlewareWebPush(userID))
	}

	handler := notifHttp.NewWebPushHandler(repo, service)

	pushGroup := router.Group("/push")
	{
		pushGroup.GET("/vapid-key", handler.GetVAPIDKey)
		pushGroup.POST("/subscribe", handler.Subscribe)
		pushGroup.POST("/unsubscribe", handler.Unsubscribe)
		pushGroup.GET("/status", handler.GetStatus)
		pushGroup.DELETE("/subscriptions/:id", handler.DeleteSubscription)
		pushGroup.POST("/test", handler.TestPush)
	}

	return router
}

// TestGetVAPIDKey tests the GetVAPIDKey endpoint
func TestGetVAPIDKey(t *testing.T) {
	tests := []struct {
		name           string
		configured     bool
		expectedStatus int
	}{
		{
			name:           "success - configured",
			configured:     true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "error - not configured",
			configured:     false,
			expectedStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockWebPushRepository)
			mockService := new(MockWebPushService)

			mockService.On("IsConfigured").Return(tt.configured)
			if tt.configured {
				mockService.On("GetVAPIDPublicKey").Return("test-public-key")
			}

			router := setupRouter(mockRepo, mockService, false, 0)

			req := httptest.NewRequest(http.MethodGet, "/push/vapid-key", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.configured {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "test-public-key", response["public_key"])
			}
		})
	}
}

// TestSubscribe tests the Subscribe endpoint
func TestSubscribe(t *testing.T) {
	tests := []struct {
		name           string
		authenticated  bool
		configured     bool
		body           map[string]any
		expectedStatus int
	}{
		{
			name:          "success",
			authenticated: true,
			configured:    true,
			body: map[string]any{
				"endpoint": "https://push.example.com/sub1",
				"p256dh":   "test-p256dh-key",
				"auth":     "test-auth-key",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized",
			authenticated:  false,
			configured:     true,
			body:           map[string]any{},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "not configured",
			authenticated:  true,
			configured:     false,
			body:           map[string]any{},
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name:          "missing endpoint",
			authenticated: true,
			configured:    true,
			body: map[string]any{
				"p256dh": "test-p256dh-key",
				"auth":   "test-auth-key",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:          "missing p256dh",
			authenticated: true,
			configured:    true,
			body: map[string]any{
				"endpoint": "https://push.example.com/sub1",
				"auth":     "test-auth-key",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:          "missing auth",
			authenticated: true,
			configured:    true,
			body: map[string]any{
				"endpoint": "https://push.example.com/sub1",
				"p256dh":   "test-p256dh-key",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockWebPushRepository)
			mockService := new(MockWebPushService)

			if tt.authenticated {
				mockService.On("IsConfigured").Return(tt.configured)
			}

			if tt.authenticated && tt.configured && tt.expectedStatus == http.StatusOK {
				mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
			}

			router := setupRouter(mockRepo, mockService, tt.authenticated, 1)

			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/push/subscribe", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestUnsubscribe tests the Unsubscribe endpoint
func TestUnsubscribe(t *testing.T) {
	tests := []struct {
		name           string
		authenticated  bool
		body           map[string]any
		expectedStatus int
	}{
		{
			name:          "success",
			authenticated: true,
			body: map[string]any{
				"endpoint": "https://push.example.com/sub1",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized",
			authenticated:  false,
			body:           map[string]any{},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "missing endpoint",
			authenticated:  true,
			body:           map[string]any{},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockWebPushRepository)
			mockService := new(MockWebPushService)

			if tt.authenticated && tt.expectedStatus == http.StatusOK {
				mockRepo.On("DeleteByEndpoint", mock.Anything, mock.Anything).Return(nil)
			}

			router := setupRouter(mockRepo, mockService, tt.authenticated, 1)

			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/push/unsubscribe", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestGetStatus tests the GetStatus endpoint
func TestGetStatus(t *testing.T) {
	tests := []struct {
		name           string
		authenticated  bool
		subscriptions  []*entities.WebPushSubscription
		expectedStatus int
	}{
		{
			name:          "success - with subscriptions",
			authenticated: true,
			subscriptions: []*entities.WebPushSubscription{
				{ID: 1, UserID: 1, IsActive: true, CreatedAt: time.Now()},
				{ID: 2, UserID: 1, IsActive: false, CreatedAt: time.Now()},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "success - no subscriptions",
			authenticated:  true,
			subscriptions:  []*entities.WebPushSubscription{},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized",
			authenticated:  false,
			subscriptions:  nil,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockWebPushRepository)
			mockService := new(MockWebPushService)

			if tt.authenticated {
				mockRepo.On("GetByUserID", mock.Anything, int64(1)).Return(tt.subscriptions, nil)
			}

			router := setupRouter(mockRepo, mockService, tt.authenticated, 1)

			req := httptest.NewRequest(http.MethodGet, "/push/status", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]any
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response, "is_enabled")
				assert.Contains(t, response, "subscriptions")
				assert.Contains(t, response, "total_devices")
			}
		})
	}
}

// TestDeleteSubscription tests the DeleteSubscription endpoint
func TestDeleteSubscription(t *testing.T) {
	tests := []struct {
		name           string
		authenticated  bool
		subscriptionID string
		ownerID        int64
		userID         int64
		subscription   *entities.WebPushSubscription
		expectedStatus int
	}{
		{
			name:           "success",
			authenticated:  true,
			subscriptionID: "1",
			ownerID:        1,
			userID:         1,
			subscription:   &entities.WebPushSubscription{ID: 1, UserID: 1},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized",
			authenticated:  false,
			subscriptionID: "1",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid id",
			authenticated:  true,
			subscriptionID: invalidValue,
			userID:         1,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "not found",
			authenticated:  true,
			subscriptionID: "999",
			userID:         1,
			subscription:   nil,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "forbidden - different owner",
			authenticated:  true,
			subscriptionID: "1",
			ownerID:        2,
			userID:         1,
			subscription:   &entities.WebPushSubscription{ID: 1, UserID: 2},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockWebPushRepository)
			mockService := new(MockWebPushService)

			if tt.authenticated && tt.subscriptionID != invalidValue {
				mockRepo.On("GetByID", mock.Anything, mock.Anything).Return(tt.subscription, nil)
				if tt.subscription != nil && tt.subscription.UserID == tt.userID {
					mockRepo.On("Delete", mock.Anything, mock.Anything).Return(nil)
				}
			}

			router := setupRouter(mockRepo, mockService, tt.authenticated, tt.userID)

			req := httptest.NewRequest(http.MethodDelete, "/push/subscriptions/"+tt.subscriptionID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestTestPush tests the TestPush endpoint
func TestTestPush(t *testing.T) {
	tests := []struct {
		name           string
		authenticated  bool
		configured     bool
		body           map[string]any
		expectedStatus int
	}{
		{
			name:          "success - with custom message",
			authenticated: true,
			configured:    true,
			body: map[string]any{
				"title":   "Test Title",
				"message": "Test Message",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "success - default message",
			authenticated:  true,
			configured:     true,
			body:           map[string]any{},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unauthorized",
			authenticated:  false,
			configured:     true,
			body:           map[string]any{},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "not configured",
			authenticated:  true,
			configured:     false,
			body:           map[string]any{},
			expectedStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockWebPushRepository)
			mockService := new(MockWebPushService)

			if tt.authenticated {
				mockService.On("IsConfigured").Return(tt.configured)
			}

			if tt.authenticated && tt.configured {
				mockService.On("SendToUser", mock.Anything, int64(1), mock.Anything).Return(nil)
			}

			router := setupRouter(mockRepo, mockService, tt.authenticated, 1)

			bodyBytes, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/push/test", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
