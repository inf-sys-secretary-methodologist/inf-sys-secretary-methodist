// Package entities holds the WorkProgram aggregate root and its inner
// entities (Goal, Competence, Topic, AssessmentCriterion, Reference,
// Revision). See docs/plans/2026-05-27-work-program-initiative.md for
// ADR rationale.
package entities

import (
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
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

// NewWorkProgram constructs a fresh draft WorkProgram. Stub for the
// RED commit — GREEN commit fills in invariant checks.
func NewWorkProgram(_ NewWorkProgramInput) (*WorkProgram, error) {
	return nil, domain.ErrInvalidWorkProgram
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
