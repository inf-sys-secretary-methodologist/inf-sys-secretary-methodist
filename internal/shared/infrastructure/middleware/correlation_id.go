package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CorrelationIDMiddleware adds a unique correlation ID to each request for distributed tracing
func CorrelationIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to get correlation ID from header
		correlationID := c.GetHeader("X-Correlation-ID")

		// If not present, generate new one
		if correlationID == "" {
			correlationID = uuid.New().String()
		}

		// Set in response header
		c.Header("X-Correlation-ID", correlationID)

		// Add to gin context
		c.Set("correlation_id", correlationID)

		// Add to request context for downstream use
		ctx := context.WithValue(c.Request.Context(), "correlation_id", correlationID)
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
