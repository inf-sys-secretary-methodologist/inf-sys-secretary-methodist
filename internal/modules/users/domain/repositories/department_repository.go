// Package repositories defines repository interfaces for the users module.
package repositories

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/domain/entities"
)

// DepartmentRepository defines the interface for department data access.
type DepartmentRepository interface {
	Create(ctx context.Context, department *entities.Department) error
	GetByID(ctx context.Context, id int64) (*entities.Department, error)
	GetByCode(ctx context.Context, code string) (*entities.Department, error)
	Update(ctx context.Context, department *entities.Department) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, limit, offset int, activeOnly bool) ([]*entities.Department, error)
	Count(ctx context.Context, activeOnly bool) (int64, error)
	GetChildren(ctx context.Context, parentID int64) ([]*entities.Department, error)
}
