package persistence

// v0.153.7 Phase 6 backfill — closes uncovered branches in
// AssignmentRepositoryPG (List scan/rowsErr/dueDate.Valid,
// AggregateGradeDistribution scan/rowsErr) and SubmissionRepositoryPG
// (ListByAssignment + ListByStudent scan/rowsErr).
//
// All tests are sqlmock-driven и mirror the existing per-file conventions
// (regexp.QuoteMeta + WithArgs pinning). No production change.

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/domain/entities"
)

// ===== AssignmentRepositoryPG.List branch coverage =====

func TestAssignmentRepositoryPG_List_ScanError(t *testing.T) {
	repo, mock := newAssignmentRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM assignments")).
		WithArgs(sql.NullInt64{}, "", "").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta("ORDER BY created_at DESC")).
		WithArgs(sql.NullInt64{}, "", "", 50, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	_, err := repo.List(context.Background(), usecases.AssignmentListFilter{Limit: 50})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "list scan")
}

func TestAssignmentRepositoryPG_List_RowsErrPropagates(t *testing.T) {
	repo, mock := newAssignmentRepoMock(t)
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM assignments")).
		WithArgs(sql.NullInt64{}, "", "").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	rows := sqlmock.NewRows([]string{
		"id", "title", "description", "teacher_id", "group_name",
		"subject", "max_score", "due_date", "created_at", "updated_at",
	}).
		AddRow(int64(1), "L1", "d", int64(42), "ИС-21", "Algo", 100, sql.NullTime{}, now, now).
		RowError(0, fmt.Errorf("connection reset during iteration"))
	mock.ExpectQuery(regexp.QuoteMeta("ORDER BY created_at DESC")).
		WithArgs(sql.NullInt64{}, "", "", 50, 0).
		WillReturnRows(rows)

	_, err := repo.List(context.Background(), usecases.AssignmentListFilter{Limit: 50})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "list iter")
}

func TestAssignmentRepositoryPG_List_DueDateValidPopulated(t *testing.T) {
	// Covers `if dueDate.Valid` branch inside the List-loop scan (line
	// 120-123) — populates pointer onto reconstituted assignment.
	repo, mock := newAssignmentRepoMock(t)
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	due := now.Add(7 * 24 * time.Hour)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM assignments")).
		WithArgs(sql.NullInt64{}, "", "").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	rows := sqlmock.NewRows([]string{
		"id", "title", "description", "teacher_id", "group_name",
		"subject", "max_score", "due_date", "created_at", "updated_at",
	}).AddRow(int64(1), "L1", "d", int64(42), "ИС-21", "Algo", 100,
		sql.NullTime{Time: due, Valid: true}, now, now)
	mock.ExpectQuery(regexp.QuoteMeta("ORDER BY created_at DESC")).
		WithArgs(sql.NullInt64{}, "", "", 50, 0).
		WillReturnRows(rows)

	got, err := repo.List(context.Background(), usecases.AssignmentListFilter{Limit: 50})
	require.NoError(t, err)
	require.Len(t, got.Items, 1)
	require.NotNil(t, got.Items[0].DueDate())
	assert.True(t, got.Items[0].DueDate().Equal(due))
}

// ===== AssignmentRepositoryPG.AggregateGradeDistribution =====

func TestAssignmentRepositoryPG_AggregateGradeDistribution_ScanError(t *testing.T) {
	repo, mock := newAssignmentRepoMock(t)
	from := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery(`SELECT a.subject, s.status, COUNT\(\*\) FROM submissions s`).
		WithArgs(from, to).
		WillReturnRows(sqlmock.NewRows([]string{"subject"}).AddRow("Algo"))

	got, err := repo.AggregateGradeDistribution(context.Background(), from, to)
	require.Error(t, err)
	require.Nil(t, got)
	assert.Contains(t, err.Error(), "aggregate grade scan")
}

func TestAssignmentRepositoryPG_AggregateGradeDistribution_RowsErrPropagates(t *testing.T) {
	repo, mock := newAssignmentRepoMock(t)
	from := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	rows := sqlmock.NewRows([]string{"subject", "status", "count"}).
		AddRow("Algo", "graded", 5).
		RowError(0, fmt.Errorf("iter failure"))
	mock.ExpectQuery(`SELECT a.subject, s.status, COUNT\(\*\) FROM submissions s`).
		WithArgs(from, to).
		WillReturnRows(rows)

	got, err := repo.AggregateGradeDistribution(context.Background(), from, to)
	require.Error(t, err)
	require.Nil(t, got)
	assert.Contains(t, err.Error(), "aggregate grade rows")
}

// ===== SubmissionRepositoryPG.ListByAssignment branch coverage =====

func TestSubmissionRepositoryPG_ListByAssignment_ScanError(t *testing.T) {
	repo, mock := newSubmissionRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("FROM submissions s")).
		WithArgs(int64(10), "").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	_, err := repo.ListByAssignment(context.Background(), 10, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "list by assignment scan")
}

func TestSubmissionRepositoryPG_ListByAssignment_RowsErrPropagates(t *testing.T) {
	repo, mock := newSubmissionRepoMock(t)
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{
		"id", "assignment_id", "student_id", "student_name",
		"grade_value", "feedback", "graded_by", "graded_at",
		"status", "return_reason", "returned_by", "returned_at",
		"created_at", "updated_at",
	}).
		AddRow(int64(1), int64(10), int64(7), "Иван",
			sql.NullInt64{}, sql.NullString{}, sql.NullInt64{}, sql.NullTime{},
			"pending", nil, nil, nil,
			now, now).
		RowError(0, fmt.Errorf("connection reset"))

	mock.ExpectQuery(regexp.QuoteMeta("FROM submissions s")).
		WithArgs(int64(10), "").
		WillReturnRows(rows)

	_, err := repo.ListByAssignment(context.Background(), 10, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "list by assignment iter")
}

// ===== SubmissionRepositoryPG.ListByStudent branch coverage =====

func TestSubmissionRepositoryPG_ListByStudent_ScanError(t *testing.T) {
	repo, mock := newSubmissionRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE s.student_id = $1")).
		WithArgs(int64(7), "").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	_, err := repo.ListByStudent(context.Background(), 7, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "list by student scan")
}

func TestSubmissionRepositoryPG_ListByStudent_RowsErrPropagates(t *testing.T) {
	repo, mock := newSubmissionRepoMock(t)
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)

	// Row order mirrors ListByStudent SELECT shape (21 columns).
	rows := sqlmock.NewRows([]string{
		"aid", "title", "description", "subject", "group_name",
		"max_score", "due_date",
		"a_created_at", "a_updated_at",
		"submission_id", "student_id",
		"grade_value", "feedback", "graded_by", "graded_at",
		"return_reason", "returned_by", "returned_at",
		"status", "s_created_at", "s_updated_at",
	}).
		AddRow(int64(10), "L1", "d", "Algo", "ИС-21",
			100, sql.NullTime{},
			now, now,
			int64(1), int64(7),
			sql.NullInt64{}, sql.NullString{}, sql.NullInt64{}, sql.NullTime{},
			sql.NullString{}, sql.NullInt64{}, sql.NullTime{},
			string(entities.StatusPending), now, now).
		RowError(0, fmt.Errorf("connection reset"))

	mock.ExpectQuery(regexp.QuoteMeta("WHERE s.student_id = $1")).
		WithArgs(int64(7), "").
		WillReturnRows(rows)

	_, err := repo.ListByStudent(context.Background(), 7, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "list by student iter")
}
