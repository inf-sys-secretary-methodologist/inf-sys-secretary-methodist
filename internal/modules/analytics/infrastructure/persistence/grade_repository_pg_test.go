package persistence

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

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/analytics/domain/entities"
)

func newGradeRepoMock(t *testing.T) (*GradeRepositoryPG, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewGradeRepositoryPG(db), mock
}

var gradeCols = []string{"id", "student_id", "subject", "grade_type", "grade_value", "max_value", "weight", "graded_by", "grade_date", "notes", "created_at", "updated_at"}

// ---- CreateGrade ----

func TestGradeCreate_Success(t *testing.T) {
	repo, mock := newGradeRepoMock(t)
	now := time.Now()
	grade := &entities.Grade{StudentID: 1, Subject: "Math", GradeType: entities.GradeTypeCurrent, GradeValue: 85, MaxValue: 100, Weight: 1.0, GradeDate: now}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO grades")).
		WithArgs(grade.StudentID, grade.Subject, grade.GradeType, grade.GradeValue, grade.MaxValue, grade.Weight, grade.GradedBy, grade.GradeDate, grade.Notes).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(int64(1), now, now))

	err := repo.CreateGrade(context.Background(), grade)
	require.NoError(t, err)
	assert.Equal(t, int64(1), grade.ID)
}

func TestGradeCreate_Error(t *testing.T) {
	repo, mock := newGradeRepoMock(t)
	grade := &entities.Grade{StudentID: 1, Subject: "Math"}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO grades")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("insert error"))

	err := repo.CreateGrade(context.Background(), grade)
	assert.Error(t, err)
}

// ---- GetGradesByStudent ----

func TestGradeGetGradesByStudent_Success(t *testing.T) {
	repo, mock := newGradeRepoMock(t)
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta("WHERE student_id = $1")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(gradeCols).
			AddRow(int64(1), int64(1), "Math", "current", 85.0, 100.0, 1.0, nil, now, nil, now, now))

	grades, err := repo.GetGradesByStudent(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, grades, 1)
}

func TestGradeGetGradesByStudent_QueryError(t *testing.T) {
	repo, mock := newGradeRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE student_id = $1")).
		WithArgs(int64(1)).
		WillReturnError(fmt.Errorf("query error"))

	_, err := repo.GetGradesByStudent(context.Background(), 1)
	assert.Error(t, err)
}

func TestGradeGetGradesByStudent_ScanError(t *testing.T) {
	repo, mock := newGradeRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE student_id = $1")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("bad"))

	_, err := repo.GetGradesByStudent(context.Background(), 1)
	assert.Error(t, err)
}

// ---- GetGradesBySubject ----

func TestGradeGetGradesBySubject_Success(t *testing.T) {
	repo, mock := newGradeRepoMock(t)
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta("WHERE student_id = $1 AND subject = $2")).
		WithArgs(int64(1), "Math").
		WillReturnRows(sqlmock.NewRows(gradeCols).
			AddRow(int64(1), int64(1), "Math", "current", 85.0, 100.0, 1.0, nil, now, nil, now, now))

	grades, err := repo.GetGradesBySubject(context.Background(), 1, "Math")
	require.NoError(t, err)
	assert.Len(t, grades, 1)
}

func TestGradeGetGradesBySubject_QueryError(t *testing.T) {
	repo, mock := newGradeRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE student_id = $1 AND subject = $2")).
		WithArgs(int64(1), "Math").
		WillReturnError(fmt.Errorf("query error"))

	_, err := repo.GetGradesBySubject(context.Background(), 1, "Math")
	assert.Error(t, err)
}

func TestGradeGetGradesBySubject_ScanError(t *testing.T) {
	repo, mock := newGradeRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE student_id = $1 AND subject = $2")).
		WithArgs(int64(1), "Math").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("bad"))

	_, err := repo.GetGradesBySubject(context.Background(), 1, "Math")
	assert.Error(t, err)
}

// ---- UpdateGrade ----

func TestGradeUpdate_Success(t *testing.T) {
	repo, mock := newGradeRepoMock(t)
	now := time.Now()
	grade := &entities.Grade{ID: 1, Subject: "Math", GradeType: entities.GradeTypeCurrent, GradeValue: 90, MaxValue: 100, Weight: 1.0, GradeDate: now}

	mock.ExpectQuery(regexp.QuoteMeta("UPDATE grades SET")).
		WithArgs(grade.Subject, grade.GradeType, grade.GradeValue, grade.MaxValue, grade.Weight, grade.GradedBy, grade.GradeDate, grade.Notes, grade.ID).
		WillReturnRows(sqlmock.NewRows([]string{"updated_at"}).AddRow(now))

	err := repo.UpdateGrade(context.Background(), grade)
	require.NoError(t, err)
}

func TestGradeUpdate_Error(t *testing.T) {
	repo, mock := newGradeRepoMock(t)
	grade := &entities.Grade{ID: 1}

	mock.ExpectQuery(regexp.QuoteMeta("UPDATE grades SET")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("update error"))

	err := repo.UpdateGrade(context.Background(), grade)
	assert.Error(t, err)
}

// ---- DeleteGrade ----

func TestGradeDelete_Success(t *testing.T) {
	repo, mock := newGradeRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM grades")).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.DeleteGrade(context.Background(), 1)
	require.NoError(t, err)
}

func TestGradeDelete_NotFound(t *testing.T) {
	repo, mock := newGradeRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM grades")).
		WithArgs(int64(999)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.DeleteGrade(context.Background(), 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "grade not found")
}

func TestGradeDelete_ExecError(t *testing.T) {
	repo, mock := newGradeRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM grades")).
		WithArgs(int64(1)).
		WillReturnError(fmt.Errorf("delete error"))

	err := repo.DeleteGrade(context.Background(), 1)
	assert.Error(t, err)
}

func TestGradeDelete_RowsAffectedError(t *testing.T) {
	repo, mock := newGradeRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM grades")).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("rows error")))

	err := repo.DeleteGrade(context.Background(), 1)
	assert.Error(t, err)
}

// ---- GetStudentGradeStats ----

func TestGradeGetStudentGradeStats_Success(t *testing.T) {
	repo, mock := newGradeRepoMock(t)
	cols := []string{"student_id", "student_name", "group_name", "total_grades", "grade_average", "weighted_average", "min_grade", "max_grade", "failing_grades_count"}
	groupName := "G1"

	mock.ExpectQuery(regexp.QuoteMeta("SELECT student_id, student_name")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(cols).AddRow(int64(1), "John", &groupName, 10, 85.0, 87.0, 60.0, 100.0, 1))

	stats, err := repo.GetStudentGradeStats(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, 85.0, stats.GradeAverage)
}

func TestGradeGetStudentGradeStats_NotFound(t *testing.T) {
	repo, mock := newGradeRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT student_id")).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetStudentGradeStats(context.Background(), 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "student not found")
}

func TestGradeGetStudentGradeStats_DBError(t *testing.T) {
	repo, mock := newGradeRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT student_id")).
		WithArgs(int64(1)).
		WillReturnError(fmt.Errorf("db error"))

	_, err := repo.GetStudentGradeStats(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get grade stats")
}
