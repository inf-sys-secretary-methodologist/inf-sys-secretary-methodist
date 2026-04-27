package repositories

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
)

type LessonFilter struct {
	SemesterID   *int64
	GroupID      *int64
	TeacherID    *int64
	ClassroomID  *int64
	DisciplineID *int64
	DayOfWeek    *domain.DayOfWeek
	WeekType     *domain.WeekType
}

type LessonRepository interface {
	Create(ctx context.Context, lesson *entities.Lesson) error
	Save(ctx context.Context, lesson *entities.Lesson) error
	GetByID(ctx context.Context, id int64) (*entities.Lesson, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, filter LessonFilter, limit, offset int) ([]*entities.Lesson, error)
	Count(ctx context.Context, filter LessonFilter) (int64, error)
	GetTimetable(ctx context.Context, filter LessonFilter) ([]*entities.Lesson, error)
}
