package repositories

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/messaging/domain/entities"
)

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
