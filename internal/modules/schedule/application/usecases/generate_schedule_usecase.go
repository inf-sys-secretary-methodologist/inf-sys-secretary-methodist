package usecases

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/solver"
)

// The generator consumes only these narrow read surfaces (DIP): the concrete
// PG repositories already satisfy them.
type (
	generateLoadLister interface {
		List(ctx context.Context, filter TeachingLoadFilter) ([]*entities.TeachingLoad, error)
	}
	generateSlotLister interface {
		List(ctx context.Context) ([]*entities.LessonSlot, error)
	}
	generateRoomLister interface {
		List(ctx context.Context, filter ClassroomFilter, limit, offset int) ([]*entities.Classroom, error)
	}
)

// GenerateScheduleUseCase turns the planned teaching load into a draft timetable
// by unfolding it into CSP solver variables and running the (pure) solver. It
// never persists on its own — Preview computes a draft; applying it is a
// separate, explicit step.
type GenerateScheduleUseCase struct {
	loads   generateLoadLister
	slots   generateSlotLister
	rooms   generateRoomLister
	weights solver.SoftWeights
}

// NewGenerateScheduleUseCase wires the use case with its read dependencies and
// the default soft-preference weights.
func NewGenerateScheduleUseCase(loads generateLoadLister, slots generateSlotLister, rooms generateRoomLister) *GenerateScheduleUseCase {
	return &GenerateScheduleUseCase{
		loads:   loads,
		slots:   slots,
		rooms:   rooms,
		weights: solver.NewDefaultWeights(),
	}
}

// GenerateParams is the request for a draft schedule.
type GenerateParams struct {
	SemesterID int64
	Days       []domain.DayOfWeek // optional; defaults to Mon-Sat
}

// GeneratedLesson is one placed lesson in a draft, with names resolved for display.
type GeneratedLesson struct {
	LoadID         int64
	GroupID        int64
	GroupName      string
	TeacherID      int64
	TeacherName    string
	DisciplineID   int64
	DisciplineName string
	LessonTypeID   int64
	LessonTypeName string
	WeekType       string
	DayOfWeek      int
	SlotNumber     int
	TimeStart      string
	TimeEnd        string
	ClassroomID    int64
	ClassroomName  string
}

// UnplacedLesson is a load line the solver could not place (best-effort).
type UnplacedLesson struct {
	LoadID         int64
	GroupName      string
	DisciplineName string
	LessonTypeName string
	WeekType       string
}

// SchedulePreview is the draft returned by Preview: placed lessons plus anything
// left unplaced, with summary counts.
type SchedulePreview struct {
	Lessons        []GeneratedLesson
	Unplaced       []UnplacedLesson
	TotalRequested int
	PlacedCount    int
	UnplacedCount  int
}

// Preview assembles the solver input for the semester, runs the solver, and
// returns the resulting draft without persisting anything.
func (uc *GenerateScheduleUseCase) Preview(ctx context.Context, params GenerateParams) (*SchedulePreview, error) {
	return &SchedulePreview{}, nil
}
