package persistence

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTeacherScopeMock(t *testing.T) (*TeacherScopeResolverPG, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewTeacherScopeResolverPG(db), mock
}

func TestTeacherScopeResolverPG_ReturnsDistinctDisciplineIDsForTeacher(t *testing.T) {
	repo, mock := newTeacherScopeMock(t)

	// The teacher's disciplines come from the schedule: the disciplines they
	// are scheduled to teach. DISTINCT collapses repeated lessons of the same
	// discipline; the ids are disciplines(id), matching student_debts
	// .discipline_id after migration 051.
	mock.ExpectQuery(regexp.QuoteMeta("SELECT DISTINCT discipline_id FROM schedule_lessons WHERE teacher_id = $1")).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"discipline_id"}).AddRow(int64(42)).AddRow(int64(43)))

	ids, err := repo.DisciplineIDsForTeacher(context.Background(), 9)
	require.NoError(t, err)
	assert.Equal(t, []int64{42, 43}, ids)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTeacherScopeResolverPG_EmptyWhenTeacherTeachesNothing(t *testing.T) {
	repo, mock := newTeacherScopeMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("FROM schedule_lessons WHERE teacher_id = $1")).
		WithArgs(int64(9)).
		WillReturnRows(sqlmock.NewRows([]string{"discipline_id"}))

	ids, err := repo.DisciplineIDsForTeacher(context.Background(), 9)
	require.NoError(t, err)
	assert.Empty(t, ids, "a teacher scheduled for nothing owns no disciplines (empty, not error)")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTeacherScopeResolverPG_QueryErrorPropagates(t *testing.T) {
	repo, mock := newTeacherScopeMock(t)
	sentinel := errors.New("schedule unavailable")

	mock.ExpectQuery(regexp.QuoteMeta("FROM schedule_lessons WHERE teacher_id = $1")).
		WithArgs(int64(9)).
		WillReturnError(sentinel)

	_, err := repo.DisciplineIDsForTeacher(context.Background(), 9)
	assert.ErrorIs(t, err, sentinel, "a query failure must propagate, not be swallowed as an empty scope")
	assert.NoError(t, mock.ExpectationsWereMet())
}
