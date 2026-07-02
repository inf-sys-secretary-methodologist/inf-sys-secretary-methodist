package usecases

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
)

// TeachingLoadFilter narrows a teaching-load listing. All fields are optional.
type TeachingLoadFilter struct {
	SemesterID *int64
	GroupID    *int64
	TeacherID  *int64
}

// TeachingLoadRepository persists planned teaching assignments. The interface
// lives in the consumer package (DIP); the PG implementation is in
// infrastructure/persistence.
type TeachingLoadRepository interface {
	// Create inserts a new load line. A duplicate (semester, group, discipline,
	// lesson_type) returns entities.ErrTeachingLoadDuplicate.
	Create(ctx context.Context, load *entities.TeachingLoad) error
	// Update mutates an existing line. Missing row -> entities.ErrTeachingLoadNotFound;
	// duplicate key -> entities.ErrTeachingLoadDuplicate.
	Update(ctx context.Context, load *entities.TeachingLoad) error
	// Delete removes a line by id. Missing row -> entities.ErrTeachingLoadNotFound.
	Delete(ctx context.Context, id int64) error
	// GetByID returns one hydrated line or entities.ErrTeachingLoadNotFound.
	GetByID(ctx context.Context, id int64) (*entities.TeachingLoad, error)
	// List returns hydrated load lines matching the filter, ordered by group then discipline.
	List(ctx context.Context, filter TeachingLoadFilter) ([]*entities.TeachingLoad, error)
}
