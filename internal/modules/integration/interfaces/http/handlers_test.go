package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/domain/entities"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupRouter() *gin.Engine {
	return gin.New()
}

func withAuth(userID int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	}
}

func performRequest(router *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	var req *http.Request
	if body != nil {
		jsonBytes, _ := json.Marshal(body)
		req = httptest.NewRequest(method, path, bytes.NewReader(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// ===== Mock Repositories =====

type mockSyncConflictRepo struct{ mock.Mock }

func (m *mockSyncConflictRepo) Create(ctx context.Context, conflict *entities.SyncConflict) error {
	return m.Called(ctx, conflict).Error(0)
}
func (m *mockSyncConflictRepo) GetByID(ctx context.Context, id int64) (*entities.SyncConflict, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.SyncConflict), args.Error(1)
}
func (m *mockSyncConflictRepo) List(ctx context.Context, filter entities.SyncConflictFilter) ([]*entities.SyncConflict, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*entities.SyncConflict), args.Get(1).(int64), args.Error(2)
}
func (m *mockSyncConflictRepo) Update(ctx context.Context, conflict *entities.SyncConflict) error {
	return m.Called(ctx, conflict).Error(0)
}
func (m *mockSyncConflictRepo) Delete(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockSyncConflictRepo) GetPending(ctx context.Context, limit, offset int) ([]*entities.SyncConflict, int64, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*entities.SyncConflict), args.Get(1).(int64), args.Error(2)
}
func (m *mockSyncConflictRepo) GetBySyncLogID(ctx context.Context, syncLogID int64) ([]*entities.SyncConflict, error) {
	args := m.Called(ctx, syncLogID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.SyncConflict), args.Error(1)
}
func (m *mockSyncConflictRepo) GetStats(ctx context.Context) (*entities.ConflictStats, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.ConflictStats), args.Error(1)
}
func (m *mockSyncConflictRepo) BulkResolve(ctx context.Context, ids []int64, resolution entities.ConflictResolution, resolvedBy int64) error {
	return m.Called(ctx, ids, resolution, resolvedBy).Error(0)
}
func (m *mockSyncConflictRepo) DeleteBySyncLogID(ctx context.Context, syncLogID int64) error {
	return m.Called(ctx, syncLogID).Error(0)
}
func (m *mockSyncConflictRepo) GetPendingByEntityType(ctx context.Context, entityType entities.SyncEntityType) ([]*entities.SyncConflict, error) {
	args := m.Called(ctx, entityType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.SyncConflict), args.Error(1)
}
func (m *mockSyncConflictRepo) Resolve(ctx context.Context, id int64, resolution entities.ConflictResolution, userID int64, resolvedData string) error {
	return m.Called(ctx, id, resolution, userID, resolvedData).Error(0)
}

type mockEmployeeRepo struct{ mock.Mock }

func (m *mockEmployeeRepo) Create(ctx context.Context, emp *entities.ExternalEmployee) error {
	return m.Called(ctx, emp).Error(0)
}
func (m *mockEmployeeRepo) GetByID(ctx context.Context, id int64) (*entities.ExternalEmployee, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.ExternalEmployee), args.Error(1)
}
func (m *mockEmployeeRepo) GetByExternalID(ctx context.Context, externalID string) (*entities.ExternalEmployee, error) {
	args := m.Called(ctx, externalID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.ExternalEmployee), args.Error(1)
}
func (m *mockEmployeeRepo) Update(ctx context.Context, emp *entities.ExternalEmployee) error {
	return m.Called(ctx, emp).Error(0)
}
func (m *mockEmployeeRepo) Delete(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockEmployeeRepo) List(ctx context.Context, filter entities.ExternalEmployeeFilter) ([]*entities.ExternalEmployee, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*entities.ExternalEmployee), args.Get(1).(int64), args.Error(2)
}
func (m *mockEmployeeRepo) GetUnlinked(ctx context.Context, limit, offset int) ([]*entities.ExternalEmployee, int64, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*entities.ExternalEmployee), args.Get(1).(int64), args.Error(2)
}
func (m *mockEmployeeRepo) LinkToLocalUser(ctx context.Context, id, localUserID int64) error {
	return m.Called(ctx, id, localUserID).Error(0)
}
func (m *mockEmployeeRepo) Unlink(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockEmployeeRepo) GetByLocalUserID(ctx context.Context, localUserID int64) (*entities.ExternalEmployee, error) {
	args := m.Called(ctx, localUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.ExternalEmployee), args.Error(1)
}
func (m *mockEmployeeRepo) CreateOrUpdate(ctx context.Context, emp *entities.ExternalEmployee) error {
	return m.Called(ctx, emp).Error(0)
}
func (m *mockEmployeeRepo) Upsert(ctx context.Context, emp *entities.ExternalEmployee) error {
	return m.Called(ctx, emp).Error(0)
}
func (m *mockEmployeeRepo) GetByCode(ctx context.Context, code string) (*entities.ExternalEmployee, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.ExternalEmployee), args.Error(1)
}
func (m *mockEmployeeRepo) GetAllExternalIDs(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}
func (m *mockEmployeeRepo) BulkUpsert(ctx context.Context, employees []*entities.ExternalEmployee) error {
	return m.Called(ctx, employees).Error(0)
}
func (m *mockEmployeeRepo) MarkInactiveExcept(ctx context.Context, activeExternalIDs []string) error {
	return m.Called(ctx, activeExternalIDs).Error(0)
}

type mockStudentRepo struct{ mock.Mock }

func (m *mockStudentRepo) Create(ctx context.Context, s *entities.ExternalStudent) error {
	return m.Called(ctx, s).Error(0)
}
func (m *mockStudentRepo) GetByID(ctx context.Context, id int64) (*entities.ExternalStudent, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.ExternalStudent), args.Error(1)
}
func (m *mockStudentRepo) GetByExternalID(ctx context.Context, externalID string) (*entities.ExternalStudent, error) {
	args := m.Called(ctx, externalID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.ExternalStudent), args.Error(1)
}
func (m *mockStudentRepo) Update(ctx context.Context, s *entities.ExternalStudent) error {
	return m.Called(ctx, s).Error(0)
}
func (m *mockStudentRepo) Delete(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockStudentRepo) List(ctx context.Context, filter entities.ExternalStudentFilter) ([]*entities.ExternalStudent, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*entities.ExternalStudent), args.Get(1).(int64), args.Error(2)
}
func (m *mockStudentRepo) GetUnlinked(ctx context.Context, limit, offset int) ([]*entities.ExternalStudent, int64, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*entities.ExternalStudent), args.Get(1).(int64), args.Error(2)
}
func (m *mockStudentRepo) LinkToLocalUser(ctx context.Context, id, localUserID int64) error {
	return m.Called(ctx, id, localUserID).Error(0)
}
func (m *mockStudentRepo) Unlink(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockStudentRepo) GetByLocalUserID(ctx context.Context, localUserID int64) (*entities.ExternalStudent, error) {
	args := m.Called(ctx, localUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.ExternalStudent), args.Error(1)
}
func (m *mockStudentRepo) CreateOrUpdate(ctx context.Context, s *entities.ExternalStudent) error {
	return m.Called(ctx, s).Error(0)
}
func (m *mockStudentRepo) Upsert(ctx context.Context, s *entities.ExternalStudent) error {
	return m.Called(ctx, s).Error(0)
}
func (m *mockStudentRepo) GetByCode(ctx context.Context, code string) (*entities.ExternalStudent, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.ExternalStudent), args.Error(1)
}
func (m *mockStudentRepo) GetByGroup(ctx context.Context, groupName string) ([]*entities.ExternalStudent, error) {
	args := m.Called(ctx, groupName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.ExternalStudent), args.Error(1)
}
func (m *mockStudentRepo) GetByFaculty(ctx context.Context, faculty string) ([]*entities.ExternalStudent, error) {
	args := m.Called(ctx, faculty)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.ExternalStudent), args.Error(1)
}
func (m *mockStudentRepo) GetAllExternalIDs(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}
func (m *mockStudentRepo) BulkUpsert(ctx context.Context, students []*entities.ExternalStudent) error {
	return m.Called(ctx, students).Error(0)
}
func (m *mockStudentRepo) MarkInactiveExcept(ctx context.Context, activeExternalIDs []string) error {
	return m.Called(ctx, activeExternalIDs).Error(0)
}
func (m *mockStudentRepo) GetGroups(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}
func (m *mockStudentRepo) GetFaculties(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

// ===== Conflict Handler Tests =====

func TestNewConflictHandler(t *testing.T) {
	h := NewConflictHandler(nil)
	assert.NotNil(t, h)
}

func TestConflictHandler_RegisterRoutes(t *testing.T) {
	repo := new(mockSyncConflictRepo)
	uc := usecases.NewConflictUseCase(repo)
	h := NewConflictHandler(uc)
	router := setupRouter()
	h.RegisterRoutes(router.Group("/api"))
}

func TestConflictHandler_List_Success(t *testing.T) {
	repo := new(mockSyncConflictRepo)
	uc := usecases.NewConflictUseCase(repo)
	h := NewConflictHandler(uc)

	repo.On("List", mock.Anything, mock.Anything).Return([]*entities.SyncConflict{}, int64(0), nil)

	router := setupRouter()
	router.GET("/conflicts", h.List)
	w := performRequest(router, http.MethodGet, "/conflicts", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestConflictHandler_List_WithFilters(t *testing.T) {
	repo := new(mockSyncConflictRepo)
	uc := usecases.NewConflictUseCase(repo)
	h := NewConflictHandler(uc)

	repo.On("List", mock.Anything, mock.Anything).Return([]*entities.SyncConflict{}, int64(0), nil)

	router := setupRouter()
	router.GET("/conflicts", h.List)
	w := performRequest(router, http.MethodGet, "/conflicts?sync_log_id=1&entity_type=employee&resolution=pending&limit=10&offset=5", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestConflictHandler_List_Error(t *testing.T) {
	repo := new(mockSyncConflictRepo)
	uc := usecases.NewConflictUseCase(repo)
	h := NewConflictHandler(uc)

	repo.On("List", mock.Anything, mock.Anything).Return(nil, int64(0), fmt.Errorf("db error"))

	router := setupRouter()
	router.GET("/conflicts", h.List)
	w := performRequest(router, http.MethodGet, "/conflicts", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestConflictHandler_GetByID_Success(t *testing.T) {
	repo := new(mockSyncConflictRepo)
	uc := usecases.NewConflictUseCase(repo)
	h := NewConflictHandler(uc)

	conflict := &entities.SyncConflict{ID: 1, EntityType: entities.SyncEntityEmployee}
	repo.On("GetByID", mock.Anything, int64(1)).Return(conflict, nil)

	router := setupRouter()
	router.GET("/conflicts/:id", h.GetByID)
	w := performRequest(router, http.MethodGet, "/conflicts/1", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestConflictHandler_GetByID_InvalidID(t *testing.T) {
	repo := new(mockSyncConflictRepo)
	uc := usecases.NewConflictUseCase(repo)
	h := NewConflictHandler(uc)

	router := setupRouter()
	router.GET("/conflicts/:id", h.GetByID)
	w := performRequest(router, http.MethodGet, "/conflicts/abc", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestConflictHandler_GetByID_NotFound(t *testing.T) {
	repo := new(mockSyncConflictRepo)
	uc := usecases.NewConflictUseCase(repo)
	h := NewConflictHandler(uc)

	repo.On("GetByID", mock.Anything, int64(1)).Return(nil, nil)

	router := setupRouter()
	router.GET("/conflicts/:id", h.GetByID)
	w := performRequest(router, http.MethodGet, "/conflicts/1", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestConflictHandler_GetByID_Error(t *testing.T) {
	repo := new(mockSyncConflictRepo)
	uc := usecases.NewConflictUseCase(repo)
	h := NewConflictHandler(uc)

	repo.On("GetByID", mock.Anything, int64(1)).Return(nil, fmt.Errorf("db error"))

	router := setupRouter()
	router.GET("/conflicts/:id", h.GetByID)
	w := performRequest(router, http.MethodGet, "/conflicts/1", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestConflictHandler_GetPending_Success(t *testing.T) {
	repo := new(mockSyncConflictRepo)
	uc := usecases.NewConflictUseCase(repo)
	h := NewConflictHandler(uc)

	repo.On("GetPending", mock.Anything, 20, 0).Return([]*entities.SyncConflict{}, int64(0), nil)

	router := setupRouter()
	router.GET("/conflicts/pending", h.GetPending)
	w := performRequest(router, http.MethodGet, "/conflicts/pending", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestConflictHandler_GetPending_Error(t *testing.T) {
	repo := new(mockSyncConflictRepo)
	uc := usecases.NewConflictUseCase(repo)
	h := NewConflictHandler(uc)

	repo.On("GetPending", mock.Anything, 20, 0).Return(nil, int64(0), fmt.Errorf("err"))

	router := setupRouter()
	router.GET("/conflicts/pending", h.GetPending)
	w := performRequest(router, http.MethodGet, "/conflicts/pending", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestConflictHandler_GetStats_Success(t *testing.T) {
	repo := new(mockSyncConflictRepo)
	uc := usecases.NewConflictUseCase(repo)
	h := NewConflictHandler(uc)

	stats := &entities.ConflictStats{TotalConflicts: 5}
	repo.On("GetStats", mock.Anything).Return(stats, nil)

	router := setupRouter()
	router.GET("/conflicts/stats", h.GetStats)
	w := performRequest(router, http.MethodGet, "/conflicts/stats", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestConflictHandler_GetStats_Error(t *testing.T) {
	repo := new(mockSyncConflictRepo)
	uc := usecases.NewConflictUseCase(repo)
	h := NewConflictHandler(uc)

	repo.On("GetStats", mock.Anything).Return(nil, fmt.Errorf("err"))

	router := setupRouter()
	router.GET("/conflicts/stats", h.GetStats)
	w := performRequest(router, http.MethodGet, "/conflicts/stats", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestConflictHandler_Resolve_Success(t *testing.T) {
	repo := new(mockSyncConflictRepo)
	uc := usecases.NewConflictUseCase(repo)
	h := NewConflictHandler(uc)

	conflict := &entities.SyncConflict{ID: 1, Resolution: entities.ConflictResolutionPending}
	repo.On("GetByID", mock.Anything, int64(1)).Return(conflict, nil)
	repo.On("Resolve", mock.Anything, int64(1), entities.ConflictResolutionUseLocal, int64(1), "").Return(nil)

	router := setupRouter()
	router.POST("/conflicts/:id/resolve", withAuth(1), h.Resolve)
	w := performRequest(router, http.MethodPost, "/conflicts/1/resolve", map[string]interface{}{
		"resolution": "use_local",
	})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestConflictHandler_Resolve_InvalidID(t *testing.T) {
	h := NewConflictHandler(nil)
	router := setupRouter()
	router.POST("/conflicts/:id/resolve", h.Resolve)
	w := performRequest(router, http.MethodPost, "/conflicts/abc/resolve", map[string]interface{}{"resolution": "use_local"})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestConflictHandler_Resolve_InvalidJSON(t *testing.T) {
	h := NewConflictHandler(nil)
	router := setupRouter()
	router.POST("/conflicts/:id/resolve", h.Resolve)
	req := httptest.NewRequest(http.MethodPost, "/conflicts/1/resolve", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestConflictHandler_Resolve_InvalidResolution(t *testing.T) {
	h := NewConflictHandler(nil)
	router := setupRouter()
	router.POST("/conflicts/:id/resolve", h.Resolve)
	w := performRequest(router, http.MethodPost, "/conflicts/1/resolve", map[string]interface{}{
		"resolution": "invalid",
	})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestConflictHandler_Resolve_Error(t *testing.T) {
	repo := new(mockSyncConflictRepo)
	uc := usecases.NewConflictUseCase(repo)
	h := NewConflictHandler(uc)

	repo.On("GetByID", mock.Anything, int64(1)).Return(nil, fmt.Errorf("err"))

	router := setupRouter()
	router.POST("/conflicts/:id/resolve", withAuth(1), h.Resolve)
	w := performRequest(router, http.MethodPost, "/conflicts/1/resolve", map[string]interface{}{"resolution": "use_local"})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestConflictHandler_Resolve_AllValidResolutions(t *testing.T) {
	for _, res := range []entities.ConflictResolution{
		entities.ConflictResolutionUseLocal,
		entities.ConflictResolutionUseExternal,
		entities.ConflictResolutionMerge,
		entities.ConflictResolutionSkip,
	} {
		t.Run(string(res), func(t *testing.T) {
			repo := new(mockSyncConflictRepo)
			uc := usecases.NewConflictUseCase(repo)
			h := NewConflictHandler(uc)

			conflict := &entities.SyncConflict{ID: 1, Resolution: entities.ConflictResolutionPending}
			repo.On("GetByID", mock.Anything, int64(1)).Return(conflict, nil)
			repo.On("Resolve", mock.Anything, int64(1), mock.Anything, int64(1), "").Return(nil)

			router := setupRouter()
			router.POST("/conflicts/:id/resolve", withAuth(1), h.Resolve)
			w := performRequest(router, http.MethodPost, "/conflicts/1/resolve", map[string]interface{}{"resolution": string(res)})
			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestConflictHandler_BulkResolve_Success(t *testing.T) {
	repo := new(mockSyncConflictRepo)
	uc := usecases.NewConflictUseCase(repo)
	h := NewConflictHandler(uc)

	repo.On("BulkResolve", mock.Anything, []int64{1, 2}, entities.ConflictResolutionSkip, int64(1)).Return(nil)

	router := setupRouter()
	router.POST("/conflicts/bulk-resolve", withAuth(1), h.BulkResolve)
	w := performRequest(router, http.MethodPost, "/conflicts/bulk-resolve", map[string]interface{}{
		"ids":        []int64{1, 2},
		"resolution": "skip",
	})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestConflictHandler_BulkResolve_InvalidJSON(t *testing.T) {
	h := NewConflictHandler(nil)
	router := setupRouter()
	router.POST("/conflicts/bulk-resolve", h.BulkResolve)
	req := httptest.NewRequest(http.MethodPost, "/conflicts/bulk-resolve", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestConflictHandler_BulkResolve_EmptyIDs(t *testing.T) {
	h := NewConflictHandler(nil)
	router := setupRouter()
	router.POST("/conflicts/bulk-resolve", h.BulkResolve)
	w := performRequest(router, http.MethodPost, "/conflicts/bulk-resolve", map[string]interface{}{
		"ids":        []int64{},
		"resolution": "skip",
	})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestConflictHandler_BulkResolve_Error(t *testing.T) {
	repo := new(mockSyncConflictRepo)
	uc := usecases.NewConflictUseCase(repo)
	h := NewConflictHandler(uc)

	repo.On("BulkResolve", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("err"))

	router := setupRouter()
	router.POST("/conflicts/bulk-resolve", withAuth(1), h.BulkResolve)
	w := performRequest(router, http.MethodPost, "/conflicts/bulk-resolve", map[string]interface{}{
		"ids":        []int64{1},
		"resolution": "skip",
	})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestConflictHandler_Delete_Success(t *testing.T) {
	repo := new(mockSyncConflictRepo)
	uc := usecases.NewConflictUseCase(repo)
	h := NewConflictHandler(uc)

	repo.On("Delete", mock.Anything, int64(1)).Return(nil)

	router := setupRouter()
	router.DELETE("/conflicts/:id", h.Delete)
	w := performRequest(router, http.MethodDelete, "/conflicts/1", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestConflictHandler_Delete_InvalidID(t *testing.T) {
	h := NewConflictHandler(nil)
	router := setupRouter()
	router.DELETE("/conflicts/:id", h.Delete)
	w := performRequest(router, http.MethodDelete, "/conflicts/abc", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestConflictHandler_Delete_Error(t *testing.T) {
	repo := new(mockSyncConflictRepo)
	uc := usecases.NewConflictUseCase(repo)
	h := NewConflictHandler(uc)

	repo.On("Delete", mock.Anything, int64(1)).Return(fmt.Errorf("err"))

	router := setupRouter()
	router.DELETE("/conflicts/:id", h.Delete)
	w := performRequest(router, http.MethodDelete, "/conflicts/1", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ===== Employee Handler Tests =====

func TestNewEmployeeHandler(t *testing.T) {
	h := NewEmployeeHandler(nil)
	assert.NotNil(t, h)
}

func TestEmployeeHandler_RegisterRoutes(t *testing.T) {
	repo := new(mockEmployeeRepo)
	uc := usecases.NewEmployeeUseCase(repo)
	h := NewEmployeeHandler(uc)
	router := setupRouter()
	h.RegisterRoutes(router.Group("/api"))
}

func TestEmployeeHandler_List_Success(t *testing.T) {
	repo := new(mockEmployeeRepo)
	uc := usecases.NewEmployeeUseCase(repo)
	h := NewEmployeeHandler(uc)

	repo.On("List", mock.Anything, mock.Anything).Return([]*entities.ExternalEmployee{}, int64(0), nil)

	router := setupRouter()
	router.GET("/employees", h.List)
	w := performRequest(router, http.MethodGet, "/employees?search=test&department=IT&position=dev&is_active=true&is_linked=false&limit=10&offset=5", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEmployeeHandler_List_Error(t *testing.T) {
	repo := new(mockEmployeeRepo)
	uc := usecases.NewEmployeeUseCase(repo)
	h := NewEmployeeHandler(uc)

	repo.On("List", mock.Anything, mock.Anything).Return(nil, int64(0), fmt.Errorf("err"))

	router := setupRouter()
	router.GET("/employees", h.List)
	w := performRequest(router, http.MethodGet, "/employees", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestEmployeeHandler_GetByID_Success(t *testing.T) {
	repo := new(mockEmployeeRepo)
	uc := usecases.NewEmployeeUseCase(repo)
	h := NewEmployeeHandler(uc)

	emp := &entities.ExternalEmployee{ID: 1, FirstName: "John", LastName: "Doe"}
	repo.On("GetByID", mock.Anything, int64(1)).Return(emp, nil)

	router := setupRouter()
	router.GET("/employees/:id", h.GetByID)
	w := performRequest(router, http.MethodGet, "/employees/1", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEmployeeHandler_GetByID_InvalidID(t *testing.T) {
	h := NewEmployeeHandler(nil)
	router := setupRouter()
	router.GET("/employees/:id", h.GetByID)
	w := performRequest(router, http.MethodGet, "/employees/abc", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEmployeeHandler_GetByID_NotFound(t *testing.T) {
	repo := new(mockEmployeeRepo)
	uc := usecases.NewEmployeeUseCase(repo)
	h := NewEmployeeHandler(uc)

	repo.On("GetByID", mock.Anything, int64(1)).Return(nil, nil)

	router := setupRouter()
	router.GET("/employees/:id", h.GetByID)
	w := performRequest(router, http.MethodGet, "/employees/1", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestEmployeeHandler_GetByID_Error(t *testing.T) {
	repo := new(mockEmployeeRepo)
	uc := usecases.NewEmployeeUseCase(repo)
	h := NewEmployeeHandler(uc)

	repo.On("GetByID", mock.Anything, int64(1)).Return(nil, fmt.Errorf("err"))

	router := setupRouter()
	router.GET("/employees/:id", h.GetByID)
	w := performRequest(router, http.MethodGet, "/employees/1", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestEmployeeHandler_GetUnlinked_Success(t *testing.T) {
	repo := new(mockEmployeeRepo)
	uc := usecases.NewEmployeeUseCase(repo)
	h := NewEmployeeHandler(uc)

	repo.On("GetUnlinked", mock.Anything, 20, 0).Return([]*entities.ExternalEmployee{}, int64(0), nil)

	router := setupRouter()
	router.GET("/employees/unlinked", h.GetUnlinked)
	w := performRequest(router, http.MethodGet, "/employees/unlinked", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEmployeeHandler_GetUnlinked_Error(t *testing.T) {
	repo := new(mockEmployeeRepo)
	uc := usecases.NewEmployeeUseCase(repo)
	h := NewEmployeeHandler(uc)

	repo.On("GetUnlinked", mock.Anything, 20, 0).Return(nil, int64(0), fmt.Errorf("err"))

	router := setupRouter()
	router.GET("/employees/unlinked", h.GetUnlinked)
	w := performRequest(router, http.MethodGet, "/employees/unlinked", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestEmployeeHandler_Link_Success(t *testing.T) {
	repo := new(mockEmployeeRepo)
	uc := usecases.NewEmployeeUseCase(repo)
	h := NewEmployeeHandler(uc)

	emp := &entities.ExternalEmployee{ID: 1} // not linked
	repo.On("GetByID", mock.Anything, int64(1)).Return(emp, nil)
	repo.On("GetByLocalUserID", mock.Anything, int64(5)).Return(nil, nil)
	repo.On("LinkToLocalUser", mock.Anything, int64(1), int64(5)).Return(nil)

	router := setupRouter()
	router.POST("/employees/:id/link", h.Link)
	w := performRequest(router, http.MethodPost, "/employees/1/link", dto.LinkEmployeeRequest{LocalUserID: 5})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEmployeeHandler_Link_InvalidID(t *testing.T) {
	h := NewEmployeeHandler(nil)
	router := setupRouter()
	router.POST("/employees/:id/link", h.Link)
	w := performRequest(router, http.MethodPost, "/employees/abc/link", dto.LinkEmployeeRequest{LocalUserID: 5})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEmployeeHandler_Link_InvalidJSON(t *testing.T) {
	h := NewEmployeeHandler(nil)
	router := setupRouter()
	router.POST("/employees/:id/link", h.Link)
	req := httptest.NewRequest(http.MethodPost, "/employees/1/link", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEmployeeHandler_Link_Error(t *testing.T) {
	repo := new(mockEmployeeRepo)
	uc := usecases.NewEmployeeUseCase(repo)
	h := NewEmployeeHandler(uc)

	emp := &entities.ExternalEmployee{ID: 1}
	repo.On("GetByID", mock.Anything, int64(1)).Return(emp, nil)
	repo.On("GetByLocalUserID", mock.Anything, int64(5)).Return(nil, nil)
	repo.On("LinkToLocalUser", mock.Anything, int64(1), int64(5)).Return(fmt.Errorf("err"))

	router := setupRouter()
	router.POST("/employees/:id/link", h.Link)
	w := performRequest(router, http.MethodPost, "/employees/1/link", dto.LinkEmployeeRequest{LocalUserID: 5})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestEmployeeHandler_Unlink_Success(t *testing.T) {
	repo := new(mockEmployeeRepo)
	uc := usecases.NewEmployeeUseCase(repo)
	h := NewEmployeeHandler(uc)

	emp := &entities.ExternalEmployee{ID: 1}
	localUserID := int64(5)
	emp.LocalUserID = &localUserID
	repo.On("GetByID", mock.Anything, int64(1)).Return(emp, nil)
	repo.On("Unlink", mock.Anything, int64(1)).Return(nil)

	router := setupRouter()
	router.DELETE("/employees/:id/link", h.Unlink)
	w := performRequest(router, http.MethodDelete, "/employees/1/link", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEmployeeHandler_Unlink_InvalidID(t *testing.T) {
	h := NewEmployeeHandler(nil)
	router := setupRouter()
	router.DELETE("/employees/:id/link", h.Unlink)
	w := performRequest(router, http.MethodDelete, "/employees/abc/link", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEmployeeHandler_Delete_Success(t *testing.T) {
	repo := new(mockEmployeeRepo)
	uc := usecases.NewEmployeeUseCase(repo)
	h := NewEmployeeHandler(uc)

	repo.On("Delete", mock.Anything, int64(1)).Return(nil)

	router := setupRouter()
	router.DELETE("/employees/:id", h.Delete)
	w := performRequest(router, http.MethodDelete, "/employees/1", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEmployeeHandler_Delete_InvalidID(t *testing.T) {
	h := NewEmployeeHandler(nil)
	router := setupRouter()
	router.DELETE("/employees/:id", h.Delete)
	w := performRequest(router, http.MethodDelete, "/employees/abc", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestEmployeeHandler_Delete_Error(t *testing.T) {
	repo := new(mockEmployeeRepo)
	uc := usecases.NewEmployeeUseCase(repo)
	h := NewEmployeeHandler(uc)

	repo.On("Delete", mock.Anything, int64(1)).Return(fmt.Errorf("err"))

	router := setupRouter()
	router.DELETE("/employees/:id", h.Delete)
	w := performRequest(router, http.MethodDelete, "/employees/1", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ===== Student Handler Tests =====

func TestNewStudentHandler(t *testing.T) {
	h := NewStudentHandler(nil)
	assert.NotNil(t, h)
}

func TestStudentHandler_RegisterRoutes(t *testing.T) {
	repo := new(mockStudentRepo)
	uc := usecases.NewStudentUseCase(repo)
	h := NewStudentHandler(uc)
	router := setupRouter()
	h.RegisterRoutes(router.Group("/api"))
}

func TestStudentHandler_List_Success(t *testing.T) {
	repo := new(mockStudentRepo)
	uc := usecases.NewStudentUseCase(repo)
	h := NewStudentHandler(uc)

	repo.On("List", mock.Anything, mock.Anything).Return([]*entities.ExternalStudent{}, int64(0), nil)

	router := setupRouter()
	router.GET("/students", h.List)
	w := performRequest(router, http.MethodGet, "/students?search=test&group_name=G1&faculty=F1&status=active&course=2&is_active=true&is_linked=false", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestStudentHandler_List_Error(t *testing.T) {
	repo := new(mockStudentRepo)
	uc := usecases.NewStudentUseCase(repo)
	h := NewStudentHandler(uc)

	repo.On("List", mock.Anything, mock.Anything).Return(nil, int64(0), fmt.Errorf("err"))

	router := setupRouter()
	router.GET("/students", h.List)
	w := performRequest(router, http.MethodGet, "/students", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestStudentHandler_GetByID_Success(t *testing.T) {
	repo := new(mockStudentRepo)
	uc := usecases.NewStudentUseCase(repo)
	h := NewStudentHandler(uc)

	s := &entities.ExternalStudent{ID: 1, FirstName: "Jane", LastName: "Doe"}
	repo.On("GetByID", mock.Anything, int64(1)).Return(s, nil)

	router := setupRouter()
	router.GET("/students/:id", h.GetByID)
	w := performRequest(router, http.MethodGet, "/students/1", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestStudentHandler_GetByID_InvalidID(t *testing.T) {
	h := NewStudentHandler(nil)
	router := setupRouter()
	router.GET("/students/:id", h.GetByID)
	w := performRequest(router, http.MethodGet, "/students/abc", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestStudentHandler_GetByID_NotFound(t *testing.T) {
	repo := new(mockStudentRepo)
	uc := usecases.NewStudentUseCase(repo)
	h := NewStudentHandler(uc)

	repo.On("GetByID", mock.Anything, int64(1)).Return(nil, nil)

	router := setupRouter()
	router.GET("/students/:id", h.GetByID)
	w := performRequest(router, http.MethodGet, "/students/1", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestStudentHandler_GetUnlinked_Success(t *testing.T) {
	repo := new(mockStudentRepo)
	uc := usecases.NewStudentUseCase(repo)
	h := NewStudentHandler(uc)

	repo.On("GetUnlinked", mock.Anything, 20, 0).Return([]*entities.ExternalStudent{}, int64(0), nil)

	router := setupRouter()
	router.GET("/students/unlinked", h.GetUnlinked)
	w := performRequest(router, http.MethodGet, "/students/unlinked", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestStudentHandler_GetGroups_Success(t *testing.T) {
	repo := new(mockStudentRepo)
	uc := usecases.NewStudentUseCase(repo)
	h := NewStudentHandler(uc)

	repo.On("GetGroups", mock.Anything).Return([]string{"G1", "G2"}, nil)

	router := setupRouter()
	router.GET("/students/groups", h.GetGroups)
	w := performRequest(router, http.MethodGet, "/students/groups", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestStudentHandler_GetGroups_Error(t *testing.T) {
	repo := new(mockStudentRepo)
	uc := usecases.NewStudentUseCase(repo)
	h := NewStudentHandler(uc)

	repo.On("GetGroups", mock.Anything).Return(nil, fmt.Errorf("err"))

	router := setupRouter()
	router.GET("/students/groups", h.GetGroups)
	w := performRequest(router, http.MethodGet, "/students/groups", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestStudentHandler_GetFaculties_Success(t *testing.T) {
	repo := new(mockStudentRepo)
	uc := usecases.NewStudentUseCase(repo)
	h := NewStudentHandler(uc)

	repo.On("GetFaculties", mock.Anything).Return([]string{"F1", "F2"}, nil)

	router := setupRouter()
	router.GET("/students/faculties", h.GetFaculties)
	w := performRequest(router, http.MethodGet, "/students/faculties", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestStudentHandler_GetFaculties_Error(t *testing.T) {
	repo := new(mockStudentRepo)
	uc := usecases.NewStudentUseCase(repo)
	h := NewStudentHandler(uc)

	repo.On("GetFaculties", mock.Anything).Return(nil, fmt.Errorf("err"))

	router := setupRouter()
	router.GET("/students/faculties", h.GetFaculties)
	w := performRequest(router, http.MethodGet, "/students/faculties", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestStudentHandler_Link_Success(t *testing.T) {
	repo := new(mockStudentRepo)
	uc := usecases.NewStudentUseCase(repo)
	h := NewStudentHandler(uc)

	s := &entities.ExternalStudent{ID: 1}
	repo.On("GetByID", mock.Anything, int64(1)).Return(s, nil)
	repo.On("GetByLocalUserID", mock.Anything, int64(5)).Return(nil, nil)
	repo.On("LinkToLocalUser", mock.Anything, int64(1), int64(5)).Return(nil)

	router := setupRouter()
	router.POST("/students/:id/link", h.Link)
	w := performRequest(router, http.MethodPost, "/students/1/link", dto.LinkStudentRequest{LocalUserID: 5})
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestStudentHandler_Link_InvalidID(t *testing.T) {
	h := NewStudentHandler(nil)
	router := setupRouter()
	router.POST("/students/:id/link", h.Link)
	w := performRequest(router, http.MethodPost, "/students/abc/link", dto.LinkStudentRequest{LocalUserID: 5})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestStudentHandler_Link_InvalidJSON(t *testing.T) {
	h := NewStudentHandler(nil)
	router := setupRouter()
	router.POST("/students/:id/link", h.Link)
	req := httptest.NewRequest(http.MethodPost, "/students/1/link", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestStudentHandler_Unlink_Success(t *testing.T) {
	repo := new(mockStudentRepo)
	uc := usecases.NewStudentUseCase(repo)
	h := NewStudentHandler(uc)

	s := &entities.ExternalStudent{ID: 1}
	localUserID := int64(5)
	s.LocalUserID = &localUserID
	repo.On("GetByID", mock.Anything, int64(1)).Return(s, nil)
	repo.On("Unlink", mock.Anything, int64(1)).Return(nil)

	router := setupRouter()
	router.DELETE("/students/:id/link", h.Unlink)
	w := performRequest(router, http.MethodDelete, "/students/1/link", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestStudentHandler_Unlink_InvalidID(t *testing.T) {
	h := NewStudentHandler(nil)
	router := setupRouter()
	router.DELETE("/students/:id/link", h.Unlink)
	w := performRequest(router, http.MethodDelete, "/students/abc/link", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestStudentHandler_Delete_Success(t *testing.T) {
	repo := new(mockStudentRepo)
	uc := usecases.NewStudentUseCase(repo)
	h := NewStudentHandler(uc)

	repo.On("Delete", mock.Anything, int64(1)).Return(nil)

	router := setupRouter()
	router.DELETE("/students/:id", h.Delete)
	w := performRequest(router, http.MethodDelete, "/students/1", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestStudentHandler_Delete_InvalidID(t *testing.T) {
	h := NewStudentHandler(nil)
	router := setupRouter()
	router.DELETE("/students/:id", h.Delete)
	w := performRequest(router, http.MethodDelete, "/students/abc", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestStudentHandler_Delete_Error(t *testing.T) {
	repo := new(mockStudentRepo)
	uc := usecases.NewStudentUseCase(repo)
	h := NewStudentHandler(uc)

	repo.On("Delete", mock.Anything, int64(1)).Return(fmt.Errorf("err"))

	router := setupRouter()
	router.DELETE("/students/:id", h.Delete)
	w := performRequest(router, http.MethodDelete, "/students/1", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ===== Sync Handler Tests =====

func TestNewSyncHandler(t *testing.T) {
	h := NewSyncHandler(nil)
	assert.NotNil(t, h)
}

func TestSyncHandler_RegisterRoutes(t *testing.T) {
	h := NewSyncHandler(nil)
	router := setupRouter()
	h.RegisterRoutes(router.Group("/api"))
}

func TestSyncHandler_StartSync_InvalidJSON(t *testing.T) {
	h := NewSyncHandler(nil)
	router := setupRouter()
	router.POST("/sync/start", h.StartSync)
	req := httptest.NewRequest(http.MethodPost, "/sync/start", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSyncHandler_StartSync_InvalidEntityType(t *testing.T) {
	h := NewSyncHandler(nil)
	router := setupRouter()
	router.POST("/sync/start", h.StartSync)
	w := performRequest(router, http.MethodPost, "/sync/start", map[string]interface{}{
		"entity_type": "invalid",
		"direction":   "import",
	})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSyncHandler_StartSync_InvalidDirection(t *testing.T) {
	h := NewSyncHandler(nil)
	router := setupRouter()
	router.POST("/sync/start", h.StartSync)
	w := performRequest(router, http.MethodPost, "/sync/start", map[string]interface{}{
		"entity_type": "employee",
		"direction":   "invalid",
	})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSyncHandler_ListSyncLogs_InvalidID(t *testing.T) {
	h := NewSyncHandler(nil)
	router := setupRouter()
	router.GET("/sync/logs/:id", h.GetSyncLog)
	w := performRequest(router, http.MethodGet, "/sync/logs/abc", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSyncHandler_CancelSync_InvalidID(t *testing.T) {
	h := NewSyncHandler(nil)
	router := setupRouter()
	router.POST("/sync/logs/:id/cancel", h.CancelSync)
	w := performRequest(router, http.MethodPost, "/sync/logs/abc/cancel", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSyncHandler_StartSync_ValidEntityTypes(t *testing.T) {
	// Test valid employee entity type but invalid direction
	h := NewSyncHandler(nil)
	router := setupRouter()
	router.POST("/sync/start", h.StartSync)
	w := performRequest(router, http.MethodPost, "/sync/start", map[string]interface{}{
		"entity_type": "student",
		"direction":   "invalid",
	})
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestStudentHandler_GetByID_Error(t *testing.T) {
	repo := new(mockStudentRepo)
	uc := usecases.NewStudentUseCase(repo)
	h := NewStudentHandler(uc)

	repo.On("GetByID", mock.Anything, int64(1)).Return(nil, fmt.Errorf("err"))

	router := setupRouter()
	router.GET("/students/:id", h.GetByID)
	w := performRequest(router, http.MethodGet, "/students/1", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestStudentHandler_GetUnlinked_Error(t *testing.T) {
	repo := new(mockStudentRepo)
	uc := usecases.NewStudentUseCase(repo)
	h := NewStudentHandler(uc)

	repo.On("GetUnlinked", mock.Anything, 20, 0).Return(nil, int64(0), fmt.Errorf("err"))

	router := setupRouter()
	router.GET("/students/unlinked", h.GetUnlinked)
	w := performRequest(router, http.MethodGet, "/students/unlinked", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestStudentHandler_Link_Error(t *testing.T) {
	repo := new(mockStudentRepo)
	uc := usecases.NewStudentUseCase(repo)
	h := NewStudentHandler(uc)

	s := &entities.ExternalStudent{ID: 1}
	repo.On("GetByID", mock.Anything, int64(1)).Return(s, nil)
	repo.On("GetByLocalUserID", mock.Anything, int64(5)).Return(nil, nil)
	repo.On("LinkToLocalUser", mock.Anything, int64(1), int64(5)).Return(fmt.Errorf("err"))

	router := setupRouter()
	router.POST("/students/:id/link", h.Link)
	w := performRequest(router, http.MethodPost, "/students/1/link", dto.LinkStudentRequest{LocalUserID: 5})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestStudentHandler_Unlink_Error(t *testing.T) {
	repo := new(mockStudentRepo)
	uc := usecases.NewStudentUseCase(repo)
	h := NewStudentHandler(uc)

	s := &entities.ExternalStudent{ID: 1}
	localUserID := int64(5)
	s.LocalUserID = &localUserID
	repo.On("GetByID", mock.Anything, int64(1)).Return(s, nil)
	repo.On("Unlink", mock.Anything, int64(1)).Return(fmt.Errorf("err"))

	router := setupRouter()
	router.DELETE("/students/:id/link", h.Unlink)
	w := performRequest(router, http.MethodDelete, "/students/1/link", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

type mockSyncLogRepo struct{ mock.Mock }

func (m *mockSyncLogRepo) Create(ctx context.Context, log *entities.SyncLog) error {
	return m.Called(ctx, log).Error(0)
}
func (m *mockSyncLogRepo) Update(ctx context.Context, log *entities.SyncLog) error {
	return m.Called(ctx, log).Error(0)
}
func (m *mockSyncLogRepo) GetByID(ctx context.Context, id int64) (*entities.SyncLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.SyncLog), args.Error(1)
}
func (m *mockSyncLogRepo) List(ctx context.Context, filter entities.SyncLogFilter) ([]*entities.SyncLog, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*entities.SyncLog), args.Get(1).(int64), args.Error(2)
}
func (m *mockSyncLogRepo) GetLatest(ctx context.Context, entityType entities.SyncEntityType) (*entities.SyncLog, error) {
	args := m.Called(ctx, entityType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.SyncLog), args.Error(1)
}
func (m *mockSyncLogRepo) GetRunning(ctx context.Context) ([]*entities.SyncLog, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.SyncLog), args.Error(1)
}
func (m *mockSyncLogRepo) GetStats(ctx context.Context, entityType *entities.SyncEntityType) (*entities.SyncStats, error) {
	args := m.Called(ctx, entityType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.SyncStats), args.Error(1)
}
func (m *mockSyncLogRepo) Delete(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockSyncLogRepo) DeleteOlderThan(ctx context.Context, days int) (int64, error) {
	args := m.Called(ctx, days)
	return args.Get(0).(int64), args.Error(1)
}

func newSyncHandler(syncLogRepo *mockSyncLogRepo, empRepo *mockEmployeeRepo, stuRepo *mockStudentRepo, conflictRepo *mockSyncConflictRepo) *SyncHandler {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	uc := usecases.NewSyncUseCase(nil, syncLogRepo, empRepo, stuRepo, conflictRepo, logger)
	return NewSyncHandler(uc)
}

func TestSyncHandler_ListSyncLogs_Success(t *testing.T) {
	syncLogRepo := new(mockSyncLogRepo)
	h := newSyncHandler(syncLogRepo, new(mockEmployeeRepo), new(mockStudentRepo), new(mockSyncConflictRepo))

	syncLogRepo.On("List", mock.Anything, mock.Anything).Return([]*entities.SyncLog{}, int64(0), nil)

	router := setupRouter()
	router.GET("/sync/logs", h.ListSyncLogs)
	w := performRequest(router, http.MethodGet, "/sync/logs?entity_type=employee&direction=import&status=completed&limit=10&offset=5", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSyncHandler_ListSyncLogs_Error(t *testing.T) {
	syncLogRepo := new(mockSyncLogRepo)
	h := newSyncHandler(syncLogRepo, new(mockEmployeeRepo), new(mockStudentRepo), new(mockSyncConflictRepo))

	syncLogRepo.On("List", mock.Anything, mock.Anything).Return(nil, int64(0), fmt.Errorf("err"))

	router := setupRouter()
	router.GET("/sync/logs", h.ListSyncLogs)
	w := performRequest(router, http.MethodGet, "/sync/logs", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSyncHandler_GetSyncLog_Success(t *testing.T) {
	syncLogRepo := new(mockSyncLogRepo)
	h := newSyncHandler(syncLogRepo, new(mockEmployeeRepo), new(mockStudentRepo), new(mockSyncConflictRepo))

	log := &entities.SyncLog{ID: 1, EntityType: entities.SyncEntityEmployee}
	syncLogRepo.On("GetByID", mock.Anything, int64(1)).Return(log, nil)

	router := setupRouter()
	router.GET("/sync/logs/:id", h.GetSyncLog)
	w := performRequest(router, http.MethodGet, "/sync/logs/1", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSyncHandler_GetSyncLog_NotFound(t *testing.T) {
	syncLogRepo := new(mockSyncLogRepo)
	h := newSyncHandler(syncLogRepo, new(mockEmployeeRepo), new(mockStudentRepo), new(mockSyncConflictRepo))

	syncLogRepo.On("GetByID", mock.Anything, int64(1)).Return(nil, nil)

	router := setupRouter()
	router.GET("/sync/logs/:id", h.GetSyncLog)
	w := performRequest(router, http.MethodGet, "/sync/logs/1", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSyncHandler_GetSyncLog_Error(t *testing.T) {
	syncLogRepo := new(mockSyncLogRepo)
	h := newSyncHandler(syncLogRepo, new(mockEmployeeRepo), new(mockStudentRepo), new(mockSyncConflictRepo))

	syncLogRepo.On("GetByID", mock.Anything, int64(1)).Return(nil, fmt.Errorf("err"))

	router := setupRouter()
	router.GET("/sync/logs/:id", h.GetSyncLog)
	w := performRequest(router, http.MethodGet, "/sync/logs/1", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSyncHandler_CancelSync_Success(t *testing.T) {
	syncLogRepo := new(mockSyncLogRepo)
	h := newSyncHandler(syncLogRepo, new(mockEmployeeRepo), new(mockStudentRepo), new(mockSyncConflictRepo))

	log := &entities.SyncLog{ID: 1, Status: entities.SyncStatusInProgress}
	syncLogRepo.On("GetByID", mock.Anything, int64(1)).Return(log, nil)
	syncLogRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

	router := setupRouter()
	router.POST("/sync/logs/:id/cancel", h.CancelSync)
	w := performRequest(router, http.MethodPost, "/sync/logs/1/cancel", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSyncHandler_CancelSync_Error(t *testing.T) {
	syncLogRepo := new(mockSyncLogRepo)
	h := newSyncHandler(syncLogRepo, new(mockEmployeeRepo), new(mockStudentRepo), new(mockSyncConflictRepo))

	syncLogRepo.On("GetByID", mock.Anything, int64(1)).Return(nil, fmt.Errorf("err"))

	router := setupRouter()
	router.POST("/sync/logs/:id/cancel", h.CancelSync)
	w := performRequest(router, http.MethodPost, "/sync/logs/1/cancel", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSyncHandler_GetSyncStats_Success(t *testing.T) {
	syncLogRepo := new(mockSyncLogRepo)
	h := newSyncHandler(syncLogRepo, new(mockEmployeeRepo), new(mockStudentRepo), new(mockSyncConflictRepo))

	stats := &entities.SyncStats{TotalSyncs: 10}
	syncLogRepo.On("GetStats", mock.Anything, mock.Anything).Return(stats, nil)

	router := setupRouter()
	router.GET("/sync/stats", h.GetSyncStats)
	w := performRequest(router, http.MethodGet, "/sync/stats?entity_type=employee", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSyncHandler_GetSyncStats_NoEntityType(t *testing.T) {
	syncLogRepo := new(mockSyncLogRepo)
	h := newSyncHandler(syncLogRepo, new(mockEmployeeRepo), new(mockStudentRepo), new(mockSyncConflictRepo))

	stats := &entities.SyncStats{TotalSyncs: 10}
	syncLogRepo.On("GetStats", mock.Anything, mock.Anything).Return(stats, nil)

	router := setupRouter()
	router.GET("/sync/stats", h.GetSyncStats)
	w := performRequest(router, http.MethodGet, "/sync/stats", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSyncHandler_GetSyncStats_Error(t *testing.T) {
	syncLogRepo := new(mockSyncLogRepo)
	h := newSyncHandler(syncLogRepo, new(mockEmployeeRepo), new(mockStudentRepo), new(mockSyncConflictRepo))

	syncLogRepo.On("GetStats", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("err"))

	router := setupRouter()
	router.GET("/sync/stats", h.GetSyncStats)
	w := performRequest(router, http.MethodGet, "/sync/stats", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSyncHandler_GetStatus(t *testing.T) {
	syncLogRepo := new(mockSyncLogRepo)
	h := newSyncHandler(syncLogRepo, new(mockEmployeeRepo), new(mockStudentRepo), new(mockSyncConflictRepo))

	router := setupRouter()
	router.GET("/sync/status", h.GetStatus)
	w := performRequest(router, http.MethodGet, "/sync/status", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestEmployeeHandler_Unlink_Error(t *testing.T) {
	repo := new(mockEmployeeRepo)
	uc := usecases.NewEmployeeUseCase(repo)
	h := NewEmployeeHandler(uc)

	emp := &entities.ExternalEmployee{ID: 1}
	localUserID := int64(5)
	emp.LocalUserID = &localUserID
	repo.On("GetByID", mock.Anything, int64(1)).Return(emp, nil)
	repo.On("Unlink", mock.Anything, int64(1)).Return(fmt.Errorf("err"))

	router := setupRouter()
	router.DELETE("/employees/:id/link", h.Unlink)
	w := performRequest(router, http.MethodDelete, "/employees/1/link", nil)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
