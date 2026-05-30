package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/repositories"
)

// ApproveRevisionInput is the public request DTO. Actor + role flow
// through Execute as separate arguments so handlers wire the JWT
// subject explicitly.
type ApproveRevisionInput struct {
	WorkProgramID int64
	RevisionID    int64
}

// approveRevisionRepo is the narrow load-mutate-persist port.
type approveRevisionRepo interface {
	GetByID(ctx context.Context, id int64) (*entities.WorkProgram, error)
	Update(ctx context.Context, wp *entities.WorkProgram) error
}

// ApproveRevisionUseCase moves a pending_approval Revision into the
// approved state, recording the approver identity. Approver role per
// ADR-018 ADR-5: methodist primary, system_admin override.
type ApproveRevisionUseCase struct {
	repo  approveRevisionRepo
	audit AuditSink
}

// NewApproveRevisionUseCase wires the use case. Repo is required.
func NewApproveRevisionUseCase(repo approveRevisionRepo, audit AuditSink) *ApproveRevisionUseCase {
	if repo == nil {
		panic("work_program: NewApproveRevisionUseCase requires non-nil repo")
	}
	return &ApproveRevisionUseCase{repo: repo, audit: audit}
}

// Execute runs the approve-revision flow:
//  1. Role gate (isApprover): methodist OR system_admin → otherwise
//     ErrWorkProgramScopeForbidden + 'forbidden_role' denial.
//  2. Load by id; ErrWorkProgramNotFound → 'not_found' denial.
//  3. wp.ApproveRevision applies the lookup + sub-FSM gate:
//     ErrRevisionNotFound → 'revision_not_found',
//     ErrInvalidStatusTransition → 'not_pending'. actorID is recorded
//     on the revision as approverID (Рособрнадзор trail).
//  4. Persist via repo.Update. Transport errors propagate without audit.
func (uc *ApproveRevisionUseCase) Execute(ctx context.Context, actorID int64, actorRole string, in ApproveRevisionInput) (*entities.WorkProgram, error) {
	if !isApprover(actorRole) {
		emitAudit(uc.audit, ctx, "work_program.revision_approve_denied",
			denialFields(actorID, in.WorkProgramID, "forbidden_role", ""))
		return nil, fmt.Errorf("%w: role %q cannot approve revisions", domain.ErrWorkProgramScopeForbidden, actorRole)
	}

	wp, err := uc.repo.GetByID(ctx, in.WorkProgramID)
	if err != nil {
		if errors.Is(err, repositories.ErrWorkProgramNotFound) {
			emitAudit(uc.audit, ctx, "work_program.revision_approve_denied",
				denialFields(actorID, in.WorkProgramID, "not_found", ""))
		}
		return nil, err
	}

	if err := wp.ApproveRevision(in.RevisionID, actorID); err != nil {
		switch {
		case errors.Is(err, domain.ErrRevisionNotFound):
			emitAudit(uc.audit, ctx, "work_program.revision_approve_denied",
				denialFields(actorID, in.WorkProgramID, "revision_not_found", wp.SpecialtyCode()))
		case errors.Is(err, domain.ErrInvalidStatusTransition):
			emitAudit(uc.audit, ctx, "work_program.revision_approve_denied",
				denialFields(actorID, in.WorkProgramID, "not_pending", wp.SpecialtyCode()))
		}
		return nil, err
	}

	if err := uc.repo.Update(ctx, wp); err != nil {
		return nil, err
	}

	fields := successFields(actorID, wp.ID(), wp.SpecialtyCode(), string(wp.Status()))
	fields["revision_id"] = in.RevisionID
	emitAudit(uc.audit, ctx, "work_program.revision_approved", fields)
	return wp, nil
}
