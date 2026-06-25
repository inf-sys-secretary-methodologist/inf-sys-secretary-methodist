package excel_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/xuri/excelize/v2"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/infrastructure/excel"
)

// registryHeaderRow mirrors the first seven (import-relevant) registry
// headers the importer validates.
var registryHeaderRow = []string{"ID", "Источник", "ФИО студента", "Группа", "Дисциплина", "Семестр", "Форма контроля"}

// buildWorkbook writes a single sheet with the given rows and returns the
// xlsx bytes. sheet="" defaults to the registry sheet name.
func buildWorkbook(t *testing.T, sheet string, rows [][]string) []byte {
	t.Helper()
	if sheet == "" {
		sheet = "Долги"
	}
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()
	if err := f.SetSheetName("Sheet1", sheet); err != nil {
		t.Fatalf("set sheet name: %v", err)
	}
	for r, row := range rows {
		for c, v := range row {
			axis, _ := excelize.CoordinatesToCellName(c+1, r+1)
			if err := f.SetCellValue(sheet, axis, v); err != nil {
				t.Fatalf("set cell: %v", err)
			}
		}
	}
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatalf("write workbook: %v", err)
	}
	return buf.Bytes()
}

func TestDebtImporter_ParsesRowsWithAndWithoutServiceID(t *testing.T) {
	data := buildWorkbook(t, "", [][]string{
		registryHeaderRow,
		{"55", "ved-7", "Иванов Иван", "ИВТ-21", "Базы данных", "3", "exam"},
		{"", "", "Петров Пётр", "ИВТ-21", "Сети", "4", "zachet"},
	})

	rows, err := excel.NewDebtImporter().Import(context.Background(), bytes.NewReader(data))
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}

	first := rows[0]
	if first.ServiceID == nil || *first.ServiceID != 55 {
		t.Fatalf("row 1 ServiceID = %v, want 55", first.ServiceID)
	}
	if first.StudentFullName != "Иванов Иван" || first.GroupName != "ИВТ-21" ||
		first.DisciplineName != "Базы данных" || first.Semester != 3 ||
		first.ControlForm != "exam" || first.SourceRef != "ved-7" {
		t.Fatalf("row 1 parsed wrong: %+v", first)
	}

	second := rows[1]
	if second.ServiceID != nil {
		t.Fatalf("row 2 ServiceID = %v, want nil (blank id → new row)", second.ServiceID)
	}
	if second.StudentFullName != "Петров Пётр" || second.Semester != 4 || second.ControlForm != "zachet" {
		t.Fatalf("row 2 parsed wrong: %+v", second)
	}
}

func TestDebtImporter_SkipsBlankTrailingRows(t *testing.T) {
	data := buildWorkbook(t, "", [][]string{
		registryHeaderRow,
		{"55", "", "Иванов Иван", "ИВТ-21", "Базы данных", "3", "exam"},
		{"", "", "", "", "", "", ""},
	})

	rows, err := excel.NewDebtImporter().Import(context.Background(), bytes.NewReader(data))
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected blank trailing row skipped (1 row), got %d", len(rows))
	}
}

func TestDebtImporter_NonNumericSemesterDefersToDomain(t *testing.T) {
	// A bad semester is a per-row validation problem: the importer yields
	// Semester 0 and lets ImportDebts/domain reject it, rather than failing
	// the whole document.
	data := buildWorkbook(t, "", [][]string{
		registryHeaderRow,
		{"", "", "Иванов Иван", "ИВТ-21", "Базы данных", "третий", "exam"},
	})

	rows, err := excel.NewDebtImporter().Import(context.Background(), bytes.NewReader(data))
	if err != nil {
		t.Fatalf("import must not fail the document for a bad semester: %v", err)
	}
	if len(rows) != 1 || rows[0].Semester != 0 {
		t.Fatalf("expected one row with Semester 0, got %+v", rows)
	}
}

func TestDebtImporter_HeaderRowShorterThanExpectedIsRejected(t *testing.T) {
	data := buildWorkbook(t, "", [][]string{
		{"ID", "Источник", "ФИО студента"}, // truncated header
		{"55", "", "Иванов Иван", "ИВТ-21", "Базы данных", "3", "exam"},
	})

	_, err := excel.NewDebtImporter().Import(context.Background(), bytes.NewReader(data))
	if err == nil {
		t.Fatal("expected a parse error for a malformed header")
	}
}

func TestDebtImporter_MissingRegistrySheetIsRejected(t *testing.T) {
	data := buildWorkbook(t, "НеТот", [][]string{registryHeaderRow})

	_, err := excel.NewDebtImporter().Import(context.Background(), bytes.NewReader(data))
	if err == nil {
		t.Fatal("expected a parse error when the registry sheet is absent")
	}
}

func TestDebtImporter_UnparseableServiceIDIsRejected(t *testing.T) {
	// A corrupt machine-written id means the file is not a valid export.
	data := buildWorkbook(t, "", [][]string{
		registryHeaderRow,
		{"не-число", "", "Иванов Иван", "ИВТ-21", "Базы данных", "3", "exam"},
	})

	_, err := excel.NewDebtImporter().Import(context.Background(), bytes.NewReader(data))
	if err == nil {
		t.Fatal("expected a parse error for an unparseable service id")
	}
}
