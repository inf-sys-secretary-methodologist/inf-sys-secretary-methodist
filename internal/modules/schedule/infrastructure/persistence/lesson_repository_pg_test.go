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

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/repositories"
)

// v0.153.1 #196 coverage push: lesson_repository_pg.go was at 0% covered
// (312 LOC file — biggest single coverage gap remaining). Tests pin
// Create + Save + GetByID (joins) + List + Count + GetTimetable +
// buildWhereClause variants + scanLessons error branches.

func newLessonRepoMock(t *testing.T) (*LessonRepositoryPG, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewLessonRepositoryPG(db), mock
}

func sampleLesson() *entities.Lesson {
	now := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	return entities.NewLesson(1, 2, 3, 4, 5, 6, domain.Monday, "09:00", "10:30",
		domain.WeekTypeAll, now, now.AddDate(0, 4, 0), now)
}

// flatLessonCols mirrors the 18-column SELECT used by List/Count/Save.
var flatLessonCols = []string{
	"id", "semester_id", "discipline_id", "lesson_type_id", "teacher_id",
	"group_id", "classroom_id", "day_of_week", "time_start", "time_end",
	"week_type", "date_start", "date_end", "notes", "is_canceled",
	"cancellation_reason", "created_at", "updated_at",
}

// joinedLessonCols mirrors the GetByID/GetTimetable SELECT with all 5 joins.
var joinedLessonCols = append(append([]string{}, flatLessonCols...),
	// discipline (9)
	"d_id", "d_name", "d_code", "d_dept", "d_credits", "d_total",
	"d_lect", "d_pract", "d_lab",
	// lesson_type (4)
	"lt_id", "lt_name", "lt_short", "lt_color",
	// classroom (9)
	"cr_id", "cr_building", "cr_number", "cr_name", "cr_capacity", "cr_type",
	"cr_avail", "cr_created", "cr_updated",
	// group (6)
	"sg_id", "sg_specialty", "sg_name", "sg_course", "sg_curator", "sg_capacity",
	// teacher (3)
	"u_id", "u_name", "u_email",
)

func addFlatLessonRow(rows *sqlmock.Rows, id int64) *sqlmock.Rows {
	now := time.Now()
	return rows.AddRow(
		id, int64(1), int64(2), int64(3), int64(4),
		int64(5), int64(6), int(1), "09:00", "10:30",
		"all", now, now.AddDate(0, 4, 0), nil, false,
		nil, now, now,
	)
}

func addJoinedLessonRow(rows *sqlmock.Rows, id int64) *sqlmock.Rows {
	now := time.Now()
	return rows.AddRow(
		// flat
		id, int64(1), int64(2), int64(3), int64(4),
		int64(5), int64(6), int(1), "09:00", "10:30",
		"all", now, now.AddDate(0, 4, 0), nil, false,
		nil, now, now,
		// discipline
		int64(2), "Высшая математика", nil, nil, nil, nil, nil, nil, nil,
		// lesson_type
		int64(3), "Лекция", "ЛК", nil,
		// classroom
		int64(6), "Главный", "101", nil, 50, nil, true, now, now,
		// student_group
		int64(5), int64(10), "ИС-21", 2, nil, 25,
		// teacher
		int64(4), "Иванов Иван Иванович", "ivanov@example.com",
	)
}

// --- Create ---

func TestLessonRepoCreate_HappyPath(t *testing.T) {
	repo, mock := newLessonRepoMock(t)
	lesson := sampleLesson()

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO schedule_lessons")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(99)))

	err := repo.Create(context.Background(), lesson)
	require.NoError(t, err)
	assert.Equal(t, int64(99), lesson.ID)
}

func TestLessonRepoCreate_DBError(t *testing.T) {
	repo, mock := newLessonRepoMock(t)
	lesson := sampleLesson()

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO schedule_lessons")).
		WillReturnError(errors.New("constraint violation"))

	err := repo.Create(context.Background(), lesson)
	require.Error(t, err)
}

// --- Save ---

func TestLessonRepoSave_HappyPath(t *testing.T) {
	repo, mock := newLessonRepoMock(t)
	lesson := sampleLesson()
	lesson.ID = 42

	mock.ExpectExec(regexp.QuoteMeta("UPDATE schedule_lessons SET")).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Save(context.Background(), lesson)
	require.NoError(t, err)
}

func TestLessonRepoSave_DBError(t *testing.T) {
	repo, mock := newLessonRepoMock(t)
	lesson := sampleLesson()
	lesson.ID = 42

	mock.ExpectExec(regexp.QuoteMeta("UPDATE schedule_lessons SET")).
		WillReturnError(errors.New("update failed"))

	err := repo.Save(context.Background(), lesson)
	require.Error(t, err)
}

// --- GetByID ---

func TestLessonRepoGetByID_HappyPath(t *testing.T) {
	repo, mock := newLessonRepoMock(t)
	rows := sqlmock.NewRows(joinedLessonCols)
	addJoinedLessonRow(rows, int64(42))

	mock.ExpectQuery(regexp.QuoteMeta("FROM schedule_lessons l")).
		WithArgs(int64(42)).
		WillReturnRows(rows)

	lesson, err := repo.GetByID(context.Background(), int64(42))
	require.NoError(t, err)
	require.NotNil(t, lesson)
	assert.Equal(t, int64(42), lesson.ID)
	require.NotNil(t, lesson.Discipline)
	assert.Equal(t, "Высшая математика", lesson.Discipline.Name)
	require.NotNil(t, lesson.LessonType)
	assert.Equal(t, "Лекция", lesson.LessonType.Name)
	require.NotNil(t, lesson.Classroom)
	assert.Equal(t, "101", lesson.Classroom.Number)
	require.NotNil(t, lesson.Group)
	assert.Equal(t, "ИС-21", lesson.Group.Name)
	require.NotNil(t, lesson.Teacher)
	assert.Equal(t, "Иванов Иван Иванович", lesson.Teacher.Name)
}

func TestLessonRepoGetByID_NotFound(t *testing.T) {
	repo, mock := newLessonRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM schedule_lessons l")).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	lesson, err := repo.GetByID(context.Background(), int64(999))
	require.NoError(t, err, "ErrNoRows must collapse to (nil, nil)")
	assert.Nil(t, lesson)
}

func TestLessonRepoGetByID_QueryError(t *testing.T) {
	repo, mock := newLessonRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM schedule_lessons l")).
		WithArgs(int64(42)).
		WillReturnError(errors.New("connection lost"))

	lesson, err := repo.GetByID(context.Background(), int64(42))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get lesson")
	assert.Nil(t, lesson)
}

// --- Delete ---

func TestLessonRepoDelete_HappyPath(t *testing.T) {
	repo, mock := newLessonRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM schedule_lessons WHERE id = $1")).
		WithArgs(int64(42)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), int64(42))
	require.NoError(t, err)
}

func TestLessonRepoDelete_DBError(t *testing.T) {
	repo, mock := newLessonRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM schedule_lessons WHERE id = $1")).
		WithArgs(int64(42)).
		WillReturnError(errors.New("fk constraint"))

	err := repo.Delete(context.Background(), int64(42))
	require.Error(t, err)
}

// --- List ---

func TestLessonRepoList_NoFilterNoLimit(t *testing.T) {
	repo, mock := newLessonRepoMock(t)
	rows := sqlmock.NewRows(flatLessonCols)
	addFlatLessonRow(rows, int64(1))
	addFlatLessonRow(rows, int64(2))

	mock.ExpectQuery(regexp.QuoteMeta("FROM schedule_lessons ORDER BY day_of_week")).
		WillReturnRows(rows)

	lessons, err := repo.List(context.Background(), repositories.LessonFilter{}, 0, 0)
	require.NoError(t, err)
	assert.Len(t, lessons, 2)
}

func TestLessonRepoList_AllFiltersAdded(t *testing.T) {
	repo, mock := newLessonRepoMock(t)
	semID, gID, tID, cID, dID := int64(1), int64(5), int64(4), int64(6), int64(2)
	day := domain.Monday
	week := domain.WeekTypeAll
	filter := repositories.LessonFilter{
		SemesterID:   &semID,
		GroupID:      &gID,
		TeacherID:    &tID,
		ClassroomID:  &cID,
		DisciplineID: &dID,
		DayOfWeek:    &day,
		WeekType:     &week,
	}

	mock.ExpectQuery(regexp.QuoteMeta("WHERE semester_id = $1 AND group_id = $2 AND teacher_id = $3 AND classroom_id = $4 AND discipline_id = $5 AND day_of_week = $6 AND week_type = $7")).
		WithArgs(semID, gID, tID, cID, dID, int(day), string(week)).
		WillReturnRows(sqlmock.NewRows(flatLessonCols))

	_, err := repo.List(context.Background(), filter, 0, 0)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLessonRepoList_PaginationAddsLimitOffset(t *testing.T) {
	repo, mock := newLessonRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("LIMIT $1 OFFSET $2")).
		WithArgs(10, 20).
		WillReturnRows(sqlmock.NewRows(flatLessonCols))

	_, err := repo.List(context.Background(), repositories.LessonFilter{}, 10, 20)
	require.NoError(t, err)
}

func TestLessonRepoList_QueryError(t *testing.T) {
	repo, mock := newLessonRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM schedule_lessons")).
		WillReturnError(errors.New("db down"))

	lessons, err := repo.List(context.Background(), repositories.LessonFilter{}, 0, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list lessons")
	assert.Nil(t, lessons)
}

func TestLessonRepoList_ScanError(t *testing.T) {
	repo, mock := newLessonRepoMock(t)
	// Column-count mismatch → Scan error.
	rows := sqlmock.NewRows([]string{"id", "semester_id"}).AddRow(int64(1), int64(1))
	mock.ExpectQuery(regexp.QuoteMeta("FROM schedule_lessons")).WillReturnRows(rows)

	lessons, err := repo.List(context.Background(), repositories.LessonFilter{}, 0, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to scan lesson")
	assert.Nil(t, lessons)
}

// --- Count ---

func TestLessonRepoCount_HappyPath(t *testing.T) {
	repo, mock := newLessonRepoMock(t)
	semID := int64(1)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM schedule_lessons")).
		WithArgs(semID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(42)))

	count, err := repo.Count(context.Background(), repositories.LessonFilter{SemesterID: &semID})
	require.NoError(t, err)
	assert.Equal(t, int64(42), count)
}

func TestLessonRepoCount_DBError(t *testing.T) {
	repo, mock := newLessonRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM schedule_lessons")).
		WillReturnError(errors.New("count failed"))

	_, err := repo.Count(context.Background(), repositories.LessonFilter{})
	require.Error(t, err)
}

// --- GetTimetable ---

func TestLessonRepoGetTimetable_HappyPath(t *testing.T) {
	repo, mock := newLessonRepoMock(t)
	rows := sqlmock.NewRows(joinedLessonCols)
	addJoinedLessonRow(rows, int64(1))
	addJoinedLessonRow(rows, int64(2))

	mock.ExpectQuery(regexp.QuoteMeta("FROM schedule_lessons l")).
		WillReturnRows(rows)

	lessons, err := repo.GetTimetable(context.Background(), repositories.LessonFilter{})
	require.NoError(t, err)
	require.Len(t, lessons, 2)
	require.NotNil(t, lessons[0].Discipline)
	require.NotNil(t, lessons[1].Teacher)
}

func TestLessonRepoGetTimetable_WithFilter(t *testing.T) {
	repo, mock := newLessonRepoMock(t)
	gID := int64(5)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE l.group_id = $1")).
		WithArgs(gID).
		WillReturnRows(sqlmock.NewRows(joinedLessonCols))

	_, err := repo.GetTimetable(context.Background(), repositories.LessonFilter{GroupID: &gID})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestLessonRepoGetTimetable_QueryError(t *testing.T) {
	repo, mock := newLessonRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM schedule_lessons l")).
		WillReturnError(errors.New("timetable query failed"))

	lessons, err := repo.GetTimetable(context.Background(), repositories.LessonFilter{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get timetable")
	assert.Nil(t, lessons)
}

func TestLessonRepoGetTimetable_ScanError(t *testing.T) {
	repo, mock := newLessonRepoMock(t)
	// Wrong column count for joined scan.
	rows := sqlmock.NewRows([]string{"id", "name"}).AddRow(int64(1), "broken")
	mock.ExpectQuery(regexp.QuoteMeta("FROM schedule_lessons l")).WillReturnRows(rows)

	lessons, err := repo.GetTimetable(context.Background(), repositories.LessonFilter{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to scan timetable lesson")
	assert.Nil(t, lessons)
}

// --- Constructor ---

func TestNewLessonRepositoryPG_StoresHandle(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	repo := NewLessonRepositoryPG(db)
	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}
