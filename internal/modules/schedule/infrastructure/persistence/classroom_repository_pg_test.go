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

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/repositories"
)

func newClassroomRepoMock(t *testing.T) (*ClassroomRepositoryPG, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewClassroomRepositoryPG(db), mock
}

func TestClassroomRepositoryPG_GetByID_Found(t *testing.T) {
	repo, mock := newClassroomRepoMock(t)
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "building", "number", "name", "capacity", "type",
		"equipment", "is_available", "created_at", "updated_at",
	}).AddRow(
		int64(42), "A", "101", nil, 30, nil,
		[]byte(`{"projector":true}`), true, now, now,
	)

	mock.ExpectQuery(regexp.QuoteMeta("FROM classrooms WHERE id = $1")).
		WithArgs(int64(42)).
		WillReturnRows(rows)

	result, err := repo.GetByID(context.Background(), 42)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, int64(42), result.ID)
	assert.Equal(t, "A", result.Building)
	assert.Equal(t, "101", result.Number)
	assert.Equal(t, true, result.Equipment["projector"])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestClassroomRepositoryPG_GetByID_NotFound(t *testing.T) {
	repo, mock := newClassroomRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM classrooms WHERE id = $1")).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	result, err := repo.GetByID(context.Background(), 999)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestClassroomRepositoryPG_GetByID_QueryError(t *testing.T) {
	repo, mock := newClassroomRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("FROM classrooms WHERE id = $1")).
		WithArgs(int64(42)).
		WillReturnError(errors.New("db down"))

	result, err := repo.GetByID(context.Background(), 42)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get classroom")
}

func TestClassroomRepositoryPG_GetByID_InvalidEquipmentJSON(t *testing.T) {
	repo, mock := newClassroomRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "building", "number", "name", "capacity", "type",
		"equipment", "is_available", "created_at", "updated_at",
	}).AddRow(
		int64(42), "A", "101", nil, 30, nil,
		[]byte(`not-json`), true, now, now,
	)
	mock.ExpectQuery(regexp.QuoteMeta("FROM classrooms WHERE id = $1")).
		WithArgs(int64(42)).
		WillReturnRows(rows)

	result, err := repo.GetByID(context.Background(), 42)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unmarshal equipment")
}

func TestClassroomRepositoryPG_List_NoFilter(t *testing.T) {
	repo, mock := newClassroomRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "building", "number", "name", "capacity", "type",
		"equipment", "is_available", "created_at", "updated_at",
	}).
		AddRow(int64(1), "A", "101", nil, 30, nil, []byte("{}"), true, now, now).
		AddRow(int64(2), "B", "202", nil, 50, nil, nil, true, now, now)

	mock.ExpectQuery(regexp.QuoteMeta("FROM classrooms ORDER BY building, number LIMIT $1 OFFSET $2")).
		WithArgs(10, 0).
		WillReturnRows(rows)

	result, err := repo.List(context.Background(), repositories.ClassroomFilter{}, 10, 0)
	require.NoError(t, err)
	assert.Len(t, result, 2)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestClassroomRepositoryPG_List_WithAllFilters(t *testing.T) {
	repo, mock := newClassroomRepoMock(t)
	now := time.Now()
	building := "A"
	classroomType := "lecture"
	minCapacity := 30
	isAvailable := true

	rows := sqlmock.NewRows([]string{
		"id", "building", "number", "name", "capacity", "type",
		"equipment", "is_available", "created_at", "updated_at",
	}).AddRow(int64(1), "A", "101", nil, 30, "lecture", []byte("{}"), true, now, now)

	// buildWhereClause composes 4 conditions in order: building / type /
	// min_capacity / is_available. Match with substring + WithArgs pin.
	mock.ExpectQuery(`FROM classrooms WHERE building = \$1 AND type = \$2 AND capacity >= \$3 AND is_available = \$4 ORDER BY building, number`).
		WithArgs("A", "lecture", 30, true).
		WillReturnRows(rows)

	result, err := repo.List(context.Background(), repositories.ClassroomFilter{
		Building:    &building,
		Type:        &classroomType,
		MinCapacity: &minCapacity,
		IsAvailable: &isAvailable,
	}, 0, 0)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestClassroomRepositoryPG_List_QueryError(t *testing.T) {
	repo, mock := newClassroomRepoMock(t)
	mock.ExpectQuery("FROM classrooms").
		WillReturnError(errors.New("db down"))

	result, err := repo.List(context.Background(), repositories.ClassroomFilter{}, 0, 0)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to list classrooms")
}

func TestClassroomRepositoryPG_List_ScanError(t *testing.T) {
	repo, mock := newClassroomRepoMock(t)
	// Wrong column count → Scan fails inside scanClassrooms.
	rows := sqlmock.NewRows([]string{"id"}).AddRow(int64(1))
	mock.ExpectQuery("FROM classrooms").WillReturnRows(rows)

	result, err := repo.List(context.Background(), repositories.ClassroomFilter{}, 0, 0)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to scan classroom")
}

func TestClassroomRepositoryPG_Count_NoFilter(t *testing.T) {
	repo, mock := newClassroomRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM classrooms")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(7)))

	count, err := repo.Count(context.Background(), repositories.ClassroomFilter{})
	require.NoError(t, err)
	assert.Equal(t, int64(7), count)
}

func TestClassroomRepositoryPG_Count_WithBuildingFilter(t *testing.T) {
	repo, mock := newClassroomRepoMock(t)
	building := "A"
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM classrooms WHERE building = \$1`).
		WithArgs("A").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(3)))

	count, err := repo.Count(context.Background(), repositories.ClassroomFilter{Building: &building})
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
}
