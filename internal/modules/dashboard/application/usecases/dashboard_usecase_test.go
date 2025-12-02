package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/dashboard/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/dashboard/domain/repositories"
)

// MockDashboardRepository is a mock implementation of DashboardRepository
type MockDashboardRepository struct {
	documentsCount *repositories.CountResult
	reportsCount   *repositories.CountResult
	tasksCount     *repositories.CountResult
	eventsCount    *repositories.CountResult
	studentsCount  *repositories.CountResult
	documentsTrend []repositories.TrendData
	reportsTrend   []repositories.TrendData
	tasksTrend     []repositories.TrendData
	eventsTrend    []repositories.TrendData
	activities     []repositories.ActivityData
	activityTotal  int64
	err            error
}

func (m *MockDashboardRepository) GetDocumentsCount(_ context.Context, _ int) (*repositories.CountResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.documentsCount, nil
}

func (m *MockDashboardRepository) GetReportsCount(_ context.Context, _ int) (*repositories.CountResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.reportsCount, nil
}

func (m *MockDashboardRepository) GetTasksCount(_ context.Context, _ int) (*repositories.CountResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.tasksCount, nil
}

func (m *MockDashboardRepository) GetEventsCount(_ context.Context, _ int) (*repositories.CountResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.eventsCount, nil
}

func (m *MockDashboardRepository) GetStudentsCount(_ context.Context, _ int) (*repositories.CountResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.studentsCount, nil
}

func (m *MockDashboardRepository) GetDocumentsTrend(_ context.Context, _, _ time.Time) ([]repositories.TrendData, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.documentsTrend, nil
}

func (m *MockDashboardRepository) GetReportsTrend(_ context.Context, _, _ time.Time) ([]repositories.TrendData, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.reportsTrend, nil
}

func (m *MockDashboardRepository) GetTasksTrend(_ context.Context, _, _ time.Time) ([]repositories.TrendData, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.tasksTrend, nil
}

func (m *MockDashboardRepository) GetEventsTrend(_ context.Context, _, _ time.Time) ([]repositories.TrendData, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.eventsTrend, nil
}

func (m *MockDashboardRepository) GetRecentActivity(_ context.Context, _ int) ([]repositories.ActivityData, int64, error) {
	if m.err != nil {
		return nil, 0, m.err
	}
	return m.activities, m.activityTotal, nil
}

func TestDashboardUseCase_GetStats(t *testing.T) {
	tests := []struct {
		name     string
		period   string
		mock     *MockDashboardRepository
		wantErr  bool
		validate func(t *testing.T, result *dto.DashboardStatsOutput)
	}{
		{
			name:   "successful stats retrieval",
			period: "month",
			mock: &MockDashboardRepository{
				documentsCount: &repositories.CountResult{Total: 100, PreviousTotal: 80},
				reportsCount:   &repositories.CountResult{Total: 50, PreviousTotal: 40},
				tasksCount:     &repositories.CountResult{Total: 200, PreviousTotal: 150},
				eventsCount:    &repositories.CountResult{Total: 30, PreviousTotal: 25},
				studentsCount:  &repositories.CountResult{Total: 500, PreviousTotal: 450},
			},
			wantErr: false,
			validate: func(t *testing.T, result *dto.DashboardStatsOutput) {
				if result.Documents.Total != 100 {
					t.Errorf("expected documents total 100, got %d", result.Documents.Total)
				}
				if result.Documents.Change != 25.0 { // (100-80)/80 * 100 = 25%
					t.Errorf("expected documents change 25.0, got %f", result.Documents.Change)
				}
				if result.Reports.Total != 50 {
					t.Errorf("expected reports total 50, got %d", result.Reports.Total)
				}
			},
		},
		{
			name:   "error from repository",
			period: "week",
			mock: &MockDashboardRepository{
				err: errors.New("database error"),
			},
			wantErr: true,
		},
		{
			name:   "zero previous total",
			period: "month",
			mock: &MockDashboardRepository{
				documentsCount: &repositories.CountResult{Total: 10, PreviousTotal: 0},
				reportsCount:   &repositories.CountResult{Total: 5, PreviousTotal: 0},
				tasksCount:     &repositories.CountResult{Total: 20, PreviousTotal: 0},
				eventsCount:    &repositories.CountResult{Total: 3, PreviousTotal: 0},
				studentsCount:  &repositories.CountResult{Total: 50, PreviousTotal: 0},
			},
			wantErr: false,
			validate: func(t *testing.T, result *dto.DashboardStatsOutput) {
				if result.Documents.Change != 100.0 {
					t.Errorf("expected documents change 100.0 for zero previous, got %f", result.Documents.Change)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewDashboardUseCase(tt.mock)
			result, err := uc.GetStats(context.Background(), tt.period)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetStats() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestDashboardUseCase_GetTrends(t *testing.T) {
	tests := []struct {
		name     string
		input    *dto.DashboardTrendsInput
		mock     *MockDashboardRepository
		wantErr  bool
		validate func(t *testing.T, result *dto.DashboardTrendsOutput)
	}{
		{
			name: "successful trends retrieval",
			input: &dto.DashboardTrendsInput{
				Period: "month",
			},
			mock: &MockDashboardRepository{
				documentsTrend: []repositories.TrendData{
					{Date: time.Now().AddDate(0, 0, -7), Count: 10},
					{Date: time.Now().AddDate(0, 0, -6), Count: 15},
				},
				reportsTrend: []repositories.TrendData{
					{Date: time.Now().AddDate(0, 0, -7), Count: 5},
				},
				tasksTrend:  []repositories.TrendData{},
				eventsTrend: []repositories.TrendData{},
			},
			wantErr: false,
			validate: func(t *testing.T, result *dto.DashboardTrendsOutput) {
				if len(result.DocumentsTrend) != 2 {
					t.Errorf("expected 2 document trend points, got %d", len(result.DocumentsTrend))
				}
				if len(result.ReportsTrend) != 1 {
					t.Errorf("expected 1 report trend point, got %d", len(result.ReportsTrend))
				}
			},
		},
		{
			name: "error from repository",
			input: &dto.DashboardTrendsInput{
				Period: "week",
			},
			mock: &MockDashboardRepository{
				err: errors.New("database error"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewDashboardUseCase(tt.mock)
			result, err := uc.GetTrends(context.Background(), tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetTrends() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestDashboardUseCase_GetActivity(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		limit    int
		mock     *MockDashboardRepository
		wantErr  bool
		validate func(t *testing.T, result *dto.DashboardActivityOutput)
	}{
		{
			name:  "successful activity retrieval",
			limit: 10,
			mock: &MockDashboardRepository{
				activities: []repositories.ActivityData{
					{
						ID:          1,
						Type:        "document",
						Action:      "created",
						Title:       "Test Document",
						Description: "Test description",
						UserID:      1,
						UserName:    "Test User",
						CreatedAt:   now,
					},
				},
				activityTotal: 100,
			},
			wantErr: false,
			validate: func(t *testing.T, result *dto.DashboardActivityOutput) {
				if len(result.Activities) != 1 {
					t.Errorf("expected 1 activity, got %d", len(result.Activities))
				}
				if result.Total != 100 {
					t.Errorf("expected total 100, got %d", result.Total)
				}
				if result.Activities[0].Title != "Test Document" {
					t.Errorf("expected title 'Test Document', got '%s'", result.Activities[0].Title)
				}
			},
		},
		{
			name:  "limit enforced to max 50",
			limit: 100,
			mock: &MockDashboardRepository{
				activities:    []repositories.ActivityData{},
				activityTotal: 0,
			},
			wantErr: false,
		},
		{
			name:  "default limit when zero",
			limit: 0,
			mock: &MockDashboardRepository{
				activities:    []repositories.ActivityData{},
				activityTotal: 0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := NewDashboardUseCase(tt.mock)
			result, err := uc.GetActivity(context.Background(), tt.limit)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetActivity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestCalculateChange(t *testing.T) {
	tests := []struct {
		name     string
		current  int64
		previous int64
		expected float64
	}{
		{"positive change", 120, 100, 20.0},
		{"negative change", 80, 100, -20.0},
		{"zero previous with current", 10, 0, 100.0},
		{"zero previous zero current", 0, 0, 0.0},
		{"no change", 100, 100, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateChange(tt.current, tt.previous)
			if result != tt.expected {
				t.Errorf("calculateChange(%d, %d) = %f, expected %f", tt.current, tt.previous, result, tt.expected)
			}
		})
	}
}

func TestGetPeriodDays(t *testing.T) {
	tests := []struct {
		period   string
		expected int
	}{
		{"week", 7},
		{"month", 30},
		{"quarter", 90},
		{"year", 365},
		{"unknown", 30},
		{"", 30},
	}

	for _, tt := range tests {
		t.Run(tt.period, func(t *testing.T) {
			result := getPeriodDays(tt.period)
			if result != tt.expected {
				t.Errorf("getPeriodDays(%s) = %d, expected %d", tt.period, result, tt.expected)
			}
		})
	}
}

func TestTruncateDescription(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{"short string", "Hello", 10, "Hello"},
		{"exact length", "Hello", 5, "Hello"},
		{"needs truncation", "Hello World", 8, "Hello..."},
		{"empty string", "", 10, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateDescription(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncateDescription(%s, %d) = %s, expected %s", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}
