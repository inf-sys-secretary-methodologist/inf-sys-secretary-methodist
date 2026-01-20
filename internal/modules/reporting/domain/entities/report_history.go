package entities

import (
	"encoding/json"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain"
)

// ReportAction represents an action performed on a report
type ReportAction string

const (
	ReportActionCreated   ReportAction = "created"
	ReportActionUpdated   ReportAction = "updated"
	ReportActionSubmitted ReportAction = "submitted"
	ReportActionApproved  ReportAction = "approved"
	ReportActionRejected  ReportAction = "rejected"
	ReportActionPublished ReportAction = "published"
	ReportActionDeleted   ReportAction = "deleted"
)

// ReportHistory represents a history entry for report changes
type ReportHistory struct {
	ID        int64           `json:"id"`
	ReportID  int64           `json:"report_id"`
	UserID    *int64          `json:"user_id,omitempty"`
	Action    ReportAction    `json:"action"`
	Details   json.RawMessage `json:"details,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}

// NewReportHistory creates a new history entry
func NewReportHistory(reportID int64, userID *int64, action ReportAction) *ReportHistory {
	return &ReportHistory{
		ReportID:  reportID,
		UserID:    userID,
		Action:    action,
		CreatedAt: time.Now(),
	}
}

// SetDetails sets additional details for the history entry
func (rh *ReportHistory) SetDetails(details any) error {
	data, err := json.Marshal(details)
	if err != nil {
		return err
	}
	rh.Details = data
	return nil
}

// GetDetails unmarshals details into the provided target
func (rh *ReportHistory) GetDetails(target any) error {
	if rh.Details == nil {
		return nil
	}
	return json.Unmarshal(rh.Details, target)
}

// ReportGenerationLog represents a log entry for report generation
type ReportGenerationLog struct {
	ID               int64                   `json:"id"`
	ReportID         int64                   `json:"report_id"`
	StartedAt        time.Time               `json:"started_at"`
	CompletedAt      *time.Time              `json:"completed_at,omitempty"`
	Status           domain.GenerationStatus `json:"status"`
	ErrorMessage     *string                 `json:"error_message,omitempty"`
	DurationSeconds  *int                    `json:"duration_seconds,omitempty"`
	RecordsProcessed *int                    `json:"records_processed,omitempty"`
}

// NewReportGenerationLog creates a new generation log entry
func NewReportGenerationLog(reportID int64) *ReportGenerationLog {
	return &ReportGenerationLog{
		ReportID:  reportID,
		StartedAt: time.Now(),
		Status:    domain.GenerationStatusStarted,
	}
}

// Complete marks generation as completed
func (rgl *ReportGenerationLog) Complete(recordsProcessed int) {
	now := time.Now()
	rgl.CompletedAt = &now
	rgl.Status = domain.GenerationStatusCompleted
	rgl.RecordsProcessed = &recordsProcessed
	duration := int(now.Sub(rgl.StartedAt).Seconds())
	rgl.DurationSeconds = &duration
}

// Fail marks generation as failed
func (rgl *ReportGenerationLog) Fail(errorMessage string) {
	now := time.Now()
	rgl.CompletedAt = &now
	rgl.Status = domain.GenerationStatusFailed
	rgl.ErrorMessage = &errorMessage
	duration := int(now.Sub(rgl.StartedAt).Seconds())
	rgl.DurationSeconds = &duration
}

// IsCompleted checks if generation is completed
func (rgl *ReportGenerationLog) IsCompleted() bool {
	return rgl.Status == domain.GenerationStatusCompleted
}

// IsFailed checks if generation failed
func (rgl *ReportGenerationLog) IsFailed() bool {
	return rgl.Status == domain.GenerationStatusFailed
}

// IsInProgress checks if generation is still in progress
func (rgl *ReportGenerationLog) IsInProgress() bool {
	return rgl.Status == domain.GenerationStatusStarted
}
