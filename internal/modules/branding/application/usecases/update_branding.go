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
//
// RED stub returns nil + nil — handler tests fail on the projected
// fields. GREEN restores the real validation + persistence path.
func (uc *UpdateBrandingUseCase) Execute(
	_ context.Context,
	_, _, _, _, _, _ string,
	_ int64,
) (*entities.BrandSettings, error) {
	return nil, nil
}
