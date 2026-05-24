// Package handlers contains HTTP request handlers for the users module.
package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	authDomain "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/sanitization"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/validation"
)

// asString safely coerces an interface{} into a string, returning ""
// if the value is nil or not a string. Used for reading role-like
// claims from the gin context without a panic on absent/mistyped
// values.
func asString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

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

	// Authorization: actor must be the target user (self-edit) OR
	// system_admin (override). The usecase enforces the rule via the
	// domain free function; the handler just supplies the inputs from
	// the JWT-bound gin context. Closes #283 ADR-1 (TIER 0 profile
	// takeover).
	actorID, exists := c.Get("user_id")
	if !exists {
		resp := response.Unauthorized("Требуется авторизация")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}
	actorIDInt, _ := actorID.(int64)
	actorRoleStr, _ := c.Get("role")
	actorRole := authDomain.RoleType(asString(actorRoleStr))

	ctx := c.Request.Context()
	if err := h.usecase.UpdateUserProfile(ctx, actorIDInt, actorRole, id, &input); err != nil {
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

	// Actor identity required for audit forensic invariant (v0.160.1
	// Item 6 — role change must record who triggered it).
	actorIDRaw, exists := c.Get("user_id")
	if !exists {
		resp := response.Unauthorized("Требуется авторизация")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}
	actorID, _ := actorIDRaw.(int64)

	ctx := c.Request.Context()
	if err := h.usecase.UpdateUserRole(ctx, actorID, id, &input); err != nil {
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

	// Actor identity required by usecase last-admin guard (#283 ADR-4).
	actorIDRaw, exists := c.Get("user_id")
	if !exists {
		resp := response.Unauthorized("Требуется авторизация")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}
	actorID, _ := actorIDRaw.(int64)

	ctx := c.Request.Context()
	if err := h.usecase.UpdateUserStatus(ctx, actorID, id, &input); err != nil {
		httpErr := response.MapDomainError(err)
		c.JSON(httpErr.Status, httpErr.Response)
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"message": "Статус обновлен"}))
}

// Delete handles DELETE /api/users/:id - deletes a user.
//
// The route group already enforces RequireRole(system_admin), so any
// caller reaching this handler is an admin. The usecase still runs
// the #283 ADR-4 guards: actor must not delete itself, and removing
// the last system_admin is rejected to keep the recovery path open.
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		resp := response.BadRequest("Неверный ID пользователя")
		c.JSON(http.StatusBadRequest, resp)
		return
	}

	actorIDRaw, exists := c.Get("user_id")
	if !exists {
		resp := response.Unauthorized("Требуется авторизация")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}
	actorID, _ := actorIDRaw.(int64)

	ctx := c.Request.Context()
	if err := h.usecase.DeleteUser(ctx, actorID, id); err != nil {
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

	// Actor identity for audit forensic invariant (v0.160.1 Item 6).
	actorIDRaw, exists := c.Get("user_id")
	if !exists {
		resp := response.Unauthorized("Требуется авторизация")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}
	actorID, _ := actorIDRaw.(int64)

	ctx := c.Request.Context()
	if err := h.usecase.BulkUpdateDepartment(ctx, actorID, &input); err != nil {
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

	// Actor identity for audit forensic invariant (v0.160.1 Item 6).
	actorIDRaw, exists := c.Get("user_id")
	if !exists {
		resp := response.Unauthorized("Требуется авторизация")
		c.JSON(http.StatusUnauthorized, resp)
		return
	}
	actorID, _ := actorIDRaw.(int64)

	ctx := c.Request.Context()
	if err := h.usecase.BulkUpdatePosition(ctx, actorID, &input); err != nil {
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
