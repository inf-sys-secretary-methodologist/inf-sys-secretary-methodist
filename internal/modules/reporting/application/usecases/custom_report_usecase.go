package usecases

import (
	"context"
	"errors"
	"math"

	"github.com/google/uuid"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain/repositories"
)

var (
	ErrCustomReportNotFound = errors.New("custom report not found")
	ErrUnauthorizedAccess   = errors.New("unauthorized access to custom report")
	ErrInvalidDataSource    = errors.New("invalid data source")
	ErrInvalidFields        = errors.New("at least one field is required")
)

// QueryBuilder defines the interface for executing and exporting reports
type QueryBuilder interface {
	Execute(ctx context.Context, report *entities.CustomReport, page, pageSize int) (*entities.ReportExecutionResult, error)
	Export(result *entities.ReportExecutionResult, options entities.ExportOptions, reportName string) ([]byte, string, error)
	GetAvailableFields(dataSource entities.DataSourceType) []entities.ReportField
}

// CustomReportUseCase handles custom report business logic
type CustomReportUseCase struct {
	repo         repositories.CustomReportRepository
	queryBuilder QueryBuilder
}

// NewCustomReportUseCase creates a new CustomReportUseCase
func NewCustomReportUseCase(
	repo repositories.CustomReportRepository,
	queryBuilder QueryBuilder,
) *CustomReportUseCase {
	return &CustomReportUseCase{
		repo:         repo,
		queryBuilder: queryBuilder,
	}
}

// Create creates a new custom report
func (uc *CustomReportUseCase) Create(ctx context.Context, input dto.CreateCustomReportInput, userID int64) (dto.CustomReportOutput, error) {
	// Validate data source
	dataSource := entities.DataSourceType(input.DataSource)
	if !dataSource.IsValid() {
		return dto.CustomReportOutput{}, ErrInvalidDataSource
	}

	// Validate fields
	if len(input.Fields) == 0 {
		return dto.CustomReportOutput{}, ErrInvalidFields
	}

	// Create entity
	report := entities.NewCustomReport(input.Name, input.Description, dataSource, userID)
	report.SetFields(dto.ToSelectedFields(input.Fields))
	report.SetFilters(dto.ToReportFilters(input.Filters))
	report.SetGroupings(dto.ToReportGroupings(input.Groupings))
	report.SetSortings(dto.ToReportSortings(input.Sortings))
	report.SetPublic(input.IsPublic)

	// Save to repository
	if err := uc.repo.Create(ctx, report); err != nil {
		return dto.CustomReportOutput{}, err
	}

	return dto.ToCustomReportOutput(report), nil
}

// GetByID retrieves a custom report by ID
func (uc *CustomReportUseCase) GetByID(ctx context.Context, id uuid.UUID, userID int64) (dto.CustomReportOutput, error) {
	report, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return dto.CustomReportOutput{}, err
	}
	if report == nil {
		return dto.CustomReportOutput{}, ErrCustomReportNotFound
	}

	// Check access: user must be creator or report must be public
	if report.CreatedBy != userID && !report.IsPublic {
		return dto.CustomReportOutput{}, ErrUnauthorizedAccess
	}

	return dto.ToCustomReportOutput(report), nil
}

// Update updates an existing custom report
func (uc *CustomReportUseCase) Update(ctx context.Context, id uuid.UUID, input dto.UpdateCustomReportInput, userID int64) (dto.CustomReportOutput, error) {
	report, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return dto.CustomReportOutput{}, err
	}
	if report == nil {
		return dto.CustomReportOutput{}, ErrCustomReportNotFound
	}

	// Only creator can update
	if report.CreatedBy != userID {
		return dto.CustomReportOutput{}, ErrUnauthorizedAccess
	}

	// Update fields if provided
	if input.Name != nil {
		report.Name = *input.Name
	}
	if input.Description != nil {
		report.Description = *input.Description
	}
	if input.DataSource != nil {
		dataSource := entities.DataSourceType(*input.DataSource)
		if !dataSource.IsValid() {
			return dto.CustomReportOutput{}, ErrInvalidDataSource
		}
		report.DataSource = dataSource
	}
	if input.Fields != nil {
		if len(input.Fields) == 0 {
			return dto.CustomReportOutput{}, ErrInvalidFields
		}
		report.SetFields(dto.ToSelectedFields(input.Fields))
	}
	if input.Filters != nil {
		report.SetFilters(dto.ToReportFilters(input.Filters))
	}
	if input.Groupings != nil {
		report.SetGroupings(dto.ToReportGroupings(input.Groupings))
	}
	if input.Sortings != nil {
		report.SetSortings(dto.ToReportSortings(input.Sortings))
	}
	if input.IsPublic != nil {
		report.SetPublic(*input.IsPublic)
	}

	if err := uc.repo.Update(ctx, report); err != nil {
		return dto.CustomReportOutput{}, err
	}

	return dto.ToCustomReportOutput(report), nil
}

// Delete deletes a custom report
func (uc *CustomReportUseCase) Delete(ctx context.Context, id uuid.UUID, userID int64) error {
	report, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if report == nil {
		return ErrCustomReportNotFound
	}

	// Only creator can delete
	if report.CreatedBy != userID {
		return ErrUnauthorizedAccess
	}

	return uc.repo.Delete(ctx, id)
}

// List lists custom reports with filtering and pagination
func (uc *CustomReportUseCase) List(ctx context.Context, input dto.CustomReportFilterInput, userID int64) (dto.CustomReportListOutput, error) {
	// Build filter - show user's own reports and public reports
	filter := dto.ToCustomReportFilter(input, nil)

	// Get total count first
	totalCount, err := uc.repo.Count(ctx, filter)
	if err != nil {
		return dto.CustomReportListOutput{}, err
	}

	// Get reports
	reports, err := uc.repo.List(ctx, filter)
	if err != nil {
		return dto.CustomReportListOutput{}, err
	}

	// Filter to show only accessible reports (own or public)
	accessibleReports := make([]*entities.CustomReport, 0)
	for _, r := range reports {
		if r.CreatedBy == userID || r.IsPublic {
			accessibleReports = append(accessibleReports, r)
		}
	}

	// Convert to DTOs
	outputs := make([]dto.CustomReportOutput, len(accessibleReports))
	for i, r := range accessibleReports {
		outputs[i] = dto.ToCustomReportOutput(r)
	}

	page := input.Page
	if page < 1 {
		page = 1
	}
	pageSize := input.PageSize
	if pageSize < 1 {
		pageSize = 10
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))

	return dto.CustomReportListOutput{
		Reports:    outputs,
		Total:      totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// Execute executes a custom report and returns the data
func (uc *CustomReportUseCase) Execute(ctx context.Context, id uuid.UUID, input dto.ExecuteReportInput, userID int64) (dto.ExecuteReportOutput, error) {
	report, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return dto.ExecuteReportOutput{}, err
	}
	if report == nil {
		return dto.ExecuteReportOutput{}, ErrCustomReportNotFound
	}

	// Check access
	if report.CreatedBy != userID && !report.IsPublic {
		return dto.ExecuteReportOutput{}, ErrUnauthorizedAccess
	}

	// Default pagination
	page := input.Page
	if page < 1 {
		page = 1
	}
	pageSize := input.PageSize
	if pageSize < 1 {
		pageSize = 50
	}
	if pageSize > 1000 {
		pageSize = 1000
	}

	// Execute the query
	result, err := uc.queryBuilder.Execute(ctx, report, page, pageSize)
	if err != nil {
		return dto.ExecuteReportOutput{}, err
	}

	// Convert columns to DTOs
	columns := make([]dto.ReportColumnOutput, len(result.Columns))
	for i, c := range result.Columns {
		columns[i] = dto.ReportColumnOutput{
			Key:   c.Key,
			Label: c.Label,
		}
	}

	return dto.ExecuteReportOutput{
		Columns:    columns,
		Rows:       result.Rows,
		TotalCount: result.TotalCount,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: result.TotalPages,
	}, nil
}

// Export exports a custom report to the specified format
func (uc *CustomReportUseCase) Export(ctx context.Context, id uuid.UUID, input dto.ExportReportInput, userID int64) ([]byte, string, error) {
	report, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, "", err
	}
	if report == nil {
		return nil, "", ErrCustomReportNotFound
	}

	// Check access
	if report.CreatedBy != userID && !report.IsPublic {
		return nil, "", ErrUnauthorizedAccess
	}

	// Execute the query to get all data (no pagination for export)
	result, err := uc.queryBuilder.Execute(ctx, report, 1, 10000)
	if err != nil {
		return nil, "", err
	}

	// Export based on format
	format := entities.ExportFormat(input.Format)
	options := entities.ExportOptions{
		Format:         format,
		IncludeHeaders: input.IncludeHeaders,
		PageSize:       input.PageSize,
		Orientation:    input.Orientation,
	}

	return uc.queryBuilder.Export(result, options, report.Name)
}

// GetMyReports gets reports created by the current user
func (uc *CustomReportUseCase) GetMyReports(ctx context.Context, page, pageSize int, userID int64) (dto.CustomReportListOutput, error) {
	input := dto.CustomReportFilterInput{
		Page:     page,
		PageSize: pageSize,
	}
	filter := dto.ToCustomReportFilter(input, &userID)

	totalCount, err := uc.repo.Count(ctx, filter)
	if err != nil {
		return dto.CustomReportListOutput{}, err
	}

	reports, err := uc.repo.List(ctx, filter)
	if err != nil {
		return dto.CustomReportListOutput{}, err
	}

	outputs := make([]dto.CustomReportOutput, len(reports))
	for i, r := range reports {
		outputs[i] = dto.ToCustomReportOutput(r)
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))

	return dto.CustomReportListOutput{
		Reports:    outputs,
		Total:      totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetPublicReports gets all public reports
func (uc *CustomReportUseCase) GetPublicReports(ctx context.Context, page, pageSize int) (dto.CustomReportListOutput, error) {
	isPublic := true
	input := dto.CustomReportFilterInput{
		IsPublic: &isPublic,
		Page:     page,
		PageSize: pageSize,
	}
	filter := dto.ToCustomReportFilter(input, nil)

	totalCount, err := uc.repo.Count(ctx, filter)
	if err != nil {
		return dto.CustomReportListOutput{}, err
	}

	reports, err := uc.repo.List(ctx, filter)
	if err != nil {
		return dto.CustomReportListOutput{}, err
	}

	outputs := make([]dto.CustomReportOutput, len(reports))
	for i, r := range reports {
		outputs[i] = dto.ToCustomReportOutput(r)
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(pageSize)))

	return dto.CustomReportListOutput{
		Reports:    outputs,
		Total:      totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}
