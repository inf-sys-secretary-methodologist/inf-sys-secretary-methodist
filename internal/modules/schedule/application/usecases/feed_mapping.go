package usecases

import (
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/ical"
)

// lessonToICalEvent maps a recurring lesson onto a single VEVENT with an RRULE.
// The bool is false when the lesson yields no occurrences and must be skipped.
func lessonToICalEvent(l *entities.Lesson, uidDomain string, loc *time.Location) (ical.Event, bool) {
	return ical.Event{}, false
}

// eventToICalEvent maps a calendar event onto a VEVENT. The bool is false when
// the event must be skipped (e.g. soft-deleted).
func eventToICalEvent(e *entities.Event, uidDomain string, loc *time.Location) (ical.Event, bool) {
	return ical.Event{}, false
}
