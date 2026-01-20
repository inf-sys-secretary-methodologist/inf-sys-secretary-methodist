// Package entities contains domain entities for the notifications module.
package entities

import (
	"time"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeInfo     NotificationType = "info"
	NotificationTypeSuccess  NotificationType = "success"
	NotificationTypeWarning  NotificationType = "warning"
	NotificationTypeError    NotificationType = "error"
	NotificationTypeReminder NotificationType = "reminder"
	NotificationTypeTask     NotificationType = "task"
	NotificationTypeDocument NotificationType = "document"
	NotificationTypeEvent    NotificationType = "event"
	NotificationTypeSystem   NotificationType = "system"
)

// NotificationChannel represents the delivery channel
type NotificationChannel string

const (
	ChannelInApp    NotificationChannel = "in_app"
	ChannelEmail    NotificationChannel = "email"
	ChannelPush     NotificationChannel = "push"
	ChannelTelegram NotificationChannel = "telegram"
	ChannelSlack    NotificationChannel = "slack"
)

// NotificationPriority represents the priority level
type NotificationPriority string

const (
	PriorityLow    NotificationPriority = "low"
	PriorityNormal NotificationPriority = "normal"
	PriorityHigh   NotificationPriority = "high"
	PriorityUrgent NotificationPriority = "urgent"
)

// Notification represents an in-app notification
type Notification struct {
	ID        int64                `json:"id"`
	UserID    int64                `json:"user_id"`
	Type      NotificationType     `json:"type"`
	Priority  NotificationPriority `json:"priority"`
	Title     string               `json:"title"`
	Message   string               `json:"message"`
	Link      string               `json:"link,omitempty"`
	ImageURL  string               `json:"image_url,omitempty"`
	IsRead    bool                 `json:"is_read"`
	ReadAt    *time.Time           `json:"read_at,omitempty"`
	ExpiresAt *time.Time           `json:"expires_at,omitempty"`
	Metadata  map[string]any       `json:"metadata,omitempty"`
	CreatedAt time.Time            `json:"created_at"`
	UpdatedAt time.Time            `json:"updated_at"`
}

// NewNotification creates a new notification with default values
func NewNotification(userID int64, notificationType NotificationType, title, message string) *Notification {
	now := time.Now()
	return &Notification{
		UserID:    userID,
		Type:      notificationType,
		Priority:  PriorityNormal,
		Title:     title,
		Message:   message,
		IsRead:    false,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// MarkAsRead marks the notification as read
func (n *Notification) MarkAsRead() {
	if !n.IsRead {
		now := time.Now()
		n.IsRead = true
		n.ReadAt = &now
		n.UpdatedAt = now
	}
}

// IsExpired checks if the notification has expired
func (n *Notification) IsExpired() bool {
	if n.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*n.ExpiresAt)
}

// NotificationFilter represents filter criteria for listing notifications
type NotificationFilter struct {
	UserID   int64
	Type     NotificationType
	Priority NotificationPriority
	IsRead   *bool
	Limit    int
	Offset   int
}

// NotificationStats represents notification statistics for a user
type NotificationStats struct {
	TotalCount   int64 `json:"total_count"`
	UnreadCount  int64 `json:"unread_count"`
	TodayCount   int64 `json:"today_count"`
	UrgentCount  int64 `json:"urgent_count"`
	ExpiredCount int64 `json:"expired_count"`
}
