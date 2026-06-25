package main

import (
	"context"

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
//
// RED stub — replaced by the real implementation in the GREEN commit.
func (s debt1CSource) Fetch(_ context.Context) ([]sdUsecases.ImportedDebt, error) {
	return nil, nil
}

// controlFormFrom1C maps a 1С Russian control-form label onto the domain wire
// code. An unrecognized label is passed through unchanged so the domain
// rejects it as a per-row error rather than silently mislabeling the debt.
//
// RED stub — replaced by the real implementation in the GREEN commit.
func controlFormFrom1C(label string) string {
	_ = sdEntities.ControlFormExam
	return label
}
