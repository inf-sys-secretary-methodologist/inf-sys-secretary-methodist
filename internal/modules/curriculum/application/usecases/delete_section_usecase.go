package usecases

import (
	"context"
	"errors"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
)

// deleteSectionRepo is the narrow port for the delete use case:
// fetch (for authorization scope) + delete.
type deleteSectionRepo interface {
	GetByID(ctx context.Context, id int64) (*entities.Section, error)
	Delete(ctx context.Context, id int64) error
}

// deleteSectionCurriculumLookup is the cross-aggregate read port for
// authorization (curriculum status + author primitives).
type deleteSectionCurriculumLookup interface {
	GetByID(ctx context.Context, id int64) (*entities.Curriculum, error)
}

// DeleteSectionUseCase removes a section after authorization. CASCADE
// in migration 034 handles eventual child-item cleanup (v0.128.1+).
// Hard-delete per ADR-4: undo via UI confirm dialog; audit trail
// captures forensics.
type DeleteSectionUseCase struct {
	repo           deleteSectionRepo
	curriculumRepo deleteSectionCurriculumLookup
	audit          AuditSink
}

// NewDeleteSectionUseCase wires the use case.
func NewDeleteSectionUseCase(
	repo deleteSectionRepo,
	curriculumRepo deleteSectionCurriculumLookup,
	audit AuditSink,
) *DeleteSectionUseCase {
	if repo == nil {
		panic("section: NewDeleteSectionUseCase requires non-nil repo")
	}
	if curriculumRepo == nil {
		panic("section: NewDeleteSectionUseCase requires non-nil curriculumRepo")
	}
	return &DeleteSectionUseCase{repo: repo, curriculumRepo: curriculumRepo, audit: audit}
}

// Execute performs the delete:
//
//  1. Load section by id; ErrSectionNotFound → 'not_found' denial.
//  2. Load parent curriculum; missing → propagate без audit
//     (orphaned-row defense, not policy).
//  3. AuthorizeEdit; sentinels mapped to 'forbidden' / 'not_editable'.
//  4. Delete via repo.
//
// CASCADE in migration 034 handles child items implicitly when sections
// disappear (v0.128.1 adds the items table).
func (uc *DeleteSectionUseCase) Execute(ctx context.Context, actorID int64, isAdmin bool, sectionID int64) error {
	s, err := uc.repo.GetByID(ctx, sectionID)
	if err != nil {
		if errors.Is(err, repositories.ErrSectionNotFound) {
			emitSectionAudit(uc.audit, ctx, "section.delete_denied",
				sectionDenialFields(actorID, sectionID, 0, "not_found"))
		}
		return err
	}

	cur, err := uc.curriculumRepo.GetByID(ctx, s.CurriculumID())
	if err != nil {
		return err
	}

	if err := s.AuthorizeEdit(actorID, isAdmin, cur.Status(), cur.CreatedBy()); err != nil {
		switch {
		case errors.Is(err, entities.ErrCannotEditSection):
			emitSectionAudit(uc.audit, ctx, "section.delete_denied",
				sectionDenialFields(actorID, s.ID, s.CurriculumID(), "not_editable"))
		case errors.Is(err, entities.ErrSectionScopeForbidden):
			emitSectionAudit(uc.audit, ctx, "section.delete_denied",
				sectionDenialFields(actorID, s.ID, s.CurriculumID(), "forbidden"))
		}
		return err
	}

	if err := uc.repo.Delete(ctx, sectionID); err != nil {
		return err
	}

	emitSectionAudit(uc.audit, ctx, "section.deleted", map[string]any{
		"actor_user_id": actorID,
		"section_id":    s.ID,
		"curriculum_id": s.CurriculumID(),
	})
	return nil
}
