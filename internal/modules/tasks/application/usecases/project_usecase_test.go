package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain"
)

const updatedProjectName = "Updated Project"

func TestProjectUseCase_Create(t *testing.T) {
	projectRepo := NewMockProjectRepository()
	uc := NewProjectUseCase(projectRepo, nil)

	ctx := context.Background()
	input := dto.CreateProjectInput{
		Name:        "Test Project",
		Description: strPtr("Test Description"),
	}

	project, err := uc.Create(ctx, 1, input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if project.ID == 0 {
		t.Error("expected project ID to be set")
	}

	if project.Name != "Test Project" {
		t.Errorf("expected name 'Test Project', got '%s'", project.Name)
	}

	if project.Status != domain.ProjectStatusPlanning {
		t.Errorf("expected status 'planning', got '%s'", project.Status)
	}

	if project.OwnerID != 1 {
		t.Errorf("expected owner_id 1, got %d", project.OwnerID)
	}
}

func TestProjectUseCase_GetByID(t *testing.T) {
	projectRepo := NewMockProjectRepository()
	uc := NewProjectUseCase(projectRepo, nil)

	ctx := context.Background()

	// Create project first
	input := dto.CreateProjectInput{Name: "Test Project"}
	created, _ := uc.Create(ctx, 1, input)

	// Get by ID
	project, err := uc.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if project.ID != created.ID {
		t.Errorf("expected project ID %d, got %d", created.ID, project.ID)
	}
}

func TestProjectUseCase_GetByID_NotFound(t *testing.T) {
	projectRepo := NewMockProjectRepository()
	uc := NewProjectUseCase(projectRepo, nil)

	ctx := context.Background()

	_, err := uc.GetByID(ctx, 999)
	if !errors.Is(err, ErrProjectNotFound) {
		t.Errorf("expected ErrProjectNotFound, got %v", err)
	}
}

func TestProjectUseCase_Update(t *testing.T) {
	projectRepo := NewMockProjectRepository()
	uc := NewProjectUseCase(projectRepo, nil)

	ctx := context.Background()

	// Create project first
	input := dto.CreateProjectInput{Name: "Test Project"}
	created, _ := uc.Create(ctx, 1, input)

	// Update
	newName := updatedProjectName
	updateInput := dto.UpdateProjectInput{Name: &newName}

	updated, err := uc.Update(ctx, 1, created.ID, updateInput)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if updated.Name != updatedProjectName {
		t.Errorf("expected name %q, got %q", updatedProjectName, updated.Name)
	}
}

func TestProjectUseCase_Update_Unauthorized(t *testing.T) {
	projectRepo := NewMockProjectRepository()
	uc := NewProjectUseCase(projectRepo, nil)

	ctx := context.Background()

	// Create project
	input := dto.CreateProjectInput{Name: "Test Project"}
	created, _ := uc.Create(ctx, 1, input)

	// Try to update by non-owner
	newName := updatedProjectName
	_, err := uc.Update(ctx, 2, created.ID, dto.UpdateProjectInput{Name: &newName})
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestProjectUseCase_Delete(t *testing.T) {
	projectRepo := NewMockProjectRepository()
	uc := NewProjectUseCase(projectRepo, nil)

	ctx := context.Background()

	// Create project
	input := dto.CreateProjectInput{Name: "Test Project"}
	created, _ := uc.Create(ctx, 1, input)

	// Delete (by owner)
	err := uc.Delete(ctx, 1, created.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify deleted
	_, err = uc.GetByID(ctx, created.ID)
	if !errors.Is(err, ErrProjectNotFound) {
		t.Error("expected project to be deleted")
	}
}

func TestProjectUseCase_Delete_Unauthorized(t *testing.T) {
	projectRepo := NewMockProjectRepository()
	uc := NewProjectUseCase(projectRepo, nil)

	ctx := context.Background()

	// Create project
	input := dto.CreateProjectInput{Name: "Test Project"}
	created, _ := uc.Create(ctx, 1, input)

	// Try to delete by non-owner
	err := uc.Delete(ctx, 2, created.ID)
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestProjectUseCase_List(t *testing.T) {
	projectRepo := NewMockProjectRepository()
	uc := NewProjectUseCase(projectRepo, nil)

	ctx := context.Background()

	// Create projects
	_, _ = uc.Create(ctx, 1, dto.CreateProjectInput{Name: "Project 1"})
	_, _ = uc.Create(ctx, 1, dto.CreateProjectInput{Name: "Project 2"})

	// List
	input := dto.ProjectFilterInput{Limit: 10}
	output, err := uc.List(ctx, input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if output.Total != 2 {
		t.Errorf("expected total 2, got %d", output.Total)
	}
}

func TestProjectUseCase_StatusWorkflow(t *testing.T) {
	projectRepo := NewMockProjectRepository()
	uc := NewProjectUseCase(projectRepo, nil)

	ctx := context.Background()

	// Create project
	project, _ := uc.Create(ctx, 1, dto.CreateProjectInput{Name: "Test Project"})
	if project.Status != domain.ProjectStatusPlanning {
		t.Errorf("expected initial status 'planning', got '%s'", project.Status)
	}

	// Activate
	project, err := uc.Activate(ctx, 1, project.ID)
	if err != nil {
		t.Fatalf("activate error: %v", err)
	}
	if project.Status != domain.ProjectStatusActive {
		t.Errorf("expected status 'active', got '%s'", project.Status)
	}

	// Put on hold
	project, err = uc.PutOnHold(ctx, 1, project.ID)
	if err != nil {
		t.Fatalf("put on hold error: %v", err)
	}
	if project.Status != domain.ProjectStatusOnHold {
		t.Errorf("expected status 'on_hold', got '%s'", project.Status)
	}

	// Complete
	project, err = uc.Complete(ctx, 1, project.ID)
	if err != nil {
		t.Fatalf("complete error: %v", err)
	}
	if project.Status != domain.ProjectStatusCompleted {
		t.Errorf("expected status 'completed', got '%s'", project.Status)
	}
}

func TestProjectUseCase_Cancel(t *testing.T) {
	projectRepo := NewMockProjectRepository()
	uc := NewProjectUseCase(projectRepo, nil)

	ctx := context.Background()

	// Create project
	project, _ := uc.Create(ctx, 1, dto.CreateProjectInput{Name: "Test Project"})

	// Cancel
	project, err := uc.Cancel(ctx, 1, project.ID)
	if err != nil {
		t.Fatalf("cancel error: %v", err)
	}
	if project.Status != domain.ProjectStatusCancelled {
		t.Errorf("expected status 'canceled', got '%s'", project.Status)
	}
}

func TestProjectUseCase_StatusChange_Unauthorized(t *testing.T) {
	projectRepo := NewMockProjectRepository()
	uc := NewProjectUseCase(projectRepo, nil)

	ctx := context.Background()

	// Create project
	project, _ := uc.Create(ctx, 1, dto.CreateProjectInput{Name: "Test Project"})

	// Try to activate by non-owner
	_, err := uc.Activate(ctx, 2, project.ID)
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}
