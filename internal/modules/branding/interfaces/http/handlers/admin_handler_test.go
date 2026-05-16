package handlers

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

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/branding/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/branding/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/branding/domain/entities"
)

// fakeRepo is a deterministic in-memory replacement for
// BrandSettingsRepository. Production wiring uses the PG repo.
type fakeRepo struct {
	mu       sync.Mutex
	settings *entities.BrandSettings
	updates  int
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
	r.updates++
	return nil
}

// fakeClock returns a fixed time so audit / updated_at assertions
// stay stable across runs.
type fakeClock struct{ now time.Time }

func (c fakeClock) Now() time.Time { return c.now }

// auditCall records one invocation of LogAuditEvent.
type auditCall struct {
	Action   string
	Resource string
	Fields   map[string]any
}

// fakeAudit is the spy sink mirroring AuditSink. Records every call.
type fakeAudit struct {
	mu    sync.Mutex
	calls []auditCall
}

func (a *fakeAudit) LogAuditEvent(_ context.Context, action, resource string, fields map[string]any) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.calls = append(a.calls, auditCall{Action: action, Resource: resource, Fields: fields})
}

// withAuth mirrors production JWT middleware contract — sets
// user_id + role context keys read by downstream RequireRole.
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

func newTestEngine(t *testing.T, repo usecases.BrandSettingsRepository, clock usecases.Clock, audit usecases.AuditSink) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	getUC := usecases.NewGetBrandingUseCase(repo)
	updateUC := usecases.NewUpdateBrandingUseCase(repo, clock, audit)
	h := NewAdminBrandingHandler(getUC, updateUC)
	r := gin.New()
	r.Use(gin.Recovery())
	api := r.Group("/api")
	api.Use(withAuth(42, "system_admin"))
	api.GET("/admin/branding", h.GetBranding)
	api.PUT("/admin/branding", h.UpdateBranding)
	return r
}

type envelope struct {
	Success bool                 `json:"success"`
	Data    dto.BrandSettingsDTO `json:"data"`
	Error   *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func TestAdminBrandingHandler_Get_ReturnsCurrentSettings(t *testing.T) {
	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	bs, err := entities.NewBrandSettings(
		"Existing Brand", "Existing tagline",
		"https://example.com/logo.png", "https://example.com/favicon.ico",
		"#112233", "#445566", now,
	)
	require.NoError(t, err)
	repo := &fakeRepo{settings: bs}

	r := newTestEngine(t, repo, fakeClock{now: now}, &fakeAudit{})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/admin/branding", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body envelope
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.True(t, body.Success)
	assert.Equal(t, "Existing Brand", body.Data.AppName)
	assert.Equal(t, "Existing tagline", body.Data.Tagline)
	assert.Equal(t, "https://example.com/logo.png", body.Data.LogoURL)
	assert.Equal(t, "https://example.com/favicon.ico", body.Data.FaviconURL)
	assert.Equal(t, "#112233", body.Data.PrimaryColor)
	assert.Equal(t, "#445566", body.Data.SecondaryColor)
	assert.True(t, body.Data.UpdatedAt.Equal(now))
}

func TestAdminBrandingHandler_Update_HappyPath_PersistsAndAudits(t *testing.T) {
	seedNow := time.Date(2026, 5, 14, 9, 0, 0, 0, time.UTC)
	bs, err := entities.NewBrandSettings("Old", "", "", "", "", "", seedNow)
	require.NoError(t, err)
	repo := &fakeRepo{settings: bs}

	updateNow := time.Date(2026, 5, 14, 12, 30, 0, 0, time.UTC)
	audit := &fakeAudit{}
	r := newTestEngine(t, repo, fakeClock{now: updateNow}, audit)

	body, _ := json.Marshal(dto.UpdateBrandSettingsRequest{
		AppName:        "New Brand",
		Tagline:        "Welcome",
		LogoURL:        "https://new.example/logo.png",
		FaviconURL:     "https://new.example/favicon.ico",
		PrimaryColor:   "#abcdef",
		SecondaryColor: "#012345",
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/admin/branding", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())
	var env envelope
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &env))
	assert.True(t, env.Success)
	assert.Equal(t, "New Brand", env.Data.AppName)
	assert.Equal(t, "Welcome", env.Data.Tagline)
	assert.Equal(t, "#abcdef", env.Data.PrimaryColor)
	assert.True(t, env.Data.UpdatedAt.Equal(updateNow), "updatedAt must come from injected clock")

	assert.Equal(t, 1, repo.updates, "repo.Update called exactly once")

	require.Len(t, audit.calls, 1, "exactly one audit event emitted")
	assert.Equal(t, "brand.updated", audit.calls[0].Action)
	assert.Equal(t, "brand", audit.calls[0].Resource)
	assert.NotNil(t, audit.calls[0].Fields)
	assert.Equal(t, int64(42), audit.calls[0].Fields["actor_user_id"],
		"actor user id from JWT context surfaces in audit payload")
}

func TestAdminBrandingHandler_Update_InvalidColor_Returns422(t *testing.T) {
	now := time.Date(2026, 5, 14, 9, 0, 0, 0, time.UTC)
	bs, _ := entities.NewBrandSettings("Old", "", "", "", "", "", now)
	repo := &fakeRepo{settings: bs}
	audit := &fakeAudit{}
	r := newTestEngine(t, repo, fakeClock{now: now}, audit)

	body, _ := json.Marshal(dto.UpdateBrandSettingsRequest{
		AppName:      "Brand",
		PrimaryColor: "not-a-color",
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/admin/branding", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnprocessableEntity, w.Code, "domain validation → 422")
	assert.Equal(t, 0, repo.updates, "rejected update does not hit repo")
	assert.Len(t, audit.calls, 0, "rejected update emits no audit")
}

func TestAdminBrandingHandler_Update_InvalidURL_Returns422(t *testing.T) {
	now := time.Date(2026, 5, 14, 9, 0, 0, 0, time.UTC)
	bs, _ := entities.NewBrandSettings("Old", "", "", "", "", "", now)
	repo := &fakeRepo{settings: bs}
	r := newTestEngine(t, repo, fakeClock{now: now}, &fakeAudit{})

	body, _ := json.Marshal(dto.UpdateBrandSettingsRequest{
		AppName: "Brand",
		LogoURL: "javascript:alert(1)",
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/admin/branding", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnprocessableEntity, w.Code,
		"javascript:/data:/file: blocked by URL whitelist → 422")
}

func TestAdminBrandingHandler_Update_EmptyAppName_Returns422(t *testing.T) {
	now := time.Date(2026, 5, 14, 9, 0, 0, 0, time.UTC)
	bs, _ := entities.NewBrandSettings("Old", "", "", "", "", "", now)
	repo := &fakeRepo{settings: bs}
	r := newTestEngine(t, repo, fakeClock{now: now}, &fakeAudit{})

	body, _ := json.Marshal(dto.UpdateBrandSettingsRequest{
		AppName: "",
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/admin/branding", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestNewAdminBrandingHandler_NilUseCase_Panics(t *testing.T) {
	assert.Panics(t, func() {
		_ = NewAdminBrandingHandler(nil, nil)
	}, "nil use case must fail DI construction")
}
