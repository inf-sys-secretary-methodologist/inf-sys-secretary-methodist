package integrations

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
)

// AdminIntegrationsHandler exposes the read-only admin endpoint
// GET /api/admin/integrations/config. Mounted under adminGroup
// with RequireRole(system_admin); handler-level role guard is
// intentionally absent because the route-level middleware is the
// canonical gate (mirror к admin/sentry and admin/backups).
type AdminIntegrationsHandler struct {
	uc *AdminIntegrationsUseCase
}

// NewAdminIntegrationsHandler wires the handler against the use
// case. Panics on a nil use case so misconfigured DI fails at
// construction.
func NewAdminIntegrationsHandler(uc *AdminIntegrationsUseCase) *AdminIntegrationsHandler {
	if uc == nil {
		panic("integrations: nil AdminIntegrationsUseCase")
	}
	return &AdminIntegrationsHandler{uc: uc}
}

// GetConfig handles GET /api/admin/integrations/config. Returns
// the combined VAPID + n8n runtime config snapshot. The VAPID
// private key is never returned — only its presence as boolean.
func (h *AdminIntegrationsHandler) GetConfig(c *gin.Context) {
	cfg := h.uc.GetConfig(c.Request.Context())
	c.JSON(http.StatusOK, response.Success(cfg))
}
