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

// GetBranding handles GET /api/admin/branding — returns the
// current settings projection.
func (h *AdminBrandingHandler) GetBranding(c *gin.Context) {
	settings, err := h.getUC.Execute(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			response.ErrorResponse("INTERNAL", "failed to load branding"))
		return
	}
	c.JSON(http.StatusOK, response.Success(projectDTO(settings)))
}

// UpdateBranding handles PUT /api/admin/branding. Domain
// validation errors map к 422 (well-formed body, conflicting
// value); other errors map к 500.
func (h *AdminBrandingHandler) UpdateBranding(c *gin.Context) {
	var req dto.UpdateBrandSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest,
			response.ErrorResponse("INVALID_BODY", err.Error()))
		return
	}
	settings, err := h.updateUC.Execute(c.Request.Context(), usecases.UpdateBrandingInput{
		AppName:        req.AppName,
		Tagline:        req.Tagline,
		LogoURL:        req.LogoURL,
		FaviconURL:     req.FaviconURL,
		PrimaryColor:   req.PrimaryColor,
		SecondaryColor: req.SecondaryColor,
		ActorUserID:    actorIDFromContext(c),
	})
	if err != nil {
		if isDomainValidationError(err) {
			c.JSON(http.StatusUnprocessableEntity,
				response.ErrorResponse(domainErrorCode(err), err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError,
			response.ErrorResponse("INTERNAL", "failed to update branding"))
		return
	}
	c.JSON(http.StatusOK, response.Success(projectDTO(settings)))
}

func projectDTO(s *entities.BrandSettings) dto.BrandSettingsDTO {
	return dto.BrandSettingsDTO{
		AppName:        s.AppName(),
		Tagline:        s.Tagline(),
		LogoURL:        s.LogoURL(),
		FaviconURL:     s.FaviconURL(),
		PrimaryColor:   s.PrimaryColor(),
		SecondaryColor: s.SecondaryColor(),
		UpdatedAt:      s.UpdatedAt(),
	}
}

// actorIDFromContext reads user_id from the gin context as set by
// the production JWTMiddleware. Returns 0 if absent — audit emit
// uses this as a sentinel for "no actor known".
func actorIDFromContext(c *gin.Context) int64 {
	v, ok := c.Get("user_id")
	if !ok {
		return 0
	}
	switch x := v.(type) {
	case int64:
		return x
	case int:
		return int64(x)
	case float64:
		return int64(x)
	}
	return 0
}

func isDomainValidationError(err error) bool {
	return errors.Is(err, entities.ErrInvalidAppName) ||
		errors.Is(err, entities.ErrInvalidTagline) ||
		errors.Is(err, entities.ErrInvalidColor) ||
		errors.Is(err, entities.ErrInvalidURL)
}

func domainErrorCode(err error) string {
	switch {
	case errors.Is(err, entities.ErrInvalidAppName):
		return "INVALID_APP_NAME"
	case errors.Is(err, entities.ErrInvalidTagline):
		return "INVALID_TAGLINE"
	case errors.Is(err, entities.ErrInvalidColor):
		return "INVALID_COLOR"
	case errors.Is(err, entities.ErrInvalidURL):
		return "INVALID_URL"
	}
	return "INVALID_INPUT"
}
