package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/application/usecases"
	authEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// Ensure MockUserRepository implements usecases.UserRepository
var _ usecases.UserRepository = (*MockUserRepository)(nil)

// MockUserRepository is a mock implementation of UserRepository for testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *authEntities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Save(ctx context.Context, user *authEntities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int64) (*authEntities.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authEntities.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*authEntities.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authEntities.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmailForAuth(ctx context.Context, email string) (*authEntities.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authEntities.User), args.Error(1)
}

func (m *MockUserRepository) GetByIDForAuth(ctx context.Context, id int64) (*authEntities.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authEntities.User), args.Error(1)
}

func (m *MockUserRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int) ([]*authEntities.User, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*authEntities.User), args.Error(1)
}

// generateTestToken creates a valid JWT token for testing
func generateTestToken(secret []byte, userID int64, role string) string {
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id": float64(userID),
		"role":    role,
		"exp":     now.Add(15 * time.Minute).Unix(),
		"iat":     now.Unix(),
		"nbf":     now.Unix(),
		"jti":     uuid.New().String(),
		"iss":     "inf-sys-auth",
		"aud":     "inf-sys-api",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(secret)
	return tokenString
}

// generateExpiredToken creates an expired JWT token for testing
func generateExpiredToken(secret []byte, userID int64, role string) string {
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id": float64(userID),
		"role":    role,
		"exp":     now.Add(-1 * time.Hour).Unix(),
		"iat":     now.Add(-2 * time.Hour).Unix(),
		"jti":     uuid.New().String(),
		"iss":     "inf-sys-auth",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(secret)
	return tokenString
}

func TestJWTMiddleware(t *testing.T) {
	jwtSecret := []byte("test-jwt-secret-key")
	refreshSecret := []byte("test-refresh-secret-key")
	mockRepo := new(MockUserRepository)
	authUseCase := usecases.NewAuthUseCase(mockRepo, jwtSecret, refreshSecret, []byte("mfa-intermediate"), nil, nil, nil)

	t.Run("valid bearer token in Authorization header", func(t *testing.T) {
		router := gin.New()
		router.Use(JWTMiddleware(authUseCase))
		router.GET("/test", func(c *gin.Context) {
			userID, _ := c.Get("user_id")
			role, _ := c.Get("role")
			c.JSON(http.StatusOK, gin.H{"user_id": userID, "role": role})
		})

		token := generateTestToken(jwtSecret, 42, "admin")
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("valid token in query parameter", func(t *testing.T) {
		router := gin.New()
		router.Use(JWTMiddleware(authUseCase))
		router.GET("/test", func(c *gin.Context) {
			userID, _ := c.Get("user_id")
			c.JSON(http.StatusOK, gin.H{"user_id": userID})
		})

		token := generateTestToken(jwtSecret, 42, "admin")
		req := httptest.NewRequest(http.MethodGet, "/test?token="+token, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("missing token", func(t *testing.T) {
		router := gin.New()
		router.Use(JWTMiddleware(authUseCase))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid Authorization header without Bearer prefix", func(t *testing.T) {
		router := gin.New()
		router.Use(JWTMiddleware(authUseCase))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Basic some-token")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("expired token", func(t *testing.T) {
		router := gin.New()
		router.Use(JWTMiddleware(authUseCase))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		token := generateExpiredToken(jwtSecret, 42, "admin")
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("token signed with wrong secret", func(t *testing.T) {
		router := gin.New()
		router.Use(JWTMiddleware(authUseCase))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		token := generateTestToken([]byte("wrong-secret"), 42, "admin")
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("malformed token string", func(t *testing.T) {
		router := gin.New()
		router.Use(JWTMiddleware(authUseCase))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer not-a-valid-jwt-token")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("sets user_id and role from claims", func(t *testing.T) {
		var gotUserID int64
		var gotRole string

		router := gin.New()
		router.Use(JWTMiddleware(authUseCase))
		router.GET("/test", func(c *gin.Context) {
			uid, _ := c.Get("user_id")
			gotUserID = uid.(int64)
			r, _ := c.Get("role")
			gotRole = r.(string)
			c.JSON(http.StatusOK, gin.H{})
		})

		token := generateTestToken(jwtSecret, 99, "teacher")
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, int64(99), gotUserID)
		assert.Equal(t, "teacher", gotRole)
	})
}

func TestRequireRole(t *testing.T) {
	t.Run("allows user with matching role", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("role", "admin")
			c.Next()
		})
		router.Use(RequireRole("admin", "moderator"))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("denies user with non-matching role", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("role", "student")
			c.Next()
		})
		router.Use(RequireRole("admin", "moderator"))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("denies when role is not set", func(t *testing.T) {
		router := gin.New()
		router.Use(RequireRole("admin"))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("denies when role is wrong type", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("role", 123) // wrong type
			c.Next()
		})
		router.Use(RequireRole("admin"))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	router := gin.New()
	router.Use(SecurityHeadersMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "max-age=31536000; includeSubDomains", w.Header().Get("Strict-Transport-Security"))
	assert.Equal(t, "default-src 'self'", w.Header().Get("Content-Security-Policy"))
	assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
}

// TestRateLimiter / TestNewRateLimiter / TestRateLimitMiddleware
// removed in v0.159.0 ADR-3 Tier 2 — the in-memory RateLimiter under
// test is gone (replaced by the Redis-backed shared limiter). See
// internal/shared/infrastructure/middleware/rate_limiting_test.go for
// the production limiter's test coverage including the new trusted-
// proxy CIDR variants.
