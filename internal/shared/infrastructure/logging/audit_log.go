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
