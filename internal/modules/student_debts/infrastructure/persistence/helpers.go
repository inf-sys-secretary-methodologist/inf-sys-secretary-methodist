package persistence

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
)

// execQuerier is the narrow surface accepted by INSERT/UPDATE helpers
// inside a transaction — both `*sql.Tx` and DBTX satisfy it.
type execQuerier interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// pqUniqueViolation is the SQLSTATE code for a unique-constraint
// violation in PostgreSQL. Mirrors the inline pattern used by the
// work_program / curriculum modules to keep this bounded context free
// of a dependency on the shared error mapper for a single sentinel.
const pqUniqueViolation = "23505"

// uqStudentDebtIdentity is the migration 050 natural-key constraint
// name. The repo matches the constraint name (not just the SQLSTATE)
// so a future uniqueness check on a different tuple gets its own
// mapping rather than collapsing onto ErrStudentDebtIdentityExists.
const uqStudentDebtIdentity = "uq_student_debts_identity"

// isIdentityViolation reports whether err is a PostgreSQL unique
// violation against the natural-key constraint.
func isIdentityViolation(err error) bool {
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		return false
	}
	if string(pqErr.Code) != pqUniqueViolation {
		return false
	}
	// Match constraint name when available — defensive against future
	// uniqueness constraints on the same table.
	return pqErr.Constraint == "" || pqErr.Constraint == uqStudentDebtIdentity
}

// nullableInt64Ptr wraps a *int64 as a SQL NULL when the pointer is nil.
// Used for optional id columns (student_user_id, discipline_id,
// recorded_by).
func nullableInt64Ptr(p *int64) sql.NullInt64 {
	if p == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *p, Valid: true}
}

// nullableIntPtr wraps an *int as SQL NULL when the pointer is nil. Used
// for the optional grade column; the grade range is institution-specific
// and small, so the int→int32 conversion cannot realistically overflow.
func nullableIntPtr(p *int) sql.NullInt32 {
	if p == nil {
		return sql.NullInt32{}
	}
	return sql.NullInt32{Int32: int32(*p), Valid: true} // #nosec G115 -- academic grades are small bounded values
}

// nullableTimePtr wraps a *time.Time as SQL NULL when the pointer is nil.
// Used for the optional recorded_at column.
func nullableTimePtr(p *time.Time) sql.NullTime {
	if p == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *p, Valid: true}
}
