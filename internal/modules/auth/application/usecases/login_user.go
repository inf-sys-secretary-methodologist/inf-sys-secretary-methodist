package usecases

import (
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/services"
)

// LoginUserCommand represents login command
type LoginUserCommand struct {
	Email    string
	Password string
}

// LoginUserResult represents login result
type LoginUserResult struct {
	Token string
	User  *entities.User
}

// LoginUserUseCase handles user login
type LoginUserUseCase struct {
	userRepo    repositories.UserRepository
	authService services.AuthService
}

// NewLoginUserUseCase creates a new login use case
func NewLoginUserUseCase(
	userRepo repositories.UserRepository,
	authService services.AuthService,
) *LoginUserUseCase {
	return &LoginUserUseCase{
		userRepo:    userRepo,
		authService: authService,
	}
}

// Execute executes the login use case
func (uc *LoginUserUseCase) Execute(cmd LoginUserCommand) (*LoginUserResult, error) {
	// Validate credentials
	user, err := uc.authService.ValidateCredentials(cmd.Email, cmd.Password)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials: %w", err)
	}

	// Check if user is active
	if !user.IsActive() {
		return nil, fmt.Errorf("user account is not active")
	}

	// Generate token
	token, err := uc.authService.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &LoginUserResult{
		Token: token,
		User:  user,
	}, nil
}
