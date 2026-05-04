package usecases

import (
	"context"
	"errors"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/repositories"
)

// GetAssignmentInput is the use-case input contract.
type GetAssignmentInput struct {
	Caller       CallerScope
	AssignmentID int64
}

// GetAssignmentUseCase loads a single assignment, enforcing the caller
// can see it. It exists because the list endpoint already filters by
// scope at the SQL level — a single-fetch endpoint would otherwise
// expose any assignment ID to any non-student.
type GetAssignmentUseCase struct {
	repo repositories.AssignmentRepository
}

// NewGetAssignmentUseCase wires the use case.
func NewGetAssignmentUseCase(repo repositories.AssignmentRepository) *GetAssignmentUseCase {
	return &GetAssignmentUseCase{repo: repo}
}

// Execute is the entry point. Stub returns ErrNotImplemented during the
// RED stage — keeps tests compiling while the failing assertions drive
// the implementation.
func (uc *GetAssignmentUseCase) Execute(ctx context.Context, in GetAssignmentInput) (*entities.Assignment, error) {
	return nil, errors.New("GetAssignmentUseCase.Execute: not implemented")
}
