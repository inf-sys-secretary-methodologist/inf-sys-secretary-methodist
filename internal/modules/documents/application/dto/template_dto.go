// Package dto contains Data Transfer Objects for the documents module.
package dto

import "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"

// TemplateListResponse represents a list of available templates
type TemplateListResponse struct {
	Templates []TemplateResponse `json:"templates"`
	Total     int                `json:"total"`
}

// TemplateResponse represents a document template
type TemplateResponse struct {
	ID                int64                        `json:"id"`
	Name              string                       `json:"name"`
	Code              string                       `json:"code"`
	Description       *string                      `json:"description,omitempty"`
	TemplateContent   *string                      `json:"template_content,omitempty"`
	TemplateVariables []entities.TemplateVariable  `json:"template_variables,omitempty"`
	HasTemplate       bool                         `json:"has_template"`
}

// CreateFromTemplateRequest represents a request to create a document from a template
type CreateFromTemplateRequest struct {
	Title      string            `json:"title" validate:"required,min=1,max=500"`
	Variables  map[string]string `json:"variables" validate:"required"`
	CategoryID *int64            `json:"category_id,omitempty"`
}

// PreviewTemplateRequest represents a request to preview a template with variables
type PreviewTemplateRequest struct {
	Variables map[string]string `json:"variables" validate:"required"`
}

// PreviewTemplateResponse represents a preview of the rendered template
type PreviewTemplateResponse struct {
	Content string `json:"content"`
}

// UpdateTemplateRequest represents a request to update a document type's template
type UpdateTemplateRequest struct {
	TemplateContent   *string                      `json:"template_content,omitempty"`
	TemplateVariables []entities.TemplateVariable  `json:"template_variables,omitempty"`
}

// ToTemplateResponse converts a DocumentType entity to TemplateResponse DTO
func ToTemplateResponse(dt *entities.DocumentType) *TemplateResponse {
	return &TemplateResponse{
		ID:                dt.ID,
		Name:              dt.Name,
		Code:              dt.Code,
		Description:       dt.Description,
		TemplateContent:   dt.TemplateContent,
		TemplateVariables: dt.TemplateVariables,
		HasTemplate:       dt.TemplateContent != nil && *dt.TemplateContent != "",
	}
}

// ToTemplateListResponse converts a slice of DocumentType entities to TemplateListResponse
func ToTemplateListResponse(types []entities.DocumentType) *TemplateListResponse {
	templates := make([]TemplateResponse, 0, len(types))
	for _, dt := range types {
		// Only include types that have templates
		if dt.TemplateContent != nil && *dt.TemplateContent != "" {
			templates = append(templates, *ToTemplateResponse(&dt))
		}
	}
	return &TemplateListResponse{
		Templates: templates,
		Total:     len(templates),
	}
}
