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

// publishedAtLayout is the wire format for MinobrnaukiOrder.published_at —
// a calendar date with no time component (the DB column is DATE).
const publishedAtLayout = "2006-01-02"

// ===== Narrow ports =====

// RecordMinobrnaukiOrderPort is the narrow port for RecordMinobrnaukiOrderUseCase.
type RecordMinobrnaukiOrderPort interface {
	Execute(ctx context.Context, actorID int64, actorRole string, in wpUsecases.RecordMinobrnaukiOrderInput) (*entities.MinobrnaukiOrder, error)
}

// GetMinobrnaukiOrderPort is the narrow port for GetMinobrnaukiOrderUseCase.
// It returns the order plus the ids of the work programs it affects.
type GetMinobrnaukiOrderPort interface {
	Execute(ctx context.Context, actorRole string, id int64) (*entities.MinobrnaukiOrder, []int64, error)
}

// ListMinobrnaukiOrdersPort is the narrow port for ListMinobrnaukiOrdersUseCase.
type ListMinobrnaukiOrdersPort interface {
	Execute(ctx context.Context, actorRole string, filter repositories.MinobrnaukiOrderListFilter) (repositories.MinobrnaukiOrderListResult, error)
}

// MinobrnaukiOrderHandler exposes the приказ Минобрнауки endpoints (ADR-11):
// record (POST), list (GET), get-by-id (GET /:id).
type MinobrnaukiOrderHandler struct {
	record RecordMinobrnaukiOrderPort
	get    GetMinobrnaukiOrderPort
	list   ListMinobrnaukiOrdersPort
}

// NewMinobrnaukiOrderHandler wires the handler. All ports are required —
// a nil dependency fails loudly at DI wiring rather than as a nil panic
// under load (mirror of NewWorkProgramHandler).
func NewMinobrnaukiOrderHandler(
	record RecordMinobrnaukiOrderPort,
	get GetMinobrnaukiOrderPort,
	list ListMinobrnaukiOrdersPort,
) *MinobrnaukiOrderHandler {
	if record == nil || get == nil || list == nil {
		panic("work_program: NewMinobrnaukiOrderHandler requires non-nil ports")
	}
	return &MinobrnaukiOrderHandler{record: record, get: get, list: list}
}

// ===== Request DTOs =====

// RecordMinobrnaukiOrderRequest is the JSON body for POST /minobrnauki-orders.
// The actor (→ uploaded_by) is derived from the JWT subject server-side,
// never from the body. published_at is a calendar date (YYYY-MM-DD).
type RecordMinobrnaukiOrderRequest struct {
	OrderNumber            string  `json:"order_number"               binding:"required"`
	Title                  string  `json:"title"                      binding:"required"`
	PublishedAt            string  `json:"published_at"               binding:"required"`
	DocumentID             *int64  `json:"document_id"`
	ChangeScope            string  `json:"change_scope"               binding:"required"`
	Summary                string  `json:"summary"`
	AffectedWorkProgramIDs []int64 `json:"affected_work_program_ids"`
}

// ===== Response DTOs =====

// MinobrnaukiOrderDTO is the full public shape for one order, including
// the affected-work-program ids. published_at is a date string; created_at
// is RFC 3339.
type MinobrnaukiOrderDTO struct {
	ID                     int64   `json:"id"`
	OrderNumber            string  `json:"order_number"`
	Title                  string  `json:"title"`
	PublishedAt            string  `json:"published_at"`
	DocumentID             *int64  `json:"document_id,omitempty"`
	ChangeScope            string  `json:"change_scope"`
	Summary                string  `json:"summary,omitempty"`
	UploadedBy             int64   `json:"uploaded_by"`
	CreatedAt              string  `json:"created_at"`
	AffectedWorkProgramIDs []int64 `json:"affected_work_program_ids"`
}

// MinobrnaukiOrderSummaryDTO is the list-row projection — order fields
// without the affected set (kept off the list to stay cheap).
type MinobrnaukiOrderSummaryDTO struct {
	ID          int64  `json:"id"`
	OrderNumber string `json:"order_number"`
	Title       string `json:"title"`
	PublishedAt string `json:"published_at"`
	DocumentID  *int64 `json:"document_id,omitempty"`
	ChangeScope string `json:"change_scope"`
	Summary     string `json:"summary,omitempty"`
	UploadedBy  int64  `json:"uploaded_by"`
	CreatedAt   string `json:"created_at"`
}

// MinobrnaukiOrdersListResponse is the page response shape.
type MinobrnaukiOrdersListResponse struct {
	Items []MinobrnaukiOrderSummaryDTO `json:"items"`
	Total int                          `json:"total"`
}

// ===== Mappers =====

func mapMinobrnaukiOrder(o *entities.MinobrnaukiOrder, affected []int64) MinobrnaukiOrderDTO {
	if affected == nil {
		affected = []int64{}
	}
	return MinobrnaukiOrderDTO{
		ID:                     o.ID(),
		OrderNumber:            o.OrderNumber(),
		Title:                  o.Title(),
		PublishedAt:            o.PublishedAt().Format(publishedAtLayout),
		DocumentID:             o.DocumentID(),
		ChangeScope:            string(o.ChangeScope()),
		Summary:                o.Summary(),
		UploadedBy:             o.UploadedBy(),
		CreatedAt:              o.CreatedAt().Format(time.RFC3339),
		AffectedWorkProgramIDs: affected,
	}
}

func mapMinobrnaukiOrderSummary(it repositories.MinobrnaukiOrderListItem) MinobrnaukiOrderSummaryDTO {
	return MinobrnaukiOrderSummaryDTO{
		ID:          it.ID,
		OrderNumber: it.OrderNumber,
		Title:       it.Title,
		PublishedAt: it.PublishedAt.Format(publishedAtLayout),
		DocumentID:  it.DocumentID,
		ChangeScope: string(it.ChangeScope),
		Summary:     it.Summary,
		UploadedBy:  it.UploadedBy,
		CreatedAt:   it.CreatedAt.Format(time.RFC3339),
	}
}

// mapMinobrnaukiOrderError maps domain / repository sentinels to HTTP
// status. Order visibility is role-flat (not ownership-scoped), so a
// denial is a true 403 rather than an enumeration oracle — no IDOR
// collapse to 404 is needed here.
func mapMinobrnaukiOrderError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repositories.ErrMinobrnaukiOrderNotFound):
		c.JSON(http.StatusNotFound, response.NotFound("minobrnauki order"))
	case errors.Is(err, domain.ErrMinobrnaukiOrderScopeForbidden):
		c.JSON(http.StatusForbidden, response.Forbidden("not authorized to operate on minobrnauki orders"))
	case errors.Is(err, domain.ErrGenerationRateLimited):
		c.JSON(http.StatusTooManyRequests, response.ErrorResponse("RATE_LIMITED", err.Error()))
	case errors.Is(err, domain.ErrInvalidMinobrnaukiOrder):
		c.JSON(http.StatusUnprocessableEntity, response.ErrorResponse("INVALID_MINOBRNAUKI_ORDER", err.Error()))
	default:
		c.JSON(http.StatusInternalServerError, response.InternalError("internal error"))
	}
}

// ===== Endpoints =====

// Record handles POST /api/v1/minobrnauki-orders.
// @Summary Record a Минобрнауки order (приказ) and optionally mark affected РПД
// @Tags    minobrnauki-orders
// @Accept  json
// @Produce json
// @Param   body body RecordMinobrnaukiOrderRequest true "Order payload"
// @Success 201 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 422 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/minobrnauki-orders [post]
func (h *MinobrnaukiOrderHandler) Record(c *gin.Context) {
	actorID, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	var body RecordMinobrnaukiOrderRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid request body: "+err.Error()))
		return
	}
	publishedAt, err := time.Parse(publishedAtLayout, body.PublishedAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid published_at; expected YYYY-MM-DD"))
		return
	}

	order, err := h.record.Execute(c.Request.Context(), actorID, role, wpUsecases.RecordMinobrnaukiOrderInput{
		OrderNumber:            body.OrderNumber,
		Title:                  body.Title,
		PublishedAt:            publishedAt,
		DocumentID:             body.DocumentID,
		ChangeScope:            body.ChangeScope,
		Summary:                body.Summary,
		AffectedWorkProgramIDs: body.AffectedWorkProgramIDs,
	})
	if err != nil {
		mapMinobrnaukiOrderError(c, err)
		return
	}
	c.JSON(http.StatusCreated, response.Success(mapMinobrnaukiOrder(order, body.AffectedWorkProgramIDs)))
}

// Get handles GET /api/v1/minobrnauki-orders/:id.
// @Summary Fetch one Минобрнауки order with the ids of the work programs it affects
// @Tags    minobrnauki-orders
// @Produce json
// @Param   id path int true "Order ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/minobrnauki-orders/{id} [get]
func (h *MinobrnaukiOrderHandler) Get(c *gin.Context) {
	_, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	id, ok := parsePositiveID(c.Param("id"))
	if !ok {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid minobrnauki order id"))
		return
	}
	order, affected, err := h.get.Execute(c.Request.Context(), role, id)
	if err != nil {
		mapMinobrnaukiOrderError(c, err)
		return
	}
	c.JSON(http.StatusOK, response.Success(mapMinobrnaukiOrder(order, affected)))
}

// List handles GET /api/v1/minobrnauki-orders.
// @Summary List Минобрнауки orders, filterable by change_scope / uploaded_by
// @Tags    minobrnauki-orders
// @Produce json
// @Param   change_scope query string false "minor / major"
// @Param   uploaded_by  query int    false "Uploader user id filter"
// @Param   limit        query int    false "Page size (default 50, max 200)"
// @Param   offset       query int    false "Page offset"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Security BearerAuth
// @Router /api/v1/minobrnauki-orders [get]
func (h *MinobrnaukiOrderHandler) List(c *gin.Context) {
	_, role, ok := authContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("missing user context"))
		return
	}
	filter := repositories.MinobrnaukiOrderListFilter{}
	if s := c.Query("change_scope"); s != "" {
		sc := domain.MinobrnaukiOrderChangeScope(s)
		filter.ChangeScope = &sc
	}
	if v := c.Query("uploaded_by"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			filter.UploadedBy = &n
		}
	}
	filter.Limit, _ = strconv.Atoi(c.Query("limit"))
	filter.Offset, _ = strconv.Atoi(c.Query("offset"))

	res, err := h.list.Execute(c.Request.Context(), role, filter)
	if err != nil {
		mapMinobrnaukiOrderError(c, err)
		return
	}
	out := MinobrnaukiOrdersListResponse{
		Items: make([]MinobrnaukiOrderSummaryDTO, 0, len(res.Items)),
		Total: res.Total,
	}
	for _, it := range res.Items {
		out.Items = append(out.Items, mapMinobrnaukiOrderSummary(it))
	}
	c.JSON(http.StatusOK, response.Success(out))
}

// RegisterMinobrnaukiOrderRoutes mounts the three endpoints under
// /minobrnauki-orders. Caller must apply auth middleware to the group —
// every endpoint requires an authenticated context.
func RegisterMinobrnaukiOrderRoutes(rg *gin.RouterGroup, h *MinobrnaukiOrderHandler) {
	g := rg.Group("/minobrnauki-orders")
	g.POST("", h.Record)
	g.GET("", h.List)
	g.GET("/:id", h.Get)
}
