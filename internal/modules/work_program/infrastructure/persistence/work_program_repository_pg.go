package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/repositories"
)

const wpSelectColumns = `id, discipline_id, specialty_code, applicable_from_year, title, annotation, status, author_id, approver_id, approved_at, reject_reason, version, created_at, updated_at`

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

// Save inserts a new WorkProgram aggregate atomically inside a single
// transaction: root row + every populated child collection (Goals,
// Competences, Topics, Assessments, References). Revisions are
// included in the iteration only when the aggregate carries any —
// fresh drafts cannot per ErrRevisionNotPermitted, but Reconstituted
// aggregates may.
//
// On success the generated root id is written back onto the aggregate
// via SetID. PostgreSQL unique-constraint violation (SQLSTATE 23505)
// against uq_wp_discipline_specialty_cohort maps to
// ErrWorkProgramIdentityExists so the use-case layer gets a
// deterministic 409 mapping without parsing pq error structs itself.
// Any child-insert failure surfaces via fmt.Errorf wrapping and the
// deferred Rollback discards the partial state.
func (r *WorkProgramRepositoryPG) Save(ctx context.Context, wp *entities.WorkProgram) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("work_program: save: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := insertWorkProgramRoot(ctx, tx, wp); err != nil {
		return err
	}

	rootID := wp.ID()
	for _, g := range wp.Goals() {
		if err := insertGoal(ctx, tx, rootID, g); err != nil {
			return err
		}
	}
	for _, c := range wp.Competences() {
		if err := insertCompetence(ctx, tx, rootID, c); err != nil {
			return err
		}
	}
	for _, t := range wp.Topics() {
		if err := insertTopic(ctx, tx, rootID, t); err != nil {
			return err
		}
	}
	for _, a := range wp.Assessments() {
		if err := insertAssessment(ctx, tx, rootID, a); err != nil {
			return err
		}
	}
	for _, ref := range wp.References() {
		if err := insertReference(ctx, tx, rootID, ref); err != nil {
			return err
		}
	}
	for _, rev := range wp.Revisions() {
		if err := insertRevision(ctx, tx, rootID, rev); err != nil {
			return err
		}
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

// wpListColumns enumerates the projection used for List — root-only
// fields, no inner aggregate slices (list endpoints stay cheap).
const wpListColumns = `id, discipline_id, specialty_code, applicable_from_year, title, status, author_id, version`

// wpListFilterClause uses cast-and-nullable predicates so a single
// filter shape works for every filter combination (mirror curriculum
// module pattern). Empty string / sql.Null* values disable a predicate.
const wpListFilterClause = `WHERE ($1 = '' OR status = $1)
		AND ($2::bigint IS NULL OR discipline_id = $2::bigint)
		AND ($3 = '' OR specialty_code = $3)
		AND ($4::bigint IS NULL OR applicable_from_year = $4::bigint)
		AND ($5::bigint IS NULL OR author_id = $5::bigint)`

// List returns a page of WorkProgram items matching the filter
// together with the total number of matching rows (ignoring Limit /
// Offset). Items carry root state only; callers needing children use
// GetByID.
func (r *WorkProgramRepositoryPG) List(ctx context.Context, filter repositories.WorkProgramListFilter) (repositories.WorkProgramListResult, error) {
	statusArg := ""
	if filter.Status != nil {
		statusArg = string(*filter.Status)
	}
	var disciplineArg sql.NullInt64
	if filter.DisciplineID != nil {
		disciplineArg = sql.NullInt64{Int64: *filter.DisciplineID, Valid: true}
	}
	var yearArg sql.NullInt64
	if filter.ApplicableFromYear != nil {
		// Carried as int64 (no int32 narrowing): the filter year is
		// untrusted query input, so a value past int32's range must not
		// wrap and spuriously match a stored cohort year. Out-of-range
		// years simply match no row.
		yearArg = sql.NullInt64{Int64: int64(*filter.ApplicableFromYear), Valid: true}
	}
	var authorArg sql.NullInt64
	if filter.AuthorID != nil {
		authorArg = sql.NullInt64{Int64: *filter.AuthorID, Valid: true}
	}

	countQuery := `SELECT COUNT(*) FROM work_programs ` + wpListFilterClause
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery,
		statusArg, disciplineArg, filter.SpecialtyCode, yearArg, authorArg,
	).Scan(&total); err != nil {
		return repositories.WorkProgramListResult{}, fmt.Errorf("work_program: list count: %w", err)
	}

	listQuery := `SELECT ` + wpListColumns + ` FROM work_programs ` + wpListFilterClause + `
		ORDER BY applicable_from_year DESC, created_at DESC, id DESC
		LIMIT $6 OFFSET $7`

	rows, err := r.db.QueryContext(ctx, listQuery,
		statusArg, disciplineArg, filter.SpecialtyCode, yearArg, authorArg,
		filter.Limit, filter.Offset,
	)
	if err != nil {
		return repositories.WorkProgramListResult{}, fmt.Errorf("work_program: list: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var items []repositories.ListItem
	for rows.Next() {
		var (
			id, disciplineID, authorID int64
			specialty, title, statusS  string
			year, version              int
		)
		if err := rows.Scan(&id, &disciplineID, &specialty, &year, &title, &statusS, &authorID, &version); err != nil {
			return repositories.WorkProgramListResult{}, fmt.Errorf("work_program: list scan: %w", err)
		}
		items = append(items, repositories.ListItem{
			ID:                 id,
			DisciplineID:       disciplineID,
			SpecialtyCode:      specialty,
			ApplicableFromYear: year,
			Title:              title,
			Status:             domain.Status(statusS),
			AuthorID:           authorID,
			Version:            version,
		})
	}
	if err := rows.Err(); err != nil {
		return repositories.WorkProgramListResult{}, fmt.Errorf("work_program: list iter: %w", err)
	}
	return repositories.WorkProgramListResult{Items: items, Total: total}, nil
}

// GetByID returns the aggregate with the given id, hydrated through
// Reconstitute*: root + every populated child collection (Goals,
// Competences, Topics, Assessments, References, Revisions). Returns
// repositories.ErrWorkProgramNotFound when no row matches.
func (r *WorkProgramRepositoryPG) GetByID(ctx context.Context, id int64) (*entities.WorkProgram, error) {
	rootIn, err := selectWorkProgramRoot(ctx, r.db, id)
	if err != nil {
		return nil, err
	}

	goals, err := selectGoals(ctx, r.db, id)
	if err != nil {
		return nil, err
	}
	competences, err := selectCompetences(ctx, r.db, id)
	if err != nil {
		return nil, err
	}
	topics, err := selectTopics(ctx, r.db, id)
	if err != nil {
		return nil, err
	}
	assessments, err := selectAssessments(ctx, r.db, id)
	if err != nil {
		return nil, err
	}
	references, err := selectReferences(ctx, r.db, id)
	if err != nil {
		return nil, err
	}
	revisions, err := selectRevisions(ctx, r.db, id)
	if err != nil {
		return nil, err
	}

	rootIn.Goals = goals
	rootIn.Competences = competences
	rootIn.Topics = topics
	rootIn.Assessments = assessments
	rootIn.References = references
	rootIn.Revisions = revisions
	return entities.ReconstituteWorkProgram(rootIn), nil
}

// selectWorkProgramRoot fetches the root row and unwraps nullable
// columns into the Reconstitute input. Inner aggregate slices are
// filled by the caller after sibling child selects.
func selectWorkProgramRoot(ctx context.Context, db DBTX, id int64) (entities.ReconstituteWorkProgramInput, error) {
	query := `SELECT ` + wpSelectColumns + ` FROM work_programs WHERE id = $1`

	var (
		idv, disciplineID, authorID int64
		specialty, title, statusStr string
		applicableFromYear, version int
		annotation, rejectReason    sql.NullString
		approverID                  sql.NullInt64
		approvedAt                  sql.NullTime
		createdAt, updatedAt        time.Time
	)
	err := db.QueryRowContext(ctx, query, id).Scan(
		&idv, &disciplineID, &specialty, &applicableFromYear,
		&title, &annotation, &statusStr, &authorID,
		&approverID, &approvedAt, &rejectReason, &version,
		&createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entities.ReconstituteWorkProgramInput{}, repositories.ErrWorkProgramNotFound
		}
		return entities.ReconstituteWorkProgramInput{}, fmt.Errorf("work_program: get by id: %w", err)
	}

	out := entities.ReconstituteWorkProgramInput{
		ID:                 idv,
		DisciplineID:       disciplineID,
		SpecialtyCode:      specialty,
		ApplicableFromYear: applicableFromYear,
		Title:              title,
		Status:             domain.Status(statusStr),
		AuthorID:           authorID,
		Version:            version,
		CreatedAt:          createdAt,
		UpdatedAt:          updatedAt,
	}
	if annotation.Valid {
		out.Annotation = annotation.String
	}
	if rejectReason.Valid {
		out.RejectReason = rejectReason.String
	}
	if approverID.Valid {
		v := approverID.Int64
		out.ApproverID = &v
	}
	if approvedAt.Valid {
		t := approvedAt.Time
		out.ApprovedAt = &t
	}
	return out, nil
}

// Update writes the (already-mutated) aggregate back atomically:
// UPDATE root with optimistic-lock guard (WHERE id=? AND version=?)
// — server-side `version = version + 1` increment is reflected back
// onto the entity on success — followed by delete + reinsert of every
// child collection inside the same tx.
//
// On RowsAffected == 0 the impl distinguishes via a follow-up SELECT 1:
//   - row missing entirely             → ErrWorkProgramNotFound
//   - row exists but stale version     → ErrWorkProgramVersionConflict
//
// Mirrors the curriculum module pattern (v0.157.0+).
func (r *WorkProgramRepositoryPG) Update(ctx context.Context, wp *entities.WorkProgram) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("work_program: update: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	currentVersion := wp.Version()

	const query = `
		UPDATE work_programs SET
			discipline_id = $1,
			specialty_code = $2,
			applicable_from_year = $3,
			title = $4,
			annotation = $5,
			status = $6,
			author_id = $7,
			approver_id = $8,
			approved_at = $9,
			reject_reason = $10,
			updated_at = $11,
			version = version + 1
		WHERE id = $12 AND version = $13`

	result, err := tx.ExecContext(ctx, query,
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
		wp.UpdatedAt(),
		wp.ID(),
		currentVersion,
	)
	if err != nil {
		return fmt.Errorf("work_program: update: exec: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("work_program: update: rows affected: %w", err)
	}
	if affected == 0 {
		return r.disambiguateAbsentUpdate(ctx, tx, wp.ID())
	}

	// Re-sync children: simplest correct algorithm is delete-all +
	// reinsert-all under the same tx. WorkProgram aggregates are
	// modest in size (≤ a few hundred children combined) so the
	// extra IO is acceptable; a diff-merge optimization is а future
	// refinement once profiling shows it matters.
	if err := deleteAllChildren(ctx, tx, wp.ID()); err != nil {
		return err
	}
	for _, g := range wp.Goals() {
		if err := insertGoal(ctx, tx, wp.ID(), g); err != nil {
			return err
		}
	}
	for _, c := range wp.Competences() {
		if err := insertCompetence(ctx, tx, wp.ID(), c); err != nil {
			return err
		}
	}
	for _, tp := range wp.Topics() {
		if err := insertTopic(ctx, tx, wp.ID(), tp); err != nil {
			return err
		}
	}
	for _, a := range wp.Assessments() {
		if err := insertAssessment(ctx, tx, wp.ID(), a); err != nil {
			return err
		}
	}
	for _, ref := range wp.References() {
		if err := insertReference(ctx, tx, wp.ID(), ref); err != nil {
			return err
		}
	}
	for _, rev := range wp.Revisions() {
		if err := insertRevision(ctx, tx, wp.ID(), rev); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("work_program: update: commit: %w", err)
	}
	wp.SetVersion(currentVersion + 1)
	return nil
}

// disambiguateAbsentUpdate runs a SELECT 1 against the row id; returns
// ErrWorkProgramNotFound when the row is gone (mid-edit deletion),
// ErrWorkProgramVersionConflict when the row exists with a different
// version (stale entity), wrapping any other DB error.
func (r *WorkProgramRepositoryPG) disambiguateAbsentUpdate(ctx context.Context, tx execQuerier, id int64) error {
	var one int
	err := tx.QueryRowContext(ctx, `SELECT 1 FROM work_programs WHERE id = $1`, id).Scan(&one)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repositories.ErrWorkProgramNotFound
		}
		return fmt.Errorf("work_program: update: disambiguate: %w", err)
	}
	return repositories.ErrWorkProgramVersionConflict
}

// Delete removes the WorkProgram row by id. Migration 048
// ON DELETE CASCADE handles every child table cleanup automatically,
// so a single DELETE statement suffices. Returns
// ErrWorkProgramNotFound when no row is deleted.
func (r *WorkProgramRepositoryPG) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM work_programs WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("work_program: delete: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("work_program: delete: rows affected: %w", err)
	}
	if affected == 0 {
		return repositories.ErrWorkProgramNotFound
	}
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
