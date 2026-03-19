package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// Tests for RequestIDMiddleware

func TestRequestIDMiddleware_GeneratesNewID(t *testing.T) {
	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
	assert.NotEmpty(t, w.Header().Get("X-Correlation-ID"))
	// Both headers should be the same
	assert.Equal(t, w.Header().Get("X-Request-ID"), w.Header().Get("X-Correlation-ID"))
}

func TestRequestIDMiddleware_UsesExistingRequestID(t *testing.T) {
	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "existing-request-id")
	router.ServeHTTP(w, req)

	assert.Equal(t, "existing-request-id", w.Header().Get("X-Request-ID"))
	assert.Equal(t, "existing-request-id", w.Header().Get("X-Correlation-ID"))
}

func TestRequestIDMiddleware_UsesCorrelationID(t *testing.T) {
	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Correlation-ID", "existing-correlation-id")
	router.ServeHTTP(w, req)

	assert.Equal(t, "existing-correlation-id", w.Header().Get("X-Request-ID"))
	assert.Equal(t, "existing-correlation-id", w.Header().Get("X-Correlation-ID"))
}

func TestRequestIDMiddleware_PrefersRequestIDOverCorrelationID(t *testing.T) {
	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "request-id-value")
	req.Header.Set("X-Correlation-ID", "correlation-id-value")
	router.ServeHTTP(w, req)

	assert.Equal(t, "request-id-value", w.Header().Get("X-Request-ID"))
}

func TestRequestIDMiddleware_SetsInGinContext(t *testing.T) {
	router := gin.New()
	router.Use(RequestIDMiddleware())
	var ctxRequestID string
	router.GET("/test", func(c *gin.Context) {
		val, _ := c.Get("request_id")
		ctxRequestID, _ = val.(string)
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "my-req-id")
	router.ServeHTTP(w, req)

	assert.Equal(t, "my-req-id", ctxRequestID)
}

func TestRequestIDMiddleware_SetsInRequestContext(t *testing.T) {
	router := gin.New()
	router.Use(RequestIDMiddleware())
	var ctxValue string
	router.GET("/test", func(c *gin.Context) {
		val := c.Request.Context().Value(contextKeyRequestID)
		if val != nil {
			ctxValue = val.(string)
		}
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "ctx-test-id")
	router.ServeHTTP(w, req)

	assert.Equal(t, "ctx-test-id", ctxValue)
}

// Tests for RequestContextMiddleware

func TestRequestContextMiddleware_SetsIPAddress(t *testing.T) {
	router := gin.New()
	router.Use(RequestContextMiddleware())
	var ipAddr string
	router.GET("/test", func(c *gin.Context) {
		val := c.Request.Context().Value(contextKeyIPAddress)
		if val != nil {
			ipAddr = val.(string)
		}
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	router.ServeHTTP(w, req)

	assert.NotEmpty(t, ipAddr)
}

func TestRequestContextMiddleware_SetsUserAgent(t *testing.T) {
	router := gin.New()
	router.Use(RequestContextMiddleware())
	var userAgent string
	router.GET("/test", func(c *gin.Context) {
		val := c.Request.Context().Value(contextKeyUserAgent)
		if val != nil {
			userAgent = val.(string)
		}
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "TestAgent/1.0")
	router.ServeHTTP(w, req)

	assert.Equal(t, "TestAgent/1.0", userAgent)
}

func TestRequestContextMiddleware_SetsHTTPMethod(t *testing.T) {
	router := gin.New()
	router.Use(RequestContextMiddleware())
	var method string
	router.POST("/test", func(c *gin.Context) {
		val := c.Request.Context().Value(contextKeyHTTPMethod)
		if val != nil {
			method = val.(string)
		}
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, "POST", method)
}

func TestRequestContextMiddleware_SetsHTTPPath(t *testing.T) {
	router := gin.New()
	router.Use(RequestContextMiddleware())
	var path string
	router.GET("/api/v1/test", func(c *gin.Context) {
		val := c.Request.Context().Value(contextKeyHTTPPath)
		if val != nil {
			path = val.(string)
		}
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, "/api/v1/test", path)
}

// Tests for TracingMiddleware

func TestTracingMiddleware_ReturnsHandler(t *testing.T) {
	handler := TracingMiddleware("test-service")
	assert.NotNil(t, handler)
}

// Tests for getEnvInt

func TestGetEnvInt_ValidValue(t *testing.T) {
	t.Setenv("TEST_ENV_INT", "42")
	assert.Equal(t, 42, getEnvInt("TEST_ENV_INT", 10))
}

func TestGetEnvInt_InvalidValue(t *testing.T) {
	t.Setenv("TEST_ENV_INT_BAD", "not_a_number")
	assert.Equal(t, 10, getEnvInt("TEST_ENV_INT_BAD", 10))
}

func TestGetEnvInt_Missing(t *testing.T) {
	assert.Equal(t, 99, getEnvInt("TEST_ENV_INT_MISSING_XYZZY", 99))
}

func TestGetEnvInt_Empty(t *testing.T) {
	t.Setenv("TEST_ENV_INT_EMPTY", "")
	assert.Equal(t, 55, getEnvInt("TEST_ENV_INT_EMPTY", 55))
}

// Tests for LoadRateLimitConfig with env vars

func TestLoadRateLimitConfig_CustomValues(t *testing.T) {
	t.Setenv("RATE_LIMIT_PUBLIC_RPM", "50")
	t.Setenv("RATE_LIMIT_PUBLIC_BURST", "25")
	t.Setenv("RATE_LIMIT_AUTH_RPM", "500")
	t.Setenv("RATE_LIMIT_AUTH_BURST", "100")

	cfg := LoadRateLimitConfig()
	assert.Equal(t, 50, cfg.PublicRequestsPerMinute)
	assert.Equal(t, 25, cfg.PublicBurst)
	assert.Equal(t, 500, cfg.AuthRequestsPerMinute)
	assert.Equal(t, 100, cfg.AuthBurst)
}

func TestLoadRateLimitConfig_InvalidValues(t *testing.T) {
	t.Setenv("RATE_LIMIT_PUBLIC_RPM", "abc")
	t.Setenv("RATE_LIMIT_PUBLIC_BURST", "xyz")

	// Unset auth vars to get defaults
	os.Unsetenv("RATE_LIMIT_AUTH_RPM")
	os.Unsetenv("RATE_LIMIT_AUTH_BURST")

	cfg := LoadRateLimitConfig()
	assert.Equal(t, 300, cfg.PublicRequestsPerMinute) // default
	assert.Equal(t, 100, cfg.PublicBurst)             // default
}

// Tests for NewRateLimiter constructor

func TestNewRateLimiter_Constructor(t *testing.T) {
	client, _ := setupTestRedis(t)
	defer func() {
		_ = client.Close()
	}()

	rl := NewRateLimiter(client, 100, 50)
	assert.NotNil(t, rl)
	assert.Equal(t, 100, rl.requests)
	assert.Equal(t, 50, rl.burst)
}
