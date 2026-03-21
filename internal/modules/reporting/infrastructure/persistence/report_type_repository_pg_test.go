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

func newRTRepoMock(t *testing.T) (*ReportTypeRepositoryPG, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewReportTypeRepositoryPG(db), mock
}

var rtCols = []string{
	"id", "name", "code", "description", "category", "template_path",
	"output_format", "is_periodic", "period_type", "created_at", "updated_at",
}

func newRTRows() *sqlmock.Rows { return sqlmock.NewRows(rtCols) }

func addRTRow(rows *sqlmock.Rows, id int64, name string) *sqlmock.Rows {
	now := time.Now()
	cat := domain.ReportCategoryAcademic
	return rows.AddRow(id, name, "code1", nil, &cat, nil, domain.OutputFormatPDF, false, nil, now, now)
}

// --- Create ---

func TestReportTypeRepositoryPG_Create_Success(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	rt := entities.NewReportType("Test", "test_code", domain.OutputFormatPDF)

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO report_types")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	err := repo.Create(context.Background(), rt)
	require.NoError(t, err)
	assert.Equal(t, int64(1), rt.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_Create_Error(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	rt := entities.NewReportType("Test", "code", domain.OutputFormatPDF)

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO report_types")).WillReturnError(sql.ErrConnDone)

	err := repo.Create(context.Background(), rt)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- Save ---

func TestReportTypeRepositoryPG_Save_Success(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	rt := entities.NewReportType("Updated", "code", domain.OutputFormatXLSX)
	rt.ID = 5

	mock.ExpectExec(regexp.QuoteMeta("UPDATE report_types SET")).WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Save(context.Background(), rt)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_Save_NotFound(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	rt := entities.NewReportType("X", "x", domain.OutputFormatPDF)
	rt.ID = 999

	mock.ExpectExec(regexp.QuoteMeta("UPDATE report_types")).WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Save(context.Background(), rt)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_Save_Error(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	rt := entities.NewReportType("X", "x", domain.OutputFormatPDF)
	rt.ID = 5

	mock.ExpectExec(regexp.QuoteMeta("UPDATE report_types")).WillReturnError(sql.ErrConnDone)

	err := repo.Save(context.Background(), rt)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_Save_RowsAffectedError(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	rt := entities.NewReportType("X", "x", domain.OutputFormatPDF)
	rt.ID = 5

	mock.ExpectExec(regexp.QuoteMeta("UPDATE report_types")).WillReturnResult(sqlmock.NewErrorResult(sql.ErrConnDone))

	err := repo.Save(context.Background(), rt)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetByID ---

func TestReportTypeRepositoryPG_GetByID_Success(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	rows := addRTRow(newRTRows(), 1, "RT1")

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, code")).WithArgs(int64(1)).WillReturnRows(rows)

	rt, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	require.NotNil(t, rt)
	assert.Equal(t, "RT1", rt.Name)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_GetByID_NotFound(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, code")).WillReturnError(sql.ErrNoRows)

	rt, err := repo.GetByID(context.Background(), 999)
	require.NoError(t, err)
	assert.Nil(t, rt)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_GetByID_Error(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, code")).WillReturnError(sql.ErrConnDone)

	_, err := repo.GetByID(context.Background(), 1)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetByCode ---

func TestReportTypeRepositoryPG_GetByCode_Success(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	rows := addRTRow(newRTRows(), 1, "RT1")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, code")).WithArgs("code1").WillReturnRows(rows)

	rt, err := repo.GetByCode(context.Background(), "code1")
	require.NoError(t, err)
	require.NotNil(t, rt)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_GetByCode_NotFound(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, code")).WillReturnError(sql.ErrNoRows)

	rt, err := repo.GetByCode(context.Background(), "nope")
	require.NoError(t, err)
	assert.Nil(t, rt)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_GetByCode_Error(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, code")).WillReturnError(sql.ErrConnDone)

	_, err := repo.GetByCode(context.Background(), "x")
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- Delete ---

func TestReportTypeRepositoryPG_Delete_Success(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM report_types")).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.Delete(context.Background(), 1))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_Delete_NotFound(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM report_types")).WillReturnResult(sqlmock.NewResult(0, 0))
	require.Error(t, repo.Delete(context.Background(), 999))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_Delete_Error(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM report_types")).WillReturnError(sql.ErrConnDone)
	require.Error(t, repo.Delete(context.Background(), 1))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_Delete_RowsAffectedError(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM report_types")).WillReturnResult(sqlmock.NewErrorResult(sql.ErrConnDone))
	require.Error(t, repo.Delete(context.Background(), 1))
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- List ---

func TestReportTypeRepositoryPG_List_NoFilter(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	rows := addRTRow(newRTRows(), 1, "RT1")
	mock.ExpectQuery("SELECT id, name, code").WillReturnRows(rows)

	rts, err := repo.List(context.Background(), repositories.ReportTypeFilter{}, 10, 0)
	require.NoError(t, err)
	assert.Len(t, rts, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_List_WithFilter(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	cat := domain.ReportCategoryAcademic
	isPeriodic := true
	rows := addRTRow(newRTRows(), 1, "RT1")
	mock.ExpectQuery("SELECT id, name, code").WillReturnRows(rows)

	rts, err := repo.List(context.Background(), repositories.ReportTypeFilter{Category: &cat, IsPeriodic: &isPeriodic}, 10, 0)
	require.NoError(t, err)
	assert.Len(t, rts, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_List_NoLimit(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	rows := addRTRow(newRTRows(), 1, "RT1")
	mock.ExpectQuery("SELECT id, name, code").WillReturnRows(rows)

	rts, err := repo.List(context.Background(), repositories.ReportTypeFilter{}, 0, 0)
	require.NoError(t, err)
	assert.Len(t, rts, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_List_Error(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	mock.ExpectQuery("SELECT id, name, code").WillReturnError(sql.ErrConnDone)

	_, err := repo.List(context.Background(), repositories.ReportTypeFilter{}, 10, 0)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_List_ScanError(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	rows := sqlmock.NewRows(rtCols).AddRow("bad", "name", "code", nil, nil, nil, "pdf", false, nil, time.Now(), time.Now())
	mock.ExpectQuery("SELECT id, name, code").WillReturnRows(rows)

	_, err := repo.List(context.Background(), repositories.ReportTypeFilter{}, 10, 0)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_List_RowsErr(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	rows := addRTRow(newRTRows(), 1, "RT1")
	rows.RowError(0, sql.ErrConnDone)
	mock.ExpectQuery("SELECT id, name, code").WillReturnRows(rows)

	_, err := repo.List(context.Background(), repositories.ReportTypeFilter{}, 10, 0)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- Count ---

func TestReportTypeRepositoryPG_Count_Success(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(3)))

	count, err := repo.Count(context.Background(), repositories.ReportTypeFilter{})
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_Count_Error(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).WillReturnError(sql.ErrConnDone)

	_, err := repo.Count(context.Background(), repositories.ReportTypeFilter{})
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetByCategory, GetPeriodic ---

func TestReportTypeRepositoryPG_GetByCategory(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	rows := addRTRow(newRTRows(), 1, "RT1")
	mock.ExpectQuery("SELECT id, name, code").WillReturnRows(rows)

	rts, err := repo.GetByCategory(context.Background(), domain.ReportCategoryAcademic)
	require.NoError(t, err)
	assert.Len(t, rts, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_GetPeriodic(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	rows := addRTRow(newRTRows(), 1, "RT1")
	mock.ExpectQuery("SELECT id, name, code").WillReturnRows(rows)

	rts, err := repo.GetPeriodic(context.Background())
	require.NoError(t, err)
	assert.Len(t, rts, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- AddParameter ---

func TestReportTypeRepositoryPG_AddParameter_Success(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	p := entities.NewReportParameter(1, "param1", domain.ParameterTypeString, true)

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO report_parameters")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))

	err := repo.AddParameter(context.Background(), p)
	require.NoError(t, err)
	assert.Equal(t, int64(1), p.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_AddParameter_Error(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	p := entities.NewReportParameter(1, "p", domain.ParameterTypeString, false)
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO report_parameters")).WillReturnError(sql.ErrConnDone)
	require.Error(t, repo.AddParameter(context.Background(), p))
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- UpdateParameter ---

func TestReportTypeRepositoryPG_UpdateParameter_Success(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	p := &entities.ReportParameter{ID: 1, ParameterName: "updated", ParameterType: domain.ParameterTypeNumber}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE report_parameters SET")).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.UpdateParameter(context.Background(), p))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_UpdateParameter_NotFound(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	p := &entities.ReportParameter{ID: 999}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE report_parameters")).WillReturnResult(sqlmock.NewResult(0, 0))
	require.Error(t, repo.UpdateParameter(context.Background(), p))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_UpdateParameter_Error(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	p := &entities.ReportParameter{ID: 1}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE report_parameters")).WillReturnError(sql.ErrConnDone)
	require.Error(t, repo.UpdateParameter(context.Background(), p))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_UpdateParameter_RowsAffectedError(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	p := &entities.ReportParameter{ID: 1}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE report_parameters")).WillReturnResult(sqlmock.NewErrorResult(sql.ErrConnDone))
	require.Error(t, repo.UpdateParameter(context.Background(), p))
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- DeleteParameter ---

func TestReportTypeRepositoryPG_DeleteParameter_Success(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM report_parameters")).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.DeleteParameter(context.Background(), 1))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_DeleteParameter_NotFound(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM report_parameters")).WillReturnResult(sqlmock.NewResult(0, 0))
	require.Error(t, repo.DeleteParameter(context.Background(), 999))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_DeleteParameter_Error(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM report_parameters")).WillReturnError(sql.ErrConnDone)
	require.Error(t, repo.DeleteParameter(context.Background(), 1))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_DeleteParameter_RowsAffectedError(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM report_parameters")).WillReturnResult(sqlmock.NewErrorResult(sql.ErrConnDone))
	require.Error(t, repo.DeleteParameter(context.Background(), 1))
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetParametersByReportType ---

func TestReportTypeRepositoryPG_GetParametersByReportType_Success(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "report_type_id", "parameter_name", "parameter_type", "is_required", "default_value", "options", "display_order", "created_at"}).
		AddRow(int64(1), int64(1), "param1", domain.ParameterTypeString, true, nil, json.RawMessage(`[]`), 0, now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, report_type_id")).WithArgs(int64(1)).WillReturnRows(rows)

	params, err := repo.GetParametersByReportType(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, params, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_GetParametersByReportType_Error(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, report_type_id")).WillReturnError(sql.ErrConnDone)
	_, err := repo.GetParametersByReportType(context.Background(), 1)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_GetParametersByReportType_ScanError(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	rows := sqlmock.NewRows([]string{"id", "report_type_id", "parameter_name", "parameter_type", "is_required", "default_value", "options", "display_order", "created_at"}).
		AddRow("bad", int64(1), "param1", "string", true, nil, nil, 0, time.Now())
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, report_type_id")).WithArgs(int64(1)).WillReturnRows(rows)
	_, err := repo.GetParametersByReportType(context.Background(), 1)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- AddTemplate ---

func TestReportTypeRepositoryPG_AddTemplate_Success(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	tmpl := entities.NewReportTemplate(1, "tmpl", "content", 10)
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO report_templates")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
	require.NoError(t, repo.AddTemplate(context.Background(), tmpl))
	assert.Equal(t, int64(1), tmpl.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_AddTemplate_Error(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	tmpl := entities.NewReportTemplate(1, "t", "c", 10)
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO report_templates")).WillReturnError(sql.ErrConnDone)
	require.Error(t, repo.AddTemplate(context.Background(), tmpl))
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- UpdateTemplate ---

func TestReportTypeRepositoryPG_UpdateTemplate_Success(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	tmpl := &entities.ReportTemplate{ID: 1, Name: "updated", Content: "c", UpdatedAt: time.Now()}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE report_templates SET")).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.UpdateTemplate(context.Background(), tmpl))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_UpdateTemplate_NotFound(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	tmpl := &entities.ReportTemplate{ID: 999}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE report_templates")).WillReturnResult(sqlmock.NewResult(0, 0))
	require.Error(t, repo.UpdateTemplate(context.Background(), tmpl))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_UpdateTemplate_Error(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	tmpl := &entities.ReportTemplate{ID: 1}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE report_templates")).WillReturnError(sql.ErrConnDone)
	require.Error(t, repo.UpdateTemplate(context.Background(), tmpl))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_UpdateTemplate_RowsAffectedError(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	tmpl := &entities.ReportTemplate{ID: 1}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE report_templates")).WillReturnResult(sqlmock.NewErrorResult(sql.ErrConnDone))
	require.Error(t, repo.UpdateTemplate(context.Background(), tmpl))
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- DeleteTemplate ---

func TestReportTypeRepositoryPG_DeleteTemplate_Success(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM report_templates")).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.DeleteTemplate(context.Background(), 1))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_DeleteTemplate_NotFound(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM report_templates")).WillReturnResult(sqlmock.NewResult(0, 0))
	require.Error(t, repo.DeleteTemplate(context.Background(), 999))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_DeleteTemplate_Error(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM report_templates")).WillReturnError(sql.ErrConnDone)
	require.Error(t, repo.DeleteTemplate(context.Background(), 1))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_DeleteTemplate_RowsAffectedError(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM report_templates")).WillReturnResult(sqlmock.NewErrorResult(sql.ErrConnDone))
	require.Error(t, repo.DeleteTemplate(context.Background(), 1))
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetTemplatesByReportType ---

func TestReportTypeRepositoryPG_GetTemplatesByReportType_Success(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "report_type_id", "name", "content", "is_default", "created_by", "created_at", "updated_at"}).
		AddRow(int64(1), int64(1), "tmpl", "content", true, int64(10), now, now)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, report_type_id, name")).WithArgs(int64(1)).WillReturnRows(rows)

	tmpls, err := repo.GetTemplatesByReportType(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, tmpls, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_GetTemplatesByReportType_Error(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, report_type_id")).WillReturnError(sql.ErrConnDone)
	_, err := repo.GetTemplatesByReportType(context.Background(), 1)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_GetTemplatesByReportType_ScanError(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	rows := sqlmock.NewRows([]string{"id", "report_type_id", "name", "content", "is_default", "created_by", "created_at", "updated_at"}).
		AddRow("bad", int64(1), "t", "c", true, int64(10), time.Now(), time.Now())
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, report_type_id")).WithArgs(int64(1)).WillReturnRows(rows)
	_, err := repo.GetTemplatesByReportType(context.Background(), 1)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetDefaultTemplate ---

func TestReportTypeRepositoryPG_GetDefaultTemplate_Success(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "report_type_id", "name", "content", "is_default", "created_by", "created_at", "updated_at"}).
		AddRow(int64(1), int64(1), "default", "content", true, int64(10), now, now)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, report_type_id, name")).WithArgs(int64(1)).WillReturnRows(rows)

	tmpl, err := repo.GetDefaultTemplate(context.Background(), 1)
	require.NoError(t, err)
	require.NotNil(t, tmpl)
	assert.True(t, tmpl.IsDefault)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_GetDefaultTemplate_NotFound(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, report_type_id")).WillReturnError(sql.ErrNoRows)

	tmpl, err := repo.GetDefaultTemplate(context.Background(), 1)
	require.NoError(t, err)
	assert.Nil(t, tmpl)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_GetDefaultTemplate_Error(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, report_type_id")).WillReturnError(sql.ErrConnDone)
	_, err := repo.GetDefaultTemplate(context.Background(), 1)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- SetDefaultTemplate ---

func TestReportTypeRepositoryPG_SetDefaultTemplate_Success(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE report_templates SET is_default = false")).WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE report_templates SET is_default = true")).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.SetDefaultTemplate(context.Background(), 1, 5)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_SetDefaultTemplate_BeginError(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	mock.ExpectBegin().WillReturnError(sql.ErrConnDone)
	require.Error(t, repo.SetDefaultTemplate(context.Background(), 1, 5))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_SetDefaultTemplate_RemoveDefaultError(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE report_templates SET is_default = false")).WillReturnError(sql.ErrConnDone)
	mock.ExpectRollback()
	require.Error(t, repo.SetDefaultTemplate(context.Background(), 1, 5))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_SetDefaultTemplate_SetDefaultError(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE report_templates SET is_default = false")).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE report_templates SET is_default = true")).WillReturnError(sql.ErrConnDone)
	mock.ExpectRollback()
	require.Error(t, repo.SetDefaultTemplate(context.Background(), 1, 5))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_SetDefaultTemplate_NotFound(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE report_templates SET is_default = false")).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE report_templates SET is_default = true")).WillReturnResult(sqlmock.NewResult(0, 0))
	// Note: err from RowsAffected is nil, so deferred rollback does not trigger;
	// the returned error is a new fmt.Errorf, not assigned to the named `err` variable.
	// The transaction will be cleaned up by db.Close in test cleanup.
	require.Error(t, repo.SetDefaultTemplate(context.Background(), 1, 999))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_SetDefaultTemplate_RowsAffectedError(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE report_templates SET is_default = false")).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE report_templates SET is_default = true")).WillReturnResult(sqlmock.NewErrorResult(sql.ErrConnDone))
	mock.ExpectRollback()
	require.Error(t, repo.SetDefaultTemplate(context.Background(), 1, 5))
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- Subscribe ---

func TestReportTypeRepositoryPG_Subscribe_Success(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	s := entities.NewReportSubscription(1, 2, domain.DeliveryMethodEmail)
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO report_subscriptions")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(1)))
	require.NoError(t, repo.Subscribe(context.Background(), s))
	assert.Equal(t, int64(1), s.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_Subscribe_Error(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	s := entities.NewReportSubscription(1, 2, domain.DeliveryMethodEmail)
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO report_subscriptions")).WillReturnError(sql.ErrConnDone)
	require.Error(t, repo.Subscribe(context.Background(), s))
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- Unsubscribe ---

func TestReportTypeRepositoryPG_Unsubscribe_Success(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM report_subscriptions")).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.Unsubscribe(context.Background(), 1, 2))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_Unsubscribe_NotFound(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM report_subscriptions")).WillReturnResult(sqlmock.NewResult(0, 0))
	require.Error(t, repo.Unsubscribe(context.Background(), 1, 999))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_Unsubscribe_Error(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM report_subscriptions")).WillReturnError(sql.ErrConnDone)
	require.Error(t, repo.Unsubscribe(context.Background(), 1, 2))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_Unsubscribe_RowsAffectedError(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM report_subscriptions")).WillReturnResult(sqlmock.NewErrorResult(sql.ErrConnDone))
	require.Error(t, repo.Unsubscribe(context.Background(), 1, 2))
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetSubscription ---

func TestReportTypeRepositoryPG_GetSubscription_Success(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "report_type_id", "user_id", "delivery_method", "is_active", "created_at"}).
		AddRow(int64(1), int64(1), int64(2), domain.DeliveryMethodEmail, true, now)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, report_type_id")).WillReturnRows(rows)

	s, err := repo.GetSubscription(context.Background(), 1, 2)
	require.NoError(t, err)
	require.NotNil(t, s)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_GetSubscription_NotFound(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, report_type_id")).WillReturnError(sql.ErrNoRows)
	s, err := repo.GetSubscription(context.Background(), 1, 999)
	require.NoError(t, err)
	assert.Nil(t, s)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_GetSubscription_Error(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, report_type_id")).WillReturnError(sql.ErrConnDone)
	_, err := repo.GetSubscription(context.Background(), 1, 2)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetSubscribersByReportType ---

func TestReportTypeRepositoryPG_GetSubscribersByReportType_Success(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "report_type_id", "user_id", "delivery_method", "is_active", "created_at"}).
		AddRow(int64(1), int64(1), int64(2), domain.DeliveryMethodEmail, true, now)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, report_type_id")).WillReturnRows(rows)

	subs, err := repo.GetSubscribersByReportType(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, subs, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_GetSubscribersByReportType_Error(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, report_type_id")).WillReturnError(sql.ErrConnDone)
	_, err := repo.GetSubscribersByReportType(context.Background(), 1)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_GetSubscribersByReportType_ScanError(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	rows := sqlmock.NewRows([]string{"id", "report_type_id", "user_id", "delivery_method", "is_active", "created_at"}).
		AddRow("bad", int64(1), int64(2), "email", true, time.Now())
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, report_type_id")).WillReturnRows(rows)
	_, err := repo.GetSubscribersByReportType(context.Background(), 1)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- GetSubscriptionsByUser ---

func TestReportTypeRepositoryPG_GetSubscriptionsByUser_Success(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "report_type_id", "user_id", "delivery_method", "is_active", "created_at"}).
		AddRow(int64(1), int64(1), int64(2), domain.DeliveryMethodBoth, true, now)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, report_type_id")).WillReturnRows(rows)

	subs, err := repo.GetSubscriptionsByUser(context.Background(), 2)
	require.NoError(t, err)
	assert.Len(t, subs, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_GetSubscriptionsByUser_Error(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, report_type_id")).WillReturnError(sql.ErrConnDone)
	_, err := repo.GetSubscriptionsByUser(context.Background(), 2)
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

// --- UpdateSubscription ---

func TestReportTypeRepositoryPG_UpdateSubscription_Success(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	s := &entities.ReportSubscription{ID: 1, DeliveryMethod: domain.DeliveryMethodBoth, IsActive: true}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE report_subscriptions SET")).WillReturnResult(sqlmock.NewResult(0, 1))
	require.NoError(t, repo.UpdateSubscription(context.Background(), s))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_UpdateSubscription_NotFound(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	s := &entities.ReportSubscription{ID: 999}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE report_subscriptions")).WillReturnResult(sqlmock.NewResult(0, 0))
	require.Error(t, repo.UpdateSubscription(context.Background(), s))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_UpdateSubscription_Error(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	s := &entities.ReportSubscription{ID: 1}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE report_subscriptions")).WillReturnError(sql.ErrConnDone)
	require.Error(t, repo.UpdateSubscription(context.Background(), s))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestReportTypeRepositoryPG_UpdateSubscription_RowsAffectedError(t *testing.T) {
	repo, mock := newRTRepoMock(t)
	s := &entities.ReportSubscription{ID: 1}
	mock.ExpectExec(regexp.QuoteMeta("UPDATE report_subscriptions")).WillReturnResult(sqlmock.NewErrorResult(sql.ErrConnDone))
	require.Error(t, repo.UpdateSubscription(context.Background(), s))
	require.NoError(t, mock.ExpectationsWereMet())
}
