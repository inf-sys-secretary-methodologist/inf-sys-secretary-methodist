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

func TestRecordResitResultUseCase_PassedClosesDebt(t *testing.T) {
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
	assert.Same(t, debt, updated)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "student_debts.resit_recorded", audit.events[0].action)
}

func TestRecordResitResultUseCase_FailedBelowThresholdReopens(t *testing.T) {
	debt := scheduledDebt(t)
	repo := &fakeDebtRepo{
		getByID: func(_ context.Context, _ int64) (*entities.StudentDebt, error) { return debt, nil },
		update:  func(_ context.Context, _ *entities.StudentDebt) error { return nil },
	}
	uc := usecases.NewRecordResitResultUseCase(repo, &recordingAudit{}, fixedClock(), 2)

	got, err := uc.Execute(context.Background(), 7, "academic_secretary", recordInput(1, entities.ResitResultFailed, nil))
	require.NoError(t, err)
	assert.Equal(t, entities.DebtStatusOpen, got.Status(), "1 fail with N=2 returns to open")
}

func TestRecordResitResultUseCase_FailedReachingThresholdEscalates(t *testing.T) {
	debt := scheduledDebt(t)
	repo := &fakeDebtRepo{
		getByID: func(_ context.Context, _ int64) (*entities.StudentDebt, error) { return debt, nil },
		update:  func(_ context.Context, _ *entities.StudentDebt) error { return nil },
	}
	// N=1: a single regular failure escalates straight to commission.
	uc := usecases.NewRecordResitResultUseCase(repo, &recordingAudit{}, fixedClock(), 1)

	got, err := uc.Execute(context.Background(), 7, "system_admin", recordInput(1, entities.ResitResultFailed, nil))
	require.NoError(t, err)
	assert.Equal(t, entities.DebtStatusCommission, got.Status())
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
