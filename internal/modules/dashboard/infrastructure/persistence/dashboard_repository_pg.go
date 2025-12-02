// Package persistence contains repository implementations for the dashboard module.
package persistence

import (
	"context"
	"database/sql"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/dashboard/domain/repositories"
)

// DashboardRepositoryPG implements DashboardRepository using PostgreSQL
type DashboardRepositoryPG struct {
	db *sql.DB
}

// NewDashboardRepositoryPG creates a new PostgreSQL dashboard repository
func NewDashboardRepositoryPG(db *sql.DB) *DashboardRepositoryPG {
	return &DashboardRepositoryPG{db: db}
}

// GetDocumentsCount returns document count with comparison to previous period
func (r *DashboardRepositoryPG) GetDocumentsCount(ctx context.Context, periodDays int) (*repositories.CountResult, error) {
	now := time.Now()
	periodStart := now.AddDate(0, 0, -periodDays)
	previousStart := periodStart.AddDate(0, 0, -periodDays)

	var total, previousTotal int64

	// Current period count
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM documents WHERE created_at >= $1
	`, periodStart).Scan(&total)
	if err != nil {
		return nil, err
	}

	// Previous period count
	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM documents WHERE created_at >= $1 AND created_at < $2
	`, previousStart, periodStart).Scan(&previousTotal)
	if err != nil {
		return nil, err
	}

	return &repositories.CountResult{
		Total:         total,
		PreviousTotal: previousTotal,
	}, nil
}

// GetReportsCount returns reports count with comparison to previous period
func (r *DashboardRepositoryPG) GetReportsCount(ctx context.Context, periodDays int) (*repositories.CountResult, error) {
	now := time.Now()
	periodStart := now.AddDate(0, 0, -periodDays)
	previousStart := periodStart.AddDate(0, 0, -periodDays)

	var total, previousTotal int64

	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM reports WHERE created_at >= $1
	`, periodStart).Scan(&total)
	if err != nil {
		return nil, err
	}

	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM reports WHERE created_at >= $1 AND created_at < $2
	`, previousStart, periodStart).Scan(&previousTotal)
	if err != nil {
		return nil, err
	}

	return &repositories.CountResult{
		Total:         total,
		PreviousTotal: previousTotal,
	}, nil
}

// GetTasksCount returns tasks count with comparison to previous period
func (r *DashboardRepositoryPG) GetTasksCount(ctx context.Context, periodDays int) (*repositories.CountResult, error) {
	now := time.Now()
	periodStart := now.AddDate(0, 0, -periodDays)
	previousStart := periodStart.AddDate(0, 0, -periodDays)

	var total, previousTotal int64

	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM tasks WHERE created_at >= $1
	`, periodStart).Scan(&total)
	if err != nil {
		return nil, err
	}

	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM tasks WHERE created_at >= $1 AND created_at < $2
	`, previousStart, periodStart).Scan(&previousTotal)
	if err != nil {
		return nil, err
	}

	return &repositories.CountResult{
		Total:         total,
		PreviousTotal: previousTotal,
	}, nil
}

// GetEventsCount returns events count with comparison to previous period
func (r *DashboardRepositoryPG) GetEventsCount(ctx context.Context, periodDays int) (*repositories.CountResult, error) {
	now := time.Now()
	periodStart := now.AddDate(0, 0, -periodDays)
	previousStart := periodStart.AddDate(0, 0, -periodDays)

	var total, previousTotal int64

	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM events WHERE created_at >= $1
	`, periodStart).Scan(&total)
	if err != nil {
		return nil, err
	}

	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM events WHERE created_at >= $1 AND created_at < $2
	`, previousStart, periodStart).Scan(&previousTotal)
	if err != nil {
		return nil, err
	}

	return &repositories.CountResult{
		Total:         total,
		PreviousTotal: previousTotal,
	}, nil
}

// GetStudentsCount returns students count (users with student role)
func (r *DashboardRepositoryPG) GetStudentsCount(ctx context.Context, periodDays int) (*repositories.CountResult, error) {
	now := time.Now()
	periodStart := now.AddDate(0, 0, -periodDays)
	previousStart := periodStart.AddDate(0, 0, -periodDays)

	var total, previousTotal int64

	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM users WHERE role = 'student' AND created_at >= $1
	`, periodStart).Scan(&total)
	if err != nil {
		return nil, err
	}

	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM users WHERE role = 'student' AND created_at >= $1 AND created_at < $2
	`, previousStart, periodStart).Scan(&previousTotal)
	if err != nil {
		return nil, err
	}

	return &repositories.CountResult{
		Total:         total,
		PreviousTotal: previousTotal,
	}, nil
}

// GetDocumentsTrend returns document creation trend for the period
func (r *DashboardRepositoryPG) GetDocumentsTrend(ctx context.Context, startDate, endDate time.Time) ([]repositories.TrendData, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT DATE(created_at) as date, COUNT(*) as count
		FROM documents
		WHERE created_at >= $1 AND created_at <= $2
		GROUP BY DATE(created_at)
		ORDER BY date
	`, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return scanTrendData(rows)
}

// GetReportsTrend returns reports creation trend for the period
func (r *DashboardRepositoryPG) GetReportsTrend(ctx context.Context, startDate, endDate time.Time) ([]repositories.TrendData, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT DATE(created_at) as date, COUNT(*) as count
		FROM reports
		WHERE created_at >= $1 AND created_at <= $2
		GROUP BY DATE(created_at)
		ORDER BY date
	`, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return scanTrendData(rows)
}

// GetTasksTrend returns tasks creation trend for the period
func (r *DashboardRepositoryPG) GetTasksTrend(ctx context.Context, startDate, endDate time.Time) ([]repositories.TrendData, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT DATE(created_at) as date, COUNT(*) as count
		FROM tasks
		WHERE created_at >= $1 AND created_at <= $2
		GROUP BY DATE(created_at)
		ORDER BY date
	`, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return scanTrendData(rows)
}

// GetEventsTrend returns events creation trend for the period
func (r *DashboardRepositoryPG) GetEventsTrend(ctx context.Context, startDate, endDate time.Time) ([]repositories.TrendData, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT DATE(created_at) as date, COUNT(*) as count
		FROM events
		WHERE created_at >= $1 AND created_at <= $2
		GROUP BY DATE(created_at)
		ORDER BY date
	`, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return scanTrendData(rows)
}

// GetRecentActivity returns recent activity across all modules
func (r *DashboardRepositoryPG) GetRecentActivity(ctx context.Context, limit int) ([]repositories.ActivityData, int64, error) {
	// Union query for activity from different modules
	query := `
		WITH activity AS (
			SELECT
				d.id,
				'document' as type,
				'created' as action,
				d.title,
				COALESCE(d.subject, '') as description,
				d.author_id as user_id,
				u.name as user_name,
				d.created_at
			FROM documents d
			LEFT JOIN users u ON d.author_id = u.id

			UNION ALL

			SELECT
				r.id,
				'report' as type,
				'created' as action,
				r.title,
				COALESCE(r.description, '') as description,
				r.author_id as user_id,
				u.name as user_name,
				r.created_at
			FROM reports r
			LEFT JOIN users u ON r.author_id = u.id

			UNION ALL

			SELECT
				t.id,
				'task' as type,
				'created' as action,
				t.title,
				COALESCE(t.description, '') as description,
				t.author_id as user_id,
				u.name as user_name,
				t.created_at
			FROM tasks t
			LEFT JOIN users u ON t.author_id = u.id

			UNION ALL

			SELECT
				e.id,
				'event' as type,
				'created' as action,
				e.title,
				COALESCE(e.description, '') as description,
				e.organizer_id as user_id,
				u.name as user_name,
				e.created_at
			FROM events e
			LEFT JOIN users u ON e.organizer_id = u.id

			UNION ALL

			SELECT
				a.id,
				'announcement' as type,
				'created' as action,
				a.title,
				COALESCE(a.content, '') as description,
				a.author_id as user_id,
				u.name as user_name,
				a.created_at
			FROM announcements a
			LEFT JOIN users u ON a.author_id = u.id
		)
		SELECT id, type, action, title, description, user_id, COALESCE(user_name, 'Unknown') as user_name, created_at
		FROM activity
		ORDER BY created_at DESC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	var activities []repositories.ActivityData
	for rows.Next() {
		var a repositories.ActivityData
		err := rows.Scan(&a.ID, &a.Type, &a.Action, &a.Title, &a.Description, &a.UserID, &a.UserName, &a.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		activities = append(activities, a)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	// Get total count
	var total int64
	err = r.db.QueryRowContext(ctx, `
		SELECT
			(SELECT COUNT(*) FROM documents) +
			(SELECT COUNT(*) FROM reports) +
			(SELECT COUNT(*) FROM tasks) +
			(SELECT COUNT(*) FROM events) +
			(SELECT COUNT(*) FROM announcements)
	`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return activities, total, nil
}

// scanTrendData scans rows into TrendData slice
func scanTrendData(rows *sql.Rows) ([]repositories.TrendData, error) {
	var result []repositories.TrendData
	for rows.Next() {
		var t repositories.TrendData
		if err := rows.Scan(&t.Date, &t.Count); err != nil {
			return nil, err
		}
		result = append(result, t)
	}
	return result, rows.Err()
}
