package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/repositories"
)

// RejectWorkProgramInput is the public DTO. Reason is mandatory —
// the author needs actionable feedback to revise (domain enforces
// non-empty via ErrRejectReasonRequired).
type RejectWorkProgramInput struct {
	ID     int64
	Reason string
}

// rejectWorkProgramRepo is the narrow port: load + write back.
type rejectWorkProgramRepo interface {
	GetByID(ctx context.Context, id int64) (*entities.WorkProgram, error)
	Update(ctx context.Context, wp *entities.WorkProgram) error
}

// RejectWorkProgramUseCase moves a pending_approval WorkProgram back
// to draft with a recorded reason. Approver role per ADR-018 ADR-5:
// methodist primary, system_admin override.
type RejectWorkProgramUseCase struct {
	repo  rejectWorkProgramRepo
	audit AuditSink
}

// NewRejectWorkProgramUseCase wires the use case. Repo is required.
func NewRejectWorkProgramUseCase(repo rejectWorkProgramRepo, audit AuditSink) *RejectWorkProgramUseCase {
	if repo == nil {
		panic("work_program: NewRejectWorkProgramUseCase requires non-nil repo")
	}
	return &RejectWorkProgramUseCase{repo: repo, audit: audit}
}

// Execute runs the reject flow:
//  1. Role gate (isApprover): methodist OR system_admin.
//  2. Load by id; ErrWorkProgramNotFound → 'not_found' denial.
//  3. Apply wp.Reject(reason):
//     - ErrInvalidStatusTransition (wrong status) → 'not_pending' denial.
//     - ErrRejectReasonRequired (empty/whitespace reason) → 'empty_reason'
//     denial; the domain enforces non-empty after TrimSpace so the
//     author always gets actionable feedback.
//  4. Persist via repo.Update. Transport errors propagate without audit.
//
// Per ADR-018 ADR-3 the reason is captured in both the entity field
// (which lives until the next Approve clears it) AND the audit row
// (the durable forensic record).
func (uc *RejectWorkProgramUseCase) Execute(ctx context.Context, actorID int64, actorRole string, in RejectWorkProgramInput) (*entities.WorkProgram, error) {
	if !isApprover(actorRole) {
		emitAudit(uc.audit, ctx, "work_program.reject_denied",
			denialFields(actorID, in.ID, "forbidden_role", ""))
		return nil, fmt.Errorf("%w: role %q cannot reject", domain.ErrWorkProgramScopeForbidden, actorRole)
	}

	wp, err := uc.repo.GetByID(ctx, in.ID)
	if err != nil {
		if errors.Is(err, repositories.ErrWorkProgramNotFound) {
			emitAudit(uc.audit, ctx, "work_program.reject_denied",
				denialFields(actorID, in.ID, "not_found", ""))
		}
		return nil, err
	}

	if err := wp.Reject(in.Reason); err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidStatusTransition):
			emitAudit(uc.audit, ctx, "work_program.reject_denied",
				denialFields(actorID, in.ID, "not_pending", wp.SpecialtyCode()))
		case errors.Is(err, domain.ErrRejectReasonRequired):
			emitAudit(uc.audit, ctx, "work_program.reject_denied",
				denialFields(actorID, in.ID, "empty_reason", wp.SpecialtyCode()))
		}
		return nil, err
	}

	if err := uc.repo.Update(ctx, wp); err != nil {
		return nil, err
	}

	emitAudit(uc.audit, ctx, "work_program.rejected", map[string]any{
		"actor_user_id":   actorID,
		"work_program_id": wp.ID(),
		"specialty_code":  wp.SpecialtyCode(),
		"status":          string(wp.Status()),
		"reject_reason":   wp.RejectReason(),
	})
	return wp, nil
}
