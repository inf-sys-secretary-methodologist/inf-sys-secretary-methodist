package usecases

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/dashboard/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/dashboard/domain/repositories"
)

const (
	testPeriodWeek    = "week"
	testPeriodQuarter = "quarter"
	testPeriodYear    = "year"
)

// MockDashboardRepository is a mock implementation of DashboardRepository
// with per-method error support for granular testing.
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
	err            error // global error for all methods

	// Per-method errors (override global err when set)
	documentsCountErr *error
	reportsCountErr   *error
	tasksCountErr     *error
	eventsCountErr    *error
	studentsCountErr  *error
	documentsTrendErr *error
	reportsTrendErr   *error
	tasksTrendErr     *error
	eventsTrendErr    *error
	activityErr       *error
}

func (m *MockDashboardRepository) getErr(specific *error) error {
	if specific != nil {
		return *specific
	}
	return m.err
}

func (m *MockDashboardRepository) GetDocumentsCount(_ context.Context, _ int) (*repositories.CountResult, error) {
	if err := m.getErr(m.documentsCountErr); err != nil {
		return nil, err
	}
	return m.documentsCount, nil
}

func (m *MockDashboardRepository) GetReportsCount(_ context.Context, _ int) (*repositories.CountResult, error) {
	if err := m.getErr(m.reportsCountErr); err != nil {
		return nil, err
	}
	return m.reportsCount, nil
}

func (m *MockDashboardRepository) GetTasksCount(_ context.Context, _ int) (*repositories.CountResult, error) {
	if err := m.getErr(m.tasksCountErr); err != nil {
		return nil, err
	}
	return m.tasksCount, nil
}

func (m *MockDashboardRepository) GetEventsCount(_ context.Context, _ int) (*repositories.CountResult, error) {
	if err := m.getErr(m.eventsCountErr); err != nil {
		return nil, err
	}
	return m.eventsCount, nil
}

func (m *MockDashboardRepository) GetStudentsCount(_ context.Context, _ int) (*repositories.CountResult, error) {
	if err := m.getErr(m.studentsCountErr); err != nil {
		return nil, err
	}
	return m.studentsCount, nil
}

func (m *MockDashboardRepository) GetDocumentsTrend(_ context.Context, _, _ time.Time) ([]repositories.TrendData, error) {
	if err := m.getErr(m.documentsTrendErr); err != nil {
		return nil, err
	}
	return m.documentsTrend, nil
}

func (m *MockDashboardRepository) GetReportsTrend(_ context.Context, _, _ time.Time) ([]repositories.TrendData, error) {
	if err := m.getErr(m.reportsTrendErr); err != nil {
		return nil, err
	}
	return m.reportsTrend, nil
}

func (m *MockDashboardRepository) GetTasksTrend(_ context.Context, _, _ time.Time) ([]repositories.TrendData, error) {
	if err := m.getErr(m.tasksTrendErr); err != nil {
		return nil, err
	}
	return m.tasksTrend, nil
}

func (m *MockDashboardRepository) GetEventsTrend(_ context.Context, _, _ time.Time) ([]repositories.TrendData, error) {
	if err := m.getErr(m.eventsTrendErr); err != nil {
		return nil, err
	}
	return m.eventsTrend, nil
}

func (m *MockDashboardRepository) GetRecentActivity(_ context.Context, _ int) ([]repositories.ActivityData, int64, error) {
	if err := m.getErr(m.activityErr); err != nil {
		return nil, 0, err
	}
	return m.activities, m.activityTotal, nil
}

// Helper to create a pointer to an error
func errPtr(err error) *error {
	return &err
}

// successMock returns a mock with all counts and trends populated (no errors).
func successMock() *MockDashboardRepository {
	return &MockDashboardRepository{
		documentsCount: &repositories.CountResult{Total: 100, PreviousTotal: 80},
		reportsCount:   &repositories.CountResult{Total: 50, PreviousTotal: 40},
		tasksCount:     &repositories.CountResult{Total: 200, PreviousTotal: 150},
		eventsCount:    &repositories.CountResult{Total: 30, PreviousTotal: 25},
		studentsCount:  &repositories.CountResult{Total: 500, PreviousTotal: 450},
		documentsTrend: []repositories.TrendData{{Date: time.Now(), Count: 10}},
		reportsTrend:   []repositories.TrendData{{Date: time.Now(), Count: 5}},
		tasksTrend:     []repositories.TrendData{{Date: time.Now(), Count: 8}},
		eventsTrend:    []repositories.TrendData{{Date: time.Now(), Count: 3}},
		activities:     []repositories.ActivityData{},
		activityTotal:  0,
	}
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
			mock:   successMock(),
			validate: func(t *testing.T, result *dto.DashboardStatsOutput) {
				if result.Documents.Total != 100 {
					t.Errorf("expected documents total 100, got %d", result.Documents.Total)
				}
				if result.Documents.Change != 25.0 {
					t.Errorf("expected documents change 25.0, got %f", result.Documents.Change)
				}
				if result.Reports.Total != 50 {
					t.Errorf("expected reports total 50, got %d", result.Reports.Total)
				}
				if result.Documents.Period != "month" {
					t.Errorf("expected period 'month', got '%s'", result.Documents.Period)
				}
			},
		},
		{
			name:    "documents count error",
			period:  "week",
			mock:    &MockDashboardRepository{err: errors.New("database error")},
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
			validate: func(t *testing.T, result *dto.DashboardStatsOutput) {
				if result.Documents.Change != 100.0 {
					t.Errorf("expected documents change 100.0 for zero previous, got %f", result.Documents.Change)
				}
			},
		},
		{
			name:   "reports count error",
			period: "month",
			mock: func() *MockDashboardRepository {
				m := successMock()
				m.reportsCountErr = errPtr(errors.New("reports error"))
				return m
			}(),
			wantErr: true,
		},
		{
			name:   "tasks count error",
			period: "month",
			mock: func() *MockDashboardRepository {
				m := successMock()
				m.tasksCountErr = errPtr(errors.New("tasks error"))
				return m
			}(),
			wantErr: true,
		},
		{
			name:   "events count error",
			period: "month",
			mock: func() *MockDashboardRepository {
				m := successMock()
				m.eventsCountErr = errPtr(errors.New("events error"))
				return m
			}(),
			wantErr: true,
		},
		{
			name:   "students count error",
			period: "month",
			mock: func() *MockDashboardRepository {
				m := successMock()
				m.studentsCountErr = errPtr(errors.New("students error"))
				return m
			}(),
			wantErr: true,
		},
		{
			name:   "week period",
			period: testPeriodWeek,
			mock:   successMock(),
			validate: func(t *testing.T, result *dto.DashboardStatsOutput) {
				if result.Documents.Period != testPeriodWeek {
					t.Errorf("expected period '%s', got '%s'", testPeriodWeek, result.Documents.Period)
				}
			},
		},
		{
			name:   "quarter period",
			period: testPeriodQuarter,
			mock:   successMock(),
			validate: func(t *testing.T, result *dto.DashboardStatsOutput) {
				if result.Documents.Period != testPeriodQuarter {
					t.Errorf("expected period '%s', got '%s'", testPeriodQuarter, result.Documents.Period)
				}
			},
		},
		{
			name:   "year period",
			period: testPeriodYear,
			mock:   successMock(),
			validate: func(t *testing.T, result *dto.DashboardStatsOutput) {
				if result.Documents.Period != testPeriodYear {
					t.Errorf("expected period '%s', got '%s'", testPeriodYear, result.Documents.Period)
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
			mock: successMock(),
			validate: func(t *testing.T, result *dto.DashboardTrendsOutput) {
				if len(result.DocumentsTrend) != 1 {
					t.Errorf("expected 1 document trend point, got %d", len(result.DocumentsTrend))
				}
				if len(result.ReportsTrend) != 1 {
					t.Errorf("expected 1 report trend point, got %d", len(result.ReportsTrend))
				}
			},
		},
		{
			name: "documents trend error",
			input: &dto.DashboardTrendsInput{
				Period: "week",
			},
			mock:    &MockDashboardRepository{err: errors.New("database error")},
			wantErr: true,
		},
		{
			name: "reports trend error",
			input: &dto.DashboardTrendsInput{
				Period: "month",
			},
			mock: func() *MockDashboardRepository {
				m := successMock()
				m.reportsTrendErr = errPtr(errors.New("reports trend error"))
				return m
			}(),
			wantErr: true,
		},
		{
			name: "tasks trend error",
			input: &dto.DashboardTrendsInput{
				Period: "month",
			},
			mock: func() *MockDashboardRepository {
				m := successMock()
				m.tasksTrendErr = errPtr(errors.New("tasks trend error"))
				return m
			}(),
			wantErr: true,
		},
		{
			name: "events trend error",
			input: &dto.DashboardTrendsInput{
				Period: "month",
			},
			mock: func() *MockDashboardRepository {
				m := successMock()
				m.eventsTrendErr = errPtr(errors.New("events trend error"))
				return m
			}(),
			wantErr: true,
		},
		{
			name: "week period",
			input: &dto.DashboardTrendsInput{
				Period: "week",
			},
			mock: successMock(),
		},
		{
			name: "quarter period",
			input: &dto.DashboardTrendsInput{
				Period: "quarter",
			},
			mock: successMock(),
		},
		{
			name: "year period",
			input: &dto.DashboardTrendsInput{
				Period: "year",
			},
			mock: successMock(),
		},
		{
			name: "custom start and end dates",
			input: &dto.DashboardTrendsInput{
				StartDate: "2024-01-01",
				EndDate:   "2024-01-31",
			},
			mock: successMock(),
		},
		{
			name: "custom start date only",
			input: &dto.DashboardTrendsInput{
				StartDate: "2024-01-01",
			},
			mock: successMock(),
		},
		{
			name: "custom end date only",
			input: &dto.DashboardTrendsInput{
				EndDate: "2024-06-30",
			},
			mock: successMock(),
		},
		{
			name: "invalid start date falls back to period",
			input: &dto.DashboardTrendsInput{
				StartDate: "not-a-date",
				Period:    "week",
			},
			mock: successMock(),
		},
		{
			name: "invalid end date falls back to now",
			input: &dto.DashboardTrendsInput{
				EndDate: "not-a-date",
				Period:  "month",
			},
			mock: successMock(),
		},
		{
			name: "default period (empty)",
			input: &dto.DashboardTrendsInput{
				Period: "",
			},
			mock: successMock(),
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
		},
		{
			name:  "default limit when zero",
			limit: 0,
			mock: &MockDashboardRepository{
				activities:    []repositories.ActivityData{},
				activityTotal: 0,
			},
		},
		{
			name:  "negative limit defaults to 10",
			limit: -5,
			mock: &MockDashboardRepository{
				activities:    []repositories.ActivityData{},
				activityTotal: 0,
			},
		},
		{
			name:    "activity retrieval error",
			limit:   10,
			mock:    &MockDashboardRepository{err: errors.New("activity error")},
			wantErr: true,
		},
		{
			name:  "long description gets truncated",
			limit: 10,
			mock: &MockDashboardRepository{
				activities: []repositories.ActivityData{
					{
						ID:          1,
						Type:        "document",
						Action:      "created",
						Title:       "Test",
						Description: strings.Repeat("a", 200),
						UserID:      1,
						UserName:    "User",
						CreatedAt:   now,
					},
				},
				activityTotal: 1,
			},
			validate: func(t *testing.T, result *dto.DashboardActivityOutput) {
				if len(result.Activities[0].Description) != 100 {
					t.Errorf("expected truncated description of length 100, got %d", len(result.Activities[0].Description))
				}
				if !strings.HasSuffix(result.Activities[0].Description, "...") {
					t.Error("expected truncated description to end with '...'")
				}
			},
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

func TestGetTrendDateRange(t *testing.T) {
	tests := []struct {
		name  string
		input *dto.DashboardTrendsInput
		check func(t *testing.T, start, end time.Time)
	}{
		{
			name:  "week period",
			input: &dto.DashboardTrendsInput{Period: "week"},
			check: func(t *testing.T, start, end time.Time) {
				diff := end.Sub(start)
				if int(diff.Hours()/24) != 7 {
					t.Errorf("expected 7 days diff for week, got %d", int(diff.Hours()/24))
				}
			},
		},
		{
			name:  "quarter period",
			input: &dto.DashboardTrendsInput{Period: "quarter"},
			check: func(t *testing.T, start, end time.Time) {
				// quarter = 3 months back
				expected := end.AddDate(0, -3, 0)
				if start.Year() != expected.Year() || start.Month() != expected.Month() || start.Day() != expected.Day() {
					t.Errorf("expected start ~%v, got %v", expected, start)
				}
			},
		},
		{
			name:  "year period",
			input: &dto.DashboardTrendsInput{Period: "year"},
			check: func(t *testing.T, start, end time.Time) {
				expected := end.AddDate(-1, 0, 0)
				if start.Year() != expected.Year() || start.Month() != expected.Month() {
					t.Errorf("expected start ~%v, got %v", expected, start)
				}
			},
		},
		{
			name:  "default (month) period",
			input: &dto.DashboardTrendsInput{Period: ""},
			check: func(t *testing.T, start, end time.Time) {
				expected := end.AddDate(0, -1, 0)
				if start.Year() != expected.Year() || start.Month() != expected.Month() {
					t.Errorf("expected start ~%v, got %v", expected, start)
				}
			},
		},
		{
			name: "custom start date",
			input: &dto.DashboardTrendsInput{
				StartDate: "2024-03-01",
				Period:    "month",
			},
			check: func(t *testing.T, start, _ time.Time) {
				if start.Year() != 2024 || start.Month() != 3 || start.Day() != 1 {
					t.Errorf("expected start 2024-03-01, got %v", start)
				}
			},
		},
		{
			name: "custom end date",
			input: &dto.DashboardTrendsInput{
				EndDate: "2024-06-30",
				Period:  "month",
			},
			check: func(t *testing.T, _, end time.Time) {
				if end.Year() != 2024 || end.Month() != 6 || end.Day() != 30 {
					t.Errorf("expected end 2024-06-30, got %v", end)
				}
			},
		},
		{
			name: "both custom dates",
			input: &dto.DashboardTrendsInput{
				StartDate: "2024-01-01",
				EndDate:   "2024-12-31",
			},
			check: func(t *testing.T, start, end time.Time) {
				if start.Year() != 2024 || start.Month() != 1 || start.Day() != 1 {
					t.Errorf("expected start 2024-01-01, got %v", start)
				}
				if end.Year() != 2024 || end.Month() != 12 || end.Day() != 31 {
					t.Errorf("expected end 2024-12-31, got %v", end)
				}
			},
		},
		{
			name: "invalid start date results in zero time",
			input: &dto.DashboardTrendsInput{
				StartDate: "bad-date",
				Period:    "week",
			},
			check: func(t *testing.T, start, _ time.Time) {
				// Invalid start date fails to parse, startDate stays zero
				// (the else branch only triggers when StartDate == "")
				if !start.IsZero() {
					t.Errorf("expected zero start time for invalid date, got %v", start)
				}
			},
		},
		{
			name: "invalid end date uses now",
			input: &dto.DashboardTrendsInput{
				EndDate: "bad-date",
				Period:  "month",
			},
			check: func(t *testing.T, _, end time.Time) {
				// Should use time.Now()
				if time.Since(end) > time.Second {
					t.Errorf("expected end to be close to now, got %v", end)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := getTrendDateRange(tt.input)
			tt.check(t, start, end)
		})
	}
}

func TestConvertTrendData(t *testing.T) {
	now := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	data := []repositories.TrendData{
		{Date: now, Count: 10},
		{Date: now.AddDate(0, 0, 1), Count: 20},
	}

	result := convertTrendData(data)

	if len(result) != 2 {
		t.Fatalf("expected 2 points, got %d", len(result))
	}
	if result[0].Date != "2024-06-15" {
		t.Errorf("expected date '2024-06-15', got '%s'", result[0].Date)
	}
	if result[0].Value != 10 {
		t.Errorf("expected value 10, got %d", result[0].Value)
	}
	if result[1].Date != "2024-06-16" {
		t.Errorf("expected date '2024-06-16', got '%s'", result[1].Date)
	}

	// Empty data
	emptyResult := convertTrendData(nil)
	if len(emptyResult) != 0 {
		t.Errorf("expected 0 points for nil data, got %d", len(emptyResult))
	}
}
