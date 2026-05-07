// Package repositories declares the persistence ports for the assignments
// module. Concrete implementations live in the infrastructure layer.
package repositories

import (
	"context"
	"errors"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/entities"
)

// ErrAssignmentNotFound signals that no Assignment exists for the
// requested ID. Use cases map this sentinel directly to a domain-level
// "not found" condition; handlers map it to HTTP 404.
var ErrAssignmentNotFound = errors.New("assignments: assignment not found")

// AssignmentListFilter narrows a List query. Zero-valued fields are
// treated as "no filter on this dimension". Limit/Offset are honored by
// the repository; a non-positive Limit means "no clamp at the
// repository layer" — use cases are responsible for choosing a
// sensible default to keep result sets bounded.
type AssignmentListFilter struct {
	// TeacherID, when non-nil, restricts results to assignments authored
	// by that teacher. The use case sets this for teacher callers; for
	// methodist/secretary/admin callers it stays nil so they see every
	// assignment.
	TeacherID *int64
	// Subject filters by exact match when non-empty.
	Subject string
	// GroupName filters by exact match when non-empty.
	GroupName string
	// Limit caps the number of returned items. Repositories must treat
	// values <= 0 as "no extra clamp" and rely on the caller's policy.
	Limit int
	// Offset is the starting index for pagination.
	Offset int
}

// AssignmentListResult bundles the page of items with the unfiltered
// total so the UI can render pagination controls without a second query.
type AssignmentListResult struct {
	Items []*entities.Assignment
	Total int
}

// AssignmentRepository is the persistence port for Assignment aggregates.
type AssignmentRepository interface {
	// GetByID returns the Assignment with the given id or ErrAssignmentNotFound.
	GetByID(ctx context.Context, id int64) (*entities.Assignment, error)

	// List returns a page of assignments matching the filter together with
	// the total number of matching rows (ignoring Limit/Offset). Empty
	// result is not an error.
	List(ctx context.Context, filter AssignmentListFilter) (AssignmentListResult, error)
}
