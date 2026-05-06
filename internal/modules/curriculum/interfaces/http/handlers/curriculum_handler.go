// Package handlers contains HTTP handlers for the curriculum module.
package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	curUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
)

// Recognised role values, mirrored from auth/domain/permission.go but
// duplicated here to avoid an HTTP layer importing the auth domain.
// Updating this list when a new role is introduced is a deliberate
// step — the failure-closed default is "unknown role → no access."
const (
	roleTeacher           = "teacher"
	roleMethodist         = "methodist"
	roleAcademicSecretary = "academic_secretary"
	roleSystemAdmin       = "system_admin"
	roleStudent           = "student"
)

// CreateCurriculumPort is the narrow port for the create use case.
type CreateCurriculumPort interface {
	Execute(ctx context.Context, actorID int64, in curUsecases.CreateCurriculumInput) (*entities.Curriculum, error)
}

// GetCurriculumPort is the narrow port for the get use case.
type GetCurriculumPort interface {
	Execute(ctx context.Context, id int64) (*entities.Curriculum, error)
}

// ListCurriculaPort is the narrow port for the list use case.
type ListCurriculaPort interface {
	Execute(ctx context.Context, in curUsecases.ListCurriculaInput) (curUsecases.CurriculaPage, error)
}

// UpdateCurriculumPort is the narrow port for the update use case.
type UpdateCurriculumPort interface {
	Execute(ctx context.Context, actorID int64, isAdmin bool, in curUsecases.UpdateCurriculumInput) (*entities.Curriculum, error)
}

// CurriculumHandler exposes the four CRUD endpoints over HTTP.
type CurriculumHandler struct {
	create CreateCurriculumPort
	get    GetCurriculumPort
	list   ListCurriculaPort
	update UpdateCurriculumPort
}

// NewCurriculumHandler wires the handler. All four ports are required
// (non-nil): nil dependencies would let requests reach a panic deeper
// in the call stack instead of failing during DI wiring. Mirrors the
// failure-closed posture established for analytics in v0.108.3 and
// the assignments line.
func NewCurriculumHandler(
	create CreateCurriculumPort,
	get GetCurriculumPort,
	list ListCurriculaPort,
	update UpdateCurriculumPort,
) *CurriculumHandler {
	if create == nil || get == nil || list == nil || update == nil {
		panic("curriculum: NewCurriculumHandler requires non-nil ports (create / get / list / update)")
	}
	return &CurriculumHandler{create: create, get: get, list: list, update: update}
}

// CurriculumDTO is the public response shape for a curriculum row.
// Timestamps are encoded as RFC 3339 strings so frontend clients
// don't depend on Go time-marshal quirks.
type CurriculumDTO struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	Code        string  `json:"code"`
	Specialty   string  `json:"specialty"`
	Year        int     `json:"year"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	CreatedBy   int64   `json:"created_by"`
	ApprovedBy  *int64  `json:"approved_by,omitempty"`
	ApprovedAt  *string `json:"approved_at,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// mapCurriculum projects the domain entity to its public DTO.
func mapCurriculum(c *entities.Curriculum) CurriculumDTO {
	dto := CurriculumDTO{
		ID:          c.ID,
		Title:       c.Title(),
		Code:        c.Code(),
		Specialty:   c.Specialty(),
		Year:        c.Year(),
		Description: c.Description(),
		Status:      string(c.Status()),
		CreatedBy:   c.CreatedBy(),
		CreatedAt:   c.CreatedAt().Format(time.RFC3339),
		UpdatedAt:   c.UpdatedAt().Format(time.RFC3339),
	}
	if ab := c.ApprovedBy(); ab != nil {
		v := *ab
		dto.ApprovedBy = &v
	}
	if aat := c.ApprovedAt(); aat != nil {
		s := aat.Format(time.RFC3339)
		dto.ApprovedAt = &s
	}
	return dto
}

// CreateCurriculumRequest is the JSON body schema for POST /api/curriculum.
// Exported so swag can generate the schema in the OpenAPI spec.
type CreateCurriculumRequest struct {
	Title       string `json:"title"       example:"ИВТ-2026 / 4 года"`
	Code        string `json:"code"        example:"09.03.04-2026"`
	Specialty   string `json:"specialty"   example:"Информатика и вычислительная техника"`
	Year        int    `json:"year"        example:"2026"`
	Description string `json:"description" example:"Учебный план направления подготовки"`
}

// Create handles POST /api/curriculum.
// @Summary Create a curriculum
// @Tags curriculum
// @Accept json
// @Produce json
// @Param body body CreateCurriculumRequest true "Curriculum payload"
// @Success 201 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 422 {object} response.Response
// @Security BearerAuth
// @Router /api/curriculum [post]
func (h *CurriculumHandler) Create(c *gin.Context) {
	actorID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	if !canWrite(role) {
		c.JSON(http.StatusForbidden, response.Forbidden("only methodist or system_admin may create curricula"))
		return
	}

	var body CreateCurriculumRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid request body: "+err.Error()))
		return
	}

	curriculum, err := h.create.Execute(c.Request.Context(), actorID, curUsecases.CreateCurriculumInput{
		Title:       body.Title,
		Code:        body.Code,
		Specialty:   body.Specialty,
		Year:        body.Year,
		Description: body.Description,
	})
	if err != nil {
		mapWriteError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response.Success(mapCurriculum(curriculum)))
}

// authContext extracts user_id + role from the gin context. Failure-
// closed: missing either signals the route was reached without going
// through the auth middleware (or with a malformed JWT) and we
// refuse to fall back to a silent admin identity.
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
	roleStr, _ := roleVal.(string)
	if roleStr == "" {
		return 0, "", false
	}
	return userID, roleStr, true
}

// canWrite is the role whitelist for create / update endpoints.
// Defense in depth on top of RequireNonStudent (which only excludes
// students) — only methodist and system_admin should mutate
// curricula per the PermissionMatrix.
func canWrite(role string) bool {
	return role == roleMethodist || role == roleSystemAdmin
}

// isAdminRole reports whether the role is system_admin — used by
// the update path to grant the admin override on AuthorizeEdit.
func isAdminRole(role string) bool {
	return role == roleSystemAdmin
}

// mapWriteError maps domain / repository sentinels coming back from
// the Create / Update use cases to the matching HTTP status. Every
// sentinel is matched explicitly via errors.Is BEFORE the generic
// MapDomainError fallback — otherwise the generic mapper falls
// through to 500 and clients lose the actionable distinction
// between e.g. "code conflict" (409) and "internal error" (500).
func mapWriteError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repositories.ErrCurriculumNotFound):
		c.JSON(http.StatusNotFound, response.NotFound("curriculum"))
		return
	case errors.Is(err, repositories.ErrCurriculumCodeExists):
		c.JSON(http.StatusConflict,
			response.ErrorResponse("CODE_EXISTS", "curriculum code already exists"))
		return
	case errors.Is(err, entities.ErrCurriculumScopeForbidden):
		c.JSON(http.StatusForbidden,
			response.Forbidden("only the author or an administrator may edit this curriculum"))
		return
	case errors.Is(err, entities.ErrCannotEditApproved):
		c.JSON(http.StatusUnprocessableEntity,
			response.ErrorResponse("NOT_EDITABLE", "curriculum is not in an editable state"))
		return
	case errors.Is(err, entities.ErrInvalidCurriculum):
		c.JSON(http.StatusUnprocessableEntity,
			response.ErrorResponse("INVALID_INPUT", err.Error()))
		return
	}
	httpErr := response.MapDomainError(err)
	c.JSON(httpErr.Status, httpErr.Response)
}

// guard against unused warnings while v0.116.0 only exercises some
// of these helpers (read endpoints land in the next two cycles).
var _ = roleTeacher
var _ = roleAcademicSecretary
var _ = roleStudent
var _ = isAdminRole
