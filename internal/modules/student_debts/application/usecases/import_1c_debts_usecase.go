package usecases

import (
	"context"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/entities"
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
func (uc *Import1CDebtsUseCase) Execute(ctx context.Context, actorID int64, actorRole string) (ImportResult, error) {
	if !isDebtManager(actorRole) {
		emitAudit(uc.audit, ctx, "student_debts.import_1c_denied", denialFields(actorID, 0, "forbidden"))
		return ImportResult{}, fmt.Errorf("%w: actor %d (role %q) cannot import debts from 1С",
			entities.ErrDebtAccessForbidden, actorID, actorRole)
	}

	rows, err := uc.source.Fetch(ctx)
	if err != nil {
		return ImportResult{}, fmt.Errorf("student_debts: 1С fetch: %w", err)
	}

	result := uc.applier.applyAll(ctx, rows)

	emitAudit(uc.audit, ctx, "student_debts.imported_1c", map[string]any{
		"actor_user_id": actorID,
		"created":       result.Created,
		"updated":       result.Updated,
		"skipped":       result.Skipped,
		"errors":        len(result.Errors),
	})
	return result, nil
}
