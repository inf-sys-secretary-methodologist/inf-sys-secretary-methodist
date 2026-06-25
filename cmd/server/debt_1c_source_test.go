package main

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	integrationEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/domain/entities"
)

type fakeDebtCatalog struct {
	rows []integrationEntities.ODataStudentDebt
	err  error
}

func (f fakeDebtCatalog) GetAllStudentDebts(_ context.Context) ([]integrationEntities.ODataStudentDebt, error) {
	return f.rows, f.err
}

func TestDebt1CSource_MapsRowsAndControlForms(t *testing.T) {
	src := debt1CSource{catalog: fakeDebtCatalog{rows: []integrationEntities.ODataStudentDebt{
		{RefKey: "d1", StudentName: "Кузнецов Дмитрий", GroupName: "БИ-21", Discipline: "Базы данных", Semester: 3, ControlForm: "Экзамен"},
		{RefKey: "d2", StudentName: "Смирнова Елена", GroupName: "БИ-21", Discipline: "Мат. анализ", Semester: 3, ControlForm: "Зачёт"},
		{RefKey: "d3", StudentName: "Волков Артём", GroupName: "ПИ-22", Discipline: "Программирование", Semester: 2, ControlForm: "Дифференцированный зачёт"},
		{RefKey: "d4", StudentName: "Волков Артём", GroupName: "ПИ-22", Discipline: "Архитектура ЭВМ", Semester: 2, ControlForm: "Курсовой проект"},
	}}}

	rows, err := src.Fetch(context.Background())
	require.NoError(t, err)
	require.Len(t, rows, 4)

	assert.Equal(t, "Кузнецов Дмитрий", rows[0].StudentFullName)
	assert.Equal(t, "БИ-21", rows[0].GroupName)
	assert.Equal(t, "Базы данных", rows[0].DisciplineName)
	assert.Equal(t, 3, rows[0].Semester)
	assert.Equal(t, "d1", rows[0].SourceRef, "1С Ref_Key carried as the source reference")

	assert.Equal(t, "exam", rows[0].ControlForm)
	assert.Equal(t, "zachet", rows[1].ControlForm)
	assert.Equal(t, "differential_zachet", rows[2].ControlForm)
	assert.Equal(t, "course_project", rows[3].ControlForm)
}

func TestDebt1CSource_SkipsDeletionMarked(t *testing.T) {
	src := debt1CSource{catalog: fakeDebtCatalog{rows: []integrationEntities.ODataStudentDebt{
		{RefKey: "live", StudentName: "Активный Долг", GroupName: "БИ-21", Discipline: "БД", Semester: 3, ControlForm: "Экзамен"},
		{RefKey: "gone", DeletionMark: true, StudentName: "Удалённый", GroupName: "БИ-21", Discipline: "БД", Semester: 3, ControlForm: "Экзамен"},
	}}}

	rows, err := src.Fetch(context.Background())
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, "live", rows[0].SourceRef)
}

func TestDebt1CSource_FetchErrorPropagates(t *testing.T) {
	src := debt1CSource{catalog: fakeDebtCatalog{err: errors.New("1С down")}}
	_, err := src.Fetch(context.Background())
	require.Error(t, err)
}

func TestControlFormFrom1C_UnknownPassesThrough(t *testing.T) {
	// An unrecognized label must pass through so the domain rejects it,
	// rather than being silently coerced to a valid form.
	assert.Equal(t, "Лабораторная", controlFormFrom1C("Лабораторная"))
}
