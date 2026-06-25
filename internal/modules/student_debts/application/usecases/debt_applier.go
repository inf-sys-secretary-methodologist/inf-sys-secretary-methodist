package usecases

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/repositories"
)

// debtApplier is the idempotent upsert core shared by the debt-import use
// cases (Excel upload and 1С OData sync). Given parsed source rows it
// creates new debts, updates changed ones, skips unchanged ones (by
// SourceHash), and collects per-row validation/conflict/persistence problems
// into the result rather than aborting the batch. Keeping it source-agnostic
// lets both ingestion paths reuse identical upsert semantics.
type debtApplier struct{ repo importDebtsRepo }

// applyAll upserts every row, collecting per-row failures as ImportResult
// errors (1-based row index + human-readable identity).
func (a debtApplier) applyAll(ctx context.Context, rows []ImportedDebt) ImportResult {
	var result ImportResult
	for i, row := range rows {
		if err := a.applyRow(ctx, row, &result); err != nil {
			result.Errors = append(result.Errors, ImportRowError{
				Row: i + 1, Identity: identityLabel(row), Message: err.Error(),
			})
		}
	}
	return result
}

// applyRow upserts a single source row, advancing the result counters.
// Returns an error (collected as a row error by the caller) for any
// validation/conflict/persistence problem.
func (a debtApplier) applyRow(ctx context.Context, row ImportedDebt, result *ImportResult) error {
	existing, isNew, err := a.resolve(ctx, row)
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
		if err := a.repo.Save(ctx, debt); err != nil {
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
	if err := a.repo.Update(ctx, existing); err != nil {
		return err
	}
	result.Updated++
	return nil
}

// resolve finds the debt a row targets: by service id (must exist — a
// dangling id is a row error) or by natural key (absent → a new row).
func (a debtApplier) resolve(ctx context.Context, row ImportedDebt) (existing *entities.StudentDebt, isNew bool, err error) {
	if row.ServiceID != nil {
		d, err := a.repo.GetByID(ctx, *row.ServiceID)
		if errors.Is(err, repositories.ErrStudentDebtNotFound) {
			return nil, false, fmt.Errorf("service id %d not found", *row.ServiceID)
		}
		if err != nil {
			return nil, false, err
		}
		return d, false, nil
	}
	d, err := a.repo.FindByIdentity(ctx, row.GroupName, row.StudentFullName, row.DisciplineName, row.Semester)
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
