package usecases

import (
	"context"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/entities"
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
	assignmentRepo AssignmentRepository
	submissionRepo SubmissionRepository
}

// NewListSubmissionsUseCase wires the use case.
func NewListSubmissionsUseCase(
	assignmentRepo AssignmentRepository,
	submissionRepo SubmissionRepository,
) *ListSubmissionsUseCase {
	return &ListSubmissionsUseCase{
		assignmentRepo: assignmentRepo,
		submissionRepo: submissionRepo,
	}
}

// Execute loads the parent assignment, enforces caller-scope, and then
// returns the submission read-models for that assignment. Errors
// surface domain sentinels:
//
//   - ErrAssignmentNotFound       → 404
//   - entities.ErrAssignmentScopeForbidden     → 403
func (uc *ListSubmissionsUseCase) Execute(ctx context.Context, in ListSubmissionsInput) ([]views.SubmissionView, error) {
	a, err := uc.assignmentRepo.GetByID(ctx, in.AssignmentID)
	if err != nil {
		return nil, fmt.Errorf("list submissions: load assignment: %w", err)
	}
	if err := a.AuthorizeAccess(in.Caller.Unrestricted, in.Caller.UserID); err != nil {
		return nil, err
	}

	subs, err := uc.submissionRepo.ListByAssignment(ctx, in.AssignmentID, in.Status)
	if err != nil {
		return nil, fmt.Errorf("list submissions: %w", err)
	}
	return subs, nil
}
