package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"
)

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
