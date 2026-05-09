package usecases

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
)

// listDisciplineItemsRepo is the narrow port для list-by-section use case.
type listDisciplineItemsRepo interface {
	ListBySectionID(ctx context.Context, sectionID int64) ([]*entities.DisciplineItem, error)
}

// listDisciplineItemsBySectionLookup is the cross-aggregate guard port:
// confirms the parent section exists before the items query is issued.
// Closes the empty-vs-absent ambiguity flagged in v0.128.1 retroactive
// review — без guard клиенты не отличают "section has 0 items" от
// "section does not exist".
type listDisciplineItemsBySectionLookup interface {
	GetByID(ctx context.Context, id int64) (*entities.Section, error)
}

// ListDisciplineItemsBySectionUseCase returns every discipline item
// attached to a section, ordered for deterministic display. Returns
// repositories.ErrSectionNotFound (not an empty slice) when the parent
// section is absent — caller maps to 404.
type ListDisciplineItemsBySectionUseCase struct {
	repo        listDisciplineItemsRepo
	sectionRepo listDisciplineItemsBySectionLookup
}

// NewListDisciplineItemsBySectionUseCase wires the read use case.
func NewListDisciplineItemsBySectionUseCase(
	repo listDisciplineItemsRepo,
	sectionRepo listDisciplineItemsBySectionLookup,
) *ListDisciplineItemsBySectionUseCase {
	if repo == nil {
		panic("discipline_item: NewListDisciplineItemsBySectionUseCase requires non-nil repo")
	}
	if sectionRepo == nil {
		panic("discipline_item: NewListDisciplineItemsBySectionUseCase requires non-nil sectionRepo")
	}
	return &ListDisciplineItemsBySectionUseCase{repo: repo, sectionRepo: sectionRepo}
}

// Execute returns slice (possibly empty) или wrapped transport error.
func (uc *ListDisciplineItemsBySectionUseCase) Execute(ctx context.Context, sectionID int64) ([]*entities.DisciplineItem, error) {
	return uc.repo.ListBySectionID(ctx, sectionID)
}
