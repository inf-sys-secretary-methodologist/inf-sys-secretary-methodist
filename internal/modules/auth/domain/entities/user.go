package entities

import (
	"fmt"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
)

// User represents a user entity in the auth domain
// Aligned with migrations/001_create_users_table.up.sql
type User struct {
	ID        int64           `db:"id" json:"id"`
	Email     string          `db:"email" json:"email"`
	Password  string          `db:"password" json:"-"` // hashed password, не отдаём в JSON
	Name      string          `db:"name" json:"name"`
	Role      domain.RoleType `db:"role" json:"role"`
	Status    UserStatus      `db:"status" json:"status"`
	CreatedAt time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt time.Time       `db:"updated_at" json:"updated_at"`
}

// UserStatus represents user account status
type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusBlocked  UserStatus = "blocked"
)

// NewUser creates a new user
func NewUser(email, passwordHash, name string, role domain.RoleType) *User {
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

// UserContext представляет контекст пользователя для проверки прав
type UserContext struct {
	UserID    int64           `json:"user_id"`
	Role      domain.RoleType `json:"role"`
	FacultyID *string         `json:"faculty_id,omitempty"`
	GroupID   *string         `json:"group_id,omitempty"`
}

// HasPermission проверяет наличие разрешения у пользователя
func (uc *UserContext) HasPermission(resource domain.ResourceType, action domain.ActionType) bool {
	rolePermissions, exists := domain.PermissionMatrix[uc.Role]
	if !exists {
		return false
	}

	actionPermissions, exists := rolePermissions[resource]
	if !exists {
		return false
	}

	accessLevel, exists := actionPermissions[action]
	if !exists {
		return false
	}

	return accessLevel > domain.AccessDenied
}

// GetAccessLevel возвращает уровень доступа для ресурса и действия
func (uc *UserContext) GetAccessLevel(resource domain.ResourceType, action domain.ActionType) domain.AccessLevel {
	rolePermissions, exists := domain.PermissionMatrix[uc.Role]
	if !exists {
		return domain.AccessDenied
	}

	actionPermissions, exists := rolePermissions[resource]
	if !exists {
		return domain.AccessDenied
	}

	accessLevel, exists := actionPermissions[action]
	if !exists {
		return domain.AccessDenied
	}

	return accessLevel
}

// ToUserContext преобразует User в UserContext
func (u *User) ToUserContext() *UserContext {
	return &UserContext{
		UserID: u.ID,
		Role:   u.Role,
	}
}
