package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/domain/entities"
	domainErrors "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/domain/errors"
)

func newDepartmentUseCase(repo *mockDepartmentRepo) *usecases.DepartmentUseCase {
	return usecases.NewDepartmentUseCase(repo, nil)
}

func setupDepartmentRouter(handler *DepartmentHandler) *gin.Engine {
	r := gin.New()
	r.POST("/departments", handler.Create)
	r.GET("/departments", handler.List)
	r.GET("/departments/:id", handler.GetByID)
	r.PUT("/departments/:id", handler.Update)
	r.DELETE("/departments/:id", handler.Delete)
	r.GET("/departments/:id/children", handler.GetChildren)
	return r
}

// --- Create ---

func TestDepartmentHandler_Create_Success(t *testing.T) {
	repo := new(mockDepartmentRepo)
	repo.On("Create", mock.Anything, mock.Anything).Return(nil)

	uc := newDepartmentUseCase(repo)
	handler := NewDepartmentHandler(uc)
	router := setupDepartmentRouter(handler)

	body := map[string]interface{}{
		"name": "IT Department",
		"code": "IT01",
	}
	req := httptest.NewRequest(http.MethodPost, "/departments", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["success"].(bool))
}

func TestDepartmentHandler_Create_InvalidJSON(t *testing.T) {
	uc := newDepartmentUseCase(new(mockDepartmentRepo))
	handler := NewDepartmentHandler(uc)
	router := setupDepartmentRouter(handler)

	req := httptest.NewRequest(http.MethodPost, "/departments", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDepartmentHandler_Create_ValidationError(t *testing.T) {
	uc := newDepartmentUseCase(new(mockDepartmentRepo))
	handler := NewDepartmentHandler(uc)
	router := setupDepartmentRouter(handler)

	// Missing required "name" and "code"
	body := map[string]interface{}{
		"description": "Something",
	}
	req := httptest.NewRequest(http.MethodPost, "/departments", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDepartmentHandler_Create_UsecaseError(t *testing.T) {
	repo := new(mockDepartmentRepo)
	repo.On("Create", mock.Anything, mock.Anything).Return(domainErrors.ErrAlreadyExists)

	uc := newDepartmentUseCase(repo)
	handler := NewDepartmentHandler(uc)
	router := setupDepartmentRouter(handler)

	body := map[string]interface{}{
		"name": "IT Department",
		"code": "IT01",
	}
	req := httptest.NewRequest(http.MethodPost, "/departments", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

// --- List ---

func TestDepartmentHandler_List_Success(t *testing.T) {
	repo := new(mockDepartmentRepo)
	depts := []*entities.Department{{ID: 1, Name: "IT"}}
	repo.On("List", mock.Anything, 10, 0, false).Return(depts, nil)
	repo.On("Count", mock.Anything, false).Return(int64(1), nil)

	uc := newDepartmentUseCase(repo)
	handler := NewDepartmentHandler(uc)
	router := setupDepartmentRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/departments", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDepartmentHandler_List_WithParams(t *testing.T) {
	repo := new(mockDepartmentRepo)
	repo.On("List", mock.Anything, 5, 5, true).Return([]*entities.Department{}, nil)
	repo.On("Count", mock.Anything, true).Return(int64(0), nil)

	uc := newDepartmentUseCase(repo)
	handler := NewDepartmentHandler(uc)
	router := setupDepartmentRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/departments?page=2&limit=5&active_only=true", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDepartmentHandler_List_UsecaseError(t *testing.T) {
	repo := new(mockDepartmentRepo)
	repo.On("List", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("db error"))

	uc := newDepartmentUseCase(repo)
	handler := NewDepartmentHandler(uc)
	router := setupDepartmentRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/departments", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- GetByID ---

func TestDepartmentHandler_GetByID_Success(t *testing.T) {
	repo := new(mockDepartmentRepo)
	dept := &entities.Department{ID: 1, Name: "IT"}
	repo.On("GetByID", mock.Anything, int64(1)).Return(dept, nil)

	uc := newDepartmentUseCase(repo)
	handler := NewDepartmentHandler(uc)
	router := setupDepartmentRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/departments/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDepartmentHandler_GetByID_InvalidID(t *testing.T) {
	uc := newDepartmentUseCase(new(mockDepartmentRepo))
	handler := NewDepartmentHandler(uc)
	router := setupDepartmentRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/departments/abc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDepartmentHandler_GetByID_NotFound(t *testing.T) {
	repo := new(mockDepartmentRepo)
	repo.On("GetByID", mock.Anything, int64(999)).Return(nil, domainErrors.ErrNotFound)

	uc := newDepartmentUseCase(repo)
	handler := NewDepartmentHandler(uc)
	router := setupDepartmentRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/departments/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// --- Update ---

func TestDepartmentHandler_Update_Success(t *testing.T) {
	repo := new(mockDepartmentRepo)
	dept := &entities.Department{ID: 1, Name: "Old"}
	repo.On("GetByID", mock.Anything, int64(1)).Return(dept, nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(nil)

	uc := newDepartmentUseCase(repo)
	handler := NewDepartmentHandler(uc)
	router := setupDepartmentRouter(handler)

	body := map[string]interface{}{
		"name": "Updated Name",
		"code": "UPD1",
	}
	req := httptest.NewRequest(http.MethodPut, "/departments/1", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDepartmentHandler_Update_InvalidID(t *testing.T) {
	uc := newDepartmentUseCase(new(mockDepartmentRepo))
	handler := NewDepartmentHandler(uc)
	router := setupDepartmentRouter(handler)

	req := httptest.NewRequest(http.MethodPut, "/departments/abc", jsonBody(t, map[string]interface{}{}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDepartmentHandler_Update_InvalidJSON(t *testing.T) {
	uc := newDepartmentUseCase(new(mockDepartmentRepo))
	handler := NewDepartmentHandler(uc)
	router := setupDepartmentRouter(handler)

	req := httptest.NewRequest(http.MethodPut, "/departments/1", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDepartmentHandler_Update_ValidationError(t *testing.T) {
	uc := newDepartmentUseCase(new(mockDepartmentRepo))
	handler := NewDepartmentHandler(uc)
	router := setupDepartmentRouter(handler)

	// Missing required fields
	body := map[string]interface{}{
		"description": "test",
	}
	req := httptest.NewRequest(http.MethodPut, "/departments/1", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDepartmentHandler_Update_UsecaseError(t *testing.T) {
	repo := new(mockDepartmentRepo)
	repo.On("GetByID", mock.Anything, int64(1)).Return(nil, domainErrors.ErrNotFound)

	uc := newDepartmentUseCase(repo)
	handler := NewDepartmentHandler(uc)
	router := setupDepartmentRouter(handler)

	body := map[string]interface{}{
		"name": "Updated",
		"code": "UPD1",
	}
	req := httptest.NewRequest(http.MethodPut, "/departments/1", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// --- Delete ---

func TestDepartmentHandler_Delete_Success(t *testing.T) {
	repo := new(mockDepartmentRepo)
	dept := &entities.Department{ID: 1, Name: "IT"}
	repo.On("GetByID", mock.Anything, int64(1)).Return(dept, nil)
	repo.On("GetChildren", mock.Anything, int64(1)).Return([]*entities.Department{}, nil)
	repo.On("Delete", mock.Anything, int64(1)).Return(nil)

	uc := newDepartmentUseCase(repo)
	handler := NewDepartmentHandler(uc)
	router := setupDepartmentRouter(handler)

	req := httptest.NewRequest(http.MethodDelete, "/departments/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDepartmentHandler_Delete_InvalidID(t *testing.T) {
	uc := newDepartmentUseCase(new(mockDepartmentRepo))
	handler := NewDepartmentHandler(uc)
	router := setupDepartmentRouter(handler)

	req := httptest.NewRequest(http.MethodDelete, "/departments/abc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDepartmentHandler_Delete_HasChildren(t *testing.T) {
	repo := new(mockDepartmentRepo)
	dept := &entities.Department{ID: 1, Name: "IT"}
	repo.On("GetByID", mock.Anything, int64(1)).Return(dept, nil)
	repo.On("GetChildren", mock.Anything, int64(1)).Return([]*entities.Department{{ID: 2}}, nil)

	uc := newDepartmentUseCase(repo)
	handler := NewDepartmentHandler(uc)
	router := setupDepartmentRouter(handler)

	req := httptest.NewRequest(http.MethodDelete, "/departments/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDepartmentHandler_Delete_NotFound(t *testing.T) {
	repo := new(mockDepartmentRepo)
	repo.On("GetByID", mock.Anything, int64(999)).Return(nil, domainErrors.ErrNotFound)

	uc := newDepartmentUseCase(repo)
	handler := NewDepartmentHandler(uc)
	router := setupDepartmentRouter(handler)

	req := httptest.NewRequest(http.MethodDelete, "/departments/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDepartmentHandler_Delete_GenericUsecaseError(t *testing.T) {
	repo := new(mockDepartmentRepo)
	dept := &entities.Department{ID: 1, Name: "IT"}
	repo.On("GetByID", mock.Anything, int64(1)).Return(dept, nil)
	repo.On("GetChildren", mock.Anything, int64(1)).Return([]*entities.Department{}, nil)
	repo.On("Delete", mock.Anything, int64(1)).Return(errors.New("db error"))

	uc := newDepartmentUseCase(repo)
	handler := NewDepartmentHandler(uc)
	router := setupDepartmentRouter(handler)

	req := httptest.NewRequest(http.MethodDelete, "/departments/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- GetChildren ---

func TestDepartmentHandler_GetChildren_Success(t *testing.T) {
	repo := new(mockDepartmentRepo)
	children := []*entities.Department{{ID: 2, Name: "Sub-IT"}}
	repo.On("GetChildren", mock.Anything, int64(1)).Return(children, nil)

	uc := newDepartmentUseCase(repo)
	handler := NewDepartmentHandler(uc)
	router := setupDepartmentRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/departments/1/children", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDepartmentHandler_GetChildren_InvalidID(t *testing.T) {
	uc := newDepartmentUseCase(new(mockDepartmentRepo))
	handler := NewDepartmentHandler(uc)
	router := setupDepartmentRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/departments/abc/children", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDepartmentHandler_GetChildren_UsecaseError(t *testing.T) {
	repo := new(mockDepartmentRepo)
	repo.On("GetChildren", mock.Anything, int64(1)).Return(nil, errors.New("db error"))

	uc := newDepartmentUseCase(repo)
	handler := NewDepartmentHandler(uc)
	router := setupDepartmentRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/departments/1/children", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestNewDepartmentHandler(t *testing.T) {
	uc := newDepartmentUseCase(new(mockDepartmentRepo))
	handler := NewDepartmentHandler(uc)
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.usecase)
	assert.NotNil(t, handler.validator)
	assert.NotNil(t, handler.sanitizer)
}
