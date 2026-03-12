package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain/repositories"
)

// ReportFieldDTO represents a field definition for API
type ReportFieldDTO struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Label      string   `json:"label"`
	Type       string   `json:"type"`
	Source     string   `json:"source"`
	EnumValues []string `json:"enumValues,omitempty"`
}

// SelectedFieldDTO represents a selected field in the report for API
type SelectedFieldDTO struct {
	Field       ReportFieldDTO `json:"field"`
	Order       int            `json:"order"`
	Alias       string         `json:"alias,omitempty"`
	Aggregation string         `json:"aggregation,omitempty"`
}

// ReportFilterDTO represents a filter configuration for API
type ReportFilterDTO struct {
	ID       string         `json:"id"`
	Field    ReportFieldDTO `json:"field"`
	Operator string         `json:"operator"`
	Value    interface{}    `json:"value"`
	Value2   interface{}    `json:"value2,omitempty"`
}

// ReportGroupingDTO represents a grouping configuration for API
type ReportGroupingDTO struct {
	Field ReportFieldDTO `json:"field"`
	Order string         `json:"order"`
}

// ReportSortingDTO represents a sorting configuration for API
type ReportSortingDTO struct {
	Field ReportFieldDTO `json:"field"`
	Order string         `json:"order"`
}

// CreateCustomReportInput represents input for creating a custom report
type CreateCustomReportInput struct {
	Name        string              `json:"name" validate:"required,min=1,max=255"`
	Description string              `json:"description,omitempty" validate:"max=1000"`
	DataSource  string              `json:"dataSource" validate:"required,oneof=documents users events tasks students"`
	Fields      []SelectedFieldDTO  `json:"fields" validate:"required,min=1"`
	Filters     []ReportFilterDTO   `json:"filters"`
	Groupings   []ReportGroupingDTO `json:"groupings"`
	Sortings    []ReportSortingDTO  `json:"sortings"`
	IsPublic    bool                `json:"isPublic"`
}

// UpdateCustomReportInput represents input for updating a custom report
type UpdateCustomReportInput struct {
	Name        *string             `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Description *string             `json:"description,omitempty" validate:"omitempty,max=1000"`
	DataSource  *string             `json:"dataSource,omitempty" validate:"omitempty,oneof=documents users events tasks students"`
	Fields      []SelectedFieldDTO  `json:"fields,omitempty"`
	Filters     []ReportFilterDTO   `json:"filters,omitempty"`
	Groupings   []ReportGroupingDTO `json:"groupings,omitempty"`
	Sortings    []ReportSortingDTO  `json:"sortings,omitempty"`
	IsPublic    *bool               `json:"isPublic,omitempty"`
}

// ExecuteReportInput represents input for executing a custom report
type ExecuteReportInput struct {
	Page     int `json:"page" validate:"min=1"`
	PageSize int `json:"pageSize" validate:"min=1,max=1000"`
}

// ExportReportInput represents input for exporting a custom report
type ExportReportInput struct {
	Format         string `json:"format" validate:"required,oneof=pdf xlsx csv"`
	IncludeHeaders bool   `json:"includeHeaders"`
	PageSize       string `json:"pageSize,omitempty" validate:"omitempty,oneof=A4 A3 Letter"`
	Orientation    string `json:"orientation,omitempty" validate:"omitempty,oneof=portrait landscape"`
}

// CustomReportOutput represents output for a custom report
type CustomReportOutput struct {
	ID          uuid.UUID           `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	DataSource  string              `json:"dataSource"`
	Fields      []SelectedFieldDTO  `json:"fields"`
	Filters     []ReportFilterDTO   `json:"filters"`
	Groupings   []ReportGroupingDTO `json:"groupings"`
	Sortings    []ReportSortingDTO  `json:"sortings"`
	CreatedAt   time.Time           `json:"createdAt"`
	UpdatedAt   time.Time           `json:"updatedAt"`
	CreatedBy   int64               `json:"createdBy"`
	IsPublic    bool                `json:"isPublic"`
}

// CustomReportListOutput represents output for a list of custom reports
type CustomReportListOutput struct {
	Reports    []CustomReportOutput `json:"reports"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"pageSize"`
	TotalPages int                  `json:"totalPages"`
}

// CustomReportFilterInput represents filter input for listing custom reports
type CustomReportFilterInput struct {
	DataSource string `json:"dataSource,omitempty" form:"dataSource" validate:"omitempty,oneof=documents users events tasks students"`
	IsPublic   *bool  `json:"isPublic,omitempty" form:"isPublic"`
	Search     string `json:"search,omitempty" form:"search"`
	Page       int    `json:"page,omitempty" form:"page" validate:"min=1"`
	PageSize   int    `json:"pageSize,omitempty" form:"pageSize" validate:"min=1,max=100"`
}

// ReportColumnOutput represents a column in the execution result
type ReportColumnOutput struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

// ExecuteReportOutput represents output for report execution
type ExecuteReportOutput struct {
	Columns    []ReportColumnOutput     `json:"columns"`
	Rows       []map[string]interface{} `json:"rows"`
	TotalCount int64                    `json:"totalCount"`
	Page       int                      `json:"page"`
	PageSize   int                      `json:"pageSize"`
	TotalPages int                      `json:"totalPages"`
}

// ToCustomReportOutput converts entity to DTO
func ToCustomReportOutput(report *entities.CustomReport) CustomReportOutput {
	fields := make([]SelectedFieldDTO, len(report.Fields))
	for i, f := range report.Fields {
		fields[i] = SelectedFieldDTO{
			Field: ReportFieldDTO{
				ID:         f.Field.ID,
				Name:       f.Field.Name,
				Label:      f.Field.Label,
				Type:       string(f.Field.Type),
				Source:     string(f.Field.Source),
				EnumValues: f.Field.EnumValues,
			},
			Order:       f.Order,
			Alias:       f.Alias,
			Aggregation: string(f.Aggregation),
		}
	}

	filters := make([]ReportFilterDTO, len(report.Filters))
	for i, f := range report.Filters {
		filters[i] = ReportFilterDTO{
			ID: f.ID,
			Field: ReportFieldDTO{
				ID:         f.Field.ID,
				Name:       f.Field.Name,
				Label:      f.Field.Label,
				Type:       string(f.Field.Type),
				Source:     string(f.Field.Source),
				EnumValues: f.Field.EnumValues,
			},
			Operator: string(f.Operator),
			Value:    f.Value,
			Value2:   f.Value2,
		}
	}

	groupings := make([]ReportGroupingDTO, len(report.Groupings))
	for i, g := range report.Groupings {
		groupings[i] = ReportGroupingDTO{
			Field: ReportFieldDTO{
				ID:         g.Field.ID,
				Name:       g.Field.Name,
				Label:      g.Field.Label,
				Type:       string(g.Field.Type),
				Source:     string(g.Field.Source),
				EnumValues: g.Field.EnumValues,
			},
			Order: string(g.Order),
		}
	}

	sortings := make([]ReportSortingDTO, len(report.Sortings))
	for i, s := range report.Sortings {
		sortings[i] = ReportSortingDTO{
			Field: ReportFieldDTO{
				ID:         s.Field.ID,
				Name:       s.Field.Name,
				Label:      s.Field.Label,
				Type:       string(s.Field.Type),
				Source:     string(s.Field.Source),
				EnumValues: s.Field.EnumValues,
			},
			Order: string(s.Order),
		}
	}

	return CustomReportOutput{
		ID:          report.ID,
		Name:        report.Name,
		Description: report.Description,
		DataSource:  string(report.DataSource),
		Fields:      fields,
		Filters:     filters,
		Groupings:   groupings,
		Sortings:    sortings,
		CreatedAt:   report.CreatedAt,
		UpdatedAt:   report.UpdatedAt,
		CreatedBy:   report.CreatedBy,
		IsPublic:    report.IsPublic,
	}
}

// ToCustomReportFilter converts filter input to repository filter
func ToCustomReportFilter(input CustomReportFilterInput, userID *int64) repositories.CustomReportFilter {
	filter := repositories.CustomReportFilter{
		Page:     input.Page,
		PageSize: input.PageSize,
		Search:   input.Search,
	}

	if input.Page < 1 {
		filter.Page = 1
	}
	if input.PageSize < 1 {
		filter.PageSize = 10
	}

	if userID != nil {
		filter.CreatedBy = userID
	}

	if input.DataSource != "" {
		ds := entities.DataSourceType(input.DataSource)
		filter.DataSource = &ds
	}

	if input.IsPublic != nil {
		filter.IsPublic = input.IsPublic
	}

	return filter
}

// ToSelectedFields converts DTO to entity fields
func ToSelectedFields(dtos []SelectedFieldDTO) []entities.SelectedField {
	fields := make([]entities.SelectedField, len(dtos))
	for i, dto := range dtos {
		fields[i] = entities.SelectedField{
			Field: entities.ReportField{
				ID:         dto.Field.ID,
				Name:       dto.Field.Name,
				Label:      dto.Field.Label,
				Type:       entities.FieldType(dto.Field.Type),
				Source:     entities.DataSourceType(dto.Field.Source),
				EnumValues: dto.Field.EnumValues,
			},
			Order:       dto.Order,
			Alias:       dto.Alias,
			Aggregation: entities.AggregationType(dto.Aggregation),
		}
	}
	return fields
}

// ToReportFilters converts DTO to entity filters
func ToReportFilters(dtos []ReportFilterDTO) []entities.ReportFilterConfig {
	filters := make([]entities.ReportFilterConfig, len(dtos))
	for i, dto := range dtos {
		filters[i] = entities.ReportFilterConfig{
			ID: dto.ID,
			Field: entities.ReportField{
				ID:         dto.Field.ID,
				Name:       dto.Field.Name,
				Label:      dto.Field.Label,
				Type:       entities.FieldType(dto.Field.Type),
				Source:     entities.DataSourceType(dto.Field.Source),
				EnumValues: dto.Field.EnumValues,
			},
			Operator: entities.FilterOperator(dto.Operator),
			Value:    dto.Value,
			Value2:   dto.Value2,
		}
	}
	return filters
}

// ToReportGroupings converts DTO to entity groupings
func ToReportGroupings(dtos []ReportGroupingDTO) []entities.ReportGrouping {
	groupings := make([]entities.ReportGrouping, len(dtos))
	for i, dto := range dtos {
		groupings[i] = entities.ReportGrouping{
			Field: entities.ReportField{
				ID:         dto.Field.ID,
				Name:       dto.Field.Name,
				Label:      dto.Field.Label,
				Type:       entities.FieldType(dto.Field.Type),
				Source:     entities.DataSourceType(dto.Field.Source),
				EnumValues: dto.Field.EnumValues,
			},
			Order: entities.SortOrder(dto.Order),
		}
	}
	return groupings
}

// ToReportSortings converts DTO to entity sortings
func ToReportSortings(dtos []ReportSortingDTO) []entities.ReportSorting {
	sortings := make([]entities.ReportSorting, len(dtos))
	for i, dto := range dtos {
		sortings[i] = entities.ReportSorting{
			Field: entities.ReportField{
				ID:         dto.Field.ID,
				Name:       dto.Field.Name,
				Label:      dto.Field.Label,
				Type:       entities.FieldType(dto.Field.Type),
				Source:     entities.DataSourceType(dto.Field.Source),
				EnumValues: dto.Field.EnumValues,
			},
			Order: entities.SortOrder(dto.Order),
		}
	}
	return sortings
}
