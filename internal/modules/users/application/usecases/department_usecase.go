// Package usecases contains business logic for the users module.
package usecases

import (
	"context"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

// DepartmentUseCase handles department management business logic.
type DepartmentUseCase struct {
	departmentRepo repositories.DepartmentRepository
	auditLogger    *logging.AuditLogger
}

// NewDepartmentUseCase creates a new department use case.
func NewDepartmentUseCase(
	departmentRepo repositories.DepartmentRepository,
	auditLogger *logging.AuditLogger,
) *DepartmentUseCase {
	return &DepartmentUseCase{
		departmentRepo: departmentRepo,
		auditLogger:    auditLogger,
	}
}

// CreateDepartment creates a new department.
func (uc *DepartmentUseCase) CreateDepartment(ctx context.Context, input *dto.CreateDepartmentInput) (*entities.Department, error) {
	// Verify parent exists if provided
	if input.ParentID != nil {
		_, err := uc.departmentRepo.GetByID(ctx, *input.ParentID)
		if err != nil {
			return nil, err
		}
	}

	department := entities.NewDepartment(
		input.Name,
		input.Code,
		input.Description,
		input.ParentID,
	)

	err := uc.departmentRepo.Create(ctx, department)
	if err != nil {
		return nil, err
	}

	if uc.auditLogger != nil {
		uc.auditLogger.LogAuditEvent(ctx, "create", "department", map[string]interface{}{
			"department_id": department.ID,
			"name":          department.Name,
			"code":          department.Code,
		})
	}

	return department, nil
}

// GetDepartment retrieves a department by ID.
func (uc *DepartmentUseCase) GetDepartment(ctx context.Context, id int64) (*entities.Department, error) {
	return uc.departmentRepo.GetByID(ctx, id)
}

// GetDepartmentByCode retrieves a department by code.
func (uc *DepartmentUseCase) GetDepartmentByCode(ctx context.Context, code string) (*entities.Department, error) {
	return uc.departmentRepo.GetByCode(ctx, code)
}

// ListDepartments retrieves a paginated list of departments.
func (uc *DepartmentUseCase) ListDepartments(ctx context.Context, page, limit int, activeOnly bool) (*dto.DepartmentListResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	departments, err := uc.departmentRepo.List(ctx, limit, offset, activeOnly)
	if err != nil {
		return nil, err
	}

	total, err := uc.departmentRepo.Count(ctx, activeOnly)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &dto.DepartmentListResponse{
		Departments: departments,
		Total:       total,
		Page:        page,
		Limit:       limit,
		TotalPages:  totalPages,
	}, nil
}

// UpdateDepartment updates an existing department.
func (uc *DepartmentUseCase) UpdateDepartment(ctx context.Context, id int64, input *dto.UpdateDepartmentInput) (*entities.Department, error) {
	department, err := uc.departmentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Verify parent exists if provided
	if input.ParentID != nil && *input.ParentID != id {
		_, err := uc.departmentRepo.GetByID(ctx, *input.ParentID)
		if err != nil {
			return nil, err
		}
	}

	department.Name = input.Name
	department.Code = input.Code
	department.Description = input.Description
	department.ParentID = input.ParentID
	department.HeadID = input.HeadID
	if input.IsActive != nil {
		department.IsActive = *input.IsActive
	}
	department.UpdatedAt = time.Now()

	err = uc.departmentRepo.Update(ctx, department)
	if err != nil {
		return nil, err
	}

	if uc.auditLogger != nil {
		uc.auditLogger.LogAuditEvent(ctx, "update", "department", map[string]interface{}{
			"department_id": id,
			"name":          department.Name,
			"code":          department.Code,
		})
	}

	return department, nil
}

// DeleteDepartment deletes a department.
func (uc *DepartmentUseCase) DeleteDepartment(ctx context.Context, id int64) error {
	// Verify department exists
	department, err := uc.departmentRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if department has children
	children, err := uc.departmentRepo.GetChildren(ctx, id)
	if err != nil {
		return err
	}
	if len(children) > 0 {
		// Could return an error or cascade delete - for now, we'll prevent deletion
		return &DepartmentHasChildrenError{DepartmentID: id}
	}

	err = uc.departmentRepo.Delete(ctx, id)
	if err != nil {
		return err
	}

	if uc.auditLogger != nil {
		uc.auditLogger.LogAuditEvent(ctx, "delete", "department", map[string]interface{}{
			"department_id": id,
			"name":          department.Name,
		})
	}

	return nil
}

// GetDepartmentChildren retrieves child departments.
func (uc *DepartmentUseCase) GetDepartmentChildren(ctx context.Context, parentID int64) ([]*entities.Department, error) {
	return uc.departmentRepo.GetChildren(ctx, parentID)
}

// DepartmentHasChildrenError indicates a department has child departments.
type DepartmentHasChildrenError struct {
	DepartmentID int64
}

func (e *DepartmentHasChildrenError) Error() string {
	return "department has child departments and cannot be deleted"
}
