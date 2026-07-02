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

// List returns hydrated load lines matching the filter.
func (uc *TeachingLoadUseCase) List(ctx context.Context, filter TeachingLoadFilter) ([]*entities.TeachingLoad, error) {
	return uc.repo.List(ctx, filter)
}

// Create validates the new load via the entity constructor and persists it.
func (uc *TeachingLoadUseCase) Create(ctx context.Context, p TeachingLoadParams) (*entities.TeachingLoad, error) {
	load, err := entities.NewTeachingLoad(p.SemesterID, p.GroupID, p.DisciplineID, p.TeacherID, p.LessonTypeID, p.PairsPerWeek, p.WeekType, uc.now())
	if err != nil {
		return nil, err
	}
	if err := uc.repo.Create(ctx, load); err != nil {
		return nil, err
	}
	return load, nil
}

// Update validates the new values, carries the target id and persists the change.
func (uc *TeachingLoadUseCase) Update(ctx context.Context, id int64, p TeachingLoadParams) (*entities.TeachingLoad, error) {
	load, err := entities.NewTeachingLoad(p.SemesterID, p.GroupID, p.DisciplineID, p.TeacherID, p.LessonTypeID, p.PairsPerWeek, p.WeekType, uc.now())
	if err != nil {
		return nil, err
	}
	load.ID = id
	// The UPDATE only touches updated_at; created_at is owned by Create.
	load.CreatedAt = time.Time{}
	if err := uc.repo.Update(ctx, load); err != nil {
		return nil, err
	}
	return load, nil
}

// Delete removes a load line by id.
func (uc *TeachingLoadUseCase) Delete(ctx context.Context, id int64) error {
	return uc.repo.Delete(ctx, id)
}
