// Package entities contains domain entities for the AI module.
package entities

import "time"

// Conversation represents an AI chat conversation
type Conversation struct {
	ID            int64      `json:"id"`
	UserID        int64      `json:"user_id"`
	Title         string     `json:"title"`
	Model         string     `json:"model"`
	MessageCount  int        `json:"message_count"`
	LastMessageAt *time.Time `json:"last_message_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// NewConversation creates a new conversation with the given model name.
func NewConversation(userID int64, title, model string) *Conversation {
	now := time.Now()
	return &Conversation{
		UserID:       userID,
		Title:        title,
		Model:        model,
		MessageCount: 0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
