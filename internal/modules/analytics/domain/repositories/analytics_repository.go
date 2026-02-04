// Package repositories defines interfaces for analytics persistence.
package repositories

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/analytics/domain/entities"
)

// AnalyticsRepository defines the interface for analytics persistence operations
type AnalyticsRepository interface {
	// Risk assessment queries
	GetAtRiskStudents(ctx context.Context, limit, offset int) ([]entities.StudentRiskScore, int64, error)
	GetStudentRisk(ctx context.Context, studentID int64) (*entities.StudentRiskScore, error)
	GetGroupSummary(ctx context.Context, groupName string) (*entities.GroupAnalyticsSummary, error)
	GetAllGroupsSummary(ctx context.Context) ([]entities.GroupAnalyticsSummary, error)

	// Filtered risk queries
	GetStudentsByRiskLevel(ctx context.Context, riskLevel entities.RiskLevel, limit, offset int) ([]entities.StudentRiskScore, int64, error)

	// Trend analysis
	GetMonthlyAttendanceTrend(ctx context.Context, months int) ([]entities.MonthlyAttendanceTrend, error)
}

// AttendanceRepository defines the interface for attendance data operations
type AttendanceRepository interface {
	// Lesson operations
	CreateLesson(ctx context.Context, lesson *entities.Lesson) error
	GetLessonByID(ctx context.Context, id int64) (*entities.Lesson, error)
	GetLessonsByGroup(ctx context.Context, groupName string) ([]entities.Lesson, error)
	GetLessonsByTeacher(ctx context.Context, teacherID int64) ([]entities.Lesson, error)

	// Attendance record operations
	MarkAttendance(ctx context.Context, record *entities.AttendanceRecord) error
	BulkMarkAttendance(ctx context.Context, records []entities.AttendanceRecord) error
	GetAttendanceByLesson(ctx context.Context, lessonID int64, date string) ([]entities.AttendanceRecord, error)
	GetAttendanceByStudent(ctx context.Context, studentID int64, fromDate, toDate string) ([]entities.AttendanceRecord, error)
	UpdateAttendance(ctx context.Context, record *entities.AttendanceRecord) error

	// Statistics
	GetStudentAttendanceStats(ctx context.Context, studentID int64) (*entities.AttendanceStats, error)
}

// GradeRepository defines the interface for grade data operations
type GradeRepository interface {
	// Grade operations
	CreateGrade(ctx context.Context, grade *entities.Grade) error
	GetGradesByStudent(ctx context.Context, studentID int64) ([]entities.Grade, error)
	GetGradesBySubject(ctx context.Context, studentID int64, subject string) ([]entities.Grade, error)
	UpdateGrade(ctx context.Context, grade *entities.Grade) error
	DeleteGrade(ctx context.Context, id int64) error

	// Statistics
	GetStudentGradeStats(ctx context.Context, studentID int64) (*entities.GradeStats, error)
}
