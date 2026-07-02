package usecases

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
)

// LessonSlotRepository persists the institution-wide bell-schedule catalog.
// The interface lives in the consumer package (DIP); the PG implementation
// is in infrastructure/persistence.
type LessonSlotRepository interface {
	// Create inserts a new slot. A duplicate Number returns entities.ErrLessonSlotNumberTaken.
	Create(ctx context.Context, slot *entities.LessonSlot) error
	// Update mutates an existing slot. A missing row returns entities.ErrLessonSlotNotFound;
	// a duplicate Number returns entities.ErrLessonSlotNumberTaken.
	Update(ctx context.Context, slot *entities.LessonSlot) error
	// Delete removes a slot by id. A missing row returns entities.ErrLessonSlotNotFound.
	Delete(ctx context.Context, id int64) error
	// GetByID returns one slot or entities.ErrLessonSlotNotFound.
	GetByID(ctx context.Context, id int64) (*entities.LessonSlot, error)
	// List returns all slots ordered by Number.
	List(ctx context.Context) ([]*entities.LessonSlot, error)
}
