package dto

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToCustomReportOutput(t *testing.T) {
	now := time.Now()
	id := uuid.New()

	report := &entities.CustomReport{
		ID:          id,
		Name:        "My Report",
		Description: "A custom report",
		DataSource:  entities.DataSourceDocuments,
		Fields: []entities.SelectedField{
			{
				Field: entities.ReportField{
					ID: "f1", Name: "title", Label: "Title",
					Type: entities.FieldTypeString, Source: entities.DataSourceDocuments,
				},
				Order: 1,
				Alias: "doc_title",
			},
		},
		Filters: []entities.ReportFilterConfig{
			{
				ID: "filter1",
				Field: entities.ReportField{
					ID: "f2", Name: "status", Label: "Status",
					Type: entities.FieldTypeEnum, Source: entities.DataSourceDocuments,
					EnumValues: []string{"draft", "published"},
				},
				Operator: entities.FilterEquals,
				Value:    "draft",
			},
		},
		Groupings: []entities.ReportGrouping{
			{
				Field: entities.ReportField{
					ID: "f1", Name: "title", Label: "Title",
					Type: entities.FieldTypeString, Source: entities.DataSourceDocuments,
				},
				Order: entities.SortOrderAsc,
			},
		},
		Sortings: []entities.ReportSorting{
			{
				Field: entities.ReportField{
					ID: "f1", Name: "title", Label: "Title",
					Type: entities.FieldTypeString, Source: entities.DataSourceDocuments,
				},
				Order: entities.SortOrderDesc,
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
		CreatedBy: 42,
		IsPublic:  true,
	}

	output := ToCustomReportOutput(report)

	assert.Equal(t, id, output.ID)
	assert.Equal(t, "My Report", output.Name)
	assert.Equal(t, "A custom report", output.Description)
	assert.Equal(t, "documents", output.DataSource)
	assert.True(t, output.IsPublic)

	require.Len(t, output.Fields, 1)
	assert.Equal(t, "title", output.Fields[0].Field.Name)
	assert.Equal(t, "doc_title", output.Fields[0].Alias)

	require.Len(t, output.Filters, 1)
	assert.Equal(t, "filter1", output.Filters[0].ID)
	assert.Equal(t, "equals", output.Filters[0].Operator)

	require.Len(t, output.Groupings, 1)
	assert.Equal(t, "asc", output.Groupings[0].Order)

	require.Len(t, output.Sortings, 1)
	assert.Equal(t, "desc", output.Sortings[0].Order)
}

func TestToCustomReportFilter(t *testing.T) {
	userID := int64(42)
	isPublic := true
	input := CustomReportFilterInput{
		DataSource: "documents",
		IsPublic:   &isPublic,
		Search:     "test",
		Page:       2,
		PageSize:   25,
	}

	filter := ToCustomReportFilter(input, &userID)

	assert.Equal(t, 2, filter.Page)
	assert.Equal(t, 25, filter.PageSize)
	assert.Equal(t, "test", filter.Search)
	assert.Equal(t, &userID, filter.CreatedBy)
	require.NotNil(t, filter.DataSource)
	assert.Equal(t, entities.DataSourceDocuments, *filter.DataSource)
	assert.Equal(t, &isPublic, filter.IsPublic)
}

func TestToCustomReportFilter_Defaults(t *testing.T) {
	input := CustomReportFilterInput{
		Page:     0,
		PageSize: 0,
	}

	filter := ToCustomReportFilter(input, nil)

	assert.Equal(t, 1, filter.Page)
	assert.Equal(t, 10, filter.PageSize)
	assert.Nil(t, filter.CreatedBy)
	assert.Nil(t, filter.DataSource)
}

func TestToSelectedFields(t *testing.T) {
	dtos := []SelectedFieldDTO{
		{
			Field: ReportFieldDTO{
				ID: "f1", Name: "title", Label: "Title",
				Type: "string", Source: "documents",
			},
			Order: 1,
			Alias: "t",
		},
	}

	fields := ToSelectedFields(dtos)

	require.Len(t, fields, 1)
	assert.Equal(t, "title", fields[0].Field.Name)
	assert.Equal(t, entities.FieldTypeString, fields[0].Field.Type)
	assert.Equal(t, entities.DataSourceDocuments, fields[0].Field.Source)
	assert.Equal(t, 1, fields[0].Order)
	assert.Equal(t, "t", fields[0].Alias)
}

func TestToReportFilters(t *testing.T) {
	dtos := []ReportFilterDTO{
		{
			ID: "f1",
			Field: ReportFieldDTO{
				ID: "f1", Name: "status", Label: "Status",
				Type: "enum", Source: "documents",
			},
			Operator: "equals",
			Value:    "draft",
		},
	}

	filters := ToReportFilters(dtos)

	require.Len(t, filters, 1)
	assert.Equal(t, "f1", filters[0].ID)
	assert.Equal(t, entities.FilterEquals, filters[0].Operator)
	assert.Equal(t, "draft", filters[0].Value)
}

func TestToReportGroupings(t *testing.T) {
	dtos := []ReportGroupingDTO{
		{
			Field: ReportFieldDTO{ID: "f1", Name: "title", Label: "Title", Type: "string", Source: "documents"},
			Order: "asc",
		},
	}

	groupings := ToReportGroupings(dtos)

	require.Len(t, groupings, 1)
	assert.Equal(t, entities.SortOrderAsc, groupings[0].Order)
}

func TestToReportSortings(t *testing.T) {
	dtos := []ReportSortingDTO{
		{
			Field: ReportFieldDTO{ID: "f1", Name: "title", Label: "Title", Type: "string", Source: "documents"},
			Order: "desc",
		},
	}

	sortings := ToReportSortings(dtos)

	require.Len(t, sortings, 1)
	assert.Equal(t, entities.SortOrderDesc, sortings[0].Order)
}
