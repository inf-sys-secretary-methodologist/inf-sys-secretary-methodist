package persistence

import (
	"context"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/application/usecases"
)

// Compile-time assertion that the adapter satisfies the application port.
var _ usecases.TeacherScopeResolver = (*TeacherScopeResolverPG)(nil)

// TeacherScopeResolverPG resolves a teacher's disciplines from the schedule:
// the disciplines they are scheduled to teach. It reads schedule_lessons
// directly at the SQL layer (no cross-module Go import — the same pattern
// analytics uses for teacher→group scoping), returning disciplines(id)
// values that match student_debts.discipline_id after migration 051.
type TeacherScopeResolverPG struct {
	db DBTX
}

// NewTeacherScopeResolverPG constructs the adapter. db can be `*sql.DB`
// (default DI) or `*sql.Tx`.
func NewTeacherScopeResolverPG(db DBTX) *TeacherScopeResolverPG {
	return &TeacherScopeResolverPG{db: db}
}

// DisciplineIDsForTeacher returns the distinct discipline ids the teacher is
// scheduled to teach, ordered for a stable result. An empty result (the
// teacher is scheduled for nothing) is not an error.
func (r *TeacherScopeResolverPG) DisciplineIDsForTeacher(ctx context.Context, teacherID int64) ([]int64, error) {
	const query = `SELECT DISTINCT discipline_id FROM schedule_lessons WHERE teacher_id = $1 ORDER BY discipline_id`

	rows, err := r.db.QueryContext(ctx, query, teacherID)
	if err != nil {
		return nil, fmt.Errorf("student_debts: teacher scope: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("student_debts: teacher scope scan: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("student_debts: teacher scope iter: %w", err)
	}
	return ids, nil
}
