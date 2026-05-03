package repositories

import "context"

// TeacherScopeRepository resolves the canonical list of student groups a
// teacher is associated with. The list is the source of truth for the
// analytics teacher scope filter — the handler builds a *TeacherScope
// from this whitelist before delegating to use cases.
//
// Implementations read authoritative scheduling data (schedule_lessons
// joined with student_groups). The cross-table query is a persistence
// detail; the analytics module does NOT import the schedule Go module.
type TeacherScopeRepository interface {
	// ListGroupNames returns the deduplicated, ordered list of group
	// names the teacher is assigned to teach. An unassigned teacher
	// yields an empty slice (deny-all scope).
	ListGroupNames(ctx context.Context, teacherID int64) ([]string, error)
}
