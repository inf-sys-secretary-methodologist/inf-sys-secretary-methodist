package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/application/usecases"
	authHandler "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/interfaces/http/handlers"
	persistence "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/infrastructure"
	testSuite "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/testing/suite"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/testing/fixtures"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/testing/helpers"
)

// AuthHandlerTestSuite tests auth HTTP handlers
type AuthHandlerTestSuite struct {
	testSuite.IntegrationSuite
	handler *authHandler.AuthHandler
	usecase *usecases.AuthUseCase
	router  *gin.Engine
}

// SetupSuite runs once before all tests
func (s *AuthHandlerTestSuite) SetupSuite() {
	s.IntegrationSuite.SetupSuite()

	// Setup dependencies
	repo := persistence.NewUserRepositoryPG(s.DB)
	jwtSecret := []byte("test-secret")
	refreshSecret := []byte("test-refresh-secret")
	s.usecase = usecases.NewAuthUseCase(repo, jwtSecret, refreshSecret, nil, nil)
	s.handler = authHandler.NewAuthHandler(s.usecase)

	// Setup router
	gin.SetMode(gin.TestMode)
	s.router = gin.New()
	s.router.POST("/register", s.handler.Register)
	s.router.POST("/login", s.handler.Login)
	s.router.POST("/refresh", s.handler.RefreshToken)
}

// TearDownTest runs after each test
func (s *AuthHandlerTestSuite) TearDownTest() {
	s.TruncateTables("users")
}

// TestRegister tests user registration endpoint
func (s *AuthHandlerTestSuite) TestRegister() {
	payload := map[string]interface{}{
		"email":    "newuser@example.com",
		"password": "SecurePass123",
		"role":     "student",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("success", response["status"])
}

// TestRegisterDuplicateEmail tests registration with duplicate email
func (s *AuthHandlerTestSuite) TestRegisterDuplicateEmail() {
	ctx := helpers.TestContext(s.T())

	// Create existing user
	user := fixtures.NewUserBuilder().
		WithEmail("existing@example.com").
		Build()
	repo := persistence.NewUserRepositoryPG(s.DB)
	err := repo.Create(ctx, user)
	s.NoError(err)

	// Try to register with same email
	payload := map[string]interface{}{
		"email":    "existing@example.com",
		"password": "SecurePass123",
		"role":     "student",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusConflict, w.Code)
}

// TestRegisterInvalidPayload tests registration with invalid data
func (s *AuthHandlerTestSuite) TestRegisterInvalidPayload() {
	testCases := []struct {
		name    string
		payload map[string]interface{}
	}{
		{
			name: "missing email",
			payload: map[string]interface{}{
				"password": "SecurePass123",
				"role":     "student",
			},
		},
		{
			name: "invalid email",
			payload: map[string]interface{}{
				"email":    "notanemail",
				"password": "SecurePass123",
				"role":     "student",
			},
		},
		{
			name: "weak password",
			payload: map[string]interface{}{
				"email":    "user@example.com",
				"password": "weak",
				"role":     "student",
			},
		},
		{
			name: "invalid role",
			payload: map[string]interface{}{
				"email":    "user@example.com",
				"password": "SecurePass123",
				"role":     "invalidrole",
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			body, _ := json.Marshal(tc.payload)
			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			s.router.ServeHTTP(w, req)

			s.Equal(http.StatusBadRequest, w.Code)
		})
	}
}

// TestLogin tests user login endpoint
func (s *AuthHandlerTestSuite) TestLogin() {
	ctx := helpers.TestContext(s.T())

	// Create user with known password
	password := "Test123456"
	user := fixtures.UserWithKnownPassword(password)
	user.Email = "testlogin@example.com"

	repo := persistence.NewUserRepositoryPG(s.DB)
	err := repo.Create(ctx, user)
	s.NoError(err)

	// Login
	payload := map[string]interface{}{
		"email":    "testlogin@example.com",
		"password": password,
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.Equal("success", response["status"])

	data, ok := response["data"].(map[string]interface{})
	s.True(ok)
	s.NotEmpty(data["access_token"])
	s.NotEmpty(data["refresh_token"])
}

// TestLoginInvalidCredentials tests login with wrong password
func (s *AuthHandlerTestSuite) TestLoginInvalidCredentials() {
	ctx := helpers.TestContext(s.T())

	// Create user
	user := fixtures.UserWithKnownPassword("CorrectPass123")
	user.Email = "testbadlogin@example.com"

	repo := persistence.NewUserRepositoryPG(s.DB)
	err := repo.Create(ctx, user)
	s.NoError(err)

	// Try login with wrong password
	payload := map[string]interface{}{
		"email":    "testbadlogin@example.com",
		"password": "WrongPassword123",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusUnauthorized, w.Code)
}

// TestLoginNonExistentUser tests login with non-existent email
func (s *AuthHandlerTestSuite) TestLoginNonExistentUser() {
	payload := map[string]interface{}{
		"email":    "nonexistent@example.com",
		"password": "SomePassword123",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusUnauthorized, w.Code)
}

// TestSuite runs the test suite
func TestAuthHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(AuthHandlerTestSuite))
}
