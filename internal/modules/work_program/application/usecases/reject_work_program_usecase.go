package usecases

import (
	"context"
	"errors"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
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

// Execute is a stub. Real implementation lands in the matching GREEN commit.
func (uc *RejectWorkProgramUseCase) Execute(_ context.Context, _ int64, _ string, _ RejectWorkProgramInput) (*entities.WorkProgram, error) {
	return nil, errors.New("work_program: RejectWorkProgramUseCase not implemented yet (RED)")
}
