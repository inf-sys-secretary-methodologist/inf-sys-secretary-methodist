// Package http contains HTTP request handlers for the documents module.
package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/sanitization"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/validation"
)

// TagHandler handles HTTP requests for tag endpoints.
type TagHandler struct {
	usecase   *usecases.TagUseCase
	validator *validation.Validator
	sanitizer *sanitization.Sanitizer
}

// NewTagHandler creates a new tag handler.
func NewTagHandler(usecase *usecases.TagUseCase) *TagHandler {
	return &TagHandler{
		usecase:   usecase,
		validator: validation.NewValidator(),
		sanitizer: sanitization.NewSanitizer(),
	}
}

// Create handles tag creation
func (h *TagHandler) Create(c *gin.Context) {
	var input dto.CreateTagInput
	if err := c.ShouldBindJSON(&input); err != nil {
		resp := response.BadRequest("Неверный формат запроса")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// Sanitize inputs
	input.Name = h.sanitizer.SanitizeString(input.Name)

	// Validate
	if err := h.validator.Validate(input); err != nil {
		resp := response.BadRequest(err.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	tag, err := h.usecase.Create(ctx, input)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	resp := response.Success(tag)
	c.JSON(http.StatusCreated, resp)
}

// GetByID handles getting a tag by ID
func (h *TagHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID тега")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	tag, err := h.usecase.GetByID(ctx, id)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	resp := response.Success(tag)
	c.JSON(http.StatusOK, resp)
}

// Update handles tag update
func (h *TagHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID тега")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	var input dto.UpdateTagInput
	if err := c.ShouldBindJSON(&input); err != nil {
		resp := response.BadRequest("Неверный формат запроса")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// Sanitize inputs
	if input.Name != nil {
		sanitized := h.sanitizer.SanitizeString(*input.Name)
		input.Name = &sanitized
	}

	// Validate
	if err := h.validator.Validate(input); err != nil {
		resp := response.BadRequest(err.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	tag, err := h.usecase.Update(ctx, id, input)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	resp := response.Success(tag)
	c.JSON(http.StatusOK, resp)
}

// Delete handles tag deletion
func (h *TagHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID тега")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	if err := h.usecase.Delete(ctx, id); err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	resp := response.Success(gin.H{"message": "Тег успешно удален"})
	c.JSON(http.StatusOK, resp)
}

// GetAll handles getting all tags
func (h *TagHandler) GetAll(c *gin.Context) {
	ctx := c.Request.Context()
	tags, err := h.usecase.GetAll(ctx)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	resp := response.Success(tags)
	c.JSON(http.StatusOK, resp)
}

// Search handles searching tags by name
func (h *TagHandler) Search(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		resp := response.BadRequest("Параметр поиска 'q' обязателен")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	limit := 10
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	ctx := c.Request.Context()
	tags, err := h.usecase.Search(ctx, query, limit)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	resp := response.Success(tags)
	c.JSON(http.StatusOK, resp)
}

// GetDocumentTags handles getting tags for a document
func (h *TagHandler) GetDocumentTags(c *gin.Context) {
	documentID, err := strconv.ParseInt(c.Param("document_id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID документа")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	result, err := h.usecase.GetDocumentTags(ctx, documentID)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	resp := response.Success(result)
	c.JSON(http.StatusOK, resp)
}

// SetDocumentTags handles setting tags for a document
func (h *TagHandler) SetDocumentTags(c *gin.Context) {
	documentID, err := strconv.ParseInt(c.Param("document_id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID документа")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	var input dto.SetDocumentTagsInput
	if err := c.ShouldBindJSON(&input); err != nil {
		resp := response.BadRequest("Неверный формат запроса")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	result, err := h.usecase.SetDocumentTags(ctx, documentID, input.TagIDs)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	resp := response.Success(result)
	c.JSON(http.StatusOK, resp)
}

// AddTagToDocument handles adding a tag to a document
func (h *TagHandler) AddTagToDocument(c *gin.Context) {
	documentID, err := strconv.ParseInt(c.Param("document_id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID документа")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	tagID, err := strconv.ParseInt(c.Param("tag_id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID тега")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	if err := h.usecase.AddTagToDocument(ctx, documentID, tagID); err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	resp := response.Success(gin.H{"message": "Тег добавлен к документу"})
	c.JSON(http.StatusOK, resp)
}

// RemoveTagFromDocument handles removing a tag from a document
func (h *TagHandler) RemoveTagFromDocument(c *gin.Context) {
	documentID, err := strconv.ParseInt(c.Param("document_id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID документа")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	tagID, err := strconv.ParseInt(c.Param("tag_id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID тега")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	if err := h.usecase.RemoveTagFromDocument(ctx, documentID, tagID); err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	resp := response.Success(gin.H{"message": "Тег удален из документа"})
	c.JSON(http.StatusOK, resp)
}

// GetDocumentsByTag handles getting documents by tag
func (h *TagHandler) GetDocumentsByTag(c *gin.Context) {
	tagID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID тега")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	page := 1
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	pageSize := 20
	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 {
			pageSize = parsed
		}
	}

	ctx := c.Request.Context()
	result, err := h.usecase.GetDocumentsByTag(ctx, tagID, page, pageSize)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	resp := response.Success(result)
	c.JSON(http.StatusOK, resp)
}
