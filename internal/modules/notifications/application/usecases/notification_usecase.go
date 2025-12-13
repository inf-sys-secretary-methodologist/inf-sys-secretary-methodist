// Package usecases contains application use cases for the notifications module.
package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/services"
)

// NotificationUseCase handles notification operations
type NotificationUseCase struct {
	notificationRepo repositories.NotificationRepository
	preferencesRepo  repositories.PreferencesRepository
	emailService     services.EmailService
}

// NewNotificationUseCase creates a new notification use case
func NewNotificationUseCase(
	notificationRepo repositories.NotificationRepository,
	preferencesRepo repositories.PreferencesRepository,
	emailService services.EmailService,
) *NotificationUseCase {
	return &NotificationUseCase{
		notificationRepo: notificationRepo,
		preferencesRepo:  preferencesRepo,
		emailService:     emailService,
	}
}

// Create creates a new notification and optionally sends it via other channels
func (uc *NotificationUseCase) Create(ctx context.Context, input *dto.CreateNotificationInput) (*dto.NotificationOutput, error) {
	notification := input.ToEntity()

	if err := uc.notificationRepo.Create(ctx, notification); err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	return dto.ToOutput(notification), nil
}

// CreateBulk creates notifications for multiple users
func (uc *NotificationUseCase) CreateBulk(ctx context.Context, input *dto.CreateBulkNotificationInput) ([]*dto.NotificationOutput, error) {
	now := time.Now()
	notifications := make([]*entities.Notification, len(input.UserIDs))

	priority := input.Priority
	if priority == "" {
		priority = entities.PriorityNormal
	}

	for i, userID := range input.UserIDs {
		notifications[i] = &entities.Notification{
			UserID:    userID,
			Type:      input.Type,
			Priority:  priority,
			Title:     input.Title,
			Message:   input.Message,
			Link:      input.Link,
			ImageURL:  input.ImageURL,
			IsRead:    false,
			ExpiresAt: input.ExpiresAt,
			Metadata:  input.Metadata,
			CreatedAt: now,
			UpdatedAt: now,
		}
	}

	if err := uc.notificationRepo.CreateBulk(ctx, notifications); err != nil {
		return nil, fmt.Errorf("failed to create bulk notifications: %w", err)
	}

	return dto.ToOutputList(notifications), nil
}

// List retrieves notifications based on filter criteria
func (uc *NotificationUseCase) List(ctx context.Context, input *dto.NotificationListInput) (*dto.NotificationListOutput, error) {
	filter := &entities.NotificationFilter{
		UserID:   input.UserID,
		Type:     input.Type,
		Priority: input.Priority,
		IsRead:   input.IsRead,
		Limit:    input.Limit,
		Offset:   input.Offset,
	}

	if filter.Limit <= 0 {
		filter.Limit = 50
	}

	notifications, err := uc.notificationRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list notifications: %w", err)
	}

	unreadCount, err := uc.notificationRepo.GetUnreadCount(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unread count: %w", err)
	}

	stats, err := uc.notificationRepo.GetStats(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	return &dto.NotificationListOutput{
		Notifications: dto.ToOutputList(notifications),
		TotalCount:    stats.TotalCount,
		UnreadCount:   unreadCount,
		Limit:         filter.Limit,
		Offset:        filter.Offset,
	}, nil
}

// GetByID retrieves a notification by ID
func (uc *NotificationUseCase) GetByID(ctx context.Context, id int64) (*dto.NotificationOutput, error) {
	notification, err := uc.notificationRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	if notification == nil {
		return nil, nil
	}

	return dto.ToOutput(notification), nil
}

// MarkAsRead marks a notification as read
func (uc *NotificationUseCase) MarkAsRead(ctx context.Context, id int64) error {
	if err := uc.notificationRepo.MarkAsRead(ctx, id); err != nil {
		return fmt.Errorf("failed to mark as read: %w", err)
	}

	return nil
}

// MarkAllAsRead marks all notifications as read for a user
func (uc *NotificationUseCase) MarkAllAsRead(ctx context.Context, userID int64) error {
	if err := uc.notificationRepo.MarkAllAsRead(ctx, userID); err != nil {
		return fmt.Errorf("failed to mark all as read: %w", err)
	}

	return nil
}

// Delete deletes a notification
func (uc *NotificationUseCase) Delete(ctx context.Context, id int64) error {
	if err := uc.notificationRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	return nil
}

// DeleteAll deletes all notifications for a user
func (uc *NotificationUseCase) DeleteAll(ctx context.Context, userID int64) error {
	if err := uc.notificationRepo.DeleteByUserID(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete all notifications: %w", err)
	}

	return nil
}

// GetUnreadCount returns the count of unread notifications for a user
func (uc *NotificationUseCase) GetUnreadCount(ctx context.Context, userID int64) (*dto.UnreadCountOutput, error) {
	count, err := uc.notificationRepo.GetUnreadCount(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unread count: %w", err)
	}

	return &dto.UnreadCountOutput{Count: count}, nil
}

// GetStats returns notification statistics for a user
func (uc *NotificationUseCase) GetStats(ctx context.Context, userID int64) (*dto.NotificationStatsOutput, error) {
	stats, err := uc.notificationRepo.GetStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	return &dto.NotificationStatsOutput{
		TotalCount:   stats.TotalCount,
		UnreadCount:  stats.UnreadCount,
		TodayCount:   stats.TodayCount,
		UrgentCount:  stats.UrgentCount,
		ExpiredCount: stats.ExpiredCount,
	}, nil
}

// CleanupExpired deletes expired notifications
func (uc *NotificationUseCase) CleanupExpired(ctx context.Context) (int64, error) {
	count, err := uc.notificationRepo.DeleteExpired(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired: %w", err)
	}

	return count, nil
}

// SendEventReminderNotification creates and sends an event reminder notification
func (uc *NotificationUseCase) SendEventReminderNotification(ctx context.Context, userID int64, eventTitle string, eventTime time.Time, link string) error {
	input := &dto.CreateNotificationInput{
		UserID:   userID,
		Type:     entities.NotificationTypeReminder,
		Priority: entities.PriorityHigh,
		Title:    "Напоминание о событии",
		Message:  fmt.Sprintf("Скоро начнётся: %s в %s", eventTitle, eventTime.Format("15:04")),
		Link:     link,
	}

	_, err := uc.Create(ctx, input)
	return err
}

// SendTaskNotification creates a task-related notification
func (uc *NotificationUseCase) SendTaskNotification(ctx context.Context, userID int64, title, message, link string) error {
	input := &dto.CreateNotificationInput{
		UserID:   userID,
		Type:     entities.NotificationTypeTask,
		Priority: entities.PriorityNormal,
		Title:    title,
		Message:  message,
		Link:     link,
	}

	_, err := uc.Create(ctx, input)
	return err
}

// SendDocumentNotification creates a document-related notification
func (uc *NotificationUseCase) SendDocumentNotification(ctx context.Context, userID int64, title, message, link string) error {
	input := &dto.CreateNotificationInput{
		UserID:   userID,
		Type:     entities.NotificationTypeDocument,
		Priority: entities.PriorityNormal,
		Title:    title,
		Message:  message,
		Link:     link,
	}

	_, err := uc.Create(ctx, input)
	return err
}

// SendSystemNotification creates a system notification for a user
func (uc *NotificationUseCase) SendSystemNotification(ctx context.Context, userID int64, title, message string) error {
	input := &dto.CreateNotificationInput{
		UserID:   userID,
		Type:     entities.NotificationTypeSystem,
		Priority: entities.PriorityNormal,
		Title:    title,
		Message:  message,
	}

	_, err := uc.Create(ctx, input)
	return err
}

// BroadcastSystemNotification sends a system notification to multiple users
func (uc *NotificationUseCase) BroadcastSystemNotification(ctx context.Context, userIDs []int64, title, message string) error {
	input := &dto.CreateBulkNotificationInput{
		UserIDs:  userIDs,
		Type:     entities.NotificationTypeSystem,
		Priority: entities.PriorityNormal,
		Title:    title,
		Message:  message,
	}

	_, err := uc.CreateBulk(ctx, input)
	return err
}
