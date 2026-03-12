package entities

import (
	"testing"
	"time"
)

func TestNewNotification(t *testing.T) {
	userID := int64(42)
	notificationType := NotificationTypeInfo
	title := "Test Notification"
	message := "This is a test notification"

	n := NewNotification(userID, notificationType, title, message)

	if n.UserID != userID {
		t.Errorf("expected user ID %d, got %d", userID, n.UserID)
	}
	if n.Type != notificationType {
		t.Errorf("expected type %q, got %q", notificationType, n.Type)
	}
	if n.Title != title {
		t.Errorf("expected title %q, got %q", title, n.Title)
	}
	if n.Message != message {
		t.Errorf("expected message %q, got %q", message, n.Message)
	}
	if n.Priority != PriorityNormal {
		t.Errorf("expected priority %q, got %q", PriorityNormal, n.Priority)
	}
	if n.IsRead {
		t.Error("expected IsRead to be false")
	}
	if n.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestNotification_MarkAsRead(t *testing.T) {
	n := NewNotification(1, NotificationTypeInfo, "Title", "Message")

	n.MarkAsRead()

	if !n.IsRead {
		t.Error("expected IsRead to be true")
	}
	if n.ReadAt == nil {
		t.Error("expected ReadAt to be set")
	}
}

func TestNotification_MarkAsRead_AlreadyRead(t *testing.T) {
	n := NewNotification(1, NotificationTypeInfo, "Title", "Message")
	n.MarkAsRead()
	firstReadAt := *n.ReadAt

	// Mark as read again
	time.Sleep(1 * time.Millisecond) // Ensure time difference
	n.MarkAsRead()

	// ReadAt should not change
	if n.ReadAt.After(firstReadAt) {
		t.Error("expected ReadAt to not change when already read")
	}
}

func TestNotification_IsExpired(t *testing.T) {
	n := NewNotification(1, NotificationTypeInfo, "Title", "Message")

	// No expiry set
	if n.IsExpired() {
		t.Error("expected notification without expiry to not be expired")
	}

	// Expiry in the past
	pastTime := time.Now().Add(-1 * time.Hour)
	n.ExpiresAt = &pastTime

	if !n.IsExpired() {
		t.Error("expected notification with past expiry to be expired")
	}

	// Expiry in the future
	futureTime := time.Now().Add(1 * time.Hour)
	n.ExpiresAt = &futureTime

	if n.IsExpired() {
		t.Error("expected notification with future expiry to not be expired")
	}
}

func TestNotificationTypeConstants(t *testing.T) {
	tests := []struct {
		name      string
		notifType NotificationType
		expected  string
	}{
		{"info", NotificationTypeInfo, "info"},
		{"success", NotificationTypeSuccess, "success"},
		{"warning", NotificationTypeWarning, "warning"},
		{"error", NotificationTypeError, "error"},
		{"reminder", NotificationTypeReminder, "reminder"},
		{"task", NotificationTypeTask, "task"},
		{"document", NotificationTypeDocument, "document"},
		{"event", NotificationTypeEvent, "event"},
		{"system", NotificationTypeSystem, "system"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.notifType) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.notifType)
			}
		})
	}
}

func TestNotificationChannelConstants(t *testing.T) {
	tests := []struct {
		name     string
		channel  NotificationChannel
		expected string
	}{
		{"in_app", ChannelInApp, "in_app"},
		{"email", ChannelEmail, "email"},
		{"push", ChannelPush, "push"},
		{"telegram", ChannelTelegram, "telegram"},
		{"slack", ChannelSlack, "slack"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.channel) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.channel)
			}
		})
	}
}

func TestNotificationPriorityConstants(t *testing.T) {
	tests := []struct {
		name     string
		priority NotificationPriority
		expected string
	}{
		{"low", PriorityLow, "low"},
		{"normal", PriorityNormal, "normal"},
		{"high", PriorityHigh, "high"},
		{"urgent", PriorityUrgent, "urgent"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.priority) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.priority)
			}
		})
	}
}
