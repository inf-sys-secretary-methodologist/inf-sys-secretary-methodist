package usecases

import (
	"context"
	"errors"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
)

// CreateDisciplineItemInput is the public DTO for CreateDisciplineItem.
type CreateDisciplineItemInput struct {
	SectionID     int64
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

// createDisciplineItemRepo is the narrow port for persistence — Save only.
type createDisciplineItemRepo interface {
	Save(ctx context.Context, d *entities.DisciplineItem) error
}

// createDisciplineItemSectionLookup is the cross-aggregate port: load
// the parent section to read its curriculum_id (one hop в two-level
// lookup chain).
type createDisciplineItemSectionLookup interface {
	GetByID(ctx context.Context, id int64) (*entities.Section, error)
}

// createDisciplineItemCurriculumLookup is the cross-aggregate port:
// load the curriculum (через section.curriculum_id) для status +
// created_by primitives. Two-level cross-aggregate per ADR-1 Beta.
type createDisciplineItemCurriculumLookup interface {
	GetByID(ctx context.Context, id int64) (*entities.Curriculum, error)
}

// CreateDisciplineItemUseCase persists a fresh discipline item after
// two-level authorization (item → section → curriculum status +
// ownership).
type CreateDisciplineItemUseCase struct {
	repo           createDisciplineItemRepo
	sectionRepo    createDisciplineItemSectionLookup
	curriculumRepo createDisciplineItemCurriculumLookup
	audit          AuditSink
	clock          func() time.Time
}

// NewCreateDisciplineItemUseCase wires the use case. All 3 repo
// arguments required (non-nil): nil dependencies would let requests
// reach a panic deeper в the call stack instead of failing at DI.
func NewCreateDisciplineItemUseCase(
	repo createDisciplineItemRepo,
	sectionRepo createDisciplineItemSectionLookup,
	curriculumRepo createDisciplineItemCurriculumLookup,
	audit AuditSink,
	clock func() time.Time,
) *CreateDisciplineItemUseCase {
	if repo == nil {
		panic("discipline_item: NewCreateDisciplineItemUseCase requires non-nil repo")
	}
	if sectionRepo == nil {
		panic("discipline_item: NewCreateDisciplineItemUseCase requires non-nil sectionRepo")
	}
	if curriculumRepo == nil {
		panic("discipline_item: NewCreateDisciplineItemUseCase requires non-nil curriculumRepo")
	}
	if clock == nil {
		clock = time.Now
	}
	return &CreateDisciplineItemUseCase{repo: repo, sectionRepo: sectionRepo, curriculumRepo: curriculumRepo, audit: audit, clock: clock}
}

// Execute runs the use case end-to-end:
//
//  1. Load parent section by SectionID; ErrSectionNotFound → 'section_not_found'
//     denial.
//  2. Load curriculum (через section.CurriculumID()); errors propagate
//     без audit (orphaned section is operational anomaly).
//  3. AuthorizeDisciplineItemEdit free function (no instance yet —
//     ADR-1 Beta primitives) → 'forbidden' / 'not_editable' denial.
//  4. NewDisciplineItem invariant gate → 'invalid' denial.
//  5. repo.Save persists; success emits 'discipline_item.created'.
//
// Transport errors propagate WITHOUT audit (logger captures stack
// traces — operational, не policy).
func (uc *CreateDisciplineItemUseCase) Execute(ctx context.Context, actorID int64, isAdmin bool, in CreateDisciplineItemInput) (*entities.DisciplineItem, error) {
	section, err := uc.sectionRepo.GetByID(ctx, in.SectionID)
	if err != nil {
		if errors.Is(err, repositories.ErrSectionNotFound) {
			emitDisciplineItemAudit(uc.audit, ctx, "discipline_item.create_denied",
				disciplineItemDenialFields(actorID, 0, in.SectionID, 0, "section_not_found"))
		}
		return nil, err
	}

	cur, err := uc.curriculumRepo.GetByID(ctx, section.CurriculumID())
	if err != nil {
		// Orphaned section path; FK CASCADE normally prevents но defense-in-depth.
		return nil, err
	}

	if err := entities.AuthorizeDisciplineItemEdit(actorID, isAdmin, cur.Status(), cur.CreatedBy()); err != nil {
		switch {
		case errors.Is(err, entities.ErrCannotEditDisciplineItem):
			emitDisciplineItemAudit(uc.audit, ctx, "discipline_item.create_denied",
				disciplineItemDenialFields(actorID, 0, in.SectionID, section.CurriculumID(), "not_editable"))
		case errors.Is(err, entities.ErrDisciplineItemScopeForbidden):
			emitDisciplineItemAudit(uc.audit, ctx, "discipline_item.create_denied",
				disciplineItemDenialFields(actorID, 0, in.SectionID, section.CurriculumID(), "forbidden"))
		}
		return nil, err
	}

	d, err := entities.NewDisciplineItem(entities.NewDisciplineItemParams{
		SectionID:     in.SectionID,
		Title:         in.Title,
		HoursLectures: in.HoursLectures,
		HoursPractice: in.HoursPractice,
		HoursLab:      in.HoursLab,
		HoursSelf:     in.HoursSelf,
		ControlForm:   in.ControlForm,
		Credits:       in.Credits,
		Semester:      in.Semester,
		OrderIndex:    in.OrderIndex,
		Now:           uc.clock(),
	})
	if err != nil {
		if errors.Is(err, entities.ErrInvalidDisciplineItem) {
			emitDisciplineItemAudit(uc.audit, ctx, "discipline_item.create_denied",
				disciplineItemDenialFields(actorID, 0, in.SectionID, section.CurriculumID(), "invalid"))
		}
		return nil, err
	}

	if err := uc.repo.Save(ctx, d); err != nil {
		return nil, err
	}

	emitDisciplineItemAudit(uc.audit, ctx, "discipline_item.created", map[string]any{
		"actor_user_id": actorID,
		"item_id":       d.ID,
		"section_id":    d.SectionID(),
		"curriculum_id": section.CurriculumID(),
		"title":         d.Title(),
		"control_form":  string(d.ControlForm()),
	})
	return d, nil
}
