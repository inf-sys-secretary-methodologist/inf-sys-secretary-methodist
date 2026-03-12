package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

// Project use case errors.
var (
	ErrProjectNotFound     = errors.New("project not found")
	ErrCannotModifyProject = errors.New("cannot modify project")
)

// ProjectUseCase provides project management operations.
type ProjectUseCase struct {
	projectRepo repositories.ProjectRepository
	auditLogger *logging.AuditLogger
}

// NewProjectUseCase creates a new ProjectUseCase.
func NewProjectUseCase(
	projectRepo repositories.ProjectRepository,
	auditLogger *logging.AuditLogger,
) *ProjectUseCase {
	return &ProjectUseCase{
		projectRepo: projectRepo,
		auditLogger: auditLogger,
	}
}

// Create creates a new project.
func (uc *ProjectUseCase) Create(ctx context.Context, userID int64, input dto.CreateProjectInput) (*entities.Project, error) {
	project := entities.NewProject(input.Name, userID)
	project.Description = input.Description
	project.StartDate = input.StartDate
	project.EndDate = input.EndDate

	if err := uc.projectRepo.Create(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	uc.logAudit(ctx, userID, "project.created", project.ID)
	return project, nil
}

// GetByID retrieves a project by ID.
func (uc *ProjectUseCase) GetByID(ctx context.Context, id int64) (*entities.Project, error) {
	project, err := uc.projectRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	if project == nil {
		return nil, ErrProjectNotFound
	}
	return project, nil
}

// Update updates a project.
func (uc *ProjectUseCase) Update(ctx context.Context, userID, projectID int64, input dto.UpdateProjectInput) (*entities.Project, error) {
	project, err := uc.GetByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Only owner can update
	if project.OwnerID != userID {
		return nil, ErrUnauthorized
	}

	if input.Name != nil {
		project.Name = *input.Name
	}

	if input.Description != nil {
		project.Description = input.Description
	}

	if input.Status != nil {
		status := domain.ProjectStatus(*input.Status)
		if !status.IsValid() {
			return nil, fmt.Errorf("%w: invalid status", ErrInvalidInput)
		}
		project.Status = status
	}

	if input.StartDate != nil {
		project.StartDate = input.StartDate
	}

	if input.EndDate != nil {
		project.EndDate = input.EndDate
	}

	if err := uc.projectRepo.Save(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	uc.logAudit(ctx, userID, "project.updated", projectID)
	return project, nil
}

// Delete deletes a project.
func (uc *ProjectUseCase) Delete(ctx context.Context, userID, projectID int64) error {
	project, err := uc.GetByID(ctx, projectID)
	if err != nil {
		return err
	}

	// Only owner can delete
	if project.OwnerID != userID {
		return ErrUnauthorized
	}

	if err := uc.projectRepo.Delete(ctx, projectID); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	uc.logAudit(ctx, userID, "project.deleted", projectID)
	return nil
}

// List lists projects with filters.
func (uc *ProjectUseCase) List(ctx context.Context, input dto.ProjectFilterInput) (*dto.ProjectListOutput, error) {
	filter := input.ToProjectFilter()

	projects, err := uc.projectRepo.List(ctx, filter, input.Limit, input.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	total, err := uc.projectRepo.Count(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count projects: %w", err)
	}

	output := &dto.ProjectListOutput{
		Projects: make([]dto.ProjectOutput, 0, len(projects)),
		Total:    total,
		Limit:    input.Limit,
		Offset:   input.Offset,
	}

	for _, project := range projects {
		output.Projects = append(output.Projects, dto.ToProjectOutput(project))
	}

	return output, nil
}

// Activate activates a project.
func (uc *ProjectUseCase) Activate(ctx context.Context, userID, projectID int64) (*entities.Project, error) {
	project, err := uc.GetByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	if project.OwnerID != userID {
		return nil, ErrUnauthorized
	}

	project.Activate()

	if err := uc.projectRepo.Save(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to activate project: %w", err)
	}

	uc.logAudit(ctx, userID, "project.activated", projectID)
	return project, nil
}

// PutOnHold puts a project on hold.
func (uc *ProjectUseCase) PutOnHold(ctx context.Context, userID, projectID int64) (*entities.Project, error) {
	project, err := uc.GetByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	if project.OwnerID != userID {
		return nil, ErrUnauthorized
	}

	project.PutOnHold()

	if err := uc.projectRepo.Save(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to put project on hold: %w", err)
	}

	uc.logAudit(ctx, userID, "project.on_hold", projectID)
	return project, nil
}

// Complete completes a project.
func (uc *ProjectUseCase) Complete(ctx context.Context, userID, projectID int64) (*entities.Project, error) {
	project, err := uc.GetByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	if project.OwnerID != userID {
		return nil, ErrUnauthorized
	}

	project.Complete()

	if err := uc.projectRepo.Save(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to complete project: %w", err)
	}

	uc.logAudit(ctx, userID, "project.completed", projectID)
	return project, nil
}

// Cancel cancels a project.
func (uc *ProjectUseCase) Cancel(ctx context.Context, userID, projectID int64) (*entities.Project, error) {
	project, err := uc.GetByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	if project.OwnerID != userID {
		return nil, ErrUnauthorized
	}

	project.Cancel()

	if err := uc.projectRepo.Save(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to cancel project: %w", err)
	}

	uc.logAudit(ctx, userID, "project.canceled", projectID)
	return project, nil
}

// logAudit logs an audit event.
func (uc *ProjectUseCase) logAudit(ctx context.Context, userID int64, action string, resourceID int64) {
	if uc.auditLogger != nil {
		uc.auditLogger.LogAuditEvent(ctx, action, "project", map[string]interface{}{
			"user_id":     userID,
			"resource_id": resourceID,
		})
	}
}
