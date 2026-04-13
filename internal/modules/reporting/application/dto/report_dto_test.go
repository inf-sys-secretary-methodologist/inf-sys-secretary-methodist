package dto

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToReportOutput(t *testing.T) {
	now := time.Now()
	desc := "Description"
	periodStart := now.Add(-30 * 24 * time.Hour)
	periodEnd := now
	fileName := "report.pdf"
	fileSize := int64(4096)
	mimeType := "application/pdf"
	reviewerComment := "Looks good"
	reviewedAt := now.Add(-time.Hour)
	publishedAt := now

	rt := &entities.ReportType{Name: "Monthly", Code: "monthly"}

	params, _ := json.Marshal(map[string]interface{}{"key": "val"})

	report := &entities.Report{
		ID:              1,
		ReportTypeID:    2,
		Title:           "Test Report",
		Description:     &desc,
		PeriodStart:     &periodStart,
		PeriodEnd:       &periodEnd,
		AuthorID:        42,
		Status:          domain.ReportStatusApproved,
		FileName:        &fileName,
		FileSize:        &fileSize,
		MimeType:        &mimeType,
		Parameters:      params,
		ReviewedBy:      ptrInt64(10),
		ReviewerComment: &reviewerComment,
		ReviewedAt:      &reviewedAt,
		PublishedAt:     &publishedAt,
		IsPublic:        true,
		CreatedAt:       now,
		UpdatedAt:       now,
		ReportType:      rt,
	}

	output := ToReportOutput(report)

	require.NotNil(t, output)
	assert.Equal(t, int64(1), output.ID)
	assert.Equal(t, int64(2), output.ReportTypeID)
	assert.Equal(t, "Test Report", output.Title)
	assert.Equal(t, &desc, output.Description)
	assert.Equal(t, "approved", output.Status)
	assert.True(t, output.HasFile)
	assert.Equal(t, "Monthly", output.ReportTypeName)
	assert.Equal(t, "monthly", output.ReportTypeCode)
	assert.Contains(t, output.Parameters, "key")
	assert.Equal(t, &reviewerComment, output.ReviewerComment)
	assert.True(t, output.IsPublic)
}

func TestToReportOutput_NoFile(t *testing.T) {
	report := &entities.Report{
		ID:       1,
		AuthorID: 1,
		Title:    "No File",
		Status:   domain.ReportStatusDraft,
	}

	output := ToReportOutput(report)
	assert.False(t, output.HasFile)
}

func TestToReportOutput_NoReportType(t *testing.T) {
	report := &entities.Report{
		ID:       1,
		AuthorID: 1,
		Title:    "Standalone",
		Status:   domain.ReportStatusDraft,
	}

	output := ToReportOutput(report)
	assert.Empty(t, output.ReportTypeName)
	assert.Empty(t, output.ReportTypeCode)
}

func TestToReportAccessOutput(t *testing.T) {
	now := time.Now()
	role := domain.AccessRoleAdmin
	access := &entities.ReportAccess{
		ID:         1,
		ReportID:   10,
		UserID:     ptrInt64(42),
		Role:       &role,
		Permission: domain.ReportPermissionApprove,
		GrantedBy:  ptrInt64(1),
		CreatedAt:  now,
	}

	output := ToReportAccessOutput(access)

	require.NotNil(t, output)
	assert.Equal(t, int64(1), output.ID)
	assert.Equal(t, int64(10), output.ReportID)
	assert.Equal(t, ptrInt64(42), output.UserID)
	roleStr := "admin"
	assert.Equal(t, &roleStr, output.Role)
	assert.Equal(t, "approve", output.Permission)
}

func TestToReportAccessOutput_NilRole(t *testing.T) {
	access := &entities.ReportAccess{
		ID:         1,
		ReportID:   10,
		Permission: domain.ReportPermissionRead,
		CreatedAt:  time.Now(),
	}

	output := ToReportAccessOutput(access)
	assert.Nil(t, output.Role)
}

func TestToReportCommentOutput(t *testing.T) {
	now := time.Now()
	comment := &entities.ReportComment{
		ID:        1,
		ReportID:  10,
		AuthorID:  42,
		Content:   "Great report",
		CreatedAt: now,
		UpdatedAt: now,
	}

	output := ToReportCommentOutput(comment)

	require.NotNil(t, output)
	assert.Equal(t, int64(1), output.ID)
	assert.Equal(t, "Great report", output.Content)
}

func TestToReportHistoryOutput(t *testing.T) {
	now := time.Now()
	details, _ := json.Marshal(map[string]interface{}{"field": "status"})
	history := &entities.ReportHistory{
		ID:        1,
		ReportID:  10,
		UserID:    ptrInt64(42),
		Action:    entities.ReportActionCreated,
		Details:   details,
		CreatedAt: now,
	}

	output := ToReportHistoryOutput(history)

	require.NotNil(t, output)
	assert.Equal(t, int64(1), output.ID)
	assert.Contains(t, output.Details, "field")
}

func TestToReportGenerationLogOutput(t *testing.T) {
	now := time.Now()
	completedAt := now.Add(time.Minute)
	duration := 60
	records := 100
	log := &entities.ReportGenerationLog{
		ID:               1,
		ReportID:         10,
		StartedAt:        now,
		CompletedAt:      &completedAt,
		Status:           domain.GenerationStatusCompleted,
		DurationSeconds:  &duration,
		RecordsProcessed: &records,
	}

	output := ToReportGenerationLogOutput(log)

	require.NotNil(t, output)
	assert.Equal(t, int64(1), output.ID)
	assert.Equal(t, &completedAt, output.CompletedAt)
	assert.Equal(t, "completed", output.Status)
	assert.Equal(t, &duration, output.DurationSeconds)
	assert.Equal(t, &records, output.RecordsProcessed)
}

func TestToReportFilter_ValidStatus(t *testing.T) {
	status := "draft"
	input := &ReportFilterInput{Status: &status}

	result, err := ToReportFilter(input)

	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, domain.ReportStatusDraft, *result)
}

func TestToReportFilter_NilStatus(t *testing.T) {
	input := &ReportFilterInput{}

	result, err := ToReportFilter(input)

	assert.NoError(t, err)
	assert.Nil(t, result)
}

func ptrInt64(v int64) *int64 {
	return &v
}
