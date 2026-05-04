// Package views holds read-side projections for the assignments
// bounded context. These types intentionally denormalise data the UI
// needs in one fetch (e.g. submission rows joined with the student's
// display name) so handlers do not have to compose multiple read paths
// per request.
//
// Read models live in the domain package because they describe the
// shape of a domain query result, but they are deliberately separate
// from entities/value objects: views are flat, public-field DTOs and
// carry no behaviour.
package views

import (
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/entities"
)

// SubmissionView is a read-side projection of a submission joined with
// the student's display fields, suitable for grading-list rendering.
//
// Field semantics mirror the persistence row 1:1; pointer fields are
// nil when the submission is still pending (no grade applied yet).
type SubmissionView struct {
	ID               int64
	AssignmentID     int64
	StudentID        int64
	StudentFirstName string
	StudentLastName  string
	GradeValue       *int
	Feedback         string
	GradedBy         *int64
	GradedAt         *time.Time
	Status           entities.SubmissionStatus
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
