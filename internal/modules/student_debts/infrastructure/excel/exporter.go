package excel

import (
	"bytes"
	"context"
	"fmt"
	"strconv"

	"github.com/xuri/excelize/v2"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/entities"
)

// Compile-time assertion that the adapter satisfies the application port.
var _ usecases.DebtExporter = (*DebtExporter)(nil)

// DebtExporter serializes the debt registry into a round-trippable xlsx
// workbook (the "Долги" registry sheet + a "Попытки" attempts sheet).
type DebtExporter struct{}

// NewDebtExporter constructs the exporter. It is stateless.
func NewDebtExporter() *DebtExporter { return &DebtExporter{} }

// Export writes every aggregate into the workbook and returns the xlsx
// bytes. The registry sheet carries one row per debt (service id + source
// ref + identity fields as round-trippable values, plus display columns for
// the latest attempt); the attempts sheet carries one row per resit attempt
// keyed by debt id. The service columns (ID, Источник) are hidden so the
// methodist sees a clean sheet while re-import stays exact.
func (e *DebtExporter) Export(_ context.Context, debts []*entities.StudentDebt) ([]byte, error) {
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	if err := f.SetSheetName("Sheet1", sheetRegistry); err != nil {
		return nil, fmt.Errorf("student_debts: excel: rename sheet: %w", err)
	}
	if _, err := f.NewSheet(sheetAttempts); err != nil {
		return nil, fmt.Errorf("student_debts: excel: new attempts sheet: %w", err)
	}

	if err := writeHeader(f, sheetRegistry, registryHeaders); err != nil {
		return nil, err
	}
	if err := writeHeader(f, sheetAttempts, attemptsHeaders); err != nil {
		return nil, err
	}

	registryRow, attemptRow := 2, 2
	for _, d := range debts {
		if err := writeRegistryRow(f, registryRow, d); err != nil {
			return nil, err
		}
		registryRow++
		for _, a := range d.Attempts() {
			if err := writeAttemptRow(f, attemptRow, d.ID, a); err != nil {
				return nil, err
			}
			attemptRow++
		}
	}

	// Hide the service columns so the registry reads cleanly; values stay
	// present for an exact re-import.
	if err := f.SetColVisible(sheetRegistry, "A:B", false); err != nil {
		return nil, fmt.Errorf("student_debts: excel: hide service columns: %w", err)
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("student_debts: excel: write: %w", err)
	}
	return buf.Bytes(), nil
}

// writeHeader writes the header row (row 1) for a sheet.
func writeHeader(f *excelize.File, sheet string, headers []string) error {
	return setRow(f, sheet, 1, headers)
}

// setRow writes a slice of string values across columns A.. of the given
// 1-based row.
func setRow(f *excelize.File, sheet string, row int, values []string) error {
	for i, v := range values {
		axis, err := excelize.CoordinatesToCellName(i+1, row)
		if err != nil {
			return fmt.Errorf("student_debts: excel: cell name: %w", err)
		}
		if err := f.SetCellValue(sheet, axis, v); err != nil {
			return fmt.Errorf("student_debts: excel: set %s!%s: %w", sheet, axis, err)
		}
	}
	return nil
}

// writeRegistryRow writes one debt's registry row: round-trippable identity
// columns plus display columns derived from the latest attempt.
func writeRegistryRow(f *excelize.File, row int, d *entities.StudentDebt) error {
	var schedule, examiner, result string
	if a := latestAttempt(d); a != nil {
		schedule = a.ScheduledDate().Format(dateLayout)
		examiner = a.Examiner()
		result = resultLabel(a.Result())
	}
	return setRow(f, sheetRegistry, row, []string{
		formatID(d.ID),
		d.SourceRef,
		d.StudentFullName,
		d.GroupName,
		d.DisciplineName,
		strconv.Itoa(d.Semester),
		string(d.ControlForm),
		statusLabel(d.Status()),
		schedule,
		examiner,
		result,
	})
}

// writeAttemptRow writes one resit attempt row keyed by debt id.
func writeAttemptRow(f *excelize.File, row int, debtID int64, a *entities.ResitAttempt) error {
	grade := ""
	if g := a.Grade(); g != nil {
		grade = strconv.Itoa(*g)
	}
	commission := "Нет"
	if a.IsCommission {
		commission = "Да"
	}
	return setRow(f, sheetAttempts, row, []string{
		formatID(debtID),
		strconv.Itoa(a.AttemptNo),
		a.ScheduledDate().Format(dateLayout),
		a.Examiner(),
		resultLabel(a.Result()),
		grade,
		commission,
	})
}

// latestAttempt returns the most recent attempt (the last element, which
// the repository hydrates and the aggregate appends in attempt-no order),
// or nil when the debt has none.
func latestAttempt(d *entities.StudentDebt) *entities.ResitAttempt {
	attempts := d.Attempts()
	if len(attempts) == 0 {
		return nil
	}
	return attempts[len(attempts)-1]
}

// formatID renders a positive id, or "" for an unset (zero) id so a
// round-tripped blank stays blank.
func formatID(id int64) string {
	if id == 0 {
		return ""
	}
	return strconv.FormatInt(id, 10)
}
