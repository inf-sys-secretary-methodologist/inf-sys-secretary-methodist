package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	wpUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
)

// GenerateOrderRevisionsPort is the narrow port for
// GenerateOrderRevisionsUseCase (AI bulk-revision, ADR-12).
type GenerateOrderRevisionsPort interface {
	Execute(ctx context.Context, actorID int64, actorRole string, orderID int64) (wpUsecases.GenerateOrderRevisionsResult, error)
}

// GenerateOrderRevisionsHandler exposes the methodist-triggered
// LLM bulk-revision endpoint. Kept separate from MinobrnaukiOrderHandler
// (mirror of RevisionHandler vs WorkProgramHandler) so the order CRUD
// handler's constructor stays untouched.
type GenerateOrderRevisionsHandler struct {
	gen GenerateOrderRevisionsPort
}

// NewGenerateOrderRevisionsHandler wires the handler. The port is required.
func NewGenerateOrderRevisionsHandler(gen GenerateOrderRevisionsPort) *GenerateOrderRevisionsHandler {
	if gen == nil {
		panic("work_program: NewGenerateOrderRevisionsHandler requires a non-nil port")
	}
	return &GenerateOrderRevisionsHandler{gen: gen}
}

// GenerateRevisionsResponse is the JSON body returned by the endpoint —
// the run summary (counts only; the generated drafts ride along on each
// affected РПД and are reviewed via the revision flow).
type GenerateRevisionsResponse struct {
	Generated int `json:"generated"`
	Skipped   int `json:"skipped"`
	Failures  int `json:"failures"`
}

// GenerateRevisions handles
// POST /api/v1/minobrnauki-orders/:id/generate-revisions — RED stub.
func (h *GenerateOrderRevisionsHandler) GenerateRevisions(c *gin.Context) {
	_ = h.gen
	c.JSON(http.StatusInternalServerError, response.InternalError("not implemented"))
}

// RegisterGenerateOrderRevisionsRoutes mounts the endpoint. Caller applies
// auth middleware to the group beforehand.
func RegisterGenerateOrderRevisionsRoutes(rg *gin.RouterGroup, h *GenerateOrderRevisionsHandler) {
	rg.POST("/minobrnauki-orders/:id/generate-revisions", h.GenerateRevisions)
}
