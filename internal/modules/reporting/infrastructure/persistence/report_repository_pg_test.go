package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain/repositories"
)

func newReportRepoMock(t *testing.T) (*ReportRepositoryPG, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewReportRepositoryPG(db), mock
}

var reportCols = []string{
	"id", "report_type_id", "title", "description", "period_start", "period_end",
	"author_id", "status", "file_name", "file_path", "file_size", "mime_type",
	"parameters", "data", "reviewer_comment", "reviewed_by", "reviewed_at",
	"published_at", "is_public", "created_at", "updated_at",
}

func newReportRows() *sqlmock.Rows { return sqlmock.NewRows(reportCols) }

func addReportRow(rows *sqlmock.Rows, id int64, title string) *sqlmock.Rows {
	now := time.Now()
	return rows.AddRow(
		id, int64(1), title, nil, nil, nil,
		int64(10), domain.ReportStatusDraft, nil, nil, nil, nil,
		json.RawMessage(`{}`), json.RawMessage(`{}`), nil, nil, nil,
		nil, false, now, now,
	)
}

// --- Create ---

func TestReportRepositoryPG_Create_Success(t *testing.T) {
	repo, mock := newReportRepoMock(t)
	r := entities.NewReport(1, "Test Report", 10)

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO reports")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	err := repo.Create(context.Background(), r)
	require.NoError(t, err)
	assert.Equal(t, int64(1), r.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_Create_Error(t *testing.T) {
	repo, mock := newReportRepoMock(t)
	r := entities.NewReport(1, "Test", 10)

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO reports")).
		WillReturnError(sql.ErrConnDone)

	err := repo.Create(context.Background(), r)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- Save ---

func TestReportRepositoryPG_Save_Success(t *testing.T) {
	repo, mock := newReportRepoMock(t)
	r := entities.NewReport(1, "Updated", 10)
	r.ID = 5

	mock.ExpectExec(regexp.QuoteMeta("UPDATE reports SET")).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Save(context.Background(), r)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_Save_NotFound(t *testing.T) {
	repo, mock := newReportRepoMock(t)
	r := entities.NewReport(1, "X", 10)
	r.ID = 999

	mock.ExpectExec(regexp.QuoteMeta("UPDATE reports SET")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Save(context.Background(), r)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_Save_Error(t *testing.T) {
	repo, mock := newReportRepoMock(t)
	r := entities.NewReport(1, "X", 10)
	r.ID = 5

	mock.ExpectExec(regexp.QuoteMeta("UPDATE reports SET")).
		WillReturnError(sql.ErrConnDone)

	err := repo.Save(context.Background(), r)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_Save_RowsAffectedError(t *testing.T) {
	repo, mock := newReportRepoMock(t)
	r := entities.NewReport(1, "X", 10)
	r.ID = 5

	mock.ExpectExec(regexp.QuoteMeta("UPDATE reports SET")).
		WillReturnResult(sqlmock.NewErrorResult(sql.ErrConnDone))

	err := repo.Save(context.Background(), r)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetByID ---

func TestReportRepositoryPG_GetByID_Success(t *testing.T) {
	repo, mock := newReportRepoMock(t)
	rows := addReportRow(newReportRows(), 1, "Report1")

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, report_type_id")).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	r, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "Report1", r.Title)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_GetByID_NotFound(t *testing.T) {
	repo, mock := newReportRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, report_type_id")).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	r, err := repo.GetByID(context.Background(), 999)
	require.NoError(t, err)
	assert.Nil(t, r)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_GetByID_Error(t *testing.T) {
	repo, mock := newReportRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, report_type_id")).
		WillReturnError(sql.ErrConnDone)

	r, err := repo.GetByID(context.Background(), 1)
	require.Error(t, err)
	assert.Nil(t, r)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- Delete ---

func TestReportRepositoryPG_Delete_Success(t *testing.T) {
	repo, mock := newReportRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM reports WHERE id = $1")).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), 1)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_Delete_NotFound(t *testing.T) {
	repo, mock := newReportRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM reports")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Delete(context.Background(), 999)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_Delete_Error(t *testing.T) {
	repo, mock := newReportRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM reports")).
		WillReturnError(sql.ErrConnDone)

	err := repo.Delete(context.Background(), 1)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_Delete_RowsAffectedError(t *testing.T) {
	repo, mock := newReportRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM reports")).
		WillReturnResult(sqlmock.NewErrorResult(sql.ErrConnDone))

	err := repo.Delete(context.Background(), 1)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- List ---

func TestReportRepositoryPG_List_NoFilter(t *testing.T) {
	repo, mock := newReportRepoMock(t)
	rows := addReportRow(newReportRows(), 1, "R1")

	mock.ExpectQuery("SELECT id, report_type_id").WillReturnRows(rows)

	reports, err := repo.List(context.Background(), repositories.ReportFilter{}, 10, 0)
	require.NoError(t, err)
	assert.Len(t, reports, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_List_AllFilters(t *testing.T) {
	repo, mock := newReportRepoMock(t)
	rtID := int64(1)
	authorID := int64(2)
	status := domain.ReportStatusDraft
	isPublic := true
	now := time.Now()
	search := "test"
	filter := repositories.ReportFilter{
		ReportTypeID: &rtID,
		AuthorID:     &authorID,
		Status:       &status,
		IsPublic:     &isPublic,
		PeriodStart:  &now,
		PeriodEnd:    &now,
		Search:       &search,
	}

	rows := addReportRow(newReportRows(), 1, "test")
	mock.ExpectQuery("SELECT id, report_type_id").WillReturnRows(rows)

	reports, err := repo.List(context.Background(), filter, 10, 0)
	require.NoError(t, err)
	assert.Len(t, reports, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_List_NoLimit(t *testing.T) {
	repo, mock := newReportRepoMock(t)
	rows := addReportRow(newReportRows(), 1, "R1")

	mock.ExpectQuery("SELECT id, report_type_id").WillReturnRows(rows)

	reports, err := repo.List(context.Background(), repositories.ReportFilter{}, 0, 0)
	require.NoError(t, err)
	assert.Len(t, reports, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_List_Error(t *testing.T) {
	repo, mock := newReportRepoMock(t)

	mock.ExpectQuery("SELECT id, report_type_id").WillReturnError(sql.ErrConnDone)

	_, err := repo.List(context.Background(), repositories.ReportFilter{}, 10, 0)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_List_ScanError(t *testing.T) {
	repo, mock := newReportRepoMock(t)
	rows := sqlmock.NewRows(reportCols).AddRow(
		"bad", int64(1), "R", nil, nil, nil,
		int64(10), "draft", nil, nil, nil, nil,
		nil, nil, nil, nil, nil, nil, false, time.Now(), time.Now(),
	)

	mock.ExpectQuery("SELECT id, report_type_id").WillReturnRows(rows)

	_, err := repo.List(context.Background(), repositories.ReportFilter{}, 10, 0)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_List_RowsErr(t *testing.T) {
	repo, mock := newReportRepoMock(t)
	rows := addReportRow(newReportRows(), 1, "R1")
	rows.RowError(0, sql.ErrConnDone)

	mock.ExpectQuery("SELECT id, report_type_id").WillReturnRows(rows)

	_, err := repo.List(context.Background(), repositories.ReportFilter{}, 10, 0)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- Count ---

func TestReportRepositoryPG_Count_Success(t *testing.T) {
	repo, mock := newReportRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM reports")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(5)))

	count, err := repo.Count(context.Background(), repositories.ReportFilter{})
	require.NoError(t, err)
	assert.Equal(t, int64(5), count)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_Count_Error(t *testing.T) {
	repo, mock := newReportRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnError(sql.ErrConnDone)

	_, err := repo.Count(context.Background(), repositories.ReportFilter{})
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetByAuthor, GetByStatus, GetByReportType, GetPublicReports ---

func TestReportRepositoryPG_GetByAuthor(t *testing.T) {
	repo, mock := newReportRepoMock(t)
	rows := addReportRow(newReportRows(), 1, "R")
	mock.ExpectQuery("SELECT id, report_type_id").WillReturnRows(rows)

	reports, err := repo.GetByAuthor(context.Background(), 10, 10, 0)
	require.NoError(t, err)
	assert.Len(t, reports, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_GetByStatus(t *testing.T) {
	repo, mock := newReportRepoMock(t)
	rows := addReportRow(newReportRows(), 1, "R")
	mock.ExpectQuery("SELECT id, report_type_id").WillReturnRows(rows)

	reports, err := repo.GetByStatus(context.Background(), domain.ReportStatusDraft, 10, 0)
	require.NoError(t, err)
	assert.Len(t, reports, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_GetByReportType(t *testing.T) {
	repo, mock := newReportRepoMock(t)
	rows := addReportRow(newReportRows(), 1, "R")
	mock.ExpectQuery("SELECT id, report_type_id").WillReturnRows(rows)

	reports, err := repo.GetByReportType(context.Background(), 1, 10, 0)
	require.NoError(t, err)
	assert.Len(t, reports, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_GetPublicReports(t *testing.T) {
	repo, mock := newReportRepoMock(t)
	rows := addReportRow(newReportRows(), 1, "R")
	mock.ExpectQuery("SELECT id, report_type_id").WillReturnRows(rows)

	reports, err := repo.GetPublicReports(context.Background(), 10, 0)
	require.NoError(t, err)
	assert.Len(t, reports, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- AddAccess ---

func TestReportRepositoryPG_AddAccess_Success(t *testing.T) {
	repo, mock := newReportRepoMock(t)
	grantedBy := int64(1)
	access := entities.NewReportAccessForUser(1, 2, domain.ReportPermissionRead, &grantedBy)

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO report_access")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	err := repo.AddAccess(context.Background(), access)
	require.NoError(t, err)
	assert.Equal(t, int64(1), access.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_AddAccess_Error(t *testing.T) {
	repo, mock := newReportRepoMock(t)
	access := &entities.ReportAccess{ReportID: 1, CreatedAt: time.Now()}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO report_access")).
		WillReturnError(sql.ErrConnDone)

	err := repo.AddAccess(context.Background(), access)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- RemoveAccess ---

func TestReportRepositoryPG_RemoveAccess_Success(t *testing.T) {
	repo, mock := newReportRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM report_access")).
		WithArgs(int64(5), int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.RemoveAccess(context.Background(), 1, 5)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_RemoveAccess_NotFound(t *testing.T) {
	repo, mock := newReportRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM report_access")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.RemoveAccess(context.Background(), 1, 999)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_RemoveAccess_Error(t *testing.T) {
	repo, mock := newReportRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM report_access")).
		WillReturnError(sql.ErrConnDone)

	err := repo.RemoveAccess(context.Background(), 1, 5)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_RemoveAccess_RowsAffectedError(t *testing.T) {
	repo, mock := newReportRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM report_access")).
		WillReturnResult(sqlmock.NewErrorResult(sql.ErrConnDone))

	err := repo.RemoveAccess(context.Background(), 1, 5)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetAccessByReport ---

func TestReportRepositoryPG_GetAccessByReport_Success(t *testing.T) {
	repo, mock := newReportRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "report_id", "user_id", "role", "permission", "granted_by", "created_at"}).
		AddRow(int64(1), int64(1), int64(2), nil, domain.ReportPermissionRead, nil, now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, report_id, user_id")).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	accesses, err := repo.GetAccessByReport(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, accesses, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_GetAccessByReport_Error(t *testing.T) {
	repo, mock := newReportRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, report_id")).
		WillReturnError(sql.ErrConnDone)

	_, err := repo.GetAccessByReport(context.Background(), 1)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_GetAccessByReport_ScanError(t *testing.T) {
	repo, mock := newReportRepoMock(t)
	rows := sqlmock.NewRows([]string{"id", "report_id", "user_id", "role", "permission", "granted_by", "created_at"}).
		AddRow("bad", int64(1), nil, nil, "read", nil, time.Now())

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, report_id")).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	_, err := repo.GetAccessByReport(context.Background(), 1)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- HasAccess ---

func TestReportRepositoryPG_HasAccess_DirectAccess(t *testing.T) {
	repo, mock := newReportRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS")).
		WithArgs(int64(1), int64(2), domain.ReportPermissionRead).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	has, err := repo.HasAccess(context.Background(), 1, 2, domain.ReportPermissionRead)
	require.NoError(t, err)
	assert.True(t, has)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_HasAccess_PublicReport(t *testing.T) {
	repo, mock := newReportRepoMock(t)

	// No direct access
	mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS")).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	// Public check for read permission
	mock.ExpectQuery(regexp.QuoteMeta("SELECT is_public FROM reports")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"is_public"}).AddRow(true))

	has, err := repo.HasAccess(context.Background(), 1, 2, domain.ReportPermissionRead)
	require.NoError(t, err)
	assert.True(t, has)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_HasAccess_RoleBased(t *testing.T) {
	repo, mock := newReportRepoMock(t)

	// No direct access
	mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS")).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	// Not public
	mock.ExpectQuery(regexp.QuoteMeta("SELECT is_public")).
		WillReturnRows(sqlmock.NewRows([]string{"is_public"}).AddRow(false))

	// Role-based
	mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS")).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	has, err := repo.HasAccess(context.Background(), 1, 2, domain.ReportPermissionRead)
	require.NoError(t, err)
	assert.True(t, has)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_HasAccess_NoAccess(t *testing.T) {
	repo, mock := newReportRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS")).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	// Write permission - no public check
	mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS")).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	has, err := repo.HasAccess(context.Background(), 1, 2, domain.ReportPermissionWrite)
	require.NoError(t, err)
	assert.False(t, has)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_HasAccess_DirectCheckError(t *testing.T) {
	repo, mock := newReportRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS")).
		WillReturnError(sql.ErrConnDone)

	_, err := repo.HasAccess(context.Background(), 1, 2, domain.ReportPermissionRead)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_HasAccess_PublicCheckError(t *testing.T) {
	repo, mock := newReportRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS")).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT is_public")).
		WillReturnError(sql.ErrConnDone)

	_, err := repo.HasAccess(context.Background(), 1, 2, domain.ReportPermissionRead)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_HasAccess_PublicCheckNoRows(t *testing.T) {
	repo, mock := newReportRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS")).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT is_public")).
		WillReturnError(sql.ErrNoRows)

	// Role check
	mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS")).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	has, err := repo.HasAccess(context.Background(), 1, 2, domain.ReportPermissionRead)
	require.NoError(t, err)
	assert.False(t, has)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_HasAccess_RoleCheckError(t *testing.T) {
	repo, mock := newReportRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS")).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	// Write - no public
	mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS")).
		WillReturnError(sql.ErrConnDone)

	_, err := repo.HasAccess(context.Background(), 1, 2, domain.ReportPermissionWrite)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- AddComment ---

func TestReportRepositoryPG_AddComment_Success(t *testing.T) {
	repo, mock := newReportRepoMock(t)
	c := entities.NewReportComment(1, 2, "content")

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO report_comments")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	err := repo.AddComment(context.Background(), c)
	require.NoError(t, err)
	assert.Equal(t, int64(1), c.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_AddComment_Error(t *testing.T) {
	repo, mock := newReportRepoMock(t)
	c := entities.NewReportComment(1, 2, "content")

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO report_comments")).
		WillReturnError(sql.ErrConnDone)

	err := repo.AddComment(context.Background(), c)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- UpdateComment ---

func TestReportRepositoryPG_UpdateComment_Success(t *testing.T) {
	repo, mock := newReportRepoMock(t)
	c := entities.NewReportComment(1, 2, "updated")
	c.ID = 5

	mock.ExpectExec(regexp.QuoteMeta("UPDATE report_comments SET")).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdateComment(context.Background(), c)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_UpdateComment_NotFound(t *testing.T) {
	repo, mock := newReportRepoMock(t)
	c := &entities.ReportComment{ID: 999}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE report_comments")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.UpdateComment(context.Background(), c)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_UpdateComment_Error(t *testing.T) {
	repo, mock := newReportRepoMock(t)
	c := &entities.ReportComment{ID: 1}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE report_comments")).
		WillReturnError(sql.ErrConnDone)

	err := repo.UpdateComment(context.Background(), c)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_UpdateComment_RowsAffectedError(t *testing.T) {
	repo, mock := newReportRepoMock(t)
	c := &entities.ReportComment{ID: 1}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE report_comments")).
		WillReturnResult(sqlmock.NewErrorResult(sql.ErrConnDone))

	err := repo.UpdateComment(context.Background(), c)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- DeleteComment ---

func TestReportRepositoryPG_DeleteComment_Success(t *testing.T) {
	repo, mock := newReportRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM report_comments")).
		WithArgs(int64(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.DeleteComment(context.Background(), 1)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_DeleteComment_NotFound(t *testing.T) {
	repo, mock := newReportRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM report_comments")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.DeleteComment(context.Background(), 999)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_DeleteComment_Error(t *testing.T) {
	repo, mock := newReportRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM report_comments")).
		WillReturnError(sql.ErrConnDone)

	err := repo.DeleteComment(context.Background(), 1)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_DeleteComment_RowsAffectedError(t *testing.T) {
	repo, mock := newReportRepoMock(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM report_comments")).
		WillReturnResult(sqlmock.NewErrorResult(sql.ErrConnDone))

	err := repo.DeleteComment(context.Background(), 1)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetCommentsByReport ---

func TestReportRepositoryPG_GetCommentsByReport_Success(t *testing.T) {
	repo, mock := newReportRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "report_id", "author_id", "content", "created_at", "updated_at"}).
		AddRow(int64(1), int64(1), int64(2), "comment", now, now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, report_id, author_id")).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	comments, err := repo.GetCommentsByReport(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, comments, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_GetCommentsByReport_Error(t *testing.T) {
	repo, mock := newReportRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, report_id")).
		WillReturnError(sql.ErrConnDone)

	_, err := repo.GetCommentsByReport(context.Background(), 1)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_GetCommentsByReport_ScanError(t *testing.T) {
	repo, mock := newReportRepoMock(t)
	rows := sqlmock.NewRows([]string{"id", "report_id", "author_id", "content", "created_at", "updated_at"}).
		AddRow("bad", int64(1), int64(2), "comment", time.Now(), time.Now())

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, report_id")).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	_, err := repo.GetCommentsByReport(context.Background(), 1)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- AddHistory ---

func TestReportRepositoryPG_AddHistory_Success(t *testing.T) {
	repo, mock := newReportRepoMock(t)
	userID := int64(2)
	h := entities.NewReportHistory(1, &userID, entities.ReportActionCreated)

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO report_history")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	err := repo.AddHistory(context.Background(), h)
	require.NoError(t, err)
	assert.Equal(t, int64(1), h.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_AddHistory_Error(t *testing.T) {
	repo, mock := newReportRepoMock(t)
	h := entities.NewReportHistory(1, nil, entities.ReportActionCreated)

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO report_history")).
		WillReturnError(sql.ErrConnDone)

	err := repo.AddHistory(context.Background(), h)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetHistoryByReport ---

func TestReportRepositoryPG_GetHistoryByReport_Success(t *testing.T) {
	repo, mock := newReportRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "report_id", "user_id", "action", "details", "created_at"}).
		AddRow(int64(1), int64(1), nil, entities.ReportActionCreated, json.RawMessage(`{}`), now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, report_id, user_id")).
		WithArgs(int64(1), 10, 0).
		WillReturnRows(rows)

	history, err := repo.GetHistoryByReport(context.Background(), 1, 10, 0)
	require.NoError(t, err)
	assert.Len(t, history, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_GetHistoryByReport_Error(t *testing.T) {
	repo, mock := newReportRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, report_id")).
		WillReturnError(sql.ErrConnDone)

	_, err := repo.GetHistoryByReport(context.Background(), 1, 10, 0)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_GetHistoryByReport_ScanError(t *testing.T) {
	repo, mock := newReportRepoMock(t)
	rows := sqlmock.NewRows([]string{"id", "report_id", "user_id", "action", "details", "created_at"}).
		AddRow("bad", int64(1), nil, "created", nil, time.Now())

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, report_id")).
		WithArgs(int64(1), 10, 0).
		WillReturnRows(rows)

	_, err := repo.GetHistoryByReport(context.Background(), 1, 10, 0)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- CreateGenerationLog ---

func TestReportRepositoryPG_CreateGenerationLog_Success(t *testing.T) {
	repo, mock := newReportRepoMock(t)
	l := entities.NewReportGenerationLog(1)

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO report_generation_log")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	err := repo.CreateGenerationLog(context.Background(), l)
	require.NoError(t, err)
	assert.Equal(t, int64(1), l.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_CreateGenerationLog_Error(t *testing.T) {
	repo, mock := newReportRepoMock(t)
	l := entities.NewReportGenerationLog(1)

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO report_generation_log")).
		WillReturnError(sql.ErrConnDone)

	err := repo.CreateGenerationLog(context.Background(), l)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- UpdateGenerationLog ---

func TestReportRepositoryPG_UpdateGenerationLog_Success(t *testing.T) {
	repo, mock := newReportRepoMock(t)
	l := entities.NewReportGenerationLog(1)
	l.ID = 5
	l.Complete(100)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE report_generation_log SET")).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdateGenerationLog(context.Background(), l)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_UpdateGenerationLog_NotFound(t *testing.T) {
	repo, mock := newReportRepoMock(t)
	l := &entities.ReportGenerationLog{ID: 999}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE report_generation_log")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.UpdateGenerationLog(context.Background(), l)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_UpdateGenerationLog_Error(t *testing.T) {
	repo, mock := newReportRepoMock(t)
	l := &entities.ReportGenerationLog{ID: 1}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE report_generation_log")).
		WillReturnError(sql.ErrConnDone)

	err := repo.UpdateGenerationLog(context.Background(), l)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_UpdateGenerationLog_RowsAffectedError(t *testing.T) {
	repo, mock := newReportRepoMock(t)
	l := &entities.ReportGenerationLog{ID: 1}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE report_generation_log")).
		WillReturnResult(sqlmock.NewErrorResult(sql.ErrConnDone))

	err := repo.UpdateGenerationLog(context.Background(), l)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetGenerationLogsByReport ---

func TestReportRepositoryPG_GetGenerationLogsByReport_Success(t *testing.T) {
	repo, mock := newReportRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "report_id", "started_at", "completed_at", "status", "error_message", "duration_seconds", "records_processed"}).
		AddRow(int64(1), int64(1), now, nil, domain.GenerationStatusStarted, nil, nil, nil)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, report_id, started_at")).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	logs, err := repo.GetGenerationLogsByReport(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, logs, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_GetGenerationLogsByReport_Error(t *testing.T) {
	repo, mock := newReportRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, report_id")).
		WillReturnError(sql.ErrConnDone)

	_, err := repo.GetGenerationLogsByReport(context.Background(), 1)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportRepositoryPG_GetGenerationLogsByReport_ScanError(t *testing.T) {
	repo, mock := newReportRepoMock(t)
	rows := sqlmock.NewRows([]string{"id", "report_id", "started_at", "completed_at", "status", "error_message", "duration_seconds", "records_processed"}).
		AddRow("bad", int64(1), time.Now(), nil, "started", nil, nil, nil)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, report_id")).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	_, err := repo.GetGenerationLogsByReport(context.Background(), 1)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
