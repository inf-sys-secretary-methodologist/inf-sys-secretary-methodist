package persistence

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
)

func newLoadRepoMock(t *testing.T) (*TeachingLoadRepositoryPG, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewTeachingLoadRepositoryPG(db), mock
}

func sampleLoad(now time.Time) *entities.TeachingLoad {
	return &entities.TeachingLoad{
		SemesterID: 1, GroupID: 2, DisciplineID: 3, TeacherID: 4, LessonTypeID: 5,
		PairsPerWeek: 2, WeekType: domain.WeekTypeAll, CreatedAt: now, UpdatedAt: now,
	}
}

func TestTeachingLoadRepositoryPG_Create_Success(t *testing.T) {
	repo, mock := newLoadRepoMock(t)
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	load := sampleLoad(now)

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO teaching_loads")).
		WithArgs(int64(1), int64(2), int64(3), int64(4), int64(5), 2, "all", now, now).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(9)))

	err := repo.Create(context.Background(), load)
	require.NoError(t, err)
	assert.Equal(t, int64(9), load.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTeachingLoadRepositoryPG_Create_Duplicate(t *testing.T) {
	repo, mock := newLoadRepoMock(t)
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	load := sampleLoad(now)

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO teaching_loads")).
		WithArgs(int64(1), int64(2), int64(3), int64(4), int64(5), 2, "all", now, now).
		WillReturnError(&pq.Error{Code: "23505"})

	err := repo.Create(context.Background(), load)
	assert.ErrorIs(t, err, entities.ErrTeachingLoadDuplicate)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTeachingLoadRepositoryPG_Update_NotFound(t *testing.T) {
	repo, mock := newLoadRepoMock(t)
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	load := sampleLoad(now)
	load.ID = 99

	mock.ExpectExec(regexp.QuoteMeta("UPDATE teaching_loads")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Update(context.Background(), load)
	assert.ErrorIs(t, err, entities.ErrTeachingLoadNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTeachingLoadRepositoryPG_Update_Duplicate(t *testing.T) {
	repo, mock := newLoadRepoMock(t)
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	load := sampleLoad(now)
	load.ID = 9

	mock.ExpectExec(regexp.QuoteMeta("UPDATE teaching_loads")).
		WillReturnError(&pq.Error{Code: "23505"})

	err := repo.Update(context.Background(), load)
	assert.ErrorIs(t, err, entities.ErrTeachingLoadDuplicate)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTeachingLoadRepositoryPG_Delete_NotFound(t *testing.T) {
	repo, mock := newLoadRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM teaching_loads WHERE id = $1")).
		WithArgs(int64(99)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Delete(context.Background(), 99)
	assert.ErrorIs(t, err, entities.ErrTeachingLoadNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTeachingLoadRepositoryPG_GetByID_NotFound(t *testing.T) {
	repo, mock := newLoadRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("FROM teaching_loads")).
		WithArgs(int64(99)).
		WillReturnError(sql.ErrNoRows)

	got, err := repo.GetByID(context.Background(), 99)
	assert.Nil(t, got)
	assert.ErrorIs(t, err, entities.ErrTeachingLoadNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func loadHydratedRows(now time.Time) *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "semester_id", "group_id", "discipline_id", "teacher_id", "lesson_type_id",
		"pairs_per_week", "week_type", "created_at", "updated_at",
		"g_id", "g_name", "d_id", "d_name", "lt_id", "lt_name", "lt_short", "u_id", "u_name", "u_email",
	}).AddRow(
		int64(9), int64(1), int64(2), int64(3), int64(4), int64(5),
		2, "all", now, now,
		int64(2), "IS-21", int64(3), "Матанализ", int64(5), "Лекция", "лек", int64(4), "Иванов И.", "iv@u.ru",
	)
}

func TestTeachingLoadRepositoryPG_GetByID_Hydrated(t *testing.T) {
	repo, mock := newLoadRepoMock(t)
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	mock.ExpectQuery(regexp.QuoteMeta("FROM teaching_loads")).
		WithArgs(int64(9)).
		WillReturnRows(loadHydratedRows(now))

	got, err := repo.GetByID(context.Background(), 9)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, int64(9), got.ID)
	assert.Equal(t, 2, got.PairsPerWeek)
	require.NotNil(t, got.Group)
	assert.Equal(t, "IS-21", got.Group.Name)
	require.NotNil(t, got.Discipline)
	assert.Equal(t, "Матанализ", got.Discipline.Name)
	require.NotNil(t, got.Teacher)
	assert.Equal(t, "Иванов И.", got.Teacher.Name)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTeachingLoadRepositoryPG_List_FilterBySemester(t *testing.T) {
	repo, mock := newLoadRepoMock(t)
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	semester := int64(1)

	mock.ExpectQuery(regexp.QuoteMeta("FROM teaching_loads")).
		WithArgs(int64(1)).
		WillReturnRows(loadHydratedRows(now))

	got, err := repo.List(context.Background(), usecases.TeachingLoadFilter{SemesterID: &semester})
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "IS-21", got[0].Group.Name)
	require.NoError(t, mock.ExpectationsWereMet())
}
