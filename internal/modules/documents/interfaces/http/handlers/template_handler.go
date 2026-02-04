// Package http provides HTTP handlers for the documents module.
package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/validation"
)

// TemplateHandler handles HTTP requests for document templates
type TemplateHandler struct {
	templateUseCase *usecases.TemplateUseCase
	validator       *validation.Validator
}

// NewTemplateHandler creates a new TemplateHandler
func NewTemplateHandler(templateUseCase *usecases.TemplateUseCase, validator *validation.Validator) *TemplateHandler {
	return &TemplateHandler{
		templateUseCase: templateUseCase,
		validator:       validator,
	}
}

// GetTemplates returns all available document templates
// @Summary Get all document templates
// @Description Returns a list of document types that have templates
// @Tags templates
// @Accept json
// @Produce json
// @Success 200 {object} dto.TemplateListResponse
// @Failure 500 {object} map[string]string
// @Router /api/templates [get]
// @Security BearerAuth
func (h *TemplateHandler) GetTemplates(c *gin.Context) {
	templates, err := h.templateUseCase.GetAllTemplates(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, templates)
}

// GetTemplate returns a specific document template
// @Summary Get a document template
// @Description Returns a specific document type's template by ID
// @Tags templates
// @Accept json
// @Produce json
// @Param id path int true "Document Type ID"
// @Success 200 {object} dto.TemplateResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/templates/{id} [get]
// @Security BearerAuth
func (h *TemplateHandler) GetTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template ID"})
		return
	}

	template, err := h.templateUseCase.GetTemplate(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, template)
}

// PreviewTemplate renders a template with provided variables
// @Summary Preview template
// @Description Renders a template with the provided variables without creating a document
// @Tags templates
// @Accept json
// @Produce json
// @Param id path int true "Document Type ID"
// @Param request body dto.PreviewTemplateRequest true "Template variables"
// @Success 200 {object} dto.PreviewTemplateResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/templates/{id}/preview [post]
// @Security BearerAuth
func (h *TemplateHandler) PreviewTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template ID"})
		return
	}

	var req dto.PreviewTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.validator.Validate(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	preview, err := h.templateUseCase.PreviewTemplate(c.Request.Context(), id, req.Variables)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, preview)
}

// CreateDocumentFromTemplate creates a new document from a template
// @Summary Create document from template
// @Description Creates a new document by filling a template with provided variables
// @Tags templates
// @Accept json
// @Produce json
// @Param id path int true "Document Type ID"
// @Param request body dto.CreateFromTemplateRequest true "Document creation data"
// @Success 201 {object} dto.DocumentResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/templates/{id}/create [post]
// @Security BearerAuth
func (h *TemplateHandler) CreateDocumentFromTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template ID"})
		return
	}

	var req dto.CreateFromTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.validator.Validate(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	doc, err := h.templateUseCase.CreateDocumentFromTemplate(c.Request.Context(), id, &req, userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.ToDocumentOutput(doc))
}

// UpdateTemplate updates a document type's template
// @Summary Update template
// @Description Updates the template content and variables for a document type
// @Tags templates
// @Accept json
// @Produce json
// @Param id path int true "Document Type ID"
// @Param request body dto.UpdateTemplateRequest true "Template data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/templates/{id} [put]
// @Security BearerAuth
func (h *TemplateHandler) UpdateTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template ID"})
		return
	}

	// Check admin role
	role, exists := c.Get("user_role")
	if !exists || (role != "admin" && role != "secretary") {
		c.JSON(http.StatusForbidden, gin.H{"error": "only admin or secretary can update templates"})
		return
	}

	var req dto.UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	err = h.templateUseCase.UpdateTemplate(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "template updated successfully"})
}
