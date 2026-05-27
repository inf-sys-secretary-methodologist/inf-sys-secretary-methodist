package usecases

import (
	"context"
	"errors"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
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

// Execute is a stub. Real implementation lands in the matching GREEN commit.
func (uc *ApproveWorkProgramUseCase) Execute(_ context.Context, _ int64, _ string, _ ApproveWorkProgramInput) (*entities.WorkProgram, error) {
	return nil, errors.New("work_program: ApproveWorkProgramUseCase not implemented yet (RED)")
}
