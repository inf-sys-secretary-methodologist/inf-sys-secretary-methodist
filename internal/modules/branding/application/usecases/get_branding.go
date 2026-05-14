package usecases

import (
	"context"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/branding/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/branding/domain/repositories"
)

// GetBrandingUseCase reads the singleton brand settings.
// Mounted under both the admin and the public HTTP groups —
// no field is sensitive so the same projection serves both.
type GetBrandingUseCase struct {
	repo repositories.BrandSettingsRepository
}

// NewGetBrandingUseCase constructs the use case. Panics on nil
// repo so misconfigured DI fails at construction.
func NewGetBrandingUseCase(repo repositories.BrandSettingsRepository) *GetBrandingUseCase {
	if repo == nil {
		panic("branding: nil BrandSettingsRepository")
	}
	return &GetBrandingUseCase{repo: repo}
}

// Execute returns the current brand settings. RED stub returns
// an empty entity so handler integration tests fail on the
// "settings surface" assertions; GREEN restores the repo call.
func (uc *GetBrandingUseCase) Execute(_ context.Context) (*entities.BrandSettings, error) {
	return entities.RehydrateBrandSettings("", "", "", "", "", "", time.Time{}), nil
}
