package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
)

type fakeLoadLister struct {
	loads []*entities.TeachingLoad
	err   error
}

func (f *fakeLoadLister) List(_ context.Context, _ TeachingLoadFilter) ([]*entities.TeachingLoad, error) {
	return f.loads, f.err
}

type fakeSlotLister struct {
	slots []*entities.LessonSlot
	err   error
}

func (f *fakeSlotLister) List(_ context.Context) ([]*entities.LessonSlot, error) {
	return f.slots, f.err
}

type fakeRoomLister struct {
	rooms []*entities.Classroom
	err   error
}

func (f *fakeRoomLister) List(_ context.Context, _ ClassroomFilter, _, _ int) ([]*entities.Classroom, error) {
	return f.rooms, f.err
}

func strPtr(s string) *string { return &s }

// hydratedLoad builds a fully-hydrated teaching load for the generator.
func hydratedLoad(id int64, pairs int, week domain.WeekType, groupCap int, lessonShort, lessonName string) *entities.TeachingLoad {
	return &entities.TeachingLoad{
		ID:           id,
		SemesterID:   1,
		GroupID:      10,
		DisciplineID: 20,
		TeacherID:    30,
		LessonTypeID: 40,
		PairsPerWeek: pairs,
		WeekType:     week,
		Group:        &entities.StudentGroup{ID: 10, Name: "ПИ-101", Capacity: groupCap},
		Discipline:   &entities.Discipline{ID: 20, Name: "Матанализ"},
		Teacher:      &entities.TeacherInfo{ID: 30, Name: "Иванов И.И."},
		LessonType:   &entities.LessonType{ID: 40, Name: lessonName, ShortName: lessonShort},
	}
}

func twoSlots() []*entities.LessonSlot {
	return []*entities.LessonSlot{
		{ID: 1, Number: 1, TimeStart: "08:30", TimeEnd: "10:00"},
		{ID: 2, Number: 2, TimeStart: "10:10", TimeEnd: "11:40"},
	}
}

func lectureRoom() []*entities.Classroom {
	return []*entities.Classroom{
		{ID: 100, Building: "A", Number: "101", Capacity: 30, Type: strPtr("lecture"), IsAvailable: true},
	}
}

func TestGenerate_Preview_InvalidSemester(t *testing.T) {
	uc := NewGenerateScheduleUseCase(&fakeLoadLister{}, &fakeSlotLister{}, &fakeRoomLister{})
	_, err := uc.Preview(context.Background(), GenerateParams{SemesterID: 0})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for semester 0, got %v", err)
	}
}

func TestGenerate_Preview_PlacesAllPairs(t *testing.T) {
	uc := NewGenerateScheduleUseCase(
		&fakeLoadLister{loads: []*entities.TeachingLoad{hydratedLoad(1, 2, domain.WeekTypeAll, 25, "Лек", "Лекция")}},
		&fakeSlotLister{slots: twoSlots()},
		&fakeRoomLister{rooms: lectureRoom()},
	)

	preview, err := uc.Preview(context.Background(), GenerateParams{SemesterID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if preview.TotalRequested != 2 || preview.PlacedCount != 2 || preview.UnplacedCount != 0 {
		t.Fatalf("counts = total %d / placed %d / unplaced %d, want 2/2/0",
			preview.TotalRequested, preview.PlacedCount, preview.UnplacedCount)
	}
	if len(preview.Lessons) != 2 {
		t.Fatalf("want 2 placed lessons, got %d", len(preview.Lessons))
	}

	got := preview.Lessons[0]
	if got.GroupName != "ПИ-101" || got.TeacherName != "Иванов И.И." ||
		got.DisciplineName != "Матанализ" || got.LessonTypeName != "Лекция" {
		t.Errorf("names not resolved from hydrated load: %+v", got)
	}
	if got.WeekType != "all" {
		t.Errorf("week type = %q, want all", got.WeekType)
	}
	if got.TimeStart == "" || got.TimeEnd == "" {
		t.Errorf("slot times not resolved: start=%q end=%q", got.TimeStart, got.TimeEnd)
	}
	if got.ClassroomID != 100 || got.ClassroomName != "A-101" {
		t.Errorf("classroom not resolved: id=%d name=%q", got.ClassroomID, got.ClassroomName)
	}
	if got.LoadID != 1 {
		t.Errorf("load id = %d, want 1", got.LoadID)
	}
}

func TestGenerate_Preview_ReportsUnplacedOnOverload(t *testing.T) {
	// 3 pairs of the same group must share the single Monday slot / single room —
	// only one fits; the other two are reported unplaced, not dropped silently.
	uc := NewGenerateScheduleUseCase(
		&fakeLoadLister{loads: []*entities.TeachingLoad{hydratedLoad(1, 3, domain.WeekTypeAll, 25, "Лек", "Лекция")}},
		&fakeSlotLister{slots: []*entities.LessonSlot{{ID: 1, Number: 1, TimeStart: "08:30", TimeEnd: "10:00"}}},
		&fakeRoomLister{rooms: lectureRoom()},
	)

	preview, err := uc.Preview(context.Background(), GenerateParams{SemesterID: 1, Days: []domain.DayOfWeek{domain.Monday}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if preview.PlacedCount != 1 || preview.UnplacedCount != 2 || preview.TotalRequested != 3 {
		t.Fatalf("counts = placed %d / unplaced %d / total %d, want 1/2/3",
			preview.PlacedCount, preview.UnplacedCount, preview.TotalRequested)
	}
	if len(preview.Unplaced) != 2 || preview.Unplaced[0].GroupName != "ПИ-101" {
		t.Errorf("unplaced not reported with names: %+v", preview.Unplaced)
	}
}

func TestGenerate_Preview_RoomTypeRuleBlocksMismatch(t *testing.T) {
	// A lab lesson with only a lecture room available cannot be placed: the
	// room-suitability rule (Лаб -> lab/computer) excludes the lecture room.
	uc := NewGenerateScheduleUseCase(
		&fakeLoadLister{loads: []*entities.TeachingLoad{hydratedLoad(1, 1, domain.WeekTypeAll, 25, "Лаб", "Лабораторная работа")}},
		&fakeSlotLister{slots: twoSlots()},
		&fakeRoomLister{rooms: lectureRoom()},
	)

	preview, err := uc.Preview(context.Background(), GenerateParams{SemesterID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if preview.PlacedCount != 0 || preview.UnplacedCount != 1 {
		t.Errorf("lab in lecture-only building must be unplaced: placed %d / unplaced %d",
			preview.PlacedCount, preview.UnplacedCount)
	}
}

func TestGenerate_Preview_UnhydratedGroupIsUnplaced(t *testing.T) {
	// A load whose group could not be hydrated has an unknown size; placing it
	// would silently disable the room-capacity check, so it must be reported
	// unplaced rather than scheduled blindly into an arbitrary room.
	load := hydratedLoad(1, 1, domain.WeekTypeAll, 25, "Лек", "Лекция")
	load.Group = nil // simulate a broken/missing hydration
	uc := NewGenerateScheduleUseCase(
		&fakeLoadLister{loads: []*entities.TeachingLoad{load}},
		&fakeSlotLister{slots: twoSlots()},
		&fakeRoomLister{rooms: lectureRoom()},
	)

	preview, err := uc.Preview(context.Background(), GenerateParams{SemesterID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if preview.PlacedCount != 0 || preview.UnplacedCount != 1 {
		t.Errorf("load with unknown group size must be unplaced: placed %d / unplaced %d",
			preview.PlacedCount, preview.UnplacedCount)
	}
}

func TestGenerate_Preview_PropagatesRepoError(t *testing.T) {
	uc := NewGenerateScheduleUseCase(
		&fakeLoadLister{err: errors.New("db down")},
		&fakeSlotLister{slots: twoSlots()},
		&fakeRoomLister{rooms: lectureRoom()},
	)
	if _, err := uc.Preview(context.Background(), GenerateParams{SemesterID: 1}); err == nil {
		t.Fatal("expected repo error to propagate")
	}
}
