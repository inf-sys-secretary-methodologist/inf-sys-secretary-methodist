// Package handlers exposes HTTP endpoints for the work_program
// (рабочая программа дисциплины / РПД) bounded context.
package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	wpUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
)

// Narrow ports — handler depends on the methods it calls, not the
// concrete use-case structs. Keeps the handler test double-friendly
// and the DI seam explicit.

// CreateWorkProgramPort is the narrow port for CreateWorkProgramUseCase.
type CreateWorkProgramPort interface {
	Execute(ctx context.Context, actorID int64, actorRole string, in wpUsecases.CreateWorkProgramInput) (*entities.WorkProgram, error)
}

// GetWorkProgramPort is the narrow port for GetWorkProgramUseCase.
type GetWorkProgramPort interface {
	Execute(ctx context.Context, actorID int64, actorRole string, in wpUsecases.GetWorkProgramInput) (*entities.WorkProgram, error)
}

// ListWorkProgramsPort is the narrow port for ListWorkProgramsUseCase.
type ListWorkProgramsPort interface {
	Execute(ctx context.Context, actorID int64, actorRole string, in wpUsecases.ListWorkProgramsInput) (wpUsecases.ListWorkProgramsResult, error)
}

// SubmitWorkProgramPort is the narrow port for SubmitWorkProgramUseCase.
type SubmitWorkProgramPort interface {
	Execute(ctx context.Context, actorID int64, actorRole string, in wpUsecases.SubmitWorkProgramInput) (*entities.WorkProgram, error)
}

// ApproveWorkProgramPort is the narrow port for ApproveWorkProgramUseCase.
type ApproveWorkProgramPort interface {
	Execute(ctx context.Context, actorID int64, actorRole string, in wpUsecases.ApproveWorkProgramInput) (*entities.WorkProgram, error)
}

// RejectWorkProgramPort is the narrow port for RejectWorkProgramUseCase.
type RejectWorkProgramPort interface {
	Execute(ctx context.Context, actorID int64, actorRole string, in wpUsecases.RejectWorkProgramInput) (*entities.WorkProgram, error)
}

// DiscardDraftWorkProgramPort is the narrow port for DiscardDraftWorkProgramUseCase.
type DiscardDraftWorkProgramPort interface {
	Execute(ctx context.Context, actorID int64, actorRole string, in wpUsecases.DiscardDraftWorkProgramInput) (*entities.WorkProgram, error)
}

// WorkProgramHandler exposes the 7 РПД endpoints over HTTP: read + create
// (PR 4a) and the four status transitions submit/approve/reject/discard
// (PR 4b).
type WorkProgramHandler struct {
	create  CreateWorkProgramPort
	get     GetWorkProgramPort
	list    ListWorkProgramsPort
	submit  SubmitWorkProgramPort
	approve ApproveWorkProgramPort
	reject  RejectWorkProgramPort
	discard DiscardDraftWorkProgramPort
}

// NewWorkProgramHandler wires the handler. All ports are required —
// a nil dependency would surface as a nil-pointer panic deep under load
// instead of failing loudly at DI wiring time (mirror к extracurricular
// + curriculum failure-closed posture).
func NewWorkProgramHandler(
	create CreateWorkProgramPort,
	get GetWorkProgramPort,
	list ListWorkProgramsPort,
	submit SubmitWorkProgramPort,
	approve ApproveWorkProgramPort,
	reject RejectWorkProgramPort,
	discard DiscardDraftWorkProgramPort,
) *WorkProgramHandler {
	if create == nil || get == nil || list == nil ||
		submit == nil || approve == nil || reject == nil || discard == nil {
		panic("work_program: NewWorkProgramHandler requires non-nil ports")
	}
	return &WorkProgramHandler{
		create: create, get: get, list: list,
		submit: submit, approve: approve, reject: reject, discard: discard,
	}
}

// CreateWorkProgramRequest is the JSON body for POST /work-programs.
// binding tags per CLAUDE.md feedback (NOT `validate:`). The author is
// derived from the JWT subject server-side — never from the body.
type CreateWorkProgramRequest struct {
	DisciplineID       int64  `json:"discipline_id"        binding:"required"`
	SpecialtyCode      string `json:"specialty_code"       binding:"required"`
	ApplicableFromYear int    `json:"applicable_from_year" binding:"required"`
	Title              string `json:"title"                binding:"required"`
	Annotation         string `json:"annotation"`
}

// RejectWorkProgramRequest is the JSON body for POST /work-programs/:id/reject.
// The reason is mandatory so the author understands what to revise — the
// domain enforces non-empty after trimming, this binding tag fails fast.
type RejectWorkProgramRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// ===== Response DTOs =====

// WorkProgramDTO is the full public response shape for one РПД, with all
// six inner collections hydrated. Timestamps are RFC 3339 strings.
type WorkProgramDTO struct {
	ID                 int64           `json:"id"`
	DisciplineID       int64           `json:"discipline_id"`
	SpecialtyCode      string          `json:"specialty_code"`
	ApplicableFromYear int             `json:"applicable_from_year"`
	Title              string          `json:"title"`
	Annotation         string          `json:"annotation"`
	Status             string          `json:"status"`
	AuthorID           int64           `json:"author_id"`
	ApproverID         *int64          `json:"approver_id,omitempty"`
	ApprovedAt         *string         `json:"approved_at,omitempty"`
	RejectReason       string          `json:"reject_reason,omitempty"`
	Version            int             `json:"version"`
	CreatedAt          string          `json:"created_at"`
	UpdatedAt          string          `json:"updated_at"`
	Goals              []GoalDTO       `json:"goals"`
	Competences        []CompetenceDTO `json:"competences"`
	Topics             []TopicDTO      `json:"topics"`
	Assessments        []AssessmentDTO `json:"assessments"`
	References         []ReferenceDTO  `json:"references"`
	Revisions          []RevisionDTO   `json:"revisions"`
}

// GoalDTO is the projected shape of one goal row.
type GoalDTO struct {
	ID         int64  `json:"id"`
	Text       string `json:"text"`
	OrderIndex int    `json:"order_index"`
}

// CompetenceDTO is the projected shape of one competence row.
type CompetenceDTO struct {
	ID          int64  `json:"id"`
	Code        string `json:"code"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

// TopicDTO is the projected shape of one topic row.
type TopicDTO struct {
	ID               int64  `json:"id"`
	Kind             string `json:"kind"`
	Title            string `json:"title"`
	Hours            int    `json:"hours"`
	WeekNumber       *int   `json:"week_number,omitempty"`
	LearningOutcomes string `json:"learning_outcomes"`
	OrderIndex       int    `json:"order_index"`
}

// AssessmentDTO is the projected shape of one ФОС item row.
type AssessmentDTO struct {
	ID               int64    `json:"id"`
	Type             string   `json:"type"`
	Description      string   `json:"description"`
	MaxScore         int      `json:"max_score"`
	ExampleQuestions []string `json:"example_questions"`
}

// ReferenceDTO is the projected shape of one literature reference row.
type ReferenceDTO struct {
	ID         int64  `json:"id"`
	Kind       string `json:"kind"`
	Citation   string `json:"citation"`
	Year       *int   `json:"year,omitempty"`
	ISBN       string `json:"isbn,omitempty"`
	URL        string `json:"url,omitempty"`
	OrderIndex int    `json:"order_index"`
}

// RevisionDTO is the projected shape of one revision (лист актуализации).
type RevisionDTO struct {
	ID             int64   `json:"id"`
	RevisionNumber int     `json:"revision_number"`
	ChangeType     string  `json:"change_type"`
	ChangeSummary  string  `json:"change_summary"`
	Status         string  `json:"status"`
	AuthorID       int64   `json:"author_id"`
	ApproverID     *int64  `json:"approver_id,omitempty"`
	ApprovedAt     *string `json:"approved_at,omitempty"`
	RejectReason   string  `json:"reject_reason,omitempty"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

// WorkProgramSummaryDTO is the lightweight list-row projection — root
// fields only (no inner collections). Mirrors repositories.ListItem.
type WorkProgramSummaryDTO struct {
	ID                 int64  `json:"id"`
	DisciplineID       int64  `json:"discipline_id"`
	SpecialtyCode      string `json:"specialty_code"`
	ApplicableFromYear int    `json:"applicable_from_year"`
	Title              string `json:"title"`
	Status             string `json:"status"`
	AuthorID           int64  `json:"author_id"`
	Version            int    `json:"version"`
}

// WorkProgramsListResponse is the page response shape.
type WorkProgramsListResponse struct {
	Items []WorkProgramSummaryDTO `json:"items"`
	Total int                     `json:"total"`
}

// ===== Mappers =====

func formatRFC3339Ptr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format(time.RFC3339)
	return &s
}

// mapWorkProgram maps a hydrated aggregate to the full public DTO.
func mapWorkProgram(wp *entities.WorkProgram) WorkProgramDTO {
	goals := wp.Goals()
	goalDTOs := make([]GoalDTO, 0, len(goals))
	for _, g := range goals {
		goalDTOs = append(goalDTOs, GoalDTO{ID: g.ID(), Text: g.Text(), OrderIndex: g.OrderIndex()})
	}
	comps := wp.Competences()
	compDTOs := make([]CompetenceDTO, 0, len(comps))
	for _, c := range comps {
		compDTOs = append(compDTOs, CompetenceDTO{
			ID: c.ID(), Code: c.Code(), Type: string(c.Type()), Description: c.Description(),
		})
	}
	topics := wp.Topics()
	topicDTOs := make([]TopicDTO, 0, len(topics))
	for _, tp := range topics {
		topicDTOs = append(topicDTOs, TopicDTO{
			ID: tp.ID(), Kind: string(tp.Kind()), Title: tp.Title(), Hours: tp.Hours(),
			WeekNumber: tp.WeekNumber(), LearningOutcomes: tp.LearningOutcomes(), OrderIndex: tp.OrderIndex(),
		})
	}
	asmts := wp.Assessments()
	asmtDTOs := make([]AssessmentDTO, 0, len(asmts))
	for _, a := range asmts {
		asmtDTOs = append(asmtDTOs, AssessmentDTO{
			ID: a.ID(), Type: string(a.Type()), Description: a.Description(),
			MaxScore: a.MaxScore(), ExampleQuestions: a.ExampleQuestions(),
		})
	}
	refs := wp.References()
	refDTOs := make([]ReferenceDTO, 0, len(refs))
	for _, r := range refs {
		refDTOs = append(refDTOs, ReferenceDTO{
			ID: r.ID(), Kind: string(r.Kind()), Citation: r.Citation(), Year: r.Year(),
			ISBN: r.ISBN(), URL: r.URL(), OrderIndex: r.OrderIndex(),
		})
	}
	revs := wp.Revisions()
	revDTOs := make([]RevisionDTO, 0, len(revs))
	for _, r := range revs {
		revDTOs = append(revDTOs, RevisionDTO{
			ID: r.ID(), RevisionNumber: r.RevisionNumber(), ChangeType: string(r.ChangeType()),
			ChangeSummary: r.ChangeSummary(), Status: string(r.Status()), AuthorID: r.AuthorID(),
			ApproverID: r.ApproverID(), ApprovedAt: formatRFC3339Ptr(r.ApprovedAt()),
			RejectReason: r.RejectReason(), CreatedAt: r.CreatedAt().Format(time.RFC3339),
			UpdatedAt: r.UpdatedAt().Format(time.RFC3339),
		})
	}
	return WorkProgramDTO{
		ID:                 wp.ID(),
		DisciplineID:       wp.DisciplineID(),
		SpecialtyCode:      wp.SpecialtyCode(),
		ApplicableFromYear: wp.ApplicableFromYear(),
		Title:              wp.Title(),
		Annotation:         wp.Annotation(),
		Status:             string(wp.Status()),
		AuthorID:           wp.AuthorID(),
		ApproverID:         wp.ApproverID(),
		ApprovedAt:         formatRFC3339Ptr(wp.ApprovedAt()),
		RejectReason:       wp.RejectReason(),
		Version:            wp.Version(),
		CreatedAt:          wp.CreatedAt().Format(time.RFC3339),
		UpdatedAt:          wp.UpdatedAt().Format(time.RFC3339),
		Goals:              goalDTOs,
		Competences:        compDTOs,
		Topics:             topicDTOs,
		Assessments:        asmtDTOs,
		References:         refDTOs,
		Revisions:          revDTOs,
	}
}

func mapSummary(it repositories.ListItem) WorkProgramSummaryDTO {
	return WorkProgramSummaryDTO{
		ID:                 it.ID,
		DisciplineID:       it.DisciplineID,
		SpecialtyCode:      it.SpecialtyCode,
		ApplicableFromYear: it.ApplicableFromYear,
		Title:              it.Title,
		Status:             string(it.Status),
		AuthorID:           it.AuthorID,
		Version:            it.Version,
	}
}

// ===== Auth + parse helpers =====

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

func parsePositiveID(raw string) (int64, bool) {
	if raw == "" {
		return 0, false
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func isAdminRole(role string) bool { return role == "system_admin" }

// mapWorkProgramError maps domain / repository sentinels to HTTP status.
//
// hideForbiddenAsNotFound implements the OWASP IDOR mitigation: on
// resource-scoped reads/transitions, a non-admin caller who is denied
// access must not be able to tell "this РПД exists but I can't see it"
// (403) apart from "no such РПД" (404). Collapsing both to 404 removes
// the enumeration oracle. Admins keep the 403 signal (they are entitled
// to know the resource exists). The use-case audit log records the true
// reason (forbidden vs not_found) for internal forensics regardless.
func mapWorkProgramError(c *gin.Context, err error, hideForbiddenAsNotFound bool) {
	switch {
	case errors.Is(err, repositories.ErrWorkProgramNotFound):
		c.JSON(http.StatusNotFound, response.NotFound("work program"))
	case errors.Is(err, domain.ErrWorkProgramScopeForbidden):
		if hideForbiddenAsNotFound {
			c.JSON(http.StatusNotFound, response.NotFound("work program"))
		} else {
			c.JSON(http.StatusForbidden, response.Forbidden("not authorized to operate on this work program"))
		}
	case errors.Is(err, repositories.ErrWorkProgramIdentityExists):
		c.JSON(http.StatusConflict, response.ErrorResponse("IDENTITY_EXISTS", "a work program with this discipline + specialty + year already exists"))
	case errors.Is(err, repositories.ErrWorkProgramVersionConflict):
		c.JSON(http.StatusConflict, response.ErrorResponse("VERSION_CONFLICT", "work program was modified concurrently; reload + retry"))
	case errors.Is(err, domain.ErrInvalidStatusTransition):
		c.JSON(http.StatusUnprocessableEntity, response.ErrorResponse("INVALID_TRANSITION", err.Error()))
	case errors.Is(err, domain.ErrInvalidWorkProgram):
		c.JSON(http.StatusUnprocessableEntity, response.ErrorResponse("INVALID_WORK_PROGRAM", err.Error()))
	default:
		c.JSON(http.StatusInternalServerError, response.InternalError("internal error"))
	}
}

// ===== Endpoints (PR 4a stubs — implemented in following GREEN commits) =====

// Create handles POST /api/v1/work-programs.
// @Summary Create a work program (РПД) draft
// @Tags    work-programs
// @Accept  json
// @Produce json
// @Param   body body CreateWorkProgramRequest true "Work program payload"
// @Success 201 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 422 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/work-programs [post]
func (h *WorkProgramHandler) Create(c *gin.Context) {
	actorID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	var body CreateWorkProgramRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid request body: "+err.Error()))
		return
	}
	wp, err := h.create.Execute(c.Request.Context(), actorID, role, wpUsecases.CreateWorkProgramInput{
		DisciplineID:       body.DisciplineID,
		SpecialtyCode:      body.SpecialtyCode,
		ApplicableFromYear: body.ApplicableFromYear,
		Title:              body.Title,
		Annotation:         body.Annotation,
	})
	if err != nil {
		// Create is a collection POST — a role-based denial is a true
		// 403, not an enumeration oracle, so no IDOR collapse here.
		mapWorkProgramError(c, err, false)
		return
	}
	c.JSON(http.StatusCreated, response.Success(mapWorkProgram(wp)))
}

// Get handles GET /api/v1/work-programs/:id.
// @Summary Fetch one work program (РПД) by id with all inner collections
// @Tags    work-programs
// @Produce json
// @Param   id path int true "Work program ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/work-programs/{id} [get]
func (h *WorkProgramHandler) Get(c *gin.Context) {
	actorID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	id, ok := parsePositiveID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid work program id"))
		return
	}
	wp, err := h.get.Execute(c.Request.Context(), actorID, role, wpUsecases.GetWorkProgramInput{ID: id})
	if err != nil {
		// Non-admins get scope-forbidden collapsed to 404 (IDOR).
		mapWorkProgramError(c, err, !isAdminRole(role))
		return
	}
	c.JSON(http.StatusOK, response.Success(mapWorkProgram(wp)))
}

// List handles GET /api/v1/work-programs.
// @Summary List work programs (РПД), role-scoped + filterable
// @Tags    work-programs
// @Produce json
// @Param   status               query string false "Lifecycle status filter"
// @Param   discipline_id        query int    false "Discipline id filter"
// @Param   specialty_code       query string false "Specialty code filter"
// @Param   applicable_from_year query int    false "Cohort year filter"
// @Param   author_id            query int    false "Author user id filter"
// @Param   limit                query int    false "Page size (default 50, max 200)"
// @Param   offset               query int    false "Page offset"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/work-programs [get]
func (h *WorkProgramHandler) List(c *gin.Context) {
	actorID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	in := wpUsecases.ListWorkProgramsInput{
		SpecialtyCode: c.Query("specialty_code"),
	}
	if s := c.Query("status"); s != "" {
		in.Status = &s
	}
	if v := c.Query("discipline_id"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			in.DisciplineID = &n
		}
	}
	if v := c.Query("applicable_from_year"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			in.ApplicableFromYear = &n
		}
	}
	if v := c.Query("author_id"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			in.AuthorID = &n
		}
	}
	// Pagination defaults / clamps live in the use case so every caller
	// inherits the same bounds — handler passes raw values through.
	in.Limit, _ = strconv.Atoi(c.Query("limit"))
	in.Offset, _ = strconv.Atoi(c.Query("offset"))

	res, err := h.list.Execute(c.Request.Context(), actorID, role, in)
	if err != nil {
		// Collection endpoint — role-based denial is a true 403.
		mapWorkProgramError(c, err, false)
		return
	}
	out := WorkProgramsListResponse{
		Items: make([]WorkProgramSummaryDTO, 0, len(res.Items)),
		Total: res.Total,
	}
	for _, it := range res.Items {
		out.Items = append(out.Items, mapSummary(it))
	}
	c.JSON(http.StatusOK, response.Success(out))
}

// Submit handles POST /api/v1/work-programs/:id/submit.
// @Summary Submit a draft work program for approval (draft → pending_approval)
// @Tags    work-programs
// @Produce json
// @Param   id path int true "Work program ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 422 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/work-programs/{id}/submit [post]
func (h *WorkProgramHandler) Submit(c *gin.Context) {
	actorID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	id, ok := parsePositiveID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid work program id"))
		return
	}
	wp, err := h.submit.Execute(c.Request.Context(), actorID, role, wpUsecases.SubmitWorkProgramInput{ID: id})
	if err != nil {
		mapWorkProgramError(c, err, !isAdminRole(role))
		return
	}
	c.JSON(http.StatusOK, response.Success(mapWorkProgram(wp)))
}

// Approve handles POST /api/v1/work-programs/:id/approve.
// @Summary Approve a pending work program (pending_approval → approved)
// @Tags    work-programs
// @Produce json
// @Param   id path int true "Work program ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 422 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/work-programs/{id}/approve [post]
func (h *WorkProgramHandler) Approve(c *gin.Context) {
	actorID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	id, ok := parsePositiveID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid work program id"))
		return
	}
	wp, err := h.approve.Execute(c.Request.Context(), actorID, role, wpUsecases.ApproveWorkProgramInput{ID: id})
	if err != nil {
		mapWorkProgramError(c, err, !isAdminRole(role))
		return
	}
	c.JSON(http.StatusOK, response.Success(mapWorkProgram(wp)))
}

// Reject handles POST /api/v1/work-programs/:id/reject.
func (h *WorkProgramHandler) Reject(c *gin.Context) { c.Status(http.StatusNotImplemented) }

// Discard handles POST /api/v1/work-programs/:id/discard.
func (h *WorkProgramHandler) Discard(c *gin.Context) { c.Status(http.StatusNotImplemented) }

// RegisterWorkProgramRoutes mounts all 7 endpoints under /work-programs.
// Caller must apply auth middleware to the group before passing it in —
// every endpoint requires an authenticated context.
func RegisterWorkProgramRoutes(rg *gin.RouterGroup, h *WorkProgramHandler) {
	wp := rg.Group("/work-programs")
	wp.POST("", h.Create)
	wp.GET("", h.List)
	wp.GET("/:id", h.Get)
	wp.POST("/:id/submit", h.Submit)
	wp.POST("/:id/approve", h.Approve)
	wp.POST("/:id/reject", h.Reject)
	wp.POST("/:id/discard", h.Discard)
}
