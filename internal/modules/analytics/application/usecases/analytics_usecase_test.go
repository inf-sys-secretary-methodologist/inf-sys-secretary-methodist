package usecases_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/analytics/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/analytics/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/analytics/domain/entities"
)

// --- Mock implementations ---

// mockAnalyticsRepository implements repositories.AnalyticsRepository
type mockAnalyticsRepository struct {
	atRiskStudents      []entities.StudentRiskScore
	atRiskTotal         int64
	studentRisk         *entities.StudentRiskScore
	groupSummary        *entities.GroupAnalyticsSummary
	allGroupsSummaries  []entities.GroupAnalyticsSummary
	riskLevelStudents   []entities.StudentRiskScore
	riskLevelTotal      int64
	monthlyTrends       []entities.MonthlyAttendanceTrend
	err                 error
	riskLevelErr        error
	trendErr            error
	groupSummaryErr     error
	allGroupsSummaryErr error
	studentRiskErr      error
}

func (m *mockAnalyticsRepository) GetAtRiskStudents(_ context.Context, _, _ int) ([]entities.StudentRiskScore, int64, error) {
	return m.atRiskStudents, m.atRiskTotal, m.err
}

func (m *mockAnalyticsRepository) GetStudentRisk(_ context.Context, _ int64) (*entities.StudentRiskScore, error) {
	if m.studentRiskErr != nil {
		return nil, m.studentRiskErr
	}
	return m.studentRisk, nil
}

func (m *mockAnalyticsRepository) GetGroupSummary(_ context.Context, _ string) (*entities.GroupAnalyticsSummary, error) {
	if m.groupSummaryErr != nil {
		return nil, m.groupSummaryErr
	}
	return m.groupSummary, nil
}

func (m *mockAnalyticsRepository) GetAllGroupsSummary(_ context.Context) ([]entities.GroupAnalyticsSummary, error) {
	if m.allGroupsSummaryErr != nil {
		return nil, m.allGroupsSummaryErr
	}
	return m.allGroupsSummaries, nil
}

func (m *mockAnalyticsRepository) GetStudentsByRiskLevel(_ context.Context, _ entities.RiskLevel, _, _ int) ([]entities.StudentRiskScore, int64, error) {
	if m.riskLevelErr != nil {
		return nil, 0, m.riskLevelErr
	}
	return m.riskLevelStudents, m.riskLevelTotal, nil
}

func (m *mockAnalyticsRepository) GetMonthlyAttendanceTrend(_ context.Context, _ int) ([]entities.MonthlyAttendanceTrend, error) {
	if m.trendErr != nil {
		return nil, m.trendErr
	}
	return m.monthlyTrends, nil
}

func (m *mockAnalyticsRepository) GetRiskWeightConfig(_ context.Context) (*entities.RiskWeightConfig, error) {
	return nil, nil
}
func (m *mockAnalyticsRepository) UpdateRiskWeightConfig(_ context.Context, _ *entities.RiskWeightConfig) error {
	return nil
}
func (m *mockAnalyticsRepository) SaveRiskHistory(_ context.Context, _ *entities.RiskHistoryEntry) error {
	return nil
}
func (m *mockAnalyticsRepository) GetStudentRiskHistory(_ context.Context, _ int64, _ int) ([]entities.RiskHistoryEntry, error) {
	return nil, nil
}

// mockAttendanceRepository implements repositories.AttendanceRepository
type mockAttendanceRepository struct {
	markErr        error
	bulkMarkErr    error
	lessonRecords  []entities.AttendanceRecord
	lessonErr      error
	createLessonFn func(lesson *entities.Lesson) error
}

func (m *mockAttendanceRepository) CreateLesson(_ context.Context, lesson *entities.Lesson) error {
	if m.createLessonFn != nil {
		return m.createLessonFn(lesson)
	}
	lesson.ID = 42
	return nil
}

func (m *mockAttendanceRepository) GetLessonByID(_ context.Context, _ int64) (*entities.Lesson, error) {
	return nil, nil
}

func (m *mockAttendanceRepository) GetLessonsByGroup(_ context.Context, _ string) ([]entities.Lesson, error) {
	return nil, nil
}

func (m *mockAttendanceRepository) GetLessonsByTeacher(_ context.Context, _ int64) ([]entities.Lesson, error) {
	return nil, nil
}

func (m *mockAttendanceRepository) MarkAttendance(_ context.Context, _ *entities.AttendanceRecord) error {
	return m.markErr
}

func (m *mockAttendanceRepository) BulkMarkAttendance(_ context.Context, _ []entities.AttendanceRecord) error {
	return m.bulkMarkErr
}

func (m *mockAttendanceRepository) GetAttendanceByLesson(_ context.Context, _ int64, _ string) ([]entities.AttendanceRecord, error) {
	if m.lessonErr != nil {
		return nil, m.lessonErr
	}
	return m.lessonRecords, nil
}

func (m *mockAttendanceRepository) GetAttendanceByStudent(_ context.Context, _ int64, _, _ string) ([]entities.AttendanceRecord, error) {
	return nil, nil
}

func (m *mockAttendanceRepository) UpdateAttendance(_ context.Context, _ *entities.AttendanceRecord) error {
	return nil
}

func (m *mockAttendanceRepository) GetStudentAttendanceStats(_ context.Context, _ int64) (*entities.AttendanceStats, error) {
	return nil, nil
}

// mockGradeRepository implements repositories.GradeRepository
type mockGradeRepository struct{}

func (m *mockGradeRepository) CreateGrade(_ context.Context, _ *entities.Grade) error { return nil }
func (m *mockGradeRepository) GetGradesByStudent(_ context.Context, _ int64) ([]entities.Grade, error) {
	return nil, nil
}
func (m *mockGradeRepository) GetGradesBySubject(_ context.Context, _ int64, _ string) ([]entities.Grade, error) {
	return nil, nil
}
func (m *mockGradeRepository) UpdateGrade(_ context.Context, _ *entities.Grade) error { return nil }
func (m *mockGradeRepository) DeleteGrade(_ context.Context, _ int64) error           { return nil }
func (m *mockGradeRepository) GetStudentGradeStats(_ context.Context, _ int64) (*entities.GradeStats, error) {
	return nil, nil
}

// --- Helper functions ---

func newTestUseCase(
	analyticsRepo *mockAnalyticsRepository,
	attendanceRepo *mockAttendanceRepository,
	gradeRepo *mockGradeRepository,
) *usecases.AnalyticsUseCase {
	if analyticsRepo == nil {
		analyticsRepo = &mockAnalyticsRepository{}
	}
	if attendanceRepo == nil {
		attendanceRepo = &mockAttendanceRepository{}
	}
	if gradeRepo == nil {
		gradeRepo = &mockGradeRepository{}
	}
	// Pass nil for auditLogger to test the nil-guard branches
	return usecases.NewAnalyticsUseCase(analyticsRepo, attendanceRepo, gradeRepo, nil)
}

func ptrFloat64(v float64) *float64 { return &v }
func ptrString(v string) *string    { return &v }
func ptrInt64(v int64) *int64       { return &v }

func sampleStudentRisk() entities.StudentRiskScore {
	return entities.StudentRiskScore{
		StudentID:      1,
		StudentName:    "John Doe",
		GroupName:      ptrString("CS-101"),
		AttendanceRate: ptrFloat64(0.65),
		GradeAverage:   ptrFloat64(3.2),
		RiskLevel:      entities.RiskLevelHigh,
		RiskScore:      75.0,
		RiskFactors: &entities.RiskFactors{
			Attendance: entities.AttendanceRiskFactor{Rate: 0.65, AbsentCount: 10, IsRisk: true},
			Grades:     entities.GradesRiskFactor{Average: 3.2, FailingCount: 2, IsRisk: true},
		},
	}
}

// --- Tests for GetAtRiskStudents ---

func TestGetAtRiskStudents_Success(t *testing.T) {
	students := []entities.StudentRiskScore{sampleStudentRisk()}
	repo := &mockAnalyticsRepository{atRiskStudents: students, atRiskTotal: 1}
	uc := newTestUseCase(repo, nil, nil)

	resp, err := uc.GetAtRiskStudents(context.Background(), 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), resp.Total)
	assert.Equal(t, 1, resp.Page)
	assert.Equal(t, 10, resp.PageSize)
	assert.Len(t, resp.Students, 1)
	assert.Equal(t, "John Doe", resp.Students[0].StudentName)
	assert.Equal(t, "high", resp.Students[0].RiskLevel)
}

func TestGetAtRiskStudents_RepoError(t *testing.T) {
	repo := &mockAnalyticsRepository{err: errors.New("db error")}
	uc := newTestUseCase(repo, nil, nil)

	resp, err := uc.GetAtRiskStudents(context.Background(), 1, 10)
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get at-risk students")
}

func TestGetAtRiskStudents_PageNormalization(t *testing.T) {
	repo := &mockAnalyticsRepository{atRiskStudents: []entities.StudentRiskScore{}, atRiskTotal: 0}
	uc := newTestUseCase(repo, nil, nil)

	// page < 1 should normalize to 1
	resp, err := uc.GetAtRiskStudents(context.Background(), 0, 10)
	require.NoError(t, err)
	assert.Equal(t, 1, resp.Page)

	// negative page
	resp, err = uc.GetAtRiskStudents(context.Background(), -5, 10)
	require.NoError(t, err)
	assert.Equal(t, 1, resp.Page)
}

func TestGetAtRiskStudents_PageSizeNormalization(t *testing.T) {
	repo := &mockAnalyticsRepository{atRiskStudents: []entities.StudentRiskScore{}, atRiskTotal: 0}
	uc := newTestUseCase(repo, nil, nil)

	// pageSize < 1 should normalize to 20
	resp, err := uc.GetAtRiskStudents(context.Background(), 1, 0)
	require.NoError(t, err)
	assert.Equal(t, 20, resp.PageSize)

	// pageSize > 100 should normalize to 20
	resp, err = uc.GetAtRiskStudents(context.Background(), 1, 101)
	require.NoError(t, err)
	assert.Equal(t, 20, resp.PageSize)

	// negative pageSize
	resp, err = uc.GetAtRiskStudents(context.Background(), 1, -1)
	require.NoError(t, err)
	assert.Equal(t, 20, resp.PageSize)
}

func TestGetAtRiskStudents_EmptyResult(t *testing.T) {
	repo := &mockAnalyticsRepository{atRiskStudents: []entities.StudentRiskScore{}, atRiskTotal: 0}
	uc := newTestUseCase(repo, nil, nil)

	resp, err := uc.GetAtRiskStudents(context.Background(), 1, 10)
	require.NoError(t, err)
	assert.Empty(t, resp.Students)
	assert.Equal(t, int64(0), resp.Total)
}

func TestGetAtRiskStudents_MultipleStudents(t *testing.T) {
	s1 := sampleStudentRisk()
	s2 := sampleStudentRisk()
	s2.StudentID = 2
	s2.StudentName = "Jane Doe"
	s2.RiskLevel = entities.RiskLevelCritical
	repo := &mockAnalyticsRepository{atRiskStudents: []entities.StudentRiskScore{s1, s2}, atRiskTotal: 2}
	uc := newTestUseCase(repo, nil, nil)

	resp, err := uc.GetAtRiskStudents(context.Background(), 1, 10)
	require.NoError(t, err)
	assert.Len(t, resp.Students, 2)
	assert.Equal(t, "critical", resp.Students[1].RiskLevel)
}

// --- Tests for GetStudentRisk ---

func TestGetStudentRisk_Success(t *testing.T) {
	risk := sampleStudentRisk()
	repo := &mockAnalyticsRepository{studentRisk: &risk}
	uc := newTestUseCase(repo, nil, nil)

	resp, err := uc.GetStudentRisk(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), resp.StudentID)
	assert.Equal(t, "high", resp.RiskLevel)
	assert.Equal(t, 75.0, resp.RiskScore)
	assert.NotNil(t, resp.RiskFactors)
}

func TestGetStudentRisk_Error(t *testing.T) {
	repo := &mockAnalyticsRepository{studentRiskErr: errors.New("not found")}
	uc := newTestUseCase(repo, nil, nil)

	resp, err := uc.GetStudentRisk(context.Background(), 999)
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get student risk")
}

// --- Tests for GetGroupSummary ---

func TestGetGroupSummary_Success(t *testing.T) {
	summary := &entities.GroupAnalyticsSummary{
		GroupName:         "CS-101",
		TotalStudents:     30,
		AvgAttendanceRate: 0.85,
		AvgGrade:          4.0,
		CriticalRiskCount: 1,
		HighRiskCount:     3,
		MediumRiskCount:   5,
		LowRiskCount:      21,
		AtRiskPercentage:  30.0,
	}
	repo := &mockAnalyticsRepository{groupSummary: summary}
	uc := newTestUseCase(repo, nil, nil)

	resp, err := uc.GetGroupSummary(context.Background(), "CS-101")
	require.NoError(t, err)
	assert.Equal(t, "CS-101", resp.GroupName)
	assert.Equal(t, 30, resp.TotalStudents)
	assert.Equal(t, 0.85, resp.AvgAttendanceRate)
	assert.Equal(t, 1, resp.RiskDistribution.Critical)
	assert.Equal(t, 3, resp.RiskDistribution.High)
	assert.Equal(t, 5, resp.RiskDistribution.Medium)
	assert.Equal(t, 21, resp.RiskDistribution.Low)
}

func TestGetGroupSummary_Error(t *testing.T) {
	repo := &mockAnalyticsRepository{groupSummaryErr: errors.New("group not found")}
	uc := newTestUseCase(repo, nil, nil)

	resp, err := uc.GetGroupSummary(context.Background(), "INVALID")
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get group summary")
}

// --- Tests for GetAllGroupsSummary ---

func TestGetAllGroupsSummary_Success(t *testing.T) {
	summaries := []entities.GroupAnalyticsSummary{
		{GroupName: "CS-101", TotalStudents: 30, AvgAttendanceRate: 0.85, AvgGrade: 4.0},
		{GroupName: "CS-102", TotalStudents: 25, AvgAttendanceRate: 0.78, AvgGrade: 3.5},
	}
	repo := &mockAnalyticsRepository{allGroupsSummaries: summaries}
	uc := newTestUseCase(repo, nil, nil)

	resp, err := uc.GetAllGroupsSummary(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 2, resp.Total)
	assert.Len(t, resp.Groups, 2)
	assert.Equal(t, "CS-101", resp.Groups[0].GroupName)
	assert.Equal(t, "CS-102", resp.Groups[1].GroupName)
}

func TestGetAllGroupsSummary_Empty(t *testing.T) {
	repo := &mockAnalyticsRepository{allGroupsSummaries: []entities.GroupAnalyticsSummary{}}
	uc := newTestUseCase(repo, nil, nil)

	resp, err := uc.GetAllGroupsSummary(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Total)
	assert.Empty(t, resp.Groups)
}

func TestGetAllGroupsSummary_Error(t *testing.T) {
	repo := &mockAnalyticsRepository{allGroupsSummaryErr: errors.New("db error")}
	uc := newTestUseCase(repo, nil, nil)

	resp, err := uc.GetAllGroupsSummary(context.Background())
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get all groups summary")
}

// --- Tests for GetStudentsByRiskLevel ---

func TestGetStudentsByRiskLevel_Success(t *testing.T) {
	s := sampleStudentRisk()
	repo := &mockAnalyticsRepository{riskLevelStudents: []entities.StudentRiskScore{s}, riskLevelTotal: 1}
	uc := newTestUseCase(repo, nil, nil)

	resp, err := uc.GetStudentsByRiskLevel(context.Background(), "high", 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), resp.Total)
	assert.Len(t, resp.Students, 1)
	assert.Equal(t, 1, resp.Page)
	assert.Equal(t, 10, resp.PageSize)
}

func TestGetStudentsByRiskLevel_PageNormalization(t *testing.T) {
	repo := &mockAnalyticsRepository{riskLevelStudents: []entities.StudentRiskScore{}, riskLevelTotal: 0}
	uc := newTestUseCase(repo, nil, nil)

	resp, err := uc.GetStudentsByRiskLevel(context.Background(), "high", -1, 10)
	require.NoError(t, err)
	assert.Equal(t, 1, resp.Page)
}

func TestGetStudentsByRiskLevel_PageSizeNormalization(t *testing.T) {
	repo := &mockAnalyticsRepository{riskLevelStudents: []entities.StudentRiskScore{}, riskLevelTotal: 0}
	uc := newTestUseCase(repo, nil, nil)

	// pageSize > 100
	resp, err := uc.GetStudentsByRiskLevel(context.Background(), "critical", 1, 200)
	require.NoError(t, err)
	assert.Equal(t, 20, resp.PageSize)

	// pageSize < 1
	resp, err = uc.GetStudentsByRiskLevel(context.Background(), "critical", 1, 0)
	require.NoError(t, err)
	assert.Equal(t, 20, resp.PageSize)
}

func TestGetStudentsByRiskLevel_Error(t *testing.T) {
	repo := &mockAnalyticsRepository{riskLevelErr: errors.New("db error")}
	uc := newTestUseCase(repo, nil, nil)

	resp, err := uc.GetStudentsByRiskLevel(context.Background(), "high", 1, 10)
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get students by risk level")
}

// --- Tests for GetAttendanceTrend ---

func TestGetAttendanceTrend_Success(t *testing.T) {
	trends := []entities.MonthlyAttendanceTrend{
		{Month: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), UniqueStudents: 100, TotalRecords: 500, PresentCount: 450, AbsentCount: 50, AttendanceRate: 0.9},
		{Month: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC), UniqueStudents: 98, TotalRecords: 490, PresentCount: 430, AbsentCount: 60, AttendanceRate: 0.88},
	}
	repo := &mockAnalyticsRepository{monthlyTrends: trends}
	uc := newTestUseCase(repo, nil, nil)

	resp, err := uc.GetAttendanceTrend(context.Background(), 6)
	require.NoError(t, err)
	assert.Equal(t, 6, resp.Months)
	assert.Len(t, resp.Trends, 2)
	assert.Equal(t, "2024-01", resp.Trends[0].Month)
	assert.Equal(t, "2024-02", resp.Trends[1].Month)
	assert.Equal(t, 100, resp.Trends[0].UniqueStudents)
}

func TestGetAttendanceTrend_MonthsNormalization(t *testing.T) {
	repo := &mockAnalyticsRepository{monthlyTrends: []entities.MonthlyAttendanceTrend{}}
	uc := newTestUseCase(repo, nil, nil)

	// months < 1 normalizes to 6
	resp, err := uc.GetAttendanceTrend(context.Background(), 0)
	require.NoError(t, err)
	assert.Equal(t, 6, resp.Months)

	// months > 24 normalizes to 6
	resp, err = uc.GetAttendanceTrend(context.Background(), 25)
	require.NoError(t, err)
	assert.Equal(t, 6, resp.Months)

	// negative months
	resp, err = uc.GetAttendanceTrend(context.Background(), -1)
	require.NoError(t, err)
	assert.Equal(t, 6, resp.Months)
}

func TestGetAttendanceTrend_ValidBoundary(t *testing.T) {
	repo := &mockAnalyticsRepository{monthlyTrends: []entities.MonthlyAttendanceTrend{}}
	uc := newTestUseCase(repo, nil, nil)

	// months = 1 (lower boundary, valid)
	resp, err := uc.GetAttendanceTrend(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, 1, resp.Months)

	// months = 24 (upper boundary, valid)
	resp, err = uc.GetAttendanceTrend(context.Background(), 24)
	require.NoError(t, err)
	assert.Equal(t, 24, resp.Months)
}

func TestGetAttendanceTrend_Error(t *testing.T) {
	repo := &mockAnalyticsRepository{trendErr: errors.New("db error")}
	uc := newTestUseCase(repo, nil, nil)

	resp, err := uc.GetAttendanceTrend(context.Background(), 6)
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get attendance trend")
}

// --- Tests for MarkAttendance ---

func TestMarkAttendance_Success(t *testing.T) {
	attendanceRepo := &mockAttendanceRepository{}
	uc := newTestUseCase(nil, attendanceRepo, nil)

	req := &dto.MarkAttendanceRequest{
		StudentID:  1,
		LessonID:   10,
		LessonDate: "2024-03-15",
		Status:     "present",
		Notes:      "on time",
	}
	markedBy := int64(5)

	resp, err := uc.MarkAttendance(context.Background(), req, markedBy)
	require.NoError(t, err)
	assert.Equal(t, int64(1), resp.StudentID)
	assert.Equal(t, int64(10), resp.LessonID)
	assert.Equal(t, "2024-03-15", resp.LessonDate)
	assert.Equal(t, "present", resp.Status)
	assert.NotNil(t, resp.MarkedBy)
	assert.Equal(t, int64(5), *resp.MarkedBy)
	assert.NotNil(t, resp.Notes)
	assert.Equal(t, "on time", *resp.Notes)
}

func TestMarkAttendance_WithoutNotes(t *testing.T) {
	attendanceRepo := &mockAttendanceRepository{}
	uc := newTestUseCase(nil, attendanceRepo, nil)

	req := &dto.MarkAttendanceRequest{
		StudentID:  1,
		LessonID:   10,
		LessonDate: "2024-03-15",
		Status:     "absent",
		Notes:      "", // empty notes
	}

	resp, err := uc.MarkAttendance(context.Background(), req, 5)
	require.NoError(t, err)
	assert.Nil(t, resp.Notes)
	assert.Equal(t, "absent", resp.Status)
}

func TestMarkAttendance_InvalidDate(t *testing.T) {
	uc := newTestUseCase(nil, nil, nil)

	req := &dto.MarkAttendanceRequest{
		StudentID:  1,
		LessonID:   10,
		LessonDate: "not-a-date",
		Status:     "present",
	}

	resp, err := uc.MarkAttendance(context.Background(), req, 5)
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid lesson date format")
}

func TestMarkAttendance_RepoError(t *testing.T) {
	attendanceRepo := &mockAttendanceRepository{markErr: errors.New("db error")}
	uc := newTestUseCase(nil, attendanceRepo, nil)

	req := &dto.MarkAttendanceRequest{
		StudentID:  1,
		LessonID:   10,
		LessonDate: "2024-03-15",
		Status:     "present",
	}

	resp, err := uc.MarkAttendance(context.Background(), req, 5)
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to mark attendance")
}

// --- Tests for BulkMarkAttendance ---

func TestBulkMarkAttendance_Success(t *testing.T) {
	attendanceRepo := &mockAttendanceRepository{}
	uc := newTestUseCase(nil, attendanceRepo, nil)

	req := &dto.BulkMarkAttendanceRequest{
		LessonID:   10,
		LessonDate: "2024-03-15",
		Records: []dto.BulkAttendanceRecord{
			{StudentID: 1, Status: "present", Notes: "good"},
			{StudentID: 2, Status: "absent", Notes: ""},
			{StudentID: 3, Status: "late", Notes: "5 minutes late"},
		},
	}

	resp, err := uc.BulkMarkAttendance(context.Background(), req, 5)
	require.NoError(t, err)
	assert.Len(t, resp, 3)
	assert.Equal(t, "present", resp[0].Status)
	assert.NotNil(t, resp[0].Notes)
	assert.Equal(t, "good", *resp[0].Notes)
	assert.Equal(t, "absent", resp[1].Status)
	assert.Nil(t, resp[1].Notes) // empty notes should be nil
	assert.Equal(t, "late", resp[2].Status)
	assert.NotNil(t, resp[2].Notes)
}

func TestBulkMarkAttendance_InvalidDate(t *testing.T) {
	uc := newTestUseCase(nil, nil, nil)

	req := &dto.BulkMarkAttendanceRequest{
		LessonID:   10,
		LessonDate: "invalid",
		Records:    []dto.BulkAttendanceRecord{{StudentID: 1, Status: "present"}},
	}

	resp, err := uc.BulkMarkAttendance(context.Background(), req, 5)
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid lesson date format")
}

func TestBulkMarkAttendance_RepoError(t *testing.T) {
	attendanceRepo := &mockAttendanceRepository{bulkMarkErr: errors.New("bulk insert failed")}
	uc := newTestUseCase(nil, attendanceRepo, nil)

	req := &dto.BulkMarkAttendanceRequest{
		LessonID:   10,
		LessonDate: "2024-03-15",
		Records:    []dto.BulkAttendanceRecord{{StudentID: 1, Status: "present"}},
	}

	resp, err := uc.BulkMarkAttendance(context.Background(), req, 5)
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to bulk mark attendance")
}

func TestBulkMarkAttendance_EmptyRecords(t *testing.T) {
	attendanceRepo := &mockAttendanceRepository{}
	uc := newTestUseCase(nil, attendanceRepo, nil)

	req := &dto.BulkMarkAttendanceRequest{
		LessonID:   10,
		LessonDate: "2024-03-15",
		Records:    []dto.BulkAttendanceRecord{},
	}

	resp, err := uc.BulkMarkAttendance(context.Background(), req, 5)
	require.NoError(t, err)
	assert.Empty(t, resp)
}

// --- Tests for GetLessonAttendance ---

func TestGetLessonAttendance_Success(t *testing.T) {
	now := time.Now()
	records := []entities.AttendanceRecord{
		{ID: 1, StudentID: 1, LessonID: 10, LessonDate: now, Status: entities.AttendanceStatusPresent, MarkedBy: ptrInt64(5)},
		{ID: 2, StudentID: 2, LessonID: 10, LessonDate: now, Status: entities.AttendanceStatusAbsent},
		{ID: 3, StudentID: 3, LessonID: 10, LessonDate: now, Status: entities.AttendanceStatusLate},
		{ID: 4, StudentID: 4, LessonID: 10, LessonDate: now, Status: entities.AttendanceStatusExcused},
	}
	attendanceRepo := &mockAttendanceRepository{lessonRecords: records}
	uc := newTestUseCase(nil, attendanceRepo, nil)

	resp, err := uc.GetLessonAttendance(context.Background(), 10, "2024-03-15")
	require.NoError(t, err)
	assert.Equal(t, int64(10), resp.LessonID)
	assert.Equal(t, "2024-03-15", resp.LessonDate)
	assert.Len(t, resp.Records, 4)
	assert.Equal(t, 4, resp.Summary.Total)
	assert.Equal(t, 1, resp.Summary.Present)
	assert.Equal(t, 1, resp.Summary.Absent)
	assert.Equal(t, 1, resp.Summary.Late)
	assert.Equal(t, 1, resp.Summary.Excused)
}

func TestGetLessonAttendance_Empty(t *testing.T) {
	attendanceRepo := &mockAttendanceRepository{lessonRecords: []entities.AttendanceRecord{}}
	uc := newTestUseCase(nil, attendanceRepo, nil)

	resp, err := uc.GetLessonAttendance(context.Background(), 10, "2024-03-15")
	require.NoError(t, err)
	assert.Empty(t, resp.Records)
	assert.Equal(t, 0, resp.Summary.Total)
}

func TestGetLessonAttendance_Error(t *testing.T) {
	attendanceRepo := &mockAttendanceRepository{lessonErr: errors.New("db error")}
	uc := newTestUseCase(nil, attendanceRepo, nil)

	resp, err := uc.GetLessonAttendance(context.Background(), 10, "2024-03-15")
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get lesson attendance")
}

func TestGetLessonAttendance_AllPresent(t *testing.T) {
	now := time.Now()
	records := []entities.AttendanceRecord{
		{ID: 1, StudentID: 1, LessonID: 10, LessonDate: now, Status: entities.AttendanceStatusPresent},
		{ID: 2, StudentID: 2, LessonID: 10, LessonDate: now, Status: entities.AttendanceStatusPresent},
	}
	attendanceRepo := &mockAttendanceRepository{lessonRecords: records}
	uc := newTestUseCase(nil, attendanceRepo, nil)

	resp, err := uc.GetLessonAttendance(context.Background(), 10, "2024-03-15")
	require.NoError(t, err)
	assert.Equal(t, 2, resp.Summary.Present)
	assert.Equal(t, 0, resp.Summary.Absent)
	assert.Equal(t, 0, resp.Summary.Late)
	assert.Equal(t, 0, resp.Summary.Excused)
	assert.Equal(t, 2, resp.Summary.Total)
}

// --- Tests for CreateLesson ---

func TestCreateLesson_Success_DefaultType(t *testing.T) {
	attendanceRepo := &mockAttendanceRepository{}
	uc := newTestUseCase(nil, attendanceRepo, nil)

	req := &dto.CreateLessonRequest{
		Name:       "Math 101",
		Subject:    "Mathematics",
		TeacherID:  ptrInt64(1),
		GroupName:  ptrString("CS-101"),
		LessonType: "", // empty should default to lecture
	}

	lesson, err := uc.CreateLesson(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, int64(42), lesson.ID)
	assert.Equal(t, "Math 101", lesson.Name)
	assert.Equal(t, "Mathematics", lesson.Subject)
	assert.Equal(t, entities.LessonTypeLecture, lesson.LessonType)
}

func TestCreateLesson_Success_CustomType(t *testing.T) {
	attendanceRepo := &mockAttendanceRepository{}
	uc := newTestUseCase(nil, attendanceRepo, nil)

	req := &dto.CreateLessonRequest{
		Name:       "Lab Session",
		Subject:    "Physics",
		LessonType: "lab",
	}

	lesson, err := uc.CreateLesson(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, entities.LessonTypeLab, lesson.LessonType)
}

func TestCreateLesson_Error(t *testing.T) {
	attendanceRepo := &mockAttendanceRepository{
		createLessonFn: func(_ *entities.Lesson) error {
			return errors.New("duplicate lesson")
		},
	}
	uc := newTestUseCase(nil, attendanceRepo, nil)

	req := &dto.CreateLessonRequest{
		Name:    "Math 101",
		Subject: "Mathematics",
	}

	lesson, err := uc.CreateLesson(context.Background(), req)
	assert.Nil(t, lesson)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create lesson")
}

// --- Tests for NewAnalyticsUseCase constructor ---

func TestNewAnalyticsUseCase(t *testing.T) {
	analyticsRepo := &mockAnalyticsRepository{}
	attendanceRepo := &mockAttendanceRepository{}
	gradeRepo := &mockGradeRepository{}

	uc := usecases.NewAnalyticsUseCase(analyticsRepo, attendanceRepo, gradeRepo, nil)
	assert.NotNil(t, uc)
}

// --- Tests with valid page/pageSize at boundary ---

func TestGetAtRiskStudents_PageSizeBoundary(t *testing.T) {
	repo := &mockAnalyticsRepository{atRiskStudents: []entities.StudentRiskScore{}, atRiskTotal: 0}
	uc := newTestUseCase(repo, nil, nil)

	// pageSize exactly 100 is valid
	resp, err := uc.GetAtRiskStudents(context.Background(), 1, 100)
	require.NoError(t, err)
	assert.Equal(t, 100, resp.PageSize)

	// pageSize exactly 1 is valid
	resp, err = uc.GetAtRiskStudents(context.Background(), 1, 1)
	require.NoError(t, err)
	assert.Equal(t, 1, resp.PageSize)
}

func TestGetStudentsByRiskLevel_PageSizeBoundary(t *testing.T) {
	repo := &mockAnalyticsRepository{riskLevelStudents: []entities.StudentRiskScore{}, riskLevelTotal: 0}
	uc := newTestUseCase(repo, nil, nil)

	// pageSize exactly 100 is valid
	resp, err := uc.GetStudentsByRiskLevel(context.Background(), "low", 1, 100)
	require.NoError(t, err)
	assert.Equal(t, 100, resp.PageSize)

	// pageSize exactly 1 is valid
	resp, err = uc.GetStudentsByRiskLevel(context.Background(), "low", 1, 1)
	require.NoError(t, err)
	assert.Equal(t, 1, resp.PageSize)
}

// --- Test risk level student with nil optional fields ---

func TestGetStudentRisk_NilOptionalFields(t *testing.T) {
	risk := entities.StudentRiskScore{
		StudentID:   2,
		StudentName: "Minimal Student",
		RiskLevel:   entities.RiskLevelLow,
		RiskScore:   10.0,
		// GroupName, AttendanceRate, GradeAverage, RiskFactors all nil
	}
	repo := &mockAnalyticsRepository{studentRisk: &risk}
	uc := newTestUseCase(repo, nil, nil)

	resp, err := uc.GetStudentRisk(context.Background(), 2)
	require.NoError(t, err)
	assert.Nil(t, resp.GroupName)
	assert.Nil(t, resp.AttendanceRate)
	assert.Nil(t, resp.GradeAverage)
	assert.Nil(t, resp.RiskFactors)
	assert.Equal(t, "low", resp.RiskLevel)
}

// --- Test pagination offset calculation ---

func TestGetAtRiskStudents_PaginationOffset(t *testing.T) {
	// We can verify indirectly by checking the response page values
	repo := &mockAnalyticsRepository{atRiskStudents: []entities.StudentRiskScore{}, atRiskTotal: 50}
	uc := newTestUseCase(repo, nil, nil)

	resp, err := uc.GetAtRiskStudents(context.Background(), 3, 10)
	require.NoError(t, err)
	assert.Equal(t, 3, resp.Page)
	assert.Equal(t, 10, resp.PageSize)
	assert.Equal(t, int64(50), resp.Total)
}

// --- Test MarkAttendance with all status types ---

func TestMarkAttendance_AllStatusTypes(t *testing.T) {
	statuses := []string{"present", "absent", "late", "excused"}

	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			attendanceRepo := &mockAttendanceRepository{}
			uc := newTestUseCase(nil, attendanceRepo, nil)

			req := &dto.MarkAttendanceRequest{
				StudentID:  1,
				LessonID:   10,
				LessonDate: "2024-03-15",
				Status:     status,
			}

			resp, err := uc.MarkAttendance(context.Background(), req, 5)
			require.NoError(t, err)
			assert.Equal(t, status, resp.Status)
		})
	}
}

// --- Test CreateLesson with all lesson types ---

func TestCreateLesson_AllLessonTypes(t *testing.T) {
	types := []string{"lecture", "practice", "lab", "seminar", "exam"}

	for _, lt := range types {
		t.Run(lt, func(t *testing.T) {
			attendanceRepo := &mockAttendanceRepository{}
			uc := newTestUseCase(nil, attendanceRepo, nil)

			req := &dto.CreateLessonRequest{
				Name:       "Test Lesson",
				Subject:    "Test Subject",
				LessonType: lt,
			}

			lesson, err := uc.CreateLesson(context.Background(), req)
			require.NoError(t, err)
			assert.Equal(t, entities.LessonType(lt), lesson.LessonType)
		})
	}
}
