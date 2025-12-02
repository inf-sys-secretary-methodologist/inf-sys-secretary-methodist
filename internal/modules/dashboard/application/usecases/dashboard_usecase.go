// Package usecases contains business logic for the dashboard module.
package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/dashboard/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/dashboard/domain/repositories"
)

// DashboardUseCase handles dashboard business logic
type DashboardUseCase struct {
	repo repositories.DashboardRepository
}

// NewDashboardUseCase creates a new DashboardUseCase
func NewDashboardUseCase(repo repositories.DashboardRepository) *DashboardUseCase {
	return &DashboardUseCase{repo: repo}
}

// GetStats returns KPI statistics for the dashboard
func (uc *DashboardUseCase) GetStats(ctx context.Context, period string) (*dto.DashboardStatsOutput, error) {
	periodDays := getPeriodDays(period)

	// Fetch all counts in parallel would be ideal, but for simplicity we do sequentially
	documents, err := uc.repo.GetDocumentsCount(ctx, periodDays)
	if err != nil {
		return nil, fmt.Errorf("failed to get documents count: %w", err)
	}

	reports, err := uc.repo.GetReportsCount(ctx, periodDays)
	if err != nil {
		return nil, fmt.Errorf("failed to get reports count: %w", err)
	}

	tasks, err := uc.repo.GetTasksCount(ctx, periodDays)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks count: %w", err)
	}

	events, err := uc.repo.GetEventsCount(ctx, periodDays)
	if err != nil {
		return nil, fmt.Errorf("failed to get events count: %w", err)
	}

	students, err := uc.repo.GetStudentsCount(ctx, periodDays)
	if err != nil {
		return nil, fmt.Errorf("failed to get students count: %w", err)
	}

	return &dto.DashboardStatsOutput{
		Documents: dto.StatItem{
			Total:  documents.Total,
			Change: calculateChange(documents.Total, documents.PreviousTotal),
			Period: period,
		},
		Reports: dto.StatItem{
			Total:  reports.Total,
			Change: calculateChange(reports.Total, reports.PreviousTotal),
			Period: period,
		},
		Tasks: dto.StatItem{
			Total:  tasks.Total,
			Change: calculateChange(tasks.Total, tasks.PreviousTotal),
			Period: period,
		},
		Events: dto.StatItem{
			Total:  events.Total,
			Change: calculateChange(events.Total, events.PreviousTotal),
			Period: period,
		},
		Students: dto.StatItem{
			Total:  students.Total,
			Change: calculateChange(students.Total, students.PreviousTotal),
			Period: period,
		},
	}, nil
}

// GetTrends returns trend data for charts
func (uc *DashboardUseCase) GetTrends(ctx context.Context, input *dto.DashboardTrendsInput) (*dto.DashboardTrendsOutput, error) {
	startDate, endDate := getTrendDateRange(input)

	documentsTrend, err := uc.repo.GetDocumentsTrend(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get documents trend: %w", err)
	}

	reportsTrend, err := uc.repo.GetReportsTrend(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get reports trend: %w", err)
	}

	tasksTrend, err := uc.repo.GetTasksTrend(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks trend: %w", err)
	}

	eventsTrend, err := uc.repo.GetEventsTrend(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get events trend: %w", err)
	}

	return &dto.DashboardTrendsOutput{
		DocumentsTrend: convertTrendData(documentsTrend),
		ReportsTrend:   convertTrendData(reportsTrend),
		TasksTrend:     convertTrendData(tasksTrend),
		EventsTrend:    convertTrendData(eventsTrend),
	}, nil
}

// GetActivity returns recent activity
func (uc *DashboardUseCase) GetActivity(ctx context.Context, limit int) (*dto.DashboardActivityOutput, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	activities, total, err := uc.repo.GetRecentActivity(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent activity: %w", err)
	}

	items := make([]dto.ActivityItem, len(activities))
	for i, a := range activities {
		items[i] = dto.ActivityItem{
			ID:          a.ID,
			Type:        a.Type,
			Action:      a.Action,
			Title:       a.Title,
			Description: truncateDescription(a.Description, 100),
			UserID:      a.UserID,
			UserName:    a.UserName,
			CreatedAt:   a.CreatedAt,
		}
	}

	return &dto.DashboardActivityOutput{
		Activities: items,
		Total:      total,
	}, nil
}

// Helper functions

func getPeriodDays(period string) int {
	switch period {
	case "week":
		return 7
	case "month":
		return 30
	case "quarter":
		return 90
	case "year":
		return 365
	default:
		return 30 // default to month
	}
}

func calculateChange(current, previous int64) float64 {
	if previous == 0 {
		if current > 0 {
			return 100.0
		}
		return 0.0
	}
	return float64(current-previous) / float64(previous) * 100
}

func getTrendDateRange(input *dto.DashboardTrendsInput) (time.Time, time.Time) {
	now := time.Now()
	endDate := now

	if input.EndDate != "" {
		if parsed, err := time.Parse("2006-01-02", input.EndDate); err == nil {
			endDate = parsed
		}
	}

	var startDate time.Time
	if input.StartDate != "" {
		if parsed, err := time.Parse("2006-01-02", input.StartDate); err == nil {
			startDate = parsed
		}
	} else {
		switch input.Period {
		case "week":
			startDate = endDate.AddDate(0, 0, -7)
		case "quarter":
			startDate = endDate.AddDate(0, -3, 0)
		case "year":
			startDate = endDate.AddDate(-1, 0, 0)
		default: // month
			startDate = endDate.AddDate(0, -1, 0)
		}
	}

	return startDate, endDate
}

func convertTrendData(data []repositories.TrendData) []dto.TrendPoint {
	result := make([]dto.TrendPoint, len(data))
	for i, d := range data {
		result[i] = dto.TrendPoint{
			Date:  d.Date.Format("2006-01-02"),
			Value: d.Count,
		}
	}
	return result
}

func truncateDescription(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
