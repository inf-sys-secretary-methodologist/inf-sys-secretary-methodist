package usecases

import (
	"context"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
)

// LessonSlotUseCase manages the bell-schedule catalog: validation lives in the
// LessonSlot entity, persistence behind LessonSlotRepository.
type LessonSlotUseCase struct {
	repo LessonSlotRepository
	now  func() time.Time
}

// LessonSlotOption overrides an optional dependency (used in tests).
type LessonSlotOption func(*LessonSlotUseCase)

// WithSlotClock overrides the time source.
func WithSlotClock(fn func() time.Time) LessonSlotOption {
	return func(uc *LessonSlotUseCase) { uc.now = fn }
}

// NewLessonSlotUseCase wires the use case with its repository.
func NewLessonSlotUseCase(repo LessonSlotRepository, opts ...LessonSlotOption) *LessonSlotUseCase {
	uc := &LessonSlotUseCase{repo: repo, now: time.Now}
	for _, opt := range opts {
		opt(uc)
	}
	return uc
}

// List returns all slots ordered by number.
func (uc *LessonSlotUseCase) List(ctx context.Context) ([]*entities.LessonSlot, error) {
	return uc.repo.List(ctx)
}

// Create validates a new slot via the entity constructor and persists it.
func (uc *LessonSlotUseCase) Create(ctx context.Context, number int, timeStart, timeEnd string) (*entities.LessonSlot, error) {
	slot, err := entities.NewLessonSlot(number, timeStart, timeEnd, uc.now())
	if err != nil {
		return nil, err
	}
	if err := uc.repo.Create(ctx, slot); err != nil {
		return nil, err
	}
	return slot, nil
}

// Update validates the new values via the entity constructor, carries the
// target id and persists the change.
func (uc *LessonSlotUseCase) Update(ctx context.Context, id int64, number int, timeStart, timeEnd string) (*entities.LessonSlot, error) {
	slot, err := entities.NewLessonSlot(number, timeStart, timeEnd, uc.now())
	if err != nil {
		return nil, err
	}
	slot.ID = id
	if err := uc.repo.Update(ctx, slot); err != nil {
		return nil, err
	}
	return slot, nil
}

// Delete removes a slot by id.
func (uc *LessonSlotUseCase) Delete(ctx context.Context, id int64) error {
	return uc.repo.Delete(ctx, id)
}
