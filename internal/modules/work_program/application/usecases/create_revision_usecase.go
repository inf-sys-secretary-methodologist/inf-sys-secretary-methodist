package usecases

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
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

// Execute proposes a revision. STUB — real flow lands in GREEN.
func (uc *CreateRevisionUseCase) Execute(ctx context.Context, actorID int64, actorRole string, in CreateRevisionInput) (*entities.WorkProgram, error) {
	_ = uc.repo
	_ = uc.audit
	return nil, nil
}
