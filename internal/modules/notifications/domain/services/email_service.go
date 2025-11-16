package services

import (
	"context"
)

// EmailService defines the interface for email operations
type EmailService interface {
	// SendEmail sends an email to one or more recipients
	SendEmail(ctx context.Context, req *SendEmailRequest) error

	// SendWelcomeEmail sends a welcome email to a new user
	SendWelcomeEmail(ctx context.Context, recipientEmail, userName string) error

	// SendPasswordResetEmail sends a password reset email
	SendPasswordResetEmail(ctx context.Context, recipientEmail, resetToken string) error

	// SendNotification sends a generic notification email
	SendNotification(ctx context.Context, recipientEmail, subject, body string) error
}

// SendEmailRequest represents a request to send an email
type SendEmailRequest struct {
	To      []string `json:"to"`
	CC      []string `json:"cc,omitempty"`
	BCC     []string `json:"bcc,omitempty"`
	Subject string   `json:"subject"`
	Body    string   `json:"body"`
	IsHTML  bool     `json:"is_html,omitempty"`
}
