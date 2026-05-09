package entities

import (
	"errors"
	"fmt"
)

// ErrInvalidControlForm signals that a ControlForm value does not
// match one of the 4 recognized РФ academic standard forms.
var ErrInvalidControlForm = errors.New("control_form: unknown form")

// ControlForm is the typed enum mirroring the SQL CHECK on
// curriculum_section_items.control_form (chk_section_items_control_form_enum,
// migration 035). Per CLAUDE.md ubiquitous-language gate: bare `string`
// with magic values → typed enum with methods.
//
// 4 РФ academic standard forms:
//   - zachet               — зачёт (pass/fail)
//   - exam                 — экзамен (graded на 5-балльной шкале)
//   - course_project       — курсовой проект
//   - differential_zachet  — дифференцированный зачёт (graded zachet)
//
// The string literals match the SQL CHECK byte-for-byte so domain
// values round-trip без translation.
type ControlForm string

// Recognized control forms.
const (
	ControlFormZachet             ControlForm = "zachet"
	ControlFormExam               ControlForm = "exam"
	ControlFormCourseProject      ControlForm = "course_project"
	ControlFormDifferentialZachet ControlForm = "differential_zachet"
)

// IsValid reports whether c is one of the recognized forms.
// Repository implementations call this on Reconstitute paths so a row
// that somehow holds an unknown form is rejected before it leaks into
// the use-case layer.
func (c ControlForm) IsValid() bool {
	switch c {
	case ControlFormZachet, ControlFormExam, ControlFormCourseProject, ControlFormDifferentialZachet:
		return true
	default:
		return false
	}
}

// Validate returns nil if c is a recognized form, or an error wrapping
// ErrInvalidControlForm with the offending value otherwise. Used by
// NewDisciplineItem invariant checks.
func (c ControlForm) Validate() error {
	if !c.IsValid() {
		return fmt.Errorf("%w: %q", ErrInvalidControlForm, string(c))
	}
	return nil
}

// String returns the canonical wire representation.
func (c ControlForm) String() string { return string(c) }
