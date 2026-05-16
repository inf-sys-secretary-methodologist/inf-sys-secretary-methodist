package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/branding/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/branding/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/branding/domain/entities"
)

// stubRepo serves a fixed BrandSettings without any state beyond
// what the constructor provides. Public handler does not write so
// no mutation surface is needed.
type stubRepo struct{ settings *entities.BrandSettings }

func (r *stubRepo) Get(_ context.Context) (*entities.BrandSettings, error) {
	return r.settings, nil
}

func (r *stubRepo) Update(_ context.Context, _ *entities.BrandSettings) error {
	return nil
}

var _ usecases.BrandSettingsRepository = (*stubRepo)(nil)

func newPublicTestEngine(t *testing.T, repo usecases.BrandSettingsRepository) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	getUC := usecases.NewGetBrandingUseCase(repo)
	h := NewPublicBrandingHandler(getUC)
	r := gin.New()
	r.Use(gin.Recovery())
	// Public group — no withAuth middleware on purpose. The
	// production publicGroup mounts the same path without any
	// JWT requirement so login pages can fetch branding before
	// authentication.
	publicAPI := r.Group("/api/public")
	publicAPI.GET("/branding", h.GetBranding)
	return r
}

type publicEnvelope struct {
	Success bool                 `json:"success"`
	Data    dto.BrandSettingsDTO `json:"data"`
}

func TestPublicBrandingHandler_Get_ReturnsFullProjection_NoAuthRequired(t *testing.T) {
	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	bs, err := entities.NewBrandSettings(
		"Public Brand", "Public tagline",
		"https://public.example/logo.png", "https://public.example/favicon.ico",
		"#aabbcc", "#ddeeff", now,
	)
	require.NoError(t, err)

	r := newPublicTestEngine(t, &stubRepo{settings: bs})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/public/branding", nil)
	// No Authorization header — public endpoint must accept this.
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code,
		"public GET must not require auth; body=%s", w.Body.String())

	var body publicEnvelope
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.True(t, body.Success)
	assert.Equal(t, "Public Brand", body.Data.AppName)
	assert.Equal(t, "Public tagline", body.Data.Tagline)
	assert.Equal(t, "https://public.example/logo.png", body.Data.LogoURL)
	assert.Equal(t, "https://public.example/favicon.ico", body.Data.FaviconURL)
	assert.Equal(t, "#aabbcc", body.Data.PrimaryColor)
	assert.Equal(t, "#ddeeff", body.Data.SecondaryColor)
	assert.True(t, body.Data.UpdatedAt.Equal(now))
}

func TestNewPublicBrandingHandler_NilUseCase_Panics(t *testing.T) {
	assert.Panics(t, func() {
		_ = NewPublicBrandingHandler(nil)
	})
}
