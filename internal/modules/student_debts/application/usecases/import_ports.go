package usecases

import (
	"context"
	"io"
)

// ImportedDebt is one parsed source row yielded by a DebtImporter, before
// domain validation. ServiceID is the stable round-trip id written by a
// previous export (nil for a row authored from scratch); the remaining
// fields are the denormalized natural key + control form as raw wire
// values (validated when the aggregate is (re)built).
type ImportedDebt struct {
	ServiceID       *int64
	StudentFullName string
	GroupName       string
	DisciplineName  string
	Semester        int
	ControlForm     string
	SourceRef       string
}

// DebtImporter parses a registry document (xlsx now, 1С OData later) into
// source rows. The concrete adapter lives in infrastructure/ and is wired
// in main.go — this package owns only the port (DIP). Implementations
// return a parse-level error only for a malformed document; per-row
// validation problems surface as ImportResult.Errors from ImportDebts.
type DebtImporter interface {
	Import(ctx context.Context, r io.Reader) ([]ImportedDebt, error)
}

// DebtSource fetches debt rows from an external system (1С OData) without a
// byte stream — its source is an API, not an uploaded document. The concrete
// adapter lives in infrastructure/ and is wired in main.go (DIP); this
// package owns only the port. A transport/parse failure is returned as an
// error; per-row validation problems surface as ImportResult.Errors from
// Import1CDebts.
type DebtSource interface {
	Fetch(ctx context.Context) ([]ImportedDebt, error)
}

// ImportResult is the lightweight log returned by ImportDebts: how many
// rows were created / updated / skipped (unchanged), plus per-row errors.
// Per design §3 a full conflict-resolution workflow is YAGNI for v1 — row
// errors are reported, not queued.
type ImportResult struct {
	Created int
	Updated int
	Skipped int
	Errors  []ImportRowError
}

// ImportRowError records why a single source row could not be applied.
// Row is the 1-based position in the source; Identity is the human-readable
// natural key for locating the row in the spreadsheet.
type ImportRowError struct {
	Row      int
	Identity string
	Message  string
}
