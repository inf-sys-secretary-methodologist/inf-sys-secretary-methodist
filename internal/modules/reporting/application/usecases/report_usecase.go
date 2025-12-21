// Package usecases contains business logic for the reporting module.
package usecases

import (
	"context"
	"errors"
	"fmt"
	"time"

	notifUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/storage"
)

// Common errors
var (
	ErrReportNotFound     = errors.New("report not found")
	ErrReportTypeNotFound = errors.New("report type not found")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrCannotModifyReport = errors.New("cannot modify report in current status")
	ErrInvalidInput       = errors.New("invalid input")
)

// ReportUseCase handles report business logic
type ReportUseCase struct {
	reportRepo          repositories.ReportRepository
	reportTypeRepo      repositories.ReportTypeRepository
	s3Client            *storage.S3Client
	auditLog            *logging.AuditLogger
	notificationUseCase *notifUsecases.NotificationUseCase
}

// NewReportUseCase creates a new report use case
func NewReportUseCase(
	reportRepo repositories.ReportRepository,
	reportTypeRepo repositories.ReportTypeRepository,
	s3Client *storage.S3Client,
	auditLog *logging.AuditLogger,
	notificationUseCase *notifUsecases.NotificationUseCase,
) *ReportUseCase {
	return &ReportUseCase{
		reportRepo:          reportRepo,
		reportTypeRepo:      reportTypeRepo,
		s3Client:            s3Client,
		auditLog:            auditLog,
		notificationUseCase: notificationUseCase,
	}
}

// Create creates a new report
func (uc *ReportUseCase) Create(ctx context.Context, authorID int64, input *dto.CreateReportInput) (*dto.ReportOutput, error) {
	// Validate report type exists
	reportType, err := uc.reportTypeRepo.GetByID(ctx, input.ReportTypeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get report type: %w", err)
	}
	if reportType == nil {
		return nil, ErrReportTypeNotFound
	}

	// Create report entity
	report := entities.NewReport(input.ReportTypeID, input.Title, authorID)
	report.Description = input.Description
	report.IsPublic = input.IsPublic

	if input.PeriodStart != nil && input.PeriodEnd != nil {
		report.SetPeriod(*input.PeriodStart, *input.PeriodEnd)
	}

	if input.Parameters != nil {
		if err := report.SetParameters(input.Parameters); err != nil {
			return nil, fmt.Errorf("failed to set parameters: %w", err)
		}
	}

	// Save to database
	if err := uc.reportRepo.Create(ctx, report); err != nil {
		return nil, fmt.Errorf("failed to create report: %w", err)
	}

	// Add history entry
	history := entities.NewReportHistory(report.ID, &authorID, entities.ReportActionCreated)
	_ = uc.reportRepo.AddHistory(ctx, history)

	// Log audit
	uc.logAudit(ctx, "create_report", authorID, report.ID, nil)

	// Load report type for output
	report.ReportType = reportType

	return dto.ToReportOutput(report), nil
}

// GetByID retrieves a report by ID
func (uc *ReportUseCase) GetByID(ctx context.Context, id, userID int64) (*dto.ReportOutput, error) {
	report, err := uc.reportRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get report: %w", err)
	}
	if report == nil {
		return nil, ErrReportNotFound
	}

	// Check access
	hasAccess, err := uc.reportRepo.HasAccess(ctx, id, userID, domain.ReportPermissionRead)
	if err != nil {
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	// Allow access if user is author or has explicit access
	if report.AuthorID != userID && !hasAccess && !report.IsPublic {
		return nil, ErrUnauthorized
	}

	// Load report type
	reportType, _ := uc.reportTypeRepo.GetByID(ctx, report.ReportTypeID)
	report.ReportType = reportType

	return dto.ToReportOutput(report), nil
}

// Update updates a report
func (uc *ReportUseCase) Update(ctx context.Context, id, userID int64, input *dto.UpdateReportInput) (*dto.ReportOutput, error) {
	report, err := uc.reportRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get report: %w", err)
	}
	if report == nil {
		return nil, ErrReportNotFound
	}

	// Check if user can edit
	if report.AuthorID != userID {
		hasAccess, err := uc.reportRepo.HasAccess(ctx, id, userID, domain.ReportPermissionWrite)
		if err != nil {
			return nil, fmt.Errorf("failed to check access: %w", err)
		}
		if !hasAccess {
			return nil, ErrUnauthorized
		}
	}

	// Check if report can be edited
	if !report.CanEdit() {
		return nil, ErrCannotModifyReport
	}

	// Apply updates
	if input.Title != nil {
		report.Title = *input.Title
	}
	if input.Description != nil {
		report.Description = input.Description
	}
	if input.PeriodStart != nil && input.PeriodEnd != nil {
		report.SetPeriod(*input.PeriodStart, *input.PeriodEnd)
	}
	if input.Parameters != nil {
		if err := report.SetParameters(input.Parameters); err != nil {
			return nil, fmt.Errorf("failed to set parameters: %w", err)
		}
	}
	if input.IsPublic != nil {
		report.IsPublic = *input.IsPublic
	}
	report.UpdatedAt = time.Now()

	// Save changes
	if err := uc.reportRepo.Save(ctx, report); err != nil {
		return nil, fmt.Errorf("failed to save report: %w", err)
	}

	// Add history entry
	history := entities.NewReportHistory(report.ID, &userID, entities.ReportActionUpdated)
	_ = uc.reportRepo.AddHistory(ctx, history)

	uc.logAudit(ctx, "update_report", userID, report.ID, nil)

	return dto.ToReportOutput(report), nil
}

// Delete deletes a report
func (uc *ReportUseCase) Delete(ctx context.Context, id, userID int64) error {
	report, err := uc.reportRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get report: %w", err)
	}
	if report == nil {
		return ErrReportNotFound
	}

	// Only author can delete
	if report.AuthorID != userID {
		return ErrUnauthorized
	}

	// Cannot delete published reports
	if report.IsFinalized() {
		return ErrCannotModifyReport
	}

	// Delete file from S3 if exists
	if report.FilePath != nil && uc.s3Client != nil {
		_ = uc.s3Client.Delete(ctx, *report.FilePath)
	}

	// Delete from database
	if err := uc.reportRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete report: %w", err)
	}

	uc.logAudit(ctx, "delete_report", userID, id, nil)

	return nil
}

// List lists reports with filtering
// TODO: Add access control filtering based on userID
func (uc *ReportUseCase) List(ctx context.Context, _ int64, input *dto.ReportFilterInput) (*dto.ReportListOutput, error) {
	// Build filter
	filter := repositories.ReportFilter{
		ReportTypeID: input.ReportTypeID,
		AuthorID:     input.AuthorID,
		IsPublic:     input.IsPublic,
		Search:       input.Search,
	}

	// Parse status
	if input.Status != nil {
		status := domain.ReportStatus(*input.Status)
		if status.IsValid() {
			filter.Status = &status
		}
	}

	// Parse dates
	if input.PeriodStart != nil {
		if t, err := time.Parse("2006-01-02", *input.PeriodStart); err == nil {
			filter.PeriodStart = &t
		}
	}
	if input.PeriodEnd != nil {
		if t, err := time.Parse("2006-01-02", *input.PeriodEnd); err == nil {
			filter.PeriodEnd = &t
		}
	}

	// Pagination
	page := input.Page
	if page < 1 {
		page = 1
	}
	pageSize := input.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// Get total count
	total, err := uc.reportRepo.Count(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count reports: %w", err)
	}

	// Get reports
	reports, err := uc.reportRepo.List(ctx, filter, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list reports: %w", err)
	}

	// Convert to output
	outputs := make([]*dto.ReportOutput, len(reports))
	for i, report := range reports {
		outputs[i] = dto.ToReportOutput(report)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &dto.ReportListOutput{
		Reports:    outputs,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// Generate starts report generation process
func (uc *ReportUseCase) Generate(ctx context.Context, id, userID int64, input *dto.GenerateReportInput) (*dto.ReportOutput, error) {
	report, err := uc.reportRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get report: %w", err)
	}
	if report == nil {
		return nil, ErrReportNotFound
	}

	// Check authorization
	if report.AuthorID != userID {
		return nil, ErrUnauthorized
	}

	// Update parameters if provided
	if input != nil && input.Parameters != nil {
		if err := report.SetParameters(input.Parameters); err != nil {
			return nil, fmt.Errorf("failed to set parameters: %w", err)
		}
	}

	// Start generation
	if err := report.StartGeneration(); err != nil {
		return nil, err
	}

	// Create generation log
	genLog := entities.NewReportGenerationLog(report.ID)
	if err := uc.reportRepo.CreateGenerationLog(ctx, genLog); err != nil {
		return nil, fmt.Errorf("failed to create generation log: %w", err)
	}

	// Save report status
	if err := uc.reportRepo.Save(ctx, report); err != nil {
		return nil, fmt.Errorf("failed to save report: %w", err)
	}

	// TODO: Trigger actual report generation asynchronously
	// For now, we'll simulate instant completion with sample data
	go uc.simulateGeneration(context.Background(), report, genLog, userID)

	uc.logAudit(ctx, "generate_report", userID, report.ID, nil)

	return dto.ToReportOutput(report), nil
}

// simulateGeneration simulates report generation (placeholder for actual implementation)
func (uc *ReportUseCase) simulateGeneration(ctx context.Context, report *entities.Report, genLog *entities.ReportGenerationLog, userID int64) {
	// Simulate processing time
	time.Sleep(2 * time.Second)

	// Set sample data
	sampleData := map[string]interface{}{
		"generated_at": time.Now().Format(time.RFC3339),
		"records":      100,
		"summary":      "Report generated successfully",
	}
	_ = report.SetData(sampleData)

	// Complete generation (without file for now)
	// In real implementation, this would generate actual file
	fileName := fmt.Sprintf("report_%d_%s.pdf", report.ID, time.Now().Format("20060102_150405"))
	_ = report.CompleteGeneration(fileName, "reports/"+fileName, 0, "application/pdf")
	_ = uc.reportRepo.Save(ctx, report)

	// Update generation log
	genLog.Complete(100)
	_ = uc.reportRepo.UpdateGenerationLog(ctx, genLog)

	// Add history
	history := entities.NewReportHistory(report.ID, &userID, entities.ReportActionUpdated)
	_ = history.SetDetails(map[string]string{"action": "generation_completed"})
	_ = uc.reportRepo.AddHistory(ctx, history)
}

// SubmitForReview submits a report for review
func (uc *ReportUseCase) SubmitForReview(ctx context.Context, id, userID int64) (*dto.ReportOutput, error) {
	report, err := uc.reportRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get report: %w", err)
	}
	if report == nil {
		return nil, ErrReportNotFound
	}

	if report.AuthorID != userID {
		return nil, ErrUnauthorized
	}

	if err := report.SubmitForReview(); err != nil {
		return nil, err
	}

	if err := uc.reportRepo.Save(ctx, report); err != nil {
		return nil, fmt.Errorf("failed to save report: %w", err)
	}

	// Add history
	history := entities.NewReportHistory(report.ID, &userID, entities.ReportActionSubmitted)
	_ = uc.reportRepo.AddHistory(ctx, history)

	uc.logAudit(ctx, "submit_report_for_review", userID, report.ID, nil)

	return dto.ToReportOutput(report), nil
}

// Review reviews a report (approve or reject)
func (uc *ReportUseCase) Review(ctx context.Context, id, reviewerID int64, input *dto.ReviewReportInput) (*dto.ReportOutput, error) {
	report, err := uc.reportRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get report: %w", err)
	}
	if report == nil {
		return nil, ErrReportNotFound
	}

	// Check reviewer has approval permission
	hasAccess, err := uc.reportRepo.HasAccess(ctx, id, reviewerID, domain.ReportPermissionApprove)
	if err != nil {
		return nil, fmt.Errorf("failed to check access: %w", err)
	}
	if !hasAccess {
		return nil, ErrUnauthorized
	}

	var action entities.ReportAction
	switch input.Action {
	case "approve":
		if err := report.Approve(reviewerID, input.Comment); err != nil {
			return nil, err
		}
		action = entities.ReportActionApproved
	case "reject":
		if err := report.Reject(reviewerID, input.Comment); err != nil {
			return nil, err
		}
		action = entities.ReportActionRejected
	default:
		return nil, ErrInvalidInput
	}

	if err := uc.reportRepo.Save(ctx, report); err != nil {
		return nil, fmt.Errorf("failed to save report: %w", err)
	}

	// Add history
	history := entities.NewReportHistory(report.ID, &reviewerID, action)
	_ = history.SetDetails(map[string]string{"comment": input.Comment})
	_ = uc.reportRepo.AddHistory(ctx, history)

	uc.logAudit(ctx, "review_report", reviewerID, report.ID, map[string]interface{}{"action": input.Action})

	return dto.ToReportOutput(report), nil
}

// Publish publishes a report
func (uc *ReportUseCase) Publish(ctx context.Context, id, userID int64, input *dto.PublishReportInput) (*dto.ReportOutput, error) {
	report, err := uc.reportRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get report: %w", err)
	}
	if report == nil {
		return nil, ErrReportNotFound
	}

	// Check publish permission
	hasAccess, err := uc.reportRepo.HasAccess(ctx, id, userID, domain.ReportPermissionPublish)
	if err != nil {
		return nil, fmt.Errorf("failed to check access: %w", err)
	}
	if report.AuthorID != userID && !hasAccess {
		return nil, ErrUnauthorized
	}

	if err := report.Publish(input.IsPublic); err != nil {
		return nil, err
	}

	if err := uc.reportRepo.Save(ctx, report); err != nil {
		return nil, fmt.Errorf("failed to save report: %w", err)
	}

	// Add history
	history := entities.NewReportHistory(report.ID, &userID, entities.ReportActionPublished)
	_ = history.SetDetails(map[string]interface{}{"is_public": input.IsPublic})
	_ = uc.reportRepo.AddHistory(ctx, history)

	uc.logAudit(ctx, "publish_report", userID, report.ID, nil)

	// Notify users with access about report publication
	if uc.notificationUseCase != nil {
		go func() {
			accesses, err := uc.reportRepo.GetAccessByReport(context.Background(), report.ID)
			if err == nil {
				for _, access := range accesses {
					if access.UserID != nil && *access.UserID != userID {
						_ = uc.notificationUseCase.SendSystemNotification(
							context.Background(),
							*access.UserID,
							"Отчёт опубликован",
							fmt.Sprintf("Отчёт «%s» был опубликован", report.Title),
						)
					}
				}
			}
		}()
	}

	return dto.ToReportOutput(report), nil
}

// AddAccess adds access permission to a report
func (uc *ReportUseCase) AddAccess(ctx context.Context, reportID, userID int64, input *dto.AddAccessInput) (*dto.ReportAccessOutput, error) {
	report, err := uc.reportRepo.GetByID(ctx, reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to get report: %w", err)
	}
	if report == nil {
		return nil, ErrReportNotFound
	}

	// Only author can manage access
	if report.AuthorID != userID {
		return nil, ErrUnauthorized
	}

	permission := domain.ReportPermission(input.Permission)
	var access *entities.ReportAccess

	switch {
	case input.UserID != nil:
		access = entities.NewReportAccessForUser(reportID, *input.UserID, permission, &userID)
	case input.Role != "":
		role := domain.AccessRole(input.Role)
		access = entities.NewReportAccessForRole(reportID, role, permission, &userID)
	default:
		return nil, ErrInvalidInput
	}

	if err := uc.reportRepo.AddAccess(ctx, access); err != nil {
		return nil, fmt.Errorf("failed to add access: %w", err)
	}

	return dto.ToReportAccessOutput(access), nil
}

// RemoveAccess removes access permission from a report
func (uc *ReportUseCase) RemoveAccess(ctx context.Context, reportID, accessID, userID int64) error {
	report, err := uc.reportRepo.GetByID(ctx, reportID)
	if err != nil {
		return fmt.Errorf("failed to get report: %w", err)
	}
	if report == nil {
		return ErrReportNotFound
	}

	if report.AuthorID != userID {
		return ErrUnauthorized
	}

	return uc.reportRepo.RemoveAccess(ctx, reportID, accessID)
}

// GetAccess retrieves access permissions for a report
func (uc *ReportUseCase) GetAccess(ctx context.Context, reportID, userID int64) ([]*dto.ReportAccessOutput, error) {
	report, err := uc.reportRepo.GetByID(ctx, reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to get report: %w", err)
	}
	if report == nil {
		return nil, ErrReportNotFound
	}

	if report.AuthorID != userID {
		return nil, ErrUnauthorized
	}

	accesses, err := uc.reportRepo.GetAccessByReport(ctx, reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to get access: %w", err)
	}

	outputs := make([]*dto.ReportAccessOutput, len(accesses))
	for i, access := range accesses {
		outputs[i] = dto.ToReportAccessOutput(access)
	}

	return outputs, nil
}

// AddComment adds a comment to a report
func (uc *ReportUseCase) AddComment(ctx context.Context, reportID, userID int64, input *dto.AddCommentInput) (*dto.ReportCommentOutput, error) {
	report, err := uc.reportRepo.GetByID(ctx, reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to get report: %w", err)
	}
	if report == nil {
		return nil, ErrReportNotFound
	}

	// Check user has at least read access
	hasAccess, err := uc.reportRepo.HasAccess(ctx, reportID, userID, domain.ReportPermissionRead)
	if err != nil {
		return nil, fmt.Errorf("failed to check access: %w", err)
	}
	if report.AuthorID != userID && !hasAccess && !report.IsPublic {
		return nil, ErrUnauthorized
	}

	comment := entities.NewReportComment(reportID, userID, input.Content)
	if err := uc.reportRepo.AddComment(ctx, comment); err != nil {
		return nil, fmt.Errorf("failed to add comment: %w", err)
	}

	return dto.ToReportCommentOutput(comment), nil
}

// GetComments retrieves comments for a report
func (uc *ReportUseCase) GetComments(ctx context.Context, reportID, userID int64) ([]*dto.ReportCommentOutput, error) {
	report, err := uc.reportRepo.GetByID(ctx, reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to get report: %w", err)
	}
	if report == nil {
		return nil, ErrReportNotFound
	}

	// Check access
	hasAccess, err := uc.reportRepo.HasAccess(ctx, reportID, userID, domain.ReportPermissionRead)
	if err != nil {
		return nil, fmt.Errorf("failed to check access: %w", err)
	}
	if report.AuthorID != userID && !hasAccess && !report.IsPublic {
		return nil, ErrUnauthorized
	}

	comments, err := uc.reportRepo.GetCommentsByReport(ctx, reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments: %w", err)
	}

	outputs := make([]*dto.ReportCommentOutput, len(comments))
	for i, comment := range comments {
		outputs[i] = dto.ToReportCommentOutput(comment)
	}

	return outputs, nil
}

// GetHistory retrieves history for a report
func (uc *ReportUseCase) GetHistory(ctx context.Context, reportID, userID int64, limit, offset int) ([]*dto.ReportHistoryOutput, error) {
	report, err := uc.reportRepo.GetByID(ctx, reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to get report: %w", err)
	}
	if report == nil {
		return nil, ErrReportNotFound
	}

	if report.AuthorID != userID {
		return nil, ErrUnauthorized
	}

	if limit <= 0 {
		limit = 50
	}

	history, err := uc.reportRepo.GetHistoryByReport(ctx, reportID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get history: %w", err)
	}

	outputs := make([]*dto.ReportHistoryOutput, len(history))
	for i, h := range history {
		outputs[i] = dto.ToReportHistoryOutput(h)
	}

	return outputs, nil
}

// GetReportTypes retrieves all report types
func (uc *ReportUseCase) GetReportTypes(ctx context.Context, input *dto.ReportTypeFilterInput) (*dto.ReportTypeListOutput, error) {
	filter := repositories.ReportTypeFilter{
		IsPeriodic: input.IsPeriodic,
	}

	if input.Category != nil {
		cat := domain.ReportCategory(*input.Category)
		if cat.IsValid() {
			filter.Category = &cat
		}
	}

	// Pagination
	page := input.Page
	if page < 1 {
		page = 1
	}
	pageSize := input.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	total, err := uc.reportTypeRepo.Count(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count report types: %w", err)
	}

	reportTypes, err := uc.reportTypeRepo.List(ctx, filter, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list report types: %w", err)
	}

	// Load parameters for each report type
	for _, rt := range reportTypes {
		params, _ := uc.reportTypeRepo.GetParametersByReportType(ctx, rt.ID)
		if params != nil {
			rt.Parameters = make([]entities.ReportParameter, len(params))
			for i, p := range params {
				rt.Parameters[i] = *p
			}
		}
	}

	outputs := make([]*dto.ReportTypeOutput, len(reportTypes))
	for i, rt := range reportTypes {
		outputs[i] = dto.ToReportTypeOutput(rt)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &dto.ReportTypeListOutput{
		ReportTypes: outputs,
		Total:       total,
		Page:        page,
		PageSize:    pageSize,
		TotalPages:  totalPages,
	}, nil
}

// GetReportTypeByID retrieves a report type by ID
func (uc *ReportUseCase) GetReportTypeByID(ctx context.Context, id int64) (*dto.ReportTypeOutput, error) {
	reportType, err := uc.reportTypeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get report type: %w", err)
	}
	if reportType == nil {
		return nil, ErrReportTypeNotFound
	}

	// Load parameters
	params, _ := uc.reportTypeRepo.GetParametersByReportType(ctx, id)
	if params != nil {
		reportType.Parameters = make([]entities.ReportParameter, len(params))
		for i, p := range params {
			reportType.Parameters[i] = *p
		}
	}

	// Load templates
	templates, _ := uc.reportTypeRepo.GetTemplatesByReportType(ctx, id)
	if templates != nil {
		reportType.Templates = make([]entities.ReportTemplate, len(templates))
		for i, t := range templates {
			reportType.Templates[i] = *t
		}
	}

	return dto.ToReportTypeOutput(reportType), nil
}

// logAudit logs audit events
func (uc *ReportUseCase) logAudit(ctx context.Context, action string, userID, resourceID int64, details map[string]interface{}) {
	if uc.auditLog == nil {
		return
	}

	fields := map[string]interface{}{
		"user_id":     userID,
		"resource_id": resourceID,
	}
	for k, v := range details {
		fields[k] = v
	}
	uc.auditLog.LogAuditEvent(ctx, action, "report", fields)
}
