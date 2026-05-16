package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	authDomain "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
	authMW "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/interfaces/http/middleware"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/branding/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/branding/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/branding/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/branding/interfaces/http/handlers"
)

// fakeRepo is a deterministic in-memory replacement for the PG repo —
// avoids a real database in the route integration test while still
// exercising the production handler chain (use case → repo → DTO).
type fakeRepo struct {
	mu       sync.Mutex
	settings *entities.BrandSettings
}

func (r *fakeRepo) Get(_ context.Context) (*entities.BrandSettings, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.settings, nil
}

func (r *fakeRepo) Update(_ context.Context, s *entities.BrandSettings) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.settings = s
	return nil
}

type fakeClock struct{ now time.Time }

func (c fakeClock) Now() time.Time { return c.now }

// withAuth mirrors production JWTMiddleware: writes user_id + role
// context keys so the downstream RequireRole admin gate can read them.
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

// newTestEngine assembles a production-shaped router:
//   - /api/public — no auth, hosts the public branding GET
//   - /api/admin — auth + RequireRole(system_admin), hosts the admin GET/PUT
//
// All wiring goes through RegisterBrandingRoutes so the test pins the
// extractor's contract (rather than re-mounting routes inline, which
// would tautologically pass even if the registrar were broken).
func newTestEngine(t *testing.T, uid int64, role string, repo usecases.BrandSettingsRepository) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())

	publicGroup := r.Group("/api/public")
	apiGroup := r.Group("/api")
	apiGroup.Use(withAuth(uid, role))
	adminGroup := apiGroup.Group("/admin")
	adminGroup.Use(authMW.RequireRole(string(authDomain.RoleSystemAdmin)))

	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	getUC := usecases.NewGetBrandingUseCase(repo)
	updateUC := usecases.NewUpdateBrandingUseCase(repo, fakeClock{now: now}, nil)
	adminH := handlers.NewAdminBrandingHandler(getUC, updateUC)
	publicH := handlers.NewPublicBrandingHandler(getUC)

	RegisterBrandingRoutes(adminGroup, publicGroup, adminH, publicH)
	return r
}

func seedRepo(t *testing.T) *fakeRepo {
	t.Helper()
	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	bs, err := entities.NewBrandSettings(
		"Brand", "Tag",
		"https://example.com/logo.png", "https://example.com/favicon.ico",
		"#112233", "#445566",
		now,
	)
	require.NoError(t, err)
	return &fakeRepo{settings: bs}
}

// TestRegisterBrandingRoutes_AdminGet_AllowedForSystemAdmin pins the
// admin GET surface. system_admin reaches the handler and gets the
// seeded projection.
func TestRegisterBrandingRoutes_AdminGet_AllowedForSystemAdmin(t *testing.T) {
	repo := seedRepo(t)
	r := newTestEngine(t, 42, "system_admin", repo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/admin/branding", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, "system_admin must reach admin GET — got body %s", w.Body.String())
	var env struct {
		Success bool                 `json:"success"`
		Data    dto.BrandSettingsDTO `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &env))
	assert.True(t, env.Success)
	assert.Equal(t, "Brand", env.Data.AppName)
}

// TestRegisterBrandingRoutes_AdminPut_AllowedForSystemAdmin verifies
// the admin PUT surface. system_admin reaches the handler with a valid
// body and gets 200.
func TestRegisterBrandingRoutes_AdminPut_AllowedForSystemAdmin(t *testing.T) {
	repo := seedRepo(t)
	r := newTestEngine(t, 42, "system_admin", repo)

	body, _ := json.Marshal(dto.UpdateBrandSettingsRequest{
		AppName:        "Updated",
		Tagline:        "Tag2",
		LogoURL:        "https://example.com/new-logo.png",
		FaviconURL:     "https://example.com/new-favicon.ico",
		PrimaryColor:   "#abcdef",
		SecondaryColor: "#012345",
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/admin/branding", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, "system_admin must reach admin PUT — got body %s", w.Body.String())
}

// TestRegisterBrandingRoutes_AdminEndpoints_DeniedForNonAdmin pins the
// security invariant: admin endpoints reject every non-system_admin
// caller with 403. The admin gate sits outside the registrar and the
// registrar must mount under the gated group so this works.
func TestRegisterBrandingRoutes_AdminEndpoints_DeniedForNonAdmin(t *testing.T) {
	deniedRoles := []string{"methodist", "academic_secretary", "teacher", "student"}
	endpoints := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/admin/branding"},
		{http.MethodPut, "/api/admin/branding"},
	}

	for _, role := range deniedRoles {
		for _, ep := range endpoints {
			t.Run(role+"_"+ep.method, func(t *testing.T) {
				repo := seedRepo(t)
				r := newTestEngine(t, 1, role, repo)
				w := httptest.NewRecorder()
				req := httptest.NewRequest(ep.method, ep.path, nil)
				r.ServeHTTP(w, req)
				assert.Equal(t, http.StatusForbidden, w.Code,
					"role %q must be denied %s %s — admin gate must wrap the mount",
					role, ep.method, ep.path)
			})
		}
	}
}

// TestRegisterBrandingRoutes_PublicGet_NoAuthRequired verifies the
// public surface is reachable without any auth context. The login page
// consumes this before the user has a JWT.
func TestRegisterBrandingRoutes_PublicGet_NoAuthRequired(t *testing.T) {
	repo := seedRepo(t)
	// uid=0 + role="" — withAuth writes nothing, so the public group
	// (which has no auth middleware) reaches the handler unconditionally.
	r := newTestEngine(t, 0, "", repo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/public/branding", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, "public GET must reach handler without auth — got body %s", w.Body.String())
	var env struct {
		Success bool                 `json:"success"`
		Data    dto.BrandSettingsDTO `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &env))
	assert.True(t, env.Success)
	assert.Equal(t, "Brand", env.Data.AppName)
}

// TestRegisterBrandingRoutes_OptionsCORS_RespondsNoContent pins the
// CORS preflight handlers. Mirror к existing admin/public surface
// CORS pattern (every admin/* and public/* endpoint in main.go has
// an OPTIONS handler). Browsers fail to preflight without them.
func TestRegisterBrandingRoutes_OptionsCORS_RespondsNoContent(t *testing.T) {
	repo := seedRepo(t)
	cases := []struct {
		name string
		uid  int64
		role string
		path string
	}{
		{"adminOptions", 42, "system_admin", "/api/admin/branding"},
		{"publicOptions", 0, "", "/api/public/branding"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := newTestEngine(t, tc.uid, tc.role, repo)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodOptions, tc.path, nil)
			r.ServeHTTP(w, req)
			assert.Equal(t, http.StatusNoContent, w.Code,
				"OPTIONS %s must respond 204 — CORS preflight handler missing", tc.path)
		})
	}
}
