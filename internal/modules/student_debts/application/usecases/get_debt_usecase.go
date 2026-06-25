package usecases

import (
	"context"
	"fmt"
	"slices"

	authDomain "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/entities"
)

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
// Repository errors (ErrStudentDebtNotFound, transport) propagate without
// an audit event — reads audit only on denial, so ID-typo not-found noise
// does not flood the log. Denied reads emit "student_debts.view_denied"
// and return ErrDebtAccessForbidden.
func (uc *GetDebtUseCase) Execute(ctx context.Context, actorID int64, actorRole string, id int64) (*entities.StudentDebt, error) {
	debt, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	allowed, err := uc.canView(ctx, actorID, actorRole, debt)
	if err != nil {
		return nil, err
	}
	if !allowed {
		emitAudit(uc.audit, ctx, "student_debts.view_denied", denialFields(actorID, id, "forbidden"))
		return nil, fmt.Errorf("%w: actor %d (role %q) cannot view debt %d",
			entities.ErrDebtAccessForbidden, actorID, actorRole, id)
	}
	return debt, nil
}

// canView evaluates the per-row access matrix. Staff see every debt; a
// teacher sees a debt only when its discipline is one they own (resolved
// per request); a student sees only debts whose student_user_id is their
// own. A resolver failure is returned as an error (infra), distinct from
// a plain "false" denial.
func (uc *GetDebtUseCase) canView(ctx context.Context, actorID int64, actorRole string, debt *entities.StudentDebt) (bool, error) {
	if isDebtManager(actorRole) {
		return true, nil
	}
	switch authDomain.RoleType(actorRole) {
	case authDomain.RoleTeacher:
		ids, err := uc.teacherScope.DisciplineIDsForTeacher(ctx, actorID)
		if err != nil {
			return false, fmt.Errorf("student_debts: resolve teacher scope: %w", err)
		}
		return debt.DisciplineID != nil && slices.Contains(ids, *debt.DisciplineID), nil
	case authDomain.RoleStudent:
		return debt.StudentUserID != nil && *debt.StudentUserID == actorID, nil
	default:
		return false, nil
	}
}
