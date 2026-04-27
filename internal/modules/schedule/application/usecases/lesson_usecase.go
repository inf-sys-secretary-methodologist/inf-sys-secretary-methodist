// Package usecases provides application use cases for the schedule module.
package usecases

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

// Use case errors.
var (
	ErrLessonNotFound = errors.New("lesson not found")
	ErrForbidden      = errors.New("forbidden")
	ErrInvalidInput   = errors.New("invalid input")
)

// CreateLessonInputForUC represents usecase-level input for creating a lesson.
type CreateLessonInputForUC struct {
	SemesterID   int64
	DisciplineID int64
	LessonTypeID int64
	TeacherID    int64
	GroupID      int64
	ClassroomID  int64
	DayOfWeek    domain.DayOfWeek
	TimeStart    string
	TimeEnd      string
	WeekType     domain.WeekType
	Notes        *string
}

// UpdateLessonInputForUC represents usecase-level input for updating a lesson.
type UpdateLessonInputForUC struct {
	SemesterID   *int64
	DisciplineID *int64
	LessonTypeID *int64
	TeacherID    *int64
	GroupID      *int64
	ClassroomID  *int64
	DayOfWeek    *domain.DayOfWeek
	TimeStart    *string
	TimeEnd      *string
	WeekType     *domain.WeekType
	Notes        *string
}

// CreateChangeInputForUC represents usecase-level input for creating a schedule change.
type CreateChangeInputForUC struct {
	LessonID       int64
	ChangeType     domain.ChangeType
	OriginalDate   time.Time
	NewDate        *time.Time
	NewClassroomID *int64
	NewTeacherID   *int64
	Reason         *string
}

// LessonUseCase provides lesson management operations.
type LessonUseCase struct {
	lessonRepo    repositories.LessonRepository
	classroomRepo repositories.ClassroomRepository
	referenceRepo repositories.ReferenceRepository
	changeRepo    repositories.ScheduleChangeRepository
	auditLogger   *logging.AuditLogger
}

// NewLessonUseCase creates a new LessonUseCase.
func NewLessonUseCase(
	lessonRepo repositories.LessonRepository,
	classroomRepo repositories.ClassroomRepository,
	referenceRepo repositories.ReferenceRepository,
	changeRepo repositories.ScheduleChangeRepository,
	auditLogger *logging.AuditLogger,
) *LessonUseCase {
	return &LessonUseCase{
		lessonRepo:    lessonRepo,
		classroomRepo: classroomRepo,
		referenceRepo: referenceRepo,
		changeRepo:    changeRepo,
		auditLogger:   auditLogger,
	}
}

// Create creates a new lesson.
func (uc *LessonUseCase) Create(ctx context.Context, userID int64, input CreateLessonInputForUC) (*entities.Lesson, error) {
	lesson := entities.NewLesson(
		input.SemesterID,
		input.DisciplineID,
		input.LessonTypeID,
		input.TeacherID,
		input.GroupID,
		input.ClassroomID,
		input.DayOfWeek,
		input.TimeStart,
		input.TimeEnd,
		input.WeekType,
	)
	lesson.Notes = input.Notes

	if err := lesson.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	if err := uc.lessonRepo.Create(ctx, lesson); err != nil {
		return nil, fmt.Errorf("failed to create lesson: %w", err)
	}

	uc.logAudit(ctx, userID, "lesson.created", lesson.ID)
	return lesson, nil
}

// GetByID retrieves a lesson by ID.
func (uc *LessonUseCase) GetByID(ctx context.Context, id int64) (*entities.Lesson, error) {
	lesson, err := uc.lessonRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get lesson: %w", err)
	}
	if lesson == nil {
		return nil, ErrLessonNotFound
	}
	return lesson, nil
}

// Update updates a lesson.
func (uc *LessonUseCase) Update(ctx context.Context, userID, lessonID int64, input UpdateLessonInputForUC) (*entities.Lesson, error) {
	lesson, err := uc.GetByID(ctx, lessonID)
	if err != nil {
		return nil, err
	}

	if input.SemesterID != nil {
		lesson.SemesterID = *input.SemesterID
	}
	if input.DisciplineID != nil {
		lesson.DisciplineID = *input.DisciplineID
	}
	if input.LessonTypeID != nil {
		lesson.LessonTypeID = *input.LessonTypeID
	}
	if input.TeacherID != nil {
		lesson.TeacherID = *input.TeacherID
	}
	if input.GroupID != nil {
		lesson.GroupID = *input.GroupID
	}
	if input.ClassroomID != nil {
		lesson.ClassroomID = *input.ClassroomID
	}
	if input.DayOfWeek != nil {
		lesson.DayOfWeek = *input.DayOfWeek
	}
	if input.TimeStart != nil {
		lesson.TimeStart = *input.TimeStart
	}
	if input.TimeEnd != nil {
		lesson.TimeEnd = *input.TimeEnd
	}
	if input.WeekType != nil {
		lesson.WeekType = *input.WeekType
	}
	if input.Notes != nil {
		lesson.Notes = input.Notes
	}

	lesson.UpdatedAt = time.Now()

	if err := lesson.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	if err := uc.lessonRepo.Save(ctx, lesson); err != nil {
		return nil, fmt.Errorf("failed to update lesson: %w", err)
	}

	uc.logAudit(ctx, userID, "lesson.updated", lessonID)
	return lesson, nil
}

// Delete deletes a lesson.
func (uc *LessonUseCase) Delete(ctx context.Context, userID, lessonID int64) error {
	_, err := uc.GetByID(ctx, lessonID)
	if err != nil {
		return err
	}

	if err := uc.lessonRepo.Delete(ctx, lessonID); err != nil {
		return fmt.Errorf("failed to delete lesson: %w", err)
	}

	uc.logAudit(ctx, userID, "lesson.deleted", lessonID)
	return nil
}

// List lists lessons with filters.
func (uc *LessonUseCase) List(ctx context.Context, filter repositories.LessonFilter, limit, offset int) ([]*entities.Lesson, error) {
	return uc.lessonRepo.List(ctx, filter, limit, offset)
}

// Count counts lessons with filters.
func (uc *LessonUseCase) Count(ctx context.Context, filter repositories.LessonFilter) (int64, error) {
	return uc.lessonRepo.Count(ctx, filter)
}

// GetTimetable returns all lessons matching the filter with associations loaded.
func (uc *LessonUseCase) GetTimetable(ctx context.Context, filter repositories.LessonFilter) ([]*entities.Lesson, error) {
	return uc.lessonRepo.GetTimetable(ctx, filter)
}

// CreateChange creates a schedule change record.
func (uc *LessonUseCase) CreateChange(ctx context.Context, userID int64, input CreateChangeInputForUC) (*entities.ScheduleChange, error) {
	if !input.ChangeType.IsValid() {
		return nil, fmt.Errorf("%w: invalid change type", ErrInvalidInput)
	}

	change := &entities.ScheduleChange{
		LessonID:       input.LessonID,
		ChangeType:     input.ChangeType,
		OriginalDate:   input.OriginalDate,
		NewDate:        input.NewDate,
		NewClassroomID: input.NewClassroomID,
		NewTeacherID:   input.NewTeacherID,
		Reason:         input.Reason,
		CreatedBy:      userID,
		CreatedAt:      time.Now(),
	}

	if err := uc.changeRepo.Create(ctx, change); err != nil {
		return nil, fmt.Errorf("failed to create schedule change: %w", err)
	}

	uc.logAudit(ctx, userID, "schedule_change.created", change.ID)
	return change, nil
}

// ListChanges returns schedule changes for a given lesson.
func (uc *LessonUseCase) ListChanges(ctx context.Context, lessonID int64) ([]*entities.ScheduleChange, error) {
	return uc.changeRepo.GetByLessonID(ctx, lessonID)
}

// ListClassrooms lists classrooms with filters.
func (uc *LessonUseCase) ListClassrooms(ctx context.Context, filter repositories.ClassroomFilter, limit, offset int) ([]*entities.Classroom, error) {
	return uc.classroomRepo.List(ctx, filter, limit, offset)
}

// ListStudentGroups lists student groups.
func (uc *LessonUseCase) ListStudentGroups(ctx context.Context, limit, offset int) ([]*entities.StudentGroup, error) {
	return uc.referenceRepo.ListStudentGroups(ctx, limit, offset)
}

// ListDisciplines lists disciplines.
func (uc *LessonUseCase) ListDisciplines(ctx context.Context, limit, offset int) ([]*entities.Discipline, error) {
	return uc.referenceRepo.ListDisciplines(ctx, limit, offset)
}

// ListSemesters lists semesters.
func (uc *LessonUseCase) ListSemesters(ctx context.Context, activeOnly bool) ([]*entities.Semester, error) {
	return uc.referenceRepo.ListSemesters(ctx, activeOnly)
}

// ListLessonTypes lists lesson types.
func (uc *LessonUseCase) ListLessonTypes(ctx context.Context) ([]*entities.LessonType, error) {
	return uc.referenceRepo.ListLessonTypes(ctx)
}

// logAudit logs an audit event.
func (uc *LessonUseCase) logAudit(ctx context.Context, userID int64, action string, resourceID int64) {
	if uc.auditLogger != nil {
		uc.auditLogger.LogAuditEvent(ctx, action, "schedule_lesson", map[string]interface{}{
			"user_id":     userID,
			"resource_id": resourceID,
		})
	}
}
