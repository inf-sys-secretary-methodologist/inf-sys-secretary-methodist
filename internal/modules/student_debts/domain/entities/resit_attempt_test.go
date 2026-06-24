package entities_test

import (
	"errors"
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/entities"
)

func validAttempt(t *testing.T) *entities.ResitAttempt {
	t.Helper()
	a, err := entities.NewResitAttempt(1, time.Now().Add(48*time.Hour), "Петров С.И.", false)
	if err != nil {
		t.Fatalf("NewResitAttempt valid: unexpected err %v", err)
	}
	return a
}

func TestNewResitAttempt_Valid(t *testing.T) {
	a := validAttempt(t)
	if a.AttemptNo != 1 || a.Examiner() != "Петров С.И." {
		t.Errorf("unexpected attempt: %+v", a)
	}
	if a.Result() != entities.ResitResultPending {
		t.Errorf("new attempt result = %q, want pending", a.Result())
	}
}

func TestNewResitAttempt_Invalid(t *testing.T) {
	now := time.Now().Add(time.Hour)
	tests := []struct {
		name      string
		attemptNo int
		date      time.Time
		examiner  string
	}{
		{"zero attempt no", 0, now, "Петров"},
		{"negative attempt no", -1, now, "Петров"},
		{"empty examiner", 1, now, ""},
		{"whitespace examiner", 1, now, "   "},
		{"zero date", 1, time.Time{}, "Петров"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := entities.NewResitAttempt(tt.attemptNo, tt.date, tt.examiner, false)
			if !errors.Is(err, entities.ErrInvalidResitAttempt) {
				t.Errorf("err = %v, want ErrInvalidResitAttempt", err)
			}
		})
	}
}

func TestResitAttempt_Record(t *testing.T) {
	grade := 4
	now := time.Now()

	t.Run("passed with grade", func(t *testing.T) {
		a := validAttempt(t)
		if err := a.Record(entities.ResitResultPassed, &grade, 42, now); err != nil {
			t.Fatalf("Record: %v", err)
		}
		if a.Result() != entities.ResitResultPassed {
			t.Errorf("result = %q, want passed", a.Result())
		}
		if a.Grade() == nil || *a.Grade() != 4 {
			t.Errorf("grade = %v, want 4", a.Grade())
		}
		if a.RecordedBy() == nil || *a.RecordedBy() != 42 {
			t.Errorf("recordedBy = %v, want 42", a.RecordedBy())
		}
		if a.RecordedAt() == nil {
			t.Error("recordedAt not set")
		}
	})

	t.Run("pending result rejected", func(t *testing.T) {
		a := validAttempt(t)
		if err := a.Record(entities.ResitResultPending, nil, 42, now); !errors.Is(err, entities.ErrInvalidResitRecord) {
			t.Errorf("err = %v, want ErrInvalidResitRecord", err)
		}
	})

	t.Run("unknown result rejected", func(t *testing.T) {
		a := validAttempt(t)
		if err := a.Record(entities.ResitResult("bogus"), nil, 42, now); !errors.Is(err, entities.ErrInvalidResitRecord) {
			t.Errorf("err = %v, want ErrInvalidResitRecord", err)
		}
	})

	t.Run("non-positive recorder rejected", func(t *testing.T) {
		a := validAttempt(t)
		if err := a.Record(entities.ResitResultFailed, nil, 0, now); !errors.Is(err, entities.ErrInvalidResitRecord) {
			t.Errorf("err = %v, want ErrInvalidResitRecord", err)
		}
	})

	t.Run("double record rejected", func(t *testing.T) {
		a := validAttempt(t)
		if err := a.Record(entities.ResitResultFailed, nil, 42, now); err != nil {
			t.Fatalf("first Record: %v", err)
		}
		if err := a.Record(entities.ResitResultPassed, &grade, 42, now); !errors.Is(err, entities.ErrAttemptAlreadyRecorded) {
			t.Errorf("err = %v, want ErrAttemptAlreadyRecorded", err)
		}
	})
}
