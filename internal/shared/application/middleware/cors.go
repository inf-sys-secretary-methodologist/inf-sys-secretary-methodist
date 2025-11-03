package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware provides CORS headers
type CORSMiddleware struct {
	allowedOrigins []string
	allowedMethods []string
	allowedHeaders []string
}

// NewCORSMiddleware creates a new CORS middleware
func NewCORSMiddleware(origins, methods, headers []string) *CORSMiddleware {
	return &CORSMiddleware{
		allowedOrigins: origins,
		allowedMethods: methods,
		allowedHeaders: headers,
	}
}

// Handler returns Gin middleware function
func (m *CORSMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Проверяем разрешенные origins
		allowedOrigin := m.getAllowedOrigin(origin)
		if allowedOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowedOrigin)
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// Устанавливаем разрешенные методы
		if len(m.allowedMethods) > 0 {
			c.Header("Access-Control-Allow-Methods", strings.Join(m.allowedMethods, ", "))
		}

		// Устанавливаем разрешенные заголовки
		if len(m.allowedHeaders) > 0 {
			c.Header("Access-Control-Allow-Headers", strings.Join(m.allowedHeaders, ", "))
		}

		// Обрабатываем preflight запросы
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// getAllowedOrigin проверяет, разрешен ли origin
func (m *CORSMiddleware) getAllowedOrigin(origin string) string {
	if len(m.allowedOrigins) == 0 {
		return ""
	}

	// Проверяем на wildcard
	for _, allowed := range m.allowedOrigins {
		if allowed == "*" {
			return origin
		}
		if allowed == origin {
			return origin
		}
	}

	return ""
}
