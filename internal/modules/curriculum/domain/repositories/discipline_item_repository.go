package repositories

import (
	"context"
	"errors"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
)

// ErrDisciplineItemNotFound signals that no DisciplineItem row exists
// for the requested id (or that the row was deleted between load and
// write). Handlers map this sentinel to HTTP 404.
var ErrDisciplineItemNotFound = errors.New("discipline_item: item not found")

// ErrDisciplineItemVersionConflict signals that an Update attempted
// to write against a stale version of the entity — another transaction
// has committed a newer version since this one was loaded. Handlers
// map this sentinel to HTTP 409 Conflict (optimistic locking per ADR-3).
//
// Distinguished from ErrDisciplineItemNotFound at the repository layer
// via a follow-up SELECT after RowsAffected == 0 — the row vanishing
// entirely is a different operational story (deleted, не stale) than
// a version race, and surfaces cleaner UX upstream (reload-and-retry
// vs "this item is gone"). Mirror к Section optimistic-lock behavior.
var ErrDisciplineItemVersionConflict = errors.New("discipline_item: version conflict")

// DisciplineItemRepository is the persistence port for DisciplineItem
// aggregates. Implementations must satisfy the documented sentinel
// contract: ErrDisciplineItemNotFound on missing rows;
// ErrDisciplineItemVersionConflict on stale-version Update attempts.
type DisciplineItemRepository interface {
	// Save inserts a new DisciplineItem and writes the generated id
	// back onto the entity. version starts at 0 в the row.
	Save(ctx context.Context, d *entities.DisciplineItem) error

	// GetByID returns the DisciplineItem with the given id or
	// ErrDisciplineItemNotFound.
	GetByID(ctx context.Context, id int64) (*entities.DisciplineItem, error)

	// ListBySectionID returns every DisciplineItem attached to the
	// given section, ordered by (order_index ASC, created_at ASC, id ASC)
	// для deterministic display. An empty result is not an error.
	ListBySectionID(ctx context.Context, sectionID int64) ([]*entities.DisciplineItem, error)

	// Update writes the (already-mutated) entity back. Implementations
	// MUST enforce optimistic locking: WHERE id = ? AND version = ?.
	// On RowsAffected == 0 the impl distinguishes via a follow-up
	// existence check:
	//   row missing entirely → ErrDisciplineItemNotFound
	//   row exists but version stale → ErrDisciplineItemVersionConflict
	// On success, the entity's version is bumped to reflect the new
	// row state.
	Update(ctx context.Context, d *entities.DisciplineItem) error

	// Delete removes the DisciplineItem row by id. Returns
	// ErrDisciplineItemNotFound if no row was deleted.
	Delete(ctx context.Context, id int64) error
}
