package auditlog

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

// AdminAuditLogHandler exposes GET /api/admin/audit-logs.
// Mounted under the admin route group with RequireRole(system_admin)
// — handler-level role guard is intentionally absent because the
// route-level middleware is the canonical gate and a handler-side
// duplicate would drift over time. Integration tests pin the
// middleware-handler pair to keep the gate honest.
type AdminAuditLogHandler struct {
	uc *AdminAuditLogUseCase
}

// NewAdminAuditLogHandler wires the handler against the use case.
// Panics on a nil use case so misconfigured DI fails at construction.
func NewAdminAuditLogHandler(uc *AdminAuditLogUseCase) *AdminAuditLogHandler {
	if uc == nil {
		panic("auditlog: nil AdminAuditLogUseCase")
	}
	return &AdminAuditLogHandler{uc: uc}
}

// LogResponse is the JSON projection of one persisted audit-log row.
// Times are RFC3339 strings so the frontend parses without timezone
// ambiguity. Fields stays as map[string]any so JSON marshaling
// preserves the JSONB column shape one-to-one.
type LogResponse struct {
	ID            int64          `json:"id"`
	CreatedAt     string         `json:"created_at"`
	Action        string         `json:"action"`
	Resource      string         `json:"resource"`
	ActorUserID   *int64         `json:"actor_user_id"`
	ActorIP       *string        `json:"actor_ip"`
	CorrelationID *string        `json:"correlation_id"`
	Fields        map[string]any `json:"fields"`
}

// List handles GET /api/admin/audit-logs.
//
// @Summary List audit logs (admin only)
// @Tags admin
// @Produce json
// @Param action query string false "Filter by exact action match"
// @Param resource query string false "Filter by exact resource match"
// @Param user_id query int false "Filter by actor user id"
// @Param from query string false "Inclusive lower bound on created_at (RFC3339)"
// @Param to query string false "Exclusive upper bound on created_at (RFC3339)"
// @Param limit query int false "Page size (default 50, max 200)"
// @Param offset query int false "Pagination offset"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /api/admin/audit-logs [get]
func (h *AdminAuditLogHandler) List(c *gin.Context) {
	input, err := parseListInput(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	result, err := h.uc.List(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, ErrInvalidTimeRange) {
			c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, response.InternalError("failed to list audit logs"))
		return
	}

	items := mapItems(result.Items)
	// Single source of truth: ClampLimit is the use-case clamp policy.
	// Without this the handler would re-implement Default/Max bounds
	// and drift away from the use case over time.
	perPage := ClampLimit(input.Limit)
	page := input.Offset/perPage + 1
	totalPages := 0
	if result.Total > 0 {
		totalPages = (result.Total + perPage - 1) / perPage
	}

	c.JSON(http.StatusOK, response.List(items, response.Pagination{
		Page:       page,
		PerPage:    perPage,
		Total:      result.Total,
		TotalPages: totalPages,
	}))
}

// parseListInput translates query-string values into a typed ListInput.
// Returns a user-facing error message ready to flow into 400 BadRequest
// — no stack traces or internal identifiers leak.
func parseListInput(c *gin.Context) (ListInput, error) {
	input := ListInput{
		Action:   c.Query("action"),
		Resource: c.Query("resource"),
	}

	if v := c.Query("user_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			return ListInput{}, errors.New("invalid user_id: positive integer required")
		}
		input.UserID = &id
	}

	if v := c.Query("from"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return ListInput{}, errors.New("invalid from: RFC3339 timestamp required")
		}
		input.From = &t
	}

	if v := c.Query("to"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return ListInput{}, errors.New("invalid to: RFC3339 timestamp required")
		}
		input.To = &t
	}

	if v := c.Query("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			return ListInput{}, errors.New("invalid limit: non-negative integer required")
		}
		input.Limit = n
	}

	if v := c.Query("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			return ListInput{}, errors.New("invalid offset: non-negative integer required")
		}
		input.Offset = n
	}

	return input, nil
}

// mapItems projects domain logs to the JSON wire shape. Returns an
// empty (non-nil) slice when the input is empty so the JSON body
// renders `"data": []` rather than `"data": null`, matching the
// pagination Meta when Total is zero.
func mapItems(items []*logging.AuditLog) []LogResponse {
	out := make([]LogResponse, 0, len(items))
	for _, log := range items {
		out = append(out, LogResponse{
			ID:            log.ID,
			CreatedAt:     log.CreatedAt.UTC().Format(time.RFC3339),
			Action:        log.Action,
			Resource:      log.Resource,
			ActorUserID:   log.ActorUserID,
			ActorIP:       log.ActorIP,
			CorrelationID: log.CorrelationID,
			Fields:        log.Fields,
		})
	}
	return out
}
