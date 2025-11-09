package entities

import (
	"fmt"
	"time"
)

// User represents a user entity in the auth domain
// Aligned with migrations/001_create_users_table.up.sql
type User struct {
	ID        int64      `db:"id" json:"id"`
	Email     string     `db:"email" json:"email"`
	Password  string     `db:"password" json:"-"` // hashed password, не отдаём в JSON
	Name      string     `db:"name" json:"name"`
	Role      UserRole   `db:"role" json:"role"`
	Status    UserStatus `db:"status" json:"status"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
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

// ===== НОВАЯ ФУНКЦИОНАЛЬНОСТЬ ДЛЯ СИСТЕМЫ РАЗРЕШЕНИЙ =====

// ResourceType представляет тип ресурса в системе
type ResourceType string

const (
	ResourceUsers       ResourceType = "users"
	ResourceCurriculum  ResourceType = "curriculum"
	ResourceSchedule    ResourceType = "schedule"
	ResourceAssignments ResourceType = "assignments"
	ResourceReports     ResourceType = "reports"
)

// ActionType представляет тип действия
type ActionType string

const (
	ActionCreate     ActionType = "create"
	ActionRead       ActionType = "read"
	ActionUpdate     ActionType = "update"
	ActionDelete     ActionType = "delete"
	ActionDeactivate ActionType = "deactivate"
	ActionApprove    ActionType = "approve"
	ActionExecute    ActionType = "execute"
	ActionExport     ActionType = "export"
)

// AccessLevel представляет уровень доступа
type AccessLevel int

const (
	AccessDenied  AccessLevel = 0
	AccessOwn     AccessLevel = 1
	AccessLimited AccessLevel = 2
	AccessFull    AccessLevel = 3
)

// PermissionMatrix определяет матрицу разрешений для ролей
var PermissionMatrix = map[UserRole]map[ResourceType]map[ActionType]AccessLevel{
	RoleAdmin: {
		ResourceUsers: {
			ActionCreate:     AccessFull,
			ActionRead:       AccessFull,
			ActionUpdate:     AccessFull,
			ActionDeactivate: AccessFull,
		},
		ResourceCurriculum: {
			ActionCreate:  AccessFull,
			ActionRead:    AccessFull,
			ActionUpdate:  AccessFull,
			ActionApprove: AccessFull,
		},
		ResourceSchedule: {
			ActionCreate: AccessFull,
			ActionRead:   AccessFull,
			ActionUpdate: AccessFull,
		},
		ResourceAssignments: {
			ActionCreate: AccessFull,
			ActionRead:   AccessFull,
			ActionUpdate: AccessFull,
		},
		ResourceReports: {
			ActionCreate: AccessFull,
			ActionRead:   AccessFull,
			ActionExport: AccessFull,
		},
	},
	RoleMethodist: {
		ResourceUsers: {
			ActionCreate:     AccessDenied,
			ActionRead:       AccessLimited,
			ActionUpdate:     AccessDenied,
			ActionDeactivate: AccessDenied,
		},
		ResourceCurriculum: {
			ActionCreate:  AccessFull,
			ActionRead:    AccessFull,
			ActionUpdate:  AccessFull,
			ActionApprove: AccessDenied,
		},
		ResourceSchedule: {
			ActionCreate: AccessDenied,
			ActionRead:   AccessFull,
			ActionUpdate: AccessLimited,
		},
		ResourceAssignments: {
			ActionCreate: AccessFull,
			ActionRead:   AccessFull,
			ActionUpdate: AccessLimited,
		},
		ResourceReports: {
			ActionCreate: AccessFull,
			ActionRead:   AccessFull,
			ActionExport: AccessFull,
		},
	},
	RoleSecretary: {
		ResourceUsers: {
			ActionCreate:     AccessDenied,
			ActionRead:       AccessLimited,
			ActionUpdate:     AccessDenied,
			ActionDeactivate: AccessDenied,
		},
		ResourceCurriculum: {
			ActionCreate:  AccessDenied,
			ActionRead:    AccessFull,
			ActionUpdate:  AccessDenied,
			ActionApprove: AccessDenied,
		},
		ResourceSchedule: {
			ActionCreate: AccessFull,
			ActionRead:   AccessFull,
			ActionUpdate: AccessFull,
		},
		ResourceAssignments: {
			ActionCreate: AccessDenied,
			ActionRead:   AccessFull,
			ActionUpdate: AccessDenied,
		},
		ResourceReports: {
			ActionCreate: AccessFull,
			ActionRead:   AccessFull,
			ActionExport: AccessFull,
		},
	},
	RoleTeacher: {
		ResourceUsers: {
			ActionCreate:     AccessDenied,
			ActionRead:       AccessLimited,
			ActionUpdate:     AccessDenied,
			ActionDeactivate: AccessDenied,
		},
		ResourceCurriculum: {
			ActionCreate:  AccessDenied,
			ActionRead:    AccessFull,
			ActionUpdate:  AccessLimited,
			ActionApprove: AccessDenied,
		},
		ResourceSchedule: {
			ActionCreate: AccessDenied,
			ActionRead:   AccessFull,
			ActionUpdate: AccessDenied,
		},
		ResourceAssignments: {
			ActionCreate: AccessFull,
			ActionRead:   AccessFull,
			ActionUpdate: AccessOwn,
		},
		ResourceReports: {
			ActionCreate: AccessFull,
			ActionRead:   AccessLimited,
			ActionExport: AccessLimited,
		},
	},
	RoleStudent: {
		ResourceUsers: {
			ActionCreate:     AccessDenied,
			ActionRead:       AccessDenied,
			ActionUpdate:     AccessOwn,
			ActionDeactivate: AccessDenied,
		},
		ResourceCurriculum: {
			ActionCreate:  AccessDenied,
			ActionRead:    AccessLimited,
			ActionUpdate:  AccessDenied,
			ActionApprove: AccessDenied,
		},
		ResourceSchedule: {
			ActionCreate: AccessDenied,
			ActionRead:   AccessFull,
			ActionUpdate: AccessDenied,
		},
		ResourceAssignments: {
			ActionCreate:  AccessDenied,
			ActionRead:    AccessOwn,
			ActionUpdate:  AccessDenied,
			ActionExecute: AccessFull,
		},
		ResourceReports: {
			ActionCreate: AccessDenied,
			ActionRead:   AccessDenied,
			ActionExport: AccessDenied,
		},
	},
}

// UserContext представляет контекст пользователя для проверки прав
type UserContext struct {
	UserID    string   `json:"user_id"`
	Role      UserRole `json:"role"`
	FacultyID *string  `json:"faculty_id,omitempty"`
	GroupID   *string  `json:"group_id,omitempty"`
}

// HasPermission проверяет наличие разрешения у пользователя
func (uc *UserContext) HasPermission(resource ResourceType, action ActionType) bool {
	rolePermissions, exists := PermissionMatrix[uc.Role]
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

	return accessLevel > AccessDenied
}

// GetAccessLevel возвращает уровень доступа для ресурса и действия
func (uc *UserContext) GetAccessLevel(resource ResourceType, action ActionType) AccessLevel {
	rolePermissions, exists := PermissionMatrix[uc.Role]
	if !exists {
		return AccessDenied
	}

	actionPermissions, exists := rolePermissions[resource]
	if !exists {
		return AccessDenied
	}

	accessLevel, exists := actionPermissions[action]
	if !exists {
		return AccessDenied
	}

	return accessLevel
}

// ToUserContext преобразует User в UserContext
func (u *User) ToUserContext() *UserContext {
	return &UserContext{
		UserID: u.Entity.ID,
		Role:   u.Role,
	}
}
