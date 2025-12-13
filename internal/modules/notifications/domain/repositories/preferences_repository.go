// Package repositories contains repository interfaces for the notifications module.
package repositories

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/entities"
)

// PreferencesRepository defines the interface for notification preferences persistence
type PreferencesRepository interface {
	// CRUD operations
	Create(ctx context.Context, preferences *entities.UserNotificationPreferences) error
	Update(ctx context.Context, preferences *entities.UserNotificationPreferences) error
	Delete(ctx context.Context, userID int64) error

	// Query operations
	GetByUserID(ctx context.Context, userID int64) (*entities.UserNotificationPreferences, error)

	// Ensure preferences exist (create if not)
	GetOrCreate(ctx context.Context, userID int64) (*entities.UserNotificationPreferences, error)

	// Batch operations for specific settings
	UpdateChannelEnabled(ctx context.Context, userID int64, channel entities.NotificationChannel, enabled bool) error
	UpdateQuietHours(ctx context.Context, userID int64, enabled bool, start, end, timezone string) error
}
