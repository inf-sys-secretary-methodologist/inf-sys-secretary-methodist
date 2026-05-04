// parseLocalDate strips the time + offset from an ISO 8601 string and
// reconstructs local midnight on the same calendar date. The grading
// UI displays due dates and graded-at timestamps; what matters to a
// teacher is the calendar day, not the exact instant. Naive
// new Date(iso) parses the instant in UTC and shifts the visible
// date in negative-UTC timezones — CLAUDE.md rule #9 forbids it.
export function parseLocalDate(iso: string): Date {
  // ISO 8601 always starts with the date portion before the 'T'.
  // For date-only inputs (no 'T') split returns the original string,
  // which is itself a valid YYYY-MM-DD prefix.
  const dateOnly = iso.split('T')[0]
  return new Date(`${dateOnly}T00:00:00`)
}
