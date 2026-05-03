// Package usecases contains application use cases for the analytics module.
package usecases

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/analytics/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/analytics/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/analytics/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

// AnalyticsUseCase handles analytics business logic
type AnalyticsUseCase struct {
	analyticsRepo  repositories.AnalyticsRepository
	attendanceRepo repositories.AttendanceRepository
	gradeRepo      repositories.GradeRepository
	auditLogger    *logging.AuditLogger
}

// NewAnalyticsUseCase creates a new AnalyticsUseCase
func NewAnalyticsUseCase(
	analyticsRepo repositories.AnalyticsRepository,
	attendanceRepo repositories.AttendanceRepository,
	gradeRepo repositories.GradeRepository,
	auditLogger *logging.AuditLogger,
) *AnalyticsUseCase {
	return &AnalyticsUseCase{
		analyticsRepo:  analyticsRepo,
		attendanceRepo: attendanceRepo,
		gradeRepo:      gradeRepo,
		auditLogger:    auditLogger,
	}
}

// GetAtRiskStudents returns students who are at risk based on attendance and grades.
// When scope is non-nil (teacher role), the underlying query is filtered to
// the scope's whitelist of group names; pagination totals reflect the
// post-filter count (filter is pushed down to SQL).
func (uc *AnalyticsUseCase) GetAtRiskStudents(ctx context.Context, scope *entities.TeacherScope, page, pageSize int) (*dto.AtRiskStudentsResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	students, total, err := uc.analyticsRepo.GetAtRiskStudents(ctx, scope, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get at-risk students: %w", err)
	}

	response := &dto.AtRiskStudentsResponse{
		Students: make([]dto.StudentRiskResponse, 0, len(students)),
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}

	for _, s := range students {
		response.Students = append(response.Students, *dto.ToStudentRiskResponse(&s))
	}

	return response, nil
}

// GetStudentRisk returns the risk assessment for a specific student.
// When scope is non-nil (teacher role), the student's group must be in
// the scope whitelist or ErrAnalyticsScopeForbidden is returned.
func (uc *AnalyticsUseCase) GetStudentRisk(ctx context.Context, scope *entities.TeacherScope, studentID int64) (*dto.StudentRiskResponse, error) {
	risk, err := uc.analyticsRepo.GetStudentRisk(ctx, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get student risk: %w", err)
	}

	if scope != nil {
		if risk == nil || !scope.AllowsGroupPtr(risk.GroupName) {
			return nil, entities.ErrAnalyticsScopeForbidden
		}
	}

	return dto.ToStudentRiskResponse(risk), nil
}

// GetGroupSummary returns analytics summary for a specific group.
// When scope is non-nil, the group must be in the scope whitelist or
// ErrAnalyticsScopeForbidden is returned without contacting the repository.
func (uc *AnalyticsUseCase) GetGroupSummary(ctx context.Context, scope *entities.TeacherScope, groupName string) (*dto.GroupSummaryResponse, error) {
	if scope != nil && !scope.AllowsGroup(groupName) {
		return nil, entities.ErrAnalyticsScopeForbidden
	}

	summary, err := uc.analyticsRepo.GetGroupSummary(ctx, groupName)
	if err != nil {
		return nil, fmt.Errorf("failed to get group summary: %w", err)
	}

	return dto.ToGroupSummaryResponse(summary), nil
}

// GetAllGroupsSummary returns analytics summary for all groups.
// When scope is non-nil, the result is filtered to the scope's whitelist
// in the repository layer.
func (uc *AnalyticsUseCase) GetAllGroupsSummary(ctx context.Context, scope *entities.TeacherScope) (*dto.AllGroupsSummaryResponse, error) {
	summaries, err := uc.analyticsRepo.GetAllGroupsSummary(ctx, scope)
	if err != nil {
		return nil, fmt.Errorf("failed to get all groups summary: %w", err)
	}

	response := &dto.AllGroupsSummaryResponse{
		Groups: make([]dto.GroupSummaryResponse, 0, len(summaries)),
		Total:  len(summaries),
	}

	for _, s := range summaries {
		response.Groups = append(response.Groups, *dto.ToGroupSummaryResponse(&s))
	}

	return response, nil
}

// GetStudentsByRiskLevel returns students filtered by risk level.
// When scope is non-nil, the result is further restricted to the scope's
// whitelist of group names.
func (uc *AnalyticsUseCase) GetStudentsByRiskLevel(ctx context.Context, scope *entities.TeacherScope, riskLevel string, page, pageSize int) (*dto.AtRiskStudentsResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	level := entities.RiskLevel(riskLevel)
	offset := (page - 1) * pageSize

	students, total, err := uc.analyticsRepo.GetStudentsByRiskLevel(ctx, scope, level, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get students by risk level: %w", err)
	}

	response := &dto.AtRiskStudentsResponse{
		Students: make([]dto.StudentRiskResponse, 0, len(students)),
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}

	for _, s := range students {
		response.Students = append(response.Students, *dto.ToStudentRiskResponse(&s))
	}

	return response, nil
}

// GetAttendanceTrend returns monthly attendance trend data
func (uc *AnalyticsUseCase) GetAttendanceTrend(ctx context.Context, months int) (*dto.AttendanceTrendResponse, error) {
	if months < 1 || months > 24 {
		months = 6
	}

	trends, err := uc.analyticsRepo.GetMonthlyAttendanceTrend(ctx, months)
	if err != nil {
		return nil, fmt.Errorf("failed to get attendance trend: %w", err)
	}

	response := &dto.AttendanceTrendResponse{
		Trends: make([]dto.MonthlyTrendResponse, 0, len(trends)),
		Months: months,
	}

	for _, t := range trends {
		response.Trends = append(response.Trends, *dto.ToMonthlyTrendResponse(&t))
	}

	return response, nil
}

// MarkAttendance marks attendance for a single student
func (uc *AnalyticsUseCase) MarkAttendance(ctx context.Context, req *dto.MarkAttendanceRequest, markedBy int64) (*dto.AttendanceRecordResponse, error) {
	lessonDate, err := time.Parse("2006-01-02", req.LessonDate)
	if err != nil {
		return nil, fmt.Errorf("invalid lesson date format: %w", err)
	}

	record := &entities.AttendanceRecord{
		StudentID:  req.StudentID,
		LessonID:   req.LessonID,
		LessonDate: lessonDate,
		Status:     entities.AttendanceStatus(req.Status),
		MarkedBy:   &markedBy,
	}

	if req.Notes != "" {
		record.Notes = &req.Notes
	}

	err = uc.attendanceRepo.MarkAttendance(ctx, record)
	if err != nil {
		return nil, fmt.Errorf("failed to mark attendance: %w", err)
	}

	if uc.auditLogger != nil {
		uc.auditLogger.LogAuditEvent(ctx, "create", "attendance_records", map[string]interface{}{
			"student_id":  req.StudentID,
			"lesson_id":   req.LessonID,
			"lesson_date": req.LessonDate,
			"status":      req.Status,
		})
	}

	return dto.ToAttendanceRecordResponse(record), nil
}

// BulkMarkAttendance marks attendance for multiple students at once
func (uc *AnalyticsUseCase) BulkMarkAttendance(ctx context.Context, req *dto.BulkMarkAttendanceRequest, markedBy int64) ([]dto.AttendanceRecordResponse, error) {
	lessonDate, err := time.Parse("2006-01-02", req.LessonDate)
	if err != nil {
		return nil, fmt.Errorf("invalid lesson date format: %w", err)
	}

	records := make([]entities.AttendanceRecord, 0, len(req.Records))
	for _, r := range req.Records {
		record := entities.AttendanceRecord{
			StudentID:  r.StudentID,
			LessonID:   req.LessonID,
			LessonDate: lessonDate,
			Status:     entities.AttendanceStatus(r.Status),
			MarkedBy:   &markedBy,
		}
		if r.Notes != "" {
			record.Notes = &r.Notes
		}
		records = append(records, record)
	}

	err = uc.attendanceRepo.BulkMarkAttendance(ctx, records)
	if err != nil {
		return nil, fmt.Errorf("failed to bulk mark attendance: %w", err)
	}

	if uc.auditLogger != nil {
		uc.auditLogger.LogAuditEvent(ctx, "create", "attendance_records", map[string]interface{}{
			"lesson_id":     req.LessonID,
			"lesson_date":   req.LessonDate,
			"records_count": len(records),
		})
	}

	response := make([]dto.AttendanceRecordResponse, 0, len(records))
	for _, r := range records {
		response = append(response, *dto.ToAttendanceRecordResponse(&r))
	}

	return response, nil
}

// GetLessonAttendance returns attendance records for a specific lesson on a date
func (uc *AnalyticsUseCase) GetLessonAttendance(ctx context.Context, lessonID int64, date string) (*dto.LessonAttendanceResponse, error) {
	records, err := uc.attendanceRepo.GetAttendanceByLesson(ctx, lessonID, date)
	if err != nil {
		return nil, fmt.Errorf("failed to get lesson attendance: %w", err)
	}

	response := &dto.LessonAttendanceResponse{
		LessonID:   lessonID,
		LessonDate: date,
		Records:    make([]dto.AttendanceRecordResponse, 0, len(records)),
	}

	for _, r := range records {
		response.Records = append(response.Records, *dto.ToAttendanceRecordResponse(&r))
		switch r.Status {
		case entities.AttendanceStatusPresent:
			response.Summary.Present++
		case entities.AttendanceStatusAbsent:
			response.Summary.Absent++
		case entities.AttendanceStatusLate:
			response.Summary.Late++
		case entities.AttendanceStatusExcused:
			response.Summary.Excused++
		}
		response.Summary.Total++
	}

	return response, nil
}

// CreateLesson creates a new lesson
func (uc *AnalyticsUseCase) CreateLesson(ctx context.Context, req *dto.CreateLessonRequest) (*entities.Lesson, error) {
	lessonType := entities.LessonTypeLecture
	if req.LessonType != "" {
		lessonType = entities.LessonType(req.LessonType)
	}

	lesson := &entities.Lesson{
		Name:       req.Name,
		Subject:    req.Subject,
		TeacherID:  req.TeacherID,
		GroupName:  req.GroupName,
		LessonType: lessonType,
	}

	err := uc.attendanceRepo.CreateLesson(ctx, lesson)
	if err != nil {
		return nil, fmt.Errorf("failed to create lesson: %w", err)
	}

	if uc.auditLogger != nil {
		uc.auditLogger.LogAuditEvent(ctx, "create", "lessons", map[string]interface{}{
			"lesson_id": lesson.ID,
			"name":      lesson.Name,
			"subject":   lesson.Subject,
		})
	}

	return lesson, nil
}

// GetRiskWeightConfig returns the current risk weight configuration.
func (uc *AnalyticsUseCase) GetRiskWeightConfig(ctx context.Context) (*dto.RiskWeightConfigResponse, error) {
	cfg, err := uc.analyticsRepo.GetRiskWeightConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get risk weight config: %w", err)
	}
	return &dto.RiskWeightConfigResponse{
		AttendanceWeight:      cfg.AttendanceWeight,
		GradeWeight:           cfg.GradeWeight,
		SubmissionWeight:      cfg.SubmissionWeight,
		InactivityWeight:      cfg.InactivityWeight,
		HighRiskThreshold:     cfg.HighRiskThreshold,
		CriticalRiskThreshold: cfg.CriticalRiskThreshold,
		UpdatedAt:             cfg.UpdatedAt,
	}, nil
}

// UpdateRiskWeightConfig updates the risk weight configuration (admin only).
func (uc *AnalyticsUseCase) UpdateRiskWeightConfig(ctx context.Context, req dto.UpdateRiskWeightConfigRequest, updatedBy int64) error {
	// Validate weights sum to 1.0
	sum := req.AttendanceWeight + req.GradeWeight + req.SubmissionWeight + req.InactivityWeight
	if math.Abs(sum-1.0) > 0.01 {
		return fmt.Errorf("weights must sum to 1.0, got %.2f", sum)
	}

	cfg := &entities.RiskWeightConfig{
		AttendanceWeight:      req.AttendanceWeight,
		GradeWeight:           req.GradeWeight,
		SubmissionWeight:      req.SubmissionWeight,
		InactivityWeight:      req.InactivityWeight,
		HighRiskThreshold:     req.HighRiskThreshold,
		CriticalRiskThreshold: req.CriticalRiskThreshold,
		UpdatedBy:             &updatedBy,
	}

	if err := uc.analyticsRepo.UpdateRiskWeightConfig(ctx, cfg); err != nil {
		return fmt.Errorf("failed to update risk weight config: %w", err)
	}

	if uc.auditLogger != nil {
		uc.auditLogger.LogAuditEvent(ctx, "update", "risk_weight_config", map[string]interface{}{
			"updated_by":        updatedBy,
			"attendance_weight": req.AttendanceWeight,
			"grade_weight":      req.GradeWeight,
		})
	}

	return nil
}

// GetStudentRiskHistory returns risk score history for a student.
// When scope is non-nil, the student's current group is loaded first and
// must be in the scope whitelist; otherwise ErrAnalyticsScopeForbidden is
// returned and the history query is skipped.
func (uc *AnalyticsUseCase) GetStudentRiskHistory(ctx context.Context, scope *entities.TeacherScope, studentID int64, limit int) (*dto.RiskHistoryResponse, error) {
	if scope != nil {
		risk, err := uc.analyticsRepo.GetStudentRisk(ctx, studentID)
		if err != nil {
			return nil, fmt.Errorf("failed to get student risk for scope check: %w", err)
		}
		if risk == nil || !scope.AllowsGroupPtr(risk.GroupName) {
			return nil, entities.ErrAnalyticsScopeForbidden
		}
	}

	if limit <= 0 {
		limit = 90
	}

	history, err := uc.analyticsRepo.GetStudentRiskHistory(ctx, studentID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get risk history: %w", err)
	}

	entries := make([]dto.RiskHistoryEntry, 0, len(history))
	for _, e := range history {
		entries = append(entries, dto.RiskHistoryEntryFromEntity(e))
	}

	return &dto.RiskHistoryResponse{
		StudentID: studentID,
		History:   entries,
		Total:     len(entries),
	}, nil
}
