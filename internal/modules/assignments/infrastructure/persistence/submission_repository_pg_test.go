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

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/repositories"
)

func newSubmissionRepoMock(t *testing.T) (*SubmissionRepositoryPG, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewSubmissionRepositoryPG(db), mock
}

func TestSubmissionRepositoryPG_GetByAssignmentAndStudent(t *testing.T) {
	now := time.Date(2026, 5, 4, 12, 0, 0, 0, time.UTC)

	t.Run("graded row hydrates non-nil grade fields", func(t *testing.T) {
		repo, mock := newSubmissionRepoMock(t)
		rows := sqlmock.NewRows([]string{
			"id", "assignment_id", "student_id", "grade_value", "feedback",
			"graded_by", "graded_at", "status", "created_at", "updated_at",
		}).AddRow(int64(5), int64(10), int64(7), int64(85), "good",
			int64(42), now, "graded", now, now)

		mock.ExpectQuery(regexp.QuoteMeta("FROM submissions WHERE assignment_id = $1 AND student_id = $2")).
			WithArgs(int64(10), int64(7)).
			WillReturnRows(rows)

		got, err := repo.GetByAssignmentAndStudent(context.Background(), 10, 7)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, int64(5), got.ID)
		assert.Equal(t, entities.StatusGraded, got.Status())
		require.NotNil(t, got.GradeValue())
		assert.Equal(t, 85, *got.GradeValue())
		require.NotNil(t, got.GradedBy())
		assert.Equal(t, int64(42), *got.GradedBy())
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("pending row leaves nullables nil", func(t *testing.T) {
		repo, mock := newSubmissionRepoMock(t)
		rows := sqlmock.NewRows([]string{
			"id", "assignment_id", "student_id", "grade_value", "feedback",
			"graded_by", "graded_at", "status", "created_at", "updated_at",
		}).AddRow(int64(5), int64(10), int64(7), nil, "",
			nil, nil, "pending", now, now)

		mock.ExpectQuery(regexp.QuoteMeta("FROM submissions WHERE assignment_id = $1 AND student_id = $2")).
			WithArgs(int64(10), int64(7)).
			WillReturnRows(rows)

		got, err := repo.GetByAssignmentAndStudent(context.Background(), 10, 7)
		require.NoError(t, err)
		assert.Equal(t, entities.StatusPending, got.Status())
		assert.Nil(t, got.GradeValue())
		assert.Nil(t, got.GradedBy())
		assert.Nil(t, got.GradedAt())
	})

	t.Run("no rows surfaces ErrSubmissionNotFound", func(t *testing.T) {
		repo, mock := newSubmissionRepoMock(t)
		mock.ExpectQuery(regexp.QuoteMeta("FROM submissions WHERE assignment_id = $1 AND student_id = $2")).
			WithArgs(int64(10), int64(7)).
			WillReturnError(sql.ErrNoRows)

		got, err := repo.GetByAssignmentAndStudent(context.Background(), 10, 7)
		assert.Nil(t, got)
		require.Error(t, err)
		assert.True(t, errors.Is(err, repositories.ErrSubmissionNotFound))
	})

	t.Run("transport error wraps with op context", func(t *testing.T) {
		repo, mock := newSubmissionRepoMock(t)
		mock.ExpectQuery(regexp.QuoteMeta("FROM submissions WHERE assignment_id = $1 AND student_id = $2")).
			WithArgs(int64(10), int64(7)).
			WillReturnError(errors.New("conn refused"))

		got, err := repo.GetByAssignmentAndStudent(context.Background(), 10, 7)
		assert.Nil(t, got)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "submissions: get by pair")
		assert.False(t, errors.Is(err, repositories.ErrSubmissionNotFound),
			"transport error must NOT be misclassified as not-found")
	})
}

func TestSubmissionRepositoryPG_Save_Upsert(t *testing.T) {
	now := time.Date(2026, 5, 4, 12, 0, 0, 0, time.UTC)
	repo, mock := newSubmissionRepoMock(t)

	score, _ := entities.NewScore(85, 100)
	sub := entities.NewSubmission(10, 7, now)
	require.NoError(t, sub.Grade(score, "good", 42, now))

	mock.ExpectQuery(regexp.QuoteMeta(
		"INSERT INTO submissions",
	)).WithArgs(
		int64(10), int64(7),
		sqlmock.AnyArg(), // grade_value
		sqlmock.AnyArg(), // feedback
		sqlmock.AnyArg(), // graded_by
		sqlmock.AnyArg(), // graded_at
		"graded",
		sqlmock.AnyArg(), // created_at
		sqlmock.AnyArg(), // updated_at
	).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(99)))

	require.NoError(t, repo.Save(context.Background(), sub))
	assert.Equal(t, int64(99), sub.ID, "Save must populate the assigned id back onto the entity")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSubmissionRepositoryPG_Save_OnConflictReturnsExistingID(t *testing.T) {
	now := time.Date(2026, 5, 4, 12, 0, 0, 0, time.UTC)
	repo, mock := newSubmissionRepoMock(t)

	score, _ := entities.NewScore(70, 100)
	sub := entities.NewSubmission(10, 7, now)
	require.NoError(t, sub.Grade(score, "ok", 42, now))

	// On a "first grade" race, a concurrent writer has already inserted
	// the (assignment_id, student_id) row. The unique constraint forces
	// our INSERT into ON CONFLICT DO UPDATE; the RETURNING id surfaces
	// the existing row's id (not a freshly generated one).
	mock.ExpectQuery(regexp.QuoteMeta("ON CONFLICT (assignment_id, student_id) DO UPDATE SET")).
		WithArgs(
			int64(10), int64(7),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), "graded",
			sqlmock.AnyArg(), sqlmock.AnyArg(),
		).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(123)))

	require.NoError(t, repo.Save(context.Background(), sub))
	assert.Equal(t, int64(123), sub.ID,
		"Save must surface the upserted row id from RETURNING (existing row on conflict)")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSubmissionRepositoryPG_Save_TransportErrorWrapped(t *testing.T) {
	now := time.Date(2026, 5, 4, 12, 0, 0, 0, time.UTC)
	repo, mock := newSubmissionRepoMock(t)

	score, _ := entities.NewScore(50, 100)
	sub := entities.NewSubmission(10, 7, now)
	require.NoError(t, sub.Grade(score, "", 42, now))

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO submissions")).
		WillReturnError(errors.New("deadlock detected"))

	err := repo.Save(context.Background(), sub)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "submissions: upsert")
	assert.Equal(t, int64(0), sub.ID, "ID must not be set when upsert fails")
}

func TestSubmissionRepositoryPG_ListByAssignment(t *testing.T) {
	now := time.Date(2026, 5, 4, 12, 0, 0, 0, time.UTC)

	t.Run("rows hydrate read-model with student name from JOIN", func(t *testing.T) {
		repo, mock := newSubmissionRepoMock(t)

		rows := sqlmock.NewRows([]string{
			"id", "assignment_id", "student_id", "student_name",
			"grade_value", "feedback", "graded_by", "graded_at",
			"status", "created_at", "updated_at",
		}).
			AddRow(int64(1), int64(10), int64(7), "Иван Петров",
				sql.NullInt64{}, sql.NullString{}, sql.NullInt64{}, sql.NullTime{},
				"pending", now, now).
			AddRow(int64(2), int64(10), int64(8), "Анна Смирнова",
				int64(85), "great", int64(42), now,
				"graded", now, now)

		mock.ExpectQuery(regexp.QuoteMeta("FROM submissions s")).
			WithArgs(int64(10), "").
			WillReturnRows(rows)

		got, err := repo.ListByAssignment(context.Background(), 10, nil)
		require.NoError(t, err)
		require.Len(t, got, 2)

		assert.Equal(t, "Иван Петров", got[0].StudentName)
		assert.Equal(t, entities.StatusPending, got[0].Status)
		assert.Nil(t, got[0].GradeValue)

		assert.Equal(t, "Анна Смирнова", got[1].StudentName)
		assert.Equal(t, entities.StatusGraded, got[1].Status)
		require.NotNil(t, got[1].GradeValue)
		assert.Equal(t, 85, *got[1].GradeValue)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("status filter is forwarded as text argument", func(t *testing.T) {
		repo, mock := newSubmissionRepoMock(t)

		mock.ExpectQuery(regexp.QuoteMeta("FROM submissions s")).
			WithArgs(int64(10), "graded").
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "assignment_id", "student_id", "student_name",
				"grade_value", "feedback", "graded_by", "graded_at",
				"status", "created_at", "updated_at",
			}))

		status := entities.StatusGraded
		got, err := repo.ListByAssignment(context.Background(), 10, &status)
		require.NoError(t, err)
		assert.Empty(t, got)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("transport error wraps", func(t *testing.T) {
		repo, mock := newSubmissionRepoMock(t)

		mock.ExpectQuery(regexp.QuoteMeta("FROM submissions s")).
			WithArgs(int64(10), "").
			WillReturnError(errors.New("conn refused"))

		_, err := repo.ListByAssignment(context.Background(), 10, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "list by assignment")
	})
}
