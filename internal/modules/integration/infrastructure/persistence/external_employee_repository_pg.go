package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/domain/repositories"
)

// ExternalEmployeeRepositoryPg implements ExternalEmployeeRepository using PostgreSQL
type ExternalEmployeeRepositoryPg struct {
	db *sql.DB
}

// NewExternalEmployeeRepositoryPg creates a new PostgreSQL external employee repository
func NewExternalEmployeeRepositoryPg(db *sql.DB) repositories.ExternalEmployeeRepository {
	return &ExternalEmployeeRepositoryPg{db: db}
}

func scanEmployeeRow(row *sql.Row) (*entities.ExternalEmployee, error) {
	var emp entities.ExternalEmployee
	var middleName, email, phone, position, department sql.NullString
	var employmentDate, dismissalDate sql.NullTime
	var localUserID sql.NullInt64
	var externalDataHash sql.NullString
	var rawData []byte

	err := row.Scan(
		&emp.ID, &emp.ExternalID, &emp.Code, &emp.FirstName, &emp.LastName,
		&middleName, &email, &phone, &position, &department,
		&employmentDate, &dismissalDate, &emp.IsActive, &localUserID,
		&emp.LastSyncAt, &externalDataHash, &rawData, &emp.CreatedAt, &emp.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if middleName.Valid {
		emp.MiddleName = middleName.String
	}
	if email.Valid {
		emp.Email = email.String
	}
	if phone.Valid {
		emp.Phone = phone.String
	}
	if position.Valid {
		emp.Position = position.String
	}
	if department.Valid {
		emp.Department = department.String
	}
	if employmentDate.Valid {
		emp.EmploymentDate = &employmentDate.Time
	}
	if dismissalDate.Valid {
		emp.DismissalDate = &dismissalDate.Time
	}
	if localUserID.Valid {
		emp.LocalUserID = &localUserID.Int64
	}
	if externalDataHash.Valid {
		emp.ExternalDataHash = externalDataHash.String
	}
	if len(rawData) > 0 {
		emp.RawData = string(rawData)
	}

	return &emp, nil
}

func scanEmployeeRows(rows *sql.Rows) ([]*entities.ExternalEmployee, error) {
	var employees []*entities.ExternalEmployee

	for rows.Next() {
		var emp entities.ExternalEmployee
		var middleName, email, phone, position, department sql.NullString
		var employmentDate, dismissalDate sql.NullTime
		var localUserID sql.NullInt64
		var externalDataHash sql.NullString
		var rawData []byte

		err := rows.Scan(
			&emp.ID, &emp.ExternalID, &emp.Code, &emp.FirstName, &emp.LastName,
			&middleName, &email, &phone, &position, &department,
			&employmentDate, &dismissalDate, &emp.IsActive, &localUserID,
			&emp.LastSyncAt, &externalDataHash, &rawData, &emp.CreatedAt, &emp.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if middleName.Valid {
			emp.MiddleName = middleName.String
		}
		if email.Valid {
			emp.Email = email.String
		}
		if phone.Valid {
			emp.Phone = phone.String
		}
		if position.Valid {
			emp.Position = position.String
		}
		if department.Valid {
			emp.Department = department.String
		}
		if employmentDate.Valid {
			emp.EmploymentDate = &employmentDate.Time
		}
		if dismissalDate.Valid {
			emp.DismissalDate = &dismissalDate.Time
		}
		if localUserID.Valid {
			emp.LocalUserID = &localUserID.Int64
		}
		if externalDataHash.Valid {
			emp.ExternalDataHash = externalDataHash.String
		}
		if len(rawData) > 0 {
			emp.RawData = string(rawData)
		}

		employees = append(employees, &emp)
	}

	return employees, rows.Err()
}

const employeeSelectFields = `id, external_id, code, first_name, last_name, middle_name,
	email, phone, position, department, employment_date, dismissal_date,
	is_active, local_user_id, last_sync_at, external_data_hash, raw_data, created_at, updated_at`

// Create creates a new external employee record
func (r *ExternalEmployeeRepositoryPg) Create(ctx context.Context, emp *entities.ExternalEmployee) error {
	var rawData []byte
	if emp.RawData != "" {
		rawData = []byte(emp.RawData)
	}

	query := `
		INSERT INTO external_employees (
			external_id, code, first_name, last_name, middle_name,
			email, phone, position, department, employment_date,
			dismissal_date, is_active, local_user_id, last_sync_at,
			external_data_hash, raw_data, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18
		) RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		emp.ExternalID, emp.Code, emp.FirstName, emp.LastName,
		nullString(emp.MiddleName), nullString(emp.Email), nullString(emp.Phone),
		nullString(emp.Position), nullString(emp.Department),
		nullTime(emp.EmploymentDate), nullTime(emp.DismissalDate),
		emp.IsActive, nullInt64(emp.LocalUserID), emp.LastSyncAt,
		nullString(emp.ExternalDataHash), rawData, emp.CreatedAt, emp.UpdatedAt,
	).Scan(&emp.ID)

	if err != nil {
		return fmt.Errorf("failed to create external employee: %w", err)
	}

	return nil
}

// Update updates an existing external employee record
func (r *ExternalEmployeeRepositoryPg) Update(ctx context.Context, emp *entities.ExternalEmployee) error {
	var rawData []byte
	if emp.RawData != "" {
		rawData = []byte(emp.RawData)
	}

	query := `
		UPDATE external_employees SET
			code = $1, first_name = $2, last_name = $3, middle_name = $4,
			email = $5, phone = $6, position = $7, department = $8,
			employment_date = $9, dismissal_date = $10, is_active = $11,
			local_user_id = $12, last_sync_at = $13, external_data_hash = $14,
			raw_data = $15, updated_at = $16
		WHERE id = $17`

	result, err := r.db.ExecContext(ctx, query,
		emp.Code, emp.FirstName, emp.LastName, nullString(emp.MiddleName),
		nullString(emp.Email), nullString(emp.Phone), nullString(emp.Position),
		nullString(emp.Department), nullTime(emp.EmploymentDate),
		nullTime(emp.DismissalDate), emp.IsActive, nullInt64(emp.LocalUserID),
		emp.LastSyncAt, nullString(emp.ExternalDataHash), rawData, time.Now(), emp.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update external employee: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// Upsert creates or updates an external employee by external ID
func (r *ExternalEmployeeRepositoryPg) Upsert(ctx context.Context, emp *entities.ExternalEmployee) error {
	var rawData []byte
	if emp.RawData != "" {
		rawData = []byte(emp.RawData)
	}

	query := `
		INSERT INTO external_employees (
			external_id, code, first_name, last_name, middle_name,
			email, phone, position, department, employment_date,
			dismissal_date, is_active, last_sync_at, external_data_hash,
			raw_data, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17
		)
		ON CONFLICT (external_id) DO UPDATE SET
			code = EXCLUDED.code,
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			middle_name = EXCLUDED.middle_name,
			email = EXCLUDED.email,
			phone = EXCLUDED.phone,
			position = EXCLUDED.position,
			department = EXCLUDED.department,
			employment_date = EXCLUDED.employment_date,
			dismissal_date = EXCLUDED.dismissal_date,
			is_active = EXCLUDED.is_active,
			last_sync_at = EXCLUDED.last_sync_at,
			external_data_hash = EXCLUDED.external_data_hash,
			raw_data = EXCLUDED.raw_data,
			updated_at = NOW()
		RETURNING id`

	now := time.Now()
	err := r.db.QueryRowContext(ctx, query,
		emp.ExternalID, emp.Code, emp.FirstName, emp.LastName,
		nullString(emp.MiddleName), nullString(emp.Email), nullString(emp.Phone),
		nullString(emp.Position), nullString(emp.Department),
		nullTime(emp.EmploymentDate), nullTime(emp.DismissalDate),
		emp.IsActive, now, nullString(emp.ExternalDataHash), rawData, now, now,
	).Scan(&emp.ID)

	if err != nil {
		return fmt.Errorf("failed to upsert external employee: %w", err)
	}

	return nil
}

// GetByID retrieves an external employee by ID
func (r *ExternalEmployeeRepositoryPg) GetByID(ctx context.Context, id int64) (*entities.ExternalEmployee, error) {
	query := fmt.Sprintf(`SELECT %s FROM external_employees WHERE id = $1`, employeeSelectFields)
	emp, err := scanEmployeeRow(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get external employee: %w", err)
	}

	return emp, nil
}

// GetByExternalID retrieves an external employee by 1C external ID
func (r *ExternalEmployeeRepositoryPg) GetByExternalID(ctx context.Context, externalID string) (*entities.ExternalEmployee, error) {
	query := fmt.Sprintf(`SELECT %s FROM external_employees WHERE external_id = $1`, employeeSelectFields)
	emp, err := scanEmployeeRow(r.db.QueryRowContext(ctx, query, externalID))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get external employee by external ID: %w", err)
	}

	return emp, nil
}

// GetByCode retrieves an external employee by 1C code
func (r *ExternalEmployeeRepositoryPg) GetByCode(ctx context.Context, code string) (*entities.ExternalEmployee, error) {
	query := fmt.Sprintf(`SELECT %s FROM external_employees WHERE code = $1`, employeeSelectFields)
	emp, err := scanEmployeeRow(r.db.QueryRowContext(ctx, query, code))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get external employee by code: %w", err)
	}

	return emp, nil
}

// GetByLocalUserID retrieves an external employee by linked local user ID
func (r *ExternalEmployeeRepositoryPg) GetByLocalUserID(ctx context.Context, localUserID int64) (*entities.ExternalEmployee, error) {
	query := fmt.Sprintf(`SELECT %s FROM external_employees WHERE local_user_id = $1`, employeeSelectFields)
	emp, err := scanEmployeeRow(r.db.QueryRowContext(ctx, query, localUserID))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get external employee by local user ID: %w", err)
	}

	return emp, nil
}

// List retrieves external employees with optional filtering
func (r *ExternalEmployeeRepositoryPg) List(ctx context.Context, filter entities.ExternalEmployeeFilter) ([]*entities.ExternalEmployee, int64, error) {
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
	if filter.Department != "" {
		conditions = append(conditions, fmt.Sprintf("department = $%d", argNum))
		args = append(args, filter.Department)
		argNum++
	}
	if filter.Position != "" {
		conditions = append(conditions, fmt.Sprintf("position = $%d", argNum))
		args = append(args, filter.Position)
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
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM external_employees %s", whereClause)
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count external employees: %w", err)
	}

	// Get items
	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}

	query := fmt.Sprintf(`
		SELECT %s FROM external_employees %s
		ORDER BY last_name, first_name
		LIMIT $%d OFFSET $%d`,
		employeeSelectFields, whereClause, argNum, argNum+1)

	args = append(args, limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list external employees: %w", err)
	}
	defer rows.Close()

	employees, err := scanEmployeeRows(rows)
	if err != nil {
		return nil, 0, err
	}

	return employees, total, nil
}

// GetUnlinked retrieves external employees not linked to local users
func (r *ExternalEmployeeRepositoryPg) GetUnlinked(ctx context.Context, limit, offset int) ([]*entities.ExternalEmployee, int64, error) {
	isLinked := false
	return r.List(ctx, entities.ExternalEmployeeFilter{
		IsLinked: &isLinked,
		Limit:    limit,
		Offset:   offset,
	})
}

// LinkToLocalUser links an external employee to a local user
func (r *ExternalEmployeeRepositoryPg) LinkToLocalUser(ctx context.Context, id int64, localUserID int64) error {
	query := `UPDATE external_employees SET local_user_id = $1, updated_at = NOW() WHERE id = $2`
	result, err := r.db.ExecContext(ctx, query, localUserID, id)
	if err != nil {
		return fmt.Errorf("failed to link external employee: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// Unlink removes the link between external employee and local user
func (r *ExternalEmployeeRepositoryPg) Unlink(ctx context.Context, id int64) error {
	query := `UPDATE external_employees SET local_user_id = NULL, updated_at = NOW() WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to unlink external employee: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// Delete deletes an external employee record
func (r *ExternalEmployeeRepositoryPg) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM external_employees WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete external employee: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// GetAllExternalIDs retrieves all external IDs for change detection
func (r *ExternalEmployeeRepositoryPg) GetAllExternalIDs(ctx context.Context) ([]string, error) {
	query := `SELECT external_id FROM external_employees`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get external IDs: %w", err)
	}
	defer rows.Close()

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

// BulkUpsert creates or updates multiple external employees
func (r *ExternalEmployeeRepositoryPg) BulkUpsert(ctx context.Context, employees []*entities.ExternalEmployee) error {
	if len(employees) == 0 {
		return nil
	}

	// Use transaction for bulk operation
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, emp := range employees {
		var rawData []byte
		if emp.RawData != "" {
			rawData = []byte(emp.RawData)
		}

		query := `
			INSERT INTO external_employees (
				external_id, code, first_name, last_name, middle_name,
				email, phone, position, department, employment_date,
				dismissal_date, is_active, last_sync_at, external_data_hash,
				raw_data, created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
				$11, $12, $13, $14, $15, NOW(), NOW()
			)
			ON CONFLICT (external_id) DO UPDATE SET
				code = EXCLUDED.code,
				first_name = EXCLUDED.first_name,
				last_name = EXCLUDED.last_name,
				middle_name = EXCLUDED.middle_name,
				email = EXCLUDED.email,
				phone = EXCLUDED.phone,
				position = EXCLUDED.position,
				department = EXCLUDED.department,
				employment_date = EXCLUDED.employment_date,
				dismissal_date = EXCLUDED.dismissal_date,
				is_active = EXCLUDED.is_active,
				last_sync_at = EXCLUDED.last_sync_at,
				external_data_hash = EXCLUDED.external_data_hash,
				raw_data = EXCLUDED.raw_data,
				updated_at = NOW()`

		_, err := tx.ExecContext(ctx, query,
			emp.ExternalID, emp.Code, emp.FirstName, emp.LastName,
			nullString(emp.MiddleName), nullString(emp.Email), nullString(emp.Phone),
			nullString(emp.Position), nullString(emp.Department),
			nullTime(emp.EmploymentDate), nullTime(emp.DismissalDate),
			emp.IsActive, time.Now(), nullString(emp.ExternalDataHash), rawData,
		)
		if err != nil {
			return fmt.Errorf("failed to upsert employee %s: %w", emp.ExternalID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// MarkInactiveExcept marks all employees as inactive except those with given external IDs
func (r *ExternalEmployeeRepositoryPg) MarkInactiveExcept(ctx context.Context, activeExternalIDs []string) error {
	if len(activeExternalIDs) == 0 {
		// Mark all as inactive
		_, err := r.db.ExecContext(ctx,
			"UPDATE external_employees SET is_active = false, updated_at = NOW()")
		return err
	}

	query := `
		UPDATE external_employees
		SET is_active = false, updated_at = NOW()
		WHERE external_id != ALL($1)`

	_, err := r.db.ExecContext(ctx, query, pq.Array(activeExternalIDs))
	if err != nil {
		return fmt.Errorf("failed to mark inactive employees: %w", err)
	}

	return nil
}

// Helper functions
func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func nullInt64(i *int64) sql.NullInt64 {
	if i == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *i, Valid: true}
}

func nullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}
