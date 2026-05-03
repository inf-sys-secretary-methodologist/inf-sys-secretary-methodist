package persistence

import (
	"context"
	"database/sql"
	"fmt"
)

// TeacherScopeRepositoryPG resolves the teacher → group whitelist by
// reading schedule_lessons and student_groups. The two tables live in
// the schedule module's storage namespace; reading them here is a
// persistence-layer detail — analytics does not import the schedule
// Go module, only its physical tables (modular monolith convention).
type TeacherScopeRepositoryPG struct {
	db *sql.DB
}

// NewTeacherScopeRepositoryPG constructs the repository.
func NewTeacherScopeRepositoryPG(db *sql.DB) *TeacherScopeRepositoryPG {
	return &TeacherScopeRepositoryPG{db: db}
}

// ListGroupNames returns the canonical group-name whitelist for teacherID.
// Empty slice (not nil error) is returned for unassigned teachers — the
// caller turns that into a deny-all TeacherScope.
func (r *TeacherScopeRepositoryPG) ListGroupNames(ctx context.Context, teacherID int64) ([]string, error) {
	const query = `SELECT DISTINCT sg.name FROM schedule_lessons sl JOIN student_groups sg ON sg.id = sl.group_id WHERE sl.teacher_id = $1`

	rows, err := r.db.QueryContext(ctx, query, teacherID)
	if err != nil {
		return nil, fmt.Errorf("teacher_scope: failed to query groups: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("teacher_scope: failed to scan group name: %w", err)
		}
		names = append(names, name)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("teacher_scope: rows iteration error: %w", err)
	}
	return names, nil
}
