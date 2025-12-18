package repositories

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/domain/entities"
)

// SyncConflictRepository defines the interface for sync conflict persistence
type SyncConflictRepository interface {
	// Create creates a new sync conflict record
	Create(ctx context.Context, conflict *entities.SyncConflict) error

	// Update updates an existing sync conflict record
	Update(ctx context.Context, conflict *entities.SyncConflict) error

	// GetByID retrieves a sync conflict by ID
	GetByID(ctx context.Context, id int64) (*entities.SyncConflict, error)

	// List retrieves sync conflicts with optional filtering
	List(ctx context.Context, filter entities.SyncConflictFilter) ([]*entities.SyncConflict, int64, error)

	// GetBySyncLogID retrieves all conflicts for a specific sync log
	GetBySyncLogID(ctx context.Context, syncLogID int64) ([]*entities.SyncConflict, error)

	// GetPending retrieves all pending (unresolved) conflicts
	GetPending(ctx context.Context, limit, offset int) ([]*entities.SyncConflict, int64, error)

	// GetPendingByEntityType retrieves pending conflicts for a specific entity type
	GetPendingByEntityType(ctx context.Context, entityType entities.SyncEntityType) ([]*entities.SyncConflict, error)

	// Resolve resolves a conflict with the specified resolution
	Resolve(ctx context.Context, id int64, resolution entities.ConflictResolution, userID int64, resolvedData string) error

	// BulkResolve resolves multiple conflicts with the same resolution
	BulkResolve(ctx context.Context, ids []int64, resolution entities.ConflictResolution, userID int64) error

	// Delete deletes a sync conflict record
	Delete(ctx context.Context, id int64) error

	// DeleteBySyncLogID deletes all conflicts for a specific sync log
	DeleteBySyncLogID(ctx context.Context, syncLogID int64) error

	// GetStats retrieves conflict statistics
	GetStats(ctx context.Context) (*entities.ConflictStats, error)
}
