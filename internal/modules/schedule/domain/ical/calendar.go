// Package ical renders schedule data as RFC 5545 (iCalendar) documents so that
// lessons and events can be consumed by external calendar clients (Google
// Calendar, Outlook, Apple Calendar) via a subscription feed.
//
// The package is a pure serializer: it has no I/O and no dependency on other
// modules. Mapping domain entities (lessons, events) onto the value types here
// belongs to the application layer.
package ical

import (
	"strconv"
	"strings"
	"time"
)

// Frequency enumerates the RRULE FREQ values supported by the feed.
type Frequency string

// Frequency values.
const (
	FreqDaily   Frequency = "DAILY"
	FreqWeekly  Frequency = "WEEKLY"
	FreqMonthly Frequency = "MONTHLY"
	FreqYearly  Frequency = "YEARLY"
)

// Weekday enumerates the two-letter RRULE BYDAY codes.
type Weekday string

// Weekday values.
const (
	Monday    Weekday = "MO"
	Tuesday   Weekday = "TU"
	Wednesday Weekday = "WE"
	Thursday  Weekday = "TH"
	Friday    Weekday = "FR"
	Saturday  Weekday = "SA"
	Sunday    Weekday = "SU"
)

// Status enumerates the VEVENT STATUS values used by the feed.
type Status string

// Status values.
const (
	StatusConfirmed Status = "CONFIRMED"
	// StatusCancelled uses the RFC 5545 spelling of the STATUS value
	// (double "L"), which external calendar clients match verbatim.
	StatusCancelled Status = "CANCELLED" //nolint:misspell // RFC 5545 STATUS value
)

// The value types below (Recurrence, Event, Calendar) are render models: plain
// serialization structs whose only invariant is producing spec-compliant
// output, which the renderer guarantees. Business invariants (e.g. a lesson's
// time range) belong to the schedule domain entities and are enforced when the
// application layer maps those entities onto these types.

// Recurrence models an RFC 5545 RRULE. A nil *Recurrence means a single,
// non-repeating event.
type Recurrence struct {
	Frequency Frequency
	Interval  int        // occurrences repeat every Interval units; <=1 omits INTERVAL
	Until     *time.Time // inclusive end of the series, always rendered in UTC
	Count     *int       // number of occurrences; ignored when Until is also set
	ByDay     []Weekday  // BYDAY component
}

// Event models a single VEVENT.
type Event struct {
	UID         string
	Summary     string
	Description string
	Location    string
	Start       time.Time
	End         time.Time
	AllDay      bool
	Stamp       time.Time // DTSTAMP; injected by the caller for deterministic output
	Recurrence  *Recurrence
	ExDates     []time.Time // EXDATE — occurrences removed from the series (e.g. cancellations)
	Status      Status
	Categories  []string
}

// Calendar models a VCALENDAR document.
type Calendar struct {
	ProdID string  // PRODID identifier
	Name   string  // X-WR-CALNAME display name
	TZID   string  // IANA time zone; timed events render as wall-clock in this zone
	Events []Event // calendar components
}

// dateTimeLayout is the RFC 5545 DATE-TIME form (basic ISO 8601, no separators).
const dateTimeLayout = "20060102T150405"

// dateLayout is the RFC 5545 DATE form used for all-day components.
const dateLayout = "20060102"

// Render serializes the calendar into an RFC 5545 document using CRLF line
// endings, 75-octet line folding and RFC-compliant text escaping.
func (c Calendar) Render() string {
	zone, hasZone := lookupZone(c.TZID)

	var lines []string
	add := func(s string) { lines = append(lines, foldLine(s)) }

	add("BEGIN:VCALENDAR")
	add("VERSION:2.0")
	add("PRODID:" + escapeText(c.ProdID))
	add("CALSCALE:GREGORIAN")
	add("METHOD:PUBLISH")
	if c.Name != "" {
		add("X-WR-CALNAME:" + escapeText(c.Name))
	}
	if hasZone {
		add("X-WR-TIMEZONE:" + c.TZID)
		for l := range strings.SplitSeq(zone.vtimezone, "\n") {
			add(l)
		}
	}
	for _, ev := range c.Events {
		c.renderEvent(add, ev, zone, hasZone)
	}
	add("END:VCALENDAR")

	return strings.Join(lines, "\r\n") + "\r\n"
}

// renderEvent appends the folded property lines of a single VEVENT via add.
func (c Calendar) renderEvent(add func(string), ev Event, zone tzInfo, hasZone bool) {
	add("BEGIN:VEVENT")
	add("UID:" + escapeText(ev.UID))
	add("DTSTAMP:" + formatUTC(ev.Stamp))

	add(c.dateProp("DTSTART", ev.Start, ev.AllDay, zone, hasZone))
	// DTEND is optional in RFC 5545; omit it rather than emit a garbage
	// zero-value instant when the caller supplies no end time.
	if !ev.End.IsZero() {
		add(c.dateProp("DTEND", ev.End, ev.AllDay, zone, hasZone))
	}

	if ev.Recurrence != nil {
		add("RRULE:" + renderRRule(*ev.Recurrence))
	}
	for _, ex := range ev.ExDates {
		add(c.dateProp("EXDATE", ex, ev.AllDay, zone, hasZone))
	}

	if ev.Summary != "" {
		add("SUMMARY:" + escapeText(ev.Summary))
	}
	if ev.Description != "" {
		add("DESCRIPTION:" + escapeText(ev.Description))
	}
	if ev.Location != "" {
		add("LOCATION:" + escapeText(ev.Location))
	}
	if ev.Status != "" {
		add("STATUS:" + string(ev.Status))
	}
	if len(ev.Categories) > 0 {
		escaped := make([]string, len(ev.Categories))
		for i, cat := range ev.Categories {
			escaped[i] = escapeText(cat)
		}
		add("CATEGORIES:" + strings.Join(escaped, ","))
	}
	add("END:VEVENT")
}

// dateProp renders a date/date-time property (DTSTART/DTEND/EXDATE) choosing
// the VALUE=DATE, TZID-local or UTC form depending on the event and zone.
//
// The all-day branch formats the value in its own location, so callers must
// pass all-day times already as the intended wall-clock date; the timed branch
// converts the instant into the calendar zone.
func (c Calendar) dateProp(name string, t time.Time, allDay bool, zone tzInfo, hasZone bool) string {
	switch {
	case allDay:
		return name + ";VALUE=DATE:" + t.Format(dateLayout)
	case hasZone:
		return name + ";TZID=" + c.TZID + ":" + t.In(zone.location).Format(dateTimeLayout)
	default:
		return name + ":" + formatUTC(t)
	}
}

// formatUTC renders an instant in RFC 5545 UTC form (trailing Z).
func formatUTC(t time.Time) string {
	return t.UTC().Format(dateTimeLayout) + "Z"
}

// renderRRule serializes a Recurrence as an RFC 5545 RRULE value.
func renderRRule(r Recurrence) string {
	parts := []string{"FREQ=" + string(r.Frequency)}
	if r.Interval > 1 {
		parts = append(parts, "INTERVAL="+strconv.Itoa(r.Interval))
	}
	// RFC 5545 §3.3.10 forbids UNTIL and COUNT in the same RRULE; UNTIL wins
	// when a caller supplies both, so the output is always spec-compliant.
	switch {
	case r.Until != nil:
		parts = append(parts, "UNTIL="+formatUTC(*r.Until))
	case r.Count != nil:
		parts = append(parts, "COUNT="+strconv.Itoa(*r.Count))
	}
	if len(r.ByDay) > 0 {
		days := make([]string, len(r.ByDay))
		for i, d := range r.ByDay {
			days[i] = string(d)
		}
		parts = append(parts, "BYDAY="+strings.Join(days, ","))
	}
	return strings.Join(parts, ";")
}
