package http

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
)

// errorKey lives в template_handler.go (same package) — reused here
// for all gin.H error payloads to close the goconst lint cluster.

// WorkflowHandler exposes the documents workflow gates as HTTP
// endpoints. Grew across multiple phases (#227 v0.148.0 — submit/
// approve/reject; #230 v0.149.0 — register; #231 v0.150.0 — routing;
// #232 v0.151.0 — execution; #233 v0.152.0 — archive+resubmit, final).
// Keeps existing DocumentHandler untouched so CRUD endpoints stay
// independent.
type WorkflowHandler struct {
	submit         *usecases.SubmitDocumentUseCase
	approve        *usecases.ApproveDocumentUseCase
	reject         *usecases.RejectDocumentUseCase
	register       *usecases.RegisterDocumentUseCase
	startRouting   *usecases.StartRoutingUseCase
	signVisa       *usecases.SignVisaUseCase
	assignExecutor *usecases.AssignExecutorUseCase
	markExecuted   *usecases.MarkExecutedUseCase
	archive        *usecases.ArchiveDocumentUseCase
	resubmit       *usecases.ResubmitDocumentUseCase
}

// NewWorkflowHandler wires the workflow handler. Phase use cases are
// optional — passing nil disables the matching route (handler returns 501).
func NewWorkflowHandler(
	submit *usecases.SubmitDocumentUseCase,
	approve *usecases.ApproveDocumentUseCase,
	reject *usecases.RejectDocumentUseCase,
	register *usecases.RegisterDocumentUseCase,
	startRouting *usecases.StartRoutingUseCase,
	signVisa *usecases.SignVisaUseCase,
	assignExecutor *usecases.AssignExecutorUseCase,
	markExecuted *usecases.MarkExecutedUseCase,
	archive *usecases.ArchiveDocumentUseCase,
	resubmit *usecases.ResubmitDocumentUseCase,
) *WorkflowHandler {
	return &WorkflowHandler{
		submit:         submit,
		approve:        approve,
		reject:         reject,
		register:       register,
		startRouting:   startRouting,
		signVisa:       signVisa,
		assignExecutor: assignExecutor,
		markExecuted:   markExecuted,
		archive:        archive,
		resubmit:       resubmit,
	}
}

// RegisterSubmitRoute mounts POST /:id/submit + /:id/resubmit. Caller
// already scoped к /documents-style group with non-student gate.
// Resubmit author-or-edit-role gated at usecase boundary (mirror к
// Submit pattern) per ADR-2 #233.
func RegisterSubmitRoute(g *gin.RouterGroup, h *WorkflowHandler) {
	g.POST("/:id/submit", h.Submit)
	// v0.152.0 Phase 5 — Resubmit endpoint (#233). Same non-admin
	// route group as Submit; author OR edit-role allowed.
	g.POST("/:id/resubmit", h.Resubmit)
}

// RegisterAdminWorkflowRoutes mounts POST /:id/approve, /:id/reject,
// /:id/register, /:id/start-routing, /:id/sign-visa, /:id/assign-executor,
// /:id/mark-executed, /:id/archive on the caller's admin group
// (already gated by RequireRole(AcademicSecretary, SystemAdmin)).
// Mirror к curriculum's admin route pattern.
func RegisterAdminWorkflowRoutes(g *gin.RouterGroup, h *WorkflowHandler) {
	g.POST("/:id/approve", h.Approve)
	g.POST("/:id/reject", h.Reject)
	// v0.149.0 Phase 2 — Register endpoint (#230).
	g.POST("/:id/register", h.Register)
	// v0.150.0 Phase 3 — Routing endpoints (#231).
	g.POST("/:id/start-routing", h.StartRouting)
	g.POST("/:id/sign-visa", h.SignVisa)
	// v0.151.0 Phase 4 — Execution endpoints (#232).
	g.POST("/:id/assign-executor", h.AssignExecutor)
	g.POST("/:id/mark-executed", h.MarkExecuted)
	// v0.152.0 Phase 5 — Archive endpoint (#233).
	g.POST("/:id/archive", h.Archive)
}

// rejectBody is the request DTO for the Reject endpoint.
type rejectBody struct {
	Reason string `json:"reason"`
}

// registerBody is the request DTO for the Register endpoint
// (v0.149.0 #230). number трим-валидируется в the entity layer.
type registerBody struct {
	Number string `json:"number"`
}

// assignExecutorBody is the request DTO for the AssignExecutor endpoint
// (v0.151.0 #232). DueDate optional — RFC3339 string ("2026-05-24" or
// "2026-05-24T00:00:00Z"); empty/omitted ⇒ no hard deadline.
type assignExecutorBody struct {
	ExecutorID int64  `json:"executor_id"`
	DueDate    string `json:"due_date,omitempty"`
}

// Submit handles POST /api/documents/:id/submit.
//
// Authorization: JWT middleware sets user_id + role in context; the
// usecase enforces the author-or-edit-role rule and returns
// ErrDocumentForbidden when violated.
func (h *WorkflowHandler) Submit(c *gin.Context) {
	id, err := parseDocID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{errorKey: "invalid document id"})
		return
	}
	userID, role, ok := readActor(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{errorKey: "unauthorized"})
		return
	}
	doc, err := h.submit.Execute(c.Request.Context(), userID, role, usecases.SubmitDocumentInput{ID: id})
	if err != nil {
		mapWorkflowError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(doc))
}

// Approve handles POST /api/admin/documents/:id/approve.
//
// Route-level admin middleware pre-gates the call; the usecase
// enforces the status invariant via the entity Approve method.
func (h *WorkflowHandler) Approve(c *gin.Context) {
	id, err := parseDocID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{errorKey: "invalid document id"})
		return
	}
	adminID, _, ok := readActor(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{errorKey: "unauthorized"})
		return
	}
	doc, err := h.approve.Execute(c.Request.Context(), adminID, usecases.ApproveDocumentInput{ID: id})
	if err != nil {
		mapWorkflowError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(doc))
}

// Register handles POST /api/admin/documents/:id/register.
//
// Body: {"number": "..."}. Entity validates 3..N rune count after
// trim; invalid → 422 ErrInvalidRegistrationNumber.
//
// Issue: #230
func (h *WorkflowHandler) Register(c *gin.Context) {
	if h.register == nil {
		c.JSON(http.StatusNotImplemented, gin.H{errorKey: "register usecase not wired"})
		return
	}
	id, err := parseDocID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{errorKey: "invalid document id"})
		return
	}
	adminID, _, ok := readActor(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{errorKey: "unauthorized"})
		return
	}
	var body registerBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{errorKey: "invalid request body"})
		return
	}
	doc, err := h.register.Execute(c.Request.Context(), adminID, usecases.RegisterDocumentInput{ID: id, Number: body.Number})
	if err != nil {
		mapWorkflowError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(doc))
}

// StartRouting handles POST /api/admin/documents/:id/start-routing.
//
// Route-level admin middleware pre-gates the call; the usecase
// enforces the status invariant via the entity SendToRouting method.
// Body empty — path id + JWT subject identify row + actor.
//
// Issue: #231
func (h *WorkflowHandler) StartRouting(c *gin.Context) {
	if h.startRouting == nil {
		c.JSON(http.StatusNotImplemented, gin.H{errorKey: "start-routing usecase not wired"})
		return
	}
	id, err := parseDocID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{errorKey: "invalid document id"})
		return
	}
	routerID, _, ok := readActor(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{errorKey: "unauthorized"})
		return
	}
	doc, err := h.startRouting.Execute(c.Request.Context(), routerID, usecases.StartRoutingInput{ID: id})
	if err != nil {
		mapWorkflowError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(doc))
}

// SignVisa handles POST /api/admin/documents/:id/sign-visa.
//
// Single-step visa per ADR-1. Route-level admin middleware pre-gates
// the call; entity SignVisa method enforces the status invariant.
// Body empty — same envelope as Approve.
//
// Issue: #231
func (h *WorkflowHandler) SignVisa(c *gin.Context) {
	if h.signVisa == nil {
		c.JSON(http.StatusNotImplemented, gin.H{errorKey: "sign-visa usecase not wired"})
		return
	}
	id, err := parseDocID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{errorKey: "invalid document id"})
		return
	}
	visaID, _, ok := readActor(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{errorKey: "unauthorized"})
		return
	}
	doc, err := h.signVisa.Execute(c.Request.Context(), visaID, usecases.SignVisaInput{ID: id})
	if err != nil {
		mapWorkflowError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(doc))
}

// AssignExecutor handles POST /api/admin/documents/:id/assign-executor.
//
// Body: {"executor_id": int64, "due_date": optional RFC3339 date string}.
// Route-level admin middleware pre-gates the call; the usecase
// validates executor_id > 0 → 422 ErrInvalidExecutor.
//
// Issue: #232
func (h *WorkflowHandler) AssignExecutor(c *gin.Context) {
	if h.assignExecutor == nil {
		c.JSON(http.StatusNotImplemented, gin.H{errorKey: "assign-executor usecase not wired"})
		return
	}
	id, err := parseDocID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{errorKey: "invalid document id"})
		return
	}
	actorID, _, ok := readActor(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{errorKey: "unauthorized"})
		return
	}
	var body assignExecutorBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{errorKey: "invalid request body"})
		return
	}
	var duePtr *time.Time
	if body.DueDate != "" {
		due, perr := parseDueDate(body.DueDate)
		if perr != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{errorKey: "invalid due_date format (use YYYY-MM-DD or RFC3339)"})
			return
		}
		duePtr = &due
	}
	doc, err := h.assignExecutor.Execute(c.Request.Context(), actorID, usecases.AssignExecutorInput{
		ID: id, ExecutorID: body.ExecutorID, DueDate: duePtr,
	})
	if err != nil {
		mapWorkflowError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(doc))
}

// MarkExecuted handles POST /api/admin/documents/:id/mark-executed.
//
// Body-less. Route-level admin middleware pre-gates the call; entity
// MarkExecuted method enforces the status invariant.
//
// Issue: #232
func (h *WorkflowHandler) MarkExecuted(c *gin.Context) {
	if h.markExecuted == nil {
		c.JSON(http.StatusNotImplemented, gin.H{errorKey: "mark-executed usecase not wired"})
		return
	}
	id, err := parseDocID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{errorKey: "invalid document id"})
		return
	}
	actorID, _, ok := readActor(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{errorKey: "unauthorized"})
		return
	}
	doc, err := h.markExecuted.Execute(c.Request.Context(), actorID, usecases.MarkExecutedInput{ID: id})
	if err != nil {
		mapWorkflowError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(doc))
}

// Archive handles POST /api/admin/documents/:id/archive.
//
// Body-less. Route-level admin middleware pre-gates the call; entity
// Archive method enforces the status invariant (must be Executed).
// Terminal — no further transitions allowed after archive.
//
// Issue: #233
func (h *WorkflowHandler) Archive(c *gin.Context) {
	if h.archive == nil {
		c.JSON(http.StatusNotImplemented, gin.H{errorKey: "archive usecase not wired"})
		return
	}
	id, err := parseDocID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{errorKey: "invalid document id"})
		return
	}
	actorID, _, ok := readActor(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{errorKey: "unauthorized"})
		return
	}
	doc, err := h.archive.Execute(c.Request.Context(), actorID, usecases.ArchiveDocumentInput{ID: id})
	if err != nil {
		mapWorkflowError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(doc))
}

// Resubmit handles POST /api/documents/:id/resubmit.
//
// Body-less. Route mounted on non-student group (not admin); usecase
// enforces author-or-edit-role gate per ADR-2. Entity Resubmit method
// enforces the status invariant (must be Rejected) and clears the
// rejection audit fields atomically.
//
// Issue: #233
func (h *WorkflowHandler) Resubmit(c *gin.Context) {
	if h.resubmit == nil {
		c.JSON(http.StatusNotImplemented, gin.H{errorKey: "resubmit usecase not wired"})
		return
	}
	id, err := parseDocID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{errorKey: "invalid document id"})
		return
	}
	userID, role, ok := readActor(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{errorKey: "unauthorized"})
		return
	}
	doc, err := h.resubmit.Execute(c.Request.Context(), userID, role, usecases.ResubmitDocumentInput{ID: id})
	if err != nil {
		mapWorkflowError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(doc))
}

// parseDueDate accepts either YYYY-MM-DD (local midnight UTC) or full
// RFC3339. Returns time.Time + error.
func parseDueDate(s string) (time.Time, error) {
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}
	return time.Parse(time.RFC3339, s)
}

// Reject handles POST /api/admin/documents/:id/reject.
//
// Body: {"reason": "10..500 chars"}. The usecase validates the reason
// VO; invalid → 422 ErrRejectionReasonInvalid.
func (h *WorkflowHandler) Reject(c *gin.Context) {
	id, err := parseDocID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{errorKey: "invalid document id"})
		return
	}
	adminID, _, ok := readActor(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{errorKey: "unauthorized"})
		return
	}
	var body rejectBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{errorKey: "invalid request body"})
		return
	}
	doc, err := h.reject.Execute(c.Request.Context(), adminID, usecases.RejectDocumentInput{ID: id, Reason: body.Reason})
	if err != nil {
		mapWorkflowError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(doc))
}

// readActor extracts (userID, role) from gin context populated by the
// production JWT middleware. Returns ok=false when either key is missing
// или has the wrong type — defense-in-depth against context-key drift.
//
// Mirror к `feedback_handler_context_key_must_match_middleware`: reads
// "user_id" + "role" exactly as auth_middleware.go sets them.
func readActor(c *gin.Context) (int64, entities.UserRole, bool) {
	uidVal, exists := c.Get("user_id")
	if !exists {
		return 0, "", false
	}
	uid, ok := uidVal.(int64)
	if !ok {
		return 0, "", false
	}
	roleVal, exists := c.Get("role")
	if !exists {
		return 0, "", false
	}
	switch r := roleVal.(type) {
	case entities.UserRole:
		return uid, r, true
	case string:
		return uid, entities.UserRole(r), true
	default:
		return 0, "", false
	}
}

// parseID is the shared :id parsing helper.
func parseDocID(c *gin.Context) (int64, error) {
	return strconv.ParseInt(c.Param("id"), 10, 64)
}

// mapWorkflowError maps usecase + domain errors к stable HTTP codes.
// 404 for not-found, 403 for forbidden, 409 for state-machine
// violations, 422 for invalid input.
func mapWorkflowError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecases.ErrDocumentNotFound):
		c.JSON(http.StatusNotFound, gin.H{errorKey: "document not found"})
	case errors.Is(err, usecases.ErrDocumentForbidden):
		c.JSON(http.StatusForbidden, gin.H{errorKey: "forbidden"})
	case errors.Is(err, entities.ErrCannotSubmit),
		errors.Is(err, entities.ErrCannotApprove),
		errors.Is(err, entities.ErrCannotReject),
		errors.Is(err, entities.ErrCannotRegister),
		errors.Is(err, entities.ErrCannotRoute),
		errors.Is(err, entities.ErrCannotSignVisa),
		errors.Is(err, entities.ErrCannotAssignExecutor),
		errors.Is(err, entities.ErrCannotMarkExecuted),
		errors.Is(err, entities.ErrCannotArchive),
		errors.Is(err, entities.ErrCannotResubmit):
		c.JSON(http.StatusConflict, gin.H{errorKey: err.Error()})
	case errors.Is(err, entities.ErrRejectionReasonInvalid),
		errors.Is(err, entities.ErrInvalidRegistrationNumber),
		errors.Is(err, usecases.ErrInvalidExecutor):
		c.JSON(http.StatusUnprocessableEntity, gin.H{errorKey: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{errorKey: "internal"})
	}
}
