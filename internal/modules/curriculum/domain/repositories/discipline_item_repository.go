package repositories

import "errors"

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

// DisciplineItemHoursAgg is a read-model row produced by
// AggregateHoursByYear: per-curriculum totals of the four hours
// columns. Consumed by the annual report pipeline. DTO only.
type DisciplineItemHoursAgg struct {
	CurriculumID    int64
	CurriculumTitle string
	Lectures        int
	Practice        int
	Lab             int
	SelfStudy       int
}

// The DisciplineItemRepository port itself lives в
// internal/modules/curriculum/application/usecases/repository_interfaces.go
// (DIP — interface lives with consumer). v0.157.1.
