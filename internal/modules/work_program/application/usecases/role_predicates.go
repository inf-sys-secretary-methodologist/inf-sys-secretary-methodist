// Authorisation predicates for the work_program bounded context.
//
// All role-string comparisons in the WorkProgram use cases go through
// these predicates; the constants come from authDomain.RoleType so a
// typo would fail at compile time on the constant reference, not
// silently at runtime through default-deny.
//
// The five predicates cover the ADR-018 ADR-5 role matrix:
//
//   - isAllowedToCreateWorkProgram → CreateWorkProgramUseCase
//   - isAuthorOrSystemAdmin        → Submit / DiscardDraft (author-scoped)
//   - isApprover                   → Approve / Reject (methodist + admin)
//   - canViewWorkProgram           → Get (single-row visibility check)
//   - applyListRoleFilter          → List (query-filter rewrite, role-scoped)
//
// Co-locating them here closes the cohesion smell flagged in the
// v0.178.0 code-review (SHIP 9.50/10) — previously the predicates
// were spread across the four use-case files that introduced them,
// which made the role matrix hard to read as a single decision table.
//
// canViewWorkProgram and applyListRoleFilter are SEPARATE predicates
// because they answer different questions: the former is a yes/no
// row-level gate on a hydrated WP, the latter rewrites a multi-row
// SQL filter before dispatch. They share the same role matrix but
// translate it into different artifacts (boolean vs filter mutation).
package usecases

import (
	"fmt"

	authDomain "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/repositories"
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

// isAllowedToRecordMinobrnaukiOrder encodes the ADR-11 write gate for
// приказы Минобрнауки: methodist records orders (primary), academic
// secretary may record normative documents, system_admin override.
// teacher and student cannot record orders.
func isAllowedToRecordMinobrnaukiOrder(role string) bool {
	r := authDomain.RoleType(role)
	return r == authDomain.RoleMethodist ||
		r == authDomain.RoleAcademicSecretary ||
		r == authDomain.RoleSystemAdmin
}

// isAllowedToViewMinobrnaukiOrders encodes the ADR-11 read gate: every
// non-student staff role may view orders (teachers need to see orders
// affecting their disciplines). Students have no business reason to read
// internal regulatory-tracking artifacts, so they are denied. Orders are
// not author-scoped, so this is a flat role check (no row-level filter).
func isAllowedToViewMinobrnaukiOrders(role string) bool {
	switch authDomain.RoleType(role) {
	case authDomain.RoleSystemAdmin, authDomain.RoleMethodist,
		authDomain.RoleAcademicSecretary, authDomain.RoleTeacher:
		return true
	default:
		return false
	}
}

// canViewWorkProgram encodes the ADR-018 ADR-5 view-rights matrix
// for the single-row Get path.
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

// applyListRoleFilter rewrites the inbound List filter in place to
// enforce row-level access policy for the actor's role. Same matrix
// as canViewWorkProgram but expressed as filter mutations rather than
// a yes/no gate so the SQL layer can use cohort indexes without
// over-fetching and post-filtering.
//
//	system_admin / methodist / academic_secretary → pass-through
//	teacher                                       → AuthorID forced to actor id (closes
//	                                                "list other teachers' drafts" enumeration;
//	                                                approved WPs from other authors reachable
//	                                                via Get deep link)
//	student                                       → Status forced to approved
//	anything else                                 → ErrWorkProgramScopeForbidden (caller
//	                                                emits the audit event so reason can carry
//	                                                use-case-specific context)
//
// The predicate intentionally returns the sentinel rather than a bool
// so the caller cannot accidentally drop the deny branch — the only
// way to ignore the error is explicit, which surfaces at code review.
func applyListRoleFilter(actorID int64, actorRole string, filter *repositories.WorkProgramListFilter) error {
	switch authDomain.RoleType(actorRole) {
	case authDomain.RoleSystemAdmin, authDomain.RoleMethodist, authDomain.RoleAcademicSecretary:
		// Pass-through — these roles see every WP.
		return nil
	case authDomain.RoleTeacher:
		actor := actorID
		filter.AuthorID = &actor
		return nil
	case authDomain.RoleStudent:
		approved := domain.StatusApproved
		filter.Status = &approved
		return nil
	default:
		return fmt.Errorf("%w: role %q cannot list work programs",
			domain.ErrWorkProgramScopeForbidden, actorRole)
	}
}
