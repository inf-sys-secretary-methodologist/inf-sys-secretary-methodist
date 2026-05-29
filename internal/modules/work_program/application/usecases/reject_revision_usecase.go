package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/repositories"
)

// RejectRevisionInput is the public request DTO. Reason is mandatory —
// the author needs actionable feedback (domain enforces non-empty via
// ErrRejectReasonRequired). Actor + role flow through Execute as
// separate arguments.
type RejectRevisionInput struct {
	WorkProgramID int64
	RevisionID    int64
	Reason        string
}

// rejectRevisionRepo is the narrow load-mutate-persist port.
type rejectRevisionRepo interface {
	GetByID(ctx context.Context, id int64) (*entities.WorkProgram, error)
	Update(ctx context.Context, wp *entities.WorkProgram) error
}

// RejectRevisionUseCase moves a pending_approval Revision into the
// rejected state with a recorded reason. Approver role per ADR-018
// ADR-5: methodist primary, system_admin override.
type RejectRevisionUseCase struct {
	repo  rejectRevisionRepo
	audit AuditSink
}

// NewRejectRevisionUseCase wires the use case. Repo is required.
func NewRejectRevisionUseCase(repo rejectRevisionRepo, audit AuditSink) *RejectRevisionUseCase {
	if repo == nil {
		panic("work_program: NewRejectRevisionUseCase requires non-nil repo")
	}
	return &RejectRevisionUseCase{repo: repo, audit: audit}
}

// Execute runs the reject-revision flow:
//  1. Role gate (isApprover): methodist OR system_admin → otherwise
//     ErrWorkProgramScopeForbidden + 'forbidden_role' denial.
//  2. Load by id; ErrWorkProgramNotFound → 'not_found' denial.
//  3. wp.RejectRevision applies the lookup + sub-FSM + reason gate:
//     ErrRevisionNotFound → 'revision_not_found',
//     ErrInvalidStatusTransition → 'not_pending',
//     ErrRejectReasonRequired → 'empty_reason'.
//  4. Persist via repo.Update. Transport errors propagate without audit.
func (uc *RejectRevisionUseCase) Execute(ctx context.Context, actorID int64, actorRole string, in RejectRevisionInput) (*entities.WorkProgram, error) {
	if !isApprover(actorRole) {
		emitAudit(uc.audit, ctx, "work_program.revision_reject_denied",
			denialFields(actorID, in.WorkProgramID, "forbidden_role", ""))
		return nil, fmt.Errorf("%w: role %q cannot reject revisions", domain.ErrWorkProgramScopeForbidden, actorRole)
	}

	wp, err := uc.repo.GetByID(ctx, in.WorkProgramID)
	if err != nil {
		if errors.Is(err, repositories.ErrWorkProgramNotFound) {
			emitAudit(uc.audit, ctx, "work_program.revision_reject_denied",
				denialFields(actorID, in.WorkProgramID, "not_found", ""))
		}
		return nil, err
	}

	if err := wp.RejectRevision(in.RevisionID, in.Reason); err != nil {
		switch {
		case errors.Is(err, domain.ErrRevisionNotFound):
			emitAudit(uc.audit, ctx, "work_program.revision_reject_denied",
				denialFields(actorID, in.WorkProgramID, "revision_not_found", wp.SpecialtyCode()))
		case errors.Is(err, domain.ErrInvalidStatusTransition):
			emitAudit(uc.audit, ctx, "work_program.revision_reject_denied",
				denialFields(actorID, in.WorkProgramID, "not_pending", wp.SpecialtyCode()))
		case errors.Is(err, domain.ErrRejectReasonRequired):
			emitAudit(uc.audit, ctx, "work_program.revision_reject_denied",
				denialFields(actorID, in.WorkProgramID, "empty_reason", wp.SpecialtyCode()))
		}
		return nil, err
	}

	if err := uc.repo.Update(ctx, wp); err != nil {
		return nil, err
	}

	fields := successFields(actorID, wp.ID(), wp.SpecialtyCode(), string(wp.Status()))
	fields["revision_id"] = in.RevisionID
	fields["reject_reason"] = in.Reason
	emitAudit(uc.audit, ctx, "work_program.revision_rejected", fields)
	return wp, nil
}
