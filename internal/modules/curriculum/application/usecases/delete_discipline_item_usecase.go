package usecases

import (
	"context"
	"errors"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
)

// deleteDisciplineItemRepo is the narrow port (item GetByID + Delete).
type deleteDisciplineItemRepo interface {
	GetByID(ctx context.Context, id int64) (*entities.DisciplineItem, error)
	Delete(ctx context.Context, id int64) error
}

// deleteDisciplineItemSectionLookup is one-hop cross-aggregate port.
type deleteDisciplineItemSectionLookup interface {
	GetByID(ctx context.Context, id int64) (*entities.Section, error)
}

// deleteDisciplineItemCurriculumLookup is the second-hop cross-aggregate
// port для status + author primitives.
type deleteDisciplineItemCurriculumLookup interface {
	GetByID(ctx context.Context, id int64) (*entities.Curriculum, error)
}

// DeleteDisciplineItemUseCase removes an item after two-level
// authorization. Hard-delete per ADR-4; audit trail captures forensics.
type DeleteDisciplineItemUseCase struct {
	repo           deleteDisciplineItemRepo
	sectionRepo    deleteDisciplineItemSectionLookup
	curriculumRepo deleteDisciplineItemCurriculumLookup
	audit          AuditSink
}

// NewDeleteDisciplineItemUseCase wires the use case.
func NewDeleteDisciplineItemUseCase(
	repo deleteDisciplineItemRepo,
	sectionRepo deleteDisciplineItemSectionLookup,
	curriculumRepo deleteDisciplineItemCurriculumLookup,
	audit AuditSink,
) *DeleteDisciplineItemUseCase {
	if repo == nil {
		panic("discipline_item: NewDeleteDisciplineItemUseCase requires non-nil repo")
	}
	if sectionRepo == nil {
		panic("discipline_item: NewDeleteDisciplineItemUseCase requires non-nil sectionRepo")
	}
	if curriculumRepo == nil {
		panic("discipline_item: NewDeleteDisciplineItemUseCase requires non-nil curriculumRepo")
	}
	return &DeleteDisciplineItemUseCase{repo: repo, sectionRepo: sectionRepo, curriculumRepo: curriculumRepo, audit: audit}
}

// Execute — implementation lands в GREEN commit (Pair 4).
func (uc *DeleteDisciplineItemUseCase) Execute(ctx context.Context, actorID int64, isAdmin bool, itemID int64) error {
	_ = isAdmin
	emitDisciplineItemAudit(uc.audit, ctx, "discipline_item.unimplemented",
		disciplineItemDenialFields(actorID, itemID, 0, 0, "stub"))
	return errors.New("discipline_item: DeleteDisciplineItemUseCase.Execute not implemented yet")
}
