package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/repositories"
)

// ApproveWorkProgramInput is the public DTO. Approver id flows through
// Execute as a positional argument (separate from the request body)
// so handlers wire the JWT subject directly.
type ApproveWorkProgramInput struct {
	ID int64
}

// approveWorkProgramRepo is the narrow port: load + write back.
type approveWorkProgramRepo interface {
	GetByID(ctx context.Context, id int64) (*entities.WorkProgram, error)
	Update(ctx context.Context, wp *entities.WorkProgram) error
}

// ApproveWorkProgramUseCase moves a pending_approval WorkProgram into
// the approved state, recording the approver identity + timestamp.
// Approver role per ADR-018 ADR-5: methodist primary, system_admin
// override. The use case is the access gate (route-level middleware
// pins the same set as defense in depth in PR 4).
type ApproveWorkProgramUseCase struct {
	repo  approveWorkProgramRepo
	audit AuditSink
}

// NewApproveWorkProgramUseCase wires the use case. Repo is required.
func NewApproveWorkProgramUseCase(repo approveWorkProgramRepo, audit AuditSink) *ApproveWorkProgramUseCase {
	if repo == nil {
		panic("work_program: NewApproveWorkProgramUseCase requires non-nil repo")
	}
	return &ApproveWorkProgramUseCase{repo: repo, audit: audit}
}

// Execute runs the approve flow:
//  1. Role gate: methodist OR system_admin per ADR-018 ADR-5.
//  2. Load by id; ErrWorkProgramNotFound → 'not_found' denial.
//  3. Apply wp.Approve(actorID); ErrInvalidStatusTransition →
//     'not_pending' denial.
//  4. Persist via repo.Update. Transport errors propagate without
//     audit (audit log = policy decisions, not infra outages).
//
// ApproverID is the actor id (JWT-derived), recorded on the entity
// for Рособрнадзор audit trail ("who approved this РПД").
func (uc *ApproveWorkProgramUseCase) Execute(ctx context.Context, actorID int64, actorRole string, in ApproveWorkProgramInput) (*entities.WorkProgram, error) {
	if !isApprover(actorRole) {
		emitAudit(uc.audit, ctx, "work_program.approve_denied",
			denialFields(actorID, in.ID, "forbidden_role", ""))
		return nil, fmt.Errorf("%w: role %q cannot approve", domain.ErrWorkProgramScopeForbidden, actorRole)
	}

	wp, err := uc.repo.GetByID(ctx, in.ID)
	if err != nil {
		if errors.Is(err, repositories.ErrWorkProgramNotFound) {
			emitAudit(uc.audit, ctx, "work_program.approve_denied",
				denialFields(actorID, in.ID, "not_found", ""))
		}
		return nil, err
	}

	if err := wp.Approve(actorID); err != nil {
		if errors.Is(err, domain.ErrInvalidStatusTransition) {
			emitAudit(uc.audit, ctx, "work_program.approve_denied",
				denialFields(actorID, in.ID, "not_pending", wp.SpecialtyCode()))
		}
		return nil, err
	}

	if err := uc.repo.Update(ctx, wp); err != nil {
		return nil, err
	}

	emitAudit(uc.audit, ctx, "work_program.approved",
		successFields(actorID, wp.ID(), wp.SpecialtyCode(), string(wp.Status())))
	return wp, nil
}
