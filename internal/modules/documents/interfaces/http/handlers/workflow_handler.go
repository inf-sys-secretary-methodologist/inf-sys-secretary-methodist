package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
)

// WorkflowHandler exposes the v0.148.0 documents workflow gates
// (Submit / Approve / Reject) as HTTP endpoints. Keeps existing
// DocumentHandler untouched so CRUD endpoints stay independent.
//
// Issue: #227
type WorkflowHandler struct {
	submit  *usecases.SubmitDocumentUseCase
	approve *usecases.ApproveDocumentUseCase
	reject  *usecases.RejectDocumentUseCase
}

// NewWorkflowHandler wires the workflow handler.
func NewWorkflowHandler(
	submit *usecases.SubmitDocumentUseCase,
	approve *usecases.ApproveDocumentUseCase,
	reject *usecases.RejectDocumentUseCase,
) *WorkflowHandler {
	return &WorkflowHandler{submit: submit, approve: approve, reject: reject}
}

// RegisterWorkflowRoutes mounts the three workflow endpoints onto the
// provided group. Mounted under one group для backwards-compat с the
// integration tests; production callers receive admin-tier endpoints
// behind an admin middleware via RegisterAdminWorkflowRoutes.
//
// Routes:
//   - POST /documents/:id/submit          (any authenticated non-student)
//   - POST /admin/documents/:id/approve   (caller-side admin gate)
//   - POST /admin/documents/:id/reject    (caller-side admin gate)
//
// Issue: #227
func RegisterWorkflowRoutes(g *gin.RouterGroup, h *WorkflowHandler) {
	g.POST("/documents/:id/submit", h.Submit)
	g.POST("/admin/documents/:id/approve", h.Approve)
	g.POST("/admin/documents/:id/reject", h.Reject)
}

// RegisterSubmitRoute mounts only POST /:id/submit. Caller already
// scoped к /documents-style group with non-student gate.
func RegisterSubmitRoute(g *gin.RouterGroup, h *WorkflowHandler) {
	g.POST("/:id/submit", h.Submit)
}

// RegisterAdminWorkflowRoutes mounts POST /:id/approve and /:id/reject
// on the caller's admin group (already gated by RequireRole
// (AcademicSecretary, SystemAdmin)). Mirror к curriculum's admin route
// pattern.
func RegisterAdminWorkflowRoutes(g *gin.RouterGroup, h *WorkflowHandler) {
	g.POST("/:id/approve", h.Approve)
	g.POST("/:id/reject", h.Reject)
}

// rejectBody is the request DTO for the Reject endpoint.
type rejectBody struct {
	Reason string `json:"reason"`
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
	c.JSON(http.StatusOK, doc)
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
	c.JSON(http.StatusOK, doc)
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
	c.JSON(http.StatusOK, doc)
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
		errors.Is(err, entities.ErrCannotReject):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, entities.ErrRejectionReasonInvalid):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal"})
	}
}
