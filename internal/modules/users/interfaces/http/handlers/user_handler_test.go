package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	authEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/domain/repositories"
	domainErrors "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/domain/errors"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// --- Mock Auth UserRepository ---

type mockAuthUserRepo struct {
	mock.Mock
}

func (m *mockAuthUserRepo) Create(ctx context.Context, user *authEntities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockAuthUserRepo) Save(ctx context.Context, user *authEntities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockAuthUserRepo) GetByID(ctx context.Context, id int64) (*authEntities.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authEntities.User), args.Error(1)
}

func (m *mockAuthUserRepo) GetByEmail(ctx context.Context, email string) (*authEntities.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authEntities.User), args.Error(1)
}

func (m *mockAuthUserRepo) GetByEmailForAuth(ctx context.Context, email string) (*authEntities.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authEntities.User), args.Error(1)
}

func (m *mockAuthUserRepo) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockAuthUserRepo) List(ctx context.Context, limit, offset int) ([]*authEntities.User, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*authEntities.User), args.Error(1)
}

// --- Mock UserProfileRepository ---

type mockUserProfileRepo struct {
	mock.Mock
}

func (m *mockUserProfileRepo) GetProfileByID(ctx context.Context, userID int64) (*entities.UserWithOrg, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.UserWithOrg), args.Error(1)
}

func (m *mockUserProfileRepo) UpdateProfile(ctx context.Context, userID int64, departmentID, positionID *int64, phone, avatar, bio string) error {
	args := m.Called(ctx, userID, departmentID, positionID, phone, avatar, bio)
	return args.Error(0)
}

func (m *mockUserProfileRepo) ListUsersWithOrg(ctx context.Context, filter *repositories.UserFilter, limit, offset int) ([]*entities.UserWithOrg, error) {
	args := m.Called(ctx, filter, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.UserWithOrg), args.Error(1)
}

func (m *mockUserProfileRepo) CountUsers(ctx context.Context, filter *repositories.UserFilter) (int64, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockUserProfileRepo) GetUsersByDepartment(ctx context.Context, departmentID int64) ([]*entities.UserWithOrg, error) {
	args := m.Called(ctx, departmentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.UserWithOrg), args.Error(1)
}

func (m *mockUserProfileRepo) GetUsersByPosition(ctx context.Context, positionID int64) ([]*entities.UserWithOrg, error) {
	args := m.Called(ctx, positionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.UserWithOrg), args.Error(1)
}

func (m *mockUserProfileRepo) BulkUpdateDepartment(ctx context.Context, userIDs []int64, departmentID *int64) error {
	args := m.Called(ctx, userIDs, departmentID)
	return args.Error(0)
}

func (m *mockUserProfileRepo) BulkUpdatePosition(ctx context.Context, userIDs []int64, positionID *int64) error {
	args := m.Called(ctx, userIDs, positionID)
	return args.Error(0)
}

// --- Mock DepartmentRepository ---

type mockDepartmentRepo struct {
	mock.Mock
}

func (m *mockDepartmentRepo) Create(ctx context.Context, dept *entities.Department) error {
	args := m.Called(ctx, dept)
	return args.Error(0)
}

func (m *mockDepartmentRepo) GetByID(ctx context.Context, id int64) (*entities.Department, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Department), args.Error(1)
}

func (m *mockDepartmentRepo) GetByCode(ctx context.Context, code string) (*entities.Department, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Department), args.Error(1)
}

func (m *mockDepartmentRepo) Update(ctx context.Context, dept *entities.Department) error {
	args := m.Called(ctx, dept)
	return args.Error(0)
}

func (m *mockDepartmentRepo) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockDepartmentRepo) List(ctx context.Context, limit, offset int, activeOnly bool) ([]*entities.Department, error) {
	args := m.Called(ctx, limit, offset, activeOnly)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Department), args.Error(1)
}

func (m *mockDepartmentRepo) Count(ctx context.Context, activeOnly bool) (int64, error) {
	args := m.Called(ctx, activeOnly)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockDepartmentRepo) GetChildren(ctx context.Context, parentID int64) ([]*entities.Department, error) {
	args := m.Called(ctx, parentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Department), args.Error(1)
}

// --- Mock PositionRepository ---

type mockPositionRepo struct {
	mock.Mock
}

func (m *mockPositionRepo) Create(ctx context.Context, pos *entities.Position) error {
	args := m.Called(ctx, pos)
	return args.Error(0)
}

func (m *mockPositionRepo) GetByID(ctx context.Context, id int64) (*entities.Position, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Position), args.Error(1)
}

func (m *mockPositionRepo) GetByCode(ctx context.Context, code string) (*entities.Position, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Position), args.Error(1)
}

func (m *mockPositionRepo) Update(ctx context.Context, pos *entities.Position) error {
	args := m.Called(ctx, pos)
	return args.Error(0)
}

func (m *mockPositionRepo) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockPositionRepo) List(ctx context.Context, limit, offset int, activeOnly bool) ([]*entities.Position, error) {
	args := m.Called(ctx, limit, offset, activeOnly)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Position), args.Error(1)
}

func (m *mockPositionRepo) Count(ctx context.Context, activeOnly bool) (int64, error) {
	args := m.Called(ctx, activeOnly)
	return args.Get(0).(int64), args.Error(1)
}

// --- Helpers ---

func newUserUseCase(authRepo *mockAuthUserRepo, profileRepo *mockUserProfileRepo, deptRepo *mockDepartmentRepo, posRepo *mockPositionRepo) *usecases.UserUseCase {
	return usecases.NewUserUseCase(authRepo, profileRepo, deptRepo, posRepo, nil, nil)
}

func setupUserRouter(handler *UserHandler) *gin.Engine {
	r := gin.New()
	r.GET("/users", handler.List)
	r.GET("/users/:id", handler.GetByID)
	r.PUT("/users/:id/profile", handler.UpdateProfile)
	r.PUT("/users/:id/role", handler.UpdateRole)
	r.PUT("/users/:id/status", handler.UpdateStatus)
	r.DELETE("/users/:id", handler.Delete)
	r.POST("/users/bulk/department", handler.BulkUpdateDepartment)
	r.POST("/users/bulk/position", handler.BulkUpdatePosition)
	r.GET("/users/by-department/:id", handler.GetByDepartment)
	r.GET("/users/by-position/:id", handler.GetByPosition)
	return r
}

func jsonBody(t *testing.T, v interface{}) *bytes.Reader {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return bytes.NewReader(b)
}

// --- User Handler Tests ---

func TestUserHandler_List_Success(t *testing.T) {
	profileRepo := new(mockUserProfileRepo)
	authRepo := new(mockAuthUserRepo)
	deptRepo := new(mockDepartmentRepo)
	posRepo := new(mockPositionRepo)

	users := []*entities.UserWithOrg{{ID: 1, Email: "a@b.com", Name: "Test"}}
	profileRepo.On("ListUsersWithOrg", mock.Anything, mock.Anything, 10, 0).Return(users, nil)
	profileRepo.On("CountUsers", mock.Anything, mock.Anything).Return(int64(1), nil)

	uc := newUserUseCase(authRepo, profileRepo, deptRepo, posRepo)
	handler := NewUserHandler(uc)
	router := setupUserRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
}

func TestUserHandler_List_WithFilters(t *testing.T) {
	profileRepo := new(mockUserProfileRepo)
	authRepo := new(mockAuthUserRepo)
	deptRepo := new(mockDepartmentRepo)
	posRepo := new(mockPositionRepo)

	users := []*entities.UserWithOrg{}
	profileRepo.On("ListUsersWithOrg", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(users, nil)
	profileRepo.On("CountUsers", mock.Anything, mock.Anything).Return(int64(0), nil)

	uc := newUserUseCase(authRepo, profileRepo, deptRepo, posRepo)
	handler := NewUserHandler(uc)
	router := setupUserRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/users?search=john&role=student&page=2&limit=5", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserHandler_List_UsecaseError(t *testing.T) {
	profileRepo := new(mockUserProfileRepo)
	authRepo := new(mockAuthUserRepo)
	deptRepo := new(mockDepartmentRepo)
	posRepo := new(mockPositionRepo)

	profileRepo.On("ListUsersWithOrg", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("db error"))

	uc := newUserUseCase(authRepo, profileRepo, deptRepo, posRepo)
	handler := NewUserHandler(uc)
	router := setupUserRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUserHandler_GetByID_Success(t *testing.T) {
	profileRepo := new(mockUserProfileRepo)
	authRepo := new(mockAuthUserRepo)
	deptRepo := new(mockDepartmentRepo)
	posRepo := new(mockPositionRepo)

	user := &entities.UserWithOrg{ID: 1, Email: "a@b.com", Name: "Test"}
	profileRepo.On("GetProfileByID", mock.Anything, int64(1)).Return(user, nil)

	uc := newUserUseCase(authRepo, profileRepo, deptRepo, posRepo)
	handler := NewUserHandler(uc)
	router := setupUserRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserHandler_GetByID_InvalidID(t *testing.T) {
	uc := newUserUseCase(new(mockAuthUserRepo), new(mockUserProfileRepo), new(mockDepartmentRepo), new(mockPositionRepo))
	handler := NewUserHandler(uc)
	router := setupUserRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/users/invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_GetByID_NotFound(t *testing.T) {
	profileRepo := new(mockUserProfileRepo)
	profileRepo.On("GetProfileByID", mock.Anything, int64(999)).Return(nil, domainErrors.ErrNotFound)

	uc := newUserUseCase(new(mockAuthUserRepo), profileRepo, new(mockDepartmentRepo), new(mockPositionRepo))
	handler := NewUserHandler(uc)
	router := setupUserRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/users/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUserHandler_UpdateProfile_Success(t *testing.T) {
	profileRepo := new(mockUserProfileRepo)
	authRepo := new(mockAuthUserRepo)
	deptRepo := new(mockDepartmentRepo)
	posRepo := new(mockPositionRepo)

	authRepo.On("GetByID", mock.Anything, int64(1)).Return(&authEntities.User{ID: 1}, nil)
	profileRepo.On("UpdateProfile", mock.Anything, int64(1), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	uc := newUserUseCase(authRepo, profileRepo, deptRepo, posRepo)
	handler := NewUserHandler(uc)
	router := setupUserRouter(handler)

	body := map[string]interface{}{
		"phone": "+79991234567",
		"bio":   "Hello world",
	}
	req := httptest.NewRequest(http.MethodPut, "/users/1/profile", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserHandler_UpdateProfile_InvalidID(t *testing.T) {
	uc := newUserUseCase(new(mockAuthUserRepo), new(mockUserProfileRepo), new(mockDepartmentRepo), new(mockPositionRepo))
	handler := NewUserHandler(uc)
	router := setupUserRouter(handler)

	req := httptest.NewRequest(http.MethodPut, "/users/abc/profile", jsonBody(t, map[string]interface{}{}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_UpdateProfile_InvalidJSON(t *testing.T) {
	uc := newUserUseCase(new(mockAuthUserRepo), new(mockUserProfileRepo), new(mockDepartmentRepo), new(mockPositionRepo))
	handler := NewUserHandler(uc)
	router := setupUserRouter(handler)

	req := httptest.NewRequest(http.MethodPut, "/users/1/profile", bytes.NewReader([]byte("not-json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_UpdateProfile_ValidationError(t *testing.T) {
	uc := newUserUseCase(new(mockAuthUserRepo), new(mockUserProfileRepo), new(mockDepartmentRepo), new(mockPositionRepo))
	handler := NewUserHandler(uc)
	router := setupUserRouter(handler)

	body := map[string]interface{}{
		"phone": "not-a-phone",
	}
	req := httptest.NewRequest(http.MethodPut, "/users/1/profile", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_UpdateProfile_UsecaseError(t *testing.T) {
	profileRepo := new(mockUserProfileRepo)
	authRepo := new(mockAuthUserRepo)

	authRepo.On("GetByID", mock.Anything, int64(1)).Return(nil, domainErrors.ErrNotFound)

	uc := newUserUseCase(authRepo, profileRepo, new(mockDepartmentRepo), new(mockPositionRepo))
	handler := NewUserHandler(uc)
	router := setupUserRouter(handler)

	body := map[string]interface{}{
		"bio": "test",
	}
	req := httptest.NewRequest(http.MethodPut, "/users/1/profile", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUserHandler_UpdateRole_Success(t *testing.T) {
	authRepo := new(mockAuthUserRepo)
	authRepo.On("GetByID", mock.Anything, int64(1)).Return(&authEntities.User{ID: 1, Role: "student"}, nil)
	authRepo.On("Save", mock.Anything, mock.Anything).Return(nil)

	uc := newUserUseCase(authRepo, new(mockUserProfileRepo), new(mockDepartmentRepo), new(mockPositionRepo))
	handler := NewUserHandler(uc)
	router := setupUserRouter(handler)

	body := map[string]interface{}{"role": "teacher"}
	req := httptest.NewRequest(http.MethodPut, "/users/1/role", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserHandler_UpdateRole_InvalidID(t *testing.T) {
	uc := newUserUseCase(new(mockAuthUserRepo), new(mockUserProfileRepo), new(mockDepartmentRepo), new(mockPositionRepo))
	handler := NewUserHandler(uc)
	router := setupUserRouter(handler)

	body := map[string]interface{}{"role": "teacher"}
	req := httptest.NewRequest(http.MethodPut, "/users/abc/role", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_UpdateRole_InvalidJSON(t *testing.T) {
	uc := newUserUseCase(new(mockAuthUserRepo), new(mockUserProfileRepo), new(mockDepartmentRepo), new(mockPositionRepo))
	handler := NewUserHandler(uc)
	router := setupUserRouter(handler)

	req := httptest.NewRequest(http.MethodPut, "/users/1/role", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_UpdateRole_ValidationError(t *testing.T) {
	uc := newUserUseCase(new(mockAuthUserRepo), new(mockUserProfileRepo), new(mockDepartmentRepo), new(mockPositionRepo))
	handler := NewUserHandler(uc)
	router := setupUserRouter(handler)

	body := map[string]interface{}{"role": "invalid_role"}
	req := httptest.NewRequest(http.MethodPut, "/users/1/role", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_UpdateRole_UsecaseError(t *testing.T) {
	authRepo := new(mockAuthUserRepo)
	authRepo.On("GetByID", mock.Anything, int64(1)).Return(nil, domainErrors.ErrNotFound)

	uc := newUserUseCase(authRepo, new(mockUserProfileRepo), new(mockDepartmentRepo), new(mockPositionRepo))
	handler := NewUserHandler(uc)
	router := setupUserRouter(handler)

	body := map[string]interface{}{"role": "teacher"}
	req := httptest.NewRequest(http.MethodPut, "/users/1/role", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUserHandler_UpdateStatus_Success(t *testing.T) {
	authRepo := new(mockAuthUserRepo)
	authRepo.On("GetByID", mock.Anything, int64(1)).Return(&authEntities.User{ID: 1, Status: "active"}, nil)
	authRepo.On("Save", mock.Anything, mock.Anything).Return(nil)

	uc := newUserUseCase(authRepo, new(mockUserProfileRepo), new(mockDepartmentRepo), new(mockPositionRepo))
	handler := NewUserHandler(uc)
	router := setupUserRouter(handler)

	body := map[string]interface{}{"status": "inactive"}
	req := httptest.NewRequest(http.MethodPut, "/users/1/status", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserHandler_UpdateStatus_InvalidID(t *testing.T) {
	uc := newUserUseCase(new(mockAuthUserRepo), new(mockUserProfileRepo), new(mockDepartmentRepo), new(mockPositionRepo))
	handler := NewUserHandler(uc)
	router := setupUserRouter(handler)

	body := map[string]interface{}{"status": "active"}
	req := httptest.NewRequest(http.MethodPut, "/users/xyz/status", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_UpdateStatus_InvalidJSON(t *testing.T) {
	uc := newUserUseCase(new(mockAuthUserRepo), new(mockUserProfileRepo), new(mockDepartmentRepo), new(mockPositionRepo))
	handler := NewUserHandler(uc)
	router := setupUserRouter(handler)

	req := httptest.NewRequest(http.MethodPut, "/users/1/status", bytes.NewReader([]byte("{bad")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_UpdateStatus_ValidationError(t *testing.T) {
	uc := newUserUseCase(new(mockAuthUserRepo), new(mockUserProfileRepo), new(mockDepartmentRepo), new(mockPositionRepo))
	handler := NewUserHandler(uc)
	router := setupUserRouter(handler)

	body := map[string]interface{}{"status": "unknown"}
	req := httptest.NewRequest(http.MethodPut, "/users/1/status", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_UpdateStatus_UsecaseError(t *testing.T) {
	authRepo := new(mockAuthUserRepo)
	authRepo.On("GetByID", mock.Anything, int64(1)).Return(nil, domainErrors.ErrNotFound)

	uc := newUserUseCase(authRepo, new(mockUserProfileRepo), new(mockDepartmentRepo), new(mockPositionRepo))
	handler := NewUserHandler(uc)
	router := setupUserRouter(handler)

	body := map[string]interface{}{"status": "active"}
	req := httptest.NewRequest(http.MethodPut, "/users/1/status", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUserHandler_Delete_Success(t *testing.T) {
	authRepo := new(mockAuthUserRepo)
	authRepo.On("GetByID", mock.Anything, int64(1)).Return(&authEntities.User{ID: 1}, nil)
	authRepo.On("Delete", mock.Anything, int64(1)).Return(nil)

	uc := newUserUseCase(authRepo, new(mockUserProfileRepo), new(mockDepartmentRepo), new(mockPositionRepo))
	handler := NewUserHandler(uc)
	router := setupUserRouter(handler)

	req := httptest.NewRequest(http.MethodDelete, "/users/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserHandler_Delete_InvalidID(t *testing.T) {
	uc := newUserUseCase(new(mockAuthUserRepo), new(mockUserProfileRepo), new(mockDepartmentRepo), new(mockPositionRepo))
	handler := NewUserHandler(uc)
	router := setupUserRouter(handler)

	req := httptest.NewRequest(http.MethodDelete, "/users/abc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_Delete_NotFound(t *testing.T) {
	authRepo := new(mockAuthUserRepo)
	authRepo.On("GetByID", mock.Anything, int64(999)).Return(nil, domainErrors.ErrNotFound)

	uc := newUserUseCase(authRepo, new(mockUserProfileRepo), new(mockDepartmentRepo), new(mockPositionRepo))
	handler := NewUserHandler(uc)
	router := setupUserRouter(handler)

	req := httptest.NewRequest(http.MethodDelete, "/users/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUserHandler_BulkUpdateDepartment_Success(t *testing.T) {
	profileRepo := new(mockUserProfileRepo)
	deptRepo := new(mockDepartmentRepo)
	deptID := int64(1)

	deptRepo.On("GetByID", mock.Anything, int64(1)).Return(&entities.Department{ID: 1}, nil)
	profileRepo.On("BulkUpdateDepartment", mock.Anything, []int64{1, 2, 3}, &deptID).Return(nil)

	uc := newUserUseCase(new(mockAuthUserRepo), profileRepo, deptRepo, new(mockPositionRepo))
	handler := NewUserHandler(uc)
	router := setupUserRouter(handler)

	body := map[string]interface{}{
		"user_ids":      []int64{1, 2, 3},
		"department_id": 1,
	}
	req := httptest.NewRequest(http.MethodPost, "/users/bulk/department", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserHandler_BulkUpdateDepartment_InvalidJSON(t *testing.T) {
	uc := newUserUseCase(new(mockAuthUserRepo), new(mockUserProfileRepo), new(mockDepartmentRepo), new(mockPositionRepo))
	handler := NewUserHandler(uc)
	router := setupUserRouter(handler)

	req := httptest.NewRequest(http.MethodPost, "/users/bulk/department", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_BulkUpdateDepartment_ValidationError(t *testing.T) {
	uc := newUserUseCase(new(mockAuthUserRepo), new(mockUserProfileRepo), new(mockDepartmentRepo), new(mockPositionRepo))
	handler := NewUserHandler(uc)
	router := setupUserRouter(handler)

	body := map[string]interface{}{
		"user_ids": []int64{},
	}
	req := httptest.NewRequest(http.MethodPost, "/users/bulk/department", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_BulkUpdateDepartment_UsecaseError(t *testing.T) {
	profileRepo := new(mockUserProfileRepo)
	profileRepo.On("BulkUpdateDepartment", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("db error"))

	uc := newUserUseCase(new(mockAuthUserRepo), profileRepo, new(mockDepartmentRepo), new(mockPositionRepo))
	handler := NewUserHandler(uc)
	router := setupUserRouter(handler)

	body := map[string]interface{}{
		"user_ids": []int64{1, 2},
	}
	req := httptest.NewRequest(http.MethodPost, "/users/bulk/department", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUserHandler_BulkUpdatePosition_Success(t *testing.T) {
	profileRepo := new(mockUserProfileRepo)
	posRepo := new(mockPositionRepo)
	posID := int64(5)

	posRepo.On("GetByID", mock.Anything, int64(5)).Return(&entities.Position{ID: 5}, nil)
	profileRepo.On("BulkUpdatePosition", mock.Anything, []int64{1, 2}, &posID).Return(nil)

	uc := newUserUseCase(new(mockAuthUserRepo), profileRepo, new(mockDepartmentRepo), posRepo)
	handler := NewUserHandler(uc)
	router := setupUserRouter(handler)

	body := map[string]interface{}{
		"user_ids":    []int64{1, 2},
		"position_id": 5,
	}
	req := httptest.NewRequest(http.MethodPost, "/users/bulk/position", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserHandler_BulkUpdatePosition_InvalidJSON(t *testing.T) {
	uc := newUserUseCase(new(mockAuthUserRepo), new(mockUserProfileRepo), new(mockDepartmentRepo), new(mockPositionRepo))
	handler := NewUserHandler(uc)
	router := setupUserRouter(handler)

	req := httptest.NewRequest(http.MethodPost, "/users/bulk/position", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_BulkUpdatePosition_ValidationError(t *testing.T) {
	uc := newUserUseCase(new(mockAuthUserRepo), new(mockUserProfileRepo), new(mockDepartmentRepo), new(mockPositionRepo))
	handler := NewUserHandler(uc)
	router := setupUserRouter(handler)

	body := map[string]interface{}{
		"user_ids": []int64{},
	}
	req := httptest.NewRequest(http.MethodPost, "/users/bulk/position", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_BulkUpdatePosition_UsecaseError(t *testing.T) {
	profileRepo := new(mockUserProfileRepo)
	profileRepo.On("BulkUpdatePosition", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("db error"))

	uc := newUserUseCase(new(mockAuthUserRepo), profileRepo, new(mockDepartmentRepo), new(mockPositionRepo))
	handler := NewUserHandler(uc)
	router := setupUserRouter(handler)

	body := map[string]interface{}{
		"user_ids": []int64{1},
	}
	req := httptest.NewRequest(http.MethodPost, "/users/bulk/position", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUserHandler_GetByDepartment_Success(t *testing.T) {
	profileRepo := new(mockUserProfileRepo)
	users := []*entities.UserWithOrg{{ID: 1, DepartmentName: "IT"}}
	profileRepo.On("GetUsersByDepartment", mock.Anything, int64(1)).Return(users, nil)

	uc := newUserUseCase(new(mockAuthUserRepo), profileRepo, new(mockDepartmentRepo), new(mockPositionRepo))
	handler := NewUserHandler(uc)
	router := setupUserRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/users/by-department/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserHandler_GetByDepartment_InvalidID(t *testing.T) {
	uc := newUserUseCase(new(mockAuthUserRepo), new(mockUserProfileRepo), new(mockDepartmentRepo), new(mockPositionRepo))
	handler := NewUserHandler(uc)
	router := setupUserRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/users/by-department/abc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_GetByDepartment_UsecaseError(t *testing.T) {
	profileRepo := new(mockUserProfileRepo)
	profileRepo.On("GetUsersByDepartment", mock.Anything, int64(1)).Return(nil, errors.New("db error"))

	uc := newUserUseCase(new(mockAuthUserRepo), profileRepo, new(mockDepartmentRepo), new(mockPositionRepo))
	handler := NewUserHandler(uc)
	router := setupUserRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/users/by-department/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUserHandler_GetByPosition_Success(t *testing.T) {
	profileRepo := new(mockUserProfileRepo)
	users := []*entities.UserWithOrg{{ID: 1, PositionName: "Dev"}}
	profileRepo.On("GetUsersByPosition", mock.Anything, int64(2)).Return(users, nil)

	uc := newUserUseCase(new(mockAuthUserRepo), profileRepo, new(mockDepartmentRepo), new(mockPositionRepo))
	handler := NewUserHandler(uc)
	router := setupUserRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/users/by-position/2", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserHandler_GetByPosition_InvalidID(t *testing.T) {
	uc := newUserUseCase(new(mockAuthUserRepo), new(mockUserProfileRepo), new(mockDepartmentRepo), new(mockPositionRepo))
	handler := NewUserHandler(uc)
	router := setupUserRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/users/by-position/abc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_GetByPosition_UsecaseError(t *testing.T) {
	profileRepo := new(mockUserProfileRepo)
	profileRepo.On("GetUsersByPosition", mock.Anything, int64(2)).Return(nil, errors.New("db error"))

	uc := newUserUseCase(new(mockAuthUserRepo), profileRepo, new(mockDepartmentRepo), new(mockPositionRepo))
	handler := NewUserHandler(uc)
	router := setupUserRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/users/by-position/2", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestNewUserHandler(t *testing.T) {
	uc := newUserUseCase(new(mockAuthUserRepo), new(mockUserProfileRepo), new(mockDepartmentRepo), new(mockPositionRepo))
	handler := NewUserHandler(uc)
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.usecase)
	assert.NotNil(t, handler.validator)
	assert.NotNil(t, handler.sanitizer)
}
