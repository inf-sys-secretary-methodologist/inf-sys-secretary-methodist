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
			"graded_by", "graded_at", "status",
			"return_reason", "returned_by", "returned_at",
			"created_at", "updated_at",
		}).AddRow(int64(5), int64(10), int64(7), int64(85), "good",
			int64(42), now, "graded",
			nil, nil, nil,
			now, now)

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
			"graded_by", "graded_at", "status",
			"return_reason", "returned_by", "returned_at",
			"created_at", "updated_at",
		}).AddRow(int64(5), int64(10), int64(7), nil, "",
			nil, nil, "pending",
			nil, nil, nil,
			now, now)

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

	t.Run("returned row hydrates return triple", func(t *testing.T) {
		repo, mock := newSubmissionRepoMock(t)
		rows := sqlmock.NewRows([]string{
			"id", "assignment_id", "student_id", "grade_value", "feedback",
			"graded_by", "graded_at", "status",
			"return_reason", "returned_by", "returned_at",
			"created_at", "updated_at",
		}).AddRow(
			int64(5), int64(10), int64(7),
			nil, "",
			nil, nil,
			"returned",
			"revisit derivation", int64(99), now,
			now, now,
		)

		mock.ExpectQuery(regexp.QuoteMeta("FROM submissions WHERE assignment_id = $1 AND student_id = $2")).
			WithArgs(int64(10), int64(7)).
			WillReturnRows(rows)

		got, err := repo.GetByAssignmentAndStudent(context.Background(), 10, 7)
		require.NoError(t, err)
		assert.Equal(t, entities.StatusReturned, got.Status())
		assert.Equal(t, "revisit derivation", got.ReturnReason())
		require.NotNil(t, got.ReturnedBy())
		assert.Equal(t, int64(99), *got.ReturnedBy())
		require.NotNil(t, got.ReturnedAt())
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

	score, _ := entities.NewScore(85)
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
		sqlmock.AnyArg(), // return_reason
		sqlmock.AnyArg(), // returned_by
		sqlmock.AnyArg(), // returned_at
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

	score, _ := entities.NewScore(70)
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
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
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

	score, _ := entities.NewScore(50)
	sub := entities.NewSubmission(10, 7, now)
	require.NoError(t, sub.Grade(score, "", 42, now))

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO submissions")).
		WillReturnError(errors.New("deadlock detected"))

	err := repo.Save(context.Background(), sub)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "submissions: upsert")
	assert.Equal(t, int64(0), sub.ID, "ID must not be set when upsert fails")
}

func TestSubmissionRepositoryPG_Save_PersistsReturnFields(t *testing.T) {
	now := time.Date(2026, 5, 4, 12, 0, 0, 0, time.UTC)
	repo, mock := newSubmissionRepoMock(t)

	a, err := entities.NewAssignment(entities.NewAssignmentParams{
		Title: "Lab", GroupName: "A", Subject: "CS", MaxScore: 100, TeacherID: 99, Now: now,
	})
	require.NoError(t, err)
	score, err := a.NewSubmissionScore(85)
	require.NoError(t, err)

	sub := entities.NewSubmission(10, 7, now)
	require.NoError(t, sub.Grade(score, "good", 99, now))
	// Now return: this clears grade fields and sets the return triple.
	require.NoError(t, sub.Return("revisit derivation", 99, now))

	// The Save SQL must now write 12 column values. Order MUST be:
	// assignment_id, student_id, grade_value, feedback, graded_by, graded_at,
	// status, return_reason, returned_by, returned_at, created_at, updated_at.
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO submissions")).
		WithArgs(
			int64(10), int64(7),
			sql.NullInt64{},                       // grade_value cleared
			"",                                    // feedback cleared
			sql.NullInt64{},                       // graded_by cleared
			sql.NullTime{},                        // graded_at cleared
			"returned",
			sql.NullString{String: "revisit derivation", Valid: true}, // return_reason
			sql.NullInt64{Int64: 99, Valid: true},                     // returned_by
			sql.NullTime{Time: now, Valid: true},                      // returned_at
			sqlmock.AnyArg(),                                          // created_at
			sqlmock.AnyArg(),                                          // updated_at
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(42)))

	require.NoError(t, repo.Save(context.Background(), sub))
	assert.Equal(t, int64(42), sub.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSubmissionRepositoryPG_Save_PersistsResubmittedFields(t *testing.T) {
	now := time.Date(2026, 5, 4, 12, 0, 0, 0, time.UTC)
	later := time.Date(2026, 5, 4, 18, 0, 0, 0, time.UTC)
	repo, mock := newSubmissionRepoMock(t)

	a, err := entities.NewAssignment(entities.NewAssignmentParams{
		Title: "Lab", GroupName: "A", Subject: "CS", MaxScore: 100, TeacherID: 99, Now: now,
	})
	require.NoError(t, err)
	score, err := a.NewSubmissionScore(85)
	require.NoError(t, err)

	sub := entities.NewSubmission(10, 7, now)
	require.NoError(t, sub.Grade(score, "good", 99, now))
	require.NoError(t, sub.Return("revisit derivation", 99, now))
	require.NoError(t, sub.Resubmit(later))
	require.Equal(t, entities.StatusPending, sub.Status())

	// After Resubmit the row must persist with status='pending' and both
	// triples NULL — the grade triple is already null (Return cleared it),
	// and Resubmit nulls the return triple. Save writes all 12 column
	// values unconditionally, so no SQL change is required; this backfill
	// pins the round-trip contract for the new entity transition.
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO submissions")).
		WithArgs(
			int64(10), int64(7),
			sql.NullInt64{},  // grade_value cleared (was set in Grade, nulled by Return)
			"",               // feedback cleared
			sql.NullInt64{},  // graded_by cleared
			sql.NullTime{},   // graded_at cleared
			"pending",        // status flipped back from returned
			sql.NullString{}, // return_reason cleared by Resubmit
			sql.NullInt64{},  // returned_by cleared by Resubmit
			sql.NullTime{},   // returned_at cleared by Resubmit
			sqlmock.AnyArg(), // created_at
			sqlmock.AnyArg(), // updated_at (later, post-Resubmit)
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(42)))

	require.NoError(t, repo.Save(context.Background(), sub))
	assert.Equal(t, int64(42), sub.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSubmissionRepositoryPG_ListByAssignment(t *testing.T) {
	now := time.Date(2026, 5, 4, 12, 0, 0, 0, time.UTC)

	t.Run("rows hydrate read-model with student name from JOIN", func(t *testing.T) {
		repo, mock := newSubmissionRepoMock(t)

		rows := sqlmock.NewRows([]string{
			"id", "assignment_id", "student_id", "student_name",
			"grade_value", "feedback", "graded_by", "graded_at",
			"status", "return_reason", "returned_by", "returned_at",
			"created_at", "updated_at",
		}).
			AddRow(int64(1), int64(10), int64(7), "Иван Петров",
				sql.NullInt64{}, sql.NullString{}, sql.NullInt64{}, sql.NullTime{},
				"pending", nil, nil, nil,
				now, now).
			AddRow(int64(2), int64(10), int64(8), "Анна Смирнова",
				int64(85), "great", int64(42), now,
				"graded", nil, nil, nil,
				now, now)

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
				"status", "return_reason", "returned_by", "returned_at",
				"created_at", "updated_at",
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

	t.Run("returned row hydrates return triple", func(t *testing.T) {
		repo, mock := newSubmissionRepoMock(t)

		rows := sqlmock.NewRows([]string{
			"id", "assignment_id", "student_id", "student_name",
			"grade_value", "feedback", "graded_by", "graded_at",
			"status", "return_reason", "returned_by", "returned_at",
			"created_at", "updated_at",
		}).AddRow(
			int64(42), int64(10), int64(7), "Иван Петров",
			nil, "",
			nil, nil,
			"returned",
			"revisit derivation", int64(99), now,
			now, now,
		)

		mock.ExpectQuery(regexp.QuoteMeta("FROM submissions s")).
			WithArgs(int64(10), "").
			WillReturnRows(rows)

		got, err := repo.ListByAssignment(context.Background(), 10, nil)
		require.NoError(t, err)
		require.Len(t, got, 1)

		v := got[0]
		assert.Equal(t, "Иван Петров", v.StudentName)
		assert.Equal(t, entities.StatusReturned, v.Status)
		assert.Nil(t, v.GradeValue)
		assert.Equal(t, "revisit derivation", v.ReturnReason)
		require.NotNil(t, v.ReturnedBy)
		assert.Equal(t, int64(99), *v.ReturnedBy)
		require.NotNil(t, v.ReturnedAt)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSubmissionRepositoryPG_ListByStudent(t *testing.T) {
	now := time.Date(2026, 5, 6, 12, 0, 0, 0, time.UTC)
	due := time.Date(2026, 5, 20, 23, 59, 0, 0, time.UTC)

	t.Run("rows hydrate denormalised assignment + submission projection", func(t *testing.T) {
		repo, mock := newSubmissionRepoMock(t)

		rows := sqlmock.NewRows([]string{
			"assignment_id", "title", "description", "subject", "group_name",
			"max_score", "due_date",
			"a_created_at", "a_updated_at",
			"submission_id", "student_id",
			"grade_value", "feedback", "graded_by", "graded_at",
			"return_reason", "returned_by", "returned_at",
			"status", "s_created_at", "s_updated_at",
		}).
			AddRow(int64(10), "Lab 1", "Solve A", "Math", "БСБО-01-22",
				100, due,
				now, now,
				int64(1), int64(7),
				sql.NullInt64{}, "", sql.NullInt64{}, sql.NullTime{},
				nil, nil, nil,
				"pending", now, now).
			AddRow(int64(11), "Lab 2", "Solve B", "Math", "БСБО-01-22",
				50, sql.NullTime{},
				now, now,
				int64(2), int64(7),
				int64(45), "good", int64(42), now,
				nil, nil, nil,
				"graded", now, now).
			AddRow(int64(12), "Essay", "Topic", "Russian", "БСБО-01-22",
				10, sql.NullTime{},
				now, now,
				int64(3), int64(7),
				nil, "",
				nil, nil,
				"please add citations", int64(99), now,
				"returned", now, now)

		mock.ExpectQuery(regexp.QuoteMeta("FROM submissions s")).
			WithArgs(int64(7), "").
			WillReturnRows(rows)

		got, err := repo.ListByStudent(context.Background(), 7, nil)
		require.NoError(t, err)
		require.Len(t, got, 3)

		// Pending row.
		assert.Equal(t, int64(10), got[0].AssignmentID)
		assert.Equal(t, "Lab 1", got[0].Title)
		assert.Equal(t, "Math", got[0].Subject)
		assert.Equal(t, 100, got[0].MaxScore)
		require.NotNil(t, got[0].DueDate)
		assert.Equal(t, entities.StatusPending, got[0].Status)
		assert.Nil(t, got[0].GradeValue)

		// Graded row.
		assert.Equal(t, entities.StatusGraded, got[1].Status)
		require.NotNil(t, got[1].GradeValue)
		assert.Equal(t, 45, *got[1].GradeValue)
		assert.Equal(t, "good", got[1].Feedback)
		assert.Nil(t, got[1].DueDate)

		// Returned row.
		assert.Equal(t, entities.StatusReturned, got[2].Status)
		assert.Equal(t, "please add citations", got[2].ReturnReason)
		require.NotNil(t, got[2].ReturnedBy)
		assert.Equal(t, int64(99), *got[2].ReturnedBy)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("status filter is forwarded as text argument", func(t *testing.T) {
		repo, mock := newSubmissionRepoMock(t)

		mock.ExpectQuery(regexp.QuoteMeta("FROM submissions s")).
			WithArgs(int64(7), "returned").
			WillReturnRows(sqlmock.NewRows([]string{
				"assignment_id", "title", "description", "subject", "group_name",
				"max_score", "due_date",
				"a_created_at", "a_updated_at",
				"submission_id", "student_id",
				"grade_value", "feedback", "graded_by", "graded_at",
				"return_reason", "returned_by", "returned_at",
				"status", "s_created_at", "s_updated_at",
			}))

		status := entities.StatusReturned
		got, err := repo.ListByStudent(context.Background(), 7, &status)
		require.NoError(t, err)
		assert.Empty(t, got)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("transport error wraps", func(t *testing.T) {
		repo, mock := newSubmissionRepoMock(t)

		mock.ExpectQuery(regexp.QuoteMeta("FROM submissions s")).
			WithArgs(int64(7), "").
			WillReturnError(errors.New("conn refused"))

		_, err := repo.ListByStudent(context.Background(), 7, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "list by student")
	})
}
