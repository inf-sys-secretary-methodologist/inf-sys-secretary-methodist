// Package services defines domain service interfaces for the notifications module.
package services

import (
	"context"
)

// SlackService defines the interface for Slack notification operations
type SlackService interface {
	// SendChannelMessage sends a message to a Slack channel
	SendChannelMessage(ctx context.Context, req *SendSlackChannelMessageRequest) error

	// SendDirectMessage sends a direct message to a Slack user
	SendDirectMessage(ctx context.Context, req *SendSlackDirectMessageRequest) error

	// SendNotification sends a notification with proper formatting
	SendNotification(ctx context.Context, channelOrUserID string, title, message string, priority string, isDirect bool) error
}

// SendSlackChannelMessageRequest represents a request to send a message to a Slack channel
type SendSlackChannelMessageRequest struct {
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

// SendSlackDirectMessageRequest represents a request to send a direct message to a Slack user
type SendSlackDirectMessageRequest struct {
	UserID string `json:"user_id"`
	Text   string `json:"text"`
}
