package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// withAuth mirrors the production JWTMiddleware contract:
// it writes the role into the "role" context key (not "user_role").
// Tests that exercise role-gated handler behaviour MUST use this
// helper — any helper writing "user_role" would mask the wrong-key
// bug class fixed in v0.126.0 (templates filter) and v0.126.1.
func withAuth(userID int64, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("role", role)
		c.Next()
	}
}

// TestAvatarHandler_Upload_RoleKey_AdminOverride pins the avatar
// upload admin override to the production JWTMiddleware contract.
// Before v0.126.1 the handler read c.Get("user_role"); the production
// middleware writes c.Set("role", ...). Result: the admin override on
// uploading another user's avatar silently 403'd in production for
// system_admin even though unit tests passed with a mirroring helper.
//
// All subcases target user 2 while authenticating as user 1, so the
// gate at avatar_handler.go:75 is exercised. Bodies are intentionally
// empty so the handler short-circuits at FormFile (post-gate, 400).
func TestAvatarHandler_Upload_RoleKey_AdminOverride(t *testing.T) {
	allowed := []string{"system_admin"}
	denied := []string{"methodist", "academic_secretary", "teacher", "student", ""}

	mount := func(role string) *gin.Engine {
		r := gin.New()
		h := &AvatarHandler{}
		r.POST("/users/:id/avatar", withAuth(1, role), h.Upload)
		return r
	}

	for _, role := range allowed {
		t.Run("allowed_"+role, func(t *testing.T) {
			r := mount(role)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/users/2/avatar", nil)
			r.ServeHTTP(w, req)
			assert.NotEqual(t, http.StatusForbidden, w.Code,
				"role %q must pass admin override on Upload when middleware writes 'role' key",
				role)
		})
	}

	for _, role := range denied {
		t.Run("denied_"+role, func(t *testing.T) {
			r := mount(role)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/users/2/avatar", nil)
			r.ServeHTTP(w, req)
			assert.Equal(t, http.StatusForbidden, w.Code,
				"role %q must be denied admin override on Upload", role)
		})
	}
}

// TestAvatarHandler_Delete_RoleKey_AdminOverride mirrors the Upload
// pin for the Delete handler, which has the same wrong-key bug at
// avatar_handler.go:200.
//
// For the allowed (system_admin) case the gate passes and the handler
// proceeds to call userUseCase.GetUser on a nil dependency — that
// panic is the observable signal that the gate let the request through.
// For denied roles the gate short-circuits at 403 before the panic.
func TestAvatarHandler_Delete_RoleKey_AdminOverride(t *testing.T) {
	denied := []string{"methodist", "academic_secretary", "teacher", "student", ""}

	mount := func(role string) *gin.Engine {
		r := gin.New()
		h := &AvatarHandler{}
		r.DELETE("/users/:id/avatar", withAuth(1, role), h.Delete)
		return r
	}

	t.Run("allowed_system_admin", func(t *testing.T) {
		r := mount("system_admin")
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/users/2/avatar", nil)
		assert.Panics(t, func() { r.ServeHTTP(w, req) },
			"system_admin must pass admin override on Delete (panics at nil userUseCase)")
	})

	for _, role := range denied {
		t.Run("denied_"+role, func(t *testing.T) {
			r := mount(role)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/users/2/avatar", nil)
			r.ServeHTTP(w, req)
			assert.Equal(t, http.StatusForbidden, w.Code,
				"role %q must be denied admin override on Delete", role)
		})
	}
}
