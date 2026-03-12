package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/domain/repositories"
)

// ExternalStudentRepositoryPg implements ExternalStudentRepository using PostgreSQL
type ExternalStudentRepositoryPg struct {
	db *sql.DB
}

// NewExternalStudentRepositoryPg creates a new PostgreSQL external student repository
func NewExternalStudentRepositoryPg(db *sql.DB) repositories.ExternalStudentRepository {
	return &ExternalStudentRepositoryPg{db: db}
}

func scanStudentRow(row *sql.Row) (*entities.ExternalStudent, error) {
	var student entities.ExternalStudent
	var middleName, email, phone, groupName, faculty, specialty sql.NullString
	var course sql.NullInt32
	var studyForm sql.NullString
	var enrollmentDate, expulsionDate, graduationDate sql.NullTime
	var localUserID sql.NullInt64
	var externalDataHash sql.NullString
	var rawData []byte

	err := row.Scan(
		&student.ID, &student.ExternalID, &student.Code, &student.FirstName, &student.LastName,
		&middleName, &email, &phone, &groupName, &faculty, &specialty, &course,
		&studyForm, &enrollmentDate, &expulsionDate, &graduationDate,
		&student.Status, &student.IsActive, &localUserID, &student.LastSyncAt,
		&externalDataHash, &rawData, &student.CreatedAt, &student.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if middleName.Valid {
		student.MiddleName = middleName.String
	}
	if email.Valid {
		student.Email = email.String
	}
	if phone.Valid {
		student.Phone = phone.String
	}
	if groupName.Valid {
		student.GroupName = groupName.String
	}
	if faculty.Valid {
		student.Faculty = faculty.String
	}
	if specialty.Valid {
		student.Specialty = specialty.String
	}
	if course.Valid {
		student.Course = int(course.Int32)
	}
	if studyForm.Valid {
		student.StudyForm = studyForm.String
	}
	if enrollmentDate.Valid {
		student.EnrollmentDate = &enrollmentDate.Time
	}
	if expulsionDate.Valid {
		student.ExpulsionDate = &expulsionDate.Time
	}
	if graduationDate.Valid {
		student.GraduationDate = &graduationDate.Time
	}
	if localUserID.Valid {
		student.LocalUserID = &localUserID.Int64
	}
	if externalDataHash.Valid {
		student.ExternalDataHash = externalDataHash.String
	}
	if len(rawData) > 0 {
		student.RawData = string(rawData)
	}

	return &student, nil
}

func scanStudentRows(rows *sql.Rows) ([]*entities.ExternalStudent, error) {
	var students []*entities.ExternalStudent

	for rows.Next() {
		var student entities.ExternalStudent
		var middleName, email, phone, groupName, faculty, specialty sql.NullString
		var course sql.NullInt32
		var studyForm sql.NullString
		var enrollmentDate, expulsionDate, graduationDate sql.NullTime
		var localUserID sql.NullInt64
		var externalDataHash sql.NullString
		var rawData []byte

		err := rows.Scan(
			&student.ID, &student.ExternalID, &student.Code, &student.FirstName, &student.LastName,
			&middleName, &email, &phone, &groupName, &faculty, &specialty, &course,
			&studyForm, &enrollmentDate, &expulsionDate, &graduationDate,
			&student.Status, &student.IsActive, &localUserID, &student.LastSyncAt,
			&externalDataHash, &rawData, &student.CreatedAt, &student.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if middleName.Valid {
			student.MiddleName = middleName.String
		}
		if email.Valid {
			student.Email = email.String
		}
		if phone.Valid {
			student.Phone = phone.String
		}
		if groupName.Valid {
			student.GroupName = groupName.String
		}
		if faculty.Valid {
			student.Faculty = faculty.String
		}
		if specialty.Valid {
			student.Specialty = specialty.String
		}
		if course.Valid {
			student.Course = int(course.Int32)
		}
		if studyForm.Valid {
			student.StudyForm = studyForm.String
		}
		if enrollmentDate.Valid {
			student.EnrollmentDate = &enrollmentDate.Time
		}
		if expulsionDate.Valid {
			student.ExpulsionDate = &expulsionDate.Time
		}
		if graduationDate.Valid {
			student.GraduationDate = &graduationDate.Time
		}
		if localUserID.Valid {
			student.LocalUserID = &localUserID.Int64
		}
		if externalDataHash.Valid {
			student.ExternalDataHash = externalDataHash.String
		}
		if len(rawData) > 0 {
			student.RawData = string(rawData)
		}

		students = append(students, &student)
	}

	return students, rows.Err()
}

const studentSelectFields = `id, external_id, code, first_name, last_name, middle_name,
	email, phone, group_name, faculty, specialty, course, study_form,
	enrollment_date, expulsion_date, graduation_date, status, is_active,
	local_user_id, last_sync_at, external_data_hash, raw_data, created_at, updated_at`

// Create creates a new external student record
func (r *ExternalStudentRepositoryPg) Create(ctx context.Context, student *entities.ExternalStudent) error {
	var rawData []byte
	if student.RawData != "" {
		rawData = []byte(student.RawData)
	}

	query := `
		INSERT INTO external_students (
			external_id, code, first_name, last_name, middle_name,
			email, phone, group_name, faculty, specialty, course,
			study_form, enrollment_date, expulsion_date, graduation_date,
			status, is_active, local_user_id, last_sync_at,
			external_data_hash, raw_data, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11,
			$12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23
		) RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		student.ExternalID, student.Code, student.FirstName, student.LastName,
		nullString(student.MiddleName), nullString(student.Email), nullString(student.Phone),
		nullString(student.GroupName), nullString(student.Faculty), nullString(student.Specialty),
		nullInt32(student.Course), nullString(student.StudyForm),
		nullTime(student.EnrollmentDate), nullTime(student.ExpulsionDate), nullTime(student.GraduationDate),
		student.Status, student.IsActive, nullInt64(student.LocalUserID), student.LastSyncAt,
		nullString(student.ExternalDataHash), rawData, student.CreatedAt, student.UpdatedAt,
	).Scan(&student.ID)

	if err != nil {
		return fmt.Errorf("failed to create external student: %w", err)
	}

	return nil
}

// Update updates an existing external student record
func (r *ExternalStudentRepositoryPg) Update(ctx context.Context, student *entities.ExternalStudent) error {
	var rawData []byte
	if student.RawData != "" {
		rawData = []byte(student.RawData)
	}

	query := `
		UPDATE external_students SET
			code = $1, first_name = $2, last_name = $3, middle_name = $4,
			email = $5, phone = $6, group_name = $7, faculty = $8,
			specialty = $9, course = $10, study_form = $11,
			enrollment_date = $12, expulsion_date = $13, graduation_date = $14,
			status = $15, is_active = $16, local_user_id = $17, last_sync_at = $18,
			external_data_hash = $19, raw_data = $20, updated_at = $21
		WHERE id = $22`

	result, err := r.db.ExecContext(ctx, query,
		student.Code, student.FirstName, student.LastName, nullString(student.MiddleName),
		nullString(student.Email), nullString(student.Phone), nullString(student.GroupName),
		nullString(student.Faculty), nullString(student.Specialty), nullInt32(student.Course),
		nullString(student.StudyForm), nullTime(student.EnrollmentDate),
		nullTime(student.ExpulsionDate), nullTime(student.GraduationDate),
		student.Status, student.IsActive, nullInt64(student.LocalUserID), student.LastSyncAt,
		nullString(student.ExternalDataHash), rawData, time.Now(), student.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update external student: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// Upsert creates or updates an external student by external ID
func (r *ExternalStudentRepositoryPg) Upsert(ctx context.Context, student *entities.ExternalStudent) error {
	var rawData []byte
	if student.RawData != "" {
		rawData = []byte(student.RawData)
	}

	query := `
		INSERT INTO external_students (
			external_id, code, first_name, last_name, middle_name,
			email, phone, group_name, faculty, specialty, course,
			study_form, enrollment_date, expulsion_date, graduation_date,
			status, is_active, last_sync_at, external_data_hash,
			raw_data, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11,
			$12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22
		)
		ON CONFLICT (external_id) DO UPDATE SET
			code = EXCLUDED.code,
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			middle_name = EXCLUDED.middle_name,
			email = EXCLUDED.email,
			phone = EXCLUDED.phone,
			group_name = EXCLUDED.group_name,
			faculty = EXCLUDED.faculty,
			specialty = EXCLUDED.specialty,
			course = EXCLUDED.course,
			study_form = EXCLUDED.study_form,
			enrollment_date = EXCLUDED.enrollment_date,
			expulsion_date = EXCLUDED.expulsion_date,
			graduation_date = EXCLUDED.graduation_date,
			status = EXCLUDED.status,
			is_active = EXCLUDED.is_active,
			last_sync_at = EXCLUDED.last_sync_at,
			external_data_hash = EXCLUDED.external_data_hash,
			raw_data = EXCLUDED.raw_data,
			updated_at = NOW()
		RETURNING id`

	now := time.Now()
	err := r.db.QueryRowContext(ctx, query,
		student.ExternalID, student.Code, student.FirstName, student.LastName,
		nullString(student.MiddleName), nullString(student.Email), nullString(student.Phone),
		nullString(student.GroupName), nullString(student.Faculty), nullString(student.Specialty),
		nullInt32(student.Course), nullString(student.StudyForm),
		nullTime(student.EnrollmentDate), nullTime(student.ExpulsionDate), nullTime(student.GraduationDate),
		student.Status, student.IsActive, now, nullString(student.ExternalDataHash), rawData, now, now,
	).Scan(&student.ID)

	if err != nil {
		return fmt.Errorf("failed to upsert external student: %w", err)
	}

	return nil
}

// GetByID retrieves an external student by ID
func (r *ExternalStudentRepositoryPg) GetByID(ctx context.Context, id int64) (*entities.ExternalStudent, error) {
	query := fmt.Sprintf(`SELECT %s FROM external_students WHERE id = $1`, studentSelectFields)
	student, err := scanStudentRow(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get external student: %w", err)
	}

	return student, nil
}

// GetByExternalID retrieves an external student by 1C external ID
func (r *ExternalStudentRepositoryPg) GetByExternalID(ctx context.Context, externalID string) (*entities.ExternalStudent, error) {
	query := fmt.Sprintf(`SELECT %s FROM external_students WHERE external_id = $1`, studentSelectFields)
	student, err := scanStudentRow(r.db.QueryRowContext(ctx, query, externalID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get external student by external ID: %w", err)
	}

	return student, nil
}

// GetByCode retrieves an external student by 1C code
func (r *ExternalStudentRepositoryPg) GetByCode(ctx context.Context, code string) (*entities.ExternalStudent, error) {
	query := fmt.Sprintf(`SELECT %s FROM external_students WHERE code = $1`, studentSelectFields)
	student, err := scanStudentRow(r.db.QueryRowContext(ctx, query, code))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get external student by code: %w", err)
	}

	return student, nil
}

// GetByLocalUserID retrieves an external student by linked local user ID
func (r *ExternalStudentRepositoryPg) GetByLocalUserID(ctx context.Context, localUserID int64) (*entities.ExternalStudent, error) {
	query := fmt.Sprintf(`SELECT %s FROM external_students WHERE local_user_id = $1`, studentSelectFields)
	student, err := scanStudentRow(r.db.QueryRowContext(ctx, query, localUserID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get external student by local user ID: %w", err)
	}

	return student, nil
}

// List retrieves external students with optional filtering
func (r *ExternalStudentRepositoryPg) List(ctx context.Context, filter entities.ExternalStudentFilter) ([]*entities.ExternalStudent, int64, error) {
	var conditions []string
	var args []any
	argNum := 1

	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf(
			"to_tsvector('russian', coalesce(first_name, '') || ' ' || coalesce(last_name, '') || ' ' || coalesce(middle_name, '')) @@ plainto_tsquery('russian', $%d)",
			argNum))
		args = append(args, filter.Search)
		argNum++
	}
	if filter.GroupName != "" {
		conditions = append(conditions, fmt.Sprintf("group_name = $%d", argNum))
		args = append(args, filter.GroupName)
		argNum++
	}
	if filter.Faculty != "" {
		conditions = append(conditions, fmt.Sprintf("faculty = $%d", argNum))
		args = append(args, filter.Faculty)
		argNum++
	}
	if filter.Course != nil {
		conditions = append(conditions, fmt.Sprintf("course = $%d", argNum))
		args = append(args, *filter.Course)
		argNum++
	}
	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argNum))
		args = append(args, filter.Status)
		argNum++
	}
	if filter.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argNum))
		args = append(args, *filter.IsActive)
		argNum++
	}
	if filter.IsLinked != nil {
		if *filter.IsLinked {
			conditions = append(conditions, "local_user_id IS NOT NULL")
		} else {
			conditions = append(conditions, "local_user_id IS NULL")
		}
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM external_students %s", whereClause)
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count external students: %w", err)
	}

	// Get items
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}

	query := fmt.Sprintf(`
		SELECT %s FROM external_students %s
		ORDER BY last_name, first_name
		LIMIT $%d OFFSET $%d`,
		studentSelectFields, whereClause, argNum, argNum+1)

	args = append(args, limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list external students: %w", err)
	}
	defer func() { _ = rows.Close() }()

	students, err := scanStudentRows(rows)
	if err != nil {
		return nil, 0, err
	}

	return students, total, nil
}

// GetUnlinked retrieves external students not linked to local users
func (r *ExternalStudentRepositoryPg) GetUnlinked(ctx context.Context, limit, offset int) ([]*entities.ExternalStudent, int64, error) {
	isLinked := false
	return r.List(ctx, entities.ExternalStudentFilter{
		IsLinked: &isLinked,
		Limit:    limit,
		Offset:   offset,
	})
}

// GetByGroup retrieves external students by group name
func (r *ExternalStudentRepositoryPg) GetByGroup(ctx context.Context, groupName string) ([]*entities.ExternalStudent, error) {
	query := fmt.Sprintf(`SELECT %s FROM external_students WHERE group_name = $1 ORDER BY last_name, first_name`, studentSelectFields)
	rows, err := r.db.QueryContext(ctx, query, groupName)
	if err != nil {
		return nil, fmt.Errorf("failed to get students by group: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return scanStudentRows(rows)
}

// GetByFaculty retrieves external students by faculty
func (r *ExternalStudentRepositoryPg) GetByFaculty(ctx context.Context, faculty string) ([]*entities.ExternalStudent, error) {
	query := fmt.Sprintf(`SELECT %s FROM external_students WHERE faculty = $1 ORDER BY last_name, first_name`, studentSelectFields)
	rows, err := r.db.QueryContext(ctx, query, faculty)
	if err != nil {
		return nil, fmt.Errorf("failed to get students by faculty: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return scanStudentRows(rows)
}

// LinkToLocalUser links an external student to a local user
func (r *ExternalStudentRepositoryPg) LinkToLocalUser(ctx context.Context, id int64, localUserID int64) error {
	query := `UPDATE external_students SET local_user_id = $1, updated_at = NOW() WHERE id = $2`
	result, err := r.db.ExecContext(ctx, query, localUserID, id)
	if err != nil {
		return fmt.Errorf("failed to link external student: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// Unlink removes the link between external student and local user
func (r *ExternalStudentRepositoryPg) Unlink(ctx context.Context, id int64) error {
	query := `UPDATE external_students SET local_user_id = NULL, updated_at = NOW() WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to unlink external student: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// Delete deletes an external student record
func (r *ExternalStudentRepositoryPg) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM external_students WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete external student: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// GetAllExternalIDs retrieves all external IDs for change detection
func (r *ExternalStudentRepositoryPg) GetAllExternalIDs(ctx context.Context) ([]string, error) {
	query := `SELECT external_id FROM external_students`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get external IDs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// BulkUpsert creates or updates multiple external students
func (r *ExternalStudentRepositoryPg) BulkUpsert(ctx context.Context, students []*entities.ExternalStudent) error {
	if len(students) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	for _, student := range students {
		var rawData []byte
		if student.RawData != "" {
			rawData = []byte(student.RawData)
		}

		query := `
			INSERT INTO external_students (
				external_id, code, first_name, last_name, middle_name,
				email, phone, group_name, faculty, specialty, course,
				study_form, enrollment_date, expulsion_date, graduation_date,
				status, is_active, last_sync_at, external_data_hash,
				raw_data, created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11,
				$12, $13, $14, $15, $16, $17, NOW(), $18, $19, NOW(), NOW()
			)
			ON CONFLICT (external_id) DO UPDATE SET
				code = EXCLUDED.code,
				first_name = EXCLUDED.first_name,
				last_name = EXCLUDED.last_name,
				middle_name = EXCLUDED.middle_name,
				email = EXCLUDED.email,
				phone = EXCLUDED.phone,
				group_name = EXCLUDED.group_name,
				faculty = EXCLUDED.faculty,
				specialty = EXCLUDED.specialty,
				course = EXCLUDED.course,
				study_form = EXCLUDED.study_form,
				enrollment_date = EXCLUDED.enrollment_date,
				expulsion_date = EXCLUDED.expulsion_date,
				graduation_date = EXCLUDED.graduation_date,
				status = EXCLUDED.status,
				is_active = EXCLUDED.is_active,
				last_sync_at = NOW(),
				external_data_hash = EXCLUDED.external_data_hash,
				raw_data = EXCLUDED.raw_data,
				updated_at = NOW()`

		_, err := tx.ExecContext(ctx, query,
			student.ExternalID, student.Code, student.FirstName, student.LastName,
			nullString(student.MiddleName), nullString(student.Email), nullString(student.Phone),
			nullString(student.GroupName), nullString(student.Faculty), nullString(student.Specialty),
			nullInt32(student.Course), nullString(student.StudyForm),
			nullTime(student.EnrollmentDate), nullTime(student.ExpulsionDate), nullTime(student.GraduationDate),
			student.Status, student.IsActive, nullString(student.ExternalDataHash), rawData,
		)
		if err != nil {
			return fmt.Errorf("failed to upsert student %s: %w", student.ExternalID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// MarkInactiveExcept marks all students as inactive except those with given external IDs
func (r *ExternalStudentRepositoryPg) MarkInactiveExcept(ctx context.Context, activeExternalIDs []string) error {
	if len(activeExternalIDs) == 0 {
		_, err := r.db.ExecContext(ctx,
			"UPDATE external_students SET is_active = false, updated_at = NOW()")
		return err
	}

	query := `
		UPDATE external_students
		SET is_active = false, updated_at = NOW()
		WHERE external_id != ALL($1)`

	_, err := r.db.ExecContext(ctx, query, pq.Array(activeExternalIDs))
	if err != nil {
		return fmt.Errorf("failed to mark inactive students: %w", err)
	}

	return nil
}

// GetGroups retrieves distinct group names
func (r *ExternalStudentRepositoryPg) GetGroups(ctx context.Context) ([]string, error) {
	query := `SELECT DISTINCT group_name FROM external_students WHERE group_name IS NOT NULL ORDER BY group_name`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get groups: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var groups []string
	for rows.Next() {
		var group string
		if err := rows.Scan(&group); err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}
	return groups, rows.Err()
}

// GetFaculties retrieves distinct faculty names
func (r *ExternalStudentRepositoryPg) GetFaculties(ctx context.Context) ([]string, error) {
	query := `SELECT DISTINCT faculty FROM external_students WHERE faculty IS NOT NULL ORDER BY faculty`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get faculties: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var faculties []string
	for rows.Next() {
		var faculty string
		if err := rows.Scan(&faculty); err != nil {
			return nil, err
		}
		faculties = append(faculties, faculty)
	}
	return faculties, rows.Err()
}

// Helper function
func nullInt32(i int) sql.NullInt32 {
	if i == 0 {
		return sql.NullInt32{}
	}
	return sql.NullInt32{Int32: int32(i), Valid: true} // #nosec G115 -- values are small bounded integers that fit in int32
}
