// Package persistence provides PostgreSQL implementations for analytics repositories.
package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/analytics/domain/entities"
)

// AnalyticsRepositoryPG implements AnalyticsRepository using PostgreSQL
type AnalyticsRepositoryPG struct {
	db *sql.DB
}

// NewAnalyticsRepositoryPG creates a new AnalyticsRepositoryPG
func NewAnalyticsRepositoryPG(db *sql.DB) *AnalyticsRepositoryPG {
	return &AnalyticsRepositoryPG{db: db}
}

// GetAtRiskStudents returns students with high or critical risk levels
func (r *AnalyticsRepositoryPG) GetAtRiskStudents(ctx context.Context, limit, offset int) ([]entities.StudentRiskScore, int64, error) {
	// Count total
	var total int64
	countQuery := `SELECT COUNT(*) FROM v_at_risk_students`
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count at-risk students: %w", err)
	}

	// Get paginated results
	query := `
		SELECT student_id, student_name, group_name, attendance_rate, grade_average,
			risk_level, risk_score, risk_factors
		FROM v_at_risk_students
		ORDER BY risk_score DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get at-risk students: %w", err)
	}
	defer rows.Close()

	var students []entities.StudentRiskScore
	for rows.Next() {
		var s entities.StudentRiskScore
		var riskFactorsJSON []byte
		err := rows.Scan(
			&s.StudentID, &s.StudentName, &s.GroupName,
			&s.AttendanceRate, &s.GradeAverage,
			&s.RiskLevel, &s.RiskScore, &riskFactorsJSON,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan student risk: %w", err)
		}
		if len(riskFactorsJSON) > 0 {
			s.RiskFactors = &entities.RiskFactors{}
			if err := json.Unmarshal(riskFactorsJSON, s.RiskFactors); err != nil {
				// Continue without risk factors if unmarshal fails
				s.RiskFactors = nil
			}
		}
		students = append(students, s)
	}

	return students, total, nil
}

// GetStudentRisk returns risk assessment for a specific student
func (r *AnalyticsRepositoryPG) GetStudentRisk(ctx context.Context, studentID int64) (*entities.StudentRiskScore, error) {
	query := `
		SELECT student_id, student_name, group_name, attendance_rate, grade_average,
			risk_level, risk_score, risk_factors
		FROM v_student_risk_score
		WHERE student_id = $1`

	var s entities.StudentRiskScore
	var riskFactorsJSON []byte
	err := r.db.QueryRowContext(ctx, query, studentID).Scan(
		&s.StudentID, &s.StudentName, &s.GroupName,
		&s.AttendanceRate, &s.GradeAverage,
		&s.RiskLevel, &s.RiskScore, &riskFactorsJSON,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("student not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get student risk: %w", err)
	}

	if len(riskFactorsJSON) > 0 {
		s.RiskFactors = &entities.RiskFactors{}
		if err := json.Unmarshal(riskFactorsJSON, s.RiskFactors); err != nil {
			s.RiskFactors = nil
		}
	}

	return &s, nil
}

// GetGroupSummary returns analytics summary for a specific group
func (r *AnalyticsRepositoryPG) GetGroupSummary(ctx context.Context, groupName string) (*entities.GroupAnalyticsSummary, error) {
	query := `
		SELECT group_name, total_students, avg_attendance_rate, avg_grade,
			critical_risk_count, high_risk_count, medium_risk_count, low_risk_count,
			at_risk_percentage
		FROM v_group_analytics_summary
		WHERE group_name = $1`

	var s entities.GroupAnalyticsSummary
	err := r.db.QueryRowContext(ctx, query, groupName).Scan(
		&s.GroupName, &s.TotalStudents, &s.AvgAttendanceRate, &s.AvgGrade,
		&s.CriticalRiskCount, &s.HighRiskCount, &s.MediumRiskCount, &s.LowRiskCount,
		&s.AtRiskPercentage,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("group not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get group summary: %w", err)
	}

	return &s, nil
}

// GetAllGroupsSummary returns analytics summary for all groups
func (r *AnalyticsRepositoryPG) GetAllGroupsSummary(ctx context.Context) ([]entities.GroupAnalyticsSummary, error) {
	query := `
		SELECT group_name, total_students, avg_attendance_rate, avg_grade,
			critical_risk_count, high_risk_count, medium_risk_count, low_risk_count,
			at_risk_percentage
		FROM v_group_analytics_summary
		ORDER BY at_risk_percentage DESC NULLS LAST`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all groups summary: %w", err)
	}
	defer rows.Close()

	var summaries []entities.GroupAnalyticsSummary
	for rows.Next() {
		var s entities.GroupAnalyticsSummary
		err := rows.Scan(
			&s.GroupName, &s.TotalStudents, &s.AvgAttendanceRate, &s.AvgGrade,
			&s.CriticalRiskCount, &s.HighRiskCount, &s.MediumRiskCount, &s.LowRiskCount,
			&s.AtRiskPercentage,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan group summary: %w", err)
		}
		summaries = append(summaries, s)
	}

	return summaries, nil
}

// GetStudentsByRiskLevel returns students filtered by risk level
func (r *AnalyticsRepositoryPG) GetStudentsByRiskLevel(ctx context.Context, riskLevel entities.RiskLevel, limit, offset int) ([]entities.StudentRiskScore, int64, error) {
	// Count total
	var total int64
	countQuery := `SELECT COUNT(*) FROM v_student_risk_score WHERE risk_level = $1`
	if err := r.db.QueryRowContext(ctx, countQuery, riskLevel).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count students: %w", err)
	}

	// Get paginated results
	query := `
		SELECT student_id, student_name, group_name, attendance_rate, grade_average,
			risk_level, risk_score, risk_factors
		FROM v_student_risk_score
		WHERE risk_level = $1
		ORDER BY risk_score DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, riskLevel, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get students by risk level: %w", err)
	}
	defer rows.Close()

	var students []entities.StudentRiskScore
	for rows.Next() {
		var s entities.StudentRiskScore
		var riskFactorsJSON []byte
		err := rows.Scan(
			&s.StudentID, &s.StudentName, &s.GroupName,
			&s.AttendanceRate, &s.GradeAverage,
			&s.RiskLevel, &s.RiskScore, &riskFactorsJSON,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan student risk: %w", err)
		}
		if len(riskFactorsJSON) > 0 {
			s.RiskFactors = &entities.RiskFactors{}
			if err := json.Unmarshal(riskFactorsJSON, s.RiskFactors); err != nil {
				s.RiskFactors = nil
			}
		}
		students = append(students, s)
	}

	return students, total, nil
}

// GetMonthlyAttendanceTrend returns monthly attendance trend data
func (r *AnalyticsRepositoryPG) GetMonthlyAttendanceTrend(ctx context.Context, months int) ([]entities.MonthlyAttendanceTrend, error) {
	query := `
		SELECT month, unique_students, total_records, present_count, absent_count, attendance_rate
		FROM v_monthly_attendance_trend
		ORDER BY month DESC
		LIMIT $1`

	rows, err := r.db.QueryContext(ctx, query, months)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly attendance trend: %w", err)
	}
	defer rows.Close()

	var trends []entities.MonthlyAttendanceTrend
	for rows.Next() {
		var t entities.MonthlyAttendanceTrend
		err := rows.Scan(
			&t.Month, &t.UniqueStudents, &t.TotalRecords,
			&t.PresentCount, &t.AbsentCount, &t.AttendanceRate,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan monthly trend: %w", err)
		}
		trends = append(trends, t)
	}

	return trends, nil
}
