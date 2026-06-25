package entities_test

import (
	"errors"
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/entities"
)

func newOpenDebt(t *testing.T) *entities.StudentDebt {
	t.Helper()
	d, err := entities.NewStudentDebt("Кузнецов Д.А.", "БИ-21", "Базы данных", 3, entities.ControlFormExam)
	if err != nil {
		t.Fatalf("NewStudentDebt valid: %v", err)
	}
	return d
}

func TestNewStudentDebt_Valid(t *testing.T) {
	d := newOpenDebt(t)
	if d.Status() != entities.DebtStatusOpen {
		t.Errorf("new debt status = %q, want open", d.Status())
	}
	if d.Version != 1 {
		t.Errorf("new debt version = %d, want 1", d.Version)
	}
	if len(d.Attempts()) != 0 {
		t.Errorf("new debt has %d attempts, want 0", len(d.Attempts()))
	}
}

func TestNewStudentDebt_Invalid(t *testing.T) {
	tests := []struct {
		name                       string
		student, group, discipline string
		semester                   int
		form                       entities.ControlForm
	}{
		{"empty student", "", "БИ-21", "БД", 3, entities.ControlFormExam},
		{"empty group", "Кузнецов", "", "БД", 3, entities.ControlFormExam},
		{"empty discipline", "Кузнецов", "БИ-21", "", 3, entities.ControlFormExam},
		{"semester too low", "Кузнецов", "БИ-21", "БД", 0, entities.ControlFormExam},
		{"semester too high", "Кузнецов", "БИ-21", "БД", 13, entities.ControlFormExam},
		{"bad control form", "Кузнецов", "БИ-21", "БД", 3, entities.ControlForm("bogus")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := entities.NewStudentDebt(tt.student, tt.group, tt.discipline, tt.semester, tt.form)
			if !errors.Is(err, entities.ErrInvalidStudentDebt) {
				t.Errorf("err = %v, want ErrInvalidStudentDebt", err)
			}
		})
	}
}

func TestUpdateSourceFields(t *testing.T) {
	t.Run("refreshes denormalized fields, preserves lifecycle", func(t *testing.T) {
		d := newOpenDebt(t)
		now := time.Now()
		_ = d.ScheduleResit(now.Add(72*time.Hour), "Петров", now) // status resit_scheduled, 1 attempt

		err := d.UpdateSourceFields("Кузнецов Денис Алексеевич", "БИ-21", "Базы данных и СУБД", 4, entities.ControlFormDifferentialZachet)
		if err != nil {
			t.Fatalf("UpdateSourceFields valid: %v", err)
		}
		if d.StudentFullName != "Кузнецов Денис Алексеевич" || d.DisciplineName != "Базы данных и СУБД" || d.Semester != 4 {
			t.Errorf("fields not refreshed: %+v", d)
		}
		if d.ControlForm != entities.ControlFormDifferentialZachet {
			t.Errorf("control form not refreshed: %q", d.ControlForm)
		}
		if d.Status() != entities.DebtStatusResitScheduled {
			t.Errorf("status changed to %q, want resit_scheduled preserved", d.Status())
		}
		if len(d.Attempts()) != 1 {
			t.Errorf("attempts changed: %d, want 1 preserved", len(d.Attempts()))
		}
	})

	t.Run("rejects invalid input", func(t *testing.T) {
		cases := []struct {
			name                       string
			student, group, discipline string
			semester                   int
			form                       entities.ControlForm
		}{
			{"empty student", "", "БИ-21", "БД", 3, entities.ControlFormExam},
			{"semester too high", "Кузнецов", "БИ-21", "БД", 13, entities.ControlFormExam},
			{"bad control form", "Кузнецов", "БИ-21", "БД", 3, entities.ControlForm("bogus")},
		}
		for _, tt := range cases {
			t.Run(tt.name, func(t *testing.T) {
				d := newOpenDebt(t)
				err := d.UpdateSourceFields(tt.student, tt.group, tt.discipline, tt.semester, tt.form)
				if !errors.Is(err, entities.ErrInvalidStudentDebt) {
					t.Errorf("err = %v, want ErrInvalidStudentDebt", err)
				}
			})
		}
	})
}

func TestScheduleResit(t *testing.T) {
	now := time.Now()
	date := now.Add(72 * time.Hour)

	t.Run("from open creates first non-commission attempt", func(t *testing.T) {
		d := newOpenDebt(t)
		if err := d.ScheduleResit(date, "Петров", now); err != nil {
			t.Fatalf("ScheduleResit: %v", err)
		}
		if d.Status() != entities.DebtStatusResitScheduled {
			t.Errorf("status = %q, want resit_scheduled", d.Status())
		}
		if len(d.Attempts()) != 1 || d.Attempts()[0].IsCommission {
			t.Errorf("attempts = %+v, want 1 non-commission", d.Attempts())
		}
	})

	t.Run("rejected when already scheduled", func(t *testing.T) {
		d := newOpenDebt(t)
		_ = d.ScheduleResit(date, "Петров", now)
		if err := d.ScheduleResit(date, "Петров", now); !errors.Is(err, entities.ErrInvalidTransition) {
			t.Errorf("err = %v, want ErrInvalidTransition", err)
		}
	})

	t.Run("rejected when closed", func(t *testing.T) {
		d := newOpenDebt(t)
		_ = d.ScheduleResit(date, "Петров", now)
		_ = d.RecordResitResult(entities.ResitResultPassed, nil, 1, now, 2)
		if err := d.ScheduleResit(date, "Петров", now); !errors.Is(err, entities.ErrDebtClosed) {
			t.Errorf("err = %v, want ErrDebtClosed", err)
		}
	})
}

func TestRecordResitResult(t *testing.T) {
	now := time.Now()
	date := now.Add(72 * time.Hour)

	t.Run("passed closes debt", func(t *testing.T) {
		d := newOpenDebt(t)
		_ = d.ScheduleResit(date, "Петров", now)
		if err := d.RecordResitResult(entities.ResitResultPassed, intptr(5), 1, now, 2); err != nil {
			t.Fatalf("Record: %v", err)
		}
		if d.Status() != entities.DebtStatusClosedPassed {
			t.Errorf("status = %q, want closed_passed", d.Status())
		}
	})

	t.Run("failed below threshold returns to open", func(t *testing.T) {
		d := newOpenDebt(t)
		_ = d.ScheduleResit(date, "Петров", now)
		_ = d.RecordResitResult(entities.ResitResultFailed, nil, 1, now, 2)
		if d.Status() != entities.DebtStatusOpen {
			t.Errorf("status = %q, want open (1 fail, N=2)", d.Status())
		}
	})

	t.Run("failed reaching threshold escalates to commission", func(t *testing.T) {
		d := newOpenDebt(t)
		// N=2: two regular failures → commission.
		_ = d.ScheduleResit(date, "Петров", now)
		_ = d.RecordResitResult(entities.ResitResultFailed, nil, 1, now, 2)
		_ = d.ScheduleResit(date, "Петров", now)
		_ = d.RecordResitResult(entities.ResitResultNoShow, nil, 1, now, 2)
		if d.Status() != entities.DebtStatusCommission {
			t.Errorf("status = %q, want commission", d.Status())
		}
	})

	t.Run("commission attempt failed closes as failed", func(t *testing.T) {
		d := newOpenDebt(t)
		// N=1: first failure escalates straight to commission.
		_ = d.ScheduleResit(date, "Петров", now)
		_ = d.RecordResitResult(entities.ResitResultFailed, nil, 1, now, 1)
		if d.Status() != entities.DebtStatusCommission {
			t.Fatalf("precondition: status = %q, want commission", d.Status())
		}
		_ = d.ScheduleResit(date, "Комиссия", now)
		if !d.Attempts()[len(d.Attempts())-1].IsCommission {
			t.Error("expected last attempt to be a commission attempt")
		}
		_ = d.RecordResitResult(entities.ResitResultFailed, nil, 1, now, 1)
		if d.Status() != entities.DebtStatusClosedFailed {
			t.Errorf("status = %q, want closed_failed", d.Status())
		}
	})

	t.Run("commission attempt passed closes as passed", func(t *testing.T) {
		d := newOpenDebt(t)
		_ = d.ScheduleResit(date, "Петров", now)
		_ = d.RecordResitResult(entities.ResitResultFailed, nil, 1, now, 1)
		_ = d.ScheduleResit(date, "Комиссия", now)
		_ = d.RecordResitResult(entities.ResitResultPassed, intptr(4), 1, now, 1)
		if d.Status() != entities.DebtStatusClosedPassed {
			t.Errorf("status = %q, want closed_passed", d.Status())
		}
	})

	t.Run("rejected when no scheduled resit", func(t *testing.T) {
		d := newOpenDebt(t)
		if err := d.RecordResitResult(entities.ResitResultPassed, nil, 1, now, 2); !errors.Is(err, entities.ErrNoScheduledResit) {
			t.Errorf("err = %v, want ErrNoScheduledResit", err)
		}
	})
}

func intptr(i int) *int { return &i }
