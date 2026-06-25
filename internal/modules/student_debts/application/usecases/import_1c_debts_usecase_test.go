package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/entities"
)

// fakeDebtSource is a reader-free DebtSource: it yields pre-canned rows (or
// an error) the way the 1С OData adapter would.
type fakeDebtSource struct {
	rows   []usecases.ImportedDebt
	err    error
	called bool
}

func (f *fakeDebtSource) Fetch(_ context.Context) ([]usecases.ImportedDebt, error) {
	f.called = true
	return f.rows, f.err
}

func TestImport1CDebtsUseCase_DeniedForNonManager(t *testing.T) {
	src := &fakeDebtSource{}
	audit := &recordingAudit{}
	uc := usecases.NewImport1CDebtsUseCase(newImportStore(), src, audit)

	_, err := uc.Execute(context.Background(), 1, "student")
	assert.ErrorIs(t, err, entities.ErrDebtAccessForbidden)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "student_debts.import_1c_denied", audit.events[0].action)
	assert.False(t, src.called, "source must not be fetched for a denied actor")
}

func TestImport1CDebtsUseCase_FetchErrorPropagates(t *testing.T) {
	src := &fakeDebtSource{err: errors.New("1С unreachable")}
	uc := usecases.NewImport1CDebtsUseCase(newImportStore(), src, &recordingAudit{})

	_, err := uc.Execute(context.Background(), 1, "methodist")
	require.Error(t, err)
	assert.NotErrorIs(t, err, entities.ErrDebtAccessForbidden)
}

func TestImport1CDebtsUseCase_CreatesNewRows(t *testing.T) {
	src := &fakeDebtSource{rows: []usecases.ImportedDebt{
		importedRow(nil, "Кузнецов Дмитрий", "БИ-21", "Базы данных", 3, "exam"),
		importedRow(nil, "Смирнова Елена", "БИ-21", "Мат. анализ", 3, "zachet"),
	}}
	store := newImportStore()
	audit := &recordingAudit{}
	uc := usecases.NewImport1CDebtsUseCase(store, src, audit)

	res, err := uc.Execute(context.Background(), 7, "academic_secretary")
	require.NoError(t, err)
	assert.Equal(t, 2, res.Created)
	assert.Empty(t, res.Errors)
	assert.Len(t, store.byID, 2)
	require.Len(t, audit.events, 1)
	assert.Equal(t, "student_debts.imported_1c", audit.events[0].action)
	assert.True(t, src.called)
}

func TestImport1CDebtsUseCase_IdempotentSkipsUnchanged(t *testing.T) {
	rows := []usecases.ImportedDebt{
		importedRow(nil, "Волков Артём", "ПИ-22", "Программирование", 2, "exam"),
	}
	store := newImportStore()
	uc := usecases.NewImport1CDebtsUseCase(store, &fakeDebtSource{rows: rows}, &recordingAudit{})

	first, err := uc.Execute(context.Background(), 1, "methodist")
	require.NoError(t, err)
	assert.Equal(t, 1, first.Created)

	second, err := uc.Execute(context.Background(), 1, "methodist")
	require.NoError(t, err)
	assert.Equal(t, 0, second.Created)
	assert.Equal(t, 1, second.Skipped)
	assert.Len(t, store.byID, 1)
}
