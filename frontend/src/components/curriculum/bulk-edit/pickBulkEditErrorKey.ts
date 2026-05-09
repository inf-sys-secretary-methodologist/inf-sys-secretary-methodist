// pickBulkEditErrorKey — pure (HTTP status, backend error code) →
// i18n key resolver для toast/banner messages on bulk-edit failures.
//
// Backend sentinel mapping (bulk_discipline_items_handler.go::
// mapBulkEditError): see plan ADR-12 + handler comments. 409 conflict
// is handled separately by bulkEditDisciplineItems mutation (returns
// {kind: 'conflict'} discriminated variant) — pickBulkEditErrorKey is
// invoked ONLY for non-409 axios exceptions.
//
// Default-deny на unknown status / code → errorGeneric (per
// feedback_status_aware_error_mapping.md). User sees recovery prompt
// rather than getting stuck on opaque failure.

export type BulkEditErrorKey =
  | 'errorEmptyBulk'
  | 'errorCrossSection'
  | 'errorNotEditable'
  | 'errorInvalidInput'
  | 'errorNotFound'
  | 'errorForbidden'
  | 'errorGeneric'

export function pickBulkEditErrorKey(
  status: number | undefined,
  errorCode: string | undefined
): BulkEditErrorKey {
  // Pair 4 RED stub — always returns errorGeneric. Tests asserting
  // specific keys для 422 sub-cases / 404 / 403 fail. GREEN replaces с
  // real switch.
  void status
  void errorCode
  return 'errorGeneric'
}
