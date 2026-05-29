package usecases

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
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

// Execute rejects a revision. STUB — real flow lands in GREEN.
func (uc *RejectRevisionUseCase) Execute(ctx context.Context, actorID int64, actorRole string, in RejectRevisionInput) (*entities.WorkProgram, error) {
	_ = uc.repo
	_ = uc.audit
	return nil, nil
}
