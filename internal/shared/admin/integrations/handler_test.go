package integrations

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func newTestEngine(
	t *testing.T,
	probe VAPIDProbe,
	vapidPublic, vapidSubject string,
	n8nEnabled bool,
	n8nWebhookURL string,
) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	uc := NewAdminIntegrationsUseCase(probe, vapidPublic, vapidSubject, n8nEnabled, n8nWebhookURL)
	h := NewAdminIntegrationsHandler(uc)
	r := gin.New()
	r.Use(gin.Recovery())
	api := r.Group("/api")
	api.Use(withAuth(1, "system_admin"))
	api.GET("/admin/integrations/config", h.GetConfig)
	return r
}

type envelopeBody struct {
	Success bool   `json:"success"`
	Data    Config `json:"data"`
}

func TestAdminIntegrationsHandler_VAPIDConfigured_N8NEnabled(t *testing.T) {
	r := newTestEngine(t,
		func() bool { return true },
		"BPublicKey123", "mailto:admin@example.com",
		true, "https://n8n.example.com",
	)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/admin/integrations/config", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body envelopeBody
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))

	assert.True(t, body.Success)
	assert.True(t, body.Data.VAPID.Configured, "probe true → configured true")
	assert.Equal(t, "BPublicKey123", body.Data.VAPID.PublicKey,
		"public key surfaces (non-secret, browser receives it via /push/public-key anyway)")
	assert.Equal(t, "mailto:admin@example.com", body.Data.VAPID.Subject)
	assert.True(t, body.Data.N8N.Enabled)
	assert.Equal(t, "https://n8n.example.com", body.Data.N8N.WebhookURL)
}

func TestAdminIntegrationsHandler_VAPIDUnconfigured_N8NDisabled(t *testing.T) {
	r := newTestEngine(t,
		func() bool { return false },
		"", "",
		false, "http://localhost:5678",
	)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/admin/integrations/config", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body envelopeBody
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))

	assert.True(t, body.Success)
	assert.False(t, body.Data.VAPID.Configured,
		"probe false → configured false (WebPush sender will fail at send time)")
	assert.Equal(t, "", body.Data.VAPID.PublicKey, "no public key when unconfigured")
	assert.False(t, body.Data.N8N.Enabled, "disabled flag honored")
	assert.Equal(t, "http://localhost:5678", body.Data.N8N.WebhookURL,
		"webhook URL surfaces even when disabled (default localhost dev value)")
}

func TestNewAdminIntegrationsUseCase_NilProbe_Panics(t *testing.T) {
	assert.Panics(t, func() {
		_ = NewAdminIntegrationsUseCase(nil, "k", "s", true, "u")
	}, "nil VAPIDProbe must fail DI construction")
}

func TestNewAdminIntegrationsHandler_NilUseCase_Panics(t *testing.T) {
	assert.Panics(t, func() {
		_ = NewAdminIntegrationsHandler(nil)
	}, "nil AdminIntegrationsUseCase must fail DI construction")
}

func TestEnvVAPIDProbe_RequiresAllThreeEnvVars(t *testing.T) {
	cases := []struct {
		name    string
		public  string
		private string
		subject string
		want    bool
	}{
		{"all empty", "", "", "", false},
		{"only public", "p", "", "", false},
		{"only private", "", "x", "", false},
		{"only subject", "", "", "s", false},
		{"public+private no subject", "p", "x", "", false},
		{"public+subject no private", "p", "", "s", false},
		{"private+subject no public", "", "x", "s", false},
		{"all three set", "p", "x", "s", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("VAPID_PUBLIC_KEY", tc.public)
			t.Setenv("VAPID_PRIVATE_KEY", tc.private)
			t.Setenv("VAPID_SUBJECT", tc.subject)
			assert.Equal(t, tc.want, EnvVAPIDProbe())
		})
	}
}
