package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// stubRole installs role into gin context, simulating successful JWT validation.
func stubRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("role", role)
		c.Next()
	}
}

// TestRequireNonStudent_BlocksStudent verifies that the convenience helper
// RequireNonStudent() denies access to users whose role is "student".
// AUDIT_REPORT critical item #1: students could call POST /api/documents,
// GET /api/reports, GET /api/analytics — all should be 403.
func TestRequireNonStudent_BlocksStudent(t *testing.T) {
	router := gin.New()
	router.Use(stubRole("student"))
	router.Use(RequireNonStudent())
	router.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// TestRequireNonStudent_AllowsOtherRoles verifies that all four non-student
// roles pass through the guard.
func TestRequireNonStudent_AllowsOtherRoles(t *testing.T) {
	roles := []string{"system_admin", "methodist", "academic_secretary", "teacher"}

	for _, role := range roles {
		t.Run(role, func(t *testing.T) {
			router := gin.New()
			router.Use(stubRole(role))
			router.Use(RequireNonStudent())
			router.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/x", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code, "role=%s must pass", role)
		})
	}
}

// TestRequireNonStudent_BlocksMissingRole verifies that a request without
// any role context (no JWTMiddleware ran) is rejected — defence in depth.
func TestRequireNonStudent_BlocksMissingRole(t *testing.T) {
	router := gin.New()
	router.Use(RequireNonStudent())
	router.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
