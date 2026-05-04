// Stub of the local-midnight ISO parser. Real implementation lands
// in the GREEN commit; this stub returns Invalid Date so the failing
// tests fail meaningfully.

export function parseLocalDate(_iso: string): Date {
  return new Date(NaN)
}
