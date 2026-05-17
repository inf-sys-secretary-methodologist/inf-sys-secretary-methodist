package persistence

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// v0.153.1 #196 coverage push: reference_repository_pg.go went from 0%
// → full happy + error coverage. The repo is a thin SELECT wrapper, so
// tests pin column lists + error wrapping; no business logic.

func newReferenceRepoMock(t *testing.T) (*ReferenceRepositoryPG, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewReferenceRepositoryPG(db), mock
}

// --- ListStudentGroups ---

func TestReferenceRepoListStudentGroups_NoLimitReturnsAll(t *testing.T) {
	repo, mock := newReferenceRepoMock(t)
	rows := sqlmock.NewRows([]string{"id", "specialty_id", "name", "course", "curator_id", "capacity"}).
		AddRow(int64(1), int64(10), "ИС-21", 2, int64(7), 25).
		AddRow(int64(2), int64(10), "ИС-22", 2, nil, 22)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, specialty_id, name, course, curator_id, capacity")).
		WillReturnRows(rows)

	groups, err := repo.ListStudentGroups(context.Background(), 0, 0)
	require.NoError(t, err)
	require.Len(t, groups, 2)
	assert.Equal(t, "ИС-21", groups[0].Name)
	assert.NotNil(t, groups[0].CuratorID)
	assert.Equal(t, int64(7), *groups[0].CuratorID)
	assert.Nil(t, groups[1].CuratorID, "nullable curator_id NULL must scan to nil pointer")
}

func TestReferenceRepoListStudentGroups_LimitAddsPagination(t *testing.T) {
	repo, mock := newReferenceRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("LIMIT $1 OFFSET $2")).
		WithArgs(10, 20).
		WillReturnRows(sqlmock.NewRows([]string{"id", "specialty_id", "name", "course", "curator_id", "capacity"}))

	groups, err := repo.ListStudentGroups(context.Background(), 10, 20)
	require.NoError(t, err)
	assert.Empty(t, groups)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReferenceRepoListStudentGroups_QueryError(t *testing.T) {
	repo, mock := newReferenceRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, specialty_id")).
		WillReturnError(errors.New("connection refused"))

	groups, err := repo.ListStudentGroups(context.Background(), 0, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list student groups")
	assert.Nil(t, groups)
}

func TestReferenceRepoListStudentGroups_ScanError(t *testing.T) {
	repo, mock := newReferenceRepoMock(t)
	// Row width mismatch — Scan returns error и use case wraps.
	rows := sqlmock.NewRows([]string{"id", "specialty_id", "name"}).
		AddRow(int64(1), int64(10), "ИС-21")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, specialty_id")).WillReturnRows(rows)

	groups, err := repo.ListStudentGroups(context.Background(), 0, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to scan student group")
	assert.Nil(t, groups)
}

// --- ListDisciplines ---

func TestReferenceRepoListDisciplines_HappyPath(t *testing.T) {
	repo, mock := newReferenceRepoMock(t)
	code := "DISC-101"
	credits := 3
	hoursTotal := 72
	rows := sqlmock.NewRows([]string{
		"id", "name", "code", "department_id", "credits",
		"hours_total", "hours_lectures", "hours_practice", "hours_labs",
	}).AddRow(int64(1), "Высшая математика", &code, int64(5), &credits, &hoursTotal, nil, nil, nil)

	mock.ExpectQuery(regexp.QuoteMeta("FROM disciplines ORDER BY name")).WillReturnRows(rows)

	disciplines, err := repo.ListDisciplines(context.Background(), 0, 0)
	require.NoError(t, err)
	require.Len(t, disciplines, 1)
	assert.Equal(t, "Высшая математика", disciplines[0].Name)
	require.NotNil(t, disciplines[0].Credits)
	assert.Equal(t, 3, *disciplines[0].Credits)
}

func TestReferenceRepoListDisciplines_LimitAddsPagination(t *testing.T) {
	repo, mock := newReferenceRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("LIMIT $1 OFFSET $2")).
		WithArgs(5, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "name", "code", "department_id", "credits",
			"hours_total", "hours_lectures", "hours_practice", "hours_labs",
		}))

	disciplines, err := repo.ListDisciplines(context.Background(), 5, 0)
	require.NoError(t, err)
	assert.Empty(t, disciplines)
}

func TestReferenceRepoListDisciplines_QueryError(t *testing.T) {
	repo, mock := newReferenceRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM disciplines")).
		WillReturnError(errors.New("table missing"))

	disciplines, err := repo.ListDisciplines(context.Background(), 0, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list disciplines")
	assert.Nil(t, disciplines)
}

// --- ListSemesters ---

func TestReferenceRepoListSemesters_AllReturnsBoth(t *testing.T) {
	repo, mock := newReferenceRepoMock(t)
	start := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2027, 1, 31, 0, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{"id", "academic_year_id", "name", "number", "start_date", "end_date", "is_active"}).
		AddRow(int64(1), int64(1), "Осень 2026", 1, start, end, true).
		AddRow(int64(2), int64(1), "Весна 2026", 2, start.AddDate(0, 6, 0), end.AddDate(0, 6, 0), false)

	mock.ExpectQuery(regexp.QuoteMeta("FROM semesters")).
		WillReturnRows(rows)

	semesters, err := repo.ListSemesters(context.Background(), false)
	require.NoError(t, err)
	require.Len(t, semesters, 2)
	assert.True(t, semesters[0].IsActive)
	assert.False(t, semesters[1].IsActive)
}

func TestReferenceRepoListSemesters_ActiveOnlyAddsFilter(t *testing.T) {
	repo, mock := newReferenceRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE is_active = true")).
		WillReturnRows(sqlmock.NewRows([]string{"id", "academic_year_id", "name", "number", "start_date", "end_date", "is_active"}).
			AddRow(int64(1), int64(1), "Осень 2026", 1, time.Now(), time.Now().AddDate(0, 5, 0), true))

	semesters, err := repo.ListSemesters(context.Background(), true)
	require.NoError(t, err)
	require.Len(t, semesters, 1)
	assert.True(t, semesters[0].IsActive)
}

func TestReferenceRepoListSemesters_QueryError(t *testing.T) {
	repo, mock := newReferenceRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM semesters")).
		WillReturnError(errors.New("pg unavailable"))

	semesters, err := repo.ListSemesters(context.Background(), false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list semesters")
	assert.Nil(t, semesters)
}

// --- ListLessonTypes ---

func TestReferenceRepoListLessonTypes_HappyPath(t *testing.T) {
	repo, mock := newReferenceRepoMock(t)
	color := "#FF0000"
	rows := sqlmock.NewRows([]string{"id", "name", "short_name", "color"}).
		AddRow(int64(1), "Лекция", "ЛК", &color).
		AddRow(int64(2), "Практика", "ПР", nil)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, short_name, color FROM lesson_types")).
		WillReturnRows(rows)

	types, err := repo.ListLessonTypes(context.Background())
	require.NoError(t, err)
	require.Len(t, types, 2)
	assert.Equal(t, "Лекция", types[0].Name)
	require.NotNil(t, types[0].Color)
	assert.Equal(t, "#FF0000", *types[0].Color)
	assert.Nil(t, types[1].Color)
}

func TestReferenceRepoListLessonTypes_QueryError(t *testing.T) {
	repo, mock := newReferenceRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM lesson_types")).
		WillReturnError(errors.New("rollback"))

	types, err := repo.ListLessonTypes(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list lesson types")
	assert.Nil(t, types)
}

// --- GetActiveSemester ---

func TestReferenceRepoGetActiveSemester_HappyPath(t *testing.T) {
	repo, mock := newReferenceRepoMock(t)
	start := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2027, 1, 31, 0, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{"id", "academic_year_id", "name", "number", "start_date", "end_date", "is_active"}).
		AddRow(int64(1), int64(1), "Осень 2026", 1, start, end, true)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE is_active = true LIMIT 1")).
		WillReturnRows(rows)

	sem, err := repo.GetActiveSemester(context.Background())
	require.NoError(t, err)
	require.NotNil(t, sem)
	assert.True(t, sem.IsActive)
	assert.Equal(t, "Осень 2026", sem.Name)
}

func TestReferenceRepoGetActiveSemester_NoRowsReturnsNilNilError(t *testing.T) {
	repo, mock := newReferenceRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE is_active = true")).
		WillReturnError(sql.ErrNoRows)

	sem, err := repo.GetActiveSemester(context.Background())
	require.NoError(t, err, "ErrNoRows должен collapse в (nil, nil) — анти-enumeration shape")
	assert.Nil(t, sem)
}

func TestReferenceRepoGetActiveSemester_QueryError(t *testing.T) {
	repo, mock := newReferenceRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE is_active = true")).
		WillReturnError(errors.New("connection broken"))

	sem, err := repo.GetActiveSemester(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get active semester")
	assert.Nil(t, sem)
}

// --- Constructor ---

func TestNewReferenceRepositoryPG_StoresHandle(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := NewReferenceRepositoryPG(db)
	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}
