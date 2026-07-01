package usecases

import (
	"context"
	"errors"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/entities"
)

// Sentinel errors used across the conversation-flow boundary (handler →
// usecase → repo). Issue #263 ADR-8: handlers translate ErrConversationNotFound
// to HTTP 404 and ErrConversationAccessDenied to HTTP 403, replacing the
// prior fmt.Errorf strings that collapsed everything to 500 (info disclosure
// + wrong status). Sentinels live in the repository package so both the
// persistence layer (returns NotFound) and the usecase layer (returns
// AccessDenied after ownership check) can reference them without a circular
// import.
var (
	// ErrConversationNotFound indicates the conversation ID does not exist.
	ErrConversationNotFound = errors.New("conversation not found")
	// ErrConversationAccessDenied indicates the caller does not own the
	// conversation; handler translates to 403 Forbidden.
	ErrConversationAccessDenied = errors.New("conversation access denied")
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
