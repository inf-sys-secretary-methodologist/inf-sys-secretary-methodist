package repositories

import "errors"

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

// The SectionRepository port itself lives в
// internal/modules/curriculum/application/usecases/repository_interfaces.go
// (DIP — interface lives with consumer). v0.157.1.
