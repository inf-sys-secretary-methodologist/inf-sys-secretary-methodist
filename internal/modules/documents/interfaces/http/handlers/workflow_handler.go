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
// provided group. Caller is responsible for pre-gating с admin
// middleware on the /admin sub-paths — keeps routes registrar
// permission-agnostic per `feedback_routes_registrar_adminMW_choice`
// (same-tier set → caller pre-gates).
//
// Routes:
//   - POST /documents/:id/submit          (any authenticated non-student)
//   - POST /admin/documents/:id/approve   (secretary/admin pre-gated)
//   - POST /admin/documents/:id/reject    (secretary/admin pre-gated)
func RegisterWorkflowRoutes(g *gin.RouterGroup, h *WorkflowHandler) {
	g.POST("/documents/:id/submit", h.Submit)
	g.POST("/admin/documents/:id/approve", h.Approve)
	g.POST("/admin/documents/:id/reject", h.Reject)
}

// rejectBody is the request DTO for the Reject endpoint.
type rejectBody struct {
	Reason string `json:"reason"`
}

// Submit handles POST /api/documents/:id/submit (RED stub —
// references helpers in dead branch so golangci `unused` stays
// satisfied without changing test observable behavior; GREEN replaces
// the body).
func (h *WorkflowHandler) Submit(c *gin.Context) {
	if false {
		_, _ = parseDocID(c)
		_ = rejectBody{}
		mapWorkflowError(c, usecases.ErrDocumentNotFound)
	}
	c.JSON(http.StatusNotImplemented, gin.H{"error": "stub"})
}

// Approve handles POST /api/admin/documents/:id/approve (RED stub).
func (h *WorkflowHandler) Approve(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "stub"})
}

// Reject handles POST /api/admin/documents/:id/reject (RED stub).
func (h *WorkflowHandler) Reject(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "stub"})
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
