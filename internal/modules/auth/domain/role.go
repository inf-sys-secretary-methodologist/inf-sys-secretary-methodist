package domain

import "time"

// RoleType представляет тип роли
type RoleType string

const (
	// RoleSystemAdmin represents the system administrator role with full access.
	RoleSystemAdmin RoleType = "system_admin"
	// RoleMethodist represents the methodist role.
	RoleMethodist RoleType = "methodist"
	// RoleAcademicSecretary represents the academic secretary role.
	RoleAcademicSecretary RoleType = "academic_secretary"
	// RoleTeacher represents the teacher role.
	RoleTeacher RoleType = "teacher"
	// RoleStudent represents the student role.
	RoleStudent RoleType = "student"
)

// ResourceType представляет тип ресурса
type ResourceType string

const (
	// ResourceUsers represents the users resource type.
	ResourceUsers ResourceType = "users"
	// ResourceCurriculum represents the curriculum resource type.
	ResourceCurriculum ResourceType = "curriculum"
	// ResourceSchedule represents the schedule resource type.
	ResourceSchedule ResourceType = "schedule"
	// ResourceAssignments represents the assignments resource type.
	ResourceAssignments ResourceType = "assignments"
	// ResourceReports represents the reports resource type.
	ResourceReports ResourceType = "reports"
)

// ActionType представляет тип действия
type ActionType string

const (
	// ActionCreate represents the create action.
	ActionCreate ActionType = "create"
	// ActionRead represents the read action.
	ActionRead ActionType = "read"
	// ActionUpdate represents the update action.
	ActionUpdate ActionType = "update"
	// ActionDelete represents the delete action.
	ActionDelete ActionType = "delete"
	// ActionDeactivate represents the deactivate action.
	ActionDeactivate ActionType = "deactivate"
	// ActionApprove represents the approve action.
	ActionApprove ActionType = "approve"
	// ActionExecute represents the execute action.
	ActionExecute ActionType = "execute"
	// ActionExport represents the export action.
	ActionExport ActionType = "export"
)

// AccessLevel представляет уровень доступа
type AccessLevel int

const (
	// AccessDenied indicates no access to the resource.
	AccessDenied AccessLevel = 0
	// AccessLimited indicates limited access to the resource.
	AccessLimited AccessLevel = 1
	// AccessOwn indicates access only to own resources.
	AccessOwn AccessLevel = 2
	// AccessFull indicates full access to the resource.
	AccessFull AccessLevel = 3
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
