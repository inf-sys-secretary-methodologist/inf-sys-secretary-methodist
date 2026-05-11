// Package middleware contains shared infrastructure middleware components.
package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

// Context key types for type-safe context values
type contextKey string

const (
	contextKeyRequestID  contextKey = "request_id"
	contextKeyIPAddress  contextKey = "ip_address"
	contextKeyUserAgent  contextKey = "user_agent"
	contextKeyHTTPMethod contextKey = "http_method"
	contextKeyHTTPPath   contextKey = "http_path"
)

// RequestIDMiddleware adds a unique request ID to each request for distributed tracing
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to get request ID from header (support both X-Request-ID and X-Correlation-ID)
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = c.GetHeader("X-Correlation-ID")
		}

		// If not present, generate new one
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Set in response headers (both for backward compatibility)
		c.Header("X-Request-ID", requestID)
		c.Header("X-Correlation-ID", requestID)

		// Add to gin context (используем request_id для консистентности с другими middleware)
		c.Set("request_id", requestID)

		// Add to request context for downstream use. Two writes:
		//   - middleware.contextKeyRequestID — preserves the original
		//     unexported-key contract for existing in-package consumers.
		//   - logging.ContextKeyCorrelationID — typed key that
		//     AuditLogger.persist reads when populating audit_logs.
		//     Without this second write every persisted row would carry
		//     correlation_id = NULL (v0.130.0 reviewer Tier 1).
		ctx := context.WithValue(c.Request.Context(), contextKeyRequestID, requestID)
		ctx = context.WithValue(ctx, logging.ContextKeyCorrelationID, requestID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// RequestContextMiddleware enriches context with request metadata.
//
// IP address is written under both the local unexported contextKey
// (preserving existing in-package consumers) and the exported
// logging.ContextKeyIPAddress so AuditLogger.persist can extract
// actor_ip when writing an audit_logs row. The other metadata keys
// remain middleware-internal — they are not consumed by the audit
// pipeline today.
func RequestContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Add IP address (both keys — see doc comment above)
		ip := c.ClientIP()
		ctx = context.WithValue(ctx, contextKeyIPAddress, ip)
		ctx = context.WithValue(ctx, logging.ContextKeyIPAddress, ip)

		// Add user agent
		ctx = context.WithValue(ctx, contextKeyUserAgent, c.Request.UserAgent())

		// Add request method and path
		ctx = context.WithValue(ctx, contextKeyHTTPMethod, c.Request.Method)
		ctx = context.WithValue(ctx, contextKeyHTTPPath, c.Request.URL.Path)

		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
