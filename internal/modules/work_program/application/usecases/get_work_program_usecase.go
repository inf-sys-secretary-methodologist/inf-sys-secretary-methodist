package usecases

import (
	"context"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
)

// GetWorkProgramInput is the public DTO. The actor (and role) flow
// through Execute as positional arguments — handlers wire those from
// the JWT subject + role separately from the request path.
type GetWorkProgramInput struct {
	ID int64
}

// getWorkProgramRepo is the narrow port: load by id only. Get does not
// mutate; using a narrow port keeps use-case tests free of unused
// Save / Update / Delete wiring.
type getWorkProgramRepo interface {
	GetByID(ctx context.Context, id int64) (*entities.WorkProgram, error)
}

// GetWorkProgramUseCase hydrates and authorizes a WorkProgram read.
// View-rights matrix per ADR-018 ADR-5 is encoded в canViewWorkProgram.
type GetWorkProgramUseCase struct {
	repo  getWorkProgramRepo
	audit AuditSink
}

// NewGetWorkProgramUseCase wires the use case. Repo is required.
func NewGetWorkProgramUseCase(repo getWorkProgramRepo, audit AuditSink) *GetWorkProgramUseCase {
	if repo == nil {
		panic("work_program: NewGetWorkProgramUseCase requires non-nil repo")
	}
	return &GetWorkProgramUseCase{repo: repo, audit: audit}
}

// Execute loads the WorkProgram by id and authorizes the read:
//  1. Repository errors (ErrWorkProgramNotFound, transport) propagate
//     without an audit event — reads audit only on denial. Not-found
//     events would otherwise flood the log with ID-typo noise.
//  2. View-rights are evaluated via canViewWorkProgram (per ADR-018
//     ADR-5). Denied reads emit "work_program.view_denied"
//     (reason=forbidden) so privilege-escalation attempts are visible
//     in /admin/audit-logs.
func (uc *GetWorkProgramUseCase) Execute(ctx context.Context, actorID int64, actorRole string, in GetWorkProgramInput) (*entities.WorkProgram, error) {
	wp, err := uc.repo.GetByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}

	if !canViewWorkProgram(actorID, actorRole, wp) {
		emitAudit(uc.audit, ctx, "work_program.view_denied",
			denialFields(actorID, in.ID, "forbidden", wp.SpecialtyCode()))
		return nil, fmt.Errorf("%w: actor %d (role %q) cannot view work_program %d",
			domain.ErrWorkProgramScopeForbidden, actorID, actorRole, in.ID)
	}

	return wp, nil
}

// canViewWorkProgram encodes the ADR-018 ADR-5 view-rights matrix.
//
//	system_admin / methodist / academic_secretary → see every status
//	teacher                                       → own at any status OR any author's approved
//	student                                       → only approved (273-ФЗ ст. 29 mandatory openness)
//	anything else                                 → denied unconditionally
//
// Centralizing the predicate here lets handlers and downstream read
// projections (List endpoint in PR 3b) reuse the same authorization
// logic without duplicating the role-string literals.
func canViewWorkProgram(actorID int64, actorRole string, wp *entities.WorkProgram) bool {
	switch actorRole {
	case "system_admin", "methodist", "academic_secretary":
		return true
	case "teacher":
		return wp.AuthorID() == actorID || wp.Status() == domain.StatusApproved
	case "student":
		return wp.Status() == domain.StatusApproved
	default:
		return false
	}
}
