// Package handlers exposes the student_debts read endpoints over HTTP.
// Handlers parse the request, delegate to a use-case port and map the
// result/error — no business logic lives here.
package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	authDomain "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
)

// Pagination bounds applied when the client omits or over-asks limit.
const (
	defaultPageLimit = 50
	maxPageLimit     = 200
)

// GetDebtPort is the read port for a single debt (with scope checks).
type GetDebtPort interface {
	Execute(ctx context.Context, actorID int64, actorRole string, id int64) (*entities.StudentDebt, error)
}

// ListDebtsPort is the registry-list read port (staff + teacher scope).
type ListDebtsPort interface {
	Execute(ctx context.Context, actorID int64, actorRole string, filter repositories.StudentDebtListFilter) (repositories.StudentDebtListResult, error)
}

// ListMyDebtsPort is the student "my debts" read port (forces ownership).
type ListMyDebtsPort interface {
	Execute(ctx context.Context, actorID int64, filter repositories.StudentDebtListFilter) (repositories.StudentDebtListResult, error)
}

// GetDebtStatsPort is the dashboard-aggregate read port.
type GetDebtStatsPort interface {
	Execute(ctx context.Context, actorID int64, actorRole string, filter repositories.StudentDebtListFilter) (repositories.StudentDebtStats, error)
}

// StudentDebtHandler serves the student_debts read endpoints.
type StudentDebtHandler struct {
	get    GetDebtPort
	list   ListDebtsPort
	listMy ListMyDebtsPort
	stats  GetDebtStatsPort
}

// NewStudentDebtHandler wires the handler. All ports are required;
// constructing with a nil port panics (failure-closed DI).
func NewStudentDebtHandler(get GetDebtPort, list ListDebtsPort, listMy ListMyDebtsPort, stats GetDebtStatsPort) *StudentDebtHandler {
	if get == nil || list == nil || listMy == nil || stats == nil {
		panic("student_debts: NewStudentDebtHandler requires non-nil ports")
	}
	return &StudentDebtHandler{get: get, list: list, listMy: listMy, stats: stats}
}

// RegisterStudentDebtRoutes mounts the read endpoints under /student-debts.
// The caller applies authentication middleware to rg first. Static segments
// (/stats, /my) are registered before /:id so they take routing priority.
func RegisterStudentDebtRoutes(rg *gin.RouterGroup, h *StudentDebtHandler) {
	g := rg.Group("/student-debts")
	g.GET("", h.List)
	g.GET("/stats", h.Stats)
	g.GET("/my", h.My)
	g.GET("/:id", h.Get)
}

// List handles GET /student-debts — the role-scoped registry page.
func (h *StudentDebtHandler) List(c *gin.Context) {
	actorID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	res, err := h.list.Execute(c.Request.Context(), actorID, role, parseListFilter(c))
	if err != nil {
		// List denial is role-based (no specific resource to hide) — a true 403.
		mapDebtError(c, err, false)
		return
	}
	c.JSON(http.StatusOK, response.Success(mapList(res)))
}

// Get handles GET /student-debts/:id — a single debt with its attempts.
func (h *StudentDebtHandler) Get(c *gin.Context) {
	actorID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	id, ok := parsePositiveID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid student debt id"))
		return
	}
	debt, err := h.get.Execute(c.Request.Context(), actorID, role, id)
	if err != nil {
		// Non-managers get scope-forbidden collapsed to 404 (OWASP IDOR):
		// never reveal a debt's existence to someone who may not see it.
		mapDebtError(c, err, !isDebtManagerRole(role))
		return
	}
	c.JSON(http.StatusOK, response.Success(mapDebt(debt)))
}

// My handles GET /student-debts/my — the caller's own debts.
func (h *StudentDebtHandler) My(c *gin.Context) {
	actorID, _, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	res, err := h.listMy.Execute(c.Request.Context(), actorID, parseListFilter(c))
	if err != nil {
		mapDebtError(c, err, false)
		return
	}
	c.JSON(http.StatusOK, response.Success(mapList(res)))
}

// Stats handles GET /student-debts/stats — the dashboard aggregate.
func (h *StudentDebtHandler) Stats(c *gin.Context) {
	actorID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	res, err := h.stats.Execute(c.Request.Context(), actorID, role, parseListFilter(c))
	if err != nil {
		mapDebtError(c, err, false)
		return
	}
	c.JSON(http.StatusOK, response.Success(mapStats(res)))
}

// --- helpers ----------------------------------------------------------------

// authContext extracts the authenticated actor id + role from the gin
// context (set by auth middleware). Failure-closed: any missing/ill-typed
// value yields ok=false so the handler returns 401.
func authContext(c *gin.Context) (userID int64, role string, ok bool) {
	uid, exists := c.Get("user_id")
	if !exists {
		return 0, "", false
	}
	switch v := uid.(type) {
	case int64:
		userID = v
	case int:
		userID = int64(v)
	case float64:
		userID = int64(v)
	default:
		return 0, "", false
	}
	roleVal, exists := c.Get("role")
	if !exists {
		return 0, "", false
	}
	role, _ = roleVal.(string)
	if role == "" {
		return 0, "", false
	}
	return userID, role, true
}

// parsePositiveID parses a strictly-positive int64 path parameter.
func parsePositiveID(raw string) (int64, bool) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

// isDebtManagerRole reports whether the role has unrestricted registry read
// (admin/methodist/secretary) — mirrors the use-case isDebtManager so the
// handler can decide whether a scope denial is an IDOR-collapse 404.
func isDebtManagerRole(role string) bool {
	switch authDomain.RoleType(role) {
	case authDomain.RoleSystemAdmin, authDomain.RoleMethodist, authDomain.RoleAcademicSecretary:
		return true
	default:
		return false
	}
}

// parseListFilter builds the registry filter from query parameters, applying
// pagination bounds (a zero limit would otherwise mean SQL LIMIT 0).
func parseListFilter(c *gin.Context) repositories.StudentDebtListFilter {
	f := repositories.StudentDebtListFilter{GroupName: c.Query("group_name")}
	if s := c.Query("status"); s != "" {
		st := entities.DebtStatus(s)
		f.Status = &st
	}
	if v := c.Query("semester"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			f.Semester = &n
		}
	}
	f.Limit, _ = strconv.Atoi(c.Query("limit"))
	switch {
	case f.Limit <= 0:
		f.Limit = defaultPageLimit
	case f.Limit > maxPageLimit:
		f.Limit = maxPageLimit
	}
	if f.Offset, _ = strconv.Atoi(c.Query("offset")); f.Offset < 0 {
		f.Offset = 0
	}
	return f
}

// mapDebtError maps domain/repository sentinels to HTTP status. When
// hideForbiddenAsNotFound is set, a scope denial is reported as 404 instead
// of 403 (IDOR-collapse for callers who must not learn the debt exists).
func mapDebtError(c *gin.Context, err error, hideForbiddenAsNotFound bool) {
	switch {
	case errors.Is(err, repositories.ErrStudentDebtNotFound):
		c.JSON(http.StatusNotFound, response.NotFound("student debt"))
	case errors.Is(err, entities.ErrDebtAccessForbidden):
		if hideForbiddenAsNotFound {
			c.JSON(http.StatusNotFound, response.NotFound("student debt"))
		} else {
			c.JSON(http.StatusForbidden, response.Forbidden("not authorized for this debt registry operation"))
		}
	case errors.Is(err, repositories.ErrStudentDebtVersionConflict):
		c.JSON(http.StatusConflict, response.ErrorResponse("VERSION_CONFLICT", "the debt was modified concurrently; reload and retry"))
	case errors.Is(err, repositories.ErrStudentDebtIdentityExists):
		c.JSON(http.StatusConflict, response.ErrorResponse("IDENTITY_EXISTS", "a debt with this identity already exists"))
	case errors.Is(err, entities.ErrDebtClosed):
		c.JSON(http.StatusConflict, response.ErrorResponse("DEBT_CLOSED", "the debt is closed; no further resits can be scheduled"))
	case errors.Is(err, entities.ErrNoScheduledResit):
		c.JSON(http.StatusConflict, response.ErrorResponse("NO_SCHEDULED_RESIT", "there is no scheduled resit to record on this attempt"))
	case errors.Is(err, entities.ErrAttemptAlreadyRecorded):
		c.JSON(http.StatusConflict, response.ErrorResponse("ALREADY_RECORDED", "this attempt's result has already been recorded"))
	case errors.Is(err, entities.ErrInvalidTransition):
		c.JSON(http.StatusConflict, response.ErrorResponse("INVALID_TRANSITION", "the requested transition is not allowed in the current state"))
	case errors.Is(err, entities.ErrInvalidStudentDebt),
		errors.Is(err, entities.ErrInvalidResitAttempt),
		errors.Is(err, entities.ErrInvalidResitRecord),
		errors.Is(err, entities.ErrInvalidResitResult):
		c.JSON(http.StatusUnprocessableEntity, response.ErrorResponse("VALIDATION_ERROR", "the request did not satisfy a domain rule"))
	default:
		c.JSON(http.StatusInternalServerError, response.InternalError("internal error"))
	}
}
