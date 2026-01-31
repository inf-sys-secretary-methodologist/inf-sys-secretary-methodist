// Package services defines domain service interfaces for the notifications module.
package services

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/entities"
)

// WebPushService defines the interface for Web Push notification operations
type WebPushService interface {
	// SendNotification sends a push notification to a specific subscription
	SendNotification(ctx context.Context, sub *entities.WebPushSubscription, payload *entities.WebPushPayload) error

	// SendToUser sends a push notification to all active subscriptions for a user
	SendToUser(ctx context.Context, userID int64, payload *entities.WebPushPayload) error

	// GetVAPIDPublicKey returns the VAPID public key for client-side subscription
	GetVAPIDPublicKey() string

	// IsConfigured returns true if the service is properly configured with VAPID keys
	IsConfigured() bool
}
