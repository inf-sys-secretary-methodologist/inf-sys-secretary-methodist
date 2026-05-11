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
// — handler-level guard is intentionally absent because the
// route-level middleware is the canonical gate and any handler-level
// duplicate would drift over time.
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

// LogResponse is the JSON projection of one persisted audit-log
// row returned by List. Times are RFC3339 strings (ISO 8601 with
// timezone) so the frontend can parse without ambiguity. Fields stays
// as map[string]any so JSON marshaling preserves the JSONB shape.
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
// Stub: returns 501 Not Implemented; replaced by the matching GREEN.
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
	_ = h.uc
	_, _ = parseListInput(c) // no-op call so the helper compiles unused-clean until GREEN
	_ = mapItems(nil)        // no-op call so the helper compiles unused-clean until GREEN
	c.JSON(http.StatusNotImplemented, response.ErrorResponse("NOT_IMPLEMENTED", "audit-log list: stub"))
}

// parseListInput translates query-string values to a typed ListInput.
// Stub: returns an empty input + the unimplemented error so the
// handler test's parser-shape cases compile.
func parseListInput(c *gin.Context) (ListInput, error) {
	_ = c
	_ = errParseStub
	_ = strconv.Atoi
	_ = time.Parse
	return ListInput{}, errParseStub
}

// mapItems projects domain logs to the JSON wire shape.
// Stub: returns nil — exercised in GREEN.
func mapItems(items []*logging.AuditLog) []LogResponse {
	_ = items
	return nil
}

var errParseStub = errors.New("audit-log parse: not implemented")
