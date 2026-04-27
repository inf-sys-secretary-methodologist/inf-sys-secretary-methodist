package repositories

import (
	"context"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
)

type ScheduleChangeRepository interface {
	Create(ctx context.Context, change *entities.ScheduleChange) error
	GetByLessonID(ctx context.Context, lessonID int64) ([]*entities.ScheduleChange, error)
	GetByDateRange(ctx context.Context, start, end time.Time) ([]*entities.ScheduleChange, error)
}
