// Package middleware contains shared infrastructure middleware components.
package middleware

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// TracingMiddleware returns OpenTelemetry tracing middleware for Gin.
// It creates spans for each HTTP request and propagates trace context.
func TracingMiddleware(serviceName string) gin.HandlerFunc {
	return otelgin.Middleware(serviceName)
}
