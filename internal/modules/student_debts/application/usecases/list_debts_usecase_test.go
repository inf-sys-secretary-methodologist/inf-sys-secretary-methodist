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

func okList(total int) repositories.StudentDebtListResult {
	return repositories.StudentDebtListResult{Total: total}
}

func TestListDebtsUseCase_StaffPassThrough(t *testing.T) {
	var seen repositories.StudentDebtListFilter
	repo := &fakeDebtRepo{list: func(_ context.Context, f repositories.StudentDebtListFilter) (repositories.StudentDebtListResult, error) {
		seen = f
		return okList(5), nil
	}}
	scope := &fakeTeacherScope{}
	uc := usecases.NewListDebtsUseCase(repo, scope, &recordingAudit{})

	in := repositories.StudentDebtListFilter{GroupName: "ИВТ-21", Limit: 20}
	res, err := uc.Execute(context.Background(), 1, "methodist", in)

	require.NoError(t, err)
	assert.Equal(t, 5, res.Total)
	assert.Equal(t, "ИВТ-21", seen.GroupName)
	assert.Nil(t, seen.DisciplineIDs, "staff filter must not be discipline-scoped")
	assert.False(t, scope.called, "staff path must not consult the teacher scope")
}

func TestListDebtsUseCase_TeacherScopedToOwnedDisciplines(t *testing.T) {
	var seen repositories.StudentDebtListFilter
	repo := &fakeDebtRepo{list: func(_ context.Context, f repositories.StudentDebtListFilter) (repositories.StudentDebtListResult, error) {
		seen = f
		return okList(2), nil
	}}
	scope := &fakeTeacherScope{ids: []int64{3, 7}}
	uc := usecases.NewListDebtsUseCase(repo, scope, &recordingAudit{})

	// Client tries to smuggle a foreign discipline id — must be overridden.
	in := repositories.StudentDebtListFilter{DisciplineIDs: []int64{999}, Limit: 20}
	res, err := uc.Execute(context.Background(), 100, "teacher", in)

	require.NoError(t, err)
	assert.Equal(t, 2, res.Total)
	assert.Equal(t, []int64{3, 7}, seen.DisciplineIDs, "teacher scope must override client DisciplineIDs")
	assert.Equal(t, int64(100), scope.askedWith)
}

func TestListDebtsUseCase_TeacherWithNoDisciplinesSeesNothing(t *testing.T) {
	repo := &fakeDebtRepo{list: func(_ context.Context, _ repositories.StudentDebtListFilter) (repositories.StudentDebtListResult, error) {
		t.Fatal("repo.List must NOT be called for a teacher with no disciplines")
		return repositories.StudentDebtListResult{}, nil
	}}
	scope := &fakeTeacherScope{ids: nil}
	uc := usecases.NewListDebtsUseCase(repo, scope, &recordingAudit{})

	res, err := uc.Execute(context.Background(), 100, "teacher", repositories.StudentDebtListFilter{Limit: 20})
	require.NoError(t, err)
	assert.Equal(t, 0, res.Total)
	assert.Empty(t, res.Items)
}

func TestListDebtsUseCase_StudentDenied(t *testing.T) {
	repo := &fakeDebtRepo{list: func(_ context.Context, _ repositories.StudentDebtListFilter) (repositories.StudentDebtListResult, error) {
		t.Fatal("repo.List must NOT be called for a denied student")
		return repositories.StudentDebtListResult{}, nil
	}}
	audit := &recordingAudit{}
	uc := usecases.NewListDebtsUseCase(repo, &fakeTeacherScope{}, audit)

	_, err := uc.Execute(context.Background(), 999, "student", repositories.StudentDebtListFilter{Limit: 20})
	assert.ErrorIs(t, err, entities.ErrDebtAccessForbidden)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "student_debts.list_denied", audit.events[0].action)
}

func TestListDebtsUseCase_TeacherScopeErrorPropagates(t *testing.T) {
	repo := &fakeDebtRepo{list: func(_ context.Context, _ repositories.StudentDebtListFilter) (repositories.StudentDebtListResult, error) {
		return okList(0), nil
	}}
	scope := &fakeTeacherScope{err: errors.New("curriculum down")}
	uc := usecases.NewListDebtsUseCase(repo, scope, &recordingAudit{})

	_, err := uc.Execute(context.Background(), 100, "teacher", repositories.StudentDebtListFilter{Limit: 20})
	require.Error(t, err)
	assert.NotErrorIs(t, err, entities.ErrDebtAccessForbidden)
}

func TestListMyDebtsUseCase_ForcesActorAsStudent(t *testing.T) {
	var seen repositories.StudentDebtListFilter
	repo := &fakeDebtRepo{list: func(_ context.Context, f repositories.StudentDebtListFilter) (repositories.StudentDebtListResult, error) {
		seen = f
		return okList(3), nil
	}}
	uc := usecases.NewListMyDebtsUseCase(repo)

	// Client tries to read another user's debts — must be overridden.
	in := repositories.StudentDebtListFilter{StudentUserID: ptr(int64(1)), Limit: 20}
	res, err := uc.Execute(context.Background(), 999, in)

	require.NoError(t, err)
	assert.Equal(t, 3, res.Total)
	require.NotNil(t, seen.StudentUserID)
	assert.Equal(t, int64(999), *seen.StudentUserID, "ListMyDebts must force StudentUserID to the actor")
}
