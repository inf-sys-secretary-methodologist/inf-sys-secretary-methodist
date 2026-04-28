package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func withLessonAuth(handler gin.HandlerFunc, userID int64, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("user_role", role)
		handler(c)
	}
}

func setupLessonRouter(handler *LessonHandler, role string) *gin.Engine {
	r := gin.New()
	r.POST("/schedule/lessons", withLessonAuth(handler.Create, 1, role))
	r.GET("/schedule/lessons", withLessonAuth(handler.List, 1, role))
	r.GET("/schedule/lessons/timetable", withLessonAuth(handler.GetTimetable, 1, role))
	r.GET("/schedule/lessons/:id", withLessonAuth(handler.GetByID, 1, role))
	r.PUT("/schedule/lessons/:id", withLessonAuth(handler.Update, 1, role))
	r.DELETE("/schedule/lessons/:id", withLessonAuth(handler.Delete, 1, role))
	r.POST("/schedule/changes", withLessonAuth(handler.CreateChange, 1, role))
	return r
}

func TestLessonHandler_Create_Unauthorized(t *testing.T) {
	handler := NewLessonHandler(nil)
	r := gin.New()
	r.POST("/schedule/lessons", handler.Create)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/schedule/lessons", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLessonHandler_Create_ForbiddenForStudent(t *testing.T) {
	handler := NewLessonHandler(nil)
	r := setupLessonRouter(handler, "student")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/schedule/lessons", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestLessonHandler_Create_ForbiddenForTeacher(t *testing.T) {
	handler := NewLessonHandler(nil)
	r := setupLessonRouter(handler, "teacher")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/schedule/lessons", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestLessonHandler_Create_ForbiddenForMethodist(t *testing.T) {
	handler := NewLessonHandler(nil)
	r := setupLessonRouter(handler, "methodist")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/schedule/lessons", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestLessonHandler_Update_ForbiddenForStudent(t *testing.T) {
	handler := NewLessonHandler(nil)
	r := setupLessonRouter(handler, "student")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/schedule/lessons/1", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestLessonHandler_Delete_ForbiddenForTeacher(t *testing.T) {
	handler := NewLessonHandler(nil)
	r := setupLessonRouter(handler, "teacher")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/schedule/lessons/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestLessonHandler_CreateChange_ForbiddenForStudent(t *testing.T) {
	handler := NewLessonHandler(nil)
	r := setupLessonRouter(handler, "student")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/schedule/changes", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestLessonHandler_PermissionMatrix_Forbidden(t *testing.T) {
	handler := NewLessonHandler(nil)

	tests := []struct {
		name   string
		method string
		path   string
		role   string
	}{
		{"student cannot create", http.MethodPost, "/schedule/lessons", "student"},
		{"teacher cannot create", http.MethodPost, "/schedule/lessons", "teacher"},
		{"methodist cannot create", http.MethodPost, "/schedule/lessons", "methodist"},
		{"student cannot update", http.MethodPut, "/schedule/lessons/1", "student"},
		{"teacher cannot update", http.MethodPut, "/schedule/lessons/1", "teacher"},
		{"methodist cannot update", http.MethodPut, "/schedule/lessons/1", "methodist"},
		{"student cannot delete", http.MethodDelete, "/schedule/lessons/1", "student"},
		{"teacher cannot delete", http.MethodDelete, "/schedule/lessons/1", "teacher"},
		{"methodist cannot delete", http.MethodDelete, "/schedule/lessons/1", "methodist"},
		{"student cannot create change", http.MethodPost, "/schedule/changes", "student"},
		{"teacher cannot create change", http.MethodPost, "/schedule/changes", "teacher"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setupLessonRouter(handler, tt.role)
			w := httptest.NewRecorder()
			var body *strings.Reader
			if tt.method == http.MethodPost || tt.method == http.MethodPut {
				body = strings.NewReader(`{}`)
			} else {
				body = strings.NewReader("")
			}
			req := httptest.NewRequest(tt.method, tt.path, body)
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusForbidden, w.Code, "role=%s method=%s path=%s should be 403", tt.role, tt.method, tt.path)
		})
	}
}

func TestLessonHandler_PermissionMatrix_Allowed(t *testing.T) {
	handler := NewLessonHandler(nil)

	tests := []struct {
		name string
		role string
	}{
		{"admin passes guard", "system_admin"},
		{"secretary passes guard", "academic_secretary"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := setupLessonRouter(handler, tt.role)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/schedule/lessons", strings.NewReader(`{}`))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.NotEqual(t, http.StatusForbidden, w.Code, "role=%s should NOT get 403", tt.role)
		})
	}
}

func TestLessonHandler_Create_InvalidJSON(t *testing.T) {
	handler := NewLessonHandler(nil)
	r := setupLessonRouter(handler, "system_admin")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/schedule/lessons", strings.NewReader(`{bad json`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLessonHandler_GetByID_InvalidID(t *testing.T) {
	handler := NewLessonHandler(nil)
	r := setupLessonRouter(handler, "system_admin")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/schedule/lessons/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLessonHandler_Update_InvalidID(t *testing.T) {
	handler := NewLessonHandler(nil)
	r := setupLessonRouter(handler, "system_admin")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/schedule/lessons/xyz", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLessonHandler_Delete_InvalidID(t *testing.T) {
	handler := NewLessonHandler(nil)
	r := setupLessonRouter(handler, "system_admin")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/schedule/lessons/xyz", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
