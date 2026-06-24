package entities

import (
	"errors"
	"fmt"
)

// ErrInvalidControlForm signals an unrecognized ControlForm value.
var ErrInvalidControlForm = errors.New("control_form: unknown form")

// ControlForm is the form of academic control a debt is owed against. This
// bounded context owns its own copy (cross-module imports are forbidden); the
// wire values mirror curriculum.ControlForm and the 4 РФ academic standards so
// imported source data round-trips without translation.
//
//	zachet              — зачёт
//	exam                — экзамен
//	course_project      — курсовой проект
//	differential_zachet — дифференцированный зачёт
type ControlForm string

// Recognized control forms.
const (
	ControlFormZachet             ControlForm = "zachet"
	ControlFormExam               ControlForm = "exam"
	ControlFormCourseProject      ControlForm = "course_project"
	ControlFormDifferentialZachet ControlForm = "differential_zachet"
)

// IsValid reports whether c is one of the recognized forms.
func (c ControlForm) IsValid() bool {
	switch c {
	case ControlFormZachet, ControlFormExam, ControlFormCourseProject, ControlFormDifferentialZachet:
		return true
	default:
		return false
	}
}

// Validate returns nil for a recognized form, else wraps ErrInvalidControlForm.
func (c ControlForm) Validate() error {
	if !c.IsValid() {
		return fmt.Errorf("%w: %q", ErrInvalidControlForm, string(c))
	}
	return nil
}

// String returns the canonical wire representation.
func (c ControlForm) String() string { return string(c) }
