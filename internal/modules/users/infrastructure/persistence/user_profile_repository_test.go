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

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/domain/repositories"
)

func newProfileRepoMock(t *testing.T) (*UserProfileRepositoryPG, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	repo := NewUserProfileRepositoryPG(db)
	return repo.(*UserProfileRepositoryPG), mock
}

var profileCols = []string{
	"id", "email", "name", "role", "status",
	"phone", "avatar", "bio",
	"department_id", "department_name",
	"position_id", "position_name",
	"created_at", "updated_at",
}

var listProfileCols = []string{
	"id", "email", "name", "role", "status",
	"phone", "avatar",
	"department_id", "department_name",
	"position_id", "position_name",
	"created_at", "updated_at",
}

func TestProfileGetByID_Success(t *testing.T) {
	repo, mock := newProfileRepoMock(t)
	now := time.Now()
	deptID := int64(1)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(profileCols).
			AddRow(int64(1), "test@test.com", "John", "admin", "active",
				"123456", "", "bio",
				deptID, "IT",
				nil, "",
				now, now))

	user, err := repo.GetProfileByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "John", user.Name)
	assert.NotNil(t, user.DepartmentID)
	assert.Equal(t, int64(1), *user.DepartmentID)
	assert.Nil(t, user.PositionID)
}

func TestProfileGetByID_NotFound(t *testing.T) {
	repo, mock := newProfileRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetProfileByID(context.Background(), 999)
	assert.Error(t, err)
}

func TestProfileUpdateProfile_Success(t *testing.T) {
	repo, mock := newProfileRepoMock(t)
	deptID := int64(1)
	posID := int64(2)

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO user_profiles")).
		WithArgs(int64(10), &deptID, &posID, "123", "/avatar.png", "bio", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdateProfile(context.Background(), 10, &deptID, &posID, "123", "/avatar.png", "bio")
	require.NoError(t, err)
}

func TestProfileUpdateProfile_Error(t *testing.T) {
	repo, mock := newProfileRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO user_profiles")).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(fmt.Errorf("upsert error"))

	err := repo.UpdateProfile(context.Background(), 10, nil, nil, "", "", "")
	assert.Error(t, err)
}

func TestProfileListUsersWithOrg_NoFilter(t *testing.T) {
	repo, mock := newProfileRepoMock(t)
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
		WithArgs(10, 0).
		WillReturnRows(sqlmock.NewRows(listProfileCols).
			AddRow(int64(1), "a@b.com", "A", "user", "active", "", "", nil, "", nil, "", now, now))

	users, err := repo.ListUsersWithOrg(context.Background(), nil, 10, 0)
	require.NoError(t, err)
	assert.Len(t, users, 1)
}

func TestProfileListUsersWithOrg_AllFilters(t *testing.T) {
	repo, mock := newProfileRepoMock(t)
	now := time.Now()
	deptID := int64(1)
	posID := int64(2)

	filter := &repositories.UserFilter{
		DepartmentID: &deptID,
		PositionID:   &posID,
		Role:         "admin",
		Status:       "active",
		Search:       "John",
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
		WithArgs(deptID, posID, "admin", "active", "%John%", 10, 0).
		WillReturnRows(sqlmock.NewRows(listProfileCols).
			AddRow(int64(1), "john@test.com", "John", "admin", "active", "123", "", deptID, "IT", posID, "Dev", now, now))

	users, err := repo.ListUsersWithOrg(context.Background(), filter, 10, 0)
	require.NoError(t, err)
	assert.Len(t, users, 1)
	assert.NotNil(t, users[0].DepartmentID)
	assert.NotNil(t, users[0].PositionID)
}

func TestProfileListUsersWithOrg_DefaultLimits(t *testing.T) {
	repo, mock := newProfileRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
		WithArgs(10, 0).
		WillReturnRows(sqlmock.NewRows(listProfileCols))

	_, err := repo.ListUsersWithOrg(context.Background(), nil, -1, -5)
	require.NoError(t, err)
}

func TestProfileListUsersWithOrg_MaxLimit(t *testing.T) {
	repo, mock := newProfileRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
		WithArgs(100, 0).
		WillReturnRows(sqlmock.NewRows(listProfileCols))

	_, err := repo.ListUsersWithOrg(context.Background(), nil, 500, 0)
	require.NoError(t, err)
}

func TestProfileListUsersWithOrg_QueryError(t *testing.T) {
	repo, mock := newProfileRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
		WithArgs(10, 0).
		WillReturnError(fmt.Errorf("query error"))

	_, err := repo.ListUsersWithOrg(context.Background(), nil, 10, 0)
	assert.Error(t, err)
}

func TestProfileListUsersWithOrg_ScanError(t *testing.T) {
	repo, mock := newProfileRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
		WithArgs(10, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("bad"))

	_, err := repo.ListUsersWithOrg(context.Background(), nil, 10, 0)
	assert.Error(t, err)
}

func TestProfileCountUsers_NoFilter(t *testing.T) {
	repo, mock := newProfileRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(10)))

	count, err := repo.CountUsers(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, int64(10), count)
}

func TestProfileCountUsers_AllFilters(t *testing.T) {
	repo, mock := newProfileRepoMock(t)
	deptID := int64(1)
	posID := int64(2)

	filter := &repositories.UserFilter{
		DepartmentID: &deptID,
		PositionID:   &posID,
		Role:         "admin",
		Status:       "active",
		Search:       "John",
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WithArgs(deptID, posID, "admin", "active", "%John%").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	count, err := repo.CountUsers(context.Background(), filter)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestProfileCountUsers_Error(t *testing.T) {
	repo, mock := newProfileRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnError(fmt.Errorf("count error"))

	_, err := repo.CountUsers(context.Background(), nil)
	assert.Error(t, err)
}

func TestProfileGetUsersByDepartment(t *testing.T) {
	repo, mock := newProfileRepoMock(t)
	now := time.Now()
	deptID := int64(1)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
		WithArgs(deptID, 100, 0).
		WillReturnRows(sqlmock.NewRows(listProfileCols).
			AddRow(int64(1), "a@b.com", "A", "user", "active", "", "", deptID, "IT", nil, "", now, now))

	users, err := repo.GetUsersByDepartment(context.Background(), deptID)
	require.NoError(t, err)
	assert.Len(t, users, 1)
}

func TestProfileGetUsersByPosition(t *testing.T) {
	repo, mock := newProfileRepoMock(t)
	now := time.Now()
	posID := int64(2)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
		WithArgs(posID, 100, 0).
		WillReturnRows(sqlmock.NewRows(listProfileCols).
			AddRow(int64(1), "a@b.com", "A", "user", "active", "", "", nil, "", posID, "Dev", now, now))

	users, err := repo.GetUsersByPosition(context.Background(), posID)
	require.NoError(t, err)
	assert.Len(t, users, 1)
}

func TestProfileBulkUpdateDepartment_Empty(t *testing.T) {
	repo, _ := newProfileRepoMock(t)

	err := repo.BulkUpdateDepartment(context.Background(), []int64{}, nil)
	require.NoError(t, err)
}

func TestProfileBulkUpdateDepartment_Success(t *testing.T) {
	repo, mock := newProfileRepoMock(t)
	deptID := int64(5)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE user_profiles")).
		WithArgs(&deptID, sqlmock.AnyArg(), int64(1), int64(2)).
		WillReturnResult(sqlmock.NewResult(0, 2))

	err := repo.BulkUpdateDepartment(context.Background(), []int64{1, 2}, &deptID)
	require.NoError(t, err)
}

func TestProfileBulkUpdateDepartment_Error(t *testing.T) {
	repo, mock := newProfileRepoMock(t)
	deptID := int64(5)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE user_profiles")).
		WithArgs(&deptID, sqlmock.AnyArg(), int64(1)).
		WillReturnError(fmt.Errorf("update error"))

	err := repo.BulkUpdateDepartment(context.Background(), []int64{1}, &deptID)
	assert.Error(t, err)
}

func TestProfileBulkUpdatePosition_Empty(t *testing.T) {
	repo, _ := newProfileRepoMock(t)

	err := repo.BulkUpdatePosition(context.Background(), []int64{}, nil)
	require.NoError(t, err)
}

func TestProfileBulkUpdatePosition_Success(t *testing.T) {
	repo, mock := newProfileRepoMock(t)
	posID := int64(3)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE user_profiles")).
		WithArgs(&posID, sqlmock.AnyArg(), int64(1), int64(2)).
		WillReturnResult(sqlmock.NewResult(0, 2))

	err := repo.BulkUpdatePosition(context.Background(), []int64{1, 2}, &posID)
	require.NoError(t, err)
}

func TestProfileBulkUpdatePosition_Error(t *testing.T) {
	repo, mock := newProfileRepoMock(t)
	posID := int64(3)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE user_profiles")).
		WithArgs(&posID, sqlmock.AnyArg(), int64(1)).
		WillReturnError(fmt.Errorf("update error"))

	err := repo.BulkUpdatePosition(context.Background(), []int64{1}, &posID)
	assert.Error(t, err)
}
