// Package services defines domain service interfaces for the notifications module.
package services

import (
	"context"
)

// TelegramService defines the interface for Telegram notification operations
type TelegramService interface {
	// SendMessage sends a text message to a Telegram chat
	SendMessage(ctx context.Context, req *SendTelegramMessageRequest) error

	// SendNotification sends a notification message with proper formatting
	SendNotification(ctx context.Context, chatID string, title, message string, priority string) error
}

// SendTelegramMessageRequest represents a request to send a Telegram message
type SendTelegramMessageRequest struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"` // "HTML" or "Markdown"
}
