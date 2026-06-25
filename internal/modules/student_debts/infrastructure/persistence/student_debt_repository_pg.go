package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/repositories"
)

// sdSelectColumns enumerates the student_debts projection used by GetByID
// (full root hydration including DB-owned timestamps).
const sdSelectColumns = `id, student_full_name, group_name, discipline_name, semester, control_form, student_user_id, discipline_id, source_ref, source_hash, status, version, created_at, updated_at`

// sdListColumns enumerates the lightweight List projection — root-only
// fields, no attempts slice (list endpoints stay cheap).
const sdListColumns = `id, student_full_name, group_name, discipline_name, semester, control_form, student_user_id, status, version`

// sdListFilterClause uses cast-and-nullable predicates so a single filter
// shape works for every combination (mirror work_program / curriculum
// pattern). Empty string / sql.Null* values disable a predicate.
const sdListFilterClause = `WHERE ($1 = '' OR group_name = $1)
		AND ($2 = '' OR status = $2)
		AND ($3::bigint IS NULL OR semester = $3::bigint)
		AND ($4::bigint IS NULL OR student_user_id = $4::bigint)
		AND ($5::bigint[] IS NULL OR discipline_id = ANY($5::bigint[]))`

// draSelectColumns enumerates the debt_resit_attempts projection used by
// the attempt hydration query.
const draSelectColumns = `id, debt_id, attempt_no, scheduled_date, examiner, is_commission, result, grade, recorded_by, recorded_at`

// Compile-time assertion that the PG impl satisfies the wide port
// declared in application/usecases (DIP).
var _ usecases.StudentDebtRepository = (*StudentDebtRepositoryPG)(nil)

// StudentDebtRepositoryPG is the SQL implementation of
// StudentDebtRepository. Accepts DBTX (not *sql.DB) so the same struct
// works in single-connection mode and against `*sql.Tx`.
type StudentDebtRepositoryPG struct {
	db DBTX
}

// NewStudentDebtRepositoryPG constructs the repository. db can be
// `*sql.DB` (default DI) or `*sql.Tx` (future transactional paths).
func NewStudentDebtRepositoryPG(db DBTX) *StudentDebtRepositoryPG {
	return &StudentDebtRepositoryPG{db: db}
}

// Save inserts a new StudentDebt aggregate atomically inside a single
// transaction: root row + every resit attempt. On success the generated
// ids are written back onto the root and its attempts. A PostgreSQL
// unique-constraint violation (SQLSTATE 23505) against
// uq_student_debts_identity maps to ErrStudentDebtIdentityExists so the
// use-case layer gets a deterministic 409 / upsert signal. Any failure
// triggers the deferred Rollback, discarding the partial state.
func (r *StudentDebtRepositoryPG) Save(ctx context.Context, debt *entities.StudentDebt) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("student_debts: save: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	const insertRoot = `
		INSERT INTO student_debts (
			student_full_name, group_name, discipline_name, semester,
			control_form, student_user_id, discipline_id, source_ref,
			source_hash, status, version
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id`

	var newID int64
	err = tx.QueryRowContext(ctx, insertRoot,
		debt.StudentFullName,
		debt.GroupName,
		debt.DisciplineName,
		debt.Semester,
		string(debt.ControlForm),
		nullableInt64Ptr(debt.StudentUserID),
		nullableInt64Ptr(debt.DisciplineID),
		debt.SourceRef,
		debt.SourceHash,
		string(debt.Status()),
		debt.Version,
	).Scan(&newID)
	if err != nil {
		if isIdentityViolation(err) {
			return repositories.ErrStudentDebtIdentityExists
		}
		return fmt.Errorf("student_debts: save: insert root: %w", err)
	}
	debt.ID = newID

	for _, a := range debt.Attempts() {
		if err := insertAttempt(ctx, tx, newID, a); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("student_debts: save: commit: %w", err)
	}
	return nil
}

// insertAttempt inserts one resit attempt inside the given tx, writing
// the generated id and debt id back onto the entity. Shared by Save and
// Update so the INSERT shape stays in one place.
func insertAttempt(ctx context.Context, tx execQuerier, debtID int64, a *entities.ResitAttempt) error {
	const query = `
		INSERT INTO debt_resit_attempts (
			debt_id, attempt_no, scheduled_date, examiner, is_commission,
			result, grade, recorded_by, recorded_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`

	var newID int64
	err := tx.QueryRowContext(ctx, query,
		debtID,
		a.AttemptNo,
		a.ScheduledDate(),
		a.Examiner(),
		a.IsCommission,
		string(a.Result()),
		nullableIntPtr(a.Grade()),
		nullableInt64Ptr(a.RecordedBy()),
		nullableTimePtr(a.RecordedAt()),
	).Scan(&newID)
	if err != nil {
		return fmt.Errorf("student_debts: insert attempt: %w", err)
	}
	a.ID = newID
	a.DebtID = debtID
	return nil
}

// GetByID returns the aggregate with the given id, hydrated through
// Reconstitute*: root + its attempts in attempt-no order. Returns
// repositories.ErrStudentDebtNotFound when no row matches.
func (r *StudentDebtRepositoryPG) GetByID(ctx context.Context, id int64) (*entities.StudentDebt, error) {
	query := `SELECT ` + sdSelectColumns + ` FROM student_debts WHERE id = $1`

	var (
		idv                            int64
		studentName, group, discipline string
		semester, version              int
		controlForm, status            string
		sourceRef, sourceHash          string
		studentUserID, disciplineID    sql.NullInt64
		createdAt, updatedAt           time.Time
	)
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&idv, &studentName, &group, &discipline, &semester, &controlForm,
		&studentUserID, &disciplineID, &sourceRef, &sourceHash, &status,
		&version, &createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repositories.ErrStudentDebtNotFound
		}
		return nil, fmt.Errorf("student_debts: get by id: %w", err)
	}

	attempts, err := r.selectAttempts(ctx, id)
	if err != nil {
		return nil, err
	}

	return entities.ReconstituteStudentDebt(
		idv, studentName, group, discipline, semester,
		entities.ControlForm(controlForm),
		nullInt64Ptr(studentUserID), nullInt64Ptr(disciplineID),
		sourceRef, sourceHash, version, entities.DebtStatus(status),
		attempts, createdAt, updatedAt,
	), nil
}

// selectAttempts hydrates the resit attempts for a debt in attempt-no
// order. An empty result is not an error.
func (r *StudentDebtRepositoryPG) selectAttempts(ctx context.Context, debtID int64) ([]*entities.ResitAttempt, error) {
	query := `SELECT ` + draSelectColumns + ` FROM debt_resit_attempts WHERE debt_id = $1 ORDER BY attempt_no`

	rows, err := r.db.QueryContext(ctx, query, debtID)
	if err != nil {
		return nil, fmt.Errorf("student_debts: select attempts: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var attempts []*entities.ResitAttempt
	for rows.Next() {
		var (
			id, dID       int64
			attemptNo     int
			isCommission  bool
			scheduledDate time.Time
			examiner      string
			result        string
			grade         sql.NullInt32
			recordedBy    sql.NullInt64
			recordedAt    sql.NullTime
		)
		if err := rows.Scan(
			&id, &dID, &attemptNo, &scheduledDate, &examiner,
			&isCommission, &result, &grade, &recordedBy, &recordedAt,
		); err != nil {
			return nil, fmt.Errorf("student_debts: scan attempt: %w", err)
		}
		attempts = append(attempts, entities.ReconstituteResitAttempt(
			id, dID, attemptNo, isCommission, scheduledDate, examiner,
			entities.ResitResult(result), nullInt32Ptr(grade),
			nullInt64Ptr(recordedBy), nullTimePtr(recordedAt),
		))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("student_debts: iter attempts: %w", err)
	}
	return attempts, nil
}

// List returns a page of StudentDebt items matching the filter together
// with the total count of matching rows (ignoring Limit / Offset). Items
// carry root state only; callers needing attempts use GetByID.
func (r *StudentDebtRepositoryPG) List(ctx context.Context, filter repositories.StudentDebtListFilter) (repositories.StudentDebtListResult, error) {
	statusArg := ""
	if filter.Status != nil {
		statusArg = string(*filter.Status)
	}
	var semesterArg sql.NullInt64
	if filter.Semester != nil {
		semesterArg = sql.NullInt64{Int64: int64(*filter.Semester), Valid: true}
	}
	var studentArg sql.NullInt64
	if filter.StudentUserID != nil {
		studentArg = sql.NullInt64{Int64: *filter.StudentUserID, Valid: true}
	}
	// nil interface → SQL NULL so the predicate disables; a non-empty
	// slice → bigint[] for the discipline_id = ANY(...) teacher scope.
	// An empty (non-nil) slice is treated as "no filter" too, never as
	// ANY('{}') which would match nothing unintentionally.
	var disciplineArg any
	if len(filter.DisciplineIDs) > 0 {
		disciplineArg = pq.Array(filter.DisciplineIDs)
	}

	countQuery := `SELECT COUNT(*) FROM student_debts ` + sdListFilterClause
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery,
		filter.GroupName, statusArg, semesterArg, studentArg, disciplineArg,
	).Scan(&total); err != nil {
		return repositories.StudentDebtListResult{}, fmt.Errorf("student_debts: list count: %w", err)
	}

	listQuery := `SELECT ` + sdListColumns + ` FROM student_debts ` + sdListFilterClause + `
		ORDER BY group_name, student_full_name, semester, id
		LIMIT $6 OFFSET $7`

	rows, err := r.db.QueryContext(ctx, listQuery,
		filter.GroupName, statusArg, semesterArg, studentArg, disciplineArg,
		filter.Limit, filter.Offset,
	)
	if err != nil {
		return repositories.StudentDebtListResult{}, fmt.Errorf("student_debts: list: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var items []repositories.StudentDebtListItem
	for rows.Next() {
		var (
			id            int64
			studentName   string
			group         string
			discipline    string
			semester      int
			controlForm   string
			studentUserID sql.NullInt64
			status        string
			version       int
		)
		if err := rows.Scan(
			&id, &studentName, &group, &discipline, &semester,
			&controlForm, &studentUserID, &status, &version,
		); err != nil {
			return repositories.StudentDebtListResult{}, fmt.Errorf("student_debts: list scan: %w", err)
		}
		items = append(items, repositories.StudentDebtListItem{
			ID:              id,
			StudentFullName: studentName,
			GroupName:       group,
			DisciplineName:  discipline,
			Semester:        semester,
			ControlForm:     entities.ControlForm(controlForm),
			StudentUserID:   nullInt64Ptr(studentUserID),
			Status:          entities.DebtStatus(status),
			Version:         version,
		})
	}
	if err := rows.Err(); err != nil {
		return repositories.StudentDebtListResult{}, fmt.Errorf("student_debts: list iter: %w", err)
	}
	return repositories.StudentDebtListResult{Items: items, Total: total}, nil
}

// Update writes the mutated aggregate back atomically: UPDATE root with
// optimistic-lock guard (WHERE id=? AND version=?) and a server-side
// version increment, then delete + reinsert every attempt inside the
// same tx. On RowsAffected == 0 a follow-up SELECT distinguishes a
// missing row (ErrStudentDebtNotFound) from a stale version
// (ErrStudentDebtVersionConflict). The DB trigger maintains updated_at.
func (r *StudentDebtRepositoryPG) Update(ctx context.Context, debt *entities.StudentDebt) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("student_debts: update: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	currentVersion := debt.Version

	const query = `
		UPDATE student_debts SET
			student_full_name = $1,
			group_name = $2,
			discipline_name = $3,
			semester = $4,
			control_form = $5,
			student_user_id = $6,
			discipline_id = $7,
			source_ref = $8,
			source_hash = $9,
			status = $10,
			version = version + 1
		WHERE id = $11 AND version = $12`

	result, err := tx.ExecContext(ctx, query,
		debt.StudentFullName,
		debt.GroupName,
		debt.DisciplineName,
		debt.Semester,
		string(debt.ControlForm),
		nullableInt64Ptr(debt.StudentUserID),
		nullableInt64Ptr(debt.DisciplineID),
		debt.SourceRef,
		debt.SourceHash,
		string(debt.Status()),
		debt.ID,
		currentVersion,
	)
	if err != nil {
		if isIdentityViolation(err) {
			return repositories.ErrStudentDebtIdentityExists
		}
		return fmt.Errorf("student_debts: update: exec: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("student_debts: update: rows affected: %w", err)
	}
	if affected == 0 {
		return r.disambiguateAbsentUpdate(ctx, tx, debt.ID)
	}

	// Re-sync attempts: delete-all + reinsert-all under the same tx.
	// A debt carries a handful of attempts at most, so the extra IO is
	// negligible and the algorithm stays trivially correct.
	if _, err := tx.ExecContext(ctx, `DELETE FROM debt_resit_attempts WHERE debt_id = $1`, debt.ID); err != nil {
		return fmt.Errorf("student_debts: update: delete attempts: %w", err)
	}
	for _, a := range debt.Attempts() {
		if err := insertAttempt(ctx, tx, debt.ID, a); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("student_debts: update: commit: %w", err)
	}
	debt.Version = currentVersion + 1
	return nil
}

// disambiguateAbsentUpdate runs a SELECT 1 against the row id; returns
// ErrStudentDebtNotFound when the row is gone (mid-edit deletion),
// ErrStudentDebtVersionConflict when the row exists with a different
// version (stale entity), wrapping any other DB error.
func (r *StudentDebtRepositoryPG) disambiguateAbsentUpdate(ctx context.Context, tx execQuerier, id int64) error {
	var one int
	err := tx.QueryRowContext(ctx, `SELECT 1 FROM student_debts WHERE id = $1`, id).Scan(&one)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repositories.ErrStudentDebtNotFound
		}
		return fmt.Errorf("student_debts: update: disambiguate: %w", err)
	}
	return repositories.ErrStudentDebtVersionConflict
}

// --- nullable column unwrappers --------------------------------------------

// nullInt64Ptr unwraps a nullable bigint column into a *int64 (nil when
// the column was SQL NULL).
func nullInt64Ptr(n sql.NullInt64) *int64 {
	if !n.Valid {
		return nil
	}
	v := n.Int64
	return &v
}

// nullInt32Ptr unwraps a nullable int column into a *int (nil when the
// column was SQL NULL) — used for the optional grade.
func nullInt32Ptr(n sql.NullInt32) *int {
	if !n.Valid {
		return nil
	}
	v := int(n.Int32)
	return &v
}

// nullTimePtr unwraps a nullable timestamp column into a *time.Time (nil
// when the column was SQL NULL) — used for the optional recorded_at.
func nullTimePtr(n sql.NullTime) *time.Time {
	if !n.Valid {
		return nil
	}
	t := n.Time
	return &t
}
