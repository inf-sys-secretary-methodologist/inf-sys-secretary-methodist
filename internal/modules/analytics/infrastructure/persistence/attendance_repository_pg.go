// Package persistence provides PostgreSQL implementations for analytics repositories.
package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/analytics/domain/entities"
)

// AttendanceRepositoryPG implements AttendanceRepository using PostgreSQL
type AttendanceRepositoryPG struct {
	db *sql.DB
}

// NewAttendanceRepositoryPG creates a new AttendanceRepositoryPG
func NewAttendanceRepositoryPG(db *sql.DB) *AttendanceRepositoryPG {
	return &AttendanceRepositoryPG{db: db}
}

// CreateLesson creates a new lesson
func (r *AttendanceRepositoryPG) CreateLesson(ctx context.Context, lesson *entities.Lesson) error {
	query := `
		INSERT INTO lessons (name, subject, teacher_id, group_name, lesson_type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, created_at, updated_at`

	return r.db.QueryRowContext(ctx, query,
		lesson.Name, lesson.Subject, lesson.TeacherID, lesson.GroupName, lesson.LessonType,
	).Scan(&lesson.ID, &lesson.CreatedAt, &lesson.UpdatedAt)
}

// GetLessonByID retrieves a lesson by ID
func (r *AttendanceRepositoryPG) GetLessonByID(ctx context.Context, id int64) (*entities.Lesson, error) {
	query := `
		SELECT id, name, subject, teacher_id, group_name, lesson_type, created_at, updated_at
		FROM lessons WHERE id = $1`

	var lesson entities.Lesson
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&lesson.ID, &lesson.Name, &lesson.Subject, &lesson.TeacherID,
		&lesson.GroupName, &lesson.LessonType, &lesson.CreatedAt, &lesson.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("lesson not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get lesson: %w", err)
	}
	return &lesson, nil
}

// GetLessonsByGroup retrieves lessons by group name
func (r *AttendanceRepositoryPG) GetLessonsByGroup(ctx context.Context, groupName string) ([]entities.Lesson, error) {
	query := `
		SELECT id, name, subject, teacher_id, group_name, lesson_type, created_at, updated_at
		FROM lessons WHERE group_name = $1 ORDER BY subject, name`

	rows, err := r.db.QueryContext(ctx, query, groupName)
	if err != nil {
		return nil, fmt.Errorf("failed to get lessons: %w", err)
	}
	defer rows.Close()

	var lessons []entities.Lesson
	for rows.Next() {
		var lesson entities.Lesson
		err := rows.Scan(
			&lesson.ID, &lesson.Name, &lesson.Subject, &lesson.TeacherID,
			&lesson.GroupName, &lesson.LessonType, &lesson.CreatedAt, &lesson.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan lesson: %w", err)
		}
		lessons = append(lessons, lesson)
	}
	return lessons, nil
}

// GetLessonsByTeacher retrieves lessons by teacher ID
func (r *AttendanceRepositoryPG) GetLessonsByTeacher(ctx context.Context, teacherID int64) ([]entities.Lesson, error) {
	query := `
		SELECT id, name, subject, teacher_id, group_name, lesson_type, created_at, updated_at
		FROM lessons WHERE teacher_id = $1 ORDER BY subject, name`

	rows, err := r.db.QueryContext(ctx, query, teacherID)
	if err != nil {
		return nil, fmt.Errorf("failed to get lessons: %w", err)
	}
	defer rows.Close()

	var lessons []entities.Lesson
	for rows.Next() {
		var lesson entities.Lesson
		err := rows.Scan(
			&lesson.ID, &lesson.Name, &lesson.Subject, &lesson.TeacherID,
			&lesson.GroupName, &lesson.LessonType, &lesson.CreatedAt, &lesson.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan lesson: %w", err)
		}
		lessons = append(lessons, lesson)
	}
	return lessons, nil
}

// MarkAttendance marks attendance for a single student
func (r *AttendanceRepositoryPG) MarkAttendance(ctx context.Context, record *entities.AttendanceRecord) error {
	query := `
		INSERT INTO attendance_records (student_id, lesson_id, lesson_date, status, marked_by, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		ON CONFLICT (student_id, lesson_id, lesson_date) DO UPDATE SET
			status = EXCLUDED.status,
			marked_by = EXCLUDED.marked_by,
			notes = EXCLUDED.notes,
			updated_at = NOW()
		RETURNING id, created_at, updated_at`

	return r.db.QueryRowContext(ctx, query,
		record.StudentID, record.LessonID, record.LessonDate,
		record.Status, record.MarkedBy, record.Notes,
	).Scan(&record.ID, &record.CreatedAt, &record.UpdatedAt)
}

// BulkMarkAttendance marks attendance for multiple students
func (r *AttendanceRepositoryPG) BulkMarkAttendance(ctx context.Context, records []entities.AttendanceRecord) error {
	if len(records) == 0 {
		return nil
	}

	// Build bulk insert query
	values := make([]string, 0, len(records))
	args := make([]interface{}, 0, len(records)*6)
	argNum := 1

	for _, r := range records {
		values = append(values, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, NOW(), NOW())",
			argNum, argNum+1, argNum+2, argNum+3, argNum+4, argNum+5))
		args = append(args, r.StudentID, r.LessonID, r.LessonDate, r.Status, r.MarkedBy, r.Notes)
		argNum += 6
	}

	query := fmt.Sprintf(`
		INSERT INTO attendance_records (student_id, lesson_id, lesson_date, status, marked_by, notes, created_at, updated_at)
		VALUES %s
		ON CONFLICT (student_id, lesson_id, lesson_date) DO UPDATE SET
			status = EXCLUDED.status,
			marked_by = EXCLUDED.marked_by,
			notes = EXCLUDED.notes,
			updated_at = NOW()`,
		strings.Join(values, ", ")) // #nosec G201 -- parameterized placeholders, not user input

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

// GetAttendanceByLesson retrieves attendance records for a specific lesson on a date
func (r *AttendanceRepositoryPG) GetAttendanceByLesson(ctx context.Context, lessonID int64, date string) ([]entities.AttendanceRecord, error) {
	query := `
		SELECT id, student_id, lesson_id, lesson_date, status, marked_by, notes, created_at, updated_at
		FROM attendance_records
		WHERE lesson_id = $1 AND lesson_date = $2
		ORDER BY student_id`

	rows, err := r.db.QueryContext(ctx, query, lessonID, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get attendance: %w", err)
	}
	defer rows.Close()

	var records []entities.AttendanceRecord
	for rows.Next() {
		var record entities.AttendanceRecord
		err := rows.Scan(
			&record.ID, &record.StudentID, &record.LessonID, &record.LessonDate,
			&record.Status, &record.MarkedBy, &record.Notes, &record.CreatedAt, &record.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan attendance record: %w", err)
		}
		records = append(records, record)
	}
	return records, nil
}

// GetAttendanceByStudent retrieves attendance records for a student in a date range
func (r *AttendanceRepositoryPG) GetAttendanceByStudent(ctx context.Context, studentID int64, fromDate, toDate string) ([]entities.AttendanceRecord, error) {
	query := `
		SELECT id, student_id, lesson_id, lesson_date, status, marked_by, notes, created_at, updated_at
		FROM attendance_records
		WHERE student_id = $1 AND lesson_date BETWEEN $2 AND $3
		ORDER BY lesson_date DESC`

	rows, err := r.db.QueryContext(ctx, query, studentID, fromDate, toDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get attendance: %w", err)
	}
	defer rows.Close()

	var records []entities.AttendanceRecord
	for rows.Next() {
		var record entities.AttendanceRecord
		err := rows.Scan(
			&record.ID, &record.StudentID, &record.LessonID, &record.LessonDate,
			&record.Status, &record.MarkedBy, &record.Notes, &record.CreatedAt, &record.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan attendance record: %w", err)
		}
		records = append(records, record)
	}
	return records, nil
}

// UpdateAttendance updates an attendance record
func (r *AttendanceRepositoryPG) UpdateAttendance(ctx context.Context, record *entities.AttendanceRecord) error {
	query := `
		UPDATE attendance_records SET
			status = $1, marked_by = $2, notes = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING updated_at`

	return r.db.QueryRowContext(ctx, query,
		record.Status, record.MarkedBy, record.Notes, record.ID,
	).Scan(&record.UpdatedAt)
}

// GetStudentAttendanceStats returns attendance statistics for a student
func (r *AttendanceRepositoryPG) GetStudentAttendanceStats(ctx context.Context, studentID int64) (*entities.AttendanceStats, error) {
	query := `
		SELECT student_id, student_name, group_name,
			total_records, present_count, absent_count, late_count, excused_count, attendance_rate
		FROM v_student_attendance_stats
		WHERE student_id = $1`

	var stats entities.AttendanceStats
	err := r.db.QueryRowContext(ctx, query, studentID).Scan(
		&stats.StudentID, &stats.StudentName, &stats.GroupName,
		&stats.TotalRecords, &stats.PresentCount, &stats.AbsentCount,
		&stats.LateCount, &stats.ExcusedCount, &stats.AttendanceRate,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("student not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get attendance stats: %w", err)
	}
	return &stats, nil
}
