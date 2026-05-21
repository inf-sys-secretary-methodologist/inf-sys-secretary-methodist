// Package middleware contains HTTP middleware for the auth module.
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/http/response"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

// extractAndValidateToken parses the Bearer token (header) or ?token=
// query parameter, validates it via the use case, and on success pins
// (user_id, role, claims) onto the gin context plus promotes user_id
// into the request context for AuditLogger. On any failure it writes
// 401, aborts the chain, and returns ok=false. Single source of truth
// for token plumbing — shared by JWTMiddleware and JWTMiddlewareWith-
// Revocation so the revocation variant can run the check BEFORE
// c.Next() instead of after (issue #279 ADR-1).
func extractAndValidateToken(c *gin.Context, authUseCase *usecases.AuthUseCase) (*jwt.MapClaims, bool) {
	var tokenString string

	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, response.Unauthorized("Требуется Bearer токен"))
			c.Abort()
			return nil, false
		}
	} else {
		tokenString = c.Query("token")
	}

	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("Требуется токен авторизации"))
		c.Abort()
		return nil, false
	}

	ctx := c.Request.Context()
	claims, err := authUseCase.ValidateAccessToken(ctx, tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("Неверный или истекший токен"))
		c.Abort()
		return nil, false
	}

	userID, _ := (*claims)["user_id"].(float64)
	uid := int64(userID)
	c.Set("user_id", uid)
	c.Set("role", (*claims)["role"])
	c.Set("claims", claims)

	// Promote actor id into the request context under the typed
	// logging.ContextKeyUserID so AuditLogger.persist can populate
	// audit_logs.actor_user_id. Gin's c.Set does not propagate to
	// c.Request.Context() — without this explicit promotion every
	// persisted audit row would carry NULL actor (v0.130.0 reviewer
	// Tier 1 finding).
	ctx = context.WithValue(c.Request.Context(), logging.ContextKeyUserID, uid)
	c.Request = c.Request.WithContext(ctx)
	return claims, true
}

// JWTMiddleware validates JWT tokens.
// Supports token from Authorization header (Bearer token) or query parameter (?token=xxx).
// Query parameter is useful for file downloads where browser can't set headers.
func JWTMiddleware(authUseCase *usecases.AuthUseCase) gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := extractAndValidateToken(c, authUseCase); !ok {
			return
		}
		c.Next()
	}
}

// JWTMiddlewareWithRevocation behaves like JWTMiddleware but additionally
// rejects access tokens whose JTI is present in the revoked-token store.
// The revocation check runs BEFORE c.Next() so a revoked token never
// reaches the protected handler — DB writes, audit emits, uploads and
// other side effects are guaranteed not to leak through. Issue #279 ADR-1.
//
// Pass revokedRepo=nil to bypass the revocation check (useful in dev or
// in tests that do not exercise logout).
func JWTMiddlewareWithRevocation(authUseCase *usecases.AuthUseCase, revokedRepo usecases.RevokedTokenRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := extractAndValidateToken(c, authUseCase)
		if !ok {
			return
		}

		if revokedRepo != nil {
			jti, _ := (*claims)["jti"].(string)
			if jti != "" {
				revoked, err := revokedRepo.IsRevoked(c.Request.Context(), jti)
				if err != nil {
					// Fail closed on storage errors — better to force re-login
					// than to risk accepting a token that may have been revoked.
					c.JSON(http.StatusUnauthorized, response.Unauthorized("Не удалось проверить токен"))
					c.Abort()
					return
				}
				if revoked {
					c.JSON(http.StatusUnauthorized, response.Unauthorized("Токен отозван"))
					c.Abort()
					return
				}
			}
		}

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

// In-memory RateLimiter + RateLimitMiddleware were removed in v0.159.0
// ADR-3 Tier 2: the production rate limiter is the Redis-backed
// shared/infrastructure/middleware/rate_limiting.go implementation,
// which supports burst, per-user keying, and trusted-proxy CIDR
// validation. The in-memory version was dead code (no production
// caller) and would have silently bypassed the new CIDR allowlist if
// someone wired it.
