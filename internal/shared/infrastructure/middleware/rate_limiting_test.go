package middleware

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testRemoteAddr = "192.168.1.1:12345"

// setupTestRedis создаёт in-memory Redis сервер для тестов
func setupTestRedis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	mr := miniredis.RunT(t)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return client, mr
}

func TestRateLimiter_WithinLimit(t *testing.T) {
	client, _ := setupTestRedis(t)
	defer func() {
		_ = client.Close()
	}()

	// Создаём rate limiter: 10 req/min + burst 5 = всего 15
	limiter := NewRateLimiter(client, 10, 5)

	// Создаём router с middleware
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(limiter.RateLimitMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Делаем первый запрос
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// Проверяем, что запрос прошёл
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "10", w.Header().Get("X-RateLimit-Limit"))
	assert.Equal(t, "5", w.Header().Get("X-RateLimit-Burst"))
	assert.Equal(t, "14", w.Header().Get("X-RateLimit-Remaining")) // 15 - 1 = 14
}

func TestRateLimiter_ExceedLimit(t *testing.T) {
	client, _ := setupTestRedis(t)
	defer func() {
		_ = client.Close()
	}()

	// Создаём rate limiter: 10 req/min + burst 5 = всего 15
	limiter := NewRateLimiter(client, 10, 5)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(limiter.RateLimitMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Делаем 16 запросов (превышение лимита)
	for i := 0; i < 16; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		if i < 15 {
			// Первые 15 должны пройти
			assert.Equal(t, http.StatusOK, w.Code, "Request %d should pass", i+1)
		} else {
			// 16-й запрос должен быть заблокирован
			assert.Equal(t, http.StatusTooManyRequests, w.Code, "Request %d should be blocked", i+1)
			assert.Equal(t, "0", w.Header().Get("X-RateLimit-Remaining"))
			assert.NotEmpty(t, w.Header().Get("Retry-After"))
		}
	}
}

func TestRateLimiter_BurstSupport(t *testing.T) {
	client, _ := setupTestRedis(t)
	defer func() {
		_ = client.Close()
	}()

	// Создаём rate limiter: 10 req/min + burst 5
	limiter := NewRateLimiter(client, 10, 5)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(limiter.RateLimitMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Делаем 12 запросов (в пределах burst)
	for i := 0; i < 12; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		// Все 12 должны пройти (базовый лимит 10 + burst 5)
		assert.Equal(t, http.StatusOK, w.Code, "Request %d should pass (within burst)", i+1)
	}

	// 13-й, 14-й, 15-й запросы тоже должны пройти
	for i := 12; i < 15; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "Request %d should pass", i+1)
	}

	// 16-й должен быть заблокирован
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code, "Request 16 should be blocked")
}

func TestRateLimiter_DifferentIPs(t *testing.T) {
	client, _ := setupTestRedis(t)
	defer func() {
		_ = client.Close()
	}()

	limiter := NewRateLimiter(client, 10, 5)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(limiter.RateLimitMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// IP 1: делаем 15 запросов (исчерпываем лимит)
	for i := 0; i < 15; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = testRemoteAddr
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// IP 1: 16-й запрос должен быть заблокирован
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = testRemoteAddr
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusTooManyRequests, w1.Code)

	// IP 2: первый запрос должен пройти (свой лимит)
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.2:12346"
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Equal(t, "14", w2.Header().Get("X-RateLimit-Remaining"))
}

func TestRateLimiter_RetryAfterHeader(t *testing.T) {
	client, _ := setupTestRedis(t)
	defer func() {
		_ = client.Close()
	}()

	limiter := NewRateLimiter(client, 10, 5)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(limiter.RateLimitMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Делаем 15 запросов (исчерпываем лимит)
	for i := 0; i < 15; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// 16-й запрос должен быть заблокирован
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	// Проверяем, что установлены правильные заголовки при блокировке
	retryAfter := w.Header().Get("Retry-After")
	assert.NotEmpty(t, retryAfter, "Retry-After header should be set")

	resetTime := w.Header().Get("X-RateLimit-Reset")
	assert.NotEmpty(t, resetTime, "X-RateLimit-Reset header should be set")

	assert.Equal(t, "0", w.Header().Get("X-RateLimit-Remaining"))
	assert.Equal(t, "10", w.Header().Get("X-RateLimit-Limit"))
	assert.Equal(t, "5", w.Header().Get("X-RateLimit-Burst"))
}

func TestRateLimiter_HeadersPresent(t *testing.T) {
	client, _ := setupTestRedis(t)
	defer func() {
		_ = client.Close()
	}()

	limiter := NewRateLimiter(client, 10, 5)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(limiter.RateLimitMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	// Проверяем все необходимые заголовки
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Limit"))
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Burst"))
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Remaining"))
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Reset"))
}

// Legacy TestGetRealIP_* tests removed in v0.159.0 ADR-3b: the bare
// getRealIP wrapper that unconditionally trusted X-Forwarded-For was
// the bug the audit flagged. Its replacement getRealIPWithTrustedProxies
// is covered by TestGetRealIPWithTrustedProxies above. Issue #279.

// TestGetRealIPWithTrustedProxies pins v0.159.0 ADR-3b: X-Forwarded-For
// is honored only when r.RemoteAddr falls inside the supplied trusted-
// proxy CIDR allowlist. With an empty CIDR list (secure default) the
// header is ignored — an internet-facing client cannot spoof its source
// IP through X-Forwarded-For. The TCP peer is used as the source of
// truth, stripped of the port for stable bucket keys. Issue #279.
func TestGetRealIPWithTrustedProxies(t *testing.T) {
	parseCIDRs := func(t *testing.T, specs ...string) []*net.IPNet {
		t.Helper()
		out := make([]*net.IPNet, 0, len(specs))
		for _, s := range specs {
			_, cidr, err := net.ParseCIDR(s)
			require.NoError(t, err, "test setup: parse CIDR %q", s)
			out = append(out, cidr)
		}
		return out
	}

	cases := []struct {
		name          string
		remoteAddr    string
		xForwardedFor string
		xRealIP       string
		trustedCIDRs  []*net.IPNet
		want          string
	}{
		{
			name:          "X-Forwarded-For trusted when RemoteAddr inside CIDR",
			remoteAddr:    "10.0.0.5:54321",
			xForwardedFor: "1.2.3.4",
			trustedCIDRs:  parseCIDRs(t, "10.0.0.0/8"),
			want:          "1.2.3.4",
		},
		{
			name:          "X-Forwarded-For IGNORED when RemoteAddr outside CIDR (spoof attempt)",
			remoteAddr:    "8.8.8.8:54321",
			xForwardedFor: "1.2.3.4",
			trustedCIDRs:  parseCIDRs(t, "10.0.0.0/8"),
			want:          "8.8.8.8",
		},
		{
			name:          "Empty trusted-CIDR list — secure default ignores X-Forwarded-For",
			remoteAddr:    "10.0.0.5:54321",
			xForwardedFor: "1.2.3.4",
			trustedCIDRs:  nil,
			want:          "10.0.0.5",
		},
		{
			name:          "First IP in X-Forwarded-For chain is taken when proxy trusted",
			remoteAddr:    "10.0.0.5:54321",
			xForwardedFor: "1.2.3.4, 5.6.7.8, 9.10.11.12",
			trustedCIDRs:  parseCIDRs(t, "10.0.0.0/8"),
			want:          "1.2.3.4",
		},
		{
			name:         "No proxy headers — peer IP returned without port",
			remoteAddr:   "13.14.15.16:5678",
			trustedCIDRs: parseCIDRs(t, "10.0.0.0/8"),
			want:         "13.14.15.16",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tc.remoteAddr
			if tc.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tc.xForwardedFor)
			}
			if tc.xRealIP != "" {
				req.Header.Set("X-Real-IP", tc.xRealIP)
			}

			got := getRealIPWithTrustedProxies(req, tc.trustedCIDRs)
			assert.Equal(t, tc.want, got)
		})
	}
}

// TestParseTrustedProxyCIDRs pins the env-spec parser: comma-separated
// list of CIDRs, malformed / blank entries silently skipped (logging
// is the caller's concern). Issue #279 ADR-3b.
func TestParseTrustedProxyCIDRs(t *testing.T) {
	cases := []struct {
		name   string
		spec   string
		wantN  int
		wantOK []string // CIDR strings that must successfully parse
	}{
		{"empty spec returns nil", "", 0, nil},
		{"single CIDR", "10.0.0.0/8", 1, []string{"10.0.0.0/8"}},
		{"multiple CIDRs", "10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16", 3, []string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}},
		{"malformed entries skipped", "10.0.0.0/8, not-a-cidr, 192.168.0.0/16", 2, []string{"10.0.0.0/8", "192.168.0.0/16"}},
		{"IPv6 CIDR accepted", "fc00::/7", 1, []string{"fc00::/7"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ParseTrustedProxyCIDRs(tc.spec)
			require.Len(t, got, tc.wantN)
			for i, want := range tc.wantOK {
				_, expected, err := net.ParseCIDR(want)
				require.NoError(t, err)
				assert.Equal(t, expected.String(), got[i].String())
			}
		})
	}
}

func TestRateLimitConfig_LoadFromEnv(t *testing.T) {
	// Устанавливаем env переменные
	t.Setenv("RATE_LIMIT_PUBLIC_RPM", "20")
	t.Setenv("RATE_LIMIT_PUBLIC_BURST", "10")
	t.Setenv("RATE_LIMIT_AUTH_RPM", "100")
	t.Setenv("RATE_LIMIT_AUTH_BURST", "20")

	config := LoadRateLimitConfig()

	assert.Equal(t, 20, config.PublicRequestsPerMinute)
	assert.Equal(t, 10, config.PublicBurst)
	assert.Equal(t, 100, config.AuthRequestsPerMinute)
	assert.Equal(t, 20, config.AuthBurst)
}

func TestRateLimitConfig_DefaultValues(t *testing.T) {
	// Очищаем env переменные
	t.Setenv("RATE_LIMIT_PUBLIC_RPM", "")
	t.Setenv("RATE_LIMIT_PUBLIC_BURST", "")
	t.Setenv("RATE_LIMIT_AUTH_RPM", "")
	t.Setenv("RATE_LIMIT_AUTH_BURST", "")

	config := LoadRateLimitConfig()

	// Проверяем default значения
	assert.Equal(t, 300, config.PublicRequestsPerMinute)
	assert.Equal(t, 100, config.PublicBurst)
	assert.Equal(t, 1000, config.AuthRequestsPerMinute)
	assert.Equal(t, 200, config.AuthBurst)
}

func TestRateLimitConfig_GetLimiters(t *testing.T) {
	client, _ := setupTestRedis(t)
	defer func() {
		_ = client.Close()
	}()

	config := &RateLimitConfig{
		PublicRequestsPerMinute: 15,
		PublicBurst:             7,
		AuthRequestsPerMinute:   80,
		AuthBurst:               12,
	}

	publicLimiter := config.GetPublicRateLimiter(client)
	require.NotNil(t, publicLimiter)
	assert.Equal(t, 15, publicLimiter.requests)
	assert.Equal(t, 7, publicLimiter.burst)

	authLimiter := config.GetAuthRateLimiter(client)
	require.NotNil(t, authLimiter)
	assert.Equal(t, 80, authLimiter.requests)
	assert.Equal(t, 12, authLimiter.burst)
}
