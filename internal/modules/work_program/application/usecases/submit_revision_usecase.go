package usecases

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
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

// Execute submits a revision. STUB — real flow lands in GREEN.
func (uc *SubmitRevisionUseCase) Execute(ctx context.Context, actorID int64, actorRole string, in SubmitRevisionInput) (*entities.WorkProgram, error) {
	_ = uc.repo
	_ = uc.audit
	return nil, nil
}
