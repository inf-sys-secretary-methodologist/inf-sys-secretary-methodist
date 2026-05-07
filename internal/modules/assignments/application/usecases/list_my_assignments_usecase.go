package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/views"
)

// MyAssignmentsRepository is the narrow read-side port the
// ListMyAssignmentsUseCase needs. Its single method intentionally
// matches SubmissionRepository.ListByStudent so the production wiring
// passes the same concrete repo, while tests can supply a tiny fake.
type MyAssignmentsRepository interface {
	ListByStudent(ctx context.Context, studentID int64, status *entities.SubmissionStatus) ([]views.StudentAssignmentView, error)
}

// ListMyAssignmentsInput is the use-case input contract.
type ListMyAssignmentsInput struct {
	// StudentID is the authenticated caller. Always required and must
	// be > 0 — defense in depth on top of the handler whitelist.
	StudentID int64
	// Status, when non-nil, filters submissions to a single lifecycle
	// state (pending / graded / returned).
	Status *entities.SubmissionStatus
}

// ListMyAssignmentsUseCase returns the student's My Assignments view —
// every assignment where the student has a submission row, joined with
// the parent assignment's metadata in one round-trip.
type ListMyAssignmentsUseCase struct {
	repo MyAssignmentsRepository
}

// NewListMyAssignmentsUseCase wires the use case. Failure-closed: a nil
// repo at construction time panics rather than letting requests reach
// a nil-pointer dereference deeper in the call stack.
func NewListMyAssignmentsUseCase(repo MyAssignmentsRepository) *ListMyAssignmentsUseCase {
	if repo == nil {
		panic("assignments: NewListMyAssignmentsUseCase requires non-nil repo")
	}
	return &ListMyAssignmentsUseCase{repo: repo}
}

// Execute fetches the student's denormalised assignment list. A
// non-positive student id is rejected here so the invariant holds for
// every caller (handler already enforces, but the use case must too —
// future callers may bypass the HTTP layer).
func (uc *ListMyAssignmentsUseCase) Execute(ctx context.Context, in ListMyAssignmentsInput) ([]views.StudentAssignmentView, error) {
	if in.StudentID <= 0 {
		return nil, errors.New("list my assignments: student id must be positive")
	}
	out, err := uc.repo.ListByStudent(ctx, in.StudentID, in.Status)
	if err != nil {
		return nil, fmt.Errorf("list my assignments: %w", err)
	}
	return out, nil
}
