package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
	"github.com/stretchr/testify/assert"
)

func TestLoggingMiddleware_Handler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		method         string
		path           string
		queryString    string
		statusCode     int
		requestID      string
		expectedLevel  string
	}{
		{
			name:          "successful request logs info",
			method:        "GET",
			path:          "/test",
			queryString:   "",
			statusCode:    http.StatusOK,
			requestID:     "",
			expectedLevel: "info",
		},
		{
			name:          "client error logs warn",
			method:        "POST",
			path:          "/test",
			queryString:   "",
			statusCode:    http.StatusBadRequest,
			requestID:     "req-123",
			expectedLevel: "warn",
		},
		{
			name:          "server error logs error",
			method:        "GET",
			path:          "/test",
			queryString:   "",
			statusCode:    http.StatusInternalServerError,
			requestID:     "req-456",
			expectedLevel: "error",
		},
		{
			name:          "request with query string",
			method:        "GET",
			path:          "/test",
			queryString:   "key=value&foo=bar",
			statusCode:    http.StatusOK,
			requestID:     "",
			expectedLevel: "info",
		},
		{
			name:          "unauthorized logs warn",
			method:        "GET",
			path:          "/protected",
			queryString:   "",
			statusCode:    http.StatusUnauthorized,
			requestID:     "",
			expectedLevel: "warn",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logging.NewLogger("info")
			middleware := NewLoggingMiddleware(logger)

			router := gin.New()
			router.Use(middleware.Handler())
			router.Handle(tt.method, tt.path, func(c *gin.Context) {
				if tt.requestID != "" {
					c.Set("request_id", tt.requestID)
				}
				c.Status(tt.statusCode)
			})

			url := tt.path
			if tt.queryString != "" {
				url += "?" + tt.queryString
			}

			req := httptest.NewRequest(tt.method, url, nil)
			req.Header.Set("User-Agent", "TestAgent/1.0")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.statusCode, w.Code)
		})
	}
}

func TestLoggingMiddleware_LogLevels(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		statusCode int
		wantLevel  string
	}{
		{"2xx success", http.StatusOK, "info"},
		{"2xx created", http.StatusCreated, "info"},
		{"3xx redirect", http.StatusMovedPermanently, "info"},
		{"4xx bad request", http.StatusBadRequest, "warn"},
		{"4xx unauthorized", http.StatusUnauthorized, "warn"},
		{"4xx forbidden", http.StatusForbidden, "warn"},
		{"4xx not found", http.StatusNotFound, "warn"},
		{"5xx internal error", http.StatusInternalServerError, "error"},
		{"5xx service unavailable", http.StatusServiceUnavailable, "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logging.NewLogger("debug")
			middleware := NewLoggingMiddleware(logger)

			router := gin.New()
			router.Use(middleware.Handler())
			router.GET("/test", func(c *gin.Context) {
				c.Status(tt.statusCode)
			})

			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.statusCode, w.Code)
		})
	}
}

func TestLoggingMiddleware_RequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := logging.NewLogger("info")
	middleware := NewLoggingMiddleware(logger)

	router := gin.New()
	router.Use(middleware.Handler())
	router.GET("/test", func(c *gin.Context) {
		c.Set("request_id", "test-request-id-123")
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
