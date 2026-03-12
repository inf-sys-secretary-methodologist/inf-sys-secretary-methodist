package entities

import (
	"strings"
	"testing"
)

func TestNewTextMessage(t *testing.T) {
	conversationID := int64(1)
	senderID := int64(42)
	content := "Hello, World!"

	msg, err := NewTextMessage(conversationID, senderID, content)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if msg.ConversationID != conversationID {
		t.Errorf("expected conversation ID %d, got %d", conversationID, msg.ConversationID)
	}
	if msg.SenderID != senderID {
		t.Errorf("expected sender ID %d, got %d", senderID, msg.SenderID)
	}
	if msg.Content != content {
		t.Errorf("expected content %q, got %q", content, msg.Content)
	}
	if msg.Type != MessageTypeText {
		t.Errorf("expected type %q, got %q", MessageTypeText, msg.Type)
	}
	if msg.IsEdited {
		t.Error("expected IsEdited to be false")
	}
	if msg.IsDeleted {
		t.Error("expected IsDeleted to be false")
	}
}

func TestNewTextMessage_EmptyContent(t *testing.T) {
	_, err := NewTextMessage(1, 1, "")

	if err != ErrEmptyMessageContent {
		t.Errorf("expected error %v, got %v", ErrEmptyMessageContent, err)
	}
}

func TestNewTextMessage_TooLong(t *testing.T) {
	longContent := strings.Repeat("a", MaxMessageLength+1)

	_, err := NewTextMessage(1, 1, longContent)

	if err != ErrMessageTooLong {
		t.Errorf("expected error %v, got %v", ErrMessageTooLong, err)
	}
}

func TestNewSystemMessage(t *testing.T) {
	conversationID := int64(1)
	content := "User joined the chat"

	msg := NewSystemMessage(conversationID, content)

	if msg.ConversationID != conversationID {
		t.Errorf("expected conversation ID %d, got %d", conversationID, msg.ConversationID)
	}
	if msg.SenderID != 0 {
		t.Errorf("expected sender ID 0, got %d", msg.SenderID)
	}
	if msg.Content != content {
		t.Errorf("expected content %q, got %q", content, msg.Content)
	}
	if msg.Type != MessageTypeSystem {
		t.Errorf("expected type %q, got %q", MessageTypeSystem, msg.Type)
	}
}

func TestNewReplyMessage(t *testing.T) {
	replyToID := int64(10)

	msg, err := NewReplyMessage(1, 42, "Reply content", replyToID)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if msg.ReplyToID == nil || *msg.ReplyToID != replyToID {
		t.Errorf("expected reply to ID %d, got %v", replyToID, msg.ReplyToID)
	}
}

func TestNewReplyMessage_EmptyContent(t *testing.T) {
	_, err := NewReplyMessage(1, 42, "", 10)

	if err != ErrEmptyMessageContent {
		t.Errorf("expected error %v, got %v", ErrEmptyMessageContent, err)
	}
}

func TestMessage_Edit(t *testing.T) {
	msg, _ := NewTextMessage(1, 42, "Original content")

	err := msg.Edit("New content")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if msg.Content != "New content" {
		t.Errorf("expected content %q, got %q", "New content", msg.Content)
	}
	if !msg.IsEdited {
		t.Error("expected IsEdited to be true")
	}
	if msg.EditedAt == nil {
		t.Error("expected EditedAt to be set")
	}
}

func TestMessage_Edit_Deleted(t *testing.T) {
	msg, _ := NewTextMessage(1, 42, "Original content")
	msg.Delete()

	err := msg.Edit("New content")

	if err != ErrCannotEditMessage {
		t.Errorf("expected error %v, got %v", ErrCannotEditMessage, err)
	}
}

func TestMessage_Edit_EmptyContent(t *testing.T) {
	msg, _ := NewTextMessage(1, 42, "Original content")

	err := msg.Edit("")

	if err != ErrEmptyMessageContent {
		t.Errorf("expected error %v, got %v", ErrEmptyMessageContent, err)
	}
}

func TestMessage_Edit_TooLong(t *testing.T) {
	msg, _ := NewTextMessage(1, 42, "Original content")
	longContent := strings.Repeat("a", MaxMessageLength+1)

	err := msg.Edit(longContent)

	if err != ErrMessageTooLong {
		t.Errorf("expected error %v, got %v", ErrMessageTooLong, err)
	}
}

func TestMessage_Delete(t *testing.T) {
	msg, _ := NewTextMessage(1, 42, "Content")

	err := msg.Delete()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !msg.IsDeleted {
		t.Error("expected IsDeleted to be true")
	}
	if msg.DeletedAt == nil {
		t.Error("expected DeletedAt to be set")
	}
	if msg.Content != "" {
		t.Errorf("expected content to be empty, got %q", msg.Content)
	}
}

func TestMessage_Delete_AlreadyDeleted(t *testing.T) {
	msg, _ := NewTextMessage(1, 42, "Content")
	msg.Delete()

	err := msg.Delete()

	if err != ErrCannotDeleteMessage {
		t.Errorf("expected error %v, got %v", ErrCannotDeleteMessage, err)
	}
}

func TestMessage_CanEdit(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(*Message)
		userID int64
		want   bool
	}{
		{
			name:   "sender can edit own message",
			setup:  func(m *Message) {},
			userID: 42,
			want:   true,
		},
		{
			name:   "other user cannot edit",
			setup:  func(m *Message) {},
			userID: 99,
			want:   false,
		},
		{
			name:   "cannot edit deleted message",
			setup:  func(m *Message) { m.Delete() },
			userID: 42,
			want:   false,
		},
		{
			name: "cannot edit system message",
			setup: func(m *Message) {
				m.Type = MessageTypeSystem
			},
			userID: 42,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, _ := NewTextMessage(1, 42, "Content")
			tt.setup(msg)

			got := msg.CanEdit(tt.userID)
			if got != tt.want {
				t.Errorf("CanEdit() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessage_CanDelete(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*Message)
		userID  int64
		isAdmin bool
		want    bool
	}{
		{
			name:    "sender can delete own message",
			setup:   func(m *Message) {},
			userID:  42,
			isAdmin: false,
			want:    true,
		},
		{
			name:    "admin can delete any message",
			setup:   func(m *Message) {},
			userID:  99,
			isAdmin: true,
			want:    true,
		},
		{
			name:    "non-sender non-admin cannot delete",
			setup:   func(m *Message) {},
			userID:  99,
			isAdmin: false,
			want:    false,
		},
		{
			name:    "cannot delete already deleted",
			setup:   func(m *Message) { m.Delete() },
			userID:  42,
			isAdmin: false,
			want:    false,
		},
		{
			name: "cannot delete system message",
			setup: func(m *Message) {
				m.Type = MessageTypeSystem
			},
			userID:  42,
			isAdmin: true,
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, _ := NewTextMessage(1, 42, "Content")
			tt.setup(msg)

			got := msg.CanDelete(tt.userID, tt.isAdmin)
			if got != tt.want {
				t.Errorf("CanDelete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessageTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		msgType  MessageType
		expected string
	}{
		{"text", MessageTypeText, "text"},
		{"image", MessageTypeImage, "image"},
		{"file", MessageTypeFile, "file"},
		{"system", MessageTypeSystem, "system"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.msgType) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.msgType)
			}
		})
	}
}

func TestMessageStatusConstants(t *testing.T) {
	tests := []struct {
		name     string
		status   MessageStatus
		expected string
	}{
		{"sent", MessageStatusSent, "sent"},
		{"delivered", MessageStatusDelivered, "delivered"},
		{"read", MessageStatusRead, "read"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.status)
			}
		})
	}
}
