package excel_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/infrastructure/excel"
)

// TestExcel_ExportImportRoundTrip is the idempotency guard: every import
// -relevant field a debt carries must survive an export → import cycle
// unchanged, so a methodist can export the registry, edit it and re-import
// without the natural key / service id drifting. Display-only columns are
// not part of the contract and are not checked.
func TestExcel_ExportImportRoundTrip(t *testing.T) {
	withAttempt, err := entities.NewStudentDebt("Иванов Иван", "ИВТ-21", "Базы данных", 3, entities.ControlFormExam)
	if err != nil {
		t.Fatalf("build debt A: %v", err)
	}
	withAttempt.ID = 55
	withAttempt.SourceRef = "ved-7"
	if err := withAttempt.ScheduleResit(time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC), "Петров П.П.", time.Now()); err != nil {
		t.Fatalf("schedule resit: %v", err)
	}

	openNoSource, err := entities.NewStudentDebt("Петров Пётр", "ИВТ-22", "Сети", 4, entities.ControlFormDifferentialZachet)
	if err != nil {
		t.Fatalf("build debt B: %v", err)
	}
	openNoSource.ID = 77

	originals := []*entities.StudentDebt{withAttempt, openNoSource}

	data, err := excel.NewDebtExporter().Export(context.Background(), originals)
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	rows, err := excel.NewDebtImporter().Import(context.Background(), bytes.NewReader(data))
	if err != nil {
		t.Fatalf("import: %v", err)
	}

	if len(rows) != len(originals) {
		t.Fatalf("round-trip row count = %d, want %d", len(rows), len(originals))
	}
	for i, want := range originals {
		got := rows[i]
		if got.ServiceID == nil || *got.ServiceID != want.ID {
			t.Fatalf("row %d ServiceID = %v, want %d", i, got.ServiceID, want.ID)
		}
		if got.StudentFullName != want.StudentFullName ||
			got.GroupName != want.GroupName ||
			got.DisciplineName != want.DisciplineName ||
			got.Semester != want.Semester ||
			got.ControlForm != string(want.ControlForm) ||
			got.SourceRef != want.SourceRef {
			t.Fatalf("row %d round-trip mismatch:\n got  %+v\n want %s/%s/%s/sem %d/%s/ref %q",
				i, got, want.StudentFullName, want.GroupName, want.DisciplineName,
				want.Semester, want.ControlForm, want.SourceRef)
		}
	}
}
