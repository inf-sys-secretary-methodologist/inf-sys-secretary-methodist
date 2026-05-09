package repositories

import (
	"context"
	"errors"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
)

// ErrSectionNotFound signals that no Section row exists for the
// requested id (or that the row was deleted between load and write).
// Handlers map this sentinel to HTTP 404.
var ErrSectionNotFound = errors.New("section: section not found")

// ErrSectionVersionConflict signals that an Update attempted to write
// against a stale version of the entity — another transaction has
// committed a newer version since this one was loaded. The caller
// should reload the section and merge or retry. Handlers map this
// sentinel to HTTP 409 Conflict (optimistic locking per ADR-3).
//
// Distinguished from ErrSectionNotFound at the repository layer via a
// follow-up SELECT after RowsAffected == 0 — the row vanishing entirely
// is a different operational story (deleted, not stale) than a version
// race, and surfaces cleaner UX upstream (reload-and-retry vs
// "this section is gone").
var ErrSectionVersionConflict = errors.New("section: version conflict")

// SectionRepository is the persistence port for Section aggregates.
// Implementations must satisfy the documented sentinel contract:
// ErrSectionNotFound on missing rows; ErrSectionVersionConflict on
// stale-version Update attempts.
type SectionRepository interface {
	// Save inserts a new Section and writes the generated id back onto
	// the entity. version starts at 0 in the row; ID is set on success.
	Save(ctx context.Context, s *entities.Section) error

	// GetByID returns the Section with the given id or
	// ErrSectionNotFound.
	GetByID(ctx context.Context, id int64) (*entities.Section, error)

	// ListByCurriculumID returns every Section attached to the given
	// curriculum, ordered by (order_index ASC, created_at ASC, id ASC)
	// for deterministic display. An empty result is not an error.
	ListByCurriculumID(ctx context.Context, curriculumID int64) ([]*entities.Section, error)

	// Update writes the (already-mutated) entity back. Implementations
	// MUST enforce optimistic locking: WHERE id = ? AND version = ?.
	// On RowsAffected == 0 the impl distinguishes via a follow-up
	// existence check:
	//   row missing entirely → ErrSectionNotFound
	//   row exists but version stale → ErrSectionVersionConflict
	// On success, the entity's version is bumped to reflect the new
	// row state so callers see a consistent post-update view.
	Update(ctx context.Context, s *entities.Section) error

	// Delete removes the Section row by id. Returns ErrSectionNotFound
	// if no row was deleted. CASCADE in migration 034 handles
	// child-item cleanup automatically (DisciplineItem in v0.128.1+).
	Delete(ctx context.Context, id int64) error
}
