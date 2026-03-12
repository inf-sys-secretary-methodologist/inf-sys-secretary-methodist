// Package services contains application services for the notifications module.
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	webpush "github.com/SherClockHolmes/webpush-go"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/services"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

// WebPushServiceImpl implements WebPushService
type WebPushServiceImpl struct {
	webpushRepo     repositories.WebPushRepository
	vapidPublicKey  string
	vapidPrivateKey string
	vapidSubject    string
	auditLog        *logging.AuditLogger
}

// NewWebPushService creates a new WebPush service
func NewWebPushService(
	webpushRepo repositories.WebPushRepository,
	vapidPublicKey, vapidPrivateKey, vapidSubject string,
	auditLog *logging.AuditLogger,
) services.WebPushService {
	return &WebPushServiceImpl{
		webpushRepo:     webpushRepo,
		vapidPublicKey:  vapidPublicKey,
		vapidPrivateKey: vapidPrivateKey,
		vapidSubject:    vapidSubject,
		auditLog:        auditLog,
	}
}

// SendNotification sends a push notification to a specific subscription
func (s *WebPushServiceImpl) SendNotification(ctx context.Context, sub *entities.WebPushSubscription, payload *entities.WebPushPayload) error {
	if !s.IsConfigured() {
		return fmt.Errorf("webpush service is not configured")
	}

	// Marshal payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create webpush subscription object
	subscription := &webpush.Subscription{
		Endpoint: sub.Endpoint,
		Keys: webpush.Keys{
			P256dh: sub.P256dhKey,
			Auth:   sub.AuthKey,
		},
	}

	// Create webpush options
	options := &webpush.Options{
		Subscriber:      s.vapidSubject,
		VAPIDPublicKey:  s.vapidPublicKey,
		VAPIDPrivateKey: s.vapidPrivateKey,
		TTL:             86400, // 24 hours
		Urgency:         webpush.UrgencyNormal,
	}

	// Set urgency based on requireInteraction flag
	if payload.RequireInteraction {
		options.Urgency = webpush.UrgencyHigh
	}

	// Send the notification
	resp, err := webpush.SendNotification(payloadBytes, subscription, options)
	if err != nil {
		slog.Error("Failed to send web push notification",
			"subscription_id", sub.ID,
			"error", err,
		)
		return fmt.Errorf("failed to send notification: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Check response status
	if resp.StatusCode == http.StatusGone || resp.StatusCode == http.StatusNotFound {
		// Subscription is no longer valid - deactivate it
		slog.Info("Web push subscription expired, deactivating",
			"subscription_id", sub.ID,
			"status_code", resp.StatusCode,
		)
		if deactivateErr := s.webpushRepo.Deactivate(ctx, sub.ID); deactivateErr != nil {
			slog.Warn("Failed to deactivate expired subscription",
				"subscription_id", sub.ID,
				"error", deactivateErr,
			)
		}
		return fmt.Errorf("subscription expired or invalid: status %d", resp.StatusCode)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("push service returned error: status %d", resp.StatusCode)
	}

	// Update last used timestamp
	if updateErr := s.webpushRepo.UpdateLastUsed(ctx, sub.ID); updateErr != nil {
		slog.Warn("Failed to update subscription last_used_at",
			"subscription_id", sub.ID,
			"error", updateErr,
		)
	}

	// Log audit event
	s.logAudit(ctx, "webpush_notification_sent", "webpush", map[string]any{
		"subscription_id": sub.ID,
		"user_id":         sub.UserID,
		"title":           payload.Title,
	})

	slog.Info("Web push notification sent",
		"subscription_id", sub.ID,
		"user_id", sub.UserID,
		"status_code", resp.StatusCode,
	)

	return nil
}

// SendToUser sends a push notification to all active subscriptions for a user
func (s *WebPushServiceImpl) SendToUser(ctx context.Context, userID int64, payload *entities.WebPushPayload) error {
	if !s.IsConfigured() {
		return fmt.Errorf("webpush service is not configured")
	}

	// Get all active subscriptions for the user
	subscriptions, err := s.webpushRepo.GetActiveByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user subscriptions: %w", err)
	}

	if len(subscriptions) == 0 {
		slog.Debug("No active web push subscriptions for user", "user_id", userID)
		return nil
	}

	// Send notification to all subscriptions
	var lastErr error
	successCount := 0
	for _, sub := range subscriptions {
		if err := s.SendNotification(ctx, sub, payload); err != nil {
			slog.Warn("Failed to send web push to subscription",
				"subscription_id", sub.ID,
				"user_id", userID,
				"error", err,
			)
			lastErr = err
		} else {
			successCount++
		}
	}

	slog.Info("Web push notifications sent to user",
		"user_id", userID,
		"total_subscriptions", len(subscriptions),
		"successful", successCount,
	)

	// Return error only if all sends failed
	if successCount == 0 && lastErr != nil {
		return lastErr
	}

	return nil
}

// GetVAPIDPublicKey returns the VAPID public key for client-side subscription
func (s *WebPushServiceImpl) GetVAPIDPublicKey() string {
	return s.vapidPublicKey
}

// IsConfigured returns true if the service is properly configured with VAPID keys
func (s *WebPushServiceImpl) IsConfigured() bool {
	return s.vapidPublicKey != "" && s.vapidPrivateKey != "" && s.vapidSubject != ""
}

// logAudit safely logs an audit event with nil check
func (s *WebPushServiceImpl) logAudit(ctx context.Context, action, resourceType string, details map[string]any) {
	if s.auditLog != nil {
		s.auditLog.LogAuditEvent(ctx, action, resourceType, details)
	}
}
