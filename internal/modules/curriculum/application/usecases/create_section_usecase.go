package usecases

import (
	"context"
	"errors"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
)

// CreateSectionInput is the public DTO for CreateSection. CurriculumID
// identifies the parent aggregate; the actor (and admin flag) flow
// through Execute as positional arguments so request body never
// carries authentication context.
type CreateSectionInput struct {
	CurriculumID int64
	Title        string
	Description  string
	OrderIndex   int
}

// createSectionRepo is the narrow port for persistence. CreateSection
// only needs Save — keeping the port single-method protects use-case
// tests from broader concrete-repository surface.
type createSectionRepo interface {
	Save(ctx context.Context, s *entities.Section) error
}

// createSectionCurriculumLookup is the cross-aggregate port: load the
// parent curriculum to read its status + author so AuthorizeSectionEdit
// can decide. Per ADR-1 Beta, Section knows nothing about Curriculum
// directly — the use case orchestrates the read.
type createSectionCurriculumLookup interface {
	GetByID(ctx context.Context, id int64) (*entities.Curriculum, error)
}

// CreateSectionUseCase persists a fresh section after authorization
// against the parent curriculum's lifecycle + ownership.
type CreateSectionUseCase struct {
	repo           createSectionRepo
	curriculumRepo createSectionCurriculumLookup
	audit          AuditSink
	clock          func() time.Time
}

// NewCreateSectionUseCase wires the use case. Both repo arguments are
// required (non-nil): nil dependencies would let requests reach a
// panic deeper in the call stack instead of failing during DI wiring.
func NewCreateSectionUseCase(
	repo createSectionRepo,
	curriculumRepo createSectionCurriculumLookup,
	audit AuditSink,
	clock func() time.Time,
) *CreateSectionUseCase {
	if repo == nil {
		panic("section: NewCreateSectionUseCase requires non-nil repo")
	}
	if curriculumRepo == nil {
		panic("section: NewCreateSectionUseCase requires non-nil curriculumRepo")
	}
	if clock == nil {
		clock = time.Now
	}
	return &CreateSectionUseCase{repo: repo, curriculumRepo: curriculumRepo, audit: audit, clock: clock}
}

// Execute — implementation lands в GREEN commit (Pair 4). The stub
// emits a no-op audit denial so the audit helpers stay live in the
// dependency graph — keeps golangci-lint's unused-symbol check happy
// across the RED commit boundary.
func (uc *CreateSectionUseCase) Execute(ctx context.Context, actorID int64, isAdmin bool, in CreateSectionInput) (*entities.Section, error) {
	_ = isAdmin
	emitSectionAudit(uc.audit, ctx, "section.unimplemented",
		sectionDenialFields(actorID, 0, in.CurriculumID, "stub"))
	return nil, errors.New("section: CreateSectionUseCase.Execute not implemented yet")
}
