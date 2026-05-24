// Package usecases owns the messaging repository ports per Clean
// Architecture DIP (gate from CLAUDE.md: "Repository interfaces — в
// пакете-потребителе (`usecase/`), НЕ в `domain/`"). Mirrors the
// curriculum v0.157.1 polish placement; the announcement port lives
// in this same package for the same reason.
//
// Sentinels and filter DTOs остаются in domain/ (ConversationFilter +
// MessageFilter on entities/; messaging has no domain-level error
// sentinels because participants/access invariants raise entity-level
// errors like ErrNotParticipant из entities/).
package usecases

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/messaging/domain/entities"
)

// ConversationRepository defines the interface for conversation persistence.
type ConversationRepository interface {
	// Create creates a new conversation with participants.
	Create(ctx context.Context, conversation *entities.Conversation) error

	// GetByID retrieves a conversation by ID.
	GetByID(ctx context.Context, id int64) (*entities.Conversation, error)

	// GetDirectConversation finds an existing direct conversation between two users.
	GetDirectConversation(ctx context.Context, userID1, userID2 int64) (*entities.Conversation, error)

	// List returns conversations for a user with pagination.
	List(ctx context.Context, filter entities.ConversationFilter) ([]*entities.Conversation, int64, error)

	// Update updates a conversation.
	Update(ctx context.Context, conversation *entities.Conversation) error

	// Delete deletes a conversation.
	Delete(ctx context.Context, id int64) error

	// AddParticipant adds a participant to a conversation.
	AddParticipant(ctx context.Context, participant *entities.Participant) error

	// RemoveParticipant marks a participant as left.
	RemoveParticipant(ctx context.Context, conversationID, userID int64) error

	// UpdateParticipant updates participant settings (role, muted).
	UpdateParticipant(ctx context.Context, participant *entities.Participant) error

	// GetParticipants returns all active participants of a conversation.
	GetParticipants(ctx context.Context, conversationID int64) ([]entities.Participant, error)

	// GetParticipant returns a specific participant.
	GetParticipant(ctx context.Context, conversationID, userID int64) (*entities.Participant, error)

	// UpdateLastRead updates the last read timestamp for a participant.
	UpdateLastRead(ctx context.Context, conversationID, userID int64, messageID int64) error

	// GetUnreadCount returns the number of unread messages for a user in a conversation.
	GetUnreadCount(ctx context.Context, conversationID, userID int64) (int, error)
}

// MessageRepository defines the interface for message persistence.
type MessageRepository interface {
	// Create creates a new message.
	Create(ctx context.Context, message *entities.Message) error

	// GetByID retrieves a message by ID.
	GetByID(ctx context.Context, id int64) (*entities.Message, error)

	// List returns messages for a conversation with pagination.
	List(ctx context.Context, filter entities.MessageFilter) ([]*entities.Message, error)

	// Update updates a message (for edit/delete).
	Update(ctx context.Context, message *entities.Message) error

	// Delete permanently deletes a message.
	Delete(ctx context.Context, id int64) error

	// GetLastMessage returns the last message in a conversation.
	GetLastMessage(ctx context.Context, conversationID int64) (*entities.Message, error)

	// CountUnread returns the count of unread messages for a user in a conversation.
	CountUnread(ctx context.Context, conversationID, userID int64, lastReadAt *int64) (int, error)

	// CreateAttachment creates a message attachment.
	CreateAttachment(ctx context.Context, attachment *entities.Attachment) error

	// GetAttachments returns attachments for a message.
	GetAttachments(ctx context.Context, messageID int64) ([]entities.Attachment, error)

	// Search searches messages in a conversation.
	Search(ctx context.Context, conversationID int64, query string, limit, offset int) ([]*entities.Message, int64, error)
}
