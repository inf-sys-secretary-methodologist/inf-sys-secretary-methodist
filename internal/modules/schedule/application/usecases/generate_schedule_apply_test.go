package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
)

type fakeLessonWriter struct {
	count         int64
	countErr      error
	created       []*entities.Lesson
	createManyErr error
}

func (f *fakeLessonWriter) Count(_ context.Context, _ LessonFilter) (int64, error) {
	return f.count, f.countErr
}

func (f *fakeLessonWriter) CreateMany(_ context.Context, lessons []*entities.Lesson) error {
	if f.createManyErr != nil {
		return f.createManyErr // atomic: a failed batch persists nothing
	}
	f.created = append(f.created, lessons...)
	return nil
}

type fakeSemesterLister struct {
	semesters []*entities.Semester
	err       error
}

func (f *fakeSemesterLister) ListSemesters(_ context.Context, _ bool) ([]*entities.Semester, error) {
	return f.semesters, f.err
}

func semesterOne() []*entities.Semester {
	return []*entities.Semester{{
		ID:        1,
		StartDate: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC),
	}}
}

func applyFixedClock() time.Time { return time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC) }

func newApplyUC(loads *fakeLoadLister, slots *fakeSlotLister, rooms *fakeRoomLister, writer *fakeLessonWriter, sems *fakeSemesterLister) *GenerateScheduleUseCase {
	return NewGenerateScheduleUseCase(loads, slots, rooms,
		WithApplyWriter(writer), WithSemesters(sems), WithGenerateClock(applyFixedClock))
}

func TestGenerate_Apply_PersistsPlacedLessons(t *testing.T) {
	writer := &fakeLessonWriter{}
	uc := newApplyUC(
		&fakeLoadLister{loads: []*entities.TeachingLoad{hydratedLoad(1, 2, domain.WeekTypeAll, 25, "Лек", "Лекция")}},
		&fakeSlotLister{slots: twoSlots()},
		&fakeRoomLister{rooms: lectureRoom()},
		writer,
		&fakeSemesterLister{semesters: semesterOne()},
	)

	res, err := uc.Apply(context.Background(), GenerateParams{SemesterID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Created != 2 || res.Unplaced != 0 {
		t.Fatalf("result = created %d / unplaced %d, want 2/0", res.Created, res.Unplaced)
	}
	if len(writer.created) != 2 {
		t.Fatalf("want 2 persisted lessons, got %d", len(writer.created))
	}

	l := writer.created[0]
	if l.SemesterID != 1 || l.DisciplineID != 20 || l.LessonTypeID != 40 ||
		l.TeacherID != 30 || l.GroupID != 10 || l.ClassroomID != 100 {
		t.Errorf("lesson references wrong: %+v", l)
	}
	if l.TimeStart == "" || l.TimeEnd == "" {
		t.Errorf("lesson times not filled from slot: %+v", l)
	}
	if !l.DateStart.Equal(semesterOne()[0].StartDate) || !l.DateEnd.Equal(semesterOne()[0].EndDate) {
		t.Errorf("lesson dates must come from semester: start=%v end=%v", l.DateStart, l.DateEnd)
	}
	if string(l.WeekType) != "all" {
		t.Errorf("week type = %q, want all", l.WeekType)
	}
}

func TestGenerate_Apply_RefusesWhenScheduleExists(t *testing.T) {
	writer := &fakeLessonWriter{count: 5}
	uc := newApplyUC(
		&fakeLoadLister{loads: []*entities.TeachingLoad{hydratedLoad(1, 2, domain.WeekTypeAll, 25, "Лек", "Лекция")}},
		&fakeSlotLister{slots: twoSlots()},
		&fakeRoomLister{rooms: lectureRoom()},
		writer,
		&fakeSemesterLister{semesters: semesterOne()},
	)

	_, err := uc.Apply(context.Background(), GenerateParams{SemesterID: 1})
	if !errors.Is(err, ErrScheduleAlreadyExists) {
		t.Fatalf("expected ErrScheduleAlreadyExists, got %v", err)
	}
	if len(writer.created) != 0 {
		t.Errorf("nothing must be created when a schedule already exists")
	}
}

func TestGenerate_Apply_SemesterNotFound(t *testing.T) {
	uc := newApplyUC(
		&fakeLoadLister{loads: []*entities.TeachingLoad{hydratedLoad(1, 1, domain.WeekTypeAll, 25, "Лек", "Лекция")}},
		&fakeSlotLister{slots: twoSlots()},
		&fakeRoomLister{rooms: lectureRoom()},
		&fakeLessonWriter{},
		&fakeSemesterLister{semesters: nil}, // no semesters
	)

	_, err := uc.Apply(context.Background(), GenerateParams{SemesterID: 1})
	if !errors.Is(err, ErrSemesterNotFound) {
		t.Fatalf("expected ErrSemesterNotFound, got %v", err)
	}
}

func TestGenerate_Apply_NotWritableWithoutWriteDeps(t *testing.T) {
	uc := NewGenerateScheduleUseCase(
		&fakeLoadLister{loads: []*entities.TeachingLoad{hydratedLoad(1, 1, domain.WeekTypeAll, 25, "Лек", "Лекция")}},
		&fakeSlotLister{slots: twoSlots()},
		&fakeRoomLister{rooms: lectureRoom()},
	)
	_, err := uc.Apply(context.Background(), GenerateParams{SemesterID: 1})
	if !errors.Is(err, ErrGenerateNotWritable) {
		t.Fatalf("expected ErrGenerateNotWritable, got %v", err)
	}
}

func TestGenerate_Apply_SkipsUnplaced(t *testing.T) {
	writer := &fakeLessonWriter{}
	uc := newApplyUC(
		&fakeLoadLister{loads: []*entities.TeachingLoad{hydratedLoad(1, 3, domain.WeekTypeAll, 25, "Лек", "Лекция")}},
		&fakeSlotLister{slots: []*entities.LessonSlot{{ID: 1, Number: 1, TimeStart: "08:30", TimeEnd: "10:00"}}},
		&fakeRoomLister{rooms: lectureRoom()},
		writer,
		&fakeSemesterLister{semesters: semesterOne()},
	)

	res, err := uc.Apply(context.Background(), GenerateParams{SemesterID: 1, Days: []domain.DayOfWeek{domain.Monday}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Created != 1 || res.Unplaced != 2 {
		t.Errorf("result = created %d / unplaced %d, want 1/2", res.Created, res.Unplaced)
	}
	if len(writer.created) != 1 {
		t.Errorf("only placed lessons persisted, got %d", len(writer.created))
	}
}

func TestGenerate_Apply_PropagatesCreateErrorAtomically(t *testing.T) {
	writer := &fakeLessonWriter{createManyErr: errors.New("insert failed")}
	uc := newApplyUC(
		&fakeLoadLister{loads: []*entities.TeachingLoad{hydratedLoad(1, 2, domain.WeekTypeAll, 25, "Лек", "Лекция")}},
		&fakeSlotLister{slots: twoSlots()},
		&fakeRoomLister{rooms: lectureRoom()},
		writer,
		&fakeSemesterLister{semesters: semesterOne()},
	)

	if _, err := uc.Apply(context.Background(), GenerateParams{SemesterID: 1}); err == nil {
		t.Fatal("expected create error to propagate")
	}
	if len(writer.created) != 0 {
		t.Errorf("a failed batch must persist nothing, got %d lessons", len(writer.created))
	}
}

func TestGenerate_Apply_RejectsInvalidDay(t *testing.T) {
	writer := &fakeLessonWriter{}
	uc := newApplyUC(
		&fakeLoadLister{loads: []*entities.TeachingLoad{hydratedLoad(1, 1, domain.WeekTypeAll, 25, "Лек", "Лекция")}},
		&fakeSlotLister{slots: twoSlots()},
		&fakeRoomLister{rooms: lectureRoom()},
		writer,
		&fakeSemesterLister{semesters: semesterOne()},
	)

	_, err := uc.Apply(context.Background(), GenerateParams{SemesterID: 1, Days: []domain.DayOfWeek{domain.DayOfWeek(99)}})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("invalid day must map to ErrInvalidInput, got %v", err)
	}
	if len(writer.created) != 0 {
		t.Errorf("nothing must be written when input is invalid")
	}
}
