// Package repositories defines repository interfaces for the users module.
package repositories

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/domain/entities"
)

// PositionRepository defines the interface for position data access.
type PositionRepository interface {
	Create(ctx context.Context, position *entities.Position) error
	GetByID(ctx context.Context, id int64) (*entities.Position, error)
	GetByCode(ctx context.Context, code string) (*entities.Position, error)
	Update(ctx context.Context, position *entities.Position) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, limit, offset int, activeOnly bool) ([]*entities.Position, error)
	Count(ctx context.Context, activeOnly bool) (int64, error)
}
