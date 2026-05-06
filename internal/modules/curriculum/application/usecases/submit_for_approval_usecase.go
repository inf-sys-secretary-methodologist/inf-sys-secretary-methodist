package usecases

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
)

// SubmitForApprovalInput is the public DTO. The actor (and admin
// flag) flow through Execute as positional arguments — handlers
// wire those from JWT subject + role separately from the request
// body.
type SubmitForApprovalInput struct {
	ID int64
}

// submitForApprovalRepo is the narrow port: load by id (so we can
// authorize against the row's createdBy) + write back the
// transitioned entity.
type submitForApprovalRepo interface {
	GetByID(ctx context.Context, id int64) (*entities.Curriculum, error)
	Update(ctx context.Context, c *entities.Curriculum) error
}

// SubmitForApprovalUseCase moves a draft curriculum into the
// pending_approval state. Author or admin may invoke it; the
// entity enforces the status invariant.
type SubmitForApprovalUseCase struct {
	repo  submitForApprovalRepo
	audit AuditSink
	clock func() time.Time
}

// NewSubmitForApprovalUseCase wires the use case. Repo is required.
func NewSubmitForApprovalUseCase(repo submitForApprovalRepo, audit AuditSink, clock func() time.Time) *SubmitForApprovalUseCase {
	if repo == nil {
		panic("curriculum: NewSubmitForApprovalUseCase requires non-nil repo")
	}
	if clock == nil {
		clock = time.Now
	}
	return &SubmitForApprovalUseCase{repo: repo, audit: audit, clock: clock}
}

// Execute runs the submit flow:
//  1. Load by id; ErrCurriculumNotFound → 'not_found' denial.
//  2. Authorize: actor must be author OR admin. Otherwise →
//     ErrCurriculumScopeForbidden + 'forbidden' denial.
//  3. Apply SubmitForApproval; ErrCannotSubmit → 'not_draft' denial.
//  4. Persist via repo.Update. Transport errors propagate without
//     audit (audit log = policy decisions, not infra outages).
//
// The "" code argument in denialFields is intentional — Submit
// does not carry a code mutation, so the field stays empty for
// forensic consistency with other denial events.
func (uc *SubmitForApprovalUseCase) Execute(
	ctx context.Context, actorID int64, isAdmin bool,
	in SubmitForApprovalInput,
) (*entities.Curriculum, error) {
	c, err := uc.repo.GetByID(ctx, in.ID)
	if err != nil {
		if errors.Is(err, repositories.ErrCurriculumNotFound) {
			emitAudit(uc.audit, ctx, "curriculum.submit_denied",
				denialFields(actorID, in.ID, "not_found", ""))
		}
		return nil, err
	}

	if !isAdmin && actorID != c.CreatedBy() {
		emitAudit(uc.audit, ctx, "curriculum.submit_denied",
			denialFields(actorID, in.ID, "forbidden", ""))
		return nil, fmt.Errorf("%w: actor %d is not the author (%d)",
			entities.ErrCurriculumScopeForbidden, actorID, c.CreatedBy())
	}

	if err := c.SubmitForApproval(uc.clock()); err != nil {
		if errors.Is(err, entities.ErrCannotSubmit) {
			emitAudit(uc.audit, ctx, "curriculum.submit_denied",
				denialFields(actorID, in.ID, "not_draft", ""))
		}
		return nil, err
	}

	if err := uc.repo.Update(ctx, c); err != nil {
		return nil, err
	}

	emitAudit(uc.audit, ctx, "curriculum.submitted", map[string]any{
		"actor_user_id": actorID,
		"curriculum_id": c.ID,
		"code":          c.Code(),
		"status":        string(c.Status()),
	})
	return c, nil
}
