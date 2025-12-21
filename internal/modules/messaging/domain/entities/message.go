package entities

import (
	"errors"
	"time"
)

// MessageType represents the type of message.
type MessageType string

const (
	MessageTypeText   MessageType = "text"
	MessageTypeImage  MessageType = "image"
	MessageTypeFile   MessageType = "file"
	MessageTypeSystem MessageType = "system"
)

// MessageStatus represents the delivery status of a message.
type MessageStatus string

const (
	MessageStatusSent      MessageStatus = "sent"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusRead      MessageStatus = "read"
)

// Validation errors.
var (
	ErrMessageNotFound     = errors.New("message not found")
	ErrEmptyMessageContent = errors.New("message content cannot be empty")
	ErrMessageTooLong      = errors.New("message content is too long")
	ErrCannotEditMessage   = errors.New("cannot edit this message")
	ErrCannotDeleteMessage = errors.New("cannot delete this message")
)

// MaxMessageLength is the maximum allowed length for a message.
const MaxMessageLength = 10000

// Message represents a chat message.
type Message struct {
	ID             int64         `db:"id" json:"id"`
	ConversationID int64         `db:"conversation_id" json:"conversation_id"`
	SenderID       int64         `db:"sender_id" json:"sender_id"`
	Type           MessageType   `db:"type" json:"type"`
	Content        string        `db:"content" json:"content"`
	ReplyToID      *int64        `db:"reply_to_id" json:"reply_to_id,omitempty"`
	ReplyTo        *Message      `db:"-" json:"reply_to,omitempty"`
	Attachments    []Attachment  `db:"-" json:"attachments,omitempty"`
	IsEdited       bool          `db:"is_edited" json:"is_edited"`
	EditedAt       *time.Time    `db:"edited_at" json:"edited_at,omitempty"`
	IsDeleted      bool          `db:"is_deleted" json:"is_deleted"`
	DeletedAt      *time.Time    `db:"deleted_at" json:"deleted_at,omitempty"`
	CreatedAt      time.Time     `db:"created_at" json:"created_at"`
	// Sender info (joined)
	SenderName      string  `db:"sender_name" json:"sender_name,omitempty"`
	SenderAvatarURL *string `db:"sender_avatar_url" json:"sender_avatar_url,omitempty"`
}

// Attachment represents a file attached to a message.
type Attachment struct {
	ID        int64     `db:"id" json:"id"`
	MessageID int64     `db:"message_id" json:"message_id"`
	FileID    int64     `db:"file_id" json:"file_id"`
	FileName  string    `db:"file_name" json:"file_name"`
	FileSize  int64     `db:"file_size" json:"file_size"`
	MimeType  string    `db:"mime_type" json:"mime_type"`
	URL       string    `db:"url" json:"url"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// NewTextMessage creates a new text message.
func NewTextMessage(conversationID, senderID int64, content string) (*Message, error) {
	if content == "" {
		return nil, ErrEmptyMessageContent
	}
	if len(content) > MaxMessageLength {
		return nil, ErrMessageTooLong
	}

	return &Message{
		ConversationID: conversationID,
		SenderID:       senderID,
		Type:           MessageTypeText,
		Content:        content,
		CreatedAt:      time.Now(),
	}, nil
}

// NewSystemMessage creates a new system message (e.g., "User joined the chat").
func NewSystemMessage(conversationID int64, content string) *Message {
	return &Message{
		ConversationID: conversationID,
		SenderID:       0, // System messages have no sender
		Type:           MessageTypeSystem,
		Content:        content,
		CreatedAt:      time.Now(),
	}
}

// NewReplyMessage creates a new message as a reply to another message.
func NewReplyMessage(conversationID, senderID int64, content string, replyToID int64) (*Message, error) {
	msg, err := NewTextMessage(conversationID, senderID, content)
	if err != nil {
		return nil, err
	}
	msg.ReplyToID = &replyToID
	return msg, nil
}

// Edit updates the message content.
func (m *Message) Edit(newContent string) error {
	if m.IsDeleted {
		return ErrCannotEditMessage
	}
	if newContent == "" {
		return ErrEmptyMessageContent
	}
	if len(newContent) > MaxMessageLength {
		return ErrMessageTooLong
	}

	m.Content = newContent
	m.IsEdited = true
	now := time.Now()
	m.EditedAt = &now
	return nil
}

// Delete marks the message as deleted.
func (m *Message) Delete() error {
	if m.IsDeleted {
		return ErrCannotDeleteMessage
	}

	m.IsDeleted = true
	now := time.Now()
	m.DeletedAt = &now
	m.Content = "" // Clear content for privacy
	return nil
}

// CanEdit checks if a user can edit this message.
func (m *Message) CanEdit(userID int64) bool {
	// Only the sender can edit their own messages
	// System messages cannot be edited
	return m.SenderID == userID && m.Type != MessageTypeSystem && !m.IsDeleted
}

// CanDelete checks if a user can delete this message.
func (m *Message) CanDelete(userID int64, isAdmin bool) bool {
	// Sender can delete their own messages
	// Admins can delete any message
	// System messages cannot be deleted
	if m.Type == MessageTypeSystem || m.IsDeleted {
		return false
	}
	return m.SenderID == userID || isAdmin
}

// MessageFilter represents filters for listing messages.
type MessageFilter struct {
	ConversationID int64
	BeforeID       *int64
	AfterID        *int64
	SenderID       *int64
	Type           *MessageType
	Search         *string
	Limit          int
}

// MessageReadReceipt tracks when a user read messages in a conversation.
type MessageReadReceipt struct {
	ConversationID int64     `db:"conversation_id" json:"conversation_id"`
	UserID         int64     `db:"user_id" json:"user_id"`
	LastReadAt     time.Time `db:"last_read_at" json:"last_read_at"`
	LastMessageID  int64     `db:"last_message_id" json:"last_message_id"`
}
