// Package persistence provides PostgreSQL implementations of the
// student_debts module's repository ports.
package persistence

import (
	"context"
	"database/sql"
)

// DBTX is the narrow database surface needed by student_debts repository
// implementations — both `*sql.DB` and `*sql.Tx` satisfy it. Mirrors the
// work_program / curriculum pattern so the same SQL runs against either.
//
// Includes BeginTx because the StudentDebt aggregate writes two tables
// atomically (root + attempts) — the repository owns the transaction
// lifecycle, not the use case.
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}
