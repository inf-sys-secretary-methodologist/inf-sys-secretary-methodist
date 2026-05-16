package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
)

// WorkflowHandler exposes the documents workflow gates as HTTP
// endpoints. Grew across multiple phases (#227 v0.148.0 — submit/
// approve/reject; #230 v0.149.0 — register; #231/#232/#233 в очереди).
// Keeps existing DocumentHandler untouched so CRUD endpoints stay
// independent.
type WorkflowHandler struct {
	submit   *usecases.SubmitDocumentUseCase
	approve  *usecases.ApproveDocumentUseCase
	reject   *usecases.RejectDocumentUseCase
	register *usecases.RegisterDocumentUseCase
}

// NewWorkflowHandler wires the workflow handler. Register use case
// is optional — passing nil disables the route (handler returns 501).
func NewWorkflowHandler(
	submit *usecases.SubmitDocumentUseCase,
	approve *usecases.ApproveDocumentUseCase,
	reject *usecases.RejectDocumentUseCase,
	register *usecases.RegisterDocumentUseCase,
) *WorkflowHandler {
	return &WorkflowHandler{submit: submit, approve: approve, reject: reject, register: register}
}

// RegisterSubmitRoute mounts only POST /:id/submit. Caller already
// scoped к /documents-style group with non-student gate.
func RegisterSubmitRoute(g *gin.RouterGroup, h *WorkflowHandler) {
	g.POST("/:id/submit", h.Submit)
}

// RegisterAdminWorkflowRoutes mounts POST /:id/approve, /:id/reject,
// /:id/register on the caller's admin group (already gated by
// RequireRole(AcademicSecretary, SystemAdmin)). Mirror к curriculum's
// admin route pattern.
func RegisterAdminWorkflowRoutes(g *gin.RouterGroup, h *WorkflowHandler) {
	g.POST("/:id/approve", h.Approve)
	g.POST("/:id/reject", h.Reject)
	// v0.149.0 Phase 2 — Register endpoint (#230).
	g.POST("/:id/register", h.Register)
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

// Submit handles POST /api/documents/:id/submit.
//
// Authorization: JWT middleware sets user_id + role in context; the
// usecase enforces the author-or-edit-role rule and returns
// ErrDocumentForbidden when violated.
func (h *WorkflowHandler) Submit(c *gin.Context) {
	id, err := parseDocID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document id"})
		return
	}
	userID, role, ok := readActor(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document id"})
		return
	}
	adminID, _, ok := readActor(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
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
		c.JSON(http.StatusNotImplemented, gin.H{"error": "register usecase not wired"})
		return
	}
	id, err := parseDocID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document id"})
		return
	}
	adminID, _, ok := readActor(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var body registerBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid request body"})
		return
	}
	doc, err := h.register.Execute(c.Request.Context(), adminID, usecases.RegisterDocumentInput{ID: id, Number: body.Number})
	if err != nil {
		mapWorkflowError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(doc))
}

// Reject handles POST /api/admin/documents/:id/reject.
//
// Body: {"reason": "10..500 chars"}. The usecase validates the reason
// VO; invalid → 422 ErrRejectionReasonInvalid.
func (h *WorkflowHandler) Reject(c *gin.Context) {
	id, err := parseDocID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document id"})
		return
	}
	adminID, _, ok := readActor(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var body rejectBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid request body"})
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
		c.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
	case errors.Is(err, usecases.ErrDocumentForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	case errors.Is(err, entities.ErrCannotSubmit),
		errors.Is(err, entities.ErrCannotApprove),
		errors.Is(err, entities.ErrCannotReject),
		errors.Is(err, entities.ErrCannotRegister):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, entities.ErrRejectionReasonInvalid),
		errors.Is(err, entities.ErrInvalidRegistrationNumber):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal"})
	}
}
