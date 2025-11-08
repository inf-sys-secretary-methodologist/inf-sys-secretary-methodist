package domain

import "time"

// RoleType представляет тип роли
type RoleType string

const (
	RoleSystemAdmin       RoleType = "system_admin"
	RoleMethodist         RoleType = "methodist"
	RoleAcademicSecretary RoleType = "academic_secretary"
	RoleTeacher           RoleType = "teacher"
	RoleStudent           RoleType = "student"
)

// ResourceType представляет тип ресурса
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
	AccessLimited AccessLevel = 1
	AccessOwn     AccessLevel = 2
	AccessFull    AccessLevel = 3
)

// Permission представляет разрешение
type Permission struct {
	ID          string       `json:"id"`
	Resource    ResourceType `json:"resource"`
	Action      ActionType   `json:"action"`
	AccessLevel AccessLevel  `json:"access_level"`
	Description string       `json:"description"`
}

// Role представляет роль пользователя
type Role struct {
	ID          string       `json:"id"`
	Type        RoleType     `json:"type"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Permissions []Permission `json:"permissions"`
	IsActive    bool         `json:"is_active"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// UserRole представляет роль пользователя с контекстом
type UserRole struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	RoleID    string     `json:"role_id"`
	Scope     *Scope     `json:"scope,omitempty"`
	IsActive  bool       `json:"is_active"`
	StartsAt  *time.Time `json:"starts_at,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// Scope представляет область видимости роли
type Scope struct {
	FacultyID    *string `json:"faculty_id,omitempty"`
	DepartmentID *string `json:"department_id,omitempty"`
	GroupID      *string `json:"group_id,omitempty"`
}
