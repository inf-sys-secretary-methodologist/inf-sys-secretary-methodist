package api

import (
	"context"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/agentsim/agent"
)

type Conversation struct {
	ID          int64  `json:"id"`
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
	CreatedBy   int64  `json:"created_by"`
	UnreadCount int    `json:"unread_count"`
	LastMessage any    `json:"last_message"`
}

type ConversationList struct {
	Conversations []Conversation `json:"conversations"`
	Total         int            `json:"total"`
}

type Message struct {
	ID             int64  `json:"id"`
	ConversationID int64  `json:"conversation_id"`
	SenderID       int64  `json:"sender_id"`
	SenderName     string `json:"sender_name"`
	Content        string `json:"content"`
	Type           string `json:"type"`
	IsEdited       bool   `json:"is_edited"`
	IsDeleted      bool   `json:"is_deleted"`
}

type MessageList struct {
	Messages []Message `json:"messages"`
	HasMore  bool      `json:"has_more"`
}

// CreateDirectConversation creates a direct conversation with another user.
func (c *Client) CreateDirectConversation(ctx context.Context, a *agent.Agent, recipientID int64) (*Conversation, error) {
	body := map[string]any{
		"recipient_id": recipientID,
	}
	resp, err := c.Post(ctx, "/api/conversations/direct", a, body)
	if err != nil {
		return nil, fmt.Errorf("create direct conversation: %w", err)
	}
	var conv Conversation
	if err := ParseData(resp, &conv); err != nil {
		return nil, err
	}
	return &conv, nil
}

// CreateGroupConversation creates a group conversation.
func (c *Client) CreateGroupConversation(ctx context.Context, a *agent.Agent, title, description string, participantIDs []int64) (*Conversation, error) {
	body := map[string]any{
		"title":           title,
		"description":     description,
		"participant_ids": participantIDs,
	}
	resp, err := c.Post(ctx, "/api/conversations/group", a, body)
	if err != nil {
		return nil, fmt.Errorf("create group conversation: %w", err)
	}
	var conv Conversation
	if err := ParseData(resp, &conv); err != nil {
		return nil, err
	}
	return &conv, nil
}

// ListConversations retrieves conversations for the agent.
func (c *Client) ListConversations(ctx context.Context, a *agent.Agent) (*ConversationList, error) {
	resp, err := c.Get(ctx, "/api/conversations", a)
	if err != nil {
		return nil, err
	}
	var list ConversationList
	if err := ParseData(resp, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

// SendMessage sends a message in a conversation.
func (c *Client) SendMessage(ctx context.Context, a *agent.Agent, conversationID int64, content string) (*Message, error) {
	body := map[string]any{
		"content": content,
		"type":    "text",
	}
	resp, err := c.Post(ctx, fmt.Sprintf("/api/conversations/%d/messages", conversationID), a, body)
	if err != nil {
		return nil, fmt.Errorf("send message: %w", err)
	}
	var msg Message
	if err := ParseData(resp, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// GetMessages retrieves messages from a conversation.
func (c *Client) GetMessages(ctx context.Context, a *agent.Agent, conversationID int64, limit int) (*MessageList, error) {
	path := fmt.Sprintf("/api/conversations/%d/messages?limit=%d", conversationID, limit)
	resp, err := c.Get(ctx, path, a)
	if err != nil {
		return nil, err
	}
	var list MessageList
	if err := ParseData(resp, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

// MarkConversationAsRead marks messages in a conversation as read.
func (c *Client) MarkConversationAsRead(ctx context.Context, a *agent.Agent, conversationID int64, messageID int64) error {
	body := map[string]any{
		"message_id": messageID,
	}
	_, err := c.Post(ctx, fmt.Sprintf("/api/conversations/%d/read", conversationID), a, body)
	return err
}
