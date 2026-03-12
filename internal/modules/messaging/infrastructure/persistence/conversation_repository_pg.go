package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/messaging/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/messaging/domain/repositories"
)

// ConversationRepositoryPG implements ConversationRepository using PostgreSQL.
type ConversationRepositoryPG struct {
	db *sql.DB
}

// NewConversationRepositoryPG creates a new PostgreSQL conversation repository.
func NewConversationRepositoryPG(db *sql.DB) repositories.ConversationRepository {
	return &ConversationRepositoryPG{db: db}
}

// Create creates a new conversation with participants.
func (r *ConversationRepositoryPG) Create(ctx context.Context, conversation *entities.Conversation) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Insert conversation
	query := `
		INSERT INTO conversations (type, title, description, avatar_url, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`

	err = tx.QueryRowContext(ctx, query,
		conversation.Type,
		conversation.Title,
		conversation.Description,
		conversation.AvatarURL,
		conversation.CreatedBy,
		conversation.CreatedAt,
		conversation.UpdatedAt,
	).Scan(&conversation.ID)
	if err != nil {
		return fmt.Errorf("failed to create conversation: %w", err)
	}

	// Insert participants
	for i := range conversation.Participants {
		p := &conversation.Participants[i]
		p.ConversationID = conversation.ID

		pQuery := `
			INSERT INTO conversation_participants (conversation_id, user_id, role, is_muted, joined_at)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id`

		err = tx.QueryRowContext(ctx, pQuery,
			p.ConversationID,
			p.UserID,
			p.Role,
			p.IsMuted,
			p.JoinedAt,
		).Scan(&p.ID)
		if err != nil {
			return fmt.Errorf("failed to add participant: %w", err)
		}
	}

	return tx.Commit()
}

// GetByID retrieves a conversation by ID.
func (r *ConversationRepositoryPG) GetByID(ctx context.Context, id int64) (*entities.Conversation, error) {
	query := `
		SELECT id, type, title, description, avatar_url, created_by, created_at, updated_at
		FROM conversations
		WHERE id = $1`

	var conv entities.Conversation
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&conv.ID,
		&conv.Type,
		&conv.Title,
		&conv.Description,
		&conv.AvatarURL,
		&conv.CreatedBy,
		&conv.CreatedAt,
		&conv.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, entities.ErrConversationNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}

	// Load participants
	participants, err := r.GetParticipants(ctx, id)
	if err != nil {
		return nil, err
	}
	conv.Participants = participants

	return &conv, nil
}

// GetDirectConversation finds an existing direct conversation between two users.
func (r *ConversationRepositoryPG) GetDirectConversation(ctx context.Context, userID1, userID2 int64) (*entities.Conversation, error) {
	query := `
		SELECT c.id
		FROM conversations c
		WHERE c.type = 'direct'
		AND EXISTS (
			SELECT 1 FROM conversation_participants cp1
			WHERE cp1.conversation_id = c.id AND cp1.user_id = $1 AND cp1.left_at IS NULL
		)
		AND EXISTS (
			SELECT 1 FROM conversation_participants cp2
			WHERE cp2.conversation_id = c.id AND cp2.user_id = $2 AND cp2.left_at IS NULL
		)
		LIMIT 1`

	var convID int64
	err := r.db.QueryRowContext(ctx, query, userID1, userID2).Scan(&convID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find direct conversation: %w", err)
	}

	return r.GetByID(ctx, convID)
}

// List returns conversations for a user with pagination.
func (r *ConversationRepositoryPG) List(ctx context.Context, filter entities.ConversationFilter) ([]*entities.Conversation, int64, error) {
	var conditions []string
	var args []any
	argNum := 1

	// Base condition: user must be a participant
	conditions = append(conditions, fmt.Sprintf(`
		EXISTS (
			SELECT 1 FROM conversation_participants cp
			WHERE cp.conversation_id = c.id AND cp.user_id = $%d AND cp.left_at IS NULL
		)`, argNum))
	args = append(args, filter.UserID)
	argNum++

	if filter.Type != nil {
		conditions = append(conditions, fmt.Sprintf("c.type = $%d", argNum))
		args = append(args, *filter.Type)
		argNum++
	}

	if filter.Search != nil && *filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(c.title ILIKE $%d OR EXISTS (SELECT 1 FROM conversation_participants cp2 JOIN users u ON u.id = cp2.user_id WHERE cp2.conversation_id = c.id AND u.name ILIKE $%d))", argNum, argNum))
		args = append(args, "%"+*filter.Search+"%")
		argNum++
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count total
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM conversations c WHERE %s`, whereClause) // #nosec G201 -- dynamic WHERE from parameterized conditions, not user input
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count conversations: %w", err)
	}

	// Fetch conversations with last message
	query := fmt.Sprintf(`
		SELECT c.id, c.type, c.title, c.description, c.avatar_url, c.created_by, c.created_at, c.updated_at,
			   m.id, m.sender_id, m.type, m.content, m.created_at,
			   u.name, up.avatar
		FROM conversations c
		LEFT JOIN LATERAL (
			SELECT id, sender_id, type, content, created_at
			FROM messages
			WHERE conversation_id = c.id AND is_deleted = FALSE
			ORDER BY created_at DESC
			LIMIT 1
		) m ON TRUE
		LEFT JOIN users u ON u.id = m.sender_id
		LEFT JOIN user_profiles up ON up.user_id = m.sender_id
		WHERE %s
		ORDER BY COALESCE(m.created_at, c.updated_at) DESC
		LIMIT $%d OFFSET $%d`, whereClause, argNum, argNum+1) // #nosec G201 -- dynamic WHERE from parameterized conditions, not user input

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list conversations: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var conversations []*entities.Conversation
	for rows.Next() {
		var conv entities.Conversation
		var msgID, msgSenderID sql.NullInt64
		var msgType, msgContent sql.NullString
		var msgCreatedAt sql.NullTime
		var senderName sql.NullString
		var senderAvatar sql.NullString

		err := rows.Scan(
			&conv.ID, &conv.Type, &conv.Title, &conv.Description, &conv.AvatarURL,
			&conv.CreatedBy, &conv.CreatedAt, &conv.UpdatedAt,
			&msgID, &msgSenderID, &msgType, &msgContent, &msgCreatedAt,
			&senderName, &senderAvatar,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan conversation: %w", err)
		}

		if msgID.Valid {
			conv.LastMessage = &entities.Message{
				ID:        msgID.Int64,
				SenderID:  msgSenderID.Int64,
				Type:      entities.MessageType(msgType.String),
				Content:   msgContent.String,
				CreatedAt: msgCreatedAt.Time,
			}
			if senderName.Valid {
				conv.LastMessage.SenderName = senderName.String
			}
			if senderAvatar.Valid {
				conv.LastMessage.SenderAvatarURL = &senderAvatar.String
			}
		}

		// Load participants
		participants, err := r.GetParticipants(ctx, conv.ID)
		if err != nil {
			return nil, 0, err
		}
		conv.Participants = participants

		// Get unread count
		unread, err := r.GetUnreadCount(ctx, conv.ID, filter.UserID)
		if err != nil {
			return nil, 0, err
		}
		conv.UnreadCount = unread

		conversations = append(conversations, &conv)
	}

	return conversations, total, nil
}

// Update updates a conversation.
func (r *ConversationRepositoryPG) Update(ctx context.Context, conversation *entities.Conversation) error {
	query := `
		UPDATE conversations
		SET title = $2, description = $3, avatar_url = $4, updated_at = NOW()
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		conversation.ID,
		conversation.Title,
		conversation.Description,
		conversation.AvatarURL,
	)
	if err != nil {
		return fmt.Errorf("failed to update conversation: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return entities.ErrConversationNotFound
	}

	return nil
}

// Delete deletes a conversation.
func (r *ConversationRepositoryPG) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM conversations WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete conversation: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return entities.ErrConversationNotFound
	}

	return nil
}

// AddParticipant adds a participant to a conversation.
func (r *ConversationRepositoryPG) AddParticipant(ctx context.Context, participant *entities.Participant) error {
	query := `
		INSERT INTO conversation_participants (conversation_id, user_id, role, is_muted, joined_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (conversation_id, user_id) DO UPDATE
		SET left_at = NULL, role = EXCLUDED.role, joined_at = EXCLUDED.joined_at
		RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		participant.ConversationID,
		participant.UserID,
		participant.Role,
		participant.IsMuted,
		participant.JoinedAt,
	).Scan(&participant.ID)
	if err != nil {
		return fmt.Errorf("failed to add participant: %w", err)
	}

	return nil
}

// RemoveParticipant marks a participant as left.
func (r *ConversationRepositoryPG) RemoveParticipant(ctx context.Context, conversationID, userID int64) error {
	query := `
		UPDATE conversation_participants
		SET left_at = NOW()
		WHERE conversation_id = $1 AND user_id = $2 AND left_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, conversationID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove participant: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return entities.ErrNotParticipant
	}

	return nil
}

// UpdateParticipant updates participant settings.
func (r *ConversationRepositoryPG) UpdateParticipant(ctx context.Context, participant *entities.Participant) error {
	query := `
		UPDATE conversation_participants
		SET role = $3, is_muted = $4
		WHERE conversation_id = $1 AND user_id = $2 AND left_at IS NULL`

	result, err := r.db.ExecContext(ctx, query,
		participant.ConversationID,
		participant.UserID,
		participant.Role,
		participant.IsMuted,
	)
	if err != nil {
		return fmt.Errorf("failed to update participant: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return entities.ErrNotParticipant
	}

	return nil
}

// GetParticipants returns all active participants of a conversation.
func (r *ConversationRepositoryPG) GetParticipants(ctx context.Context, conversationID int64) ([]entities.Participant, error) {
	query := `
		SELECT cp.id, cp.conversation_id, cp.user_id, cp.role, cp.last_read_at, cp.is_muted, cp.joined_at, cp.left_at,
			   u.name, up.avatar
		FROM conversation_participants cp
		JOIN users u ON u.id = cp.user_id
		LEFT JOIN user_profiles up ON up.user_id = cp.user_id
		WHERE cp.conversation_id = $1 AND cp.left_at IS NULL
		ORDER BY cp.joined_at`

	rows, err := r.db.QueryContext(ctx, query, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participants: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var participants []entities.Participant
	for rows.Next() {
		var p entities.Participant
		err := rows.Scan(
			&p.ID, &p.ConversationID, &p.UserID, &p.Role, &p.LastReadAt, &p.IsMuted, &p.JoinedAt, &p.LeftAt,
			&p.UserName, &p.UserAvatarURL,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan participant: %w", err)
		}
		participants = append(participants, p)
	}

	return participants, nil
}

// GetParticipant returns a specific participant.
func (r *ConversationRepositoryPG) GetParticipant(ctx context.Context, conversationID, userID int64) (*entities.Participant, error) {
	query := `
		SELECT cp.id, cp.conversation_id, cp.user_id, cp.role, cp.last_read_at, cp.is_muted, cp.joined_at, cp.left_at,
			   u.name, up.avatar
		FROM conversation_participants cp
		JOIN users u ON u.id = cp.user_id
		LEFT JOIN user_profiles up ON up.user_id = cp.user_id
		WHERE cp.conversation_id = $1 AND cp.user_id = $2 AND cp.left_at IS NULL`

	var p entities.Participant
	err := r.db.QueryRowContext(ctx, query, conversationID, userID).Scan(
		&p.ID, &p.ConversationID, &p.UserID, &p.Role, &p.LastReadAt, &p.IsMuted, &p.JoinedAt, &p.LeftAt,
		&p.UserName, &p.UserAvatarURL,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, entities.ErrNotParticipant
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get participant: %w", err)
	}

	return &p, nil
}

// UpdateLastRead updates the last read timestamp for a participant.
func (r *ConversationRepositoryPG) UpdateLastRead(ctx context.Context, conversationID, userID int64, messageID int64) error {
	query := `
		UPDATE conversation_participants
		SET last_read_at = NOW(), last_read_message_id = $3
		WHERE conversation_id = $1 AND user_id = $2 AND left_at IS NULL`

	_, err := r.db.ExecContext(ctx, query, conversationID, userID, messageID)
	if err != nil {
		return fmt.Errorf("failed to update last read: %w", err)
	}

	return nil
}

// GetUnreadCount returns the number of unread messages for a user in a conversation.
func (r *ConversationRepositoryPG) GetUnreadCount(ctx context.Context, conversationID, userID int64) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM messages m
		WHERE m.conversation_id = $1
		AND m.is_deleted = FALSE
		AND m.sender_id != $2
		AND m.created_at > COALESCE(
			(SELECT last_read_at FROM conversation_participants
			 WHERE conversation_id = $1 AND user_id = $2),
			'1970-01-01'::timestamptz
		)`

	var count int
	err := r.db.QueryRowContext(ctx, query, conversationID, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}

	return count, nil
}
