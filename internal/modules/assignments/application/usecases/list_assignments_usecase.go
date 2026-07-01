package usecases

import (
	"context"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/entities"
)

// DefaultListLimit and MaxListLimit cap page sizes for assignment listings.
// Callers passing a non-positive Limit get the default; values above the
// max are clamped down. Bounded result sets keep the API predictable
// under accidental "page_size=999999" scrapes.
const (
	DefaultListLimit = 50
	MaxListLimit     = 100
)

// CallerScope expresses the slice of the system a caller is allowed to
// query. The use case maps this to a repository filter rather than
// reasoning about HTTP roles directly — keeping authentication strings
// out of the application layer. Handlers translate roles to scope at
// the seam.
type CallerScope struct {
	// UserID identifies the caller. Always required.
	UserID int64
	// Unrestricted is true for methodist / academic_secretary /
	// system_admin — they see every assignment. False for teacher: the
	// use case forces a TeacherID filter to UserID.
	Unrestricted bool
}

// ListAssignmentsInput is the use-case input contract.
type ListAssignmentsInput struct {
	Caller    CallerScope
	Subject   string
	GroupName string
	Limit     int
	Offset    int
}

// ListAssignmentsOutput is the use-case output contract.
type ListAssignmentsOutput struct {
	Items []*entities.Assignment
	Total int
}

// ListAssignmentsUseCase returns assignments visible to the caller.
type ListAssignmentsUseCase struct {
	repo AssignmentRepository
}

// NewListAssignmentsUseCase wires the use case.
func NewListAssignmentsUseCase(repo AssignmentRepository) *ListAssignmentsUseCase {
	return &ListAssignmentsUseCase{repo: repo}
}

// Execute builds a repository filter from the caller's scope, queries
// the repository and returns the assignments visible to that caller.
//
// Caller-scope mapping:
//   - Unrestricted=true → no TeacherID filter (methodist / secretary / admin)
//   - Unrestricted=false → forced TeacherID = caller.UserID (teacher)
//
// Pagination policy lives here, not in the repository, so callers can
// not opt out by passing huge Limit values: <=0 defaults to
// DefaultListLimit, anything above MaxListLimit is clamped down.
func (uc *ListAssignmentsUseCase) Execute(ctx context.Context, in ListAssignmentsInput) (ListAssignmentsOutput, error) {
	filter := AssignmentListFilter{
		Subject:   in.Subject,
		GroupName: in.GroupName,
		Limit:     clampLimit(in.Limit),
		Offset:    clampOffset(in.Offset),
	}
	if !in.Caller.Unrestricted {
		teacherID := in.Caller.UserID
		filter.TeacherID = &teacherID
	}

	res, err := uc.repo.List(ctx, filter)
	if err != nil {
		return ListAssignmentsOutput{}, fmt.Errorf("list assignments: %w", err)
	}
	return ListAssignmentsOutput(res), nil
}

func clampLimit(v int) int {
	switch {
	case v <= 0:
		return DefaultListLimit
	case v > MaxListLimit:
		return MaxListLimit
	default:
		return v
	}
}

func clampOffset(v int) int {
	if v < 0 {
		return 0
	}
	return v
}
