package usecases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"
	notifEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/entities"
	notifRepos "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/repositories"
	notifUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

// Ensure notifRepos is used (compile check)
var _ notifRepos.NotificationRepository = (*stubNotificationRepo)(nil)

// stubNotificationRepo is a minimal mock for NotificationRepository
type stubNotificationRepo struct{}

func (s *stubNotificationRepo) Create(_ context.Context, n *notifEntities.Notification) error {
	n.ID = 1
	return nil
}
func (s *stubNotificationRepo) Update(_ context.Context, _ *notifEntities.Notification) error {
	return nil
}
func (s *stubNotificationRepo) Delete(_ context.Context, _ int64) error        { return nil }
func (s *stubNotificationRepo) GetByID(_ context.Context, _ int64) (*notifEntities.Notification, error) {
	return nil, nil
}
func (s *stubNotificationRepo) List(_ context.Context, _ *notifEntities.NotificationFilter) ([]*notifEntities.Notification, error) {
	return nil, nil
}
func (s *stubNotificationRepo) GetByUserID(_ context.Context, _ int64, _, _ int) ([]*notifEntities.Notification, error) {
	return nil, nil
}
func (s *stubNotificationRepo) GetUnreadByUserID(_ context.Context, _ int64) ([]*notifEntities.Notification, error) {
	return nil, nil
}
func (s *stubNotificationRepo) MarkAsRead(_ context.Context, _ int64) error    { return nil }
func (s *stubNotificationRepo) MarkAllAsRead(_ context.Context, _ int64) error { return nil }
func (s *stubNotificationRepo) DeleteByUserID(_ context.Context, _ int64) error {
	return nil
}
func (s *stubNotificationRepo) DeleteExpired(_ context.Context) (int64, error) { return 0, nil }
func (s *stubNotificationRepo) GetUnreadCount(_ context.Context, _ int64) (int64, error) {
	return 0, nil
}
func (s *stubNotificationRepo) GetStats(_ context.Context, _ int64) (*notifEntities.NotificationStats, error) {
	return nil, nil
}
func (s *stubNotificationRepo) CreateBulk(_ context.Context, _ []*notifEntities.Notification) error {
	return nil
}

func newTestNotifUC() *notifUsecases.NotificationUseCase {
	return notifUsecases.NewNotificationUseCase(&stubNotificationRepo{}, nil, nil, nil, nil, nil)
}

// Mock repository for testing
type mockUserRepository struct {
	users       map[string]*entities.User
	shouldError bool
	createError bool
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users: make(map[string]*entities.User),
	}
}

func (m *mockUserRepository) Create(_ context.Context, user *entities.User) error {
	if m.createError {
		return errors.New("database error")
	}
	user.ID = 1
	m.users[user.Email] = user
	return nil
}

func (m *mockUserRepository) Save(_ context.Context, _ *entities.User) error {
	return nil
}

func (m *mockUserRepository) GetByID(_ context.Context, id int64) (*entities.User, error) {
	if m.shouldError {
		return nil, errors.New("user not found")
	}
	for _, user := range m.users {
		if user.ID == id {
			return user, nil
		}
	}
	return &entities.User{
		ID:       1,
		Email:    "test@example.com",
		Password: "$2a$14$dlaPdzfteiUkTlRwHMp/DuuMyviurDbsQnBwQ1MPSuUM4VnpyQJBK",
		Role:     domain.RoleSystemAdmin,
		Status:   entities.UserStatusActive,
	}, nil
}

func (m *mockUserRepository) GetByEmail(_ context.Context, email string) (*entities.User, error) {
	if m.shouldError {
		return nil, errors.New("user not found")
	}
	if user, ok := m.users[email]; ok {
		return user, nil
	}
	// Password is: Admin123456!
	// Hash generated with bcrypt cost 14
	return &entities.User{
		ID:       1,
		Email:    email,
		Password: "$2a$14$dlaPdzfteiUkTlRwHMp/DuuMyviurDbsQnBwQ1MPSuUM4VnpyQJBK",
		Role:     domain.RoleSystemAdmin,
		Status:   entities.UserStatusActive,
	}, nil
}

func (m *mockUserRepository) GetByEmailForAuth(ctx context.Context, email string) (*entities.User, error) {
	// Same as GetByEmail for mock
	return m.GetByEmail(ctx, email)
}

func (m *mockUserRepository) Delete(_ context.Context, _ int64) error {
	return nil
}

func (m *mockUserRepository) List(_ context.Context, _, _ int) ([]*entities.User, error) {
	return []*entities.User{}, nil
}

func TestRegister(t *testing.T) {
	repo := newMockUserRepository()
	useCase := usecases.NewAuthUseCase(repo, []byte("secret"), []byte("refresh"), nil, nil, nil)

	input := dto.RegisterInput{
		Email:    "newuser@example.com",
		Password: "SecurePass123!",
		Role:     "admin",
	}

	err := useCase.Register(context.Background(), input)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
}

func TestLogin(t *testing.T) {
	repo := newMockUserRepository()
	useCase := usecases.NewAuthUseCase(repo, []byte("secret"), []byte("refresh"), nil, nil, nil)

	input := dto.LoginInput{
		Email:    "test@example.com",
		Password: "Admin123456!",
	}

	accessToken, refreshToken, err := useCase.Login(context.Background(), input)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if accessToken == "" || refreshToken == "" {
		t.Fatal("Tokens should not be empty")
	}
}

func TestValidateAccessToken(t *testing.T) {
	repo := newMockUserRepository()
	useCase := usecases.NewAuthUseCase(repo, []byte("secret"), []byte("refresh"), nil, nil, nil)

	// First login to get a token
	input := dto.LoginInput{
		Email:    "test@example.com",
		Password: "Admin123456!",
	}

	accessToken, _, err := useCase.Login(context.Background(), input)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	// Validate the token
	claims, err := useCase.ValidateAccessToken(context.Background(), accessToken)
	if err != nil {
		t.Fatalf("Token validation failed: %v", err)
	}

	if claims == nil {
		t.Fatal("Claims should not be nil")
	}

	userID, ok := (*claims)["user_id"]
	if !ok || userID == nil {
		t.Fatal("user_id claim missing")
	}

	// Verify all security claims are present
	role, ok := (*claims)["role"]
	if !ok || role == nil {
		t.Fatal("role claim missing")
	}

	jti, ok := (*claims)["jti"]
	if !ok || jti == nil {
		t.Fatal("jti claim missing")
	}

	iss, ok := (*claims)["iss"]
	if !ok || iss == nil {
		t.Fatal("iss claim missing")
	}
	assert.Equal(t, "inf-sys-auth", iss)

	aud, ok := (*claims)["aud"]
	if !ok || aud == nil {
		t.Fatal("aud claim missing")
	}
	assert.Equal(t, "inf-sys-api", aud)
}

func TestLoginWithUser(t *testing.T) {
	repo := newMockUserRepository()
	useCase := usecases.NewAuthUseCase(repo, []byte("secret"), []byte("refresh"), nil, nil, nil)

	t.Run("successful login returns user", func(t *testing.T) {
		input := dto.LoginInput{
			Email:    "test@example.com",
			Password: "Admin123456!",
		}

		accessToken, refreshToken, user, err := useCase.LoginWithUser(context.Background(), input)

		assert.NoError(t, err)
		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, refreshToken)
		assert.NotNil(t, user)
		assert.Equal(t, "test@example.com", user.Email)
	})

	t.Run("invalid password returns error", func(t *testing.T) {
		input := dto.LoginInput{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}

		accessToken, refreshToken, user, err := useCase.LoginWithUser(context.Background(), input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "authentication failed")
		assert.Empty(t, accessToken)
		assert.Empty(t, refreshToken)
		assert.Nil(t, user)
	})

	t.Run("user not found returns error", func(t *testing.T) {
		repo := newMockUserRepository()
		repo.shouldError = true
		useCase := usecases.NewAuthUseCase(repo, []byte("secret"), []byte("refresh"), nil, nil, nil)

		input := dto.LoginInput{
			Email:    "notfound@example.com",
			Password: "Admin123456!",
		}

		accessToken, refreshToken, user, err := useCase.LoginWithUser(context.Background(), input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "authentication failed")
		assert.Empty(t, accessToken)
		assert.Empty(t, refreshToken)
		assert.Nil(t, user)
	})

	t.Run("blocked user cannot login", func(t *testing.T) {
		repo := newMockUserRepository()
		repo.users["blocked@example.com"] = &entities.User{
			ID:       2,
			Email:    "blocked@example.com",
			Password: "$2a$14$dlaPdzfteiUkTlRwHMp/DuuMyviurDbsQnBwQ1MPSuUM4VnpyQJBK",
			Role:     domain.RoleSystemAdmin,
			Status:   entities.UserStatusBlocked,
		}
		useCase := usecases.NewAuthUseCase(repo, []byte("secret"), []byte("refresh"), nil, nil, nil)

		input := dto.LoginInput{
			Email:    "blocked@example.com",
			Password: "Admin123456!",
		}

		accessToken, refreshToken, user, err := useCase.LoginWithUser(context.Background(), input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot login")
		assert.Empty(t, accessToken)
		assert.Empty(t, refreshToken)
		assert.Nil(t, user)
	})

	t.Run("inactive user cannot login", func(t *testing.T) {
		repo := newMockUserRepository()
		repo.users["inactive@example.com"] = &entities.User{
			ID:       3,
			Email:    "inactive@example.com",
			Password: "$2a$14$dlaPdzfteiUkTlRwHMp/DuuMyviurDbsQnBwQ1MPSuUM4VnpyQJBK",
			Role:     domain.RoleSystemAdmin,
			Status:   entities.UserStatusInactive,
		}
		useCase := usecases.NewAuthUseCase(repo, []byte("secret"), []byte("refresh"), nil, nil, nil)

		input := dto.LoginInput{
			Email:    "inactive@example.com",
			Password: "Admin123456!",
		}

		accessToken, refreshToken, user, err := useCase.LoginWithUser(context.Background(), input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot login")
		assert.Empty(t, accessToken)
		assert.Empty(t, refreshToken)
		assert.Nil(t, user)
	})
}

func TestLogin_InvalidPassword(t *testing.T) {
	repo := newMockUserRepository()
	useCase := usecases.NewAuthUseCase(repo, []byte("secret"), []byte("refresh"), nil, nil, nil)

	input := dto.LoginInput{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	accessToken, refreshToken, err := useCase.Login(context.Background(), input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authentication failed")
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
}

func TestLogin_UserNotFound(t *testing.T) {
	repo := newMockUserRepository()
	repo.shouldError = true
	useCase := usecases.NewAuthUseCase(repo, []byte("secret"), []byte("refresh"), nil, nil, nil)

	input := dto.LoginInput{
		Email:    "notfound@example.com",
		Password: "Admin123456!",
	}

	accessToken, refreshToken, err := useCase.Login(context.Background(), input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authentication failed")
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
}

func TestLogin_BlockedUser(t *testing.T) {
	repo := newMockUserRepository()
	repo.users["blocked@example.com"] = &entities.User{
		ID:       2,
		Email:    "blocked@example.com",
		Password: "$2a$14$dlaPdzfteiUkTlRwHMp/DuuMyviurDbsQnBwQ1MPSuUM4VnpyQJBK",
		Role:     domain.RoleSystemAdmin,
		Status:   entities.UserStatusBlocked,
	}
	useCase := usecases.NewAuthUseCase(repo, []byte("secret"), []byte("refresh"), nil, nil, nil)

	input := dto.LoginInput{
		Email:    "blocked@example.com",
		Password: "Admin123456!",
	}

	accessToken, refreshToken, err := useCase.Login(context.Background(), input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot login")
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
}

func TestRefreshToken(t *testing.T) {
	repo := newMockUserRepository()
	useCase := usecases.NewAuthUseCase(repo, []byte("secret"), []byte("refresh"), nil, nil, nil)

	t.Run("successful token refresh", func(t *testing.T) {
		// First login to get a refresh token
		input := dto.LoginInput{
			Email:    "test@example.com",
			Password: "Admin123456!",
		}
		_, refreshToken, err := useCase.Login(context.Background(), input)
		assert.NoError(t, err)

		// Refresh the token
		newAccessToken, newRefreshToken, err := useCase.RefreshToken(context.Background(), refreshToken)

		assert.NoError(t, err)
		assert.NotEmpty(t, newAccessToken)
		assert.NotEmpty(t, newRefreshToken)
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		newAccessToken, newRefreshToken, err := useCase.RefreshToken(context.Background(), "invalid-token")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid refresh token")
		assert.Empty(t, newAccessToken)
		assert.Empty(t, newRefreshToken)
	})

	t.Run("expired refresh token", func(t *testing.T) {
		// This is an expired token (exp in the past)
		expiredToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJleHAiOjE2MDAwMDAwMDAsImlhdCI6MTYwMDAwMDAwMH0.invalid"

		newAccessToken, newRefreshToken, err := useCase.RefreshToken(context.Background(), expiredToken)

		assert.Error(t, err)
		assert.Empty(t, newAccessToken)
		assert.Empty(t, newRefreshToken)
	})

	t.Run("user not found during refresh", func(t *testing.T) {
		repo := newMockUserRepository()
		useCase := usecases.NewAuthUseCase(repo, []byte("secret"), []byte("refresh"), nil, nil, nil)

		// First login to get a valid refresh token
		input := dto.LoginInput{
			Email:    "test@example.com",
			Password: "Admin123456!",
		}
		_, refreshToken, err := useCase.Login(context.Background(), input)
		assert.NoError(t, err)

		// Now simulate user being deleted
		repo.shouldError = true

		newAccessToken, newRefreshToken, err := useCase.RefreshToken(context.Background(), refreshToken)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
		assert.Empty(t, newAccessToken)
		assert.Empty(t, newRefreshToken)
	})

	t.Run("blocked user cannot refresh token", func(t *testing.T) {
		repo := newMockUserRepository()
		repo.users["testblocked@example.com"] = &entities.User{
			ID:       5,
			Email:    "testblocked@example.com",
			Password: "$2a$14$dlaPdzfteiUkTlRwHMp/DuuMyviurDbsQnBwQ1MPSuUM4VnpyQJBK",
			Role:     domain.RoleSystemAdmin,
			Status:   entities.UserStatusActive,
		}
		useCase := usecases.NewAuthUseCase(repo, []byte("secret"), []byte("refresh"), nil, nil, nil)

		// First login with active user
		input := dto.LoginInput{
			Email:    "testblocked@example.com",
			Password: "Admin123456!",
		}
		_, refreshToken, err := useCase.Login(context.Background(), input)
		assert.NoError(t, err)

		// Block the user
		repo.users["testblocked@example.com"].Status = entities.UserStatusBlocked

		// Try to refresh
		newAccessToken, newRefreshToken, err := useCase.RefreshToken(context.Background(), refreshToken)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot refresh token")
		assert.Empty(t, newAccessToken)
		assert.Empty(t, newRefreshToken)
	})
}

func TestValidateAccessToken_Invalid(t *testing.T) {
	repo := newMockUserRepository()
	useCase := usecases.NewAuthUseCase(repo, []byte("secret"), []byte("refresh"), nil, nil, nil)

	t.Run("invalid token format", func(t *testing.T) {
		claims, err := useCase.ValidateAccessToken(context.Background(), "invalid-token")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid access token")
		assert.Nil(t, claims)
	})

	t.Run("token with wrong secret", func(t *testing.T) {
		// Create a token with different secret
		otherUseCase := usecases.NewAuthUseCase(repo, []byte("other-secret"), []byte("refresh"), nil, nil, nil)

		input := dto.LoginInput{
			Email:    "test@example.com",
			Password: "Admin123456!",
		}
		accessToken, _, _ := otherUseCase.Login(context.Background(), input)

		// Try to validate with original useCase (different secret)
		claims, err := useCase.ValidateAccessToken(context.Background(), accessToken)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid access token")
		assert.Nil(t, claims)
	})

	t.Run("token with wrong signing method", func(t *testing.T) {
		// Create a token with RSA signing method header but HMAC body (will fail validation)
		// We can't easily create such a token, so test with none algorithm
		token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
			"user_id": 1,
			"role":    "admin",
			"exp":     time.Now().Add(time.Hour).Unix(),
		})
		tokenString, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

		claims, err := useCase.ValidateAccessToken(context.Background(), tokenString)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})
}

func TestRegister_Errors(t *testing.T) {
	t.Run("database error on create", func(t *testing.T) {
		repo := newMockUserRepository()
		repo.createError = true
		useCase := usecases.NewAuthUseCase(repo, []byte("secret"), []byte("refresh"), nil, nil, nil)

		input := dto.RegisterInput{
			Email:    "newuser@example.com",
			Password: "SecurePass123!",
			Role:     "admin",
		}

		err := useCase.Register(context.Background(), input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create user")
	})
}

func TestRegister_WithName(t *testing.T) {
	repo := newMockUserRepository()
	useCase := usecases.NewAuthUseCase(repo, []byte("secret"), []byte("refresh"), nil, nil, nil)

	input := dto.RegisterInput{
		Email:    "named@example.com",
		Password: "SecurePass123!",
		Name:     "Test User",
		Role:     string(domain.RoleTeacher),
	}

	err := useCase.Register(context.Background(), input)
	assert.NoError(t, err)

	// User should be in the repo
	user, ok := repo.users["named@example.com"]
	assert.True(t, ok)
	assert.Equal(t, "Test User", user.Name)
}


func TestLogin_InactiveUser(t *testing.T) {
	repo := newMockUserRepository()
	repo.users["inactive@example.com"] = &entities.User{
		ID:       3,
		Email:    "inactive@example.com",
		Password: "$2a$14$dlaPdzfteiUkTlRwHMp/DuuMyviurDbsQnBwQ1MPSuUM4VnpyQJBK",
		Role:     domain.RoleSystemAdmin,
		Status:   entities.UserStatusInactive,
	}
	useCase := usecases.NewAuthUseCase(repo, []byte("secret"), []byte("refresh"), nil, nil, nil)

	input := dto.LoginInput{
		Email:    "inactive@example.com",
		Password: "Admin123456!",
	}

	accessToken, refreshToken, err := useCase.Login(context.Background(), input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot login")
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
}

func TestValidateAccessToken_EmptyString(t *testing.T) {
	repo := newMockUserRepository()
	useCase := usecases.NewAuthUseCase(repo, []byte("secret"), []byte("refresh"), nil, nil, nil)

	claims, err := useCase.ValidateAccessToken(context.Background(), "")
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestRefreshToken_WithAccessToken(t *testing.T) {
	// Access token should not work as refresh token (different secret)
	repo := newMockUserRepository()
	useCase := usecases.NewAuthUseCase(repo, []byte("access-secret"), []byte("refresh-secret"), nil, nil, nil)

	input := dto.LoginInput{
		Email:    "test@example.com",
		Password: "Admin123456!",
	}
	accessToken, _, err := useCase.Login(context.Background(), input)
	assert.NoError(t, err)

	// Try to use access token as refresh token
	_, _, err = useCase.RefreshToken(context.Background(), accessToken)
	assert.Error(t, err)
}

func TestNewAuthUseCase(t *testing.T) {
	repo := newMockUserRepository()
	uc := usecases.NewAuthUseCase(repo, []byte("jwt"), []byte("refresh"), nil, nil, nil)
	assert.NotNil(t, uc)
}

func TestLogin_WithLoggers(t *testing.T) {
	repo := newMockUserRepository()
	logger := logging.NewLogger("debug")
	secLog := logging.NewSecurityLogger(logger)
	auditLog := logging.NewAuditLogger(logger)
	useCase := usecases.NewAuthUseCase(repo, []byte("secret"), []byte("refresh"), secLog, auditLog, nil)

	t.Run("successful login with loggers", func(t *testing.T) {
		input := dto.LoginInput{
			Email:    "test@example.com",
			Password: "Admin123456!",
		}
		accessToken, refreshToken, err := useCase.Login(context.Background(), input)
		assert.NoError(t, err)
		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, refreshToken)
	})

	t.Run("failed login with loggers - wrong password", func(t *testing.T) {
		input := dto.LoginInput{
			Email:    "test@example.com",
			Password: "wrong",
		}
		_, _, err := useCase.Login(context.Background(), input)
		assert.Error(t, err)
	})

	t.Run("failed login with loggers - user not found", func(t *testing.T) {
		repo := newMockUserRepository()
		repo.shouldError = true
		uc := usecases.NewAuthUseCase(repo, []byte("secret"), []byte("refresh"), secLog, auditLog, nil)
		input := dto.LoginInput{
			Email:    "notfound@example.com",
			Password: "Admin123456!",
		}
		_, _, err := uc.Login(context.Background(), input)
		assert.Error(t, err)
	})

	t.Run("failed login with loggers - blocked user", func(t *testing.T) {
		repo := newMockUserRepository()
		repo.users["blocked@test.com"] = &entities.User{
			ID:       99,
			Email:    "blocked@test.com",
			Password: "$2a$14$dlaPdzfteiUkTlRwHMp/DuuMyviurDbsQnBwQ1MPSuUM4VnpyQJBK",
			Role:     domain.RoleSystemAdmin,
			Status:   entities.UserStatusBlocked,
		}
		uc := usecases.NewAuthUseCase(repo, []byte("secret"), []byte("refresh"), secLog, auditLog, nil)
		input := dto.LoginInput{
			Email:    "blocked@test.com",
			Password: "Admin123456!",
		}
		_, _, err := uc.Login(context.Background(), input)
		assert.Error(t, err)
	})
}

func TestLoginWithUser_WithLoggers(t *testing.T) {
	repo := newMockUserRepository()
	logger := logging.NewLogger("debug")
	secLog := logging.NewSecurityLogger(logger)
	auditLog := logging.NewAuditLogger(logger)
	useCase := usecases.NewAuthUseCase(repo, []byte("secret"), []byte("refresh"), secLog, auditLog, nil)

	t.Run("successful login with user and loggers", func(t *testing.T) {
		input := dto.LoginInput{
			Email:    "test@example.com",
			Password: "Admin123456!",
		}
		accessToken, refreshToken, user, err := useCase.LoginWithUser(context.Background(), input)
		assert.NoError(t, err)
		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, refreshToken)
		assert.NotNil(t, user)
	})

	t.Run("failed login with user - wrong password with loggers", func(t *testing.T) {
		input := dto.LoginInput{
			Email:    "test@example.com",
			Password: "wrong",
		}
		_, _, _, err := useCase.LoginWithUser(context.Background(), input)
		assert.Error(t, err)
	})

	t.Run("failed login with user - not found with loggers", func(t *testing.T) {
		repo := newMockUserRepository()
		repo.shouldError = true
		uc := usecases.NewAuthUseCase(repo, []byte("secret"), []byte("refresh"), secLog, auditLog, nil)
		input := dto.LoginInput{
			Email:    "notfound@test.com",
			Password: "Admin123456!",
		}
		_, _, _, err := uc.LoginWithUser(context.Background(), input)
		assert.Error(t, err)
	})

	t.Run("failed login with user - blocked with loggers", func(t *testing.T) {
		repo := newMockUserRepository()
		repo.users["blocked2@test.com"] = &entities.User{
			ID:       88,
			Email:    "blocked2@test.com",
			Password: "$2a$14$dlaPdzfteiUkTlRwHMp/DuuMyviurDbsQnBwQ1MPSuUM4VnpyQJBK",
			Role:     domain.RoleSystemAdmin,
			Status:   entities.UserStatusBlocked,
		}
		uc := usecases.NewAuthUseCase(repo, []byte("secret"), []byte("refresh"), secLog, auditLog, nil)
		input := dto.LoginInput{
			Email:    "blocked2@test.com",
			Password: "Admin123456!",
		}
		_, _, _, err := uc.LoginWithUser(context.Background(), input)
		assert.Error(t, err)
	})

	t.Run("failed login with user - inactive with loggers", func(t *testing.T) {
		repo := newMockUserRepository()
		repo.users["inactive2@test.com"] = &entities.User{
			ID:       77,
			Email:    "inactive2@test.com",
			Password: "$2a$14$dlaPdzfteiUkTlRwHMp/DuuMyviurDbsQnBwQ1MPSuUM4VnpyQJBK",
			Role:     domain.RoleSystemAdmin,
			Status:   entities.UserStatusInactive,
		}
		uc := usecases.NewAuthUseCase(repo, []byte("secret"), []byte("refresh"), secLog, auditLog, nil)
		input := dto.LoginInput{
			Email:    "inactive2@test.com",
			Password: "Admin123456!",
		}
		_, _, _, err := uc.LoginWithUser(context.Background(), input)
		assert.Error(t, err)
	})
}

func TestRegister_WithLoggers(t *testing.T) {
	repo := newMockUserRepository()
	logger := logging.NewLogger("debug")
	secLog := logging.NewSecurityLogger(logger)
	auditLog := logging.NewAuditLogger(logger)
	useCase := usecases.NewAuthUseCase(repo, []byte("secret"), []byte("refresh"), secLog, auditLog, nil)

	t.Run("successful register with loggers", func(t *testing.T) {
		input := dto.RegisterInput{
			Email:    "newlogged@example.com",
			Password: "SecurePass123!",
			Role:     string(domain.RoleTeacher),
		}
		err := useCase.Register(context.Background(), input)
		assert.NoError(t, err)
	})

	t.Run("failed register with loggers", func(t *testing.T) {
		repo := newMockUserRepository()
		repo.createError = true
		uc := usecases.NewAuthUseCase(repo, []byte("secret"), []byte("refresh"), secLog, auditLog, nil)
		input := dto.RegisterInput{
			Email:    "fail@example.com",
			Password: "SecurePass123!",
			Role:     string(domain.RoleTeacher),
		}
		err := uc.Register(context.Background(), input)
		assert.Error(t, err)
	})
}

func TestRefreshToken_WithLoggers(t *testing.T) {
	repo := newMockUserRepository()
	logger := logging.NewLogger("debug")
	secLog := logging.NewSecurityLogger(logger)
	auditLog := logging.NewAuditLogger(logger)
	useCase := usecases.NewAuthUseCase(repo, []byte("secret"), []byte("refresh"), secLog, auditLog, nil)

	t.Run("successful refresh with loggers", func(t *testing.T) {
		input := dto.LoginInput{
			Email:    "test@example.com",
			Password: "Admin123456!",
		}
		_, refreshToken, err := useCase.Login(context.Background(), input)
		assert.NoError(t, err)

		newAccess, newRefresh, err := useCase.RefreshToken(context.Background(), refreshToken)
		assert.NoError(t, err)
		assert.NotEmpty(t, newAccess)
		assert.NotEmpty(t, newRefresh)
	})

	t.Run("failed refresh with loggers - invalid token", func(t *testing.T) {
		_, _, err := useCase.RefreshToken(context.Background(), "bad-token")
		assert.Error(t, err)
	})

	t.Run("failed refresh with loggers - user not found", func(t *testing.T) {
		input := dto.LoginInput{
			Email:    "test@example.com",
			Password: "Admin123456!",
		}
		_, refreshToken, err := useCase.Login(context.Background(), input)
		assert.NoError(t, err)

		repo.shouldError = true
		_, _, err = useCase.RefreshToken(context.Background(), refreshToken)
		assert.Error(t, err)
		repo.shouldError = false
	})
}

func TestRefreshToken_InactiveUser(t *testing.T) {
	repo := newMockUserRepository()
	repo.users["deactivated@example.com"] = &entities.User{
		ID:       10,
		Email:    "deactivated@example.com",
		Password: "$2a$14$dlaPdzfteiUkTlRwHMp/DuuMyviurDbsQnBwQ1MPSuUM4VnpyQJBK",
		Role:     domain.RoleSystemAdmin,
		Status:   entities.UserStatusActive,
	}
	useCase := usecases.NewAuthUseCase(repo, []byte("secret"), []byte("refresh"), nil, nil, nil)

	input := dto.LoginInput{
		Email:    "deactivated@example.com",
		Password: "Admin123456!",
	}
	_, refreshToken, err := useCase.Login(context.Background(), input)
	assert.NoError(t, err)

	// Deactivate user
	repo.users["deactivated@example.com"].Status = entities.UserStatusInactive

	_, _, err = useCase.RefreshToken(context.Background(), refreshToken)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot refresh token")
}

func TestRefreshToken_MissingUserID(t *testing.T) {
	// Create a refresh token without user_id claim
	secret := []byte("refresh-secret")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	})
	tokenString, err := token.SignedString(secret)
	assert.NoError(t, err)

	repo := newMockUserRepository()
	useCase := usecases.NewAuthUseCase(repo, []byte("access-secret"), secret, nil, nil, nil)

	_, _, err = useCase.RefreshToken(context.Background(), tokenString)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing user_id")
}

func TestRefreshToken_WrongSigningMethod(t *testing.T) {
	// Create token with none signing method
	token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
		"user_id": float64(1),
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

	repo := newMockUserRepository()
	useCase := usecases.NewAuthUseCase(repo, []byte("secret"), []byte("refresh"), nil, nil, nil)

	_, _, err := useCase.RefreshToken(context.Background(), tokenString)
	assert.Error(t, err)
}

func TestRegister_PasswordTooLong(t *testing.T) {
	repo := newMockUserRepository()
	logger := logging.NewLogger("debug")
	secLog := logging.NewSecurityLogger(logger)
	useCase := usecases.NewAuthUseCase(repo, []byte("secret"), []byte("refresh"), secLog, nil, nil)

	// bcrypt returns ErrPasswordTooLong for passwords > 72 bytes
	input := dto.RegisterInput{
		Email:    "longpass@example.com",
		Password: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", // 73 'a' chars
		Role:     "admin",
	}

	err := useCase.Register(context.Background(), input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to hash password")
}

func TestRegister_WithNotification(t *testing.T) {
	repo := newMockUserRepository()
	useCase := usecases.NewAuthUseCase(repo, []byte("secret"), []byte("refresh"), nil, nil, newTestNotifUC())

	input := dto.RegisterInput{
		Email:    "withnotif@example.com",
		Password: "SecurePass123!",
		Name:     "Notified User",
		Role:     string(domain.RoleTeacher),
	}

	err := useCase.Register(context.Background(), input)
	assert.NoError(t, err)
	// Give goroutine time to complete
	time.Sleep(50 * time.Millisecond)
}

func TestRegister_WithAllLoggers_AndNotification(t *testing.T) {
	repo := newMockUserRepository()
	logger := logging.NewLogger("debug")
	secLog := logging.NewSecurityLogger(logger)
	auditLog := logging.NewAuditLogger(logger)
	useCase := usecases.NewAuthUseCase(repo, []byte("secret"), []byte("refresh"), secLog, auditLog, newTestNotifUC())

	input := dto.RegisterInput{
		Email:    "fulltest@example.com",
		Password: "SecurePass123!",
		Name:     "Full Test",
		Role:     string(domain.RoleTeacher),
	}

	err := useCase.Register(context.Background(), input)
	assert.NoError(t, err)
	// Give goroutine time to complete
	time.Sleep(50 * time.Millisecond)
}

