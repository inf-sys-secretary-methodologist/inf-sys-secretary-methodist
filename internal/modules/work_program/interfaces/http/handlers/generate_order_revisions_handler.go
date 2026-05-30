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
// POST /api/v1/minobrnauki-orders/:id/generate-revisions — a methodist
// triggers LLM generation of a draft лист актуализации for every РПД
// affected by the order. The actor + role derive from the JWT subject
// (never the body); the run summary is returned as counts.
//
// @Summary Generate draft revisions for every РПД affected by an order (AI bulk-revision)
// @Tags    minobrnauki-orders
// @Produce json
// @Param   id path int true "Order ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 429 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/minobrnauki-orders/{id}/generate-revisions [post]
func (h *GenerateOrderRevisionsHandler) GenerateRevisions(c *gin.Context) {
	actorID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	id, ok := parsePositiveID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid minobrnauki order id"))
		return
	}
	res, err := h.gen.Execute(c.Request.Context(), actorID, role, id)
	if err != nil {
		mapMinobrnaukiOrderError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(GenerateRevisionsResponse{
		Generated: res.Generated,
		Skipped:   res.Skipped,
		Failures:  res.Failures,
	}))
}

// RegisterGenerateOrderRevisionsRoutes mounts the endpoint. Caller applies
// auth middleware to the group beforehand.
func RegisterGenerateOrderRevisionsRoutes(rg *gin.RouterGroup, h *GenerateOrderRevisionsHandler) {
	rg.POST("/minobrnauki-orders/:id/generate-revisions", h.GenerateRevisions)
}
