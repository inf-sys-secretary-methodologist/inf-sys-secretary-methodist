import { validateGrade } from '../validation'

describe('validateGrade', () => {
  describe('accepts valid values', () => {
    const cases: Array<{ raw: string; max: number; expected: number }> = [
      { raw: '0', max: 100, expected: 0 },
      { raw: '50', max: 100, expected: 50 },
      { raw: '100', max: 100, expected: 100 },
      { raw: '85', max: 100, expected: 85 },
      { raw: '5', max: 5, expected: 5 },
    ]
    it.each(cases)('value $raw against max $max → $expected', ({ raw, max, expected }) => {
      const r = validateGrade(raw, max)
      expect(r.ok).toBe(true)
      if (r.ok) expect(r.value).toBe(expected)
    })
  })

  describe('rejects invalid values', () => {
    const cases: Array<{ raw: string; max: number; reason: string }> = [
      { raw: '', max: 100, reason: 'NOT_A_NUMBER' },
      { raw: 'abc', max: 100, reason: 'NOT_A_NUMBER' },
      { raw: '-1', max: 100, reason: 'NEGATIVE' },
      { raw: '-50', max: 100, reason: 'NEGATIVE' },
      { raw: '101', max: 100, reason: 'OVER_MAX' },
      { raw: '500', max: 100, reason: 'OVER_MAX' },
      { raw: '85.5', max: 100, reason: 'NOT_INTEGER' },
    ]
    it.each(cases)('value "$raw" / max $max → $reason', ({ raw, max, reason }) => {
      const r = validateGrade(raw, max)
      expect(r.ok).toBe(false)
      if (!r.ok) expect(r.reason).toBe(reason)
    })
  })
})
