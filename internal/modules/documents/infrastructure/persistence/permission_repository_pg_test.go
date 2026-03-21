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

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
	domainErrors "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/domain/errors"
)

func newPermRepoMock(t *testing.T) (*PermissionRepositoryPG, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewPermissionRepositoryPG(db), mock
}

var permCols = []string{"id", "document_id", "user_id", "role", "permission", "granted_by", "expires_at", "created_at"}
var permJoinCols = []string{
	"id", "document_id", "user_id", "role", "permission",
	"granted_by", "expires_at", "created_at",
	"user_name", "user_email", "granted_by_name",
}

func TestNewPermissionRepositoryPG(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer func() { _ = db.Close() }()
	assert.NotNil(t, NewPermissionRepositoryPG(db))
}

func TestPermRepo_Create_Success(t *testing.T) {
	repo, mock := newPermRepoMock(t)
	p := &entities.DocumentPermission{DocumentID: 1, UserID: ptrInt64(10), Permission: "read", GrantedBy: ptrInt64(5)}
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO document_permissions")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	err := repo.Create(context.Background(), p)
	require.NoError(t, err)
	assert.Equal(t, int64(1), p.ID)
}

func TestPermRepo_Create_Error(t *testing.T) {
	repo, mock := newPermRepoMock(t)
	p := &entities.DocumentPermission{DocumentID: 1}
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO document_permissions")).
		WillReturnError(fmt.Errorf("db err"))

	err := repo.Create(context.Background(), p)
	assert.Error(t, err)
}

func TestPermRepo_Update_Success(t *testing.T) {
	repo, mock := newPermRepoMock(t)
	p := &entities.DocumentPermission{ID: 1, Permission: "write"}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE document_permissions")).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), p)
	require.NoError(t, err)
}

func TestPermRepo_Update_NotFound(t *testing.T) {
	repo, mock := newPermRepoMock(t)
	p := &entities.DocumentPermission{ID: 999, Permission: "write"}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE document_permissions")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Update(context.Background(), p)
	assert.ErrorIs(t, err, domainErrors.ErrNotFound)
}

func TestPermRepo_Delete_Success(t *testing.T) {
	repo, mock := newPermRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM document_permissions WHERE id = $1")).
		WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), 1)
	require.NoError(t, err)
}

func TestPermRepo_Delete_NotFound(t *testing.T) {
	repo, mock := newPermRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM document_permissions WHERE id = $1")).
		WithArgs(int64(999)).WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Delete(context.Background(), 999)
	assert.ErrorIs(t, err, domainErrors.ErrNotFound)
}

func TestPermRepo_GetByID_Success(t *testing.T) {
	repo, mock := newPermRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows(permJoinCols).AddRow(
		1, 1, ptrInt64(10), nil, "read", int64(5), nil, now,
		ptrStr("User"), ptrStr("user@test.com"), ptrStr("Admin"),
	)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT dp.id")).
		WithArgs(int64(1)).WillReturnRows(rows)

	p, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), p.ID)
}

func TestPermRepo_GetByID_NotFound(t *testing.T) {
	repo, mock := newPermRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT dp.id")).
		WithArgs(int64(999)).WillReturnError(sql.ErrNoRows)

	_, err := repo.GetByID(context.Background(), 999)
	assert.ErrorIs(t, err, domainErrors.ErrNotFound)
}

func TestPermRepo_GetByDocumentID(t *testing.T) {
	repo, mock := newPermRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows(permJoinCols).AddRow(
		1, 1, ptrInt64(10), nil, "read", int64(5), nil, now,
		ptrStr("User"), ptrStr("user@test.com"), ptrStr("Admin"),
	)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE dp.document_id = $1")).
		WithArgs(int64(1)).WillReturnRows(rows)

	perms, err := repo.GetByDocumentID(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, perms, 1)
}

func TestPermRepo_GetByUserID(t *testing.T) {
	repo, mock := newPermRepoMock(t)
	rows := sqlmock.NewRows(permJoinCols)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE dp.user_id = $1")).
		WithArgs(int64(10)).WillReturnRows(rows)

	perms, err := repo.GetByUserID(context.Background(), 10)
	require.NoError(t, err)
	assert.Empty(t, perms)
}

func TestPermRepo_GetByUserIDOrRole(t *testing.T) {
	repo, mock := newPermRepoMock(t)
	rows := sqlmock.NewRows(permJoinCols)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE (dp.user_id = $1 OR dp.role = $2)")).
		WithArgs(int64(10), "admin").WillReturnRows(rows)

	perms, err := repo.GetByUserIDOrRole(context.Background(), 10, "admin")
	require.NoError(t, err)
	assert.Empty(t, perms)
}

func TestPermRepo_GetByGrantedBy(t *testing.T) {
	repo, mock := newPermRepoMock(t)
	rows := sqlmock.NewRows(permJoinCols)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE dp.granted_by = $1")).
		WithArgs(int64(5)).WillReturnRows(rows)

	perms, err := repo.GetByGrantedBy(context.Background(), 5)
	require.NoError(t, err)
	assert.Empty(t, perms)
}

func TestPermRepo_GetByDocumentAndUser(t *testing.T) {
	repo, mock := newPermRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows(permCols).AddRow(1, 1, ptrInt64(10), nil, "read", int64(5), nil, now)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE document_id = $1 AND user_id = $2")).
		WithArgs(int64(1), int64(10)).WillReturnRows(rows)

	p, err := repo.GetByDocumentAndUser(context.Background(), 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), p.ID)
}

func TestPermRepo_GetByDocumentAndUser_NotFound(t *testing.T) {
	repo, mock := newPermRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE document_id = $1 AND user_id = $2")).
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetByDocumentAndUser(context.Background(), 1, 999)
	assert.ErrorIs(t, err, domainErrors.ErrNotFound)
}

func TestPermRepo_GetByDocumentAndRole(t *testing.T) {
	repo, mock := newPermRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows(permCols).AddRow(1, 1, nil, ptrStr("admin"), "read", int64(5), nil, now)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE document_id = $1 AND role = $2")).
		WithArgs(int64(1), entities.UserRole("admin")).WillReturnRows(rows)

	p, err := repo.GetByDocumentAndRole(context.Background(), 1, "admin")
	require.NoError(t, err)
	assert.Equal(t, int64(1), p.ID)
}

func TestPermRepo_HasPermission(t *testing.T) {
	repo, mock := newPermRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS")).
		WithArgs(int64(1), int64(10), entities.PermissionLevel("read")).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	has, err := repo.HasPermission(context.Background(), 1, 10, "read")
	require.NoError(t, err)
	assert.True(t, has)
}

func TestPermRepo_HasAnyPermission(t *testing.T) {
	repo, mock := newPermRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS")).
		WithArgs(int64(1), int64(10)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	has, err := repo.HasAnyPermission(context.Background(), 1, 10)
	require.NoError(t, err)
	assert.False(t, has)
}

func TestPermRepo_GetUserPermissionLevel(t *testing.T) {
	repo, mock := newPermRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT dp.permission")).
		WithArgs(int64(1), int64(10), entities.UserRole("user")).
		WillReturnRows(sqlmock.NewRows([]string{"permission"}).AddRow("write"))

	level, err := repo.GetUserPermissionLevel(context.Background(), 1, 10, "user")
	require.NoError(t, err)
	assert.NotNil(t, level)
	assert.Equal(t, entities.PermissionLevel("write"), *level)
}

func TestPermRepo_GetUserPermissionLevel_NoPermission(t *testing.T) {
	repo, mock := newPermRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT dp.permission")).
		WillReturnError(sql.ErrNoRows)

	level, err := repo.GetUserPermissionLevel(context.Background(), 1, 10, "user")
	require.NoError(t, err)
	assert.Nil(t, level)
}

func TestPermRepo_DeleteByDocumentID(t *testing.T) {
	repo, mock := newPermRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM document_permissions WHERE document_id = $1")).
		WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 3))

	err := repo.DeleteByDocumentID(context.Background(), 1)
	require.NoError(t, err)
}

func TestPermRepo_DeleteByUserID(t *testing.T) {
	repo, mock := newPermRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM document_permissions WHERE user_id = $1")).
		WithArgs(int64(10)).WillReturnResult(sqlmock.NewResult(0, 2))

	err := repo.DeleteByUserID(context.Background(), 10)
	require.NoError(t, err)
}

func TestPermRepo_DeleteExpired(t *testing.T) {
	repo, mock := newPermRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM document_permissions WHERE expires_at IS NOT NULL")).
		WillReturnResult(sqlmock.NewResult(0, 5))

	count, err := repo.DeleteExpired(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(5), count)
}

func TestPermRepo_GetByUserID_Success(t *testing.T) {
	repo, mock := newPermRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows(permJoinCols).AddRow(
		int64(1), int64(10), ptrInt64(5), ptrStr("admin"), "read",
		int64(1), nil, now, ptrStr("User"), ptrStr("u@e.com"), ptrStr("Admin"),
	)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE dp.user_id")).
		WithArgs(int64(5)).
		WillReturnRows(rows)
	perms, err := repo.GetByUserID(context.Background(), 5)
	require.NoError(t, err)
	assert.Len(t, perms, 1)
}

func TestPermRepo_GetByUserID_Error(t *testing.T) {
	repo, mock := newPermRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE dp.user_id")).
		WithArgs(int64(5)).
		WillReturnError(fmt.Errorf("db error"))
	_, err := repo.GetByUserID(context.Background(), 5)
	assert.Error(t, err)
}

func TestPermRepo_GetByUserIDOrRole_Success(t *testing.T) {
	repo, mock := newPermRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows(permJoinCols).AddRow(
		int64(1), int64(10), ptrInt64(5), ptrStr("teacher"), "read",
		int64(1), nil, now, ptrStr("User"), ptrStr("u@e.com"), ptrStr("Admin"),
	)
	mock.ExpectQuery(regexp.QuoteMeta("dp.user_id = $1 OR dp.role = $2")).
		WithArgs(int64(5), "teacher").
		WillReturnRows(rows)
	perms, err := repo.GetByUserIDOrRole(context.Background(), 5, "teacher")
	require.NoError(t, err)
	assert.Len(t, perms, 1)
}

func TestPermRepo_GetByUserIDOrRole_Error(t *testing.T) {
	repo, mock := newPermRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("dp.user_id = $1 OR dp.role = $2")).
		WithArgs(int64(5), "teacher").
		WillReturnError(fmt.Errorf("db error"))
	_, err := repo.GetByUserIDOrRole(context.Background(), 5, "teacher")
	assert.Error(t, err)
}

func TestPermRepo_GetByGrantedBy_Success(t *testing.T) {
	repo, mock := newPermRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows(permJoinCols).AddRow(
		int64(1), int64(10), ptrInt64(5), ptrStr("admin"), "write",
		int64(1), nil, now, ptrStr("User"), ptrStr("u@e.com"), ptrStr("Admin"),
	)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE dp.granted_by")).
		WithArgs(int64(1)).
		WillReturnRows(rows)
	perms, err := repo.GetByGrantedBy(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, perms, 1)
}

func TestPermRepo_GetByGrantedBy_Error(t *testing.T) {
	repo, mock := newPermRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE dp.granted_by")).
		WithArgs(int64(1)).
		WillReturnError(fmt.Errorf("db error"))
	_, err := repo.GetByGrantedBy(context.Background(), 1)
	assert.Error(t, err)
}

func TestPermRepo_GetByDocumentAndRole_Success(t *testing.T) {
	repo, mock := newPermRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows(permCols).AddRow(
		int64(1), int64(10), nil, ptrStr("admin"), "read",
		int64(1), nil, now,
	)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE document_id = $1 AND role = $2")).
		WithArgs(int64(10), entities.UserRole("admin")).
		WillReturnRows(rows)
	perm, err := repo.GetByDocumentAndRole(context.Background(), 10, "admin")
	require.NoError(t, err)
	assert.NotNil(t, perm)
}

func TestPermRepo_GetByDocumentAndRole_NotFound(t *testing.T) {
	repo, mock := newPermRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE document_id = $1 AND role = $2")).
		WillReturnError(sql.ErrNoRows)
	_, err := repo.GetByDocumentAndRole(context.Background(), 10, "admin")
	assert.ErrorIs(t, err, domainErrors.ErrNotFound)
}

func TestPermRepo_GetByDocumentAndRole_Error(t *testing.T) {
	repo, mock := newPermRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("WHERE document_id = $1 AND role = $2")).
		WillReturnError(fmt.Errorf("db error"))
	_, err := repo.GetByDocumentAndRole(context.Background(), 10, "admin")
	assert.Error(t, err)
}

func TestPermRepo_Update_Error(t *testing.T) {
	repo, mock := newPermRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE document_permissions SET")).WillReturnError(fmt.Errorf("db"))
	assert.Error(t, repo.Update(context.Background(), &entities.DocumentPermission{ID: 1}))
}

func TestPermRepo_Delete_Error(t *testing.T) {
	repo, mock := newPermRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM document_permissions WHERE id")).WillReturnError(fmt.Errorf("db"))
	assert.Error(t, repo.Delete(context.Background(), 1))
}

func TestPermRepo_DeleteByDocumentID_Error(t *testing.T) {
	repo, mock := newPermRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM document_permissions WHERE document_id")).WillReturnError(fmt.Errorf("db"))
	assert.Error(t, repo.DeleteByDocumentID(context.Background(), 1))
}

func TestPermRepo_DeleteByUserID_Error(t *testing.T) {
	repo, mock := newPermRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM document_permissions WHERE user_id")).WillReturnError(fmt.Errorf("db"))
	assert.Error(t, repo.DeleteByUserID(context.Background(), 1))
}

func TestPermRepo_DeleteExpired_Error(t *testing.T) {
	repo, mock := newPermRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM document_permissions WHERE expires_at")).WillReturnError(fmt.Errorf("db"))
	_, err := repo.DeleteExpired(context.Background())
	assert.Error(t, err)
}

// Helper functions
func ptrInt64(v int64) *int64 { return &v }
func ptrStr(v string) *string { return &v }
