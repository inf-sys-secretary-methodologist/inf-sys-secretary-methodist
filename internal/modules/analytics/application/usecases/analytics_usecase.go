// Package usecases contains application use cases for the analytics module.
package usecases

import (
	"context"
	"fmt"
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

// GetAtRiskStudents returns students who are at risk based on attendance and grades
func (uc *AnalyticsUseCase) GetAtRiskStudents(ctx context.Context, page, pageSize int) (*dto.AtRiskStudentsResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	students, total, err := uc.analyticsRepo.GetAtRiskStudents(ctx, pageSize, offset)
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

// GetStudentRisk returns the risk assessment for a specific student
func (uc *AnalyticsUseCase) GetStudentRisk(ctx context.Context, studentID int64) (*dto.StudentRiskResponse, error) {
	risk, err := uc.analyticsRepo.GetStudentRisk(ctx, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get student risk: %w", err)
	}

	return dto.ToStudentRiskResponse(risk), nil
}

// GetGroupSummary returns analytics summary for a specific group
func (uc *AnalyticsUseCase) GetGroupSummary(ctx context.Context, groupName string) (*dto.GroupSummaryResponse, error) {
	summary, err := uc.analyticsRepo.GetGroupSummary(ctx, groupName)
	if err != nil {
		return nil, fmt.Errorf("failed to get group summary: %w", err)
	}

	return dto.ToGroupSummaryResponse(summary), nil
}

// GetAllGroupsSummary returns analytics summary for all groups
func (uc *AnalyticsUseCase) GetAllGroupsSummary(ctx context.Context) (*dto.AllGroupsSummaryResponse, error) {
	summaries, err := uc.analyticsRepo.GetAllGroupsSummary(ctx)
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

// GetStudentsByRiskLevel returns students filtered by risk level
func (uc *AnalyticsUseCase) GetStudentsByRiskLevel(ctx context.Context, riskLevel string, page, pageSize int) (*dto.AtRiskStudentsResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	level := entities.RiskLevel(riskLevel)
	offset := (page - 1) * pageSize

	students, total, err := uc.analyticsRepo.GetStudentsByRiskLevel(ctx, level, pageSize, offset)
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
