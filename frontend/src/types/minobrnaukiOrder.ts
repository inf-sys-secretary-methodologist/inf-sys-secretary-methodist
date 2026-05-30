// Минобрнауки order (приказ) types. Mirror the backend response DTOs in
// internal/modules/work_program/interfaces/http/handlers/minobrnauki_order_handler.go
// (MinobrnaukiOrderDTO / MinobrnaukiOrderSummaryDTO / MinobrnaukiOrdersListResponse).
// An order is the regulatory record (ADR-11) that drives the AI
// bulk-revision flow (ADR-12) over every affected РПД.

// MinobrnaukiOrderChangeScope mirrors the domain enum
// (domain/types.go MinobrnaukiOrderChangeScope{Minor,Major}).
export type MinobrnaukiOrderChangeScope = 'minor' | 'major'

export const MINOBRNAUKI_ORDER_CHANGE_SCOPES: MinobrnaukiOrderChangeScope[] = ['minor', 'major']

// MinobrnaukiOrderSummary is the list-row projection — order fields
// without the affected set (kept off the list to stay cheap, mirror of
// MinobrnaukiOrderSummaryDTO). published_at is a calendar date
// (YYYY-MM-DD); created_at is RFC 3339.
export interface MinobrnaukiOrderSummary {
  id: number
  order_number: string
  title: string
  published_at: string
  document_id?: number
  change_scope: MinobrnaukiOrderChangeScope
  summary?: string
  uploaded_by: number
  created_at: string
}

// MinobrnaukiOrder is the full single-order shape — the summary plus the
// ids of the work programs the order affects (mirror of
// MinobrnaukiOrderDTO). The detail endpoint hydrates the affected set.
export interface MinobrnaukiOrder extends MinobrnaukiOrderSummary {
  affected_work_program_ids: number[]
}

// MinobrnaukiOrdersListResponse is the page response shape
// (MinobrnaukiOrdersListResponse on the server).
export interface MinobrnaukiOrdersListResponse {
  items: MinobrnaukiOrderSummary[]
  total: number
}

// MinobrnaukiOrderListFilter matches the query string accepted by GET
// /api/v1/minobrnauki-orders. All fields optional — the use case applies
// a flat role gate (non-student staff) + pagination defaults server-side.
export interface MinobrnaukiOrderListFilter {
  change_scope?: MinobrnaukiOrderChangeScope
  uploaded_by?: number
  limit?: number
  offset?: number
}
