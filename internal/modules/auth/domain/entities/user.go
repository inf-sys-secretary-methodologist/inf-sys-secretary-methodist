package entities

import (
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/domain/common"
)

// User represents a user entity in the auth domain
type User struct {
	common.Entity
	common.AggregateRoot

	Email        string
	PasswordHash string
	Name         string
	Role         UserRole
	Status       UserStatus
}

// UserRole represents user role
type UserRole string

const (
	RoleAdmin     UserRole = "admin"
	RoleSecretary UserRole = "secretary"
	RoleMethodist UserRole = "methodist"
	RoleTeacher   UserRole = "teacher"
	RoleStudent   UserRole = "student"
)

// UserStatus represents user account status
type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusBlocked  UserStatus = "blocked"
)

// NewUser creates a new user
func NewUser(id, email, passwordHash, name string, role UserRole) *User {
	now := time.Now()
	return &User{
		Entity: common.Entity{
			ID:        id,
			CreatedAt: now,
			UpdatedAt: now,
		},
		Email:        email,
		PasswordHash: passwordHash,
		Name:         name,
		Role:         role,
		Status:       UserStatusActive,
	}
}

// Activate activates the user account
func (u *User) Activate() {
	u.Status = UserStatusActive
	u.Entity.Touch()
}

// Deactivate deactivates the user account
func (u *User) Deactivate() {
	u.Status = UserStatusInactive
	u.Entity.Touch()
}

// Block blocks the user account
func (u *User) Block() {
	u.Status = UserStatusBlocked
	u.Entity.Touch()
}

// IsActive checks if user is active
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}
