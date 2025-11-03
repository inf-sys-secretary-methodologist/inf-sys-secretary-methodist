package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCORSMiddleware_Handler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name            string
		allowedOrigins  []string
		allowedMethods  []string
		allowedHeaders  []string
		requestOrigin   string
		requestMethod   string
		expectedOrigin  string
		expectedMethods string
		expectedHeaders string
		shouldAbort     bool
	}{
		{
			name:            "wildcard origin",
			allowedOrigins:  []string{"*"},
			allowedMethods:  []string{"GET", "POST"},
			allowedHeaders:  []string{"Content-Type"},
			requestOrigin:   "http://localhost:3000",
			requestMethod:   "GET",
			expectedOrigin:  "http://localhost:3000",
			expectedMethods: "GET, POST",
			expectedHeaders: "Content-Type",
			shouldAbort:     false,
		},
		{
			name:            "exact origin match",
			allowedOrigins:  []string{"http://localhost:3000"},
			allowedMethods:  []string{"GET", "POST", "PUT"},
			allowedHeaders:  []string{"Content-Type", "Authorization"},
			requestOrigin:   "http://localhost:3000",
			requestMethod:   "GET",
			expectedOrigin:  "http://localhost:3000",
			expectedMethods: "GET, POST, PUT",
			expectedHeaders: "Content-Type, Authorization",
			shouldAbort:     false,
		},
		{
			name:           "origin not allowed",
			allowedOrigins: []string{"http://allowed.com"},
			allowedMethods: []string{"GET"},
			allowedHeaders: []string{"Content-Type"},
			requestOrigin:  "http://notallowed.com",
			requestMethod:  "GET",
			expectedOrigin: "",
			shouldAbort:    false,
		},
		{
			name:            "preflight request with OPTIONS",
			allowedOrigins:  []string{"http://localhost:3000"},
			allowedMethods:  []string{"GET", "POST"},
			allowedHeaders:  []string{"Content-Type"},
			requestOrigin:   "http://localhost:3000",
			requestMethod:   "OPTIONS",
			expectedOrigin:  "http://localhost:3000",
			expectedMethods: "GET, POST",
			expectedHeaders: "Content-Type",
			shouldAbort:     true,
		},
		{
			name:           "empty allowed origins",
			allowedOrigins: []string{},
			allowedMethods: []string{"GET"},
			allowedHeaders: []string{"Content-Type"},
			requestOrigin:  "http://localhost:3000",
			requestMethod:  "GET",
			expectedOrigin: "",
			shouldAbort:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewCORSMiddleware(tt.allowedOrigins, tt.allowedMethods, tt.allowedHeaders)

			router := gin.New()
			router.Use(middleware.Handler())
			router.GET("/test", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(tt.requestMethod, "/test", nil)
			req.Header.Set("Origin", tt.requestOrigin)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if tt.expectedOrigin != "" {
				assert.Equal(t, tt.expectedOrigin, w.Header().Get("Access-Control-Allow-Origin"))
				assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
			} else {
				assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
			}

			if tt.expectedMethods != "" {
				assert.Equal(t, tt.expectedMethods, w.Header().Get("Access-Control-Allow-Methods"))
			}

			if tt.expectedHeaders != "" {
				assert.Equal(t, tt.expectedHeaders, w.Header().Get("Access-Control-Allow-Headers"))
			}

			if tt.shouldAbort {
				assert.Equal(t, http.StatusNoContent, w.Code)
			} else {
				assert.Equal(t, http.StatusOK, w.Code)
			}
		})
	}
}

func TestCORSMiddleware_getAllowedOrigin(t *testing.T) {
	tests := []struct {
		name           string
		allowedOrigins []string
		requestOrigin  string
		expected       string
	}{
		{
			name:           "wildcard allows any origin",
			allowedOrigins: []string{"*"},
			requestOrigin:  "http://example.com",
			expected:       "http://example.com",
		},
		{
			name:           "exact match",
			allowedOrigins: []string{"http://localhost:3000", "http://example.com"},
			requestOrigin:  "http://localhost:3000",
			expected:       "http://localhost:3000",
		},
		{
			name:           "no match",
			allowedOrigins: []string{"http://allowed.com"},
			requestOrigin:  "http://notallowed.com",
			expected:       "",
		},
		{
			name:           "empty allowed origins",
			allowedOrigins: []string{},
			requestOrigin:  "http://example.com",
			expected:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewCORSMiddleware(tt.allowedOrigins, nil, nil)
			result := middleware.getAllowedOrigin(tt.requestOrigin)
			assert.Equal(t, tt.expected, result)
		})
	}
}
