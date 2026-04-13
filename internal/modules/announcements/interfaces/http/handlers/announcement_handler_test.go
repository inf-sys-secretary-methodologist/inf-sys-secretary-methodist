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

func setupAnnouncementRouter(handler *AnnouncementHandler) *gin.Engine {
	r := gin.New()
	r.POST("/announcements", handler.Create)
	r.GET("/announcements/:id", handler.GetByID)
	r.PUT("/announcements/:id", handler.Update)
	r.DELETE("/announcements/:id", handler.Delete)
	r.GET("/announcements", handler.List)
	r.GET("/announcements/published", handler.GetPublished)
	r.GET("/announcements/pinned", handler.GetPinned)
	r.GET("/announcements/recent", handler.GetRecent)
	r.POST("/announcements/:id/publish", handler.Publish)
	r.POST("/announcements/:id/unpublish", handler.Unpublish)
	r.POST("/announcements/:id/archive", handler.Archive)
	return r
}

func TestAnnouncementHandler_Create_Unauthorized(t *testing.T) {
	handler := NewAnnouncementHandler(nil)
	r := setupAnnouncementRouter(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/announcements", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "user not authenticated")
}

func TestAnnouncementHandler_Create_InvalidUserIDType(t *testing.T) {
	handler := NewAnnouncementHandler(nil)
	r := gin.New()
	r.POST("/announcements", func(c *gin.Context) {
		c.Set("user_id", "not-an-int64")
		handler.Create(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/announcements", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid user ID type")
}

func TestAnnouncementHandler_Create_InvalidJSON(t *testing.T) {
	handler := NewAnnouncementHandler(nil)
	r := gin.New()
	r.POST("/announcements", func(c *gin.Context) {
		c.Set("user_id", int64(42))
		handler.Create(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/announcements", strings.NewReader(`{invalid`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAnnouncementHandler_GetByID_InvalidID(t *testing.T) {
	handler := NewAnnouncementHandler(nil)
	r := setupAnnouncementRouter(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/announcements/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid id")
}

func TestAnnouncementHandler_Update_Unauthorized(t *testing.T) {
	handler := NewAnnouncementHandler(nil)
	r := setupAnnouncementRouter(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/announcements/1", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAnnouncementHandler_Update_InvalidID(t *testing.T) {
	handler := NewAnnouncementHandler(nil)
	r := gin.New()
	r.PUT("/announcements/:id", func(c *gin.Context) {
		c.Set("user_id", int64(42))
		handler.Update(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/announcements/abc", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAnnouncementHandler_Update_InvalidJSON(t *testing.T) {
	handler := NewAnnouncementHandler(nil)
	r := gin.New()
	r.PUT("/announcements/:id", func(c *gin.Context) {
		c.Set("user_id", int64(42))
		handler.Update(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/announcements/1", strings.NewReader(`{bad`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAnnouncementHandler_Delete_Unauthorized(t *testing.T) {
	handler := NewAnnouncementHandler(nil)
	r := setupAnnouncementRouter(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/announcements/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAnnouncementHandler_Delete_InvalidID(t *testing.T) {
	handler := NewAnnouncementHandler(nil)
	r := gin.New()
	r.DELETE("/announcements/:id", func(c *gin.Context) {
		c.Set("user_id", int64(42))
		handler.Delete(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/announcements/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAnnouncementHandler_Publish_Unauthorized(t *testing.T) {
	handler := NewAnnouncementHandler(nil)
	r := setupAnnouncementRouter(handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/announcements/1/publish", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAnnouncementHandler_Publish_InvalidID(t *testing.T) {
	handler := NewAnnouncementHandler(nil)
	r := gin.New()
	r.POST("/announcements/:id/publish", func(c *gin.Context) {
		c.Set("user_id", int64(42))
		handler.Publish(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/announcements/abc/publish", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAnnouncementHandler_IsAdmin(t *testing.T) {
	handler := NewAnnouncementHandler(nil)

	tests := []struct {
		name     string
		role     interface{}
		exists   bool
		expected bool
	}{
		{"admin role", "admin", true, true},
		{"non-admin role", "user", true, false},
		{"no role", nil, false, false},
		{"invalid type", 123, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			if tt.exists {
				c.Set("user_role", tt.role)
			}
			assert.Equal(t, tt.expected, handler.isAdmin(c))
		})
	}
}
