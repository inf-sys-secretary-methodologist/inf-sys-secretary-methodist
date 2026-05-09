package usecases

import (
	"context"
	"errors"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
)

// listSectionsRepo is the narrow port for the list-by-curriculum use case.
type listSectionsRepo interface {
	ListByCurriculumID(ctx context.Context, curriculumID int64) ([]*entities.Section, error)
}

// ListSectionsByCurriculumUseCase returns every section attached to a
// curriculum, ordered for deterministic display by the repository.
// Reads have no authorization beyond the route-level RequireNonStudent
// middleware — same contract as GetSection.
type ListSectionsByCurriculumUseCase struct {
	repo listSectionsRepo
}

// NewListSectionsByCurriculumUseCase wires the read use case.
func NewListSectionsByCurriculumUseCase(repo listSectionsRepo) *ListSectionsByCurriculumUseCase {
	if repo == nil {
		panic("section: NewListSectionsByCurriculumUseCase requires non-nil repo")
	}
	return &ListSectionsByCurriculumUseCase{repo: repo}
}

// Execute — implementation lands в GREEN commit (Pair 4).
func (uc *ListSectionsByCurriculumUseCase) Execute(ctx context.Context, curriculumID int64) ([]*entities.Section, error) {
	_, _ = ctx, curriculumID
	return nil, errors.New("section: ListSectionsByCurriculumUseCase.Execute not implemented yet")
}
