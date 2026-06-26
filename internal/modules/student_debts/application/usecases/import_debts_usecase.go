package usecases

import (
	"context"
	"fmt"
	"io"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/entities"
)

// importDebtsRepo is the narrow port the debt-import use cases need: probe
// existing debts (by service id or natural key) and persist new/changed ones.
type importDebtsRepo interface {
	Save(ctx context.Context, debt *entities.StudentDebt) error
	Update(ctx context.Context, debt *entities.StudentDebt) error
	GetByID(ctx context.Context, id int64) (*entities.StudentDebt, error)
	FindByIdentity(ctx context.Context, groupName, studentFullName, disciplineName string, semester int) (*entities.StudentDebt, error)
}

// ImportDebtsUseCase ingests an uploaded registry document (xlsx) into the
// debt registry with idempotent upsert semantics (EDIT_ROLES only). A row
// carrying a service id updates that debt by id; a row without one is matched
// by natural key (group, student, discipline, semester) — found → update
// (skipped when the SourceHash is unchanged), absent → insert. Per-row
// validation/conflict problems are collected into ImportResult.Errors rather
// than aborting the whole import. The shared upsert core lives in debtApplier
// so the 1С sync path reuses identical semantics.
type ImportDebtsUseCase struct {
	applier  debtApplier
	importer DebtImporter
	audit    AuditSink
}

// NewImportDebtsUseCase wires the use case. repo and importer are
// required; audit may be nil.
func NewImportDebtsUseCase(repo importDebtsRepo, importer DebtImporter, audit AuditSink) *ImportDebtsUseCase {
	if repo == nil || importer == nil {
		panic("student_debts: NewImportDebtsUseCase requires non-nil repo and importer")
	}
	return &ImportDebtsUseCase{applier: debtApplier{repo: repo}, importer: importer, audit: audit}
}

// Execute parses the source and applies every row, returning the import
// log. A malformed document is a hard error; per-row problems are in
// ImportResult.Errors.
func (uc *ImportDebtsUseCase) Execute(ctx context.Context, actorID int64, actorRole string, src io.Reader) (ImportResult, error) {
	if !isDebtManager(actorRole) {
		emitAudit(uc.audit, ctx, "student_debts.import_denied", denialFields(actorID, 0, "forbidden"))
		return ImportResult{}, fmt.Errorf("%w: actor %d (role %q) cannot import debts",
			entities.ErrDebtAccessForbidden, actorID, actorRole)
	}

	rows, err := uc.importer.Import(ctx, src)
	if err != nil {
		return ImportResult{}, fmt.Errorf("student_debts: import parse: %w", err)
	}

	result := uc.applier.applyAll(ctx, rows)

	emitAudit(uc.audit, ctx, "student_debts.imported", map[string]any{
		"actor_user_id": actorID,
		"created":       result.Created,
		"updated":       result.Updated,
		"skipped":       result.Skipped,
		"errors":        len(result.Errors),
	})
	return result, nil
}
