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

func newDeptRepoMock(t *testing.T) (*DepartmentRepositoryPG, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	repo := NewDepartmentRepositoryPG(db)
	return repo.(*DepartmentRepositoryPG), mock
}

var deptCols = []string{"id", "name", "code", "description", "parent_id", "head_id", "is_active", "created_at", "updated_at"}

func TestDepartmentCreate_Success(t *testing.T) {
	repo, mock := newDeptRepoMock(t)
	now := time.Now()
	dept := &entities.Department{Name: "IT", Code: "IT", Description: "IT Dept", IsActive: true, CreatedAt: now, UpdatedAt: now}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO org_departments")).
		WithArgs(dept.Name, dept.Code, dept.Description, dept.ParentID, dept.HeadID, dept.IsActive, dept.CreatedAt, dept.UpdatedAt).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	err := repo.Create(context.Background(), dept)
	require.NoError(t, err)
	assert.Equal(t, int64(1), dept.ID)
}

func TestDepartmentCreate_Error(t *testing.T) {
	repo, mock := newDeptRepoMock(t)
	dept := &entities.Department{Name: "IT", Code: "IT"}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO org_departments")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("insert error"))

	err := repo.Create(context.Background(), dept)
	assert.Error(t, err)
}

func TestDepartmentGetByID_Success(t *testing.T) {
	repo, mock := newDeptRepoMock(t)
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, code, description")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(deptCols).
			AddRow(int64(1), "IT", "IT", "IT Dept", nil, nil, true, now, now))

	dept, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "IT", dept.Name)
}

func TestDepartmentGetByID_NotFound(t *testing.T) {
	repo, mock := newDeptRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, code")).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetByID(context.Background(), 999)
	assert.Error(t, err)
}

func TestDepartmentGetByCode_Success(t *testing.T) {
	repo, mock := newDeptRepoMock(t)
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, code, description")).
		WithArgs("IT").
		WillReturnRows(sqlmock.NewRows(deptCols).
			AddRow(int64(1), "IT", "IT", "IT Dept", nil, nil, true, now, now))

	dept, err := repo.GetByCode(context.Background(), "IT")
	require.NoError(t, err)
	assert.Equal(t, "IT", dept.Code)
}

func TestDepartmentGetByCode_NotFound(t *testing.T) {
	repo, mock := newDeptRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, code")).
		WithArgs("NOPE").
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetByCode(context.Background(), "NOPE")
	assert.Error(t, err)
}

func TestDepartmentUpdate_Success(t *testing.T) {
	repo, mock := newDeptRepoMock(t)
	now := time.Now()
	dept := &entities.Department{ID: 1, Name: "Updated", Code: "UPD", IsActive: true, UpdatedAt: now}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE org_departments")).
		WithArgs(dept.Name, dept.Code, dept.Description, dept.ParentID, dept.HeadID, dept.IsActive, dept.UpdatedAt, dept.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), dept)
	require.NoError(t, err)
}

func TestDepartmentUpdate_NotFound(t *testing.T) {
	repo, mock := newDeptRepoMock(t)
	dept := &entities.Department{ID: 999, Name: "Updated", Code: "UPD"}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE org_departments")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Update(context.Background(), dept)
	assert.Error(t, err)
}

func TestDepartmentUpdate_ExecError(t *testing.T) {
	repo, mock := newDeptRepoMock(t)
	dept := &entities.Department{ID: 1, Name: "Updated"}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE org_departments")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("exec error"))

	err := repo.Update(context.Background(), dept)
	assert.Error(t, err)
}

func TestDepartmentUpdate_RowsAffectedError(t *testing.T) {
	repo, mock := newDeptRepoMock(t)
	dept := &entities.Department{ID: 1, Name: "Updated"}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE org_departments")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("rows error")))

	err := repo.Update(context.Background(), dept)
	assert.Error(t, err)
}

func TestDepartmentDelete_Success(t *testing.T) {
	repo, mock := newDeptRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM org_departments")).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), 1)
	require.NoError(t, err)
}

func TestDepartmentDelete_NotFound(t *testing.T) {
	repo, mock := newDeptRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM org_departments")).
		WithArgs(int64(999)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Delete(context.Background(), 999)
	assert.Error(t, err)
}

func TestDepartmentDelete_ExecError(t *testing.T) {
	repo, mock := newDeptRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM org_departments")).
		WithArgs(int64(1)).
		WillReturnError(fmt.Errorf("delete error"))

	err := repo.Delete(context.Background(), 1)
	assert.Error(t, err)
}

func TestDepartmentDelete_RowsAffectedError(t *testing.T) {
	repo, mock := newDeptRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM org_departments")).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("rows error")))

	err := repo.Delete(context.Background(), 1)
	assert.Error(t, err)
}

func TestDepartmentList_Success(t *testing.T) {
	repo, mock := newDeptRepoMock(t)
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, code, description")).
		WithArgs(10, 0).
		WillReturnRows(sqlmock.NewRows(deptCols).
			AddRow(int64(1), "IT", "IT", "IT Dept", nil, nil, true, now, now))

	depts, err := repo.List(context.Background(), 10, 0, false)
	require.NoError(t, err)
	assert.Len(t, depts, 1)
}

func TestDepartmentList_ActiveOnly(t *testing.T) {
	repo, mock := newDeptRepoMock(t)
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta("WHERE is_active = true")).
		WithArgs(10, 0).
		WillReturnRows(sqlmock.NewRows(deptCols).
			AddRow(int64(1), "IT", "IT", "IT Dept", nil, nil, true, now, now))

	depts, err := repo.List(context.Background(), 10, 0, true)
	require.NoError(t, err)
	assert.Len(t, depts, 1)
}

func TestDepartmentList_DefaultLimits(t *testing.T) {
	repo, mock := newDeptRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, code")).
		WithArgs(10, 0).
		WillReturnRows(sqlmock.NewRows(deptCols))

	depts, err := repo.List(context.Background(), -1, -5, false)
	require.NoError(t, err)
	assert.Len(t, depts, 0)
}

func TestDepartmentList_MaxLimit(t *testing.T) {
	repo, mock := newDeptRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, code")).
		WithArgs(100, 0).
		WillReturnRows(sqlmock.NewRows(deptCols))

	_, err := repo.List(context.Background(), 500, 0, false)
	require.NoError(t, err)
}

func TestDepartmentList_QueryError(t *testing.T) {
	repo, mock := newDeptRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, code")).
		WithArgs(10, 0).
		WillReturnError(fmt.Errorf("query error"))

	_, err := repo.List(context.Background(), 10, 0, false)
	assert.Error(t, err)
}

func TestDepartmentList_ScanError(t *testing.T) {
	repo, mock := newDeptRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, code")).
		WithArgs(10, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("bad"))

	_, err := repo.List(context.Background(), 10, 0, false)
	assert.Error(t, err)
}

func TestDepartmentCount_Success(t *testing.T) {
	repo, mock := newDeptRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(5)))

	count, err := repo.Count(context.Background(), false)
	require.NoError(t, err)
	assert.Equal(t, int64(5), count)
}

func TestDepartmentCount_ActiveOnly(t *testing.T) {
	repo, mock := newDeptRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE is_active = true")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(3)))

	count, err := repo.Count(context.Background(), true)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestDepartmentCount_Error(t *testing.T) {
	repo, mock := newDeptRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnError(fmt.Errorf("count error"))

	_, err := repo.Count(context.Background(), false)
	assert.Error(t, err)
}

func TestDepartmentGetChildren_Success(t *testing.T) {
	repo, mock := newDeptRepoMock(t)
	now := time.Now()
	parentID := int64(1)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE parent_id = $1")).
		WithArgs(parentID).
		WillReturnRows(sqlmock.NewRows(deptCols).
			AddRow(int64(2), "Sub-IT", "SIT", "Sub dept", &parentID, nil, true, now, now))

	depts, err := repo.GetChildren(context.Background(), parentID)
	require.NoError(t, err)
	assert.Len(t, depts, 1)
}

func TestDepartmentGetChildren_QueryError(t *testing.T) {
	repo, mock := newDeptRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE parent_id = $1")).
		WithArgs(int64(1)).
		WillReturnError(fmt.Errorf("query error"))

	_, err := repo.GetChildren(context.Background(), 1)
	assert.Error(t, err)
}

func TestDepartmentGetChildren_ScanError(t *testing.T) {
	repo, mock := newDeptRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE parent_id = $1")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("bad"))

	_, err := repo.GetChildren(context.Background(), 1)
	assert.Error(t, err)
}
