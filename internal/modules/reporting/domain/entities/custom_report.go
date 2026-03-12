package entities

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// DataSourceType represents the data source for custom reports
type DataSourceType string

// DataSourceType values.
const (
	DataSourceDocuments DataSourceType = "documents"
	DataSourceUsers     DataSourceType = "users"
	DataSourceEvents    DataSourceType = "events"
	DataSourceTasks     DataSourceType = "tasks"
	DataSourceStudents  DataSourceType = "students"
)

// IsValid checks if the data source type is valid
func (d DataSourceType) IsValid() bool {
	switch d {
	case DataSourceDocuments, DataSourceUsers, DataSourceEvents, DataSourceTasks, DataSourceStudents:
		return true
	}
	return false
}

// FieldType represents the type of a report field
type FieldType string

// FieldType values.
const (
	FieldTypeString  FieldType = "string"
	FieldTypeNumber  FieldType = "number"
	FieldTypeDate    FieldType = "date"
	FieldTypeBoolean FieldType = "boolean"
	FieldTypeEnum    FieldType = "enum"
)

// AggregationType represents the aggregation type for a field
type AggregationType string

// AggregationType values.
const (
	AggregationNone  AggregationType = ""
	AggregationCount AggregationType = "count"
	AggregationSum   AggregationType = "sum"
	AggregationAvg   AggregationType = "avg"
	AggregationMin   AggregationType = "min"
	AggregationMax   AggregationType = "max"
)

// FilterOperator represents the operator for filtering
type FilterOperator string

// FilterOperator values.
const (
	FilterEquals         FilterOperator = "equals"
	FilterNotEquals      FilterOperator = "not_equals"
	FilterContains       FilterOperator = "contains"
	FilterNotContains    FilterOperator = "not_contains"
	FilterStartsWith     FilterOperator = "starts_with"
	FilterEndsWith       FilterOperator = "ends_with"
	FilterGreaterThan    FilterOperator = "greater_than"
	FilterLessThan       FilterOperator = "less_than"
	FilterGreaterOrEqual FilterOperator = "greater_or_equal"
	FilterLessOrEqual    FilterOperator = "less_or_equal"
	FilterBetween        FilterOperator = "between"
	FilterIn             FilterOperator = "in"
	FilterNotIn          FilterOperator = "not_in"
	FilterIsNull         FilterOperator = "is_null"
	FilterIsNotNull      FilterOperator = "is_not_null"
)

// SortOrder represents the sort order
type SortOrder string

// SortOrder values.
const (
	SortOrderAsc  SortOrder = "asc"
	SortOrderDesc SortOrder = "desc"
)

// ExportFormat represents the export format
type ExportFormat string

// ExportFormat values.
const (
	ExportFormatPDF  ExportFormat = "pdf"
	ExportFormatXLSX ExportFormat = "xlsx"
	ExportFormatCSV  ExportFormat = "csv"
)

// IsValid checks if the export format is valid
func (e ExportFormat) IsValid() bool {
	switch e {
	case ExportFormatPDF, ExportFormatXLSX, ExportFormatCSV:
		return true
	}
	return false
}

// ReportField represents a field definition
type ReportField struct {
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	Label      string         `json:"label"`
	Type       FieldType      `json:"type"`
	Source     DataSourceType `json:"source"`
	EnumValues []string       `json:"enumValues,omitempty"`
}

// SelectedField represents a selected field in the report
type SelectedField struct {
	Field       ReportField     `json:"field"`
	Order       int             `json:"order"`
	Alias       string          `json:"alias,omitempty"`
	Aggregation AggregationType `json:"aggregation,omitempty"`
}

// ReportFilterConfig represents a filter configuration
type ReportFilterConfig struct {
	ID       string         `json:"id"`
	Field    ReportField    `json:"field"`
	Operator FilterOperator `json:"operator"`
	Value    interface{}    `json:"value"`
	Value2   interface{}    `json:"value2,omitempty"` // For 'between' operator
}

// ReportGrouping represents a grouping configuration
type ReportGrouping struct {
	Field ReportField `json:"field"`
	Order SortOrder   `json:"order"`
}

// ReportSorting represents a sorting configuration
type ReportSorting struct {
	Field ReportField `json:"field"`
	Order SortOrder   `json:"order"`
}

// CustomReport represents a custom report template
type CustomReport struct {
	ID          uuid.UUID            `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description,omitempty"`
	DataSource  DataSourceType       `json:"dataSource"`
	Fields      []SelectedField      `json:"fields"`
	Filters     []ReportFilterConfig `json:"filters"`
	Groupings   []ReportGrouping     `json:"groupings"`
	Sortings    []ReportSorting      `json:"sortings"`
	CreatedAt   time.Time            `json:"createdAt"`
	UpdatedAt   time.Time            `json:"updatedAt"`
	CreatedBy   int64                `json:"createdBy"`
	IsPublic    bool                 `json:"isPublic"`
}

// NewCustomReport creates a new custom report
func NewCustomReport(name string, description string, dataSource DataSourceType, createdBy int64) *CustomReport {
	now := time.Now()
	return &CustomReport{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		DataSource:  dataSource,
		Fields:      []SelectedField{},
		Filters:     []ReportFilterConfig{},
		Groupings:   []ReportGrouping{},
		Sortings:    []ReportSorting{},
		CreatedAt:   now,
		UpdatedAt:   now,
		CreatedBy:   createdBy,
		IsPublic:    false,
	}
}

// SetFields sets the selected fields for the report
func (r *CustomReport) SetFields(fields []SelectedField) {
	r.Fields = fields
	r.UpdatedAt = time.Now()
}

// SetFilters sets the filters for the report
func (r *CustomReport) SetFilters(filters []ReportFilterConfig) {
	r.Filters = filters
	r.UpdatedAt = time.Now()
}

// SetGroupings sets the groupings for the report
func (r *CustomReport) SetGroupings(groupings []ReportGrouping) {
	r.Groupings = groupings
	r.UpdatedAt = time.Now()
}

// SetSortings sets the sortings for the report
func (r *CustomReport) SetSortings(sortings []ReportSorting) {
	r.Sortings = sortings
	r.UpdatedAt = time.Now()
}

// SetPublic sets the public visibility of the report
func (r *CustomReport) SetPublic(isPublic bool) {
	r.IsPublic = isPublic
	r.UpdatedAt = time.Now()
}

// GetFieldsJSON returns fields as JSON bytes
func (r *CustomReport) GetFieldsJSON() ([]byte, error) {
	return json.Marshal(r.Fields)
}

// GetFiltersJSON returns filters as JSON bytes
func (r *CustomReport) GetFiltersJSON() ([]byte, error) {
	return json.Marshal(r.Filters)
}

// GetGroupingsJSON returns groupings as JSON bytes
func (r *CustomReport) GetGroupingsJSON() ([]byte, error) {
	return json.Marshal(r.Groupings)
}

// GetSortingsJSON returns sortings as JSON bytes
func (r *CustomReport) GetSortingsJSON() ([]byte, error) {
	return json.Marshal(r.Sortings)
}

// SetFieldsFromJSON sets fields from JSON bytes
func (r *CustomReport) SetFieldsFromJSON(data []byte) error {
	return json.Unmarshal(data, &r.Fields)
}

// SetFiltersFromJSON sets filters from JSON bytes
func (r *CustomReport) SetFiltersFromJSON(data []byte) error {
	return json.Unmarshal(data, &r.Filters)
}

// SetGroupingsFromJSON sets groupings from JSON bytes
func (r *CustomReport) SetGroupingsFromJSON(data []byte) error {
	return json.Unmarshal(data, &r.Groupings)
}

// SetSortingsFromJSON sets sortings from JSON bytes
func (r *CustomReport) SetSortingsFromJSON(data []byte) error {
	return json.Unmarshal(data, &r.Sortings)
}

// CanEdit checks if the user can edit the report
func (r *CustomReport) CanEdit(userID int64) bool {
	return r.CreatedBy == userID
}

// ExportOptions represents export options for a report
type ExportOptions struct {
	Format         ExportFormat `json:"format"`
	IncludeHeaders bool         `json:"includeHeaders"`
	PageSize       string       `json:"pageSize,omitempty"`
	Orientation    string       `json:"orientation,omitempty"`
}

// ReportExecutionResult represents the result of report execution
type ReportExecutionResult struct {
	Columns    []ReportColumn           `json:"columns"`
	Rows       []map[string]interface{} `json:"rows"`
	TotalCount int64                    `json:"totalCount"`
	Page       int                      `json:"page"`
	PageSize   int                      `json:"pageSize"`
	TotalPages int                      `json:"totalPages"`
}

// ReportColumn represents a column in the result
type ReportColumn struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}
