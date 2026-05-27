// Package entities holds the WorkProgram aggregate root and its inner
// entities (Goal, Competence, Topic, AssessmentCriterion, Reference,
// Revision). See docs/plans/2026-05-27-work-program-initiative.md for
// ADR rationale.
package entities

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
)

// Invariant bounds. Mirrored by migration 047 CHECK constraints
// (defense in depth) — single source of truth lives here.
const (
	minApplicableYear = 2000
	maxApplicableYear = 2100
	maxAnnotationLen  = 8192
)

// NewWorkProgramInput carries the constructor parameters for a fresh
// WorkProgram aggregate. Identity tuple per ADR-3:
// (DisciplineID, SpecialtyCode, ApplicableFromYear).
type NewWorkProgramInput struct {
	DisciplineID       int64
	SpecialtyCode      string
	ApplicableFromYear int
	Title              string
	Annotation         string
	AuthorID           int64
}

// WorkProgram — aggregate root for рабочая программа дисциплины (РПД).
// Identity = (DisciplineID, SpecialtyCode, ApplicableFromYear) per
// ADR-3. Status FSM per ADR-2. Inner aggregates (Goal/Competence/
// Topic/AssessmentCriterion/Reference/Revision) land in subsequent
// TDD pairs of this PR.
type WorkProgram struct {
	id                 int64
	disciplineID       int64
	specialtyCode      string
	applicableFromYear int
	title              string
	annotation         string
	status             domain.Status
	authorID           int64
	approverID         *int64
	approvedAt         *time.Time
	rejectReason       string
	version            int
	createdAt          time.Time
	updatedAt          time.Time
}

// NewWorkProgram constructs a fresh draft WorkProgram. Inputs are
// trimmed prior to invariant checks. All five field-level invariants
// surface as ErrInvalidWorkProgram with the offending field named
// (so handlers can map to 422 with a usable message). On success the
// aggregate is in status=draft, version=0 (optimistic-lock starting
// point per v0.157.0 ADR-2), approved_at=nil.
func NewWorkProgram(in NewWorkProgramInput) (*WorkProgram, error) {
	title := strings.TrimSpace(in.Title)
	specialty := strings.TrimSpace(in.SpecialtyCode)
	annotation := strings.TrimSpace(in.Annotation)

	if title == "" {
		return nil, fmt.Errorf("%w: title is required", domain.ErrInvalidWorkProgram)
	}
	if specialty == "" {
		return nil, fmt.Errorf("%w: specialty_code is required", domain.ErrInvalidWorkProgram)
	}
	if in.DisciplineID <= 0 {
		return nil, fmt.Errorf("%w: discipline_id must be positive", domain.ErrInvalidWorkProgram)
	}
	if in.AuthorID <= 0 {
		return nil, fmt.Errorf("%w: author_id must be positive", domain.ErrInvalidWorkProgram)
	}
	if in.ApplicableFromYear < minApplicableYear || in.ApplicableFromYear > maxApplicableYear {
		return nil, fmt.Errorf("%w: applicable_from_year must be in [%d, %d]",
			domain.ErrInvalidWorkProgram, minApplicableYear, maxApplicableYear)
	}
	if utf8.RuneCountInString(annotation) > maxAnnotationLen {
		return nil, fmt.Errorf("%w: annotation must be <= %d runes", domain.ErrInvalidWorkProgram, maxAnnotationLen)
	}

	now := time.Now().UTC()
	return &WorkProgram{
		disciplineID:       in.DisciplineID,
		specialtyCode:      specialty,
		applicableFromYear: in.ApplicableFromYear,
		title:              title,
		annotation:         annotation,
		status:             domain.StatusDraft,
		authorID:           in.AuthorID,
		version:            0,
		createdAt:          now,
		updatedAt:          now,
	}, nil
}

// --- Status FSM transitions per ADR-2 ---

// Submit transitions the WorkProgram from draft or needs_revision to
// pending_approval. Author-only operation; caller (use case) handles
// the role check.
func (w *WorkProgram) Submit() error {
	if w.status != domain.StatusDraft && w.status != domain.StatusNeedsRevision {
		return domain.ErrInvalidStatusTransition
	}
	w.status = domain.StatusPendingApproval
	w.updatedAt = time.Now().UTC()
	return nil
}

// Approve transitions the WorkProgram from pending_approval to
// approved. Methodist-only operation per ADR-5; approverID is the
// acting user's ID, recorded for audit / Рособрнадзор-trail.
func (w *WorkProgram) Approve(approverID int64) error {
	if w.status != domain.StatusPendingApproval {
		return domain.ErrInvalidStatusTransition
	}
	if approverID <= 0 {
		return fmt.Errorf("%w: approver_id must be positive", domain.ErrInvalidWorkProgram)
	}
	now := time.Now().UTC()
	w.status = domain.StatusApproved
	w.approverID = &approverID
	w.approvedAt = &now
	w.rejectReason = ""
	w.updatedAt = now
	return nil
}

// MarkNeedsRevision transitions the WorkProgram from approved to
// needs_revision. Auto-triggered by DisciplineItem.Updated event
// handler per ADR-8; safe-noop if already in needs_revision (event
// dispatch may double-fire on retry).
func (w *WorkProgram) MarkNeedsRevision() error {
	if w.status == domain.StatusNeedsRevision {
		return nil
	}
	if w.status != domain.StatusApproved {
		return domain.ErrInvalidStatusTransition
	}
	w.status = domain.StatusNeedsRevision
	w.updatedAt = time.Now().UTC()
	return nil
}

// Archive transitions the WorkProgram to archived (terminal). Allowed
// from draft / approved / needs_revision per ADR-2. Cannot archive
// from pending_approval — methodist must Reject first так чтобы
// reason is recorded.
func (w *WorkProgram) Archive() error {
	switch w.status {
	case domain.StatusDraft, domain.StatusApproved, domain.StatusNeedsRevision:
		w.status = domain.StatusArchived
		w.updatedAt = time.Now().UTC()
		return nil
	default:
		return domain.ErrInvalidStatusTransition
	}
}

// Read-only accessors. Aggregate fields stay unexported so invariants
// can only be mutated via aggregate methods.

// ID returns the persistent identifier (0 for fresh, not-yet-saved aggregates).
func (w *WorkProgram) ID() int64 { return w.id }

// DisciplineID returns the linked discipline identifier.
func (w *WorkProgram) DisciplineID() int64 { return w.disciplineID }

// SpecialtyCode returns the specialty/program code (e.g. "09.03.01").
func (w *WorkProgram) SpecialtyCode() string { return w.specialtyCode }

// ApplicableFromYear returns the cohort year (year of student intake).
func (w *WorkProgram) ApplicableFromYear() int { return w.applicableFromYear }

// Title returns the program title.
func (w *WorkProgram) Title() string { return w.title }

// Annotation returns the free-form annotation (≤ 8192 chars).
func (w *WorkProgram) Annotation() string { return w.annotation }

// Status returns the current FSM state.
func (w *WorkProgram) Status() domain.Status { return w.status }

// AuthorID returns the original author identifier.
func (w *WorkProgram) AuthorID() int64 { return w.authorID }

// ApproverID returns the methodist who approved this WP, or nil if not approved.
func (w *WorkProgram) ApproverID() *int64 { return w.approverID }

// ApprovedAt returns the approval timestamp, or nil if not approved.
func (w *WorkProgram) ApprovedAt() *time.Time { return w.approvedAt }

// RejectReason returns the methodist's rejection rationale (empty if not rejected).
func (w *WorkProgram) RejectReason() string { return w.rejectReason }

// Version returns the optimistic-lock counter (starts at 0).
func (w *WorkProgram) Version() int { return w.version }

// CreatedAt returns the creation timestamp.
func (w *WorkProgram) CreatedAt() time.Time { return w.createdAt }

// UpdatedAt returns the last mutation timestamp.
func (w *WorkProgram) UpdatedAt() time.Time { return w.updatedAt }
