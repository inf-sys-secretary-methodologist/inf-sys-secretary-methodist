package repositories

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"
)

// UserRepository defines the interface for user persistence
type UserRepository interface {
	Create(ctx context.Context, user *entities.User) error
	Save(ctx context.Context, user *entities.User) error
	GetByID(ctx context.Context, id int64) (*entities.User, error)
	GetByEmail(ctx context.Context, email string) (*entities.User, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, limit, offset int) ([]*entities.User, error)
}
