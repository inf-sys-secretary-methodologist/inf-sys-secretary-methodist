package usecases

import (
	"context"
	"errors"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
)

// getSectionRepo is the narrow port for the read use case.
type getSectionRepo interface {
	GetByID(ctx context.Context, id int64) (*entities.Section, error)
}

// GetSectionUseCase fetches a section by id. Reads have no
// authorization beyond the route-level RequireNonStudent middleware —
// teachers / methodists / secretaries / admins may all view section
// content; per-section ownership filtering is a UI concern.
type GetSectionUseCase struct {
	repo getSectionRepo
}

// NewGetSectionUseCase wires the read use case.
func NewGetSectionUseCase(repo getSectionRepo) *GetSectionUseCase {
	if repo == nil {
		panic("section: NewGetSectionUseCase requires non-nil repo")
	}
	return &GetSectionUseCase{repo: repo}
}

// Execute — implementation lands в GREEN commit (Pair 4).
func (uc *GetSectionUseCase) Execute(ctx context.Context, id int64) (*entities.Section, error) {
	_, _ = ctx, id
	return nil, errors.New("section: GetSectionUseCase.Execute not implemented yet")
}
