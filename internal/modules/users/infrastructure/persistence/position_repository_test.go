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

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/domain/entities"
)

func newPosRepoMock(t *testing.T) (*PositionRepositoryPG, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	repo := NewPositionRepositoryPG(db)
	return repo.(*PositionRepositoryPG), mock
}

var posCols = []string{"id", "name", "code", "description", "level", "is_active", "created_at", "updated_at"}

func TestPositionCreate_Success(t *testing.T) {
	repo, mock := newPosRepoMock(t)
	now := time.Now()
	pos := &entities.Position{Name: "Dev", Code: "DEV", Level: 1, IsActive: true, CreatedAt: now, UpdatedAt: now}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO positions")).
		WithArgs(pos.Name, pos.Code, pos.Description, pos.Level, pos.IsActive, pos.CreatedAt, pos.UpdatedAt).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	err := repo.Create(context.Background(), pos)
	require.NoError(t, err)
	assert.Equal(t, int64(1), pos.ID)
}

func TestPositionCreate_Error(t *testing.T) {
	repo, mock := newPosRepoMock(t)
	pos := &entities.Position{Name: "Dev", Code: "DEV"}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO positions")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("insert error"))

	err := repo.Create(context.Background(), pos)
	assert.Error(t, err)
}

func TestPositionGetByID_Success(t *testing.T) {
	repo, mock := newPosRepoMock(t)
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, code")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(posCols).AddRow(int64(1), "Dev", "DEV", "Developer", 1, true, now, now))

	pos, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "Dev", pos.Name)
}

func TestPositionGetByID_NotFound(t *testing.T) {
	repo, mock := newPosRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, code")).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetByID(context.Background(), 999)
	assert.Error(t, err)
}

func TestPositionGetByCode_Success(t *testing.T) {
	repo, mock := newPosRepoMock(t)
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, code")).
		WithArgs("DEV").
		WillReturnRows(sqlmock.NewRows(posCols).AddRow(int64(1), "Dev", "DEV", "Developer", 1, true, now, now))

	pos, err := repo.GetByCode(context.Background(), "DEV")
	require.NoError(t, err)
	assert.Equal(t, "DEV", pos.Code)
}

func TestPositionGetByCode_NotFound(t *testing.T) {
	repo, mock := newPosRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, code")).
		WithArgs("NOPE").
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetByCode(context.Background(), "NOPE")
	assert.Error(t, err)
}

func TestPositionUpdate_Success(t *testing.T) {
	repo, mock := newPosRepoMock(t)
	now := time.Now()
	pos := &entities.Position{ID: 1, Name: "Updated", Code: "UPD", Level: 2, IsActive: true, UpdatedAt: now}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE positions")).
		WithArgs(pos.Name, pos.Code, pos.Description, pos.Level, pos.IsActive, pos.UpdatedAt, pos.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), pos)
	require.NoError(t, err)
}

func TestPositionUpdate_NotFound(t *testing.T) {
	repo, mock := newPosRepoMock(t)
	pos := &entities.Position{ID: 999, Name: "Updated"}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE positions")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Update(context.Background(), pos)
	assert.Error(t, err)
}

func TestPositionUpdate_ExecError(t *testing.T) {
	repo, mock := newPosRepoMock(t)
	pos := &entities.Position{ID: 1, Name: "Updated"}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE positions")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("exec error"))

	err := repo.Update(context.Background(), pos)
	assert.Error(t, err)
}

func TestPositionUpdate_RowsAffectedError(t *testing.T) {
	repo, mock := newPosRepoMock(t)
	pos := &entities.Position{ID: 1, Name: "Updated"}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE positions")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("rows error")))

	err := repo.Update(context.Background(), pos)
	assert.Error(t, err)
}

func TestPositionDelete_Success(t *testing.T) {
	repo, mock := newPosRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM positions")).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), 1)
	require.NoError(t, err)
}

func TestPositionDelete_NotFound(t *testing.T) {
	repo, mock := newPosRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM positions")).
		WithArgs(int64(999)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Delete(context.Background(), 999)
	assert.Error(t, err)
}

func TestPositionDelete_ExecError(t *testing.T) {
	repo, mock := newPosRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM positions")).
		WithArgs(int64(1)).
		WillReturnError(fmt.Errorf("delete error"))

	err := repo.Delete(context.Background(), 1)
	assert.Error(t, err)
}

func TestPositionDelete_RowsAffectedError(t *testing.T) {
	repo, mock := newPosRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM positions")).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("rows error")))

	err := repo.Delete(context.Background(), 1)
	assert.Error(t, err)
}

func TestPositionList_Success(t *testing.T) {
	repo, mock := newPosRepoMock(t)
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, code")).
		WithArgs(10, 0).
		WillReturnRows(sqlmock.NewRows(posCols).AddRow(int64(1), "Dev", "DEV", "Developer", 1, true, now, now))

	positions, err := repo.List(context.Background(), 10, 0, false)
	require.NoError(t, err)
	assert.Len(t, positions, 1)
}

func TestPositionList_ActiveOnly(t *testing.T) {
	repo, mock := newPosRepoMock(t)
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta("WHERE is_active = true")).
		WithArgs(10, 0).
		WillReturnRows(sqlmock.NewRows(posCols).AddRow(int64(1), "Dev", "DEV", "", 1, true, now, now))

	_, err := repo.List(context.Background(), 10, 0, true)
	require.NoError(t, err)
}

func TestPositionList_DefaultLimits(t *testing.T) {
	repo, mock := newPosRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, code")).
		WithArgs(10, 0).
		WillReturnRows(sqlmock.NewRows(posCols))

	_, err := repo.List(context.Background(), -1, -5, false)
	require.NoError(t, err)
}

func TestPositionList_MaxLimit(t *testing.T) {
	repo, mock := newPosRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, code")).
		WithArgs(100, 0).
		WillReturnRows(sqlmock.NewRows(posCols))

	_, err := repo.List(context.Background(), 500, 0, false)
	require.NoError(t, err)
}

func TestPositionList_QueryError(t *testing.T) {
	repo, mock := newPosRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, code")).
		WithArgs(10, 0).
		WillReturnError(fmt.Errorf("query error"))

	_, err := repo.List(context.Background(), 10, 0, false)
	assert.Error(t, err)
}

func TestPositionList_ScanError(t *testing.T) {
	repo, mock := newPosRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, code")).
		WithArgs(10, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("bad"))

	_, err := repo.List(context.Background(), 10, 0, false)
	assert.Error(t, err)
}

func TestPositionCount_Success(t *testing.T) {
	repo, mock := newPosRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(5)))

	count, err := repo.Count(context.Background(), false)
	require.NoError(t, err)
	assert.Equal(t, int64(5), count)
}

func TestPositionCount_ActiveOnly(t *testing.T) {
	repo, mock := newPosRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE is_active = true")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(3)))

	count, err := repo.Count(context.Background(), true)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestPositionCount_Error(t *testing.T) {
	repo, mock := newPosRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnError(fmt.Errorf("count error"))

	_, err := repo.Count(context.Background(), false)
	assert.Error(t, err)
}
