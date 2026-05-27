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
// ADR-3. Status FSM per ADR-2. Inner aggregates per ADR-1 mutate only
// through the AddX collection methods so the root can enforce
// aggregate-wide invariants (frozen status, code uniqueness, monotonic
// revision numbering).
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
	goals              []*Goal
	competences        []*Competence
	topics             []*Topic
	assessments        []*AssessmentCriterion
	references         []*Reference
	revisions          []*Revision
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

// Reject transitions the WorkProgram from pending_approval back to
// draft with a recorded reason. Methodist-only per ADR-5. Reason is
// trimmed before storage; empty/whitespace-only after trim is
// rejected via ErrRejectReasonRequired (the author needs actionable
// feedback).
//
// Transition guard runs before reason guard so an empty reason on a
// wrong-status WP returns the status error (more informative). When
// both would fire the status mismatch wins.
func (w *WorkProgram) Reject(reason string) error {
	if w.status != domain.StatusPendingApproval {
		return domain.ErrInvalidStatusTransition
	}
	trimmed := strings.TrimSpace(reason)
	if trimmed == "" {
		return domain.ErrRejectReasonRequired
	}
	w.status = domain.StatusDraft
	w.rejectReason = trimmed
	w.updatedAt = time.Now().UTC()
	return nil
}

// DiscardDraft transitions the WorkProgram from draft to archived
// without going through approval — author abandons their own draft.
// Allowed from draft only; other states have proper Archive/Reject
// paths that preserve audit trail.
func (w *WorkProgram) DiscardDraft() error {
	if w.status != domain.StatusDraft {
		return domain.ErrInvalidStatusTransition
	}
	w.status = domain.StatusArchived
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

// --- Inner-aggregate collection methods (ADR-1) ---

// canEditContent reports whether content mutations (Goal / Competence
// / Topic / Assessment / Reference) are allowed in the current status.
// Per ADR-2, only draft and needs_revision permit content edits;
// pending_approval / approved / archived are frozen.
func (w *WorkProgram) canEditContent() bool {
	return w.status == domain.StatusDraft || w.status == domain.StatusNeedsRevision
}

// AddGoal appends a Goal to the aggregate. Returns
// ErrCannotEditFrozenStatus if the program is in a frozen status.
func (w *WorkProgram) AddGoal(g *Goal) error {
	if g == nil {
		return fmt.Errorf("%w: goal must not be nil", domain.ErrInvalidWorkProgram)
	}
	if !w.canEditContent() {
		return domain.ErrCannotEditFrozenStatus
	}
	w.goals = append(w.goals, g)
	w.updatedAt = time.Now().UTC()
	return nil
}

// AddCompetence appends a Competence. Code must be unique within the
// program (mirrors uq_wpc_program_code at the DB level).
func (w *WorkProgram) AddCompetence(c *Competence) error {
	if c == nil {
		return fmt.Errorf("%w: competence must not be nil", domain.ErrInvalidWorkProgram)
	}
	if !w.canEditContent() {
		return domain.ErrCannotEditFrozenStatus
	}
	for _, existing := range w.competences {
		if existing.Code() == c.Code() {
			return fmt.Errorf("%w: code %q", domain.ErrDuplicateCompetenceCode, c.Code())
		}
	}
	w.competences = append(w.competences, c)
	w.updatedAt = time.Now().UTC()
	return nil
}

// AddTopic appends a Topic. HoursTotal cross-aggregate validation
// (sum vs учебный план) lives in the use-case layer per ADR-1.
func (w *WorkProgram) AddTopic(t *Topic) error {
	if t == nil {
		return fmt.Errorf("%w: topic must not be nil", domain.ErrInvalidWorkProgram)
	}
	if !w.canEditContent() {
		return domain.ErrCannotEditFrozenStatus
	}
	w.topics = append(w.topics, t)
	w.updatedAt = time.Now().UTC()
	return nil
}

// AddAssessment appends an AssessmentCriterion (ФОС item).
func (w *WorkProgram) AddAssessment(a *AssessmentCriterion) error {
	if a == nil {
		return fmt.Errorf("%w: assessment must not be nil", domain.ErrInvalidWorkProgram)
	}
	if !w.canEditContent() {
		return domain.ErrCannotEditFrozenStatus
	}
	w.assessments = append(w.assessments, a)
	w.updatedAt = time.Now().UTC()
	return nil
}

// AddReference appends a Reference (литература/источник).
func (w *WorkProgram) AddReference(r *Reference) error {
	if r == nil {
		return fmt.Errorf("%w: reference must not be nil", domain.ErrInvalidWorkProgram)
	}
	if !w.canEditContent() {
		return domain.ErrCannotEditFrozenStatus
	}
	w.references = append(w.references, r)
	w.updatedAt = time.Now().UTC()
	return nil
}

// AddRevision appends a Revision (лист актуализации). Permitted only
// when the parent is approved or needs_revision per ADR-10 — drafts
// have no baseline; pending_approval / archived programs cannot
// accept revisions. revision_number must equal NextRevisionNumber()
// (monotonic, no gaps).
func (w *WorkProgram) AddRevision(r *Revision) error {
	if r == nil {
		return fmt.Errorf("%w: revision must not be nil", domain.ErrInvalidWorkProgram)
	}
	if w.status != domain.StatusApproved && w.status != domain.StatusNeedsRevision {
		return domain.ErrRevisionNotPermitted
	}
	expected := w.NextRevisionNumber()
	if r.RevisionNumber() != expected {
		return fmt.Errorf("%w: revision_number must equal %d (next monotonic), got %d",
			domain.ErrInvalidWorkProgram, expected, r.RevisionNumber())
	}
	w.revisions = append(w.revisions, r)
	w.updatedAt = time.Now().UTC()
	return nil
}

// NextRevisionNumber returns the expected revision_number for the
// next AddRevision call: 1 if no revisions yet, else max + 1.
func (w *WorkProgram) NextRevisionNumber() int {
	maxN := 0
	for _, r := range w.revisions {
		if r.RevisionNumber() > maxN {
			maxN = r.RevisionNumber()
		}
	}
	return maxN + 1
}

// ReconstituteWorkProgramInput collects fields for repository
// hydration. Mirror migration 047 + 048 columns.
type ReconstituteWorkProgramInput struct {
	ID                 int64
	DisciplineID       int64
	SpecialtyCode      string
	ApplicableFromYear int
	Title              string
	Annotation         string
	Status             domain.Status
	AuthorID           int64
	ApproverID         *int64
	ApprovedAt         *time.Time
	RejectReason       string
	Version            int
	CreatedAt          time.Time
	UpdatedAt          time.Time
	Goals              []*Goal
	Competences        []*Competence
	Topics             []*Topic
	Assessments        []*AssessmentCriterion
	References         []*Reference
	Revisions          []*Revision
}

// ReconstituteWorkProgram builds an aggregate from persisted state.
// Skips invariant checks — DB CHECK constraints + inner-entity
// Reconstitute calls already validated. Inner slices are stored by
// reference; the repository owns lifetime semantics.
func ReconstituteWorkProgram(in ReconstituteWorkProgramInput) *WorkProgram {
	return &WorkProgram{
		id:                 in.ID,
		disciplineID:       in.DisciplineID,
		specialtyCode:      in.SpecialtyCode,
		applicableFromYear: in.ApplicableFromYear,
		title:              in.Title,
		annotation:         in.Annotation,
		status:             in.Status,
		authorID:           in.AuthorID,
		approverID:         in.ApproverID,
		approvedAt:         in.ApprovedAt,
		rejectReason:       in.RejectReason,
		version:            in.Version,
		createdAt:          in.CreatedAt,
		updatedAt:          in.UpdatedAt,
		goals:              in.Goals,
		competences:        in.Competences,
		topics:             in.Topics,
		assessments:        in.Assessments,
		references:         in.References,
		revisions:          in.Revisions,
	}
}

// HoursTotal aggregates Topic.Hours per kind. The returned map always
// contains all four canonical TopicKinds (initialized to zero) so
// callers can index without nil-map / missing-key hazards. Cross-
// aggregate validation (sum vs учебный план) is the use-case layer's
// job per ADR-1.
func (w *WorkProgram) HoursTotal() map[domain.TopicKind]int {
	result := map[domain.TopicKind]int{
		domain.TopicKindLecture:   0,
		domain.TopicKindPractice:  0,
		domain.TopicKindLab:       0,
		domain.TopicKindSelfStudy: 0,
	}
	for _, t := range w.topics {
		result[t.Kind()] += t.Hours()
	}
	return result
}

// Goals returns a defensive copy of the goals slice.
func (w *WorkProgram) Goals() []*Goal {
	out := make([]*Goal, len(w.goals))
	copy(out, w.goals)
	return out
}

// Competences returns a defensive copy of the competences slice.
func (w *WorkProgram) Competences() []*Competence {
	out := make([]*Competence, len(w.competences))
	copy(out, w.competences)
	return out
}

// Topics returns a defensive copy of the topics slice.
func (w *WorkProgram) Topics() []*Topic {
	out := make([]*Topic, len(w.topics))
	copy(out, w.topics)
	return out
}

// Assessments returns a defensive copy of the assessments slice.
func (w *WorkProgram) Assessments() []*AssessmentCriterion {
	out := make([]*AssessmentCriterion, len(w.assessments))
	copy(out, w.assessments)
	return out
}

// References returns a defensive copy of the references slice.
func (w *WorkProgram) References() []*Reference {
	out := make([]*Reference, len(w.references))
	copy(out, w.references)
	return out
}

// Revisions returns a defensive copy of the revisions slice.
func (w *WorkProgram) Revisions() []*Revision {
	out := make([]*Revision, len(w.revisions))
	copy(out, w.revisions)
	return out
}

// Read-only accessors. Aggregate fields stay unexported so invariants
// can only be mutated via aggregate methods.

// ID returns the persistent identifier (0 for fresh, not-yet-saved aggregates).
func (w *WorkProgram) ID() int64 { return w.id }

// SetID assigns the persistent identifier after a successful repository
// INSERT ... RETURNING id. Repository-only contract — use cases and
// handlers get a fully-formed aggregate from the repo and MUST NOT
// call this. Mirrors the curriculum module pattern (v0.157.0+).
func (w *WorkProgram) SetID(id int64) { w.id = id }

// SetVersion writes the optimistic-lock counter after a successful
// repository UPDATE. Repository-only contract (same caveat as SetID).
// Callers see a consistent post-update view without a separate reload.
func (w *WorkProgram) SetVersion(v int) { w.version = v }

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
