// Authorisation predicates for the student_debts bounded context.
//
// All role-string comparisons in the use cases go through these
// predicates; the constants come from authDomain.RoleType so a typo
// fails at compile time on the constant reference, not silently at
// runtime through default-deny.
//
// The access matrix (design §5):
//
//	                 | admin | methodist | secretary | teacher        | student
//	read all/by group|  ✅   |    ✅     |    ✅     | ✅ own disc.   |   ❌
//	read own (/my)   |   —   |     —     |     —     |      —         |   ✅
//	import/export,   |  ✅   |    ✅     |    ✅     |      ❌        |   ❌
//	schedule, record |       |           |           |               |
//
// isDebtManager covers the {admin, methodist, secretary} set that has
// both unrestricted read and every write/import right (EDIT_ROLES).
// teacher is scoped by owned disciplines (resolved per request, not a
// pure predicate); student sees only their own debts. Those two cases
// live in the use cases because they need request context (the resolver
// call / the actor id), not a static role check.
package usecases

import (
	authDomain "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
)

// isDebtManager reports whether the role is one of the staff roles with
// unrestricted read of the debt registry and every edit right
// (import/export, schedule resit, record result). Mirrors EDIT_ROLES.
func isDebtManager(role string) bool {
	switch authDomain.RoleType(role) {
	case authDomain.RoleSystemAdmin, authDomain.RoleMethodist, authDomain.RoleAcademicSecretary:
		return true
	default:
		return false
	}
}
