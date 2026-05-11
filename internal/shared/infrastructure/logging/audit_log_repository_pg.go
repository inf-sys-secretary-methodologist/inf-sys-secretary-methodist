package logging

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
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

// auditLogListColumns is the projection used by List. Keeping it as
// a package-level constant gives the matching test fixture a single
// source of truth for row layout.
const auditLogListColumns = `id, created_at, action, resource,
	actor_user_id, actor_ip, correlation_id, fields`

// auditLogListFilterClause gates each filter dimension on its own
// sentinel-arg null-coalesce so the plan stays stable across filter
// combinations. From/To form a half-open [from, to) range — adding a
// "to" of 2026-05-02 with from 2026-05-01 yields exactly one day of
// rows without double-counting midnight boundaries.
const auditLogListFilterClause = `WHERE ($1 = '' OR action = $1)
	AND ($2 = '' OR resource = $2)
	AND ($3::bigint IS NULL OR actor_user_id = $3::bigint)
	AND ($4::timestamptz IS NULL OR created_at >= $4::timestamptz)
	AND ($5::timestamptz IS NULL OR created_at < $5::timestamptz)`

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
func (r *AuditLogRepositoryPG) List(ctx context.Context, filter AuditLogFilter) (AuditLogListResult, error) {
	var userArg sql.NullInt64
	if filter.UserID != nil {
		userArg = sql.NullInt64{Int64: *filter.UserID, Valid: true}
	}
	var fromArg sql.NullTime
	if filter.From != nil {
		fromArg = sql.NullTime{Time: *filter.From, Valid: true}
	}
	var toArg sql.NullTime
	if filter.To != nil {
		toArg = sql.NullTime{Time: *filter.To, Valid: true}
	}

	countQuery := `SELECT COUNT(*) FROM audit_logs ` + auditLogListFilterClause
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery,
		filter.Action, filter.Resource, userArg, fromArg, toArg,
	).Scan(&total); err != nil {
		return AuditLogListResult{}, fmt.Errorf("audit_logs: count: %w", err)
	}

	listQuery := `SELECT ` + auditLogListColumns + ` FROM audit_logs ` + auditLogListFilterClause + `
		ORDER BY created_at DESC, id DESC
		LIMIT $6 OFFSET $7`

	rows, err := r.db.QueryContext(ctx, listQuery,
		filter.Action, filter.Resource, userArg, fromArg, toArg,
		filter.Limit, filter.Offset,
	)
	if err != nil {
		return AuditLogListResult{}, fmt.Errorf("audit_logs: list: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var items []*AuditLog
	for rows.Next() {
		var (
			id            int64
			createdAt     time.Time
			action        string
			resource      string
			actorUserID   sql.NullInt64
			actorIP       sql.NullString
			correlationID sql.NullString
			fieldsRaw     []byte
		)
		if err := rows.Scan(&id, &createdAt, &action, &resource,
			&actorUserID, &actorIP, &correlationID, &fieldsRaw); err != nil {
			return AuditLogListResult{}, fmt.Errorf("audit_logs: list scan: %w", err)
		}

		fields := map[string]any{}
		if len(fieldsRaw) > 0 {
			if err := json.Unmarshal(fieldsRaw, &fields); err != nil {
				return AuditLogListResult{}, fmt.Errorf("audit_logs: list unmarshal fields: %w", err)
			}
		}

		log := &AuditLog{
			ID:        id,
			CreatedAt: createdAt,
			Action:    action,
			Resource:  resource,
			Fields:    fields,
		}
		if actorUserID.Valid {
			v := actorUserID.Int64
			log.ActorUserID = &v
		}
		if actorIP.Valid {
			v := actorIP.String
			log.ActorIP = &v
		}
		if correlationID.Valid {
			v := correlationID.String
			log.CorrelationID = &v
		}
		items = append(items, log)
	}
	if err := rows.Err(); err != nil {
		return AuditLogListResult{}, fmt.Errorf("audit_logs: list iter: %w", err)
	}

	return AuditLogListResult{Items: items, Total: total}, nil
}
