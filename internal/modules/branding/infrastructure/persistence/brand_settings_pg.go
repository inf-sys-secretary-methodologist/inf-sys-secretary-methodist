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

// Get reads the singleton row (id = 1). Returns
// ErrBrandSettingsMissing if the seed row is absent — that means
// migration 037 was not applied and main.go should fail fast.
func (r *BrandSettingsRepositoryPG) Get(ctx context.Context) (*entities.BrandSettings, error) {
	query := `SELECT ` + brandSettingsSelectColumns + ` FROM brand_settings WHERE id = 1`

	var (
		appName        string
		tagline        string
		logoURL        string
		faviconURL     string
		primaryColor   string
		secondaryColor string
		updatedAt      time.Time
	)
	err := r.db.QueryRowContext(ctx, query).Scan(
		&appName, &tagline, &logoURL, &faviconURL,
		&primaryColor, &secondaryColor, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrBrandSettingsMissing
		}
		return nil, fmt.Errorf("branding: failed to read settings: %w", err)
	}
	return entities.RehydrateBrandSettings(
		appName, tagline, logoURL, faviconURL,
		primaryColor, secondaryColor, updatedAt,
	), nil
}

// Update overwrites the singleton row. RowsAffected == 0 means the
// seed row is absent — surface as ErrBrandSettingsMissing so callers
// can disambiguate infrastructure rot from a normal write.
func (r *BrandSettingsRepositoryPG) Update(ctx context.Context, settings *entities.BrandSettings) error {
	const query = `UPDATE brand_settings SET app_name = $1, tagline = $2, logo_url = $3, favicon_url = $4, primary_color = $5, secondary_color = $6, updated_at = $7 WHERE id = 1`

	result, err := r.db.ExecContext(ctx, query,
		settings.AppName(),
		settings.Tagline(),
		settings.LogoURL(),
		settings.FaviconURL(),
		settings.PrimaryColor(),
		settings.SecondaryColor(),
		settings.UpdatedAt(),
	)
	if err != nil {
		return fmt.Errorf("branding: failed to update settings: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("branding: failed to inspect update result: %w", err)
	}
	if rows == 0 {
		return ErrBrandSettingsMissing
	}
	return nil
}
