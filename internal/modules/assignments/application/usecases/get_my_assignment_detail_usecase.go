package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/views"
)

// GetMyAssignmentDetailInput is the use-case input contract.
type GetMyAssignmentDetailInput struct {
	AssignmentID int64
	StudentID    int64
}

// GetMyAssignmentDetailUseCase loads a single assignment alongside the
// student's own submission row and returns them as a denormalised view
// — the same shape the list endpoint uses, so the frontend renders
// the detail page with no extra mapping.
type GetMyAssignmentDetailUseCase struct {
	assignmentRepo repositories.AssignmentRepository
	submissionRepo repositories.SubmissionRepository
}

// NewGetMyAssignmentDetailUseCase wires the use case. Failure-closed:
// nil assignment repo OR nil submission repo panics rather than
// surfacing nil-pointer dereferences at request time.
func NewGetMyAssignmentDetailUseCase(
	assignmentRepo repositories.AssignmentRepository,
	submissionRepo repositories.SubmissionRepository,
) *GetMyAssignmentDetailUseCase {
	if assignmentRepo == nil {
		panic("assignments: NewGetMyAssignmentDetailUseCase requires non-nil assignmentRepo")
	}
	if submissionRepo == nil {
		panic("assignments: NewGetMyAssignmentDetailUseCase requires non-nil submissionRepo")
	}
	return &GetMyAssignmentDetailUseCase{
		assignmentRepo: assignmentRepo,
		submissionRepo: submissionRepo,
	}
}

// Execute loads (assignment, submission) for the (assignmentID, studentID)
// pair, enforces the read-side ownership invariant, and assembles the
// denormalised view. Surfaces sentinels:
//
//   - repositories.ErrAssignmentNotFound  → 404
//   - repositories.ErrSubmissionNotFound  → 404 (no submission yet)
//   - entities.ErrSubmissionOwnerOnly     → 403 (defense-in-depth, see ADR-2)
func (uc *GetMyAssignmentDetailUseCase) Execute(ctx context.Context, in GetMyAssignmentDetailInput) (*views.StudentAssignmentView, error) {
	if in.StudentID <= 0 {
		return nil, errors.New("get my assignment: student id must be positive")
	}
	if in.AssignmentID <= 0 {
		return nil, errors.New("get my assignment: assignment id must be positive")
	}

	a, err := uc.assignmentRepo.GetByID(ctx, in.AssignmentID)
	if err != nil {
		return nil, fmt.Errorf("get my assignment: load assignment: %w", err)
	}

	s, err := uc.submissionRepo.GetByAssignmentAndStudent(ctx, in.AssignmentID, in.StudentID)
	if err != nil {
		return nil, fmt.Errorf("get my assignment: load submission: %w", err)
	}

	if err := s.AuthorizeReader(in.StudentID); err != nil {
		return nil, err
	}

	return buildStudentAssignmentView(a, s), nil
}

// buildStudentAssignmentView merges the assignment + submission into the
// flat read-model. Kept private to the use case — repository impls have
// their own scan-time construction (single SQL JOIN); this path is for
// the detail endpoint where two round-trips are unavoidable.
func buildStudentAssignmentView(a *entities.Assignment, s *entities.Submission) *views.StudentAssignmentView {
	return &views.StudentAssignmentView{
		AssignmentID:        a.ID,
		Title:               a.Title(),
		Description:         a.Description(),
		Subject:             a.Subject(),
		GroupName:           a.GroupName(),
		MaxScore:            a.MaxScore(),
		DueDate:             a.DueDate(),
		AssignmentCreatedAt: a.CreatedAt(),
		AssignmentUpdatedAt: a.UpdatedAt(),

		SubmissionID:        s.ID,
		StudentID:           s.StudentID,
		GradeValue:          s.GradeValue(),
		Feedback:            s.Feedback(),
		GradedBy:            s.GradedBy(),
		GradedAt:            s.GradedAt(),
		ReturnReason:        s.ReturnReason(),
		ReturnedBy:          s.ReturnedBy(),
		ReturnedAt:          s.ReturnedAt(),
		Status:              s.Status(),
		SubmissionCreatedAt: s.CreatedAt(),
		SubmissionUpdatedAt: s.UpdatedAt(),
	}
}
