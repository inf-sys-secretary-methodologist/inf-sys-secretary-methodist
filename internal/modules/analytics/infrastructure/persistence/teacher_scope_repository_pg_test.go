package persistence

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTeacherScopeRepoMock(t *testing.T) (*TeacherScopeRepositoryPG, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewTeacherScopeRepositoryPG(db), mock
}

// SQL contract — the repository must read authoritative scheduling data
// (schedule_lessons + student_groups) and return DISTINCT group names
// the teacher is assigned to. The cross-table read is a persistence
// detail; analytics does not import the schedule Go module.

func TestTeacherScopeRepository_ListGroupNames_ReturnsDistinctGroupsForTeacher(t *testing.T) {
	repo, mock := newTeacherScopeRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT DISTINCT sg.name FROM schedule_lessons sl JOIN student_groups sg ON sg.id = sl.group_id WHERE sl.teacher_id = $1")).
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"name"}).
			AddRow("ИС-21").
			AddRow("ПИ-31"))

	got, err := repo.ListGroupNames(context.Background(), 7)
	require.NoError(t, err)
	assert.Equal(t, []string{"ИС-21", "ПИ-31"}, got)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTeacherScopeRepository_ListGroupNames_EmptyForUnassignedTeacher(t *testing.T) {
	repo, mock := newTeacherScopeRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("schedule_lessons")).
		WithArgs(int64(99)).
		WillReturnRows(sqlmock.NewRows([]string{"name"}))

	got, err := repo.ListGroupNames(context.Background(), 99)
	require.NoError(t, err)
	assert.Empty(t, got, "teacher with no scheduled lessons must yield empty whitelist (deny-all scope)")
}

func TestTeacherScopeRepository_ListGroupNames_QueryError(t *testing.T) {
	repo, mock := newTeacherScopeRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("schedule_lessons")).
		WithArgs(int64(7)).
		WillReturnError(fmt.Errorf("connection refused"))

	got, err := repo.ListGroupNames(context.Background(), 7)
	assert.Error(t, err)
	assert.Nil(t, got)
}
