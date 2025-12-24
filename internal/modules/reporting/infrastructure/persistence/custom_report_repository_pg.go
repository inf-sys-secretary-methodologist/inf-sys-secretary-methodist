package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain/repositories"
)

// CustomReportRepositoryPG is a PostgreSQL implementation of CustomReportRepository
type CustomReportRepositoryPG struct {
	db *sql.DB
}

// NewCustomReportRepositoryPG creates a new CustomReportRepositoryPG
func NewCustomReportRepositoryPG(db *sql.DB) *CustomReportRepositoryPG {
	return &CustomReportRepositoryPG{db: db}
}

// Create creates a new custom report
func (repo *CustomReportRepositoryPG) Create(ctx context.Context, report *entities.CustomReport) error {
	fieldsJSON, err := json.Marshal(report.Fields)
	if err != nil {
		return fmt.Errorf("failed to marshal fields: %w", err)
	}

	filtersJSON, err := json.Marshal(report.Filters)
	if err != nil {
		return fmt.Errorf("failed to marshal filters: %w", err)
	}

	groupingsJSON, err := json.Marshal(report.Groupings)
	if err != nil {
		return fmt.Errorf("failed to marshal groupings: %w", err)
	}

	sortingsJSON, err := json.Marshal(report.Sortings)
	if err != nil {
		return fmt.Errorf("failed to marshal sortings: %w", err)
	}

	query := `
		INSERT INTO custom_reports (id, name, description, data_source, fields, filters, groupings, sortings, created_at, updated_at, created_by, is_public)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	var description *string
	if report.Description != "" {
		description = &report.Description
	}

	_, err = repo.db.ExecContext(ctx, query,
		report.ID,
		report.Name,
		description,
		string(report.DataSource),
		fieldsJSON,
		filtersJSON,
		groupingsJSON,
		sortingsJSON,
		report.CreatedAt,
		report.UpdatedAt,
		report.CreatedBy,
		report.IsPublic,
	)

	if err != nil {
		return fmt.Errorf("failed to create custom report: %w", err)
	}

	return nil
}

// Update updates an existing custom report
func (repo *CustomReportRepositoryPG) Update(ctx context.Context, report *entities.CustomReport) error {
	fieldsJSON, err := json.Marshal(report.Fields)
	if err != nil {
		return fmt.Errorf("failed to marshal fields: %w", err)
	}

	filtersJSON, err := json.Marshal(report.Filters)
	if err != nil {
		return fmt.Errorf("failed to marshal filters: %w", err)
	}

	groupingsJSON, err := json.Marshal(report.Groupings)
	if err != nil {
		return fmt.Errorf("failed to marshal groupings: %w", err)
	}

	sortingsJSON, err := json.Marshal(report.Sortings)
	if err != nil {
		return fmt.Errorf("failed to marshal sortings: %w", err)
	}

	query := `
		UPDATE custom_reports
		SET name = $1, description = $2, data_source = $3, fields = $4, filters = $5,
		    groupings = $6, sortings = $7, updated_at = $8, is_public = $9
		WHERE id = $10
	`

	var description *string
	if report.Description != "" {
		description = &report.Description
	}

	report.UpdatedAt = time.Now()

	result, err := repo.db.ExecContext(ctx, query,
		report.Name,
		description,
		string(report.DataSource),
		fieldsJSON,
		filtersJSON,
		groupingsJSON,
		sortingsJSON,
		report.UpdatedAt,
		report.IsPublic,
		report.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update custom report: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("custom report not found")
	}

	return nil
}

// GetByID retrieves a custom report by ID
func (repo *CustomReportRepositoryPG) GetByID(ctx context.Context, id uuid.UUID) (*entities.CustomReport, error) {
	query := `
		SELECT id, name, description, data_source, fields, filters, groupings, sortings, created_at, updated_at, created_by, is_public
		FROM custom_reports
		WHERE id = $1
	`

	var (
		report      entities.CustomReport
		description sql.NullString
		fields      []byte
		filters     []byte
		groupings   []byte
		sortings    []byte
		dataSource  string
	)

	err := repo.db.QueryRowContext(ctx, query, id).Scan(
		&report.ID,
		&report.Name,
		&description,
		&dataSource,
		&fields,
		&filters,
		&groupings,
		&sortings,
		&report.CreatedAt,
		&report.UpdatedAt,
		&report.CreatedBy,
		&report.IsPublic,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get custom report: %w", err)
	}

	report.DataSource = entities.DataSourceType(dataSource)
	if description.Valid {
		report.Description = description.String
	}

	if len(fields) > 0 {
		if err := json.Unmarshal(fields, &report.Fields); err != nil {
			return nil, fmt.Errorf("failed to unmarshal fields: %w", err)
		}
	} else {
		report.Fields = []entities.SelectedField{}
	}

	if len(filters) > 0 {
		if err := json.Unmarshal(filters, &report.Filters); err != nil {
			return nil, fmt.Errorf("failed to unmarshal filters: %w", err)
		}
	} else {
		report.Filters = []entities.ReportFilterConfig{}
	}

	if len(groupings) > 0 {
		if err := json.Unmarshal(groupings, &report.Groupings); err != nil {
			return nil, fmt.Errorf("failed to unmarshal groupings: %w", err)
		}
	} else {
		report.Groupings = []entities.ReportGrouping{}
	}

	if len(sortings) > 0 {
		if err := json.Unmarshal(sortings, &report.Sortings); err != nil {
			return nil, fmt.Errorf("failed to unmarshal sortings: %w", err)
		}
	} else {
		report.Sortings = []entities.ReportSorting{}
	}

	return &report, nil
}

// Delete deletes a custom report by ID
func (repo *CustomReportRepositoryPG) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM custom_reports WHERE id = $1`

	result, err := repo.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete custom report: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("custom report not found")
	}

	return nil
}

// List lists custom reports with filtering and pagination
func (repo *CustomReportRepositoryPG) List(ctx context.Context, filter repositories.CustomReportFilter) ([]*entities.CustomReport, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.CreatedBy != nil {
		conditions = append(conditions, fmt.Sprintf("created_by = $%d", argIndex))
		args = append(args, *filter.CreatedBy)
		argIndex++
	}

	if filter.DataSource != nil {
		conditions = append(conditions, fmt.Sprintf("data_source = $%d", argIndex))
		args = append(args, string(*filter.DataSource))
		argIndex++
	}

	if filter.IsPublic != nil {
		conditions = append(conditions, fmt.Sprintf("is_public = $%d", argIndex))
		args = append(args, *filter.IsPublic)
		argIndex++
	}

	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+filter.Search+"%")
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Default pagination
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize

	query := fmt.Sprintf(`
		SELECT id, name, description, data_source, fields, filters, groupings, sortings, created_at, updated_at, created_by, is_public
		FROM custom_reports
		%s
		ORDER BY updated_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, pageSize, offset)

	rows, err := repo.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list custom reports: %w", err)
	}
	defer func() { _ = rows.Close() }()

	reports := make([]*entities.CustomReport, 0)
	for rows.Next() {
		var (
			report      entities.CustomReport
			description sql.NullString
			fields      []byte
			filters     []byte
			groupings   []byte
			sortings    []byte
			dataSource  string
		)

		err := rows.Scan(
			&report.ID,
			&report.Name,
			&description,
			&dataSource,
			&fields,
			&filters,
			&groupings,
			&sortings,
			&report.CreatedAt,
			&report.UpdatedAt,
			&report.CreatedBy,
			&report.IsPublic,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		report.DataSource = entities.DataSourceType(dataSource)
		if description.Valid {
			report.Description = description.String
		}

		if len(fields) > 0 {
			if err := json.Unmarshal(fields, &report.Fields); err != nil {
				return nil, fmt.Errorf("failed to unmarshal fields: %w", err)
			}
		} else {
			report.Fields = []entities.SelectedField{}
		}

		if len(filters) > 0 {
			if err := json.Unmarshal(filters, &report.Filters); err != nil {
				return nil, fmt.Errorf("failed to unmarshal filters: %w", err)
			}
		} else {
			report.Filters = []entities.ReportFilterConfig{}
		}

		if len(groupings) > 0 {
			if err := json.Unmarshal(groupings, &report.Groupings); err != nil {
				return nil, fmt.Errorf("failed to unmarshal groupings: %w", err)
			}
		} else {
			report.Groupings = []entities.ReportGrouping{}
		}

		if len(sortings) > 0 {
			if err := json.Unmarshal(sortings, &report.Sortings); err != nil {
				return nil, fmt.Errorf("failed to unmarshal sortings: %w", err)
			}
		} else {
			report.Sortings = []entities.ReportSorting{}
		}

		reports = append(reports, &report)
	}

	return reports, nil
}

// Count counts custom reports matching the filter
func (repo *CustomReportRepositoryPG) Count(ctx context.Context, filter repositories.CustomReportFilter) (int64, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.CreatedBy != nil {
		conditions = append(conditions, fmt.Sprintf("created_by = $%d", argIndex))
		args = append(args, *filter.CreatedBy)
		argIndex++
	}

	if filter.DataSource != nil {
		conditions = append(conditions, fmt.Sprintf("data_source = $%d", argIndex))
		args = append(args, string(*filter.DataSource))
		argIndex++
	}

	if filter.IsPublic != nil {
		conditions = append(conditions, fmt.Sprintf("is_public = $%d", argIndex))
		args = append(args, *filter.IsPublic)
		argIndex++
	}

	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+filter.Search+"%")
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	query := fmt.Sprintf(`SELECT COUNT(*) FROM custom_reports %s`, whereClause)

	var count int64
	err := repo.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count custom reports: %w", err)
	}

	return count, nil
}

// GetByCreator retrieves all custom reports created by a user
func (repo *CustomReportRepositoryPG) GetByCreator(ctx context.Context, creatorID int64, page, pageSize int) ([]*entities.CustomReport, error) {
	filter := repositories.CustomReportFilter{
		CreatedBy: &creatorID,
		Page:      page,
		PageSize:  pageSize,
	}
	return repo.List(ctx, filter)
}

// GetPublicReports retrieves all public custom reports
func (repo *CustomReportRepositoryPG) GetPublicReports(ctx context.Context, page, pageSize int) ([]*entities.CustomReport, error) {
	isPublic := true
	filter := repositories.CustomReportFilter{
		IsPublic: &isPublic,
		Page:     page,
		PageSize: pageSize,
	}
	return repo.List(ctx, filter)
}
