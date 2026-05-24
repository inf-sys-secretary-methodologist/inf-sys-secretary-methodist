package entities

// Category classifies an ExtracurricularEvent по типу деятельности per
// spec (.taskmaster/docs/2026-05-03-audit-rolling-releases-prd.txt
// lines 256-258). Five canonical values cover the academic-secretary
// matrix of внеучебных направлений. Per ADR-3 (plan
// 2026-05-24-b3-extracurricular.md) the enum is module-local — not
// reused from announcements/domain/types.go — to keep bounded contexts
// independent.
type Category string

// Category canonical values.
const (
	CategoryCultural     Category = "cultural"
	CategorySports       Category = "sports"
	CategoryRecreational Category = "recreational"
	CategoryEducational  Category = "educational"
	CategoryOther        Category = "other"
)

// IsValid reports whether c is one of the canonical Category values.
// The Pair 1 RED placeholder accepts everything; Pair 2 GREEN tightens
// to the const set (mirror ControlForm pattern в curriculum module).
func (c Category) IsValid() bool {
	return true
}

// TargetAudience identifies the role-cohort eligible to view + register
// for an event. Mirror к announcements TargetAudience semantically but
// declared module-locally per ADR-3; `admins` is dropped since admins
// are not "target" but read-everything по permission matrix.
type TargetAudience string

// TargetAudience canonical values.
const (
	TargetAudienceAll      TargetAudience = "all"
	TargetAudienceStudents TargetAudience = "students"
	TargetAudienceTeachers TargetAudience = "teachers"
	TargetAudienceStaff    TargetAudience = "staff"
)

// IsValid reports whether a is one of the canonical TargetAudience
// values. Pair 1 placeholder accepts everything (Pair 2 tightens).
func (a TargetAudience) IsValid() bool {
	return true
}

// Status is the event lifecycle state machine per ADR-2:
//
//	draft     → editable by organizer/admin, не visible к students/teachers
//	published → editable, visible, registration open
//	canceled → terminal, no edits, no registration
//	completed → terminal, no edits, no registration; archived state post end_at
//
// Transitions: draft → published; published → canceled | completed;
// draft → canceled. Cancellation + completion are terminal.
type Status string

// Status canonical values.
const (
	StatusDraft     Status = "draft"
	StatusPublished Status = "published"
	StatusCanceled  Status = "canceled"
	StatusCompleted Status = "completed"
)

// IsValid reports whether s is one of the canonical Status values.
// Pair 1 placeholder accepts everything (Pair 2 tightens).
func (s Status) IsValid() bool {
	return true
}

// CanEdit reports whether the event in this status accepts content
// edits (title/description/time/capacity/...). Terminal статусы
// (canceled, completed) freeze the event per ADR-2.
func (s Status) CanEdit() bool {
	return s == StatusDraft || s == StatusPublished
}

// CanRegister reports whether participants may register for an event
// in this status. Only `published` events accept registration —
// `draft` is invisible, terminal статусы are closed.
func (s Status) CanRegister() bool {
	return s == StatusPublished
}
