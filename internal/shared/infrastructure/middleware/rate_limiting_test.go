package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
		req.RemoteAddr = "192.168.1.1:12345"
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// IP 1: 16-й запрос должен быть заблокирован
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.1:12345"
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

func TestGetRealIP_XForwardedFor(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	req.RemoteAddr = "5.6.7.8:1234"

	ip := getRealIP(req)
	assert.Equal(t, "1.2.3.4", ip)
}

func TestGetRealIP_XRealIP(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Real-IP", "9.10.11.12")
	req.RemoteAddr = "5.6.7.8:1234"

	ip := getRealIP(req)
	assert.Equal(t, "9.10.11.12", ip)
}

func TestGetRealIP_RemoteAddr(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "13.14.15.16:5678"

	ip := getRealIP(req)
	assert.Equal(t, "13.14.15.16:5678", ip)
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
	assert.Equal(t, 10, config.PublicRequestsPerMinute)
	assert.Equal(t, 5, config.PublicBurst)
	assert.Equal(t, 60, config.AuthRequestsPerMinute)
	assert.Equal(t, 10, config.AuthBurst)
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
