// Package repositories defines repository interfaces for the users module.
package repositories

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/domain/entities"
)

// UserFilter contains filter options for listing users.
type UserFilter struct {
	DepartmentID *int64
	PositionID   *int64
	Role         string
	Status       string
	Search       string // search by name or email
}

// UserProfileRepository defines the interface for extended user profile data access.
type UserProfileRepository interface {
	// GetProfileByID retrieves user profile with organizational info
	GetProfileByID(ctx context.Context, userID int64) (*entities.UserWithOrg, error)

	// UpdateProfile updates user's organizational information (department, position, phone, etc.)
	UpdateProfile(ctx context.Context, userID int64, departmentID, positionID *int64, phone, avatar, bio string) error

	// ListUsersWithOrg retrieves paginated list of users with their organizational info
	ListUsersWithOrg(ctx context.Context, filter *UserFilter, limit, offset int) ([]*entities.UserWithOrg, error)

	// CountUsers returns total count of users matching the filter
	CountUsers(ctx context.Context, filter *UserFilter) (int64, error)

	// GetUsersByDepartment retrieves all users in a specific department
	GetUsersByDepartment(ctx context.Context, departmentID int64) ([]*entities.UserWithOrg, error)

	// GetUsersByPosition retrieves all users with a specific position
	GetUsersByPosition(ctx context.Context, positionID int64) ([]*entities.UserWithOrg, error)

	// BulkUpdateDepartment moves multiple users to a new department
	BulkUpdateDepartment(ctx context.Context, userIDs []int64, departmentID *int64) error

	// BulkUpdatePosition assigns multiple users to a new position
	BulkUpdatePosition(ctx context.Context, userIDs []int64, positionID *int64) error
}
