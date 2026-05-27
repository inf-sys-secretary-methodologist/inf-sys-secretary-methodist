package usecases

import (
	"context"
	"errors"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
)

// SubmitWorkProgramInput is the public DTO. The actor (and role) flow
// through Execute as positional arguments — handlers wire those from
// the JWT subject + role separately from the request body.
type SubmitWorkProgramInput struct {
	ID int64
}

// submitWorkProgramRepo is the narrow port: load by id (so we can
// authorize against the row's AuthorID) + write back the transitioned
// aggregate.
type submitWorkProgramRepo interface {
	GetByID(ctx context.Context, id int64) (*entities.WorkProgram, error)
	Update(ctx context.Context, wp *entities.WorkProgram) error
}

// SubmitWorkProgramUseCase moves a draft WorkProgram into the
// pending_approval state. Author or system_admin may invoke it; the
// entity enforces the status invariant.
type SubmitWorkProgramUseCase struct {
	repo  submitWorkProgramRepo
	audit AuditSink
}

// NewSubmitWorkProgramUseCase wires the use case. Repo is required.
func NewSubmitWorkProgramUseCase(repo submitWorkProgramRepo, audit AuditSink) *SubmitWorkProgramUseCase {
	if repo == nil {
		panic("work_program: NewSubmitWorkProgramUseCase requires non-nil repo")
	}
	return &SubmitWorkProgramUseCase{repo: repo, audit: audit}
}

// Execute is a stub. Real implementation lands in the matching GREEN commit.
func (uc *SubmitWorkProgramUseCase) Execute(_ context.Context, _ int64, _ string, _ SubmitWorkProgramInput) (*entities.WorkProgram, error) {
	return nil, errors.New("work_program: SubmitWorkProgramUseCase not implemented yet (RED)")
}
