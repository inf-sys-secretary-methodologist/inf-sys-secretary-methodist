package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	wpUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
)

// Narrow ports for the revision (лист актуализации) write-workflow. The
// handler depends on the methods it calls, not the concrete use-case
// structs — keeps the DI seam explicit and the handler test-friendly.

// CreateRevisionPort is the narrow port for CreateRevisionUseCase.
type CreateRevisionPort interface {
	Execute(ctx context.Context, actorID int64, actorRole string, in wpUsecases.CreateRevisionInput) (*entities.WorkProgram, error)
}

// SubmitRevisionPort is the narrow port for SubmitRevisionUseCase.
type SubmitRevisionPort interface {
	Execute(ctx context.Context, actorID int64, actorRole string, in wpUsecases.SubmitRevisionInput) (*entities.WorkProgram, error)
}

// ApproveRevisionPort is the narrow port for ApproveRevisionUseCase.
type ApproveRevisionPort interface {
	Execute(ctx context.Context, actorID int64, actorRole string, in wpUsecases.ApproveRevisionInput) (*entities.WorkProgram, error)
}

// RejectRevisionPort is the narrow port for RejectRevisionUseCase.
type RejectRevisionPort interface {
	Execute(ctx context.Context, actorID int64, actorRole string, in wpUsecases.RejectRevisionInput) (*entities.WorkProgram, error)
}

// RevisionHandler exposes the лист-актуализации write-workflow endpoints
// nested under /work-programs/:id/revisions. Read projection lives on the
// parent WorkProgram GET (RevisionDTO); this handler only mutates.
type RevisionHandler struct {
	create  CreateRevisionPort
	submit  SubmitRevisionPort
	approve ApproveRevisionPort
	reject  RejectRevisionPort
}

// NewRevisionHandler wires the handler. All ports are required — a nil
// dependency would surface as a nil-pointer panic under load instead of
// failing loudly at DI wiring time.
func NewRevisionHandler(create CreateRevisionPort, submit SubmitRevisionPort, approve ApproveRevisionPort, reject RejectRevisionPort) *RevisionHandler {
	if create == nil || submit == nil || approve == nil || reject == nil {
		panic("work_program: NewRevisionHandler requires non-nil ports")
	}
	return &RevisionHandler{create: create, submit: submit, approve: approve, reject: reject}
}

// CreateRevisionRequest is the JSON body for POST /work-programs/:id/revisions.
// The author derives from the JWT subject server-side. diff_payload is an
// optional raw-JSON before/after blob — being part of the request body it
// is inherently valid JSON when present.
type CreateRevisionRequest struct {
	ChangeType    string          `json:"change_type"    binding:"required"`
	ChangeSummary string          `json:"change_summary" binding:"required"`
	DiffPayload   json.RawMessage `json:"diff_payload"`
}

// RejectRevisionRequest is the JSON body for the reject endpoint. Reason
// is mandatory (domain enforces non-empty after trim; this fails fast).
type RejectRevisionRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// Create handles POST /api/v1/work-programs/:id/revisions — author
// proposes a draft лист актуализации. Author derives from the JWT
// subject; forbidden is hidden as 404 for non-admins (IDOR).
//
// @Summary Propose a revision (лист актуализации) on a РПД
// @Tags    work-programs
// @Accept  json
// @Produce json
// @Param   id   path int                   true "Work program ID"
// @Param   body body CreateRevisionRequest true "Revision proposal"
// @Success 201 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 422 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/work-programs/{id}/revisions [post]
func (h *RevisionHandler) Create(c *gin.Context) {
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
	var body CreateRevisionRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid request body: "+err.Error()))
		return
	}
	wp, err := h.create.Execute(c.Request.Context(), actorID, role, wpUsecases.CreateRevisionInput{
		WorkProgramID: id,
		ChangeType:    body.ChangeType,
		ChangeSummary: body.ChangeSummary,
		DiffPayload:   body.DiffPayload,
	})
	if err != nil {
		mapWorkProgramError(c, err, !isAdminRole(role))
		return
	}
	c.JSON(http.StatusCreated, response.Success(mapWorkProgram(wp)))
}

// Submit handles POST /api/v1/work-programs/:id/revisions/:rid/submit —
// author moves a draft revision to pending_approval.
//
// @Summary Submit a revision for approval (draft → pending_approval)
// @Tags    work-programs
// @Produce json
// @Param   id  path int true "Work program ID"
// @Param   rid path int true "Revision ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 422 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/work-programs/{id}/revisions/{rid}/submit [post]
func (h *RevisionHandler) Submit(c *gin.Context) {
	actorID, role, id, rid, ok := h.parseIDs(c)
	if !ok {
		return
	}
	wp, err := h.submit.Execute(c.Request.Context(), actorID, role, wpUsecases.SubmitRevisionInput{
		WorkProgramID: id, RevisionID: rid,
	})
	if err != nil {
		mapWorkProgramError(c, err, !isAdminRole(role))
		return
	}
	c.JSON(http.StatusOK, response.Success(mapWorkProgram(wp)))
}

// Approve handles POST /api/v1/work-programs/:id/revisions/:rid/approve —
// methodist approves a pending revision.
//
// @Summary Approve a revision (pending_approval → approved)
// @Tags    work-programs
// @Produce json
// @Param   id  path int true "Work program ID"
// @Param   rid path int true "Revision ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 422 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/work-programs/{id}/revisions/{rid}/approve [post]
func (h *RevisionHandler) Approve(c *gin.Context) {
	actorID, role, id, rid, ok := h.parseIDs(c)
	if !ok {
		return
	}
	wp, err := h.approve.Execute(c.Request.Context(), actorID, role, wpUsecases.ApproveRevisionInput{
		WorkProgramID: id, RevisionID: rid,
	})
	if err != nil {
		mapWorkProgramError(c, err, !isAdminRole(role))
		return
	}
	c.JSON(http.StatusOK, response.Success(mapWorkProgram(wp)))
}

// Reject handles POST /api/v1/work-programs/:id/revisions/:rid/reject —
// methodist rejects a pending revision with a mandatory reason.
//
// @Summary Reject a revision with a reason (pending_approval → rejected)
// @Tags    work-programs
// @Accept  json
// @Produce json
// @Param   id   path int                   true "Work program ID"
// @Param   rid  path int                   true "Revision ID"
// @Param   body body RejectRevisionRequest true "Rejection reason"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 422 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/work-programs/{id}/revisions/{rid}/reject [post]
func (h *RevisionHandler) Reject(c *gin.Context) {
	actorID, role, id, rid, ok := h.parseIDs(c)
	if !ok {
		return
	}
	var body RejectRevisionRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid request body: "+err.Error()))
		return
	}
	wp, err := h.reject.Execute(c.Request.Context(), actorID, role, wpUsecases.RejectRevisionInput{
		WorkProgramID: id, RevisionID: rid, Reason: body.Reason,
	})
	if err != nil {
		mapWorkProgramError(c, err, !isAdminRole(role))
		return
	}
	c.JSON(http.StatusOK, response.Success(mapWorkProgram(wp)))
}

// parseIDs extracts the auth context + :id + :rid path params shared by
// the transition endpoints, writing the appropriate error response and
// returning ok=false on any failure.
func (h *RevisionHandler) parseIDs(c *gin.Context) (actorID int64, role string, id, rid int64, ok bool) {
	actorID, role, ok = authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return 0, "", 0, 0, false
	}
	id, ok = parsePositiveID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid work program id"))
		return 0, "", 0, 0, false
	}
	rid, ok = parsePositiveID(c.Param("rid"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid revision id"))
		return 0, "", 0, 0, false
	}
	return actorID, role, id, rid, true
}

// RegisterRevisionRoutes mounts the 4 revision endpoints under
// /work-programs/:id/revisions. Caller applies auth middleware to the
// group beforehand.
func RegisterRevisionRoutes(rg *gin.RouterGroup, h *RevisionHandler) {
	g := rg.Group("/work-programs/:id/revisions")
	g.POST("", h.Create)
	g.POST("/:rid/submit", h.Submit)
	g.POST("/:rid/approve", h.Approve)
	g.POST("/:rid/reject", h.Reject)
}
