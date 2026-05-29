package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/repositories"
)

// SubmitRevisionInput is the public request DTO. Actor + role flow
// through Execute as separate arguments so handlers wire the JWT
// subject explicitly.
type SubmitRevisionInput struct {
	WorkProgramID int64
	RevisionID    int64
}

// submitRevisionRepo is the narrow load-mutate-persist port.
type submitRevisionRepo interface {
	GetByID(ctx context.Context, id int64) (*entities.WorkProgram, error)
	Update(ctx context.Context, wp *entities.WorkProgram) error
}

// SubmitRevisionUseCase moves a draft Revision into pending_approval.
// Author-scoped (author or system_admin) — the same authorship set as
// proposing the revision; a methodist approves it afterwards.
type SubmitRevisionUseCase struct {
	repo  submitRevisionRepo
	audit AuditSink
}

// NewSubmitRevisionUseCase wires the use case. Repo is required.
func NewSubmitRevisionUseCase(repo submitRevisionRepo, audit AuditSink) *SubmitRevisionUseCase {
	if repo == nil {
		panic("work_program: NewSubmitRevisionUseCase requires non-nil repo")
	}
	return &SubmitRevisionUseCase{repo: repo, audit: audit}
}

// Execute runs the submit-revision flow:
//  1. Load by id; ErrWorkProgramNotFound → 'not_found' denial.
//  2. Authorize: actor must be author OR system_admin → otherwise
//     ErrWorkProgramScopeForbidden + 'forbidden' denial.
//  3. wp.SubmitRevision applies the lookup + sub-FSM gate:
//     ErrRevisionNotFound → 'revision_not_found',
//     ErrInvalidStatusTransition → 'not_submittable'.
//  4. Persist via repo.Update. Transport errors propagate without audit.
func (uc *SubmitRevisionUseCase) Execute(ctx context.Context, actorID int64, actorRole string, in SubmitRevisionInput) (*entities.WorkProgram, error) {
	wp, err := uc.repo.GetByID(ctx, in.WorkProgramID)
	if err != nil {
		if errors.Is(err, repositories.ErrWorkProgramNotFound) {
			emitAudit(uc.audit, ctx, "work_program.revision_submit_denied",
				denialFields(actorID, in.WorkProgramID, "not_found", ""))
		}
		return nil, err
	}

	if !isAuthorOrSystemAdmin(actorID, actorRole, wp.AuthorID()) {
		emitAudit(uc.audit, ctx, "work_program.revision_submit_denied",
			denialFields(actorID, in.WorkProgramID, "forbidden", wp.SpecialtyCode()))
		return nil, fmt.Errorf("%w: actor %d is not the author (%d) and not system_admin",
			domain.ErrWorkProgramScopeForbidden, actorID, wp.AuthorID())
	}

	if err := wp.SubmitRevision(in.RevisionID); err != nil {
		switch {
		case errors.Is(err, domain.ErrRevisionNotFound):
			emitAudit(uc.audit, ctx, "work_program.revision_submit_denied",
				denialFields(actorID, in.WorkProgramID, "revision_not_found", wp.SpecialtyCode()))
		case errors.Is(err, domain.ErrInvalidStatusTransition):
			emitAudit(uc.audit, ctx, "work_program.revision_submit_denied",
				denialFields(actorID, in.WorkProgramID, "not_submittable", wp.SpecialtyCode()))
		}
		return nil, err
	}

	if err := uc.repo.Update(ctx, wp); err != nil {
		return nil, err
	}

	fields := successFields(actorID, wp.ID(), wp.SpecialtyCode(), string(wp.Status()))
	fields["revision_id"] = in.RevisionID
	emitAudit(uc.audit, ctx, "work_program.revision_submitted", fields)
	return wp, nil
}
