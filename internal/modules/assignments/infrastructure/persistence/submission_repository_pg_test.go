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
