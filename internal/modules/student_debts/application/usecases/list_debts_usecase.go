package usecases

import (
	"context"
	"fmt"

	authDomain "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/repositories"
)

// listDebtsRepo is the narrow read port ListDebts needs.
type listDebtsRepo interface {
	List(ctx context.Context, filter repositories.StudentDebtListFilter) (repositories.StudentDebtListResult, error)
}

// ListDebtsUseCase lists the debt registry for staff and teachers.
// Staff (admin/methodist/secretary) see every debt subject to the
// inbound filter; a teacher's view is forced to the disciplines they own
// (any client-supplied DisciplineIDs are overridden, closing
// cross-discipline enumeration). Students are denied here — they use
// ListMyDebtsUseCase via the /my endpoint.
type ListDebtsUseCase struct {
	repo         listDebtsRepo
	teacherScope TeacherScopeResolver
	audit        AuditSink
}

// NewListDebtsUseCase wires the use case. repo and teacherScope are
// required; audit may be nil.
func NewListDebtsUseCase(repo listDebtsRepo, teacherScope TeacherScopeResolver, audit AuditSink) *ListDebtsUseCase {
	if repo == nil || teacherScope == nil {
		panic("student_debts: NewListDebtsUseCase requires non-nil repo and teacherScope")
	}
	return &ListDebtsUseCase{repo: repo, teacherScope: teacherScope, audit: audit}
}

// Execute applies the role-scoped filter and lists matching debts:
//   - staff (admin/methodist/secretary) → inbound filter pass-through;
//   - teacher → DisciplineIDs forced to the disciplines they own (any
//     client value overridden). A teacher who owns no disciplines gets
//     an empty page WITHOUT hitting the repo — an empty DisciplineIDs
//     would otherwise disable the predicate and leak the whole registry;
//   - anyone else (student, unknown) → denied + audit; students read
//     their own debts through ListMyDebtsUseCase.
func (uc *ListDebtsUseCase) Execute(ctx context.Context, actorID int64, actorRole string, filter repositories.StudentDebtListFilter) (repositories.StudentDebtListResult, error) {
	if isDebtManager(actorRole) {
		return uc.repo.List(ctx, filter)
	}

	if authDomain.RoleType(actorRole) == authDomain.RoleTeacher {
		ids, err := uc.teacherScope.DisciplineIDsForTeacher(ctx, actorID)
		if err != nil {
			return repositories.StudentDebtListResult{}, fmt.Errorf("student_debts: resolve teacher scope: %w", err)
		}
		if len(ids) == 0 {
			return repositories.StudentDebtListResult{}, nil
		}
		filter.DisciplineIDs = ids
		return uc.repo.List(ctx, filter)
	}

	emitAudit(uc.audit, ctx, "student_debts.list_denied", denialFields(actorID, 0, "forbidden"))
	return repositories.StudentDebtListResult{}, fmt.Errorf("%w: role %q cannot list the debt registry",
		entities.ErrDebtAccessForbidden, actorRole)
}
