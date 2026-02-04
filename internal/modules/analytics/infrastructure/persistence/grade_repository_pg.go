// Package persistence provides PostgreSQL implementations for analytics repositories.
package persistence

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/analytics/domain/entities"
)

// GradeRepositoryPG implements GradeRepository using PostgreSQL
type GradeRepositoryPG struct {
	db *sql.DB
}

// NewGradeRepositoryPG creates a new GradeRepositoryPG
func NewGradeRepositoryPG(db *sql.DB) *GradeRepositoryPG {
	return &GradeRepositoryPG{db: db}
}

// CreateGrade creates a new grade record
func (r *GradeRepositoryPG) CreateGrade(ctx context.Context, grade *entities.Grade) error {
	query := `
		INSERT INTO grades (student_id, subject, grade_type, grade_value, max_value, weight, graded_by, grade_date, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		RETURNING id, created_at, updated_at`

	return r.db.QueryRowContext(ctx, query,
		grade.StudentID, grade.Subject, grade.GradeType, grade.GradeValue,
		grade.MaxValue, grade.Weight, grade.GradedBy, grade.GradeDate, grade.Notes,
	).Scan(&grade.ID, &grade.CreatedAt, &grade.UpdatedAt)
}

// GetGradesByStudent retrieves all grades for a student
func (r *GradeRepositoryPG) GetGradesByStudent(ctx context.Context, studentID int64) ([]entities.Grade, error) {
	query := `
		SELECT id, student_id, subject, grade_type, grade_value, max_value, weight, graded_by, grade_date, notes, created_at, updated_at
		FROM grades
		WHERE student_id = $1
		ORDER BY grade_date DESC, subject`

	rows, err := r.db.QueryContext(ctx, query, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get grades: %w", err)
	}
	defer rows.Close()

	var grades []entities.Grade
	for rows.Next() {
		var grade entities.Grade
		err := rows.Scan(
			&grade.ID, &grade.StudentID, &grade.Subject, &grade.GradeType,
			&grade.GradeValue, &grade.MaxValue, &grade.Weight, &grade.GradedBy,
			&grade.GradeDate, &grade.Notes, &grade.CreatedAt, &grade.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan grade: %w", err)
		}
		grades = append(grades, grade)
	}
	return grades, nil
}

// GetGradesBySubject retrieves grades for a student in a specific subject
func (r *GradeRepositoryPG) GetGradesBySubject(ctx context.Context, studentID int64, subject string) ([]entities.Grade, error) {
	query := `
		SELECT id, student_id, subject, grade_type, grade_value, max_value, weight, graded_by, grade_date, notes, created_at, updated_at
		FROM grades
		WHERE student_id = $1 AND subject = $2
		ORDER BY grade_date DESC`

	rows, err := r.db.QueryContext(ctx, query, studentID, subject)
	if err != nil {
		return nil, fmt.Errorf("failed to get grades: %w", err)
	}
	defer rows.Close()

	var grades []entities.Grade
	for rows.Next() {
		var grade entities.Grade
		err := rows.Scan(
			&grade.ID, &grade.StudentID, &grade.Subject, &grade.GradeType,
			&grade.GradeValue, &grade.MaxValue, &grade.Weight, &grade.GradedBy,
			&grade.GradeDate, &grade.Notes, &grade.CreatedAt, &grade.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan grade: %w", err)
		}
		grades = append(grades, grade)
	}
	return grades, nil
}

// UpdateGrade updates an existing grade
func (r *GradeRepositoryPG) UpdateGrade(ctx context.Context, grade *entities.Grade) error {
	query := `
		UPDATE grades SET
			subject = $1, grade_type = $2, grade_value = $3, max_value = $4,
			weight = $5, graded_by = $6, grade_date = $7, notes = $8, updated_at = NOW()
		WHERE id = $9
		RETURNING updated_at`

	return r.db.QueryRowContext(ctx, query,
		grade.Subject, grade.GradeType, grade.GradeValue, grade.MaxValue,
		grade.Weight, grade.GradedBy, grade.GradeDate, grade.Notes, grade.ID,
	).Scan(&grade.UpdatedAt)
}

// DeleteGrade deletes a grade
func (r *GradeRepositoryPG) DeleteGrade(ctx context.Context, id int64) error {
	query := `DELETE FROM grades WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete grade: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("grade not found")
	}
	return nil
}

// GetStudentGradeStats returns grade statistics for a student
func (r *GradeRepositoryPG) GetStudentGradeStats(ctx context.Context, studentID int64) (*entities.GradeStats, error) {
	query := `
		SELECT student_id, student_name, group_name,
			total_grades, grade_average, weighted_average, min_grade, max_grade, failing_grades_count
		FROM v_student_grade_stats
		WHERE student_id = $1`

	var stats entities.GradeStats
	err := r.db.QueryRowContext(ctx, query, studentID).Scan(
		&stats.StudentID, &stats.StudentName, &stats.GroupName,
		&stats.TotalGrades, &stats.GradeAverage, &stats.WeightedAverage,
		&stats.MinGrade, &stats.MaxGrade, &stats.FailingGradesCount,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("student not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get grade stats: %w", err)
	}
	return &stats, nil
}
