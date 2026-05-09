package entities

import (
	"errors"
	"fmt"
	"strings"
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

// NewDisciplineItem validates invariants and returns a fresh
// DisciplineItem at version 0 (optimistic-locking baseline).
//
// Invariants (mirroring SQL CHECKs in migration 035):
//   - section_id > 0
//   - title trimmed-non-empty, ≤ 255 runes
//   - hours_lectures / hours_practice / hours_lab / hours_self ≥ 0 each
//   - credits ≥ 0
//   - semester ∈ [1, 12] (covers bachelor 8 + master 4)
//   - control_form ∈ {zachet, exam, course_project, differential_zachet}
//   - order_index ≥ 0
//
// Each violation wraps ErrInvalidDisciplineItem with the offending
// field so errors.Is resolves the sentinel for the 422 mapping in
// handlers, and the message identifies which field for the operator.
func NewDisciplineItem(p NewDisciplineItemParams) (*DisciplineItem, error) {
	if p.SectionID <= 0 {
		return nil, fmt.Errorf("%w: section_id must be positive, got %d",
			ErrInvalidDisciplineItem, p.SectionID)
	}
	title := strings.TrimSpace(p.Title)
	if title == "" {
		return nil, fmt.Errorf("%w: title must not be empty", ErrInvalidDisciplineItem)
	}
	if len([]rune(title)) > maxDisciplineItemTitleLen {
		return nil, fmt.Errorf("%w: title length %d exceeds max %d",
			ErrInvalidDisciplineItem, len([]rune(title)), maxDisciplineItemTitleLen)
	}
	if p.HoursLectures < 0 {
		return nil, fmt.Errorf("%w: hours_lectures must be non-negative, got %d",
			ErrInvalidDisciplineItem, p.HoursLectures)
	}
	if p.HoursPractice < 0 {
		return nil, fmt.Errorf("%w: hours_practice must be non-negative, got %d",
			ErrInvalidDisciplineItem, p.HoursPractice)
	}
	if p.HoursLab < 0 {
		return nil, fmt.Errorf("%w: hours_lab must be non-negative, got %d",
			ErrInvalidDisciplineItem, p.HoursLab)
	}
	if p.HoursSelf < 0 {
		return nil, fmt.Errorf("%w: hours_self must be non-negative, got %d",
			ErrInvalidDisciplineItem, p.HoursSelf)
	}
	if p.Credits < 0 {
		return nil, fmt.Errorf("%w: credits must be non-negative, got %d",
			ErrInvalidDisciplineItem, p.Credits)
	}
	if p.Semester < minSemester || p.Semester > maxSemester {
		return nil, fmt.Errorf("%w: semester %d outside [%d, %d]",
			ErrInvalidDisciplineItem, p.Semester, minSemester, maxSemester)
	}
	if err := p.ControlForm.Validate(); err != nil {
		return nil, fmt.Errorf("%w: control_form %s", ErrInvalidDisciplineItem, err.Error())
	}
	if p.OrderIndex < 0 {
		return nil, fmt.Errorf("%w: order_index must be non-negative, got %d",
			ErrInvalidDisciplineItem, p.OrderIndex)
	}
	return &DisciplineItem{
		sectionID:     p.SectionID,
		title:         title,
		hoursLectures: p.HoursLectures,
		hoursPractice: p.HoursPractice,
		hoursLab:      p.HoursLab,
		hoursSelf:     p.HoursSelf,
		controlForm:   p.ControlForm,
		credits:       p.Credits,
		semester:      p.Semester,
		orderIndex:    p.OrderIndex,
		version:       0,
		createdAt:     p.Now,
		updatedAt:     p.Now,
	}, nil
}

// ReconstituteDisciplineItem rebuilds a DisciplineItem from
// authoritative storage. It bypasses NewDisciplineItem's invariant
// checks because the values are already canonical (the DB enforces
// the same CHECKs at write time). Used exclusively by repository
// implementations.
func ReconstituteDisciplineItem(
	id, sectionID int64,
	title string,
	hoursLectures, hoursPractice, hoursLab, hoursSelf int,
	controlForm ControlForm,
	credits, semester, orderIndex, version int,
	createdAt, updatedAt time.Time,
) *DisciplineItem {
	return &DisciplineItem{
		ID:            id,
		sectionID:     sectionID,
		title:         title,
		hoursLectures: hoursLectures,
		hoursPractice: hoursPractice,
		hoursLab:      hoursLab,
		hoursSelf:     hoursSelf,
		controlForm:   controlForm,
		credits:       credits,
		semester:      semester,
		orderIndex:    orderIndex,
		version:       version,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
	}
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

// UpdateBasics applies a content edit (all 9 mutable fields) to the
// discipline item. The method is atomic: if any invariant fails, the
// entity is left untouched and the wrapped ErrInvalidDisciplineItem
// is returned.
//
// The status gate (curriculum lifecycle inheritance per ADR-2) is NOT
// enforced here — DisciplineItem keeps no Curriculum/Section reference.
// Callers must invoke AuthorizeDisciplineItemEdit (free function) or
// AuthorizeEdit (method) first and only then UpdateBasics. The
// two-step shape keeps DisciplineItem pure of cross-aggregate knowledge.
//
// Version is repo-managed (ADR-3); UpdateBasics does not touch it
// (mirror к Section behavior).
func (d *DisciplineItem) UpdateBasics(
	title string,
	hoursLectures, hoursPractice, hoursLab, hoursSelf int,
	controlForm ControlForm,
	credits, semester, orderIndex int,
	now time.Time,
) error {
	trimmedTitle := strings.TrimSpace(title)
	if trimmedTitle == "" {
		return fmt.Errorf("%w: title must not be empty", ErrInvalidDisciplineItem)
	}
	if len([]rune(trimmedTitle)) > maxDisciplineItemTitleLen {
		return fmt.Errorf("%w: title length %d exceeds max %d",
			ErrInvalidDisciplineItem, len([]rune(trimmedTitle)), maxDisciplineItemTitleLen)
	}
	if hoursLectures < 0 {
		return fmt.Errorf("%w: hours_lectures must be non-negative, got %d",
			ErrInvalidDisciplineItem, hoursLectures)
	}
	if hoursPractice < 0 {
		return fmt.Errorf("%w: hours_practice must be non-negative, got %d",
			ErrInvalidDisciplineItem, hoursPractice)
	}
	if hoursLab < 0 {
		return fmt.Errorf("%w: hours_lab must be non-negative, got %d",
			ErrInvalidDisciplineItem, hoursLab)
	}
	if hoursSelf < 0 {
		return fmt.Errorf("%w: hours_self must be non-negative, got %d",
			ErrInvalidDisciplineItem, hoursSelf)
	}
	if err := controlForm.Validate(); err != nil {
		return fmt.Errorf("%w: control_form %s", ErrInvalidDisciplineItem, err.Error())
	}
	if credits < 0 {
		return fmt.Errorf("%w: credits must be non-negative, got %d",
			ErrInvalidDisciplineItem, credits)
	}
	if semester < minSemester || semester > maxSemester {
		return fmt.Errorf("%w: semester %d outside [%d, %d]",
			ErrInvalidDisciplineItem, semester, minSemester, maxSemester)
	}
	if orderIndex < 0 {
		return fmt.Errorf("%w: order_index must be non-negative, got %d",
			ErrInvalidDisciplineItem, orderIndex)
	}
	// All validation passed — apply mutations atomically.
	d.title = trimmedTitle
	d.hoursLectures = hoursLectures
	d.hoursPractice = hoursPractice
	d.hoursLab = hoursLab
	d.hoursSelf = hoursSelf
	d.controlForm = controlForm
	d.credits = credits
	d.semester = semester
	d.orderIndex = orderIndex
	d.updatedAt = now
	return nil
}

// AuthorizeDisciplineItemEdit decides whether a caller may operate on
// a discipline item inside the given curriculum. Free function — rule
// depends entirely on actor + curriculum primitives, no entity state
// consulted (mirror к AuthorizeSectionEdit pattern from v0.128.0;
// declared free from первого draft to avoid Pair 2 → Pair 4 refactor
// leak experienced в Section).
//
// Gate ordering (matches Section.AuthorizeEdit contract):
//
//  1. curStatus.CanEdit() — non-editable lifecycle freezes items для
//     всех (включая admins). ErrCannotEditDisciplineItem.
//  2. isAdmin — system_admin / academic_secretary override ownership.
//  3. actorID > 0 && actorID == curCreatedBy — author methodist.
//  4. Otherwise → ErrDisciplineItemScopeForbidden.
//
// The actorID > 0 guard is defense-in-depth against a JWT subject
// lost upstream.
func AuthorizeDisciplineItemEdit(actorID int64, isAdmin bool, curStatus CurriculumStatus, curCreatedBy int64) error {
	if !curStatus.CanEdit() {
		return fmt.Errorf("%w: curriculum status %q is not editable",
			ErrCannotEditDisciplineItem, string(curStatus))
	}
	if isAdmin {
		return nil
	}
	if actorID > 0 && actorID == curCreatedBy {
		return nil
	}
	return fmt.Errorf("%w: actor %d is not the curriculum author (%d)",
		ErrDisciplineItemScopeForbidden, actorID, curCreatedBy)
}

// AuthorizeEdit is the method-form alias of AuthorizeDisciplineItemEdit
// kept for the read-mutate-save use cases (Update / Delete) where a
// loaded DisciplineItem is in scope. Create uses the free function
// directly (no instance yet). Both forms share the same logic via
// delegation — eliminating drift risk (pinned by
// TestDisciplineItem_AuthorizeEdit_MethodDelegatesToFreeFunction).
func (d *DisciplineItem) AuthorizeEdit(actorID int64, isAdmin bool, curStatus CurriculumStatus, curCreatedBy int64) error {
	_ = d
	return AuthorizeDisciplineItemEdit(actorID, isAdmin, curStatus, curCreatedBy)
}
