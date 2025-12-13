// Package services contains application services for the notifications module.
package services

import (
	"context"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/services"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/composio"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

// ComposioTelegramService implements TelegramService using Composio
type ComposioTelegramService struct {
	client   *composio.Client
	entityID string // User ID for Composio authentication
	auditLog *logging.AuditLogger
}

// NewComposioTelegramService creates a new Telegram service using Composio
func NewComposioTelegramService(apiKey, entityID string, auditLog *logging.AuditLogger) services.TelegramService {
	return &ComposioTelegramService{
		client:   composio.NewClient(apiKey),
		entityID: entityID,
		auditLog: auditLog,
	}
}

// SendMessage sends a text message to a Telegram chat
func (s *ComposioTelegramService) SendMessage(ctx context.Context, req *services.SendTelegramMessageRequest) error {
	if req.ChatID == "" {
		return fmt.Errorf("chat_id is required")
	}
	if req.Text == "" {
		return fmt.Errorf("text is required")
	}

	telegramReq := &composio.SendTelegramMessageRequest{
		ChatID:    req.ChatID,
		Text:      req.Text,
		ParseMode: req.ParseMode,
	}

	_, err := s.client.SendTelegramMessage(ctx, s.entityID, telegramReq)
	if err != nil {
		return fmt.Errorf("failed to send Telegram message: %w", err)
	}

	// Log audit event
	s.logAudit(ctx, "telegram_message_sent", "telegram", map[string]any{
		"chat_id":    req.ChatID,
		"parse_mode": req.ParseMode,
	})

	return nil
}

// SendNotification sends a notification message with proper formatting
func (s *ComposioTelegramService) SendNotification(ctx context.Context, chatID string, title, message string, priority string) error {
	// Format notification with emoji based on priority
	var emoji string
	switch priority {
	case "urgent":
		emoji = "🚨"
	case "high":
		emoji = "⚠️"
	case "normal":
		emoji = "📬"
	case "low":
		emoji = "ℹ️"
	default:
		emoji = "📬"
	}

	// Build HTML formatted message
	text := fmt.Sprintf(
		"%s <b>%s</b>\n\n%s",
		emoji,
		escapeHTML(title),
		escapeHTML(message),
	)

	req := &services.SendTelegramMessageRequest{
		ChatID:    chatID,
		Text:      text,
		ParseMode: "HTML",
	}

	return s.SendMessage(ctx, req)
}

// logAudit safely logs an audit event with nil check
func (s *ComposioTelegramService) logAudit(ctx context.Context, action, resourceType string, details map[string]any) {
	if s.auditLog != nil {
		s.auditLog.LogAuditEvent(ctx, action, resourceType, details)
	}
}

// escapeHTML escapes special HTML characters for Telegram HTML parse mode
func escapeHTML(text string) string {
	replacer := map[string]string{
		"&": "&amp;",
		"<": "&lt;",
		">": "&gt;",
	}

	result := text
	for old, new := range replacer {
		for i := 0; i < len(result); i++ {
			if string(result[i]) == old {
				result = result[:i] + new + result[i+1:]
				i += len(new) - 1
			}
		}
	}
	return result
}
