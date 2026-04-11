package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/entities"
	analyticsEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/analytics/domain/entities"
	dashboardRepos "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/dashboard/domain/repositories"
)

// --- Mock DashboardRepository (struct-based, following project pattern) ---

type mockDashboardRepo struct {
	documentsCount *dashboardRepos.CountResult
	err            error
}

func (m *mockDashboardRepo) GetDocumentsCount(_ context.Context, _ int) (*dashboardRepos.CountResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.documentsCount, nil
}

func (m *mockDashboardRepo) GetReportsCount(_ context.Context, _ int) (*dashboardRepos.CountResult, error) {
	return &dashboardRepos.CountResult{}, nil
}

func (m *mockDashboardRepo) GetTasksCount(_ context.Context, _ int) (*dashboardRepos.CountResult, error) {
	return &dashboardRepos.CountResult{}, nil
}

func (m *mockDashboardRepo) GetEventsCount(_ context.Context, _ int) (*dashboardRepos.CountResult, error) {
	return &dashboardRepos.CountResult{}, nil
}

func (m *mockDashboardRepo) GetStudentsCount(_ context.Context, _ int) (*dashboardRepos.CountResult, error) {
	return &dashboardRepos.CountResult{}, nil
}

func (m *mockDashboardRepo) GetDocumentsTrend(_ context.Context, _, _ time.Time) ([]dashboardRepos.TrendData, error) {
	return nil, nil
}

func (m *mockDashboardRepo) GetReportsTrend(_ context.Context, _, _ time.Time) ([]dashboardRepos.TrendData, error) {
	return nil, nil
}

func (m *mockDashboardRepo) GetTasksTrend(_ context.Context, _, _ time.Time) ([]dashboardRepos.TrendData, error) {
	return nil, nil
}

func (m *mockDashboardRepo) GetEventsTrend(_ context.Context, _, _ time.Time) ([]dashboardRepos.TrendData, error) {
	return nil, nil
}

func (m *mockDashboardRepo) GetRecentActivity(_ context.Context, _ int) ([]dashboardRepos.ActivityData, int64, error) {
	return nil, 0, nil
}

// --- Mock AnalyticsRepository (struct-based) ---

type mockAnalyticsRepo struct {
	atRiskStudents  []analyticsEntities.StudentRiskScore
	totalAtRisk     int64
	attendanceTrend []analyticsEntities.MonthlyAttendanceTrend
	err             error
}

func (m *mockAnalyticsRepo) GetAtRiskStudents(_ context.Context, _, _ int) ([]analyticsEntities.StudentRiskScore, int64, error) {
	if m.err != nil {
		return nil, 0, m.err
	}
	return m.atRiskStudents, m.totalAtRisk, nil
}

func (m *mockAnalyticsRepo) GetStudentRisk(_ context.Context, _ int64) (*analyticsEntities.StudentRiskScore, error) {
	return nil, nil
}

func (m *mockAnalyticsRepo) GetGroupSummary(_ context.Context, _ string) (*analyticsEntities.GroupAnalyticsSummary, error) {
	return nil, nil
}

func (m *mockAnalyticsRepo) GetAllGroupsSummary(_ context.Context) ([]analyticsEntities.GroupAnalyticsSummary, error) {
	return nil, nil
}

func (m *mockAnalyticsRepo) GetStudentsByRiskLevel(_ context.Context, _ analyticsEntities.RiskLevel, _, _ int) ([]analyticsEntities.StudentRiskScore, int64, error) {
	return nil, 0, nil
}

func (m *mockAnalyticsRepo) GetRiskWeightConfig(_ context.Context) (*analyticsEntities.RiskWeightConfig, error) {
	return nil, nil
}
func (m *mockAnalyticsRepo) UpdateRiskWeightConfig(_ context.Context, _ *analyticsEntities.RiskWeightConfig) error {
	return nil
}
func (m *mockAnalyticsRepo) SaveRiskHistory(_ context.Context, _ *analyticsEntities.RiskHistoryEntry) error {
	return nil
}
func (m *mockAnalyticsRepo) GetStudentRiskHistory(_ context.Context, _ int64, _ int) ([]analyticsEntities.RiskHistoryEntry, error) {
	return nil, nil
}

func (m *mockAnalyticsRepo) GetMonthlyAttendanceTrend(_ context.Context, _ int) ([]analyticsEntities.MonthlyAttendanceTrend, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.attendanceTrend, nil
}

// --- Tests ---

func TestComputeMood_Panicking(t *testing.T) {
	dashRepo := &mockDashboardRepo{
		documentsCount: &dashboardRepos.CountResult{Total: 115, PreviousTotal: 100}, // overdue = 15
	}
	analyticsRepo := &mockAnalyticsRepo{
		atRiskStudents: makeCriticalStudents(6),
		totalAtRisk:    10,
	}
	pp := &mockPersonalityProvider{}
	uc := NewMoodUseCase(dashRepo, analyticsRepo, nil, pp)

	mood, err := uc.ComputeMood(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, entities.MoodPanicking, mood.State)
	assert.Equal(t, 1.0, mood.Intensity)
}

func TestComputeMood_Stressed(t *testing.T) {
	dashRepo := &mockDashboardRepo{
		documentsCount: &dashboardRepos.CountResult{Total: 107, PreviousTotal: 100}, // overdue = 7
	}
	analyticsRepo := &mockAnalyticsRepo{
		atRiskStudents: makeCriticalStudents(4),
		totalAtRisk:    6,
	}
	pp := &mockPersonalityProvider{}
	uc := NewMoodUseCase(dashRepo, analyticsRepo, nil, pp)

	mood, err := uc.ComputeMood(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, entities.MoodStressed, mood.State)
	assert.Equal(t, 0.8, mood.Intensity)
}

func TestComputeMood_Happy(t *testing.T) {
	dashRepo := &mockDashboardRepo{
		documentsCount: &dashboardRepos.CountResult{Total: 100, PreviousTotal: 100}, // overdue = 0
	}
	analyticsRepo := &mockAnalyticsRepo{
		atRiskStudents: []analyticsEntities.StudentRiskScore{},
		totalAtRisk:    0,
		attendanceTrend: []analyticsEntities.MonthlyAttendanceTrend{
			{AttendanceRate: 80},
			{AttendanceRate: 80}, // stable, not improving
		},
	}
	pp := &mockPersonalityProvider{}
	uc := NewMoodUseCase(dashRepo, analyticsRepo, nil, pp)

	mood, err := uc.ComputeMood(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, entities.MoodHappy, mood.State)
	assert.Equal(t, 0.8, mood.Intensity)
}

func TestComputeMood_Inspired(t *testing.T) {
	dashRepo := &mockDashboardRepo{
		documentsCount: &dashboardRepos.CountResult{Total: 100, PreviousTotal: 100}, // overdue = 0
	}
	analyticsRepo := &mockAnalyticsRepo{
		atRiskStudents: []analyticsEntities.StudentRiskScore{},
		totalAtRisk:    0,
		attendanceTrend: []analyticsEntities.MonthlyAttendanceTrend{
			{AttendanceRate: 70},
			{AttendanceRate: 85}, // improving (+15 > +2 threshold)
		},
	}
	pp := &mockPersonalityProvider{}
	uc := NewMoodUseCase(dashRepo, analyticsRepo, nil, pp)

	mood, err := uc.ComputeMood(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, entities.MoodInspired, mood.State)
	assert.Equal(t, 0.9, mood.Intensity)
}

func TestComputeMood_DefaultContent(t *testing.T) {
	dashRepo := &mockDashboardRepo{
		documentsCount: &dashboardRepos.CountResult{Total: 102, PreviousTotal: 100}, // overdue = 2 (small)
	}
	analyticsRepo := &mockAnalyticsRepo{
		atRiskStudents: []analyticsEntities.StudentRiskScore{
			{RiskLevel: analyticsEntities.RiskLevelLow},
		},
		totalAtRisk: 1,
		attendanceTrend: []analyticsEntities.MonthlyAttendanceTrend{
			{AttendanceRate: 80},
			{AttendanceRate: 80}, // stable
		},
	}
	pp := &mockPersonalityProvider{}
	uc := NewMoodUseCase(dashRepo, analyticsRepo, nil, pp)

	mood, err := uc.ComputeMood(context.Background())

	assert.NoError(t, err)
	// State depends on time of day / day of week, but it should not be panicking/stressed/happy/inspired
	assert.NotEqual(t, entities.MoodPanicking, mood.State)
	assert.NotEqual(t, entities.MoodStressed, mood.State)
	// Could be content, relaxed, or worried depending on current time
	assert.Contains(t, []entities.MoodState{
		entities.MoodContent,
		entities.MoodRelaxed,
		entities.MoodWorried,
	}, mood.State)
}

func TestComputeMood_DashboardRepoError(t *testing.T) {
	dashRepo := &mockDashboardRepo{
		err: errors.New("database connection lost"),
	}
	analyticsRepo := &mockAnalyticsRepo{
		atRiskStudents: []analyticsEntities.StudentRiskScore{},
		totalAtRisk:    0,
	}
	pp := &mockPersonalityProvider{}
	uc := NewMoodUseCase(dashRepo, analyticsRepo, nil, pp)

	mood, err := uc.ComputeMood(context.Background())

	// Should not panic, should still return a mood
	assert.NoError(t, err)
	assert.NotNil(t, mood)
	assert.Equal(t, 0, mood.OverdueDocuments) // graceful degradation
}

func TestComputeMood_AnalyticsRepoError(t *testing.T) {
	dashRepo := &mockDashboardRepo{
		documentsCount: &dashboardRepos.CountResult{Total: 100, PreviousTotal: 100},
	}
	analyticsRepo := &mockAnalyticsRepo{
		err: errors.New("analytics service unavailable"),
	}
	pp := &mockPersonalityProvider{}
	uc := NewMoodUseCase(dashRepo, analyticsRepo, nil, pp)

	mood, err := uc.ComputeMood(context.Background())

	// Should not panic, returns default mood
	assert.NoError(t, err)
	assert.NotNil(t, mood)
}

func TestGetCurrentMood_NilCache(t *testing.T) {
	dashRepo := &mockDashboardRepo{
		documentsCount: &dashboardRepos.CountResult{Total: 100, PreviousTotal: 100},
	}
	analyticsRepo := &mockAnalyticsRepo{
		atRiskStudents: []analyticsEntities.StudentRiskScore{},
		totalAtRisk:    0,
	}
	pp := &mockPersonalityProvider{}
	uc := NewMoodUseCase(dashRepo, analyticsRepo, nil, pp)

	response, err := uc.GetCurrentMood(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.State)
	assert.NotEmpty(t, response.Message)
	assert.NotEmpty(t, response.Greeting)
}

// --- Helpers ---

func makeCriticalStudents(count int) []analyticsEntities.StudentRiskScore {
	students := make([]analyticsEntities.StudentRiskScore, count)
	for i := range count {
		students[i] = analyticsEntities.StudentRiskScore{
			StudentID: int64(i + 1),
			RiskLevel: analyticsEntities.RiskLevelCritical,
		}
	}
	return students
}
