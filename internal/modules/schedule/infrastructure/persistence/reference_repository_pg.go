package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
)

// ReferenceRepositoryPG implements ReferenceRepository using PostgreSQL.
type ReferenceRepositoryPG struct {
	db *sql.DB
}

// NewReferenceRepositoryPG creates a new ReferenceRepositoryPG.
func NewReferenceRepositoryPG(db *sql.DB) *ReferenceRepositoryPG {
	return &ReferenceRepositoryPG{db: db}
}

// ListStudentGroups returns all student groups ordered by name.
func (r *ReferenceRepositoryPG) ListStudentGroups(ctx context.Context, limit, offset int) ([]*entities.StudentGroup, error) {
	query := `SELECT id, specialty_id, name, course, curator_id, capacity
		FROM student_groups ORDER BY name`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)
	}

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list student groups: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var groups []*entities.StudentGroup
	for rows.Next() {
		g := &entities.StudentGroup{}
		if err := rows.Scan(&g.ID, &g.SpecialtyID, &g.Name, &g.Course, &g.CuratorID, &g.Capacity); err != nil {
			return nil, fmt.Errorf("failed to scan student group: %w", err)
		}
		groups = append(groups, g)
	}

	return groups, rows.Err()
}

// ListDisciplines returns all disciplines ordered by name.
func (r *ReferenceRepositoryPG) ListDisciplines(ctx context.Context, limit, offset int) ([]*entities.Discipline, error) {
	query := `SELECT id, name, code, department_id, credits, hours_total,
		hours_lectures, hours_practice, hours_labs
		FROM disciplines ORDER BY name`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)
	}

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list disciplines: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var disciplines []*entities.Discipline
	for rows.Next() {
		d := &entities.Discipline{}
		if err := rows.Scan(&d.ID, &d.Name, &d.Code, &d.DepartmentID, &d.Credits,
			&d.HoursTotal, &d.HoursLectures, &d.HoursPractice, &d.HoursLabs); err != nil {
			return nil, fmt.Errorf("failed to scan discipline: %w", err)
		}
		disciplines = append(disciplines, d)
	}

	return disciplines, rows.Err()
}

// ListSemesters returns semesters, optionally filtered by active status.
func (r *ReferenceRepositoryPG) ListSemesters(ctx context.Context, activeOnly bool) ([]*entities.Semester, error) {
	query := `SELECT id, academic_year_id, name, number, start_date, end_date, is_active
		FROM semesters`

	if activeOnly {
		query += " WHERE is_active = true"
	}

	query += " ORDER BY start_date DESC"

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list semesters: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var semesters []*entities.Semester
	for rows.Next() {
		s := &entities.Semester{}
		if err := rows.Scan(&s.ID, &s.AcademicYearID, &s.Name, &s.Number, &s.StartDate, &s.EndDate, &s.IsActive); err != nil {
			return nil, fmt.Errorf("failed to scan semester: %w", err)
		}
		semesters = append(semesters, s)
	}

	return semesters, rows.Err()
}

// ListLessonTypes returns all lesson types ordered by ID.
func (r *ReferenceRepositoryPG) ListLessonTypes(ctx context.Context) ([]*entities.LessonType, error) {
	query := `SELECT id, name, short_name, color FROM lesson_types ORDER BY id`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list lesson types: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var types []*entities.LessonType
	for rows.Next() {
		lt := &entities.LessonType{}
		if err := rows.Scan(&lt.ID, &lt.Name, &lt.ShortName, &lt.Color); err != nil {
			return nil, fmt.Errorf("failed to scan lesson type: %w", err)
		}
		types = append(types, lt)
	}

	return types, rows.Err()
}

// GetActiveSemester returns the currently active semester.
func (r *ReferenceRepositoryPG) GetActiveSemester(ctx context.Context) (*entities.Semester, error) {
	query := `SELECT id, academic_year_id, name, number, start_date, end_date, is_active
		FROM semesters WHERE is_active = true LIMIT 1`

	s := &entities.Semester{}
	err := r.db.QueryRowContext(ctx, query).Scan(
		&s.ID, &s.AcademicYearID, &s.Name, &s.Number, &s.StartDate, &s.EndDate, &s.IsActive,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get active semester: %w", err)
	}

	return s, nil
}
