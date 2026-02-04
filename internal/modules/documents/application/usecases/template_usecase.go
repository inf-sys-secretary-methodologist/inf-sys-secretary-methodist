// Package usecases contains application use cases for the documents module.
package usecases

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

// TemplateRepository defines the interface for template operations
type TemplateRepository interface {
	GetAll(ctx context.Context) ([]entities.DocumentType, error)
	GetByID(ctx context.Context, id int64) (*entities.DocumentType, error)
	UpdateTemplate(ctx context.Context, id int64, content *string, variables []entities.TemplateVariable) error
}

// TemplateUseCase handles template-related business logic
type TemplateUseCase struct {
	templateRepo TemplateRepository
	docRepo      repositories.DocumentRepository
	auditLogger  *logging.AuditLogger
}

// NewTemplateUseCase creates a new TemplateUseCase
func NewTemplateUseCase(
	templateRepo TemplateRepository,
	docRepo repositories.DocumentRepository,
	auditLogger *logging.AuditLogger,
) *TemplateUseCase {
	return &TemplateUseCase{
		templateRepo: templateRepo,
		docRepo:      docRepo,
		auditLogger:  auditLogger,
	}
}

// GetAllTemplates returns all document types with templates
func (uc *TemplateUseCase) GetAllTemplates(ctx context.Context) (*dto.TemplateListResponse, error) {
	types, err := uc.templateRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get templates: %w", err)
	}

	return dto.ToTemplateListResponse(types), nil
}

// GetTemplate returns a specific template by document type ID
func (uc *TemplateUseCase) GetTemplate(ctx context.Context, typeID int64) (*dto.TemplateResponse, error) {
	docType, err := uc.templateRepo.GetByID(ctx, typeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	return dto.ToTemplateResponse(docType), nil
}

// PreviewTemplate renders a template with given variables
func (uc *TemplateUseCase) PreviewTemplate(ctx context.Context, typeID int64, variables map[string]string) (*dto.PreviewTemplateResponse, error) {
	docType, err := uc.templateRepo.GetByID(ctx, typeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	if docType.TemplateContent == nil || *docType.TemplateContent == "" {
		return nil, fmt.Errorf("template not found for document type %d", typeID)
	}

	// Validate required variables
	if err := uc.validateVariables(docType.TemplateVariables, variables); err != nil {
		return nil, err
	}

	// Render template
	content := uc.renderTemplate(*docType.TemplateContent, variables)

	return &dto.PreviewTemplateResponse{
		Content: content,
	}, nil
}

// CreateDocumentFromTemplate creates a new document from a template
func (uc *TemplateUseCase) CreateDocumentFromTemplate(
	ctx context.Context,
	typeID int64,
	req *dto.CreateFromTemplateRequest,
	authorID int64,
) (*entities.Document, error) {
	docType, err := uc.templateRepo.GetByID(ctx, typeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	if docType.TemplateContent == nil || *docType.TemplateContent == "" {
		return nil, fmt.Errorf("template not found for document type %d", typeID)
	}

	// Validate required variables
	if err := uc.validateVariables(docType.TemplateVariables, req.Variables); err != nil {
		return nil, err
	}

	// Render template
	content := uc.renderTemplate(*docType.TemplateContent, req.Variables)

	// Extract subject from variables if present
	var subject *string
	if s, ok := req.Variables["subject"]; ok {
		subject = &s
	}

	// Create document
	doc := &entities.Document{
		DocumentTypeID: typeID,
		CategoryID:     req.CategoryID,
		Title:          req.Title,
		Subject:        subject,
		Content:        &content,
		AuthorID:       authorID,
		Status:         entities.DocumentStatusDraft,
		Version:        1,
	}

	err = uc.docRepo.Create(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	// Log audit
	if uc.auditLogger != nil {
		uc.auditLogger.LogAuditEvent(ctx, "create", "documents", map[string]interface{}{
			"document_id":      doc.ID,
			"title":            req.Title,
			"document_type_id": typeID,
			"from_template":    true,
		})
	}

	return doc, nil
}

// UpdateTemplate updates a document type's template
func (uc *TemplateUseCase) UpdateTemplate(
	ctx context.Context,
	typeID int64,
	req *dto.UpdateTemplateRequest,
) error {
	err := uc.templateRepo.UpdateTemplate(ctx, typeID, req.TemplateContent, req.TemplateVariables)
	if err != nil {
		return fmt.Errorf("failed to update template: %w", err)
	}

	// Log audit
	if uc.auditLogger != nil {
		uc.auditLogger.LogAuditEvent(ctx, "update", "document_types", map[string]interface{}{
			"document_type_id": typeID,
			"template_updated": true,
		})
	}

	return nil
}

// validateVariables validates that all required variables are provided
func (uc *TemplateUseCase) validateVariables(definitions []entities.TemplateVariable, values map[string]string) error {
	for _, def := range definitions {
		if def.Required {
			val, ok := values[def.Name]
			if !ok || strings.TrimSpace(val) == "" {
				return fmt.Errorf("required variable '%s' (%s) is missing", def.Name, def.Label)
			}
		}
	}
	return nil
}

// renderTemplate replaces template variables with provided values
func (uc *TemplateUseCase) renderTemplate(template string, variables map[string]string) string {
	result := template

	// Replace all {{variable}} patterns with values
	re := regexp.MustCompile(`\{\{(\w+)\}\}`)
	result = re.ReplaceAllStringFunc(result, func(match string) string {
		// Extract variable name from {{name}}
		varName := strings.Trim(match, "{}")
		if val, ok := variables[varName]; ok {
			return val
		}
		// Keep original placeholder if no value provided
		return match
	})

	return result
}
