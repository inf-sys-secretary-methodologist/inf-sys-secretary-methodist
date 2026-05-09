package usecases

import (
	"context"
	"errors"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
)

// UpdateSectionInput is the public DTO for the section edit use case.
// ID identifies the target row; mutable content fields follow.
//
// Version is intentionally NOT part of the input — the use case loads
// the section first (which carries the freshly-fetched version from
// the row) and the repository's optimistic-lock SQL guards the write.
// Clients that want explicit conflict detection should use a future
// "If-Match" header convention or the bulk-edit endpoint (B1b).
type UpdateSectionInput struct {
	ID          int64
	Title       string
	Description string
	OrderIndex  int
}

// updateSectionRepo is the narrow port for persistence: section
// fetch (so we can authorize against the row's curriculum) and
// section write (the optimistic-lock UPDATE).
type updateSectionRepo interface {
	GetByID(ctx context.Context, id int64) (*entities.Section, error)
	Update(ctx context.Context, s *entities.Section) error
}

// updateSectionCurriculumLookup is the cross-aggregate read port —
// we need the parent curriculum's status + author to decide
// authorization (ADR-1 Beta primitives).
type updateSectionCurriculumLookup interface {
	GetByID(ctx context.Context, id int64) (*entities.Curriculum, error)
}

// UpdateSectionUseCase loads a section, loads its parent curriculum,
// runs AuthorizeEdit, applies UpdateBasics, and persists with
// optimistic locking.
type UpdateSectionUseCase struct {
	repo           updateSectionRepo
	curriculumRepo updateSectionCurriculumLookup
	audit          AuditSink
	clock          func() time.Time
}

// NewUpdateSectionUseCase wires the use case.
func NewUpdateSectionUseCase(
	repo updateSectionRepo,
	curriculumRepo updateSectionCurriculumLookup,
	audit AuditSink,
	clock func() time.Time,
) *UpdateSectionUseCase {
	if repo == nil {
		panic("section: NewUpdateSectionUseCase requires non-nil repo")
	}
	if curriculumRepo == nil {
		panic("section: NewUpdateSectionUseCase requires non-nil curriculumRepo")
	}
	if clock == nil {
		clock = time.Now
	}
	return &UpdateSectionUseCase{repo: repo, curriculumRepo: curriculumRepo, audit: audit, clock: clock}
}

// Execute — implementation lands в GREEN commit (Pair 4).
func (uc *UpdateSectionUseCase) Execute(ctx context.Context, actorID int64, isAdmin bool, in UpdateSectionInput) (*entities.Section, error) {
	_ = isAdmin
	emitSectionAudit(uc.audit, ctx, "section.unimplemented",
		sectionDenialFields(actorID, in.ID, 0, "stub"))
	return nil, errors.New("section: UpdateSectionUseCase.Execute not implemented yet")
}
