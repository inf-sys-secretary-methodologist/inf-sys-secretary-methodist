package integration

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	authMiddleware "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/interfaces/http/middleware"
	integrationHttp "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/interfaces/http"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/config"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

// stubRoleMiddleware injects role into gin context to simulate post-JWT state.
func stubRoleMiddleware(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("role", role)
		c.Set("user_id", int64(42))
		c.Next()
	}
}

func newTestModule() *Module {
	return &Module{
		config:          &config.IntegrationConfig{Enabled: true},
		logger:          logging.NewLogger("error"),
		syncHandler:     integrationHttp.NewSyncHandler(nil),
		employeeHandler: integrationHttp.NewEmployeeHandler(nil),
		studentHandler:  integrationHttp.NewStudentHandler(nil),
		conflictHandler: integrationHttp.NewConflictHandler(nil),
	}
}

// TestRegisterRoutes_BlocksNonAdmin verifies that any role other than
// system_admin gets 403 on /api/integration/* endpoints. Without the admin
// guard wired into RegisterRoutes, sync, employee, student, conflict routes
// are accessible to any authenticated user — a critical security flaw.
func TestRegisterRoutes_BlocksNonAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		role string
	}{
		{"student"},
		{"teacher"},
		{"methodist"},
		{"academic_secretary"},
	}

	endpoints := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/integration/sync/start"},
		{http.MethodGet, "/api/integration/sync/logs"},
		{http.MethodGet, "/api/integration/sync/status"},
		{http.MethodGet, "/api/integration/employees"},
		{http.MethodGet, "/api/integration/students"},
		{http.MethodGet, "/api/integration/conflicts"},
	}

	for _, tc := range cases {
		for _, ep := range endpoints {
			t.Run(tc.role+" "+ep.method+" "+ep.path, func(t *testing.T) {
				router := gin.New()
				router.Use(stubRoleMiddleware(tc.role))
				apiGroup := router.Group("/api")

				m := newTestModule()
				m.RegisterRoutes(apiGroup, authMiddleware.RequireRole("system_admin"))

				req := httptest.NewRequest(ep.method, ep.path, bytes.NewReader([]byte("{}")))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusForbidden, w.Code,
					"role=%s must be forbidden on %s %s", tc.role, ep.method, ep.path)
			})
		}
	}
}

// TestRegisterRoutes_AllowsAdmin verifies that system_admin role passes through
// the admin guard. We don't care what handler returns afterwards (likely panic
// from nil usecase) — only that middleware does not abort with 403.
func TestRegisterRoutes_AllowsAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	// Recover panics so handler-level nil deref does not fail the test.
	router.Use(gin.CustomRecovery(func(c *gin.Context, _ interface{}) {
		c.AbortWithStatus(http.StatusInternalServerError)
	}))
	router.Use(stubRoleMiddleware("system_admin"))
	apiGroup := router.Group("/api")

	m := newTestModule()
	m.RegisterRoutes(apiGroup, authMiddleware.RequireRole("system_admin"))

	req := httptest.NewRequest(http.MethodGet, "/api/integration/sync/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusForbidden, w.Code,
		"system_admin must not be forbidden by admin guard")
}
