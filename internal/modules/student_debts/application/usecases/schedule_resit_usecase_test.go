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

func fixedClock() func() time.Time {
	t := time.Date(2026, 6, 25, 10, 0, 0, 0, time.UTC)
	return func() time.Time { return t }
}

func scheduleInput() usecases.ScheduleResitInput {
	return usecases.ScheduleResitInput{
		DebtID:        55,
		ScheduledDate: time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC),
		Examiner:      "Петров П.П.",
	}
}

func TestScheduleResitUseCase_HappyPath_PersistsNotifiesAudits(t *testing.T) {
	debt := debtWith(55, ptr(int64(999)), ptr(int64(7)))
	var updated *entities.StudentDebt
	repo := &fakeDebtRepo{
		getByID: func(_ context.Context, _ int64) (*entities.StudentDebt, error) { return debt, nil },
		update:  func(_ context.Context, d *entities.StudentDebt) error { updated = d; return nil },
	}
	notifier := &fakeNotifier{}
	audit := &recordingAudit{}
	uc := usecases.NewScheduleResitUseCase(repo, notifier, audit, fixedClock())

	got, err := uc.Execute(context.Background(), 1, "methodist", scheduleInput())
	require.NoError(t, err)
	assert.Equal(t, entities.DebtStatusResitScheduled, got.Status())
	require.Len(t, got.Attempts(), 1)
	assert.Same(t, debt, updated, "must persist the mutated aggregate")

	require.Len(t, notifier.calls, 1)
	assert.Equal(t, int64(999), notifier.calls[0].studentUserID)
	assert.Equal(t, int64(55), notifier.calls[0].debtID)

	require.Len(t, audit.events, 1)
	assert.Equal(t, "student_debts.resit_scheduled", audit.events[0].action)
}

func TestScheduleResitUseCase_NotifierSkippedWhenNoStudentLink(t *testing.T) {
	debt := debtWith(55, nil, ptr(int64(7))) // no local student account
	repo := &fakeDebtRepo{
		getByID: func(_ context.Context, _ int64) (*entities.StudentDebt, error) { return debt, nil },
		update:  func(_ context.Context, _ *entities.StudentDebt) error { return nil },
	}
	notifier := &fakeNotifier{}
	uc := usecases.NewScheduleResitUseCase(repo, notifier, &recordingAudit{}, fixedClock())

	_, err := uc.Execute(context.Background(), 1, "methodist", scheduleInput())
	require.NoError(t, err)
	assert.Empty(t, notifier.calls, "no notification without a resolved student account")
}

func TestScheduleResitUseCase_DeniedForNonManager(t *testing.T) {
	for _, role := range []string{"teacher", "student", "guest"} {
		t.Run(role, func(t *testing.T) {
			repo := &fakeDebtRepo{
				getByID: func(_ context.Context, _ int64) (*entities.StudentDebt, error) {
					t.Fatal("repo.GetByID must NOT be called for a denied actor")
					return nil, nil
				},
			}
			audit := &recordingAudit{}
			uc := usecases.NewScheduleResitUseCase(repo, &fakeNotifier{}, audit, fixedClock())

			_, err := uc.Execute(context.Background(), 1, role, scheduleInput())
			assert.ErrorIs(t, err, entities.ErrDebtAccessForbidden)
			require.Len(t, audit.events, 1)
			assert.Equal(t, "student_debts.schedule_denied", audit.events[0].action)
		})
	}
}

func TestScheduleResitUseCase_ClosedDebt_DomainErrorNotPersisted(t *testing.T) {
	closed := entities.ReconstituteStudentDebt(55, "Иванов", "ИВТ-21", "БД", 3,
		entities.ControlFormExam, ptr(int64(999)), ptr(int64(7)), "", "", 2,
		entities.DebtStatusClosedPassed, nil, time.Now(), time.Now())
	repo := &fakeDebtRepo{
		getByID: func(_ context.Context, _ int64) (*entities.StudentDebt, error) { return closed, nil },
		update: func(_ context.Context, _ *entities.StudentDebt) error {
			t.Fatal("repo.Update must NOT be called when the FSM rejects the transition")
			return nil
		},
	}
	uc := usecases.NewScheduleResitUseCase(repo, &fakeNotifier{}, &recordingAudit{}, fixedClock())

	_, err := uc.Execute(context.Background(), 1, "methodist", scheduleInput())
	assert.ErrorIs(t, err, entities.ErrDebtClosed)
}

func TestScheduleResitUseCase_RepoErrorsPropagate(t *testing.T) {
	t.Run("get by id", func(t *testing.T) {
		repo := &fakeDebtRepo{getByID: func(_ context.Context, _ int64) (*entities.StudentDebt, error) {
			return nil, repositories.ErrStudentDebtNotFound
		}}
		uc := usecases.NewScheduleResitUseCase(repo, &fakeNotifier{}, &recordingAudit{}, fixedClock())
		_, err := uc.Execute(context.Background(), 1, "methodist", scheduleInput())
		assert.ErrorIs(t, err, repositories.ErrStudentDebtNotFound)
	})

	t.Run("update version conflict", func(t *testing.T) {
		debt := debtWith(55, ptr(int64(999)), ptr(int64(7)))
		notifier := &fakeNotifier{}
		repo := &fakeDebtRepo{
			getByID: func(_ context.Context, _ int64) (*entities.StudentDebt, error) { return debt, nil },
			update: func(_ context.Context, _ *entities.StudentDebt) error {
				return repositories.ErrStudentDebtVersionConflict
			},
		}
		uc := usecases.NewScheduleResitUseCase(repo, notifier, &recordingAudit{}, fixedClock())
		_, err := uc.Execute(context.Background(), 1, "methodist", scheduleInput())
		assert.ErrorIs(t, err, repositories.ErrStudentDebtVersionConflict)
		assert.Empty(t, notifier.calls, "no notification when the persist failed")
	})
}
