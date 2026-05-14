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

// GetBranding handles GET /api/public/branding. RED stub returns
// 500 — GREEN restores the projection.
func (h *PublicBrandingHandler) GetBranding(c *gin.Context) {
	c.JSON(http.StatusInternalServerError,
		response.ErrorResponse("INTERNAL", "not implemented yet"))
}
