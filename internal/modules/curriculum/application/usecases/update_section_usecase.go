package usecases

import (
	"context"
	"errors"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
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

// updateSectionRepo is the narrow port for persistence.
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

// Execute performs the edit:
//
//  1. Load section by id; ErrSectionNotFound → 'not_found' denial.
//  2. Load parent curriculum (via section.CurriculumID());
//     ErrCurriculumNotFound propagates without audit (defense-in-depth
//     against orphaned section row, no productive policy event).
//  3. AuthorizeEdit; sentinels mapped to 'forbidden' / 'not_editable'.
//  4. UpdateBasics; ErrInvalidSection → 'invalid' denial.
//  5. Persist via repo.Update; ErrSectionVersionConflict →
//     'version_conflict' denial (audited as policy event since clients
//     act on the conflict — reload + retry).
//
// Transport errors propagate WITHOUT audit events (logger captures
// stack traces — operational, not policy).
func (uc *UpdateSectionUseCase) Execute(ctx context.Context, actorID int64, isAdmin bool, in UpdateSectionInput) (*entities.Section, error) {
	s, err := uc.repo.GetByID(ctx, in.ID)
	if err != nil {
		if errors.Is(err, repositories.ErrSectionNotFound) {
			emitSectionAudit(uc.audit, ctx, "section.update_denied",
				sectionDenialFields(actorID, in.ID, 0, "not_found"))
		}
		return nil, err
	}

	cur, err := uc.curriculumRepo.GetByID(ctx, s.CurriculumID())
	if err != nil {
		// Orphaned section: parent curriculum gone. FK CASCADE in
		// migration 034 normally prevents this; this path is
		// defense-in-depth only. No audit — operational anomaly,
		// not a policy event.
		return nil, err
	}

	if err := s.AuthorizeEdit(actorID, isAdmin, cur.Status(), cur.CreatedBy()); err != nil {
		switch {
		case errors.Is(err, entities.ErrCannotEditSection):
			emitSectionAudit(uc.audit, ctx, "section.update_denied",
				sectionDenialFields(actorID, s.ID, s.CurriculumID(), "not_editable"))
		case errors.Is(err, entities.ErrSectionScopeForbidden):
			emitSectionAudit(uc.audit, ctx, "section.update_denied",
				sectionDenialFields(actorID, s.ID, s.CurriculumID(), "forbidden"))
		}
		return nil, err
	}

	if err := s.UpdateBasics(in.Title, in.Description, in.OrderIndex, uc.clock()); err != nil {
		if errors.Is(err, entities.ErrInvalidSection) {
			emitSectionAudit(uc.audit, ctx, "section.update_denied",
				sectionDenialFields(actorID, s.ID, s.CurriculumID(), "invalid"))
		}
		return nil, err
	}

	if err := uc.repo.Update(ctx, s); err != nil {
		if errors.Is(err, repositories.ErrSectionVersionConflict) {
			emitSectionAudit(uc.audit, ctx, "section.update_denied",
				sectionDenialFields(actorID, s.ID, s.CurriculumID(), "version_conflict"))
		}
		return nil, err
	}

	emitSectionAudit(uc.audit, ctx, "section.updated", map[string]any{
		"actor_user_id": actorID,
		"section_id":    s.ID,
		"curriculum_id": s.CurriculumID(),
		"title":         s.Title(),
		"order_index":   s.OrderIndex(),
	})
	return s, nil
}
