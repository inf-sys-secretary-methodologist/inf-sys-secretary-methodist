// Package handlers contains HTTP request handlers for the users module.
package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/sanitization"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/validation"
)

// DepartmentHandler handles HTTP requests for department management.
type DepartmentHandler struct {
	usecase   *usecases.DepartmentUseCase
	validator *validation.Validator
	sanitizer *sanitization.Sanitizer
}

// NewDepartmentHandler creates a new department handler.
func NewDepartmentHandler(usecase *usecases.DepartmentUseCase) *DepartmentHandler {
	return &DepartmentHandler{
		usecase:   usecase,
		validator: validation.NewValidator(),
		sanitizer: sanitization.NewSanitizer(),
	}
}

// Create handles POST /api/departments - creates a new department.
func (h *DepartmentHandler) Create(c *gin.Context) {
	var input dto.CreateDepartmentInput
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
	department, err := h.usecase.CreateDepartment(ctx, &input)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	c.JSON(http.StatusCreated, response.Success(department))
}

// List handles GET /api/departments - lists all departments.
func (h *DepartmentHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	activeOnly := c.DefaultQuery("active_only", "false") == "true"

	ctx := c.Request.Context()
	result, err := h.usecase.ListDepartments(ctx, page, limit, activeOnly)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	c.JSON(http.StatusOK, response.Success(result))
}

// GetByID handles GET /api/departments/:id - gets a single department.
func (h *DepartmentHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID подразделения")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	department, err := h.usecase.GetDepartment(ctx, id)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	c.JSON(http.StatusOK, response.Success(department))
}

// Update handles PUT /api/departments/:id - updates a department.
func (h *DepartmentHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID подразделения")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	var input dto.UpdateDepartmentInput
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
	department, err := h.usecase.UpdateDepartment(ctx, id, &input)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	c.JSON(http.StatusOK, response.Success(department))
}

// Delete handles DELETE /api/departments/:id - deletes a department.
func (h *DepartmentHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID подразделения")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	if err := h.usecase.DeleteDepartment(ctx, id); err != nil {
		// Check for specific error type
		var deptErr *usecases.DepartmentHasChildrenError
		if errors.As(err, &deptErr) {
			resp := response.BadRequest("Подразделение имеет дочерние подразделения и не может быть удалено")
			c.JSON(http.StatusBadRequest, resp)
			return
		}
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"message": "Подразделение удалено"}))
}

// GetChildren handles GET /api/departments/:id/children - gets child departments.
func (h *DepartmentHandler) GetChildren(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID подразделения")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	children, err := h.usecase.GetDepartmentChildren(ctx, id)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"departments": children}))
}
