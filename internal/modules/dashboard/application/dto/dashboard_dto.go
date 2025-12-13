// Package dto contains Data Transfer Objects for the dashboard module.
package dto

import "time"

// StatItem represents a single KPI statistic with trend information
type StatItem struct {
	Total  int64   `json:"total"`
	Change float64 `json:"change"` // percentage change from previous period
	Period string  `json:"period"` // "week" or "month"
}

// DashboardStatsOutput represents the KPI statistics for the dashboard
type DashboardStatsOutput struct {
	Documents StatItem `json:"documents"`
	Students  StatItem `json:"students"`
	Events    StatItem `json:"events"`
	Reports   StatItem `json:"reports"`
	Tasks     StatItem `json:"tasks"`
}

// TrendPoint represents a single data point for charts
type TrendPoint struct {
	Date  string `json:"date"`
	Value int64  `json:"value"`
}

// DashboardTrendsInput represents filter options for trends data
type DashboardTrendsInput struct {
	Period    string `form:"period" validate:"omitempty,oneof=week month quarter year"`
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
}

// DashboardTrendsOutput represents trend data for charts
type DashboardTrendsOutput struct {
	DocumentsTrend []TrendPoint `json:"documents_trend"`
	ReportsTrend   []TrendPoint `json:"reports_trend"`
	TasksTrend     []TrendPoint `json:"tasks_trend"`
	EventsTrend    []TrendPoint `json:"events_trend"`
}

// ActivityItem represents a single activity entry
type ActivityItem struct {
	ID          int64     `json:"id"`
	Type        string    `json:"type"`   // "document", "report", "task", "event", "announcement"
	Action      string    `json:"action"` // "created", "updated", "deleted", "completed", etc.
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	UserID      int64     `json:"user_id"`
	UserName    string    `json:"user_name"`
	CreatedAt   time.Time `json:"created_at"`
}

// DashboardActivityOutput represents recent activity for the dashboard
type DashboardActivityOutput struct {
	Activities []ActivityItem `json:"activities"`
	Total      int64          `json:"total"`
}

// ExportDashboardInput represents input for exporting dashboard data
type ExportDashboardInput struct {
	Format    string   `json:"format" validate:"required,oneof=pdf xlsx"`
	StartDate string   `json:"start_date,omitempty"`
	EndDate   string   `json:"end_date,omitempty"`
	Sections  []string `json:"sections,omitempty"` // "stats", "trends", "activity"
}

// ExportDashboardOutput represents output after export
type ExportDashboardOutput struct {
	FileURL   string `json:"file_url"`
	FileName  string `json:"file_name"`
	FileSize  int64  `json:"file_size"`
	ExpiresAt string `json:"expires_at"`
}
