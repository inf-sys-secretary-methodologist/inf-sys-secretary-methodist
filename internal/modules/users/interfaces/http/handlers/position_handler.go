// Package handlers contains HTTP request handlers for the users module.
package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/sanitization"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/validation"
)

// PositionHandler handles HTTP requests for position management.
type PositionHandler struct {
	usecase   *usecases.PositionUseCase
	validator *validation.Validator
	sanitizer *sanitization.Sanitizer
}

// NewPositionHandler creates a new position handler.
func NewPositionHandler(usecase *usecases.PositionUseCase) *PositionHandler {
	return &PositionHandler{
		usecase:   usecase,
		validator: validation.NewValidator(),
		sanitizer: sanitization.NewSanitizer(),
	}
}

// Create handles POST /api/positions - creates a new position.
func (h *PositionHandler) Create(c *gin.Context) {
	var input dto.CreatePositionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		resp := response.BadRequest("Неверный формат запроса")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// Sanitize inputs
	input.Name = h.sanitizer.SanitizeString(input.Name)
	input.Code = h.sanitizer.SanitizeString(input.Code)
	input.Description = h.sanitizer.SanitizeString(input.Description)

	if err := h.validator.Validate(input); err != nil {
		resp := response.BadRequest(err.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	position, err := h.usecase.CreatePosition(ctx, &input)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	c.JSON(http.StatusCreated, response.Success(position))
}

// List handles GET /api/positions - lists all positions.
func (h *PositionHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	activeOnly := c.DefaultQuery("active_only", "false") == "true"

	ctx := c.Request.Context()
	result, err := h.usecase.ListPositions(ctx, page, limit, activeOnly)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	c.JSON(http.StatusOK, response.Success(result))
}

// GetByID handles GET /api/positions/:id - gets a single position.
func (h *PositionHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID должности")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	position, err := h.usecase.GetPosition(ctx, id)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	c.JSON(http.StatusOK, response.Success(position))
}

// Update handles PUT /api/positions/:id - updates a position.
func (h *PositionHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID должности")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	var input dto.UpdatePositionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		resp := response.BadRequest("Неверный формат запроса")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// Sanitize inputs
	input.Name = h.sanitizer.SanitizeString(input.Name)
	input.Code = h.sanitizer.SanitizeString(input.Code)
	input.Description = h.sanitizer.SanitizeString(input.Description)

	if err := h.validator.Validate(input); err != nil {
		resp := response.BadRequest(err.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	position, err := h.usecase.UpdatePosition(ctx, id, &input)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	c.JSON(http.StatusOK, response.Success(position))
}

// Delete handles DELETE /api/positions/:id - deletes a position.
func (h *PositionHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID должности")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	if err := h.usecase.DeletePosition(ctx, id); err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"message": "Должность удалена"}))
}
