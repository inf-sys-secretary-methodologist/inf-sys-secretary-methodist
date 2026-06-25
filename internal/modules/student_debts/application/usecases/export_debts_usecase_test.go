package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/repositories"
)

// fakeExportRepo is a function-backed double for the ListForExport port.
type fakeExportRepo struct {
	listForExport func(ctx context.Context, filter repositories.StudentDebtListFilter) ([]*entities.StudentDebt, error)
	calledWith    *repositories.StudentDebtListFilter
}

func (f *fakeExportRepo) ListForExport(ctx context.Context, filter repositories.StudentDebtListFilter) ([]*entities.StudentDebt, error) {
	f.calledWith = &filter
	return f.listForExport(ctx, filter)
}

// fakeExporter records the debts it was asked to serialize and returns a
// fixed document (or error).
type fakeExporter struct {
	data       []byte
	err        error
	calledWith []*entities.StudentDebt
	called     bool
}

func (f *fakeExporter) Export(_ context.Context, debts []*entities.StudentDebt) ([]byte, error) {
	f.called = true
	f.calledWith = debts
	return f.data, f.err
}

func TestExportDebtsUseCase_StaffExportsWholeRegistry(t *testing.T) {
	debts := []*entities.StudentDebt{debtWith(1, nil, nil), debtWith(2, nil, nil)}
	repo := &fakeExportRepo{listForExport: func(_ context.Context, _ repositories.StudentDebtListFilter) ([]*entities.StudentDebt, error) {
		return debts, nil
	}}
	exporter := &fakeExporter{data: []byte("XLSX-BYTES")}
	audit := &recordingAudit{}
	uc := usecases.NewExportDebtsUseCase(repo, &fakeTeacherScope{}, exporter, audit)

	filter := repositories.StudentDebtListFilter{GroupName: "ИВТ-21"}
	data, err := uc.Execute(context.Background(), 7, "methodist", filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != "XLSX-BYTES" {
		t.Fatalf("expected exporter bytes, got %q", data)
	}
	if repo.calledWith == nil || repo.calledWith.GroupName != "ИВТ-21" {
		t.Fatalf("expected repo called with inbound filter, got %+v", repo.calledWith)
	}
	if !exporter.called || len(exporter.calledWith) != 2 {
		t.Fatalf("expected exporter called with 2 debts, got called=%v len=%d", exporter.called, len(exporter.calledWith))
	}
	if len(audit.events) != 1 || audit.events[0].action != "student_debts.exported" {
		t.Fatalf("expected one exported audit event, got %+v", audit.events)
	}
	if got := audit.events[0].fields["count"]; got != 2 {
		t.Fatalf("expected audit count 2, got %v", got)
	}
}

func TestExportDebtsUseCase_TeacherScopedToOwnedDisciplines(t *testing.T) {
	repo := &fakeExportRepo{listForExport: func(_ context.Context, _ repositories.StudentDebtListFilter) ([]*entities.StudentDebt, error) {
		return []*entities.StudentDebt{debtWith(1, nil, ptr(int64(42)))}, nil
	}}
	exporter := &fakeExporter{data: []byte("DOC")}
	scope := &fakeTeacherScope{ids: []int64{42, 43}}
	uc := usecases.NewExportDebtsUseCase(repo, scope, exporter, &recordingAudit{})

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

func TestExportDebtsUseCase_TeacherWithNoDisciplinesExportsEmptyWithoutRepo(t *testing.T) {
	repo := &fakeExportRepo{listForExport: func(_ context.Context, _ repositories.StudentDebtListFilter) ([]*entities.StudentDebt, error) {
		t.Fatal("repo must not be queried for a teacher owning no disciplines")
		return nil, nil
	}}
	exporter := &fakeExporter{data: []byte("EMPTY")}
	audit := &recordingAudit{}
	uc := usecases.NewExportDebtsUseCase(repo, &fakeTeacherScope{ids: nil}, exporter, audit)

	data, err := uc.Execute(context.Background(), 9, "teacher", repositories.StudentDebtListFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != "EMPTY" {
		t.Fatalf("expected empty-registry document, got %q", data)
	}
	if !exporter.called || len(exporter.calledWith) != 0 {
		t.Fatalf("expected exporter called with empty slice, got called=%v len=%d", exporter.called, len(exporter.calledWith))
	}
	if got := audit.events[0].fields["count"]; got != 0 {
		t.Fatalf("expected audit count 0, got %v", got)
	}
}

func TestExportDebtsUseCase_DeniedRolesGetForbiddenNoSideEffects(t *testing.T) {
	for _, role := range []string{"student", "guest", ""} {
		t.Run(role, func(t *testing.T) {
			repo := &fakeExportRepo{listForExport: func(_ context.Context, _ repositories.StudentDebtListFilter) ([]*entities.StudentDebt, error) {
				t.Fatal("repo must not be queried on denial")
				return nil, nil
			}}
			exporter := &fakeExporter{}
			audit := &recordingAudit{}
			uc := usecases.NewExportDebtsUseCase(repo, &fakeTeacherScope{}, exporter, audit)

			_, err := uc.Execute(context.Background(), 5, role, repositories.StudentDebtListFilter{})
			if !errors.Is(err, entities.ErrDebtAccessForbidden) {
				t.Fatalf("expected ErrDebtAccessForbidden, got %v", err)
			}
			if exporter.called {
				t.Fatal("exporter must not run on denial")
			}
			if len(audit.events) != 1 || audit.events[0].action != "student_debts.export_denied" {
				t.Fatalf("expected one export_denied audit event, got %+v", audit.events)
			}
		})
	}
}

func TestExportDebtsUseCase_RepoErrorPropagatesAndSkipsExporter(t *testing.T) {
	sentinel := errors.New("db down")
	repo := &fakeExportRepo{listForExport: func(_ context.Context, _ repositories.StudentDebtListFilter) ([]*entities.StudentDebt, error) {
		return nil, sentinel
	}}
	exporter := &fakeExporter{}
	uc := usecases.NewExportDebtsUseCase(repo, &fakeTeacherScope{}, exporter, &recordingAudit{})

	_, err := uc.Execute(context.Background(), 7, "system_admin", repositories.StudentDebtListFilter{})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected wrapped repo error, got %v", err)
	}
	if exporter.called {
		t.Fatal("exporter must not run when the repo fails")
	}
}

func TestExportDebtsUseCase_ExporterErrorPropagates(t *testing.T) {
	sentinel := errors.New("encode failed")
	repo := &fakeExportRepo{listForExport: func(_ context.Context, _ repositories.StudentDebtListFilter) ([]*entities.StudentDebt, error) {
		return []*entities.StudentDebt{debtWith(1, nil, nil)}, nil
	}}
	exporter := &fakeExporter{err: sentinel}
	uc := usecases.NewExportDebtsUseCase(repo, &fakeTeacherScope{}, exporter, &recordingAudit{})

	_, err := uc.Execute(context.Background(), 7, "academic_secretary", repositories.StudentDebtListFilter{})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected wrapped exporter error, got %v", err)
	}
}

func TestNewExportDebtsUseCase_PanicsOnNilRequiredDeps(t *testing.T) {
	repo := &fakeExportRepo{}
	scope := &fakeTeacherScope{}
	exporter := &fakeExporter{}
	cases := map[string]func(){
		"nil repo":     func() { usecases.NewExportDebtsUseCase(nil, scope, exporter, nil) },
		"nil scope":    func() { usecases.NewExportDebtsUseCase(repo, nil, exporter, nil) },
		"nil exporter": func() { usecases.NewExportDebtsUseCase(repo, scope, nil, nil) },
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
