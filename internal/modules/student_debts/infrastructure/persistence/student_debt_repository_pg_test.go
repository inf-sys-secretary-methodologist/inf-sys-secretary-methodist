package persistence

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/domain/repositories"
)

func newSDRepoMock(t *testing.T) (*StudentDebtRepositoryPG, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewStudentDebtRepositoryPG(db), mock
}

func freshOpenDebt(t *testing.T) *entities.StudentDebt {
	t.Helper()
	d, err := entities.NewStudentDebt("Иванов Иван", "ИВТ-21", "Базы данных", 3, entities.ControlFormExam)
	require.NoError(t, err)
	return d
}

func debtWithScheduledResit(t *testing.T) *entities.StudentDebt {
	t.Helper()
	d := freshOpenDebt(t)
	require.NoError(t, d.ScheduleResit(
		time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC), "Петров П.П.", time.Now()))
	return d
}

// --- Save ------------------------------------------------------------------

func TestStudentDebtRepositoryPG_Save_RootOnly_HappyPath(t *testing.T) {
	repo, mock := newSDRepoMock(t)
	d := freshOpenDebt(t)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO student_debts")).
		WithArgs(
			"Иванов Иван",    // student_full_name
			"ИВТ-21",         // group_name
			"Базы данных",    // discipline_name
			3,                // semester
			"exam",           // control_form
			sqlmock.AnyArg(), // student_user_id (NULL)
			sqlmock.AnyArg(), // discipline_id (NULL)
			"",               // source_ref
			"",               // source_hash
			"open",           // status
			1,                // version
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(55)))
	mock.ExpectCommit()

	err := repo.Save(context.Background(), d)
	require.NoError(t, err)
	assert.Equal(t, int64(55), d.ID, "Save should write the generated id back onto the aggregate")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStudentDebtRepositoryPG_Save_WithAttempt_InsertsBoth(t *testing.T) {
	repo, mock := newSDRepoMock(t)
	d := debtWithScheduledResit(t)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO student_debts")).
		WithArgs(
			"Иванов Иван",     // student_full_name
			"ИВТ-21",          // group_name
			"Базы данных",     // discipline_name
			3,                 // semester
			"exam",            // control_form
			sqlmock.AnyArg(),  // student_user_id (NULL)
			sqlmock.AnyArg(),  // discipline_id (NULL)
			"",                // source_ref
			"",                // source_hash
			"resit_scheduled", // status (ScheduleResit moved it off open)
			1,                 // version
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(55)))
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO debt_resit_attempts")).
		WithArgs(
			int64(55),        // debt_id
			1,                // attempt_no
			sqlmock.AnyArg(), // scheduled_date
			"Петров П.П.",    // examiner
			false,            // is_commission
			"pending",        // result
			sqlmock.AnyArg(), // grade (NULL)
			sqlmock.AnyArg(), // recorded_by (NULL)
			sqlmock.AnyArg(), // recorded_at (NULL)
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(900)))
	mock.ExpectCommit()

	err := repo.Save(context.Background(), d)
	require.NoError(t, err)
	assert.Equal(t, int64(900), d.Attempts()[0].ID, "Save should write the attempt id back")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStudentDebtRepositoryPG_Save_IdentityConflict_MapsSentinel(t *testing.T) {
	repo, mock := newSDRepoMock(t)
	d := freshOpenDebt(t)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO student_debts")).
		WillReturnError(&pq.Error{
			Code:       "23505",
			Constraint: "uq_student_debts_identity",
		})
	mock.ExpectRollback()

	err := repo.Save(context.Background(), d)
	assert.ErrorIs(t, err, repositories.ErrStudentDebtIdentityExists,
		"unique violation on the natural key must map to ErrStudentDebtIdentityExists")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// --- GetByID ---------------------------------------------------------------

func TestStudentDebtRepositoryPG_GetByID_HydratesRootAndAttempts(t *testing.T) {
	repo, mock := newSDRepoMock(t)

	now := time.Date(2026, 6, 24, 12, 0, 0, 0, time.UTC)
	sched := time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC)

	mock.ExpectQuery(regexp.QuoteMeta("FROM student_debts WHERE id = $1")).
		WithArgs(int64(55)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "student_full_name", "group_name", "discipline_name",
			"semester", "control_form", "student_user_id", "discipline_id",
			"source_ref", "source_hash", "status", "version",
			"created_at", "updated_at",
		}).AddRow(
			int64(55), "Иванов Иван", "ИВТ-21", "Базы данных",
			3, "exam", nil, nil, "ved-7", "abc", "resit_scheduled", 2,
			now, now,
		))
	mock.ExpectQuery(regexp.QuoteMeta("FROM debt_resit_attempts WHERE debt_id = $1")).
		WithArgs(int64(55)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "debt_id", "attempt_no", "scheduled_date", "examiner",
			"is_commission", "result", "grade", "recorded_by", "recorded_at",
		}).AddRow(
			int64(900), int64(55), 1, sched, "Петров П.П.",
			false, "pending", nil, nil, nil,
		))

	d, err := repo.GetByID(context.Background(), 55)
	require.NoError(t, err)
	require.NotNil(t, d)
	assert.Equal(t, int64(55), d.ID)
	assert.Equal(t, "Иванов Иван", d.StudentFullName)
	assert.Equal(t, entities.DebtStatusResitScheduled, d.Status())
	assert.Equal(t, 2, d.Version)
	require.Len(t, d.Attempts(), 1)
	assert.Equal(t, 1, d.Attempts()[0].AttemptNo)
	assert.Equal(t, "Петров П.П.", d.Attempts()[0].Examiner())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStudentDebtRepositoryPG_GetByID_HydratesNullableColumns(t *testing.T) {
	repo, mock := newSDRepoMock(t)

	now := time.Date(2026, 6, 24, 12, 0, 0, 0, time.UTC)
	sched := time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC)
	recorded := time.Date(2026, 7, 1, 11, 0, 0, 0, time.UTC)

	// Every nullable column is populated so the non-null unwrap branch of
	// nullInt64Ptr / nullInt32Ptr / nullTimePtr is exercised — a wrong
	// pointer or lost value would surface here, not in the all-NULL case.
	mock.ExpectQuery(regexp.QuoteMeta("FROM student_debts WHERE id = $1")).
		WithArgs(int64(55)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "student_full_name", "group_name", "discipline_name",
			"semester", "control_form", "student_user_id", "discipline_id",
			"source_ref", "source_hash", "status", "version",
			"created_at", "updated_at",
		}).AddRow(
			int64(55), "Иванов Иван", "ИВТ-21", "Базы данных",
			3, "exam", int64(42), int64(7), "ved-7", "abc", "closed_passed", 4,
			now, now,
		))
	mock.ExpectQuery(regexp.QuoteMeta("FROM debt_resit_attempts WHERE debt_id = $1")).
		WithArgs(int64(55)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "debt_id", "attempt_no", "scheduled_date", "examiner",
			"is_commission", "result", "grade", "recorded_by", "recorded_at",
		}).AddRow(
			int64(900), int64(55), 1, sched, "Петров П.П.",
			true, "passed", int64(5), int64(42), recorded,
		))

	d, err := repo.GetByID(context.Background(), 55)
	require.NoError(t, err)
	require.NotNil(t, d.StudentUserID)
	assert.Equal(t, int64(42), *d.StudentUserID)
	require.NotNil(t, d.DisciplineID)
	assert.Equal(t, int64(7), *d.DisciplineID)

	require.Len(t, d.Attempts(), 1)
	a := d.Attempts()[0]
	assert.True(t, a.IsCommission)
	require.NotNil(t, a.Grade())
	assert.Equal(t, 5, *a.Grade())
	require.NotNil(t, a.RecordedBy())
	assert.Equal(t, int64(42), *a.RecordedBy())
	require.NotNil(t, a.RecordedAt())
	assert.Equal(t, recorded, *a.RecordedAt())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStudentDebtRepositoryPG_GetByID_NotFound_MapsSentinel(t *testing.T) {
	repo, mock := newSDRepoMock(t)

	// Empty result set → QueryRow.Scan returns sql.ErrNoRows, which the
	// impl maps to the NotFound sentinel.
	mock.ExpectQuery(regexp.QuoteMeta("FROM student_debts WHERE id = $1")).
		WithArgs(int64(404)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	_, err := repo.GetByID(context.Background(), 404)
	assert.ErrorIs(t, err, repositories.ErrStudentDebtNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// --- List ------------------------------------------------------------------

func TestStudentDebtRepositoryPG_List_ReturnsPageAndTotal(t *testing.T) {
	repo, mock := newSDRepoMock(t)

	status := entities.DebtStatusOpen
	filter := repositories.StudentDebtListFilter{
		Status: &status,
		Limit:  20,
		Offset: 0,
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM student_debts")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery(regexp.QuoteMeta("FROM student_debts")).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "student_full_name", "group_name", "discipline_name",
			"semester", "control_form", "student_user_id", "status", "version",
		}).
			AddRow(int64(1), "Иванов Иван", "ИВТ-21", "Базы данных", 3, "exam", nil, "open", 1).
			AddRow(int64(2), "Петров Пётр", "ИВТ-21", "Сети", 4, "zachet", nil, "open", 1))

	res, err := repo.List(context.Background(), filter)
	require.NoError(t, err)
	assert.Equal(t, 2, res.Total)
	require.Len(t, res.Items, 2)
	assert.Equal(t, "Иванов Иван", res.Items[0].StudentFullName)
	assert.Equal(t, entities.DebtStatusOpen, res.Items[1].Status)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStudentDebtRepositoryPG_List_FiltersByDisciplineIDs(t *testing.T) {
	repo, mock := newSDRepoMock(t)

	filter := repositories.StudentDebtListFilter{
		DisciplineIDs: []int64{7, 9},
		Limit:         20,
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM student_debts")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	// The list query must carry a discipline_id = ANY(...) predicate when
	// DisciplineIDs is set — this is the teacher-scope mechanism.
	mock.ExpectQuery(regexp.QuoteMeta("discipline_id = ANY")).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "student_full_name", "group_name", "discipline_name",
			"semester", "control_form", "student_user_id", "status", "version",
		}).AddRow(int64(1), "Иванов Иван", "ИВТ-21", "Базы данных", 3, "exam", nil, "open", 1))

	res, err := repo.List(context.Background(), filter)
	require.NoError(t, err)
	assert.Equal(t, 1, res.Total)
	require.Len(t, res.Items, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// --- Update ----------------------------------------------------------------

func TestStudentDebtRepositoryPG_Update_HappyPath_BumpsVersion(t *testing.T) {
	repo, mock := newSDRepoMock(t)
	d := debtWithScheduledResit(t)
	d.ID = 55 // reconstituted aggregate id

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE student_debts SET")).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM debt_resit_attempts WHERE debt_id = $1")).
		WithArgs(int64(55)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO debt_resit_attempts")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(901)))
	mock.ExpectCommit()

	err := repo.Update(context.Background(), d)
	require.NoError(t, err)
	assert.Equal(t, 2, d.Version, "Update should reflect the server-side version bump")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStudentDebtRepositoryPG_Update_IdentityConflict_MapsSentinel(t *testing.T) {
	repo, mock := newSDRepoMock(t)
	d := freshOpenDebt(t)
	d.ID = 55

	// Re-import can correct a typo in the natural key (group / name /
	// discipline / semester); if the corrected key collides with another
	// debt, the UPDATE raises 23505 and the repo must surface the same
	// identity sentinel as Save so the use case treats it consistently.
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE student_debts SET")).
		WillReturnError(&pq.Error{
			Code:       "23505",
			Constraint: "uq_student_debts_identity",
		})
	mock.ExpectRollback()

	err := repo.Update(context.Background(), d)
	assert.ErrorIs(t, err, repositories.ErrStudentDebtIdentityExists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStudentDebtRepositoryPG_Update_StaleVersion_MapsConflict(t *testing.T) {
	repo, mock := newSDRepoMock(t)
	d := freshOpenDebt(t)
	d.ID = 55

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE student_debts SET")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM student_debts WHERE id = $1")).
		WithArgs(int64(55)).
		WillReturnRows(sqlmock.NewRows([]string{"?column?"}).AddRow(1))
	mock.ExpectRollback()

	err := repo.Update(context.Background(), d)
	assert.ErrorIs(t, err, repositories.ErrStudentDebtVersionConflict)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStudentDebtRepositoryPG_Update_RowMissing_MapsNotFound(t *testing.T) {
	repo, mock := newSDRepoMock(t)
	d := freshOpenDebt(t)
	d.ID = 55

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE student_debts SET")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	// Empty result set → the disambiguation SELECT scans sql.ErrNoRows,
	// meaning the row is gone entirely.
	mock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM student_debts WHERE id = $1")).
		WithArgs(int64(55)).
		WillReturnRows(sqlmock.NewRows([]string{"?column?"}))
	mock.ExpectRollback()

	err := repo.Update(context.Background(), d)
	assert.ErrorIs(t, err, repositories.ErrStudentDebtNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}
