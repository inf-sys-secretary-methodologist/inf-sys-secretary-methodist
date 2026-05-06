package usecases

import (
	"context"
	"errors"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
)

// UpdateCurriculumInput is the public DTO for the edit use case.
// Carries the target id plus the five mutable content fields. The
// actor (and admin flag) flow through Execute as positional
// arguments — handlers wire those from JWT subject + role,
// keeping authentication context separate from the request body.
type UpdateCurriculumInput struct {
	ID          int64
	Title       string
	Code        string
	Specialty   string
	Year        int
	Description string
}

// updateCurriculumRepo is the narrow port the Update use case
// requires from persistence: fetch (so we can authorize the actor
// against the row's ownership) and write (so the mutation lands).
type updateCurriculumRepo interface {
	GetByID(ctx context.Context, id int64) (*entities.Curriculum, error)
	Update(ctx context.Context, c *entities.Curriculum) error
}

// UpdateCurriculumUseCase loads a curriculum, runs AuthorizeEdit
// against the actor + admin flag, applies UpdateBasics, and
// persists the result. Every outcome is reflected in the audit
// log so a forensic trail captures both successful edits and
// every flavour of denial.
type UpdateCurriculumUseCase struct {
	repo  updateCurriculumRepo
	audit AuditSink
	clock func() time.Time
}

// NewUpdateCurriculumUseCase wires the use case. Repo is required.
func NewUpdateCurriculumUseCase(repo updateCurriculumRepo, audit AuditSink, clock func() time.Time) *UpdateCurriculumUseCase {
	if repo == nil {
		panic("curriculum: NewUpdateCurriculumUseCase requires non-nil repo")
	}
	if clock == nil {
		clock = time.Now
	}
	return &UpdateCurriculumUseCase{repo: repo, audit: audit, clock: clock}
}

// Execute performs the edit:
//
//  1. Load row by id; ErrCurriculumNotFound → 'not_found' denial.
//  2. AuthorizeEdit(actorID, isAdmin) → 'forbidden' or 'not_editable'
//     denial depending on which sentinel surfaces.
//  3. Apply UpdateBasics; ErrInvalidCurriculum → 'invalid' denial.
//  4. Persist via repo.Update; ErrCurriculumCodeExists →
//     'code_conflict' denial.
//
// Transport errors propagate WITHOUT producing any audit event so
// the log doesn't conflate infrastructure outages with policy
// decisions (operators read transport failures from logger stack
// traces).
func (uc *UpdateCurriculumUseCase) Execute(
	ctx context.Context, actorID int64, isAdmin bool,
	in UpdateCurriculumInput,
) (*entities.Curriculum, error) {
	c, err := uc.repo.GetByID(ctx, in.ID)
	if err != nil {
		if errors.Is(err, repositories.ErrCurriculumNotFound) {
			uc.emitDenial(ctx, actorID, in.ID, "not_found", in.Code)
		}
		return nil, err
	}

	if err := c.AuthorizeEdit(actorID, isAdmin); err != nil {
		switch {
		case errors.Is(err, entities.ErrCannotEditApproved):
			uc.emitDenial(ctx, actorID, in.ID, "not_editable", in.Code)
		case errors.Is(err, entities.ErrCurriculumScopeForbidden):
			uc.emitDenial(ctx, actorID, in.ID, "forbidden", in.Code)
		}
		return nil, err
	}

	if err := c.UpdateBasics(in.Title, in.Code, in.Specialty, in.Year, in.Description, uc.clock()); err != nil {
		switch {
		case errors.Is(err, entities.ErrInvalidCurriculum):
			uc.emitDenial(ctx, actorID, in.ID, "invalid", in.Code)
		case errors.Is(err, entities.ErrCannotEditApproved):
			// Defense in depth: AuthorizeEdit already guarded this above,
			// but UpdateBasics enforces it again. If a future maintainer
			// removes the AuthorizeEdit call this denial path still trips.
			uc.emitDenial(ctx, actorID, in.ID, "not_editable", in.Code)
		}
		return nil, err
	}

	if err := uc.repo.Update(ctx, c); err != nil {
		if errors.Is(err, repositories.ErrCurriculumCodeExists) {
			uc.emitDenial(ctx, actorID, in.ID, "code_conflict", in.Code)
		}
		return nil, err
	}

	if uc.audit != nil {
		uc.audit.LogAuditEvent(ctx, "curriculum.updated", auditResource, map[string]any{
			"actor_user_id": actorID,
			"curriculum_id": c.ID,
			"code":          c.Code(),
			"year":          c.Year(),
			"specialty":     c.Specialty(),
			"status":        string(c.Status()),
		})
	}
	return c, nil
}

func (uc *UpdateCurriculumUseCase) emitDenial(ctx context.Context, actorID, curriculumID int64, reason, code string) {
	if uc.audit == nil {
		return
	}
	uc.audit.LogAuditEvent(ctx, "curriculum.update_denied", auditResource, map[string]any{
		"actor_user_id": actorID,
		"curriculum_id": curriculumID,
		"reason":        reason,
		"code":          code,
	})
}
