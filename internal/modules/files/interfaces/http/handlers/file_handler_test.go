package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupFileRouter(handler *FileHandler) *gin.Engine {
	r := gin.New()
	r.POST("/files/upload", handler.Upload)
	r.GET("/files/:id", handler.GetByID)
	r.GET("/files/:id/download", handler.Download)
	r.POST("/files/:id/attach", handler.Attach)
	r.DELETE("/files/:id", handler.Delete)
	r.GET("/files", handler.List)
	r.GET("/documents/:document_id/files", handler.GetByDocument)
	r.GET("/tasks/:task_id/files", handler.GetByTask)
	r.GET("/announcements/:announcement_id/files", handler.GetByAnnouncement)
	r.POST("/files/:id/versions", handler.CreateVersion)
	r.GET("/files/:id/versions", handler.GetVersions)
	r.GET("/files/:id/versions/:version", handler.DownloadVersion)
	r.POST("/files/cleanup", handler.CleanupExpired)
	return r
}

func TestFileHandler_Upload_NoFile(t *testing.T) {
	handler := NewFileHandler(nil, nil)
	r := setupFileRouter(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/files/upload", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFileHandler_Upload_Unauthorized(t *testing.T) {
	handler := NewFileHandler(nil, nil)
	r := gin.New()
	r.POST("/files/upload", handler.Upload)

	// Use multipart form with no user_id in context
	body := strings.NewReader("")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/files/upload", body)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFileHandler_GetByID_InvalidID(t *testing.T) {
	handler := NewFileHandler(nil, nil)
	r := setupFileRouter(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/files/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFileHandler_Download_InvalidID(t *testing.T) {
	handler := NewFileHandler(nil, nil)
	r := setupFileRouter(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/files/abc/download", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFileHandler_Attach_InvalidID(t *testing.T) {
	handler := NewFileHandler(nil, nil)
	r := setupFileRouter(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/files/abc/attach", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFileHandler_Attach_InvalidJSON(t *testing.T) {
	handler := NewFileHandler(nil, nil)
	r := setupFileRouter(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/files/1/attach", strings.NewReader(`{bad`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFileHandler_Delete_InvalidID(t *testing.T) {
	handler := NewFileHandler(nil, nil)
	r := setupFileRouter(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/files/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFileHandler_Delete_Unauthorized(t *testing.T) {
	handler := NewFileHandler(nil, nil)
	r := gin.New()
	r.DELETE("/files/:id", handler.Delete)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/files/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestFileHandler_GetByDocument_InvalidID(t *testing.T) {
	handler := NewFileHandler(nil, nil)
	r := setupFileRouter(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/documents/abc/files", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFileHandler_GetByTask_InvalidID(t *testing.T) {
	handler := NewFileHandler(nil, nil)
	r := setupFileRouter(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/tasks/abc/files", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFileHandler_GetByAnnouncement_InvalidID(t *testing.T) {
	handler := NewFileHandler(nil, nil)
	r := setupFileRouter(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/announcements/abc/files", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFileHandler_CreateVersion_InvalidID(t *testing.T) {
	handler := NewFileHandler(nil, nil)
	r := setupFileRouter(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/files/abc/versions", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFileHandler_GetVersions_InvalidID(t *testing.T) {
	handler := NewFileHandler(nil, nil)
	r := setupFileRouter(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/files/abc/versions", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFileHandler_DownloadVersion_InvalidFileID(t *testing.T) {
	handler := NewFileHandler(nil, nil)
	r := setupFileRouter(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/files/abc/versions/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFileHandler_DownloadVersion_InvalidVersion(t *testing.T) {
	handler := NewFileHandler(nil, nil)
	r := setupFileRouter(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/files/1/versions/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
