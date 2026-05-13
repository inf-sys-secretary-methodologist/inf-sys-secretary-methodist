package composio

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
)

// AdminComposioHandler exposes the read-only admin endpoint
// GET /api/admin/composio/config. Mounted under adminGroup with
// RequireRole(system_admin); handler-level role guard is
// intentionally absent because the route-level middleware is the
// canonical gate (mirror к admin/sentry, admin/integrations,
// admin/backups).
type AdminComposioHandler struct {
	uc *AdminComposioUseCase
}

// NewAdminComposioHandler wires the handler against the use case.
// Panics on a nil use case so misconfigured DI fails at
// construction.
func NewAdminComposioHandler(uc *AdminComposioUseCase) *AdminComposioHandler {
	if uc == nil {
		panic("composio: nil AdminComposioUseCase")
	}
	return &AdminComposioHandler{uc: uc}
}

// GetConfig handles GET /api/admin/composio/config. Returns the
// Composio runtime config projection. The API key is never
// returned — only its presence as boolean. EntityID and
// MCPConfigID values are opaque and similarly only their
// presence surfaces.
func (h *AdminComposioHandler) GetConfig(c *gin.Context) {
	cfg := h.uc.GetConfig(c.Request.Context())
	c.JSON(http.StatusOK, response.Success(cfg))
}
