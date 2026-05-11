package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

// This test pins the production contract between JWTMiddleware and
// AuditLogger's typed context-key extractors. After successful token
// validation the middleware must write the authenticated user id into
// the request context under logging.ContextKeyUserID so the audit_logs
// row can carry actor_user_id; without this write every persisted row
// reads NULL actor (v0.130.0 reviewer Tier 1 finding).

func TestJWTMiddleware_WritesTypedUserIDKey(t *testing.T) {
	jwtSecret := []byte("test-jwt-secret-key")
	refreshSecret := []byte("test-refresh-secret-key")
	authUseCase := usecases.NewAuthUseCase(
		new(MockUserRepository),
		jwtSecret, refreshSecret,
		[]byte("mfa-intermediate"),
		nil, nil, nil,
	)

	router := gin.New()
	router.Use(JWTMiddleware(authUseCase))

	var captured any
	router.GET("/probe", func(c *gin.Context) {
		captured = c.Request.Context().Value(logging.ContextKeyUserID)
		c.Status(http.StatusOK)
	})

	token := generateTestToken(jwtSecret, 42, "admin")
	req := httptest.NewRequest(http.MethodGet, "/probe", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, captured,
		"JWTMiddleware must write logging.ContextKeyUserID into request context so "+
			"AuditLogger.persist can extract actor_user_id for audit_logs rows")
	id, ok := captured.(int64)
	require.True(t, ok, "typed key value must be int64 — Audit extractor type-asserts on int64")
	require.Equal(t, int64(42), id)
}
