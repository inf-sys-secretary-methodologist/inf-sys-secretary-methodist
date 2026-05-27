package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/repositories"
)

// SubmitWorkProgramInput is the public DTO. The actor (and role) flow
// through Execute as positional arguments — handlers wire those from
// the JWT subject + role separately from the request body.
type SubmitWorkProgramInput struct {
	ID int64
}

// submitWorkProgramRepo is the narrow port: load by id (so we can
// authorize against the row's AuthorID) + write back the transitioned
// aggregate.
type submitWorkProgramRepo interface {
	GetByID(ctx context.Context, id int64) (*entities.WorkProgram, error)
	Update(ctx context.Context, wp *entities.WorkProgram) error
}

// SubmitWorkProgramUseCase moves a draft (or needs_revision) WorkProgram
// into the pending_approval state. Author or system_admin may invoke
// it; the entity enforces the status FSM invariant.
type SubmitWorkProgramUseCase struct {
	repo  submitWorkProgramRepo
	audit AuditSink
}

// NewSubmitWorkProgramUseCase wires the use case. Repo is required.
func NewSubmitWorkProgramUseCase(repo submitWorkProgramRepo, audit AuditSink) *SubmitWorkProgramUseCase {
	if repo == nil {
		panic("work_program: NewSubmitWorkProgramUseCase requires non-nil repo")
	}
	return &SubmitWorkProgramUseCase{repo: repo, audit: audit}
}

// Execute runs the submit flow:
//  1. Load by id; ErrWorkProgramNotFound → 'not_found' denial.
//  2. Authorize: actor must be author OR system_admin. Otherwise →
//     ErrWorkProgramScopeForbidden + 'forbidden' denial.
//  3. Apply wp.Submit(); ErrInvalidStatusTransition → 'not_submittable'
//     denial.
//  4. Persist via repo.Update. Transport errors propagate without
//     audit (audit log = policy decisions, not infra outages).
func (uc *SubmitWorkProgramUseCase) Execute(ctx context.Context, actorID int64, actorRole string, in SubmitWorkProgramInput) (*entities.WorkProgram, error) {
	wp, err := uc.repo.GetByID(ctx, in.ID)
	if err != nil {
		if errors.Is(err, repositories.ErrWorkProgramNotFound) {
			emitAudit(uc.audit, ctx, "work_program.submit_denied",
				denialFields(actorID, in.ID, "not_found", ""))
		}
		return nil, err
	}

	if !isAuthorOrSystemAdmin(actorID, actorRole, wp.AuthorID()) {
		emitAudit(uc.audit, ctx, "work_program.submit_denied",
			denialFields(actorID, in.ID, "forbidden", wp.SpecialtyCode()))
		return nil, fmt.Errorf("%w: actor %d is not the author (%d) and not system_admin",
			domain.ErrWorkProgramScopeForbidden, actorID, wp.AuthorID())
	}

	if err := wp.Submit(); err != nil {
		if errors.Is(err, domain.ErrInvalidStatusTransition) {
			emitAudit(uc.audit, ctx, "work_program.submit_denied",
				denialFields(actorID, in.ID, "not_submittable", wp.SpecialtyCode()))
		}
		return nil, err
	}

	if err := uc.repo.Update(ctx, wp); err != nil {
		return nil, err
	}

	emitAudit(uc.audit, ctx, "work_program.submitted", map[string]any{
		"actor_user_id":   actorID,
		"work_program_id": wp.ID(),
		"specialty_code":  wp.SpecialtyCode(),
		"status":          string(wp.Status()),
	})
	return wp, nil
}
