// Package repositories contains repository interfaces for the notifications module.
package repositories

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/entities"
)

// NotificationRepository defines the interface for notification persistence
type NotificationRepository interface {
	// CRUD operations
	Create(ctx context.Context, notification *entities.Notification) error
	Update(ctx context.Context, notification *entities.Notification) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*entities.Notification, error)

	// Query operations
	List(ctx context.Context, filter *entities.NotificationFilter) ([]*entities.Notification, error)
	GetByUserID(ctx context.Context, userID int64, limit, offset int) ([]*entities.Notification, error)
	GetUnreadByUserID(ctx context.Context, userID int64) ([]*entities.Notification, error)

	// Batch operations
	MarkAsRead(ctx context.Context, id int64) error
	MarkAllAsRead(ctx context.Context, userID int64) error
	DeleteByUserID(ctx context.Context, userID int64) error
	DeleteExpired(ctx context.Context) (int64, error)

	// Statistics
	GetUnreadCount(ctx context.Context, userID int64) (int64, error)
	GetStats(ctx context.Context, userID int64) (*entities.NotificationStats, error)

	// Bulk create for sending to multiple users
	CreateBulk(ctx context.Context, notifications []*entities.Notification) error
}
