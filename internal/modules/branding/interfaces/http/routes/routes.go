// Package routes wires the branding module HTTP endpoints under the
// caller-provided admin and public groups. The split mirrors the
// production main.go shape — admin endpoints sit under a system_admin
// gate, the public endpoint sits under an unauthenticated rate-limited
// group consumed by the login page before the user authenticates.
//
// Per ADR-7 (v0.136.0 plan): branding module owns its own routes via
// this registrar instead of being mounted inline in main.go. v0.137.1
// closes the deviation that left the v0.136.0 release shipping the
// inline mount as a placeholder.
package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/branding/interfaces/http/handlers"
)

// RegisterBrandingRoutes mounts /branding endpoints under the caller-
// provided groups. Both groups arrive pre-gated:
//
//   - adminGroup MUST already have system_admin RequireRole applied;
//     the registrar does not double-check.
//   - publicGroup is unauthenticated; rate-limiter middleware (if any)
//     is the caller's responsibility (the production wiring layers
//     publicRateLimiter on the group before this call).
//
// Admin routes: GET + PUT + OPTIONS /branding.
// Public routes: GET + OPTIONS /branding (same DTO shape — no field
// is sensitive, see brand_settings_dto.go).
func RegisterBrandingRoutes(
	adminGroup *gin.RouterGroup,
	publicGroup *gin.RouterGroup,
	adminHandler *handlers.AdminBrandingHandler,
	publicHandler *handlers.PublicBrandingHandler,
) {
	adminGroup.GET("/branding", adminHandler.GetBranding)
	adminGroup.PUT("/branding", adminHandler.UpdateBranding)
	adminGroup.OPTIONS("/branding", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	publicGroup.GET("/branding", publicHandler.GetBranding)
	publicGroup.OPTIONS("/branding", func(c *gin.Context) { c.Status(http.StatusNoContent) })
}
