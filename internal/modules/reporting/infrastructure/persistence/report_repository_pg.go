// Package persistence provides PostgreSQL implementations of reporting repositories.
package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain/repositories"
)

// ReportRepositoryPG implements ReportRepository using PostgreSQL
type ReportRepositoryPG struct {
	db *sql.DB
}

// NewReportRepositoryPG creates a new PostgreSQL report repository
func NewReportRepositoryPG(db *sql.DB) *ReportRepositoryPG {
	return &ReportRepositoryPG{db: db}
}

// Ensure ReportRepositoryPG implements ReportRepository
var _ repositories.ReportRepository = (*ReportRepositoryPG)(nil)

// Create inserts a new report into the database
func (r *ReportRepositoryPG) Create(ctx context.Context, report *entities.Report) error {
	query := `
		INSERT INTO reports (
			report_type_id, title, description, period_start, period_end,
			author_id, status, file_name, file_path, file_size, mime_type,
			parameters, data, reviewer_comment, reviewed_by, reviewed_at,
			published_at, is_public, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
		RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		report.ReportTypeID, report.Title, report.Description,
		report.PeriodStart, report.PeriodEnd, report.AuthorID,
		report.Status, report.FileName, report.FilePath, report.FileSize,
		report.MimeType, report.Parameters, report.Data,
		report.ReviewerComment, report.ReviewedBy, report.ReviewedAt,
		report.PublishedAt, report.IsPublic, report.CreatedAt, report.UpdatedAt,
	).Scan(&report.ID)

	if err != nil {
		return fmt.Errorf("failed to create report: %w", err)
	}
	return nil
}

// Save updates an existing report in the database
func (r *ReportRepositoryPG) Save(ctx context.Context, report *entities.Report) error {
	query := `
		UPDATE reports SET
			report_type_id = $1, title = $2, description = $3,
			period_start = $4, period_end = $5, status = $6,
			file_name = $7, file_path = $8, file_size = $9, mime_type = $10,
			parameters = $11, data = $12, reviewer_comment = $13,
			reviewed_by = $14, reviewed_at = $15, published_at = $16,
			is_public = $17, updated_at = $18
		WHERE id = $19`

	result, err := r.db.ExecContext(ctx, query,
		report.ReportTypeID, report.Title, report.Description,
		report.PeriodStart, report.PeriodEnd, report.Status,
		report.FileName, report.FilePath, report.FileSize, report.MimeType,
		report.Parameters, report.Data, report.ReviewerComment,
		report.ReviewedBy, report.ReviewedAt, report.PublishedAt,
		report.IsPublic, report.UpdatedAt, report.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to save report: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("report not found: %d", report.ID)
	}
	return nil
}

// GetByID retrieves a report by its ID
func (r *ReportRepositoryPG) GetByID(ctx context.Context, id int64) (*entities.Report, error) {
	query := `
		SELECT id, report_type_id, title, description, period_start, period_end,
			author_id, status, file_name, file_path, file_size, mime_type,
			parameters, data, reviewer_comment, reviewed_by, reviewed_at,
			published_at, is_public, created_at, updated_at
		FROM reports WHERE id = $1`

	report := &entities.Report{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&report.ID, &report.ReportTypeID, &report.Title, &report.Description,
		&report.PeriodStart, &report.PeriodEnd, &report.AuthorID, &report.Status,
		&report.FileName, &report.FilePath, &report.FileSize, &report.MimeType,
		&report.Parameters, &report.Data, &report.ReviewerComment,
		&report.ReviewedBy, &report.ReviewedAt, &report.PublishedAt,
		&report.IsPublic, &report.CreatedAt, &report.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get report: %w", err)
	}
	return report, nil
}

// Delete removes a report from the database
func (r *ReportRepositoryPG) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM reports WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete report: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("report not found: %d", id)
	}
	return nil
}

// List retrieves reports with filtering and pagination
func (r *ReportRepositoryPG) List(ctx context.Context, filter repositories.ReportFilter, limit, offset int) ([]*entities.Report, error) {
	query, args := r.buildListQuery(filter, limit, offset, false)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list reports: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return r.scanReports(rows)
}

// Count returns the total count of reports matching the filter
func (r *ReportRepositoryPG) Count(ctx context.Context, filter repositories.ReportFilter) (int64, error) {
	query, args := r.buildListQuery(filter, 0, 0, true)

	var count int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count reports: %w", err)
	}
	return count, nil
}

// GetByAuthor retrieves reports by author ID
func (r *ReportRepositoryPG) GetByAuthor(ctx context.Context, authorID int64, limit, offset int) ([]*entities.Report, error) {
	filter := repositories.ReportFilter{AuthorID: &authorID}
	return r.List(ctx, filter, limit, offset)
}

// GetByStatus retrieves reports by status
func (r *ReportRepositoryPG) GetByStatus(ctx context.Context, status domain.ReportStatus, limit, offset int) ([]*entities.Report, error) {
	filter := repositories.ReportFilter{Status: &status}
	return r.List(ctx, filter, limit, offset)
}

// GetByReportType retrieves reports by report type ID
func (r *ReportRepositoryPG) GetByReportType(ctx context.Context, reportTypeID int64, limit, offset int) ([]*entities.Report, error) {
	filter := repositories.ReportFilter{ReportTypeID: &reportTypeID}
	return r.List(ctx, filter, limit, offset)
}

// GetPublicReports retrieves public reports
func (r *ReportRepositoryPG) GetPublicReports(ctx context.Context, limit, offset int) ([]*entities.Report, error) {
	isPublic := true
	filter := repositories.ReportFilter{IsPublic: &isPublic}
	return r.List(ctx, filter, limit, offset)
}

// buildListQuery constructs the SQL query for listing reports
func (r *ReportRepositoryPG) buildListQuery(filter repositories.ReportFilter, limit, offset int, countOnly bool) (string, []any) {
	var conditions []string
	var args []any
	argNum := 1

	if filter.ReportTypeID != nil {
		conditions = append(conditions, fmt.Sprintf("report_type_id = $%d", argNum))
		args = append(args, *filter.ReportTypeID)
		argNum++
	}
	if filter.AuthorID != nil {
		conditions = append(conditions, fmt.Sprintf("author_id = $%d", argNum))
		args = append(args, *filter.AuthorID)
		argNum++
	}
	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argNum))
		args = append(args, *filter.Status)
		argNum++
	}
	if filter.IsPublic != nil {
		conditions = append(conditions, fmt.Sprintf("is_public = $%d", argNum))
		args = append(args, *filter.IsPublic)
		argNum++
	}
	if filter.PeriodStart != nil {
		conditions = append(conditions, fmt.Sprintf("period_start >= $%d", argNum))
		args = append(args, *filter.PeriodStart)
		argNum++
	}
	if filter.PeriodEnd != nil {
		conditions = append(conditions, fmt.Sprintf("period_end <= $%d", argNum))
		args = append(args, *filter.PeriodEnd)
		argNum++
	}
	if filter.Search != nil && *filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(title ILIKE $%d OR description ILIKE $%d)", argNum, argNum))
		args = append(args, "%"+*filter.Search+"%")
		argNum++
	}

	var query string
	if countOnly {
		query = "SELECT COUNT(*) FROM reports"
	} else {
		query = `
			SELECT id, report_type_id, title, description, period_start, period_end,
				author_id, status, file_name, file_path, file_size, mime_type,
				parameters, data, reviewer_comment, reviewed_by, reviewed_at,
				published_at, is_public, created_at, updated_at
			FROM reports`
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	if !countOnly {
		query += " ORDER BY created_at DESC"
		if limit > 0 {
			query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argNum, argNum+1)
			args = append(args, limit, offset)
		}
	}

	return query, args
}

// scanReports scans report rows into entities
func (r *ReportRepositoryPG) scanReports(rows *sql.Rows) ([]*entities.Report, error) {
	var reports []*entities.Report
	for rows.Next() {
		report := &entities.Report{}
		err := rows.Scan(
			&report.ID, &report.ReportTypeID, &report.Title, &report.Description,
			&report.PeriodStart, &report.PeriodEnd, &report.AuthorID, &report.Status,
			&report.FileName, &report.FilePath, &report.FileSize, &report.MimeType,
			&report.Parameters, &report.Data, &report.ReviewerComment,
			&report.ReviewedBy, &report.ReviewedAt, &report.PublishedAt,
			&report.IsPublic, &report.CreatedAt, &report.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan report: %w", err)
		}
		reports = append(reports, report)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating reports: %w", err)
	}
	return reports, nil
}

// AddAccess adds access permission for a report
func (r *ReportRepositoryPG) AddAccess(ctx context.Context, access *entities.ReportAccess) error {
	query := `
		INSERT INTO report_access (report_id, user_id, role, permission, granted_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		access.ReportID, access.UserID, access.Role,
		access.Permission, access.GrantedBy, access.CreatedAt,
	).Scan(&access.ID)

	if err != nil {
		return fmt.Errorf("failed to add report access: %w", err)
	}
	return nil
}

// RemoveAccess removes access permission from a report
func (r *ReportRepositoryPG) RemoveAccess(ctx context.Context, reportID, accessID int64) error {
	query := `DELETE FROM report_access WHERE id = $1 AND report_id = $2`
	result, err := r.db.ExecContext(ctx, query, accessID, reportID)
	if err != nil {
		return fmt.Errorf("failed to remove report access: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("access not found: %d", accessID)
	}
	return nil
}

// GetAccessByReport retrieves all access permissions for a report
func (r *ReportRepositoryPG) GetAccessByReport(ctx context.Context, reportID int64) ([]*entities.ReportAccess, error) {
	query := `
		SELECT id, report_id, user_id, role, permission, granted_by, created_at
		FROM report_access WHERE report_id = $1`

	rows, err := r.db.QueryContext(ctx, query, reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to get report access: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var accesses []*entities.ReportAccess
	for rows.Next() {
		access := &entities.ReportAccess{}
		err := rows.Scan(
			&access.ID, &access.ReportID, &access.UserID, &access.Role,
			&access.Permission, &access.GrantedBy, &access.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan report access: %w", err)
		}
		accesses = append(accesses, access)
	}
	return accesses, rows.Err()
}

// HasAccess checks if a user has specific permission for a report
func (r *ReportRepositoryPG) HasAccess(ctx context.Context, reportID, userID int64, permission domain.ReportPermission) (bool, error) {
	// First check direct user access
	query := `
		SELECT EXISTS(
			SELECT 1 FROM report_access
			WHERE report_id = $1 AND user_id = $2 AND permission = $3
		)`

	var hasDirectAccess bool
	err := r.db.QueryRowContext(ctx, query, reportID, userID, permission).Scan(&hasDirectAccess)
	if err != nil {
		return false, fmt.Errorf("failed to check report access: %w", err)
	}
	if hasDirectAccess {
		return true, nil
	}

	// Check if report is public and permission is read
	if permission == domain.ReportPermissionRead {
		query = `SELECT is_public FROM reports WHERE id = $1`
		var isPublic bool
		err = r.db.QueryRowContext(ctx, query, reportID).Scan(&isPublic)
		if err != nil && err != sql.ErrNoRows {
			return false, fmt.Errorf("failed to check report public status: %w", err)
		}
		if isPublic {
			return true, nil
		}
	}

	// Check role-based access (requires user's role from users table)
	query = `
		SELECT EXISTS(
			SELECT 1 FROM report_access ra
			JOIN users u ON u.role = ra.role::varchar
			WHERE ra.report_id = $1 AND u.id = $2 AND ra.permission = $3
		)`

	var hasRoleAccess bool
	err = r.db.QueryRowContext(ctx, query, reportID, userID, permission).Scan(&hasRoleAccess)
	if err != nil {
		return false, fmt.Errorf("failed to check role-based access: %w", err)
	}

	return hasRoleAccess, nil
}

// AddComment adds a comment to a report
func (r *ReportRepositoryPG) AddComment(ctx context.Context, comment *entities.ReportComment) error {
	query := `
		INSERT INTO report_comments (report_id, author_id, content, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		comment.ReportID, comment.AuthorID, comment.Content,
		comment.CreatedAt, comment.UpdatedAt,
	).Scan(&comment.ID)

	if err != nil {
		return fmt.Errorf("failed to add report comment: %w", err)
	}
	return nil
}

// UpdateComment updates a comment
func (r *ReportRepositoryPG) UpdateComment(ctx context.Context, comment *entities.ReportComment) error {
	query := `UPDATE report_comments SET content = $1, updated_at = $2 WHERE id = $3`
	result, err := r.db.ExecContext(ctx, query, comment.Content, comment.UpdatedAt, comment.ID)
	if err != nil {
		return fmt.Errorf("failed to update comment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("comment not found: %d", comment.ID)
	}
	return nil
}

// DeleteComment deletes a comment
func (r *ReportRepositoryPG) DeleteComment(ctx context.Context, commentID int64) error {
	query := `DELETE FROM report_comments WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, commentID)
	if err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("comment not found: %d", commentID)
	}
	return nil
}

// GetCommentsByReport retrieves all comments for a report
func (r *ReportRepositoryPG) GetCommentsByReport(ctx context.Context, reportID int64) ([]*entities.ReportComment, error) {
	query := `
		SELECT id, report_id, author_id, content, created_at, updated_at
		FROM report_comments WHERE report_id = $1
		ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to get report comments: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var comments []*entities.ReportComment
	for rows.Next() {
		comment := &entities.ReportComment{}
		err := rows.Scan(
			&comment.ID, &comment.ReportID, &comment.AuthorID,
			&comment.Content, &comment.CreatedAt, &comment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, comment)
	}
	return comments, rows.Err()
}

// AddHistory adds a history entry for a report
func (r *ReportRepositoryPG) AddHistory(ctx context.Context, history *entities.ReportHistory) error {
	query := `
		INSERT INTO report_history (report_id, user_id, action, details, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		history.ReportID, history.UserID, history.Action,
		history.Details, history.CreatedAt,
	).Scan(&history.ID)

	if err != nil {
		return fmt.Errorf("failed to add report history: %w", err)
	}
	return nil
}

// GetHistoryByReport retrieves history entries for a report
func (r *ReportRepositoryPG) GetHistoryByReport(ctx context.Context, reportID int64, limit, offset int) ([]*entities.ReportHistory, error) {
	query := `
		SELECT id, report_id, user_id, action, details, created_at
		FROM report_history WHERE report_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, reportID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get report history: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var histories []*entities.ReportHistory
	for rows.Next() {
		history := &entities.ReportHistory{}
		err := rows.Scan(
			&history.ID, &history.ReportID, &history.UserID,
			&history.Action, &history.Details, &history.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan history: %w", err)
		}
		histories = append(histories, history)
	}
	return histories, rows.Err()
}

// CreateGenerationLog creates a generation log entry
func (r *ReportRepositoryPG) CreateGenerationLog(ctx context.Context, log *entities.ReportGenerationLog) error {
	query := `
		INSERT INTO report_generation_log (report_id, started_at, status)
		VALUES ($1, $2, $3)
		RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		log.ReportID, log.StartedAt, log.Status,
	).Scan(&log.ID)

	if err != nil {
		return fmt.Errorf("failed to create generation log: %w", err)
	}
	return nil
}

// UpdateGenerationLog updates a generation log entry
func (r *ReportRepositoryPG) UpdateGenerationLog(ctx context.Context, log *entities.ReportGenerationLog) error {
	query := `
		UPDATE report_generation_log SET
			completed_at = $1, status = $2, error_message = $3,
			duration_seconds = $4, records_processed = $5
		WHERE id = $6`

	result, err := r.db.ExecContext(ctx, query,
		log.CompletedAt, log.Status, log.ErrorMessage,
		log.DurationSeconds, log.RecordsProcessed, log.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update generation log: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("generation log not found: %d", log.ID)
	}
	return nil
}

// GetGenerationLogsByReport retrieves generation logs for a report
func (r *ReportRepositoryPG) GetGenerationLogsByReport(ctx context.Context, reportID int64) ([]*entities.ReportGenerationLog, error) {
	query := `
		SELECT id, report_id, started_at, completed_at, status,
			error_message, duration_seconds, records_processed
		FROM report_generation_log WHERE report_id = $1
		ORDER BY started_at DESC`

	rows, err := r.db.QueryContext(ctx, query, reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to get generation logs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var logs []*entities.ReportGenerationLog
	for rows.Next() {
		log := &entities.ReportGenerationLog{}
		err := rows.Scan(
			&log.ID, &log.ReportID, &log.StartedAt, &log.CompletedAt,
			&log.Status, &log.ErrorMessage, &log.DurationSeconds, &log.RecordsProcessed,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan generation log: %w", err)
		}
		logs = append(logs, log)
	}
	return logs, rows.Err()
}
