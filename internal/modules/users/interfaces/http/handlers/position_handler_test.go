package handlers

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/domain/entities"
	domainErrors "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/domain/errors"
)

func newPositionUseCase(repo *mockPositionRepo) *usecases.PositionUseCase {
	return usecases.NewPositionUseCase(repo, nil)
}

func setupPositionRouter(handler *PositionHandler) *gin.Engine {
	r := gin.New()
	r.POST("/positions", handler.Create)
	r.GET("/positions", handler.List)
	r.GET("/positions/:id", handler.GetByID)
	r.PUT("/positions/:id", handler.Update)
	r.DELETE("/positions/:id", handler.Delete)
	return r
}

// --- Create ---

func TestPositionHandler_Create_Success(t *testing.T) {
	repo := new(mockPositionRepo)
	repo.On("Create", mock.Anything, mock.Anything).Return(nil)

	uc := newPositionUseCase(repo)
	handler := NewPositionHandler(uc)
	router := setupPositionRouter(handler)

	body := map[string]interface{}{
		"name":  "Developer",
		"code":  "DEV1",
		"level": 3,
	}
	req := httptest.NewRequest(http.MethodPost, "/positions", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestPositionHandler_Create_InvalidJSON(t *testing.T) {
	uc := newPositionUseCase(new(mockPositionRepo))
	handler := NewPositionHandler(uc)
	router := setupPositionRouter(handler)

	req := httptest.NewRequest(http.MethodPost, "/positions", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPositionHandler_Create_ValidationError(t *testing.T) {
	uc := newPositionUseCase(new(mockPositionRepo))
	handler := NewPositionHandler(uc)
	router := setupPositionRouter(handler)

	// Missing required name and code
	body := map[string]interface{}{
		"level": 3,
	}
	req := httptest.NewRequest(http.MethodPost, "/positions", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPositionHandler_Create_UsecaseError(t *testing.T) {
	repo := new(mockPositionRepo)
	repo.On("Create", mock.Anything, mock.Anything).Return(domainErrors.ErrAlreadyExists)

	uc := newPositionUseCase(repo)
	handler := NewPositionHandler(uc)
	router := setupPositionRouter(handler)

	body := map[string]interface{}{
		"name":  "Developer",
		"code":  "DEV1",
		"level": 3,
	}
	req := httptest.NewRequest(http.MethodPost, "/positions", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

// --- List ---

func TestPositionHandler_List_Success(t *testing.T) {
	repo := new(mockPositionRepo)
	positions := []*entities.Position{{ID: 1, Name: "Dev"}}
	repo.On("List", mock.Anything, 10, 0, false).Return(positions, nil)
	repo.On("Count", mock.Anything, false).Return(int64(1), nil)

	uc := newPositionUseCase(repo)
	handler := NewPositionHandler(uc)
	router := setupPositionRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/positions", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPositionHandler_List_WithParams(t *testing.T) {
	repo := new(mockPositionRepo)
	repo.On("List", mock.Anything, 5, 5, true).Return([]*entities.Position{}, nil)
	repo.On("Count", mock.Anything, true).Return(int64(0), nil)

	uc := newPositionUseCase(repo)
	handler := NewPositionHandler(uc)
	router := setupPositionRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/positions?page=2&limit=5&active_only=true", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPositionHandler_List_UsecaseError(t *testing.T) {
	repo := new(mockPositionRepo)
	repo.On("List", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("db error"))

	uc := newPositionUseCase(repo)
	handler := NewPositionHandler(uc)
	router := setupPositionRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/positions", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- GetByID ---

func TestPositionHandler_GetByID_Success(t *testing.T) {
	repo := new(mockPositionRepo)
	pos := &entities.Position{ID: 1, Name: "Dev"}
	repo.On("GetByID", mock.Anything, int64(1)).Return(pos, nil)

	uc := newPositionUseCase(repo)
	handler := NewPositionHandler(uc)
	router := setupPositionRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/positions/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPositionHandler_GetByID_InvalidID(t *testing.T) {
	uc := newPositionUseCase(new(mockPositionRepo))
	handler := NewPositionHandler(uc)
	router := setupPositionRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/positions/abc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPositionHandler_GetByID_NotFound(t *testing.T) {
	repo := new(mockPositionRepo)
	repo.On("GetByID", mock.Anything, int64(999)).Return(nil, domainErrors.ErrNotFound)

	uc := newPositionUseCase(repo)
	handler := NewPositionHandler(uc)
	router := setupPositionRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/positions/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// --- Update ---

func TestPositionHandler_Update_Success(t *testing.T) {
	repo := new(mockPositionRepo)
	pos := &entities.Position{ID: 1, Name: "Old"}
	repo.On("GetByID", mock.Anything, int64(1)).Return(pos, nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(nil)

	uc := newPositionUseCase(repo)
	handler := NewPositionHandler(uc)
	router := setupPositionRouter(handler)

	body := map[string]interface{}{
		"name":  "Updated",
		"code":  "UPD1",
		"level": 5,
	}
	req := httptest.NewRequest(http.MethodPut, "/positions/1", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPositionHandler_Update_InvalidID(t *testing.T) {
	uc := newPositionUseCase(new(mockPositionRepo))
	handler := NewPositionHandler(uc)
	router := setupPositionRouter(handler)

	req := httptest.NewRequest(http.MethodPut, "/positions/abc", jsonBody(t, map[string]interface{}{}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPositionHandler_Update_InvalidJSON(t *testing.T) {
	uc := newPositionUseCase(new(mockPositionRepo))
	handler := NewPositionHandler(uc)
	router := setupPositionRouter(handler)

	req := httptest.NewRequest(http.MethodPut, "/positions/1", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPositionHandler_Update_ValidationError(t *testing.T) {
	uc := newPositionUseCase(new(mockPositionRepo))
	handler := NewPositionHandler(uc)
	router := setupPositionRouter(handler)

	body := map[string]interface{}{
		"level": 5,
	}
	req := httptest.NewRequest(http.MethodPut, "/positions/1", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPositionHandler_Update_UsecaseError(t *testing.T) {
	repo := new(mockPositionRepo)
	repo.On("GetByID", mock.Anything, int64(1)).Return(nil, domainErrors.ErrNotFound)

	uc := newPositionUseCase(repo)
	handler := NewPositionHandler(uc)
	router := setupPositionRouter(handler)

	body := map[string]interface{}{
		"name":  "Updated",
		"code":  "UPD1",
		"level": 5,
	}
	req := httptest.NewRequest(http.MethodPut, "/positions/1", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// --- Delete ---

func TestPositionHandler_Delete_Success(t *testing.T) {
	repo := new(mockPositionRepo)
	pos := &entities.Position{ID: 1, Name: "Dev"}
	repo.On("GetByID", mock.Anything, int64(1)).Return(pos, nil)
	repo.On("Delete", mock.Anything, int64(1)).Return(nil)

	uc := newPositionUseCase(repo)
	handler := NewPositionHandler(uc)
	router := setupPositionRouter(handler)

	req := httptest.NewRequest(http.MethodDelete, "/positions/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPositionHandler_Delete_InvalidID(t *testing.T) {
	uc := newPositionUseCase(new(mockPositionRepo))
	handler := NewPositionHandler(uc)
	router := setupPositionRouter(handler)

	req := httptest.NewRequest(http.MethodDelete, "/positions/abc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPositionHandler_Delete_NotFound(t *testing.T) {
	repo := new(mockPositionRepo)
	repo.On("GetByID", mock.Anything, int64(999)).Return(nil, domainErrors.ErrNotFound)

	uc := newPositionUseCase(repo)
	handler := NewPositionHandler(uc)
	router := setupPositionRouter(handler)

	req := httptest.NewRequest(http.MethodDelete, "/positions/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestPositionHandler_Delete_UsecaseError(t *testing.T) {
	repo := new(mockPositionRepo)
	pos := &entities.Position{ID: 1, Name: "Dev"}
	repo.On("GetByID", mock.Anything, int64(1)).Return(pos, nil)
	repo.On("Delete", mock.Anything, int64(1)).Return(errors.New("db error"))

	uc := newPositionUseCase(repo)
	handler := NewPositionHandler(uc)
	router := setupPositionRouter(handler)

	req := httptest.NewRequest(http.MethodDelete, "/positions/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestNewPositionHandler(t *testing.T) {
	uc := newPositionUseCase(new(mockPositionRepo))
	handler := NewPositionHandler(uc)
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.usecase)
	assert.NotNil(t, handler.validator)
	assert.NotNil(t, handler.sanitizer)
}
