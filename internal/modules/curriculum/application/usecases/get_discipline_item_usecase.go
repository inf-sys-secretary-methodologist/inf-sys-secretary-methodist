package usecases

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
)

// getDisciplineItemRepo is the narrow port для read use case.
type getDisciplineItemRepo interface {
	GetByID(ctx context.Context, id int64) (*entities.DisciplineItem, error)
}

// GetDisciplineItemUseCase fetches a discipline item by id. Reads have
// no authorization beyond route-level RequireNonStudent middleware.
type GetDisciplineItemUseCase struct {
	repo getDisciplineItemRepo
}

// NewGetDisciplineItemUseCase wires the read use case.
func NewGetDisciplineItemUseCase(repo getDisciplineItemRepo) *GetDisciplineItemUseCase {
	if repo == nil {
		panic("discipline_item: NewGetDisciplineItemUseCase requires non-nil repo")
	}
	return &GetDisciplineItemUseCase{repo: repo}
}

// Execute returns item or repo sentinel.
func (uc *GetDisciplineItemUseCase) Execute(ctx context.Context, id int64) (*entities.DisciplineItem, error) {
	return uc.repo.GetByID(ctx, id)
}
