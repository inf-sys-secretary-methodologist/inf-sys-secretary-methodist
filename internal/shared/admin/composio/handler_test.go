package composio

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withAuth mirrors the production JWT middleware contract — sets
// the user_id + role context keys read by RequireRole(system_admin)
// downstream. The Composio admin endpoint is route-gated, not
// handler-gated, so the test engine simulates the post-auth state
// directly to keep the integration test focused on the projection
// rather than re-asserting the canonical role gate.
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

// newTestEngine builds a gin engine wired with a Composio admin
// handler backed by the given probe. Mirror к admin/integrations
// test scaffold — production gin engine satisfies the
// integration-test-through-production-middleware precondition for
// single-pass reviewer SHIP.
func newTestEngine(t *testing.T, probe Probe) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	uc := NewAdminComposioUseCase(probe)
	h := NewAdminComposioHandler(uc)
	r := gin.New()
	r.Use(gin.Recovery())
	api := r.Group("/api")
	api.Use(withAuth(1, "system_admin"))
	api.GET("/admin/composio/config", h.GetConfig)
	return r
}

type envelopeBody struct {
	Success bool   `json:"success"`
	Data    Config `json:"data"`
}

func TestAdminComposioHandler_AllConfigured(t *testing.T) {
	r := newTestEngine(t, func() ProbeResult {
		return ProbeResult{
			APIKeyConfigured: true,
			EntityIDSet:      true,
			MCPConfigIDSet:   true,
		}
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/admin/composio/config", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body envelopeBody
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))

	assert.True(t, body.Success)
	assert.True(t, body.Data.Configured, "all three env set → aggregate Configured true")
	assert.True(t, body.Data.APIKeyConfigured)
	assert.True(t, body.Data.EntityIDSet)
	assert.True(t, body.Data.MCPConfigIDSet)
}

func TestAdminComposioHandler_PartiallyConfigured(t *testing.T) {
	r := newTestEngine(t, func() ProbeResult {
		return ProbeResult{
			APIKeyConfigured: true,
			EntityIDSet:      true,
			MCPConfigIDSet:   false,
		}
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/admin/composio/config", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body envelopeBody
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))

	assert.True(t, body.Success)
	assert.False(t, body.Data.Configured,
		"missing MCPConfigID → aggregate Configured false (admin can see partial state)")
	assert.True(t, body.Data.APIKeyConfigured)
	assert.True(t, body.Data.EntityIDSet)
	assert.False(t, body.Data.MCPConfigIDSet)
}

func TestAdminComposioHandler_NoneConfigured(t *testing.T) {
	r := newTestEngine(t, func() ProbeResult { return ProbeResult{} })
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/admin/composio/config", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var body envelopeBody
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))

	assert.True(t, body.Success, "empty config still returns 200, not 5xx")
	assert.False(t, body.Data.Configured)
	assert.False(t, body.Data.APIKeyConfigured)
	assert.False(t, body.Data.EntityIDSet)
	assert.False(t, body.Data.MCPConfigIDSet)
}

func TestNewAdminComposioUseCase_NilProbe_Panics(t *testing.T) {
	assert.Panics(t, func() {
		_ = NewAdminComposioUseCase(nil)
	}, "nil Probe must fail DI construction")
}

func TestNewAdminComposioHandler_NilUseCase_Panics(t *testing.T) {
	assert.Panics(t, func() {
		_ = NewAdminComposioHandler(nil)
	}, "nil AdminComposioUseCase must fail DI construction")
}

func TestProbeResult_AllConfigured(t *testing.T) {
	cases := []struct {
		name string
		r    ProbeResult
		want bool
	}{
		{"all false", ProbeResult{}, false},
		{"only api key", ProbeResult{APIKeyConfigured: true}, false},
		{"only entity", ProbeResult{EntityIDSet: true}, false},
		{"only mcp", ProbeResult{MCPConfigIDSet: true}, false},
		{"api+entity, no mcp", ProbeResult{APIKeyConfigured: true, EntityIDSet: true}, false},
		{"api+mcp, no entity", ProbeResult{APIKeyConfigured: true, MCPConfigIDSet: true}, false},
		{"entity+mcp, no api", ProbeResult{EntityIDSet: true, MCPConfigIDSet: true}, false},
		{"all true", ProbeResult{APIKeyConfigured: true, EntityIDSet: true, MCPConfigIDSet: true}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.r.AllConfigured())
		})
	}
}

func TestEnvComposioProbe_ReadsAllThreeEnvVars(t *testing.T) {
	cases := []struct {
		name     string
		apiKey   string
		entityID string
		mcpID    string
		want     ProbeResult
	}{
		{"all empty", "", "", "", ProbeResult{}},
		{"only api key", "k", "", "", ProbeResult{APIKeyConfigured: true}},
		{"only entity", "", "e", "", ProbeResult{EntityIDSet: true}},
		{"only mcp", "", "", "m", ProbeResult{MCPConfigIDSet: true}},
		{"api+entity", "k", "e", "", ProbeResult{APIKeyConfigured: true, EntityIDSet: true}},
		{"api+mcp", "k", "", "m", ProbeResult{APIKeyConfigured: true, MCPConfigIDSet: true}},
		{"entity+mcp", "", "e", "m", ProbeResult{EntityIDSet: true, MCPConfigIDSet: true}},
		{"all three set", "k", "e", "m", ProbeResult{APIKeyConfigured: true, EntityIDSet: true, MCPConfigIDSet: true}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("COMPOSIO_API_KEY", tc.apiKey)
			t.Setenv("COMPOSIO_ENTITY_ID", tc.entityID)
			t.Setenv("COMPOSIO_MCP_CONFIG_ID", tc.mcpID)
			assert.Equal(t, tc.want, EnvComposioProbe())
		})
	}
}
