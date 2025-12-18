package usecases

import (
	"context"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/domain/repositories"
)

// EmployeeUseCase handles external employee operations
type EmployeeUseCase struct {
	employeeRepo repositories.ExternalEmployeeRepository
}

// NewEmployeeUseCase creates a new employee use case
func NewEmployeeUseCase(employeeRepo repositories.ExternalEmployeeRepository) *EmployeeUseCase {
	return &EmployeeUseCase{
		employeeRepo: employeeRepo,
	}
}

// List retrieves external employees with filtering
func (uc *EmployeeUseCase) List(ctx context.Context, req *dto.ExternalEmployeeListRequest) (*dto.ExternalEmployeeListResponse, error) {
	filter := entities.ExternalEmployeeFilter{
		Search:     req.Search,
		Department: req.Department,
		Position:   req.Position,
		IsActive:   req.IsActive,
		IsLinked:   req.IsLinked,
		Limit:      req.Limit,
		Offset:     req.Offset,
	}

	employees, total, err := uc.employeeRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list employees: %w", err)
	}

	items := make([]*dto.ExternalEmployeeDTO, len(employees))
	for i, emp := range employees {
		items[i] = dto.FromExternalEmployee(emp)
	}

	return &dto.ExternalEmployeeListResponse{
		Items: items,
		Total: total,
	}, nil
}

// GetByID retrieves an external employee by ID
func (uc *EmployeeUseCase) GetByID(ctx context.Context, id int64) (*dto.ExternalEmployeeDTO, error) {
	emp, err := uc.employeeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get employee: %w", err)
	}
	if emp == nil {
		return nil, nil
	}
	return dto.FromExternalEmployee(emp), nil
}

// GetByExternalID retrieves an external employee by 1C external ID
func (uc *EmployeeUseCase) GetByExternalID(ctx context.Context, externalID string) (*dto.ExternalEmployeeDTO, error) {
	emp, err := uc.employeeRepo.GetByExternalID(ctx, externalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get employee by external ID: %w", err)
	}
	if emp == nil {
		return nil, nil
	}
	return dto.FromExternalEmployee(emp), nil
}

// LinkToLocalUser links an external employee to a local user
func (uc *EmployeeUseCase) LinkToLocalUser(ctx context.Context, id int64, localUserID int64) error {
	// Check if employee exists
	emp, err := uc.employeeRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get employee: %w", err)
	}
	if emp == nil {
		return fmt.Errorf("employee not found")
	}

	// Check if already linked
	if emp.IsLinked() {
		return fmt.Errorf("employee is already linked to user %d", *emp.LocalUserID)
	}

	// Check if local user is already linked to another employee
	existing, err := uc.employeeRepo.GetByLocalUserID(ctx, localUserID)
	if err != nil {
		return fmt.Errorf("failed to check existing link: %w", err)
	}
	if existing != nil {
		return fmt.Errorf("local user %d is already linked to employee %d", localUserID, existing.ID)
	}

	if err := uc.employeeRepo.LinkToLocalUser(ctx, id, localUserID); err != nil {
		return fmt.Errorf("failed to link employee: %w", err)
	}

	return nil
}

// Unlink removes the link between external employee and local user
func (uc *EmployeeUseCase) Unlink(ctx context.Context, id int64) error {
	emp, err := uc.employeeRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get employee: %w", err)
	}
	if emp == nil {
		return fmt.Errorf("employee not found")
	}

	if !emp.IsLinked() {
		return fmt.Errorf("employee is not linked")
	}

	if err := uc.employeeRepo.Unlink(ctx, id); err != nil {
		return fmt.Errorf("failed to unlink employee: %w", err)
	}

	return nil
}

// GetUnlinked retrieves employees not linked to local users
func (uc *EmployeeUseCase) GetUnlinked(ctx context.Context, limit, offset int) (*dto.ExternalEmployeeListResponse, error) {
	employees, total, err := uc.employeeRepo.GetUnlinked(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get unlinked employees: %w", err)
	}

	items := make([]*dto.ExternalEmployeeDTO, len(employees))
	for i, emp := range employees {
		items[i] = dto.FromExternalEmployee(emp)
	}

	return &dto.ExternalEmployeeListResponse{
		Items: items,
		Total: total,
	}, nil
}

// Delete deletes an external employee
func (uc *EmployeeUseCase) Delete(ctx context.Context, id int64) error {
	if err := uc.employeeRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete employee: %w", err)
	}
	return nil
}
