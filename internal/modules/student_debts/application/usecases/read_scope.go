package usecases

import (
	"context"
	"fmt"

	authDomain "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/repositories"
)

// resolveDebtReadScope applies the registry read-access matrix (design §5)
// to filter for the given actor. It is the single gate shared by every
// staff/teacher read use case (ListDebts, ExportDebts, GetDebtStats) so the
// scoping rule lives in one place rather than being re-derived per use case.
//
// Returns:
//   - scoped: the filter to query the repository with. For staff
//     (admin/methodist/secretary) the inbound filter passes through
//     unchanged; for a teacher DisciplineIDs is forced to the disciplines
//     they own (any client-supplied value is overridden, closing
//     cross-discipline enumeration).
//   - proceed: false when the actor is a teacher owning no disciplines —
//     a statically-empty result. The caller returns an empty response
//     WITHOUT hitting the repository, because an empty DisciplineIDs slice
//     would otherwise disable the predicate and leak the whole registry.
//   - err: ErrDebtAccessForbidden (with a denial audit emitted via action)
//     for students and unknown roles, or a wrapped teacher-scope-resolution
//     error. On any error scoped is the zero filter and proceed is false.
//
// action is the audit action string for the denial event (e.g.
// "student_debts.list_denied") so each use case keeps its own event name.
func resolveDebtReadScope(
	ctx context.Context,
	teacherScope TeacherScopeResolver,
	audit AuditSink,
	action string,
	actorID int64,
	actorRole string,
	filter repositories.StudentDebtListFilter,
) (scoped repositories.StudentDebtListFilter, proceed bool, err error) {
	if isDebtManager(actorRole) {
		return filter, true, nil
	}

	if authDomain.RoleType(actorRole) == authDomain.RoleTeacher {
		ids, err := teacherScope.DisciplineIDsForTeacher(ctx, actorID)
		if err != nil {
			return repositories.StudentDebtListFilter{}, false,
				fmt.Errorf("student_debts: resolve teacher scope: %w", err)
		}
		if len(ids) == 0 {
			return repositories.StudentDebtListFilter{}, false, nil
		}
		filter.DisciplineIDs = ids
		return filter, true, nil
	}

	emitAudit(audit, ctx, action, denialFields(actorID, 0, "forbidden"))
	return repositories.StudentDebtListFilter{}, false,
		fmt.Errorf("%w: role %q cannot read the debt registry", entities.ErrDebtAccessForbidden, actorRole)
}
