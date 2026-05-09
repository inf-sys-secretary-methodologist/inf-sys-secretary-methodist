package usecases

import (
	"context"
	"errors"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
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

// Execute — implementation lands в GREEN commit (Pair 4).
func (uc *UpdateDisciplineItemUseCase) Execute(ctx context.Context, actorID int64, isAdmin bool, in UpdateDisciplineItemInput) (*entities.DisciplineItem, error) {
	_ = isAdmin
	emitDisciplineItemAudit(uc.audit, ctx, "discipline_item.unimplemented",
		disciplineItemDenialFields(actorID, in.ID, 0, 0, "stub"))
	return nil, errors.New("discipline_item: UpdateDisciplineItemUseCase.Execute not implemented yet")
}
