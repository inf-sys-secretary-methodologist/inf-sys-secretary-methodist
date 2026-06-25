package usecases

import (
	"context"
	"errors"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/entities"
)

// errNotImplemented marks the RED-state stubs in this package. The GREEN
// commit replaces every stubbed Execute with the real orchestration.
var errNotImplemented = errors.New("student_debts: not implemented")

// getDebtRepo is the narrow read port GetDebt needs.
type getDebtRepo interface {
	GetByID(ctx context.Context, id int64) (*entities.StudentDebt, error)
}

// GetDebtUseCase hydrates and authorizes a single-debt read. Staff
// (admin/methodist/secretary) see any debt; a teacher sees a debt only
// when its discipline is one they own; a student sees only their own
// debt. Denied reads return ErrDebtAccessForbidden (handlers collapse to
// 404 per OWASP IDOR).
type GetDebtUseCase struct {
	repo         getDebtRepo
	teacherScope TeacherScopeResolver
	audit        AuditSink
}

// NewGetDebtUseCase wires the use case. repo and teacherScope are
// required; audit may be nil (no-op).
func NewGetDebtUseCase(repo getDebtRepo, teacherScope TeacherScopeResolver, audit AuditSink) *GetDebtUseCase {
	if repo == nil || teacherScope == nil {
		panic("student_debts: NewGetDebtUseCase requires non-nil repo and teacherScope")
	}
	return &GetDebtUseCase{repo: repo, teacherScope: teacherScope, audit: audit}
}

// Execute loads the debt and authorizes the read per the access matrix.
func (uc *GetDebtUseCase) Execute(ctx context.Context, actorID int64, actorRole string, id int64) (*entities.StudentDebt, error) {
	return nil, errNotImplemented
}
