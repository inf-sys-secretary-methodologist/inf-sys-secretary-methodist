package logging

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

// AuditLogRepositoryPG is the PostgreSQL adapter for AuditLogWriter.
// Holds its own *sql.DB handle so writes happen on an independent
// connection — never inside a business transaction (ADR-2). This is
// what guarantees that denied/failed business operations still record
// an audit row: their tx rolls back, the audit INSERT does not.
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
func (r *AuditLogRepositoryPG) Write(ctx context.Context, log *AuditLog) error {
	fieldsJSON, err := json.Marshal(log.Fields)
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
