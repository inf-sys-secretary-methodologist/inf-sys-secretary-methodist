package api

import (
	"context"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/agentsim/agent"
)

type AIConversation struct {
	ID           int64  `json:"id"`
	Title        string `json:"title"`
	Model        string `json:"model"`
	MessageCount int    `json:"message_count"`
}

type AIChatResponse struct {
	Message        any    `json:"message"`
	ConversationID int64  `json:"conversation_id"`
	MoodState      string `json:"mood_state"`
}

// AIChat sends a message to the AI assistant.
func (c *Client) AIChat(ctx context.Context, a *agent.Agent, content string, conversationID int64) (*AIChatResponse, error) {
	body := map[string]any{
		"content": content,
	}
	if conversationID > 0 {
		body["conversation_id"] = conversationID
	}
	resp, err := c.Post(ctx, "/api/ai/chat", a, body)
	if err != nil {
		return nil, fmt.Errorf("AI chat: %w", err)
	}
	var chatResp AIChatResponse
	if err := ParseData(resp, &chatResp); err != nil {
		return nil, err
	}
	return &chatResp, nil
}

// AISearch performs a semantic search.
func (c *Client) AISearch(ctx context.Context, a *agent.Agent, query string) error {
	body := map[string]any{
		"query": query,
		"limit": 5,
	}
	_, err := c.Post(ctx, "/api/ai/search", a, body)
	return err
}
