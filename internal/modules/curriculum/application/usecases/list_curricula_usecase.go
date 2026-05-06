package usecases

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
)

// Pagination policy for the use case. Centralised here (rather than
// in the handler) so any future caller — internal scheduler, batch
// job, or alternate transport — inherits the same boundedness
// without re-implementing the limits.
const (
	defaultListLimit = 50
	maxListLimit     = 200
)

// ListCurriculaInput is the public DTO for the list use case. It
// mirrors the persistence-layer CurriculumListFilter shape but lives
// in the use-case package so handlers don't import repository types
// directly.
type ListCurriculaInput struct {
	Status    *entities.CurriculumStatus
	Year      *int
	Specialty string
	CreatedBy *int64
	Limit     int
	Offset    int
}

// CurriculaPage is the public response shape: the page of items
// together with the unfiltered total so the UI can render
// pagination controls without a second round-trip.
type CurriculaPage struct {
	Items []*entities.Curriculum
	Total int
}

// listCurriculaRepo is the narrow port from the use case to
// persistence — only List, no writes.
type listCurriculaRepo interface {
	List(ctx context.Context, filter repositories.CurriculumListFilter) (repositories.CurriculumListResult, error)
}

// ListCurriculaUseCase loads a page of curricula matching the input
// filters.
type ListCurriculaUseCase struct {
	repo listCurriculaRepo
}

// NewListCurriculaUseCase wires the use case. Repo is required.
func NewListCurriculaUseCase(repo listCurriculaRepo) *ListCurriculaUseCase {
	if repo == nil {
		panic("curriculum: NewListCurriculaUseCase requires non-nil repo")
	}
	return &ListCurriculaUseCase{repo: repo}
}

// Execute applies pagination defaults / clamps and dispatches to
// the repository.
func (uc *ListCurriculaUseCase) Execute(ctx context.Context, in ListCurriculaInput) (CurriculaPage, error) {
	limit := in.Limit
	if limit <= 0 {
		limit = defaultListLimit
	}
	if limit > maxListLimit {
		limit = maxListLimit
	}
	offset := in.Offset
	if offset < 0 {
		offset = 0
	}

	res, err := uc.repo.List(ctx, repositories.CurriculumListFilter{
		Status:    in.Status,
		Year:      in.Year,
		Specialty: in.Specialty,
		CreatedBy: in.CreatedBy,
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		return CurriculaPage{}, err
	}
	return CurriculaPage{Items: res.Items, Total: res.Total}, nil
}
