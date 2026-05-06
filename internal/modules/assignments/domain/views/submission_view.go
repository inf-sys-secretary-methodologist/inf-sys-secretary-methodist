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
// the student's display name, suitable for grading-list rendering.
//
// Field semantics mirror the persistence row 1:1; pointer fields are
// nil when the submission is still pending (no grade applied yet).
type SubmissionView struct {
	ID           int64
	AssignmentID int64
	StudentID    int64
	// StudentName comes from the users.name column via JOIN. It is
	// rendered as-is by the grading UI; per the project schema, name
	// is a single varchar column rather than first/last/middle.
	StudentName string
	GradeValue  *int
	Feedback    string
	GradedBy    *int64
	GradedAt    *time.Time
	// ReturnReason is the explanation captured when status=='returned';
	// empty string for non-returned rows.
	ReturnReason string
	// ReturnedBy is the user id of the actor who returned the
	// submission; nil for non-returned rows.
	ReturnedBy *int64
	// ReturnedAt is when the submission was returned; nil for
	// non-returned rows.
	ReturnedAt *time.Time
	Status     entities.SubmissionStatus
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// StudentAssignmentView is a read-side projection used by the student
// "My assignments" surface. It denormalises the parent assignment
// alongside the student's submission state so a single round-trip
// renders both the list page and the detail page without cross-
// aggregate composition in the handler.
//
// The submission columns are nullable for the same reason as
// SubmissionView; assignment columns are always present (a row only
// appears here because the parent assignment exists).
type StudentAssignmentView struct {
	// Assignment columns (always present).
	AssignmentID        int64
	Title               string
	Description         string
	Subject             string
	GroupName           string
	MaxScore            int
	DueDate             *time.Time
	AssignmentCreatedAt time.Time
	AssignmentUpdatedAt time.Time

	// Submission columns (always present because the JOIN selects rows
	// where a submission exists for this student).
	SubmissionID        int64
	StudentID           int64
	GradeValue          *int
	Feedback            string
	GradedBy            *int64
	GradedAt            *time.Time
	ReturnReason        string
	ReturnedBy          *int64
	ReturnedAt          *time.Time
	Status              entities.SubmissionStatus
	SubmissionCreatedAt time.Time
	SubmissionUpdatedAt time.Time
}
