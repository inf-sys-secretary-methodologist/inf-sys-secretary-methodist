package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/domain/entities"
)

// ===================== Employee Repository Tests =====================

const testRawDataJSON = `{"x":1}`

func newEmpRepoMock(t *testing.T) (*ExternalEmployeeRepositoryPg, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	repo := NewExternalEmployeeRepositoryPg(db)
	return repo.(*ExternalEmployeeRepositoryPg), mock
}

var empCols = []string{
	"id", "external_id", "code", "first_name", "last_name", "middle_name",
	"email", "phone", "position", "department", "employment_date", "dismissal_date",
	"is_active", "local_user_id", "last_sync_at", "external_data_hash", "raw_data",
	"created_at", "updated_at",
}

func newEmpRows() *sqlmock.Rows { return sqlmock.NewRows(empCols) }

func addEmpRow(rows *sqlmock.Rows, id int64, extID string) *sqlmock.Rows {
	now := time.Now()
	localUserID := int64(100)
	empDate := now
	return rows.AddRow(
		id, extID, "CODE1", "John", "Doe", "M",
		"email@test.com", "+7999", "Dev", "IT", empDate, nil,
		true, localUserID, now, "hash123", []byte(`{"key":"val"}`),
		now, now,
	)
}

// --- Create ---

func TestExternalEmployeeRepositoryPg_Create_Success(t *testing.T) {
	repo, mock := newEmpRepoMock(t)
	emp := entities.NewExternalEmployee("ext1", "C1")
	emp.FirstName = "John"
	emp.LastName = "Doe"
	emp.RawData = `{"test":true}`

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO external_employees")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	err := repo.Create(context.Background(), emp)
	require.NoError(t, err)
	assert.Equal(t, int64(1), emp.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalEmployeeRepositoryPg_Create_NoRawData(t *testing.T) {
	repo, mock := newEmpRepoMock(t)
	emp := entities.NewExternalEmployee("ext1", "C1")

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO external_employees")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	require.NoError(t, repo.Create(context.Background(), emp))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalEmployeeRepositoryPg_Create_Error(t *testing.T) {
	repo, mock := newEmpRepoMock(t)
	emp := entities.NewExternalEmployee("ext1", "C1")

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO external_employees")).
		WillReturnError(sql.ErrConnDone)

	require.Error(t, repo.Create(context.Background(), emp))
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- Update ---

func TestExternalEmployeeRepositoryPg_Update_Success(t *testing.T) {
	repo, mock := newEmpRepoMock(t)
	emp := entities.NewExternalEmployee("ext1", "C1")
	emp.ID = 5
	emp.RawData = `{"key":"val"}`

	mock.ExpectExec(regexp.QuoteMeta("UPDATE external_employees SET")).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.Update(context.Background(), emp))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalEmployeeRepositoryPg_Update_NoRawData(t *testing.T) {
	repo, mock := newEmpRepoMock(t)
	emp := entities.NewExternalEmployee("ext1", "C1")
	emp.ID = 5

	mock.ExpectExec(regexp.QuoteMeta("UPDATE external_employees SET")).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.Update(context.Background(), emp))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalEmployeeRepositoryPg_Update_NotFound(t *testing.T) {
	repo, mock := newEmpRepoMock(t)
	emp := entities.NewExternalEmployee("ext1", "C1")
	emp.ID = 999

	mock.ExpectExec(regexp.QuoteMeta("UPDATE external_employees")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Update(context.Background(), emp)
	require.ErrorIs(t, err, sql.ErrNoRows)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalEmployeeRepositoryPg_Update_Error(t *testing.T) {
	repo, mock := newEmpRepoMock(t)
	emp := entities.NewExternalEmployee("ext1", "C1")
	emp.ID = 5

	mock.ExpectExec(regexp.QuoteMeta("UPDATE external_employees")).
		WillReturnError(sql.ErrConnDone)

	require.Error(t, repo.Update(context.Background(), emp))
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- Upsert ---

func TestExternalEmployeeRepositoryPg_Upsert_Success(t *testing.T) {
	repo, mock := newEmpRepoMock(t)
	emp := entities.NewExternalEmployee("ext1", "C1")
	emp.RawData = `{"data":true}`

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO external_employees")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	require.NoError(t, repo.Upsert(context.Background(), emp))
	assert.Equal(t, int64(1), emp.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalEmployeeRepositoryPg_Upsert_NoRawData(t *testing.T) {
	repo, mock := newEmpRepoMock(t)
	emp := entities.NewExternalEmployee("ext1", "C1")

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO external_employees")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	require.NoError(t, repo.Upsert(context.Background(), emp))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalEmployeeRepositoryPg_Upsert_Error(t *testing.T) {
	repo, mock := newEmpRepoMock(t)
	emp := entities.NewExternalEmployee("ext1", "C1")

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO external_employees")).
		WillReturnError(sql.ErrConnDone)

	require.Error(t, repo.Upsert(context.Background(), emp))
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetByID ---

func TestExternalEmployeeRepositoryPg_GetByID_Success(t *testing.T) {
	repo, mock := newEmpRepoMock(t)
	rows := addEmpRow(newEmpRows(), 1, "ext1")

	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WithArgs(int64(1)).WillReturnRows(rows)

	emp, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	require.NotNil(t, emp)
	assert.Equal(t, "John", emp.FirstName)
	assert.Equal(t, "M", emp.MiddleName)
	assert.Equal(t, "email@test.com", emp.Email)
	assert.NotNil(t, emp.EmploymentDate)
	assert.NotNil(t, emp.LocalUserID)
	assert.Equal(t, "hash123", emp.ExternalDataHash)
	assert.Contains(t, emp.RawData, "key")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalEmployeeRepositoryPg_GetByID_NotFound(t *testing.T) {
	repo, mock := newEmpRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnError(sql.ErrNoRows)

	emp, err := repo.GetByID(context.Background(), 999)
	require.NoError(t, err)
	assert.Nil(t, emp)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalEmployeeRepositoryPg_GetByID_Error(t *testing.T) {
	repo, mock := newEmpRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnError(sql.ErrConnDone)

	_, err := repo.GetByID(context.Background(), 1)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetByExternalID ---

func TestExternalEmployeeRepositoryPg_GetByExternalID_Success(t *testing.T) {
	repo, mock := newEmpRepoMock(t)
	rows := addEmpRow(newEmpRows(), 1, "ext1")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnRows(rows)

	emp, err := repo.GetByExternalID(context.Background(), "ext1")
	require.NoError(t, err)
	require.NotNil(t, emp)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalEmployeeRepositoryPg_GetByExternalID_NotFound(t *testing.T) {
	repo, mock := newEmpRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnError(sql.ErrNoRows)

	emp, err := repo.GetByExternalID(context.Background(), "nope")
	require.NoError(t, err)
	assert.Nil(t, emp)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalEmployeeRepositoryPg_GetByExternalID_Error(t *testing.T) {
	repo, mock := newEmpRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnError(sql.ErrConnDone)

	_, err := repo.GetByExternalID(context.Background(), "ext1")
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetByCode ---

func TestExternalEmployeeRepositoryPg_GetByCode_Success(t *testing.T) {
	repo, mock := newEmpRepoMock(t)
	rows := addEmpRow(newEmpRows(), 1, "ext1")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnRows(rows)

	emp, err := repo.GetByCode(context.Background(), "CODE1")
	require.NoError(t, err)
	require.NotNil(t, emp)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalEmployeeRepositoryPg_GetByCode_NotFound(t *testing.T) {
	repo, mock := newEmpRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnError(sql.ErrNoRows)

	emp, err := repo.GetByCode(context.Background(), "nope")
	require.NoError(t, err)
	assert.Nil(t, emp)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalEmployeeRepositoryPg_GetByCode_Error(t *testing.T) {
	repo, mock := newEmpRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnError(sql.ErrConnDone)

	_, err := repo.GetByCode(context.Background(), "C1")
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetByLocalUserID ---

func TestExternalEmployeeRepositoryPg_GetByLocalUserID_Success(t *testing.T) {
	repo, mock := newEmpRepoMock(t)
	rows := addEmpRow(newEmpRows(), 1, "ext1")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnRows(rows)

	emp, err := repo.GetByLocalUserID(context.Background(), 100)
	require.NoError(t, err)
	require.NotNil(t, emp)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalEmployeeRepositoryPg_GetByLocalUserID_NotFound(t *testing.T) {
	repo, mock := newEmpRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnError(sql.ErrNoRows)

	emp, err := repo.GetByLocalUserID(context.Background(), 999)
	require.NoError(t, err)
	assert.Nil(t, emp)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalEmployeeRepositoryPg_GetByLocalUserID_Error(t *testing.T) {
	repo, mock := newEmpRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnError(sql.ErrConnDone)

	_, err := repo.GetByLocalUserID(context.Background(), 100)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- List ---

func TestExternalEmployeeRepositoryPg_List_NoFilter(t *testing.T) {
	repo, mock := newEmpRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	rows := addEmpRow(newEmpRows(), 1, "ext1")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnRows(rows)

	emps, total, err := repo.List(context.Background(), entities.ExternalEmployeeFilter{})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, emps, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalEmployeeRepositoryPg_List_AllFilters(t *testing.T) {
	repo, mock := newEmpRepoMock(t)
	isActive := true
	isLinked := true
	filter := entities.ExternalEmployeeFilter{
		Search:     "John",
		Department: "IT",
		Position:   "Dev",
		IsActive:   &isActive,
		IsLinked:   &isLinked,
		Limit:      5,
		Offset:     0,
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	rows := addEmpRow(newEmpRows(), 1, "ext1")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnRows(rows)

	emps, _, err := repo.List(context.Background(), filter)
	require.NoError(t, err)
	assert.Len(t, emps, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalEmployeeRepositoryPg_List_IsLinkedFalse(t *testing.T) {
	repo, mock := newEmpRepoMock(t)
	isLinked := false
	filter := entities.ExternalEmployeeFilter{IsLinked: &isLinked}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnRows(newEmpRows())

	_, total, err := repo.List(context.Background(), filter)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalEmployeeRepositoryPg_List_CountError(t *testing.T) {
	repo, mock := newEmpRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnError(sql.ErrConnDone)

	_, _, err := repo.List(context.Background(), entities.ExternalEmployeeFilter{})
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalEmployeeRepositoryPg_List_QueryError(t *testing.T) {
	repo, mock := newEmpRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnError(sql.ErrConnDone)

	_, _, err := repo.List(context.Background(), entities.ExternalEmployeeFilter{})
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalEmployeeRepositoryPg_List_ScanError(t *testing.T) {
	repo, mock := newEmpRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	rows := sqlmock.NewRows(empCols).AddRow(
		"bad", "ext", "C", "F", "L", nil, nil, nil, nil, nil, nil, nil,
		true, nil, time.Now(), nil, nil, time.Now(), time.Now(),
	)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnRows(rows)

	_, _, err := repo.List(context.Background(), entities.ExternalEmployeeFilter{})
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetUnlinked ---

func TestExternalEmployeeRepositoryPg_GetUnlinked(t *testing.T) {
	repo, mock := newEmpRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnRows(newEmpRows())

	emps, _, err := repo.GetUnlinked(context.Background(), 10, 0)
	require.NoError(t, err)
	assert.Empty(t, emps)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- LinkToLocalUser ---

func TestExternalEmployeeRepositoryPg_LinkToLocalUser_Success(t *testing.T) {
	repo, mock := newEmpRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE external_employees SET local_user_id")).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.LinkToLocalUser(context.Background(), 1, 100))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalEmployeeRepositoryPg_LinkToLocalUser_NotFound(t *testing.T) {
	repo, mock := newEmpRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE external_employees")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	require.ErrorIs(t, repo.LinkToLocalUser(context.Background(), 999, 100), sql.ErrNoRows)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalEmployeeRepositoryPg_LinkToLocalUser_Error(t *testing.T) {
	repo, mock := newEmpRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE external_employees")).
		WillReturnError(sql.ErrConnDone)

	require.Error(t, repo.LinkToLocalUser(context.Background(), 1, 100))
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- Unlink ---

func TestExternalEmployeeRepositoryPg_Unlink_Success(t *testing.T) {
	repo, mock := newEmpRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE external_employees SET local_user_id = NULL")).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.Unlink(context.Background(), 1))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalEmployeeRepositoryPg_Unlink_NotFound(t *testing.T) {
	repo, mock := newEmpRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE external_employees")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	require.ErrorIs(t, repo.Unlink(context.Background(), 999), sql.ErrNoRows)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalEmployeeRepositoryPg_Unlink_Error(t *testing.T) {
	repo, mock := newEmpRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE external_employees")).
		WillReturnError(sql.ErrConnDone)

	require.Error(t, repo.Unlink(context.Background(), 1))
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- Delete ---

func TestExternalEmployeeRepositoryPg_Delete_Success(t *testing.T) {
	repo, mock := newEmpRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM external_employees")).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, repo.Delete(context.Background(), 1))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalEmployeeRepositoryPg_Delete_NotFound(t *testing.T) {
	repo, mock := newEmpRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM external_employees")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	require.ErrorIs(t, repo.Delete(context.Background(), 999), sql.ErrNoRows)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalEmployeeRepositoryPg_Delete_Error(t *testing.T) {
	repo, mock := newEmpRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM external_employees")).
		WillReturnError(sql.ErrConnDone)

	require.Error(t, repo.Delete(context.Background(), 1))
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetAllExternalIDs ---

func TestExternalEmployeeRepositoryPg_GetAllExternalIDs_Success(t *testing.T) {
	repo, mock := newEmpRepoMock(t)
	rows := sqlmock.NewRows([]string{"external_id"}).AddRow("ext1").AddRow("ext2")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT external_id FROM external_employees")).WillReturnRows(rows)

	ids, err := repo.GetAllExternalIDs(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []string{"ext1", "ext2"}, ids)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalEmployeeRepositoryPg_GetAllExternalIDs_Error(t *testing.T) {
	repo, mock := newEmpRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT external_id")).WillReturnError(sql.ErrConnDone)

	_, err := repo.GetAllExternalIDs(context.Background())
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalEmployeeRepositoryPg_GetAllExternalIDs_ScanError(t *testing.T) {
	repo, mock := newEmpRepoMock(t)
	rows := sqlmock.NewRows([]string{"external_id"}).AddRow(123)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT external_id")).WillReturnRows(rows)

	// int can be scanned to string, so let's use a row error instead
	rows2 := sqlmock.NewRows([]string{"external_id"}).AddRow("e1")
	rows2.RowError(0, fmt.Errorf("scan error"))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT external_id")).WillReturnRows(rows2)

	// First call just works with int->string, second gets row error
	_, _ = repo.GetAllExternalIDs(context.Background())
	_, err := repo.GetAllExternalIDs(context.Background())
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- BulkUpsert ---

func TestExternalEmployeeRepositoryPg_BulkUpsert_Empty(t *testing.T) {
	repo, _ := newEmpRepoMock(t)
	require.NoError(t, repo.BulkUpsert(context.Background(), nil))
}

func TestExternalEmployeeRepositoryPg_BulkUpsert_Success(t *testing.T) {
	repo, mock := newEmpRepoMock(t)
	emp1 := entities.NewExternalEmployee("ext1", "C1")
	emp1.RawData = testRawDataJSON
	emp2 := entities.NewExternalEmployee("ext2", "C2")

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO external_employees")).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO external_employees")).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	require.NoError(t, repo.BulkUpsert(context.Background(), []*entities.ExternalEmployee{emp1, emp2}))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalEmployeeRepositoryPg_BulkUpsert_BeginError(t *testing.T) {
	repo, mock := newEmpRepoMock(t)
	emp := entities.NewExternalEmployee("ext1", "C1")
	mock.ExpectBegin().WillReturnError(sql.ErrConnDone)

	require.Error(t, repo.BulkUpsert(context.Background(), []*entities.ExternalEmployee{emp}))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalEmployeeRepositoryPg_BulkUpsert_ExecError(t *testing.T) {
	repo, mock := newEmpRepoMock(t)
	emp := entities.NewExternalEmployee("ext1", "C1")
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO external_employees")).WillReturnError(sql.ErrConnDone)
	mock.ExpectRollback()

	require.Error(t, repo.BulkUpsert(context.Background(), []*entities.ExternalEmployee{emp}))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalEmployeeRepositoryPg_BulkUpsert_CommitError(t *testing.T) {
	repo, mock := newEmpRepoMock(t)
	emp := entities.NewExternalEmployee("ext1", "C1")
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO external_employees")).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit().WillReturnError(sql.ErrConnDone)

	require.Error(t, repo.BulkUpsert(context.Background(), []*entities.ExternalEmployee{emp}))
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- MarkInactiveExcept ---

func TestExternalEmployeeRepositoryPg_MarkInactiveExcept_Empty(t *testing.T) {
	repo, mock := newEmpRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE external_employees SET is_active = false")).
		WillReturnResult(sqlmock.NewResult(0, 5))

	require.NoError(t, repo.MarkInactiveExcept(context.Background(), nil))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalEmployeeRepositoryPg_MarkInactiveExcept_WithIDs(t *testing.T) {
	repo, mock := newEmpRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE external_employees")).
		WithArgs(pq.Array([]string{"ext1", "ext2"})).
		WillReturnResult(sqlmock.NewResult(0, 3))

	require.NoError(t, repo.MarkInactiveExcept(context.Background(), []string{"ext1", "ext2"}))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalEmployeeRepositoryPg_MarkInactiveExcept_Error(t *testing.T) {
	repo, mock := newEmpRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE external_employees")).
		WillReturnError(sql.ErrConnDone)

	require.Error(t, repo.MarkInactiveExcept(context.Background(), []string{"ext1"}))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalEmployeeRepositoryPg_MarkInactiveExcept_EmptyError(t *testing.T) {
	repo, mock := newEmpRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE external_employees")).
		WillReturnError(sql.ErrConnDone)

	require.Error(t, repo.MarkInactiveExcept(context.Background(), nil))
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- Helper functions ---

func TestNullString(t *testing.T) {
	ns := nullString("test")
	assert.True(t, ns.Valid)
	assert.Equal(t, "test", ns.String)

	ns = nullString("")
	assert.False(t, ns.Valid)
}

func TestNullInt64(t *testing.T) {
	val := int64(42)
	ni := nullInt64(&val)
	assert.True(t, ni.Valid)
	assert.Equal(t, int64(42), ni.Int64)

	ni = nullInt64(nil)
	assert.False(t, ni.Valid)
}

func TestNullTime(t *testing.T) {
	now := time.Now()
	nt := nullTime(&now)
	assert.True(t, nt.Valid)

	nt = nullTime(nil)
	assert.False(t, nt.Valid)
}

// ===================== Student Repository Tests =====================

func newStudRepoMock(t *testing.T) (*ExternalStudentRepositoryPg, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	repo := NewExternalStudentRepositoryPg(db)
	return repo.(*ExternalStudentRepositoryPg), mock
}

var studCols = []string{
	"id", "external_id", "code", "first_name", "last_name", "middle_name",
	"email", "phone", "group_name", "faculty", "specialty", "course",
	"study_form", "enrollment_date", "expulsion_date", "graduation_date",
	"status", "is_active", "local_user_id", "last_sync_at",
	"external_data_hash", "raw_data", "created_at", "updated_at",
}

func newStudRows() *sqlmock.Rows { return sqlmock.NewRows(studCols) }

func addStudRow(rows *sqlmock.Rows, id int64, extID string) *sqlmock.Rows {
	now := time.Now()
	localID := int64(100)
	return rows.AddRow(
		id, extID, "CODE1", "Jane", "Doe", "M",
		"jane@test.com", "+7999", "GR-01", "CS", "SE", int32(2),
		"full-time", now, nil, nil,
		"enrolled", true, localID, now,
		"hash123", []byte(`{"k":"v"}`), now, now,
	)
}

func TestExternalStudentRepositoryPg_Create_Success(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	s := entities.NewExternalStudent("ext1", "C1")
	s.RawData = `{"test":true}`

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO external_students")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	require.NoError(t, repo.Create(context.Background(), s))
	assert.Equal(t, int64(1), s.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_Create_NoRawData(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	s := entities.NewExternalStudent("ext1", "C1")

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO external_students")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	require.NoError(t, repo.Create(context.Background(), s))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_Create_Error(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	s := entities.NewExternalStudent("ext1", "C1")

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO external_students")).WillReturnError(sql.ErrConnDone)
	require.Error(t, repo.Create(context.Background(), s))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_Update_Success(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	s := entities.NewExternalStudent("ext1", "C1")
	s.ID = 5
	s.RawData = testRawDataJSON

	mock.ExpectExec(regexp.QuoteMeta("UPDATE external_students SET")).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.Update(context.Background(), s))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_Update_NoRawData(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	s := entities.NewExternalStudent("ext1", "C1")
	s.ID = 5

	mock.ExpectExec(regexp.QuoteMeta("UPDATE external_students")).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.Update(context.Background(), s))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_Update_NotFound(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	s := entities.NewExternalStudent("ext1", "C1")
	s.ID = 999
	mock.ExpectExec(regexp.QuoteMeta("UPDATE external_students")).WillReturnResult(sqlmock.NewResult(0, 0))
	require.ErrorIs(t, repo.Update(context.Background(), s), sql.ErrNoRows)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_Update_Error(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	s := entities.NewExternalStudent("ext1", "C1")
	s.ID = 5
	mock.ExpectExec(regexp.QuoteMeta("UPDATE external_students")).WillReturnError(sql.ErrConnDone)
	require.Error(t, repo.Update(context.Background(), s))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_Upsert_Success(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	s := entities.NewExternalStudent("ext1", "C1")
	s.RawData = `{"d":1}`
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO external_students")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
	require.NoError(t, repo.Upsert(context.Background(), s))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_Upsert_Error(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	s := entities.NewExternalStudent("ext1", "C1")
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO external_students")).WillReturnError(sql.ErrConnDone)
	require.Error(t, repo.Upsert(context.Background(), s))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_GetByID_Success(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	rows := addStudRow(newStudRows(), 1, "ext1")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnRows(rows)

	s, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	require.NotNil(t, s)
	assert.Equal(t, "Jane", s.FirstName)
	assert.Equal(t, 2, s.Course)
	assert.NotNil(t, s.EnrollmentDate)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_GetByID_NotFound(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnError(sql.ErrNoRows)
	s, err := repo.GetByID(context.Background(), 999)
	require.NoError(t, err)
	assert.Nil(t, s)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_GetByID_Error(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnError(sql.ErrConnDone)
	_, err := repo.GetByID(context.Background(), 1)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_GetByExternalID(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	rows := addStudRow(newStudRows(), 1, "ext1")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnRows(rows)
	s, err := repo.GetByExternalID(context.Background(), "ext1")
	require.NoError(t, err)
	require.NotNil(t, s)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_GetByExternalID_NotFound(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnError(sql.ErrNoRows)
	s, err := repo.GetByExternalID(context.Background(), "nope")
	require.NoError(t, err)
	assert.Nil(t, s)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_GetByExternalID_Error(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnError(sql.ErrConnDone)
	_, err := repo.GetByExternalID(context.Background(), "x")
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_GetByCode(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	rows := addStudRow(newStudRows(), 1, "ext1")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnRows(rows)
	s, err := repo.GetByCode(context.Background(), "C1")
	require.NoError(t, err)
	require.NotNil(t, s)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_GetByCode_NotFound(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnError(sql.ErrNoRows)
	s, err := repo.GetByCode(context.Background(), "nope")
	require.NoError(t, err)
	assert.Nil(t, s)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_GetByCode_Error(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnError(sql.ErrConnDone)
	_, err := repo.GetByCode(context.Background(), "x")
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_GetByLocalUserID(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	rows := addStudRow(newStudRows(), 1, "ext1")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnRows(rows)
	s, err := repo.GetByLocalUserID(context.Background(), 100)
	require.NoError(t, err)
	require.NotNil(t, s)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_GetByLocalUserID_NotFound(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnError(sql.ErrNoRows)
	s, err := repo.GetByLocalUserID(context.Background(), 999)
	require.NoError(t, err)
	assert.Nil(t, s)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_GetByLocalUserID_Error(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnError(sql.ErrConnDone)
	_, err := repo.GetByLocalUserID(context.Background(), 100)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_List_AllFilters(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	isActive := true
	isLinked := true
	course := 2
	filter := entities.ExternalStudentFilter{
		Search: "Jane", GroupName: "GR-01", Faculty: "CS",
		Course: &course, Status: "enrolled",
		IsActive: &isActive, IsLinked: &isLinked,
		Limit: 5,
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	rows := addStudRow(newStudRows(), 1, "ext1")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnRows(rows)

	students, _, err := repo.List(context.Background(), filter)
	require.NoError(t, err)
	assert.Len(t, students, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_List_IsLinkedFalse(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	isLinked := false
	filter := entities.ExternalStudentFilter{IsLinked: &isLinked}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnRows(newStudRows())

	_, _, err := repo.List(context.Background(), filter)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_List_CountError(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).WillReturnError(sql.ErrConnDone)
	_, _, err := repo.List(context.Background(), entities.ExternalStudentFilter{})
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_List_QueryError(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnError(sql.ErrConnDone)
	_, _, err := repo.List(context.Background(), entities.ExternalStudentFilter{})
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_GetUnlinked(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnRows(newStudRows())
	_, _, err := repo.GetUnlinked(context.Background(), 10, 0)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_GetByGroup(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	rows := addStudRow(newStudRows(), 1, "ext1")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnRows(rows)
	students, err := repo.GetByGroup(context.Background(), "GR-01")
	require.NoError(t, err)
	assert.Len(t, students, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_GetByGroup_Error(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnError(sql.ErrConnDone)
	_, err := repo.GetByGroup(context.Background(), "GR-01")
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_GetByFaculty(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	rows := addStudRow(newStudRows(), 1, "ext1")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnRows(rows)
	students, err := repo.GetByFaculty(context.Background(), "CS")
	require.NoError(t, err)
	assert.Len(t, students, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_GetByFaculty_Error(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnError(sql.ErrConnDone)
	_, err := repo.GetByFaculty(context.Background(), "CS")
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_LinkToLocalUser(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE external_students SET local_user_id")).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.LinkToLocalUser(context.Background(), 1, 100))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_LinkToLocalUser_NotFound(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE external_students")).WillReturnResult(sqlmock.NewResult(0, 0))
	require.ErrorIs(t, repo.LinkToLocalUser(context.Background(), 999, 100), sql.ErrNoRows)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_LinkToLocalUser_Error(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE external_students")).WillReturnError(sql.ErrConnDone)
	require.Error(t, repo.LinkToLocalUser(context.Background(), 1, 100))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_Unlink(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE external_students SET local_user_id = NULL")).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.Unlink(context.Background(), 1))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_Unlink_NotFound(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE external_students")).WillReturnResult(sqlmock.NewResult(0, 0))
	require.ErrorIs(t, repo.Unlink(context.Background(), 999), sql.ErrNoRows)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_Unlink_Error(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE external_students")).WillReturnError(sql.ErrConnDone)
	require.Error(t, repo.Unlink(context.Background(), 1))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_Delete(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM external_students")).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.Delete(context.Background(), 1))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_Delete_NotFound(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM external_students")).WillReturnResult(sqlmock.NewResult(0, 0))
	require.ErrorIs(t, repo.Delete(context.Background(), 999), sql.ErrNoRows)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_Delete_Error(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM external_students")).WillReturnError(sql.ErrConnDone)
	require.Error(t, repo.Delete(context.Background(), 1))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_GetAllExternalIDs(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT external_id FROM external_students")).
		WillReturnRows(sqlmock.NewRows([]string{"external_id"}).AddRow("e1").AddRow("e2"))
	ids, err := repo.GetAllExternalIDs(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []string{"e1", "e2"}, ids)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_GetAllExternalIDs_Error(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT external_id")).WillReturnError(sql.ErrConnDone)
	_, err := repo.GetAllExternalIDs(context.Background())
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_BulkUpsert_Empty(t *testing.T) {
	repo, _ := newStudRepoMock(t)
	require.NoError(t, repo.BulkUpsert(context.Background(), nil))
}

func TestExternalStudentRepositoryPg_BulkUpsert_Success(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	s1 := entities.NewExternalStudent("ext1", "C1")
	s1.RawData = testRawDataJSON
	s2 := entities.NewExternalStudent("ext2", "C2")

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO external_students")).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO external_students")).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	require.NoError(t, repo.BulkUpsert(context.Background(), []*entities.ExternalStudent{s1, s2}))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_BulkUpsert_BeginError(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	mock.ExpectBegin().WillReturnError(sql.ErrConnDone)
	require.Error(t, repo.BulkUpsert(context.Background(), []*entities.ExternalStudent{entities.NewExternalStudent("e", "c")}))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_BulkUpsert_ExecError(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO external_students")).WillReturnError(sql.ErrConnDone)
	mock.ExpectRollback()
	require.Error(t, repo.BulkUpsert(context.Background(), []*entities.ExternalStudent{entities.NewExternalStudent("e", "c")}))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_BulkUpsert_CommitError(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO external_students")).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit().WillReturnError(sql.ErrConnDone)
	require.Error(t, repo.BulkUpsert(context.Background(), []*entities.ExternalStudent{entities.NewExternalStudent("e", "c")}))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_MarkInactiveExcept_Empty(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE external_students SET is_active = false")).WillReturnResult(sqlmock.NewResult(0, 5))
	require.NoError(t, repo.MarkInactiveExcept(context.Background(), nil))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_MarkInactiveExcept_WithIDs(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE external_students")).WillReturnResult(sqlmock.NewResult(0, 3))
	require.NoError(t, repo.MarkInactiveExcept(context.Background(), []string{"e1"}))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_MarkInactiveExcept_Error(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE external_students")).WillReturnError(sql.ErrConnDone)
	require.Error(t, repo.MarkInactiveExcept(context.Background(), []string{"e1"}))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_MarkInactiveExcept_EmptyError(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE external_students")).WillReturnError(sql.ErrConnDone)
	require.Error(t, repo.MarkInactiveExcept(context.Background(), nil))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_GetGroups(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT DISTINCT group_name")).
		WillReturnRows(sqlmock.NewRows([]string{"group_name"}).AddRow("GR-01").AddRow("GR-02"))
	groups, err := repo.GetGroups(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []string{"GR-01", "GR-02"}, groups)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_GetGroups_Error(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT DISTINCT group_name")).WillReturnError(sql.ErrConnDone)
	_, err := repo.GetGroups(context.Background())
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_GetFaculties(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT DISTINCT faculty")).
		WillReturnRows(sqlmock.NewRows([]string{"faculty"}).AddRow("CS").AddRow("Math"))
	facs, err := repo.GetFaculties(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []string{"CS", "Math"}, facs)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestExternalStudentRepositoryPg_GetFaculties_Error(t *testing.T) {
	repo, mock := newStudRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT DISTINCT faculty")).WillReturnError(sql.ErrConnDone)
	_, err := repo.GetFaculties(context.Background())
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestNullInt32(t *testing.T) {
	ni := nullInt32(5)
	assert.True(t, ni.Valid)
	assert.Equal(t, int32(5), ni.Int32)

	ni = nullInt32(0)
	assert.False(t, ni.Valid)
}

// ===================== SyncLog Repository Tests =====================

func newSyncLogRepoMock(t *testing.T) (*SyncLogRepositoryPg, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	repo := NewSyncLogRepositoryPg(db)
	return repo.(*SyncLogRepositoryPg), mock
}

var syncLogCols = []string{
	"id", "entity_type", "direction", "status", "started_at", "completed_at",
	"total_records", "processed_count", "success_count", "error_count",
	"conflict_count", "error_message", "metadata", "created_at", "updated_at",
}

func newSyncLogRows() *sqlmock.Rows { return sqlmock.NewRows(syncLogCols) }

func addSyncLogRow(rows *sqlmock.Rows, id int64) *sqlmock.Rows {
	now := time.Now()
	meta, _ := json.Marshal(map[string]string{"k": "v"})
	return rows.AddRow(
		id, "employee", "import", "completed", now, now,
		100, 100, 95, 5, 2, "some error", meta, now, now,
	)
}

func TestSyncLogRepositoryPg_Create_Success(t *testing.T) {
	repo, mock := newSyncLogRepoMock(t)
	l := entities.NewSyncLog(entities.SyncEntityEmployee, entities.SyncDirectionImport)

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO sync_logs")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
	require.NoError(t, repo.Create(context.Background(), l))
	assert.Equal(t, int64(1), l.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncLogRepositoryPg_Create_Error(t *testing.T) {
	repo, mock := newSyncLogRepoMock(t)
	l := entities.NewSyncLog(entities.SyncEntityEmployee, entities.SyncDirectionImport)
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO sync_logs")).WillReturnError(sql.ErrConnDone)
	require.Error(t, repo.Create(context.Background(), l))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncLogRepositoryPg_Update_Success(t *testing.T) {
	repo, mock := newSyncLogRepoMock(t)
	l := entities.NewSyncLog(entities.SyncEntityEmployee, entities.SyncDirectionImport)
	l.ID = 5
	mock.ExpectExec(regexp.QuoteMeta("UPDATE sync_logs SET")).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.Update(context.Background(), l))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncLogRepositoryPg_Update_NotFound(t *testing.T) {
	repo, mock := newSyncLogRepoMock(t)
	l := entities.NewSyncLog(entities.SyncEntityEmployee, entities.SyncDirectionImport)
	l.ID = 999
	mock.ExpectExec(regexp.QuoteMeta("UPDATE sync_logs")).WillReturnResult(sqlmock.NewResult(0, 0))
	require.ErrorIs(t, repo.Update(context.Background(), l), sql.ErrNoRows)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncLogRepositoryPg_Update_Error(t *testing.T) {
	repo, mock := newSyncLogRepoMock(t)
	l := entities.NewSyncLog(entities.SyncEntityEmployee, entities.SyncDirectionImport)
	l.ID = 5
	mock.ExpectExec(regexp.QuoteMeta("UPDATE sync_logs")).WillReturnError(sql.ErrConnDone)
	require.Error(t, repo.Update(context.Background(), l))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncLogRepositoryPg_GetByID_Success(t *testing.T) {
	repo, mock := newSyncLogRepoMock(t)
	rows := addSyncLogRow(newSyncLogRows(), 1)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, entity_type")).WillReturnRows(rows)
	l, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	require.NotNil(t, l)
	assert.Equal(t, entities.SyncEntityEmployee, l.EntityType)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncLogRepositoryPg_GetByID_NotFound(t *testing.T) {
	repo, mock := newSyncLogRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, entity_type")).WillReturnError(sql.ErrNoRows)
	l, err := repo.GetByID(context.Background(), 999)
	require.NoError(t, err)
	assert.Nil(t, l)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncLogRepositoryPg_GetByID_Error(t *testing.T) {
	repo, mock := newSyncLogRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, entity_type")).WillReturnError(sql.ErrConnDone)
	_, err := repo.GetByID(context.Background(), 1)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncLogRepositoryPg_GetByID_EmptyMetadata(t *testing.T) {
	repo, mock := newSyncLogRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows(syncLogCols).AddRow(
		int64(1), "employee", "import", "completed", now, nil,
		0, 0, 0, 0, 0, nil, nil, now, now,
	)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, entity_type")).WillReturnRows(rows)
	l, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	require.NotNil(t, l)
	assert.NotNil(t, l.Metadata) // should be empty map
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncLogRepositoryPg_List_AllFilters(t *testing.T) {
	repo, mock := newSyncLogRepoMock(t)
	et := entities.SyncEntityEmployee
	dir := entities.SyncDirectionImport
	status := entities.SyncStatusCompleted
	now := time.Now()
	filter := entities.SyncLogFilter{
		EntityType: &et, Direction: &dir, Status: &status,
		StartDate: &now, EndDate: &now, Limit: 5,
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	rows := addSyncLogRow(newSyncLogRows(), 1)
	mock.ExpectQuery("SELECT id, entity_type").WillReturnRows(rows)

	logs, total, err := repo.List(context.Background(), filter)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, logs, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncLogRepositoryPg_List_CountError(t *testing.T) {
	repo, mock := newSyncLogRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).WillReturnError(sql.ErrConnDone)
	_, _, err := repo.List(context.Background(), entities.SyncLogFilter{})
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncLogRepositoryPg_List_QueryError(t *testing.T) {
	repo, mock := newSyncLogRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery("SELECT id, entity_type").WillReturnError(sql.ErrConnDone)
	_, _, err := repo.List(context.Background(), entities.SyncLogFilter{})
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncLogRepositoryPg_GetLatest(t *testing.T) {
	repo, mock := newSyncLogRepoMock(t)
	rows := addSyncLogRow(newSyncLogRows(), 1)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, entity_type")).WillReturnRows(rows)
	l, err := repo.GetLatest(context.Background(), entities.SyncEntityEmployee)
	require.NoError(t, err)
	require.NotNil(t, l)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncLogRepositoryPg_GetLatest_NotFound(t *testing.T) {
	repo, mock := newSyncLogRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, entity_type")).WillReturnError(sql.ErrNoRows)
	l, err := repo.GetLatest(context.Background(), entities.SyncEntityEmployee)
	require.NoError(t, err)
	assert.Nil(t, l)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncLogRepositoryPg_GetLatest_Error(t *testing.T) {
	repo, mock := newSyncLogRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, entity_type")).WillReturnError(sql.ErrConnDone)
	_, err := repo.GetLatest(context.Background(), entities.SyncEntityEmployee)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncLogRepositoryPg_GetRunning(t *testing.T) {
	repo, mock := newSyncLogRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, entity_type")).WillReturnRows(newSyncLogRows())
	logs, err := repo.GetRunning(context.Background())
	require.NoError(t, err)
	assert.Empty(t, logs)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncLogRepositoryPg_GetRunning_Error(t *testing.T) {
	repo, mock := newSyncLogRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, entity_type")).WillReturnError(sql.ErrConnDone)
	_, err := repo.GetRunning(context.Background())
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncLogRepositoryPg_GetStats(t *testing.T) {
	repo, mock := newSyncLogRepoMock(t)
	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
		WillReturnRows(sqlmock.NewRows([]string{"total_syncs", "successful_syncs", "failed_syncs", "total_records", "total_conflicts", "last_sync_at"}).
			AddRow(int64(10), int64(8), int64(2), int64(1000), int64(5), now))

	stats, err := repo.GetStats(context.Background(), nil)
	require.NoError(t, err)
	require.NotNil(t, stats)
	assert.Equal(t, int64(10), stats.TotalSyncs)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncLogRepositoryPg_GetStats_WithEntityType(t *testing.T) {
	repo, mock := newSyncLogRepoMock(t)
	et := entities.SyncEntityEmployee
	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
		WillReturnRows(sqlmock.NewRows([]string{"total_syncs", "successful_syncs", "failed_syncs", "total_records", "total_conflicts", "last_sync_at"}).
			AddRow(int64(5), int64(4), int64(1), int64(500), int64(2), now))

	stats, err := repo.GetStats(context.Background(), &et)
	require.NoError(t, err)
	require.NotNil(t, stats)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncLogRepositoryPg_GetStats_Error(t *testing.T) {
	repo, mock := newSyncLogRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnError(sql.ErrConnDone)
	_, err := repo.GetStats(context.Background(), nil)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncLogRepositoryPg_Delete(t *testing.T) {
	repo, mock := newSyncLogRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM sync_logs")).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.Delete(context.Background(), 1))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncLogRepositoryPg_Delete_NotFound(t *testing.T) {
	repo, mock := newSyncLogRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM sync_logs")).WillReturnResult(sqlmock.NewResult(0, 0))
	require.ErrorIs(t, repo.Delete(context.Background(), 999), sql.ErrNoRows)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncLogRepositoryPg_Delete_Error(t *testing.T) {
	repo, mock := newSyncLogRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM sync_logs")).WillReturnError(sql.ErrConnDone)
	require.Error(t, repo.Delete(context.Background(), 1))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncLogRepositoryPg_DeleteOlderThan(t *testing.T) {
	repo, mock := newSyncLogRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM sync_logs WHERE created_at")).WillReturnResult(sqlmock.NewResult(0, 3))
	count, err := repo.DeleteOlderThan(context.Background(), 30)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncLogRepositoryPg_DeleteOlderThan_Error(t *testing.T) {
	repo, mock := newSyncLogRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM sync_logs")).WillReturnError(sql.ErrConnDone)
	_, err := repo.DeleteOlderThan(context.Background(), 30)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// ===================== SyncConflict Repository Tests =====================

func newConflictRepoMock(t *testing.T) (*SyncConflictRepositoryPg, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	repo := NewSyncConflictRepositoryPg(db)
	return repo.(*SyncConflictRepositoryPg), mock
}

var conflictCols = []string{
	"id", "sync_log_id", "entity_type", "entity_id", "local_data", "external_data",
	"conflict_type", "conflict_fields", "resolution", "resolved_by", "resolved_at",
	"resolved_data", "notes", "created_at", "updated_at",
}

func newConflictRows() *sqlmock.Rows { return sqlmock.NewRows(conflictCols) }

func addConflictRow(rows *sqlmock.Rows, id int64) *sqlmock.Rows {
	now := time.Now()
	resolvedBy := int64(10)
	return rows.AddRow(
		id, int64(1), "employee", "ext1", []byte(`{"local":true}`), []byte(`{"external":true}`),
		"update", pq.StringArray{"name", "email"}, "pending", resolvedBy, now,
		[]byte(`{"resolved":true}`), "some notes", now, now,
	)
}

func TestSyncConflictRepositoryPg_Create_Success(t *testing.T) {
	repo, mock := newConflictRepoMock(t)
	c := entities.NewSyncConflict(1, entities.SyncEntityEmployee, "ext1")
	c.LocalData = `{"local":true}`
	c.ExternalData = `{"external":true}`

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO sync_conflicts")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
	require.NoError(t, repo.Create(context.Background(), c))
	assert.Equal(t, int64(1), c.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncConflictRepositoryPg_Create_Error(t *testing.T) {
	repo, mock := newConflictRepoMock(t)
	c := entities.NewSyncConflict(1, entities.SyncEntityEmployee, "ext1")
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO sync_conflicts")).WillReturnError(sql.ErrConnDone)
	require.Error(t, repo.Create(context.Background(), c))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncConflictRepositoryPg_Update_Success(t *testing.T) {
	repo, mock := newConflictRepoMock(t)
	c := entities.NewSyncConflict(1, entities.SyncEntityEmployee, "ext1")
	c.ID = 5
	mock.ExpectExec(regexp.QuoteMeta("UPDATE sync_conflicts SET")).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.Update(context.Background(), c))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncConflictRepositoryPg_Update_NotFound(t *testing.T) {
	repo, mock := newConflictRepoMock(t)
	c := entities.NewSyncConflict(1, entities.SyncEntityEmployee, "ext1")
	c.ID = 999
	mock.ExpectExec(regexp.QuoteMeta("UPDATE sync_conflicts")).WillReturnResult(sqlmock.NewResult(0, 0))
	require.ErrorIs(t, repo.Update(context.Background(), c), sql.ErrNoRows)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncConflictRepositoryPg_Update_Error(t *testing.T) {
	repo, mock := newConflictRepoMock(t)
	c := entities.NewSyncConflict(1, entities.SyncEntityEmployee, "ext1")
	c.ID = 5
	mock.ExpectExec(regexp.QuoteMeta("UPDATE sync_conflicts")).WillReturnError(sql.ErrConnDone)
	require.Error(t, repo.Update(context.Background(), c))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncConflictRepositoryPg_GetByID_Success(t *testing.T) {
	repo, mock := newConflictRepoMock(t)
	rows := addConflictRow(newConflictRows(), 1)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnRows(rows)
	c, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	require.NotNil(t, c)
	assert.Equal(t, "some notes", c.Notes)
	assert.NotNil(t, c.ResolvedBy)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncConflictRepositoryPg_GetByID_NotFound(t *testing.T) {
	repo, mock := newConflictRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnError(sql.ErrNoRows)
	c, err := repo.GetByID(context.Background(), 999)
	require.NoError(t, err)
	assert.Nil(t, c)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncConflictRepositoryPg_GetByID_Error(t *testing.T) {
	repo, mock := newConflictRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnError(sql.ErrConnDone)
	_, err := repo.GetByID(context.Background(), 1)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncConflictRepositoryPg_List_AllFilters(t *testing.T) {
	repo, mock := newConflictRepoMock(t)
	syncLogID := int64(1)
	et := entities.SyncEntityEmployee
	res := entities.ConflictResolutionPending
	filter := entities.SyncConflictFilter{
		SyncLogID: &syncLogID, EntityType: &et, Resolution: &res, Limit: 5,
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	rows := addConflictRow(newConflictRows(), 1)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnRows(rows)

	conflicts, total, err := repo.List(context.Background(), filter)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, conflicts, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncConflictRepositoryPg_List_CountError(t *testing.T) {
	repo, mock := newConflictRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).WillReturnError(sql.ErrConnDone)
	_, _, err := repo.List(context.Background(), entities.SyncConflictFilter{})
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncConflictRepositoryPg_List_QueryError(t *testing.T) {
	repo, mock := newConflictRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnError(sql.ErrConnDone)
	_, _, err := repo.List(context.Background(), entities.SyncConflictFilter{})
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncConflictRepositoryPg_GetBySyncLogID(t *testing.T) {
	repo, mock := newConflictRepoMock(t)
	rows := addConflictRow(newConflictRows(), 1)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnRows(rows)
	conflicts, err := repo.GetBySyncLogID(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, conflicts, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncConflictRepositoryPg_GetBySyncLogID_Error(t *testing.T) {
	repo, mock := newConflictRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnError(sql.ErrConnDone)
	_, err := repo.GetBySyncLogID(context.Background(), 1)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncConflictRepositoryPg_GetPending(t *testing.T) {
	repo, mock := newConflictRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnRows(newConflictRows())
	_, _, err := repo.GetPending(context.Background(), 10, 0)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncConflictRepositoryPg_GetPendingByEntityType(t *testing.T) {
	repo, mock := newConflictRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnRows(newConflictRows())
	conflicts, err := repo.GetPendingByEntityType(context.Background(), entities.SyncEntityEmployee)
	require.NoError(t, err)
	assert.Empty(t, conflicts)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncConflictRepositoryPg_GetPendingByEntityType_Error(t *testing.T) {
	repo, mock := newConflictRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnError(sql.ErrConnDone)
	_, err := repo.GetPendingByEntityType(context.Background(), entities.SyncEntityEmployee)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncConflictRepositoryPg_Resolve_Success(t *testing.T) {
	repo, mock := newConflictRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE sync_conflicts SET")).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.Resolve(context.Background(), 1, entities.ConflictResolutionUseLocal, 10, `{"merged":true}`))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncConflictRepositoryPg_Resolve_NotFound(t *testing.T) {
	repo, mock := newConflictRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE sync_conflicts")).WillReturnResult(sqlmock.NewResult(0, 0))
	require.ErrorIs(t, repo.Resolve(context.Background(), 999, entities.ConflictResolutionUseLocal, 10, ""), sql.ErrNoRows)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncConflictRepositoryPg_Resolve_Error(t *testing.T) {
	repo, mock := newConflictRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE sync_conflicts")).WillReturnError(sql.ErrConnDone)
	require.Error(t, repo.Resolve(context.Background(), 1, entities.ConflictResolutionUseLocal, 10, ""))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncConflictRepositoryPg_BulkResolve_Empty(t *testing.T) {
	repo, _ := newConflictRepoMock(t)
	require.NoError(t, repo.BulkResolve(context.Background(), nil, entities.ConflictResolutionSkip, 10))
}

func TestSyncConflictRepositoryPg_BulkResolve_Success(t *testing.T) {
	repo, mock := newConflictRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE sync_conflicts SET")).WillReturnResult(sqlmock.NewResult(0, 2))
	require.NoError(t, repo.BulkResolve(context.Background(), []int64{1, 2}, entities.ConflictResolutionSkip, 10))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncConflictRepositoryPg_BulkResolve_Error(t *testing.T) {
	repo, mock := newConflictRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE sync_conflicts")).WillReturnError(sql.ErrConnDone)
	require.Error(t, repo.BulkResolve(context.Background(), []int64{1}, entities.ConflictResolutionSkip, 10))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncConflictRepositoryPg_Delete(t *testing.T) {
	repo, mock := newConflictRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM sync_conflicts WHERE id")).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.Delete(context.Background(), 1))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncConflictRepositoryPg_Delete_NotFound(t *testing.T) {
	repo, mock := newConflictRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM sync_conflicts WHERE id")).WillReturnResult(sqlmock.NewResult(0, 0))
	require.ErrorIs(t, repo.Delete(context.Background(), 999), sql.ErrNoRows)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncConflictRepositoryPg_Delete_Error(t *testing.T) {
	repo, mock := newConflictRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM sync_conflicts")).WillReturnError(sql.ErrConnDone)
	require.Error(t, repo.Delete(context.Background(), 1))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncConflictRepositoryPg_DeleteBySyncLogID(t *testing.T) {
	repo, mock := newConflictRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM sync_conflicts WHERE sync_log_id")).WillReturnResult(sqlmock.NewResult(0, 3))
	require.NoError(t, repo.DeleteBySyncLogID(context.Background(), 1))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncConflictRepositoryPg_DeleteBySyncLogID_Error(t *testing.T) {
	repo, mock := newConflictRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM sync_conflicts")).WillReturnError(sql.ErrConnDone)
	require.Error(t, repo.DeleteBySyncLogID(context.Background(), 1))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncConflictRepositoryPg_GetStats_Success(t *testing.T) {
	repo, mock := newConflictRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
		WillReturnRows(sqlmock.NewRows([]string{"total_conflicts", "pending_conflicts", "resolved_conflicts"}).
			AddRow(int64(10), int64(3), int64(7)))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT entity_type, COUNT(*)")).
		WillReturnRows(sqlmock.NewRows([]string{"entity_type", "count"}).AddRow("employee", int64(3)))

	stats, err := repo.GetStats(context.Background())
	require.NoError(t, err)
	require.NotNil(t, stats)
	assert.Equal(t, int64(10), stats.TotalConflicts)
	assert.Equal(t, int64(3), stats.PendingConflicts)
	assert.Equal(t, int64(3), stats.ByEntityType[entities.SyncEntityEmployee])
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncConflictRepositoryPg_GetStats_OverallError(t *testing.T) {
	repo, mock := newConflictRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).WillReturnError(sql.ErrConnDone)
	_, err := repo.GetStats(context.Background())
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncConflictRepositoryPg_GetStats_TypeQueryError(t *testing.T) {
	repo, mock := newConflictRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
		WillReturnRows(sqlmock.NewRows([]string{"total_conflicts", "pending_conflicts", "resolved_conflicts"}).
			AddRow(int64(10), int64(3), int64(7)))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT entity_type")).WillReturnError(sql.ErrConnDone)
	_, err := repo.GetStats(context.Background())
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSyncConflictRepositoryPg_GetStats_TypeScanError(t *testing.T) {
	repo, mock := newConflictRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT")).
		WillReturnRows(sqlmock.NewRows([]string{"total_conflicts", "pending_conflicts", "resolved_conflicts"}).
			AddRow(int64(10), int64(3), int64(7)))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT entity_type")).
		WillReturnRows(sqlmock.NewRows([]string{"entity_type", "count"}).AddRow("emp", "not-a-number"))
	_, err := repo.GetStats(context.Background())
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- nullBytes helper ---

func TestNullBytes(t *testing.T) {
	assert.Nil(t, nullBytes(""))

	// Valid JSON
	result := nullBytes(`{"key":"value"}`)
	assert.Equal(t, `{"key":"value"}`, string(result))

	// Non-JSON string
	result = nullBytes("plain text")
	assert.Equal(t, `"plain text"`, string(result))
}
