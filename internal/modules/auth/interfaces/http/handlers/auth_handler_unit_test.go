package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
	authEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/repositories"
	emailServices "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/services"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// Ensure MockUserRepository implements repositories.UserRepository
var _ repositories.UserRepository = (*MockUserRepository)(nil)

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

func (m *MockUserRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int) ([]*authEntities.User, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*authEntities.User), args.Error(1)
}

// MockEmailService is a mock implementation of EmailService
type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) SendEmail(ctx context.Context, req *emailServices.SendEmailRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockEmailService) SendWelcomeEmail(ctx context.Context, recipientEmail, userName string) error {
	args := m.Called(ctx, recipientEmail, userName)
	return args.Error(0)
}

func (m *MockEmailService) SendPasswordResetEmail(ctx context.Context, recipientEmail, resetToken string) error {
	args := m.Called(ctx, recipientEmail, resetToken)
	return args.Error(0)
}

func (m *MockEmailService) SendNotification(ctx context.Context, recipientEmail, subject, body string) error {
	args := m.Called(ctx, recipientEmail, subject, body)
	return args.Error(0)
}

// Helper constants
var (
	testJWTSecret     = []byte("test-jwt-secret-for-unit-tests")
	testRefreshSecret = []byte("test-refresh-secret-for-unit-tests")
	testPassword      = "SecurePass123"
)

// hashPassword creates a bcrypt hash for testing
func hashPassword(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	return string(hash)
}

// createActiveUser creates a test user with active status
func createActiveUser(email, password string) *authEntities.User {
	return &authEntities.User{
		ID:        1,
		Email:     email,
		Password:  hashPassword(password),
		Name:      "Test User",
		Role:      domain.RoleTeacher,
		Status:    authEntities.UserStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// setupHandler creates a handler with mock dependencies
func setupHandler(mockRepo *MockUserRepository) *AuthHandler {
	uc := usecases.NewAuthUseCase(mockRepo, testJWTSecret, testRefreshSecret, nil, nil, nil)
	return NewAuthHandler(uc, nil)
}

// setupHandlerWithEmail creates a handler with mock email service
func setupHandlerWithEmail(mockRepo *MockUserRepository, emailService *MockEmailService) *AuthHandler {
	uc := usecases.NewAuthUseCase(mockRepo, testJWTSecret, testRefreshSecret, nil, nil, nil)
	return NewAuthHandler(uc, emailService)
}

// generateRefreshToken creates a valid refresh token for testing
func generateRefreshToken(userID int64) string {
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id": float64(userID),
		"exp":     now.Add(7 * 24 * time.Hour).Unix(),
		"iat":     now.Unix(),
		"jti":     uuid.New().String(),
		"iss":     "inf-sys-auth",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(testRefreshSecret)
	return tokenString
}

func TestRegister(t *testing.T) {
	t.Run("successful registration with auto-login", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		handler := setupHandler(mockRepo)

		user := createActiveUser("newuser@example.com", testPassword)

		mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
		mockRepo.On("GetByEmailForAuth", mock.Anything, "newuser@example.com").Return(user, nil)

		router := gin.New()
		router.POST("/register", handler.Register)

		payload := map[string]string{
			"email":    "newuser@example.com",
			"password": testPassword,
			"name":     "New User",
			"role":     "teacher",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, true, resp["success"])
	})

	t.Run("registration with email service", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		mockEmail := new(MockEmailService)
		handler := setupHandlerWithEmail(mockRepo, mockEmail)

		user := createActiveUser("emailuser@example.com", testPassword)

		mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
		mockRepo.On("GetByEmailForAuth", mock.Anything, "emailuser@example.com").Return(user, nil)
		mockEmail.On("SendWelcomeEmail", mock.Anything, "emailuser@example.com", mock.Anything).Return(nil)

		router := gin.New()
		router.POST("/register", handler.Register)

		payload := map[string]string{
			"email":    "emailuser@example.com",
			"password": testPassword,
			"name":     "Email User",
			"role":     "teacher",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		// Give goroutine time to send email
		time.Sleep(50 * time.Millisecond)
	})

	t.Run("invalid JSON body", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		handler := setupHandler(mockRepo)

		router := gin.New()
		router.POST("/register", handler.Register)

		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing required fields", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		handler := setupHandler(mockRepo)

		router := gin.New()
		router.POST("/register", handler.Register)

		payload := map[string]string{
			"email": "test@example.com",
			// missing password and role
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("weak password fails custom validation", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		handler := setupHandler(mockRepo)

		router := gin.New()
		router.POST("/register", handler.Register)

		payload := map[string]string{
			"email":    "user@example.com",
			"password": "12345678", // passes min=8 but no letters -> fails strong_password
			"name":     "Test User",
			"role":     "teacher",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("duplicate email", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		handler := setupHandler(mockRepo)

		mockRepo.On("Create", mock.Anything, mock.Anything).Return(
			&duplicateError{},
		)

		router := gin.New()
		router.POST("/register", handler.Register)

		payload := map[string]string{
			"email":    "existing@example.com",
			"password": testPassword,
			"name":     "Existing User",
			"role":     "teacher",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should be some error status (409 or 500 depending on error mapping)
		assert.True(t, w.Code >= 400)
	})

	t.Run("returns 403 for privileged role", func(t *testing.T) {
		privilegedRoles := []string{
			string(domain.RoleSystemAdmin),
			string(domain.RoleMethodist),
			string(domain.RoleAcademicSecretary),
		}
		for _, role := range privilegedRoles {
			t.Run(role, func(t *testing.T) {
				mockRepo := new(MockUserRepository)
				handler := setupHandler(mockRepo)

				router := gin.New()
				router.POST("/register", handler.Register)

				payload := map[string]string{
					"email":    "attacker-" + role + "@example.com",
					"password": testPassword,
					"name":     "Attacker User",
					"role":     role,
				}
				body, _ := json.Marshal(payload)
				req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusForbidden, w.Code,
					"privileged role %s must be rejected with 403", role)
				mockRepo.AssertNotCalled(t, "Create")
			})
		}
	})

	t.Run("registration succeeds but auto-login fails", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		handler := setupHandler(mockRepo)

		mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
		// Auto-login fails because GetByEmailForAuth returns an error
		mockRepo.On("GetByEmailForAuth", mock.Anything, "faillogin@example.com").Return(nil, assert.AnError)

		router := gin.New()
		router.POST("/register", handler.Register)

		payload := map[string]string{
			"email":    "faillogin@example.com",
			"password": testPassword,
			"name":     "Fail Login User",
			"role":     "teacher",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
	})
}

func TestLogin(t *testing.T) {
	t.Run("successful login", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		handler := setupHandler(mockRepo)

		user := createActiveUser("login@example.com", testPassword)
		mockRepo.On("GetByEmailForAuth", mock.Anything, "login@example.com").Return(user, nil)

		router := gin.New()
		router.POST("/login", handler.Login)

		payload := map[string]string{
			"email":    "login@example.com",
			"password": testPassword,
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, true, resp["success"])

		data, ok := resp["data"].(map[string]interface{})
		assert.True(t, ok)
		assert.NotEmpty(t, data["token"])
		assert.NotEmpty(t, data["refreshToken"])
		assert.NotNil(t, data["user"])
	})

	t.Run("invalid JSON body", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		handler := setupHandler(mockRepo)

		router := gin.New()
		router.POST("/login", handler.Login)

		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader([]byte("not json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing email", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		handler := setupHandler(mockRepo)

		router := gin.New()
		router.POST("/login", handler.Login)

		payload := map[string]string{
			"password": testPassword,
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("SQL injection in email fails custom validation", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		handler := setupHandler(mockRepo)

		router := gin.New()
		router.POST("/login", handler.Login)

		// "select" in local part passes Gin email validation but fails no_sql_injection
		payload := map[string]string{
			"email":    "select@example.com",
			"password": testPassword,
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("wrong password", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		handler := setupHandler(mockRepo)

		user := createActiveUser("login@example.com", testPassword)
		mockRepo.On("GetByEmailForAuth", mock.Anything, "login@example.com").Return(user, nil)

		router := gin.New()
		router.POST("/login", handler.Login)

		payload := map[string]string{
			"email":    "login@example.com",
			"password": "WrongPassword123",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("non-existent user", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		handler := setupHandler(mockRepo)

		mockRepo.On("GetByEmailForAuth", mock.Anything, "nobody@example.com").Return(nil, assert.AnError)

		router := gin.New()
		router.POST("/login", handler.Login)

		payload := map[string]string{
			"email":    "nobody@example.com",
			"password": testPassword,
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestRefreshToken(t *testing.T) {
	t.Run("successful token refresh", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		handler := setupHandler(mockRepo)

		user := createActiveUser("refresh@example.com", testPassword)
		refreshToken := generateRefreshToken(user.ID)

		mockRepo.On("GetByID", mock.Anything, user.ID).Return(user, nil)

		router := gin.New()
		router.POST("/refresh", handler.RefreshToken)

		payload := map[string]string{
			"refresh_token": refreshToken,
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, true, resp["success"])
	})

	t.Run("invalid JSON body", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		handler := setupHandler(mockRepo)

		router := gin.New()
		router.POST("/refresh", handler.RefreshToken)

		req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewReader([]byte("not json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing refresh token", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		handler := setupHandler(mockRepo)

		router := gin.New()
		router.POST("/refresh", handler.RefreshToken)

		payload := map[string]string{}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		handler := setupHandler(mockRepo)

		router := gin.New()
		router.POST("/refresh", handler.RefreshToken)

		payload := map[string]string{
			"refresh_token": "invalid-token-string",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Token validation fails
		assert.True(t, w.Code >= 400)
	})

	t.Run("SQL injection in refresh token fails custom validation", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		handler := setupHandler(mockRepo)

		router := gin.New()
		router.POST("/refresh", handler.RefreshToken)

		payload := map[string]string{
			"refresh_token": "select * from users; --",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("expired refresh token", func(t *testing.T) {
		mockRepo := new(MockUserRepository)
		handler := setupHandler(mockRepo)

		// Create expired token
		now := time.Now()
		claims := jwt.MapClaims{
			"user_id": float64(1),
			"exp":     now.Add(-1 * time.Hour).Unix(),
			"iat":     now.Add(-2 * time.Hour).Unix(),
			"jti":     uuid.New().String(),
			"iss":     "inf-sys-auth",
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		expiredToken, _ := token.SignedString(testRefreshSecret)

		router := gin.New()
		router.POST("/refresh", handler.RefreshToken)

		payload := map[string]string{
			"refresh_token": expiredToken,
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.True(t, w.Code >= 400)
	})
}

// duplicateError implements error for testing duplicate key scenarios
type duplicateError struct{}

func (e *duplicateError) Error() string {
	return "duplicate key value violates unique constraint"
}
