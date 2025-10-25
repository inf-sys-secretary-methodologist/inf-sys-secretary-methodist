package repositories

import "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"

type UserRepository interface {
	Create(user *entities.User) error
	Save(user *entities.User) error // <-- обязательно
	Delete(userID string) error
	GetByEmail(email string) (*entities.User, error)
	GetByID(userID string) (*entities.User, error)
	List(page int, limit int) ([]*entities.User, error)
}
