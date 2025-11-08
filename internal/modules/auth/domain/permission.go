package domain

import "time"

// PermissionMatrix определяет матрицу разрешений для 5 основных ролей
var PermissionMatrix = map[RoleType]map[ResourceType]map[ActionType]AccessLevel{
	RoleSystemAdmin: {
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
	RoleAcademicSecretary: {
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

// RoleDefinitions определяет основные роли системы
var RoleDefinitions = map[RoleType]Role{
	RoleSystemAdmin: {
		Type:        RoleSystemAdmin,
		Name:        "Системный администратор",
		Description: "Полное управление системой",
		IsActive:    true,
	},
	RoleMethodist: {
		Type:        RoleMethodist,
		Name:        "Методист",
		Description: "Методическое обеспечение учебного процесса",
		IsActive:    true,
	},
	RoleAcademicSecretary: {
		Type:        RoleAcademicSecretary,
		Name:        "Академический секретарь",
		Description: "Административное сопровождение учебного процесса",
		IsActive:    true,
	},
	RoleTeacher: {
		Type:        RoleTeacher,
		Name:        "Преподаватель",
		Description: "Реализация образовательного процесса",
		IsActive:    true,
	},
	RoleStudent: {
		Type:        RoleStudent,
		Name:        "Студент",
		Description: "Участие в образовательном процессе",
		IsActive:    true,
	},
}

// WorkflowStep представляет шаг workflow согласования
type WorkflowStep struct {
	ID         string     `json:"id"`
	DocumentID string     `json:"document_id"`
	StepNumber int        `json:"step_number"`
	RoleType   RoleType   `json:"role_type"`
	Action     ActionType `json:"action"`
	AssignedTo *string    `json:"assigned_to,omitempty"`
	Status     string     `json:"status"`
	Comments   *string    `json:"comments,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}
