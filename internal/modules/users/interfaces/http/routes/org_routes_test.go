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

// newOrgTestEngine builds a production-shaped router with departments
// + positions routes mounted under withAuth + RegisterXxxRoutes, using
// the same RequireRole(system_admin) middleware production uses.
// Handler receivers are nil — denied requests short-circuit at the
// admin gate (Forbidden); allowed requests reach the handler and
// either panic on nil receiver or fail body parsing. gin.Recovery()
// converts panics to 500 so tests can assert status codes uniformly.
func newOrgTestEngine(uid int64, role string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())
	api := r.Group("/api")
	api.Use(withAuth(uid, role))
	adminMW := authMW.RequireRole(string(authDomain.RoleSystemAdmin))
	deptGroup := api.Group("/departments")
	RegisterDepartmentRoutes(deptGroup, adminMW, &handlers.DepartmentHandler{})
	posGroup := api.Group("/positions")
	RegisterPositionRoutes(posGroup, adminMW, &handlers.PositionHandler{})
	return r
}

// Admin-only destructive endpoints across organizational structure.
// Pre-v0.160.0 these were ALL exposed to any authenticated caller —
// v0.133.0 admin-gate split applied only to /users.
var orgWriteEndpoints = []endpointCase{
	{"deptCreate", http.MethodPost, "/api/departments"},
	{"deptUpdate", http.MethodPut, "/api/departments/1"},
	{"deptDelete", http.MethodDelete, "/api/departments/1"},
	{"posCreate", http.MethodPost, "/api/positions"},
	{"posUpdate", http.MethodPut, "/api/positions/1"},
	{"posDelete", http.MethodDelete, "/api/positions/1"},
}

// Read endpoints stay permissive for any authenticated role — frontend
// dropdowns and cross-module resolvers depend on them.
var orgReadEndpoints = []endpointCase{
	{"deptList", http.MethodGet, "/api/departments"},
	{"deptGetByID", http.MethodGet, "/api/departments/1"},
	{"deptChildren", http.MethodGet, "/api/departments/1/children"},
	{"posList", http.MethodGet, "/api/positions"},
	{"posGetByID", http.MethodGet, "/api/positions/1"},
}

// TestRegisterOrgRoutes_WriteEndpoints_DeniedForNonAdmin pins #283
// ADR-2 TIER 0: only system_admin may mutate departments/positions.
// Pre-fix any authenticated user — including students — could POST a
// new department, freely renaming the organizational structure.
func TestRegisterOrgRoutes_WriteEndpoints_DeniedForNonAdmin(t *testing.T) {
	deniedRoles := []string{"methodist", "academic_secretary", "teacher", "student"}

	for _, role := range deniedRoles {
		for _, tc := range orgWriteEndpoints {
			t.Run(role+"_"+tc.name, func(t *testing.T) {
				r := newOrgTestEngine(1, role)
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

// TestRegisterOrgRoutes_WriteEndpoints_AllowedForAdmin verifies the
// gate lets system_admin through. Allowed requests do not return 403;
// they panic at the nil handler / fail body parsing (signaling the
// handler was reached). gin.Recovery() converts panics to 500.
func TestRegisterOrgRoutes_WriteEndpoints_AllowedForAdmin(t *testing.T) {
	for _, tc := range orgWriteEndpoints {
		t.Run("system_admin_"+tc.name, func(t *testing.T) {
			r := newOrgTestEngine(1, "system_admin")
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.path, nil)
			r.ServeHTTP(w, req)
			assert.NotEqual(t, http.StatusForbidden, w.Code,
				"system_admin must pass admin gate for %s %s",
				tc.method, tc.path)
		})
	}
}

// TestRegisterOrgRoutes_ReadEndpoints_PermissiveForAllRoles confirms
// read access stays open for any authenticated role.
func TestRegisterOrgRoutes_ReadEndpoints_PermissiveForAllRoles(t *testing.T) {
	allRoles := []string{"methodist", "academic_secretary", "teacher", "student", "system_admin"}

	for _, role := range allRoles {
		for _, tc := range orgReadEndpoints {
			t.Run(role+"_"+tc.name, func(t *testing.T) {
				r := newOrgTestEngine(1, role)
				w := httptest.NewRecorder()
				req := httptest.NewRequest(tc.method, tc.path, nil)
				r.ServeHTTP(w, req)
				assert.NotEqual(t, http.StatusForbidden, w.Code,
					"role %q must reach handler for %s %s",
					role, tc.method, tc.path)
			})
		}
	}
}
