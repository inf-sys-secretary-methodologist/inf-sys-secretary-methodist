// Package persistence provides PostgreSQL implementations of the
// work_program module's repository ports.
package persistence

import (
	"context"
	"database/sql"
)

// DBTX is the narrow database surface needed by work_program repository
// implementations — both `*sql.DB` and `*sql.Tx` satisfy it. Mirrors
// the curriculum module pattern (sqlc-style) so the same SQL queries
// can run against either without code duplication.
//
// Includes BeginTx because the WorkProgram aggregate writes 7 tables
// atomically (root + 6 child types per ADR-1) — the repository owns
// the transaction lifecycle, not the use case.
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}
