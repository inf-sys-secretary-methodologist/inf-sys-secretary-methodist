package usecases

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
)

// listDisciplineItemsRepo is the narrow port для list-by-section use case.
type listDisciplineItemsRepo interface {
	ListBySectionID(ctx context.Context, sectionID int64) ([]*entities.DisciplineItem, error)
}

// ListDisciplineItemsBySectionUseCase returns every discipline item
// attached to a section, ordered for deterministic display.
type ListDisciplineItemsBySectionUseCase struct {
	repo listDisciplineItemsRepo
}

// NewListDisciplineItemsBySectionUseCase wires the read use case.
func NewListDisciplineItemsBySectionUseCase(repo listDisciplineItemsRepo) *ListDisciplineItemsBySectionUseCase {
	if repo == nil {
		panic("discipline_item: NewListDisciplineItemsBySectionUseCase requires non-nil repo")
	}
	return &ListDisciplineItemsBySectionUseCase{repo: repo}
}

// Execute returns slice (possibly empty) или wrapped transport error.
func (uc *ListDisciplineItemsBySectionUseCase) Execute(ctx context.Context, sectionID int64) ([]*entities.DisciplineItem, error) {
	return uc.repo.ListBySectionID(ctx, sectionID)
}
