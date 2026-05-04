// Stub of grade-input validation. The real rules land in the GREEN
// commit of cycle 9 (validateGrade); this file exists so the failing
// tests compile against the exported symbol.

export type GradeValidationResult =
  | { ok: true; value: number }
  | { ok: false; reason: 'NOT_A_NUMBER' | 'NEGATIVE' | 'OVER_MAX' | 'NOT_INTEGER' }

export function validateGrade(_raw: string, _maxScore: number): GradeValidationResult {
  return { ok: false, reason: 'NOT_A_NUMBER' }
}
