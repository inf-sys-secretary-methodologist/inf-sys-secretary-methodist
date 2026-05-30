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

// Create handles POST /api/v1/work-programs/:id/revisions. STUB.
func (h *RevisionHandler) Create(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, response.InternalError("stub"))
}

// Submit handles POST /api/v1/work-programs/:id/revisions/:rid/submit. STUB.
func (h *RevisionHandler) Submit(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, response.InternalError("stub"))
}

// Approve handles POST /api/v1/work-programs/:id/revisions/:rid/approve. STUB.
func (h *RevisionHandler) Approve(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, response.InternalError("stub"))
}

// Reject handles POST /api/v1/work-programs/:id/revisions/:rid/reject. STUB.
func (h *RevisionHandler) Reject(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, response.InternalError("stub"))
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
