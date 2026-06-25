// Package excel implements the student_debts DebtExporter / DebtImporter
// ports over an xlsx workbook (github.com/xuri/excelize/v2). The workbook
// is round-trippable: a methodist exports the current registry, edits it
// and imports it back. The adapters live in infrastructure and depend on
// the domain entities and the application ports (allowed dependency
// directions); they are wired in main.go.
package excel

import "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/entities"

// Sheet names for the registry workbook.
const (
	sheetRegistry = "Долги"
	sheetAttempts = "Попытки"
)

// Registry sheet column headers in order. The first two (ID, Источник) are
// the service columns the importer round-trips and the exporter hides; the
// next five are the editable identity fields read on import; the remaining
// columns are export-only display.
var registryHeaders = []string{
	"ID", "Источник", "ФИО студента", "Группа", "Дисциплина",
	"Семестр", "Форма контроля", "Статус", "Срок пересдачи",
	"Преподаватель", "Результат",
}

// Attempts sheet column headers in order (export-only; attempts are written
// by RecordResitResult, never re-imported).
var attemptsHeaders = []string{
	"ID долга", "№ попытки", "Срок", "Экзаменатор", "Результат", "Оценка", "Комиссия",
}

// dateLayout is the unambiguous ISO date used for the scheduled-resit
// columns.
const dateLayout = "2006-01-02"

// statusLabels renders a DebtStatus as a methodist-facing Russian label for
// the export-only Статус column. The wire codes round-trip through the
// identity columns, so these labels need not be machine-readable.
var statusLabels = map[entities.DebtStatus]string{
	entities.DebtStatusOpen:           "Открыт",
	entities.DebtStatusResitScheduled: "Назначена пересдача",
	entities.DebtStatusCommission:     "Комиссия",
	entities.DebtStatusClosedPassed:   "Закрыт (сдал)",
	entities.DebtStatusClosedFailed:   "Закрыт (не сдал)",
}

// resultLabels renders a ResitResult as a Russian label for the export-only
// Результат columns.
var resultLabels = map[entities.ResitResult]string{
	entities.ResitResultPending: "Ожидается",
	entities.ResitResultPassed:  "Сдал",
	entities.ResitResultFailed:  "Не сдал",
	entities.ResitResultNoShow:  "Неявка",
}

// statusLabel / resultLabel fall back to the raw wire code for an
// unrecognized value (defensive — the domain enums are closed).
func statusLabel(s entities.DebtStatus) string {
	if l, ok := statusLabels[s]; ok {
		return l
	}
	return string(s)
}

func resultLabel(r entities.ResitResult) string {
	if l, ok := resultLabels[r]; ok {
		return l
	}
	return string(r)
}
