package usecases

import (
	"fmt"
	"strings"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/ical"
)

// weekdayToICal maps the domain ISO weekday to the RFC 5545 BYDAY code.
var weekdayToICal = map[domain.DayOfWeek]ical.Weekday{
	domain.Monday:    ical.Monday,
	domain.Tuesday:   ical.Tuesday,
	domain.Wednesday: ical.Wednesday,
	domain.Thursday:  ical.Thursday,
	domain.Friday:    ical.Friday,
	domain.Saturday:  ical.Saturday,
	domain.Sunday:    ical.Sunday,
}

// frequencyToICal maps the domain recurrence frequency to the RRULE FREQ value.
var frequencyToICal = map[entities.RecurrenceFrequency]ical.Frequency{
	entities.FrequencyDaily:   ical.FreqDaily,
	entities.FrequencyWeekly:  ical.FreqWeekly,
	entities.FrequencyMonthly: ical.FreqMonthly,
	entities.FrequencyYearly:  ical.FreqYearly,
}

// lessonToICalEvent maps a recurring lesson onto a single VEVENT with an RRULE.
// The bool is false when the lesson yields no occurrences within its date range
// and must be skipped.
func lessonToICalEvent(l *entities.Lesson, uidDomain string, loc *time.Location) (ical.Event, bool) {
	first := firstLessonOccurrence(l.DateStart, l.DayOfWeek, l.WeekType, loc)
	until := endOfDay(l.DateEnd, loc)
	if first.After(until) {
		return ical.Event{}, false
	}

	sh, sm, ss := parseDayTime(l.TimeStart)
	eh, em, es := parseDayTime(l.TimeEnd)
	start := withTime(first, sh, sm, ss, loc)
	end := withTime(first, eh, em, es, loc)

	interval := 1
	if l.WeekType == domain.WeekTypeOdd || l.WeekType == domain.WeekTypeEven {
		interval = 2
	}

	status := ical.StatusConfirmed
	if l.IsCancelled {
		status = ical.StatusCancelled
	}

	return ical.Event{
		UID:         fmt.Sprintf("lesson-%d@%s", l.ID, uidDomain),
		Summary:     lessonSummary(l),
		Description: lessonDescription(l),
		Location:    classroomLabel(l.Classroom),
		Start:       start,
		End:         end,
		Stamp:       l.UpdatedAt,
		Recurrence: &ical.Recurrence{
			Frequency: ical.FreqWeekly,
			Interval:  interval,
			Until:     &until,
			ByDay:     []ical.Weekday{weekdayToICal[l.DayOfWeek]},
		},
		Status: status,
	}, true
}

// eventToICalEvent maps a calendar event onto a VEVENT. The bool is false when
// the event must be skipped (soft-deleted).
func eventToICalEvent(e *entities.Event, uidDomain string, loc *time.Location) (ical.Event, bool) {
	if e.IsDeleted() {
		return ical.Event{}, false
	}

	out := ical.Event{
		UID:     fmt.Sprintf("event-%d@%s", e.ID, uidDomain),
		Summary: e.Title,
		Start:   e.StartTime.In(loc),
		AllDay:  e.AllDay,
		Stamp:   e.UpdatedAt,
		Status:  ical.StatusConfirmed,
	}
	if e.EndTime != nil {
		out.End = e.EndTime.In(loc)
	}
	if e.Description != nil {
		out.Description = *e.Description
	}
	if e.Location != nil {
		out.Location = *e.Location
	}
	if e.Status == entities.EventStatusCancelled {
		out.Status = ical.StatusCancelled
	}
	if e.IsRecurring && e.RecurrenceRule != nil {
		out.Recurrence = recurrenceToICal(e.RecurrenceRule)
	}
	return out, true
}

// recurrenceToICal maps a domain recurrence rule onto an ical.Recurrence.
func recurrenceToICal(r *entities.RecurrenceRule) *ical.Recurrence {
	byDay := make([]ical.Weekday, 0, len(r.ByWeekday))
	for _, wd := range r.ByWeekday {
		byDay = append(byDay, ical.Weekday(string(wd)))
	}
	return &ical.Recurrence{
		Frequency: frequencyToICal[r.Frequency],
		Interval:  r.Interval,
		Until:     r.Until,
		Count:     r.Count,
		ByDay:     byDay,
	}
}

// firstLessonOccurrence returns the first date on or after dateStart that falls
// on the lesson's weekday and, for odd/even lessons, on a week of the matching
// ISO-week parity.
func firstLessonOccurrence(dateStart time.Time, day domain.DayOfWeek, wt domain.WeekType, loc *time.Location) time.Time {
	d := time.Date(dateStart.Year(), dateStart.Month(), dateStart.Day(), 0, 0, 0, 0, loc)
	for isoWeekday(d) != int(day) {
		d = d.AddDate(0, 0, 1)
	}
	if wt == domain.WeekTypeOdd || wt == domain.WeekTypeEven {
		wantOdd := wt == domain.WeekTypeOdd
		for {
			_, week := d.ISOWeek()
			if (week%2 == 1) == wantOdd {
				break
			}
			d = d.AddDate(0, 0, 7)
		}
	}
	return d
}

// isoWeekday returns the ISO weekday (Monday=1 … Sunday=7) of t.
func isoWeekday(t time.Time) int {
	if wd := int(t.Weekday()); wd != 0 {
		return wd
	}
	return 7 // time.Sunday is 0
}

// withTime combines a date with an hour/minute/second in loc.
func withTime(date time.Time, h, m, s int, loc *time.Location) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), h, m, s, 0, loc)
}

// endOfDay returns 23:59:59 of the given date in loc (the inclusive UNTIL bound).
func endOfDay(date time.Time, loc *time.Location) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 0, loc)
}

// parseDayTime parses a PostgreSQL TIME value ("HH:MM" or "HH:MM:SS").
func parseDayTime(s string) (h, m, sec int) {
	_, _ = fmt.Sscanf(s, "%d:%d:%d", &h, &m, &sec)
	return h, m, sec
}

// lessonSummary builds "<discipline> (<type>)" from the hydrated lesson.
func lessonSummary(l *entities.Lesson) string {
	name := "Занятие"
	if l.Discipline != nil && l.Discipline.Name != "" {
		name = l.Discipline.Name
	}
	if l.LessonType != nil {
		label := l.LessonType.ShortName
		if label == "" {
			label = l.LessonType.Name
		}
		if label != "" {
			return fmt.Sprintf("%s (%s)", name, label)
		}
	}
	return name
}

// lessonDescription assembles teacher and notes lines.
func lessonDescription(l *entities.Lesson) string {
	var parts []string
	if l.Teacher != nil && l.Teacher.Name != "" {
		parts = append(parts, "Преподаватель: "+l.Teacher.Name)
	}
	if l.Notes != nil && *l.Notes != "" {
		parts = append(parts, *l.Notes)
	}
	return strings.Join(parts, "\n")
}

// classroomLabel formats a classroom as its name, or "<building>-<number>".
func classroomLabel(c *entities.Classroom) string {
	if c == nil {
		return ""
	}
	if c.Name != nil && *c.Name != "" {
		return *c.Name
	}
	return c.Building + "-" + c.Number
}
