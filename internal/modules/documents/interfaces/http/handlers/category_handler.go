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

// CategoryHandler handles HTTP requests for category endpoints.
type CategoryHandler struct {
	usecase   *usecases.CategoryUseCase
	validator *validation.Validator
	sanitizer *sanitization.Sanitizer
}

// NewCategoryHandler creates a new category handler.
func NewCategoryHandler(usecase *usecases.CategoryUseCase) *CategoryHandler {
	return &CategoryHandler{
		usecase:   usecase,
		validator: validation.NewValidator(),
		sanitizer: sanitization.NewSanitizer(),
	}
}

// Create handles category creation
func (h *CategoryHandler) Create(c *gin.Context) {
	var input dto.CreateCategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		resp := response.BadRequest("Неверный формат запроса")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// Sanitize inputs
	input.Name = h.sanitizer.SanitizeString(input.Name)
	if input.Description != nil {
		sanitized := h.sanitizer.SanitizeString(*input.Description)
		input.Description = &sanitized
	}

	// Validate
	if err := h.validator.Validate(input); err != nil {
		resp := response.BadRequest(err.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	category, err := h.usecase.Create(ctx, input)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	resp := response.Success(category)
	c.JSON(http.StatusCreated, resp)
}

// GetByID handles getting a category by ID
func (h *CategoryHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID категории")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	category, err := h.usecase.GetByID(ctx, id)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	resp := response.Success(category)
	c.JSON(http.StatusOK, resp)
}

// Update handles category update
func (h *CategoryHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID категории")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	var input dto.UpdateCategoryInput
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
	if input.Description != nil {
		sanitized := h.sanitizer.SanitizeString(*input.Description)
		input.Description = &sanitized
	}

	// Validate
	if err := h.validator.Validate(input); err != nil {
		resp := response.BadRequest(err.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	category, err := h.usecase.Update(ctx, id, input)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	resp := response.Success(category)
	c.JSON(http.StatusOK, resp)
}

// Delete handles category deletion
func (h *CategoryHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID категории")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	if err := h.usecase.Delete(ctx, id); err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	resp := response.Success(gin.H{"message": "Категория успешно удалена"})
	c.JSON(http.StatusOK, resp)
}

// GetAll handles getting all categories
func (h *CategoryHandler) GetAll(c *gin.Context) {
	ctx := c.Request.Context()
	categories, err := h.usecase.GetAll(ctx)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	resp := response.Success(categories)
	c.JSON(http.StatusOK, resp)
}

// GetTree handles getting the full category tree
func (h *CategoryHandler) GetTree(c *gin.Context) {
	ctx := c.Request.Context()
	tree, err := h.usecase.GetTree(ctx)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	resp := response.Success(tree)
	c.JSON(http.StatusOK, resp)
}

// GetChildren handles getting direct children of a category
func (h *CategoryHandler) GetChildren(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID категории")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	children, err := h.usecase.GetChildren(ctx, id)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	resp := response.Success(children)
	c.JSON(http.StatusOK, resp)
}

// GetRootCategories handles getting root categories
func (h *CategoryHandler) GetRootCategories(c *gin.Context) {
	ctx := c.Request.Context()
	categories, err := h.usecase.GetRootCategories(ctx)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	resp := response.Success(categories)
	c.JSON(http.StatusOK, resp)
}

// GetWithBreadcrumb handles getting a category with breadcrumb path
func (h *CategoryHandler) GetWithBreadcrumb(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID категории")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	result, err := h.usecase.GetWithBreadcrumb(ctx, id)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	resp := response.Success(result)
	c.JSON(http.StatusOK, resp)
}
