// Package repositories contains repository interfaces for the AI module.
package repositories

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/entities"
)

// ConversationRepository defines the interface for AI conversation persistence
type ConversationRepository interface {
	// Create creates a new conversation
	Create(ctx context.Context, conversation *entities.Conversation) error

	// GetByID retrieves a conversation by ID
	GetByID(ctx context.Context, id int64) (*entities.Conversation, error)

	// GetByUserID retrieves conversations for a user with pagination
	GetByUserID(ctx context.Context, userID int64, limit, offset int) ([]entities.Conversation, int, error)

	// Update updates a conversation
	Update(ctx context.Context, conversation *entities.Conversation) error

	// Delete deletes a conversation
	Delete(ctx context.Context, id int64) error

	// Search searches conversations by title
	Search(ctx context.Context, userID int64, query string, limit, offset int) ([]entities.Conversation, int, error)
}

// MessageRepository defines the interface for AI message persistence
type MessageRepository interface {
	// Create creates a new message
	Create(ctx context.Context, message *entities.Message) error

	// GetByConversationID retrieves messages for a conversation
	GetByConversationID(ctx context.Context, conversationID int64, limit int, beforeID *int64) ([]entities.Message, bool, error)

	// GetByID retrieves a message by ID
	GetByID(ctx context.Context, id int64) (*entities.Message, error)

	// CreateMessageSource creates a message source citation
	CreateMessageSource(ctx context.Context, messageID, chunkID int64, score float64) error

	// GetMessageSources retrieves sources for a message
	GetMessageSources(ctx context.Context, messageID int64) ([]entities.MessageSource, error)

	// DeleteByConversationID deletes all messages in a conversation
	DeleteByConversationID(ctx context.Context, conversationID int64) error
}
