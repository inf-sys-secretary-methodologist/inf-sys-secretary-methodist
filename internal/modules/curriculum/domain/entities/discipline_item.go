package entities

import (
	"errors"
	"time"
)

// ErrInvalidDisciplineItem signals a violation of one of the
// DisciplineItem construction invariants (empty/oversize title,
// non-positive section_id, negative hours/credits, semester out of
// range, invalid control_form, negative order_index). Handlers map
// this sentinel to HTTP 422.
var ErrInvalidDisciplineItem = errors.New("discipline_item: invalid item")

// ErrDisciplineItemScopeForbidden indicates that a user is not
// authorized to operate on a particular DisciplineItem — the user is
// a methodist who did not author the parent Curriculum. Admins
// (system_admin / academic_secretary) override via isAdmin flag.
// Handlers map this sentinel to HTTP 403.
var ErrDisciplineItemScopeForbidden = errors.New("discipline_item: caller cannot operate on this item")

// ErrCannotEditDisciplineItem indicates that the parent Curriculum
// is in a status that does not permit content edits to its sections /
// items (anything other than draft per ADR-2 lifecycle inheritance).
// Handlers map this sentinel to HTTP 422.
var ErrCannotEditDisciplineItem = errors.New("discipline_item: cannot edit, curriculum is not in editable status")

// DisciplineItem is the Layer 2 aggregate root of the curriculum
// hierarchy (Curriculum → Sections → DisciplineItems) per plan
// 2026-05-09-v0128-section-aggregate.md ADR-1 Beta — independent AR
// carrying only `sectionID int64` FK, no navigable Section reference.
//
// Lifecycle inheritance ADR-2: no own status; editability inherits
// curriculum.status (через section.curriculum_id). Use case fetches
// section + curriculum для AuthorizeDisciplineItemEdit primitives.
//
// Optimistic locking foundation per ADR-3 (mirror к Section).
//
// Rich invariants vs Section: hours_lectures / hours_practice /
// hours_lab / hours_self (each ≥ 0), credits ≥ 0, semester ∈ [1, 12],
// control_form ∈ enum, order_index ≥ 0.
type DisciplineItem struct {
	ID            int64
	sectionID     int64
	title         string
	hoursLectures int
	hoursPractice int
	hoursLab      int
	hoursSelf     int
	controlForm   ControlForm
	credits       int
	semester      int
	orderIndex    int
	version       int
	createdAt     time.Time
	updatedAt     time.Time
}

// NewDisciplineItemParams bundles the constructor inputs.
type NewDisciplineItemParams struct {
	SectionID     int64
	Title         string
	HoursLectures int
	HoursPractice int
	HoursLab      int
	HoursSelf     int
	ControlForm   ControlForm
	Credits       int
	Semester      int
	OrderIndex    int
	Now           time.Time
}

// DisciplineItem text/numeric bounds. Mirrored exactly by the
// chk_section_items_* CHECKs в migration 035.
const (
	maxDisciplineItemTitleLen = 255
	minSemester               = 1
	maxSemester               = 12
)

// NewDisciplineItem — implementation lands в GREEN commit (Pair 1).
// Stub returns sentinel for RED tests + keeps pre-commit hook clean.
// Constants referenced as no-op so unused-symbol checker stays quiet
// across the RED→GREEN boundary.
func NewDisciplineItem(p NewDisciplineItemParams) (*DisciplineItem, error) {
	_ = p
	_, _, _ = maxDisciplineItemTitleLen, minSemester, maxSemester
	return nil, errors.New("discipline_item: NewDisciplineItem not implemented yet")
}

// ReconstituteDisciplineItem — implementation lands в GREEN commit (Pair 1).
func ReconstituteDisciplineItem(
	id, sectionID int64,
	title string,
	hoursLectures, hoursPractice, hoursLab, hoursSelf int,
	controlForm ControlForm,
	credits, semester, orderIndex, version int,
	createdAt, updatedAt time.Time,
) *DisciplineItem {
	_, _, _, _, _ = id, sectionID, title, hoursLectures, hoursPractice
	_, _, _, _, _ = hoursLab, hoursSelf, controlForm, credits, semester
	_, _, _, _ = orderIndex, version, createdAt, updatedAt
	return nil
}

// SectionID returns the FK to curriculum_sections.id.
func (d *DisciplineItem) SectionID() int64 { return d.sectionID }

// Title returns the discipline title (e.g. "Математический анализ").
func (d *DisciplineItem) Title() string { return d.title }

// HoursLectures returns the lecture hours (≥ 0).
func (d *DisciplineItem) HoursLectures() int { return d.hoursLectures }

// HoursPractice returns the practice/seminar hours (≥ 0).
func (d *DisciplineItem) HoursPractice() int { return d.hoursPractice }

// HoursLab returns the laboratory hours (≥ 0).
func (d *DisciplineItem) HoursLab() int { return d.hoursLab }

// HoursSelf returns the self-study hours (≥ 0).
func (d *DisciplineItem) HoursSelf() int { return d.hoursSelf }

// ControlForm returns the typed control form (ControlForm enum).
func (d *DisciplineItem) ControlForm() ControlForm { return d.controlForm }

// Credits returns ECTS-style credit count (≥ 0).
func (d *DisciplineItem) Credits() int { return d.credits }

// Semester returns the academic semester (1..12 — covers bachelor + master).
func (d *DisciplineItem) Semester() int { return d.semester }

// OrderIndex returns the display ordering hint (≥ 0).
func (d *DisciplineItem) OrderIndex() int { return d.orderIndex }

// Version returns the optimistic-locking version (≥ 0).
func (d *DisciplineItem) Version() int { return d.version }

// CreatedAt returns the creation timestamp.
func (d *DisciplineItem) CreatedAt() time.Time { return d.createdAt }

// UpdatedAt returns the last-mutation timestamp.
func (d *DisciplineItem) UpdatedAt() time.Time { return d.updatedAt }
