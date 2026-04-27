package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/repositories"
)

// LessonRepositoryPG implements LessonRepository using PostgreSQL.
type LessonRepositoryPG struct {
	db *sql.DB
}

// NewLessonRepositoryPG creates a new LessonRepositoryPG.
func NewLessonRepositoryPG(db *sql.DB) *LessonRepositoryPG {
	return &LessonRepositoryPG{db: db}
}

// Create inserts a new lesson.
func (r *LessonRepositoryPG) Create(ctx context.Context, lesson *entities.Lesson) error {
	query := `
		INSERT INTO schedule_lessons (
			semester_id, discipline_id, lesson_type_id, teacher_id,
			group_id, classroom_id, day_of_week, time_start, time_end,
			week_type, date_start, date_end, notes, is_cancelled,
			cancellation_reason, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING id`

	return r.db.QueryRowContext(ctx, query,
		lesson.SemesterID, lesson.DisciplineID, lesson.LessonTypeID, lesson.TeacherID,
		lesson.GroupID, lesson.ClassroomID, lesson.DayOfWeek, lesson.TimeStart, lesson.TimeEnd,
		lesson.WeekType, lesson.DateStart, lesson.DateEnd, lesson.Notes, lesson.IsCancelled,
		lesson.CancelReason, lesson.CreatedAt, lesson.UpdatedAt,
	).Scan(&lesson.ID)
}

// Save updates an existing lesson.
func (r *LessonRepositoryPG) Save(ctx context.Context, lesson *entities.Lesson) error {
	query := `
		UPDATE schedule_lessons SET
			semester_id = $1, discipline_id = $2, lesson_type_id = $3, teacher_id = $4,
			group_id = $5, classroom_id = $6, day_of_week = $7, time_start = $8, time_end = $9,
			week_type = $10, date_start = $11, date_end = $12, notes = $13, is_cancelled = $14,
			cancellation_reason = $15, updated_at = $16
		WHERE id = $17`

	_, err := r.db.ExecContext(ctx, query,
		lesson.SemesterID, lesson.DisciplineID, lesson.LessonTypeID, lesson.TeacherID,
		lesson.GroupID, lesson.ClassroomID, lesson.DayOfWeek, lesson.TimeStart, lesson.TimeEnd,
		lesson.WeekType, lesson.DateStart, lesson.DateEnd, lesson.Notes, lesson.IsCancelled,
		lesson.CancelReason, lesson.UpdatedAt, lesson.ID,
	)
	return err
}

// GetByID retrieves a lesson by ID with associated data.
func (r *LessonRepositoryPG) GetByID(ctx context.Context, id int64) (*entities.Lesson, error) {
	query := `
		SELECT
			l.id, l.semester_id, l.discipline_id, l.lesson_type_id, l.teacher_id,
			l.group_id, l.classroom_id, l.day_of_week, l.time_start, l.time_end,
			l.week_type, l.date_start, l.date_end, l.notes, l.is_cancelled,
			l.cancellation_reason, l.created_at, l.updated_at,
			d.id, d.name, d.code, d.department_id, d.credits, d.hours_total,
			d.hours_lectures, d.hours_practice, d.hours_labs,
			lt.id, lt.name, lt.short_name, lt.color,
			cr.id, cr.building, cr.number, cr.name, cr.capacity, cr.type,
			cr.is_available, cr.created_at, cr.updated_at,
			sg.id, sg.specialty_id, sg.name, sg.course, sg.curator_id, sg.capacity,
			u.id, u.first_name || ' ' || u.last_name, u.email
		FROM schedule_lessons l
		LEFT JOIN disciplines d ON l.discipline_id = d.id
		LEFT JOIN lesson_types lt ON l.lesson_type_id = lt.id
		LEFT JOIN classrooms cr ON l.classroom_id = cr.id
		LEFT JOIN student_groups sg ON l.group_id = sg.id
		LEFT JOIN users u ON l.teacher_id = u.id
		WHERE l.id = $1`

	lesson := &entities.Lesson{}
	disc := &entities.Discipline{}
	ltype := &entities.LessonType{}
	classroom := &entities.Classroom{}
	group := &entities.StudentGroup{}
	teacher := &entities.TeacherInfo{}

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&lesson.ID, &lesson.SemesterID, &lesson.DisciplineID, &lesson.LessonTypeID, &lesson.TeacherID,
		&lesson.GroupID, &lesson.ClassroomID, &lesson.DayOfWeek, &lesson.TimeStart, &lesson.TimeEnd,
		&lesson.WeekType, &lesson.DateStart, &lesson.DateEnd, &lesson.Notes, &lesson.IsCancelled,
		&lesson.CancelReason, &lesson.CreatedAt, &lesson.UpdatedAt,
		&disc.ID, &disc.Name, &disc.Code, &disc.DepartmentID, &disc.Credits, &disc.HoursTotal,
		&disc.HoursLectures, &disc.HoursPractice, &disc.HoursLabs,
		&ltype.ID, &ltype.Name, &ltype.ShortName, &ltype.Color,
		&classroom.ID, &classroom.Building, &classroom.Number, &classroom.Name, &classroom.Capacity,
		&classroom.Type, &classroom.IsAvailable, &classroom.CreatedAt, &classroom.UpdatedAt,
		&group.ID, &group.SpecialtyID, &group.Name, &group.Course, &group.CuratorID, &group.Capacity,
		&teacher.ID, &teacher.Name, &teacher.Email,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get lesson: %w", err)
	}

	lesson.Discipline = disc
	lesson.LessonType = ltype
	lesson.Classroom = classroom
	lesson.Group = group
	lesson.Teacher = teacher

	return lesson, nil
}

// Delete deletes a lesson.
func (r *LessonRepositoryPG) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM schedule_lessons WHERE id = $1", id)
	return err
}

// List lists lessons with filters and pagination.
func (r *LessonRepositoryPG) List(ctx context.Context, filter repositories.LessonFilter, limit, offset int) ([]*entities.Lesson, error) {
	whereClause, args := r.buildWhereClause(filter)

	query := `
		SELECT id, semester_id, discipline_id, lesson_type_id, teacher_id,
			group_id, classroom_id, day_of_week, time_start, time_end,
			week_type, date_start, date_end, notes, is_cancelled,
			cancellation_reason, created_at, updated_at
		FROM schedule_lessons` + whereClause + ` ORDER BY day_of_week, time_start`

	argNum := len(args) + 1
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argNum, argNum+1)
		args = append(args, limit, offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list lessons: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return r.scanLessons(rows)
}

// Count counts lessons matching the filter.
func (r *LessonRepositoryPG) Count(ctx context.Context, filter repositories.LessonFilter) (int64, error) {
	whereClause, args := r.buildWhereClause(filter)

	var count int64
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM schedule_lessons"+whereClause, args...).Scan(&count)
	return count, err
}

// GetTimetable returns all lessons matching the filter with associations loaded.
func (r *LessonRepositoryPG) GetTimetable(ctx context.Context, filter repositories.LessonFilter) ([]*entities.Lesson, error) {
	whereClauseAliased, argsAliased := r.buildWhereClauseAliased(filter, "l")

	query := `
		SELECT
			l.id, l.semester_id, l.discipline_id, l.lesson_type_id, l.teacher_id,
			l.group_id, l.classroom_id, l.day_of_week, l.time_start, l.time_end,
			l.week_type, l.date_start, l.date_end, l.notes, l.is_cancelled,
			l.cancellation_reason, l.created_at, l.updated_at,
			d.id, d.name, d.code, d.department_id, d.credits, d.hours_total,
			d.hours_lectures, d.hours_practice, d.hours_labs,
			lt.id, lt.name, lt.short_name, lt.color,
			cr.id, cr.building, cr.number, cr.name, cr.capacity, cr.type,
			cr.is_available, cr.created_at, cr.updated_at,
			sg.id, sg.specialty_id, sg.name, sg.course, sg.curator_id, sg.capacity,
			u.id, u.first_name || ' ' || u.last_name, u.email
		FROM schedule_lessons l
		LEFT JOIN disciplines d ON l.discipline_id = d.id
		LEFT JOIN lesson_types lt ON l.lesson_type_id = lt.id
		LEFT JOIN classrooms cr ON l.classroom_id = cr.id
		LEFT JOIN student_groups sg ON l.group_id = sg.id
		LEFT JOIN users u ON l.teacher_id = u.id` + whereClauseAliased + ` ORDER BY l.day_of_week, l.time_start`

	rows, err := r.db.QueryContext(ctx, query, argsAliased...)
	if err != nil {
		return nil, fmt.Errorf("failed to get timetable: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return r.scanTimetableLessons(rows)
}

func (r *LessonRepositoryPG) buildWhereClause(filter repositories.LessonFilter) (string, []interface{}) {
	return r.buildWhereClauseAliased(filter, "")
}

func (r *LessonRepositoryPG) buildWhereClauseAliased(filter repositories.LessonFilter, alias string) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argNum := 1

	prefix := ""
	if alias != "" {
		prefix = alias + "."
	}

	if filter.SemesterID != nil {
		conditions = append(conditions, fmt.Sprintf("%ssemester_id = $%d", prefix, argNum))
		args = append(args, *filter.SemesterID)
		argNum++
	}
	if filter.GroupID != nil {
		conditions = append(conditions, fmt.Sprintf("%sgroup_id = $%d", prefix, argNum))
		args = append(args, *filter.GroupID)
		argNum++
	}
	if filter.TeacherID != nil {
		conditions = append(conditions, fmt.Sprintf("%steacher_id = $%d", prefix, argNum))
		args = append(args, *filter.TeacherID)
		argNum++
	}
	if filter.ClassroomID != nil {
		conditions = append(conditions, fmt.Sprintf("%sclassroom_id = $%d", prefix, argNum))
		args = append(args, *filter.ClassroomID)
		argNum++
	}
	if filter.DisciplineID != nil {
		conditions = append(conditions, fmt.Sprintf("%sdiscipline_id = $%d", prefix, argNum))
		args = append(args, *filter.DisciplineID)
		argNum++
	}
	if filter.DayOfWeek != nil {
		conditions = append(conditions, fmt.Sprintf("%sday_of_week = $%d", prefix, argNum))
		args = append(args, int(*filter.DayOfWeek))
		argNum++
	}
	if filter.WeekType != nil {
		conditions = append(conditions, fmt.Sprintf("%sweek_type = $%d", prefix, argNum))
		args = append(args, string(*filter.WeekType))
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	return whereClause, args
}

func (r *LessonRepositoryPG) scanLessons(rows *sql.Rows) ([]*entities.Lesson, error) {
	var lessons []*entities.Lesson

	for rows.Next() {
		l := &entities.Lesson{}
		err := rows.Scan(
			&l.ID, &l.SemesterID, &l.DisciplineID, &l.LessonTypeID, &l.TeacherID,
			&l.GroupID, &l.ClassroomID, &l.DayOfWeek, &l.TimeStart, &l.TimeEnd,
			&l.WeekType, &l.DateStart, &l.DateEnd, &l.Notes, &l.IsCancelled,
			&l.CancelReason, &l.CreatedAt, &l.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan lesson: %w", err)
		}
		lessons = append(lessons, l)
	}

	return lessons, rows.Err()
}

func (r *LessonRepositoryPG) scanTimetableLessons(rows *sql.Rows) ([]*entities.Lesson, error) {
	var lessons []*entities.Lesson

	for rows.Next() {
		l := &entities.Lesson{}
		disc := &entities.Discipline{}
		ltype := &entities.LessonType{}
		classroom := &entities.Classroom{}
		group := &entities.StudentGroup{}
		teacher := &entities.TeacherInfo{}

		err := rows.Scan(
			&l.ID, &l.SemesterID, &l.DisciplineID, &l.LessonTypeID, &l.TeacherID,
			&l.GroupID, &l.ClassroomID, &l.DayOfWeek, &l.TimeStart, &l.TimeEnd,
			&l.WeekType, &l.DateStart, &l.DateEnd, &l.Notes, &l.IsCancelled,
			&l.CancelReason, &l.CreatedAt, &l.UpdatedAt,
			&disc.ID, &disc.Name, &disc.Code, &disc.DepartmentID, &disc.Credits, &disc.HoursTotal,
			&disc.HoursLectures, &disc.HoursPractice, &disc.HoursLabs,
			&ltype.ID, &ltype.Name, &ltype.ShortName, &ltype.Color,
			&classroom.ID, &classroom.Building, &classroom.Number, &classroom.Name, &classroom.Capacity,
			&classroom.Type, &classroom.IsAvailable, &classroom.CreatedAt, &classroom.UpdatedAt,
			&group.ID, &group.SpecialtyID, &group.Name, &group.Course, &group.CuratorID, &group.Capacity,
			&teacher.ID, &teacher.Name, &teacher.Email,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan timetable lesson: %w", err)
		}

		l.Discipline = disc
		l.LessonType = ltype
		l.Classroom = classroom
		l.Group = group
		l.Teacher = teacher

		lessons = append(lessons, l)
	}

	return lessons, rows.Err()
}
