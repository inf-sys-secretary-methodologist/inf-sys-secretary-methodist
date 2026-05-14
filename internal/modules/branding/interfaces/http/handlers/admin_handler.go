// Package handlers exposes the branding module HTTP endpoints.
package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/branding/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/branding/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/branding/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
)

// AdminBrandingHandler exposes GET + PUT /api/admin/branding under
// adminGroup. Route-level gate is RequireRole(system_admin); the
// handler does not double-check.
type AdminBrandingHandler struct {
	getUC    *usecases.GetBrandingUseCase
	updateUC *usecases.UpdateBrandingUseCase
}

// NewAdminBrandingHandler wires the handler against the two use
// cases. Panics on nil so misconfigured DI fails at construction.
func NewAdminBrandingHandler(
	getUC *usecases.GetBrandingUseCase,
	updateUC *usecases.UpdateBrandingUseCase,
) *AdminBrandingHandler {
	if getUC == nil {
		panic("branding: nil GetBrandingUseCase")
	}
	if updateUC == nil {
		panic("branding: nil UpdateBrandingUseCase")
	}
	return &AdminBrandingHandler{getUC: getUC, updateUC: updateUC}
}

// GetBranding handles GET /api/admin/branding. Returns the
// current settings. RED stub returns 500 — GREEN restores the
// real call.
func (h *AdminBrandingHandler) GetBranding(c *gin.Context) {
	c.JSON(http.StatusInternalServerError,
		response.ErrorResponse("INTERNAL", "not implemented yet"))
}

// UpdateBranding handles PUT /api/admin/branding. RED stub
// returns 500.
func (h *AdminBrandingHandler) UpdateBranding(c *gin.Context) {
	c.JSON(http.StatusInternalServerError,
		response.ErrorResponse("INTERNAL", "not implemented yet"))
}

// Helpers referenced by GREEN — keep imports honest in RED commit.
var (
	_ = dto.BrandSettingsDTO{}
	_ = entities.ErrInvalidAppName
	_ = errors.Is
)
