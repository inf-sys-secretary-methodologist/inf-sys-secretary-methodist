package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/messaging/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/messaging/domain/repositories"
)

// MessageRepositoryPG implements MessageRepository using PostgreSQL.
type MessageRepositoryPG struct {
	db *sql.DB
}

// NewMessageRepositoryPG creates a new PostgreSQL message repository.
func NewMessageRepositoryPG(db *sql.DB) repositories.MessageRepository {
	return &MessageRepositoryPG{db: db}
}

// Create creates a new message.
func (r *MessageRepositoryPG) Create(ctx context.Context, message *entities.Message) error {
	query := `
		INSERT INTO messages (conversation_id, sender_id, type, content, reply_to_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		message.ConversationID,
		message.SenderID,
		message.Type,
		message.Content,
		message.ReplyToID,
		message.CreatedAt,
	).Scan(&message.ID)
	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	return nil
}

// GetByID retrieves a message by ID.
func (r *MessageRepositoryPG) GetByID(ctx context.Context, id int64) (*entities.Message, error) {
	query := `
		SELECT m.id, m.conversation_id, m.sender_id, m.type, m.content, m.reply_to_id,
			   m.is_edited, m.edited_at, m.is_deleted, m.deleted_at, m.created_at,
			   u.name, up.avatar
		FROM messages m
		LEFT JOIN users u ON u.id = m.sender_id
		LEFT JOIN user_profiles up ON up.user_id = m.sender_id
		WHERE m.id = $1`

	var msg entities.Message
	var senderName sql.NullString
	var senderAvatar sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&msg.ID, &msg.ConversationID, &msg.SenderID, &msg.Type, &msg.Content, &msg.ReplyToID,
		&msg.IsEdited, &msg.EditedAt, &msg.IsDeleted, &msg.DeletedAt, &msg.CreatedAt,
		&senderName, &senderAvatar,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, entities.ErrMessageNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	if senderName.Valid {
		msg.SenderName = senderName.String
	}
	if senderAvatar.Valid {
		msg.SenderAvatarURL = &senderAvatar.String
	}

	// Load reply if exists
	if msg.ReplyToID != nil {
		reply, err := r.GetByID(ctx, *msg.ReplyToID)
		if err == nil {
			msg.ReplyTo = reply
		}
	}

	// Load attachments
	attachments, err := r.GetAttachments(ctx, msg.ID)
	if err != nil {
		return nil, err
	}
	msg.Attachments = attachments

	return &msg, nil
}

// List returns messages for a conversation with pagination.
func (r *MessageRepositoryPG) List(ctx context.Context, filter entities.MessageFilter) ([]*entities.Message, error) {
	var conditions []string
	var args []any
	argNum := 1

	conditions = append(conditions, fmt.Sprintf("m.conversation_id = $%d", argNum))
	args = append(args, filter.ConversationID)
	argNum++

	if filter.BeforeID != nil {
		conditions = append(conditions, fmt.Sprintf("m.id < $%d", argNum))
		args = append(args, *filter.BeforeID)
		argNum++
	}

	if filter.AfterID != nil {
		conditions = append(conditions, fmt.Sprintf("m.id > $%d", argNum))
		args = append(args, *filter.AfterID)
		argNum++
	}

	if filter.SenderID != nil {
		conditions = append(conditions, fmt.Sprintf("m.sender_id = $%d", argNum))
		args = append(args, *filter.SenderID)
		argNum++
	}

	if filter.Type != nil {
		conditions = append(conditions, fmt.Sprintf("m.type = $%d", argNum))
		args = append(args, *filter.Type)
		argNum++
	}

	whereClause := ""
	for i, cond := range conditions {
		if i == 0 {
			whereClause = cond
		} else {
			whereClause += " AND " + cond
		}
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}

	query := fmt.Sprintf(`
		SELECT m.id, m.conversation_id, m.sender_id, m.type, m.content, m.reply_to_id,
			   m.is_edited, m.edited_at, m.is_deleted, m.deleted_at, m.created_at,
			   u.name, up.avatar
		FROM messages m
		LEFT JOIN users u ON u.id = m.sender_id
		LEFT JOIN user_profiles up ON up.user_id = m.sender_id
		WHERE %s
		ORDER BY m.created_at DESC
		LIMIT $%d`, whereClause, argNum)

	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var messages []*entities.Message
	for rows.Next() {
		var msg entities.Message
		var senderName sql.NullString
		var senderAvatar sql.NullString

		err := rows.Scan(
			&msg.ID, &msg.ConversationID, &msg.SenderID, &msg.Type, &msg.Content, &msg.ReplyToID,
			&msg.IsEdited, &msg.EditedAt, &msg.IsDeleted, &msg.DeletedAt, &msg.CreatedAt,
			&senderName, &senderAvatar,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		if senderName.Valid {
			msg.SenderName = senderName.String
		}
		if senderAvatar.Valid {
			msg.SenderAvatarURL = &senderAvatar.String
		}

		messages = append(messages, &msg)
	}

	// Load replies and attachments for each message
	for _, msg := range messages {
		if msg.ReplyToID != nil {
			reply, err := r.getMessageBasic(ctx, *msg.ReplyToID)
			if err == nil {
				msg.ReplyTo = reply
			}
		}

		attachments, err := r.GetAttachments(ctx, msg.ID)
		if err == nil {
			msg.Attachments = attachments
		}
	}

	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// getMessageBasic retrieves basic message info (without nested data).
func (r *MessageRepositoryPG) getMessageBasic(ctx context.Context, id int64) (*entities.Message, error) {
	query := `
		SELECT m.id, m.conversation_id, m.sender_id, m.type, m.content,
			   m.is_edited, m.is_deleted, m.created_at,
			   u.name, up.avatar
		FROM messages m
		LEFT JOIN users u ON u.id = m.sender_id
		LEFT JOIN user_profiles up ON up.user_id = m.sender_id
		WHERE m.id = $1`

	var msg entities.Message
	var senderName sql.NullString
	var senderAvatar sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&msg.ID, &msg.ConversationID, &msg.SenderID, &msg.Type, &msg.Content,
		&msg.IsEdited, &msg.IsDeleted, &msg.CreatedAt,
		&senderName, &senderAvatar,
	)
	if err != nil {
		return nil, err
	}

	if senderName.Valid {
		msg.SenderName = senderName.String
	}
	if senderAvatar.Valid {
		msg.SenderAvatarURL = &senderAvatar.String
	}

	return &msg, nil
}

// Update updates a message.
func (r *MessageRepositoryPG) Update(ctx context.Context, message *entities.Message) error {
	query := `
		UPDATE messages
		SET content = $2, is_edited = $3, edited_at = $4, is_deleted = $5, deleted_at = $6
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		message.ID,
		message.Content,
		message.IsEdited,
		message.EditedAt,
		message.IsDeleted,
		message.DeletedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update message: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return entities.ErrMessageNotFound
	}

	return nil
}

// Delete permanently deletes a message.
func (r *MessageRepositoryPG) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM messages WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return entities.ErrMessageNotFound
	}

	return nil
}

// GetLastMessage returns the last message in a conversation.
func (r *MessageRepositoryPG) GetLastMessage(ctx context.Context, conversationID int64) (*entities.Message, error) {
	query := `
		SELECT m.id, m.conversation_id, m.sender_id, m.type, m.content, m.reply_to_id,
			   m.is_edited, m.edited_at, m.is_deleted, m.deleted_at, m.created_at,
			   u.name, up.avatar
		FROM messages m
		LEFT JOIN users u ON u.id = m.sender_id
		LEFT JOIN user_profiles up ON up.user_id = m.sender_id
		WHERE m.conversation_id = $1 AND m.is_deleted = FALSE
		ORDER BY m.created_at DESC
		LIMIT 1`

	var msg entities.Message
	var senderName sql.NullString
	var senderAvatar sql.NullString

	err := r.db.QueryRowContext(ctx, query, conversationID).Scan(
		&msg.ID, &msg.ConversationID, &msg.SenderID, &msg.Type, &msg.Content, &msg.ReplyToID,
		&msg.IsEdited, &msg.EditedAt, &msg.IsDeleted, &msg.DeletedAt, &msg.CreatedAt,
		&senderName, &senderAvatar,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get last message: %w", err)
	}

	if senderName.Valid {
		msg.SenderName = senderName.String
	}
	if senderAvatar.Valid {
		msg.SenderAvatarURL = &senderAvatar.String
	}

	return &msg, nil
}

// CountUnread returns the count of unread messages.
func (r *MessageRepositoryPG) CountUnread(ctx context.Context, conversationID, userID int64, lastReadMsgID *int64) (int, error) {
	var query string
	var args []any

	if lastReadMsgID != nil {
		query = `
			SELECT COUNT(*)
			FROM messages
			WHERE conversation_id = $1 AND id > $2 AND sender_id != $3 AND is_deleted = FALSE`
		args = []any{conversationID, *lastReadMsgID, userID}
	} else {
		query = `
			SELECT COUNT(*)
			FROM messages
			WHERE conversation_id = $1 AND sender_id != $2 AND is_deleted = FALSE`
		args = []any{conversationID, userID}
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count unread: %w", err)
	}

	return count, nil
}

// CreateAttachment creates a message attachment.
func (r *MessageRepositoryPG) CreateAttachment(ctx context.Context, attachment *entities.Attachment) error {
	query := `
		INSERT INTO message_attachments (message_id, file_id, file_name, file_size, mime_type, url, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		attachment.MessageID,
		attachment.FileID,
		attachment.FileName,
		attachment.FileSize,
		attachment.MimeType,
		attachment.URL,
		attachment.CreatedAt,
	).Scan(&attachment.ID)
	if err != nil {
		return fmt.Errorf("failed to create attachment: %w", err)
	}

	return nil
}

// GetAttachments returns attachments for a message.
func (r *MessageRepositoryPG) GetAttachments(ctx context.Context, messageID int64) ([]entities.Attachment, error) {
	query := `
		SELECT id, message_id, file_id, file_name, file_size, mime_type, url, created_at
		FROM message_attachments
		WHERE message_id = $1
		ORDER BY created_at`

	rows, err := r.db.QueryContext(ctx, query, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get attachments: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var attachments []entities.Attachment
	for rows.Next() {
		var a entities.Attachment
		err := rows.Scan(&a.ID, &a.MessageID, &a.FileID, &a.FileName, &a.FileSize, &a.MimeType, &a.URL, &a.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan attachment: %w", err)
		}
		attachments = append(attachments, a)
	}

	return attachments, nil
}

// Search searches messages in a conversation.
func (r *MessageRepositoryPG) Search(ctx context.Context, conversationID int64, query string, limit, offset int) ([]*entities.Message, int64, error) {
	// Count total matches
	countQuery := `
		SELECT COUNT(*)
		FROM messages
		WHERE conversation_id = $1 AND is_deleted = FALSE AND search_vector @@ plainto_tsquery('russian', $2)`

	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, conversationID, query).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count search results: %w", err)
	}

	// Fetch matches
	searchQuery := `
		SELECT m.id, m.conversation_id, m.sender_id, m.type, m.content,
			   m.is_edited, m.edited_at, m.is_deleted, m.deleted_at, m.created_at,
			   u.name, up.avatar,
			   ts_rank(m.search_vector, plainto_tsquery('russian', $2)) as rank
		FROM messages m
		LEFT JOIN users u ON u.id = m.sender_id
		LEFT JOIN user_profiles up ON up.user_id = m.sender_id
		WHERE m.conversation_id = $1 AND m.is_deleted = FALSE
		AND m.search_vector @@ plainto_tsquery('russian', $2)
		ORDER BY rank DESC, m.created_at DESC
		LIMIT $3 OFFSET $4`

	rows, err := r.db.QueryContext(ctx, searchQuery, conversationID, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search messages: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var messages []*entities.Message
	for rows.Next() {
		var msg entities.Message
		var senderName sql.NullString
		var senderAvatar sql.NullString
		var rank float64

		err := rows.Scan(
			&msg.ID, &msg.ConversationID, &msg.SenderID, &msg.Type, &msg.Content,
			&msg.IsEdited, &msg.EditedAt, &msg.IsDeleted, &msg.DeletedAt, &msg.CreatedAt,
			&senderName, &senderAvatar, &rank,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan message: %w", err)
		}

		if senderName.Valid {
			msg.SenderName = senderName.String
		}
		if senderAvatar.Valid {
			msg.SenderAvatarURL = &senderAvatar.String
		}

		messages = append(messages, &msg)
	}

	return messages, total, nil
}
