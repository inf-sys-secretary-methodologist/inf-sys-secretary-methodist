package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/application/usecases"
)

// fakeRevokedRepo — минимальный mock RevokedTokenRepository.
type fakeRevokedRepo struct {
	mock.Mock
}

func (f *fakeRevokedRepo) Revoke(ctx context.Context, jti string, ttl time.Duration) error {
	return f.Called(ctx, jti, ttl).Error(0)
}

func (f *fakeRevokedRepo) IsRevoked(ctx context.Context, jti string) (bool, error) {
	args := f.Called(ctx, jti)
	return args.Bool(0), args.Error(1)
}

func signValidToken(t *testing.T, secret []byte, jti string) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"jti":     jti,
		"user_id": float64(11),
		"role":    "teacher",
		"exp":     time.Now().Add(10 * time.Minute).Unix(),
		"iat":     time.Now().Add(-time.Minute).Unix(),
	})
	signed, err := tok.SignedString(secret)
	assert.NoError(t, err)
	return signed
}

// buildAuthUC creates a minimal AuthUseCase capable of validating tokens.
// Repositories and notification dependencies are nil — ValidateAccessToken
// only touches jwtSecret.
func buildAuthUC(secret []byte) *usecases.AuthUseCase {
	return usecases.NewAuthUseCase(nil, secret, secret, nil, nil, nil)
}

// TestJWTMiddlewareWithRevocation_RejectsRevokedToken — revoked JTI must
// surface as HTTP 401 even though the token is otherwise valid.
func TestJWTMiddlewareWithRevocation_RejectsRevokedToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := []byte("test-secret-revocation")

	repo := new(fakeRevokedRepo)
	repo.On("IsRevoked", mock.Anything, "revoked-jti").Return(true, nil)

	router := gin.New()
	router.Use(JWTMiddlewareWithRevocation(buildAuthUC(secret), repo))
	router.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer "+signValidToken(t, secret, "revoked-jti"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	repo.AssertExpectations(t)
}

// TestJWTMiddlewareWithRevocation_AcceptsActiveToken — non-revoked JTI must
// pass through to the protected handler.
func TestJWTMiddlewareWithRevocation_AcceptsActiveToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := []byte("test-secret-revocation")

	repo := new(fakeRevokedRepo)
	repo.On("IsRevoked", mock.Anything, "active-jti").Return(false, nil)

	router := gin.New()
	router.Use(JWTMiddlewareWithRevocation(buildAuthUC(secret), repo))
	router.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer "+signValidToken(t, secret, "active-jti"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	repo.AssertExpectations(t)
}

// TestJWTMiddlewareWithRevocation_NilRepoSkipsCheck — passing nil repo
// must keep the middleware usable (defensive: useful when revocation is
// disabled in dev). Validation still happens; revocation lookup is skipped.
func TestJWTMiddlewareWithRevocation_NilRepoSkipsCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := []byte("test-secret-revocation")

	router := gin.New()
	router.Use(JWTMiddlewareWithRevocation(buildAuthUC(secret), nil))
	router.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer "+signValidToken(t, secret, "any-jti"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
