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

// ReportTypeRepositoryPG implements ReportTypeRepository using PostgreSQL
type ReportTypeRepositoryPG struct {
	db *sql.DB
}

// NewReportTypeRepositoryPG creates a new PostgreSQL report type repository
func NewReportTypeRepositoryPG(db *sql.DB) *ReportTypeRepositoryPG {
	return &ReportTypeRepositoryPG{db: db}
}

// Ensure ReportTypeRepositoryPG implements ReportTypeRepository
var _ repositories.ReportTypeRepository = (*ReportTypeRepositoryPG)(nil)

// Create inserts a new report type into the database
func (r *ReportTypeRepositoryPG) Create(ctx context.Context, reportType *entities.ReportType) error {
	query := `
		INSERT INTO report_types (
			name, code, description, category, template_path,
			output_format, is_periodic, period_type, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		reportType.Name, reportType.Code, reportType.Description,
		reportType.Category, reportType.TemplatePath, reportType.OutputFormat,
		reportType.IsPeriodic, reportType.PeriodType,
		reportType.CreatedAt, reportType.UpdatedAt,
	).Scan(&reportType.ID)

	if err != nil {
		return fmt.Errorf("failed to create report type: %w", err)
	}
	return nil
}

// Save updates an existing report type in the database
func (r *ReportTypeRepositoryPG) Save(ctx context.Context, reportType *entities.ReportType) error {
	query := `
		UPDATE report_types SET
			name = $1, code = $2, description = $3, category = $4,
			template_path = $5, output_format = $6, is_periodic = $7,
			period_type = $8, updated_at = $9
		WHERE id = $10`

	result, err := r.db.ExecContext(ctx, query,
		reportType.Name, reportType.Code, reportType.Description,
		reportType.Category, reportType.TemplatePath, reportType.OutputFormat,
		reportType.IsPeriodic, reportType.PeriodType,
		reportType.UpdatedAt, reportType.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to save report type: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("report type not found: %d", reportType.ID)
	}
	return nil
}

// GetByID retrieves a report type by its ID
func (r *ReportTypeRepositoryPG) GetByID(ctx context.Context, id int64) (*entities.ReportType, error) {
	query := `
		SELECT id, name, code, description, category, template_path,
			output_format, is_periodic, period_type, created_at, updated_at
		FROM report_types WHERE id = $1`

	reportType := &entities.ReportType{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&reportType.ID, &reportType.Name, &reportType.Code,
		&reportType.Description, &reportType.Category, &reportType.TemplatePath,
		&reportType.OutputFormat, &reportType.IsPeriodic, &reportType.PeriodType,
		&reportType.CreatedAt, &reportType.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get report type: %w", err)
	}
	return reportType, nil
}

// GetByCode retrieves a report type by its code
func (r *ReportTypeRepositoryPG) GetByCode(ctx context.Context, code string) (*entities.ReportType, error) {
	query := `
		SELECT id, name, code, description, category, template_path,
			output_format, is_periodic, period_type, created_at, updated_at
		FROM report_types WHERE code = $1`

	reportType := &entities.ReportType{}
	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&reportType.ID, &reportType.Name, &reportType.Code,
		&reportType.Description, &reportType.Category, &reportType.TemplatePath,
		&reportType.OutputFormat, &reportType.IsPeriodic, &reportType.PeriodType,
		&reportType.CreatedAt, &reportType.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get report type by code: %w", err)
	}
	return reportType, nil
}

// Delete removes a report type from the database
func (r *ReportTypeRepositoryPG) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM report_types WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete report type: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("report type not found: %d", id)
	}
	return nil
}

// List retrieves report types with filtering and pagination
func (r *ReportTypeRepositoryPG) List(ctx context.Context, filter repositories.ReportTypeFilter, limit, offset int) ([]*entities.ReportType, error) {
	query, args := r.buildListQuery(filter, limit, offset, false)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list report types: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return r.scanReportTypes(rows)
}

// Count returns the total count of report types matching the filter
func (r *ReportTypeRepositoryPG) Count(ctx context.Context, filter repositories.ReportTypeFilter) (int64, error) {
	query, args := r.buildListQuery(filter, 0, 0, true)

	var count int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count report types: %w", err)
	}
	return count, nil
}

// GetByCategory retrieves report types by category
func (r *ReportTypeRepositoryPG) GetByCategory(ctx context.Context, category domain.ReportCategory) ([]*entities.ReportType, error) {
	filter := repositories.ReportTypeFilter{Category: &category}
	return r.List(ctx, filter, 0, 0)
}

// GetPeriodic retrieves periodic report types
func (r *ReportTypeRepositoryPG) GetPeriodic(ctx context.Context) ([]*entities.ReportType, error) {
	isPeriodic := true
	filter := repositories.ReportTypeFilter{IsPeriodic: &isPeriodic}
	return r.List(ctx, filter, 0, 0)
}

// buildListQuery constructs the SQL query for listing report types
func (r *ReportTypeRepositoryPG) buildListQuery(filter repositories.ReportTypeFilter, limit, offset int, countOnly bool) (string, []any) {
	var conditions []string
	var args []any
	argNum := 1

	if filter.Category != nil {
		conditions = append(conditions, fmt.Sprintf("category = $%d", argNum))
		args = append(args, *filter.Category)
		argNum++
	}
	if filter.IsPeriodic != nil {
		conditions = append(conditions, fmt.Sprintf("is_periodic = $%d", argNum))
		args = append(args, *filter.IsPeriodic)
		argNum++
	}

	var query string
	if countOnly {
		query = "SELECT COUNT(*) FROM report_types"
	} else {
		query = `
			SELECT id, name, code, description, category, template_path,
				output_format, is_periodic, period_type, created_at, updated_at
			FROM report_types`
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	if !countOnly {
		query += " ORDER BY name ASC"
		if limit > 0 {
			query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argNum, argNum+1)
			args = append(args, limit, offset)
		}
	}

	return query, args
}

// scanReportTypes scans report type rows into entities
func (r *ReportTypeRepositoryPG) scanReportTypes(rows *sql.Rows) ([]*entities.ReportType, error) {
	var reportTypes []*entities.ReportType
	for rows.Next() {
		reportType := &entities.ReportType{}
		err := rows.Scan(
			&reportType.ID, &reportType.Name, &reportType.Code,
			&reportType.Description, &reportType.Category, &reportType.TemplatePath,
			&reportType.OutputFormat, &reportType.IsPeriodic, &reportType.PeriodType,
			&reportType.CreatedAt, &reportType.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan report type: %w", err)
		}
		reportTypes = append(reportTypes, reportType)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating report types: %w", err)
	}
	return reportTypes, nil
}

// AddParameter adds a parameter to a report type
func (r *ReportTypeRepositoryPG) AddParameter(ctx context.Context, param *entities.ReportParameter) error {
	query := `
		INSERT INTO report_parameters (
			report_type_id, parameter_name, parameter_type,
			is_required, default_value, options, display_order, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		param.ReportTypeID, param.ParameterName, param.ParameterType,
		param.IsRequired, param.DefaultValue, param.Options,
		param.DisplayOrder, param.CreatedAt,
	).Scan(&param.ID)

	if err != nil {
		return fmt.Errorf("failed to add parameter: %w", err)
	}
	return nil
}

// UpdateParameter updates a parameter
func (r *ReportTypeRepositoryPG) UpdateParameter(ctx context.Context, param *entities.ReportParameter) error {
	query := `
		UPDATE report_parameters SET
			parameter_name = $1, parameter_type = $2, is_required = $3,
			default_value = $4, options = $5, display_order = $6
		WHERE id = $7`

	result, err := r.db.ExecContext(ctx, query,
		param.ParameterName, param.ParameterType, param.IsRequired,
		param.DefaultValue, param.Options, param.DisplayOrder, param.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update parameter: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("parameter not found: %d", param.ID)
	}
	return nil
}

// DeleteParameter deletes a parameter
func (r *ReportTypeRepositoryPG) DeleteParameter(ctx context.Context, paramID int64) error {
	query := `DELETE FROM report_parameters WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, paramID)
	if err != nil {
		return fmt.Errorf("failed to delete parameter: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("parameter not found: %d", paramID)
	}
	return nil
}

// GetParametersByReportType retrieves parameters for a report type
func (r *ReportTypeRepositoryPG) GetParametersByReportType(ctx context.Context, reportTypeID int64) ([]*entities.ReportParameter, error) {
	query := `
		SELECT id, report_type_id, parameter_name, parameter_type,
			is_required, default_value, options, display_order, created_at
		FROM report_parameters WHERE report_type_id = $1
		ORDER BY display_order ASC`

	rows, err := r.db.QueryContext(ctx, query, reportTypeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get parameters: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var params []*entities.ReportParameter
	for rows.Next() {
		param := &entities.ReportParameter{}
		err := rows.Scan(
			&param.ID, &param.ReportTypeID, &param.ParameterName,
			&param.ParameterType, &param.IsRequired, &param.DefaultValue,
			&param.Options, &param.DisplayOrder, &param.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan parameter: %w", err)
		}
		params = append(params, param)
	}
	return params, rows.Err()
}

// AddTemplate adds a template to a report type
func (r *ReportTypeRepositoryPG) AddTemplate(ctx context.Context, template *entities.ReportTemplate) error {
	query := `
		INSERT INTO report_templates (
			report_type_id, name, content, is_default, created_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		template.ReportTypeID, template.Name, template.Content,
		template.IsDefault, template.CreatedBy,
		template.CreatedAt, template.UpdatedAt,
	).Scan(&template.ID)

	if err != nil {
		return fmt.Errorf("failed to add template: %w", err)
	}
	return nil
}

// UpdateTemplate updates a template
func (r *ReportTypeRepositoryPG) UpdateTemplate(ctx context.Context, template *entities.ReportTemplate) error {
	query := `
		UPDATE report_templates SET
			name = $1, content = $2, is_default = $3, updated_at = $4
		WHERE id = $5`

	result, err := r.db.ExecContext(ctx, query,
		template.Name, template.Content, template.IsDefault,
		template.UpdatedAt, template.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update template: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("template not found: %d", template.ID)
	}
	return nil
}

// DeleteTemplate deletes a template
func (r *ReportTypeRepositoryPG) DeleteTemplate(ctx context.Context, templateID int64) error {
	query := `DELETE FROM report_templates WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, templateID)
	if err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("template not found: %d", templateID)
	}
	return nil
}

// GetTemplatesByReportType retrieves templates for a report type
func (r *ReportTypeRepositoryPG) GetTemplatesByReportType(ctx context.Context, reportTypeID int64) ([]*entities.ReportTemplate, error) {
	query := `
		SELECT id, report_type_id, name, content, is_default, created_by, created_at, updated_at
		FROM report_templates WHERE report_type_id = $1
		ORDER BY name ASC`

	rows, err := r.db.QueryContext(ctx, query, reportTypeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get templates: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var templates []*entities.ReportTemplate
	for rows.Next() {
		template := &entities.ReportTemplate{}
		err := rows.Scan(
			&template.ID, &template.ReportTypeID, &template.Name,
			&template.Content, &template.IsDefault, &template.CreatedBy,
			&template.CreatedAt, &template.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan template: %w", err)
		}
		templates = append(templates, template)
	}
	return templates, rows.Err()
}

// GetDefaultTemplate retrieves the default template for a report type
func (r *ReportTypeRepositoryPG) GetDefaultTemplate(ctx context.Context, reportTypeID int64) (*entities.ReportTemplate, error) {
	query := `
		SELECT id, report_type_id, name, content, is_default, created_by, created_at, updated_at
		FROM report_templates WHERE report_type_id = $1 AND is_default = true`

	template := &entities.ReportTemplate{}
	err := r.db.QueryRowContext(ctx, query, reportTypeID).Scan(
		&template.ID, &template.ReportTypeID, &template.Name,
		&template.Content, &template.IsDefault, &template.CreatedBy,
		&template.CreatedAt, &template.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get default template: %w", err)
	}
	return template, nil
}

// SetDefaultTemplate sets a template as the default for its report type
func (r *ReportTypeRepositoryPG) SetDefaultTemplate(ctx context.Context, reportTypeID, templateID int64) error {
	// Start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Remove default from all templates of this report type
	_, err = tx.ExecContext(ctx,
		`UPDATE report_templates SET is_default = false WHERE report_type_id = $1`,
		reportTypeID,
	)
	if err != nil {
		return fmt.Errorf("failed to remove default: %w", err)
	}

	// Set new default
	result, err := tx.ExecContext(ctx,
		`UPDATE report_templates SET is_default = true WHERE id = $1 AND report_type_id = $2`,
		templateID, reportTypeID,
	)
	if err != nil {
		return fmt.Errorf("failed to set default template: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("template not found: %d", templateID)
	}

	return tx.Commit()
}

// Subscribe creates a subscription to a report type
func (r *ReportTypeRepositoryPG) Subscribe(ctx context.Context, subscription *entities.ReportSubscription) error {
	query := `
		INSERT INTO report_subscriptions (
			report_type_id, user_id, delivery_method, is_active, created_at
		) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (report_type_id, user_id) DO UPDATE SET
			delivery_method = EXCLUDED.delivery_method,
			is_active = EXCLUDED.is_active
		RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		subscription.ReportTypeID, subscription.UserID,
		subscription.DeliveryMethod, subscription.IsActive,
		subscription.CreatedAt,
	).Scan(&subscription.ID)

	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}
	return nil
}

// Unsubscribe removes a subscription
func (r *ReportTypeRepositoryPG) Unsubscribe(ctx context.Context, reportTypeID, userID int64) error {
	query := `DELETE FROM report_subscriptions WHERE report_type_id = $1 AND user_id = $2`
	result, err := r.db.ExecContext(ctx, query, reportTypeID, userID)
	if err != nil {
		return fmt.Errorf("failed to unsubscribe: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("subscription not found")
	}
	return nil
}

// GetSubscription retrieves a specific subscription
func (r *ReportTypeRepositoryPG) GetSubscription(ctx context.Context, reportTypeID, userID int64) (*entities.ReportSubscription, error) {
	query := `
		SELECT id, report_type_id, user_id, delivery_method, is_active, created_at
		FROM report_subscriptions WHERE report_type_id = $1 AND user_id = $2`

	subscription := &entities.ReportSubscription{}
	err := r.db.QueryRowContext(ctx, query, reportTypeID, userID).Scan(
		&subscription.ID, &subscription.ReportTypeID, &subscription.UserID,
		&subscription.DeliveryMethod, &subscription.IsActive, &subscription.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}
	return subscription, nil
}

// GetSubscribersByReportType retrieves all subscribers for a report type
func (r *ReportTypeRepositoryPG) GetSubscribersByReportType(ctx context.Context, reportTypeID int64) ([]*entities.ReportSubscription, error) {
	query := `
		SELECT id, report_type_id, user_id, delivery_method, is_active, created_at
		FROM report_subscriptions WHERE report_type_id = $1 AND is_active = true`

	rows, err := r.db.QueryContext(ctx, query, reportTypeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscribers: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return r.scanSubscriptions(rows)
}

// GetSubscriptionsByUser retrieves all subscriptions for a user
func (r *ReportTypeRepositoryPG) GetSubscriptionsByUser(ctx context.Context, userID int64) ([]*entities.ReportSubscription, error) {
	query := `
		SELECT id, report_type_id, user_id, delivery_method, is_active, created_at
		FROM report_subscriptions WHERE user_id = $1`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user subscriptions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return r.scanSubscriptions(rows)
}

// UpdateSubscription updates a subscription
func (r *ReportTypeRepositoryPG) UpdateSubscription(ctx context.Context, subscription *entities.ReportSubscription) error {
	query := `
		UPDATE report_subscriptions SET
			delivery_method = $1, is_active = $2
		WHERE id = $3`

	result, err := r.db.ExecContext(ctx, query,
		subscription.DeliveryMethod, subscription.IsActive, subscription.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("subscription not found: %d", subscription.ID)
	}
	return nil
}

// scanSubscriptions scans subscription rows into entities
func (r *ReportTypeRepositoryPG) scanSubscriptions(rows *sql.Rows) ([]*entities.ReportSubscription, error) {
	var subscriptions []*entities.ReportSubscription
	for rows.Next() {
		subscription := &entities.ReportSubscription{}
		err := rows.Scan(
			&subscription.ID, &subscription.ReportTypeID, &subscription.UserID,
			&subscription.DeliveryMethod, &subscription.IsActive, &subscription.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subscription: %w", err)
		}
		subscriptions = append(subscriptions, subscription)
	}
	return subscriptions, rows.Err()
}
