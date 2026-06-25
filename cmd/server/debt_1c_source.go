package main

import (
	"context"
	"strings"

	integrationEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/domain/entities"
	sdUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/application/usecases"
	sdEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/entities"
)

// studentDebtCatalog is the slice of the integration OData client the 1С debt
// source needs. Declaring it here (rather than importing the concrete client)
// keeps the adapter unit-testable and the dependency narrow.
type studentDebtCatalog interface {
	GetAllStudentDebts(ctx context.Context) ([]integrationEntities.ODataStudentDebt, error)
}

// debt1CSource adapts the integration OData client to the student_debts
// DebtSource port (a cross-module bridge, hence its place at the DI layer):
// it fetches the 1С academic-debt catalog and maps each ODataStudentDebt onto
// an ImportedDebt — translating the Russian control-form label to the domain
// wire code and carrying the 1С Ref_Key as the SourceRef. Soft-deleted
// (DeletionMark) rows are dropped.
type debt1CSource struct {
	catalog studentDebtCatalog
}

// Fetch implements usecases.DebtSource.
func (s debt1CSource) Fetch(ctx context.Context) ([]sdUsecases.ImportedDebt, error) {
	raw, err := s.catalog.GetAllStudentDebts(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]sdUsecases.ImportedDebt, 0, len(raw))
	for _, d := range raw {
		if d.DeletionMark {
			continue
		}
		out = append(out, sdUsecases.ImportedDebt{
			StudentFullName: d.StudentName,
			GroupName:       d.GroupName,
			DisciplineName:  d.Discipline,
			Semester:        d.Semester,
			ControlForm:     controlFormFrom1C(d.ControlForm),
			SourceRef:       d.RefKey,
		})
	}
	return out, nil
}

// controlFormFrom1C maps a 1С Russian control-form label onto the domain wire
// code. Matching is case-insensitive and tolerates the е/ё spelling of
// «зачёт». An unrecognized label is passed through unchanged so the domain
// rejects it as a per-row error rather than silently mislabeling the debt.
func controlFormFrom1C(label string) string {
	switch normalizeControlFormLabel(label) {
	case "экзамен":
		return string(sdEntities.ControlFormExam)
	case "зачет":
		return string(sdEntities.ControlFormZachet)
	case "дифференцированныйзачет", "диф.зачет", "дифзачет":
		return string(sdEntities.ControlFormDifferentialZachet)
	case "курсовойпроект", "курсоваяработа":
		return string(sdEntities.ControlFormCourseProject)
	default:
		return label
	}
}

// normalizeControlFormLabel lowercases, drops «ё»→«е», and strips spaces so
// label variants ("Дифференцированный зачёт" / "диф. зачет") collapse to one
// matchable key.
func normalizeControlFormLabel(label string) string {
	r := strings.NewReplacer("ё", "е", " ", "", "\t", "")
	return r.Replace(strings.ToLower(strings.TrimSpace(label)))
}
