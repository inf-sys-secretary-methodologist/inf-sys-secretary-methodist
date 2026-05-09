package entities

import (
	"errors"
	"testing"
)

func TestControlForm_IsValid(t *testing.T) {
	cases := []struct {
		name string
		c    ControlForm
		want bool
	}{
		{"zachet", ControlFormZachet, true},
		{"exam", ControlFormExam, true},
		{"course_project", ControlFormCourseProject, true},
		{"differential_zachet", ControlFormDifferentialZachet, true},
		{"empty", ControlForm(""), false},
		{"unknown", ControlForm("vyzov"), false},
		{"capitalised", ControlForm("Exam"), false}, // case-sensitive — DB CHECK uses lower-case literals
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.c.IsValid(); got != tc.want {
				t.Errorf("ControlForm(%q).IsValid() = %v, want %v", string(tc.c), got, tc.want)
			}
		})
	}
}

func TestControlForm_Validate(t *testing.T) {
	if err := ControlFormZachet.Validate(); err != nil {
		t.Errorf("Validate() rejected valid form zachet: %v", err)
	}
	err := ControlForm("unknown").Validate()
	if err == nil {
		t.Fatal("Validate() accepted unknown form")
	}
	if !errors.Is(err, ErrInvalidControlForm) {
		t.Errorf("err %v does not wrap ErrInvalidControlForm", err)
	}
}

func TestControlForm_String(t *testing.T) {
	if got, want := ControlFormDifferentialZachet.String(), "differential_zachet"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
