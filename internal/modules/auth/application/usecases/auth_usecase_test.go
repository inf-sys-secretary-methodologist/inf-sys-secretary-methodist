package usecases_test

import (
	"context"
	"testing"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"
)

// Mock repository for testing
type mockUserRepository struct{}

func (m *mockUserRepository) Create(ctx context.Context, user *entities.User) error {
	user.ID = 1
	return nil
}

func (m *mockUserRepository) Save(ctx context.Context, user *entities.User) error {
	return nil
}

func (m *mockUserRepository) GetByID(ctx context.Context, id int64) (*entities.User, error) {
	return &entities.User{
		ID:       1,
		Email:    "test@example.com",
		Password: "$2a$14$test",
		Role:     entities.RoleAdmin,
		Status:   entities.UserStatusActive,
	}, nil
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	// Password is: Admin123456!
	return &entities.User{
		ID:       1,
		Email:    email,
		Password: "$2a$14$ZKHqFX3vJT8kZY7ZJy.zfOEzBxD8YmBqGqN1xPJvJ1Y1xYJPqJ5qW",
		Role:     entities.RoleAdmin,
		Status:   entities.UserStatusActive,
	}, nil
}

func (m *mockUserRepository) Delete(ctx context.Context, id int64) error {
	return nil
}

func (m *mockUserRepository) List(ctx context.Context, limit, offset int) ([]*entities.User, error) {
	return []*entities.User{}, nil
}

func TestRegister(t *testing.T) {
	repo := &mockUserRepository{}
	useCase := usecases.NewAuthUseCase(repo, []byte("secret"), []byte("refresh"))

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
	repo := &mockUserRepository{}
	useCase := usecases.NewAuthUseCase(repo, []byte("secret"), []byte("refresh"))

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
	repo := &mockUserRepository{}
	useCase := usecases.NewAuthUseCase(repo, []byte("secret"), []byte("refresh"))

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
