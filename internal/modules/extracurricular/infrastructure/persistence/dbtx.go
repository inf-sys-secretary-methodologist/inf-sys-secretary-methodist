package persistence

import (
	"context"
	"database/sql"
)

// DBTX is the narrow database surface needed by extracurricular
// repository implementations — both `*sql.DB` (single-connection mode)
// и `*sql.Tx` (transactional mode) satisfy it. Mirror к curriculum
// DBTX pattern (v0.128.3 ADR-10) so the same Save/GetByID/Update
// queries run против either driver mode.
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}
