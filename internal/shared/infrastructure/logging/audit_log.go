package logging

import (
	"context"
	"time"
)

// AuditLog is the persisted row in audit_logs (migration 036). DTO only —
// no business invariants on the Go side; CHECK constraints in SQL enforce
// non-empty action/resource. The struct mirrors the column set 1:1 so the
// repository writer can map without an intermediate translation layer.
//
// Nullable columns (actor_user_id / actor_ip / correlation_id) use
// pointer types so a missing value writes a SQL NULL rather than zero.
//
// CreatedAt is NOT passed to INSERT — the column has DEFAULT
// CURRENT_TIMESTAMP and is set authoritatively by Postgres. The field is
// populated on subsequent reads (v0.131.0 read API).
type AuditLog struct {
	ID            int64
	CreatedAt     time.Time
	Action        string
	Resource      string
	ActorUserID   *int64
	ActorIP       *string
	CorrelationID *string
	Fields        map[string]any
}

// AuditLogWriter is the narrow port used by AuditLogger to persist an
// audit event after structured-log emission (ADR-2 sync write,
// independent of any business transaction; ADR-3 failure is logged and
// not propagated). Concrete implementation AuditLogRepositoryPG
// satisfies structurally.
type AuditLogWriter interface {
	Write(ctx context.Context, log *AuditLog) error
}

// AuditLogFilter narrows a List query. Zero-valued fields are treated
// as "no filter on this dimension". Limit/Offset are honored by the
// repository; a non-positive Limit means "no extra clamp" — callers
// (use cases / handlers) are responsible for capping per-page size.
type AuditLogFilter struct {
	// Action filters by exact match when non-empty (e.g.
	// "curriculum.approved").
	Action string
	// Resource filters by exact match when non-empty (e.g. "curriculum",
	// "document").
	Resource string
	// UserID, when non-nil, restricts to events emitted by that actor.
	UserID *int64
	// From is the inclusive lower bound on created_at. Nil = no bound.
	From *time.Time
	// To is the exclusive upper bound on created_at (half-open range).
	// Nil = no bound.
	To *time.Time
	// Limit caps the returned page size. Repositories must treat
	// non-positive values as "no clamp".
	Limit int
	// Offset is the starting index for pagination.
	Offset int
}

// AuditLogListResult bundles a page of audit events with the
// unfiltered total so the UI can render pagination controls without
// a second query.
type AuditLogListResult struct {
	Items []*AuditLog
	Total int
}

// AuditLogReader is the narrow port for the audit-log read API.
// Concrete implementation AuditLogRepositoryPG satisfies structurally.
// Kept separate from AuditLogWriter so the AuditLogger collaborator
// stays write-only and read consumers (admin handler) depend on the
// narrower interface.
type AuditLogReader interface {
	List(ctx context.Context, filter AuditLogFilter) (AuditLogListResult, error)
}
