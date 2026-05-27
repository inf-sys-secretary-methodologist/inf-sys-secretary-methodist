// Wide repository ports for the work_program module. Defined in the
// application layer per CLAUDE.md DIP gate ("Repository interfaces —
// в пакете-потребителе (usecase/), НЕ в domain/. DIP по классике.").
//
// Sentinels (ErrWorkProgramNotFound etc.) live in
// internal/modules/work_program/domain/repositories — they encode
// domain values rather than persistence contracts, so handlers chain
// errors.Is against them without depending on the persistence port
// itself.
//
// Concrete implementations satisfy these ports via Go's structural
// typing, backed by compile-time `var _ usecases.<Port> = (*XxxPG)(nil)`
// assertions in the infrastructure layer (`*_pg.go` files) so signature
// drift surfaces at the implementation's compile site rather than only
// at DI wiring.
package usecases

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/repositories"
)

// WorkProgramRepository is the persistence port for WorkProgram
// aggregates. Implementations MUST persist the root + its inner
// aggregates (Goals, Competences, Topics, Assessments, References,
// Revisions) atomically — either the full aggregate lands or none of
// it does.
//
// Sentinel contract:
//   - repositories.ErrWorkProgramNotFound on missing rows
//   - repositories.ErrWorkProgramIdentityExists on uniqueness violation
//     against (discipline_id, specialty_code, applicable_from_year)
//   - repositories.ErrWorkProgramVersionConflict on stale-version Update
//
// PR 2a (v0.173.0) shipped Save; PR 2b (v0.174.0) adds GetByID with
// full child hydration; List / Update / Delete land in subsequent PRs
// of the persistence slice.
type WorkProgramRepository interface {
	// Save inserts a new WorkProgram aggregate (root row + all inner
	// aggregate rows) atomically. On success the generated id is
	// written back onto the root entity. Returns
	// repositories.ErrWorkProgramIdentityExists if a row with the
	// same identity tuple already exists.
	Save(ctx context.Context, wp *entities.WorkProgram) error

	// GetByID returns the WorkProgram aggregate with the given id,
	// hydrated through Reconstitute* — root + every populated child
	// collection (Goals, Competences, Topics, Assessments, References,
	// Revisions). Returns repositories.ErrWorkProgramNotFound when no
	// matching row exists.
	GetByID(ctx context.Context, id int64) (*entities.WorkProgram, error)

	// List returns a page of WorkProgram items matching the filter
	// together with the total number of matching rows (ignoring
	// Limit/Offset). Items carry root state only — list endpoints
	// stay cheap; callers needing full child hydration use GetByID.
	// An empty result is not an error.
	List(ctx context.Context, filter repositories.WorkProgramListFilter) (repositories.WorkProgramListResult, error)
}
