package entities

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrInvalidAssignment signals a violation of one of the Assignment
// construction invariants (empty title/group_name, non-positive max_score).
// Handlers map this sentinel to HTTP 422.
var ErrInvalidAssignment = errors.New("assignments: invalid assignment")

// ErrAssignmentScopeForbidden indicates that a user is not authorized to
// operate on a particular Assignment — typically because the user is a
// teacher who did not author the Assignment. Handlers map this sentinel
// to HTTP 403.
//
// This is a stricter check than the analytics-level TeacherScope, which
// only filters by group membership. An assignment carries a single
// authoring TeacherID; even if two teachers share a group, only the
// author may grade — defense in depth against accidental cross-grading.
var ErrAssignmentScopeForbidden = errors.New("assignments: caller cannot operate on this assignment")

// Assignment is the aggregate root for a published academic task.
// Submissions belong to it (one per student in the group).
type Assignment struct {
	ID          int64
	title       string
	description string
	teacherID   int64
	groupName   string
	subject     string
	maxScore    int
	dueDate     *time.Time
	createdAt   time.Time
	updatedAt   time.Time
}

// NewAssignmentParams bundles the constructor inputs so call sites stay
// readable when more optional fields are added (rubric, attachments, ...).
type NewAssignmentParams struct {
	Title       string
	Description string
	TeacherID   int64
	GroupName   string
	Subject     string
	MaxScore    int
	DueDate     *time.Time
	Now         time.Time
}

// NewAssignment validates invariants and returns a fresh Assignment.
//
// Invariants (mirroring the SQL CHECK constraints in migration 029):
//   - title trimmed-non-empty
//   - group_name trimmed-non-empty
//   - max_score > 0
//
// Each violation wraps ErrInvalidAssignment with the offending field so
// errors.Is still resolves the sentinel for the 422 mapping.
func NewAssignment(p NewAssignmentParams) (*Assignment, error) {
	title := strings.TrimSpace(p.Title)
	if title == "" {
		return nil, fmt.Errorf("%w: title must not be empty", ErrInvalidAssignment)
	}
	groupName := strings.TrimSpace(p.GroupName)
	if groupName == "" {
		return nil, fmt.Errorf("%w: group_name must not be empty", ErrInvalidAssignment)
	}
	if p.MaxScore <= 0 {
		return nil, fmt.Errorf("%w: max_score must be positive, got %d", ErrInvalidAssignment, p.MaxScore)
	}
	return &Assignment{
		title:       title,
		description: p.Description,
		teacherID:   p.TeacherID,
		groupName:   groupName,
		subject:     p.Subject,
		maxScore:    p.MaxScore,
		dueDate:     p.DueDate,
		createdAt:   p.Now,
		updatedAt:   p.Now,
	}, nil
}

// ReconstituteAssignment rebuilds an Assignment from authoritative
// storage. It bypasses NewAssignment's invariant checks because the
// values are already canonical (the DB enforces the same CHECKs at
// write time). Used exclusively by repository implementations.
func ReconstituteAssignment(
	id int64, title, description string, teacherID int64,
	groupName, subject string, maxScore int, dueDate *time.Time,
	createdAt, updatedAt time.Time,
) *Assignment {
	return &Assignment{
		ID:          id,
		title:       title,
		description: description,
		teacherID:   teacherID,
		groupName:   groupName,
		subject:     subject,
		maxScore:    maxScore,
		dueDate:     dueDate,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

// Title returns the assignment title.
func (a *Assignment) Title() string { return a.title }

// Description returns the assignment description.
func (a *Assignment) Description() string { return a.description }

// TeacherID returns the authoring teacher's user ID.
func (a *Assignment) TeacherID() int64 { return a.teacherID }

// GroupName returns the target student group.
func (a *Assignment) GroupName() string { return a.groupName }

// Subject returns the academic subject the assignment relates to.
func (a *Assignment) Subject() string { return a.subject }

// MaxScore returns the maximum score that can be awarded.
func (a *Assignment) MaxScore() int { return a.maxScore }

// DueDate returns the due-by timestamp, if any.
func (a *Assignment) DueDate() *time.Time { return a.dueDate }

// CreatedAt returns the creation timestamp.
func (a *Assignment) CreatedAt() time.Time { return a.createdAt }

// UpdatedAt returns the last-update timestamp.
func (a *Assignment) UpdatedAt() time.Time { return a.updatedAt }

// AuthorizeGrader returns nil if the user identified by userID is allowed
// to grade submissions on this assignment, ErrAssignmentScopeForbidden
// otherwise. Today the rule is "only the author"; if co-teaching is
// added later, this is the single place to extend.
func (a *Assignment) AuthorizeGrader(userID int64) error {
	if userID == a.teacherID {
		return nil
	}
	return fmt.Errorf("%w: user %d is not the author (%d)",
		ErrAssignmentScopeForbidden, userID, a.teacherID)
}

// AuthorizeAccess returns nil if the caller may read this assignment.
// "Unrestricted" callers (methodist / academic_secretary /
// system_admin at the application layer) always pass; otherwise the
// rule is the same as AuthorizeGrader — the user must be the author.
//
// Centralizing the read-side rule in the aggregate keeps the
// GetAssignment and ListSubmissions use cases from re-implementing
// (and accidentally diverging from) the same predicate.
func (a *Assignment) AuthorizeAccess(unrestricted bool, userID int64) error {
	if unrestricted {
		return nil
	}
	return a.AuthorizeGrader(userID)
}

// NewSubmissionScore is the domain method that owns the cross-aggregate
// rule "a submission's score must be within this assignment's
// maxScore." Moving this construction here closes the leak the
// SaveGrade use case used to carry, where the application layer had
// to know that a Score needs (value, max) and that max comes from the
// Assignment aggregate.
//
// The non-negative invariant is delegated to NewScore; the upper-bound
// check is enforced here because only Assignment knows its maxScore.
// Both rejections wrap ErrInvalidScore so the handler 422 mapping
// stays a single errors.Is dispatch.
func (a *Assignment) NewSubmissionScore(value int) (Score, error) {
	if value > a.maxScore {
		return Score{}, fmt.Errorf("%w: value %d exceeds max %d", ErrInvalidScore, value, a.maxScore)
	}
	return NewScore(value)
}
