package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/application/usecases"
)

// StudentGroupResolverPG resolves a student user to their schedule group by
// joining the 1C-synced external_students record (group_name) to student_groups
// by name. It queries those tables directly via SQL and imports no other
// module's Go code, honoring the no-cross-module-import rule.
type StudentGroupResolverPG struct {
	db *sql.DB
}

var _ usecases.StudentGroupResolver = (*StudentGroupResolverPG)(nil)

// NewStudentGroupResolverPG creates a new StudentGroupResolverPG.
func NewStudentGroupResolverPG(db *sql.DB) *StudentGroupResolverPG {
	return &StudentGroupResolverPG{db: db}
}

// ResolveGroupID returns the student's group id, or ErrStudentGroupNotFound.
func (r *StudentGroupResolverPG) ResolveGroupID(ctx context.Context, userID int64) (int64, error) {
	const query = `
		SELECT sg.id
		FROM external_students es
		JOIN student_groups sg ON sg.name = es.group_name
		WHERE es.local_user_id = $1 AND es.is_active = TRUE
		ORDER BY sg.id
		LIMIT 1`

	var groupID int64
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&groupID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, usecases.ErrStudentGroupNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("failed to resolve student group: %w", err)
	}
	return groupID, nil
}
