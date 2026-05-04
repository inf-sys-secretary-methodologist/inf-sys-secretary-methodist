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

// AssignmentRepository is the persistence port for Assignment aggregates.
type AssignmentRepository interface {
	// GetByID returns the Assignment with the given id or ErrAssignmentNotFound.
	GetByID(ctx context.Context, id int64) (*entities.Assignment, error)
}
