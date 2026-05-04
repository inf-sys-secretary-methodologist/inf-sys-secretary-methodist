import { parseLocalDate } from '../dates'

// CLAUDE.md rule #9: "Date inputs: parse as local midnight
// new Date('YYYY-MM-DDT00:00:00'), not UTC."
//
// The backend returns due_date as a TIMESTAMPTZ ISO-8601 string.
// new Date(iso) parses the instant in UTC; when rendered in a
// negative-UTC timezone it shifts by a calendar day. The grading UI
// only cares about the date portion ("when is it due"), so we drop
// the time + offset and reconstruct local midnight.
describe('parseLocalDate', () => {
  it('parses an ISO Z string and yields the same calendar date in local time', () => {
    const d = parseLocalDate('2026-05-15T00:00:00Z')
    expect(d.getFullYear()).toBe(2026)
    expect(d.getMonth()).toBe(4) // May (zero-based)
    expect(d.getDate()).toBe(15)
    expect(d.getHours()).toBe(0)
    expect(d.getMinutes()).toBe(0)
  })

  it('strips the time and offset for a tz-stamped ISO', () => {
    const d = parseLocalDate('2026-05-15T18:30:00+03:00')
    expect(d.getFullYear()).toBe(2026)
    expect(d.getMonth()).toBe(4)
    expect(d.getDate()).toBe(15)
    expect(d.getHours()).toBe(0)
  })

  it('handles a date-only string', () => {
    const d = parseLocalDate('2026-05-15')
    expect(d.getFullYear()).toBe(2026)
    expect(d.getMonth()).toBe(4)
    expect(d.getDate()).toBe(15)
  })

  it('handles end-of-month boundary that would shift in UTC− zones', () => {
    // 2026-12-31 with a UTC time of 00:00 displays as 2026-12-30
    // in any negative-UTC timezone if naively parsed via new Date(iso).
    const d = parseLocalDate('2026-12-31T00:00:00Z')
    expect(d.getFullYear()).toBe(2026)
    expect(d.getMonth()).toBe(11)
    expect(d.getDate()).toBe(31)
  })
})
