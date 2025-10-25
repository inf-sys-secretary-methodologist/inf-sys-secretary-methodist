package services

import (
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"

	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AuthService defines authentication domain service
type AuthService interface {
	ValidateCredentials(email, password string) (*entities.User, error)
	GenerateToken(user *entities.User) (string, error)
	ValidateToken(token string) (*entities.User, error)
}

type JWTService struct {
	secretKey       string
	refreshSecret   string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

func NewJWTService(secret, refreshSecret string, accessTTL, refreshTTL time.Duration) *JWTService {
	return &JWTService{
		secretKey:       secret,
		refreshSecret:   refreshSecret,
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
	}
}

func (s *JWTService) GenerateTokens(userID int64, role string) (*TokenPair, error) {
	accessClaims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(s.accessTokenTTL).Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessStr, err := accessToken.SignedString([]byte(s.secretKey))
	if err != nil {
		return nil, err
	}

	refreshClaims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(s.refreshTokenTTL).Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshStr, err := refreshToken.SignedString([]byte(s.refreshSecret))
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessStr,
		RefreshToken: refreshStr,
	}, nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
