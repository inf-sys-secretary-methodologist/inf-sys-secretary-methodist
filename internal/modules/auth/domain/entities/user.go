package entities

import (
	"fmt"
	"time"
)

// User represents a user entity in the auth domain
type User struct {
	ID        int64      `db:"id"`
	Email     string     `db:"email"`
	Password  string     `db:"password"` // hashed password
	Name      string     `db:"name"`
	Role      UserRole   `db:"role"`
	Status    UserStatus `db:"status"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
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
func NewUser(email, passwordHash, name string, role UserRole) *User {
	now := time.Now()
	return &User{
		Email:     email,
		Password:  passwordHash,
		Name:      name,
		Role:      role,
		Status:    UserStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Activate activates the user account
func (u *User) Activate() {
	u.Status = UserStatusActive
	u.UpdatedAt = time.Now()
}

// Deactivate deactivates the user account
func (u *User) Deactivate() {
	u.Status = UserStatusInactive
	u.UpdatedAt = time.Now()
}

// Block blocks the user account
func (u *User) Block() {
	u.Status = UserStatusBlocked
	u.UpdatedAt = time.Now()
}

// CanLogin checks if user can login
func (u *User) CanLogin() error {
	if !u.IsActive() {
		return ErrAccountNotActive
	}
	if u.Status == UserStatusBlocked {
		return ErrAccountBlocked
	}
	return nil
}

var (
	ErrAccountNotActive = fmt.Errorf("account is not active")
	ErrAccountBlocked   = fmt.Errorf("account is blocked")
)

// IsActive checks if user is active
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}
