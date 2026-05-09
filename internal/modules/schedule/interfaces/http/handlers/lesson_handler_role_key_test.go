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
//
// Each subcase sends an input that short-circuits between the role gate
// and the use case (invalid JSON for body endpoints, invalid id for
// Delete) so that the assertion isolates the gate decision (403 vs not).
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

	post := func(method, path, body string) *http.Request {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		return req
	}

	for _, role := range allowed {
		t.Run("allowed_create_"+role, func(t *testing.T) {
			r := mount(role)
			w := httptest.NewRecorder()
			// `{}` parses but empty date_start fails time.Parse → 400 (post-gate).
			r.ServeHTTP(w, post(http.MethodPost, "/schedule/lessons", `{}`))
			assert.NotEqual(t, http.StatusForbidden, w.Code,
				"role %q must pass canModifySchedule when middleware writes 'role' key", role)
		})
		t.Run("allowed_update_"+role, func(t *testing.T) {
			r := mount(role)
			w := httptest.NewRecorder()
			// Invalid JSON forces ShouldBindJSON to 400 — short-circuits before nil usecase.
			r.ServeHTTP(w, post(http.MethodPut, "/schedule/lessons/1", `{bad`))
			assert.NotEqual(t, http.StatusForbidden, w.Code,
				"role %q must pass canModifySchedule on Update", role)
		})
		t.Run("allowed_delete_"+role, func(t *testing.T) {
			r := mount(role)
			w := httptest.NewRecorder()
			// Invalid id "xyz" short-circuits at getIDParam (post-gate) → 400.
			req := httptest.NewRequest(http.MethodDelete, "/schedule/lessons/xyz", nil)
			r.ServeHTTP(w, req)
			assert.NotEqual(t, http.StatusForbidden, w.Code,
				"role %q must pass canModifySchedule on Delete", role)
		})
		t.Run("allowed_create_change_"+role, func(t *testing.T) {
			r := mount(role)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, post(http.MethodPost, "/schedule/changes", `{bad`))
			assert.NotEqual(t, http.StatusForbidden, w.Code,
				"role %q must pass canModifySchedule on CreateChange", role)
		})
	}

	for _, role := range denied {
		t.Run("denied_create_"+role, func(t *testing.T) {
			r := mount(role)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, post(http.MethodPost, "/schedule/lessons", `{}`))
			assert.Equal(t, http.StatusForbidden, w.Code,
				"role %q must be denied by canModifySchedule", role)
		})
	}
}
