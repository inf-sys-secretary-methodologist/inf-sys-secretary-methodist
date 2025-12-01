// Package dto contains Data Transfer Objects for the reporting module.
package dto

import (
	"encoding/json"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain/entities"
)

// CreateReportInput represents input for creating a new report
type CreateReportInput struct {
	ReportTypeID int64                  `json:"report_type_id" validate:"required"`
	Title        string                 `json:"title" validate:"required,min=1,max=500"`
	Description  *string                `json:"description,omitempty"`
	PeriodStart  *time.Time             `json:"period_start,omitempty"`
	PeriodEnd    *time.Time             `json:"period_end,omitempty"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
	IsPublic     bool                   `json:"is_public"`
}

// UpdateReportInput represents input for updating a report
type UpdateReportInput struct {
	Title       *string                `json:"title,omitempty" validate:"omitempty,min=1,max=500"`
	Description *string                `json:"description,omitempty"`
	PeriodStart *time.Time             `json:"period_start,omitempty"`
	PeriodEnd   *time.Time             `json:"period_end,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	IsPublic    *bool                  `json:"is_public,omitempty"`
}

// GenerateReportInput represents input for generating a report
type GenerateReportInput struct {
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// ReviewReportInput represents input for reviewing a report
type ReviewReportInput struct {
	Action  string `json:"action" validate:"required,oneof=approve reject"`
	Comment string `json:"comment,omitempty"`
}

// PublishReportInput represents input for publishing a report
type PublishReportInput struct {
	IsPublic bool `json:"is_public"`
}

// AddAccessInput represents input for adding report access
type AddAccessInput struct {
	UserID     *int64 `json:"user_id,omitempty"`
	Role       string `json:"role,omitempty" validate:"omitempty,oneof=admin secretary methodist teacher student"`
	Permission string `json:"permission" validate:"required,oneof=read write approve publish"`
}

// AddCommentInput represents input for adding a comment
type AddCommentInput struct {
	Content string `json:"content" validate:"required,min=1"`
}

// UpdateCommentInput represents input for updating a comment
type UpdateCommentInput struct {
	Content string `json:"content" validate:"required,min=1"`
}

// ReportOutput represents output for a single report
type ReportOutput struct {
	ID              int64                  `json:"id"`
	ReportTypeID    int64                  `json:"report_type_id"`
	ReportTypeName  string                 `json:"report_type_name,omitempty"`
	ReportTypeCode  string                 `json:"report_type_code,omitempty"`
	Title           string                 `json:"title"`
	Description     *string                `json:"description,omitempty"`
	PeriodStart     *time.Time             `json:"period_start,omitempty"`
	PeriodEnd       *time.Time             `json:"period_end,omitempty"`
	AuthorID        int64                  `json:"author_id"`
	AuthorName      string                 `json:"author_name,omitempty"`
	Status          string                 `json:"status"`
	FileName        *string                `json:"file_name,omitempty"`
	FileSize        *int64                 `json:"file_size,omitempty"`
	MimeType        *string                `json:"mime_type,omitempty"`
	HasFile         bool                   `json:"has_file"`
	Parameters      map[string]interface{} `json:"parameters,omitempty"`
	ReviewerID      *int64                 `json:"reviewer_id,omitempty"`
	ReviewerName    *string                `json:"reviewer_name,omitempty"`
	ReviewerComment *string                `json:"reviewer_comment,omitempty"`
	ReviewedAt      *time.Time             `json:"reviewed_at,omitempty"`
	PublishedAt     *time.Time             `json:"published_at,omitempty"`
	IsPublic        bool                   `json:"is_public"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// ReportListOutput represents paginated list of reports
type ReportListOutput struct {
	Reports    []*ReportOutput `json:"reports"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
}

// ReportFilterInput represents filter options for listing reports
type ReportFilterInput struct {
	ReportTypeID *int64  `form:"report_type_id"`
	AuthorID     *int64  `form:"author_id"`
	Status       *string `form:"status" validate:"omitempty,oneof=draft generating ready reviewing approved rejected published"`
	IsPublic     *bool   `form:"is_public"`
	PeriodStart  *string `form:"period_start"`
	PeriodEnd    *string `form:"period_end"`
	Search       *string `form:"search"`
	Page         int     `form:"page,default=1"`
	PageSize     int     `form:"page_size,default=20"`
}

// ReportAccessOutput represents output for report access
type ReportAccessOutput struct {
	ID         int64     `json:"id"`
	ReportID   int64     `json:"report_id"`
	UserID     *int64    `json:"user_id,omitempty"`
	Role       *string   `json:"role,omitempty"`
	Permission string    `json:"permission"`
	GrantedBy  *int64    `json:"granted_by,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// ReportCommentOutput represents output for a report comment
type ReportCommentOutput struct {
	ID         int64     `json:"id"`
	ReportID   int64     `json:"report_id"`
	AuthorID   int64     `json:"author_id"`
	AuthorName string    `json:"author_name,omitempty"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ReportHistoryOutput represents output for a report history entry
type ReportHistoryOutput struct {
	ID        int64                  `json:"id"`
	ReportID  int64                  `json:"report_id"`
	UserID    *int64                 `json:"user_id,omitempty"`
	UserName  *string                `json:"user_name,omitempty"`
	Action    string                 `json:"action"`
	Details   map[string]interface{} `json:"details,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

// ReportGenerationLogOutput represents output for a generation log entry
type ReportGenerationLogOutput struct {
	ID               int64      `json:"id"`
	ReportID         int64      `json:"report_id"`
	StartedAt        time.Time  `json:"started_at"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
	Status           string     `json:"status"`
	ErrorMessage     *string    `json:"error_message,omitempty"`
	DurationSeconds  *int       `json:"duration_seconds,omitempty"`
	RecordsProcessed *int       `json:"records_processed,omitempty"`
}

// ToReportOutput converts a Report entity to ReportOutput DTO
func ToReportOutput(report *entities.Report) *ReportOutput {
	output := &ReportOutput{
		ID:              report.ID,
		ReportTypeID:    report.ReportTypeID,
		Title:           report.Title,
		Description:     report.Description,
		PeriodStart:     report.PeriodStart,
		PeriodEnd:       report.PeriodEnd,
		AuthorID:        report.AuthorID,
		Status:          string(report.Status),
		FileName:        report.FileName,
		FileSize:        report.FileSize,
		MimeType:        report.MimeType,
		HasFile:         report.FileName != nil && *report.FileName != "",
		ReviewerID:      report.ReviewedBy,
		ReviewerComment: report.ReviewerComment,
		ReviewedAt:      report.ReviewedAt,
		PublishedAt:     report.PublishedAt,
		IsPublic:        report.IsPublic,
		CreatedAt:       report.CreatedAt,
		UpdatedAt:       report.UpdatedAt,
	}

	// Parse parameters if present
	if report.Parameters != nil {
		var params map[string]interface{}
		if err := json.Unmarshal(report.Parameters, &params); err == nil {
			output.Parameters = params
		}
	}

	// Include report type info if loaded
	if report.ReportType != nil {
		output.ReportTypeName = report.ReportType.Name
		output.ReportTypeCode = report.ReportType.Code
	}

	return output
}

// ToReportAccessOutput converts a ReportAccess entity to DTO
func ToReportAccessOutput(access *entities.ReportAccess) *ReportAccessOutput {
	output := &ReportAccessOutput{
		ID:         access.ID,
		ReportID:   access.ReportID,
		UserID:     access.UserID,
		Permission: string(access.Permission),
		GrantedBy:  access.GrantedBy,
		CreatedAt:  access.CreatedAt,
	}
	if access.Role != nil {
		role := string(*access.Role)
		output.Role = &role
	}
	return output
}

// ToReportCommentOutput converts a ReportComment entity to DTO
func ToReportCommentOutput(comment *entities.ReportComment) *ReportCommentOutput {
	return &ReportCommentOutput{
		ID:        comment.ID,
		ReportID:  comment.ReportID,
		AuthorID:  comment.AuthorID,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt,
		UpdatedAt: comment.UpdatedAt,
	}
}

// ToReportHistoryOutput converts a ReportHistory entity to DTO
func ToReportHistoryOutput(history *entities.ReportHistory) *ReportHistoryOutput {
	output := &ReportHistoryOutput{
		ID:        history.ID,
		ReportID:  history.ReportID,
		UserID:    history.UserID,
		Action:    string(history.Action),
		CreatedAt: history.CreatedAt,
	}

	if history.Details != nil {
		var details map[string]interface{}
		if err := json.Unmarshal(history.Details, &details); err == nil {
			output.Details = details
		}
	}

	return output
}

// ToReportGenerationLogOutput converts a ReportGenerationLog entity to DTO
func ToReportGenerationLogOutput(log *entities.ReportGenerationLog) *ReportGenerationLogOutput {
	return &ReportGenerationLogOutput{
		ID:               log.ID,
		ReportID:         log.ReportID,
		StartedAt:        log.StartedAt,
		CompletedAt:      log.CompletedAt,
		Status:           string(log.Status),
		ErrorMessage:     log.ErrorMessage,
		DurationSeconds:  log.DurationSeconds,
		RecordsProcessed: log.RecordsProcessed,
	}
}

// ToReportFilter converts filter input to repository filter
func ToReportFilter(input *ReportFilterInput) (*domain.ReportStatus, error) {
	if input.Status == nil {
		return nil, nil
	}
	status := domain.ReportStatus(*input.Status)
	if !status.IsValid() {
		return nil, nil
	}
	return &status, nil
}
