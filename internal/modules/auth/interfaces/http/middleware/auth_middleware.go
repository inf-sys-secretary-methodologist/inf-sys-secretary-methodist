package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/application/usecases"
)

// JWTMiddleware validates JWT tokens
func JWTMiddleware(authUseCase *usecases.AuthUseCase) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Bearer token required"})
			c.Abort()
			return
		}

		ctx := c.Request.Context()
		claims, err := authUseCase.ValidateAccessToken(ctx, tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Add claims to context
		userID, _ := (*claims)["user_id"].(float64)
		c.Set("user_id", int64(userID))
		c.Set("role", (*claims)["role"])
		c.Set("claims", claims)

		c.Next()
	}
}

// RequireRole checks if user has required role
func RequireRole(roles ...string) gin.HandlerFunc {
	roleMap := make(map[string]bool)
	for _, role := range roles {
		roleMap[role] = true
	}

	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			c.Abort()
			return
		}

		roleStr, ok := userRole.(string)
		if !ok || !roleMap[roleStr] {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// SecurityHeadersMiddleware adds security headers to responses
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

// RateLimiter implements simple in-memory rate limiting
type RateLimiter struct {
	mu       sync.Mutex
	requests map[string]*rateLimitEntry
	max      int
	window   time.Duration
}

type rateLimitEntry struct {
	count     int
	resetTime time.Time
}

func NewRateLimiter(maxRequests int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]*rateLimitEntry),
		max:      maxRequests,
		window:   window,
	}

	// Cleanup goroutine
	go rl.cleanup()

	return rl
}

func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	entry, exists := rl.requests[key]
	if !exists || now.After(entry.resetTime) {
		rl.requests[key] = &rateLimitEntry{
			count:     1,
			resetTime: now.Add(rl.window),
		}
		return true
	}

	if entry.count >= rl.max {
		return false
	}

	entry.count++
	return true
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, entry := range rl.requests {
			if now.After(entry.resetTime) {
				delete(rl.requests, key)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimitMiddleware applies rate limiting per IP address
func RateLimitMiddleware(maxRequests int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(maxRequests, window)

	return func(c *gin.Context) {
		key := c.ClientIP()

		if !limiter.Allow(key) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
