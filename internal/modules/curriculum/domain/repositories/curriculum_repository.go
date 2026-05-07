// Package repositories declares the persistence ports for the
// curriculum module. Concrete implementations live in the
// infrastructure layer.
package repositories

import (
	"context"
	"errors"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
)

// ErrCurriculumNotFound signals that no Curriculum exists for the
// requested id (or that an Update touched zero rows). Use cases map
// this sentinel directly to a domain-level "not found" condition;
// handlers map it to HTTP 404.
var ErrCurriculumNotFound = errors.New("curriculum: curriculum not found")

// ErrCurriculumCodeExists signals that an attempt to write a
// Curriculum row would violate the unique constraint on the code
// column. Surfaces from both Save (insert) and Update (rename).
// Handlers map this sentinel to HTTP 409 so the UI can ask the
// user to pick a different code.
var ErrCurriculumCodeExists = errors.New("curriculum: code already exists")

// CurriculumListFilter narrows a List query. Zero-valued fields are
// treated as "no filter on this dimension". Limit/Offset are honored
// by the repository; a non-positive Limit means "no clamp at the
// repository layer" — use cases are responsible for choosing a
// sensible default to keep result sets bounded.
type CurriculumListFilter struct {
	// Status, when non-nil, restricts results to curricula in that
	// lifecycle state. Useful for the admin "to approve" tab.
	Status *entities.CurriculumStatus
	// Year filters by exact match when non-nil.
	Year *int
	// Specialty filters by exact match when non-empty.
	Specialty string
	// CreatedBy filters by author when non-nil. The use case sets this
	// for "my curricula" views; for unrestricted callers it stays nil.
	CreatedBy *int64
	// Limit caps the number of returned items. Repositories must treat
	// values <= 0 as "no extra clamp" and rely on the caller's policy.
	Limit int
	// Offset is the starting index for pagination.
	Offset int
}

// CurriculumListResult bundles the page of items with the unfiltered
// total so the UI can render pagination controls without a second
// query.
type CurriculumListResult struct {
	Items []*entities.Curriculum
	Total int
}

// CurriculumRepository is the persistence port for Curriculum
// aggregates. Implementations must satisfy the documented sentinel
// contract: ErrCurriculumNotFound on missing rows, ErrCurriculumCodeExists
// on unique-constraint violations against the code column.
type CurriculumRepository interface {
	// GetByID returns the Curriculum with the given id or
	// ErrCurriculumNotFound.
	GetByID(ctx context.Context, id int64) (*entities.Curriculum, error)

	// List returns a page of curricula matching the filter together
	// with the total number of matching rows (ignoring Limit/Offset).
	// Empty result is not an error.
	List(ctx context.Context, filter CurriculumListFilter) (CurriculumListResult, error)

	// Save inserts a new Curriculum and assigns the generated id back
	// onto the entity. Returns ErrCurriculumCodeExists if a row with
	// the same code already exists.
	Save(ctx context.Context, c *entities.Curriculum) error

	// Update writes the (already-mutated) entity back. Returns
	// ErrCurriculumNotFound if the underlying row vanished and
	// ErrCurriculumCodeExists if the rename collides with an
	// existing code.
	Update(ctx context.Context, c *entities.Curriculum) error
}
