package logging

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

// AuditLogRepositoryPG is the PostgreSQL adapter for AuditLogWriter.
// Calls db.ExecContext directly — never via *sql.Tx — so writes happen
// independently of any business transaction (ADR-2). This is what
// guarantees that denied/failed business operations still record an
// audit row: their tx rolls back, the audit INSERT does not.
//
// The *sql.DB handle is the singleton pool shared with every other
// repository; "independent" here refers to the transaction boundary,
// not the connection pool. Postgres leases a fresh pooled connection
// for every ExecContext outside a tx.
type AuditLogRepositoryPG struct {
	db *sql.DB
}

// NewAuditLogRepositoryPG constructs the repository.
func NewAuditLogRepositoryPG(db *sql.DB) *AuditLogRepositoryPG {
	return &AuditLogRepositoryPG{db: db}
}

// Write persists one audit event to audit_logs. CreatedAt is intentionally
// not passed — the column has DEFAULT CURRENT_TIMESTAMP authoritative on
// the server (ADR-4); whatever the caller put in log.CreatedAt is
// ignored on the write path (relevant only when AuditLog is later read
// back from the database).
//
// A nil Fields map is normalized to an empty map so the JSONB column
// receives the literal {} rather than null — caller-side guards in
// AuditLogger already cover this path, the writer-side normalization
// is defense-in-depth for direct callers (e.g., the v0.131.0 read API
// reseed path if one is ever wired).
func (r *AuditLogRepositoryPG) Write(ctx context.Context, log *AuditLog) error {
	fields := log.Fields
	if fields == nil {
		fields = map[string]any{}
	}
	fieldsJSON, err := json.Marshal(fields)
	if err != nil {
		return fmt.Errorf("audit_logs: marshal fields: %w", err)
	}

	const query = `INSERT INTO audit_logs
		(action, resource, actor_user_id, actor_ip, correlation_id, fields)
		VALUES ($1, $2, $3, $4, $5, $6)`

	if _, err := r.db.ExecContext(ctx, query,
		log.Action,
		log.Resource,
		log.ActorUserID,
		log.ActorIP,
		log.CorrelationID,
		fieldsJSON,
	); err != nil {
		return fmt.Errorf("audit_logs: write: %w", err)
	}
	return nil
}

// List returns a page of audit_logs rows matching the filter together
// with the total number of matching rows (ignoring Limit/Offset). The
// COUNT and SELECT queries share the same WHERE predicate so an empty
// page past the result-set tail still reports the correct dataset
// size for pagination.
//
// Filter dimensions are independent: zero-valued fields disable that
// dimension. Sentinel-arg style mirrors the project's List repos
// (curriculum, etc.) — each predicate is gated by a null-coalesce
// against the typed parameter so the query plan stays stable.
//
// Stub: behavior deferred to the matching GREEN commit. The signature
// + return shape are declared here so the RED test file compiles
// against the AuditLogReader port without breaking the package build.
func (r *AuditLogRepositoryPG) List(ctx context.Context, filter AuditLogFilter) (AuditLogListResult, error) {
	_ = ctx
	_ = filter
	return AuditLogListResult{}, errAuditLogListNotImplemented
}

// errAuditLogListNotImplemented marks the RED stub so the upcoming
// GREEN commit replaces both the body and removes this sentinel. Not
// exported — never crosses the package boundary; tests assert against
// the error message in the RED commit only.
var errAuditLogListNotImplemented = fmt.Errorf("audit_logs: list: not implemented")
