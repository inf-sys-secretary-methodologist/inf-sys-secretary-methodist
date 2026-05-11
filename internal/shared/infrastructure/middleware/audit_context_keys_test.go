package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/middleware"
)

// These tests pin the production contract between the shared middleware
// stack and AuditLogger's typed context-key extractors. Audit_logs row
// writes (v0.130.0) read actor data from request context via
// logging.ContextKey* keys; production middlewares historically wrote
// values under a different unexported key type, so audit rows would
// land with actor_user_id / actor_ip / correlation_id = NULL. These
// tests prevent regression on that bug class.

func init() { gin.SetMode(gin.TestMode) }

func TestRequestIDMiddleware_WritesTypedCorrelationKey(t *testing.T) {
	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())

	var captured any
	router.GET("/probe", func(c *gin.Context) {
		captured = c.Request.Context().Value(logging.ContextKeyCorrelationID)
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	req.Header.Set("X-Request-ID", "test-corr-id")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, captured,
		"RequestIDMiddleware must write logging.ContextKeyCorrelationID into request context "+
			"so AuditLogger.persist can extract correlation_id for audit_logs rows")
	require.Equal(t, "test-corr-id", captured,
		"correlation id value must match the inbound X-Request-ID header verbatim")
}

func TestRequestContextMiddleware_WritesTypedIPAddressKey(t *testing.T) {
	router := gin.New()
	router.Use(middleware.RequestContextMiddleware())

	var captured any
	router.GET("/probe", func(c *gin.Context) {
		captured = c.Request.Context().Value(logging.ContextKeyIPAddress)
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	req.RemoteAddr = "10.0.0.55:1234"
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, captured,
		"RequestContextMiddleware must write logging.ContextKeyIPAddress into request context "+
			"so AuditLogger.persist can extract actor_ip for audit_logs rows")
	require.NotEmpty(t, captured)
}
