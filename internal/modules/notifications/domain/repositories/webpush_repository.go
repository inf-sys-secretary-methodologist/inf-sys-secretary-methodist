// Package repositories defines repository interfaces for the notifications module.
package repositories

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/entities"
)

// WebPushRepository defines the interface for Web Push subscription persistence operations
type WebPushRepository interface {
	// Create creates a new web push subscription
	Create(ctx context.Context, sub *entities.WebPushSubscription) error

	// GetByID retrieves a subscription by ID
	GetByID(ctx context.Context, id int64) (*entities.WebPushSubscription, error)

	// GetByEndpoint retrieves a subscription by endpoint
	GetByEndpoint(ctx context.Context, endpoint string) (*entities.WebPushSubscription, error)

	// GetByUserID retrieves all subscriptions for a user
	GetByUserID(ctx context.Context, userID int64) ([]*entities.WebPushSubscription, error)

	// GetActiveByUserID retrieves all active subscriptions for a user
	GetActiveByUserID(ctx context.Context, userID int64) ([]*entities.WebPushSubscription, error)

	// Update updates an existing subscription
	Update(ctx context.Context, sub *entities.WebPushSubscription) error

	// UpdateLastUsed updates the last_used_at timestamp
	UpdateLastUsed(ctx context.Context, id int64) error

	// Deactivate marks a subscription as inactive
	Deactivate(ctx context.Context, id int64) error

	// Delete deletes a subscription by ID
	Delete(ctx context.Context, id int64) error

	// DeleteByEndpoint deletes a subscription by endpoint
	DeleteByEndpoint(ctx context.Context, endpoint string) error

	// DeleteByUserID deletes all subscriptions for a user
	DeleteByUserID(ctx context.Context, userID int64) error

	// CountByUserID counts the number of active subscriptions for a user
	CountByUserID(ctx context.Context, userID int64) (int64, error)
}
