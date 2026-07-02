package ical

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

// msk is the fixed Moscow offset used to build deterministic wall-clock times.
var msk = time.FixedZone("MSK", 3*3600)

func loadGolden(t *testing.T, name string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read golden %s: %v", name, err)
	}
	// Golden files are stored with LF for readability; RFC 5545 mandates CRLF.
	return strings.ReplaceAll(string(raw), "\n", "\r\n")
}

const prodID = "-//Secretary Methodist//Calendar Feed//EN"

func TestRender_Basic(t *testing.T) {
	cal := Calendar{
		ProdID: prodID,
		Name:   "Расписание",
		TZID:   "Europe/Moscow",
		Events: []Event{{
			UID:     "lesson-1@methodist",
			Summary: "Матанализ",
			Start:   time.Date(2026, 2, 10, 9, 0, 0, 0, msk),
			End:     time.Date(2026, 2, 10, 10, 40, 0, 0, msk),
			Stamp:   time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
			Status:  StatusConfirmed,
		}},
	}

	got := cal.Render()
	want := loadGolden(t, "basic.ics")
	if got != want {
		t.Errorf("Render() mismatch\n--- got ---\n%q\n--- want ---\n%q", got, want)
	}
}

func TestRender_RecurringWithExdate(t *testing.T) {
	until := time.Date(2026, 6, 30, 20, 59, 59, 0, time.UTC)
	cal := Calendar{
		ProdID: prodID,
		Name:   "Расписание",
		TZID:   "Europe/Moscow",
		Events: []Event{{
			UID:      "lesson-42@methodist",
			Summary:  "Физика (лекция)",
			Location: "Ауд. 305",
			Start:    time.Date(2026, 2, 10, 9, 0, 0, 0, msk),
			End:      time.Date(2026, 2, 10, 10, 40, 0, 0, msk),
			Stamp:    time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
			Recurrence: &Recurrence{
				Frequency: FreqWeekly,
				Interval:  2,
				Until:     &until,
				ByDay:     []Weekday{Tuesday},
			},
			ExDates: []time.Time{time.Date(2026, 3, 10, 9, 0, 0, 0, msk)},
			Status:  StatusConfirmed,
		}},
	}

	got := cal.Render()
	want := loadGolden(t, "recurring.ics")
	if got != want {
		t.Errorf("Render() mismatch\n--- got ---\n%q\n--- want ---\n%q", got, want)
	}
}

func TestRender_AllDay(t *testing.T) {
	cal := Calendar{
		ProdID: prodID,
		Name:   "Календарь",
		TZID:   "Europe/Moscow",
		Events: []Event{{
			UID:     "event-7@methodist",
			Summary: "День защитника Отечества",
			Start:   time.Date(2026, 2, 23, 0, 0, 0, 0, msk),
			End:     time.Date(2026, 2, 24, 0, 0, 0, 0, msk),
			AllDay:  true,
			Stamp:   time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
			Status:  StatusConfirmed,
		}},
	}

	got := cal.Render()
	want := loadGolden(t, "allday.ics")
	if got != want {
		t.Errorf("Render() mismatch\n--- got ---\n%q\n--- want ---\n%q", got, want)
	}
}

func TestRender_EscapesTextValues(t *testing.T) {
	cal := Calendar{
		ProdID: prodID,
		TZID:   "Europe/Moscow",
		Events: []Event{{
			UID:         "e1",
			Summary:     "Лекция; тема: A, B \\ C",
			Description: "Строка 1\nСтрока 2, важно",
			Start:       time.Date(2026, 2, 10, 9, 0, 0, 0, msk),
			End:         time.Date(2026, 2, 10, 10, 0, 0, 0, msk),
			Stamp:       time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
			Status:      StatusConfirmed,
		}},
	}

	got := cal.Render()
	if !strings.Contains(got, `SUMMARY:Лекция\; тема: A\, B \\ C`) {
		t.Errorf("summary not escaped per RFC 5545; got:\n%s", got)
	}
	if !strings.Contains(got, `DESCRIPTION:Строка 1\nСтрока 2\, важно`) {
		t.Errorf("description not escaped per RFC 5545; got:\n%s", got)
	}
}

func TestRender_FoldsLongLinesAtOctetBoundary(t *testing.T) {
	longDesc := strings.Repeat("длинное описание ", 30) // multi-byte, well over 75 octets
	cal := Calendar{
		ProdID: prodID,
		TZID:   "Europe/Moscow",
		Events: []Event{{
			UID:         "e1",
			Summary:     "S",
			Description: longDesc,
			Start:       time.Date(2026, 2, 10, 9, 0, 0, 0, msk),
			End:         time.Date(2026, 2, 10, 10, 0, 0, 0, msk),
			Stamp:       time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
			Status:      StatusConfirmed,
		}},
	}

	got := cal.Render()
	for line := range strings.SplitSeq(got, "\r\n") {
		if len(line) > 75 {
			t.Errorf("line exceeds 75 octets (%d): %q", len(line), line)
		}
		if !utf8.ValidString(line) {
			t.Errorf("folding split a UTF-8 sequence: %q", line)
		}
	}
	// Continuation lines must begin with a single space.
	if !strings.Contains(got, "\r\n ") {
		t.Errorf("expected folded continuation lines starting with a space")
	}
}

func TestRender_CountAndCategories(t *testing.T) {
	count := 5
	cal := Calendar{
		ProdID: prodID,
		TZID:   "Europe/Moscow",
		Events: []Event{{
			UID:        "e1",
			Summary:    "Консультация",
			Start:      time.Date(2026, 2, 10, 9, 0, 0, 0, msk),
			End:        time.Date(2026, 2, 10, 10, 0, 0, 0, msk),
			Stamp:      time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
			Recurrence: &Recurrence{Frequency: FreqDaily, Count: &count},
			Categories: []string{"Лекция", "Важно, срочно"},
			Status:     StatusConfirmed,
		}},
	}

	got := cal.Render()
	if !strings.Contains(got, "RRULE:FREQ=DAILY;COUNT=5") {
		t.Errorf("expected RRULE with COUNT; got:\n%s", got)
	}
	// Category members are comma-separated; commas inside a member are escaped.
	if !strings.Contains(got, `CATEGORIES:Лекция,Важно\, срочно`) {
		t.Errorf("expected escaped CATEGORIES; got:\n%s", got)
	}
}

func TestRender_UTCFallbackForUnknownZone(t *testing.T) {
	cal := Calendar{
		ProdID: prodID,
		TZID:   "", // no zone → render instants in UTC, no VTIMEZONE
		Events: []Event{{
			UID:     "e1",
			Summary: "S",
			Start:   time.Date(2026, 2, 10, 9, 0, 0, 0, time.UTC),
			End:     time.Date(2026, 2, 10, 10, 0, 0, 0, time.UTC),
			Stamp:   time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
			Status:  StatusConfirmed,
		}},
	}

	got := cal.Render()
	if strings.Contains(got, "BEGIN:VTIMEZONE") {
		t.Errorf("no VTIMEZONE expected when zone is unknown; got:\n%s", got)
	}
	if !strings.Contains(got, "DTSTART:20260210T090000Z") {
		t.Errorf("expected UTC DTSTART with Z suffix; got:\n%s", got)
	}
}

func TestRender_OmitsCancelledStatusBlockWhenEmpty(t *testing.T) {
	// A canceled event must carry the RFC 5545 cancellation status so
	// subscribers drop it from their view.
	cal := Calendar{
		ProdID: prodID,
		TZID:   "Europe/Moscow",
		Events: []Event{{
			UID:     "e1",
			Summary: "Отменённая пара",
			Start:   time.Date(2026, 2, 10, 9, 0, 0, 0, msk),
			End:     time.Date(2026, 2, 10, 10, 0, 0, 0, msk),
			Stamp:   time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
			Status:  StatusCancelled,
		}},
	}

	got := cal.Render()
	wantStatus := "STATUS:CANCELLED" //nolint:misspell // RFC 5545 STATUS value
	if !strings.Contains(got, wantStatus) {
		t.Errorf("expected cancellation status; got:\n%s", got)
	}
}
