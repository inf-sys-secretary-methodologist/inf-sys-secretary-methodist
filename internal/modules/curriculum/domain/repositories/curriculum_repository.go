// Package repositories declares the persistence-side domain values for
// the curriculum module: sentinel errors callers chain with errors.Is
// (ErrCurriculumNotFound / ErrCurriculumCodeExists /
// ErrCurriculumVersionConflict) and query-shape DTOs consumed by
// read-model pipelines (CurriculumListFilter / CurriculumListResult /
// CurriculumYearSpecialtyAgg + their Section / DisciplineItem siblings).
//
// The wide repository ports themselves (CurriculumRepository etc.)
// live в internal/modules/curriculum/application/usecases per CLAUDE.md
// DIP gate. v0.157.1 (ADR-1 carry-forward from #269).
package repositories

import (
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

// ErrCurriculumVersionConflict signals that an Update was attempted
// against a stale version of the entity — another transaction has
// committed a newer version since this one was loaded. Callers should
// reload + reapply (no automatic retry — admin-internal write fitness).
// Handlers map к HTTP 409 (mirror Section pattern, v0.128.0+).
//
// Distinct from ErrCurriculumNotFound which signals "row vanished
// entirely" — different operational story (deleted, not stale) than
// the lost-update race. v0.157.0 #269 ADR-2.
var ErrCurriculumVersionConflict = errors.New("curriculum: version conflict")

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

// CurriculumYearSpecialtyAgg is a read-model row produced by
// AggregateByYearSpecialty: count of curricula grouped by (specialty,
// status) for a given academic year. Consumed by the annual report
// pipeline (internal/modules/reports/annual). Not an entity — DTO only.
type CurriculumYearSpecialtyAgg struct {
	Specialty string
	Status    entities.CurriculumStatus
	Count     int
}

// The CurriculumRepository port itself lives в
// internal/modules/curriculum/application/usecases/repository_interfaces.go
// (DIP — interface lives with consumer).
