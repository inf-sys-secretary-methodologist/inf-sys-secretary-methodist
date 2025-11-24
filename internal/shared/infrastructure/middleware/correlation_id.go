// Package middleware contains shared infrastructure middleware components.
package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

		// Add to request context for downstream use
		ctx := context.WithValue(c.Request.Context(), "request_id", requestID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// RequestContextMiddleware enriches context with request metadata
func RequestContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Add IP address
		ctx = context.WithValue(ctx, "ip_address", c.ClientIP())

		// Add user agent
		ctx = context.WithValue(ctx, "user_agent", c.Request.UserAgent())

		// Add request method and path
		ctx = context.WithValue(ctx, "http_method", c.Request.Method)
		ctx = context.WithValue(ctx, "http_path", c.Request.URL.Path)

		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
