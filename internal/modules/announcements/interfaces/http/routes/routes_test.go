package routes_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/interfaces/http/handlers"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/interfaces/http/routes"
	authDomain "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
	authMW "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/interfaces/http/middleware"
)

// setAuthContext mounts a fake JWT-context middleware mirroring what
// production auth middleware injects (user_id + role).
func setAuthContext(role authDomain.RoleType) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", int64(1))
		c.Set("role", string(role))
		c.Next()
	}
}

// TestRegisterAnnouncementRoutes_StudentBlockedOnMutation pins
// v0.163.0 ADR-1 (#303 TIER 0): students must be denied on all
// mutation routes (POST/PUT/DELETE + publish/unpublish/archive +
// attachment upload/delete). Pre-fix the entire announcementsGroup
// lived under JWT-only middleware, so a student could publish
// admin-broadcasts с target_audience=admins.
func TestRegisterAnnouncementRoutes_StudentBlockedOnMutation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mutations := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/announcements"},
		{http.MethodPut, "/announcements/1"},
		{http.MethodDelete, "/announcements/1"},
		{http.MethodPost, "/announcements/1/publish"},
		{http.MethodPost, "/announcements/1/unpublish"},
		{http.MethodPost, "/announcements/1/archive"},
		{http.MethodPost, "/announcements/1/attachments"},
		{http.MethodDelete, "/announcements/1/attachments/2"},
	}

	for _, tc := range mutations {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			r := gin.New()
			grp := r.Group("/announcements")
			grp.Use(setAuthContext(authDomain.RoleStudent))
			routes.RegisterAnnouncementRoutes(grp, authMW.RequireNonStudent(), handlers.NewAnnouncementHandler(nil))

			w := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(`{}`))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusForbidden, w.Code,
				"student must be rejected by RequireNonStudent before reaching handler")
		})
	}
}
