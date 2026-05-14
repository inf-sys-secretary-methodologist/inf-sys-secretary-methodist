package usecases

import (
	"context"

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

// Execute returns the current brand settings via the injected
// repository. Errors propagate unchanged for the handler to map.
func (uc *GetBrandingUseCase) Execute(ctx context.Context) (*entities.BrandSettings, error) {
	return uc.repo.Get(ctx)
}
