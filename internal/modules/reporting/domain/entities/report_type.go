package entities

import (
	"encoding/json"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain"
)

// ReportType represents a type/template of report
// Aligned with migrations/006_create_reports_schema.up.sql - report_types table
type ReportType struct {
	ID           int64                  `db:"id" json:"id"`
	Name         string                 `db:"name" json:"name"`
	Code         string                 `db:"code" json:"code"`
	Description  *string                `db:"description" json:"description,omitempty"`
	Category     *domain.ReportCategory `db:"category" json:"category,omitempty"`
	TemplatePath *string                `db:"template_path" json:"template_path,omitempty"`
	OutputFormat domain.OutputFormat    `db:"output_format" json:"output_format"`
	IsPeriodic   bool                   `db:"is_periodic" json:"is_periodic"`
	PeriodType   *domain.PeriodType     `db:"period_type" json:"period_type,omitempty"`
	CreatedAt    time.Time              `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time              `db:"updated_at" json:"updated_at"`

	// Associations (not stored in DB, loaded separately)
	Parameters []ReportParameter `db:"-" json:"parameters,omitempty"`
	Templates  []ReportTemplate  `db:"-" json:"templates,omitempty"`
}

// NewReportType creates a new report type
func NewReportType(name, code string, outputFormat domain.OutputFormat) *ReportType {
	now := time.Now()
	return &ReportType{
		Name:         name,
		Code:         code,
		OutputFormat: outputFormat,
		IsPeriodic:   false,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// SetPeriodic sets the report type as periodic with specified period
func (rt *ReportType) SetPeriodic(periodType domain.PeriodType) {
	rt.IsPeriodic = true
	rt.PeriodType = &periodType
	rt.UpdatedAt = time.Now()
}

// SetCategory sets the category of the report type
func (rt *ReportType) SetCategory(category domain.ReportCategory) {
	rt.Category = &category
	rt.UpdatedAt = time.Now()
}

// ReportParameter represents a parameter definition for a report type
// Aligned with migrations/006_create_reports_schema.up.sql - report_parameters table
type ReportParameter struct {
	ID            int64                `db:"id" json:"id"`
	ReportTypeID  int64                `db:"report_type_id" json:"report_type_id"`
	ParameterName string               `db:"parameter_name" json:"parameter_name"`
	ParameterType domain.ParameterType `db:"parameter_type" json:"parameter_type"`
	IsRequired    bool                 `db:"is_required" json:"is_required"`
	DefaultValue  *string              `db:"default_value" json:"default_value,omitempty"`
	Options       json.RawMessage      `db:"options" json:"options,omitempty"`
	DisplayOrder  int                  `db:"display_order" json:"display_order"`
	CreatedAt     time.Time            `db:"created_at" json:"created_at"`
}

// NewReportParameter creates a new report parameter
func NewReportParameter(reportTypeID int64, name string, paramType domain.ParameterType, required bool) *ReportParameter {
	return &ReportParameter{
		ReportTypeID:  reportTypeID,
		ParameterName: name,
		ParameterType: paramType,
		IsRequired:    required,
		DisplayOrder:  0,
		CreatedAt:     time.Now(),
	}
}

// SetOptions sets the options for select/multiselect parameters
func (rp *ReportParameter) SetOptions(options any) error {
	data, err := json.Marshal(options)
	if err != nil {
		return err
	}
	rp.Options = data
	return nil
}

// GetOptions unmarshals options into the provided target
func (rp *ReportParameter) GetOptions(target any) error {
	if rp.Options == nil {
		return nil
	}
	return json.Unmarshal(rp.Options, target)
}

// ReportTemplate represents a template for generating reports
// Aligned with migrations/006_create_reports_schema.up.sql - report_templates table
type ReportTemplate struct {
	ID           int64     `db:"id" json:"id"`
	ReportTypeID int64     `db:"report_type_id" json:"report_type_id"`
	Name         string    `db:"name" json:"name"`
	Content      string    `db:"content" json:"content"`
	IsDefault    bool      `db:"is_default" json:"is_default"`
	CreatedBy    int64     `db:"created_by" json:"created_by"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

// NewReportTemplate creates a new report template
func NewReportTemplate(reportTypeID int64, name, content string, createdBy int64) *ReportTemplate {
	now := time.Now()
	return &ReportTemplate{
		ReportTypeID: reportTypeID,
		Name:         name,
		Content:      content,
		IsDefault:    false,
		CreatedBy:    createdBy,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// SetAsDefault marks this template as the default for its report type
func (rt *ReportTemplate) SetAsDefault() {
	rt.IsDefault = true
	rt.UpdatedAt = time.Now()
}
