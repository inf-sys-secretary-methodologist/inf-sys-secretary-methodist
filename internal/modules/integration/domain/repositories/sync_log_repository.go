package repositories

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/domain/entities"
)

// SyncLogRepository defines the interface for sync log persistence
type SyncLogRepository interface {
	// Create creates a new sync log entry
	Create(ctx context.Context, log *entities.SyncLog) error

	// Update updates an existing sync log entry
	Update(ctx context.Context, log *entities.SyncLog) error

	// GetByID retrieves a sync log by ID
	GetByID(ctx context.Context, id int64) (*entities.SyncLog, error)

	// List retrieves sync logs with optional filtering
	List(ctx context.Context, filter entities.SyncLogFilter) ([]*entities.SyncLog, int64, error)

	// GetLatest retrieves the most recent sync log for an entity type
	GetLatest(ctx context.Context, entityType entities.SyncEntityType) (*entities.SyncLog, error)

	// GetRunning retrieves all currently running sync operations
	GetRunning(ctx context.Context) ([]*entities.SyncLog, error)

	// GetStats retrieves sync statistics
	GetStats(ctx context.Context, entityType *entities.SyncEntityType) (*entities.SyncStats, error)

	// Delete deletes a sync log entry
	Delete(ctx context.Context, id int64) error

	// DeleteOlderThan deletes sync logs older than the specified date
	DeleteOlderThan(ctx context.Context, days int) (int64, error)
}
