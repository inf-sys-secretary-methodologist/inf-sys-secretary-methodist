// Package entities contains domain entities for the users module.
package entities

import (
	"time"

	authEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"
)

// UserProfile extends the auth User with organizational information.
type UserProfile struct {
	*authEntities.User
	DepartmentID *int64  `db:"department_id" json:"department_id,omitempty"`
	PositionID   *int64  `db:"position_id" json:"position_id,omitempty"`
	Phone        string  `db:"phone" json:"phone,omitempty"`
	Avatar       string  `db:"avatar" json:"avatar,omitempty"`
	Bio          string  `db:"bio" json:"bio,omitempty"`

	// Denormalized fields for convenience (populated from joins)
	DepartmentName string `db:"-" json:"department_name,omitempty"`
	PositionName   string `db:"-" json:"position_name,omitempty"`
}

// UserWithOrg represents a user with organizational details for list/detail views.
type UserWithOrg struct {
	ID             int64     `json:"id"`
	Email          string    `json:"email"`
	Name           string    `json:"name"`
	Role           string    `json:"role"`
	Status         string    `json:"status"`
	Phone          string    `json:"phone,omitempty"`
	Avatar         string    `json:"avatar,omitempty"`
	DepartmentID   *int64    `json:"department_id,omitempty"`
	DepartmentName string    `json:"department_name,omitempty"`
	PositionID     *int64    `json:"position_id,omitempty"`
	PositionName   string    `json:"position_name,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
