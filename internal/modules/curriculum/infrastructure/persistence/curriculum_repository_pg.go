// Package persistence provides PostgreSQL implementations of the
// curriculum module's repository ports.
package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/domain/repositories"
)

// CurriculumRepositoryPG is the SQL implementation of CurriculumRepository.
//
// Accepts DBTX (not *sql.DB) so the same struct can run against single-
// connection mode или a `*sql.Tx` inside BulkDisciplineItemsUnitOfWork
// (v0.128.3 ADR-10).
type CurriculumRepositoryPG struct {
	db DBTX
}

// NewCurriculumRepositoryPG constructs the repository. db can be
// `*sql.DB` (default DI) или `*sql.Tx` (bulk-edit transactional path).
func NewCurriculumRepositoryPG(db DBTX) *CurriculumRepositoryPG {
	return &CurriculumRepositoryPG{db: db}
}

const curriculumSelectColumns = `id, title, code, specialty, year, description, status, created_by, approved_by, approved_at, created_at, updated_at`

// pqUniqueViolation is the SQLSTATE code for a unique-constraint
// violation in PostgreSQL. The pg-error mapping is centralized in
// shared/infrastructure/database/errors.go for cross-module use, but
// this module mirrors the small inline pattern that existing modules
// (assignments, tasks) use to keep their bounded contexts free of a
// dependency on the shared mapper for a single sentinel.
const pqUniqueViolation = "23505"

// GetByID returns the curriculum with the given id.
func (r *CurriculumRepositoryPG) GetByID(ctx context.Context, id int64) (*entities.Curriculum, error) {
	query := `SELECT ` + curriculumSelectColumns + ` FROM curricula WHERE id = $1`

	var (
		idv         int64
		title       string
		code        string
		specialty   string
		year        int
		description sql.NullString
		statusStr   string
		createdBy   int64
		approvedBy  sql.NullInt64
		approvedAt  sql.NullTime
		createdAt   time.Time
		updatedAt   time.Time
	)
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&idv, &title, &code, &specialty, &year, &description,
		&statusStr, &createdBy, &approvedBy, &approvedAt,
		&createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repositories.ErrCurriculumNotFound
		}
		return nil, fmt.Errorf("curriculum: get by id: %w", err)
	}

	var ab *int64
	if approvedBy.Valid {
		v := approvedBy.Int64
		ab = &v
	}
	var aat *time.Time
	if approvedAt.Valid {
		t := approvedAt.Time
		aat = &t
	}
	return entities.ReconstituteCurriculum(
		idv, title, code, specialty, year, description.String,
		entities.CurriculumStatus(statusStr), createdBy,
		ab, aat, createdAt, updatedAt,
	), nil
}

// List returns a page of curricula matching the filter and the total
// count of matching rows.
//
// Filter encoding follows the assignments module convention:
//   - status: empty-string sentinel disables the predicate;
//   - year / created_by: NULL-pointer disables via sql.NullInt64;
//   - specialty: empty-string sentinel disables.
//
// COUNT and SELECT share the same WHERE clause so an empty page past
// the result-set tail still reports the correct dataset size for
// pagination.
func (r *CurriculumRepositoryPG) List(ctx context.Context, filter repositories.CurriculumListFilter) (repositories.CurriculumListResult, error) {
	statusFilter := ""
	if filter.Status != nil {
		statusFilter = string(*filter.Status)
	}
	var yearArg sql.NullInt64
	if filter.Year != nil {
		yearArg = sql.NullInt64{Int64: int64(*filter.Year), Valid: true}
	}
	var creatorArg sql.NullInt64
	if filter.CreatedBy != nil {
		creatorArg = sql.NullInt64{Int64: *filter.CreatedBy, Valid: true}
	}

	const filterClause = `WHERE ($1 = '' OR status = $1)
		AND ($2::bigint IS NULL OR year = $2::bigint)
		AND ($3 = '' OR specialty = $3)
		AND ($4::bigint IS NULL OR created_by = $4::bigint)`

	countQuery := `SELECT COUNT(*) FROM curricula ` + filterClause
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery,
		statusFilter, yearArg, filter.Specialty, creatorArg,
	).Scan(&total); err != nil {
		return repositories.CurriculumListResult{}, fmt.Errorf("curriculum: count: %w", err)
	}

	listQuery := `SELECT ` + curriculumSelectColumns + ` FROM curricula ` + filterClause + `
		ORDER BY year DESC, created_at DESC, id DESC
		LIMIT $5 OFFSET $6`

	rows, err := r.db.QueryContext(ctx, listQuery,
		statusFilter, yearArg, filter.Specialty, creatorArg,
		filter.Limit, filter.Offset)
	if err != nil {
		return repositories.CurriculumListResult{}, fmt.Errorf("curriculum: list: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var items []*entities.Curriculum
	for rows.Next() {
		var (
			id          int64
			title       string
			code        string
			specialty   string
			year        int
			description sql.NullString
			statusStr   string
			createdBy   int64
			approvedBy  sql.NullInt64
			approvedAt  sql.NullTime
			createdAt   time.Time
			updatedAt   time.Time
		)
		if err := rows.Scan(&id, &title, &code, &specialty, &year, &description,
			&statusStr, &createdBy, &approvedBy, &approvedAt,
			&createdAt, &updatedAt); err != nil {
			return repositories.CurriculumListResult{}, fmt.Errorf("curriculum: list scan: %w", err)
		}
		var ab *int64
		if approvedBy.Valid {
			v := approvedBy.Int64
			ab = &v
		}
		var aat *time.Time
		if approvedAt.Valid {
			t := approvedAt.Time
			aat = &t
		}
		items = append(items, entities.ReconstituteCurriculum(
			id, title, code, specialty, year, description.String,
			entities.CurriculumStatus(statusStr), createdBy,
			ab, aat, createdAt, updatedAt,
		))
	}
	if err := rows.Err(); err != nil {
		return repositories.CurriculumListResult{}, fmt.Errorf("curriculum: list iter: %w", err)
	}
	return repositories.CurriculumListResult{Items: items, Total: total}, nil
}

// Save inserts a new Curriculum row and writes the generated id back
// onto the entity. Maps PostgreSQL unique-constraint violations
// (SQLSTATE 23505 against curricula_code_key) to ErrCurriculumCodeExists
// so the use-case layer can produce a deterministic 409 mapping
// without parsing pq error structs itself.
func (r *CurriculumRepositoryPG) Save(ctx context.Context, c *entities.Curriculum) error {
	query := `
		INSERT INTO curricula (
			title, code, specialty, year, description,
			status, created_by, approved_by, approved_at,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id`

	var newID int64
	err := r.db.QueryRowContext(ctx, query,
		c.Title(), c.Code(), c.Specialty(), c.Year(), nullableDescription(c.Description()),
		string(c.Status()), c.CreatedBy(),
		nullableInt64Ptr(c.ApprovedBy()), nullableTimePtr(c.ApprovedAt()),
		c.CreatedAt(), c.UpdatedAt(),
	).Scan(&newID)
	if err != nil {
		if isUniqueViolation(err) {
			return repositories.ErrCurriculumCodeExists
		}
		return fmt.Errorf("curriculum: save: %w", err)
	}
	c.ID = newID
	return nil
}

// Update writes the (already-mutated) entity back to the row keyed
// by ID. Returns ErrCurriculumNotFound when no row is touched (the
// row vanished between load and write — likely a stale entity).
// Maps unique-constraint violations on a code rename to
// ErrCurriculumCodeExists.
func (r *CurriculumRepositoryPG) Update(ctx context.Context, c *entities.Curriculum) error {
	query := `
		UPDATE curricula SET
			title       = $1,
			code        = $2,
			specialty   = $3,
			year        = $4,
			description = $5,
			status      = $6,
			approved_by = $7,
			approved_at = $8,
			updated_at  = $9
		WHERE id = $10`

	res, err := r.db.ExecContext(ctx, query,
		c.Title(), c.Code(), c.Specialty(), c.Year(), nullableDescription(c.Description()),
		string(c.Status()),
		nullableInt64Ptr(c.ApprovedBy()), nullableTimePtr(c.ApprovedAt()),
		c.UpdatedAt(), c.ID,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return repositories.ErrCurriculumCodeExists
		}
		return fmt.Errorf("curriculum: update: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("curriculum: update: rows affected: %w", err)
	}
	if rows == 0 {
		return repositories.ErrCurriculumNotFound
	}
	return nil
}

// nullableDescription maps an empty Go string to a SQL NULL so the
// description column stays NULL when absent (the migration leaves
// description nullable; storing ” would create a needless distinction
// from the JSON-side optional value).
func nullableDescription(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func nullableInt64Ptr(p *int64) sql.NullInt64 {
	if p == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *p, Valid: true}
}

func nullableTimePtr(p *time.Time) sql.NullTime {
	if p == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *p, Valid: true}
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == pqUniqueViolation
	}
	return false
}
