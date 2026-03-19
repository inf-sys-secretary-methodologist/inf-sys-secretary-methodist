package entities

import (
	"testing"
)

func TestNewWebPushSubscription(t *testing.T) {
	userID := int64(42)
	endpoint := "https://push.example.com/sub123"
	p256dhKey := "key123"
	authKey := "auth456"

	s := NewWebPushSubscription(userID, endpoint, p256dhKey, authKey)

	if s.UserID != userID {
		t.Errorf("expected user ID %d, got %d", userID, s.UserID)
	}
	if s.Endpoint != endpoint {
		t.Errorf("expected endpoint %q, got %q", endpoint, s.Endpoint)
	}
	if s.P256dhKey != p256dhKey {
		t.Errorf("expected p256dh key %q, got %q", p256dhKey, s.P256dhKey)
	}
	if s.AuthKey != authKey {
		t.Errorf("expected auth key %q, got %q", authKey, s.AuthKey)
	}
	if !s.IsActive {
		t.Error("expected IsActive to be true")
	}
	if s.LastUsedAt != nil {
		t.Error("expected LastUsedAt to be nil")
	}
	if s.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestWebPushSubscription_Deactivate(t *testing.T) {
	s := NewWebPushSubscription(1, "ep", "key", "auth")
	s.Deactivate()
	if s.IsActive {
		t.Error("expected subscription to be deactivated")
	}
}

func TestWebPushSubscription_Activate(t *testing.T) {
	s := NewWebPushSubscription(1, "ep", "key", "auth")
	s.Deactivate()
	s.Activate()
	if !s.IsActive {
		t.Error("expected subscription to be activated")
	}
}

func TestWebPushSubscription_UpdateLastUsed(t *testing.T) {
	s := NewWebPushSubscription(1, "ep", "key", "auth")
	s.UpdateLastUsed()
	if s.LastUsedAt == nil {
		t.Error("expected LastUsedAt to be set")
	}
}

func TestNewWebPushPayload(t *testing.T) {
	p := NewWebPushPayload("Title", "Body")
	if p.Title != "Title" {
		t.Errorf("expected title 'Title', got %q", p.Title)
	}
	if p.Body != "Body" {
		t.Errorf("expected body 'Body', got %q", p.Body)
	}
	if p.Icon != "/icons/icon-192x192.png" {
		t.Errorf("expected default icon, got %q", p.Icon)
	}
	if p.Badge != "/icons/icon-72x72.png" {
		t.Errorf("expected default badge, got %q", p.Badge)
	}
}

func TestWebPushPayload_WithURL(t *testing.T) {
	p := NewWebPushPayload("Title", "Body").WithURL("/test")
	if p.URL != "/test" {
		t.Errorf("expected URL '/test', got %q", p.URL)
	}
}

func TestWebPushPayload_WithTag(t *testing.T) {
	p := NewWebPushPayload("Title", "Body").WithTag("tag1")
	if p.Tag != "tag1" {
		t.Errorf("expected tag 'tag1', got %q", p.Tag)
	}
}

func TestWebPushPayload_WithRequireInteraction(t *testing.T) {
	p := NewWebPushPayload("Title", "Body").WithRequireInteraction(true)
	if !p.RequireInteraction {
		t.Error("expected RequireInteraction to be true")
	}
}

func TestWebPushPayload_WithData(t *testing.T) {
	data := map[string]any{"key": "value"}
	p := NewWebPushPayload("Title", "Body").WithData(data)
	if p.Data == nil {
		t.Error("expected Data to be set")
	}
	if p.Data["key"] != "value" {
		t.Errorf("expected data key=value, got %v", p.Data["key"])
	}
}

func TestWebPushPayload_AddAction(t *testing.T) {
	p := NewWebPushPayload("Title", "Body").
		AddAction("open", "Open").
		AddAction("dismiss", "Dismiss")
	if len(p.Actions) != 2 {
		t.Errorf("expected 2 actions, got %d", len(p.Actions))
	}
	if p.Actions[0].Action != "open" {
		t.Errorf("expected action 'open', got %q", p.Actions[0].Action)
	}
	if p.Actions[0].Title != "Open" {
		t.Errorf("expected title 'Open', got %q", p.Actions[0].Title)
	}
	if p.Actions[1].Action != "dismiss" {
		t.Errorf("expected action 'dismiss', got %q", p.Actions[1].Action)
	}
}

func TestWebPushPayloadFromNotification(t *testing.T) {
	n := NewNotification(1, NotificationTypeTask, "Task Title", "You have a new task")
	n.ID = 42
	n.Link = "/tasks/1"
	n.ImageURL = "/custom-icon.png"
	n.Priority = PriorityHigh
	n.Metadata = map[string]any{"task_id": 1}

	p := WebPushPayloadFromNotification(n)

	if p.Title != "Task Title" {
		t.Errorf("expected title 'Task Title', got %q", p.Title)
	}
	if p.Body != "You have a new task" {
		t.Errorf("expected body, got %q", p.Body)
	}
	if p.URL != "/tasks/1" {
		t.Errorf("expected URL '/tasks/1', got %q", p.URL)
	}
	if p.Icon != "/custom-icon.png" {
		t.Errorf("expected custom icon, got %q", p.Icon)
	}
	if p.Tag != "task" {
		t.Errorf("expected tag 'task', got %q", p.Tag)
	}
	if !p.RequireInteraction {
		t.Error("expected RequireInteraction for high priority")
	}
	if p.Data == nil {
		t.Fatal("expected data to be set")
	}
	if p.Data["notification_id"] != int64(42) {
		t.Errorf("expected notification_id=42, got %v", p.Data["notification_id"])
	}
	if p.Data["type"] != "task" {
		t.Errorf("expected type=task, got %v", p.Data["type"])
	}
	if p.Data["task_id"] != 1 {
		t.Errorf("expected task_id=1, got %v", p.Data["task_id"])
	}
}

func TestWebPushPayloadFromNotification_Urgent(t *testing.T) {
	n := NewNotification(1, NotificationTypeSystem, "System", "Alert")
	n.Priority = PriorityUrgent

	p := WebPushPayloadFromNotification(n)
	if !p.RequireInteraction {
		t.Error("expected RequireInteraction for urgent priority")
	}
}

func TestWebPushPayloadFromNotification_NoLinkNoImage(t *testing.T) {
	n := NewNotification(1, NotificationTypeInfo, "Info", "Message")

	p := WebPushPayloadFromNotification(n)
	if p.URL != "" {
		t.Errorf("expected empty URL, got %q", p.URL)
	}
	// Icon should be the default from NewWebPushPayload
	if p.Icon != "/icons/icon-192x192.png" {
		t.Errorf("expected default icon, got %q", p.Icon)
	}
	if p.RequireInteraction {
		t.Error("expected no RequireInteraction for normal priority")
	}
}

func TestWebPushPayloadFromNotification_NilMetadata(t *testing.T) {
	n := NewNotification(1, NotificationTypeInfo, "Info", "Message")
	n.Metadata = nil

	p := WebPushPayloadFromNotification(n)
	if p.Data == nil {
		t.Fatal("expected data to be set even with nil metadata")
	}
	if p.Data["type"] != "info" {
		t.Errorf("expected type=info, got %v", p.Data["type"])
	}
}

func TestWebPushAction_Struct(t *testing.T) {
	action := WebPushAction{Action: "open", Title: "Open", Icon: "/icon.png"}
	if action.Action != "open" {
		t.Errorf("expected action 'open', got %q", action.Action)
	}
	if action.Title != "Open" {
		t.Errorf("expected title 'Open', got %q", action.Title)
	}
	if action.Icon != "/icon.png" {
		t.Errorf("expected icon '/icon.png', got %q", action.Icon)
	}
}
