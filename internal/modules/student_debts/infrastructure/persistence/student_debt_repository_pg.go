package persistence

import (
	"context"
	"errors"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/repositories"
)

// errNotImplemented marks the RED-state stub. The GREEN commit replaces
// every method body with the real SQL implementation.
var errNotImplemented = errors.New("student_debts: not implemented")

// Compile-time assertion that the PG impl satisfies the wide port
// declared in application/usecases (DIP).
var _ usecases.StudentDebtRepository = (*StudentDebtRepositoryPG)(nil)

// StudentDebtRepositoryPG is the SQL implementation of
// StudentDebtRepository. Accepts DBTX (not *sql.DB) so the same struct
// works in single-connection mode and against `*sql.Tx`.
type StudentDebtRepositoryPG struct {
	db DBTX
}

// NewStudentDebtRepositoryPG constructs the repository. db can be
// `*sql.DB` (default DI) or `*sql.Tx` (future transactional paths).
func NewStudentDebtRepositoryPG(db DBTX) *StudentDebtRepositoryPG {
	return &StudentDebtRepositoryPG{db: db}
}

// Save inserts a new StudentDebt aggregate atomically (RED stub).
func (r *StudentDebtRepositoryPG) Save(_ context.Context, _ *entities.StudentDebt) error {
	return errNotImplemented
}

// GetByID returns the StudentDebt aggregate with the given id (RED stub).
func (r *StudentDebtRepositoryPG) GetByID(_ context.Context, _ int64) (*entities.StudentDebt, error) {
	return nil, errNotImplemented
}

// List returns a page of StudentDebt items matching the filter (RED stub).
func (r *StudentDebtRepositoryPG) List(_ context.Context, _ repositories.StudentDebtListFilter) (repositories.StudentDebtListResult, error) {
	return repositories.StudentDebtListResult{}, errNotImplemented
}

// Update writes the mutated aggregate back atomically (RED stub).
func (r *StudentDebtRepositoryPG) Update(_ context.Context, _ *entities.StudentDebt) error {
	return errNotImplemented
}
