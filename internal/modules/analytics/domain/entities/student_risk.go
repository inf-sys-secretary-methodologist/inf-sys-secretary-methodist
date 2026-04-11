// Package entities contains domain entities for the analytics module.
package entities

import "time"

// RiskLevel represents the risk level of a student
type RiskLevel string

// RiskLevel values.
const (
	RiskLevelLow      RiskLevel = "low"
	RiskLevelMedium   RiskLevel = "medium"
	RiskLevelHigh     RiskLevel = "high"
	RiskLevelCritical RiskLevel = "critical"
	RiskLevelUnknown  RiskLevel = "unknown"
)

// StudentRiskScore represents a student's risk assessment
type StudentRiskScore struct {
	StudentID      int64        `json:"student_id"`
	StudentName    string       `json:"student_name"`
	GroupName      *string      `json:"group_name,omitempty"`
	AttendanceRate *float64     `json:"attendance_rate,omitempty"`
	GradeAverage   *float64     `json:"grade_average,omitempty"`
	RiskLevel      RiskLevel    `json:"risk_level"`
	RiskScore      float64      `json:"risk_score"` // 0-100, higher = more at risk
	RiskFactors    *RiskFactors `json:"risk_factors,omitempty"`
	CreatedAt      time.Time    `json:"created_at,omitempty"`
	UpdatedAt      time.Time    `json:"updated_at,omitempty"`
}

// RiskFactors contains detailed risk factor information
type RiskFactors struct {
	Attendance AttendanceRiskFactor `json:"attendance"`
	Grades     GradesRiskFactor     `json:"grades"`
}

// AttendanceRiskFactor contains attendance-related risk info
type AttendanceRiskFactor struct {
	Rate        float64 `json:"rate"`
	AbsentCount int     `json:"absent_count"`
	IsRisk      bool    `json:"is_risk"`
}

// GradesRiskFactor contains grade-related risk info
type GradesRiskFactor struct {
	Average      float64 `json:"average"`
	FailingCount int     `json:"failing_count"`
	IsRisk       bool    `json:"is_risk"`
}

// StudentRiskAssessment represents a cached risk assessment
type StudentRiskAssessment struct {
	ID               int64       `json:"id"`
	StudentID        int64       `json:"student_id"`
	AttendanceRate   *float64    `json:"attendance_rate,omitempty"`
	GradeAverage     *float64    `json:"grade_average,omitempty"`
	RiskLevel        RiskLevel   `json:"risk_level"`
	RiskScore        *float64    `json:"risk_score,omitempty"`
	RiskFactors      interface{} `json:"risk_factors,omitempty"`
	LastCalculatedAt *time.Time  `json:"last_calculated_at,omitempty"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
}

// GroupAnalyticsSummary represents analytics summary for a student group
type GroupAnalyticsSummary struct {
	GroupName         string  `json:"group_name"`
	TotalStudents     int     `json:"total_students"`
	AvgAttendanceRate float64 `json:"avg_attendance_rate"`
	AvgGrade          float64 `json:"avg_grade"`
	CriticalRiskCount int     `json:"critical_risk_count"`
	HighRiskCount     int     `json:"high_risk_count"`
	MediumRiskCount   int     `json:"medium_risk_count"`
	LowRiskCount      int     `json:"low_risk_count"`
	AtRiskPercentage  float64 `json:"at_risk_percentage"`
}

// MonthlyAttendanceTrend represents monthly attendance statistics
type MonthlyAttendanceTrend struct {
	Month          time.Time `json:"month"`
	UniqueStudents int       `json:"unique_students"`
	TotalRecords   int       `json:"total_records"`
	PresentCount   int       `json:"present_count"`
	AbsentCount    int       `json:"absent_count"`
	AttendanceRate float64   `json:"attendance_rate"`
}

// RiskWeightConfig holds admin-configurable risk score weights
type RiskWeightConfig struct {
	ID                     int       `json:"id"`
	AttendanceWeight       float64   `json:"attendance_weight"`
	GradeWeight            float64   `json:"grade_weight"`
	SubmissionWeight       float64   `json:"submission_weight"`
	InactivityWeight       float64   `json:"inactivity_weight"`
	HighRiskThreshold      float64   `json:"high_risk_threshold"`
	CriticalRiskThreshold  float64   `json:"critical_risk_threshold"`
	UpdatedBy              *int64    `json:"updated_by,omitempty"`
	UpdatedAt              time.Time `json:"updated_at"`
}

// RiskHistoryEntry represents a daily risk score snapshot for a student
type RiskHistoryEntry struct {
	ID             int64      `json:"id"`
	StudentID      int64      `json:"student_id"`
	RiskScore      float64    `json:"risk_score"`
	RiskLevel      RiskLevel  `json:"risk_level"`
	AttendanceRate *float64   `json:"attendance_rate,omitempty"`
	GradeAverage   *float64   `json:"grade_average,omitempty"`
	SubmissionRate *float64   `json:"submission_rate,omitempty"`
	RiskFactors    *RiskFactors `json:"risk_factors,omitempty"`
	CalculatedAt   time.Time  `json:"calculated_at"`
}

// AttendanceStats holds attendance statistics for a student
type AttendanceStats struct {
	StudentID      int64   `json:"student_id"`
	StudentName    string  `json:"student_name"`
	GroupName      *string `json:"group_name,omitempty"`
	TotalRecords   int     `json:"total_records"`
	PresentCount   int     `json:"present_count"`
	AbsentCount    int     `json:"absent_count"`
	LateCount      int     `json:"late_count"`
	ExcusedCount   int     `json:"excused_count"`
	AttendanceRate float64 `json:"attendance_rate"`
}

// GradeStats holds grade statistics for a student
type GradeStats struct {
	StudentID          int64   `json:"student_id"`
	StudentName        string  `json:"student_name"`
	GroupName          *string `json:"group_name,omitempty"`
	TotalGrades        int     `json:"total_grades"`
	GradeAverage       float64 `json:"grade_average"`
	WeightedAverage    float64 `json:"weighted_average"`
	MinGrade           float64 `json:"min_grade"`
	MaxGrade           float64 `json:"max_grade"`
	FailingGradesCount int     `json:"failing_grades_count"`
}
