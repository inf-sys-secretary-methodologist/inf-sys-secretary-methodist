package repositories

import (
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"
)

// UserRepository defines the interface for user persistence
type UserRepository interface {
	Save(user *entities.User) error
	GetByID(id string) (*entities.User, error)
	GetByEmail(email string) (*entities.User, error)
	Delete(id string) error
	List(limit, offset int) ([]*entities.User, error)
}
