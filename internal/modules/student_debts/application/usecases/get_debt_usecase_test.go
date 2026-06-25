package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/repositories"
)

func TestGetDebtUseCase_Authorization(t *testing.T) {
	const actorID = int64(100)
	// Debt owned by student 999, linked to discipline 7.
	debt := debtWith(55, ptr(int64(999)), ptr(int64(7)))

	tests := []struct {
		name           string
		role           string
		teacherDiscIDs []int64
		wantOK         bool
		wantDenied     bool // expect an audit denial event
	}{
		{name: "system_admin sees any debt", role: "system_admin", wantOK: true},
		{name: "methodist sees any debt", role: "methodist", wantOK: true},
		{name: "secretary sees any debt", role: "academic_secretary", wantOK: true},
		{name: "teacher owning discipline sees debt", role: "teacher", teacherDiscIDs: []int64{3, 7}, wantOK: true},
		{name: "teacher not owning discipline denied", role: "teacher", teacherDiscIDs: []int64{1, 2}, wantDenied: true},
		{name: "teacher with no disciplines denied", role: "teacher", teacherDiscIDs: nil, wantDenied: true},
		{name: "student not owner denied", role: "student", wantDenied: true},
		{name: "unknown role denied", role: "guest", wantDenied: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeDebtRepo{getByID: func(_ context.Context, id int64) (*entities.StudentDebt, error) {
				require.Equal(t, int64(55), id)
				return debt, nil
			}}
			scope := &fakeTeacherScope{ids: tt.teacherDiscIDs}
			audit := &recordingAudit{}
			uc := usecases.NewGetDebtUseCase(repo, scope, audit)

			got, err := uc.Execute(context.Background(), actorID, tt.role, 55)

			if tt.wantOK {
				require.NoError(t, err)
				assert.Same(t, debt, got)
				assert.Empty(t, audit.events, "successful read must not emit denial audit")
				return
			}
			assert.Nil(t, got)
			assert.ErrorIs(t, err, entities.ErrDebtAccessForbidden)
			if tt.wantDenied {
				require.Len(t, audit.events, 1)
				assert.Equal(t, "student_debts.view_denied", audit.events[0].action)
			}
		})
	}
}

func TestGetDebtUseCase_StudentSeesOwnDebt(t *testing.T) {
	const actorID = int64(999)
	debt := debtWith(55, ptr(int64(999)), ptr(int64(7)))
	repo := &fakeDebtRepo{getByID: func(_ context.Context, _ int64) (*entities.StudentDebt, error) {
		return debt, nil
	}}
	uc := usecases.NewGetDebtUseCase(repo, &fakeTeacherScope{}, &recordingAudit{})

	got, err := uc.Execute(context.Background(), actorID, "student", 55)
	require.NoError(t, err)
	assert.Same(t, debt, got)
}

func TestGetDebtUseCase_RepoNotFoundPropagates(t *testing.T) {
	repo := &fakeDebtRepo{getByID: func(_ context.Context, _ int64) (*entities.StudentDebt, error) {
		return nil, repositories.ErrStudentDebtNotFound
	}}
	audit := &recordingAudit{}
	uc := usecases.NewGetDebtUseCase(repo, &fakeTeacherScope{}, audit)

	_, err := uc.Execute(context.Background(), 1, "methodist", 404)
	assert.ErrorIs(t, err, repositories.ErrStudentDebtNotFound)
	assert.Empty(t, audit.events, "not-found reads must not flood the audit log")
}

func TestGetDebtUseCase_TeacherScopeErrorPropagates(t *testing.T) {
	debt := debtWith(55, ptr(int64(999)), ptr(int64(7)))
	repo := &fakeDebtRepo{getByID: func(_ context.Context, _ int64) (*entities.StudentDebt, error) {
		return debt, nil
	}}
	scope := &fakeTeacherScope{err: errors.New("curriculum down")}
	uc := usecases.NewGetDebtUseCase(repo, scope, &recordingAudit{})

	_, err := uc.Execute(context.Background(), 100, "teacher", 55)
	require.Error(t, err)
	assert.NotErrorIs(t, err, entities.ErrDebtAccessForbidden,
		"a resolver failure is an infra error, not an access denial")
}
