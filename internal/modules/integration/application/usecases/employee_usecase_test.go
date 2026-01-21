package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/domain/entities"
)

// MockExternalEmployeeRepository implements ExternalEmployeeRepository for testing.
type MockExternalEmployeeRepository struct {
	employees     map[int64]*entities.ExternalEmployee
	employeesByEx map[string]*entities.ExternalEmployee
	nextID        int64
}

func NewMockExternalEmployeeRepository() *MockExternalEmployeeRepository {
	return &MockExternalEmployeeRepository{
		employees:     make(map[int64]*entities.ExternalEmployee),
		employeesByEx: make(map[string]*entities.ExternalEmployee),
		nextID:        1,
	}
}

func (m *MockExternalEmployeeRepository) Create(_ context.Context, employee *entities.ExternalEmployee) error {
	employee.ID = m.nextID
	m.nextID++
	m.employees[employee.ID] = employee
	m.employeesByEx[employee.ExternalID] = employee
	return nil
}

func (m *MockExternalEmployeeRepository) Update(_ context.Context, employee *entities.ExternalEmployee) error {
	m.employees[employee.ID] = employee
	m.employeesByEx[employee.ExternalID] = employee
	return nil
}

func (m *MockExternalEmployeeRepository) Upsert(ctx context.Context, employee *entities.ExternalEmployee) error {
	if existing, ok := m.employeesByEx[employee.ExternalID]; ok {
		employee.ID = existing.ID
		return m.Update(ctx, employee)
	}
	return m.Create(ctx, employee)
}

func (m *MockExternalEmployeeRepository) GetByID(_ context.Context, id int64) (*entities.ExternalEmployee, error) {
	if emp, exists := m.employees[id]; exists {
		copy := *emp
		return &copy, nil
	}
	return nil, nil
}

func (m *MockExternalEmployeeRepository) GetByExternalID(_ context.Context, externalID string) (*entities.ExternalEmployee, error) {
	if emp, exists := m.employeesByEx[externalID]; exists {
		copy := *emp
		return &copy, nil
	}
	return nil, nil
}

func (m *MockExternalEmployeeRepository) GetByCode(_ context.Context, code string) (*entities.ExternalEmployee, error) {
	for _, emp := range m.employees {
		if emp.Code == code {
			copy := *emp
			return &copy, nil
		}
	}
	return nil, nil
}

func (m *MockExternalEmployeeRepository) GetByLocalUserID(_ context.Context, localUserID int64) (*entities.ExternalEmployee, error) {
	for _, emp := range m.employees {
		if emp.LocalUserID != nil && *emp.LocalUserID == localUserID {
			copy := *emp
			return &copy, nil
		}
	}
	return nil, nil
}

func (m *MockExternalEmployeeRepository) List(_ context.Context, filter entities.ExternalEmployeeFilter) ([]*entities.ExternalEmployee, int64, error) {
	var result []*entities.ExternalEmployee
	for _, emp := range m.employees {
		// Simple filtering
		if filter.IsActive != nil && emp.IsActive != *filter.IsActive {
			continue
		}
		if filter.IsLinked != nil {
			isLinked := emp.IsLinked()
			if isLinked != *filter.IsLinked {
				continue
			}
		}
		result = append(result, emp)
	}

	total := int64(len(result))

	// Apply pagination
	if filter.Offset > 0 && filter.Offset < len(result) {
		result = result[filter.Offset:]
	}
	if filter.Limit > 0 && filter.Limit < len(result) {
		result = result[:filter.Limit]
	}

	return result, total, nil
}

func (m *MockExternalEmployeeRepository) GetUnlinked(_ context.Context, limit, offset int) ([]*entities.ExternalEmployee, int64, error) {
	var result []*entities.ExternalEmployee
	for _, emp := range m.employees {
		if !emp.IsLinked() {
			result = append(result, emp)
		}
	}

	total := int64(len(result))

	// Apply pagination
	if offset > 0 && offset < len(result) {
		result = result[offset:]
	}
	if limit > 0 && limit < len(result) {
		result = result[:limit]
	}

	return result, total, nil
}

func (m *MockExternalEmployeeRepository) LinkToLocalUser(_ context.Context, id int64, localUserID int64) error {
	if emp, exists := m.employees[id]; exists {
		emp.LinkToLocalUser(localUserID)
		return nil
	}
	return errors.New("employee not found")
}

func (m *MockExternalEmployeeRepository) Unlink(_ context.Context, id int64) error {
	if emp, exists := m.employees[id]; exists {
		emp.Unlink()
		return nil
	}
	return errors.New("employee not found")
}

func (m *MockExternalEmployeeRepository) Delete(_ context.Context, id int64) error {
	if emp, exists := m.employees[id]; exists {
		delete(m.employeesByEx, emp.ExternalID)
		delete(m.employees, id)
		return nil
	}
	return nil
}

func (m *MockExternalEmployeeRepository) MarkInactive(_ context.Context, ids []int64) error {
	for _, id := range ids {
		if emp, exists := m.employees[id]; exists {
			emp.IsActive = false
		}
	}
	return nil
}

func (m *MockExternalEmployeeRepository) GetActiveIDs(_ context.Context) ([]int64, error) {
	var ids []int64
	for id, emp := range m.employees {
		if emp.IsActive {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func (m *MockExternalEmployeeRepository) GetAllExternalIDs(_ context.Context) ([]string, error) {
	var ids []string
	for _, emp := range m.employees {
		ids = append(ids, emp.ExternalID)
	}
	return ids, nil
}

func (m *MockExternalEmployeeRepository) BulkUpsert(ctx context.Context, employees []*entities.ExternalEmployee) error {
	for _, emp := range employees {
		if err := m.Upsert(ctx, emp); err != nil {
			return err
		}
	}
	return nil
}

func (m *MockExternalEmployeeRepository) MarkInactiveExcept(_ context.Context, activeExternalIDs []string) error {
	activeSet := make(map[string]bool)
	for _, id := range activeExternalIDs {
		activeSet[id] = true
	}
	for _, emp := range m.employees {
		if !activeSet[emp.ExternalID] {
			emp.IsActive = false
		}
	}
	return nil
}

// Helper to create test employee
func createTestEmployee(repo *MockExternalEmployeeRepository, externalID, firstName, lastName string) *entities.ExternalEmployee {
	emp := entities.NewExternalEmployee(externalID, "CODE-"+externalID)
	emp.FirstName = firstName
	emp.LastName = lastName
	repo.Create(context.Background(), emp)
	return emp
}

// Tests

func TestEmployeeUseCase_List(t *testing.T) {
	repo := NewMockExternalEmployeeRepository()
	uc := NewEmployeeUseCase(repo)

	ctx := context.Background()

	// Create employees
	createTestEmployee(repo, "ext1", "John", "Doe")
	createTestEmployee(repo, "ext2", "Jane", "Smith")
	createTestEmployee(repo, "ext3", "Bob", "Johnson")

	// List all
	req := &dto.ExternalEmployeeListRequest{
		Limit: 10,
	}
	result, err := uc.List(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.Total != 3 {
		t.Errorf("expected total 3, got %d", result.Total)
	}

	if len(result.Items) != 3 {
		t.Errorf("expected 3 items, got %d", len(result.Items))
	}
}

func TestEmployeeUseCase_List_WithFilter(t *testing.T) {
	repo := NewMockExternalEmployeeRepository()
	uc := NewEmployeeUseCase(repo)

	ctx := context.Background()

	// Create employees
	emp1 := createTestEmployee(repo, "ext1", "John", "Doe")
	emp1.IsActive = true
	repo.Update(ctx, emp1)

	emp2 := createTestEmployee(repo, "ext2", "Jane", "Smith")
	emp2.IsActive = false
	repo.Update(ctx, emp2)

	// Filter by active
	isActive := true
	req := &dto.ExternalEmployeeListRequest{
		IsActive: &isActive,
		Limit:    10,
	}
	result, err := uc.List(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.Total != 1 {
		t.Errorf("expected total 1, got %d", result.Total)
	}
}

func TestEmployeeUseCase_GetByID(t *testing.T) {
	repo := NewMockExternalEmployeeRepository()
	uc := NewEmployeeUseCase(repo)

	ctx := context.Background()

	// Create employee
	emp := createTestEmployee(repo, "ext1", "John", "Doe")

	// Get by ID
	result, err := uc.GetByID(ctx, emp.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if result.FirstName != "John" {
		t.Errorf("expected first name 'John', got '%s'", result.FirstName)
	}
}

func TestEmployeeUseCase_GetByID_NotFound(t *testing.T) {
	repo := NewMockExternalEmployeeRepository()
	uc := NewEmployeeUseCase(repo)

	ctx := context.Background()

	// Get non-existent
	result, err := uc.GetByID(ctx, 999)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result != nil {
		t.Error("expected nil result for non-existent employee")
	}
}

func TestEmployeeUseCase_GetByExternalID(t *testing.T) {
	repo := NewMockExternalEmployeeRepository()
	uc := NewEmployeeUseCase(repo)

	ctx := context.Background()

	// Create employee
	createTestEmployee(repo, "ext1", "John", "Doe")

	// Get by external ID
	result, err := uc.GetByExternalID(ctx, "ext1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if result.ExternalID != "ext1" {
		t.Errorf("expected external ID 'ext1', got '%s'", result.ExternalID)
	}
}

func TestEmployeeUseCase_GetByExternalID_NotFound(t *testing.T) {
	repo := NewMockExternalEmployeeRepository()
	uc := NewEmployeeUseCase(repo)

	ctx := context.Background()

	// Get non-existent
	result, err := uc.GetByExternalID(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result != nil {
		t.Error("expected nil result for non-existent external ID")
	}
}

func TestEmployeeUseCase_LinkToLocalUser(t *testing.T) {
	repo := NewMockExternalEmployeeRepository()
	uc := NewEmployeeUseCase(repo)

	ctx := context.Background()

	// Create employee
	emp := createTestEmployee(repo, "ext1", "John", "Doe")

	// Link to local user
	err := uc.LinkToLocalUser(ctx, emp.ID, 42)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify linked
	result, _ := repo.GetByID(ctx, emp.ID)
	if !result.IsLinked() {
		t.Error("expected employee to be linked")
	}
	if *result.LocalUserID != 42 {
		t.Errorf("expected local user ID 42, got %d", *result.LocalUserID)
	}
}

func TestEmployeeUseCase_LinkToLocalUser_NotFound(t *testing.T) {
	repo := NewMockExternalEmployeeRepository()
	uc := NewEmployeeUseCase(repo)

	ctx := context.Background()

	// Try to link non-existent employee
	err := uc.LinkToLocalUser(ctx, 999, 42)
	if err == nil {
		t.Error("expected error for non-existent employee")
	}
}

func TestEmployeeUseCase_LinkToLocalUser_AlreadyLinked(t *testing.T) {
	repo := NewMockExternalEmployeeRepository()
	uc := NewEmployeeUseCase(repo)

	ctx := context.Background()

	// Create and link employee
	emp := createTestEmployee(repo, "ext1", "John", "Doe")
	repo.LinkToLocalUser(ctx, emp.ID, 42)

	// Try to link again
	err := uc.LinkToLocalUser(ctx, emp.ID, 43)
	if err == nil {
		t.Error("expected error for already linked employee")
	}
}

func TestEmployeeUseCase_LinkToLocalUser_LocalUserAlreadyLinked(t *testing.T) {
	repo := NewMockExternalEmployeeRepository()
	uc := NewEmployeeUseCase(repo)

	ctx := context.Background()

	// Create and link first employee
	emp1 := createTestEmployee(repo, "ext1", "John", "Doe")
	repo.LinkToLocalUser(ctx, emp1.ID, 42)

	// Create second employee
	emp2 := createTestEmployee(repo, "ext2", "Jane", "Smith")

	// Try to link second employee to same local user
	err := uc.LinkToLocalUser(ctx, emp2.ID, 42)
	if err == nil {
		t.Error("expected error when local user is already linked to another employee")
	}
}

func TestEmployeeUseCase_Unlink(t *testing.T) {
	repo := NewMockExternalEmployeeRepository()
	uc := NewEmployeeUseCase(repo)

	ctx := context.Background()

	// Create and link employee
	emp := createTestEmployee(repo, "ext1", "John", "Doe")
	repo.LinkToLocalUser(ctx, emp.ID, 42)

	// Unlink
	err := uc.Unlink(ctx, emp.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify unlinked
	result, _ := repo.GetByID(ctx, emp.ID)
	if result.IsLinked() {
		t.Error("expected employee to be unlinked")
	}
}

func TestEmployeeUseCase_Unlink_NotFound(t *testing.T) {
	repo := NewMockExternalEmployeeRepository()
	uc := NewEmployeeUseCase(repo)

	ctx := context.Background()

	// Try to unlink non-existent employee
	err := uc.Unlink(ctx, 999)
	if err == nil {
		t.Error("expected error for non-existent employee")
	}
}

func TestEmployeeUseCase_Unlink_NotLinked(t *testing.T) {
	repo := NewMockExternalEmployeeRepository()
	uc := NewEmployeeUseCase(repo)

	ctx := context.Background()

	// Create unlinked employee
	emp := createTestEmployee(repo, "ext1", "John", "Doe")

	// Try to unlink
	err := uc.Unlink(ctx, emp.ID)
	if err == nil {
		t.Error("expected error for unlinked employee")
	}
}

func TestEmployeeUseCase_GetUnlinked(t *testing.T) {
	repo := NewMockExternalEmployeeRepository()
	uc := NewEmployeeUseCase(repo)

	ctx := context.Background()

	// Create employees
	emp1 := createTestEmployee(repo, "ext1", "John", "Doe")
	createTestEmployee(repo, "ext2", "Jane", "Smith")
	createTestEmployee(repo, "ext3", "Bob", "Johnson")

	// Link first employee
	repo.LinkToLocalUser(ctx, emp1.ID, 42)

	// Get unlinked
	result, err := uc.GetUnlinked(ctx, 10, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.Total != 2 {
		t.Errorf("expected total 2 unlinked, got %d", result.Total)
	}
}

func TestEmployeeUseCase_Delete(t *testing.T) {
	repo := NewMockExternalEmployeeRepository()
	uc := NewEmployeeUseCase(repo)

	ctx := context.Background()

	// Create employee
	emp := createTestEmployee(repo, "ext1", "John", "Doe")

	// Delete
	err := uc.Delete(ctx, emp.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify deleted
	result, _ := repo.GetByID(ctx, emp.ID)
	if result != nil {
		t.Error("expected employee to be deleted")
	}
}
