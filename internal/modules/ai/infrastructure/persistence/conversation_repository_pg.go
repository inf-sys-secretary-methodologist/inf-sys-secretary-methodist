// Package persistence contains repository implementations for the AI module.
package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/repositories"
)

// ConversationRepositoryPg implements ConversationRepository using PostgreSQL
type ConversationRepositoryPg struct {
	db *sql.DB
}

// NewConversationRepositoryPg creates a new PostgreSQL conversation repository
func NewConversationRepositoryPg(db *sql.DB) repositories.ConversationRepository {
	return &ConversationRepositoryPg{db: db}
}

// Create creates a new conversation
func (r *ConversationRepositoryPg) Create(ctx context.Context, conversation *entities.Conversation) error {
	query := `
		INSERT INTO ai_conversations (user_id, title, model, message_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	now := time.Now()
	conversation.CreatedAt = now
	conversation.UpdatedAt = now

	return r.db.QueryRowContext(
		ctx, query,
		conversation.UserID,
		conversation.Title,
		conversation.Model,
		conversation.MessageCount,
		conversation.CreatedAt,
		conversation.UpdatedAt,
	).Scan(&conversation.ID)
}

// GetByID retrieves a conversation by ID
func (r *ConversationRepositoryPg) GetByID(ctx context.Context, id int64) (*entities.Conversation, error) {
	query := `
		SELECT id, user_id, title, model, message_count, last_message_at, created_at, updated_at
		FROM ai_conversations
		WHERE id = $1`

	conversation := &entities.Conversation{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&conversation.ID,
		&conversation.UserID,
		&conversation.Title,
		&conversation.Model,
		&conversation.MessageCount,
		&conversation.LastMessageAt,
		&conversation.CreatedAt,
		&conversation.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("conversation not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	return conversation, nil
}

// GetByUserID retrieves conversations for a user with pagination
func (r *ConversationRepositoryPg) GetByUserID(ctx context.Context, userID int64, limit, offset int) ([]entities.Conversation, int, error) {
	// Count total
	var total int
	countQuery := `SELECT COUNT(*) FROM ai_conversations WHERE user_id = $1`
	if err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count conversations: %w", err)
	}

	// Get conversations
	query := `
		SELECT id, user_id, title, model, message_count, last_message_at, created_at, updated_at
		FROM ai_conversations
		WHERE user_id = $1
		ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query conversations: %w", err)
	}
	defer func() { _ = rows.Close() }()

	conversations := make([]entities.Conversation, 0)
	for rows.Next() {
		var c entities.Conversation
		if err := rows.Scan(
			&c.ID,
			&c.UserID,
			&c.Title,
			&c.Model,
			&c.MessageCount,
			&c.LastMessageAt,
			&c.CreatedAt,
			&c.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan conversation: %w", err)
		}
		conversations = append(conversations, c)
	}

	return conversations, total, nil
}

// Update updates a conversation
func (r *ConversationRepositoryPg) Update(ctx context.Context, conversation *entities.Conversation) error {
	query := `
		UPDATE ai_conversations
		SET title = $1, model = $2, updated_at = $3
		WHERE id = $4`

	conversation.UpdatedAt = time.Now()
	_, err := r.db.ExecContext(ctx, query,
		conversation.Title,
		conversation.Model,
		conversation.UpdatedAt,
		conversation.ID,
	)
	return err
}

// Delete deletes a conversation
func (r *ConversationRepositoryPg) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM ai_conversations WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// Search searches conversations by title
func (r *ConversationRepositoryPg) Search(ctx context.Context, userID int64, query string, limit, offset int) ([]entities.Conversation, int, error) {
	// Count total matching
	var total int
	countQuery := `SELECT COUNT(*) FROM ai_conversations WHERE user_id = $1 AND title ILIKE '%' || $2 || '%'`
	if err := r.db.QueryRowContext(ctx, countQuery, userID, query).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count conversations: %w", err)
	}

	// Get conversations
	searchQuery := `
		SELECT id, user_id, title, model, message_count, last_message_at, created_at, updated_at
		FROM ai_conversations
		WHERE user_id = $1 AND title ILIKE '%' || $2 || '%'
		ORDER BY updated_at DESC
		LIMIT $3 OFFSET $4`

	rows, err := r.db.QueryContext(ctx, searchQuery, userID, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search conversations: %w", err)
	}
	defer func() { _ = rows.Close() }()

	conversations := make([]entities.Conversation, 0)
	for rows.Next() {
		var c entities.Conversation
		if err := rows.Scan(
			&c.ID,
			&c.UserID,
			&c.Title,
			&c.Model,
			&c.MessageCount,
			&c.LastMessageAt,
			&c.CreatedAt,
			&c.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan conversation: %w", err)
		}
		conversations = append(conversations, c)
	}

	return conversations, total, nil
}

// MessageRepositoryPg implements MessageRepository using PostgreSQL
type MessageRepositoryPg struct {
	db *sql.DB
}

// NewMessageRepositoryPg creates a new PostgreSQL message repository
func NewMessageRepositoryPg(db *sql.DB) repositories.MessageRepository {
	return &MessageRepositoryPg{db: db}
}

// Create creates a new message
func (r *MessageRepositoryPg) Create(ctx context.Context, message *entities.Message) error {
	query := `
		INSERT INTO ai_messages (conversation_id, role, content, tokens_used, model, error_message, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`

	message.CreatedAt = time.Now()

	return r.db.QueryRowContext(
		ctx, query,
		message.ConversationID,
		message.Role,
		message.Content,
		message.TokensUsed,
		message.Model,
		message.ErrorMessage,
		message.CreatedAt,
	).Scan(&message.ID)
}

// GetByConversationID retrieves messages for a conversation
func (r *MessageRepositoryPg) GetByConversationID(ctx context.Context, conversationID int64, limit int, beforeID *int64) ([]entities.Message, bool, error) {
	var query string
	var args []interface{}

	if beforeID != nil {
		query = `
			SELECT id, conversation_id, role, content, tokens_used, model, error_message, created_at
			FROM ai_messages
			WHERE conversation_id = $1 AND id < $2
			ORDER BY created_at DESC
			LIMIT $3`
		args = []interface{}{conversationID, *beforeID, limit + 1}
	} else {
		query = `
			SELECT id, conversation_id, role, content, tokens_used, model, error_message, created_at
			FROM ai_messages
			WHERE conversation_id = $1
			ORDER BY created_at DESC
			LIMIT $2`
		args = []interface{}{conversationID, limit + 1}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, false, fmt.Errorf("failed to query messages: %w", err)
	}
	defer func() { _ = rows.Close() }()

	messages := make([]entities.Message, 0)
	for rows.Next() {
		var m entities.Message
		if err := rows.Scan(
			&m.ID,
			&m.ConversationID,
			&m.Role,
			&m.Content,
			&m.TokensUsed,
			&m.Model,
			&m.ErrorMessage,
			&m.CreatedAt,
		); err != nil {
			return nil, false, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, m)
	}

	hasMore := len(messages) > limit
	if hasMore {
		messages = messages[:limit]
	}

	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, hasMore, nil
}

// GetByID retrieves a message by ID
func (r *MessageRepositoryPg) GetByID(ctx context.Context, id int64) (*entities.Message, error) {
	query := `
		SELECT id, conversation_id, role, content, tokens_used, model, error_message, created_at
		FROM ai_messages
		WHERE id = $1`

	message := &entities.Message{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&message.ID,
		&message.ConversationID,
		&message.Role,
		&message.Content,
		&message.TokensUsed,
		&message.Model,
		&message.ErrorMessage,
		&message.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("message not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	return message, nil
}

// CreateMessageSource creates a message source citation
func (r *MessageRepositoryPg) CreateMessageSource(ctx context.Context, messageID, chunkID int64, score float64) error {
	query := `
		INSERT INTO ai_message_sources (message_id, chunk_id, similarity_score, created_at)
		VALUES ($1, $2, $3, $4)`

	_, err := r.db.ExecContext(ctx, query, messageID, chunkID, score, time.Now())
	return err
}

// GetMessageSources retrieves sources for a message
func (r *MessageRepositoryPg) GetMessageSources(ctx context.Context, messageID int64) ([]entities.MessageSource, error) {
	query := `
		SELECT
			ms.id,
			ms.message_id,
			ms.chunk_id,
			c.document_id,
			d.title as document_title,
			c.chunk_text,
			ms.similarity_score,
			c.page_number
		FROM ai_message_sources ms
		JOIN ai_document_chunks c ON ms.chunk_id = c.id
		JOIN documents d ON c.document_id = d.id
		WHERE ms.message_id = $1
		ORDER BY ms.similarity_score DESC`

	rows, err := r.db.QueryContext(ctx, query, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to query message sources: %w", err)
	}
	defer func() { _ = rows.Close() }()

	sources := make([]entities.MessageSource, 0)
	for rows.Next() {
		var s entities.MessageSource
		if err := rows.Scan(
			&s.ID,
			&s.MessageID,
			&s.ChunkID,
			&s.DocumentID,
			&s.DocumentTitle,
			&s.ChunkText,
			&s.SimilarityScore,
			&s.PageNumber,
		); err != nil {
			return nil, fmt.Errorf("failed to scan message source: %w", err)
		}
		sources = append(sources, s)
	}

	return sources, nil
}

// DeleteByConversationID deletes all messages in a conversation
func (r *MessageRepositoryPg) DeleteByConversationID(ctx context.Context, conversationID int64) error {
	// First delete message sources
	sourceQuery := `
		DELETE FROM ai_message_sources
		WHERE message_id IN (SELECT id FROM ai_messages WHERE conversation_id = $1)`
	if _, err := r.db.ExecContext(ctx, sourceQuery, conversationID); err != nil {
		return fmt.Errorf("failed to delete message sources: %w", err)
	}

	// Then delete messages
	query := `DELETE FROM ai_messages WHERE conversation_id = $1`
	_, err := r.db.ExecContext(ctx, query, conversationID)
	return err
}
