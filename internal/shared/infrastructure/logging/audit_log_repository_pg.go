package logging

import (
	"context"
	"database/sql"
	"errors"
)

// AuditLogRepositoryPG is the PostgreSQL adapter for AuditLogWriter.
// Stub: behavior wired in Pair 1 GREEN.
type AuditLogRepositoryPG struct {
	db *sql.DB
}

// NewAuditLogRepositoryPG constructs the repository.
func NewAuditLogRepositoryPG(db *sql.DB) *AuditLogRepositoryPG {
	return &AuditLogRepositoryPG{db: db}
}

// Write — stub (Pair 1 RED). Real INSERT wired in GREEN.
func (r *AuditLogRepositoryPG) Write(_ context.Context, _ *AuditLog) error {
	return errors.New("audit_logs: AuditLogRepositoryPG.Write not implemented (stub)")
}
