package usecases

import (
	"context"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/domain/repositories"
)

// StudentUseCase handles external student operations
type StudentUseCase struct {
	studentRepo repositories.ExternalStudentRepository
}

// NewStudentUseCase creates a new student use case
func NewStudentUseCase(studentRepo repositories.ExternalStudentRepository) *StudentUseCase {
	return &StudentUseCase{
		studentRepo: studentRepo,
	}
}

// List retrieves external students with filtering
func (uc *StudentUseCase) List(ctx context.Context, req *dto.ExternalStudentListRequest) (*dto.ExternalStudentListResponse, error) {
	filter := entities.ExternalStudentFilter{
		Search:    req.Search,
		GroupName: req.GroupName,
		Faculty:   req.Faculty,
		Course:    req.Course,
		Status:    req.Status,
		IsActive:  req.IsActive,
		IsLinked:  req.IsLinked,
		Limit:     req.Limit,
		Offset:    req.Offset,
	}

	students, total, err := uc.studentRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list students: %w", err)
	}

	items := make([]*dto.ExternalStudentDTO, len(students))
	for i, student := range students {
		items[i] = dto.FromExternalStudent(student)
	}

	return &dto.ExternalStudentListResponse{
		Items: items,
		Total: total,
	}, nil
}

// GetByID retrieves an external student by ID
func (uc *StudentUseCase) GetByID(ctx context.Context, id int64) (*dto.ExternalStudentDTO, error) {
	student, err := uc.studentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get student: %w", err)
	}
	if student == nil {
		return nil, nil
	}
	return dto.FromExternalStudent(student), nil
}

// GetByExternalID retrieves an external student by 1C external ID
func (uc *StudentUseCase) GetByExternalID(ctx context.Context, externalID string) (*dto.ExternalStudentDTO, error) {
	student, err := uc.studentRepo.GetByExternalID(ctx, externalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get student by external ID: %w", err)
	}
	if student == nil {
		return nil, nil
	}
	return dto.FromExternalStudent(student), nil
}

// LinkToLocalUser links an external student to a local user
func (uc *StudentUseCase) LinkToLocalUser(ctx context.Context, id int64, localUserID int64) error {
	// Check if student exists
	student, err := uc.studentRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get student: %w", err)
	}
	if student == nil {
		return fmt.Errorf("student not found")
	}

	// Check if already linked
	if student.IsLinked() {
		return fmt.Errorf("student is already linked to user %d", *student.LocalUserID)
	}

	// Check if local user is already linked to another student
	existing, err := uc.studentRepo.GetByLocalUserID(ctx, localUserID)
	if err != nil {
		return fmt.Errorf("failed to check existing link: %w", err)
	}
	if existing != nil {
		return fmt.Errorf("local user %d is already linked to student %d", localUserID, existing.ID)
	}

	if err := uc.studentRepo.LinkToLocalUser(ctx, id, localUserID); err != nil {
		return fmt.Errorf("failed to link student: %w", err)
	}

	return nil
}

// Unlink removes the link between external student and local user
func (uc *StudentUseCase) Unlink(ctx context.Context, id int64) error {
	student, err := uc.studentRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get student: %w", err)
	}
	if student == nil {
		return fmt.Errorf("student not found")
	}

	if !student.IsLinked() {
		return fmt.Errorf("student is not linked")
	}

	if err := uc.studentRepo.Unlink(ctx, id); err != nil {
		return fmt.Errorf("failed to unlink student: %w", err)
	}

	return nil
}

// GetUnlinked retrieves students not linked to local users
func (uc *StudentUseCase) GetUnlinked(ctx context.Context, limit, offset int) (*dto.ExternalStudentListResponse, error) {
	students, total, err := uc.studentRepo.GetUnlinked(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get unlinked students: %w", err)
	}

	items := make([]*dto.ExternalStudentDTO, len(students))
	for i, student := range students {
		items[i] = dto.FromExternalStudent(student)
	}

	return &dto.ExternalStudentListResponse{
		Items: items,
		Total: total,
	}, nil
}

// GetByGroup retrieves students by group
func (uc *StudentUseCase) GetByGroup(ctx context.Context, groupName string) ([]*dto.ExternalStudentDTO, error) {
	students, err := uc.studentRepo.GetByGroup(ctx, groupName)
	if err != nil {
		return nil, fmt.Errorf("failed to get students by group: %w", err)
	}

	items := make([]*dto.ExternalStudentDTO, len(students))
	for i, student := range students {
		items[i] = dto.FromExternalStudent(student)
	}

	return items, nil
}

// GetByFaculty retrieves students by faculty
func (uc *StudentUseCase) GetByFaculty(ctx context.Context, faculty string) ([]*dto.ExternalStudentDTO, error) {
	students, err := uc.studentRepo.GetByFaculty(ctx, faculty)
	if err != nil {
		return nil, fmt.Errorf("failed to get students by faculty: %w", err)
	}

	items := make([]*dto.ExternalStudentDTO, len(students))
	for i, student := range students {
		items[i] = dto.FromExternalStudent(student)
	}

	return items, nil
}

// GetGroups retrieves all distinct group names
func (uc *StudentUseCase) GetGroups(ctx context.Context) (*dto.GroupsResponse, error) {
	groups, err := uc.studentRepo.GetGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get groups: %w", err)
	}
	return &dto.GroupsResponse{Groups: groups}, nil
}

// GetFaculties retrieves all distinct faculty names
func (uc *StudentUseCase) GetFaculties(ctx context.Context) (*dto.FacultiesResponse, error) {
	faculties, err := uc.studentRepo.GetFaculties(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get faculties: %w", err)
	}
	return &dto.FacultiesResponse{Faculties: faculties}, nil
}

// Delete deletes an external student
func (uc *StudentUseCase) Delete(ctx context.Context, id int64) error {
	if err := uc.studentRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete student: %w", err)
	}
	return nil
}
