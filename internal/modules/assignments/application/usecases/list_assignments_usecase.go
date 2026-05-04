package usecases

import (
	"context"
	"errors"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/repositories"
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
	repo repositories.AssignmentRepository
}

// NewListAssignmentsUseCase wires the use case.
func NewListAssignmentsUseCase(repo repositories.AssignmentRepository) *ListAssignmentsUseCase {
	return &ListAssignmentsUseCase{repo: repo}
}

// Execute is the use-case entry point. Stub returns ErrNotImplemented
// during the RED stage of the TDD cycle — it makes the failing tests
// compile while clearly signalling the missing behaviour.
func (uc *ListAssignmentsUseCase) Execute(ctx context.Context, in ListAssignmentsInput) (ListAssignmentsOutput, error) {
	return ListAssignmentsOutput{}, errors.New("ListAssignmentsUseCase.Execute: not implemented")
}
