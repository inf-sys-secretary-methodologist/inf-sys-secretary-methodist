package services

import (
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"
)

// AuthService defines authentication domain service
type AuthService interface {
	ValidateCredentials(email, password string) (*entities.User, error)
	GenerateToken(user *entities.User) (string, error)
	ValidateToken(token string) (*entities.User, error)
}
