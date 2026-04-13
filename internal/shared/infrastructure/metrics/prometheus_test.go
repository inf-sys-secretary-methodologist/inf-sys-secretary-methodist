package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestMetricsRegistered(t *testing.T) {
	// Verify that all metric variables are non-nil (promauto registers them at init)
	assert.NotNil(t, HTTPRequestsTotal)
	assert.NotNil(t, HTTPRequestDuration)
	assert.NotNil(t, HTTPRequestsInFlight)
	assert.NotNil(t, DatabaseQueriesTotal)
	assert.NotNil(t, DatabaseQueryDuration)
	assert.NotNil(t, CacheOperationsTotal)
	assert.NotNil(t, AuthEventsTotal)
	assert.NotNil(t, BusinessOperationsTotal)
	assert.NotNil(t, ActiveConnections)
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name        string
		fullPath    string
		requestPath string
		want        string
	}{
		{"uses full path", "/users/:id", "/users/123", "/users/:id"},
		{"fallback to request path", "", "/users/123", "/users/123"},
		{"both empty", "", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, normalizePath(tt.fullPath, tt.requestPath))
		})
	}
}

func TestRecordDatabaseQuery(t *testing.T) {
	// Should not panic
	RecordDatabaseQuery("SELECT", "users", "success", 100*time.Millisecond)
	RecordDatabaseQuery("INSERT", "users", "failure", 50*time.Millisecond)
}

func TestRecordCacheOperation(t *testing.T) {
	RecordCacheOperation("get", true)
	RecordCacheOperation("get", false)
}

func TestRecordAuthEvent(t *testing.T) {
	RecordAuthEvent("login", true)
	RecordAuthEvent("login", false)
}

func TestRecordBusinessOperation(t *testing.T) {
	RecordBusinessOperation("schedule", "create", true)
	RecordBusinessOperation("schedule", "create", false)
}

func TestSetActiveConnections(t *testing.T) {
	SetActiveConnections("websocket", 5)
	SetActiveConnections("database", 10)
}

func TestPrometheusMiddleware(t *testing.T) {
	router := gin.New()
	router.Use(PrometheusMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "ok", w.Body.String())
}

func TestPrometheusMiddleware_SkipsMetricsEndpoint(t *testing.T) {
	router := gin.New()
	router.Use(PrometheusMiddleware())
	router.GET("/metrics", func(c *gin.Context) {
		c.String(http.StatusOK, "metrics")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler(t *testing.T) {
	router := gin.New()
	router.GET("/metrics", Handler())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "http_requests_total")
}
