// Package middleware contains HTTP middleware for the auth module.
package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
)

// JWTMiddleware validates JWT tokens
// Supports token from Authorization header (Bearer token) or query parameter (?token=xxx)
// Query parameter is useful for file downloads where browser can't set headers
func JWTMiddleware(authUseCase *usecases.AuthUseCase) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		// First try Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				resp := response.Unauthorized("Требуется Bearer токен")
				c.JSON(http.StatusUnauthorized, resp)
				c.Abort()
				return
			}
		} else {
			// Fallback to query parameter for file downloads
			tokenString = c.Query("token")
		}

		if tokenString == "" {
			resp := response.Unauthorized("Требуется токен авторизации")
			c.JSON(http.StatusUnauthorized, resp)
			c.Abort()
			return
		}

		ctx := c.Request.Context()
		claims, err := authUseCase.ValidateAccessToken(ctx, tokenString)
		if err != nil {
			resp := response.Unauthorized("Неверный или истекший токен")
			c.JSON(http.StatusUnauthorized, resp)
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
			resp := response.Forbidden("Роль пользователя не найдена")
			c.JSON(http.StatusForbidden, resp)
			c.Abort()
			return
		}

		roleStr, ok := userRole.(string)
		if !ok || !roleMap[roleStr] {
			resp := response.Forbidden("Недостаточно прав доступа")
			c.JSON(http.StatusForbidden, resp)
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireNonStudent blocks any request whose role is "student" or whose
// role is missing from context. Use on endpoints that students must not
// reach: document creation, reports, analytics. Convenience wrapper around
// RequireRole with the four non-student roles whitelisted.
func RequireNonStudent() gin.HandlerFunc {
	return RequireRole("system_admin", "methodist", "academic_secretary", "teacher")
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

// NewRateLimiter creates a new rate limiter instance.
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

// Allow checks if a request is allowed based on the rate limit.
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
			resp := response.ErrorResponse("RATE_LIMIT_EXCEEDED", "Слишком много запросов. Пожалуйста, попробуйте позже.")
			c.JSON(http.StatusTooManyRequests, resp)
			c.Abort()
			return
		}

		c.Next()
	}
}
