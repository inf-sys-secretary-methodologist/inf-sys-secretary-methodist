package entities

import (
	"encoding/json"
	"errors"
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

func TestReport_SetParameters_InvalidJSON(t *testing.T) {
	report := NewReport(1, "Report", 1)
	// channels are not marshallable
	err := report.SetParameters(make(chan int))
	if err == nil {
		t.Error("expected error for unmarshallable value")
	}
}

func TestReport_GetParameters(t *testing.T) {
	report := NewReport(1, "Report", 1)
	params := map[string]string{"key": "value"}
	_ = report.SetParameters(params)

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

func TestReport_SetData_InvalidJSON(t *testing.T) {
	report := NewReport(1, "Report", 1)
	err := report.SetData(make(chan int))
	if err == nil {
		t.Error("expected error for unmarshallable value")
	}
}

func TestReport_GetData(t *testing.T) {
	report := NewReport(1, "Report", 1)
	data := map[string]int{"count": 42}
	_ = report.SetData(data)

	var result map[string]int
	err := report.GetData(&result)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result["count"] != 42 {
		t.Errorf("expected count=42, got count=%d", result["count"])
	}
}

func TestReport_GetData_Nil(t *testing.T) {
	report := NewReport(1, "Report", 1)

	var result map[string]int
	err := report.GetData(&result)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
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

	if !errors.Is(err, ErrInvalidStatusTransition) {
		t.Errorf("expected error %v, got %v", ErrInvalidStatusTransition, err)
	}
}

func TestReport_CompleteGeneration(t *testing.T) {
	report := NewReport(1, "Report", 1)
	_ = report.StartGeneration()

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
	if report.FilePath == nil || *report.FilePath != "/files/report.pdf" {
		t.Errorf("expected file path, got %v", report.FilePath)
	}
	if report.FileSize == nil || *report.FileSize != 1024 {
		t.Errorf("expected file size 1024, got %v", report.FileSize)
	}
	if report.MimeType == nil || *report.MimeType != "application/pdf" {
		t.Errorf("expected mime type, got %v", report.MimeType)
	}
}

func TestReport_CompleteGeneration_InvalidStatus(t *testing.T) {
	report := NewReport(1, "Report", 1)

	err := report.CompleteGeneration("report.pdf", "/files/report.pdf", 1024, "application/pdf")

	if !errors.Is(err, ErrInvalidStatusTransition) {
		t.Errorf("expected error %v, got %v", ErrInvalidStatusTransition, err)
	}
}

func TestReport_FailGeneration(t *testing.T) {
	report := NewReport(1, "Report", 1)
	_ = report.StartGeneration()

	err := report.FailGeneration()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if report.Status != domain.ReportStatusDraft {
		t.Errorf("expected status %q, got %q", domain.ReportStatusDraft, report.Status)
	}
}

func TestReport_FailGeneration_InvalidStatus(t *testing.T) {
	report := NewReport(1, "Report", 1)

	err := report.FailGeneration()
	if !errors.Is(err, ErrInvalidStatusTransition) {
		t.Errorf("expected error %v, got %v", ErrInvalidStatusTransition, err)
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

	if !errors.Is(err, ErrReportNotReady) {
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
	if report.ReviewedAt == nil {
		t.Error("expected ReviewedAt to be set")
	}
}

func TestReport_Approve_EmptyComment(t *testing.T) {
	report := NewReport(1, "Report", 1)
	report.Status = domain.ReportStatusReviewing

	err := report.Approve(99, "")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if report.ReviewerComment != nil {
		t.Error("expected ReviewerComment to be nil for empty comment")
	}
}

func TestReport_Approve_InvalidStatus(t *testing.T) {
	report := NewReport(1, "Report", 1)

	err := report.Approve(99, "OK")
	if !errors.Is(err, ErrInvalidStatusTransition) {
		t.Errorf("expected error %v, got %v", ErrInvalidStatusTransition, err)
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
	if report.ReviewedBy == nil || *report.ReviewedBy != reviewerID {
		t.Errorf("expected reviewed by %d, got %v", reviewerID, report.ReviewedBy)
	}
	if report.ReviewerComment == nil || *report.ReviewerComment != comment {
		t.Errorf("expected comment %q, got %v", comment, report.ReviewerComment)
	}
}

func TestReport_Reject_InvalidStatus(t *testing.T) {
	report := NewReport(1, "Report", 1)

	err := report.Reject(99, "Bad")
	if !errors.Is(err, ErrInvalidStatusTransition) {
		t.Errorf("expected error %v, got %v", ErrInvalidStatusTransition, err)
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

func TestReport_Publish_NotPublic(t *testing.T) {
	report := NewReport(1, "Report", 1)
	report.Status = domain.ReportStatusApproved

	err := report.Publish(false)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if report.IsPublic {
		t.Error("expected IsPublic to be false")
	}
}

func TestReport_Publish_InvalidStatus(t *testing.T) {
	report := NewReport(1, "Report", 1)

	err := report.Publish(true)
	if !errors.Is(err, ErrInvalidStatusTransition) {
		t.Errorf("expected error %v, got %v", ErrInvalidStatusTransition, err)
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

func TestReport_ReturnToDraft_InvalidStatus(t *testing.T) {
	report := NewReport(1, "Report", 1)

	err := report.ReturnToDraft()
	if !errors.Is(err, ErrInvalidStatusTransition) {
		t.Errorf("expected error %v, got %v", ErrInvalidStatusTransition, err)
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
		{"generating cannot edit", domain.ReportStatusGenerating, false},
		{"reviewing cannot edit", domain.ReportStatusReviewing, false},
		{"approved cannot edit", domain.ReportStatusApproved, false},
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

// --- ReportType tests ---

func TestNewReportType(t *testing.T) {
	rt := NewReportType("Monthly", "monthly_report", domain.OutputFormatPDF)

	if rt.Name != "Monthly" {
		t.Errorf("expected name 'Monthly', got %q", rt.Name)
	}
	if rt.Code != "monthly_report" {
		t.Errorf("expected code 'monthly_report', got %q", rt.Code)
	}
	if rt.OutputFormat != domain.OutputFormatPDF {
		t.Errorf("expected format %q, got %q", domain.OutputFormatPDF, rt.OutputFormat)
	}
	if rt.IsPeriodic {
		t.Error("expected IsPeriodic to be false")
	}
	if rt.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestReportType_SetPeriodic(t *testing.T) {
	rt := NewReportType("Monthly", "monthly", domain.OutputFormatPDF)
	rt.SetPeriodic(domain.PeriodTypeMonthly)

	if !rt.IsPeriodic {
		t.Error("expected IsPeriodic to be true")
	}
	if rt.PeriodType == nil || *rt.PeriodType != domain.PeriodTypeMonthly {
		t.Errorf("expected period type %q, got %v", domain.PeriodTypeMonthly, rt.PeriodType)
	}
}

func TestReportType_SetCategory(t *testing.T) {
	rt := NewReportType("Report", "report", domain.OutputFormatPDF)
	rt.SetCategory(domain.ReportCategoryAcademic)

	if rt.Category == nil || *rt.Category != domain.ReportCategoryAcademic {
		t.Errorf("expected category %q, got %v", domain.ReportCategoryAcademic, rt.Category)
	}
}

// --- ReportParameter tests ---

func TestNewReportParameter(t *testing.T) {
	rp := NewReportParameter(1, "start_date", domain.ParameterTypeDate, true)

	if rp.ReportTypeID != 1 {
		t.Errorf("expected report type ID 1, got %d", rp.ReportTypeID)
	}
	if rp.ParameterName != "start_date" {
		t.Errorf("expected name 'start_date', got %q", rp.ParameterName)
	}
	if rp.ParameterType != domain.ParameterTypeDate {
		t.Errorf("expected type %q, got %q", domain.ParameterTypeDate, rp.ParameterType)
	}
	if !rp.IsRequired {
		t.Error("expected IsRequired to be true")
	}
	if rp.DisplayOrder != 0 {
		t.Errorf("expected display order 0, got %d", rp.DisplayOrder)
	}
}

func TestReportParameter_SetOptions(t *testing.T) {
	rp := NewReportParameter(1, "status", domain.ParameterTypeSelect, false)
	options := []string{"active", "inactive"}

	err := rp.SetOptions(options)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if rp.Options == nil {
		t.Error("expected Options to be set")
	}
}

func TestReportParameter_SetOptions_Invalid(t *testing.T) {
	rp := NewReportParameter(1, "status", domain.ParameterTypeSelect, false)
	err := rp.SetOptions(make(chan int))
	if err == nil {
		t.Error("expected error for unmarshallable value")
	}
}

func TestReportParameter_GetOptions(t *testing.T) {
	rp := NewReportParameter(1, "status", domain.ParameterTypeSelect, false)
	options := []string{"active", "inactive"}
	_ = rp.SetOptions(options)

	var result []string
	err := rp.GetOptions(&result)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 options, got %d", len(result))
	}
}

func TestReportParameter_GetOptions_Nil(t *testing.T) {
	rp := NewReportParameter(1, "status", domain.ParameterTypeSelect, false)

	var result []string
	err := rp.GetOptions(&result)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// --- ReportTemplate tests ---

func TestNewReportTemplate(t *testing.T) {
	rt := NewReportTemplate(1, "Default Template", "<html>content</html>", 42)

	if rt.ReportTypeID != 1 {
		t.Errorf("expected report type ID 1, got %d", rt.ReportTypeID)
	}
	if rt.Name != "Default Template" {
		t.Errorf("expected name 'Default Template', got %q", rt.Name)
	}
	if rt.Content != "<html>content</html>" {
		t.Errorf("expected content, got %q", rt.Content)
	}
	if rt.IsDefault {
		t.Error("expected IsDefault to be false")
	}
	if rt.CreatedBy != 42 {
		t.Errorf("expected created by 42, got %d", rt.CreatedBy)
	}
}

func TestReportTemplate_SetAsDefault(t *testing.T) {
	rt := NewReportTemplate(1, "Template", "content", 1)
	rt.SetAsDefault()

	if !rt.IsDefault {
		t.Error("expected IsDefault to be true")
	}
}

// --- ReportAccess tests ---

func TestNewReportAccessForUser(t *testing.T) {
	grantedBy := int64(99)
	ra := NewReportAccessForUser(1, 42, domain.ReportPermissionRead, &grantedBy)

	if ra.ReportID != 1 {
		t.Errorf("expected report ID 1, got %d", ra.ReportID)
	}
	if ra.UserID == nil || *ra.UserID != 42 {
		t.Errorf("expected user ID 42, got %v", ra.UserID)
	}
	if ra.Permission != domain.ReportPermissionRead {
		t.Errorf("expected permission %q, got %q", domain.ReportPermissionRead, ra.Permission)
	}
	if ra.GrantedBy == nil || *ra.GrantedBy != 99 {
		t.Errorf("expected granted by 99, got %v", ra.GrantedBy)
	}
	if !ra.IsForUser() {
		t.Error("expected IsForUser to return true")
	}
	if ra.IsForRole() {
		t.Error("expected IsForRole to return false")
	}
}

func TestNewReportAccessForRole(t *testing.T) {
	ra := NewReportAccessForRole(1, domain.AccessRoleAdmin, domain.ReportPermissionWrite, nil)

	if ra.ReportID != 1 {
		t.Errorf("expected report ID 1, got %d", ra.ReportID)
	}
	if ra.Role == nil || *ra.Role != domain.AccessRoleAdmin {
		t.Errorf("expected role 'admin', got %v", ra.Role)
	}
	if ra.Permission != domain.ReportPermissionWrite {
		t.Errorf("expected permission %q, got %q", domain.ReportPermissionWrite, ra.Permission)
	}
	if ra.GrantedBy != nil {
		t.Error("expected GrantedBy to be nil")
	}
	if ra.IsForUser() {
		t.Error("expected IsForUser to return false")
	}
	if !ra.IsForRole() {
		t.Error("expected IsForRole to return true")
	}
}

// --- ReportComment tests ---

func TestNewReportComment(t *testing.T) {
	rc := NewReportComment(1, 42, "Good report")

	if rc.ReportID != 1 {
		t.Errorf("expected report ID 1, got %d", rc.ReportID)
	}
	if rc.AuthorID != 42 {
		t.Errorf("expected author ID 42, got %d", rc.AuthorID)
	}
	if rc.Content != "Good report" {
		t.Errorf("expected content 'Good report', got %q", rc.Content)
	}
	if rc.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestReportComment_Update(t *testing.T) {
	rc := NewReportComment(1, 1, "Old")
	rc.Update("New")

	if rc.Content != "New" {
		t.Errorf("expected content 'New', got %q", rc.Content)
	}
}

// --- ReportSubscription tests ---

func TestNewReportSubscription(t *testing.T) {
	rs := NewReportSubscription(1, 42, domain.DeliveryMethodEmail)

	if rs.ReportTypeID != 1 {
		t.Errorf("expected report type ID 1, got %d", rs.ReportTypeID)
	}
	if rs.UserID != 42 {
		t.Errorf("expected user ID 42, got %d", rs.UserID)
	}
	if rs.DeliveryMethod != domain.DeliveryMethodEmail {
		t.Errorf("expected delivery method %q, got %q", domain.DeliveryMethodEmail, rs.DeliveryMethod)
	}
	if !rs.IsActive {
		t.Error("expected IsActive to be true")
	}
}

func TestReportSubscription_Activate(t *testing.T) {
	rs := NewReportSubscription(1, 1, domain.DeliveryMethodEmail)
	rs.IsActive = false
	rs.Activate()
	if !rs.IsActive {
		t.Error("expected IsActive to be true")
	}
}

func TestReportSubscription_Deactivate(t *testing.T) {
	rs := NewReportSubscription(1, 1, domain.DeliveryMethodEmail)
	rs.Deactivate()
	if rs.IsActive {
		t.Error("expected IsActive to be false")
	}
}

func TestReportSubscription_SetDeliveryMethod(t *testing.T) {
	rs := NewReportSubscription(1, 1, domain.DeliveryMethodEmail)
	rs.SetDeliveryMethod(domain.DeliveryMethodBoth)
	if rs.DeliveryMethod != domain.DeliveryMethodBoth {
		t.Errorf("expected delivery method %q, got %q", domain.DeliveryMethodBoth, rs.DeliveryMethod)
	}
}

// --- ReportHistory tests ---

func TestNewReportHistory(t *testing.T) {
	userID := int64(42)
	rh := NewReportHistory(1, &userID, ReportActionCreated)

	if rh.ReportID != 1 {
		t.Errorf("expected report ID 1, got %d", rh.ReportID)
	}
	if rh.UserID == nil || *rh.UserID != 42 {
		t.Errorf("expected user ID 42, got %v", rh.UserID)
	}
	if rh.Action != ReportActionCreated {
		t.Errorf("expected action %q, got %q", ReportActionCreated, rh.Action)
	}
	if rh.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestNewReportHistory_NilUser(t *testing.T) {
	rh := NewReportHistory(1, nil, ReportActionUpdated)
	if rh.UserID != nil {
		t.Error("expected UserID to be nil")
	}
}

func TestReportHistory_SetDetails(t *testing.T) {
	rh := NewReportHistory(1, nil, ReportActionUpdated)
	details := map[string]string{"field": "title", "old": "old title"}

	err := rh.SetDetails(details)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if rh.Details == nil {
		t.Error("expected Details to be set")
	}
}

func TestReportHistory_SetDetails_Invalid(t *testing.T) {
	rh := NewReportHistory(1, nil, ReportActionUpdated)
	err := rh.SetDetails(make(chan int))
	if err == nil {
		t.Error("expected error for unmarshallable value")
	}
}

func TestReportHistory_GetDetails(t *testing.T) {
	rh := NewReportHistory(1, nil, ReportActionUpdated)
	details := map[string]string{"field": "title"}
	_ = rh.SetDetails(details)

	var result map[string]string
	err := rh.GetDetails(&result)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result["field"] != "title" {
		t.Errorf("expected field='title', got %q", result["field"])
	}
}

func TestReportHistory_GetDetails_Nil(t *testing.T) {
	rh := NewReportHistory(1, nil, ReportActionUpdated)

	var result map[string]string
	err := rh.GetDetails(&result)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestReportActionConstants(t *testing.T) {
	tests := []struct {
		action   ReportAction
		expected string
	}{
		{ReportActionCreated, "created"},
		{ReportActionUpdated, "updated"},
		{ReportActionSubmitted, "submitted"},
		{ReportActionApproved, "approved"},
		{ReportActionRejected, "rejected"},
		{ReportActionPublished, "published"},
		{ReportActionDeleted, "deleted"},
	}

	for _, tt := range tests {
		if string(tt.action) != tt.expected {
			t.Errorf("expected %q, got %q", tt.expected, tt.action)
		}
	}
}

// --- ReportGenerationLog tests ---

func TestNewReportGenerationLog(t *testing.T) {
	rgl := NewReportGenerationLog(1)

	if rgl.ReportID != 1 {
		t.Errorf("expected report ID 1, got %d", rgl.ReportID)
	}
	if rgl.Status != domain.GenerationStatusStarted {
		t.Errorf("expected status %q, got %q", domain.GenerationStatusStarted, rgl.Status)
	}
	if rgl.StartedAt.IsZero() {
		t.Error("expected StartedAt to be set")
	}
	if rgl.CompletedAt != nil {
		t.Error("expected CompletedAt to be nil")
	}
	if !rgl.IsInProgress() {
		t.Error("expected IsInProgress to return true")
	}
}

func TestReportGenerationLog_Complete(t *testing.T) {
	rgl := NewReportGenerationLog(1)
	rgl.Complete(100)

	if rgl.Status != domain.GenerationStatusCompleted {
		t.Errorf("expected status %q, got %q", domain.GenerationStatusCompleted, rgl.Status)
	}
	if rgl.CompletedAt == nil {
		t.Error("expected CompletedAt to be set")
	}
	if rgl.RecordsProcessed == nil || *rgl.RecordsProcessed != 100 {
		t.Errorf("expected records processed 100, got %v", rgl.RecordsProcessed)
	}
	if rgl.DurationSeconds == nil {
		t.Error("expected DurationSeconds to be set")
	}
	if !rgl.IsCompleted() {
		t.Error("expected IsCompleted to return true")
	}
	if rgl.IsFailed() {
		t.Error("expected IsFailed to return false")
	}
	if rgl.IsInProgress() {
		t.Error("expected IsInProgress to return false")
	}
}

func TestReportGenerationLog_Fail(t *testing.T) {
	rgl := NewReportGenerationLog(1)
	rgl.Fail("something went wrong")

	if rgl.Status != domain.GenerationStatusFailed {
		t.Errorf("expected status %q, got %q", domain.GenerationStatusFailed, rgl.Status)
	}
	if rgl.CompletedAt == nil {
		t.Error("expected CompletedAt to be set")
	}
	if rgl.ErrorMessage == nil || *rgl.ErrorMessage != "something went wrong" {
		t.Errorf("expected error message, got %v", rgl.ErrorMessage)
	}
	if rgl.DurationSeconds == nil {
		t.Error("expected DurationSeconds to be set")
	}
	if !rgl.IsFailed() {
		t.Error("expected IsFailed to return true")
	}
	if rgl.IsCompleted() {
		t.Error("expected IsCompleted to return false")
	}
}

// --- CustomReport tests ---

func TestNewCustomReport(t *testing.T) {
	cr := NewCustomReport("My Report", "Description", DataSourceDocuments, 42)

	if cr.Name != "My Report" {
		t.Errorf("expected name 'My Report', got %q", cr.Name)
	}
	if cr.Description != "Description" {
		t.Errorf("expected description 'Description', got %q", cr.Description)
	}
	if cr.DataSource != DataSourceDocuments {
		t.Errorf("expected data source %q, got %q", DataSourceDocuments, cr.DataSource)
	}
	if cr.CreatedBy != 42 {
		t.Errorf("expected created by 42, got %d", cr.CreatedBy)
	}
	if cr.IsPublic {
		t.Error("expected IsPublic to be false")
	}
	if len(cr.Fields) != 0 {
		t.Errorf("expected 0 fields, got %d", len(cr.Fields))
	}
	if len(cr.Filters) != 0 {
		t.Errorf("expected 0 filters, got %d", len(cr.Filters))
	}
	if len(cr.Groupings) != 0 {
		t.Errorf("expected 0 groupings, got %d", len(cr.Groupings))
	}
	if len(cr.Sortings) != 0 {
		t.Errorf("expected 0 sortings, got %d", len(cr.Sortings))
	}
	if cr.ID.String() == "" {
		t.Error("expected UUID to be generated")
	}
}

func TestCustomReport_SetFields(t *testing.T) {
	cr := NewCustomReport("Report", "", DataSourceDocuments, 1)
	fields := []SelectedField{{Order: 1}}
	cr.SetFields(fields)
	if len(cr.Fields) != 1 {
		t.Errorf("expected 1 field, got %d", len(cr.Fields))
	}
}

func TestCustomReport_SetFilters(t *testing.T) {
	cr := NewCustomReport("Report", "", DataSourceDocuments, 1)
	filters := []ReportFilterConfig{{ID: "f1"}}
	cr.SetFilters(filters)
	if len(cr.Filters) != 1 {
		t.Errorf("expected 1 filter, got %d", len(cr.Filters))
	}
}

func TestCustomReport_SetGroupings(t *testing.T) {
	cr := NewCustomReport("Report", "", DataSourceDocuments, 1)
	groupings := []ReportGrouping{{Order: SortOrderAsc}}
	cr.SetGroupings(groupings)
	if len(cr.Groupings) != 1 {
		t.Errorf("expected 1 grouping, got %d", len(cr.Groupings))
	}
}

func TestCustomReport_SetSortings(t *testing.T) {
	cr := NewCustomReport("Report", "", DataSourceDocuments, 1)
	sortings := []ReportSorting{{Order: SortOrderDesc}}
	cr.SetSortings(sortings)
	if len(cr.Sortings) != 1 {
		t.Errorf("expected 1 sorting, got %d", len(cr.Sortings))
	}
}

func TestCustomReport_SetPublic(t *testing.T) {
	cr := NewCustomReport("Report", "", DataSourceDocuments, 1)
	cr.SetPublic(true)
	if !cr.IsPublic {
		t.Error("expected IsPublic to be true")
	}
	cr.SetPublic(false)
	if cr.IsPublic {
		t.Error("expected IsPublic to be false")
	}
}

func TestCustomReport_CanEdit(t *testing.T) {
	cr := NewCustomReport("Report", "", DataSourceDocuments, 42)
	if !cr.CanEdit(42) {
		t.Error("expected creator to be able to edit")
	}
	if cr.CanEdit(99) {
		t.Error("expected non-creator to not be able to edit")
	}
}

func TestCustomReport_GetFieldsJSON(t *testing.T) {
	cr := NewCustomReport("Report", "", DataSourceDocuments, 1)
	cr.Fields = []SelectedField{{Order: 1, Alias: "test"}}

	data, err := cr.GetFieldsJSON()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if data == nil {
		t.Error("expected JSON data")
	}
}

func TestCustomReport_GetFiltersJSON(t *testing.T) {
	cr := NewCustomReport("Report", "", DataSourceDocuments, 1)
	data, err := cr.GetFiltersJSON()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if data == nil {
		t.Error("expected JSON data")
	}
}

func TestCustomReport_GetGroupingsJSON(t *testing.T) {
	cr := NewCustomReport("Report", "", DataSourceDocuments, 1)
	data, err := cr.GetGroupingsJSON()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if data == nil {
		t.Error("expected JSON data")
	}
}

func TestCustomReport_GetSortingsJSON(t *testing.T) {
	cr := NewCustomReport("Report", "", DataSourceDocuments, 1)
	data, err := cr.GetSortingsJSON()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if data == nil {
		t.Error("expected JSON data")
	}
}

func TestCustomReport_SetFieldsFromJSON(t *testing.T) {
	cr := NewCustomReport("Report", "", DataSourceDocuments, 1)
	fields := []SelectedField{{Order: 1, Alias: "test"}}
	data, _ := json.Marshal(fields)

	err := cr.SetFieldsFromJSON(data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(cr.Fields) != 1 {
		t.Errorf("expected 1 field, got %d", len(cr.Fields))
	}
}

func TestCustomReport_SetFiltersFromJSON(t *testing.T) {
	cr := NewCustomReport("Report", "", DataSourceDocuments, 1)
	filters := []ReportFilterConfig{{ID: "f1"}}
	data, _ := json.Marshal(filters)

	err := cr.SetFiltersFromJSON(data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(cr.Filters) != 1 {
		t.Errorf("expected 1 filter, got %d", len(cr.Filters))
	}
}

func TestCustomReport_SetGroupingsFromJSON(t *testing.T) {
	cr := NewCustomReport("Report", "", DataSourceDocuments, 1)
	groupings := []ReportGrouping{{Order: SortOrderAsc}}
	data, _ := json.Marshal(groupings)

	err := cr.SetGroupingsFromJSON(data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(cr.Groupings) != 1 {
		t.Errorf("expected 1 grouping, got %d", len(cr.Groupings))
	}
}

func TestCustomReport_SetSortingsFromJSON(t *testing.T) {
	cr := NewCustomReport("Report", "", DataSourceDocuments, 1)
	sortings := []ReportSorting{{Order: SortOrderDesc}}
	data, _ := json.Marshal(sortings)

	err := cr.SetSortingsFromJSON(data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(cr.Sortings) != 1 {
		t.Errorf("expected 1 sorting, got %d", len(cr.Sortings))
	}
}

func TestCustomReport_SetFieldsFromJSON_Invalid(t *testing.T) {
	cr := NewCustomReport("Report", "", DataSourceDocuments, 1)
	err := cr.SetFieldsFromJSON([]byte("invalid"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestDataSourceType_IsValid(t *testing.T) {
	tests := []struct {
		name string
		ds   DataSourceType
		want bool
	}{
		{"documents", DataSourceDocuments, true},
		{"users", DataSourceUsers, true},
		{"events", DataSourceEvents, true},
		{"tasks", DataSourceTasks, true},
		{"students", DataSourceStudents, true},
		{"invalid", DataSourceType("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ds.IsValid()
			if got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExportFormat_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		format ExportFormat
		want   bool
	}{
		{"pdf", ExportFormatPDF, true},
		{"xlsx", ExportFormatXLSX, true},
		{"csv", ExportFormatCSV, true},
		{"invalid", ExportFormat("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.format.IsValid()
			if got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
