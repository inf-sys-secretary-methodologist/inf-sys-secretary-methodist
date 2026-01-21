package entities

import (
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain"
)

func TestNewReport(t *testing.T) {
	reportTypeID := int64(1)
	title := "Monthly Report"
	authorID := int64(42)

	report := NewReport(reportTypeID, title, authorID)

	if report.ReportTypeID != reportTypeID {
		t.Errorf("expected report type ID %d, got %d", reportTypeID, report.ReportTypeID)
	}
	if report.Title != title {
		t.Errorf("expected title %q, got %q", title, report.Title)
	}
	if report.AuthorID != authorID {
		t.Errorf("expected author ID %d, got %d", authorID, report.AuthorID)
	}
	if report.Status != domain.ReportStatusDraft {
		t.Errorf("expected status %q, got %q", domain.ReportStatusDraft, report.Status)
	}
	if report.IsPublic {
		t.Error("expected IsPublic to be false")
	}
	if report.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestReport_SetPeriod(t *testing.T) {
	report := NewReport(1, "Report", 1)
	start := time.Now()
	end := start.Add(30 * 24 * time.Hour)

	report.SetPeriod(start, end)

	if report.PeriodStart == nil || !report.PeriodStart.Equal(start) {
		t.Errorf("expected period start %v, got %v", start, report.PeriodStart)
	}
	if report.PeriodEnd == nil || !report.PeriodEnd.Equal(end) {
		t.Errorf("expected period end %v, got %v", end, report.PeriodEnd)
	}
}

func TestReport_SetParameters(t *testing.T) {
	report := NewReport(1, "Report", 1)
	params := map[string]string{"key": "value"}

	err := report.SetParameters(params)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if report.Parameters == nil {
		t.Error("expected Parameters to be set")
	}
}

func TestReport_GetParameters(t *testing.T) {
	report := NewReport(1, "Report", 1)
	params := map[string]string{"key": "value"}
	report.SetParameters(params)

	var result map[string]string
	err := report.GetParameters(&result)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result["key"] != "value" {
		t.Errorf("expected key=value, got key=%q", result["key"])
	}
}

func TestReport_GetParameters_Nil(t *testing.T) {
	report := NewReport(1, "Report", 1)

	var result map[string]string
	err := report.GetParameters(&result)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestReport_SetData(t *testing.T) {
	report := NewReport(1, "Report", 1)
	data := map[string]int{"count": 42}

	err := report.SetData(data)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if report.Data == nil {
		t.Error("expected Data to be set")
	}
}

func TestReport_GetData(t *testing.T) {
	report := NewReport(1, "Report", 1)
	data := map[string]int{"count": 42}
	report.SetData(data)

	var result map[string]int
	err := report.GetData(&result)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result["count"] != 42 {
		t.Errorf("expected count=42, got count=%d", result["count"])
	}
}

func TestReport_StartGeneration(t *testing.T) {
	report := NewReport(1, "Report", 1)

	err := report.StartGeneration()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if report.Status != domain.ReportStatusGenerating {
		t.Errorf("expected status %q, got %q", domain.ReportStatusGenerating, report.Status)
	}
}

func TestReport_StartGeneration_InvalidStatus(t *testing.T) {
	report := NewReport(1, "Report", 1)
	report.Status = domain.ReportStatusReady

	err := report.StartGeneration()

	if err != ErrInvalidStatusTransition {
		t.Errorf("expected error %v, got %v", ErrInvalidStatusTransition, err)
	}
}

func TestReport_CompleteGeneration(t *testing.T) {
	report := NewReport(1, "Report", 1)
	report.StartGeneration()

	err := report.CompleteGeneration("report.pdf", "/files/report.pdf", 1024, "application/pdf")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if report.Status != domain.ReportStatusReady {
		t.Errorf("expected status %q, got %q", domain.ReportStatusReady, report.Status)
	}
	if report.FileName == nil || *report.FileName != "report.pdf" {
		t.Errorf("expected file name %q, got %v", "report.pdf", report.FileName)
	}
}

func TestReport_CompleteGeneration_InvalidStatus(t *testing.T) {
	report := NewReport(1, "Report", 1)

	err := report.CompleteGeneration("report.pdf", "/files/report.pdf", 1024, "application/pdf")

	if err != ErrInvalidStatusTransition {
		t.Errorf("expected error %v, got %v", ErrInvalidStatusTransition, err)
	}
}

func TestReport_FailGeneration(t *testing.T) {
	report := NewReport(1, "Report", 1)
	report.StartGeneration()

	err := report.FailGeneration()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if report.Status != domain.ReportStatusDraft {
		t.Errorf("expected status %q, got %q", domain.ReportStatusDraft, report.Status)
	}
}

func TestReport_SubmitForReview(t *testing.T) {
	report := NewReport(1, "Report", 1)
	report.Status = domain.ReportStatusReady

	err := report.SubmitForReview()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if report.Status != domain.ReportStatusReviewing {
		t.Errorf("expected status %q, got %q", domain.ReportStatusReviewing, report.Status)
	}
}

func TestReport_SubmitForReview_NotReady(t *testing.T) {
	report := NewReport(1, "Report", 1)

	err := report.SubmitForReview()

	if err != ErrReportNotReady {
		t.Errorf("expected error %v, got %v", ErrReportNotReady, err)
	}
}

func TestReport_Approve(t *testing.T) {
	report := NewReport(1, "Report", 1)
	report.Status = domain.ReportStatusReviewing
	reviewerID := int64(99)
	comment := "Approved"

	err := report.Approve(reviewerID, comment)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if report.Status != domain.ReportStatusApproved {
		t.Errorf("expected status %q, got %q", domain.ReportStatusApproved, report.Status)
	}
	if report.ReviewedBy == nil || *report.ReviewedBy != reviewerID {
		t.Errorf("expected reviewed by %d, got %v", reviewerID, report.ReviewedBy)
	}
	if report.ReviewerComment == nil || *report.ReviewerComment != comment {
		t.Errorf("expected comment %q, got %v", comment, report.ReviewerComment)
	}
}

func TestReport_Reject(t *testing.T) {
	report := NewReport(1, "Report", 1)
	report.Status = domain.ReportStatusReviewing
	reviewerID := int64(99)
	comment := "Needs revision"

	err := report.Reject(reviewerID, comment)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if report.Status != domain.ReportStatusRejected {
		t.Errorf("expected status %q, got %q", domain.ReportStatusRejected, report.Status)
	}
}

func TestReport_Publish(t *testing.T) {
	report := NewReport(1, "Report", 1)
	report.Status = domain.ReportStatusApproved

	err := report.Publish(true)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if report.Status != domain.ReportStatusPublished {
		t.Errorf("expected status %q, got %q", domain.ReportStatusPublished, report.Status)
	}
	if !report.IsPublic {
		t.Error("expected IsPublic to be true")
	}
	if report.PublishedAt == nil {
		t.Error("expected PublishedAt to be set")
	}
}

func TestReport_ReturnToDraft(t *testing.T) {
	report := NewReport(1, "Report", 1)
	report.Status = domain.ReportStatusRejected

	err := report.ReturnToDraft()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if report.Status != domain.ReportStatusDraft {
		t.Errorf("expected status %q, got %q", domain.ReportStatusDraft, report.Status)
	}
}

func TestReport_CanEdit(t *testing.T) {
	tests := []struct {
		name   string
		status domain.ReportStatus
		want   bool
	}{
		{"draft can edit", domain.ReportStatusDraft, true},
		{"rejected can edit", domain.ReportStatusRejected, true},
		{"ready cannot edit", domain.ReportStatusReady, false},
		{"published cannot edit", domain.ReportStatusPublished, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := NewReport(1, "Report", 1)
			report.Status = tt.status

			got := report.CanEdit()
			if got != tt.want {
				t.Errorf("CanEdit() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReport_IsFinalized(t *testing.T) {
	tests := []struct {
		name   string
		status domain.ReportStatus
		want   bool
	}{
		{"published is finalized", domain.ReportStatusPublished, true},
		{"approved is not finalized", domain.ReportStatusApproved, false},
		{"draft is not finalized", domain.ReportStatusDraft, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := NewReport(1, "Report", 1)
			report.Status = tt.status

			got := report.IsFinalized()
			if got != tt.want {
				t.Errorf("IsFinalized() = %v, want %v", got, tt.want)
			}
		})
	}
}
