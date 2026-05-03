package http

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

type fakeRevokedRepoForHandler struct {
	mock.Mock
}

func (f *fakeRevokedRepoForHandler) Revoke(ctx context.Context, jti string, ttl time.Duration) error {
	return f.Called(ctx, jti, ttl).Error(0)
}

func (f *fakeRevokedRepoForHandler) IsRevoked(ctx context.Context, jti string) (bool, error) {
	args := f.Called(ctx, jti)
	return args.Bool(0), args.Error(1)
}

func makeAccessToken(t *testing.T, secret []byte, jti string) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"jti":     jti,
		"user_id": float64(5),
		"role":    "student",
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
		"iat":     time.Now().Add(-time.Minute).Unix(),
	})
	signed, err := tok.SignedString(secret)
	assert.NoError(t, err)
	return signed
}

// TestLogoutHandler_RevokesAndReturnsNoContent — happy path: valid Bearer
// token; usecase revokes JTI; handler returns 204 No Content.
func TestLogoutHandler_RevokesAndReturnsNoContent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := []byte("h-secret")

	repo := new(fakeRevokedRepoForHandler)
	repo.On("Revoke", mock.Anything, "logout-jti-1", mock.Anything).Return(nil)

	uc := usecases.NewLogoutUseCase(repo, secret)
	h := NewLogoutHandler(uc)

	router := gin.New()
	router.POST("/auth/logout", h.Logout)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+makeAccessToken(t, secret, "logout-jti-1"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	repo.AssertExpectations(t)
}

// TestLogoutHandler_MissingAuthorization — no Bearer header → 401.
func TestLogoutHandler_MissingAuthorization(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := []byte("h-secret")

	repo := new(fakeRevokedRepoForHandler)
	uc := usecases.NewLogoutUseCase(repo, secret)
	h := NewLogoutHandler(uc)

	router := gin.New()
	router.POST("/auth/logout", h.Logout)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	repo.AssertNotCalled(t, "Revoke", mock.Anything, mock.Anything, mock.Anything)
}

// TestLogoutHandler_InvalidToken — unparseable Bearer → 401.
func TestLogoutHandler_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := []byte("h-secret")

	repo := new(fakeRevokedRepoForHandler)
	uc := usecases.NewLogoutUseCase(repo, secret)
	h := NewLogoutHandler(uc)

	router := gin.New()
	router.POST("/auth/logout", h.Logout)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer not.a.token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
