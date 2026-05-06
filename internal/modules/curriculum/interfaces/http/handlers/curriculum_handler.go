// Package handlers contains HTTP handlers for the curriculum module.
package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
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

// SubmitForApprovalPort is the narrow port for the submit use case
// (v0.117.0 lifecycle transition: draft → pending_approval).
type SubmitForApprovalPort interface {
	Execute(ctx context.Context, actorID int64, isAdmin bool, in curUsecases.SubmitForApprovalInput) (*entities.Curriculum, error)
}

// ApproveCurriculumPort is the narrow port for the admin-only approve
// use case (v0.117.0 lifecycle transition: pending_approval → approved).
type ApproveCurriculumPort interface {
	Execute(ctx context.Context, adminID int64, in curUsecases.ApproveCurriculumInput) (*entities.Curriculum, error)
}

// RejectCurriculumPort is the narrow port for the admin-only reject
// use case (v0.117.0 lifecycle transition: pending_approval → draft,
// reason captured in audit per ADR-3).
type RejectCurriculumPort interface {
	Execute(ctx context.Context, adminID int64, in curUsecases.RejectCurriculumInput) (*entities.Curriculum, error)
}

// CurriculumHandler exposes the seven curriculum endpoints (four CRUD
// + three lifecycle transitions) over HTTP.
type CurriculumHandler struct {
	create  CreateCurriculumPort
	get     GetCurriculumPort
	list    ListCurriculaPort
	update  UpdateCurriculumPort
	submit  SubmitForApprovalPort
	approve ApproveCurriculumPort
	reject  RejectCurriculumPort
}

// NewCurriculumHandler wires the handler. All seven ports are required
// (non-nil): nil dependencies would let requests reach a panic deeper
// in the call stack instead of failing during DI wiring. Mirrors the
// failure-closed posture established for analytics in v0.108.3 and
// the assignments line.
func NewCurriculumHandler(
	create CreateCurriculumPort,
	get GetCurriculumPort,
	list ListCurriculaPort,
	update UpdateCurriculumPort,
	submit SubmitForApprovalPort,
	approve ApproveCurriculumPort,
	reject RejectCurriculumPort,
) *CurriculumHandler {
	if create == nil || get == nil || list == nil || update == nil ||
		submit == nil || approve == nil || reject == nil {
		panic("curriculum: NewCurriculumHandler requires non-nil ports (create / get / list / update / submit / approve / reject)")
	}
	return &CurriculumHandler{
		create: create, get: get, list: list, update: update,
		submit: submit, approve: approve, reject: reject,
	}
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
		mapCurriculumError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response.Success(mapCurriculum(curriculum)))
}

// Get handles GET /api/curriculum/:id.
// @Summary Fetch a single curriculum by id
// @Tags curriculum
// @Produce json
// @Param id path int true "Curriculum ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Security BearerAuth
// @Router /api/curriculum/{id} [get]
func (h *CurriculumHandler) Get(c *gin.Context) {
	_, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	if !canRead(role) {
		c.JSON(http.StatusForbidden, response.Forbidden("students cannot read this curriculum view"))
		return
	}

	id, ok := parsePositiveID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid curriculum id"))
		return
	}

	curriculum, err := h.get.Execute(c.Request.Context(), id)
	if err != nil {
		mapCurriculumError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(mapCurriculum(curriculum)))
}

// Submit handles POST /api/curriculum/:id/submit.
// @Summary Submit a draft curriculum for approval
// @Tags curriculum
// @Produce json
// @Param id path int true "Curriculum ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 422 {object} response.Response
// @Security BearerAuth
// @Router /api/curriculum/{id}/submit [post]
func (h *CurriculumHandler) Submit(c *gin.Context) {
	actorID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	if !canWrite(role) {
		c.JSON(http.StatusForbidden, response.Forbidden("only methodist or system_admin may submit curricula"))
		return
	}
	id, ok := parsePositiveID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid curriculum id"))
		return
	}

	curriculum, err := h.submit.Execute(c.Request.Context(), actorID, isAdminRole(role),
		curUsecases.SubmitForApprovalInput{ID: id})
	if err != nil {
		mapCurriculumError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(mapCurriculum(curriculum)))
}

// Approve handles POST /api/curriculum/:id/approve.
// Admin-only — both the route-level RequireRole(SystemAdmin) and
// the handler-level canApprove whitelist must agree.
//
// @Summary Approve a curriculum (admin only)
// @Tags curriculum
// @Produce json
// @Param id path int true "Curriculum ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 422 {object} response.Response
// @Security BearerAuth
// @Router /api/curriculum/{id}/approve [post]
func (h *CurriculumHandler) Approve(c *gin.Context) {
	adminID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	if !canApprove(role) {
		c.JSON(http.StatusForbidden, response.Forbidden("only system_admin may approve curricula"))
		return
	}
	id, ok := parsePositiveID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid curriculum id"))
		return
	}

	curriculum, err := h.approve.Execute(c.Request.Context(), adminID,
		curUsecases.ApproveCurriculumInput{ID: id})
	if err != nil {
		mapCurriculumError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(mapCurriculum(curriculum)))
}

// RejectCurriculumRequest is the JSON body schema for POST /api/curriculum/:id/reject.
// Reason is required and non-empty (handler enforces); the use case
// audits it verbatim per ADR-3 (audit-only, not persisted on the entity).
type RejectCurriculumRequest struct {
	Reason string `json:"reason" example:"Не соответствует ФГОС, переделать раздел дисциплин"`
}

// Reject handles POST /api/curriculum/:id/reject.
// Admin-only — the canApprove whitelist gates write access on top of
// the route-level RequireRole(SystemAdmin) middleware (the same role
// admits both Approve and Reject — both are admin-only lifecycle
// transitions).
//
// @Summary Reject a curriculum (admin only)
// @Tags curriculum
// @Accept json
// @Produce json
// @Param id   path int                       true "Curriculum ID"
// @Param body body RejectCurriculumRequest   true "Rejection reason"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 422 {object} response.Response
// @Security BearerAuth
// @Router /api/curriculum/{id}/reject [post]
func (h *CurriculumHandler) Reject(c *gin.Context) {
	adminID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	if !canApprove(role) {
		c.JSON(http.StatusForbidden, response.Forbidden("only system_admin may reject curricula"))
		return
	}
	id, ok := parsePositiveID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid curriculum id"))
		return
	}
	var body RejectCurriculumRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid request body: "+err.Error()))
		return
	}
	reason := strings.TrimSpace(body.Reason)
	if reason == "" {
		c.JSON(http.StatusBadRequest, response.BadRequest("reason must not be empty"))
		return
	}

	curriculum, err := h.reject.Execute(c.Request.Context(), adminID,
		curUsecases.RejectCurriculumInput{ID: id, Reason: reason})
	if err != nil {
		mapCurriculumError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(mapCurriculum(curriculum)))
}

// canApprove is the role whitelist for the admin-only lifecycle
// endpoints (Approve, Reject). Defense in depth on top of the
// route-level RequireRole(SystemAdmin) middleware — if the route
// gate is ever stripped or relaxed, this handler still refuses
// to dispatch the use case.
func canApprove(role string) bool {
	return role == roleSystemAdmin
}

// UpdateCurriculumRequest is the JSON body schema for PUT /api/curriculum/:id.
// Exported so swag can generate the schema in the OpenAPI spec.
type UpdateCurriculumRequest struct {
	Title       string `json:"title"`
	Code        string `json:"code"`
	Specialty   string `json:"specialty"`
	Year        int    `json:"year"`
	Description string `json:"description"`
}

// Update handles PUT /api/curriculum/:id.
// @Summary Update a draft curriculum
// @Tags curriculum
// @Accept json
// @Produce json
// @Param id   path int                       true "Curriculum ID"
// @Param body body UpdateCurriculumRequest   true "Curriculum payload"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 422 {object} response.Response
// @Security BearerAuth
// @Router /api/curriculum/{id} [put]
func (h *CurriculumHandler) Update(c *gin.Context) {
	actorID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	if !canWrite(role) {
		c.JSON(http.StatusForbidden, response.Forbidden("only methodist or system_admin may edit curricula"))
		return
	}

	id, ok := parsePositiveID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid curriculum id"))
		return
	}

	var body UpdateCurriculumRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid request body: "+err.Error()))
		return
	}

	curriculum, err := h.update.Execute(c.Request.Context(), actorID, isAdminRole(role),
		curUsecases.UpdateCurriculumInput{
			ID:          id,
			Title:       body.Title,
			Code:        body.Code,
			Specialty:   body.Specialty,
			Year:        body.Year,
			Description: body.Description,
		})
	if err != nil {
		mapCurriculumError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(mapCurriculum(curriculum)))
}

// CurriculaListResponse is the response shape for the list endpoint.
// Exported so swag picks it up in the OpenAPI spec.
type CurriculaListResponse struct {
	Items []CurriculumDTO `json:"items"`
	Total int             `json:"total"`
}

// List handles GET /api/curriculum with optional filters.
// @Summary List curricula matching the supplied filters
// @Tags curriculum
// @Produce json
// @Param status      query string false "Status filter (draft / pending_approval / approved / archived)"
// @Param year        query int    false "Academic year of programme start"
// @Param specialty   query string false "Specialty exact match"
// @Param created_by  query int    false "Filter to a specific methodist's curricula"
// @Param limit       query int    false "Page size (1..200, default 50)"
// @Param offset      query int    false "Page offset (default 0)"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Security BearerAuth
// @Router /api/curriculum [get]
func (h *CurriculumHandler) List(c *gin.Context) {
	_, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	if !canRead(role) {
		c.JSON(http.StatusForbidden, response.Forbidden("students cannot read this curriculum view"))
		return
	}

	in, err := parseListInput(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	page, err := h.list.Execute(c.Request.Context(), in)
	if err != nil {
		mapCurriculumError(c, err)
		return
	}

	dtos := make([]CurriculumDTO, 0, len(page.Items))
	for _, c := range page.Items {
		dtos = append(dtos, mapCurriculum(c))
	}
	c.JSON(http.StatusOK, response.Success(CurriculaListResponse{
		Items: dtos,
		Total: page.Total,
	}))
}

// parseListInput converts gin's query strings into a typed
// ListCurriculaInput, rejecting any value that is guaranteed to
// fail validation downstream (unknown status literal, year
// outside the entity's accepted range, non-positive created_by,
// non-numeric pagination). Use-case-side defaults (limit/offset
// clamps) still apply for valid-but-extreme inputs.
func parseListInput(c *gin.Context) (curUsecases.ListCurriculaInput, error) {
	var in curUsecases.ListCurriculaInput

	if raw := c.Query("status"); raw != "" {
		st := entities.CurriculumStatus(raw)
		if !st.IsValid() {
			return in, errors.New("invalid status filter")
		}
		in.Status = &st
	}
	if raw := c.Query("year"); raw != "" {
		y, err := strconv.Atoi(raw)
		if err != nil {
			return in, errors.New("invalid year filter")
		}
		if y < 2000 || y > 2100 {
			return in, errors.New("year filter must be in [2000, 2100]")
		}
		in.Year = &y
	}
	if raw := c.Query("specialty"); raw != "" {
		in.Specialty = raw
	}
	if raw := c.Query("created_by"); raw != "" {
		v, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || v <= 0 {
			return in, errors.New("invalid created_by filter")
		}
		in.CreatedBy = &v
	}
	if raw := c.Query("limit"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 0 {
			return in, errors.New("invalid limit")
		}
		in.Limit = v
	}
	if raw := c.Query("offset"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 0 {
			return in, errors.New("invalid offset")
		}
		in.Offset = v
	}
	return in, nil
}

// canRead is the role whitelist for the read endpoints. v0.116.0
// admits the four non-student roles (methodist, system_admin,
// academic_secretary, teacher); student access requires the
// specialty-scoped variant landing in a future release (ADR-3).
func canRead(role string) bool {
	switch role {
	case roleMethodist, roleSystemAdmin, roleAcademicSecretary, roleTeacher:
		return true
	default:
		return false
	}
}

// parsePositiveID parses a path component as a strict positive
// integer. Empty string, negative, zero and fractional values are
// rejected at the boundary so the use case never sees a
// guaranteed-4xx id and the caller learns the issue immediately.
// Mirrors the Number.isInteger-style discipline established for
// student-facing detail pages in v0.114.0.
func parsePositiveID(raw string) (int64, bool) {
	if raw == "" {
		return 0, false
	}
	// Reject any non-digit byte upfront — strconv.ParseInt would
	// accept '+5' or leading whitespace; we want strict digits only.
	for i, r := range raw {
		if r < '0' || r > '9' {
			// Allow a single leading '-' so we can produce a clean
			// "negative" rejection rather than the strconv generic
			// error path; everything else fails fast.
			if i == 0 && r == '-' {
				continue
			}
			return 0, false
		}
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
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

// mapCurriculumError maps domain / repository sentinels surfaced by
// any of the four use cases to the matching HTTP status. Every
// sentinel is matched explicitly via errors.Is BEFORE the generic
// MapDomainError fallback — otherwise the generic mapper falls
// through to 500 and clients lose the actionable distinction
// between e.g. "code conflict" (409) and "internal error" (500).
//
// Used by Create / Get / List / Update; only a subset of sentinels
// is reachable from each (e.g. ErrCannotEditApproved only from
// Update), but a single mapper keeps the response shape uniform.
func mapCurriculumError(c *gin.Context, err error) {
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
	case errors.Is(err, entities.ErrCannotSubmit):
		c.JSON(http.StatusUnprocessableEntity,
			response.ErrorResponse("NOT_DRAFT", "curriculum must be in draft status to submit"))
		return
	case errors.Is(err, entities.ErrCannotApprove):
		c.JSON(http.StatusUnprocessableEntity,
			response.ErrorResponse("NOT_PENDING", "curriculum must be in pending_approval status to approve"))
		return
	case errors.Is(err, entities.ErrCannotReject):
		c.JSON(http.StatusUnprocessableEntity,
			response.ErrorResponse("NOT_PENDING", "curriculum must be in pending_approval status to reject"))
		return
	case errors.Is(err, entities.ErrInvalidCurriculum):
		c.JSON(http.StatusUnprocessableEntity,
			response.ErrorResponse("INVALID_INPUT", err.Error()))
		return
	}
	httpErr := response.MapDomainError(err)
	c.JSON(httpErr.Status, httpErr.Response)
}

// roleStudent is mirrored from auth/domain/permission.go for
// completeness even though no curriculum endpoint exposes the
// student role today; canRead / canWrite reject student via the
// switch default. Kept named (not removed) so a future student-
// scoped read endpoint has the constant ready.
var _ = roleStudent
