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
  if (status === 404) return 'errorNotFound'
  if (status === 403) return 'errorForbidden'
  if (status === 422) {
    switch (errorCode) {
      case 'EMPTY_BULK_INPUT':
        return 'errorEmptyBulk'
      case 'CROSS_SECTION_BULK_EDIT':
        return 'errorCrossSection'
      case 'NOT_EDITABLE':
        return 'errorNotEditable'
      case 'INVALID_INPUT':
        return 'errorInvalidInput'
      default:
        return 'errorGeneric'
    }
  }
  return 'errorGeneric'
}
