// Wide repository ports for the curriculum module. Defined in the
// application layer per CLAUDE.md DIP gate ("Repository interfaces —
// в пакете-потребителе (usecase/), НЕ в domain/. DIP по классике.").
//
// Sentinels (ErrCurriculumNotFound etc.) and query-shape DTOs
// (CurriculumListFilter / CurriculumListResult / *Agg) remain in
// internal/modules/curriculum/domain/repositories — they encode domain
// values rather than persistence contracts, so handlers chain
// errors.Is against them and read-model consumers reference the DTOs
// без depending on the persistence port itself.
//
// Concrete implementations satisfy these ports via Go's structural
// typing, backed by compile-time `var _ usecases.<Port> = (*XxxPG)(nil)`
// assertions in the infrastructure layer (`*_pg.go` files) so signature
// drift surfaces at the implementation's compile site rather than only
// at DI wiring. The infra layer also references these ports explicitly
// in the `Begin` return type of `BulkDisciplineItemsUnitOfWorkPG` and
// the `Items` / `Sections` / `Curricula` accessor return types of
// `bulkTxPG`.
//
// Relocated from internal/modules/curriculum/domain/repositories/
// in v0.157.1 (ADR-1 carry-forward from #269).
package usecases

import (
	"context"
	"database/sql"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
)

// CurriculumRepository is the persistence port for Curriculum
// aggregates. Implementations must satisfy the documented sentinel
// contract: repositories.ErrCurriculumNotFound on missing rows,
// repositories.ErrCurriculumCodeExists on unique-constraint violations
// against the code column.
type CurriculumRepository interface {
	// GetByID returns the Curriculum with the given id or
	// repositories.ErrCurriculumNotFound.
	GetByID(ctx context.Context, id int64) (*entities.Curriculum, error)

	// List returns a page of curricula matching the filter together
	// with the total number of matching rows (ignoring Limit/Offset).
	// Empty result is not an error.
	List(ctx context.Context, filter repositories.CurriculumListFilter) (repositories.CurriculumListResult, error)

	// Save inserts a new Curriculum and assigns the generated id back
	// onto the entity. Returns repositories.ErrCurriculumCodeExists if
	// a row with the same code already exists.
	Save(ctx context.Context, c *entities.Curriculum) error

	// Update writes the (already-mutated) entity back. Returns
	// repositories.ErrCurriculumNotFound if the underlying row vanished
	// and repositories.ErrCurriculumCodeExists if the rename collides
	// with an existing code.
	Update(ctx context.Context, c *entities.Curriculum) error

	// AggregateByYearSpecialty returns one row per (specialty, status)
	// combination for curricula with the given year, counting matching
	// rows. Empty result is not an error. Used by the annual report
	// pipeline to render the curricula summary section.
	AggregateByYearSpecialty(ctx context.Context, year int) ([]repositories.CurriculumYearSpecialtyAgg, error)
}

// SectionRepository is the persistence port for Section aggregates.
// Implementations must satisfy the documented sentinel contract:
// repositories.ErrSectionNotFound on missing rows;
// repositories.ErrSectionVersionConflict on stale-version Update
// attempts.
type SectionRepository interface {
	// Save inserts a new Section and writes the generated id back onto
	// the entity. version starts at 0 in the row; ID is set on success.
	Save(ctx context.Context, s *entities.Section) error

	// GetByID returns the Section with the given id or
	// repositories.ErrSectionNotFound.
	GetByID(ctx context.Context, id int64) (*entities.Section, error)

	// ListByCurriculumID returns every Section attached to the given
	// curriculum, ordered by (order_index ASC, created_at ASC, id ASC)
	// for deterministic display. An empty result is not an error.
	ListByCurriculumID(ctx context.Context, curriculumID int64) ([]*entities.Section, error)

	// Update writes the (already-mutated) entity back. Implementations
	// MUST enforce optimistic locking: WHERE id = ? AND version = ?.
	// On RowsAffected == 0 the impl distinguishes via a follow-up
	// existence check:
	//   row missing entirely → repositories.ErrSectionNotFound
	//   row exists but version stale → repositories.ErrSectionVersionConflict
	// On success, the entity's version is bumped to reflect the new
	// row state so callers see a consistent post-update view.
	Update(ctx context.Context, s *entities.Section) error

	// Delete removes the Section row by id. Returns
	// repositories.ErrSectionNotFound if no row was deleted. CASCADE
	// in migration 034 handles child-item cleanup automatically
	// (DisciplineItem in v0.128.1+).
	Delete(ctx context.Context, id int64) error
}

// DisciplineItemRepository is the persistence port for DisciplineItem
// aggregates. Implementations must satisfy the documented sentinel
// contract: repositories.ErrDisciplineItemNotFound on missing rows;
// repositories.ErrDisciplineItemVersionConflict on stale-version
// Update attempts.
type DisciplineItemRepository interface {
	// Save inserts a new DisciplineItem and writes the generated id
	// back onto the entity. version starts at 0 в the row.
	Save(ctx context.Context, d *entities.DisciplineItem) error

	// GetByID returns the DisciplineItem with the given id or
	// repositories.ErrDisciplineItemNotFound.
	GetByID(ctx context.Context, id int64) (*entities.DisciplineItem, error)

	// ListBySectionID returns every DisciplineItem attached to the
	// given section, ordered by (order_index ASC, created_at ASC, id ASC)
	// для deterministic display. An empty result is not an error.
	ListBySectionID(ctx context.Context, sectionID int64) ([]*entities.DisciplineItem, error)

	// Update writes the (already-mutated) entity back. Implementations
	// MUST enforce optimistic locking: WHERE id = ? AND version = ?.
	// On RowsAffected == 0 the impl distinguishes via a follow-up
	// existence check:
	//   row missing entirely → repositories.ErrDisciplineItemNotFound
	//   row exists but version stale → repositories.ErrDisciplineItemVersionConflict
	// On success, the entity's version is bumped to reflect the new
	// row state.
	Update(ctx context.Context, d *entities.DisciplineItem) error

	// Delete removes the DisciplineItem row by id. Returns
	// repositories.ErrDisciplineItemNotFound if no row was deleted.
	Delete(ctx context.Context, id int64) error

	// AggregateHoursByYear sums hours (lectures / practice / lab / self)
	// across all discipline items belonging to curricula with
	// curricula.year = year, grouped per curriculum. Empty result is
	// not an error. Used by the annual report pipeline.
	AggregateHoursByYear(ctx context.Context, year int) ([]repositories.DisciplineItemHoursAgg, error)
}

// BulkDisciplineItemsUnitOfWork opens a transactional boundary spanning
// DisciplineItem + Section + Curriculum repositories. Used by the
// bulk-edit endpoint (v0.128.3 ADR-10) to apply creates+updates+deletes
// atomically — все succeed (Commit) или ничего не applies (Rollback).
//
// The pattern intentionally excludes leaking `*sql.Tx` или any infra
// primitive — caller получает only the narrow BulkDisciplineItemsTx
// interface, which composes tx-bound repository ports for use by the
// usecase. Mirror к Vaughn Vernon's "Aggregate transactional consistency
// boundary" pattern, but spans 3 ARs because bulk-edit-РПД is a single
// editorial action в the academic secretary's mental model (per
// v0.158.0+ the author of curricula and their РПД is the secretary).
type BulkDisciplineItemsUnitOfWork interface {
	// Begin opens a fresh transaction. opts may pass an isolation level
	// (Repeatable Read recommended per ADR-12 — phantom-prevention для
	// long bulk operations); pass nil for PostgreSQL default (Read
	// Committed). Each call returns an independent BulkDisciplineItemsTx
	// — UoW is concurrency-safe (state lives on the returned Tx, not
	// the UoW itself).
	Begin(ctx context.Context, opts *sql.TxOptions) (BulkDisciplineItemsTx, error)
}

// BulkDisciplineItemsTx is the tx-bound view exposing repository ports
// backed by the same `*sql.Tx`. Caller MUST eventually call Commit() or
// Rollback() — never both, never neither (recommend `defer tx.Rollback()`
// pattern; second call returns repositories.ErrBulkTxFinished
// idempotent).
type BulkDisciplineItemsTx interface {
	// Items returns DisciplineItem repository operations bound to the tx.
	Items() DisciplineItemRepository

	// Sections returns Section repository operations bound to the tx.
	Sections() SectionRepository

	// Curricula returns Curriculum repository operations bound to the tx.
	Curricula() CurriculumRepository

	// Commit finalizes the transaction. Subsequent calls return
	// repositories.ErrBulkTxFinished.
	Commit() error

	// Rollback aborts the transaction. Subsequent calls return
	// repositories.ErrBulkTxFinished. Safe to defer immediately after
	// Begin — Rollback after Commit is a no-op (returns
	// repositories.ErrBulkTxFinished which caller may ignore).
	Rollback() error
}
