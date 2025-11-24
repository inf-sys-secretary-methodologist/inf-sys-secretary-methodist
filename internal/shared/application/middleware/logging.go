package middleware

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

// LoggingMiddleware provides request logging
type LoggingMiddleware struct {
	logger *logging.Logger
}

// NewLoggingMiddleware creates a new logging middleware
func NewLoggingMiddleware(logger *logging.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger,
	}
}

// Handler returns Gin middleware function
func (m *LoggingMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Обрабатываем запрос
		c.Next()

		// Вычисляем время выполнения
		duration := time.Since(start)

		// Логируем запрос
		fields := map[string]interface{}{
			"method":      c.Request.Method,
			"path":        path,
			"status_code": c.Writer.Status(),
			"duration_ms": duration.Milliseconds(),
			"remote_addr": c.ClientIP(),
			"user_agent":  c.Request.UserAgent(),
		}

		if query != "" {
			fields["query"] = query
		}

		// Добавляем request_id если есть
		if requestID := c.GetString("request_id"); requestID != "" {
			fields["request_id"] = requestID
		}

		// Логируем с соответствующим уровнем в зависимости от статуса
		statusCode := c.Writer.Status()
		if statusCode >= 500 {
			m.logger.Error("HTTP request", fields)
		} else if statusCode >= 400 {
			m.logger.Warn("HTTP request", fields)
		} else {
			m.logger.Info("HTTP request", fields)
		}
	}
}
