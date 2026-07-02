package usecases

import (
	"context"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
)

// TeachingLoadUseCase manages the planned teaching load: validation lives in
// the TeachingLoad entity, persistence behind TeachingLoadRepository.
type TeachingLoadUseCase struct {
	repo TeachingLoadRepository
	now  func() time.Time
}

// TeachingLoadOption overrides an optional dependency (used in tests).
type TeachingLoadOption func(*TeachingLoadUseCase)

// WithLoadClock overrides the time source.
func WithLoadClock(fn func() time.Time) TeachingLoadOption {
	return func(uc *TeachingLoadUseCase) { uc.now = fn }
}

// NewTeachingLoadUseCase wires the use case with its repository.
func NewTeachingLoadUseCase(repo TeachingLoadRepository, opts ...TeachingLoadOption) *TeachingLoadUseCase {
	uc := &TeachingLoadUseCase{repo: repo, now: time.Now}
	for _, opt := range opts {
		opt(uc)
	}
	return uc
}

// TeachingLoadParams carries the mutable fields for create/update.
type TeachingLoadParams struct {
	SemesterID   int64
	GroupID      int64
	DisciplineID int64
	TeacherID    int64
	LessonTypeID int64
	PairsPerWeek int
	WeekType     domain.WeekType
}

// List returns hydrated load lines matching the filter. STUB — see GREEN commit.
func (uc *TeachingLoadUseCase) List(ctx context.Context, filter TeachingLoadFilter) ([]*entities.TeachingLoad, error) {
	return nil, nil
}

// Create validates and persists a new load line. STUB — see GREEN commit.
func (uc *TeachingLoadUseCase) Create(ctx context.Context, p TeachingLoadParams) (*entities.TeachingLoad, error) {
	return nil, nil
}

// Update validates and persists changes to an existing line. STUB — see GREEN commit.
func (uc *TeachingLoadUseCase) Update(ctx context.Context, id int64, p TeachingLoadParams) (*entities.TeachingLoad, error) {
	return nil, nil
}

// Delete removes a load line. STUB — see GREEN commit.
func (uc *TeachingLoadUseCase) Delete(ctx context.Context, id int64) error {
	return nil
}
