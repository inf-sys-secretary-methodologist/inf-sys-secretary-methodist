package usecases

import (
	"context"
	"errors"
)

// ErrStudentGroupNotFound is returned when a user is not linked to any student
// group (e.g. staff, or a student not yet synced from 1C).
var ErrStudentGroupNotFound = errors.New("student group not found for user")

// StudentGroupResolver resolves the schedule group a student user belongs to,
// so their group's lessons can be included in a personal calendar feed. The
// link lives outside the schedule module (1C-synced student records), so this
// is a port implemented at the composition root.
type StudentGroupResolver interface {
	// ResolveGroupID returns the student's group id, or ErrStudentGroupNotFound.
	ResolveGroupID(ctx context.Context, userID int64) (int64, error)
}
