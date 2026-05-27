package persistence

import (
	"context"
	"database/sql"
	"time"
)

// execQuerier is the narrow surface accepted by INSERT/UPDATE helpers
// inside a transaction — both `*sql.Tx` and DBTX satisfy it.
type execQuerier interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// nullableString wraps an empty string as a SQL NULL. Used for
// optional text columns (annotation, reject_reason) where the domain
// uses "" to signal absence.
func nullableString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

// nullableInt64Ptr wraps a *int64 as a SQL NULL when the pointer is
// nil. Used for optional id columns (approver_id).
func nullableInt64Ptr(p *int64) sql.NullInt64 {
	if p == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *p, Valid: true}
}

// nullableTimePtr wraps a *time.Time as a SQL NULL when the pointer
// is nil. Used for optional timestamp columns (approved_at).
func nullableTimePtr(p *time.Time) sql.NullTime {
	if p == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *p, Valid: true}
}

// nullableIntPtr wraps an *int as SQL NULL when the pointer is nil.
// Used for optional numeric columns (week_number ∈ [1, 52], year ∈
// [1900, 2100]) — both ranges enforced by domain invariants and DB
// CHECK constraints, so the int→int32 conversion cannot overflow.
func nullableIntPtr(p *int) sql.NullInt32 {
	if p == nil {
		return sql.NullInt32{}
	}
	return sql.NullInt32{Int32: int32(*p), Valid: true} // #nosec G115 -- domain bounds enforce week ≤ 52, year ≤ 2100
}
