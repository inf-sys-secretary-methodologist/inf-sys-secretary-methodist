package usecases

import (
	"context"
	"fmt"

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

// Execute fetches the assignment, then enforces the caller-scope rule
// before returning it. Errors surface domain sentinels:
//
//   - repositories.ErrAssignmentNotFound       → 404
//   - entities.ErrAssignmentScopeForbidden     → 403
func (uc *GetAssignmentUseCase) Execute(ctx context.Context, in GetAssignmentInput) (*entities.Assignment, error) {
	a, err := uc.repo.GetByID(ctx, in.AssignmentID)
	if err != nil {
		return nil, fmt.Errorf("get assignment: %w", err)
	}
	if !in.Caller.Unrestricted && a.TeacherID() != in.Caller.UserID {
		return nil, fmt.Errorf("%w: user %d is not the author (%d)",
			entities.ErrAssignmentScopeForbidden, in.Caller.UserID, a.TeacherID())
	}
	return a, nil
}
