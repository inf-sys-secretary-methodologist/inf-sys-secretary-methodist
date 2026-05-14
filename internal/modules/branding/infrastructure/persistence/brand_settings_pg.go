// Package persistence provides the PostgreSQL implementation of the
// branding module's BrandSettingsRepository port.
package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/branding/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/branding/domain/repositories"
)

// ErrBrandSettingsMissing is the infrastructure-level sentinel
// returned when the seed row is absent — only happens if migration
// 037 was not applied. Use cases should not catch this; main.go
// should fail fast on first read at boot.
var ErrBrandSettingsMissing = errors.New("branding: seed row missing — migration 037 not applied")

const brandSettingsSelectColumns = `app_name, tagline, logo_url, favicon_url, primary_color, secondary_color, updated_at`

// BrandSettingsRepositoryPG persists the singleton brand settings
// in the brand_settings table. There is no constructor parameter
// besides the DB handle because there is no parameterisation —
// the table holds exactly one row, addressed by id=1.
type BrandSettingsRepositoryPG struct {
	db *sql.DB
}

// NewBrandSettingsRepositoryPG builds the repository against the
// given DB handle. Caller owns the lifecycle of *sql.DB.
func NewBrandSettingsRepositoryPG(db *sql.DB) *BrandSettingsRepositoryPG {
	return &BrandSettingsRepositoryPG{db: db}
}

// Compile-time assertion that the concrete type satisfies the port.
var _ repositories.BrandSettingsRepository = (*BrandSettingsRepositoryPG)(nil)

// Get reads the singleton row. RED stub returns an empty entity
// so admin handler integration tests see "happy path" payload
// without asserting on specific seed values; row-missing detection
// lands in GREEN.
func (r *BrandSettingsRepositoryPG) Get(_ context.Context) (*entities.BrandSettings, error) {
	return entities.RehydrateBrandSettings("", "", "", "", "", "", time.Time{}), nil
}

// Update overwrites the singleton row. RED stub does nothing —
// admin PUT tests will see "200 OK" but a subsequent Get won't
// reflect the write. GREEN restores the real UPDATE.
func (r *BrandSettingsRepositoryPG) Update(_ context.Context, _ *entities.BrandSettings) error {
	return nil
}

// References reserved for the GREEN impl — silence "unused" linter
// until the real Get/Update SQL is restored.
var (
	_ = fmt.Sprintf
	_ = brandSettingsSelectColumns
)
