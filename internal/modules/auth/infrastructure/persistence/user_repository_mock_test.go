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

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"
)

func newUserRepoMock(t *testing.T) (*UserRepositoryPG, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return &UserRepositoryPG{db: db}, mock
}

var userCols = []string{"id", "email", "password", "name", "role", "status", "created_at", "updated_at"}

func TestNewUserRepositoryPG(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer func() { _ = db.Close() }()
	repo := NewUserRepositoryPG(db)
	assert.NotNil(t, repo)
}

// --- Create ---

func TestUserRepo_Create_Success(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	user := entities.NewUser("test@test.com", "hash", "Test", domain.RoleTeacher)
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO users")).
		WithArgs(user.Email, user.Password, user.Name, user.Role, user.Status, user.CreatedAt, user.UpdatedAt).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
	err := repo.Create(context.Background(), user)
	require.NoError(t, err)
	assert.Equal(t, int64(1), user.ID)
}

func TestUserRepo_Create_Error(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	user := entities.NewUser("test@test.com", "hash", "Test", domain.RoleTeacher)
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO users")).
		WillReturnError(fmt.Errorf("db error"))
	err := repo.Create(context.Background(), user)
	assert.Error(t, err)
}

// --- Save ---

func TestUserRepo_Save_Success(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	user := &entities.User{ID: 1, Email: "test@test.com", Password: "hash", Name: "Test", Role: domain.RoleTeacher, Status: entities.UserStatusActive, UpdatedAt: time.Now()}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE users")).
		WithArgs(user.Email, user.Password, user.Name, user.Role, user.Status, user.UpdatedAt, user.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	err := repo.Save(context.Background(), user)
	require.NoError(t, err)
}

func TestUserRepo_Save_NotFound(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	user := &entities.User{ID: 999, Email: "test@test.com", UpdatedAt: time.Now()}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE users")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	err := repo.Save(context.Background(), user)
	assert.Error(t, err)
}

func TestUserRepo_Save_DBError(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	user := &entities.User{ID: 1, UpdatedAt: time.Now()}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE users")).
		WillReturnError(fmt.Errorf("db error"))
	err := repo.Save(context.Background(), user)
	assert.Error(t, err)
}

func TestUserRepo_Save_RowsAffectedError(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	user := &entities.User{ID: 1, UpdatedAt: time.Now()}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE users")).
		WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("rows affected error")))
	err := repo.Save(context.Background(), user)
	assert.Error(t, err)
}

// --- GetByID ---

func TestUserRepo_GetByID_Success(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows(userCols).AddRow(int64(1), "test@test.com", "hash", "Test", domain.RoleTeacher, entities.UserStatusActive, now, now)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, password, name, role, status, created_at, updated_at")).
		WithArgs(int64(1)).WillReturnRows(rows)
	user, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, "test@test.com", user.Email)
	assert.Equal(t, "Test", user.Name)
}

func TestUserRepo_GetByID_NotFound(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, password, name, role, status, created_at, updated_at")).
		WithArgs(int64(999)).WillReturnError(sql.ErrNoRows)
	_, err := repo.GetByID(context.Background(), 999)
	assert.Error(t, err)
}

func TestUserRepo_GetByID_DBError(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, password, name, role, status, created_at, updated_at")).
		WithArgs(int64(1)).WillReturnError(fmt.Errorf("connection error"))
	_, err := repo.GetByID(context.Background(), 1)
	assert.Error(t, err)
}

// --- GetByEmail ---

func TestUserRepo_GetByEmail_Success(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows(userCols).AddRow(int64(1), "test@test.com", "hash", "Test", domain.RoleTeacher, entities.UserStatusActive, now, now)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, password, name, role, status, created_at, updated_at")).
		WithArgs("test@test.com").WillReturnRows(rows)
	user, err := repo.GetByEmail(context.Background(), "test@test.com")
	require.NoError(t, err)
	assert.Equal(t, int64(1), user.ID)
}

func TestUserRepo_GetByEmail_NotFound(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, password, name, role, status, created_at, updated_at")).
		WithArgs("nonexistent@test.com").WillReturnError(sql.ErrNoRows)
	_, err := repo.GetByEmail(context.Background(), "nonexistent@test.com")
	assert.Error(t, err)
}

// --- GetByEmailForAuth ---

func TestUserRepo_GetByEmailForAuth_Success(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows(userCols).AddRow(int64(1), "test@test.com", "hash", "Test", domain.RoleTeacher, entities.UserStatusActive, now, now)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, password, name, role, status, created_at, updated_at")).
		WithArgs("test@test.com").WillReturnRows(rows)
	user, err := repo.GetByEmailForAuth(context.Background(), "test@test.com")
	require.NoError(t, err)
	assert.Equal(t, "hash", user.Password)
}

// --- Delete ---

func TestUserRepo_Delete_Success(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM users WHERE id = $1")).
		WithArgs(int64(1)).WillReturnResult(sqlmock.NewResult(0, 1))
	err := repo.Delete(context.Background(), 1)
	require.NoError(t, err)
}

func TestUserRepo_Delete_NotFound(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM users WHERE id = $1")).
		WithArgs(int64(999)).WillReturnResult(sqlmock.NewResult(0, 0))
	err := repo.Delete(context.Background(), 999)
	assert.Error(t, err)
}

func TestUserRepo_Delete_DBError(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM users WHERE id = $1")).
		WithArgs(int64(1)).WillReturnError(fmt.Errorf("db error"))
	err := repo.Delete(context.Background(), 1)
	assert.Error(t, err)
}

func TestUserRepo_Delete_RowsAffectedError(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM users WHERE id = $1")).
		WithArgs(int64(1)).WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("rows error")))
	err := repo.Delete(context.Background(), 1)
	assert.Error(t, err)
}

// --- List ---

func TestUserRepo_List_Success(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows(userCols).
		AddRow(int64(1), "a@test.com", "hash", "A", domain.RoleTeacher, entities.UserStatusActive, now, now).
		AddRow(int64(2), "b@test.com", "hash", "B", domain.RoleStudent, entities.UserStatusActive, now, now)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, password, name, role, status, created_at, updated_at")).
		WithArgs(10, 0).WillReturnRows(rows)
	users, err := repo.List(context.Background(), 10, 0)
	require.NoError(t, err)
	assert.Len(t, users, 2)
}

func TestUserRepo_List_Empty(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	rows := sqlmock.NewRows(userCols)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, password, name, role, status, created_at, updated_at")).
		WithArgs(10, 0).WillReturnRows(rows)
	users, err := repo.List(context.Background(), 10, 0)
	require.NoError(t, err)
	assert.Empty(t, users)
}

func TestUserRepo_List_DBError(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, password, name, role, status, created_at, updated_at")).
		WillReturnError(fmt.Errorf("db error"))
	_, err := repo.List(context.Background(), 10, 0)
	assert.Error(t, err)
}

func TestUserRepo_List_DefaultLimit(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	rows := sqlmock.NewRows(userCols)
	// limit <= 0 should be set to 10
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, password, name, role, status, created_at, updated_at")).
		WithArgs(10, 0).WillReturnRows(rows)
	_, err := repo.List(context.Background(), 0, 0)
	require.NoError(t, err)
}

func TestUserRepo_List_MaxLimit(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	rows := sqlmock.NewRows(userCols)
	// limit > 100 should be clamped to 100
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, password, name, role, status, created_at, updated_at")).
		WithArgs(100, 0).WillReturnRows(rows)
	_, err := repo.List(context.Background(), 200, 0)
	require.NoError(t, err)
}

func TestUserRepo_List_NegativeOffset(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	rows := sqlmock.NewRows(userCols)
	// negative offset should be set to 0
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, password, name, role, status, created_at, updated_at")).
		WithArgs(10, 0).WillReturnRows(rows)
	_, err := repo.List(context.Background(), 10, -5)
	require.NoError(t, err)
}

func TestUserRepo_List_ScanError(t *testing.T) {
	repo, mock := newUserRepoMock(t)
	// Return row with wrong number of columns to trigger scan error
	rows := sqlmock.NewRows([]string{"id"}).AddRow(int64(1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, email, password, name, role, status, created_at, updated_at")).
		WithArgs(10, 0).WillReturnRows(rows)
	_, err := repo.List(context.Background(), 10, 0)
	assert.Error(t, err)
}
