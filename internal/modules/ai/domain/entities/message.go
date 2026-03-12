// Package entities contains domain entities for the AI module.
package entities

import "time"

// MessageRole represents the role of a message sender
type MessageRole string

// MessageRole constants define the possible roles in a conversation.
const (
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
	MessageRoleSystem    MessageRole = "system"
)

// Message represents a message in an AI conversation
type Message struct {
	ID             int64           `json:"id"`
	ConversationID int64           `json:"conversation_id"`
	Role           MessageRole     `json:"role"`
	Content        string          `json:"content"`
	Sources        []MessageSource `json:"sources,omitempty"`
	TokensUsed     *int            `json:"tokens_used,omitempty"`
	Model          *string         `json:"model,omitempty"`
	ErrorMessage   *string         `json:"error_message,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}

// MessageSource represents a document source cited in an AI response
type MessageSource struct {
	ID              int64   `json:"id"`
	MessageID       int64   `json:"message_id"`
	ChunkID         int64   `json:"chunk_id"`
	DocumentID      int64   `json:"document_id"`
	DocumentTitle   string  `json:"document_title"`
	ChunkText       string  `json:"chunk_text"`
	SimilarityScore float64 `json:"similarity_score"`
	PageNumber      *int    `json:"page_number,omitempty"`
}

// NewUserMessage creates a new user message
func NewUserMessage(conversationID int64, content string) *Message {
	return &Message{
		ConversationID: conversationID,
		Role:           MessageRoleUser,
		Content:        content,
		CreatedAt:      time.Now(),
	}
}

// NewAssistantMessage creates a new assistant message
func NewAssistantMessage(conversationID int64, content string, model string, tokensUsed int) *Message {
	return &Message{
		ConversationID: conversationID,
		Role:           MessageRoleAssistant,
		Content:        content,
		Model:          &model,
		TokensUsed:     &tokensUsed,
		CreatedAt:      time.Now(),
	}
}
