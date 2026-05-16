// Package usecases contains business logic and the repository ports
// (interfaces) consumed by the auth module use cases. Repository ports
// live in the consumer package per Dependency Inversion Principle
// (CLAUDE.md DDD-гейт: "Repository interfaces — в пакете-потребителе
// (usecase/), НЕ в domain/").
package usecases

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
	// GetByEmailForAuth retrieves user by email for authentication purposes
	// Always fetches from database (bypasses cache) to ensure password field is populated
	GetByEmailForAuth(ctx context.Context, email string) (*entities.User, error)
	// GetByIDForAuth retrieves user by ID bypassing cache so the MFA secret
	// (excluded from cache via json:"-") is available for TOTP verification.
	GetByIDForAuth(ctx context.Context, id int64) (*entities.User, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, limit, offset int) ([]*entities.User, error)
}
