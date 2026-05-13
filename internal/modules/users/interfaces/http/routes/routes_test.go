package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	authDomain "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
	authMW "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/interfaces/http/middleware"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/interfaces/http/handlers"
)

// withAuth mirrors the production JWTMiddleware contract — writes
// "user_id" + "role" into gin context keys (same keys auth_middleware
// reads). v0.126.0 templates-filter incident class.
func withAuth(uid int64, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if uid != 0 {
			c.Set("user_id", uid)
		}
		if role != "" {
			c.Set("role", role)
		}
		c.Next()
	}
}

// newTestEngine builds a production-shaped router: a protected group
// with withAuth + RegisterUserRoutes wired against the same admin
// middleware production uses. Handlers carry nil usecases — denied
// requests short-circuit at the gate (Forbidden); allowed requests
// reach the handler and either panic on nil receiver or return 400
// on empty body. gin.Recovery() converts the panic to 500 so tests
// can assert via status codes uniformly.
func newTestEngine(uid int64, role string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())
	api := r.Group("/api")
	api.Use(withAuth(uid, role))
	adminMW := authMW.RequireRole(string(authDomain.RoleSystemAdmin))
	usersGroup := api.Group("/users")
	RegisterUserRoutes(usersGroup, adminMW, &handlers.UserHandler{}, &handlers.AvatarHandler{})
	return r
}

type endpointCase struct {
	name   string
	method string
	path   string
}

var writeEndpoints = []endpointCase{
	{"updateProfile", http.MethodPut, "/api/users/2/profile"},
	{"updateRole", http.MethodPut, "/api/users/2/role"},
	{"updateStatus", http.MethodPut, "/api/users/2/status"},
	{"deleteUser", http.MethodDelete, "/api/users/2"},
	{"bulkDepartment", http.MethodPost, "/api/users/bulk/department"},
	{"bulkPosition", http.MethodPost, "/api/users/bulk/position"},
	{"uploadAvatar", http.MethodPost, "/api/users/2/avatar"},
	{"deleteAvatar", http.MethodDelete, "/api/users/2/avatar"},
}

var readEndpoints = []endpointCase{
	{"list", http.MethodGet, "/api/users"},
	{"getById", http.MethodGet, "/api/users/2"},
	{"byDepartment", http.MethodGet, "/api/users/by-department/3"},
	{"byPosition", http.MethodGet, "/api/users/by-position/3"},
	{"getAvatar", http.MethodGet, "/api/users/2/avatar"},
}

// TestRegisterUserRoutes_WriteEndpoints_DeniedForNonAdmin pins the
// security invariant: only system_admin may invoke destructive user
// management endpoints. Pre-fix any authenticated user could
// DELETE /api/users/:id or PUT /:id/role — a TIER 0 privilege
// escalation.
func TestRegisterUserRoutes_WriteEndpoints_DeniedForNonAdmin(t *testing.T) {
	deniedRoles := []string{"methodist", "academic_secretary", "teacher", "student"}

	for _, role := range deniedRoles {
		for _, tc := range writeEndpoints {
			t.Run(role+"_"+tc.name, func(t *testing.T) {
				r := newTestEngine(1, role)
				w := httptest.NewRecorder()
				req := httptest.NewRequest(tc.method, tc.path, nil)
				r.ServeHTTP(w, req)
				assert.Equal(t, http.StatusForbidden, w.Code,
					"role %q must be denied %s %s — admin gate missing",
					role, tc.method, tc.path)
			})
		}
	}
}

// TestRegisterUserRoutes_WriteEndpoints_AllowedForAdmin verifies the
// gate lets system_admin through. Allowed requests do not return 403;
// they panic at the nil usecase / fail body parsing (signaling the
// handler was reached). gin.Recovery() converts panics to 500.
func TestRegisterUserRoutes_WriteEndpoints_AllowedForAdmin(t *testing.T) {
	for _, tc := range writeEndpoints {
		t.Run("system_admin_"+tc.name, func(t *testing.T) {
			r := newTestEngine(1, "system_admin")
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.path, nil)
			r.ServeHTTP(w, req)
			assert.NotEqual(t, http.StatusForbidden, w.Code,
				"system_admin must pass admin gate for %s %s",
				tc.method, tc.path)
		})
	}
}

// TestRegisterUserRoutes_ReadEndpoints_PermissiveForAllRoles confirms
// that read-only lookups stay reachable for any authenticated role.
// Cross-module consumers (documents author lookup, curriculum methodist
// resolver) depend on this surface.
func TestRegisterUserRoutes_ReadEndpoints_PermissiveForAllRoles(t *testing.T) {
	allRoles := []string{"methodist", "academic_secretary", "teacher", "student", "system_admin"}

	for _, role := range allRoles {
		for _, tc := range readEndpoints {
			t.Run(role+"_"+tc.name, func(t *testing.T) {
				r := newTestEngine(1, role)
				w := httptest.NewRecorder()
				req := httptest.NewRequest(tc.method, tc.path, nil)
				r.ServeHTTP(w, req)
				assert.NotEqual(t, http.StatusForbidden, w.Code,
					"role %q must reach read endpoint %s %s — read stays permissive",
					role, tc.method, tc.path)
			})
		}
	}
}

// TestRegisterUserRoutes_MissingRole_AllEndpoints_403 covers the
// stripped-context case: no role key set. RequireRole denies; read
// endpoints stay reachable (no role gate). Pre-fix the read path would
// remain reachable (same as today); the write path would also be
// reachable (the bug). After the split the write path returns 403.
func TestRegisterUserRoutes_MissingRole_WriteDenied(t *testing.T) {
	for _, tc := range writeEndpoints {
		t.Run("no_role_"+tc.name, func(t *testing.T) {
			r := newTestEngine(1, "")
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.path, nil)
			r.ServeHTTP(w, req)
			assert.Equal(t, http.StatusForbidden, w.Code,
				"missing role must be denied %s %s",
				tc.method, tc.path)
		})
	}
}
