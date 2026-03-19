package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/application/services"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/entities"
	domainServices "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/services"
	notifHttp "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/interfaces/http"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/telegram"
)

// --- Mock Notification Repository ---

type MockNotificationRepository struct {
	mock.Mock
}

func (m *MockNotificationRepository) Create(ctx context.Context, n *entities.Notification) error {
	args := m.Called(ctx, n)
	return args.Error(0)
}

func (m *MockNotificationRepository) Update(ctx context.Context, n *entities.Notification) error {
	args := m.Called(ctx, n)
	return args.Error(0)
}

func (m *MockNotificationRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockNotificationRepository) GetByID(ctx context.Context, id int64) (*entities.Notification, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Notification), args.Error(1)
}

func (m *MockNotificationRepository) List(ctx context.Context, filter *entities.NotificationFilter) ([]*entities.Notification, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*entities.Notification), args.Error(1)
}

func (m *MockNotificationRepository) GetByUserID(ctx context.Context, userID int64, limit, offset int) ([]*entities.Notification, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]*entities.Notification), args.Error(1)
}

func (m *MockNotificationRepository) GetUnreadByUserID(ctx context.Context, userID int64) ([]*entities.Notification, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*entities.Notification), args.Error(1)
}

func (m *MockNotificationRepository) MarkAsRead(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockNotificationRepository) MarkAllAsRead(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockNotificationRepository) DeleteByUserID(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockNotificationRepository) DeleteExpired(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockNotificationRepository) GetUnreadCount(ctx context.Context, userID int64) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockNotificationRepository) GetStats(ctx context.Context, userID int64) (*entities.NotificationStats, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.NotificationStats), args.Error(1)
}

func (m *MockNotificationRepository) CreateBulk(ctx context.Context, notifications []*entities.Notification) error {
	args := m.Called(ctx, notifications)
	return args.Error(0)
}

// --- Mock Preferences Repository ---

type MockPreferencesRepository struct {
	mock.Mock
}

func (m *MockPreferencesRepository) Create(ctx context.Context, p *entities.UserNotificationPreferences) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *MockPreferencesRepository) Update(ctx context.Context, p *entities.UserNotificationPreferences) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *MockPreferencesRepository) Delete(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockPreferencesRepository) GetByUserID(ctx context.Context, userID int64) (*entities.UserNotificationPreferences, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.UserNotificationPreferences), args.Error(1)
}

func (m *MockPreferencesRepository) GetOrCreate(ctx context.Context, userID int64) (*entities.UserNotificationPreferences, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.UserNotificationPreferences), args.Error(1)
}

func (m *MockPreferencesRepository) UpdateChannelEnabled(ctx context.Context, userID int64, channel entities.NotificationChannel, enabled bool) error {
	args := m.Called(ctx, userID, channel, enabled)
	return args.Error(0)
}

func (m *MockPreferencesRepository) UpdateQuietHours(ctx context.Context, userID int64, enabled bool, start, end, timezone string) error {
	args := m.Called(ctx, userID, enabled, start, end, timezone)
	return args.Error(0)
}

// --- Mock Telegram Repository ---

type MockTelegramRepository struct {
	mock.Mock
}

func (m *MockTelegramRepository) CreateVerificationCode(ctx context.Context, code *entities.TelegramVerificationCode) error {
	args := m.Called(ctx, code)
	return args.Error(0)
}

func (m *MockTelegramRepository) GetVerificationCodeByCode(ctx context.Context, code string) (*entities.TelegramVerificationCode, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.TelegramVerificationCode), args.Error(1)
}

func (m *MockTelegramRepository) GetActiveVerificationCodeByUserID(ctx context.Context, userID int64) (*entities.TelegramVerificationCode, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.TelegramVerificationCode), args.Error(1)
}

func (m *MockTelegramRepository) MarkCodeAsUsed(ctx context.Context, codeID int64) error {
	args := m.Called(ctx, codeID)
	return args.Error(0)
}

func (m *MockTelegramRepository) DeleteExpiredCodes(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockTelegramRepository) CreateConnection(ctx context.Context, conn *entities.TelegramConnection) error {
	args := m.Called(ctx, conn)
	return args.Error(0)
}

func (m *MockTelegramRepository) GetConnectionByUserID(ctx context.Context, userID int64) (*entities.TelegramConnection, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.TelegramConnection), args.Error(1)
}

func (m *MockTelegramRepository) GetConnectionByChatID(ctx context.Context, chatID int64) (*entities.TelegramConnection, error) {
	args := m.Called(ctx, chatID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.TelegramConnection), args.Error(1)
}

func (m *MockTelegramRepository) GetActiveConnections(ctx context.Context) ([]entities.TelegramConnection, error) {
	args := m.Called(ctx)
	return args.Get(0).([]entities.TelegramConnection), args.Error(1)
}

func (m *MockTelegramRepository) UpdateConnection(ctx context.Context, conn *entities.TelegramConnection) error {
	args := m.Called(ctx, conn)
	return args.Error(0)
}

func (m *MockTelegramRepository) DeleteConnection(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// --- Mock Telegram Domain Service ---

type MockTelegramDomainService struct {
	mock.Mock
}

func (m *MockTelegramDomainService) SendMessage(ctx context.Context, req *domainServices.SendTelegramMessageRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockTelegramDomainService) SendNotification(ctx context.Context, chatID string, title, message string, priority string) error {
	args := m.Called(ctx, chatID, title, message, priority)
	return args.Error(0)
}

// --- Helpers ---

func setAuthMiddleware(userID int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("role", "student")
		c.Next()
	}
}

func defaultPrefs(userID int64) *entities.UserNotificationPreferences {
	return &entities.UserNotificationPreferences{
		ID:           1,
		UserID:       userID,
		EmailEnabled: true,
		PushEnabled:  true,
		InAppEnabled: true,
		Timezone:     "Europe/Moscow",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// ===================== NOTIFICATION HANDLER TESTS =====================

func TestNotificationHandler_List(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockTgRepo := new(MockTelegramRepository)
		uc := usecases.NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, mockTgRepo, nil, nil, nil)
		handler := notifHttp.NewNotificationHandler(uc)

		notifications := []*entities.Notification{
			{ID: 1, UserID: 1, Type: entities.NotificationTypeSystem, Title: "Test", Message: "Msg", CreatedAt: time.Now()},
		}
		mockNotifRepo.On("List", mock.Anything, mock.Anything).Return(notifications, nil)
		mockNotifRepo.On("GetUnreadCount", mock.Anything, int64(1)).Return(int64(1), nil)
		mockNotifRepo.On("GetStats", mock.Anything, int64(1)).Return(&entities.NotificationStats{TotalCount: 1}, nil)

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.GET("/notifications", handler.List)

		req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("success with all query params", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockTgRepo := new(MockTelegramRepository)
		uc := usecases.NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, mockTgRepo, nil, nil, nil)
		handler := notifHttp.NewNotificationHandler(uc)

		mockNotifRepo.On("List", mock.Anything, mock.Anything).Return([]*entities.Notification{}, nil)
		mockNotifRepo.On("GetUnreadCount", mock.Anything, int64(1)).Return(int64(0), nil)
		mockNotifRepo.On("GetStats", mock.Anything, int64(1)).Return(&entities.NotificationStats{}, nil)

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.GET("/notifications", handler.List)

		req := httptest.NewRequest(http.MethodGet, "/notifications?type=system&priority=high&is_read=true&limit=10&offset=5", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("success with invalid limit and offset", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockTgRepo := new(MockTelegramRepository)
		uc := usecases.NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, mockTgRepo, nil, nil, nil)
		handler := notifHttp.NewNotificationHandler(uc)

		mockNotifRepo.On("List", mock.Anything, mock.Anything).Return([]*entities.Notification{}, nil)
		mockNotifRepo.On("GetUnreadCount", mock.Anything, int64(1)).Return(int64(0), nil)
		mockNotifRepo.On("GetStats", mock.Anything, int64(1)).Return(&entities.NotificationStats{}, nil)

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.GET("/notifications", handler.List)

		req := httptest.NewRequest(http.MethodGet, "/notifications?limit=abc&offset=xyz", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		uc := usecases.NewNotificationUseCase(nil, nil, nil, nil, nil, nil)
		handler := notifHttp.NewNotificationHandler(uc)

		router := gin.New()
		router.GET("/notifications", handler.List)

		req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("usecase error", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockTgRepo := new(MockTelegramRepository)
		uc := usecases.NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, mockTgRepo, nil, nil, nil)
		handler := notifHttp.NewNotificationHandler(uc)

		mockNotifRepo.On("List", mock.Anything, mock.Anything).Return([]*entities.Notification{}, errors.New("db error"))

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.GET("/notifications", handler.List)

		req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestNotificationHandler_GetByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		uc := usecases.NewNotificationUseCase(mockNotifRepo, nil, nil, nil, nil, nil)
		handler := notifHttp.NewNotificationHandler(uc)

		n := &entities.Notification{ID: 1, UserID: 1, Title: "Test", Message: "Msg", Type: entities.NotificationTypeSystem, CreatedAt: time.Now()}
		mockNotifRepo.On("GetByID", mock.Anything, int64(1)).Return(n, nil)

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.GET("/notifications/:id", handler.GetByID)

		req := httptest.NewRequest(http.MethodGet, "/notifications/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("not found", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		uc := usecases.NewNotificationUseCase(mockNotifRepo, nil, nil, nil, nil, nil)
		handler := notifHttp.NewNotificationHandler(uc)

		mockNotifRepo.On("GetByID", mock.Anything, int64(999)).Return(nil, nil)

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.GET("/notifications/:id", handler.GetByID)

		req := httptest.NewRequest(http.MethodGet, "/notifications/999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		uc := usecases.NewNotificationUseCase(nil, nil, nil, nil, nil, nil)
		handler := notifHttp.NewNotificationHandler(uc)

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.GET("/notifications/:id", handler.GetByID)

		req := httptest.NewRequest(http.MethodGet, "/notifications/abc", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("usecase error", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		uc := usecases.NewNotificationUseCase(mockNotifRepo, nil, nil, nil, nil, nil)
		handler := notifHttp.NewNotificationHandler(uc)

		mockNotifRepo.On("GetByID", mock.Anything, int64(1)).Return(nil, errors.New("db error"))

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.GET("/notifications/:id", handler.GetByID)

		req := httptest.NewRequest(http.MethodGet, "/notifications/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestNotificationHandler_MarkAsRead(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		uc := usecases.NewNotificationUseCase(mockNotifRepo, nil, nil, nil, nil, nil)
		handler := notifHttp.NewNotificationHandler(uc)

		mockNotifRepo.On("MarkAsRead", mock.Anything, int64(1)).Return(nil)

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.PUT("/notifications/:id/read", handler.MarkAsRead)

		req := httptest.NewRequest(http.MethodPut, "/notifications/1/read", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		uc := usecases.NewNotificationUseCase(nil, nil, nil, nil, nil, nil)
		handler := notifHttp.NewNotificationHandler(uc)

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.PUT("/notifications/:id/read", handler.MarkAsRead)

		req := httptest.NewRequest(http.MethodPut, "/notifications/abc/read", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("usecase error", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		uc := usecases.NewNotificationUseCase(mockNotifRepo, nil, nil, nil, nil, nil)
		handler := notifHttp.NewNotificationHandler(uc)

		mockNotifRepo.On("MarkAsRead", mock.Anything, int64(1)).Return(errors.New("db error"))

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.PUT("/notifications/:id/read", handler.MarkAsRead)

		req := httptest.NewRequest(http.MethodPut, "/notifications/1/read", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestNotificationHandler_MarkAllAsRead(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		uc := usecases.NewNotificationUseCase(mockNotifRepo, nil, nil, nil, nil, nil)
		handler := notifHttp.NewNotificationHandler(uc)

		mockNotifRepo.On("MarkAllAsRead", mock.Anything, int64(1)).Return(nil)

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.PUT("/notifications/read-all", handler.MarkAllAsRead)

		req := httptest.NewRequest(http.MethodPut, "/notifications/read-all", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		uc := usecases.NewNotificationUseCase(nil, nil, nil, nil, nil, nil)
		handler := notifHttp.NewNotificationHandler(uc)

		router := gin.New()
		router.PUT("/notifications/read-all", handler.MarkAllAsRead)

		req := httptest.NewRequest(http.MethodPut, "/notifications/read-all", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("usecase error", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		uc := usecases.NewNotificationUseCase(mockNotifRepo, nil, nil, nil, nil, nil)
		handler := notifHttp.NewNotificationHandler(uc)

		mockNotifRepo.On("MarkAllAsRead", mock.Anything, int64(1)).Return(errors.New("db error"))

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.PUT("/notifications/read-all", handler.MarkAllAsRead)

		req := httptest.NewRequest(http.MethodPut, "/notifications/read-all", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestNotificationHandler_Delete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		uc := usecases.NewNotificationUseCase(mockNotifRepo, nil, nil, nil, nil, nil)
		handler := notifHttp.NewNotificationHandler(uc)

		mockNotifRepo.On("Delete", mock.Anything, int64(1)).Return(nil)

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.DELETE("/notifications/:id", handler.Delete)

		req := httptest.NewRequest(http.MethodDelete, "/notifications/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		uc := usecases.NewNotificationUseCase(nil, nil, nil, nil, nil, nil)
		handler := notifHttp.NewNotificationHandler(uc)

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.DELETE("/notifications/:id", handler.Delete)

		req := httptest.NewRequest(http.MethodDelete, "/notifications/abc", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("usecase error", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		uc := usecases.NewNotificationUseCase(mockNotifRepo, nil, nil, nil, nil, nil)
		handler := notifHttp.NewNotificationHandler(uc)

		mockNotifRepo.On("Delete", mock.Anything, int64(1)).Return(errors.New("db error"))

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.DELETE("/notifications/:id", handler.Delete)

		req := httptest.NewRequest(http.MethodDelete, "/notifications/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestNotificationHandler_DeleteAll(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		uc := usecases.NewNotificationUseCase(mockNotifRepo, nil, nil, nil, nil, nil)
		handler := notifHttp.NewNotificationHandler(uc)

		mockNotifRepo.On("DeleteByUserID", mock.Anything, int64(1)).Return(nil)

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.DELETE("/notifications", handler.DeleteAll)

		req := httptest.NewRequest(http.MethodDelete, "/notifications", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		uc := usecases.NewNotificationUseCase(nil, nil, nil, nil, nil, nil)
		handler := notifHttp.NewNotificationHandler(uc)

		router := gin.New()
		router.DELETE("/notifications", handler.DeleteAll)

		req := httptest.NewRequest(http.MethodDelete, "/notifications", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("usecase error", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		uc := usecases.NewNotificationUseCase(mockNotifRepo, nil, nil, nil, nil, nil)
		handler := notifHttp.NewNotificationHandler(uc)

		mockNotifRepo.On("DeleteByUserID", mock.Anything, int64(1)).Return(errors.New("db error"))

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.DELETE("/notifications", handler.DeleteAll)

		req := httptest.NewRequest(http.MethodDelete, "/notifications", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestNotificationHandler_GetUnreadCount(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		uc := usecases.NewNotificationUseCase(mockNotifRepo, nil, nil, nil, nil, nil)
		handler := notifHttp.NewNotificationHandler(uc)

		mockNotifRepo.On("GetUnreadCount", mock.Anything, int64(1)).Return(int64(5), nil)

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.GET("/notifications/unread-count", handler.GetUnreadCount)

		req := httptest.NewRequest(http.MethodGet, "/notifications/unread-count", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp dto.UnreadCountOutput
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, int64(5), resp.Count)
	})

	t.Run("unauthorized", func(t *testing.T) {
		uc := usecases.NewNotificationUseCase(nil, nil, nil, nil, nil, nil)
		handler := notifHttp.NewNotificationHandler(uc)

		router := gin.New()
		router.GET("/notifications/unread-count", handler.GetUnreadCount)

		req := httptest.NewRequest(http.MethodGet, "/notifications/unread-count", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("usecase error", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		uc := usecases.NewNotificationUseCase(mockNotifRepo, nil, nil, nil, nil, nil)
		handler := notifHttp.NewNotificationHandler(uc)

		mockNotifRepo.On("GetUnreadCount", mock.Anything, int64(1)).Return(int64(0), errors.New("db error"))

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.GET("/notifications/unread-count", handler.GetUnreadCount)

		req := httptest.NewRequest(http.MethodGet, "/notifications/unread-count", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestNotificationHandler_GetStats(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		uc := usecases.NewNotificationUseCase(mockNotifRepo, nil, nil, nil, nil, nil)
		handler := notifHttp.NewNotificationHandler(uc)

		stats := &entities.NotificationStats{TotalCount: 10, UnreadCount: 5}
		mockNotifRepo.On("GetStats", mock.Anything, int64(1)).Return(stats, nil)

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.GET("/notifications/stats", handler.GetStats)

		req := httptest.NewRequest(http.MethodGet, "/notifications/stats", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		uc := usecases.NewNotificationUseCase(nil, nil, nil, nil, nil, nil)
		handler := notifHttp.NewNotificationHandler(uc)

		router := gin.New()
		router.GET("/notifications/stats", handler.GetStats)

		req := httptest.NewRequest(http.MethodGet, "/notifications/stats", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("usecase error", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		uc := usecases.NewNotificationUseCase(mockNotifRepo, nil, nil, nil, nil, nil)
		handler := notifHttp.NewNotificationHandler(uc)

		mockNotifRepo.On("GetStats", mock.Anything, int64(1)).Return(nil, errors.New("db error"))

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.GET("/notifications/stats", handler.GetStats)

		req := httptest.NewRequest(http.MethodGet, "/notifications/stats", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestNotificationHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		uc := usecases.NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, nil, nil, nil, nil)
		handler := notifHttp.NewNotificationHandler(uc)

		mockNotifRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.POST("/notifications", handler.Create)

		payload := dto.CreateNotificationInput{
			UserID:  2,
			Type:    entities.NotificationTypeSystem,
			Title:   "Test",
			Message: "Test message",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/notifications", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		uc := usecases.NewNotificationUseCase(nil, nil, nil, nil, nil, nil)
		handler := notifHttp.NewNotificationHandler(uc)

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.POST("/notifications", handler.Create)

		req := httptest.NewRequest(http.MethodPost, "/notifications", bytes.NewReader([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("validation error", func(t *testing.T) {
		uc := usecases.NewNotificationUseCase(nil, nil, nil, nil, nil, nil)
		handler := notifHttp.NewNotificationHandler(uc)

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.POST("/notifications", handler.Create)

		// Missing required fields
		payload := map[string]interface{}{
			"title": "Test",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/notifications", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("usecase error", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		uc := usecases.NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, nil, nil, nil, nil)
		handler := notifHttp.NewNotificationHandler(uc)

		mockNotifRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.POST("/notifications", handler.Create)

		payload := dto.CreateNotificationInput{
			UserID:  2,
			Type:    entities.NotificationTypeSystem,
			Title:   "Test",
			Message: "Test message",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/notifications", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestNotificationHandler_CreateBulk(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		uc := usecases.NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, nil, nil, nil, nil)
		handler := notifHttp.NewNotificationHandler(uc)

		mockNotifRepo.On("CreateBulk", mock.Anything, mock.Anything).Return(nil)

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.POST("/notifications/bulk", handler.CreateBulk)

		payload := dto.CreateBulkNotificationInput{
			UserIDs: []int64{1, 2, 3},
			Type:    entities.NotificationTypeSystem,
			Title:   "Bulk Test",
			Message: "Bulk message",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/notifications/bulk", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		uc := usecases.NewNotificationUseCase(nil, nil, nil, nil, nil, nil)
		handler := notifHttp.NewNotificationHandler(uc)

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.POST("/notifications/bulk", handler.CreateBulk)

		req := httptest.NewRequest(http.MethodPost, "/notifications/bulk", bytes.NewReader([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("validation error", func(t *testing.T) {
		uc := usecases.NewNotificationUseCase(nil, nil, nil, nil, nil, nil)
		handler := notifHttp.NewNotificationHandler(uc)

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.POST("/notifications/bulk", handler.CreateBulk)

		payload := map[string]interface{}{
			"title": "Test",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/notifications/bulk", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("usecase error", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		uc := usecases.NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, nil, nil, nil, nil)
		handler := notifHttp.NewNotificationHandler(uc)

		mockNotifRepo.On("CreateBulk", mock.Anything, mock.Anything).Return(errors.New("db error"))

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.POST("/notifications/bulk", handler.CreateBulk)

		payload := dto.CreateBulkNotificationInput{
			UserIDs: []int64{1, 2},
			Type:    entities.NotificationTypeSystem,
			Title:   "Test",
			Message: "Test",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/notifications/bulk", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// ===================== PREFERENCES HANDLER TESTS =====================

func TestPreferencesHandler_Get(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		mockPrefsRepo := new(MockPreferencesRepository)
		uc := usecases.NewPreferencesUseCase(mockPrefsRepo)
		handler := notifHttp.NewPreferencesHandler(uc)

		prefs := defaultPrefs(1)
		mockPrefsRepo.On("GetOrCreate", mock.Anything, int64(1)).Return(prefs, nil)

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.GET("/preferences", handler.Get)

		req := httptest.NewRequest(http.MethodGet, "/preferences", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		uc := usecases.NewPreferencesUseCase(nil)
		handler := notifHttp.NewPreferencesHandler(uc)

		router := gin.New()
		router.GET("/preferences", handler.Get)

		req := httptest.NewRequest(http.MethodGet, "/preferences", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("usecase error", func(t *testing.T) {
		mockPrefsRepo := new(MockPreferencesRepository)
		uc := usecases.NewPreferencesUseCase(mockPrefsRepo)
		handler := notifHttp.NewPreferencesHandler(uc)

		mockPrefsRepo.On("GetOrCreate", mock.Anything, int64(1)).Return(nil, errors.New("db error"))

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.GET("/preferences", handler.Get)

		req := httptest.NewRequest(http.MethodGet, "/preferences", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestPreferencesHandler_Update(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		mockPrefsRepo := new(MockPreferencesRepository)
		uc := usecases.NewPreferencesUseCase(mockPrefsRepo)
		handler := notifHttp.NewPreferencesHandler(uc)

		prefs := defaultPrefs(1)
		mockPrefsRepo.On("GetOrCreate", mock.Anything, int64(1)).Return(prefs, nil)
		mockPrefsRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.PUT("/preferences", handler.Update)

		emailEnabled := true
		payload := dto.PreferencesInput{EmailEnabled: &emailEnabled}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPut, "/preferences", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		uc := usecases.NewPreferencesUseCase(nil)
		handler := notifHttp.NewPreferencesHandler(uc)

		router := gin.New()
		router.PUT("/preferences", handler.Update)

		body, _ := json.Marshal(dto.PreferencesInput{})
		req := httptest.NewRequest(http.MethodPut, "/preferences", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		uc := usecases.NewPreferencesUseCase(nil)
		handler := notifHttp.NewPreferencesHandler(uc)

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.PUT("/preferences", handler.Update)

		req := httptest.NewRequest(http.MethodPut, "/preferences", bytes.NewReader([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("usecase error", func(t *testing.T) {
		mockPrefsRepo := new(MockPreferencesRepository)
		uc := usecases.NewPreferencesUseCase(mockPrefsRepo)
		handler := notifHttp.NewPreferencesHandler(uc)

		mockPrefsRepo.On("GetOrCreate", mock.Anything, int64(1)).Return(nil, errors.New("db error"))

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.PUT("/preferences", handler.Update)

		payload := dto.PreferencesInput{}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPut, "/preferences", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestPreferencesHandler_ToggleChannel(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		mockPrefsRepo := new(MockPreferencesRepository)
		uc := usecases.NewPreferencesUseCase(mockPrefsRepo)
		handler := notifHttp.NewPreferencesHandler(uc)

		prefs := defaultPrefs(1)
		mockPrefsRepo.On("UpdateChannelEnabled", mock.Anything, int64(1), entities.ChannelEmail, true).Return(nil)
		mockPrefsRepo.On("GetOrCreate", mock.Anything, int64(1)).Return(prefs, nil)

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.PUT("/preferences/channel", handler.ToggleChannel)

		payload := dto.ChannelToggleInput{Channel: "email", Enabled: true}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPut, "/preferences/channel", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		uc := usecases.NewPreferencesUseCase(nil)
		handler := notifHttp.NewPreferencesHandler(uc)

		router := gin.New()
		router.PUT("/preferences/channel", handler.ToggleChannel)

		body, _ := json.Marshal(dto.ChannelToggleInput{})
		req := httptest.NewRequest(http.MethodPut, "/preferences/channel", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		uc := usecases.NewPreferencesUseCase(nil)
		handler := notifHttp.NewPreferencesHandler(uc)

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.PUT("/preferences/channel", handler.ToggleChannel)

		req := httptest.NewRequest(http.MethodPut, "/preferences/channel", bytes.NewReader([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("validation error - invalid channel", func(t *testing.T) {
		uc := usecases.NewPreferencesUseCase(nil)
		handler := notifHttp.NewPreferencesHandler(uc)

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.PUT("/preferences/channel", handler.ToggleChannel)

		payload := dto.ChannelToggleInput{Channel: "invalid_channel", Enabled: true}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPut, "/preferences/channel", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("usecase error", func(t *testing.T) {
		mockPrefsRepo := new(MockPreferencesRepository)
		uc := usecases.NewPreferencesUseCase(mockPrefsRepo)
		handler := notifHttp.NewPreferencesHandler(uc)

		mockPrefsRepo.On("UpdateChannelEnabled", mock.Anything, int64(1), entities.ChannelEmail, true).Return(errors.New("db error"))

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.PUT("/preferences/channel", handler.ToggleChannel)

		payload := dto.ChannelToggleInput{Channel: "email", Enabled: true}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPut, "/preferences/channel", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestPreferencesHandler_UpdateQuietHours(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		mockPrefsRepo := new(MockPreferencesRepository)
		uc := usecases.NewPreferencesUseCase(mockPrefsRepo)
		handler := notifHttp.NewPreferencesHandler(uc)

		prefs := defaultPrefs(1)
		mockPrefsRepo.On("UpdateQuietHours", mock.Anything, int64(1), true, "22:00", "08:00", "Europe/Moscow").Return(nil)
		mockPrefsRepo.On("GetOrCreate", mock.Anything, int64(1)).Return(prefs, nil)

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.PUT("/preferences/quiet-hours", handler.UpdateQuietHours)

		payload := dto.QuietHoursInput{
			Enabled:   true,
			StartTime: "22:00",
			EndTime:   "08:00",
			Timezone:  "Europe/Moscow",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPut, "/preferences/quiet-hours", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		uc := usecases.NewPreferencesUseCase(nil)
		handler := notifHttp.NewPreferencesHandler(uc)

		router := gin.New()
		router.PUT("/preferences/quiet-hours", handler.UpdateQuietHours)

		body, _ := json.Marshal(dto.QuietHoursInput{})
		req := httptest.NewRequest(http.MethodPut, "/preferences/quiet-hours", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid body", func(t *testing.T) {
		uc := usecases.NewPreferencesUseCase(nil)
		handler := notifHttp.NewPreferencesHandler(uc)

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.PUT("/preferences/quiet-hours", handler.UpdateQuietHours)

		req := httptest.NewRequest(http.MethodPut, "/preferences/quiet-hours", bytes.NewReader([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("validation error - wrong time format", func(t *testing.T) {
		uc := usecases.NewPreferencesUseCase(nil)
		handler := notifHttp.NewPreferencesHandler(uc)

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.PUT("/preferences/quiet-hours", handler.UpdateQuietHours)

		// start_time should be len=5
		payload := dto.QuietHoursInput{
			Enabled:   true,
			StartTime: "22:00:00",
			EndTime:   "08:00",
			Timezone:  "Europe/Moscow",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPut, "/preferences/quiet-hours", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("usecase error", func(t *testing.T) {
		mockPrefsRepo := new(MockPreferencesRepository)
		uc := usecases.NewPreferencesUseCase(mockPrefsRepo)
		handler := notifHttp.NewPreferencesHandler(uc)

		mockPrefsRepo.On("UpdateQuietHours", mock.Anything, int64(1), true, "22:00", "08:00", "Europe/Moscow").Return(errors.New("db error"))

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.PUT("/preferences/quiet-hours", handler.UpdateQuietHours)

		payload := dto.QuietHoursInput{
			Enabled:   true,
			StartTime: "22:00",
			EndTime:   "08:00",
			Timezone:  "Europe/Moscow",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPut, "/preferences/quiet-hours", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestPreferencesHandler_Reset(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		mockPrefsRepo := new(MockPreferencesRepository)
		uc := usecases.NewPreferencesUseCase(mockPrefsRepo)
		handler := notifHttp.NewPreferencesHandler(uc)

		prefs := defaultPrefs(1)
		mockPrefsRepo.On("Delete", mock.Anything, int64(1)).Return(nil)
		mockPrefsRepo.On("GetOrCreate", mock.Anything, int64(1)).Return(prefs, nil)

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.POST("/preferences/reset", handler.Reset)

		req := httptest.NewRequest(http.MethodPost, "/preferences/reset", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		uc := usecases.NewPreferencesUseCase(nil)
		handler := notifHttp.NewPreferencesHandler(uc)

		router := gin.New()
		router.POST("/preferences/reset", handler.Reset)

		req := httptest.NewRequest(http.MethodPost, "/preferences/reset", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("usecase error", func(t *testing.T) {
		mockPrefsRepo := new(MockPreferencesRepository)
		uc := usecases.NewPreferencesUseCase(mockPrefsRepo)
		handler := notifHttp.NewPreferencesHandler(uc)

		mockPrefsRepo.On("Delete", mock.Anything, int64(1)).Return(nil)
		mockPrefsRepo.On("GetOrCreate", mock.Anything, int64(1)).Return(nil, errors.New("db error"))

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.POST("/preferences/reset", handler.Reset)

		req := httptest.NewRequest(http.MethodPost, "/preferences/reset", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestPreferencesHandler_GetTimezones(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockPrefsRepo := new(MockPreferencesRepository)
	uc := usecases.NewPreferencesUseCase(mockPrefsRepo)
	handler := notifHttp.NewPreferencesHandler(uc)

	router := gin.New()
	router.GET("/timezones", handler.GetTimezones)

	req := httptest.NewRequest(http.MethodGet, "/timezones", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string][]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp["timezones"])
}

// ===================== TELEGRAM HANDLER TESTS =====================

func TestTelegramHandler_GenerateVerificationCode(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		mockTgRepo := new(MockTelegramRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockTgService := new(MockTelegramDomainService)

		svc := services.NewTelegramVerificationService(mockTgRepo, mockPrefsRepo, mockTgService, nil, "testbot")
		handler := notifHttp.NewTelegramHandler(svc)

		// No existing code
		mockTgRepo.On("GetActiveVerificationCodeByUserID", mock.Anything, int64(1)).Return(nil, nil)
		mockTgRepo.On("CreateVerificationCode", mock.Anything, mock.Anything).Return(nil)

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.POST("/telegram/verification-code", handler.GenerateVerificationCode)

		req := httptest.NewRequest(http.MethodPost, "/telegram/verification-code", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		svc := services.NewTelegramVerificationService(nil, nil, nil, nil, "testbot")
		handler := notifHttp.NewTelegramHandler(svc)

		router := gin.New()
		router.POST("/telegram/verification-code", handler.GenerateVerificationCode)

		req := httptest.NewRequest(http.MethodPost, "/telegram/verification-code", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		mockTgRepo := new(MockTelegramRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockTgService := new(MockTelegramDomainService)

		svc := services.NewTelegramVerificationService(mockTgRepo, mockPrefsRepo, mockTgService, nil, "testbot")
		handler := notifHttp.NewTelegramHandler(svc)

		mockTgRepo.On("GetActiveVerificationCodeByUserID", mock.Anything, int64(1)).Return(nil, errors.New("db error"))

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.POST("/telegram/verification-code", handler.GenerateVerificationCode)

		req := httptest.NewRequest(http.MethodPost, "/telegram/verification-code", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestTelegramHandler_GetConnectionStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("connected", func(t *testing.T) {
		mockTgRepo := new(MockTelegramRepository)
		svc := services.NewTelegramVerificationService(mockTgRepo, nil, nil, nil, "testbot")
		handler := notifHttp.NewTelegramHandler(svc)

		conn := &entities.TelegramConnection{
			UserID:            1,
			TelegramChatID:    12345,
			TelegramUsername:  "testuser",
			TelegramFirstName: "Test",
			IsActive:          true,
			ConnectedAt:       time.Now(),
		}
		mockTgRepo.On("GetConnectionByUserID", mock.Anything, int64(1)).Return(conn, nil)

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.GET("/telegram/status", handler.GetConnectionStatus)

		req := httptest.NewRequest(http.MethodGet, "/telegram/status", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp notifHttp.TelegramConnectionResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.True(t, resp.Connected)
	})

	t.Run("not connected", func(t *testing.T) {
		mockTgRepo := new(MockTelegramRepository)
		svc := services.NewTelegramVerificationService(mockTgRepo, nil, nil, nil, "testbot")
		handler := notifHttp.NewTelegramHandler(svc)

		mockTgRepo.On("GetConnectionByUserID", mock.Anything, int64(1)).Return(nil, nil)

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.GET("/telegram/status", handler.GetConnectionStatus)

		req := httptest.NewRequest(http.MethodGet, "/telegram/status", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp notifHttp.TelegramConnectionResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.False(t, resp.Connected)
	})

	t.Run("unauthorized", func(t *testing.T) {
		svc := services.NewTelegramVerificationService(nil, nil, nil, nil, "testbot")
		handler := notifHttp.NewTelegramHandler(svc)

		router := gin.New()
		router.GET("/telegram/status", handler.GetConnectionStatus)

		req := httptest.NewRequest(http.MethodGet, "/telegram/status", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		mockTgRepo := new(MockTelegramRepository)
		svc := services.NewTelegramVerificationService(mockTgRepo, nil, nil, nil, "testbot")
		handler := notifHttp.NewTelegramHandler(svc)

		mockTgRepo.On("GetConnectionByUserID", mock.Anything, int64(1)).Return(nil, errors.New("db error"))

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.GET("/telegram/status", handler.GetConnectionStatus)

		req := httptest.NewRequest(http.MethodGet, "/telegram/status", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestTelegramHandler_DisconnectTelegram(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		mockTgRepo := new(MockTelegramRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		svc := services.NewTelegramVerificationService(mockTgRepo, mockPrefsRepo, nil, nil, "testbot")
		handler := notifHttp.NewTelegramHandler(svc)

		mockTgRepo.On("DeleteConnection", mock.Anything, int64(1)).Return(nil)
		mockPrefsRepo.On("UpdateChannelEnabled", mock.Anything, int64(1), entities.ChannelTelegram, false).Return(nil)

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.POST("/telegram/disconnect", handler.DisconnectTelegram)

		req := httptest.NewRequest(http.MethodPost, "/telegram/disconnect", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		svc := services.NewTelegramVerificationService(nil, nil, nil, nil, "testbot")
		handler := notifHttp.NewTelegramHandler(svc)

		router := gin.New()
		router.POST("/telegram/disconnect", handler.DisconnectTelegram)

		req := httptest.NewRequest(http.MethodPost, "/telegram/disconnect", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("error from service", func(t *testing.T) {
		mockTgRepo := new(MockTelegramRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		svc := services.NewTelegramVerificationService(mockTgRepo, mockPrefsRepo, nil, nil, "testbot")
		handler := notifHttp.NewTelegramHandler(svc)

		mockTgRepo.On("DeleteConnection", mock.Anything, int64(1)).Return(errors.New("some db error"))

		router := gin.New()
		router.Use(setAuthMiddleware(1))
		router.POST("/telegram/disconnect", handler.DisconnectTelegram)

		req := httptest.NewRequest(http.MethodPost, "/telegram/disconnect", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// The service wraps the error, so the handler sees "failed to delete connection: ..."
		// which doesn't match "connection not found" exactly -> returns 500
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// ===================== TELEGRAM WEBHOOK HANDLER TESTS =====================

func TestTelegramWebhookHandler_HandleWebhook(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("valid webhook with secret", func(t *testing.T) {
		mockTgRepo := new(MockTelegramRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockTgService := new(MockTelegramDomainService)
		svc := services.NewTelegramVerificationService(mockTgRepo, mockPrefsRepo, mockTgService, nil, "testbot")

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		handler := notifHttp.NewTelegramWebhookHandler(svc, mockTgService, "test-secret", logger, nil, nil)

		// The message will trigger ProcessUpdate in a goroutine, mock the service call
		mockTgService.On("SendMessage", mock.Anything, mock.Anything).Return(nil).Maybe()

		router := gin.New()
		router.POST("/webhook", handler.HandleWebhook)

		update := map[string]interface{}{
			"update_id": 1,
			"message": map[string]interface{}{
				"message_id": 1,
				"text":       "/help",
				"chat": map[string]interface{}{
					"id":   12345,
					"type": "private",
				},
			},
		}
		body, _ := json.Marshal(update)
		req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Telegram-Bot-Api-Secret-Token", "test-secret")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid secret", func(t *testing.T) {
		mockTgRepo := new(MockTelegramRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockTgService := new(MockTelegramDomainService)
		svc := services.NewTelegramVerificationService(mockTgRepo, mockPrefsRepo, mockTgService, nil, "testbot")

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		handler := notifHttp.NewTelegramWebhookHandler(svc, mockTgService, "test-secret", logger, nil, nil)

		router := gin.New()
		router.POST("/webhook", handler.HandleWebhook)

		body, _ := json.Marshal(map[string]interface{}{"update_id": 1})
		req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Telegram-Bot-Api-Secret-Token", "wrong-secret")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("no secret configured - allows all", func(t *testing.T) {
		mockTgRepo := new(MockTelegramRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockTgService := new(MockTelegramDomainService)
		svc := services.NewTelegramVerificationService(mockTgRepo, mockPrefsRepo, mockTgService, nil, "testbot")

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		handler := notifHttp.NewTelegramWebhookHandler(svc, mockTgService, "", logger, nil, nil)

		mockTgService.On("SendMessage", mock.Anything, mock.Anything).Return(nil).Maybe()

		router := gin.New()
		router.POST("/webhook", handler.HandleWebhook)

		update := map[string]interface{}{
			"update_id": 1,
		}
		body, _ := json.Marshal(update)
		req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid JSON body", func(t *testing.T) {
		mockTgService := new(MockTelegramDomainService)
		svc := services.NewTelegramVerificationService(nil, nil, mockTgService, nil, "testbot")

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		handler := notifHttp.NewTelegramWebhookHandler(svc, mockTgService, "", logger, nil, nil)

		router := gin.New()
		router.POST("/webhook", handler.HandleWebhook)

		req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader([]byte("not json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestTelegramWebhookHandler_ProcessUpdate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("nil message", func(t *testing.T) {
		mockTgService := new(MockTelegramDomainService)
		svc := services.NewTelegramVerificationService(nil, nil, mockTgService, nil, "testbot")
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		handler := notifHttp.NewTelegramWebhookHandler(svc, mockTgService, "", logger, nil, nil)

		update := &telegram.Update{UpdateID: 1, Message: nil}
		handler.ProcessUpdate(update) // Should not panic
	})

	t.Run("nil chat", func(t *testing.T) {
		mockTgService := new(MockTelegramDomainService)
		svc := services.NewTelegramVerificationService(nil, nil, mockTgService, nil, "testbot")
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		handler := notifHttp.NewTelegramWebhookHandler(svc, mockTgService, "", logger, nil, nil)

		update := &telegram.Update{
			UpdateID: 1,
			Message:  &telegram.Message{MessageID: 1, Chat: nil},
		}
		handler.ProcessUpdate(update) // Should not panic
	})

	t.Run("start command without code", func(t *testing.T) {
		mockTgService := new(MockTelegramDomainService)
		svc := services.NewTelegramVerificationService(nil, nil, mockTgService, nil, "testbot")
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		handler := notifHttp.NewTelegramWebhookHandler(svc, mockTgService, "", logger, nil, nil)

		mockTgService.On("SendMessage", mock.Anything, mock.Anything).Return(nil)

		update := &telegram.Update{
			UpdateID: 1,
			Message: &telegram.Message{
				MessageID: 1,
				Chat:      &telegram.Chat{ID: 12345, Type: "private"},
				Text:      "/start",
			},
		}
		handler.ProcessUpdate(update)

		// Verify send was called (welcome message)
		time.Sleep(50 * time.Millisecond)
		mockTgService.AssertCalled(t, "SendMessage", mock.Anything, mock.Anything)
	})

	t.Run("start command with verification code", func(t *testing.T) {
		mockTgRepo := new(MockTelegramRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockTgService := new(MockTelegramDomainService)
		svc := services.NewTelegramVerificationService(mockTgRepo, mockPrefsRepo, mockTgService, nil, "testbot")
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		handler := notifHttp.NewTelegramWebhookHandler(svc, mockTgService, "", logger, nil, nil)

		// Code not found
		mockTgRepo.On("GetVerificationCodeByCode", mock.Anything, "abc12345").Return(nil, nil)
		mockTgService.On("SendMessage", mock.Anything, mock.Anything).Return(nil)

		update := &telegram.Update{
			UpdateID: 1,
			Message: &telegram.Message{
				MessageID: 1,
				Chat:      &telegram.Chat{ID: 12345, Type: "private", Username: "user", FirstName: "Test"},
				Text:      "/start abc12345",
			},
		}
		handler.ProcessUpdate(update)
		time.Sleep(50 * time.Millisecond)
	})

	t.Run("plain verification code (8 hex chars)", func(t *testing.T) {
		mockTgRepo := new(MockTelegramRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockTgService := new(MockTelegramDomainService)
		svc := services.NewTelegramVerificationService(mockTgRepo, mockPrefsRepo, mockTgService, nil, "testbot")
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		handler := notifHttp.NewTelegramWebhookHandler(svc, mockTgService, "", logger, nil, nil)

		mockTgRepo.On("GetVerificationCodeByCode", mock.Anything, "abcd1234").Return(nil, nil)
		mockTgService.On("SendMessage", mock.Anything, mock.Anything).Return(nil)

		update := &telegram.Update{
			UpdateID: 1,
			Message: &telegram.Message{
				MessageID: 1,
				Chat:      &telegram.Chat{ID: 12345, Type: "private"},
				Text:      "abcd1234",
			},
		}
		handler.ProcessUpdate(update)
		time.Sleep(50 * time.Millisecond)
	})

	t.Run("help command", func(t *testing.T) {
		mockTgService := new(MockTelegramDomainService)
		svc := services.NewTelegramVerificationService(nil, nil, mockTgService, nil, "testbot")
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		handler := notifHttp.NewTelegramWebhookHandler(svc, mockTgService, "", logger, nil, nil)

		mockTgService.On("SendMessage", mock.Anything, mock.Anything).Return(nil)

		update := &telegram.Update{
			UpdateID: 1,
			Message: &telegram.Message{
				MessageID: 1,
				Chat:      &telegram.Chat{ID: 12345, Type: "private"},
				Text:      "/help",
			},
		}
		handler.ProcessUpdate(update)
		time.Sleep(50 * time.Millisecond)
	})

	t.Run("status command", func(t *testing.T) {
		mockTgRepo := new(MockTelegramRepository)
		mockTgService := new(MockTelegramDomainService)
		svc := services.NewTelegramVerificationService(mockTgRepo, nil, mockTgService, nil, "testbot")
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		handler := notifHttp.NewTelegramWebhookHandler(svc, mockTgService, "", logger, nil, nil)

		mockTgRepo.On("GetConnectionByUserID", mock.Anything, int64(0)).Return(nil, nil)
		mockTgService.On("SendMessage", mock.Anything, mock.Anything).Return(nil)

		update := &telegram.Update{
			UpdateID: 1,
			Message: &telegram.Message{
				MessageID: 1,
				Chat:      &telegram.Chat{ID: 12345, Type: "private"},
				Text:      "/status",
			},
		}
		handler.ProcessUpdate(update)
		time.Sleep(50 * time.Millisecond)
	})

	t.Run("status command with active connection", func(t *testing.T) {
		mockTgRepo := new(MockTelegramRepository)
		mockTgService := new(MockTelegramDomainService)
		svc := services.NewTelegramVerificationService(mockTgRepo, nil, mockTgService, nil, "testbot")
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		handler := notifHttp.NewTelegramWebhookHandler(svc, mockTgService, "", logger, nil, nil)

		conn := &entities.TelegramConnection{
			IsActive:    true,
			ConnectedAt: time.Now(),
		}
		mockTgRepo.On("GetConnectionByUserID", mock.Anything, int64(0)).Return(conn, nil)
		mockTgService.On("SendMessage", mock.Anything, mock.Anything).Return(nil)

		update := &telegram.Update{
			UpdateID: 1,
			Message: &telegram.Message{
				MessageID: 1,
				Chat:      &telegram.Chat{ID: 12345, Type: "private"},
				Text:      "/status",
			},
		}
		handler.ProcessUpdate(update)
		time.Sleep(50 * time.Millisecond)
	})

	t.Run("status command with inactive connection", func(t *testing.T) {
		mockTgRepo := new(MockTelegramRepository)
		mockTgService := new(MockTelegramDomainService)
		svc := services.NewTelegramVerificationService(mockTgRepo, nil, mockTgService, nil, "testbot")
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		handler := notifHttp.NewTelegramWebhookHandler(svc, mockTgService, "", logger, nil, nil)

		conn := &entities.TelegramConnection{
			IsActive:    false,
			ConnectedAt: time.Now(),
		}
		mockTgRepo.On("GetConnectionByUserID", mock.Anything, int64(0)).Return(conn, nil)
		mockTgService.On("SendMessage", mock.Anything, mock.Anything).Return(nil)

		update := &telegram.Update{
			UpdateID: 1,
			Message: &telegram.Message{
				MessageID: 1,
				Chat:      &telegram.Chat{ID: 12345, Type: "private"},
				Text:      "/status",
			},
		}
		handler.ProcessUpdate(update)
		time.Sleep(50 * time.Millisecond)
	})

	t.Run("fact command without services", func(t *testing.T) {
		mockTgService := new(MockTelegramDomainService)
		svc := services.NewTelegramVerificationService(nil, nil, mockTgService, nil, "testbot")
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		handler := notifHttp.NewTelegramWebhookHandler(svc, mockTgService, "", logger, nil, nil)

		mockTgService.On("SendMessage", mock.Anything, mock.Anything).Return(nil)

		update := &telegram.Update{
			UpdateID: 1,
			Message: &telegram.Message{
				MessageID: 1,
				Chat:      &telegram.Chat{ID: 12345, Type: "private"},
				Text:      "/fact",
			},
		}
		handler.ProcessUpdate(update)
		time.Sleep(50 * time.Millisecond)
	})

	t.Run("mood command without services", func(t *testing.T) {
		mockTgService := new(MockTelegramDomainService)
		svc := services.NewTelegramVerificationService(nil, nil, mockTgService, nil, "testbot")
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		handler := notifHttp.NewTelegramWebhookHandler(svc, mockTgService, "", logger, nil, nil)

		mockTgService.On("SendMessage", mock.Anything, mock.Anything).Return(nil)

		update := &telegram.Update{
			UpdateID: 1,
			Message: &telegram.Message{
				MessageID: 1,
				Chat:      &telegram.Chat{ID: 12345, Type: "private"},
				Text:      "/mood",
			},
		}
		handler.ProcessUpdate(update)
		time.Sleep(50 * time.Millisecond)
	})

	t.Run("unknown command", func(t *testing.T) {
		mockTgService := new(MockTelegramDomainService)
		svc := services.NewTelegramVerificationService(nil, nil, mockTgService, nil, "testbot")
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		handler := notifHttp.NewTelegramWebhookHandler(svc, mockTgService, "", logger, nil, nil)

		mockTgService.On("SendMessage", mock.Anything, mock.Anything).Return(nil)

		update := &telegram.Update{
			UpdateID: 1,
			Message: &telegram.Message{
				MessageID: 1,
				Chat:      &telegram.Chat{ID: 12345, Type: "private"},
				Text:      "random message",
			},
		}
		handler.ProcessUpdate(update)
		time.Sleep(50 * time.Millisecond)
	})

	t.Run("verification code error from service", func(t *testing.T) {
		mockTgRepo := new(MockTelegramRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockTgService := new(MockTelegramDomainService)
		svc := services.NewTelegramVerificationService(mockTgRepo, mockPrefsRepo, mockTgService, nil, "testbot")
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		handler := notifHttp.NewTelegramWebhookHandler(svc, mockTgService, "", logger, nil, nil)

		mockTgRepo.On("GetVerificationCodeByCode", mock.Anything, "ab12cd34").Return(nil, errors.New("db error"))
		mockTgService.On("SendMessage", mock.Anything, mock.Anything).Return(nil)

		update := &telegram.Update{
			UpdateID: 1,
			Message: &telegram.Message{
				MessageID: 1,
				Chat:      &telegram.Chat{ID: 12345, Type: "private", Username: "user", FirstName: "Test"},
				Text:      "ab12cd34",
			},
		}
		handler.ProcessUpdate(update)
		time.Sleep(50 * time.Millisecond)
	})

	t.Run("successful verification code flow", func(t *testing.T) {
		mockTgRepo := new(MockTelegramRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockTgService := new(MockTelegramDomainService)
		svc := services.NewTelegramVerificationService(mockTgRepo, mockPrefsRepo, mockTgService, nil, "testbot")
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		handler := notifHttp.NewTelegramWebhookHandler(svc, mockTgService, "", logger, nil, nil)

		code := &entities.TelegramVerificationCode{
			ID:        1,
			UserID:    42,
			Code:      "abcd1234",
			ExpiresAt: time.Now().Add(15 * time.Minute),
			CreatedAt: time.Now(),
		}
		mockTgRepo.On("GetVerificationCodeByCode", mock.Anything, "abcd1234").Return(code, nil)
		mockTgRepo.On("GetConnectionByChatID", mock.Anything, int64(12345)).Return(nil, nil)
		mockTgRepo.On("MarkCodeAsUsed", mock.Anything, int64(1)).Return(nil)
		mockTgRepo.On("CreateConnection", mock.Anything, mock.Anything).Return(nil)
		mockPrefsRepo.On("UpdateChannelEnabled", mock.Anything, int64(42), entities.ChannelTelegram, true).Return(nil)
		mockTgService.On("SendNotification", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		mockTgService.On("SendMessage", mock.Anything, mock.Anything).Return(nil)

		update := &telegram.Update{
			UpdateID: 1,
			Message: &telegram.Message{
				MessageID: 1,
				Chat:      &telegram.Chat{ID: 12345, Type: "private", Username: "testuser", FirstName: "Test"},
				Text:      "abcd1234",
			},
		}
		handler.ProcessUpdate(update)
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("verification code for used code", func(t *testing.T) {
		mockTgRepo := new(MockTelegramRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockTgService := new(MockTelegramDomainService)
		svc := services.NewTelegramVerificationService(mockTgRepo, mockPrefsRepo, mockTgService, nil, "testbot")
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		handler := notifHttp.NewTelegramWebhookHandler(svc, mockTgService, "", logger, nil, nil)

		usedAt := time.Now()
		code := &entities.TelegramVerificationCode{
			ID:        1,
			UserID:    42,
			Code:      "abcd1234",
			ExpiresAt: time.Now().Add(15 * time.Minute),
			UsedAt:    &usedAt,
			CreatedAt: time.Now(),
		}
		mockTgRepo.On("GetVerificationCodeByCode", mock.Anything, "abcd1234").Return(code, nil)
		mockTgService.On("SendMessage", mock.Anything, mock.Anything).Return(nil)

		update := &telegram.Update{
			UpdateID: 1,
			Message: &telegram.Message{
				MessageID: 1,
				Chat:      &telegram.Chat{ID: 12345, Type: "private"},
				Text:      "abcd1234",
			},
		}
		handler.ProcessUpdate(update)
		time.Sleep(50 * time.Millisecond)
	})

	t.Run("non-hex 8 char text goes to unknown command", func(t *testing.T) {
		mockTgService := new(MockTelegramDomainService)
		svc := services.NewTelegramVerificationService(nil, nil, mockTgService, nil, "testbot")
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		handler := notifHttp.NewTelegramWebhookHandler(svc, mockTgService, "", logger, nil, nil)

		mockTgService.On("SendMessage", mock.Anything, mock.Anything).Return(nil)

		// "gggggggg" is 8 chars but not hex (g is not a hex digit)
		update := &telegram.Update{
			UpdateID: 1,
			Message: &telegram.Message{
				MessageID: 1,
				Chat:      &telegram.Chat{ID: 12345, Type: "private"},
				Text:      "gggggggg",
			},
		}
		handler.ProcessUpdate(update)
		time.Sleep(50 * time.Millisecond)
	})

	t.Run("send message error does not panic", func(t *testing.T) {
		mockTgService := new(MockTelegramDomainService)
		svc := services.NewTelegramVerificationService(nil, nil, mockTgService, nil, "testbot")
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		handler := notifHttp.NewTelegramWebhookHandler(svc, mockTgService, "", logger, nil, nil)

		mockTgService.On("SendMessage", mock.Anything, mock.Anything).Return(errors.New("send failed"))

		update := &telegram.Update{
			UpdateID: 1,
			Message: &telegram.Message{
				MessageID: 1,
				Chat:      &telegram.Chat{ID: 12345, Type: "private"},
				Text:      "/help",
			},
		}
		handler.ProcessUpdate(update)
		time.Sleep(50 * time.Millisecond)
	})

	t.Run("status command with error", func(t *testing.T) {
		mockTgRepo := new(MockTelegramRepository)
		mockTgService := new(MockTelegramDomainService)
		svc := services.NewTelegramVerificationService(mockTgRepo, nil, mockTgService, nil, "testbot")
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		handler := notifHttp.NewTelegramWebhookHandler(svc, mockTgService, "", logger, nil, nil)

		mockTgRepo.On("GetConnectionByUserID", mock.Anything, int64(0)).Return(nil, errors.New("db error"))
		mockTgService.On("SendMessage", mock.Anything, mock.Anything).Return(nil)

		update := &telegram.Update{
			UpdateID: 1,
			Message: &telegram.Message{
				MessageID: 1,
				Chat:      &telegram.Chat{ID: 12345, Type: "private"},
				Text:      "/status",
			},
		}
		handler.ProcessUpdate(update)
		time.Sleep(50 * time.Millisecond)
	})

	t.Run("welcome message send error on verification", func(t *testing.T) {
		mockTgRepo := new(MockTelegramRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockTgService := new(MockTelegramDomainService)
		svc := services.NewTelegramVerificationService(mockTgRepo, mockPrefsRepo, mockTgService, nil, "testbot")
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		handler := notifHttp.NewTelegramWebhookHandler(svc, mockTgService, "", logger, nil, nil)

		code := &entities.TelegramVerificationCode{
			ID:        1,
			UserID:    42,
			Code:      "abcd1234",
			ExpiresAt: time.Now().Add(15 * time.Minute),
			CreatedAt: time.Now(),
		}
		mockTgRepo.On("GetVerificationCodeByCode", mock.Anything, "abcd1234").Return(code, nil)
		mockTgRepo.On("GetConnectionByChatID", mock.Anything, int64(12345)).Return(nil, nil)
		mockTgRepo.On("MarkCodeAsUsed", mock.Anything, int64(1)).Return(nil)
		mockTgRepo.On("CreateConnection", mock.Anything, mock.Anything).Return(nil)
		mockPrefsRepo.On("UpdateChannelEnabled", mock.Anything, int64(42), entities.ChannelTelegram, true).Return(nil)
		// SendNotification (welcome) fails
		mockTgService.On("SendNotification", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("send failed"))
		mockTgService.On("SendMessage", mock.Anything, mock.Anything).Return(nil)

		update := &telegram.Update{
			UpdateID: 1,
			Message: &telegram.Message{
				MessageID: 1,
				Chat:      &telegram.Chat{ID: 12345, Type: "private", Username: "user", FirstName: ""},
				Text:      "abcd1234",
			},
		}
		handler.ProcessUpdate(update)
		time.Sleep(100 * time.Millisecond)
	})
}

// ===================== WEBPUSH HANDLER ERROR PATH TESTS =====================

func TestWebPushHandler_SubscribeRepoError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockWebPushRepository)
	mockService := new(MockWebPushService)

	mockService.On("IsConfigured").Return(true)
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))

	router := setupRouter(mockRepo, mockService, true, 1)

	body, _ := json.Marshal(map[string]any{
		"endpoint": "https://push.example.com/sub1",
		"p256dh":   "test-p256dh-key",
		"auth":     "test-auth-key",
	})
	req := httptest.NewRequest(http.MethodPost, "/push/subscribe", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestWebPushHandler_UnsubscribeRepoError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockWebPushRepository)
	mockService := new(MockWebPushService)

	mockRepo.On("DeleteByEndpoint", mock.Anything, mock.Anything).Return(errors.New("db error"))

	router := setupRouter(mockRepo, mockService, true, 1)

	body, _ := json.Marshal(map[string]any{
		"endpoint": "https://push.example.com/sub1",
	})
	req := httptest.NewRequest(http.MethodPost, "/push/unsubscribe", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestWebPushHandler_GetStatusRepoError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockWebPushRepository)
	mockService := new(MockWebPushService)

	mockRepo.On("GetByUserID", mock.Anything, int64(1)).Return([]*entities.WebPushSubscription{}, errors.New("db error"))

	router := setupRouter(mockRepo, mockService, true, 1)

	req := httptest.NewRequest(http.MethodGet, "/push/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestWebPushHandler_DeleteSubscriptionRepoError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockWebPushRepository)
	mockService := new(MockWebPushService)

	// GetByID error
	mockRepo.On("GetByID", mock.Anything, int64(1)).Return(nil, errors.New("db error"))

	router := setupRouter(mockRepo, mockService, true, 1)

	req := httptest.NewRequest(http.MethodDelete, "/push/subscriptions/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestWebPushHandler_DeleteSubscriptionDeleteError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockWebPushRepository)
	mockService := new(MockWebPushService)

	sub := &entities.WebPushSubscription{ID: 1, UserID: 1}
	mockRepo.On("GetByID", mock.Anything, int64(1)).Return(sub, nil)
	mockRepo.On("Delete", mock.Anything, int64(1)).Return(errors.New("delete error"))

	router := setupRouter(mockRepo, mockService, true, 1)

	req := httptest.NewRequest(http.MethodDelete, "/push/subscriptions/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestWebPushHandler_TestPushSendError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockWebPushRepository)
	mockService := new(MockWebPushService)

	mockService.On("IsConfigured").Return(true)
	mockService.On("SendToUser", mock.Anything, int64(1), mock.Anything).Return(errors.New("send error"))

	router := setupRouter(mockRepo, mockService, true, 1)

	body, _ := json.Marshal(map[string]any{
		"title":   "Test",
		"message": "Hello",
	})
	req := httptest.NewRequest(http.MethodPost, "/push/test", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestWebPushHandler_SubscribeInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockWebPushRepository)
	mockService := new(MockWebPushService)

	mockService.On("IsConfigured").Return(true)

	router := setupRouter(mockRepo, mockService, true, 1)

	req := httptest.NewRequest(http.MethodPost, "/push/subscribe", bytes.NewBuffer([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWebPushHandler_UnsubscribeInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockWebPushRepository)
	mockService := new(MockWebPushService)

	router := setupRouter(mockRepo, mockService, true, 1)

	req := httptest.NewRequest(http.MethodPost, "/push/unsubscribe", bytes.NewBuffer([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWebPushHandler_TestPushInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockWebPushRepository)
	mockService := new(MockWebPushService)

	mockService.On("IsConfigured").Return(true)

	router := setupRouter(mockRepo, mockService, true, 1)

	req := httptest.NewRequest(http.MethodPost, "/push/test", bytes.NewBuffer([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ===================== PREFERENCES HANDLER VALIDATION TESTS =====================

func TestPreferencesHandler_UpdateValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uc := usecases.NewPreferencesUseCase(nil)
	handler := notifHttp.NewPreferencesHandler(uc)

	router := gin.New()
	router.Use(setAuthMiddleware(1))
	router.PUT("/preferences", handler.Update)

	// digest_frequency must be "daily" or "weekly"
	payload := map[string]interface{}{
		"digest_frequency": "invalid_freq_value_that_is_not_daily_or_weekly",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPut, "/preferences", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
