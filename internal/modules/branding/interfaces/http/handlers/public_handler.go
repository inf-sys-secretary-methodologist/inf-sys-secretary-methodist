package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/branding/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
)

// PublicBrandingHandler exposes GET /api/public/branding under
// publicGroup — no auth required, rate-limited via existing
// publicRateLimiter. Login page consumes this to render branded
// chrome BEFORE the user authenticates.
type PublicBrandingHandler struct {
	getUC *usecases.GetBrandingUseCase
}

// NewPublicBrandingHandler wires the public handler against the
// shared GetBrandingUseCase (no separate use case needed — public
// and admin reads return the same projection).
func NewPublicBrandingHandler(getUC *usecases.GetBrandingUseCase) *PublicBrandingHandler {
	if getUC == nil {
		panic("branding: nil GetBrandingUseCase")
	}
	return &PublicBrandingHandler{getUC: getUC}
}

// GetBranding handles GET /api/public/branding. Returns the same
// BrandSettingsDTO projection as the admin GET — no field is
// sensitive, so single shape serves both consumers.
func (h *PublicBrandingHandler) GetBranding(c *gin.Context) {
	settings, err := h.getUC.Execute(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			response.ErrorResponse("INTERNAL", "failed to load branding"))
		return
	}
	c.JSON(http.StatusOK, response.Success(projectDTO(settings)))
}
