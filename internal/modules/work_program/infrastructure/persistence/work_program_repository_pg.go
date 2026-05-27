package persistence

import (
	"context"
	"errors"
	"fmt"

	"github.com/lib/pq"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/repositories"
)

// Compile-time assertion that the PG impl satisfies the wide port
// declared in application/usecases (DIP). Catches signature drift at
// the impl's compile site rather than only at DI wiring.
var _ usecases.WorkProgramRepository = (*WorkProgramRepositoryPG)(nil)

// WorkProgramRepositoryPG is the SQL implementation of
// WorkProgramRepository. Accepts DBTX (not *sql.DB) so the same struct
// works in single-connection mode and against `*sql.Tx`.
type WorkProgramRepositoryPG struct {
	db DBTX
}

// NewWorkProgramRepositoryPG constructs the repository. db can be
// `*sql.DB` (default DI) or `*sql.Tx` (future transactional paths).
func NewWorkProgramRepositoryPG(db DBTX) *WorkProgramRepositoryPG {
	return &WorkProgramRepositoryPG{db: db}
}

// pqUniqueViolation is the SQLSTATE code for a unique-constraint
// violation in PostgreSQL. Mirrors the inline pattern used by the
// curriculum / tasks / assignments modules to keep this bounded
// context free of a dependency on the shared error mapper for a
// single sentinel.
const pqUniqueViolation = "23505"

// uqWPIdentityConstraint is the migration 047 constraint name. The
// repo matches the constraint name (not just the SQLSTATE) so a future
// uniqueness check on a different tuple gets its own mapping rather
// than collapsing both onto ErrWorkProgramIdentityExists.
const uqWPIdentityConstraint = "uq_wp_discipline_specialty_cohort"

// Save inserts a new WorkProgram aggregate root atomically inside a
// transaction. PR 2a ships only the root insert; child aggregate
// persistence (Goals, Competences, Topics, Assessments, References,
// Revisions) lands in the next RED/GREEN pair so the tx scope grows
// incrementally with test coverage.
//
// On success the generated id is written back onto the aggregate via
// SetID. PostgreSQL unique-constraint violation (SQLSTATE 23505)
// against uq_wp_discipline_specialty_cohort maps to
// ErrWorkProgramIdentityExists so the use-case layer gets a
// deterministic 409 mapping without parsing pq error structs itself.
func (r *WorkProgramRepositoryPG) Save(ctx context.Context, wp *entities.WorkProgram) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("work_program: save: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := insertWorkProgramRoot(ctx, tx, wp); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("work_program: save: commit: %w", err)
	}
	return nil
}

// insertWorkProgramRoot performs the root INSERT inside the given tx,
// writing the generated id back onto the aggregate. Extracted so the
// next RED/GREEN pair (child entity persistence) can call it as the
// first step of an extended Save flow without duplication.
func insertWorkProgramRoot(ctx context.Context, tx execQuerier, wp *entities.WorkProgram) error {
	const query = `
		INSERT INTO work_programs (
			discipline_id, specialty_code, applicable_from_year, title,
			annotation, status, author_id, approver_id, approved_at,
			reject_reason, version, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id`

	var newID int64
	err := tx.QueryRowContext(ctx, query,
		wp.DisciplineID(),
		wp.SpecialtyCode(),
		wp.ApplicableFromYear(),
		wp.Title(),
		nullableString(wp.Annotation()),
		string(wp.Status()),
		wp.AuthorID(),
		nullableInt64Ptr(wp.ApproverID()),
		nullableTimePtr(wp.ApprovedAt()),
		nullableString(wp.RejectReason()),
		wp.Version(),
		wp.CreatedAt(),
		wp.UpdatedAt(),
	).Scan(&newID)
	if err != nil {
		if isIdentityViolation(err) {
			return repositories.ErrWorkProgramIdentityExists
		}
		return fmt.Errorf("work_program: save: insert root: %w", err)
	}
	wp.SetID(newID)
	return nil
}

// isIdentityViolation reports whether err is a PostgreSQL unique
// violation against the identity tuple constraint.
func isIdentityViolation(err error) bool {
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		return false
	}
	if string(pqErr.Code) != pqUniqueViolation {
		return false
	}
	// Match constraint name when available — defensive against future
	// uniqueness constraints on the same table.
	return pqErr.Constraint == "" || pqErr.Constraint == uqWPIdentityConstraint
}
