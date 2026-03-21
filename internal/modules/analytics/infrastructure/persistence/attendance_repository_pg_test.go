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

func newAttendanceRepoMock(t *testing.T) (*AttendanceRepositoryPG, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewAttendanceRepositoryPG(db), mock
}

var lessonCols = []string{"id", "name", "subject", "teacher_id", "group_name", "lesson_type", "created_at", "updated_at"}
var attendanceCols = []string{"id", "student_id", "lesson_id", "lesson_date", "status", "marked_by", "notes", "created_at", "updated_at"}

// ---- CreateLesson ----

func TestAttendanceCreateLesson_Success(t *testing.T) {
	repo, mock := newAttendanceRepoMock(t)
	now := time.Now()
	teacherID := int64(1)
	groupName := "G1"
	lesson := &entities.Lesson{Name: "Math", Subject: "Mathematics", TeacherID: &teacherID, GroupName: &groupName, LessonType: entities.LessonTypeLecture}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO lessons")).
		WithArgs(lesson.Name, lesson.Subject, lesson.TeacherID, lesson.GroupName, lesson.LessonType).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(int64(1), now, now))

	err := repo.CreateLesson(context.Background(), lesson)
	require.NoError(t, err)
	assert.Equal(t, int64(1), lesson.ID)
}

func TestAttendanceCreateLesson_Error(t *testing.T) {
	repo, mock := newAttendanceRepoMock(t)
	lesson := &entities.Lesson{Name: "Math"}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO lessons")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("insert error"))

	err := repo.CreateLesson(context.Background(), lesson)
	assert.Error(t, err)
}

// ---- GetLessonByID ----

func TestAttendanceGetLessonByID_Success(t *testing.T) {
	repo, mock := newAttendanceRepoMock(t)
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, subject")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(lessonCols).AddRow(int64(1), "Math", "Mathematics", nil, nil, "lecture", now, now))

	lesson, err := repo.GetLessonByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "Math", lesson.Name)
}

func TestAttendanceGetLessonByID_NotFound(t *testing.T) {
	repo, mock := newAttendanceRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name")).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetLessonByID(context.Background(), 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "lesson not found")
}

func TestAttendanceGetLessonByID_DBError(t *testing.T) {
	repo, mock := newAttendanceRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name")).
		WithArgs(int64(1)).
		WillReturnError(fmt.Errorf("db error"))

	_, err := repo.GetLessonByID(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get lesson")
}

// ---- GetLessonsByGroup ----

func TestAttendanceGetLessonsByGroup_Success(t *testing.T) {
	repo, mock := newAttendanceRepoMock(t)
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta("WHERE group_name = $1")).
		WithArgs("G1").
		WillReturnRows(sqlmock.NewRows(lessonCols).AddRow(int64(1), "Math", "Mathematics", nil, "G1", "lecture", now, now))

	lessons, err := repo.GetLessonsByGroup(context.Background(), "G1")
	require.NoError(t, err)
	assert.Len(t, lessons, 1)
}

func TestAttendanceGetLessonsByGroup_QueryError(t *testing.T) {
	repo, mock := newAttendanceRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE group_name = $1")).
		WithArgs("G1").
		WillReturnError(fmt.Errorf("query error"))

	_, err := repo.GetLessonsByGroup(context.Background(), "G1")
	assert.Error(t, err)
}

func TestAttendanceGetLessonsByGroup_ScanError(t *testing.T) {
	repo, mock := newAttendanceRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE group_name = $1")).
		WithArgs("G1").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("bad"))

	_, err := repo.GetLessonsByGroup(context.Background(), "G1")
	assert.Error(t, err)
}

// ---- GetLessonsByTeacher ----

func TestAttendanceGetLessonsByTeacher_Success(t *testing.T) {
	repo, mock := newAttendanceRepoMock(t)
	now := time.Now()
	teacherID := int64(1)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE teacher_id = $1")).
		WithArgs(teacherID).
		WillReturnRows(sqlmock.NewRows(lessonCols).AddRow(int64(1), "Math", "Mathematics", teacherID, nil, "lecture", now, now))

	lessons, err := repo.GetLessonsByTeacher(context.Background(), teacherID)
	require.NoError(t, err)
	assert.Len(t, lessons, 1)
}

func TestAttendanceGetLessonsByTeacher_QueryError(t *testing.T) {
	repo, mock := newAttendanceRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE teacher_id = $1")).
		WithArgs(int64(1)).
		WillReturnError(fmt.Errorf("query error"))

	_, err := repo.GetLessonsByTeacher(context.Background(), 1)
	assert.Error(t, err)
}

func TestAttendanceGetLessonsByTeacher_ScanError(t *testing.T) {
	repo, mock := newAttendanceRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE teacher_id = $1")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("bad"))

	_, err := repo.GetLessonsByTeacher(context.Background(), 1)
	assert.Error(t, err)
}

// ---- MarkAttendance ----

func TestAttendanceMarkAttendance_Success(t *testing.T) {
	repo, mock := newAttendanceRepoMock(t)
	now := time.Now()
	record := &entities.AttendanceRecord{StudentID: 1, LessonID: 1, LessonDate: now, Status: entities.AttendanceStatusPresent}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO attendance_records")).
		WithArgs(record.StudentID, record.LessonID, record.LessonDate, record.Status, record.MarkedBy, record.Notes).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(int64(1), now, now))

	err := repo.MarkAttendance(context.Background(), record)
	require.NoError(t, err)
	assert.Equal(t, int64(1), record.ID)
}

func TestAttendanceMarkAttendance_Error(t *testing.T) {
	repo, mock := newAttendanceRepoMock(t)
	record := &entities.AttendanceRecord{StudentID: 1, LessonID: 1}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO attendance_records")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("insert error"))

	err := repo.MarkAttendance(context.Background(), record)
	assert.Error(t, err)
}

// ---- BulkMarkAttendance ----

func TestAttendanceBulkMarkAttendance_Empty(t *testing.T) {
	repo, _ := newAttendanceRepoMock(t)
	err := repo.BulkMarkAttendance(context.Background(), []entities.AttendanceRecord{})
	require.NoError(t, err)
}

func TestAttendanceBulkMarkAttendance_Success(t *testing.T) {
	repo, mock := newAttendanceRepoMock(t)
	now := time.Now()
	records := []entities.AttendanceRecord{
		{StudentID: 1, LessonID: 1, LessonDate: now, Status: entities.AttendanceStatusPresent},
		{StudentID: 2, LessonID: 1, LessonDate: now, Status: entities.AttendanceStatusAbsent},
	}

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO attendance_records")).
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(0, 2))

	err := repo.BulkMarkAttendance(context.Background(), records)
	require.NoError(t, err)
}

func TestAttendanceBulkMarkAttendance_Error(t *testing.T) {
	repo, mock := newAttendanceRepoMock(t)
	now := time.Now()
	records := []entities.AttendanceRecord{{StudentID: 1, LessonID: 1, LessonDate: now, Status: entities.AttendanceStatusPresent}}

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO attendance_records")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("exec error"))

	err := repo.BulkMarkAttendance(context.Background(), records)
	assert.Error(t, err)
}

// ---- GetAttendanceByLesson ----

func TestAttendanceGetAttendanceByLesson_Success(t *testing.T) {
	repo, mock := newAttendanceRepoMock(t)
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta("WHERE lesson_id = $1 AND lesson_date = $2")).
		WithArgs(int64(1), "2024-01-15").
		WillReturnRows(sqlmock.NewRows(attendanceCols).
			AddRow(int64(1), int64(10), int64(1), now, "present", nil, nil, now, now))

	records, err := repo.GetAttendanceByLesson(context.Background(), 1, "2024-01-15")
	require.NoError(t, err)
	assert.Len(t, records, 1)
}

func TestAttendanceGetAttendanceByLesson_QueryError(t *testing.T) {
	repo, mock := newAttendanceRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE lesson_id = $1")).
		WithArgs(int64(1), "2024-01-15").
		WillReturnError(fmt.Errorf("query error"))

	_, err := repo.GetAttendanceByLesson(context.Background(), 1, "2024-01-15")
	assert.Error(t, err)
}

func TestAttendanceGetAttendanceByLesson_ScanError(t *testing.T) {
	repo, mock := newAttendanceRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE lesson_id = $1")).
		WithArgs(int64(1), "2024-01-15").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("bad"))

	_, err := repo.GetAttendanceByLesson(context.Background(), 1, "2024-01-15")
	assert.Error(t, err)
}

// ---- GetAttendanceByStudent ----

func TestAttendanceGetAttendanceByStudent_Success(t *testing.T) {
	repo, mock := newAttendanceRepoMock(t)
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta("WHERE student_id = $1")).
		WithArgs(int64(1), "2024-01-01", "2024-01-31").
		WillReturnRows(sqlmock.NewRows(attendanceCols).
			AddRow(int64(1), int64(1), int64(10), now, "present", nil, nil, now, now))

	records, err := repo.GetAttendanceByStudent(context.Background(), 1, "2024-01-01", "2024-01-31")
	require.NoError(t, err)
	assert.Len(t, records, 1)
}

func TestAttendanceGetAttendanceByStudent_QueryError(t *testing.T) {
	repo, mock := newAttendanceRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE student_id = $1")).
		WithArgs(int64(1), "2024-01-01", "2024-01-31").
		WillReturnError(fmt.Errorf("query error"))

	_, err := repo.GetAttendanceByStudent(context.Background(), 1, "2024-01-01", "2024-01-31")
	assert.Error(t, err)
}

func TestAttendanceGetAttendanceByStudent_ScanError(t *testing.T) {
	repo, mock := newAttendanceRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE student_id = $1")).
		WithArgs(int64(1), "2024-01-01", "2024-01-31").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("bad"))

	_, err := repo.GetAttendanceByStudent(context.Background(), 1, "2024-01-01", "2024-01-31")
	assert.Error(t, err)
}

// ---- UpdateAttendance ----

func TestAttendanceUpdateAttendance_Success(t *testing.T) {
	repo, mock := newAttendanceRepoMock(t)
	now := time.Now()
	record := &entities.AttendanceRecord{ID: 1, Status: entities.AttendanceStatusLate}

	mock.ExpectQuery(regexp.QuoteMeta("UPDATE attendance_records SET")).
		WithArgs(record.Status, record.MarkedBy, record.Notes, record.ID).
		WillReturnRows(sqlmock.NewRows([]string{"updated_at"}).AddRow(now))

	err := repo.UpdateAttendance(context.Background(), record)
	require.NoError(t, err)
}

func TestAttendanceUpdateAttendance_Error(t *testing.T) {
	repo, mock := newAttendanceRepoMock(t)
	record := &entities.AttendanceRecord{ID: 1}

	mock.ExpectQuery(regexp.QuoteMeta("UPDATE attendance_records SET")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("update error"))

	err := repo.UpdateAttendance(context.Background(), record)
	assert.Error(t, err)
}

// ---- GetStudentAttendanceStats ----

func TestAttendanceGetStudentAttendanceStats_Success(t *testing.T) {
	repo, mock := newAttendanceRepoMock(t)
	cols := []string{"student_id", "student_name", "group_name", "total_records", "present_count", "absent_count", "late_count", "excused_count", "attendance_rate"}
	groupName := "G1"

	mock.ExpectQuery(regexp.QuoteMeta("SELECT student_id, student_name")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(cols).AddRow(int64(1), "John", &groupName, 20, 18, 1, 1, 0, 0.9))

	stats, err := repo.GetStudentAttendanceStats(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, 0.9, stats.AttendanceRate)
}

func TestAttendanceGetStudentAttendanceStats_NotFound(t *testing.T) {
	repo, mock := newAttendanceRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT student_id")).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetStudentAttendanceStats(context.Background(), 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "student not found")
}

func TestAttendanceGetStudentAttendanceStats_DBError(t *testing.T) {
	repo, mock := newAttendanceRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT student_id")).
		WithArgs(int64(1)).
		WillReturnError(fmt.Errorf("db error"))

	_, err := repo.GetStudentAttendanceStats(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get attendance stats")
}
