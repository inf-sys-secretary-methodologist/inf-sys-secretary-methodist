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

// UserHandler handles HTTP requests for user management.
type UserHandler struct {
	usecase   *usecases.UserUseCase
	validator *validation.Validator
	sanitizer *sanitization.Sanitizer
}

// NewUserHandler creates a new user handler.
func NewUserHandler(usecase *usecases.UserUseCase) *UserHandler {
	return &UserHandler{
		usecase:   usecase,
		validator: validation.NewValidator(),
		sanitizer: sanitization.NewSanitizer(),
	}
}

// List handles GET /api/users - lists users with filtering and pagination.
func (h *UserHandler) List(c *gin.Context) {
	var filter dto.UserListFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		resp := response.BadRequest("Неверные параметры запроса")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// Sanitize search
	filter.Search = h.sanitizer.SanitizeString(filter.Search)

	ctx := c.Request.Context()
	result, err := h.usecase.ListUsers(ctx, &filter)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	c.JSON(http.StatusOK, response.Success(result))
}

// GetByID handles GET /api/users/:id - gets a single user.
func (h *UserHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID пользователя")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	user, err := h.usecase.GetUser(ctx, id)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	c.JSON(http.StatusOK, response.Success(user))
}

// UpdateProfile handles PUT /api/users/:id/profile - updates user profile.
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID пользователя")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	var input dto.UpdateUserProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		resp := response.BadRequest("Неверный формат запроса")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	// Sanitize inputs
	input.Phone = h.sanitizer.SanitizeString(input.Phone)
	input.Avatar = h.sanitizer.SanitizeString(input.Avatar)
	input.Bio = h.sanitizer.SanitizeString(input.Bio)

	if err := h.validator.Validate(input); err != nil {
		resp := response.BadRequest(err.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	if err := h.usecase.UpdateUserProfile(ctx, id, &input); err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"message": "Профиль обновлен"}))
}

// UpdateRole handles PUT /api/users/:id/role - updates user role.
func (h *UserHandler) UpdateRole(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID пользователя")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	var input dto.UpdateUserRoleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		resp := response.BadRequest("Неверный формат запроса")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	input.Role = h.sanitizer.SanitizeString(input.Role)

	if err := h.validator.Validate(input); err != nil {
		resp := response.BadRequest(err.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	if err := h.usecase.UpdateUserRole(ctx, id, &input); err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"message": "Роль обновлена"}))
}

// UpdateStatus handles PUT /api/users/:id/status - updates user status.
func (h *UserHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID пользователя")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	var input dto.UpdateUserStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		resp := response.BadRequest("Неверный формат запроса")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	input.Status = h.sanitizer.SanitizeString(input.Status)

	if err := h.validator.Validate(input); err != nil {
		resp := response.BadRequest(err.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	if err := h.usecase.UpdateUserStatus(ctx, id, &input); err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"message": "Статус обновлен"}))
}

// Delete handles DELETE /api/users/:id - deletes a user.
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID пользователя")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	if err := h.usecase.DeleteUser(ctx, id); err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"message": "Пользователь удален"}))
}

// BulkUpdateDepartment handles POST /api/users/bulk/department - bulk assigns users to department.
func (h *UserHandler) BulkUpdateDepartment(c *gin.Context) {
	var input dto.BulkUpdateDepartmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		resp := response.BadRequest("Неверный формат запроса")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	if err := h.validator.Validate(input); err != nil {
		resp := response.BadRequest(err.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	if err := h.usecase.BulkUpdateDepartment(ctx, &input); err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"message": "Подразделение обновлено"}))
}

// BulkUpdatePosition handles POST /api/users/bulk/position - bulk assigns users to position.
func (h *UserHandler) BulkUpdatePosition(c *gin.Context) {
	var input dto.BulkUpdatePositionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		resp := response.BadRequest("Неверный формат запроса")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	if err := h.validator.Validate(input); err != nil {
		resp := response.BadRequest(err.Error())
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	if err := h.usecase.BulkUpdatePosition(ctx, &input); err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"message": "Должность обновлена"}))
}

// GetByDepartment handles GET /api/users/by-department/:id - gets users by department.
func (h *UserHandler) GetByDepartment(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID подразделения")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	users, err := h.usecase.GetUsersByDepartment(ctx, id)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"users": users}))
}

// GetByPosition handles GET /api/users/by-position/:id - gets users by position.
func (h *UserHandler) GetByPosition(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID должности")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	ctx := c.Request.Context()
	users, err := h.usecase.GetUsersByPosition(ctx, id)
	if err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"users": users}))
}
