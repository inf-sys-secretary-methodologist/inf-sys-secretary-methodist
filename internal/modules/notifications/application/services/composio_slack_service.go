// Package services contains application services for the notifications module.
package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/services"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/composio"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

// ComposioSlackService implements SlackService using Composio
type ComposioSlackService struct {
	client   *composio.Client
	entityID string // User ID for Composio authentication
	auditLog *logging.AuditLogger
}

// NewComposioSlackService creates a new Slack service using Composio
func NewComposioSlackService(apiKey, entityID string, auditLog *logging.AuditLogger) services.SlackService {
	return &ComposioSlackService{
		client:   composio.NewClient(apiKey),
		entityID: entityID,
		auditLog: auditLog,
	}
}

// SendChannelMessage sends a message to a Slack channel
func (s *ComposioSlackService) SendChannelMessage(ctx context.Context, req *services.SendSlackChannelMessageRequest) error {
	if req.Channel == "" {
		return fmt.Errorf("channel is required")
	}
	if req.Text == "" {
		return fmt.Errorf("text is required")
	}

	slackReq := &composio.SendSlackMessageRequest{
		Channel: req.Channel,
		Text:    req.Text,
	}

	_, err := s.client.SendSlackMessage(ctx, s.entityID, slackReq)
	if err != nil {
		return fmt.Errorf("failed to send Slack channel message: %w", err)
	}

	// Log audit event
	s.logAudit(ctx, "slack_channel_message_sent", "slack", map[string]any{
		"channel": req.Channel,
	})

	return nil
}

// SendDirectMessage sends a direct message to a Slack user
func (s *ComposioSlackService) SendDirectMessage(ctx context.Context, req *services.SendSlackDirectMessageRequest) error {
	if req.UserID == "" {
		return fmt.Errorf("user_id is required")
	}
	if req.Text == "" {
		return fmt.Errorf("text is required")
	}

	slackReq := &composio.SendSlackDirectMessageRequest{
		UserID: req.UserID,
		Text:   req.Text,
	}

	_, err := s.client.SendSlackDirectMessage(ctx, s.entityID, slackReq)
	if err != nil {
		return fmt.Errorf("failed to send Slack direct message: %w", err)
	}

	// Log audit event
	s.logAudit(ctx, "slack_direct_message_sent", "slack", map[string]any{
		"user_id": req.UserID,
	})

	return nil
}

// SendNotification sends a notification with proper formatting
func (s *ComposioSlackService) SendNotification(ctx context.Context, channelOrUserID string, title, message string, priority string, isDirect bool) error {
	// Format notification with emoji based on priority
	var emoji string
	switch priority {
	case "urgent":
		emoji = ":rotating_light:"
	case "high":
		emoji = ":warning:"
	case "normal":
		emoji = ":mailbox_with_mail:"
	case "low":
		emoji = ":information_source:"
	default:
		emoji = ":mailbox_with_mail:"
	}

	// Build Slack mrkdwn formatted message
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s *%s*\n\n", emoji, title))
	sb.WriteString(message)

	text := sb.String()

	if isDirect {
		return s.SendDirectMessage(ctx, &services.SendSlackDirectMessageRequest{
			UserID: channelOrUserID,
			Text:   text,
		})
	}

	return s.SendChannelMessage(ctx, &services.SendSlackChannelMessageRequest{
		Channel: channelOrUserID,
		Text:    text,
	})
}

// logAudit safely logs an audit event with nil check
func (s *ComposioSlackService) logAudit(ctx context.Context, action, resourceType string, details map[string]any) {
	if s.auditLog != nil {
		s.auditLog.LogAuditEvent(ctx, action, resourceType, details)
	}
}
