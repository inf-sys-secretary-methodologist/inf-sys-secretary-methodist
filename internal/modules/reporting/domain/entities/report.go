package entities

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain"
)

var (
	// ErrInvalidStatusTransition is returned when attempting an invalid status change
	ErrInvalidStatusTransition = errors.New("invalid status transition")
	// ErrReportNotReady is returned when trying to approve/publish a report that's not ready
	ErrReportNotReady = errors.New("report is not ready for this action")
	// ErrReportAlreadyPublished is returned when trying to modify a published report
	ErrReportAlreadyPublished = errors.New("cannot modify published report")
)

// Report represents a generated report
// Aligned with migrations/006_create_reports_schema.up.sql - reports table
type Report struct {
	ID           int64               `db:"id" json:"id"`
	ReportTypeID int64               `db:"report_type_id" json:"report_type_id"`
	Title        string              `db:"title" json:"title"`
	Description  *string             `db:"description" json:"description,omitempty"`
	PeriodStart  *time.Time          `db:"period_start" json:"period_start,omitempty"`
	PeriodEnd    *time.Time          `db:"period_end" json:"period_end,omitempty"`
	AuthorID     int64               `db:"author_id" json:"author_id"`
	Status       domain.ReportStatus `db:"status" json:"status"`

	// File information
	FileName *string `db:"file_name" json:"file_name,omitempty"`
	FilePath *string `db:"file_path" json:"file_path,omitempty"`
	FileSize *int64  `db:"file_size" json:"file_size,omitempty"`
	MimeType *string `db:"mime_type" json:"mime_type,omitempty"`

	// Report data
	Parameters json.RawMessage `db:"parameters" json:"parameters,omitempty"`
	Data       json.RawMessage `db:"data" json:"data,omitempty"`

	// Review information
	ReviewerComment *string    `db:"reviewer_comment" json:"reviewer_comment,omitempty"`
	ReviewedBy      *int64     `db:"reviewed_by" json:"reviewed_by,omitempty"`
	ReviewedAt      *time.Time `db:"reviewed_at" json:"reviewed_at,omitempty"`

	// Publication
	PublishedAt *time.Time `db:"published_at" json:"published_at,omitempty"`
	IsPublic    bool       `db:"is_public" json:"is_public"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`

	// Associations (not stored in DB, loaded separately)
	ReportType *ReportType     `db:"-" json:"report_type,omitempty"`
	Access     []ReportAccess  `db:"-" json:"access,omitempty"`
	Comments   []ReportComment `db:"-" json:"comments,omitempty"`
}

// NewReport creates a new report in draft status
func NewReport(reportTypeID int64, title string, authorID int64) *Report {
	now := time.Now()
	return &Report{
		ReportTypeID: reportTypeID,
		Title:        title,
		AuthorID:     authorID,
		Status:       domain.ReportStatusDraft,
		IsPublic:     false,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// SetPeriod sets the reporting period
func (r *Report) SetPeriod(start, end time.Time) {
	r.PeriodStart = &start
	r.PeriodEnd = &end
	r.UpdatedAt = time.Now()
}

// SetParameters sets the report parameters
func (r *Report) SetParameters(params any) error {
	data, err := json.Marshal(params)
	if err != nil {
		return err
	}
	r.Parameters = data
	r.UpdatedAt = time.Now()
	return nil
}

// GetParameters unmarshals parameters into the provided target
func (r *Report) GetParameters(target any) error {
	if r.Parameters == nil {
		return nil
	}
	return json.Unmarshal(r.Parameters, target)
}

// SetData sets the report data
func (r *Report) SetData(data any) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	r.Data = jsonData
	r.UpdatedAt = time.Now()
	return nil
}

// GetData unmarshals data into the provided target
func (r *Report) GetData(target any) error {
	if r.Data == nil {
		return nil
	}
	return json.Unmarshal(r.Data, target)
}

// StartGeneration transitions report to generating status
func (r *Report) StartGeneration() error {
	if r.Status != domain.ReportStatusDraft {
		return ErrInvalidStatusTransition
	}
	r.Status = domain.ReportStatusGenerating
	r.UpdatedAt = time.Now()
	return nil
}

// CompleteGeneration marks report generation as complete
func (r *Report) CompleteGeneration(fileName, filePath string, fileSize int64, mimeType string) error {
	if r.Status != domain.ReportStatusGenerating {
		return ErrInvalidStatusTransition
	}
	r.Status = domain.ReportStatusReady
	r.FileName = &fileName
	r.FilePath = &filePath
	r.FileSize = &fileSize
	r.MimeType = &mimeType
	r.UpdatedAt = time.Now()
	return nil
}

// FailGeneration marks report generation as failed, returns to draft
func (r *Report) FailGeneration() error {
	if r.Status != domain.ReportStatusGenerating {
		return ErrInvalidStatusTransition
	}
	r.Status = domain.ReportStatusDraft
	r.UpdatedAt = time.Now()
	return nil
}

// SubmitForReview submits the report for review
func (r *Report) SubmitForReview() error {
	if r.Status != domain.ReportStatusReady {
		return ErrReportNotReady
	}
	r.Status = domain.ReportStatusReviewing
	r.UpdatedAt = time.Now()
	return nil
}

// Approve approves the report
func (r *Report) Approve(reviewerID int64, comment string) error {
	if r.Status != domain.ReportStatusReviewing {
		return ErrInvalidStatusTransition
	}
	now := time.Now()
	r.Status = domain.ReportStatusApproved
	r.ReviewedBy = &reviewerID
	r.ReviewedAt = &now
	if comment != "" {
		r.ReviewerComment = &comment
	}
	r.UpdatedAt = now
	return nil
}

// Reject rejects the report
func (r *Report) Reject(reviewerID int64, comment string) error {
	if r.Status != domain.ReportStatusReviewing {
		return ErrInvalidStatusTransition
	}
	now := time.Now()
	r.Status = domain.ReportStatusRejected
	r.ReviewedBy = &reviewerID
	r.ReviewedAt = &now
	r.ReviewerComment = &comment
	r.UpdatedAt = now
	return nil
}

// Publish publishes the report
func (r *Report) Publish(isPublic bool) error {
	if r.Status != domain.ReportStatusApproved {
		return ErrInvalidStatusTransition
	}
	now := time.Now()
	r.Status = domain.ReportStatusPublished
	r.PublishedAt = &now
	r.IsPublic = isPublic
	r.UpdatedAt = now
	return nil
}

// ReturnToDraft returns rejected report to draft for editing
func (r *Report) ReturnToDraft() error {
	if r.Status != domain.ReportStatusRejected {
		return ErrInvalidStatusTransition
	}
	r.Status = domain.ReportStatusDraft
	r.UpdatedAt = time.Now()
	return nil
}

// CanEdit checks if report can be edited
func (r *Report) CanEdit() bool {
	return r.Status == domain.ReportStatusDraft || r.Status == domain.ReportStatusRejected
}

// IsFinalized checks if report is in a final state
func (r *Report) IsFinalized() bool {
	return r.Status == domain.ReportStatusPublished
}
