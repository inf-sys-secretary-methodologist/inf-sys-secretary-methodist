package entities

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
)

const maxRevisionSummaryLen = 4096

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

// NewRevision constructs a fresh Revision in status=draft. All field
// invariants surface as ErrInvalidWorkProgram with the offending field
// named. RevisionNumber monotonicity (next = max + 1) is an
// aggregate-level invariant enforced by WorkProgram.AddRevision; the
// constructor only checks positivity.
func NewRevision(in NewRevisionInput) (*Revision, error) {
	trimmedSummary := strings.TrimSpace(in.ChangeSummary)

	if in.WorkProgramID <= 0 {
		return nil, fmt.Errorf("%w: work_program_id must be positive", domain.ErrInvalidWorkProgram)
	}
	if in.RevisionNumber <= 0 {
		return nil, fmt.Errorf("%w: revision_number must be positive", domain.ErrInvalidWorkProgram)
	}
	if !in.ChangeType.IsValid() {
		return nil, fmt.Errorf("%w: change_type %q must be one of hours/semester/literature/assessment/other",
			domain.ErrInvalidWorkProgram, in.ChangeType)
	}
	if trimmedSummary == "" {
		return nil, fmt.Errorf("%w: change_summary is required", domain.ErrInvalidWorkProgram)
	}
	if utf8.RuneCountInString(trimmedSummary) > maxRevisionSummaryLen {
		return nil, fmt.Errorf("%w: change_summary must be <= %d runes",
			domain.ErrInvalidWorkProgram, maxRevisionSummaryLen)
	}
	if in.AuthorID <= 0 {
		return nil, fmt.Errorf("%w: author_id must be positive", domain.ErrInvalidWorkProgram)
	}

	now := time.Now().UTC()
	return &Revision{
		workProgramID:  in.WorkProgramID,
		revisionNumber: in.RevisionNumber,
		changeType:     in.ChangeType,
		changeSummary:  trimmedSummary,
		status:         domain.RevisionStatusDraft,
		authorID:       in.AuthorID,
		diffPayload:    in.DiffPayload,
		createdAt:      now,
		updatedAt:      now,
	}, nil
}

// --- Sub-FSM transitions per ADR-10 ---

// Submit transitions the Revision from draft to pending_approval.
// Author-only operation; caller (use case) handles the role check.
func (r *Revision) Submit() error {
	if r.status != domain.RevisionStatusDraft {
		return domain.ErrInvalidStatusTransition
	}
	r.status = domain.RevisionStatusPendingApproval
	r.updatedAt = time.Now().UTC()
	return nil
}

// Approve transitions the Revision from pending_approval to approved.
// approverID is the acting methodist's ID, recorded for audit /
// Рособрнадзор-trail (chk_wprev_approved_consistency). Transition
// guard runs before the approverID invariant so wrong-status calls
// surface the status error (more informative for the caller).
func (r *Revision) Approve(approverID int64) error {
	if r.status != domain.RevisionStatusPendingApproval {
		return domain.ErrInvalidStatusTransition
	}
	if approverID <= 0 {
		return fmt.Errorf("%w: approver_id must be positive", domain.ErrInvalidWorkProgram)
	}
	now := time.Now().UTC()
	r.status = domain.RevisionStatusApproved
	r.approverID = &approverID
	r.approvedAt = &now
	r.updatedAt = now
	return nil
}

// Reject transitions the Revision from pending_approval to rejected
// with a recorded reason. Reason is trimmed; empty/whitespace-only
// after trim is rejected via ErrRejectReasonRequired. Methodist-only
// per ADR-5. Status guard fires first so wrong-status callers get the
// FSM error rather than the reason error.
func (r *Revision) Reject(reason string) error {
	if r.status != domain.RevisionStatusPendingApproval {
		return domain.ErrInvalidStatusTransition
	}
	trimmed := strings.TrimSpace(reason)
	if trimmed == "" {
		return domain.ErrRejectReasonRequired
	}
	r.status = domain.RevisionStatusRejected
	r.rejectReason = trimmed
	r.updatedAt = time.Now().UTC()
	return nil
}

// ReconstituteRevisionInput collects fields for repository hydration.
type ReconstituteRevisionInput struct {
	ID             int64
	WorkProgramID  int64
	RevisionNumber int
	ChangeType     domain.RevisionChangeType
	ChangeSummary  string
	Status         domain.RevisionStatus
	AuthorID       int64
	ApproverID     *int64
	ApprovedAt     *time.Time
	RejectReason   string
	DiffPayload    []byte
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// ReconstituteRevision builds a Revision from persisted state. Skips
// invariant checks — DB CHECK constraints and the original NewRevision
// call already validated.
func ReconstituteRevision(in ReconstituteRevisionInput) *Revision {
	return &Revision{
		id:             in.ID,
		workProgramID:  in.WorkProgramID,
		revisionNumber: in.RevisionNumber,
		changeType:     in.ChangeType,
		changeSummary:  in.ChangeSummary,
		status:         in.Status,
		authorID:       in.AuthorID,
		approverID:     in.ApproverID,
		approvedAt:     in.ApprovedAt,
		rejectReason:   in.RejectReason,
		diffPayload:    in.DiffPayload,
		createdAt:      in.CreatedAt,
		updatedAt:      in.UpdatedAt,
	}
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
