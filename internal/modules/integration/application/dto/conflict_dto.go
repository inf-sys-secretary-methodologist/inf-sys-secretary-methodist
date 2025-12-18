package dto

import (
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/domain/entities"
)

// SyncConflictDTO represents a sync conflict response
type SyncConflictDTO struct {
	ID             int64                       `json:"id"`
	SyncLogID      int64                       `json:"sync_log_id"`
	EntityType     entities.SyncEntityType     `json:"entity_type"`
	EntityID       string                      `json:"entity_id"`
	LocalData      string                      `json:"local_data,omitempty"`
	ExternalData   string                      `json:"external_data,omitempty"`
	ConflictType   string                      `json:"conflict_type"`
	ConflictFields []string                    `json:"conflict_fields"`
	Resolution     entities.ConflictResolution `json:"resolution"`
	ResolvedBy     *int64                      `json:"resolved_by,omitempty"`
	ResolvedAt     *time.Time                  `json:"resolved_at,omitempty"`
	ResolvedData   string                      `json:"resolved_data,omitempty"`
	Notes          string                      `json:"notes,omitempty"`
	CreatedAt      time.Time                   `json:"created_at"`
}

// FromSyncConflict converts entity to DTO
func FromSyncConflict(conflict *entities.SyncConflict) *SyncConflictDTO {
	return &SyncConflictDTO{
		ID:             conflict.ID,
		SyncLogID:      conflict.SyncLogID,
		EntityType:     conflict.EntityType,
		EntityID:       conflict.EntityID,
		LocalData:      conflict.LocalData,
		ExternalData:   conflict.ExternalData,
		ConflictType:   conflict.ConflictType,
		ConflictFields: conflict.ConflictFields,
		Resolution:     conflict.Resolution,
		ResolvedBy:     conflict.ResolvedBy,
		ResolvedAt:     conflict.ResolvedAt,
		ResolvedData:   conflict.ResolvedData,
		Notes:          conflict.Notes,
		CreatedAt:      conflict.CreatedAt,
	}
}

// ConflictListRequest represents a request to list conflicts
type ConflictListRequest struct {
	SyncLogID  *int64                       `json:"sync_log_id,omitempty" form:"sync_log_id"`
	EntityType *entities.SyncEntityType     `json:"entity_type,omitempty" form:"entity_type"`
	Resolution *entities.ConflictResolution `json:"resolution,omitempty" form:"resolution"`
	Limit      int                          `json:"limit,omitempty" form:"limit"`
	Offset     int                          `json:"offset,omitempty" form:"offset"`
}

// ConflictListResponse represents a paginated list of conflicts
type ConflictListResponse struct {
	Items []*SyncConflictDTO `json:"items"`
	Total int64              `json:"total"`
}

// ResolveConflictRequest represents a request to resolve a conflict
type ResolveConflictRequest struct {
	Resolution   entities.ConflictResolution `json:"resolution" validate:"required,oneof=use_local use_external merge skip"`
	ResolvedData string                      `json:"resolved_data,omitempty"`
	Notes        string                      `json:"notes,omitempty"`
}

// BulkResolveRequest represents a request to resolve multiple conflicts
type BulkResolveRequest struct {
	IDs        []int64                     `json:"ids" validate:"required,min=1"`
	Resolution entities.ConflictResolution `json:"resolution" validate:"required,oneof=use_local use_external skip"`
}

// ConflictStatsDTO represents conflict statistics
type ConflictStatsDTO struct {
	TotalConflicts    int64                             `json:"total_conflicts"`
	PendingConflicts  int64                             `json:"pending_conflicts"`
	ResolvedConflicts int64                             `json:"resolved_conflicts"`
	ByEntityType      map[entities.SyncEntityType]int64 `json:"by_entity_type"`
}

// FromConflictStats converts entity to DTO
func FromConflictStats(stats *entities.ConflictStats) *ConflictStatsDTO {
	return &ConflictStatsDTO{
		TotalConflicts:    stats.TotalConflicts,
		PendingConflicts:  stats.PendingConflicts,
		ResolvedConflicts: stats.ResolvedConflicts,
		ByEntityType:      stats.ByEntityType,
	}
}
