package usecases

import (
	"context"
	"errors"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/views"
)

// ListSubmissionsInput is the use-case input contract.
type ListSubmissionsInput struct {
	Caller       CallerScope
	AssignmentID int64
	// Status, when non-nil, filters submissions to that lifecycle state.
	// The handler is responsible for converting a raw query string into
	// a typed pointer (or rejecting it as 400 if the value is unknown).
	Status *entities.SubmissionStatus
}

// ListSubmissionsUseCase returns the per-student submission rows for a
// given assignment, after enforcing the same caller-scope rule as
// GetAssignment.
type ListSubmissionsUseCase struct {
	assignmentRepo repositories.AssignmentRepository
	submissionRepo repositories.SubmissionRepository
}

// NewListSubmissionsUseCase wires the use case.
func NewListSubmissionsUseCase(
	assignmentRepo repositories.AssignmentRepository,
	submissionRepo repositories.SubmissionRepository,
) *ListSubmissionsUseCase {
	return &ListSubmissionsUseCase{
		assignmentRepo: assignmentRepo,
		submissionRepo: submissionRepo,
	}
}

// Execute is the entry point. Stub returns ErrNotImplemented during the
// RED stage — keeps tests compiling while the failing assertions drive
// the implementation.
func (uc *ListSubmissionsUseCase) Execute(ctx context.Context, in ListSubmissionsInput) ([]views.SubmissionView, error) {
	return nil, errors.New("ListSubmissionsUseCase.Execute: not implemented")
}
