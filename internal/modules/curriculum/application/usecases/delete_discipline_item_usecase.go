package usecases

import (
	"context"
	"errors"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
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

// Execute performs the delete:
//
//  1. Load item by id; ErrDisciplineItemNotFound → 'not_found' denial.
//  2. Load section + curriculum (two-level chain) для authorize primitives.
//  3. AuthorizeEdit; sentinels mapped к 'forbidden' / 'not_editable'.
//  4. repo.Delete; success emits 'discipline_item.deleted'.
//
// CASCADE migration 035 handles future child rows implicitly (none yet
// в Layer 2; bulk-edit endpoint v0.128.2 будет first consumer).
func (uc *DeleteDisciplineItemUseCase) Execute(ctx context.Context, actorID int64, isAdmin bool, itemID int64) error {
	d, err := uc.repo.GetByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, repositories.ErrDisciplineItemNotFound) {
			emitDisciplineItemAudit(uc.audit, ctx, "discipline_item.delete_denied",
				disciplineItemDenialFields(actorID, itemID, 0, 0, "not_found"))
		}
		return err
	}

	section, err := uc.sectionRepo.GetByID(ctx, d.SectionID())
	if err != nil {
		return err
	}

	cur, err := uc.curriculumRepo.GetByID(ctx, section.CurriculumID())
	if err != nil {
		return err
	}

	if err := d.AuthorizeEdit(actorID, isAdmin, cur.Status(), cur.CreatedBy()); err != nil {
		switch {
		case errors.Is(err, entities.ErrCannotEditDisciplineItem):
			emitDisciplineItemAudit(uc.audit, ctx, "discipline_item.delete_denied",
				disciplineItemDenialFields(actorID, d.ID, d.SectionID(), section.CurriculumID(), "not_editable"))
		case errors.Is(err, entities.ErrDisciplineItemScopeForbidden):
			emitDisciplineItemAudit(uc.audit, ctx, "discipline_item.delete_denied",
				disciplineItemDenialFields(actorID, d.ID, d.SectionID(), section.CurriculumID(), "forbidden"))
		}
		return err
	}

	if err := uc.repo.Delete(ctx, itemID); err != nil {
		return err
	}

	emitDisciplineItemAudit(uc.audit, ctx, "discipline_item.deleted", map[string]any{
		"actor_user_id": actorID,
		"item_id":       d.ID,
		"section_id":    d.SectionID(),
		"curriculum_id": section.CurriculumID(),
	})
	return nil
}
