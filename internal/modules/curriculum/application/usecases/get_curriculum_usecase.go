package usecases

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
)

// getCurriculumRepo is the narrow port the GetCurriculum use case
// requires from persistence — only GetByID, no writes.
type getCurriculumRepo interface {
	GetByID(ctx context.Context, id int64) (*entities.Curriculum, error)
}

// GetCurriculumUseCase loads a single curriculum by id.
//
// v0.116.0 deliberately omits an entity-level AuthorizeView gate —
// the read scope is the same RequireNonStudent role set across all
// curriculum reads, enforced at the route-middleware boundary.
// Student read with specialty filtering is a future scope (see
// docs/plans/2026-05-06-curriculum-v0116.md ADR-3).
type GetCurriculumUseCase struct {
	repo getCurriculumRepo
}

// NewGetCurriculumUseCase wires the use case. Repo is required.
func NewGetCurriculumUseCase(repo getCurriculumRepo) *GetCurriculumUseCase {
	if repo == nil {
		panic("curriculum: NewGetCurriculumUseCase requires non-nil repo")
	}
	return &GetCurriculumUseCase{repo: repo}
}

// Execute fetches the curriculum with the given id. The not-found
// sentinel from the repo (ErrCurriculumNotFound) propagates so
// handlers can map it to HTTP 404 via errors.Is. Transport errors
// propagate unwrapped.
func (uc *GetCurriculumUseCase) Execute(ctx context.Context, id int64) (*entities.Curriculum, error) {
	return uc.repo.GetByID(ctx, id)
}
