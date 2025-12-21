package usecases

import (
	"context"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/messaging/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/messaging/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/messaging/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/messaging/infrastructure/websocket"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

// MessageNotifier is an interface for sending notifications about new messages.
// This allows the messaging module to send notifications without depending on the notification module.
type MessageNotifier interface {
	// NotifyNewMessage sends a notification about a new message to the specified user.
	NotifyNewMessage(ctx context.Context, userID int64, senderName, content string, conversationID, messageID int64) error
}

// MessagingUseCase implements messaging business logic.
type MessagingUseCase struct {
	conversationRepo repositories.ConversationRepository
	messageRepo      repositories.MessageRepository
	hub              *websocket.Hub
	logger           *logging.Logger
	notifier         MessageNotifier
}

// NewMessagingUseCase creates a new messaging use case.
func NewMessagingUseCase(
	conversationRepo repositories.ConversationRepository,
	messageRepo repositories.MessageRepository,
	hub *websocket.Hub,
	logger *logging.Logger,
	notifier MessageNotifier,
) *MessagingUseCase {
	return &MessagingUseCase{
		conversationRepo: conversationRepo,
		messageRepo:      messageRepo,
		hub:              hub,
		logger:           logger,
		notifier:         notifier,
	}
}

// CreateDirectConversation creates a new direct conversation between two users.
func (uc *MessagingUseCase) CreateDirectConversation(ctx context.Context, creatorID int64, input dto.CreateDirectConversationInput) (*entities.Conversation, error) {
	// Check if direct conversation already exists
	existing, err := uc.conversationRepo.GetDirectConversation(ctx, creatorID, input.RecipientID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing conversation: %w", err)
	}
	if existing != nil {
		return existing, nil
	}

	// Create new direct conversation
	conv := entities.NewDirectConversation(creatorID, input.RecipientID)
	if err := uc.conversationRepo.Create(ctx, conv); err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	uc.logger.Info("direct conversation created", map[string]interface{}{
		"conversation_id": conv.ID,
		"creator_id":      creatorID,
		"recipient_id":    input.RecipientID,
	})

	return conv, nil
}

// CreateGroupConversation creates a new group conversation.
func (uc *MessagingUseCase) CreateGroupConversation(ctx context.Context, creatorID int64, input dto.CreateGroupConversationInput) (*entities.Conversation, error) {
	conv := entities.NewGroupConversation(creatorID, input.Title, input.ParticipantIDs)
	if input.Description != nil {
		conv.Description = input.Description
	}

	if err := uc.conversationRepo.Create(ctx, conv); err != nil {
		return nil, fmt.Errorf("failed to create group conversation: %w", err)
	}

	// Send system message
	systemMsg := entities.NewSystemMessage(conv.ID, "Group created")
	if err := uc.messageRepo.Create(ctx, systemMsg); err != nil {
		uc.logger.Error("failed to create system message", map[string]interface{}{
			"error":           err.Error(),
			"conversation_id": conv.ID,
		})
	}

	uc.logger.Info("group conversation created", map[string]interface{}{
		"conversation_id":    conv.ID,
		"creator_id":         creatorID,
		"participants_count": len(conv.Participants),
	})

	return conv, nil
}

// GetConversation retrieves a conversation by ID.
func (uc *MessagingUseCase) GetConversation(ctx context.Context, userID, conversationID int64) (*entities.Conversation, error) {
	conv, err := uc.conversationRepo.GetByID(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	// Check if user is a participant
	if !conv.HasParticipant(userID) {
		return nil, entities.ErrNotParticipant
	}

	// Get unread count
	unread, err := uc.conversationRepo.GetUnreadCount(ctx, conversationID, userID)
	if err != nil {
		uc.logger.Error("failed to get unread count", map[string]interface{}{
			"error":           err.Error(),
			"conversation_id": conversationID,
		})
	}
	conv.UnreadCount = unread

	// Get last message
	lastMsg, err := uc.messageRepo.GetLastMessage(ctx, conversationID)
	if err != nil {
		uc.logger.Error("failed to get last message", map[string]interface{}{
			"error":           err.Error(),
			"conversation_id": conversationID,
		})
	}
	conv.LastMessage = lastMsg

	return conv, nil
}

// ListConversations lists conversations for a user.
func (uc *MessagingUseCase) ListConversations(ctx context.Context, userID int64, input dto.ConversationFilterInput) ([]*entities.Conversation, int64, error) {
	filter := entities.ConversationFilter{
		UserID: userID,
		Search: input.Search,
		Limit:  input.Limit,
		Offset: input.Offset,
	}

	if input.Type != nil {
		convType := entities.ConversationType(*input.Type)
		filter.Type = &convType
	}

	if filter.Limit <= 0 {
		filter.Limit = 20
	}

	return uc.conversationRepo.List(ctx, filter)
}

// UpdateConversation updates a conversation.
func (uc *MessagingUseCase) UpdateConversation(ctx context.Context, userID, conversationID int64, input dto.UpdateConversationInput) (*entities.Conversation, error) {
	conv, err := uc.conversationRepo.GetByID(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	// Only admins can update group conversations
	if conv.IsGroupConversation() && !conv.IsAdmin(userID) {
		return nil, entities.ErrNotParticipant
	}

	if input.Title != nil {
		conv.Title = input.Title
	}
	if input.Description != nil {
		conv.Description = input.Description
	}
	if input.AvatarURL != nil {
		conv.AvatarURL = input.AvatarURL
	}

	if err := uc.conversationRepo.Update(ctx, conv); err != nil {
		return nil, fmt.Errorf("failed to update conversation: %w", err)
	}

	// Broadcast update to all participants
	uc.hub.BroadcastToConversation(conversationID, &websocket.Event{
		Type:           websocket.EventTypeConvUpdated,
		ConversationID: conversationID,
		Payload:        dto.ToConversationOutput(conv, userID),
	}, 0)

	return conv, nil
}

// AddParticipants adds participants to a group conversation.
func (uc *MessagingUseCase) AddParticipants(ctx context.Context, userID, conversationID int64, input dto.AddParticipantsInput) error {
	conv, err := uc.conversationRepo.GetByID(ctx, conversationID)
	if err != nil {
		return err
	}

	if conv.IsDirectConversation() {
		return entities.ErrCannotAddToDirectChat
	}

	if !conv.IsAdmin(userID) {
		return entities.ErrNotParticipant
	}

	for _, newUserID := range input.UserIDs {
		if conv.HasParticipant(newUserID) {
			continue
		}

		participant := &entities.Participant{
			ConversationID: conversationID,
			UserID:         newUserID,
			Role:           entities.ParticipantRoleMember,
		}
		if err := uc.conversationRepo.AddParticipant(ctx, participant); err != nil {
			return fmt.Errorf("failed to add participant: %w", err)
		}

		// Send system message
		systemMsg := entities.NewSystemMessage(conversationID, "User joined the chat")
		if err := uc.messageRepo.Create(ctx, systemMsg); err != nil {
			uc.logger.Error("failed to create system message", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	return nil
}

// LeaveConversation removes the current user from a conversation.
func (uc *MessagingUseCase) LeaveConversation(ctx context.Context, userID, conversationID int64) error {
	conv, err := uc.conversationRepo.GetByID(ctx, conversationID)
	if err != nil {
		return err
	}

	if conv.IsDirectConversation() {
		return entities.ErrCannotLeaveDirectChat
	}

	if !conv.HasParticipant(userID) {
		return entities.ErrNotParticipant
	}

	// Check if user is the last admin
	if conv.IsAdmin(userID) {
		adminCount := 0
		for _, p := range conv.Participants {
			if p.Role == entities.ParticipantRoleAdmin && p.LeftAt == nil {
				adminCount++
			}
		}
		if adminCount <= 1 {
			return entities.ErrCannotRemoveLastAdmin
		}
	}

	if err := uc.conversationRepo.RemoveParticipant(ctx, conversationID, userID); err != nil {
		return fmt.Errorf("failed to leave conversation: %w", err)
	}

	// Send system message
	systemMsg := entities.NewSystemMessage(conversationID, "User left the chat")
	if err := uc.messageRepo.Create(ctx, systemMsg); err != nil {
		uc.logger.Error("failed to create system message", map[string]interface{}{
			"error": err.Error(),
		})
	}

	return nil
}

// SendMessage sends a message to a conversation.
func (uc *MessagingUseCase) SendMessage(ctx context.Context, userID, conversationID int64, input dto.SendMessageInput) (*entities.Message, error) {
	// Verify user is a participant
	participant, err := uc.conversationRepo.GetParticipant(ctx, conversationID, userID)
	if err != nil {
		return nil, err
	}

	var msg *entities.Message
	if input.ReplyToID != nil {
		msg, err = entities.NewReplyMessage(conversationID, userID, input.Content, *input.ReplyToID)
	} else {
		msg, err = entities.NewTextMessage(conversationID, userID, input.Content)
	}
	if err != nil {
		return nil, err
	}

	if err := uc.messageRepo.Create(ctx, msg); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Create attachments if provided
	if len(input.Attachments) > 0 {
		msg.Attachments = make([]entities.Attachment, 0, len(input.Attachments))
		for _, attachInput := range input.Attachments {
			attachment := &entities.Attachment{
				MessageID: msg.ID,
				FileID:    attachInput.FileID,
				FileName:  attachInput.FileName,
				FileSize:  attachInput.FileSize,
				MimeType:  attachInput.MimeType,
				URL:       attachInput.URL,
			}
			if err := uc.messageRepo.CreateAttachment(ctx, attachment); err != nil {
				uc.logger.Error("failed to create attachment", map[string]interface{}{
					"error":      err.Error(),
					"message_id": msg.ID,
					"file_id":    attachInput.FileID,
				})
				continue
			}
			msg.Attachments = append(msg.Attachments, *attachment)
		}

		// Update message type based on attachments
		if msg.Type == entities.MessageTypeText && len(msg.Attachments) > 0 {
			// Determine type from first attachment
			mimeType := msg.Attachments[0].MimeType
			if len(mimeType) >= 5 && mimeType[:5] == "image" {
				msg.Type = entities.MessageTypeImage
			} else {
				msg.Type = entities.MessageTypeFile
			}
		}
	}

	// Fill sender info
	msg.SenderName = participant.UserName
	msg.SenderAvatarURL = participant.UserAvatarURL

	// Load reply if exists
	if msg.ReplyToID != nil {
		reply, err := uc.messageRepo.GetByID(ctx, *msg.ReplyToID)
		if err == nil {
			msg.ReplyTo = reply
		}
	}

	// Broadcast message to conversation
	uc.hub.BroadcastToConversation(conversationID, &websocket.Event{
		Type:           websocket.EventTypeNewMessage,
		ConversationID: conversationID,
		UserID:         userID,
		Payload:        dto.ToMessageOutput(msg),
	}, 0) // Don't exclude sender - they should see their own message

	// Send notifications to participants who are not online in the conversation
	if uc.notifier != nil {
		go uc.notifyParticipants(ctx, conversationID, userID, msg)
	}

	uc.logger.Debug("message sent", map[string]any{
		"message_id":      msg.ID,
		"conversation_id": conversationID,
		"sender_id":       userID,
	})

	return msg, nil
}

// notifyParticipants sends notifications to all participants except the sender.
func (uc *MessagingUseCase) notifyParticipants(ctx context.Context, conversationID, senderID int64, msg *entities.Message) {
	// Get all participants
	participants, err := uc.conversationRepo.GetParticipants(ctx, conversationID)
	if err != nil {
		uc.logger.Error("failed to get participants for notification", map[string]any{
			"error":           err.Error(),
			"conversation_id": conversationID,
		})
		return
	}

	// Truncate content for notification preview
	content := msg.Content
	if len(content) > 100 {
		content = content[:97] + "..."
	}

	// Send notification to each participant (except sender)
	for _, p := range participants {
		if p.UserID == senderID || p.LeftAt != nil {
			continue
		}

		// Skip if user is currently online in the conversation
		if uc.hub.IsUserOnline(p.UserID) {
			continue
		}

		if err := uc.notifier.NotifyNewMessage(ctx, p.UserID, msg.SenderName, content, conversationID, msg.ID); err != nil {
			uc.logger.Error("failed to send message notification", map[string]any{
				"error":           err.Error(),
				"user_id":         p.UserID,
				"conversation_id": conversationID,
				"message_id":      msg.ID,
			})
		}
	}
}

// GetMessages retrieves messages from a conversation.
func (uc *MessagingUseCase) GetMessages(ctx context.Context, userID, conversationID int64, input dto.MessageFilterInput) ([]*entities.Message, bool, error) {
	// Verify user is a participant
	if _, err := uc.conversationRepo.GetParticipant(ctx, conversationID, userID); err != nil {
		return nil, false, err
	}

	limit := input.Limit
	if limit <= 0 {
		limit = 50
	}

	filter := entities.MessageFilter{
		ConversationID: conversationID,
		BeforeID:       input.BeforeID,
		AfterID:        input.AfterID,
		Search:         input.Search,
		Limit:          limit + 1, // Fetch one extra to check if there are more
	}

	messages, err := uc.messageRepo.List(ctx, filter)
	if err != nil {
		return nil, false, fmt.Errorf("failed to list messages: %w", err)
	}

	hasMore := len(messages) > limit
	if hasMore {
		messages = messages[:limit]
	}

	return messages, hasMore, nil
}

// EditMessage edits a message.
func (uc *MessagingUseCase) EditMessage(ctx context.Context, userID, messageID int64, input dto.EditMessageInput) (*entities.Message, error) {
	msg, err := uc.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		return nil, err
	}

	if !msg.CanEdit(userID) {
		return nil, entities.ErrCannotEditMessage
	}

	if err := msg.Edit(input.Content); err != nil {
		return nil, err
	}

	if err := uc.messageRepo.Update(ctx, msg); err != nil {
		return nil, fmt.Errorf("failed to update message: %w", err)
	}

	// Broadcast update
	uc.hub.BroadcastToConversation(msg.ConversationID, &websocket.Event{
		Type:           websocket.EventTypeMessageUpdated,
		ConversationID: msg.ConversationID,
		Payload:        dto.ToMessageOutput(msg),
	}, 0)

	return msg, nil
}

// DeleteMessage deletes a message.
func (uc *MessagingUseCase) DeleteMessage(ctx context.Context, userID, messageID int64) error {
	msg, err := uc.messageRepo.GetByID(ctx, messageID)
	if err != nil {
		return err
	}

	// Check if user is admin of conversation
	conv, err := uc.conversationRepo.GetByID(ctx, msg.ConversationID)
	if err != nil {
		return err
	}

	if !msg.CanDelete(userID, conv.IsAdmin(userID)) {
		return entities.ErrCannotDeleteMessage
	}

	if err := msg.Delete(); err != nil {
		return err
	}

	if err := uc.messageRepo.Update(ctx, msg); err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	// Broadcast deletion
	uc.hub.BroadcastToConversation(msg.ConversationID, &websocket.Event{
		Type:           websocket.EventTypeMessageDeleted,
		ConversationID: msg.ConversationID,
		Payload: map[string]int64{
			"message_id": messageID,
		},
	}, 0)

	return nil
}

// MarkAsRead marks messages as read up to a specific message.
func (uc *MessagingUseCase) MarkAsRead(ctx context.Context, userID, conversationID, messageID int64) error {
	// Verify user is a participant
	if _, err := uc.conversationRepo.GetParticipant(ctx, conversationID, userID); err != nil {
		return err
	}

	if err := uc.conversationRepo.UpdateLastRead(ctx, conversationID, userID, messageID); err != nil {
		return fmt.Errorf("failed to mark as read: %w", err)
	}

	// Broadcast read receipt
	uc.hub.BroadcastToConversation(conversationID, &websocket.Event{
		Type:           websocket.EventTypeRead,
		ConversationID: conversationID,
		UserID:         userID,
		Payload: map[string]int64{
			"message_id": messageID,
		},
	}, userID)

	return nil
}

// SearchMessages searches messages in a conversation.
func (uc *MessagingUseCase) SearchMessages(ctx context.Context, userID, conversationID int64, query string, limit, offset int) ([]*entities.Message, int64, error) {
	// Verify user is a participant
	if _, err := uc.conversationRepo.GetParticipant(ctx, conversationID, userID); err != nil {
		return nil, 0, err
	}

	if limit <= 0 {
		limit = 20
	}

	return uc.messageRepo.Search(ctx, conversationID, query, limit, offset)
}
