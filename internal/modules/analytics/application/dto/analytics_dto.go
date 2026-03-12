// Package dto contains Data Transfer Objects for the analytics module.
package dto

import (
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/analytics/domain/entities"
)

// AtRiskStudentsResponse represents the response for at-risk students list
type AtRiskStudentsResponse struct {
	Students []StudentRiskResponse `json:"students"`
	Total    int64                 `json:"total"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"page_size"`
}

// StudentRiskResponse represents a single student's risk data
type StudentRiskResponse struct {
	StudentID      int64                 `json:"student_id"`
	StudentName    string                `json:"student_name"`
	GroupName      *string               `json:"group_name,omitempty"`
	AttendanceRate *float64              `json:"attendance_rate,omitempty"`
	GradeAverage   *float64              `json:"grade_average,omitempty"`
	RiskLevel      string                `json:"risk_level"`
	RiskScore      float64               `json:"risk_score"`
	RiskFactors    *entities.RiskFactors `json:"risk_factors,omitempty"`
}

// GroupSummaryResponse represents analytics summary for a group
type GroupSummaryResponse struct {
	GroupName         string  `json:"group_name"`
	TotalStudents     int     `json:"total_students"`
	AvgAttendanceRate float64 `json:"avg_attendance_rate"`
	AvgGrade          float64 `json:"avg_grade"`
	RiskDistribution  struct {
		Critical int `json:"critical"`
		High     int `json:"high"`
		Medium   int `json:"medium"`
		Low      int `json:"low"`
	} `json:"risk_distribution"`
	AtRiskPercentage float64 `json:"at_risk_percentage"`
}

// AllGroupsSummaryResponse represents summary for all groups
type AllGroupsSummaryResponse struct {
	Groups []GroupSummaryResponse `json:"groups"`
	Total  int                    `json:"total"`
}

// MonthlyTrendResponse represents monthly attendance trend data
type MonthlyTrendResponse struct {
	Month          string  `json:"month"` // Format: "2024-01"
	UniqueStudents int     `json:"unique_students"`
	TotalRecords   int     `json:"total_records"`
	PresentCount   int     `json:"present_count"`
	AbsentCount    int     `json:"absent_count"`
	AttendanceRate float64 `json:"attendance_rate"`
}

// AttendanceTrendResponse represents the full trend response
type AttendanceTrendResponse struct {
	Trends []MonthlyTrendResponse `json:"trends"`
	Months int                    `json:"months"`
}

// MarkAttendanceRequest represents a request to mark attendance
type MarkAttendanceRequest struct {
	StudentID  int64  `json:"student_id" validate:"required"`
	LessonID   int64  `json:"lesson_id" validate:"required"`
	LessonDate string `json:"lesson_date" validate:"required"` // Format: "2024-01-15"
	Status     string `json:"status" validate:"required,oneof=present absent late excused"`
	Notes      string `json:"notes,omitempty"`
}

// BulkMarkAttendanceRequest represents a request to mark attendance for multiple students
type BulkMarkAttendanceRequest struct {
	LessonID   int64                  `json:"lesson_id" validate:"required"`
	LessonDate string                 `json:"lesson_date" validate:"required"`
	Records    []BulkAttendanceRecord `json:"records" validate:"required,min=1"`
}

// BulkAttendanceRecord represents a single record in bulk attendance marking
type BulkAttendanceRecord struct {
	StudentID int64  `json:"student_id" validate:"required"`
	Status    string `json:"status" validate:"required,oneof=present absent late excused"`
	Notes     string `json:"notes,omitempty"`
}

// AttendanceRecordResponse represents an attendance record response
type AttendanceRecordResponse struct {
	ID         int64     `json:"id"`
	StudentID  int64     `json:"student_id"`
	LessonID   int64     `json:"lesson_id"`
	LessonDate string    `json:"lesson_date"`
	Status     string    `json:"status"`
	MarkedBy   *int64    `json:"marked_by,omitempty"`
	Notes      *string   `json:"notes,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// LessonAttendanceResponse represents attendance for a specific lesson
type LessonAttendanceResponse struct {
	LessonID   int64                      `json:"lesson_id"`
	LessonDate string                     `json:"lesson_date"`
	Records    []AttendanceRecordResponse `json:"records"`
	Summary    struct {
		Total   int `json:"total"`
		Present int `json:"present"`
		Absent  int `json:"absent"`
		Late    int `json:"late"`
		Excused int `json:"excused"`
	} `json:"summary"`
}

// CreateLessonRequest represents a request to create a lesson
type CreateLessonRequest struct {
	Name       string  `json:"name" validate:"required,min=1,max=255"`
	Subject    string  `json:"subject" validate:"required,min=1,max=255"`
	TeacherID  *int64  `json:"teacher_id,omitempty"`
	GroupName  *string `json:"group_name,omitempty"`
	LessonType string  `json:"lesson_type" validate:"omitempty,oneof=lecture practice lab seminar exam"`
}

// ToStudentRiskResponse converts entity to DTO
func ToStudentRiskResponse(risk *entities.StudentRiskScore) *StudentRiskResponse {
	return &StudentRiskResponse{
		StudentID:      risk.StudentID,
		StudentName:    risk.StudentName,
		GroupName:      risk.GroupName,
		AttendanceRate: risk.AttendanceRate,
		GradeAverage:   risk.GradeAverage,
		RiskLevel:      string(risk.RiskLevel),
		RiskScore:      risk.RiskScore,
		RiskFactors:    risk.RiskFactors,
	}
}

// ToGroupSummaryResponse converts entity to DTO
func ToGroupSummaryResponse(summary *entities.GroupAnalyticsSummary) *GroupSummaryResponse {
	resp := &GroupSummaryResponse{
		GroupName:         summary.GroupName,
		TotalStudents:     summary.TotalStudents,
		AvgAttendanceRate: summary.AvgAttendanceRate,
		AvgGrade:          summary.AvgGrade,
		AtRiskPercentage:  summary.AtRiskPercentage,
	}
	resp.RiskDistribution.Critical = summary.CriticalRiskCount
	resp.RiskDistribution.High = summary.HighRiskCount
	resp.RiskDistribution.Medium = summary.MediumRiskCount
	resp.RiskDistribution.Low = summary.LowRiskCount
	return resp
}

// ToMonthlyTrendResponse converts entity to DTO
func ToMonthlyTrendResponse(trend *entities.MonthlyAttendanceTrend) *MonthlyTrendResponse {
	return &MonthlyTrendResponse{
		Month:          trend.Month.Format("2006-01"),
		UniqueStudents: trend.UniqueStudents,
		TotalRecords:   trend.TotalRecords,
		PresentCount:   trend.PresentCount,
		AbsentCount:    trend.AbsentCount,
		AttendanceRate: trend.AttendanceRate,
	}
}

// ToAttendanceRecordResponse converts entity to DTO
func ToAttendanceRecordResponse(record *entities.AttendanceRecord) *AttendanceRecordResponse {
	return &AttendanceRecordResponse{
		ID:         record.ID,
		StudentID:  record.StudentID,
		LessonID:   record.LessonID,
		LessonDate: record.LessonDate.Format("2006-01-02"),
		Status:     string(record.Status),
		MarkedBy:   record.MarkedBy,
		Notes:      record.Notes,
		CreatedAt:  record.CreatedAt,
	}
}
