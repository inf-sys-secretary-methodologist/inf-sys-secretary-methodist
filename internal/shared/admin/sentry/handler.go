package sentry

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
)

// AdminSentryHandler exposes the read-only admin endpoint
// `GET /api/admin/sentry/config`. Mounted under the admin route group
// with RequireRole(system_admin); handler-level role guard is
// intentionally absent because the route-level middleware is the
// canonical gate, mirroring the AdminBackupHandler pattern.
type AdminSentryHandler struct {
	uc *AdminSentryUseCase
}

// NewAdminSentryHandler wires the handler against the use case.
// Panics on a nil use case so misconfigured DI fails at construction.
func NewAdminSentryHandler(uc *AdminSentryUseCase) *AdminSentryHandler {
	if uc == nil {
		panic("sentry: nil AdminSentryUseCase")
	}
	return &AdminSentryHandler{uc: uc}
}

// GetConfig handles GET /api/admin/sentry/config. Returns the runtime
// Sentry configuration snapshot. The DSN value itself is never
// returned — only its presence as a boolean.
func (h *AdminSentryHandler) GetConfig(c *gin.Context) {
	cfg := h.uc.GetConfig(c.Request.Context())
	c.JSON(http.StatusOK, response.Success(cfg))
}
