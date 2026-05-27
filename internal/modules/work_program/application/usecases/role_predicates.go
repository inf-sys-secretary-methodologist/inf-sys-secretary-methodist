// Authorisation predicates for the work_program bounded context.
//
// All role-string comparisons in the WorkProgram use cases go through
// these predicates; the constants come from authDomain.RoleType so a
// typo would fail at compile time on the constant reference, not
// silently at runtime through default-deny.
//
// The four predicates cover the ADR-018 ADR-5 role matrix:
//
//   - isAllowedToCreateWorkProgram → CreateWorkProgramUseCase
//   - isAuthorOrSystemAdmin        → Submit / DiscardDraft (author-scoped)
//   - isApprover                   → Approve / Reject (methodist + admin)
//   - canViewWorkProgram           → Get / List (read-side row-level)
//
// Co-locating them here closes the cohesion smell flagged in the
// v0.178.0 code-review (SHIP 9.50/10) — previously the predicates
// were spread across the four use-case files that introduced them,
// which made the role matrix hard to read as a single decision table.
package usecases

import (
	authDomain "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
)

// isAllowedToCreateWorkProgram encodes the ADR-018 ADR-5 role matrix
// for the create operation. teacher creates own; methodist may create
// as backup ("резервно creates если teacher не успевает");
// system_admin can override. academic_secretary is view-only on РПД
// (curriculum is their author surface); student is denied.
func isAllowedToCreateWorkProgram(role string) bool {
	r := authDomain.RoleType(role)
	return r == authDomain.RoleTeacher ||
		r == authDomain.RoleMethodist ||
		r == authDomain.RoleSystemAdmin
}

// isAuthorOrSystemAdmin is the canonical authorship predicate for
// author-scoped operations (Submit / DiscardDraft). The predicate
// intentionally does NOT accept methodist — methodist's authorship
// rights are bounded to Create per ADR-018 ADR-5; ongoing operations
// on an existing WP belong either to its actual author or to
// system_admin override.
func isAuthorOrSystemAdmin(actorID int64, actorRole string, authorID int64) bool {
	if authDomain.RoleType(actorRole) == authDomain.RoleSystemAdmin {
		return true
	}
	return actorID == authorID
}

// isApprover encodes the ADR-018 ADR-5 approver role set for the
// FSM-advancing operations (Approve / Reject): methodist primary,
// system_admin override.
func isApprover(role string) bool {
	r := authDomain.RoleType(role)
	return r == authDomain.RoleMethodist || r == authDomain.RoleSystemAdmin
}

// canViewWorkProgram encodes the ADR-018 ADR-5 view-rights matrix.
//
//	system_admin / methodist / academic_secretary → see every status
//	teacher                                       → own at any status OR any author's approved
//	student                                       → only approved (273-ФЗ ст. 29 mandatory openness)
//	anything else                                 → denied unconditionally
func canViewWorkProgram(actorID int64, actorRole string, wp *entities.WorkProgram) bool {
	switch authDomain.RoleType(actorRole) {
	case authDomain.RoleSystemAdmin, authDomain.RoleMethodist, authDomain.RoleAcademicSecretary:
		return true
	case authDomain.RoleTeacher:
		return wp.AuthorID() == actorID || wp.Status() == domain.StatusApproved
	case authDomain.RoleStudent:
		return wp.Status() == domain.StatusApproved
	default:
		return false
	}
}
