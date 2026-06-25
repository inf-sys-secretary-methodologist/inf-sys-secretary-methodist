package usecases_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/repositories"
)

// scheduledDebt returns a debt in resit_scheduled with one pending
// regular attempt (AttemptNo 1).
func scheduledDebt(t *testing.T) *entities.StudentDebt {
	t.Helper()
	d := debtWith(55, ptr(int64(999)), ptr(int64(7)))
	require.NoError(t, d.ScheduleResit(time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC), "Петров П.П.", time.Now()))
	return d
}

func recordInput(attemptNo int, result entities.ResitResult, grade *int) usecases.RecordResitResultInput {
	return usecases.RecordResitResultInput{DebtID: 55, AttemptNo: attemptNo, Result: result, Grade: grade}
}

func TestRecordResitResultUseCase_FSMOutcomes(t *testing.T) {
	tests := []struct {
		name       string
		role       string
		n          int
		result     entities.ResitResult
		grade      *int
		wantStatus entities.DebtStatus
	}{
		{"passed closes debt", "methodist", 2, entities.ResitResultPassed, ptr(5), entities.DebtStatusClosedPassed},
		{"failed below N returns to open", "academic_secretary", 2, entities.ResitResultFailed, nil, entities.DebtStatusOpen},
		{"no_show below N returns to open", "methodist", 2, entities.ResitResultNoShow, nil, entities.DebtStatusOpen},
		{"failed reaching N escalates to commission", "system_admin", 1, entities.ResitResultFailed, nil, entities.DebtStatusCommission},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			debt := scheduledDebt(t)
			updated := false
			repo := &fakeDebtRepo{
				getByID: func(_ context.Context, _ int64) (*entities.StudentDebt, error) { return debt, nil },
				update:  func(_ context.Context, _ *entities.StudentDebt) error { updated = true; return nil },
			}
			uc := usecases.NewRecordResitResultUseCase(repo, &recordingAudit{}, fixedClock(), tt.n)

			got, err := uc.Execute(context.Background(), 7, tt.role, recordInput(1, tt.result, tt.grade))
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, got.Status())
			assert.True(t, updated, "the advanced aggregate must be persisted")
		})
	}
}

func TestRecordResitResultUseCase_PersistsAndAudits(t *testing.T) {
	debt := scheduledDebt(t)
	var updated *entities.StudentDebt
	repo := &fakeDebtRepo{
		getByID: func(_ context.Context, _ int64) (*entities.StudentDebt, error) { return debt, nil },
		update:  func(_ context.Context, d *entities.StudentDebt) error { updated = d; return nil },
	}
	audit := &recordingAudit{}
	uc := usecases.NewRecordResitResultUseCase(repo, audit, fixedClock(), 2)

	got, err := uc.Execute(context.Background(), 7, "methodist", recordInput(1, entities.ResitResultPassed, ptr(5)))
	require.NoError(t, err)
	assert.Equal(t, entities.DebtStatusClosedPassed, got.Status())
	assert.Same(t, debt, updated, "must persist the mutated aggregate")
	require.Len(t, audit.events, 1)
	assert.Equal(t, "student_debts.resit_recorded", audit.events[0].action)
}

func TestRecordResitResultUseCase_AttemptNoMismatchRejected(t *testing.T) {
	debt := scheduledDebt(t) // latest attempt is 1
	repo := &fakeDebtRepo{
		getByID: func(_ context.Context, _ int64) (*entities.StudentDebt, error) { return debt, nil },
		update: func(_ context.Context, _ *entities.StudentDebt) error {
			t.Fatal("repo.Update must NOT be called on attempt-no mismatch")
			return nil
		},
	}
	uc := usecases.NewRecordResitResultUseCase(repo, &recordingAudit{}, fixedClock(), 2)

	_, err := uc.Execute(context.Background(), 7, "methodist", recordInput(2, entities.ResitResultPassed, ptr(5)))
	assert.ErrorIs(t, err, entities.ErrNoScheduledResit)
}

func TestRecordResitResultUseCase_DeniedForNonManager(t *testing.T) {
	for _, role := range []string{"teacher", "student", "guest"} {
		t.Run(role, func(t *testing.T) {
			repo := &fakeDebtRepo{getByID: func(_ context.Context, _ int64) (*entities.StudentDebt, error) {
				t.Fatal("repo.GetByID must NOT be called for a denied actor")
				return nil, nil
			}}
			audit := &recordingAudit{}
			uc := usecases.NewRecordResitResultUseCase(repo, audit, fixedClock(), 2)

			_, err := uc.Execute(context.Background(), 7, role, recordInput(1, entities.ResitResultPassed, ptr(5)))
			assert.ErrorIs(t, err, entities.ErrDebtAccessForbidden)
			require.Len(t, audit.events, 1)
			assert.Equal(t, "student_debts.record_denied", audit.events[0].action)
		})
	}
}

func TestRecordResitResultUseCase_NoScheduledResit(t *testing.T) {
	open := debtWith(55, ptr(int64(999)), ptr(int64(7))) // status open, no attempts
	repo := &fakeDebtRepo{
		getByID: func(_ context.Context, _ int64) (*entities.StudentDebt, error) { return open, nil },
		update: func(_ context.Context, _ *entities.StudentDebt) error {
			t.Fatal("repo.Update must NOT be called when there is no scheduled resit")
			return nil
		},
	}
	uc := usecases.NewRecordResitResultUseCase(repo, &recordingAudit{}, fixedClock(), 2)

	_, err := uc.Execute(context.Background(), 7, "methodist", recordInput(1, entities.ResitResultPassed, ptr(5)))
	assert.ErrorIs(t, err, entities.ErrNoScheduledResit)
}

func TestRecordResitResultUseCase_RepoErrorsPropagate(t *testing.T) {
	debt := scheduledDebt(t)
	repo := &fakeDebtRepo{
		getByID: func(_ context.Context, _ int64) (*entities.StudentDebt, error) { return debt, nil },
		update: func(_ context.Context, _ *entities.StudentDebt) error {
			return repositories.ErrStudentDebtVersionConflict
		},
	}
	uc := usecases.NewRecordResitResultUseCase(repo, &recordingAudit{}, fixedClock(), 2)

	_, err := uc.Execute(context.Background(), 7, "methodist", recordInput(1, entities.ResitResultPassed, ptr(5)))
	assert.ErrorIs(t, err, repositories.ErrStudentDebtVersionConflict)
}
