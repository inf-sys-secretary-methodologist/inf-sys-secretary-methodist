package usecases

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/repositories"
)

// importDebtsRepo is the narrow port ImportDebts needs: probe existing
// debts (by service id or natural key) and persist new/changed ones.
type importDebtsRepo interface {
	Save(ctx context.Context, debt *entities.StudentDebt) error
	Update(ctx context.Context, debt *entities.StudentDebt) error
	GetByID(ctx context.Context, id int64) (*entities.StudentDebt, error)
	FindByIdentity(ctx context.Context, groupName, studentFullName, disciplineName string, semester int) (*entities.StudentDebt, error)
}

// ImportDebtsUseCase ingests a registry document (xlsx now, 1С later)
// into the debt registry with idempotent upsert semantics (EDIT_ROLES
// only). A row carrying a service id updates that debt by id; a row
// without one is matched by natural key (group, student, discipline,
// semester) — found → update (skipped when the SourceHash is unchanged),
// absent → insert. Per-row validation/conflict problems are collected
// into ImportResult.Errors rather than aborting the whole import.
type ImportDebtsUseCase struct {
	repo     importDebtsRepo
	importer DebtImporter
	audit    AuditSink
}

// NewImportDebtsUseCase wires the use case. repo and importer are
// required; audit may be nil.
func NewImportDebtsUseCase(repo importDebtsRepo, importer DebtImporter, audit AuditSink) *ImportDebtsUseCase {
	if repo == nil || importer == nil {
		panic("student_debts: NewImportDebtsUseCase requires non-nil repo and importer")
	}
	return &ImportDebtsUseCase{repo: repo, importer: importer, audit: audit}
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

	var result ImportResult
	for i, row := range rows {
		if err := uc.applyRow(ctx, row, &result); err != nil {
			result.Errors = append(result.Errors, ImportRowError{
				Row: i + 1, Identity: identityLabel(row), Message: err.Error(),
			})
		}
	}

	emitAudit(uc.audit, ctx, "student_debts.imported", map[string]any{
		"actor_user_id": actorID,
		"created":       result.Created,
		"updated":       result.Updated,
		"skipped":       result.Skipped,
		"errors":        len(result.Errors),
	})
	return result, nil
}

// applyRow upserts a single source row, advancing the result counters.
// Returns an error (collected as a row error by the caller) for any
// validation/conflict/persistence problem.
func (uc *ImportDebtsUseCase) applyRow(ctx context.Context, row ImportedDebt, result *ImportResult) error {
	existing, isNew, err := uc.resolve(ctx, row)
	if err != nil {
		return err
	}

	hash := sourceHash(row)
	if isNew {
		debt, err := entities.NewStudentDebt(row.StudentFullName, row.GroupName, row.DisciplineName, row.Semester, entities.ControlForm(row.ControlForm))
		if err != nil {
			return err
		}
		debt.SourceRef = row.SourceRef
		debt.SourceHash = hash
		if err := uc.repo.Save(ctx, debt); err != nil {
			return err
		}
		result.Created++
		return nil
	}

	if existing.SourceHash == hash {
		result.Skipped++
		return nil
	}
	if err := existing.UpdateSourceFields(row.StudentFullName, row.GroupName, row.DisciplineName, row.Semester, entities.ControlForm(row.ControlForm)); err != nil {
		return err
	}
	existing.SourceRef = row.SourceRef
	existing.SourceHash = hash
	if err := uc.repo.Update(ctx, existing); err != nil {
		return err
	}
	result.Updated++
	return nil
}

// resolve finds the debt a row targets: by service id (must exist — a
// dangling id is a row error) or by natural key (absent → a new row).
func (uc *ImportDebtsUseCase) resolve(ctx context.Context, row ImportedDebt) (existing *entities.StudentDebt, isNew bool, err error) {
	if row.ServiceID != nil {
		d, err := uc.repo.GetByID(ctx, *row.ServiceID)
		if errors.Is(err, repositories.ErrStudentDebtNotFound) {
			return nil, false, fmt.Errorf("service id %d not found", *row.ServiceID)
		}
		if err != nil {
			return nil, false, err
		}
		return d, false, nil
	}
	d, err := uc.repo.FindByIdentity(ctx, row.GroupName, row.StudentFullName, row.DisciplineName, row.Semester)
	if errors.Is(err, repositories.ErrStudentDebtNotFound) {
		return nil, true, nil
	}
	if err != nil {
		return nil, false, err
	}
	return d, false, nil
}

// sourceHash is the SHA-256 over a row's denormalized content, used to
// skip unchanged rows on re-import (idempotency). Fields are joined with
// a unit separator so distinct field boundaries cannot collide.
func sourceHash(row ImportedDebt) string {
	payload := fmt.Sprintf("%s\x1f%s\x1f%s\x1f%d\x1f%s\x1f%s",
		row.StudentFullName, row.GroupName, row.DisciplineName, row.Semester, row.ControlForm, row.SourceRef)
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}

// identityLabel renders a row's natural key for a human-readable row error.
func identityLabel(row ImportedDebt) string {
	return fmt.Sprintf("%s / %s / %s / сем. %d", row.GroupName, row.StudentFullName, row.DisciplineName, row.Semester)
}
