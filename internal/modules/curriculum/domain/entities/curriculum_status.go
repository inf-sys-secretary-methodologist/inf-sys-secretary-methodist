// Package entities contains the domain entities for the curriculum
// bounded context: the Curriculum aggregate root (учебный план) and the
// CurriculumStatus typed enum.
package entities

// CurriculumStatus is the typed enum mirroring the SQL CHECK on
// curricula.status (chk_curricula_status_enum, migration 031). It exists
// so domain code never passes a "magic string" status around — the type
// system catches typos at compile time.
//
// Lifecycle (transitions land in v0.117.0):
//
//	draft ─SubmitForApproval→ pending_approval ─Approve→ approved ─Archive→ archived
//	                                            ─Reject──→ draft
type CurriculumStatus string

// Recognised statuses. The string literals match
// chk_curricula_status_enum from migration 031 byte-for-byte —
// TestCurriculumStatus_StringMatchesDBLiteral pins the parity.
const (
	StatusDraft           CurriculumStatus = "draft"
	StatusPendingApproval CurriculumStatus = "pending_approval"
	StatusApproved        CurriculumStatus = "approved"
	StatusArchived        CurriculumStatus = "archived"
)

// IsValid reports whether s is one of the recognised statuses.
// Repository implementations call this on Reconstitute paths so a row
// that somehow holds an unknown status (a future migration adding a
// status without releasing the matching domain constant, for example)
// is rejected before it leaks into the use-case layer.
func (s CurriculumStatus) IsValid() bool {
	switch s {
	case StatusDraft, StatusPendingApproval, StatusApproved, StatusArchived:
		return true
	default:
		return false
	}
}

// CanEdit reports whether a curriculum in this status may be modified
// via UpdateBasics. Only draft curricula are editable; the other states
// are either awaiting approval, already approved (and therefore frozen
// by policy), or archived.
func (s CurriculumStatus) CanEdit() bool {
	return s == StatusDraft
}

// IsApproved reports whether the curriculum has reached the approved
// terminal state.
func (s CurriculumStatus) IsApproved() bool {
	return s == StatusApproved
}
