// Audit-log frontend types matching backend DTOs at
// internal/shared/admin/auditlog/handler.go (LogResponse) and the
// AuditLogFilter query-string contract documented inline on the
// handler's GET swagger annotations.
//
// Bounded context: admin observability (only `system_admin` reads this
// surface). Do not mix into other module types — audit-log shape is
// independent of any business aggregate it describes.

// AuditLog is one persisted audit_logs row. Times are RFC3339 strings
// — the backend formats `created_at` in UTC so the client can parse
// without timezone ambiguity. Nullable cells are `null` rather than
// missing keys so downstream UI code does not need optional chaining
// guards on every field access.
export interface AuditLog {
  id: number
  created_at: string
  action: string
  resource: string
  actor_user_id: number | null
  actor_ip: string | null
  correlation_id: string | null
  fields: Record<string, unknown>
}

// AuditLogFilter mirrors the query-string accepted by GET
// /api/admin/audit-logs. All fields optional — backend defaults to
// limit=50 / offset=0 when unset and ignores empty-string filters.
export interface AuditLogFilter {
  action?: string
  resource?: string
  user_id?: number
  // from/to are RFC3339 timestamps; the backend treats the range as
  // half-open [from, to) so daily buckets do not double-count.
  from?: string
  to?: string
  limit?: number
  offset?: number
}

// AuditLogPagination mirrors response.Pagination from
// internal/shared/infrastructure/http/response/response.go — the
// shape lives under meta.pagination on the wire and the hook lifts
// it into the result alongside items for ergonomic consumption.
export interface AuditLogPagination {
  page: number
  per_page: number
  total: number
  total_pages: number
}

// AuditLogListResult is the hook-shaped projection — items + lifted
// pagination meta — so consumers do not need to thread the
// raw envelope shape (success/data/meta) past the hook boundary.
export interface AuditLogListResult {
  items: AuditLog[]
  pagination: AuditLogPagination
}
