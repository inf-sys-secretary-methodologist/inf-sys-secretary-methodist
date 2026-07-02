package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
)

// TeachingLoadRepositoryPG implements TeachingLoadRepository on PostgreSQL.
type TeachingLoadRepositoryPG struct {
	db *sql.DB
}

var _ usecases.TeachingLoadRepository = (*TeachingLoadRepositoryPG)(nil)

// NewTeachingLoadRepositoryPG creates a new TeachingLoadRepositoryPG.
func NewTeachingLoadRepositoryPG(db *sql.DB) *TeachingLoadRepositoryPG {
	return &TeachingLoadRepositoryPG{db: db}
}

// teachingLoadSelect is the hydrated read model shared by GetByID and List.
const teachingLoadSelect = `
	SELECT
		tl.id, tl.semester_id, tl.group_id, tl.discipline_id, tl.teacher_id, tl.lesson_type_id,
		tl.pairs_per_week, tl.week_type, tl.created_at, tl.updated_at,
		sg.id, sg.name, d.id, d.name, lt.id, lt.name, lt.short_name,
		u.id, COALESCE(u.name, ''), COALESCE(u.email, '')
	FROM teaching_loads tl
	LEFT JOIN student_groups sg ON tl.group_id = sg.id
	LEFT JOIN disciplines d ON tl.discipline_id = d.id
	LEFT JOIN lesson_types lt ON tl.lesson_type_id = lt.id
	LEFT JOIN users u ON tl.teacher_id = u.id`

// Create inserts a new load line, translating a duplicate key into ErrTeachingLoadDuplicate.
func (r *TeachingLoadRepositoryPG) Create(ctx context.Context, load *entities.TeachingLoad) error {
	const query = `
		INSERT INTO teaching_loads
			(semester_id, group_id, discipline_id, teacher_id, lesson_type_id, pairs_per_week, week_type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		load.SemesterID, load.GroupID, load.DisciplineID, load.TeacherID, load.LessonTypeID,
		load.PairsPerWeek, string(load.WeekType), load.CreatedAt, load.UpdatedAt,
	).Scan(&load.ID)
	if isUniqueViolation(err) {
		return entities.ErrTeachingLoadDuplicate
	}
	if err != nil {
		return fmt.Errorf("failed to create teaching load: %w", err)
	}
	return nil
}

// Update mutates an existing line. Missing row -> ErrTeachingLoadNotFound; duplicate -> ErrTeachingLoadDuplicate.
func (r *TeachingLoadRepositoryPG) Update(ctx context.Context, load *entities.TeachingLoad) error {
	const query = `
		UPDATE teaching_loads
		SET semester_id = $1, group_id = $2, discipline_id = $3, teacher_id = $4,
			lesson_type_id = $5, pairs_per_week = $6, week_type = $7, updated_at = $8
		WHERE id = $9`

	res, err := r.db.ExecContext(ctx, query,
		load.SemesterID, load.GroupID, load.DisciplineID, load.TeacherID, load.LessonTypeID,
		load.PairsPerWeek, string(load.WeekType), load.UpdatedAt, load.ID,
	)
	if isUniqueViolation(err) {
		return entities.ErrTeachingLoadDuplicate
	}
	if err != nil {
		return fmt.Errorf("failed to update teaching load: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to read update result: %w", err)
	}
	if affected == 0 {
		return entities.ErrTeachingLoadNotFound
	}
	return nil
}

// Delete removes a line by id, returning ErrTeachingLoadNotFound if none matched.
func (r *TeachingLoadRepositoryPG) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM teaching_loads WHERE id = $1`

	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete teaching load: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to read delete result: %w", err)
	}
	if affected == 0 {
		return entities.ErrTeachingLoadNotFound
	}
	return nil
}

// GetByID returns one hydrated line or ErrTeachingLoadNotFound.
func (r *TeachingLoadRepositoryPG) GetByID(ctx context.Context, id int64) (*entities.TeachingLoad, error) {
	load, err := scanTeachingLoad(r.db.QueryRowContext(ctx, teachingLoadSelect+" WHERE tl.id = $1", id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, entities.ErrTeachingLoadNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get teaching load: %w", err)
	}
	return load, nil
}

// List returns hydrated load lines matching the filter, ordered by group then discipline.
func (r *TeachingLoadRepositoryPG) List(ctx context.Context, filter usecases.TeachingLoadFilter) ([]*entities.TeachingLoad, error) {
	var conds []string
	var args []any
	i := 1
	add := func(col string, val *int64) {
		if val != nil {
			conds = append(conds, fmt.Sprintf("tl.%s = $%s", col, strconv.Itoa(i)))
			args = append(args, *val)
			i++
		}
	}
	add("semester_id", filter.SemesterID)
	add("group_id", filter.GroupID)
	add("teacher_id", filter.TeacherID)

	query := teachingLoadSelect
	if len(conds) > 0 {
		query += " WHERE " + strings.Join(conds, " AND ")
	}
	query += " ORDER BY tl.group_id, tl.discipline_id"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list teaching loads: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var loads []*entities.TeachingLoad
	for rows.Next() {
		load, err := scanTeachingLoad(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan teaching load: %w", err)
		}
		loads = append(loads, load)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate teaching loads: %w", err)
	}
	return loads, nil
}

// scanner abstracts *sql.Row and *sql.Rows for the shared hydration scan.
type scanner interface {
	Scan(dest ...any) error
}

// scanTeachingLoad maps one hydrated row into a TeachingLoad with associations.
func scanTeachingLoad(row scanner) (*entities.TeachingLoad, error) {
	load := &entities.TeachingLoad{}
	group := &entities.StudentGroup{}
	disc := &entities.Discipline{}
	ltype := &entities.LessonType{}
	teacher := &entities.TeacherInfo{}
	var weekType string

	err := row.Scan(
		&load.ID, &load.SemesterID, &load.GroupID, &load.DisciplineID, &load.TeacherID, &load.LessonTypeID,
		&load.PairsPerWeek, &weekType, &load.CreatedAt, &load.UpdatedAt,
		&group.ID, &group.Name, &disc.ID, &disc.Name, &ltype.ID, &ltype.Name, &ltype.ShortName,
		&teacher.ID, &teacher.Name, &teacher.Email,
	)
	if err != nil {
		return nil, err
	}
	load.WeekType = domain.WeekType(weekType)
	load.Group = group
	load.Discipline = disc
	load.LessonType = ltype
	load.Teacher = teacher
	return load, nil
}
