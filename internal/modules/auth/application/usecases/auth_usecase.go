// internal/modules/auth/application/usecases/auth_usecase.go
package usecases

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/repositories"
)

type AuthUseCase struct {
	userRepo      repositories.UserRepository
	jwtSecret     []byte
	refreshSecret []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

// Конструктор
func NewAuthUseCase(userRepo repositories.UserRepository, jwtSecret, refreshSecret []byte) *AuthUseCase {
	return &AuthUseCase{
		userRepo:      userRepo,
		jwtSecret:     jwtSecret,
		refreshSecret: refreshSecret,
		accessExpiry:  time.Minute * 15,
		refreshExpiry: time.Hour * 24 * 7,
	}
}

// Login - вход пользователя, выдача JWT
func (u *AuthUseCase) Login(input dto.LoginInput) (accessToken string, refreshToken string, err error) {
	user, err := u.userRepo.GetByEmail(input.Email)
	if err != nil {
		return "", "", errors.New("invalid email or password")
	}

	// проверка пароля
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return "", "", errors.New("invalid email or password")
	}

	// создаём токены
	accessToken, refreshToken, err = u.generateTokens(user)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// Вспомогательная функция генерации токенов
func (u *AuthUseCase) generateTokens(user *entities.User) (string, string, error) {
	// Access Token
	atClaims := jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(u.accessExpiry).Unix(),
	}
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	accessToken, err := at.SignedString(u.jwtSecret)
	if err != nil {
		return "", "", err
	}

	// Refresh Token
	rtClaims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(u.refreshExpiry).Unix(),
	}
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	refreshToken, err := rt.SignedString(u.refreshSecret)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
