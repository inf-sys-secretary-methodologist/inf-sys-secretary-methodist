// Package domain contains core domain logic for authentication and authorization.
package domain

import "time"

// PermissionMatrix определяет матрицу разрешений для 5 основных ролей
//
// БИЗНЕС-ЛОГИКА РОЛЕЙ:
//
// RoleSystemAdmin - Системный администратор
//   Полное управление системой, пользователями, всеми ресурсами
//
// RoleMethodist - Методист
//   Методическое обеспечение: документы, учебные планы, мероприятия
//   Может создавать объявления, работать с задачами
//
// RoleAcademicSecretary - Академический секретарь
//   Административное сопровождение: студенты, расписание, отчёты
//   Полное управление студентами и расписанием
//
// RoleTeacher - Преподаватель
//   Реализация учебного процесса: просмотр студентов, задания
//   Загрузка своих документов, работа со своими задачами
//
// RoleStudent - Студент
//   Участие в учебном процессе: просмотр расписания, задач, событий
//   Выполнение заданий, просмотр объявлений
var PermissionMatrix = map[RoleType]map[ResourceType]map[ActionType]AccessLevel{
	// =========================================================================
	// RoleSystemAdmin - Полный доступ ко всему
	// =========================================================================
	RoleSystemAdmin: {
		ResourceUsers: {
			ActionCreate:     AccessFull,
			ActionRead:       AccessFull,
			ActionUpdate:     AccessFull,
			ActionDelete:     AccessFull,
			ActionDeactivate: AccessFull,
		},
		ResourceCurriculum: {
			ActionCreate:  AccessFull,
			ActionRead:    AccessFull,
			ActionUpdate:  AccessFull,
			ActionApprove: AccessFull,
		},
		ResourceDocuments: {
			ActionCreate: AccessFull,
			ActionRead:   AccessFull,
			ActionUpdate: AccessFull,
			ActionDelete: AccessFull,
			ActionUpload: AccessFull,
			ActionExport: AccessFull,
		},
		ResourceStudents: {
			ActionCreate: AccessFull,
			ActionRead:   AccessFull,
			ActionUpdate: AccessFull,
			ActionDelete: AccessFull,
		},
		ResourceEvents: {
			ActionCreate: AccessFull,
			ActionRead:   AccessFull,
			ActionUpdate: AccessFull,
			ActionDelete: AccessFull,
		},
		ResourceTasks: {
			ActionCreate: AccessFull,
			ActionRead:   AccessFull,
			ActionUpdate: AccessFull,
			ActionDelete: AccessFull,
		},
		ResourceAssignments: {
			ActionCreate: AccessFull,
			ActionRead:   AccessFull,
			ActionUpdate: AccessFull,
			ActionDelete: AccessFull,
		},
		ResourceSchedule: {
			ActionCreate: AccessFull,
			ActionRead:   AccessFull,
			ActionUpdate: AccessFull,
		},
		ResourceReports: {
			ActionCreate:  AccessFull,
			ActionRead:    AccessFull,
			ActionExport:  AccessFull,
			ActionApprove: AccessFull,
		},
		ResourceAnnouncements: {
			ActionCreate: AccessFull,
			ActionRead:   AccessFull,
			ActionUpdate: AccessFull,
			ActionDelete: AccessFull,
		},
	},

	// =========================================================================
	// RoleMethodist - Методическое обеспечение учебного процесса
	// =========================================================================
	RoleMethodist: {
		ResourceUsers: {
			ActionRead: AccessLimited, // Только базовая информация
		},
		ResourceCurriculum: {
			ActionCreate: AccessFull,
			ActionRead:   AccessFull,
			ActionUpdate: AccessFull,
			// Нет ActionApprove - утверждает только админ
		},
		ResourceDocuments: {
			ActionCreate: AccessFull,
			ActionRead:   AccessFull,
			ActionUpdate: AccessFull,
			ActionUpload: AccessFull,
			ActionExport: AccessFull,
		},
		ResourceStudents: {
			ActionRead:   AccessFull,    // Просмотр всех студентов
			ActionUpdate: AccessLimited, // Ограниченное редактирование
		},
		ResourceEvents: {
			ActionCreate: AccessFull,
			ActionRead:   AccessFull,
			ActionUpdate: AccessFull,
		},
		ResourceTasks: {
			ActionCreate: AccessFull,
			ActionRead:   AccessFull,
			ActionUpdate: AccessFull,
		},
		ResourceAssignments: {
			ActionCreate: AccessFull,
			ActionRead:   AccessFull,
			ActionUpdate: AccessLimited,
		},
		ResourceSchedule: {
			ActionRead:   AccessFull,
			ActionUpdate: AccessLimited, // Может предлагать изменения
		},
		ResourceReports: {
			ActionCreate: AccessFull,
			ActionRead:   AccessFull,
			ActionExport: AccessFull,
		},
		ResourceAnnouncements: {
			ActionCreate: AccessFull,
			ActionRead:   AccessFull,
			ActionUpdate: AccessOwn, // Только свои объявления
			ActionDelete: AccessOwn,
		},
	},

	// =========================================================================
	// RoleAcademicSecretary - Административное сопровождение
	// =========================================================================
	RoleAcademicSecretary: {
		ResourceUsers: {
			ActionRead: AccessLimited,
		},
		ResourceCurriculum: {
			ActionRead: AccessFull,
			// Не создаёт и не редактирует учебные планы
		},
		ResourceDocuments: {
			ActionCreate: AccessFull,
			ActionRead:   AccessFull,
			ActionUpdate: AccessFull,
			ActionUpload: AccessFull,
			ActionExport: AccessFull,
		},
		ResourceStudents: {
			ActionCreate: AccessFull, // Полное управление студентами
			ActionRead:   AccessFull,
			ActionUpdate: AccessFull,
			ActionDelete: AccessFull,
		},
		ResourceEvents: {
			ActionCreate: AccessFull,
			ActionRead:   AccessFull,
			ActionUpdate: AccessFull,
			ActionDelete: AccessFull,
		},
		ResourceTasks: {
			ActionCreate: AccessFull,
			ActionRead:   AccessFull,
			ActionUpdate: AccessFull,
		},
		ResourceAssignments: {
			ActionRead: AccessFull,
			// Не создаёт задания
		},
		ResourceSchedule: {
			ActionCreate: AccessFull, // Полное управление расписанием
			ActionRead:   AccessFull,
			ActionUpdate: AccessFull,
		},
		ResourceReports: {
			ActionCreate: AccessFull,
			ActionRead:   AccessFull,
			ActionExport: AccessFull,
		},
		ResourceAnnouncements: {
			ActionCreate: AccessFull,
			ActionRead:   AccessFull,
			ActionUpdate: AccessOwn,
		},
	},

	// =========================================================================
	// RoleTeacher - Преподаватель
	// =========================================================================
	RoleTeacher: {
		ResourceUsers: {
			ActionRead: AccessLimited, // Базовая информация коллег
		},
		ResourceCurriculum: {
			ActionRead:   AccessFull,
			ActionUpdate: AccessLimited, // Может предлагать изменения
		},
		ResourceDocuments: {
			ActionRead:   AccessFull,
			ActionUpload: AccessOwn, // Только свои документы
			ActionExport: AccessLimited,
		},
		ResourceStudents: {
			ActionRead: AccessFull, // Просмотр студентов для выставления оценок
		},
		ResourceEvents: {
			ActionRead:   AccessFull,
			ActionCreate: AccessOwn, // Может создавать свои мероприятия
		},
		ResourceTasks: {
			ActionRead:   AccessFull, // Просмотр всех задач
			ActionUpdate: AccessOwn,  // Редактирование своих
			ActionCreate: AccessOwn,  // Создание задач для студентов
		},
		ResourceAssignments: {
			ActionCreate: AccessFull, // Создание заданий
			ActionRead:   AccessFull,
			ActionUpdate: AccessOwn, // Редактирование своих
		},
		ResourceSchedule: {
			ActionRead: AccessFull,
		},
		ResourceReports: {
			ActionCreate: AccessFull,    // Отчёты по своим предметам
			ActionRead:   AccessLimited, // Только свои отчёты
			ActionExport: AccessLimited,
		},
		ResourceAnnouncements: {
			ActionRead:   AccessFull,
			ActionCreate: AccessOwn, // Может создавать для своих групп
		},
	},

	// =========================================================================
	// RoleStudent - Студент (ВАЖНО: видит задачи в режиме просмотра!)
	// =========================================================================
	RoleStudent: {
		ResourceUsers: {
			ActionRead:   AccessOwn, // Только свой профиль
			ActionUpdate: AccessOwn, // Редактирование профиля
		},
		ResourceCurriculum: {
			ActionRead: AccessLimited, // Свой учебный план
		},
		ResourceDocuments: {
			ActionRead: AccessLimited, // Доступные документы
		},
		ResourceStudents: {
			ActionRead:   AccessOwn, // Только свои данные
			ActionUpdate: AccessOwn, // Редактирование профиля
		},
		ResourceEvents: {
			ActionRead: AccessFull, // Все мероприятия
		},
		ResourceTasks: {
			ActionRead:    AccessFull, // ВАЖНО: Видит все задачи (только просмотр!)
			ActionExecute: AccessFull, // Может выполнять задачи
		},
		ResourceAssignments: {
			ActionRead:    AccessOwn,  // Свои задания
			ActionExecute: AccessFull, // Выполнение заданий
		},
		ResourceSchedule: {
			ActionRead: AccessFull, // Всё расписание
		},
		ResourceReports: {
			ActionRead: AccessOwn, // Свои отчёты/оценки
		},
		ResourceAnnouncements: {
			ActionRead: AccessFull, // Все объявления
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
