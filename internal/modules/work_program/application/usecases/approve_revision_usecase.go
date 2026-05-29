package usecases

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
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

// Execute approves a revision. STUB — real flow lands in GREEN.
func (uc *ApproveRevisionUseCase) Execute(ctx context.Context, actorID int64, actorRole string, in ApproveRevisionInput) (*entities.WorkProgram, error) {
	_ = uc.repo
	_ = uc.audit
	return nil, nil
}
