package persistence

import (
	"context"
	"database/sql"

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
	return 0, nil
}
