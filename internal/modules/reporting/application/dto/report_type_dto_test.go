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

func TestToReportTypeOutput(t *testing.T) {
	now := time.Now()
	desc := "Monthly report"
	cat := domain.ReportCategoryAcademic
	pt := domain.PeriodTypeMonthly
	tplPath := "/templates/monthly.tmpl"

	rt := &entities.ReportType{
		ID:           1,
		Name:         "Monthly Report",
		Code:         "monthly",
		Description:  &desc,
		Category:     &cat,
		TemplatePath: &tplPath,
		OutputFormat: domain.OutputFormatPDF,
		IsPeriodic:   true,
		PeriodType:   &pt,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	output := ToReportTypeOutput(rt)

	require.NotNil(t, output)
	assert.Equal(t, int64(1), output.ID)
	assert.Equal(t, "Monthly Report", output.Name)
	assert.Equal(t, "monthly", output.Code)
	assert.Equal(t, &desc, output.Description)
	catStr := "academic"
	assert.Equal(t, &catStr, output.Category)
	ptStr := "monthly"
	assert.Equal(t, &ptStr, output.PeriodType)
	assert.Equal(t, "pdf", output.OutputFormat)
	assert.True(t, output.IsPeriodic)
}

func TestToReportTypeOutput_WithParameters(t *testing.T) {
	now := time.Now()
	opts, _ := json.Marshal([]string{"a", "b"})

	rt := &entities.ReportType{
		ID:           1,
		Name:         "Test",
		Code:         "test",
		OutputFormat: domain.OutputFormatXLSX,
		Parameters: []entities.ReportParameter{
			{
				ID:            1,
				ReportTypeID:  1,
				ParameterName: "department",
				ParameterType: domain.ParameterTypeString,
				IsRequired:    true,
				Options:       opts,
				DisplayOrder:  1,
				CreatedAt:     now,
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	output := ToReportTypeOutput(rt)

	require.Len(t, output.Parameters, 1)
	assert.Equal(t, "department", output.Parameters[0].ParameterName)
	assert.Equal(t, "string", output.Parameters[0].ParameterType)
	assert.True(t, output.Parameters[0].IsRequired)
	assert.NotNil(t, output.Parameters[0].Options)
}

func TestToReportTypeOutput_WithTemplates(t *testing.T) {
	now := time.Now()
	rt := &entities.ReportType{
		ID:           1,
		Name:         "Test",
		Code:         "test",
		OutputFormat: domain.OutputFormatPDF,
		Templates: []entities.ReportTemplate{
			{
				ID:           1,
				ReportTypeID: 1,
				Name:         "Default",
				Content:      "template content",
				IsDefault:    true,
				CreatedBy:    42,
				CreatedAt:    now,
				UpdatedAt:    now,
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	output := ToReportTypeOutput(rt)

	require.Len(t, output.Templates, 1)
	assert.Equal(t, "Default", output.Templates[0].Name)
	assert.True(t, output.Templates[0].IsDefault)
}

func TestToReportSubscriptionOutput(t *testing.T) {
	now := time.Now()
	sub := &entities.ReportSubscription{
		ID:             1,
		ReportTypeID:   2,
		UserID:         42,
		DeliveryMethod: domain.DeliveryMethodEmail,
		IsActive:       true,
		CreatedAt:      now,
	}

	output := ToReportSubscriptionOutput(sub)

	require.NotNil(t, output)
	assert.Equal(t, int64(1), output.ID)
	assert.Equal(t, int64(2), output.ReportTypeID)
	assert.Equal(t, int64(42), output.UserID)
	assert.Equal(t, "email", output.DeliveryMethod)
	assert.True(t, output.IsActive)
}
