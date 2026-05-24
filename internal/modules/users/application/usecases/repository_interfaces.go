// Package usecases owns the users repository ports per Clean
// Architecture DIP (gate from CLAUDE.md: "Repository interfaces — в
// пакете-потребителе (`usecase/`), НЕ в `domain/`"). Mirror
// к v0.157.1 curriculum + v0.162.1 messaging + v0.163.1 announcements
// precedent.
//
// Sentinels (none — users has none today) and query DTOs (UserFilter)
// remain in domain/repositories/ как value types whose ownership is
// the producer module. UserAccountRepository уже жил в this package
// since v0.139.0 as the narrow port over the auth user store.
package usecases

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/domain/repositories"
)

// UserProfileRepository defines the interface for extended user
// profile data access. UserFilter query DTO остаётся в
// domain/repositories — оно бывшее значениеv-type filter, не
// implementation-detail of the repository contract.
type UserProfileRepository interface {
	// GetProfileByID retrieves user profile with organizational info
	GetProfileByID(ctx context.Context, userID int64) (*entities.UserWithOrg, error)

	// UpdateProfile updates user's organizational information (department, position, phone, etc.)
	UpdateProfile(ctx context.Context, userID int64, departmentID, positionID *int64, phone, avatar, bio string) error

	// ListUsersWithOrg retrieves paginated list of users with their organizational info
	ListUsersWithOrg(ctx context.Context, filter *repositories.UserFilter, limit, offset int) ([]*entities.UserWithOrg, error)

	// CountUsers returns total count of users matching the filter
	CountUsers(ctx context.Context, filter *repositories.UserFilter) (int64, error)

	// GetUsersByDepartment retrieves all users in a specific department
	GetUsersByDepartment(ctx context.Context, departmentID int64) ([]*entities.UserWithOrg, error)

	// GetUsersByPosition retrieves all users with a specific position
	GetUsersByPosition(ctx context.Context, positionID int64) ([]*entities.UserWithOrg, error)

	// BulkUpdateDepartment moves multiple users to a new department
	BulkUpdateDepartment(ctx context.Context, userIDs []int64, departmentID *int64) error

	// BulkUpdatePosition assigns multiple users to a new position
	BulkUpdatePosition(ctx context.Context, userIDs []int64, positionID *int64) error
}

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
