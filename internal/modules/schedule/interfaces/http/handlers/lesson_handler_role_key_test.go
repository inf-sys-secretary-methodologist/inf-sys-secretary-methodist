package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// withAuth mirrors the production JWTMiddleware contract:
// it writes the role into the "role" context key (not "user_role").
// Tests that assert role-gated handler behaviour MUST use this helper —
// any helper writing "user_role" would mask the wrong-key bug class
// fixed in v0.126.0 (templates filter) and v0.126.1.
func withAuth(handler gin.HandlerFunc, userID int64, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("role", role)
		handler(c)
	}
}

// TestLessonHandler_RoleKey_FromProductionMiddleware pins the schedule
// permission gate to the JWTMiddleware contract. Before v0.126.1 the
// handler read c.Get("user_role") while the production middleware writes
// c.Set("role", ...) — every schedule write op silently 403'd in prod
// even though unit tests passed with a mirroring helper.
func TestLessonHandler_RoleKey_FromProductionMiddleware(t *testing.T) {
	handler := NewLessonHandler(nil)

	allowed := []string{"system_admin", "academic_secretary"}
	denied := []string{"student", "teacher", "methodist"}

	mount := func(role string) *gin.Engine {
		r := gin.New()
		r.POST("/schedule/lessons", withAuth(handler.Create, 1, role))
		r.PUT("/schedule/lessons/:id", withAuth(handler.Update, 1, role))
		r.DELETE("/schedule/lessons/:id", withAuth(handler.Delete, 1, role))
		r.POST("/schedule/changes", withAuth(handler.CreateChange, 1, role))
		return r
	}

	for _, role := range allowed {
		t.Run("allowed_create_"+role, func(t *testing.T) {
			r := mount(role)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/schedule/lessons",
				strings.NewReader(`{}`))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
			assert.NotEqual(t, http.StatusForbidden, w.Code,
				"role %q must pass canModifySchedule when middleware writes 'role' key", role)
		})
		t.Run("allowed_update_"+role, func(t *testing.T) {
			r := mount(role)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPut, "/schedule/lessons/1",
				strings.NewReader(`{}`))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
			assert.NotEqual(t, http.StatusForbidden, w.Code,
				"role %q must pass canModifySchedule on Update", role)
		})
		t.Run("allowed_delete_"+role, func(t *testing.T) {
			r := mount(role)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/schedule/lessons/1", nil)
			r.ServeHTTP(w, req)
			assert.NotEqual(t, http.StatusForbidden, w.Code,
				"role %q must pass canModifySchedule on Delete", role)
		})
		t.Run("allowed_create_change_"+role, func(t *testing.T) {
			r := mount(role)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/schedule/changes",
				strings.NewReader(`{}`))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
			assert.NotEqual(t, http.StatusForbidden, w.Code,
				"role %q must pass canModifySchedule on CreateChange", role)
		})
	}

	for _, role := range denied {
		t.Run("denied_create_"+role, func(t *testing.T) {
			r := mount(role)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/schedule/lessons",
				strings.NewReader(`{}`))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
			assert.Equal(t, http.StatusForbidden, w.Code,
				"role %q must be denied by canModifySchedule", role)
		})
	}
}
