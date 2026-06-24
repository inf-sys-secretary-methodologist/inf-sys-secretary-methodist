package entities_test

import (
	"errors"
	"testing"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/entities"
)

func TestDebtStatus(t *testing.T) {
	valid := []entities.DebtStatus{
		entities.DebtStatusOpen, entities.DebtStatusResitScheduled,
		entities.DebtStatusCommission, entities.DebtStatusClosedPassed,
		entities.DebtStatusClosedFailed,
	}
	for _, s := range valid {
		if !s.IsValid() {
			t.Errorf("DebtStatus(%q).IsValid() = false, want true", s)
		}
		if err := s.Validate(); err != nil {
			t.Errorf("DebtStatus(%q).Validate() = %v, want nil", s, err)
		}
		if s.String() != string(s) {
			t.Errorf("DebtStatus(%q).String() = %q", s, s.String())
		}
	}

	if entities.DebtStatus("bogus").IsValid() {
		t.Error("DebtStatus(bogus).IsValid() = true, want false")
	}
	if err := entities.DebtStatus("bogus").Validate(); !errors.Is(err, entities.ErrInvalidDebtStatus) {
		t.Errorf("Validate(bogus) err = %v, want ErrInvalidDebtStatus", err)
	}

	closed := map[entities.DebtStatus]bool{
		entities.DebtStatusOpen: false, entities.DebtStatusResitScheduled: false,
		entities.DebtStatusCommission: false, entities.DebtStatusClosedPassed: true,
		entities.DebtStatusClosedFailed: true,
	}
	for s, want := range closed {
		if s.IsClosed() != want {
			t.Errorf("DebtStatus(%q).IsClosed() = %v, want %v", s, s.IsClosed(), want)
		}
	}
}

func TestResitResult(t *testing.T) {
	valid := []entities.ResitResult{
		entities.ResitResultPending, entities.ResitResultPassed,
		entities.ResitResultFailed, entities.ResitResultNoShow,
	}
	for _, r := range valid {
		if !r.IsValid() {
			t.Errorf("ResitResult(%q).IsValid() = false, want true", r)
		}
		if err := r.Validate(); err != nil {
			t.Errorf("ResitResult(%q).Validate() = %v, want nil", r, err)
		}
	}
	if entities.ResitResult("bogus").IsValid() {
		t.Error("ResitResult(bogus).IsValid() = true, want false")
	}
	if err := entities.ResitResult("bogus").Validate(); !errors.Is(err, entities.ErrInvalidResitResult) {
		t.Errorf("Validate(bogus) err = %v, want ErrInvalidResitResult", err)
	}

	final := map[entities.ResitResult]bool{
		entities.ResitResultPending: false, entities.ResitResultPassed: true,
		entities.ResitResultFailed: true, entities.ResitResultNoShow: true,
	}
	for r, want := range final {
		if r.IsFinal() != want {
			t.Errorf("ResitResult(%q).IsFinal() = %v, want %v", r, r.IsFinal(), want)
		}
	}
}

func TestControlForm(t *testing.T) {
	valid := []entities.ControlForm{
		entities.ControlFormZachet, entities.ControlFormExam,
		entities.ControlFormCourseProject, entities.ControlFormDifferentialZachet,
	}
	for _, c := range valid {
		if !c.IsValid() {
			t.Errorf("ControlForm(%q).IsValid() = false, want true", c)
		}
		if err := c.Validate(); err != nil {
			t.Errorf("ControlForm(%q).Validate() = %v, want nil", c, err)
		}
	}
	if entities.ControlForm("bogus").IsValid() {
		t.Error("ControlForm(bogus).IsValid() = true, want false")
	}
	if err := entities.ControlForm("bogus").Validate(); !errors.Is(err, entities.ErrInvalidControlForm) {
		t.Errorf("Validate(bogus) err = %v, want ErrInvalidControlForm", err)
	}
}
