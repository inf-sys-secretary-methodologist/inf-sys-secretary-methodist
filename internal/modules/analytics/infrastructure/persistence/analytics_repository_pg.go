// Package persistence provides PostgreSQL implementations for analytics repositories.
package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
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

// GetAtRiskStudents returns students with high or critical risk levels.
// When scope is non-nil, results and the count are filtered to the
// scope's whitelist (group_name = ANY); the SQL implementation is
// supplied in Cycle 3b — for now scope is plumbed through but ignored
// so the rest of the stack compiles.
func (r *AnalyticsRepositoryPG) GetAtRiskStudents(ctx context.Context, _ *entities.TeacherScope, limit, offset int) ([]entities.StudentRiskScore, int64, error) {
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
	defer func() { _ = rows.Close() }()

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
	if errors.Is(err, sql.ErrNoRows) {
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
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("group not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get group summary: %w", err)
	}

	return &s, nil
}

// GetAllGroupsSummary returns analytics summary for all groups.
// scope is plumbed through; SQL filter added in Cycle 3b.
func (r *AnalyticsRepositoryPG) GetAllGroupsSummary(ctx context.Context, _ *entities.TeacherScope) ([]entities.GroupAnalyticsSummary, error) {
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
	defer func() { _ = rows.Close() }()

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

// GetStudentsByRiskLevel returns students filtered by risk level.
// scope is plumbed through; SQL filter added in Cycle 3b.
func (r *AnalyticsRepositoryPG) GetStudentsByRiskLevel(ctx context.Context, _ *entities.TeacherScope, riskLevel entities.RiskLevel, limit, offset int) ([]entities.StudentRiskScore, int64, error) {
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
	defer func() { _ = rows.Close() }()

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
	defer func() { _ = rows.Close() }()

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

// GetRiskWeightConfig returns the current risk weight configuration.
func (r *AnalyticsRepositoryPG) GetRiskWeightConfig(ctx context.Context) (*entities.RiskWeightConfig, error) {
	query := `SELECT id, attendance_weight, grade_weight, submission_weight, inactivity_weight,
		high_risk_threshold, critical_risk_threshold, updated_by, updated_at
		FROM risk_weight_config ORDER BY id LIMIT 1`

	var cfg entities.RiskWeightConfig
	err := r.db.QueryRowContext(ctx, query).Scan(
		&cfg.ID, &cfg.AttendanceWeight, &cfg.GradeWeight, &cfg.SubmissionWeight, &cfg.InactivityWeight,
		&cfg.HighRiskThreshold, &cfg.CriticalRiskThreshold, &cfg.UpdatedBy, &cfg.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Return defaults
			return &entities.RiskWeightConfig{
				AttendanceWeight:      0.35,
				GradeWeight:           0.30,
				SubmissionWeight:      0.20,
				InactivityWeight:      0.15,
				HighRiskThreshold:     70.0,
				CriticalRiskThreshold: 85.0,
			}, nil
		}
		return nil, fmt.Errorf("failed to get risk weight config: %w", err)
	}
	return &cfg, nil
}

// UpdateRiskWeightConfig updates the risk weight configuration.
func (r *AnalyticsRepositoryPG) UpdateRiskWeightConfig(ctx context.Context, cfg *entities.RiskWeightConfig) error {
	query := `UPDATE risk_weight_config SET
		attendance_weight = $1, grade_weight = $2, submission_weight = $3, inactivity_weight = $4,
		high_risk_threshold = $5, critical_risk_threshold = $6,
		updated_by = $7, updated_at = NOW()
		WHERE id = (SELECT id FROM risk_weight_config ORDER BY id LIMIT 1)`

	_, err := r.db.ExecContext(ctx, query,
		cfg.AttendanceWeight, cfg.GradeWeight, cfg.SubmissionWeight, cfg.InactivityWeight,
		cfg.HighRiskThreshold, cfg.CriticalRiskThreshold, cfg.UpdatedBy,
	)
	if err != nil {
		return fmt.Errorf("failed to update risk weight config: %w", err)
	}
	return nil
}

// SaveRiskHistory saves a risk score snapshot for a student.
func (r *AnalyticsRepositoryPG) SaveRiskHistory(ctx context.Context, entry *entities.RiskHistoryEntry) error {
	var factorsJSON []byte
	if entry.RiskFactors != nil {
		var err error
		factorsJSON, err = json.Marshal(entry.RiskFactors)
		if err != nil {
			return fmt.Errorf("failed to marshal risk factors: %w", err)
		}
	}

	query := `INSERT INTO student_risk_history
		(student_id, risk_score, risk_level, attendance_rate, grade_average, submission_rate, risk_factors, calculated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.ExecContext(ctx, query,
		entry.StudentID, entry.RiskScore, entry.RiskLevel,
		entry.AttendanceRate, entry.GradeAverage, entry.SubmissionRate,
		factorsJSON, entry.CalculatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save risk history: %w", err)
	}
	return nil
}

// GetStudentRiskHistory returns risk history for a student, ordered by date descending.
func (r *AnalyticsRepositoryPG) GetStudentRiskHistory(ctx context.Context, studentID int64, limit int) ([]entities.RiskHistoryEntry, error) {
	if limit <= 0 || limit > 365 {
		limit = 90
	}

	query := `SELECT id, student_id, risk_score, risk_level, attendance_rate, grade_average, submission_rate, risk_factors, calculated_at
		FROM student_risk_history
		WHERE student_id = $1
		ORDER BY calculated_at DESC
		LIMIT $2`

	rows, err := r.db.QueryContext(ctx, query, studentID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get risk history: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var history []entities.RiskHistoryEntry
	for rows.Next() {
		var e entities.RiskHistoryEntry
		var factorsJSON []byte
		err := rows.Scan(
			&e.ID, &e.StudentID, &e.RiskScore, &e.RiskLevel,
			&e.AttendanceRate, &e.GradeAverage, &e.SubmissionRate,
			&factorsJSON, &e.CalculatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan risk history: %w", err)
		}
		if len(factorsJSON) > 0 {
			var factors entities.RiskFactors
			if jsonErr := json.Unmarshal(factorsJSON, &factors); jsonErr == nil {
				e.RiskFactors = &factors
			}
		}
		history = append(history, e)
	}

	return history, nil
}
