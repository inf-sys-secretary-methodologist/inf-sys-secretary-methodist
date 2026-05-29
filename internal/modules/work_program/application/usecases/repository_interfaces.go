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

	// Update writes the (already-mutated) aggregate back atomically:
	// UPDATE root with optimistic-lock guard (WHERE id=? AND version=?)
	// then delete + reinsert every child collection inside the same
	// tx. On RowsAffected == 0 the impl distinguishes:
	//   - row missing entirely → ErrWorkProgramNotFound
	//   - row exists but stale version → ErrWorkProgramVersionConflict
	// On success the entity's version is bumped to reflect the new
	// row state so callers see a consistent post-update view without
	// reloading.
	Update(ctx context.Context, wp *entities.WorkProgram) error

	// Delete removes the WorkProgram row by id. Returns
	// ErrWorkProgramNotFound when no row is deleted. Migration 048
	// ON DELETE CASCADE handles child cleanup automatically; the
	// repo issues a single DELETE statement.
	//
	// Note: РПД are normally archived (status=archived), never deleted —
	// see ADR-1 / Рособрнадзор 6-year retention. Delete exists for
	// admin-grade cleanup paths (test fixtures, GDPR-style erasure).
	Delete(ctx context.Context, id int64) error
}

// MinobrnaukiOrderRepository is the persistence port for MinobrnaukiOrder
// entities (приказы Минобрнауки) per ADR-11 — the external regulatory
// trigger for РПД revisions. Implementations persist the order row plus
// its affected-work-program junction (minobrnauki_order_affected)
// atomically.
//
// Sentinel contract:
//   - repositories.ErrMinobrnaukiOrderNotFound on missing rows
//
// PR 6a (v0.191.0) ships the full port; the use cases that consume it
// (record order, methodist trigger, delegate-to-teacher per ADR-11)
// land in PR 6b. The order is an immutable artifact — no Update/Delete
// (corrections are made by recording a new order).
type MinobrnaukiOrderRepository interface {
	// Save inserts a new order row and its affected-work-program links
	// (one minobrnauki_order_affected row per id) atomically inside a
	// single transaction. On success the generated id is written back
	// onto the entity via SetID. affectedWorkProgramIDs may be empty —
	// the methodist records the order first and marks affected programs
	// later (ADR-11 pipeline step 2).
	Save(ctx context.Context, order *entities.MinobrnaukiOrder, affectedWorkProgramIDs []int64) error

	// GetByID returns the order with the given id. Returns
	// repositories.ErrMinobrnaukiOrderNotFound when no row matches. The
	// affected-work-program set is fetched separately via FindAffected.
	GetByID(ctx context.Context, id int64) (*entities.MinobrnaukiOrder, error)

	// List returns a page of orders matching the filter together with the
	// total count of matching rows (ignoring Limit / Offset). An empty
	// result is not an error.
	List(ctx context.Context, filter repositories.MinobrnaukiOrderListFilter) (repositories.MinobrnaukiOrderListResult, error)

	// FindAffected returns the work_program ids linked to the given order
	// via minobrnauki_order_affected, in ascending id order. An order
	// with no recorded affected programs yields an empty slice (not an
	// error).
	FindAffected(ctx context.Context, orderID int64) ([]int64, error)
}
