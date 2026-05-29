package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/repositories"
)

// CreateRevisionInput is the public request DTO for proposing a лист
// актуализации on an existing РПД. The actor (→ revision AuthorID) and
// role flow through Execute as separate arguments so handlers wire the
// JWT subject explicitly. ChangeType is a string mapped to the domain
// enum inside the use case (a bad value fails the domain constructor).
type CreateRevisionInput struct {
	WorkProgramID int64
	ChangeType    string
	ChangeSummary string
	DiffPayload   []byte // optional structured before/after JSON
}

// createRevisionRepo is the narrow load-mutate-persist port.
type createRevisionRepo interface {
	GetByID(ctx context.Context, id int64) (*entities.WorkProgram, error)
	Update(ctx context.Context, wp *entities.WorkProgram) error
}

// CreateRevisionUseCase appends a draft Revision to an approved /
// needs_revision РПД. Author-scoped (author or system_admin) per
// ADR-10 — the РПД author proposes актуализация; a methodist approves
// it later via the approve flow.
type CreateRevisionUseCase struct {
	repo  createRevisionRepo
	audit AuditSink
}

// NewCreateRevisionUseCase wires the use case. Repo is required.
func NewCreateRevisionUseCase(repo createRevisionRepo, audit AuditSink) *CreateRevisionUseCase {
	if repo == nil {
		panic("work_program: NewCreateRevisionUseCase requires non-nil repo")
	}
	return &CreateRevisionUseCase{repo: repo, audit: audit}
}

// Execute runs the propose-revision flow:
//  1. Load by id; ErrWorkProgramNotFound → 'not_found' denial.
//  2. Authorize: actor must be author OR system_admin → otherwise
//     ErrWorkProgramScopeForbidden + 'forbidden' denial.
//  3. Build the Revision through NewRevision (invariant gate);
//     ErrInvalidWorkProgram → 'invalid' denial.
//  4. wp.AddRevision applies the parent-status + monotonic-number
//     gate; ErrRevisionNotPermitted → 'not_permitted' denial.
//  5. Persist via repo.Update. Transport errors propagate without
//     audit (audit log = policy decisions, not infra outages).
func (uc *CreateRevisionUseCase) Execute(ctx context.Context, actorID int64, actorRole string, in CreateRevisionInput) (*entities.WorkProgram, error) {
	wp, err := uc.repo.GetByID(ctx, in.WorkProgramID)
	if err != nil {
		if errors.Is(err, repositories.ErrWorkProgramNotFound) {
			emitAudit(uc.audit, ctx, "work_program.revision_create_denied",
				denialFields(actorID, in.WorkProgramID, "not_found", ""))
		}
		return nil, err
	}

	if !isAuthorOrSystemAdmin(actorID, actorRole, wp.AuthorID()) {
		emitAudit(uc.audit, ctx, "work_program.revision_create_denied",
			denialFields(actorID, in.WorkProgramID, "forbidden", wp.SpecialtyCode()))
		return nil, fmt.Errorf("%w: actor %d is not the author (%d) and not system_admin",
			domain.ErrWorkProgramScopeForbidden, actorID, wp.AuthorID())
	}

	rev, err := entities.NewRevision(entities.NewRevisionInput{
		WorkProgramID:  wp.ID(),
		RevisionNumber: wp.NextRevisionNumber(),
		ChangeType:     domain.RevisionChangeType(in.ChangeType),
		ChangeSummary:  in.ChangeSummary,
		AuthorID:       actorID,
		DiffPayload:    in.DiffPayload,
	})
	if err != nil {
		emitAudit(uc.audit, ctx, "work_program.revision_create_denied",
			denialFields(actorID, in.WorkProgramID, "invalid", wp.SpecialtyCode()))
		return nil, err
	}

	if err := wp.AddRevision(rev); err != nil {
		if errors.Is(err, domain.ErrRevisionNotPermitted) {
			emitAudit(uc.audit, ctx, "work_program.revision_create_denied",
				denialFields(actorID, in.WorkProgramID, "not_permitted", wp.SpecialtyCode()))
		}
		return nil, err
	}

	if err := uc.repo.Update(ctx, wp); err != nil {
		return nil, err
	}

	fields := successFields(actorID, wp.ID(), wp.SpecialtyCode(), string(wp.Status()))
	fields["revision_number"] = rev.RevisionNumber()
	fields["change_type"] = string(rev.ChangeType())
	emitAudit(uc.audit, ctx, "work_program.revision_created", fields)
	return wp, nil
}
