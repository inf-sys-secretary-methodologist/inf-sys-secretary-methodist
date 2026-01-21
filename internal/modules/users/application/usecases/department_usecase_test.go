package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/domain/entities"
)

// MockDepartmentRepository implements DepartmentRepository for testing.
type MockDepartmentRepository struct {
	departments map[int64]*entities.Department
	nextID      int64
}

func NewMockDepartmentRepository() *MockDepartmentRepository {
	return &MockDepartmentRepository{
		departments: make(map[int64]*entities.Department),
		nextID:      1,
	}
}

func (m *MockDepartmentRepository) Create(_ context.Context, department *entities.Department) error {
	department.ID = m.nextID
	m.nextID++
	m.departments[department.ID] = department
	return nil
}

func (m *MockDepartmentRepository) GetByID(_ context.Context, id int64) (*entities.Department, error) {
	dept, exists := m.departments[id]
	if !exists {
		return nil, errors.New("department not found")
	}
	return dept, nil
}

func (m *MockDepartmentRepository) GetByCode(_ context.Context, code string) (*entities.Department, error) {
	for _, dept := range m.departments {
		if dept.Code == code {
			return dept, nil
		}
	}
	return nil, errors.New("department not found")
}

func (m *MockDepartmentRepository) Update(_ context.Context, department *entities.Department) error {
	m.departments[department.ID] = department
	return nil
}

func (m *MockDepartmentRepository) Delete(_ context.Context, id int64) error {
	delete(m.departments, id)
	return nil
}

func (m *MockDepartmentRepository) List(_ context.Context, limit, offset int, _ bool) ([]*entities.Department, error) {
	var departments []*entities.Department
	i := 0
	for _, dept := range m.departments {
		if i >= offset && len(departments) < limit {
			departments = append(departments, dept)
		}
		i++
	}
	return departments, nil
}

func (m *MockDepartmentRepository) Count(_ context.Context, _ bool) (int64, error) {
	return int64(len(m.departments)), nil
}

func (m *MockDepartmentRepository) GetChildren(_ context.Context, parentID int64) ([]*entities.Department, error) {
	var children []*entities.Department
	for _, dept := range m.departments {
		if dept.ParentID != nil && *dept.ParentID == parentID {
			children = append(children, dept)
		}
	}
	return children, nil
}

// Tests

func TestDepartmentUseCase_CreateDepartment(t *testing.T) {
	repo := NewMockDepartmentRepository()
	uc := NewDepartmentUseCase(repo, nil)

	ctx := context.Background()
	input := &dto.CreateDepartmentInput{
		Name:        "IT Department",
		Code:        "IT",
		Description: "Information Technology",
	}

	dept, err := uc.CreateDepartment(ctx, input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if dept.ID == 0 {
		t.Error("expected department ID to be set")
	}

	if dept.Name != "IT Department" {
		t.Errorf("expected name 'IT Department', got '%s'", dept.Name)
	}

	if dept.Code != "IT" {
		t.Errorf("expected code 'IT', got '%s'", dept.Code)
	}

	if !dept.IsActive {
		t.Error("expected new department to be active")
	}
}

func TestDepartmentUseCase_CreateDepartment_WithParent(t *testing.T) {
	repo := NewMockDepartmentRepository()
	uc := NewDepartmentUseCase(repo, nil)

	ctx := context.Background()

	// Create parent department
	parent, _ := uc.CreateDepartment(ctx, &dto.CreateDepartmentInput{
		Name: "Parent Dept",
		Code: "PARENT",
	})

	// Create child department
	parentID := parent.ID
	input := &dto.CreateDepartmentInput{
		Name:     "Child Dept",
		Code:     "CHILD",
		ParentID: &parentID,
	}

	dept, err := uc.CreateDepartment(ctx, input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if dept.ParentID == nil || *dept.ParentID != parent.ID {
		t.Errorf("expected parent ID %d, got %v", parent.ID, dept.ParentID)
	}
}

func TestDepartmentUseCase_CreateDepartment_ParentNotFound(t *testing.T) {
	repo := NewMockDepartmentRepository()
	uc := NewDepartmentUseCase(repo, nil)

	ctx := context.Background()
	parentID := int64(999)
	input := &dto.CreateDepartmentInput{
		Name:     "Child Dept",
		Code:     "CHILD",
		ParentID: &parentID,
	}

	_, err := uc.CreateDepartment(ctx, input)
	if err == nil {
		t.Error("expected error for non-existent parent")
	}
}

func TestDepartmentUseCase_GetDepartment(t *testing.T) {
	repo := NewMockDepartmentRepository()
	uc := NewDepartmentUseCase(repo, nil)

	ctx := context.Background()

	// Create department
	created, _ := uc.CreateDepartment(ctx, &dto.CreateDepartmentInput{
		Name: "IT Department",
		Code: "IT",
	})

	// Get department
	dept, err := uc.GetDepartment(ctx, created.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if dept.ID != created.ID {
		t.Errorf("expected ID %d, got %d", created.ID, dept.ID)
	}
}

func TestDepartmentUseCase_GetDepartment_NotFound(t *testing.T) {
	repo := NewMockDepartmentRepository()
	uc := NewDepartmentUseCase(repo, nil)

	ctx := context.Background()

	_, err := uc.GetDepartment(ctx, 999)
	if err == nil {
		t.Error("expected error for non-existent department")
	}
}

func TestDepartmentUseCase_GetDepartmentByCode(t *testing.T) {
	repo := NewMockDepartmentRepository()
	uc := NewDepartmentUseCase(repo, nil)

	ctx := context.Background()

	// Create department
	_, _ = uc.CreateDepartment(ctx, &dto.CreateDepartmentInput{
		Name: "IT Department",
		Code: "IT",
	})

	// Get by code
	dept, err := uc.GetDepartmentByCode(ctx, "IT")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if dept.Code != "IT" {
		t.Errorf("expected code 'IT', got '%s'", dept.Code)
	}
}

func TestDepartmentUseCase_ListDepartments(t *testing.T) {
	repo := NewMockDepartmentRepository()
	uc := NewDepartmentUseCase(repo, nil)

	ctx := context.Background()

	// Create departments
	_, _ = uc.CreateDepartment(ctx, &dto.CreateDepartmentInput{Name: "Dept 1", Code: "D1"})
	_, _ = uc.CreateDepartment(ctx, &dto.CreateDepartmentInput{Name: "Dept 2", Code: "D2"})
	_, _ = uc.CreateDepartment(ctx, &dto.CreateDepartmentInput{Name: "Dept 3", Code: "D3"})

	// List
	result, err := uc.ListDepartments(ctx, 1, 10, true)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.Total != 3 {
		t.Errorf("expected total 3, got %d", result.Total)
	}

	if result.Page != 1 {
		t.Errorf("expected page 1, got %d", result.Page)
	}
}

func TestDepartmentUseCase_ListDepartments_Pagination(t *testing.T) {
	repo := NewMockDepartmentRepository()
	uc := NewDepartmentUseCase(repo, nil)

	ctx := context.Background()

	// Default page
	result, _ := uc.ListDepartments(ctx, 0, 10, true)
	if result.Page != 1 {
		t.Errorf("expected default page 1, got %d", result.Page)
	}

	// Default limit
	result, _ = uc.ListDepartments(ctx, 1, 0, true)
	if result.Limit != 10 {
		t.Errorf("expected default limit 10, got %d", result.Limit)
	}

	// Max limit
	result, _ = uc.ListDepartments(ctx, 1, 200, true)
	if result.Limit != 100 {
		t.Errorf("expected max limit 100, got %d", result.Limit)
	}
}

func TestDepartmentUseCase_UpdateDepartment(t *testing.T) {
	repo := NewMockDepartmentRepository()
	uc := NewDepartmentUseCase(repo, nil)

	ctx := context.Background()

	// Create department
	created, _ := uc.CreateDepartment(ctx, &dto.CreateDepartmentInput{
		Name: "IT Department",
		Code: "IT",
	})

	// Update
	isActive := false
	input := &dto.UpdateDepartmentInput{
		Name:        "IT Dept Updated",
		Code:        "ITU",
		Description: "Updated description",
		IsActive:    &isActive,
	}

	updated, err := uc.UpdateDepartment(ctx, created.ID, input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if updated.Name != "IT Dept Updated" {
		t.Errorf("expected name 'IT Dept Updated', got '%s'", updated.Name)
	}

	if updated.Code != "ITU" {
		t.Errorf("expected code 'ITU', got '%s'", updated.Code)
	}

	if updated.IsActive {
		t.Error("expected IsActive to be false")
	}
}

func TestDepartmentUseCase_UpdateDepartment_NotFound(t *testing.T) {
	repo := NewMockDepartmentRepository()
	uc := NewDepartmentUseCase(repo, nil)

	ctx := context.Background()

	_, err := uc.UpdateDepartment(ctx, 999, &dto.UpdateDepartmentInput{
		Name: "Test",
		Code: "TEST",
	})

	if err == nil {
		t.Error("expected error for non-existent department")
	}
}

func TestDepartmentUseCase_UpdateDepartment_WithParent(t *testing.T) {
	repo := NewMockDepartmentRepository()
	uc := NewDepartmentUseCase(repo, nil)

	ctx := context.Background()

	// Create departments
	parent, _ := uc.CreateDepartment(ctx, &dto.CreateDepartmentInput{Name: "Parent", Code: "P"})
	child, _ := uc.CreateDepartment(ctx, &dto.CreateDepartmentInput{Name: "Child", Code: "C"})

	// Update with parent
	parentID := parent.ID
	input := &dto.UpdateDepartmentInput{
		Name:     "Child Updated",
		Code:     "CU",
		ParentID: &parentID,
	}

	updated, err := uc.UpdateDepartment(ctx, child.ID, input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if updated.ParentID == nil || *updated.ParentID != parent.ID {
		t.Errorf("expected parent ID %d, got %v", parent.ID, updated.ParentID)
	}
}

func TestDepartmentUseCase_DeleteDepartment(t *testing.T) {
	repo := NewMockDepartmentRepository()
	uc := NewDepartmentUseCase(repo, nil)

	ctx := context.Background()

	// Create department
	created, _ := uc.CreateDepartment(ctx, &dto.CreateDepartmentInput{
		Name: "IT Department",
		Code: "IT",
	})

	// Delete
	err := uc.DeleteDepartment(ctx, created.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify deleted
	_, err = uc.GetDepartment(ctx, created.ID)
	if err == nil {
		t.Error("expected error for deleted department")
	}
}

func TestDepartmentUseCase_DeleteDepartment_NotFound(t *testing.T) {
	repo := NewMockDepartmentRepository()
	uc := NewDepartmentUseCase(repo, nil)

	ctx := context.Background()

	err := uc.DeleteDepartment(ctx, 999)
	if err == nil {
		t.Error("expected error for non-existent department")
	}
}

func TestDepartmentUseCase_DeleteDepartment_HasChildren(t *testing.T) {
	repo := NewMockDepartmentRepository()
	uc := NewDepartmentUseCase(repo, nil)

	ctx := context.Background()

	// Create parent
	parent, _ := uc.CreateDepartment(ctx, &dto.CreateDepartmentInput{
		Name: "Parent",
		Code: "P",
	})

	// Create child
	parentID := parent.ID
	_, _ = uc.CreateDepartment(ctx, &dto.CreateDepartmentInput{
		Name:     "Child",
		Code:     "C",
		ParentID: &parentID,
	})

	// Try to delete parent
	err := uc.DeleteDepartment(ctx, parent.ID)
	if err == nil {
		t.Error("expected error when deleting department with children")
	}

	var hasChildrenErr *DepartmentHasChildrenError
	if !errors.As(err, &hasChildrenErr) {
		t.Errorf("expected DepartmentHasChildrenError, got %T", err)
	}
}

func TestDepartmentUseCase_GetDepartmentChildren(t *testing.T) {
	repo := NewMockDepartmentRepository()
	uc := NewDepartmentUseCase(repo, nil)

	ctx := context.Background()

	// Create parent
	parent, _ := uc.CreateDepartment(ctx, &dto.CreateDepartmentInput{
		Name: "Parent",
		Code: "P",
	})

	// Create children
	parentID := parent.ID
	_, _ = uc.CreateDepartment(ctx, &dto.CreateDepartmentInput{Name: "Child 1", Code: "C1", ParentID: &parentID})
	_, _ = uc.CreateDepartment(ctx, &dto.CreateDepartmentInput{Name: "Child 2", Code: "C2", ParentID: &parentID})

	// Get children
	children, err := uc.GetDepartmentChildren(ctx, parent.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(children) != 2 {
		t.Errorf("expected 2 children, got %d", len(children))
	}
}

func TestDepartmentHasChildrenError_Error(t *testing.T) {
	err := &DepartmentHasChildrenError{DepartmentID: 42}
	expected := "department has child departments and cannot be deleted"

	if err.Error() != expected {
		t.Errorf("expected '%s', got '%s'", expected, err.Error())
	}
}
