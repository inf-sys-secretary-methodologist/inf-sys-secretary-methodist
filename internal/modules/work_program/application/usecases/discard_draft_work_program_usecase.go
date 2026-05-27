package usecases

import (
	"context"
	"errors"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
)

// DiscardDraftWorkProgramInput is the public DTO.
type DiscardDraftWorkProgramInput struct {
	ID int64
}

// discardDraftWorkProgramRepo is the narrow port: load by id (so we
// can authorize against AuthorID) + write back the archived aggregate.
type discardDraftWorkProgramRepo interface {
	GetByID(ctx context.Context, id int64) (*entities.WorkProgram, error)
	Update(ctx context.Context, wp *entities.WorkProgram) error
}

// DiscardDraftWorkProgramUseCase archives a draft WorkProgram without
// going through approval — author abandons their own work. Authorized
// for the author or for a system_admin override (mirrors Submit's
// authorship predicate — methodist is intentionally excluded from
// non-author destructive actions even on draft state).
type DiscardDraftWorkProgramUseCase struct {
	repo  discardDraftWorkProgramRepo
	audit AuditSink
}

// NewDiscardDraftWorkProgramUseCase wires the use case. Repo is required.
func NewDiscardDraftWorkProgramUseCase(repo discardDraftWorkProgramRepo, audit AuditSink) *DiscardDraftWorkProgramUseCase {
	if repo == nil {
		panic("work_program: NewDiscardDraftWorkProgramUseCase requires non-nil repo")
	}
	return &DiscardDraftWorkProgramUseCase{repo: repo, audit: audit}
}

// Execute is a stub. Real implementation lands in the matching GREEN commit.
func (uc *DiscardDraftWorkProgramUseCase) Execute(_ context.Context, _ int64, _ string, _ DiscardDraftWorkProgramInput) (*entities.WorkProgram, error) {
	return nil, errors.New("work_program: DiscardDraftWorkProgramUseCase not implemented yet (RED)")
}
