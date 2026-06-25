package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/repositories"
)

// fakeStatsRepo is a function-backed double for the Stats port.
type fakeStatsRepo struct {
	stats      func(ctx context.Context, filter repositories.StudentDebtListFilter) (repositories.StudentDebtStats, error)
	calledWith *repositories.StudentDebtListFilter
}

func (f *fakeStatsRepo) Stats(ctx context.Context, filter repositories.StudentDebtListFilter) (repositories.StudentDebtStats, error) {
	f.calledWith = &filter
	return f.stats(ctx, filter)
}

func TestGetDebtStatsUseCase_StaffGetsAggregate(t *testing.T) {
	want := repositories.StudentDebtStats{
		Total: 9, Open: 4, ResitScheduled: 2, Commission: 1, ClosedPassed: 1, ClosedFailed: 1,
	}
	repo := &fakeStatsRepo{stats: func(_ context.Context, _ repositories.StudentDebtListFilter) (repositories.StudentDebtStats, error) {
		return want, nil
	}}
	uc := usecases.NewGetDebtStatsUseCase(repo, &fakeTeacherScope{}, &recordingAudit{})

	got, err := uc.Execute(context.Background(), 7, "methodist", repositories.StudentDebtListFilter{GroupName: "ИВТ-21"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Fatalf("expected stats %+v, got %+v", want, got)
	}
	if repo.calledWith == nil || repo.calledWith.GroupName != "ИВТ-21" {
		t.Fatalf("expected repo called with inbound filter, got %+v", repo.calledWith)
	}
}

func TestGetDebtStatsUseCase_TeacherScopedToOwnedDisciplines(t *testing.T) {
	repo := &fakeStatsRepo{stats: func(_ context.Context, _ repositories.StudentDebtListFilter) (repositories.StudentDebtStats, error) {
		return repositories.StudentDebtStats{Total: 3}, nil
	}}
	scope := &fakeTeacherScope{ids: []int64{42, 43}}
	uc := usecases.NewGetDebtStatsUseCase(repo, scope, &recordingAudit{})

	// Client tries to widen scope; teacher scope must override it.
	_, err := uc.Execute(context.Background(), 9, "teacher", repositories.StudentDebtListFilter{DisciplineIDs: []int64{999}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !scope.called || scope.askedWith != 9 {
		t.Fatalf("expected teacher scope resolved for actor 9, got called=%v askedWith=%d", scope.called, scope.askedWith)
	}
	if repo.calledWith == nil || len(repo.calledWith.DisciplineIDs) != 2 ||
		repo.calledWith.DisciplineIDs[0] != 42 || repo.calledWith.DisciplineIDs[1] != 43 {
		t.Fatalf("expected repo filter forced to owned disciplines, got %+v", repo.calledWith)
	}
}

func TestGetDebtStatsUseCase_TeacherWithNoDisciplinesGetsZeroWithoutRepo(t *testing.T) {
	repo := &fakeStatsRepo{stats: func(_ context.Context, _ repositories.StudentDebtListFilter) (repositories.StudentDebtStats, error) {
		t.Fatal("repo must not be queried for a teacher owning no disciplines")
		return repositories.StudentDebtStats{}, nil
	}}
	uc := usecases.NewGetDebtStatsUseCase(repo, &fakeTeacherScope{ids: nil}, &recordingAudit{})

	got, err := uc.Execute(context.Background(), 9, "teacher", repositories.StudentDebtListFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != (repositories.StudentDebtStats{}) {
		t.Fatalf("expected zero stats, got %+v", got)
	}
}

func TestGetDebtStatsUseCase_DeniedRolesGetForbiddenNoSideEffects(t *testing.T) {
	for _, role := range []string{"student", "guest", ""} {
		t.Run(role, func(t *testing.T) {
			repo := &fakeStatsRepo{stats: func(_ context.Context, _ repositories.StudentDebtListFilter) (repositories.StudentDebtStats, error) {
				t.Fatal("repo must not be queried on denial")
				return repositories.StudentDebtStats{}, nil
			}}
			audit := &recordingAudit{}
			uc := usecases.NewGetDebtStatsUseCase(repo, &fakeTeacherScope{}, audit)

			_, err := uc.Execute(context.Background(), 5, role, repositories.StudentDebtListFilter{})
			if !errors.Is(err, entities.ErrDebtAccessForbidden) {
				t.Fatalf("expected ErrDebtAccessForbidden, got %v", err)
			}
			if len(audit.events) != 1 || audit.events[0].action != "student_debts.stats_denied" {
				t.Fatalf("expected one stats_denied audit event, got %+v", audit.events)
			}
		})
	}
}

func TestGetDebtStatsUseCase_RepoErrorPropagates(t *testing.T) {
	sentinel := errors.New("db down")
	repo := &fakeStatsRepo{stats: func(_ context.Context, _ repositories.StudentDebtListFilter) (repositories.StudentDebtStats, error) {
		return repositories.StudentDebtStats{}, sentinel
	}}
	uc := usecases.NewGetDebtStatsUseCase(repo, &fakeTeacherScope{}, &recordingAudit{})

	_, err := uc.Execute(context.Background(), 7, "system_admin", repositories.StudentDebtListFilter{})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected wrapped repo error, got %v", err)
	}
}

func TestNewGetDebtStatsUseCase_PanicsOnNilRequiredDeps(t *testing.T) {
	repo := &fakeStatsRepo{}
	scope := &fakeTeacherScope{}
	cases := map[string]func(){
		"nil repo":  func() { usecases.NewGetDebtStatsUseCase(nil, scope, nil) },
		"nil scope": func() { usecases.NewGetDebtStatsUseCase(repo, nil, nil) },
	}
	for name, build := range cases {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("expected panic on nil required dependency")
				}
			}()
			build()
		})
	}
}
