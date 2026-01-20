package entities

import (
	"time"
)

// SyncConflict represents a data conflict during synchronization
type SyncConflict struct {
	ID             int64              `json:"id"`
	SyncLogID      int64              `json:"sync_log_id"`
	EntityType     SyncEntityType     `json:"entity_type"`
	EntityID       string             `json:"entity_id"`       // External ID or local ID
	LocalData      string             `json:"local_data"`      // JSON of local record
	ExternalData   string             `json:"external_data"`   // JSON of external record
	ConflictType   string             `json:"conflict_type"`   // update, delete, create
	ConflictFields []string           `json:"conflict_fields"` // Fields with conflicts
	Resolution     ConflictResolution `json:"resolution"`
	ResolvedBy     *int64             `json:"resolved_by,omitempty"`
	ResolvedAt     *time.Time         `json:"resolved_at,omitempty"`
	ResolvedData   string             `json:"resolved_data,omitempty"` // Final merged data
	Notes          string             `json:"notes,omitempty"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
}

// Conflict types
const (
	ConflictTypeUpdate = "update" // Both sides modified the same record
	ConflictTypeDelete = "delete" // One side deleted, other modified
	ConflictTypeCreate = "create" // Duplicate creation
)

// NewSyncConflict creates a new sync conflict record
func NewSyncConflict(syncLogID int64, entityType SyncEntityType, entityID string) *SyncConflict {
	now := time.Now()
	return &SyncConflict{
		SyncLogID:      syncLogID,
		EntityType:     entityType,
		EntityID:       entityID,
		Resolution:     ConflictResolutionPending,
		ConflictFields: []string{},
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// Resolve marks the conflict as resolved
func (c *SyncConflict) Resolve(resolution ConflictResolution, userID int64, resolvedData string) {
	now := time.Now()
	c.Resolution = resolution
	c.ResolvedBy = &userID
	c.ResolvedAt = &now
	c.ResolvedData = resolvedData
	c.UpdatedAt = now
}

// IsPending returns true if the conflict is not yet resolved
func (c *SyncConflict) IsPending() bool {
	return c.Resolution == ConflictResolutionPending
}

// SetLocalData sets the local data as JSON string
func (c *SyncConflict) SetLocalData(data string) {
	c.LocalData = data
	c.UpdatedAt = time.Now()
}

// SetExternalData sets the external data as JSON string
func (c *SyncConflict) SetExternalData(data string) {
	c.ExternalData = data
	c.UpdatedAt = time.Now()
}

// SetConflictFields sets the list of conflicting fields
func (c *SyncConflict) SetConflictFields(fields []string) {
	c.ConflictFields = fields
	c.UpdatedAt = time.Now()
}

// SyncConflictFilter represents filter options for sync conflicts
type SyncConflictFilter struct {
	SyncLogID  *int64              `json:"sync_log_id,omitempty"`
	EntityType *SyncEntityType     `json:"entity_type,omitempty"`
	Resolution *ConflictResolution `json:"resolution,omitempty"`
	Limit      int                 `json:"limit,omitempty"`
	Offset     int                 `json:"offset,omitempty"`
}

// ConflictStats represents statistics about sync conflicts
type ConflictStats struct {
	TotalConflicts    int64                    `json:"total_conflicts"`
	PendingConflicts  int64                    `json:"pending_conflicts"`
	ResolvedConflicts int64                    `json:"resolved_conflicts"`
	ByEntityType      map[SyncEntityType]int64 `json:"by_entity_type"`
}
