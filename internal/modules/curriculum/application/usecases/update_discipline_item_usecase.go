package usecases

import (
	"context"
	"errors"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
)

// UpdateDisciplineItemInput is the public DTO для item edit use case.
// Version intentionally NOT in input — repo loads item which carries
// freshly-fetched version, optimistic-lock SQL guards.
type UpdateDisciplineItemInput struct {
	ID            int64
	Title         string
	HoursLectures int
	HoursPractice int
	HoursLab      int
	HoursSelf     int
	ControlForm   entities.ControlForm
	Credits       int
	Semester      int
	OrderIndex    int
}

// updateDisciplineItemRepo is the narrow port для persistence
// (item GetByID + Update).
type updateDisciplineItemRepo interface {
	GetByID(ctx context.Context, id int64) (*entities.DisciplineItem, error)
	Update(ctx context.Context, d *entities.DisciplineItem) error
}

// updateDisciplineItemSectionLookup is the cross-aggregate read port
// (one hop chain: item.section_id → section).
type updateDisciplineItemSectionLookup interface {
	GetByID(ctx context.Context, id int64) (*entities.Section, error)
}

// updateDisciplineItemCurriculumLookup is the cross-aggregate read
// port (second hop: section.curriculum_id → curriculum).
type updateDisciplineItemCurriculumLookup interface {
	GetByID(ctx context.Context, id int64) (*entities.Curriculum, error)
}

// UpdateDisciplineItemUseCase loads an item, walks two-level
// cross-aggregate chain (item → section → curriculum), runs Authorize,
// applies UpdateBasics, persists с optimistic locking.
type UpdateDisciplineItemUseCase struct {
	repo           updateDisciplineItemRepo
	sectionRepo    updateDisciplineItemSectionLookup
	curriculumRepo updateDisciplineItemCurriculumLookup
	audit          AuditSink
	clock          func() time.Time
}

// NewUpdateDisciplineItemUseCase wires the use case.
func NewUpdateDisciplineItemUseCase(
	repo updateDisciplineItemRepo,
	sectionRepo updateDisciplineItemSectionLookup,
	curriculumRepo updateDisciplineItemCurriculumLookup,
	audit AuditSink,
	clock func() time.Time,
) *UpdateDisciplineItemUseCase {
	if repo == nil {
		panic("discipline_item: NewUpdateDisciplineItemUseCase requires non-nil repo")
	}
	if sectionRepo == nil {
		panic("discipline_item: NewUpdateDisciplineItemUseCase requires non-nil sectionRepo")
	}
	if curriculumRepo == nil {
		panic("discipline_item: NewUpdateDisciplineItemUseCase requires non-nil curriculumRepo")
	}
	if clock == nil {
		clock = time.Now
	}
	return &UpdateDisciplineItemUseCase{repo: repo, sectionRepo: sectionRepo, curriculumRepo: curriculumRepo, audit: audit, clock: clock}
}

// Execute performs the edit:
//
//  1. Load item by id; ErrDisciplineItemNotFound → 'not_found' denial.
//  2. Load section (через item.SectionID()); ErrSectionNotFound propagates
//     без audit (orphaned-row defense).
//  3. Load curriculum (через section.CurriculumID()); errors propagate без audit.
//  4. AuthorizeEdit (cross-aggregate primitives) → 'forbidden' / 'not_editable'.
//  5. UpdateBasics → 'invalid' denial если invariant fail.
//  6. repo.Update; ErrDisciplineItemVersionConflict → 'version_conflict' denial
//     (audited as policy event — clients act on conflict via reload+retry).
func (uc *UpdateDisciplineItemUseCase) Execute(ctx context.Context, actorID int64, isAdmin bool, in UpdateDisciplineItemInput) (*entities.DisciplineItem, error) {
	d, err := uc.repo.GetByID(ctx, in.ID)
	if err != nil {
		if errors.Is(err, repositories.ErrDisciplineItemNotFound) {
			emitDisciplineItemAudit(uc.audit, ctx, "discipline_item.update_denied",
				disciplineItemDenialFields(actorID, in.ID, 0, 0, "not_found"))
		}
		return nil, err
	}

	section, err := uc.sectionRepo.GetByID(ctx, d.SectionID())
	if err != nil {
		return nil, err
	}

	cur, err := uc.curriculumRepo.GetByID(ctx, section.CurriculumID())
	if err != nil {
		return nil, err
	}

	if err := d.AuthorizeEdit(actorID, isAdmin, cur.Status(), cur.CreatedBy()); err != nil {
		switch {
		case errors.Is(err, entities.ErrCannotEditDisciplineItem):
			emitDisciplineItemAudit(uc.audit, ctx, "discipline_item.update_denied",
				disciplineItemDenialFields(actorID, d.ID, d.SectionID(), section.CurriculumID(), "not_editable"))
		case errors.Is(err, entities.ErrDisciplineItemScopeForbidden):
			emitDisciplineItemAudit(uc.audit, ctx, "discipline_item.update_denied",
				disciplineItemDenialFields(actorID, d.ID, d.SectionID(), section.CurriculumID(), "forbidden"))
		}
		return nil, err
	}

	if err := d.UpdateBasics(in.Title,
		in.HoursLectures, in.HoursPractice, in.HoursLab, in.HoursSelf,
		in.ControlForm, in.Credits, in.Semester, in.OrderIndex,
		uc.clock()); err != nil {
		if errors.Is(err, entities.ErrInvalidDisciplineItem) {
			emitDisciplineItemAudit(uc.audit, ctx, "discipline_item.update_denied",
				disciplineItemDenialFields(actorID, d.ID, d.SectionID(), section.CurriculumID(), "invalid"))
		}
		return nil, err
	}

	if err := uc.repo.Update(ctx, d); err != nil {
		if errors.Is(err, repositories.ErrDisciplineItemVersionConflict) {
			emitDisciplineItemAudit(uc.audit, ctx, "discipline_item.update_denied",
				disciplineItemDenialFields(actorID, d.ID, d.SectionID(), section.CurriculumID(), "version_conflict"))
		}
		return nil, err
	}

	emitDisciplineItemAudit(uc.audit, ctx, "discipline_item.updated", map[string]any{
		"actor_user_id": actorID,
		"item_id":       d.ID,
		"section_id":    d.SectionID(),
		"curriculum_id": section.CurriculumID(),
		"title":         d.Title(),
	})
	return d, nil
}
