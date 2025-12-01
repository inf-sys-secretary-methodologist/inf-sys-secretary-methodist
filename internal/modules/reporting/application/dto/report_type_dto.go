// Package dto contains Data Transfer Objects for the reporting module.
package dto

import (
	"encoding/json"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain/entities"
)

// ReportTypeOutput represents output for a report type
type ReportTypeOutput struct {
	ID           int64                    `json:"id"`
	Name         string                   `json:"name"`
	Code         string                   `json:"code"`
	Description  *string                  `json:"description,omitempty"`
	Category     *string                  `json:"category,omitempty"`
	TemplatePath *string                  `json:"template_path,omitempty"`
	OutputFormat string                   `json:"output_format"`
	IsPeriodic   bool                     `json:"is_periodic"`
	PeriodType   *string                  `json:"period_type,omitempty"`
	Parameters   []*ReportParameterOutput `json:"parameters,omitempty"`
	Templates    []*ReportTemplateOutput  `json:"templates,omitempty"`
	CreatedAt    time.Time                `json:"created_at"`
	UpdatedAt    time.Time                `json:"updated_at"`
}

// ReportTypeListOutput represents paginated list of report types
type ReportTypeListOutput struct {
	ReportTypes []*ReportTypeOutput `json:"report_types"`
	Total       int64               `json:"total"`
	Page        int                 `json:"page"`
	PageSize    int                 `json:"page_size"`
	TotalPages  int                 `json:"total_pages"`
}

// ReportTypeFilterInput represents filter options for listing report types
type ReportTypeFilterInput struct {
	Category   *string `form:"category" validate:"omitempty,oneof=academic administrative financial methodical"`
	IsPeriodic *bool   `form:"is_periodic"`
	Page       int     `form:"page,default=1"`
	PageSize   int     `form:"page_size,default=20"`
}

// ReportParameterOutput represents output for a report parameter
type ReportParameterOutput struct {
	ID            int64       `json:"id"`
	ReportTypeID  int64       `json:"report_type_id"`
	ParameterName string      `json:"parameter_name"`
	ParameterType string      `json:"parameter_type"`
	IsRequired    bool        `json:"is_required"`
	DefaultValue  *string     `json:"default_value,omitempty"`
	Options       interface{} `json:"options,omitempty"`
	DisplayOrder  int         `json:"display_order"`
	CreatedAt     time.Time   `json:"created_at"`
}

// ReportTemplateOutput represents output for a report template
type ReportTemplateOutput struct {
	ID           int64     `json:"id"`
	ReportTypeID int64     `json:"report_type_id"`
	Name         string    `json:"name"`
	Content      string    `json:"content,omitempty"`
	IsDefault    bool      `json:"is_default"`
	CreatedBy    int64     `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ReportSubscriptionOutput represents output for a report subscription
type ReportSubscriptionOutput struct {
	ID             int64     `json:"id"`
	ReportTypeID   int64     `json:"report_type_id"`
	ReportTypeName string    `json:"report_type_name,omitempty"`
	UserID         int64     `json:"user_id"`
	DeliveryMethod string    `json:"delivery_method"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
}

// SubscribeInput represents input for subscribing to a report type
type SubscribeInput struct {
	DeliveryMethod string `json:"delivery_method" validate:"required,oneof=email notification both"`
}

// UpdateSubscriptionInput represents input for updating a subscription
type UpdateSubscriptionInput struct {
	DeliveryMethod *string `json:"delivery_method,omitempty" validate:"omitempty,oneof=email notification both"`
	IsActive       *bool   `json:"is_active,omitempty"`
}

// ToReportTypeOutput converts a ReportType entity to DTO
func ToReportTypeOutput(rt *entities.ReportType) *ReportTypeOutput {
	output := &ReportTypeOutput{
		ID:           rt.ID,
		Name:         rt.Name,
		Code:         rt.Code,
		Description:  rt.Description,
		TemplatePath: rt.TemplatePath,
		OutputFormat: string(rt.OutputFormat),
		IsPeriodic:   rt.IsPeriodic,
		CreatedAt:    rt.CreatedAt,
		UpdatedAt:    rt.UpdatedAt,
	}

	if rt.Category != nil {
		cat := string(*rt.Category)
		output.Category = &cat
	}
	if rt.PeriodType != nil {
		pt := string(*rt.PeriodType)
		output.PeriodType = &pt
	}

	// Convert parameters if loaded
	if len(rt.Parameters) > 0 {
		output.Parameters = make([]*ReportParameterOutput, len(rt.Parameters))
		for i, p := range rt.Parameters {
			output.Parameters[i] = ToReportParameterOutput(&p)
		}
	}

	// Convert templates if loaded
	if len(rt.Templates) > 0 {
		output.Templates = make([]*ReportTemplateOutput, len(rt.Templates))
		for i, t := range rt.Templates {
			output.Templates[i] = ToReportTemplateOutput(&t)
		}
	}

	return output
}

// ToReportParameterOutput converts a ReportParameter entity to DTO
func ToReportParameterOutput(param *entities.ReportParameter) *ReportParameterOutput {
	output := &ReportParameterOutput{
		ID:            param.ID,
		ReportTypeID:  param.ReportTypeID,
		ParameterName: param.ParameterName,
		ParameterType: string(param.ParameterType),
		IsRequired:    param.IsRequired,
		DefaultValue:  param.DefaultValue,
		DisplayOrder:  param.DisplayOrder,
		CreatedAt:     param.CreatedAt,
	}

	if param.Options != nil {
		var options interface{}
		if err := json.Unmarshal(param.Options, &options); err == nil {
			output.Options = options
		}
	}

	return output
}

// ToReportTemplateOutput converts a ReportTemplate entity to DTO
func ToReportTemplateOutput(template *entities.ReportTemplate) *ReportTemplateOutput {
	return &ReportTemplateOutput{
		ID:           template.ID,
		ReportTypeID: template.ReportTypeID,
		Name:         template.Name,
		Content:      template.Content,
		IsDefault:    template.IsDefault,
		CreatedBy:    template.CreatedBy,
		CreatedAt:    template.CreatedAt,
		UpdatedAt:    template.UpdatedAt,
	}
}

// ToReportSubscriptionOutput converts a ReportSubscription entity to DTO
func ToReportSubscriptionOutput(sub *entities.ReportSubscription) *ReportSubscriptionOutput {
	return &ReportSubscriptionOutput{
		ID:             sub.ID,
		ReportTypeID:   sub.ReportTypeID,
		UserID:         sub.UserID,
		DeliveryMethod: string(sub.DeliveryMethod),
		IsActive:       sub.IsActive,
		CreatedAt:      sub.CreatedAt,
	}
}
