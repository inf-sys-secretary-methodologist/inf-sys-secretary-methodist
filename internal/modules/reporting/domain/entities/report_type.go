package entities

import (
	"encoding/json"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain"
)

// ReportType represents a type/template of report
type ReportType struct {
	ID           int64                  `json:"id"`
	Name         string                 `json:"name"`
	Code         string                 `json:"code"`
	Description  *string                `json:"description,omitempty"`
	Category     *domain.ReportCategory `json:"category,omitempty"`
	TemplatePath *string                `json:"template_path,omitempty"`
	OutputFormat domain.OutputFormat    `json:"output_format"`
	IsPeriodic   bool                   `json:"is_periodic"`
	PeriodType   *domain.PeriodType     `json:"period_type,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`

	// Associations (loaded separately)
	Parameters []ReportParameter `json:"parameters,omitempty"`
	Templates  []ReportTemplate  `json:"templates,omitempty"`
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
type ReportParameter struct {
	ID            int64                `json:"id"`
	ReportTypeID  int64                `json:"report_type_id"`
	ParameterName string               `json:"parameter_name"`
	ParameterType domain.ParameterType `json:"parameter_type"`
	IsRequired    bool                 `json:"is_required"`
	DefaultValue  *string              `json:"default_value,omitempty"`
	Options       json.RawMessage      `json:"options,omitempty"`
	DisplayOrder  int                  `json:"display_order"`
	CreatedAt     time.Time            `json:"created_at"`
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
type ReportTemplate struct {
	ID           int64     `json:"id"`
	ReportTypeID int64     `json:"report_type_id"`
	Name         string    `json:"name"`
	Content      string    `json:"content"`
	IsDefault    bool      `json:"is_default"`
	CreatedBy    int64     `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
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
