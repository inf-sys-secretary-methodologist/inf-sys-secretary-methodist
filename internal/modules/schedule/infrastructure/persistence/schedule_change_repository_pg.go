package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
)

// ScheduleChangeRepositoryPG implements ScheduleChangeRepository using PostgreSQL.
type ScheduleChangeRepositoryPG struct {
	db *sql.DB
}

// NewScheduleChangeRepositoryPG creates a new ScheduleChangeRepositoryPG.
func NewScheduleChangeRepositoryPG(db *sql.DB) *ScheduleChangeRepositoryPG {
	return &ScheduleChangeRepositoryPG{db: db}
}

// Create inserts a new schedule change.
func (r *ScheduleChangeRepositoryPG) Create(ctx context.Context, change *entities.ScheduleChange) error {
	query := `
		INSERT INTO schedule_changes (
			lesson_id, change_type, original_date, new_date,
			new_classroom_id, new_teacher_id, reason, created_by, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`

	return r.db.QueryRowContext(ctx, query,
		change.LessonID, change.ChangeType, change.OriginalDate, change.NewDate,
		change.NewClassroomID, change.NewTeacherID, change.Reason,
		change.CreatedBy, change.CreatedAt,
	).Scan(&change.ID)
}

// GetByLessonID retrieves schedule changes for a lesson.
func (r *ScheduleChangeRepositoryPG) GetByLessonID(ctx context.Context, lessonID int64) ([]*entities.ScheduleChange, error) {
	query := `
		SELECT id, lesson_id, change_type, original_date, new_date,
			new_classroom_id, new_teacher_id, reason, created_by, created_at
		FROM schedule_changes WHERE lesson_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, lessonID)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedule changes: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return r.scanChanges(rows)
}

// GetByDateRange retrieves schedule changes in a date range.
func (r *ScheduleChangeRepositoryPG) GetByDateRange(ctx context.Context, start, end time.Time) ([]*entities.ScheduleChange, error) {
	query := `
		SELECT id, lesson_id, change_type, original_date, new_date,
			new_classroom_id, new_teacher_id, reason, created_by, created_at
		FROM schedule_changes
		WHERE original_date >= $1 AND original_date <= $2
		ORDER BY original_date`

	rows, err := r.db.QueryContext(ctx, query, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedule changes by date range: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return r.scanChanges(rows)
}

func (r *ScheduleChangeRepositoryPG) scanChanges(rows *sql.Rows) ([]*entities.ScheduleChange, error) {
	var changes []*entities.ScheduleChange

	for rows.Next() {
		c := &entities.ScheduleChange{}
		err := rows.Scan(
			&c.ID, &c.LessonID, &c.ChangeType, &c.OriginalDate, &c.NewDate,
			&c.NewClassroomID, &c.NewTeacherID, &c.Reason, &c.CreatedBy, &c.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan schedule change: %w", err)
		}
		changes = append(changes, c)
	}

	return changes, rows.Err()
}
