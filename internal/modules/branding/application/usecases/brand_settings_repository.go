// Package usecases contains branding business logic.
package usecases

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/branding/domain/entities"
)

// BrandSettingsRepository persists the singleton BrandSettings row
// (id=1, enforced by chk_brand_settings_singleton in migration
// 037). The Get / Update pair is sufficient — there is no Save
// (the seed row exists from migration time) and no Delete (brand
// is always present, even if all optional fields are empty).
type BrandSettingsRepository interface {
	// Get returns the current brand settings. A successful migration
	// guarantees the seed row exists, so Get is expected to always
	// return a non-nil entity — implementations should panic if the
	// row is missing because that means migrations were not run.
	Get(ctx context.Context) (*entities.BrandSettings, error)
	// Update overwrites the singleton row with the entity's current
	// state. Implementations issue UPDATE WHERE id=1; RowsAffected
	// must equal 1.
	Update(ctx context.Context, settings *entities.BrandSettings) error
}
