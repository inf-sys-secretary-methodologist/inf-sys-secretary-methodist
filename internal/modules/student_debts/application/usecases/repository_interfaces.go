// Wide repository ports for the student_debts module. Defined in the
// application layer per CLAUDE.md DIP gate ("Repository interfaces —
// в пакете-потребителе (usecase/), НЕ в domain/. DIP по классике.").
//
// Sentinels (ErrStudentDebtNotFound etc.) live in
// internal/modules/student_debts/domain/repositories — they encode
// domain values rather than persistence contracts, so handlers chain
// errors.Is against them without depending on the persistence port.
//
// Concrete implementations satisfy these ports via Go's structural
// typing, backed by a compile-time
// `var _ usecases.StudentDebtRepository = (*StudentDebtRepositoryPG)(nil)`
// assertion in the infrastructure layer so signature drift surfaces at
// the implementation's compile site rather than only at DI wiring.
package usecases

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/repositories"
)

// StudentDebtRepository is the persistence port for StudentDebt
// aggregates. Implementations MUST persist the root + its resit attempts
// atomically — either the full aggregate lands or none of it does.
//
// Sentinel contract:
//   - repositories.ErrStudentDebtNotFound on missing rows
//   - repositories.ErrStudentDebtIdentityExists on natural-key uniqueness
//     violation (group_name, student_full_name, discipline_name, semester)
//   - repositories.ErrStudentDebtVersionConflict on stale-version Update
type StudentDebtRepository interface {
	// Save inserts a new StudentDebt aggregate (root row + every resit
	// attempt) atomically. On success the generated ids are written back
	// onto the root and its attempts. Returns
	// repositories.ErrStudentDebtIdentityExists when a row with the same
	// natural key already exists.
	Save(ctx context.Context, debt *entities.StudentDebt) error

	// GetByID returns the StudentDebt aggregate with the given id,
	// hydrated through Reconstitute* — root + its attempts in attempt-no
	// order. Returns repositories.ErrStudentDebtNotFound when no matching
	// row exists.
	GetByID(ctx context.Context, id int64) (*entities.StudentDebt, error)

	// FindByIdentity returns the aggregate matching the natural key
	// (group_name, student_full_name, discipline_name, semester) — the
	// importer's insert-vs-update probe for a source row carrying no
	// service id. Returns repositories.ErrStudentDebtNotFound when absent.
	FindByIdentity(ctx context.Context, groupName, studentFullName, disciplineName string, semester int) (*entities.StudentDebt, error)

	// List returns a page of StudentDebt items matching the filter
	// together with the total number of matching rows (ignoring Limit /
	// Offset). Items carry root state only — attempts are not hydrated.
	// An empty result is not an error.
	List(ctx context.Context, filter repositories.StudentDebtListFilter) (repositories.StudentDebtListResult, error)

	// Update writes the (already-mutated) aggregate back atomically:
	// UPDATE root with optimistic-lock guard (WHERE id=? AND version=?)
	// then delete + reinsert every attempt inside the same tx. On
	// RowsAffected == 0 the impl distinguishes:
	//   - row missing entirely → ErrStudentDebtNotFound
	//   - row exists but stale version → ErrStudentDebtVersionConflict
	// On success the entity's version is bumped to reflect the new row
	// state so callers see a consistent post-update view without
	// reloading.
	Update(ctx context.Context, debt *entities.StudentDebt) error
}
