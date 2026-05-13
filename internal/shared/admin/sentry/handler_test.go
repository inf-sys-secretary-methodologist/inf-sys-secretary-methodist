package sentry

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withAuth mirrors the production JWTMiddleware contract for these
// handler tests — same "user_id" + "role" gin context keys the
// production auth_middleware writes. v0.126.0 wrong-key bug class.
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

func newTestEngine(t *testing.T, probe DSNProbe, environment, release string) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	uc := NewAdminSentryUseCase(probe, environment, release)
	h := NewAdminSentryHandler(uc)
	r := gin.New()
	r.Use(gin.Recovery())
	api := r.Group("/api")
	api.Use(withAuth(1, "system_admin"))
	api.GET("/admin/sentry/config", h.GetConfig)
	return r
}

type configBody struct {
	Success bool   `json:"success"`
	Data    Config `json:"data"`
}

func TestAdminSentryHandler_GetConfig_DSNConfigured(t *testing.T) {
	r := newTestEngine(t, func() bool { return true }, "production", "0.133.0")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/admin/sentry/config", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body configBody
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))

	assert.True(t, body.Success)
	assert.True(t, body.Data.DSNConfigured, "DSN probe returned true → DSNConfigured must be true")
	assert.Equal(t, "production", body.Data.Environment)
	assert.Equal(t, "0.133.0", body.Data.Release)
	assert.InDelta(t, 0.1, body.Data.TracesSampleRate, 1e-9,
		"TracesSampleRate must mirror initSentry constant (0.1)")
	assert.True(t, body.Data.TracingEnabled,
		"TracingEnabled must mirror initSentry constant (true)")
}

func TestAdminSentryHandler_GetConfig_DSNUnconfigured(t *testing.T) {
	r := newTestEngine(t, func() bool { return false }, "development", "0.133.0")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/admin/sentry/config", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body configBody
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))

	assert.True(t, body.Success)
	assert.False(t, body.Data.DSNConfigured,
		"DSN probe returned false → DSNConfigured must be false (Sentry inactive)")
	assert.Equal(t, "development", body.Data.Environment)
	assert.Equal(t, "0.133.0", body.Data.Release)
}

func TestNewAdminSentryUseCase_NilProbe_Panics(t *testing.T) {
	assert.Panics(t, func() {
		_ = NewAdminSentryUseCase(nil, "production", "0.133.0")
	}, "nil DSNProbe must fail DI construction")
}

func TestNewAdminSentryHandler_NilUseCase_Panics(t *testing.T) {
	assert.Panics(t, func() {
		_ = NewAdminSentryHandler(nil)
	}, "nil AdminSentryUseCase must fail DI construction")
}

func TestEnvDSNProbe_ReadsSentryDSNEnv(t *testing.T) {
	t.Setenv("SENTRY_DSN", "")
	assert.False(t, EnvDSNProbe(), "empty SENTRY_DSN → false")

	t.Setenv("SENTRY_DSN", "https://example@sentry.io/123")
	assert.True(t, EnvDSNProbe(), "non-empty SENTRY_DSN → true")
}
