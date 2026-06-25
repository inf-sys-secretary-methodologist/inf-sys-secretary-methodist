package usecases

import (
	"context"
	"errors"
	"io"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/entities"
)

// errNotImplemented marks the RED-state stub in this slice.
var errNotImplemented = errors.New("student_debts: not implemented")

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
	return ImportResult{}, errNotImplemented
}
