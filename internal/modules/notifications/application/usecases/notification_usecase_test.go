package usecases

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/services"
)

// MockNotificationRepository is a mock implementation of NotificationRepository
type MockNotificationRepository struct {
	mock.Mock
}

func (m *MockNotificationRepository) Create(ctx context.Context, notification *entities.Notification) error {
	args := m.Called(ctx, notification)
	if args.Get(0) == nil {
		notification.ID = 1
	}
	return args.Error(0)
}

func (m *MockNotificationRepository) Update(ctx context.Context, notification *entities.Notification) error {
	args := m.Called(ctx, notification)
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

// MockPreferencesRepository is a mock implementation of PreferencesRepository
type MockPreferencesRepository struct {
	mock.Mock
}

func (m *MockPreferencesRepository) GetByUserID(ctx context.Context, userID int64) (*entities.UserNotificationPreferences, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.UserNotificationPreferences), args.Error(1)
}

func (m *MockPreferencesRepository) Create(ctx context.Context, prefs *entities.UserNotificationPreferences) error {
	args := m.Called(ctx, prefs)
	return args.Error(0)
}

func (m *MockPreferencesRepository) Update(ctx context.Context, prefs *entities.UserNotificationPreferences) error {
	args := m.Called(ctx, prefs)
	return args.Error(0)
}

func (m *MockPreferencesRepository) Delete(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
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

// MockEmailService is a mock implementation of EmailService
type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) SendEmail(ctx context.Context, req *services.SendEmailRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockEmailService) SendWelcomeEmail(ctx context.Context, recipientEmail, userName string) error {
	args := m.Called(ctx, recipientEmail, userName)
	return args.Error(0)
}

func (m *MockEmailService) SendPasswordResetEmail(ctx context.Context, recipientEmail, resetToken string) error {
	args := m.Called(ctx, recipientEmail, resetToken)
	return args.Error(0)
}

func (m *MockEmailService) SendNotification(ctx context.Context, recipientEmail, subject, body string) error {
	args := m.Called(ctx, recipientEmail, subject, body)
	return args.Error(0)
}

func TestNotificationUseCase_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully creates notification", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockEmailSvc := new(MockEmailService)

		uc := NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, nil, mockEmailSvc, nil, nil)

		input := &dto.CreateNotificationInput{
			UserID:   1,
			Type:     entities.NotificationTypeSystem,
			Priority: entities.PriorityNormal,
			Title:    "Test Notification",
			Message:  "Test message",
		}

		mockNotifRepo.On("Create", ctx, mock.AnythingOfType("*entities.Notification")).Return(nil)

		output, err := uc.Create(ctx, input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, input.Title, output.Title)
		assert.Equal(t, input.Message, output.Message)
		mockNotifRepo.AssertExpectations(t)
	})
}

func TestNotificationUseCase_List(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully lists notifications", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockEmailSvc := new(MockEmailService)

		uc := NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, nil, mockEmailSvc, nil, nil)

		now := time.Now()
		notifications := []*entities.Notification{
			{
				ID:        1,
				UserID:    1,
				Type:      entities.NotificationTypeSystem,
				Priority:  entities.PriorityNormal,
				Title:     "Notification 1",
				Message:   "Message 1",
				CreatedAt: now,
				UpdatedAt: now,
			},
			{
				ID:        2,
				UserID:    1,
				Type:      entities.NotificationTypeTask,
				Priority:  entities.PriorityHigh,
				Title:     "Notification 2",
				Message:   "Message 2",
				CreatedAt: now,
				UpdatedAt: now,
			},
		}

		stats := &entities.NotificationStats{
			TotalCount:   2,
			UnreadCount:  1,
			TodayCount:   2,
			UrgentCount:  0,
			ExpiredCount: 0,
		}

		input := &dto.NotificationListInput{
			UserID: 1,
			Limit:  50,
		}

		mockNotifRepo.On("List", ctx, mock.AnythingOfType("*entities.NotificationFilter")).Return(notifications, nil)
		mockNotifRepo.On("GetUnreadCount", ctx, int64(1)).Return(int64(1), nil)
		mockNotifRepo.On("GetStats", ctx, int64(1)).Return(stats, nil)

		output, err := uc.List(ctx, input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Len(t, output.Notifications, 2)
		assert.Equal(t, int64(2), output.TotalCount)
		assert.Equal(t, int64(1), output.UnreadCount)
		mockNotifRepo.AssertExpectations(t)
	})
}

func TestNotificationUseCase_GetByID(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully gets notification by ID", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockEmailSvc := new(MockEmailService)

		uc := NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, nil, mockEmailSvc, nil, nil)

		now := time.Now()
		notification := &entities.Notification{
			ID:        1,
			UserID:    1,
			Type:      entities.NotificationTypeSystem,
			Priority:  entities.PriorityNormal,
			Title:     "Test",
			Message:   "Test message",
			CreatedAt: now,
			UpdatedAt: now,
		}

		mockNotifRepo.On("GetByID", ctx, int64(1)).Return(notification, nil)

		output, err := uc.GetByID(ctx, 1)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, int64(1), output.ID)
		mockNotifRepo.AssertExpectations(t)
	})

	t.Run("returns nil for non-existent notification", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockEmailSvc := new(MockEmailService)

		uc := NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, nil, mockEmailSvc, nil, nil)

		mockNotifRepo.On("GetByID", ctx, int64(999)).Return(nil, nil)

		output, err := uc.GetByID(ctx, 999)

		assert.NoError(t, err)
		assert.Nil(t, output)
		mockNotifRepo.AssertExpectations(t)
	})
}

func TestNotificationUseCase_MarkAsRead(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully marks notification as read", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockEmailSvc := new(MockEmailService)

		uc := NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, nil, mockEmailSvc, nil, nil)

		mockNotifRepo.On("MarkAsRead", ctx, int64(1)).Return(nil)

		err := uc.MarkAsRead(ctx, 1)

		assert.NoError(t, err)
		mockNotifRepo.AssertExpectations(t)
	})
}

func TestNotificationUseCase_MarkAllAsRead(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully marks all notifications as read", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockEmailSvc := new(MockEmailService)

		uc := NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, nil, mockEmailSvc, nil, nil)

		mockNotifRepo.On("MarkAllAsRead", ctx, int64(1)).Return(nil)

		err := uc.MarkAllAsRead(ctx, 1)

		assert.NoError(t, err)
		mockNotifRepo.AssertExpectations(t)
	})
}

func TestNotificationUseCase_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully deletes notification", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockEmailSvc := new(MockEmailService)

		uc := NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, nil, mockEmailSvc, nil, nil)

		mockNotifRepo.On("Delete", ctx, int64(1)).Return(nil)

		err := uc.Delete(ctx, 1)

		assert.NoError(t, err)
		mockNotifRepo.AssertExpectations(t)
	})
}

func TestNotificationUseCase_GetUnreadCount(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully gets unread count", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockEmailSvc := new(MockEmailService)

		uc := NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, nil, mockEmailSvc, nil, nil)

		mockNotifRepo.On("GetUnreadCount", ctx, int64(1)).Return(int64(5), nil)

		output, err := uc.GetUnreadCount(ctx, 1)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, int64(5), output.Count)
		mockNotifRepo.AssertExpectations(t)
	})
}

func TestNotificationUseCase_GetStats(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully gets stats", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockEmailSvc := new(MockEmailService)

		uc := NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, nil, mockEmailSvc, nil, nil)

		stats := &entities.NotificationStats{
			TotalCount:   10,
			UnreadCount:  3,
			TodayCount:   5,
			UrgentCount:  1,
			ExpiredCount: 2,
		}

		mockNotifRepo.On("GetStats", ctx, int64(1)).Return(stats, nil)

		output, err := uc.GetStats(ctx, 1)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, int64(10), output.TotalCount)
		assert.Equal(t, int64(3), output.UnreadCount)
		assert.Equal(t, int64(5), output.TodayCount)
		assert.Equal(t, int64(1), output.UrgentCount)
		assert.Equal(t, int64(2), output.ExpiredCount)
		mockNotifRepo.AssertExpectations(t)
	})
}

func TestNotificationUseCase_SendEventReminderNotification(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully sends event reminder notification", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockEmailSvc := new(MockEmailService)

		uc := NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, nil, mockEmailSvc, nil, nil)

		eventTime := time.Now().Add(time.Hour)

		mockNotifRepo.On("Create", ctx, mock.AnythingOfType("*entities.Notification")).Return(nil)

		err := uc.SendEventReminderNotification(ctx, 1, "Test Event", eventTime, "/events/1")

		assert.NoError(t, err)
		mockNotifRepo.AssertExpectations(t)
	})
}

func TestNotificationUseCase_CreateBulk(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully creates bulk notifications", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockEmailSvc := new(MockEmailService)

		uc := NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, nil, mockEmailSvc, nil, nil)

		input := &dto.CreateBulkNotificationInput{
			UserIDs:  []int64{1, 2, 3},
			Type:     entities.NotificationTypeSystem,
			Priority: entities.PriorityNormal,
			Title:    "System Announcement",
			Message:  "Important system announcement",
		}

		mockNotifRepo.On("CreateBulk", ctx, mock.AnythingOfType("[]*entities.Notification")).Return(nil)

		output, err := uc.CreateBulk(ctx, input)

		assert.NoError(t, err)
		assert.Len(t, output, 3)
		mockNotifRepo.AssertExpectations(t)
	})

	t.Run("uses default priority when not specified", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockEmailSvc := new(MockEmailService)

		uc := NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, nil, mockEmailSvc, nil, nil)

		input := &dto.CreateBulkNotificationInput{
			UserIDs: []int64{1},
			Type:    entities.NotificationTypeSystem,
			Title:   "Test",
			Message: "Test message",
		}

		mockNotifRepo.On("CreateBulk", ctx, mock.AnythingOfType("[]*entities.Notification")).Return(nil)

		output, err := uc.CreateBulk(ctx, input)

		assert.NoError(t, err)
		assert.Len(t, output, 1)
		mockNotifRepo.AssertExpectations(t)
	})

	t.Run("returns error when repository fails", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockEmailSvc := new(MockEmailService)

		uc := NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, nil, mockEmailSvc, nil, nil)

		input := &dto.CreateBulkNotificationInput{
			UserIDs: []int64{1, 2},
			Type:    entities.NotificationTypeSystem,
			Title:   "Test",
			Message: "Test message",
		}

		mockNotifRepo.On("CreateBulk", ctx, mock.AnythingOfType("[]*entities.Notification")).Return(assert.AnError)

		output, err := uc.CreateBulk(ctx, input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create bulk notifications")
		assert.Nil(t, output)
		mockNotifRepo.AssertExpectations(t)
	})
}

func TestNotificationUseCase_DeleteAll(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully deletes all notifications for user", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockEmailSvc := new(MockEmailService)

		uc := NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, nil, mockEmailSvc, nil, nil)

		mockNotifRepo.On("DeleteByUserID", ctx, int64(1)).Return(nil)

		err := uc.DeleteAll(ctx, 1)

		assert.NoError(t, err)
		mockNotifRepo.AssertExpectations(t)
	})

	t.Run("returns error when repository fails", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockEmailSvc := new(MockEmailService)

		uc := NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, nil, mockEmailSvc, nil, nil)

		mockNotifRepo.On("DeleteByUserID", ctx, int64(1)).Return(assert.AnError)

		err := uc.DeleteAll(ctx, 1)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete all notifications")
		mockNotifRepo.AssertExpectations(t)
	})
}

func TestNotificationUseCase_CleanupExpired(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully cleans up expired notifications", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockEmailSvc := new(MockEmailService)

		uc := NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, nil, mockEmailSvc, nil, nil)

		mockNotifRepo.On("DeleteExpired", ctx).Return(int64(5), nil)

		count, err := uc.CleanupExpired(ctx)

		assert.NoError(t, err)
		assert.Equal(t, int64(5), count)
		mockNotifRepo.AssertExpectations(t)
	})

	t.Run("returns zero when no expired notifications", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockEmailSvc := new(MockEmailService)

		uc := NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, nil, mockEmailSvc, nil, nil)

		mockNotifRepo.On("DeleteExpired", ctx).Return(int64(0), nil)

		count, err := uc.CleanupExpired(ctx)

		assert.NoError(t, err)
		assert.Equal(t, int64(0), count)
		mockNotifRepo.AssertExpectations(t)
	})

	t.Run("returns error when repository fails", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockEmailSvc := new(MockEmailService)

		uc := NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, nil, mockEmailSvc, nil, nil)

		mockNotifRepo.On("DeleteExpired", ctx).Return(int64(0), assert.AnError)

		count, err := uc.CleanupExpired(ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to cleanup expired")
		assert.Equal(t, int64(0), count)
		mockNotifRepo.AssertExpectations(t)
	})
}

func TestNotificationUseCase_SendTaskNotification(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully sends task notification", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockEmailSvc := new(MockEmailService)

		uc := NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, nil, mockEmailSvc, nil, nil)

		mockNotifRepo.On("Create", ctx, mock.AnythingOfType("*entities.Notification")).Return(nil)

		err := uc.SendTaskNotification(ctx, 1, "Task Assigned", "You have been assigned a new task", "/tasks/1")

		assert.NoError(t, err)
		mockNotifRepo.AssertExpectations(t)
	})

	t.Run("returns error when create fails", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockEmailSvc := new(MockEmailService)

		uc := NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, nil, mockEmailSvc, nil, nil)

		mockNotifRepo.On("Create", ctx, mock.AnythingOfType("*entities.Notification")).Return(assert.AnError)

		err := uc.SendTaskNotification(ctx, 1, "Task Assigned", "You have been assigned a new task", "/tasks/1")

		assert.Error(t, err)
		mockNotifRepo.AssertExpectations(t)
	})
}

func TestNotificationUseCase_SendDocumentNotification(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully sends document notification", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockEmailSvc := new(MockEmailService)

		uc := NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, nil, mockEmailSvc, nil, nil)

		mockNotifRepo.On("Create", ctx, mock.AnythingOfType("*entities.Notification")).Return(nil)

		err := uc.SendDocumentNotification(ctx, 1, "Document Shared", "A document has been shared with you", "/documents/1")

		assert.NoError(t, err)
		mockNotifRepo.AssertExpectations(t)
	})

	t.Run("returns error when create fails", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockEmailSvc := new(MockEmailService)

		uc := NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, nil, mockEmailSvc, nil, nil)

		mockNotifRepo.On("Create", ctx, mock.AnythingOfType("*entities.Notification")).Return(assert.AnError)

		err := uc.SendDocumentNotification(ctx, 1, "Document Shared", "A document has been shared with you", "/documents/1")

		assert.Error(t, err)
		mockNotifRepo.AssertExpectations(t)
	})
}

func TestNotificationUseCase_SendSystemNotification(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully sends system notification", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockEmailSvc := new(MockEmailService)

		uc := NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, nil, mockEmailSvc, nil, nil)

		mockNotifRepo.On("Create", ctx, mock.AnythingOfType("*entities.Notification")).Return(nil)

		err := uc.SendSystemNotification(ctx, 1, "System Update", "System will be under maintenance")

		assert.NoError(t, err)
		mockNotifRepo.AssertExpectations(t)
	})

	t.Run("returns error when create fails", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockEmailSvc := new(MockEmailService)

		uc := NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, nil, mockEmailSvc, nil, nil)

		mockNotifRepo.On("Create", ctx, mock.AnythingOfType("*entities.Notification")).Return(assert.AnError)

		err := uc.SendSystemNotification(ctx, 1, "System Update", "System will be under maintenance")

		assert.Error(t, err)
		mockNotifRepo.AssertExpectations(t)
	})
}

func TestNotificationUseCase_BroadcastSystemNotification(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully broadcasts system notification", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockEmailSvc := new(MockEmailService)

		uc := NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, nil, mockEmailSvc, nil, nil)

		mockNotifRepo.On("CreateBulk", ctx, mock.AnythingOfType("[]*entities.Notification")).Return(nil)

		err := uc.BroadcastSystemNotification(ctx, []int64{1, 2, 3}, "System Announcement", "Important update for all users")

		assert.NoError(t, err)
		mockNotifRepo.AssertExpectations(t)
	})

	t.Run("returns error when bulk create fails", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockEmailSvc := new(MockEmailService)

		uc := NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, nil, mockEmailSvc, nil, nil)

		mockNotifRepo.On("CreateBulk", ctx, mock.AnythingOfType("[]*entities.Notification")).Return(assert.AnError)

		err := uc.BroadcastSystemNotification(ctx, []int64{1, 2, 3}, "System Announcement", "Important update for all users")

		assert.Error(t, err)
		mockNotifRepo.AssertExpectations(t)
	})
}

func TestNotificationUseCase_Create_Error(t *testing.T) {
	ctx := context.Background()

	t.Run("returns error when repository fails", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockEmailSvc := new(MockEmailService)

		uc := NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, nil, mockEmailSvc, nil, nil)

		input := &dto.CreateNotificationInput{
			UserID:   1,
			Type:     entities.NotificationTypeSystem,
			Priority: entities.PriorityNormal,
			Title:    "Test Notification",
			Message:  "Test message",
		}

		mockNotifRepo.On("Create", ctx, mock.AnythingOfType("*entities.Notification")).Return(assert.AnError)

		output, err := uc.Create(ctx, input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create notification")
		assert.Nil(t, output)
		mockNotifRepo.AssertExpectations(t)
	})
}

func TestNotificationUseCase_List_Errors(t *testing.T) {
	ctx := context.Background()

	t.Run("returns error when list fails", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockEmailSvc := new(MockEmailService)

		uc := NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, nil, mockEmailSvc, nil, nil)

		input := &dto.NotificationListInput{
			UserID: 1,
			Limit:  50,
		}

		mockNotifRepo.On("List", ctx, mock.AnythingOfType("*entities.NotificationFilter")).Return([]*entities.Notification(nil), assert.AnError)

		output, err := uc.List(ctx, input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list notifications")
		assert.Nil(t, output)
		mockNotifRepo.AssertExpectations(t)
	})

	t.Run("returns error when get unread count fails", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockEmailSvc := new(MockEmailService)

		uc := NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, nil, mockEmailSvc, nil, nil)

		input := &dto.NotificationListInput{
			UserID: 1,
			Limit:  50,
		}

		mockNotifRepo.On("List", ctx, mock.AnythingOfType("*entities.NotificationFilter")).Return([]*entities.Notification{}, nil)
		mockNotifRepo.On("GetUnreadCount", ctx, int64(1)).Return(int64(0), assert.AnError)

		output, err := uc.List(ctx, input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get unread count")
		assert.Nil(t, output)
		mockNotifRepo.AssertExpectations(t)
	})

	t.Run("returns error when get stats fails", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockEmailSvc := new(MockEmailService)

		uc := NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, nil, mockEmailSvc, nil, nil)

		input := &dto.NotificationListInput{
			UserID: 1,
			Limit:  50,
		}

		mockNotifRepo.On("List", ctx, mock.AnythingOfType("*entities.NotificationFilter")).Return([]*entities.Notification{}, nil)
		mockNotifRepo.On("GetUnreadCount", ctx, int64(1)).Return(int64(0), nil)
		mockNotifRepo.On("GetStats", ctx, int64(1)).Return(nil, assert.AnError)

		output, err := uc.List(ctx, input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get stats")
		assert.Nil(t, output)
		mockNotifRepo.AssertExpectations(t)
	})

	t.Run("applies default limit when limit is zero", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockEmailSvc := new(MockEmailService)

		uc := NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, nil, mockEmailSvc, nil, nil)

		input := &dto.NotificationListInput{
			UserID: 1,
			Limit:  0,
		}

		stats := &entities.NotificationStats{TotalCount: 0}

		mockNotifRepo.On("List", ctx, mock.MatchedBy(func(f *entities.NotificationFilter) bool {
			return f.Limit == 50
		})).Return([]*entities.Notification{}, nil)
		mockNotifRepo.On("GetUnreadCount", ctx, int64(1)).Return(int64(0), nil)
		mockNotifRepo.On("GetStats", ctx, int64(1)).Return(stats, nil)

		output, err := uc.List(ctx, input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, 50, output.Limit)
		mockNotifRepo.AssertExpectations(t)
	})
}

func TestNotificationUseCase_GetByID_Error(t *testing.T) {
	ctx := context.Background()

	t.Run("returns error when repository fails", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockEmailSvc := new(MockEmailService)

		uc := NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, nil, mockEmailSvc, nil, nil)

		mockNotifRepo.On("GetByID", ctx, int64(1)).Return(nil, assert.AnError)

		output, err := uc.GetByID(ctx, 1)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get notification")
		assert.Nil(t, output)
		mockNotifRepo.AssertExpectations(t)
	})
}

func TestNotificationUseCase_MarkAsRead_Error(t *testing.T) {
	ctx := context.Background()

	t.Run("returns error when repository fails", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockEmailSvc := new(MockEmailService)

		uc := NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, nil, mockEmailSvc, nil, nil)

		mockNotifRepo.On("MarkAsRead", ctx, int64(1)).Return(assert.AnError)

		err := uc.MarkAsRead(ctx, 1)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to mark as read")
		mockNotifRepo.AssertExpectations(t)
	})
}

func TestNotificationUseCase_MarkAllAsRead_Error(t *testing.T) {
	ctx := context.Background()

	t.Run("returns error when repository fails", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockEmailSvc := new(MockEmailService)

		uc := NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, nil, mockEmailSvc, nil, nil)

		mockNotifRepo.On("MarkAllAsRead", ctx, int64(1)).Return(assert.AnError)

		err := uc.MarkAllAsRead(ctx, 1)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to mark all as read")
		mockNotifRepo.AssertExpectations(t)
	})
}

func TestNotificationUseCase_Delete_Error(t *testing.T) {
	ctx := context.Background()

	t.Run("returns error when repository fails", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockEmailSvc := new(MockEmailService)

		uc := NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, nil, mockEmailSvc, nil, nil)

		mockNotifRepo.On("Delete", ctx, int64(1)).Return(assert.AnError)

		err := uc.Delete(ctx, 1)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete notification")
		mockNotifRepo.AssertExpectations(t)
	})
}

func TestNotificationUseCase_GetUnreadCount_Error(t *testing.T) {
	ctx := context.Background()

	t.Run("returns error when repository fails", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockEmailSvc := new(MockEmailService)

		uc := NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, nil, mockEmailSvc, nil, nil)

		mockNotifRepo.On("GetUnreadCount", ctx, int64(1)).Return(int64(0), assert.AnError)

		output, err := uc.GetUnreadCount(ctx, 1)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get unread count")
		assert.Nil(t, output)
		mockNotifRepo.AssertExpectations(t)
	})
}

func TestNotificationUseCase_GetStats_Error(t *testing.T) {
	ctx := context.Background()

	t.Run("returns error when repository fails", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockEmailSvc := new(MockEmailService)

		uc := NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, nil, mockEmailSvc, nil, nil)

		mockNotifRepo.On("GetStats", ctx, int64(1)).Return(nil, assert.AnError)

		output, err := uc.GetStats(ctx, 1)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get stats")
		assert.Nil(t, output)
		mockNotifRepo.AssertExpectations(t)
	})
}

func TestNotificationUseCase_SendEventReminderNotification_Error(t *testing.T) {
	ctx := context.Background()

	t.Run("returns error when create fails", func(t *testing.T) {
		mockNotifRepo := new(MockNotificationRepository)
		mockPrefsRepo := new(MockPreferencesRepository)
		mockEmailSvc := new(MockEmailService)

		uc := NewNotificationUseCase(mockNotifRepo, mockPrefsRepo, nil, mockEmailSvc, nil, nil)

		eventTime := time.Now().Add(time.Hour)

		mockNotifRepo.On("Create", ctx, mock.AnythingOfType("*entities.Notification")).Return(assert.AnError)

		err := uc.SendEventReminderNotification(ctx, 1, "Test Event", eventTime, "/events/1")

		assert.Error(t, err)
		mockNotifRepo.AssertExpectations(t)
	})
}
