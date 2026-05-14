package usecases

import (
	"context"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/branding/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/branding/domain/repositories"
)

// Clock is the narrow port for current-time injection. Lets tests
// substitute deterministic clocks without rebuilding the entire
// use case tree.
type Clock interface {
	Now() time.Time
}

// SystemClock returns time.Now() — production wiring in main.go.
type SystemClock struct{}

// Now returns the current wall-clock time.
func (SystemClock) Now() time.Time { return time.Now() }

// UpdateBrandingInput is the use case input — replaces a 7-positional
// Execute signature with a named-field DTO (mirror к assignments /
// curriculum XxxInput convention). Frontend / handler maps from the
// HTTP DTO; ActorUserID is sourced from the JWT context.
type UpdateBrandingInput struct {
	AppName        string
	Tagline        string
	LogoURL        string
	FaviconURL     string
	PrimaryColor   string
	SecondaryColor string
	ActorUserID    int64
}

// UpdateBrandingUseCase replaces the singleton brand settings with
// the caller-provided values. The new entity is constructed via
// NewBrandSettings so all domain invariants run; the repo Update
// then persists. An audit event "brand.updated" is emitted on
// success (fire-and-forget — failure is not propagated).
type UpdateBrandingUseCase struct {
	repo  repositories.BrandSettingsRepository
	clock Clock
	audit AuditSink
}

// NewUpdateBrandingUseCase constructs the use case. Panics on a
// nil repo. clock defaults to SystemClock if nil. audit may be
// nil — emission is skipped in that case (test-friendly).
func NewUpdateBrandingUseCase(
	repo repositories.BrandSettingsRepository,
	clock Clock,
	audit AuditSink,
) *UpdateBrandingUseCase {
	if repo == nil {
		panic("branding: nil BrandSettingsRepository")
	}
	if clock == nil {
		clock = SystemClock{}
	}
	return &UpdateBrandingUseCase{repo: repo, clock: clock, audit: audit}
}

// Execute validates the input via NewBrandSettings, persists via
// repo.Update, and emits a brand.updated audit event on success.
// Returns the resulting entity. Domain validation errors propagate
// for the handler to map к 422; repo errors propagate as-is.
func (uc *UpdateBrandingUseCase) Execute(
	ctx context.Context,
	in UpdateBrandingInput,
) (*entities.BrandSettings, error) {
	now := uc.clock.Now()
	settings, err := entities.NewBrandSettings(
		in.AppName, in.Tagline, in.LogoURL, in.FaviconURL,
		in.PrimaryColor, in.SecondaryColor, now,
	)
	if err != nil {
		return nil, err
	}
	if err := uc.repo.Update(ctx, settings); err != nil {
		return nil, err
	}
	uc.emitAudit(ctx, settings, in.ActorUserID)
	return settings, nil
}

// emitAudit logs a brand.updated forensic event with the
// resulting field snapshot. Nil-safe: callers passing a nil
// AuditSink skip emission entirely (test-friendly).
func (uc *UpdateBrandingUseCase) emitAudit(ctx context.Context, settings *entities.BrandSettings, actorUserID int64) {
	if uc.audit == nil {
		return
	}
	uc.audit.LogAuditEvent(ctx, "brand.updated", "brand", map[string]any{
		"actor_user_id":   actorUserID,
		"app_name":        settings.AppName(),
		"tagline":         settings.Tagline(),
		"logo_url":        settings.LogoURL(),
		"favicon_url":     settings.FaviconURL(),
		"primary_color":   settings.PrimaryColor(),
		"secondary_color": settings.SecondaryColor(),
	})
}
