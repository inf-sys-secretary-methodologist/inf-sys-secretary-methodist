package entities

import (
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
)

// NewRevisionInput collects constructor parameters for a Revision
// (лист актуализации) per ADR-10. RevisionNumber monotonicity is
// enforced at the aggregate level (WorkProgram.AddRevision) — the
// constructor only validates positivity.
type NewRevisionInput struct {
	WorkProgramID  int64
	RevisionNumber int
	ChangeType     domain.RevisionChangeType
	ChangeSummary  string
	AuthorID       int64
	DiffPayload    []byte // optional, raw JSON; PG validates JSONB at repo layer
}

// Revision — лист актуализации, минор-изменение к РПД без полного
// reapproval per ADR-10. Independent sub-FSM (draft / pending_approval
// / approved / rejected) — parent WorkProgram remains approved while
// a revision is in flight. Inner aggregate of WorkProgram per ADR-1.
type Revision struct {
	id             int64
	workProgramID  int64
	revisionNumber int
	changeType     domain.RevisionChangeType
	changeSummary  string
	status         domain.RevisionStatus
	authorID       int64
	approverID     *int64
	approvedAt     *time.Time
	rejectReason   string
	diffPayload    []byte
	createdAt      time.Time
	updatedAt      time.Time
}

// NewRevision stub for PR 1c RED phase — real invariants land in the
// GREEN commit.
func NewRevision(_ NewRevisionInput) (*Revision, error) {
	return nil, domain.ErrInvalidWorkProgram
}

// ID returns the persistent identifier.
func (r *Revision) ID() int64 { return r.id }

// WorkProgramID returns the parent aggregate identifier.
func (r *Revision) WorkProgramID() int64 { return r.workProgramID }

// RevisionNumber returns the monotonic revision counter (1, 2, 3, ...).
func (r *Revision) RevisionNumber() int { return r.revisionNumber }

// ChangeType returns the categorized nature of the change.
func (r *Revision) ChangeType() domain.RevisionChangeType { return r.changeType }

// ChangeSummary returns the human-readable summary (trimmed, ≤ 4096 runes).
func (r *Revision) ChangeSummary() string { return r.changeSummary }

// Status returns the current sub-FSM state.
func (r *Revision) Status() domain.RevisionStatus { return r.status }

// AuthorID returns the methodist/teacher who proposed the revision.
func (r *Revision) AuthorID() int64 { return r.authorID }

// ApproverID returns the methodist who approved this revision, or nil if not yet approved.
func (r *Revision) ApproverID() *int64 { return r.approverID }

// ApprovedAt returns the approval timestamp, or nil if not approved.
func (r *Revision) ApprovedAt() *time.Time { return r.approvedAt }

// RejectReason returns the methodist's rejection rationale (empty unless rejected).
func (r *Revision) RejectReason() string { return r.rejectReason }

// DiffPayload returns the optional structured before/after JSON.
func (r *Revision) DiffPayload() []byte { return r.diffPayload }

// CreatedAt returns the creation timestamp.
func (r *Revision) CreatedAt() time.Time { return r.createdAt }

// UpdatedAt returns the last mutation timestamp.
func (r *Revision) UpdatedAt() time.Time { return r.updatedAt }
