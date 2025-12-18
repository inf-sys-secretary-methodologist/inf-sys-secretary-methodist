package dto

import (
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/domain/entities"
)

// SyncLogDTO represents a sync log response
type SyncLogDTO struct {
	ID             int64                    `json:"id"`
	EntityType     entities.SyncEntityType  `json:"entity_type"`
	Direction      entities.SyncDirection   `json:"direction"`
	Status         entities.SyncStatus      `json:"status"`
	StartedAt      time.Time                `json:"started_at"`
	CompletedAt    *time.Time               `json:"completed_at,omitempty"`
	TotalRecords   int                      `json:"total_records"`
	ProcessedCount int                      `json:"processed_count"`
	SuccessCount   int                      `json:"success_count"`
	ErrorCount     int                      `json:"error_count"`
	ConflictCount  int                      `json:"conflict_count"`
	Progress       float64                  `json:"progress"`
	ErrorMessage   string                   `json:"error_message,omitempty"`
	CreatedAt      time.Time                `json:"created_at"`
}

// FromSyncLog converts entity to DTO
func FromSyncLog(log *entities.SyncLog) *SyncLogDTO {
	return &SyncLogDTO{
		ID:             log.ID,
		EntityType:     log.EntityType,
		Direction:      log.Direction,
		Status:         log.Status,
		StartedAt:      log.StartedAt,
		CompletedAt:    log.CompletedAt,
		TotalRecords:   log.TotalRecords,
		ProcessedCount: log.ProcessedCount,
		SuccessCount:   log.SuccessCount,
		ErrorCount:     log.ErrorCount,
		ConflictCount:  log.ConflictCount,
		Progress:       log.GetProgress(),
		ErrorMessage:   log.ErrorMessage,
		CreatedAt:      log.CreatedAt,
	}
}

// SyncStatsDTO represents sync statistics
type SyncStatsDTO struct {
	TotalSyncs      int64     `json:"total_syncs"`
	SuccessfulSyncs int64     `json:"successful_syncs"`
	FailedSyncs     int64     `json:"failed_syncs"`
	TotalRecords    int64     `json:"total_records"`
	TotalConflicts  int64     `json:"total_conflicts"`
	LastSyncAt      time.Time `json:"last_sync_at"`
}

// FromSyncStats converts entity to DTO
func FromSyncStats(stats *entities.SyncStats) *SyncStatsDTO {
	return &SyncStatsDTO{
		TotalSyncs:      stats.TotalSyncs,
		SuccessfulSyncs: stats.SuccessfulSyncs,
		FailedSyncs:     stats.FailedSyncs,
		TotalRecords:    stats.TotalRecords,
		TotalConflicts:  stats.TotalConflicts,
		LastSyncAt:      stats.LastSyncAt,
	}
}

// StartSyncRequest represents a request to start synchronization
type StartSyncRequest struct {
	EntityType entities.SyncEntityType  `json:"entity_type" validate:"required,oneof=employee student finance"`
	Direction  entities.SyncDirection   `json:"direction" validate:"required,oneof=import export both"`
	Force      bool                     `json:"force"` // Force sync even if one is running
}

// SyncListRequest represents a request to list sync logs
type SyncListRequest struct {
	EntityType *entities.SyncEntityType `json:"entity_type,omitempty"`
	Direction  *entities.SyncDirection  `json:"direction,omitempty"`
	Status     *entities.SyncStatus     `json:"status,omitempty"`
	Limit      int                      `json:"limit,omitempty"`
	Offset     int                      `json:"offset,omitempty"`
}

// SyncListResponse represents a paginated list of sync logs
type SyncListResponse struct {
	Items []*SyncLogDTO `json:"items"`
	Total int64         `json:"total"`
}

// SyncResultDTO represents the result of a sync operation
type SyncResultDTO struct {
	SyncLogID      int64  `json:"sync_log_id"`
	Status         string `json:"status"`
	TotalRecords   int    `json:"total_records"`
	ProcessedCount int    `json:"processed_count"`
	SuccessCount   int    `json:"success_count"`
	ErrorCount     int    `json:"error_count"`
	ConflictCount  int    `json:"conflict_count"`
	Message        string `json:"message,omitempty"`
}
