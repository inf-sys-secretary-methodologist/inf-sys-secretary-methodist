package repositories

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
