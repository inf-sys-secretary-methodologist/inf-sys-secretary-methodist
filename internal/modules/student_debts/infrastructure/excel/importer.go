package excel

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/application/usecases"
)

// Compile-time assertion that the adapter satisfies the application port.
var _ usecases.DebtImporter = (*DebtImporter)(nil)

// Zero-based registry column indices the importer reads back. importColumns
// is the count of import-relevant leading columns (the rest are export-only
// display and ignored on import).
const (
	colID = iota
	colSourceRef
	colStudentName
	colGroup
	colDiscipline
	colSemester
	colControlForm
	importColumns
)

// DebtImporter parses the "Долги" registry sheet of an xlsx workbook into
// ImportedDebt rows. It is lenient on per-row validation (a bad semester
// becomes Semester 0 for the domain to reject downstream) but strict on
// structure (missing sheet, malformed header, corrupt machine-written id).
type DebtImporter struct{}

// NewDebtImporter constructs the importer. It is stateless.
func NewDebtImporter() *DebtImporter { return &DebtImporter{} }

// Import reads the registry sheet and returns one ImportedDebt per
// non-blank data row. A structural problem (unreadable workbook, missing
// registry sheet, header mismatch, unparseable service id) is a
// document-level error; per-row validation (semester range, control-form
// validity, empty identity) is left to ImportDebts/the domain.
func (i *DebtImporter) Import(_ context.Context, r io.Reader) ([]usecases.ImportedDebt, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, fmt.Errorf("student_debts: excel: open: %w", err)
	}
	defer func() { _ = f.Close() }()

	idx, err := f.GetSheetIndex(sheetRegistry)
	if err != nil || idx < 0 {
		return nil, fmt.Errorf("student_debts: excel: registry sheet %q not found", sheetRegistry)
	}
	rows, err := f.GetRows(sheetRegistry)
	if err != nil {
		return nil, fmt.Errorf("student_debts: excel: read rows: %w", err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("student_debts: excel: empty registry sheet")
	}
	if err := validateHeader(rows[0]); err != nil {
		return nil, err
	}

	var out []usecases.ImportedDebt
	for n, row := range rows[1:] {
		if isBlankImportRow(row) {
			continue
		}
		parsed, err := parseRow(row)
		if err != nil {
			return nil, fmt.Errorf("student_debts: excel: row %d: %w", n+2, err)
		}
		out = append(out, parsed)
	}
	return out, nil
}

// validateHeader checks the import-relevant header columns match the
// canonical registry schema, so a structurally wrong file is rejected
// before any row is misread.
func validateHeader(header []string) error {
	for i := range importColumns {
		if at(header, i) != registryHeaders[i] {
			return fmt.Errorf("student_debts: excel: unexpected header column %d: got %q, want %q",
				i+1, at(header, i), registryHeaders[i])
		}
	}
	return nil
}

// parseRow maps one registry data row to an ImportedDebt. Semester defers
// to the domain on a non-numeric value; a populated-but-unparseable service
// id is a document error (a corrupt machine-written id ⇒ invalid export).
func parseRow(row []string) (usecases.ImportedDebt, error) {
	serviceID, err := parseServiceID(at(row, colID))
	if err != nil {
		return usecases.ImportedDebt{}, err
	}
	return usecases.ImportedDebt{
		ServiceID:       serviceID,
		StudentFullName: at(row, colStudentName),
		GroupName:       at(row, colGroup),
		DisciplineName:  at(row, colDiscipline),
		Semester:        parseSemester(at(row, colSemester)),
		ControlForm:     at(row, colControlForm),
		SourceRef:       at(row, colSourceRef),
	}, nil
}

// parseServiceID returns nil for a blank id (a new row), the parsed id for a
// numeric one, or an error for a non-empty unparseable id.
func parseServiceID(s string) (*int64, error) {
	if s == "" {
		return nil, nil
	}
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid service id %q", s)
	}
	return &id, nil
}

// parseSemester returns the parsed semester, or 0 for a non-numeric value so
// the domain's semester ∈ [1,12] invariant rejects it as a per-row error.
func parseSemester(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}

// isBlankImportRow reports whether every import-relevant column is
// empty/whitespace. Display-only columns are ignored, so a row carrying
// only a stray value in (say) the Статус column is treated as blank and
// skipped rather than parsed into an empty-identity row that the domain
// would reject as noise.
func isBlankImportRow(row []string) bool {
	for i := range importColumns {
		if at(row, i) != "" {
			return false
		}
	}
	return true
}

// at returns the trimmed cell at index i, or "" when the row is shorter
// (excelize trims trailing empty cells).
func at(row []string, i int) string {
	if i < 0 || i >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[i])
}
