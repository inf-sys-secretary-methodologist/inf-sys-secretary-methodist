package usecases

import (
	"context"
	"errors"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
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

// Execute runs the use case end-to-end:
//
//  1. Load parent curriculum; ErrCurriculumNotFound → 'curriculum_not_found'
//     denial (operator reads in.CurriculumID from request log).
//  2. AuthorizeSectionEdit(actor, admin, curStatus, curCreatedBy) →
//     'forbidden' or 'not_editable' denial depending on which sentinel
//     surfaces.
//  3. Build entity through NewSection (invariant gate); ErrInvalidSection
//     → 'invalid' denial.
//  4. Persist via repo.Save.
//
// Transport errors propagate WITHOUT producing any audit event so
// the log doesn't conflate infrastructure outages with policy
// decisions (operators read transport failures from logger stack
// traces).
func (uc *CreateSectionUseCase) Execute(ctx context.Context, actorID int64, isAdmin bool, in CreateSectionInput) (*entities.Section, error) {
	cur, err := uc.curriculumRepo.GetByID(ctx, in.CurriculumID)
	if err != nil {
		if errors.Is(err, repositories.ErrCurriculumNotFound) {
			emitSectionAudit(uc.audit, ctx, "section.create_denied",
				sectionDenialFields(actorID, 0, in.CurriculumID, "curriculum_not_found"))
		}
		return nil, err
	}

	if err := entities.AuthorizeSectionEdit(actorID, isAdmin, cur.Status(), cur.CreatedBy()); err != nil {
		switch {
		case errors.Is(err, entities.ErrCannotEditSection):
			emitSectionAudit(uc.audit, ctx, "section.create_denied",
				sectionDenialFields(actorID, 0, in.CurriculumID, "not_editable"))
		case errors.Is(err, entities.ErrSectionScopeForbidden):
			emitSectionAudit(uc.audit, ctx, "section.create_denied",
				sectionDenialFields(actorID, 0, in.CurriculumID, "forbidden"))
		}
		return nil, err
	}

	s, err := entities.NewSection(entities.NewSectionParams{
		CurriculumID: in.CurriculumID,
		Title:        in.Title,
		Description:  in.Description,
		OrderIndex:   in.OrderIndex,
		Now:          uc.clock(),
	})
	if err != nil {
		if errors.Is(err, entities.ErrInvalidSection) {
			emitSectionAudit(uc.audit, ctx, "section.create_denied",
				sectionDenialFields(actorID, 0, in.CurriculumID, "invalid"))
		}
		return nil, err
	}

	if err := uc.repo.Save(ctx, s); err != nil {
		return nil, err
	}

	emitSectionAudit(uc.audit, ctx, "section.created", map[string]any{
		"actor_user_id": actorID,
		"section_id":    s.ID,
		"curriculum_id": s.CurriculumID(),
		"title":         s.Title(),
		"order_index":   s.OrderIndex(),
	})
	return s, nil
}
