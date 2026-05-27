// Package domain holds value objects, enums, and sentinel errors for the
// work_program bounded context. See docs/plans/2026-05-27-work-program-initiative.md
// for ADR rationale.
package domain

// Status is the lifecycle state of a WorkProgram aggregate.
//
// FSM (ADR-2):
//
//	draft → pending_approval (Submit, author)
//	draft → archived         (DiscardDraft, author/admin)
//	pending_approval → approved   (Approve, methodist)
//	pending_approval → draft      (Reject + reason, methodist)
//	approved → needs_revision     (DisciplineItem.Updated event, ADR-8)
//	needs_revision → pending_approval (Resubmit, author)
//	approved → archived           (Archive, admin/methodist)
//
// Terminal: archived. РПД никогда не удаляется (Рособрнадзор 6 лет per ADR-1 research).
type Status string

// Status values. Use these constants only — bare string literals в domain
// прямо запрещены per CLAUDE.md ubiquitous-language gate.
const (
	StatusDraft           Status = "draft"
	StatusPendingApproval Status = "pending_approval"
	StatusApproved        Status = "approved"
	StatusNeedsRevision   Status = "needs_revision"
	StatusArchived        Status = "archived"
)

// IsValid reports whether the receiver is one of the five canonical values.
func (s Status) IsValid() bool {
	switch s {
	case StatusDraft, StatusPendingApproval, StatusApproved, StatusNeedsRevision, StatusArchived:
		return true
	default:
		return false
	}
}

// String returns the wire-form of the status (matches DB CHECK constraint).
func (s Status) String() string { return string(s) }

// CompetenceType — ФГОС-derived classification of учебной компетенции.
// ПК (профессиональная), ОК (общекультурная), УК (универсальная).
type CompetenceType string

// CompetenceType values.
const (
	CompetenceTypePK CompetenceType = "pk"
	CompetenceTypeOK CompetenceType = "ok"
	CompetenceTypeUK CompetenceType = "uk"
)

// IsValid reports whether c is one of the three FGOS-recognized types.
func (c CompetenceType) IsValid() bool {
	switch c {
	case CompetenceTypePK, CompetenceTypeOK, CompetenceTypeUK:
		return true
	default:
		return false
	}
}

// String returns the wire-form.
func (c CompetenceType) String() string { return string(c) }

// TopicKind classifies a Topic by учебная нагрузка type.
type TopicKind string

// TopicKind values.
const (
	TopicKindLecture   TopicKind = "lecture"
	TopicKindPractice  TopicKind = "practice"
	TopicKindLab       TopicKind = "lab"
	TopicKindSelfStudy TopicKind = "self_study"
)

// IsValid reports whether k is one of the four canonical kinds.
func (k TopicKind) IsValid() bool {
	switch k {
	case TopicKindLecture, TopicKindPractice, TopicKindLab, TopicKindSelfStudy:
		return true
	default:
		return false
	}
}

// String returns the wire-form.
func (k TopicKind) String() string { return string(k) }

// AssessmentType — type of ФОС item (current control / intermediate /
// final attestation).
type AssessmentType string

// AssessmentType values.
const (
	AssessmentTypeCurrent      AssessmentType = "current"
	AssessmentTypeIntermediate AssessmentType = "intermediate"
	AssessmentTypeFinal        AssessmentType = "final"
)

// IsValid reports whether a is one of the three canonical assessment types.
func (a AssessmentType) IsValid() bool {
	switch a {
	case AssessmentTypeCurrent, AssessmentTypeIntermediate, AssessmentTypeFinal:
		return true
	default:
		return false
	}
}

// String returns the wire-form.
func (a AssessmentType) String() string { return string(a) }

// ReferenceKind classifies a Reference (литература/источник) by importance tier.
type ReferenceKind string

// ReferenceKind values.
const (
	ReferenceKindMain       ReferenceKind = "main"
	ReferenceKindAdditional ReferenceKind = "additional"
	ReferenceKindElectronic ReferenceKind = "electronic"
)

// IsValid reports whether r is one of the three canonical reference kinds.
func (r ReferenceKind) IsValid() bool {
	switch r {
	case ReferenceKindMain, ReferenceKindAdditional, ReferenceKindElectronic:
		return true
	default:
		return false
	}
}

// String returns the wire-form.
func (r ReferenceKind) String() string { return string(r) }
