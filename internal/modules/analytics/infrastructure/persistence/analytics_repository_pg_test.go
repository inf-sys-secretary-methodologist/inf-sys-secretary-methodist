package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/analytics/domain/entities"
)

func newAnalyticsRepoMock(t *testing.T) (*AnalyticsRepositoryPG, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return NewAnalyticsRepositoryPG(db), mock
}

var riskCols = []string{"student_id", "student_name", "group_name", "attendance_rate", "grade_average", "risk_level", "risk_score", "risk_factors"}

// ---- GetAtRiskStudents ----

func TestAnalyticsGetAtRiskStudents_Success(t *testing.T) {
	repo, mock := newAnalyticsRepoMock(t)
	groupName := "Group1"
	ar := 0.75
	ga := 3.5

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT student_id, student_name")).
		WithArgs(10, 0).
		WillReturnRows(sqlmock.NewRows(riskCols).
			AddRow(int64(1), "John", &groupName, &ar, &ga, "high", 75.0, []byte(`{"attendance":{"rate":0.75,"absent_count":5,"is_risk":true},"grades":{"average":3.5,"failing_count":2,"is_risk":true}}`)))

	students, total, err := repo.GetAtRiskStudents(context.Background(), nil, 10, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, students, 1)
	assert.NotNil(t, students[0].RiskFactors)
}

func TestAnalyticsGetAtRiskStudents_InvalidJSON(t *testing.T) {
	repo, mock := newAnalyticsRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT student_id, student_name")).
		WithArgs(10, 0).
		WillReturnRows(sqlmock.NewRows(riskCols).
			AddRow(int64(1), "John", nil, nil, nil, "low", 10.0, []byte(`{invalid`)))

	students, _, err := repo.GetAtRiskStudents(context.Background(), nil, 10, 0)
	require.NoError(t, err)
	assert.Nil(t, students[0].RiskFactors)
}

func TestAnalyticsGetAtRiskStudents_CountError(t *testing.T) {
	repo, mock := newAnalyticsRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnError(fmt.Errorf("count error"))

	_, _, err := repo.GetAtRiskStudents(context.Background(), nil, 10, 0)
	assert.Error(t, err)
}

func TestAnalyticsGetAtRiskStudents_QueryError(t *testing.T) {
	repo, mock := newAnalyticsRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT student_id")).
		WithArgs(10, 0).
		WillReturnError(fmt.Errorf("query error"))

	_, _, err := repo.GetAtRiskStudents(context.Background(), nil, 10, 0)
	assert.Error(t, err)
}

func TestAnalyticsGetAtRiskStudents_ScanError(t *testing.T) {
	repo, mock := newAnalyticsRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT student_id")).
		WithArgs(10, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("bad"))

	_, _, err := repo.GetAtRiskStudents(context.Background(), nil, 10, 0)
	assert.Error(t, err)
}

// ---- GetStudentRisk ----

func TestAnalyticsGetStudentRisk_Success(t *testing.T) {
	repo, mock := newAnalyticsRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT student_id, student_name")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(riskCols).
			AddRow(int64(1), "John", nil, nil, nil, "low", 10.0, nil))

	risk, err := repo.GetStudentRisk(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, entities.RiskLevel("low"), risk.RiskLevel)
}

func TestAnalyticsGetStudentRisk_WithRiskFactors(t *testing.T) {
	repo, mock := newAnalyticsRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT student_id, student_name")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(riskCols).
			AddRow(int64(1), "John", nil, nil, nil, "high", 80.0, []byte(`{"attendance":{"rate":0.5,"absent_count":10,"is_risk":true},"grades":{"average":2.0,"failing_count":3,"is_risk":true}}`)))

	risk, err := repo.GetStudentRisk(context.Background(), 1)
	require.NoError(t, err)
	assert.NotNil(t, risk.RiskFactors)
}

func TestAnalyticsGetStudentRisk_InvalidJSON(t *testing.T) {
	repo, mock := newAnalyticsRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT student_id, student_name")).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows(riskCols).
			AddRow(int64(1), "John", nil, nil, nil, "low", 10.0, []byte(`{bad`)))

	risk, err := repo.GetStudentRisk(context.Background(), 1)
	require.NoError(t, err)
	assert.Nil(t, risk.RiskFactors)
}

func TestAnalyticsGetStudentRisk_NotFound(t *testing.T) {
	repo, mock := newAnalyticsRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT student_id")).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetStudentRisk(context.Background(), 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "student not found")
}

func TestAnalyticsGetStudentRisk_DBError(t *testing.T) {
	repo, mock := newAnalyticsRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT student_id")).
		WithArgs(int64(1)).
		WillReturnError(fmt.Errorf("db error"))

	_, err := repo.GetStudentRisk(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get student risk")
}

// ---- GetGroupSummary ----

func TestAnalyticsGetGroupSummary_Success(t *testing.T) {
	repo, mock := newAnalyticsRepoMock(t)
	cols := []string{"group_name", "total_students", "avg_attendance_rate", "avg_grade", "critical_risk_count", "high_risk_count", "medium_risk_count", "low_risk_count", "at_risk_percentage"}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT group_name")).
		WithArgs("Group1").
		WillReturnRows(sqlmock.NewRows(cols).AddRow("Group1", 30, 0.85, 4.0, 1, 2, 5, 22, 10.0))

	summary, err := repo.GetGroupSummary(context.Background(), "Group1")
	require.NoError(t, err)
	assert.Equal(t, "Group1", summary.GroupName)
	assert.Equal(t, 30, summary.TotalStudents)
}

func TestAnalyticsGetGroupSummary_NotFound(t *testing.T) {
	repo, mock := newAnalyticsRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT group_name")).
		WithArgs("NOPE").
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetGroupSummary(context.Background(), "NOPE")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "group not found")
}

func TestAnalyticsGetGroupSummary_DBError(t *testing.T) {
	repo, mock := newAnalyticsRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT group_name")).
		WithArgs("Group1").
		WillReturnError(fmt.Errorf("db error"))

	_, err := repo.GetGroupSummary(context.Background(), "Group1")
	assert.Error(t, err)
}

// ---- GetAllGroupsSummary ----

func TestAnalyticsGetAllGroupsSummary_Success(t *testing.T) {
	repo, mock := newAnalyticsRepoMock(t)
	cols := []string{"group_name", "total_students", "avg_attendance_rate", "avg_grade", "critical_risk_count", "high_risk_count", "medium_risk_count", "low_risk_count", "at_risk_percentage"}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT group_name")).
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow("G1", 20, 0.9, 4.5, 0, 1, 2, 17, 5.0).
			AddRow("G2", 25, 0.8, 3.5, 2, 3, 5, 15, 20.0))

	summaries, err := repo.GetAllGroupsSummary(context.Background(), nil)
	require.NoError(t, err)
	assert.Len(t, summaries, 2)
}

func TestAnalyticsGetAllGroupsSummary_QueryError(t *testing.T) {
	repo, mock := newAnalyticsRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT group_name")).
		WillReturnError(fmt.Errorf("query error"))

	_, err := repo.GetAllGroupsSummary(context.Background(), nil)
	assert.Error(t, err)
}

func TestAnalyticsGetAllGroupsSummary_ScanError(t *testing.T) {
	repo, mock := newAnalyticsRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT group_name")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("bad"))

	_, err := repo.GetAllGroupsSummary(context.Background(), nil)
	assert.Error(t, err)
}

// ---- GetStudentsByRiskLevel ----

func TestAnalyticsGetStudentsByRiskLevel_Success(t *testing.T) {
	repo, mock := newAnalyticsRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WithArgs(entities.RiskLevelHigh).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT student_id")).
		WithArgs(entities.RiskLevelHigh, 10, 0).
		WillReturnRows(sqlmock.NewRows(riskCols).
			AddRow(int64(1), "John", nil, nil, nil, "high", 80.0, nil))

	students, total, err := repo.GetStudentsByRiskLevel(context.Background(), nil, entities.RiskLevelHigh, 10, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, students, 1)
}

func TestAnalyticsGetStudentsByRiskLevel_WithRiskFactorsJSON(t *testing.T) {
	repo, mock := newAnalyticsRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WithArgs(entities.RiskLevelHigh).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT student_id")).
		WithArgs(entities.RiskLevelHigh, 10, 0).
		WillReturnRows(sqlmock.NewRows(riskCols).
			AddRow(int64(1), "John", nil, nil, nil, "high", 80.0, []byte(`{bad`)))

	students, _, err := repo.GetStudentsByRiskLevel(context.Background(), nil, entities.RiskLevelHigh, 10, 0)
	require.NoError(t, err)
	assert.Nil(t, students[0].RiskFactors)
}

func TestAnalyticsGetStudentsByRiskLevel_CountError(t *testing.T) {
	repo, mock := newAnalyticsRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WithArgs(entities.RiskLevelHigh).
		WillReturnError(fmt.Errorf("count error"))

	_, _, err := repo.GetStudentsByRiskLevel(context.Background(), nil, entities.RiskLevelHigh, 10, 0)
	assert.Error(t, err)
}

func TestAnalyticsGetStudentsByRiskLevel_QueryError(t *testing.T) {
	repo, mock := newAnalyticsRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WithArgs(entities.RiskLevelHigh).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT student_id")).
		WithArgs(entities.RiskLevelHigh, 10, 0).
		WillReturnError(fmt.Errorf("query error"))

	_, _, err := repo.GetStudentsByRiskLevel(context.Background(), nil, entities.RiskLevelHigh, 10, 0)
	assert.Error(t, err)
}

func TestAnalyticsGetStudentsByRiskLevel_ScanError(t *testing.T) {
	repo, mock := newAnalyticsRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WithArgs(entities.RiskLevelHigh).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT student_id")).
		WithArgs(entities.RiskLevelHigh, 10, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("bad"))

	_, _, err := repo.GetStudentsByRiskLevel(context.Background(), nil, entities.RiskLevelHigh, 10, 0)
	assert.Error(t, err)
}

// ---- GetMonthlyAttendanceTrend ----

func TestAnalyticsGetMonthlyAttendanceTrend_Success(t *testing.T) {
	repo, mock := newAnalyticsRepoMock(t)
	now := time.Now()
	cols := []string{"month", "unique_students", "total_records", "present_count", "absent_count", "attendance_rate"}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT month")).
		WithArgs(6).
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow(now, 100, 500, 450, 50, 0.9))

	trends, err := repo.GetMonthlyAttendanceTrend(context.Background(), 6)
	require.NoError(t, err)
	assert.Len(t, trends, 1)
	assert.Equal(t, 0.9, trends[0].AttendanceRate)
}

func TestAnalyticsGetMonthlyAttendanceTrend_QueryError(t *testing.T) {
	repo, mock := newAnalyticsRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT month")).
		WithArgs(6).
		WillReturnError(fmt.Errorf("query error"))

	_, err := repo.GetMonthlyAttendanceTrend(context.Background(), 6)
	assert.Error(t, err)
}

func TestAnalyticsGetMonthlyAttendanceTrend_ScanError(t *testing.T) {
	repo, mock := newAnalyticsRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT month")).
		WithArgs(6).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("bad"))

	_, err := repo.GetMonthlyAttendanceTrend(context.Background(), 6)
	assert.Error(t, err)
}

// ---- Scope-filtered SQL (Cycle 3b): WHERE group_name = ANY($N) ----
//
// Pinning the SQL contract: when a non-nil *TeacherScope is passed, the
// repository MUST push the whitelist into the WHERE clause so that
// pagination COUNT(*) and the data query both reflect the post-filter
// row set.

func TestAnalyticsGetAtRiskStudents_AppliesScopeFilter(t *testing.T) {
	repo, mock := newAnalyticsRepoMock(t)
	scope := entities.NewTeacherScope(7, []string{"ИС-21"})

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM v_at_risk_students WHERE group_name = ANY")).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	mock.ExpectQuery(regexp.QuoteMeta("WHERE group_name = ANY")).
		WithArgs(sqlmock.AnyArg(), 10, 0).
		WillReturnRows(sqlmock.NewRows(riskCols))

	students, total, err := repo.GetAtRiskStudents(context.Background(), scope, 10, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, students)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAnalyticsGetAllGroupsSummary_AppliesScopeFilter(t *testing.T) {
	repo, mock := newAnalyticsRepoMock(t)
	scope := entities.NewTeacherScope(7, []string{"ИС-21", "ПИ-31"})

	mock.ExpectQuery(regexp.QuoteMeta("WHERE group_name = ANY")).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{
			"group_name", "total_students", "avg_attendance_rate", "avg_grade",
			"critical_risk_count", "high_risk_count", "medium_risk_count", "low_risk_count",
			"at_risk_percentage",
		}))

	summaries, err := repo.GetAllGroupsSummary(context.Background(), scope)
	require.NoError(t, err)
	assert.Empty(t, summaries)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAnalyticsGetStudentsByRiskLevel_AppliesScopeFilter(t *testing.T) {
	repo, mock := newAnalyticsRepoMock(t)
	scope := entities.NewTeacherScope(7, []string{"ИС-21"})

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM v_student_risk_score WHERE risk_level = $1 AND group_name = ANY")).
		WithArgs(entities.RiskLevelHigh, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	mock.ExpectQuery(regexp.QuoteMeta("WHERE risk_level = $1 AND group_name = ANY")).
		WithArgs(entities.RiskLevelHigh, sqlmock.AnyArg(), 10, 0).
		WillReturnRows(sqlmock.NewRows(riskCols))

	students, total, err := repo.GetStudentsByRiskLevel(context.Background(), scope, entities.RiskLevelHigh, 10, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, students)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAnalyticsGetAtRiskStudents_NilScopeDoesNotFilter(t *testing.T) {
	repo, mock := newAnalyticsRepoMock(t)

	// nil scope → unchanged legacy SQL without WHERE clause.
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM v_at_risk_students")).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
	mock.ExpectQuery(regexp.QuoteMeta("FROM v_at_risk_students\n\t\tORDER BY")).
		WithArgs(10, 0).
		WillReturnRows(sqlmock.NewRows(riskCols))

	_, _, err := repo.GetAtRiskStudents(context.Background(), nil, 10, 0)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAnalyticsGetAtRiskStudents_EmptyScopeYieldsZeroRows(t *testing.T) {
	repo, mock := newAnalyticsRepoMock(t)
	scope := entities.NewTeacherScope(7, nil) // teacher with no scheduled groups

	// Both queries still run (deny-all is enforced at SQL via '{}'::text[]),
	// but the row set is empty.
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM v_at_risk_students WHERE group_name = ANY")).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
	mock.ExpectQuery(regexp.QuoteMeta("WHERE group_name = ANY")).
		WithArgs(sqlmock.AnyArg(), 10, 0).
		WillReturnRows(sqlmock.NewRows(riskCols))

	students, total, err := repo.GetAtRiskStudents(context.Background(), scope, 10, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total, "empty whitelist must produce zero count (deny-all)")
	assert.Empty(t, students)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---- GetRiskWeightConfig ----

var riskWeightCols = []string{
	"id", "attendance_weight", "grade_weight", "submission_weight", "inactivity_weight",
	"high_risk_threshold", "critical_risk_threshold", "updated_by", "updated_at",
}

func TestAnalyticsGetRiskWeightConfig_Success(t *testing.T) {
	repo, mock := newAnalyticsRepoMock(t)
	updatedBy := int64(42)
	updatedAt := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta("FROM risk_weight_config ORDER BY id LIMIT 1")).
		WillReturnRows(sqlmock.NewRows(riskWeightCols).
			AddRow(1, 0.40, 0.30, 0.20, 0.10, 75.0, 90.0, updatedBy, updatedAt))

	cfg, err := repo.GetRiskWeightConfig(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, cfg.ID)
	assert.Equal(t, 0.40, cfg.AttendanceWeight)
	assert.Equal(t, 75.0, cfg.HighRiskThreshold)
	require.NotNil(t, cfg.UpdatedBy)
	assert.Equal(t, int64(42), *cfg.UpdatedBy)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAnalyticsGetRiskWeightConfig_NotFoundReturnsDefaults(t *testing.T) {
	repo, mock := newAnalyticsRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("FROM risk_weight_config ORDER BY id LIMIT 1")).
		WillReturnError(sql.ErrNoRows)

	cfg, err := repo.GetRiskWeightConfig(context.Background())
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, 0.35, cfg.AttendanceWeight)
	assert.Equal(t, 0.30, cfg.GradeWeight)
	assert.Equal(t, 0.20, cfg.SubmissionWeight)
	assert.Equal(t, 0.15, cfg.InactivityWeight)
	assert.Equal(t, 70.0, cfg.HighRiskThreshold)
	assert.Equal(t, 85.0, cfg.CriticalRiskThreshold)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAnalyticsGetRiskWeightConfig_DBError(t *testing.T) {
	repo, mock := newAnalyticsRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("FROM risk_weight_config ORDER BY id LIMIT 1")).
		WillReturnError(fmt.Errorf("connection refused"))

	_, err := repo.GetRiskWeightConfig(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get risk weight config")
}

// ---- UpdateRiskWeightConfig ----

func TestAnalyticsUpdateRiskWeightConfig_Success(t *testing.T) {
	repo, mock := newAnalyticsRepoMock(t)
	updatedBy := int64(7)
	cfg := &entities.RiskWeightConfig{
		AttendanceWeight:      0.40,
		GradeWeight:           0.30,
		SubmissionWeight:      0.20,
		InactivityWeight:      0.10,
		HighRiskThreshold:     70.0,
		CriticalRiskThreshold: 85.0,
		UpdatedBy:             &updatedBy,
	}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE risk_weight_config SET")).
		WithArgs(0.40, 0.30, 0.20, 0.10, 70.0, 85.0, &updatedBy).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdateRiskWeightConfig(context.Background(), cfg)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAnalyticsUpdateRiskWeightConfig_DBError(t *testing.T) {
	repo, mock := newAnalyticsRepoMock(t)
	cfg := &entities.RiskWeightConfig{AttendanceWeight: 0.5}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE risk_weight_config SET")).
		WillReturnError(fmt.Errorf("constraint violation"))

	err := repo.UpdateRiskWeightConfig(context.Background(), cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update risk weight config")
}

// ---- SaveRiskHistory ----

func TestAnalyticsSaveRiskHistory_WithoutFactors(t *testing.T) {
	repo, mock := newAnalyticsRepoMock(t)
	entry := &entities.RiskHistoryEntry{
		StudentID:    101,
		RiskScore:    72.5,
		RiskLevel:    entities.RiskLevelHigh,
		CalculatedAt: time.Now(),
	}

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO student_risk_history")).
		WithArgs(int64(101), 72.5, entities.RiskLevelHigh, (*float64)(nil), (*float64)(nil), (*float64)(nil), []byte(nil), entry.CalculatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.SaveRiskHistory(context.Background(), entry)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAnalyticsSaveRiskHistory_WithFactors(t *testing.T) {
	repo, mock := newAnalyticsRepoMock(t)
	attendance := 0.85
	grade := 4.2
	submission := 0.9
	entry := &entities.RiskHistoryEntry{
		StudentID:      101,
		RiskScore:      88.0,
		RiskLevel:      entities.RiskLevelCritical,
		AttendanceRate: &attendance,
		GradeAverage:   &grade,
		SubmissionRate: &submission,
		RiskFactors:    &entities.RiskFactors{},
		CalculatedAt:   time.Now(),
	}

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO student_risk_history")).
		WithArgs(int64(101), 88.0, entities.RiskLevelCritical, &attendance, &grade, &submission, sqlmock.AnyArg(), entry.CalculatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.SaveRiskHistory(context.Background(), entry)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAnalyticsSaveRiskHistory_DBError(t *testing.T) {
	repo, mock := newAnalyticsRepoMock(t)
	entry := &entities.RiskHistoryEntry{StudentID: 101, CalculatedAt: time.Now()}

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO student_risk_history")).
		WillReturnError(fmt.Errorf("disk full"))

	err := repo.SaveRiskHistory(context.Background(), entry)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save risk history")
}

// ---- GetStudentRiskHistory ----

var riskHistoryCols = []string{
	"id", "student_id", "risk_score", "risk_level",
	"attendance_rate", "grade_average", "submission_rate",
	"risk_factors", "calculated_at",
}

func TestAnalyticsGetStudentRiskHistory_Success(t *testing.T) {
	repo, mock := newAnalyticsRepoMock(t)
	calculatedAt := time.Now()
	factorsJSON := []byte(`{"days_absent":3}`)

	mock.ExpectQuery(regexp.QuoteMeta("FROM student_risk_history")).
		WithArgs(int64(101), 50).
		WillReturnRows(sqlmock.NewRows(riskHistoryCols).
			AddRow(int64(1), int64(101), 72.5, entities.RiskLevelHigh, nil, nil, nil, factorsJSON, calculatedAt).
			AddRow(int64(2), int64(101), 88.0, entities.RiskLevelCritical, nil, nil, nil, nil, calculatedAt))

	history, err := repo.GetStudentRiskHistory(context.Background(), 101, 50)
	require.NoError(t, err)
	assert.Len(t, history, 2)
	assert.Equal(t, int64(1), history[0].ID)
	assert.NotNil(t, history[0].RiskFactors, "row 1 should parse risk_factors JSON")
	assert.Nil(t, history[1].RiskFactors, "row 2 has nil risk_factors")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAnalyticsGetStudentRiskHistory_InvalidFactorsJSONSilentlySkipped(t *testing.T) {
	repo, mock := newAnalyticsRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("FROM student_risk_history")).
		WithArgs(int64(101), 90).
		WillReturnRows(sqlmock.NewRows(riskHistoryCols).
			AddRow(int64(1), int64(101), 72.5, entities.RiskLevelHigh, nil, nil, nil, []byte("not-json"), time.Now()))

	history, err := repo.GetStudentRiskHistory(context.Background(), 101, 0)
	require.NoError(t, err)
	require.Len(t, history, 1)
	assert.Nil(t, history[0].RiskFactors, "invalid JSON must NOT crash, just leave RiskFactors nil")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAnalyticsGetStudentRiskHistory_LimitClampedToDefault(t *testing.T) {
	tests := []struct {
		name      string
		inLimit   int
		wantQuery int
	}{
		{name: "zero clamped to 90", inLimit: 0, wantQuery: 90},
		{name: "negative clamped to 90", inLimit: -5, wantQuery: 90},
		{name: "over 365 clamped to 90", inLimit: 1000, wantQuery: 90},
		{name: "within bounds passed through", inLimit: 30, wantQuery: 30},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo, mock := newAnalyticsRepoMock(t)
			mock.ExpectQuery(regexp.QuoteMeta("FROM student_risk_history")).
				WithArgs(int64(101), tc.wantQuery).
				WillReturnRows(sqlmock.NewRows(riskHistoryCols))

			_, err := repo.GetStudentRiskHistory(context.Background(), 101, tc.inLimit)
			require.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAnalyticsGetStudentRiskHistory_QueryError(t *testing.T) {
	repo, mock := newAnalyticsRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("FROM student_risk_history")).
		WithArgs(int64(101), 90).
		WillReturnError(fmt.Errorf("connection lost"))

	_, err := repo.GetStudentRiskHistory(context.Background(), 101, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get risk history")
}

func TestAnalyticsGetStudentRiskHistory_ScanError(t *testing.T) {
	repo, mock := newAnalyticsRepoMock(t)

	mock.ExpectQuery(regexp.QuoteMeta("FROM student_risk_history")).
		WithArgs(int64(101), 90).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("not-an-int64"))

	_, err := repo.GetStudentRiskHistory(context.Background(), 101, 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to scan risk history")
}
