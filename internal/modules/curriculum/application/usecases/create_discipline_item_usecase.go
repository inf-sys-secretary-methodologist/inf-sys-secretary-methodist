package usecases

import (
	"context"
	"errors"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
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

// Execute — implementation lands в GREEN commit (Pair 4).
func (uc *CreateDisciplineItemUseCase) Execute(ctx context.Context, actorID int64, isAdmin bool, in CreateDisciplineItemInput) (*entities.DisciplineItem, error) {
	_ = isAdmin
	emitDisciplineItemAudit(uc.audit, ctx, "discipline_item.unimplemented",
		disciplineItemDenialFields(actorID, 0, in.SectionID, 0, "stub"))
	return nil, errors.New("discipline_item: CreateDisciplineItemUseCase.Execute not implemented yet")
}
