package usecases_test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/repositories"
)

// importStore is a stateful in-memory importDebtsRepo so idempotency can
// be exercised across two Execute calls (SourceHash skip on re-import).
type importStore struct {
	byID      map[int64]*entities.StudentDebt
	nextID    int64
	saveErr   error
	updateErr error
}

func newImportStore() *importStore { return &importStore{byID: map[int64]*entities.StudentDebt{}} }

func (s *importStore) Save(_ context.Context, d *entities.StudentDebt) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	s.nextID++
	d.ID = s.nextID
	s.byID[d.ID] = d
	return nil
}

func (s *importStore) Update(_ context.Context, d *entities.StudentDebt) error {
	if s.updateErr != nil {
		return s.updateErr
	}
	s.byID[d.ID] = d
	return nil
}

func (s *importStore) GetByID(_ context.Context, id int64) (*entities.StudentDebt, error) {
	if d, ok := s.byID[id]; ok {
		return d, nil
	}
	return nil, repositories.ErrStudentDebtNotFound
}

func (s *importStore) FindByIdentity(_ context.Context, group, student, discipline string, semester int) (*entities.StudentDebt, error) {
	for _, d := range s.byID {
		if d.GroupName == group && d.StudentFullName == student && d.DisciplineName == discipline && d.Semester == semester {
			return d, nil
		}
	}
	return nil, repositories.ErrStudentDebtNotFound
}

type fakeImporter struct {
	rows []usecases.ImportedDebt
	err  error
}

func (f *fakeImporter) Import(_ context.Context, _ io.Reader) ([]usecases.ImportedDebt, error) {
	return f.rows, f.err
}

func importedRow(id *int64, name, group, disc string, sem int, form string) usecases.ImportedDebt {
	return usecases.ImportedDebt{
		ServiceID: id, StudentFullName: name, GroupName: group,
		DisciplineName: disc, Semester: sem, ControlForm: form, SourceRef: "ved-1",
	}
}

func src() io.Reader { return strings.NewReader("doc") }

func TestImportDebtsUseCase_DeniedForNonManager(t *testing.T) {
	imp := &fakeImporter{}
	audit := &recordingAudit{}
	uc := usecases.NewImportDebtsUseCase(newImportStore(), imp, audit)

	_, err := uc.Execute(context.Background(), 1, "student", src())
	assert.ErrorIs(t, err, entities.ErrDebtAccessForbidden)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "student_debts.import_denied", audit.events[0].action)
	assert.Nil(t, imp.rows, "importer must not run for a denied actor")
}

func TestImportDebtsUseCase_ParseErrorPropagates(t *testing.T) {
	imp := &fakeImporter{err: errors.New("corrupt xlsx")}
	uc := usecases.NewImportDebtsUseCase(newImportStore(), imp, &recordingAudit{})

	_, err := uc.Execute(context.Background(), 1, "methodist", src())
	require.Error(t, err)
	assert.NotErrorIs(t, err, entities.ErrDebtAccessForbidden)
}

func TestImportDebtsUseCase_CreatesNewRows(t *testing.T) {
	imp := &fakeImporter{rows: []usecases.ImportedDebt{
		importedRow(nil, "Иванов Иван", "ИВТ-21", "Базы данных", 3, "exam"),
		importedRow(nil, "Петров Пётр", "ИВТ-21", "Сети", 4, "zachet"),
	}}
	store := newImportStore()
	audit := &recordingAudit{}
	uc := usecases.NewImportDebtsUseCase(store, imp, audit)

	res, err := uc.Execute(context.Background(), 1, "methodist", src())
	require.NoError(t, err)
	assert.Equal(t, 2, res.Created)
	assert.Equal(t, 0, res.Updated)
	assert.Empty(t, res.Errors)
	assert.Len(t, store.byID, 2)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "student_debts.imported", audit.events[0].action)
}

func TestImportDebtsUseCase_IdempotentSkipsUnchanged(t *testing.T) {
	rows := []usecases.ImportedDebt{importedRow(nil, "Иванов Иван", "ИВТ-21", "Базы данных", 3, "exam")}
	store := newImportStore()
	uc := usecases.NewImportDebtsUseCase(store, &fakeImporter{rows: rows}, &recordingAudit{})

	first, err := uc.Execute(context.Background(), 1, "methodist", src())
	require.NoError(t, err)
	assert.Equal(t, 1, first.Created)

	// Re-import the identical document — SourceHash matches → skipped.
	second, err := uc.Execute(context.Background(), 1, "methodist", src())
	require.NoError(t, err)
	assert.Equal(t, 0, second.Created)
	assert.Equal(t, 1, second.Skipped)
	assert.Len(t, store.byID, 1, "no duplicate created on re-import")
}

func TestImportDebtsUseCase_UpdatesByNaturalKey_NonKeyField(t *testing.T) {
	store := newImportStore()
	// Seed via a first import.
	uc := usecases.NewImportDebtsUseCase(store,
		&fakeImporter{rows: []usecases.ImportedDebt{importedRow(nil, "Иванов Иван", "ИВТ-21", "Базы данных", 3, "exam")}},
		&recordingAudit{})
	_, err := uc.Execute(context.Background(), 1, "methodist", src())
	require.NoError(t, err)

	// Same natural key (group/student/discipline/semester), corrected
	// control form (a non-key field) → matched by identity → updated.
	uc2 := usecases.NewImportDebtsUseCase(store,
		&fakeImporter{rows: []usecases.ImportedDebt{importedRow(nil, "Иванов Иван", "ИВТ-21", "Базы данных", 3, "differential_zachet")}},
		&recordingAudit{})
	res, err := uc2.Execute(context.Background(), 1, "methodist", src())
	require.NoError(t, err)
	assert.Equal(t, 1, res.Updated)
	assert.Len(t, store.byID, 1)
	for _, d := range store.byID {
		assert.Equal(t, entities.ControlFormDifferentialZachet, d.ControlForm)
	}
}

func TestImportDebtsUseCase_UpdatesByServiceID_CorrectsName(t *testing.T) {
	store := newImportStore()
	// Seed and capture the assigned service id.
	uc := usecases.NewImportDebtsUseCase(store,
		&fakeImporter{rows: []usecases.ImportedDebt{importedRow(nil, "Иванов И.", "ИВТ-21", "Базы данных", 3, "exam")}},
		&recordingAudit{})
	_, err := uc.Execute(context.Background(), 1, "methodist", src())
	require.NoError(t, err)
	require.Len(t, store.byID, 1)
	var id int64
	for k := range store.byID {
		id = k
	}

	// Re-import with the service id and a corrected full name (a key
	// field) → matched by id → updated, no duplicate.
	uc2 := usecases.NewImportDebtsUseCase(store,
		&fakeImporter{rows: []usecases.ImportedDebt{importedRow(&id, "Иванов Иван Иванович", "ИВТ-21", "Базы данных", 3, "exam")}},
		&recordingAudit{})
	res, err := uc2.Execute(context.Background(), 1, "methodist", src())
	require.NoError(t, err)
	assert.Equal(t, 1, res.Updated)
	assert.Len(t, store.byID, 1)
	assert.Equal(t, "Иванов Иван Иванович", store.byID[id].StudentFullName)
}

func TestImportDebtsUseCase_ServiceIDNotFound_RowError(t *testing.T) {
	missing := int64(404)
	imp := &fakeImporter{rows: []usecases.ImportedDebt{importedRow(&missing, "Иванов Иван", "ИВТ-21", "Базы данных", 3, "exam")}}
	uc := usecases.NewImportDebtsUseCase(newImportStore(), imp, &recordingAudit{})

	res, err := uc.Execute(context.Background(), 1, "methodist", src())
	require.NoError(t, err)
	assert.Equal(t, 0, res.Created)
	require.Len(t, res.Errors, 1)
	assert.Equal(t, 1, res.Errors[0].Row)
}

func TestImportDebtsUseCase_InvalidRow_RowError(t *testing.T) {
	imp := &fakeImporter{rows: []usecases.ImportedDebt{
		importedRow(nil, "Хороший Студент", "ИВТ-21", "Базы данных", 3, "exam"), // ok
		importedRow(nil, "Плохой Студент", "ИВТ-21", "Сети", 13, "exam"),        // semester out of range
		importedRow(nil, "Тоже Плохой", "ИВТ-21", "Графы", 2, "bogus_form"),     // bad control form
	}}
	store := newImportStore()
	uc := usecases.NewImportDebtsUseCase(store, imp, &recordingAudit{})

	res, err := uc.Execute(context.Background(), 1, "methodist", src())
	require.NoError(t, err)
	assert.Equal(t, 1, res.Created, "only the valid row is created")
	require.Len(t, res.Errors, 2)
	assert.Len(t, store.byID, 1)
}
