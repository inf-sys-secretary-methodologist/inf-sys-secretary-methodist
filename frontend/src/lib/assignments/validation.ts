// Grade-input validation mirroring the backend invariants:
//   0 ≤ value ≤ maxScore, value integer.
// Pre-flighting them on the client avoids a round-trip through 422
// for the obvious typos and gives the user a localised reason
// immediately. The backend remains the source of truth — this is
// guidance, not security.

export type GradeValidationResult =
  | { ok: true; value: number }
  | { ok: false; reason: 'NOT_A_NUMBER' | 'NEGATIVE' | 'OVER_MAX' | 'NOT_INTEGER' }

export function validateGrade(raw: string, maxScore: number): GradeValidationResult {
  const trimmed = raw.trim()
  if (trimmed === '') return { ok: false, reason: 'NOT_A_NUMBER' }

  const num = Number(trimmed)
  if (!Number.isFinite(num)) return { ok: false, reason: 'NOT_A_NUMBER' }
  if (!Number.isInteger(num)) return { ok: false, reason: 'NOT_INTEGER' }
  if (num < 0) return { ok: false, reason: 'NEGATIVE' }
  if (num > maxScore) return { ok: false, reason: 'OVER_MAX' }

  return { ok: true, value: num }
}
