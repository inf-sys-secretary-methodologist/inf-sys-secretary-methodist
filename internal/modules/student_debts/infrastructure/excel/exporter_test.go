package excel_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/infrastructure/excel"
)

// sampleScheduledDebt builds an open debt moved to resit_scheduled with one
// pending attempt — enough state to exercise both sheets of the export.
func sampleScheduledDebt(t *testing.T) *entities.StudentDebt {
	t.Helper()
	d, err := entities.NewStudentDebt("Иванов Иван", "ИВТ-21", "Базы данных", 3, entities.ControlFormExam)
	if err != nil {
		t.Fatalf("build debt: %v", err)
	}
	d.ID = 55
	d.SourceRef = "ved-7"
	if err := d.ScheduleResit(time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC), "Петров П.П.", time.Now()); err != nil {
		t.Fatalf("schedule resit: %v", err)
	}
	return d
}

func openExport(t *testing.T, data []byte) *excelize.File {
	t.Helper()
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("open exported workbook: %v", err)
	}
	t.Cleanup(func() { _ = f.Close() })
	return f
}

func cell(t *testing.T, f *excelize.File, sheet, axis string) string {
	t.Helper()
	v, err := f.GetCellValue(sheet, axis)
	if err != nil {
		t.Fatalf("get %s!%s: %v", sheet, axis, err)
	}
	return v
}

func TestDebtExporter_WritesRegistryAndAttemptsSheets(t *testing.T) {
	exporter := excel.NewDebtExporter()
	data, err := exporter.Export(context.Background(), []*entities.StudentDebt{sampleScheduledDebt(t)})
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	f := openExport(t, data)

	sheets := f.GetSheetList()
	if len(sheets) != 2 || sheets[0] != "Долги" || sheets[1] != "Попытки" {
		t.Fatalf("expected sheets [Долги Попытки], got %v", sheets)
	}

	// Registry header row + the import-critical data cells (round-trip
	// fidelity): ID, source ref, identity fields, control form as the wire
	// code. Display columns (status/срок/преподаватель/результат) are not
	// pinned to exact wording here — the round-trip test guards what matters.
	if got := cell(t, f, "Долги", "A1"); got != "ID" {
		t.Fatalf("A1 header = %q, want ID", got)
	}
	want := map[string]string{
		"A2": "55", "B2": "ved-7", "C2": "Иванов Иван", "D2": "ИВТ-21",
		"E2": "Базы данных", "F2": "3", "G2": "exam",
	}
	for axis, exp := range want {
		if got := cell(t, f, "Долги", axis); got != exp {
			t.Fatalf("Долги!%s = %q, want %q", axis, got, exp)
		}
	}
	if got := cell(t, f, "Долги", "I2"); got != "2026-07-01" {
		t.Fatalf("Долги!I2 (срок пересдачи) = %q, want 2026-07-01", got)
	}
	if got := cell(t, f, "Долги", "J2"); got != "Петров П.П." {
		t.Fatalf("Долги!J2 (преподаватель) = %q, want Петров П.П.", got)
	}

	// Attempts sheet: the one pending attempt, keyed by debt id.
	if got := cell(t, f, "Попытки", "A2"); got != "55" {
		t.Fatalf("Попытки!A2 (id долга) = %q, want 55", got)
	}
	if got := cell(t, f, "Попытки", "B2"); got != "1" {
		t.Fatalf("Попытки!B2 (№ попытки) = %q, want 1", got)
	}
	if got := cell(t, f, "Попытки", "D2"); got != "Петров П.П." {
		t.Fatalf("Попытки!D2 (экзаменатор) = %q, want Петров П.П.", got)
	}
	if got := cell(t, f, "Попытки", "G2"); got != "Нет" {
		t.Fatalf("Попытки!G2 (комиссия) = %q, want Нет", got)
	}
}

func TestDebtExporter_EmptyRegistryIsHeaderOnly(t *testing.T) {
	exporter := excel.NewDebtExporter()
	data, err := exporter.Export(context.Background(), nil)
	if err != nil {
		t.Fatalf("export empty: %v", err)
	}
	f := openExport(t, data)

	if got := cell(t, f, "Долги", "A1"); got != "ID" {
		t.Fatalf("empty export must still carry the header, A1 = %q", got)
	}
	if got := cell(t, f, "Долги", "A2"); got != "" {
		t.Fatalf("empty export must have no data rows, A2 = %q", got)
	}
}
