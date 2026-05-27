package usecases

import (
	"context"
	"errors"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
)

// CreateWorkProgramInput is the public request DTO. The actor
// (created_by → AuthorID) and actor's role are supplied as separate
// arguments to Execute so handlers wire the JWT subject + role
// explicitly rather than through the same struct that may be
// deserialised from untrusted JSON.
type CreateWorkProgramInput struct {
	DisciplineID       int64
	SpecialtyCode      string
	ApplicableFromYear int
	Title              string
	Annotation         string
}

// createWorkProgramRepo is the narrow port the use case requires from
// the persistence layer. Defining it here (rather than importing the
// wide *WorkProgramRepository) keeps use-case tests free of GetByID /
// Update / Delete / List wiring they do not exercise.
type createWorkProgramRepo interface {
	Save(ctx context.Context, wp *entities.WorkProgram) error
}

// CreateWorkProgramUseCase persists a fresh draft WorkProgram and
// emits the matching audit event. Role gate per ADR-018 ADR-5: only
// teacher / methodist / system_admin may create. academic_secretary is
// view-only on РПД (curriculum is their author surface); student is
// denied. methodist is allowed as a backup author per ADR-018 ADR-5
// ("резервно creates если teacher не успевает").
type CreateWorkProgramUseCase struct {
	repo  createWorkProgramRepo
	audit AuditSink
}

// NewCreateWorkProgramUseCase wires the use case. The repo is required
// (non-nil): a nil dependency would let requests reach a panic deeper
// in the call stack instead of failing during DI wiring. Nil audit
// sink is tolerated (tests may opt out).
func NewCreateWorkProgramUseCase(repo createWorkProgramRepo, audit AuditSink) *CreateWorkProgramUseCase {
	if repo == nil {
		panic("work_program: NewCreateWorkProgramUseCase requires non-nil repo")
	}
	return &CreateWorkProgramUseCase{repo: repo, audit: audit}
}

// Execute is a stub. Real implementation lands in the matching GREEN commit.
func (uc *CreateWorkProgramUseCase) Execute(_ context.Context, _ int64, _ string, _ CreateWorkProgramInput) (*entities.WorkProgram, error) {
	return nil, errors.New("work_program: CreateWorkProgramUseCase not implemented yet (RED)")
}
