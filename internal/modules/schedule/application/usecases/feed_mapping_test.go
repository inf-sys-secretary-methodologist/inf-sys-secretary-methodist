package usecases

import (
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/ical"
)

var feedLoc = time.FixedZone("MSK", 3*3600)

func strptr(s string) *string { return &s }

func baseLesson() *entities.Lesson {
	return &entities.Lesson{
		ID:         7,
		DayOfWeek:  domain.Monday,
		TimeStart:  "09:00:00",
		TimeEnd:    "10:40:00",
		WeekType:   domain.WeekTypeAll,
		DateStart:  time.Date(2026, 2, 2, 0, 0, 0, 0, feedLoc), // Monday
		DateEnd:    time.Date(2026, 6, 30, 0, 0, 0, 0, feedLoc),
		UpdatedAt:  time.Date(2026, 1, 15, 8, 0, 0, 0, time.UTC),
		Discipline: &entities.Discipline{Name: "Математический анализ"},
		LessonType: &entities.LessonType{Name: "Лекция", ShortName: "лек"},
		Classroom:  &entities.Classroom{Building: "A", Number: "305"},
		Teacher:    &entities.TeacherInfo{Name: "Иванов И.И."},
	}
}

func TestLessonToICalEvent_BasicWeekly(t *testing.T) {
	ev, ok := lessonToICalEvent(baseLesson(), "methodist", feedLoc)
	if !ok {
		t.Fatal("expected lesson to be included")
	}
	if ev.UID != "lesson-7@methodist" {
		t.Errorf("UID = %q", ev.UID)
	}
	wantStart := time.Date(2026, 2, 2, 9, 0, 0, 0, feedLoc)
	if !ev.Start.Equal(wantStart) {
		t.Errorf("Start = %v, want %v", ev.Start, wantStart)
	}
	wantEnd := time.Date(2026, 2, 2, 10, 40, 0, 0, feedLoc)
	if !ev.End.Equal(wantEnd) {
		t.Errorf("End = %v, want %v", ev.End, wantEnd)
	}
	if ev.Recurrence == nil || ev.Recurrence.Frequency != ical.FreqWeekly {
		t.Fatalf("expected weekly recurrence, got %+v", ev.Recurrence)
	}
	if ev.Recurrence.Interval > 1 {
		t.Errorf("weekType all must be interval 1, got %d", ev.Recurrence.Interval)
	}
	if len(ev.Recurrence.ByDay) != 1 || ev.Recurrence.ByDay[0] != ical.Monday {
		t.Errorf("ByDay = %v", ev.Recurrence.ByDay)
	}
	if ev.Recurrence.Until == nil {
		t.Error("expected UNTIL bound to DateEnd")
	}
	if ev.Summary != "Математический анализ (лек)" {
		t.Errorf("Summary = %q", ev.Summary)
	}
	if ev.Status != ical.StatusConfirmed {
		t.Errorf("Status = %q", ev.Status)
	}
}

func TestLessonToICalEvent_AdvancesToFirstMatchingWeekday(t *testing.T) {
	l := baseLesson()
	l.DateStart = time.Date(2026, 2, 4, 0, 0, 0, 0, feedLoc) // Wednesday
	// DayOfWeek is Monday → first occurrence is Monday 2026-02-09.
	ev, ok := lessonToICalEvent(l, "methodist", feedLoc)
	if !ok {
		t.Fatal("expected included")
	}
	if ev.Start.Weekday() != time.Monday {
		t.Errorf("first occurrence weekday = %v, want Monday", ev.Start.Weekday())
	}
	if ev.Start.Day() != 9 {
		t.Errorf("first occurrence day = %d, want 9", ev.Start.Day())
	}
}

// liveOccurrences expands the weekly rule and removes EXDATEs, returning the
// dates a calendar client would actually show.
func liveOccurrences(ev ical.Event) []time.Time {
	excluded := map[string]bool{}
	for _, e := range ev.ExDates {
		excluded[e.Format("20060102")] = true
	}
	step := max(ev.Recurrence.Interval, 1)
	var out []time.Time
	for d := ev.Start; !d.After(*ev.Recurrence.Until); d = d.AddDate(0, 0, 7*step) {
		if !excluded[d.Format("20060102")] {
			out = append(out, d)
		}
	}
	return out
}

func assertAllParity(t *testing.T, ev ical.Event, wantOdd bool) {
	t.Helper()
	occ := liveOccurrences(ev)
	if len(occ) == 0 {
		t.Fatal("expected at least one occurrence")
	}
	for _, d := range occ {
		_, wk := d.ISOWeek()
		if (wk%2 == 1) != wantOdd {
			t.Errorf("occurrence %s is on ISO week %d, wrong parity", d.Format("2006-01-02"), wk)
		}
	}
}

func TestLessonToICalEvent_OddWeekOnlyOddOccurrences(t *testing.T) {
	l := baseLesson()
	l.WeekType = domain.WeekTypeOdd
	ev, ok := lessonToICalEvent(l, "methodist", feedLoc)
	if !ok {
		t.Fatal("expected included")
	}
	if ev.Start.Weekday() != time.Monday {
		t.Errorf("anchor weekday = %v, want Monday", ev.Start.Weekday())
	}
	if len(ev.ExDates) == 0 {
		t.Error("odd-week lesson must EXDATE the off-parity weeks")
	}
	assertAllParity(t, ev, true)
}

func TestLessonToICalEvent_EvenWeekOnlyEvenOccurrences(t *testing.T) {
	l := baseLesson()
	l.WeekType = domain.WeekTypeEven
	ev, ok := lessonToICalEvent(l, "methodist", feedLoc)
	if !ok {
		t.Fatal("expected included")
	}
	assertAllParity(t, ev, false)
}

func TestLessonToICalEvent_OddWeekParityHoldsAcrossYearBoundary(t *testing.T) {
	// 2026 is a 53-week ISO year: a fixed 14-day step would flip parity after
	// New Year. Spanning Dec 2026 -> Feb 2027 must still yield only odd weeks.
	l := baseLesson()
	l.WeekType = domain.WeekTypeOdd
	l.DateStart = time.Date(2026, 12, 1, 0, 0, 0, 0, feedLoc)
	l.DateEnd = time.Date(2027, 2, 28, 0, 0, 0, 0, feedLoc)
	ev, ok := lessonToICalEvent(l, "methodist", feedLoc)
	if !ok {
		t.Fatal("expected included")
	}
	occ := liveOccurrences(ev)
	sawAfterBoundary := false
	for _, d := range occ {
		if d.Year() == 2027 {
			sawAfterBoundary = true
		}
	}
	if !sawAfterBoundary {
		t.Fatal("expected occurrences after the year boundary")
	}
	assertAllParity(t, ev, true)
}

func TestLessonToICalEvent_SkipsWhenParityAnchorPastEnd(t *testing.T) {
	l := baseLesson()
	l.WeekType = domain.WeekTypeOdd
	// Pick a range whose only Monday is on an even ISO week; the next (odd)
	// Monday falls past DateEnd, so there is no valid occurrence.
	l.DateStart = time.Date(2026, 1, 5, 0, 0, 0, 0, feedLoc) // Mon, ISO week 2 (even)
	l.DateEnd = time.Date(2026, 1, 8, 0, 0, 0, 0, feedLoc)   // Thu same week
	if _, ok := lessonToICalEvent(l, "methodist", feedLoc); ok {
		t.Error("expected skip: no odd-week Monday within range")
	}
}

func TestLessonToICalEvent_CancelledStatus(t *testing.T) {
	l := baseLesson()
	l.IsCancelled = true
	ev, ok := lessonToICalEvent(l, "methodist", feedLoc)
	if !ok {
		t.Fatal("expected included")
	}
	if ev.Status != ical.StatusCancelled {
		t.Errorf("Status = %q, want cancellation", ev.Status)
	}
}

func TestLessonToICalEvent_SkipsWhenFirstOccurrenceAfterEnd(t *testing.T) {
	l := baseLesson()
	l.DateStart = time.Date(2026, 2, 3, 0, 0, 0, 0, feedLoc) // Tuesday
	l.DateEnd = time.Date(2026, 2, 5, 0, 0, 0, 0, feedLoc)   // Thursday same week
	// DayOfWeek Monday → next Monday 2026-02-09 is after DateEnd → no occurrences.
	if _, ok := lessonToICalEvent(l, "methodist", feedLoc); ok {
		t.Error("expected lesson to be skipped (no occurrences within range)")
	}
}

func TestLessonToICalEvent_LocationAndTeacherFallbacks(t *testing.T) {
	l := baseLesson()
	l.Classroom = &entities.Classroom{Building: "B", Number: "12", Name: strptr("Актовый зал")}
	ev, _ := lessonToICalEvent(l, "methodist", feedLoc)
	if ev.Location != "Актовый зал" {
		t.Errorf("named classroom should win: Location = %q", ev.Location)
	}
}

func baseEvent() *entities.Event {
	return &entities.Event{
		ID:        11,
		Title:     "Педсовет",
		EventType: entities.EventTypeMeeting,
		Status:    entities.EventStatusScheduled,
		StartTime: time.Date(2026, 3, 1, 14, 0, 0, 0, feedLoc),
		Timezone:  "Europe/Moscow",
		UpdatedAt: time.Date(2026, 2, 1, 8, 0, 0, 0, time.UTC),
	}
}

func TestEventToICalEvent_Basic(t *testing.T) {
	e := baseEvent()
	end := time.Date(2026, 3, 1, 15, 0, 0, 0, feedLoc)
	e.EndTime = &end
	e.Description = strptr("Обсуждение сессии")
	e.Location = strptr("Ауд. 1")

	ev, ok := eventToICalEvent(e, "methodist", feedLoc)
	if !ok {
		t.Fatal("expected included")
	}
	if ev.UID != "event-11@methodist" {
		t.Errorf("UID = %q", ev.UID)
	}
	if ev.Summary != "Педсовет" {
		t.Errorf("Summary = %q", ev.Summary)
	}
	if !ev.Start.Equal(e.StartTime) || !ev.End.Equal(end) {
		t.Errorf("times mismatch: %v..%v", ev.Start, ev.End)
	}
	if ev.Status != ical.StatusConfirmed {
		t.Errorf("Status = %q", ev.Status)
	}
}

func TestEventToICalEvent_SkipsSoftDeleted(t *testing.T) {
	e := baseEvent()
	del := time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC)
	e.DeletedAt = &del
	if _, ok := eventToICalEvent(e, "methodist", feedLoc); ok {
		t.Error("soft-deleted event must be skipped")
	}
}

func TestEventToICalEvent_CancelledStatus(t *testing.T) {
	e := baseEvent()
	e.Status = entities.EventStatusCancelled
	ev, ok := eventToICalEvent(e, "methodist", feedLoc)
	if !ok {
		t.Fatal("expected included (canceled events carry a cancellation status)")
	}
	if ev.Status != ical.StatusCancelled {
		t.Errorf("Status = %q, want cancellation", ev.Status)
	}
}

func TestEventToICalEvent_AllDay(t *testing.T) {
	e := baseEvent()
	e.AllDay = true
	ev, _ := eventToICalEvent(e, "methodist", feedLoc)
	if !ev.AllDay {
		t.Error("expected AllDay event")
	}
}

func TestEventToICalEvent_RecurrenceMapping(t *testing.T) {
	e := baseEvent()
	until := time.Date(2026, 6, 30, 0, 0, 0, 0, feedLoc)
	e.IsRecurring = true
	e.RecurrenceRule = &entities.RecurrenceRule{
		Frequency: entities.FrequencyWeekly,
		Interval:  2,
		Until:     &until,
		ByWeekday: []entities.Weekday{entities.WeekdayMonday, entities.WeekdayWednesday},
	}
	ev, ok := eventToICalEvent(e, "methodist", feedLoc)
	if !ok {
		t.Fatal("expected included")
	}
	if ev.Recurrence == nil {
		t.Fatal("expected recurrence")
	}
	if ev.Recurrence.Frequency != ical.FreqWeekly || ev.Recurrence.Interval != 2 {
		t.Errorf("recurrence = %+v", ev.Recurrence)
	}
	if len(ev.Recurrence.ByDay) != 2 || ev.Recurrence.ByDay[0] != ical.Monday {
		t.Errorf("ByDay = %v", ev.Recurrence.ByDay)
	}
	if ev.Recurrence.Until == nil {
		t.Error("expected UNTIL")
	}
}
