package persistence

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/repositories"
)

// Compile-time assertion that the PG impl satisfies the wide port
// declared in application/usecases (DIP). Catches signature drift at
// the impl's compile site rather than only at DI wiring.
var _ usecases.WorkProgramRepository = (*WorkProgramRepositoryPG)(nil)

// WorkProgramRepositoryPG is the SQL implementation of
// WorkProgramRepository. Accepts DBTX (not *sql.DB) so the same struct
// works in single-connection mode and against `*sql.Tx`.
type WorkProgramRepositoryPG struct {
	db DBTX
}

// NewWorkProgramRepositoryPG constructs the repository. db can be
// `*sql.DB` (default DI) or `*sql.Tx` (future transactional paths).
func NewWorkProgramRepositoryPG(db DBTX) *WorkProgramRepositoryPG {
	return &WorkProgramRepositoryPG{db: db}
}

// Save stub for PR 2a RED phase — real implementation lands in the
// matching GREEN commit.
func (r *WorkProgramRepositoryPG) Save(_ context.Context, _ *entities.WorkProgram) error {
	return repositories.ErrWorkProgramNotFound
}
