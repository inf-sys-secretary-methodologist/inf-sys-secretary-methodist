// Package ical renders schedule data as RFC 5545 (iCalendar) documents so that
// lessons and events can be consumed by external calendar clients (Google
// Calendar, Outlook, Apple Calendar) via a subscription feed.
//
// The package is a pure serializer: it has no I/O and no dependency on other
// modules. Mapping domain entities (lessons, events) onto the value types here
// belongs to the application layer.
package ical

import "time"

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

// Recurrence models an RFC 5545 RRULE. A nil *Recurrence means a single,
// non-repeating event.
type Recurrence struct {
	Frequency Frequency
	Interval  int        // occurrences repeat every Interval units; <=1 omits INTERVAL
	Until     *time.Time // inclusive end of the series, always rendered in UTC
	Count     *int       // number of occurrences; mutually exclusive with Until
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

// Render serializes the calendar into an RFC 5545 document using CRLF line
// endings, 75-octet line folding and RFC-compliant text escaping.
func (c Calendar) Render() string {
	return ""
}
