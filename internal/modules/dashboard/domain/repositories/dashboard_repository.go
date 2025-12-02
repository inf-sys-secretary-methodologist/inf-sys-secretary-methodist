// Package repositories contains repository interfaces for the dashboard module.
package repositories

import (
	"context"
	"time"
)

// CountResult represents a count with optional change calculation
type CountResult struct {
	Total         int64
	PreviousTotal int64
}

// TrendData represents trend data for a specific date
type TrendData struct {
	Date  time.Time
	Count int64
}

// ActivityData represents raw activity data from the database
type ActivityData struct {
	ID          int64
	Type        string
	Action      string
	Title       string
	Description string
	UserID      int64
	UserName    string
	CreatedAt   time.Time
}

// DashboardRepository defines the interface for dashboard data operations
type DashboardRepository interface {
	// GetDocumentsCount returns document count with comparison to previous period
	GetDocumentsCount(ctx context.Context, periodDays int) (*CountResult, error)

	// GetReportsCount returns reports count with comparison to previous period
	GetReportsCount(ctx context.Context, periodDays int) (*CountResult, error)

	// GetTasksCount returns tasks count with comparison to previous period
	GetTasksCount(ctx context.Context, periodDays int) (*CountResult, error)

	// GetEventsCount returns events count with comparison to previous period
	GetEventsCount(ctx context.Context, periodDays int) (*CountResult, error)

	// GetStudentsCount returns students count (users with student role)
	GetStudentsCount(ctx context.Context, periodDays int) (*CountResult, error)

	// GetDocumentsTrend returns document creation trend for the period
	GetDocumentsTrend(ctx context.Context, startDate, endDate time.Time) ([]TrendData, error)

	// GetReportsTrend returns reports creation trend for the period
	GetReportsTrend(ctx context.Context, startDate, endDate time.Time) ([]TrendData, error)

	// GetTasksTrend returns tasks creation trend for the period
	GetTasksTrend(ctx context.Context, startDate, endDate time.Time) ([]TrendData, error)

	// GetEventsTrend returns events creation trend for the period
	GetEventsTrend(ctx context.Context, startDate, endDate time.Time) ([]TrendData, error)

	// GetRecentActivity returns recent activity across all modules
	GetRecentActivity(ctx context.Context, limit int) ([]ActivityData, int64, error)
}
