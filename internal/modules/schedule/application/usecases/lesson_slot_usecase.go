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

// List returns all slots ordered by number. STUB — see GREEN commit.
func (uc *LessonSlotUseCase) List(ctx context.Context) ([]*entities.LessonSlot, error) {
	return nil, nil
}

// Create validates and persists a new slot. STUB — see GREEN commit.
func (uc *LessonSlotUseCase) Create(ctx context.Context, number int, timeStart, timeEnd string) (*entities.LessonSlot, error) {
	return nil, nil
}

// Update validates and persists changes to an existing slot. STUB — see GREEN commit.
func (uc *LessonSlotUseCase) Update(ctx context.Context, id int64, number int, timeStart, timeEnd string) (*entities.LessonSlot, error) {
	return nil, nil
}

// Delete removes a slot. STUB — see GREEN commit.
func (uc *LessonSlotUseCase) Delete(ctx context.Context, id int64) error {
	return nil
}
