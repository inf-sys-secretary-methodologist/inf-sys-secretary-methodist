package persistence

import (
	"context"
	"database/sql"
)

// DBTX is the narrow database surface needed by curriculum repository
// implementations — both `*sql.DB` (single-connection mode) и `*sql.Tx`
// (transactional mode) satisfy it. Repos accept DBTX so the same SQL
// queries can run against either, without code duplication.
//
// This unlocks the BulkDisciplineItemsUnitOfWork pattern (v0.128.3 ADR-10):
// usecase Begins a tx, builds repos backed by `*sql.Tx`, then Commit или
// Rollback wraps the whole operation atomically.
//
// Common Go pattern (mirrors sqlc-generated code). DBTX intentionally
// excludes `BeginTx` — the UoW (not the repo) owns transaction lifecycle.
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}
