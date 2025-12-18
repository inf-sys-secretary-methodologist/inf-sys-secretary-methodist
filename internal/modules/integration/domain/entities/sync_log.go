package entities

import (
	"time"
)

// SyncLog represents a synchronization operation log entry
type SyncLog struct {
	ID             int64          `db:"id" json:"id"`
	EntityType     SyncEntityType `db:"entity_type" json:"entity_type"`
	Direction      SyncDirection  `db:"direction" json:"direction"`
	Status         SyncStatus     `db:"status" json:"status"`
	StartedAt      time.Time      `db:"started_at" json:"started_at"`
	CompletedAt    *time.Time     `db:"completed_at" json:"completed_at,omitempty"`
	TotalRecords   int            `db:"total_records" json:"total_records"`
	ProcessedCount int            `db:"processed_count" json:"processed_count"`
	SuccessCount   int            `db:"success_count" json:"success_count"`
	ErrorCount     int            `db:"error_count" json:"error_count"`
	ConflictCount  int            `db:"conflict_count" json:"conflict_count"`
	ErrorMessage   string         `db:"error_message" json:"error_message,omitempty"`
	Metadata       map[string]any `db:"metadata" json:"metadata,omitempty"`
	CreatedAt      time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at" json:"updated_at"`
}

// NewSyncLog creates a new sync log entry
func NewSyncLog(entityType SyncEntityType, direction SyncDirection) *SyncLog {
	now := time.Now()
	return &SyncLog{
		EntityType:     entityType,
		Direction:      direction,
		Status:         SyncStatusPending,
		StartedAt:      now,
		TotalRecords:   0,
		ProcessedCount: 0,
		SuccessCount:   0,
		ErrorCount:     0,
		ConflictCount:  0,
		Metadata:       make(map[string]any),
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// Start marks the sync as in progress
func (s *SyncLog) Start() {
	s.Status = SyncStatusInProgress
	s.StartedAt = time.Now()
	s.UpdatedAt = time.Now()
}

// Complete marks the sync as completed
func (s *SyncLog) Complete() {
	now := time.Now()
	s.Status = SyncStatusCompleted
	s.CompletedAt = &now
	s.UpdatedAt = now
}

// Fail marks the sync as failed
func (s *SyncLog) Fail(errMsg string) {
	now := time.Now()
	s.Status = SyncStatusFailed
	s.CompletedAt = &now
	s.ErrorMessage = errMsg
	s.UpdatedAt = now
}

// Cancel marks the sync as cancelled
func (s *SyncLog) Cancel() {
	now := time.Now()
	s.Status = SyncStatusCancelled
	s.CompletedAt = &now
	s.UpdatedAt = now
}

// IncrementProcessed increments the processed count
func (s *SyncLog) IncrementProcessed() {
	s.ProcessedCount++
	s.UpdatedAt = time.Now()
}

// IncrementSuccess increments the success count
func (s *SyncLog) IncrementSuccess() {
	s.SuccessCount++
	s.UpdatedAt = time.Now()
}

// IncrementError increments the error count
func (s *SyncLog) IncrementError() {
	s.ErrorCount++
	s.UpdatedAt = time.Now()
}

// IncrementConflict increments the conflict count
func (s *SyncLog) IncrementConflict() {
	s.ConflictCount++
	s.UpdatedAt = time.Now()
}

// SetTotalRecords sets the total number of records to process
func (s *SyncLog) SetTotalRecords(total int) {
	s.TotalRecords = total
	s.UpdatedAt = time.Now()
}

// GetProgress returns the sync progress as percentage
func (s *SyncLog) GetProgress() float64 {
	if s.TotalRecords == 0 {
		return 0
	}
	return float64(s.ProcessedCount) / float64(s.TotalRecords) * 100
}

// IsRunning returns true if sync is currently running
func (s *SyncLog) IsRunning() bool {
	return s.Status == SyncStatusInProgress
}

// SyncLogFilter represents filter options for sync logs
type SyncLogFilter struct {
	EntityType *SyncEntityType `json:"entity_type,omitempty"`
	Direction  *SyncDirection  `json:"direction,omitempty"`
	Status     *SyncStatus     `json:"status,omitempty"`
	StartDate  *time.Time      `json:"start_date,omitempty"`
	EndDate    *time.Time      `json:"end_date,omitempty"`
	Limit      int             `json:"limit,omitempty"`
	Offset     int             `json:"offset,omitempty"`
}

// SyncStats represents synchronization statistics
type SyncStats struct {
	TotalSyncs      int64     `json:"total_syncs"`
	SuccessfulSyncs int64     `json:"successful_syncs"`
	FailedSyncs     int64     `json:"failed_syncs"`
	TotalRecords    int64     `json:"total_records"`
	TotalConflicts  int64     `json:"total_conflicts"`
	LastSyncAt      time.Time `json:"last_sync_at"`
}
