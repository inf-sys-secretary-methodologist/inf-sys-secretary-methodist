package usecases

import (
	"context"
	"errors"

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

// Execute is a stub. Real implementation lands in the matching GREEN commit.
func (uc *GetWorkProgramUseCase) Execute(_ context.Context, _ int64, _ string, _ GetWorkProgramInput) (*entities.WorkProgram, error) {
	return nil, errors.New("work_program: GetWorkProgramUseCase not implemented yet (RED)")
}
