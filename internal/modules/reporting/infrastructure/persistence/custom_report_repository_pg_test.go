package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain/repositories"
)

func newCRRepoMock(t *testing.T) (*CustomReportRepositoryPG, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewCustomReportRepositoryPG(db), mock
}

var crCols = []string{
	"id", "name", "description", "data_source", "fields", "filters",
	"groupings", "sortings", "created_at", "updated_at", "created_by", "is_public",
}

func newCRRows() *sqlmock.Rows { return sqlmock.NewRows(crCols) }

func addCRRow(rows *sqlmock.Rows, id uuid.UUID, name string) *sqlmock.Rows {
	now := time.Now()
	return rows.AddRow(
		id, name, nil, "documents", json.RawMessage(`[]`), json.RawMessage(`[]`),
		json.RawMessage(`[]`), json.RawMessage(`[]`), now, now, int64(1), false,
	)
}

// --- Create ---

func TestCustomReportRepositoryPG_Create_Success(t *testing.T) {
	repo, mock := newCRRepoMock(t)
	cr := entities.NewCustomReport("Test", "desc", entities.DataSourceDocuments, 1)

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO custom_reports")).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Create(context.Background(), cr)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCustomReportRepositoryPG_Create_NoDescription(t *testing.T) {
	repo, mock := newCRRepoMock(t)
	cr := entities.NewCustomReport("Test", "", entities.DataSourceDocuments, 1)

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO custom_reports")).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Create(context.Background(), cr)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCustomReportRepositoryPG_Create_Error(t *testing.T) {
	repo, mock := newCRRepoMock(t)
	cr := entities.NewCustomReport("Test", "", entities.DataSourceDocuments, 1)

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO custom_reports")).
		WillReturnError(sql.ErrConnDone)

	err := repo.Create(context.Background(), cr)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- Update ---

func TestCustomReportRepositoryPG_Update_Success(t *testing.T) {
	repo, mock := newCRRepoMock(t)
	cr := entities.NewCustomReport("Updated", "desc", entities.DataSourceUsers, 1)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE custom_reports")).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), cr)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCustomReportRepositoryPG_Update_NoDescription(t *testing.T) {
	repo, mock := newCRRepoMock(t)
	cr := entities.NewCustomReport("Updated", "", entities.DataSourceUsers, 1)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE custom_reports")).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), cr)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCustomReportRepositoryPG_Update_NotFound(t *testing.T) {
	repo, mock := newCRRepoMock(t)
	cr := entities.NewCustomReport("X", "", entities.DataSourceDocuments, 1)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE custom_reports")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Update(context.Background(), cr)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCustomReportRepositoryPG_Update_Error(t *testing.T) {
	repo, mock := newCRRepoMock(t)
	cr := entities.NewCustomReport("X", "", entities.DataSourceDocuments, 1)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE custom_reports")).
		WillReturnError(sql.ErrConnDone)

	err := repo.Update(context.Background(), cr)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCustomReportRepositoryPG_Update_RowsAffectedError(t *testing.T) {
	repo, mock := newCRRepoMock(t)
	cr := entities.NewCustomReport("X", "", entities.DataSourceDocuments, 1)

	mock.ExpectExec(regexp.QuoteMeta("UPDATE custom_reports")).
		WillReturnResult(sqlmock.NewErrorResult(sql.ErrConnDone))

	err := repo.Update(context.Background(), cr)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetByID ---

func TestCustomReportRepositoryPG_GetByID_Success(t *testing.T) {
	repo, mock := newCRRepoMock(t)
	id := uuid.New()
	rows := addCRRow(newCRRows(), id, "Report1")

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, description")).
		WithArgs(id).
		WillReturnRows(rows)

	cr, err := repo.GetByID(context.Background(), id)
	require.NoError(t, err)
	require.NotNil(t, cr)
	assert.Equal(t, "Report1", cr.Name)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCustomReportRepositoryPG_GetByID_WithDescription(t *testing.T) {
	repo, mock := newCRRepoMock(t)
	id := uuid.New()
	now := time.Now()
	rows := sqlmock.NewRows(crCols).AddRow(
		id, "Report", "my desc", "documents", nil, nil, nil, nil, now, now, int64(1), true,
	)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, description")).
		WithArgs(id).
		WillReturnRows(rows)

	cr, err := repo.GetByID(context.Background(), id)
	require.NoError(t, err)
	require.NotNil(t, cr)
	assert.Equal(t, "my desc", cr.Description)
	assert.True(t, cr.IsPublic)
	// All JSON fields should default to empty slices
	assert.Equal(t, []entities.SelectedField{}, cr.Fields)
	assert.Equal(t, []entities.ReportFilterConfig{}, cr.Filters)
	assert.Equal(t, []entities.ReportGrouping{}, cr.Groupings)
	assert.Equal(t, []entities.ReportSorting{}, cr.Sortings)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCustomReportRepositoryPG_GetByID_NotFound(t *testing.T) {
	repo, mock := newCRRepoMock(t)
	id := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name")).
		WithArgs(id).
		WillReturnError(sql.ErrNoRows)

	cr, err := repo.GetByID(context.Background(), id)
	require.NoError(t, err)
	assert.Nil(t, cr)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCustomReportRepositoryPG_GetByID_Error(t *testing.T) {
	repo, mock := newCRRepoMock(t)
	id := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name")).
		WillReturnError(sql.ErrConnDone)

	cr, err := repo.GetByID(context.Background(), id)
	require.Error(t, err)
	assert.Nil(t, cr)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- Delete ---

func TestCustomReportRepositoryPG_Delete_Success(t *testing.T) {
	repo, mock := newCRRepoMock(t)
	id := uuid.New()

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM custom_reports")).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), id)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCustomReportRepositoryPG_Delete_NotFound(t *testing.T) {
	repo, mock := newCRRepoMock(t)
	id := uuid.New()

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM custom_reports")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Delete(context.Background(), id)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCustomReportRepositoryPG_Delete_Error(t *testing.T) {
	repo, mock := newCRRepoMock(t)
	id := uuid.New()

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM custom_reports")).
		WillReturnError(sql.ErrConnDone)

	err := repo.Delete(context.Background(), id)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCustomReportRepositoryPG_Delete_RowsAffectedError(t *testing.T) {
	repo, mock := newCRRepoMock(t)
	id := uuid.New()

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM custom_reports")).
		WillReturnResult(sqlmock.NewErrorResult(sql.ErrConnDone))

	err := repo.Delete(context.Background(), id)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- List ---

func TestCustomReportRepositoryPG_List_NoFilter(t *testing.T) {
	repo, mock := newCRRepoMock(t)
	id := uuid.New()
	rows := addCRRow(newCRRows(), id, "R1")

	mock.ExpectQuery("SELECT id, name, description").WillReturnRows(rows)

	reports, err := repo.List(context.Background(), repositories.CustomReportFilter{})
	require.NoError(t, err)
	assert.Len(t, reports, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCustomReportRepositoryPG_List_AllFilters(t *testing.T) {
	repo, mock := newCRRepoMock(t)
	createdBy := int64(1)
	ds := entities.DataSourceDocuments
	isPublic := true
	filter := repositories.CustomReportFilter{
		CreatedBy:  &createdBy,
		DataSource: &ds,
		IsPublic:   &isPublic,
		Search:     "test",
		Page:       1,
		PageSize:   5,
	}

	id := uuid.New()
	rows := addCRRow(newCRRows(), id, "test report")
	mock.ExpectQuery("SELECT id, name, description").WillReturnRows(rows)

	reports, err := repo.List(context.Background(), filter)
	require.NoError(t, err)
	assert.Len(t, reports, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCustomReportRepositoryPG_List_Pagination(t *testing.T) {
	repo, mock := newCRRepoMock(t)
	filter := repositories.CustomReportFilter{Page: 0, PageSize: 0} // defaults

	mock.ExpectQuery("SELECT id, name, description").WillReturnRows(newCRRows())

	reports, err := repo.List(context.Background(), filter)
	require.NoError(t, err)
	assert.Empty(t, reports)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCustomReportRepositoryPG_List_MaxPageSize(t *testing.T) {
	repo, mock := newCRRepoMock(t)
	filter := repositories.CustomReportFilter{Page: 1, PageSize: 200} // should be capped to 100

	mock.ExpectQuery("SELECT id, name, description").WillReturnRows(newCRRows())

	reports, err := repo.List(context.Background(), filter)
	require.NoError(t, err)
	assert.Empty(t, reports)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCustomReportRepositoryPG_List_Error(t *testing.T) {
	repo, mock := newCRRepoMock(t)

	mock.ExpectQuery("SELECT id, name").WillReturnError(sql.ErrConnDone)

	_, err := repo.List(context.Background(), repositories.CustomReportFilter{})
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCustomReportRepositoryPG_List_ScanError(t *testing.T) {
	repo, mock := newCRRepoMock(t)
	rows := sqlmock.NewRows(crCols).AddRow(
		"not-uuid", "name", nil, "documents", nil, nil, nil, nil, time.Now(), time.Now(), int64(1), false,
	)

	mock.ExpectQuery("SELECT id, name").WillReturnRows(rows)

	_, err := repo.List(context.Background(), repositories.CustomReportFilter{})
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- Count ---

func TestCustomReportRepositoryPG_Count_NoFilter(t *testing.T) {
	repo, mock := newCRRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(5)))

	count, err := repo.Count(context.Background(), repositories.CustomReportFilter{})
	require.NoError(t, err)
	assert.Equal(t, int64(5), count)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCustomReportRepositoryPG_Count_AllFilters(t *testing.T) {
	repo, mock := newCRRepoMock(t)
	createdBy := int64(1)
	ds := entities.DataSourceDocuments
	isPublic := true
	filter := repositories.CustomReportFilter{
		CreatedBy:  &createdBy,
		DataSource: &ds,
		IsPublic:   &isPublic,
		Search:     "test",
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(2)))

	count, err := repo.Count(context.Background(), filter)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCustomReportRepositoryPG_Count_Error(t *testing.T) {
	repo, mock := newCRRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnError(sql.ErrConnDone)

	_, err := repo.Count(context.Background(), repositories.CustomReportFilter{})
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetByCreator ---

func TestCustomReportRepositoryPG_GetByCreator(t *testing.T) {
	repo, mock := newCRRepoMock(t)
	id := uuid.New()
	rows := addCRRow(newCRRows(), id, "R1")
	mock.ExpectQuery("SELECT id, name").WillReturnRows(rows)

	reports, err := repo.GetByCreator(context.Background(), 1, 1, 10)
	require.NoError(t, err)
	assert.Len(t, reports, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetPublicReports ---

func TestCustomReportRepositoryPG_GetPublicReports(t *testing.T) {
	repo, mock := newCRRepoMock(t)
	id := uuid.New()
	rows := addCRRow(newCRRows(), id, "R1")
	mock.ExpectQuery("SELECT id, name").WillReturnRows(rows)

	reports, err := repo.GetPublicReports(context.Background(), 1, 10)
	require.NoError(t, err)
	assert.Len(t, reports, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}
