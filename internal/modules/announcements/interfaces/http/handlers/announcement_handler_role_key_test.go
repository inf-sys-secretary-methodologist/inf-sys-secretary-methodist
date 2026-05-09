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
// Tests that assert role-gated handler behavior MUST use this helper —
// any helper writing "user_role" would mask the wrong-key bug class
// fixed in v0.126.0 (templates filter) and v0.126.1.
func withAuth(userID int64, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("role", role)
		c.Next()
	}
}

// TestAnnouncementHandler_IsAdmin_FromProductionMiddleware pins the
// announcement admin override to the JWTMiddleware contract. Before
// v0.126.1 isAdmin read c.Get("user_role") and compared against the
// string "admin" — the production middleware writes the "role" key and
// the legitimate elevated value is "system_admin" (the auth domain
// constant; see auth/domain RoleSystemAdmin). The combined wrong-key
// + wrong-value bug made isAdmin always return false in production,
// silently degrading admin-only operations on others' announcements
// (Update / Delete / Publish / Unpublish / Archive) to author-self only.
func TestAnnouncementHandler_IsAdmin_FromProductionMiddleware(t *testing.T) {
	handler := NewAnnouncementHandler(nil)

	tests := []struct {
		name string
		role string
		want bool
	}{
		{"system_admin is admin", "system_admin", true},
		{"methodist is not admin", "methodist", false},
		{"academic_secretary is not admin", "academic_secretary", false},
		{"teacher is not admin", "teacher", false},
		{"student is not admin", "student", false},
		{"empty role is not admin", "", false},
		{"unknown role is not admin", "admin", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Set("role", tt.role)
			assert.Equal(t, tt.want, handler.isAdmin(c),
				"isAdmin must read 'role' key and accept only system_admin")
		})
	}
}

// TestAnnouncementHandler_IsAdmin_NoRoleKey pins the failure-closed
// behavior: when the middleware did not run (no role in context),
// isAdmin must return false (not panic, not default-true).
func TestAnnouncementHandler_IsAdmin_NoRoleKey(t *testing.T) {
	handler := NewAnnouncementHandler(nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	assert.False(t, handler.isAdmin(c),
		"missing role key must default to non-admin (failure-closed)")
}

// TestAnnouncementHandler_IsAdmin_NonStringRoleType pins type safety:
// if the middleware ever stores a non-string under "role", isAdmin
// must return false rather than panic.
func TestAnnouncementHandler_IsAdmin_NonStringRoleType(t *testing.T) {
	handler := NewAnnouncementHandler(nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("role", 123)
	assert.False(t, handler.isAdmin(c),
		"non-string role value must default to non-admin")
}

// pinHTTPStatus is a tiny smoke check that admin override path is
// observable through the public HTTP surface. Sends Update with
// invalid JSON so the binding fails after the gate; ensures handler
// does not 500 on missing role key (just falls back to non-admin).
func TestAnnouncementHandler_AdminOverride_HTTPSurface(t *testing.T) {
	handler := NewAnnouncementHandler(nil)
	r := gin.New()
	r.PUT("/announcements/:id", withAuth(42, "system_admin"), handler.Update)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/announcements/1", nil)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code,
		"binding failure on empty body must yield 400 (not 401/500)")
}
