package usecases

import (
	"context"
	"fmt"
)

// Import1CDebtsUseCase syncs the academic-debt registry from 1С via a
// DebtSource, applying the same idempotent upsert semantics as the Excel
// import (EDIT_ROLES only). It reuses debtApplier so created/updated/skipped
// behavior is identical across ingestion paths.
type Import1CDebtsUseCase struct {
	applier debtApplier
	source  DebtSource
	audit   AuditSink
}

// NewImport1CDebtsUseCase wires the use case. repo and source are required;
// audit may be nil.
func NewImport1CDebtsUseCase(repo importDebtsRepo, source DebtSource, audit AuditSink) *Import1CDebtsUseCase {
	if repo == nil || source == nil {
		panic("student_debts: NewImport1CDebtsUseCase requires non-nil repo and source")
	}
	return &Import1CDebtsUseCase{applier: debtApplier{repo: repo}, source: source, audit: audit}
}

// Execute fetches debts from 1С and applies every row. A transport/parse
// failure is a hard error; per-row problems are in ImportResult.Errors.
//
// RED stub — replaced by the real implementation in the GREEN commit.
func (uc *Import1CDebtsUseCase) Execute(_ context.Context, _ int64, _ string) (ImportResult, error) {
	return ImportResult{}, fmt.Errorf("not implemented")
}
