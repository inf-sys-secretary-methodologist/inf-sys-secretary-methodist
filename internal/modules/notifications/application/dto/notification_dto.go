// Package dto contains data transfer objects for the notifications module.
package dto

import (
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/entities"
)

// CreateNotificationInput represents input for creating a notification
type CreateNotificationInput struct {
	UserID    int64                         `json:"user_id" validate:"required"`
	Type      entities.NotificationType     `json:"type" validate:"required,oneof=info success warning error reminder task document event system"`
	Priority  entities.NotificationPriority `json:"priority" validate:"omitempty,oneof=low normal high urgent"`
	Title     string                        `json:"title" validate:"required,max=500"`
	Message   string                        `json:"message" validate:"required"`
	Link      string                        `json:"link,omitempty" validate:"omitempty,url,max=1000"`
	ImageURL  string                        `json:"image_url,omitempty" validate:"omitempty,url,max=1000"`
	ExpiresAt *time.Time                    `json:"expires_at,omitempty"`
	Metadata  map[string]any                `json:"metadata,omitempty"`
}

// CreateBulkNotificationInput represents input for creating notifications for multiple users
type CreateBulkNotificationInput struct {
	UserIDs   []int64                       `json:"user_ids" validate:"required,min=1"`
	Type      entities.NotificationType     `json:"type" validate:"required,oneof=info success warning error reminder task document event system"`
	Priority  entities.NotificationPriority `json:"priority" validate:"omitempty,oneof=low normal high urgent"`
	Title     string                        `json:"title" validate:"required,max=500"`
	Message   string                        `json:"message" validate:"required"`
	Link      string                        `json:"link,omitempty" validate:"omitempty,url,max=1000"`
	ImageURL  string                        `json:"image_url,omitempty" validate:"omitempty,url,max=1000"`
	ExpiresAt *time.Time                    `json:"expires_at,omitempty"`
	Metadata  map[string]any                `json:"metadata,omitempty"`
}

// NotificationOutput represents a notification in API response
type NotificationOutput struct {
	ID        int64                         `json:"id"`
	UserID    int64                         `json:"user_id"`
	Type      entities.NotificationType     `json:"type"`
	Priority  entities.NotificationPriority `json:"priority"`
	Title     string                        `json:"title"`
	Message   string                        `json:"message"`
	Link      string                        `json:"link,omitempty"`
	ImageURL  string                        `json:"image_url,omitempty"`
	IsRead    bool                          `json:"is_read"`
	ReadAt    *time.Time                    `json:"read_at,omitempty"`
	ExpiresAt *time.Time                    `json:"expires_at,omitempty"`
	Metadata  map[string]any                `json:"metadata,omitempty"`
	CreatedAt time.Time                     `json:"created_at"`
	TimeAgo   string                        `json:"time_ago"`
}

// NotificationListInput represents input for listing notifications
type NotificationListInput struct {
	UserID   int64                         `json:"user_id"`
	Type     entities.NotificationType     `json:"type,omitempty"`
	Priority entities.NotificationPriority `json:"priority,omitempty"`
	IsRead   *bool                         `json:"is_read,omitempty"`
	Limit    int                           `json:"limit" validate:"omitempty,min=1,max=100"`
	Offset   int                           `json:"offset" validate:"omitempty,min=0"`
}

// NotificationListOutput represents a list of notifications with pagination
type NotificationListOutput struct {
	Notifications []*NotificationOutput `json:"notifications"`
	TotalCount    int64                 `json:"total_count"`
	UnreadCount   int64                 `json:"unread_count"`
	Limit         int                   `json:"limit"`
	Offset        int                   `json:"offset"`
}

// NotificationStatsOutput represents notification statistics
type NotificationStatsOutput struct {
	TotalCount   int64 `json:"total_count"`
	UnreadCount  int64 `json:"unread_count"`
	TodayCount   int64 `json:"today_count"`
	UrgentCount  int64 `json:"urgent_count"`
	ExpiredCount int64 `json:"expired_count"`
}

// UnreadCountOutput represents unread count response
type UnreadCountOutput struct {
	Count int64 `json:"count"`
}

// ToOutput converts an entity to output DTO
func ToOutput(n *entities.Notification) *NotificationOutput {
	return &NotificationOutput{
		ID:        n.ID,
		UserID:    n.UserID,
		Type:      n.Type,
		Priority:  n.Priority,
		Title:     n.Title,
		Message:   n.Message,
		Link:      n.Link,
		ImageURL:  n.ImageURL,
		IsRead:    n.IsRead,
		ReadAt:    n.ReadAt,
		ExpiresAt: n.ExpiresAt,
		Metadata:  n.Metadata,
		CreatedAt: n.CreatedAt,
		TimeAgo:   formatTimeAgo(n.CreatedAt),
	}
}

// ToOutputList converts a list of entities to output DTOs
func ToOutputList(notifications []*entities.Notification) []*NotificationOutput {
	result := make([]*NotificationOutput, len(notifications))
	for i, n := range notifications {
		result[i] = ToOutput(n)
	}
	return result
}

// ToEntity converts input DTO to entity
func (input *CreateNotificationInput) ToEntity() *entities.Notification {
	priority := input.Priority
	if priority == "" {
		priority = entities.PriorityNormal
	}

	notification := entities.NewNotification(input.UserID, input.Type, input.Title, input.Message)
	notification.Priority = priority
	notification.Link = input.Link
	notification.ImageURL = input.ImageURL
	notification.ExpiresAt = input.ExpiresAt
	notification.Metadata = input.Metadata

	return notification
}

// formatTimeAgo formats time as a human-readable relative time
func formatTimeAgo(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "только что"
	case diff < time.Hour:
		mins := int(diff.Minutes())
		return formatRussianPlural(mins, "минуту", "минуты", "минут") + " назад"
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		return formatRussianPlural(hours, "час", "часа", "часов") + " назад"
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		return formatRussianPlural(days, "день", "дня", "дней") + " назад"
	case diff < 30*24*time.Hour:
		weeks := int(diff.Hours() / 24 / 7)
		return formatRussianPlural(weeks, "неделю", "недели", "недель") + " назад"
	default:
		return t.Format("02.01.2006")
	}
}

// formatRussianPlural formats a number with Russian plural forms
func formatRussianPlural(n int, one, few, many string) string {
	if n%10 == 1 && n%100 != 11 {
		return one
	}
	if n%10 >= 2 && n%10 <= 4 && (n%100 < 10 || n%100 >= 20) {
		return few
	}
	return many
}
