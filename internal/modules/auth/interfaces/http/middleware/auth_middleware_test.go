package middleware

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

func TestRequireRole(t *testing.T) {
	t.Run("allows user with matching role", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("role", "admin")
			c.Next()
		})
		router.Use(RequireRole("admin", "moderator"))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("denies user with non-matching role", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("role", "student")
			c.Next()
		})
		router.Use(RequireRole("admin", "moderator"))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("denies when role is not set", func(t *testing.T) {
		router := gin.New()
		router.Use(RequireRole("admin"))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("denies when role is wrong type", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("role", 123) // wrong type
			c.Next()
		})
		router.Use(RequireRole("admin"))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	router := gin.New()
	router.Use(SecurityHeadersMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "max-age=31536000; includeSubDomains", w.Header().Get("Strict-Transport-Security"))
	assert.Equal(t, "default-src 'self'", w.Header().Get("Content-Security-Policy"))
	assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
}

func TestRateLimiter(t *testing.T) {
	t.Run("allows requests within limit", func(t *testing.T) {
		rl := &RateLimiter{
			requests: make(map[string]*rateLimitEntry),
			max:      5,
			window:   time.Minute,
		}

		for i := 0; i < 5; i++ {
			assert.True(t, rl.Allow("test-key"), "request %d should be allowed", i+1)
		}
	})

	t.Run("blocks requests exceeding limit", func(t *testing.T) {
		rl := &RateLimiter{
			requests: make(map[string]*rateLimitEntry),
			max:      3,
			window:   time.Minute,
		}

		// First 3 should be allowed
		for i := 0; i < 3; i++ {
			assert.True(t, rl.Allow("test-key"))
		}

		// 4th should be blocked
		assert.False(t, rl.Allow("test-key"))
	})

	t.Run("resets after window expires", func(t *testing.T) {
		rl := &RateLimiter{
			requests: make(map[string]*rateLimitEntry),
			max:      2,
			window:   10 * time.Millisecond,
		}

		// Use up the limit
		assert.True(t, rl.Allow("test-key"))
		assert.True(t, rl.Allow("test-key"))
		assert.False(t, rl.Allow("test-key"))

		// Wait for window to expire
		time.Sleep(20 * time.Millisecond)

		// Should be allowed again
		assert.True(t, rl.Allow("test-key"))
	})

	t.Run("tracks different keys independently", func(t *testing.T) {
		rl := &RateLimiter{
			requests: make(map[string]*rateLimitEntry),
			max:      2,
			window:   time.Minute,
		}

		// Use up limit for key1
		assert.True(t, rl.Allow("key1"))
		assert.True(t, rl.Allow("key1"))
		assert.False(t, rl.Allow("key1"))

		// key2 should still be allowed
		assert.True(t, rl.Allow("key2"))
		assert.True(t, rl.Allow("key2"))
	})
}

func TestRateLimitMiddleware(t *testing.T) {
	t.Run("allows requests within limit", func(t *testing.T) {
		router := gin.New()
		router.Use(RateLimitMiddleware(3, time.Minute))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		for i := 0; i < 3; i++ {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = "192.168.1.1:12345"
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
		}
	})

	t.Run("blocks requests exceeding limit", func(t *testing.T) {
		router := gin.New()
		router.Use(RateLimitMiddleware(2, time.Minute))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// First 2 should pass
		for i := 0; i < 2; i++ {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = "192.168.1.2:12345"
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
		}

		// 3rd should be blocked
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.2:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusTooManyRequests, w.Code)
	})
}
