package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/domain/entities"
)

// MockExternalStudentRepository implements ExternalStudentRepository for testing.
type MockExternalStudentRepository struct {
	students     map[int64]*entities.ExternalStudent
	studentsByEx map[string]*entities.ExternalStudent
	nextID       int64
}

func NewMockExternalStudentRepository() *MockExternalStudentRepository {
	return &MockExternalStudentRepository{
		students:     make(map[int64]*entities.ExternalStudent),
		studentsByEx: make(map[string]*entities.ExternalStudent),
		nextID:       1,
	}
}

func (m *MockExternalStudentRepository) Create(_ context.Context, student *entities.ExternalStudent) error {
	student.ID = m.nextID
	m.nextID++
	m.students[student.ID] = student
	m.studentsByEx[student.ExternalID] = student
	return nil
}

func (m *MockExternalStudentRepository) Update(_ context.Context, student *entities.ExternalStudent) error {
	m.students[student.ID] = student
	m.studentsByEx[student.ExternalID] = student
	return nil
}

func (m *MockExternalStudentRepository) Upsert(ctx context.Context, student *entities.ExternalStudent) error {
	if existing, ok := m.studentsByEx[student.ExternalID]; ok {
		student.ID = existing.ID
		return m.Update(ctx, student)
	}
	return m.Create(ctx, student)
}

func (m *MockExternalStudentRepository) GetByID(_ context.Context, id int64) (*entities.ExternalStudent, error) {
	if student, exists := m.students[id]; exists {
		copiedStudent := *student
		return &copiedStudent, nil
	}
	return nil, nil
}

func (m *MockExternalStudentRepository) GetByExternalID(_ context.Context, externalID string) (*entities.ExternalStudent, error) {
	if student, exists := m.studentsByEx[externalID]; exists {
		cloned := *student
		return &cloned, nil
	}
	return nil, nil
}

func (m *MockExternalStudentRepository) GetByCode(_ context.Context, code string) (*entities.ExternalStudent, error) {
	for _, student := range m.students {
		if student.Code == code {
			cloned := *student
			return &cloned, nil
		}
	}
	return nil, nil
}

func (m *MockExternalStudentRepository) GetByLocalUserID(_ context.Context, localUserID int64) (*entities.ExternalStudent, error) {
	for _, student := range m.students {
		if student.LocalUserID != nil && *student.LocalUserID == localUserID {
			cloned := *student
			return &cloned, nil
		}
	}
	return nil, nil
}

func (m *MockExternalStudentRepository) List(_ context.Context, filter entities.ExternalStudentFilter) ([]*entities.ExternalStudent, int64, error) {
	var result []*entities.ExternalStudent
	for _, student := range m.students {
		// Apply filters
		if filter.IsActive != nil && student.IsActive != *filter.IsActive {
			continue
		}
		if filter.IsLinked != nil {
			isLinked := student.IsLinked()
			if isLinked != *filter.IsLinked {
				continue
			}
		}
		if filter.GroupName != "" && student.GroupName != filter.GroupName {
			continue
		}
		if filter.Faculty != "" && student.Faculty != filter.Faculty {
			continue
		}
		result = append(result, student)
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

func (m *MockExternalStudentRepository) GetUnlinked(_ context.Context, limit, offset int) ([]*entities.ExternalStudent, int64, error) {
	var result []*entities.ExternalStudent
	for _, student := range m.students {
		if !student.IsLinked() {
			result = append(result, student)
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

func (m *MockExternalStudentRepository) GetByGroup(_ context.Context, groupName string) ([]*entities.ExternalStudent, error) {
	var result []*entities.ExternalStudent
	for _, student := range m.students {
		if student.GroupName == groupName {
			result = append(result, student)
		}
	}
	return result, nil
}

func (m *MockExternalStudentRepository) GetByFaculty(_ context.Context, faculty string) ([]*entities.ExternalStudent, error) {
	var result []*entities.ExternalStudent
	for _, student := range m.students {
		if student.Faculty == faculty {
			result = append(result, student)
		}
	}
	return result, nil
}

func (m *MockExternalStudentRepository) LinkToLocalUser(_ context.Context, id int64, localUserID int64) error {
	if student, exists := m.students[id]; exists {
		student.LinkToLocalUser(localUserID)
		return nil
	}
	return errors.New("student not found")
}

func (m *MockExternalStudentRepository) Unlink(_ context.Context, id int64) error {
	if student, exists := m.students[id]; exists {
		student.Unlink()
		return nil
	}
	return errors.New("student not found")
}

func (m *MockExternalStudentRepository) Delete(_ context.Context, id int64) error {
	if student, exists := m.students[id]; exists {
		delete(m.studentsByEx, student.ExternalID)
		delete(m.students, id)
		return nil
	}
	return nil
}

func (m *MockExternalStudentRepository) GetAllExternalIDs(_ context.Context) ([]string, error) {
	var ids []string
	for _, student := range m.students {
		ids = append(ids, student.ExternalID)
	}
	return ids, nil
}

func (m *MockExternalStudentRepository) BulkUpsert(ctx context.Context, students []*entities.ExternalStudent) error {
	for _, student := range students {
		if err := m.Upsert(ctx, student); err != nil {
			return err
		}
	}
	return nil
}

func (m *MockExternalStudentRepository) MarkInactiveExcept(_ context.Context, activeExternalIDs []string) error {
	activeSet := make(map[string]bool)
	for _, id := range activeExternalIDs {
		activeSet[id] = true
	}
	for _, student := range m.students {
		if !activeSet[student.ExternalID] {
			student.IsActive = false
		}
	}
	return nil
}

func (m *MockExternalStudentRepository) GetGroups(_ context.Context) ([]string, error) {
	groupSet := make(map[string]bool)
	for _, student := range m.students {
		if student.GroupName != "" {
			groupSet[student.GroupName] = true
		}
	}
	var groups []string
	for group := range groupSet {
		groups = append(groups, group)
	}
	return groups, nil
}

func (m *MockExternalStudentRepository) GetFaculties(_ context.Context) ([]string, error) {
	facultySet := make(map[string]bool)
	for _, student := range m.students {
		if student.Faculty != "" {
			facultySet[student.Faculty] = true
		}
	}
	var faculties []string
	for faculty := range facultySet {
		faculties = append(faculties, faculty)
	}
	return faculties, nil
}

// Helper to create test student
func createTestStudent(repo *MockExternalStudentRepository, externalID, firstName, lastName, groupName, faculty string) *entities.ExternalStudent {
	student := entities.NewExternalStudent(externalID, "CODE-"+externalID)
	student.FirstName = firstName
	student.LastName = lastName
	student.GroupName = groupName
	student.Faculty = faculty
	_ = repo.Create(context.Background(), student)
	return student
}

// Tests

func TestStudentUseCase_List(t *testing.T) {
	repo := NewMockExternalStudentRepository()
	uc := NewStudentUseCase(repo)

	ctx := context.Background()

	// Create students
	createTestStudent(repo, "ext1", testFirstNameJohn, testLastNameDoe, "CS-101", "Computer Science")
	createTestStudent(repo, "ext2", "Jane", "Smith", "CS-102", "Computer Science")
	createTestStudent(repo, "ext3", "Bob", "Johnson", "MATH-101", "Mathematics")

	// List all
	req := &dto.ExternalStudentListRequest{
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

func TestStudentUseCase_List_WithGroupFilter(t *testing.T) {
	repo := NewMockExternalStudentRepository()
	uc := NewStudentUseCase(repo)

	ctx := context.Background()

	// Create students
	createTestStudent(repo, "ext1", testFirstNameJohn, testLastNameDoe, "CS-101", "Computer Science")
	createTestStudent(repo, "ext2", "Jane", "Smith", "CS-101", "Computer Science")
	createTestStudent(repo, "ext3", "Bob", "Johnson", "MATH-101", "Mathematics")

	// Filter by group
	req := &dto.ExternalStudentListRequest{
		GroupName: "CS-101",
		Limit:     10,
	}
	result, err := uc.List(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.Total != 2 {
		t.Errorf("expected total 2, got %d", result.Total)
	}
}

func TestStudentUseCase_GetByID(t *testing.T) {
	repo := NewMockExternalStudentRepository()
	uc := NewStudentUseCase(repo)

	ctx := context.Background()

	// Create student
	student := createTestStudent(repo, "ext1", testFirstNameJohn, testLastNameDoe, "CS-101", "Computer Science")

	// Get by ID
	result, err := uc.GetByID(ctx, student.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if result.FirstName != testFirstNameJohn {
		t.Errorf("expected first name 'John', got '%s'", result.FirstName)
	}
}

func TestStudentUseCase_GetByID_NotFound(t *testing.T) {
	repo := NewMockExternalStudentRepository()
	uc := NewStudentUseCase(repo)

	ctx := context.Background()

	// Get non-existent
	result, err := uc.GetByID(ctx, 999)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result != nil {
		t.Error("expected nil result for non-existent student")
	}
}

func TestStudentUseCase_GetByExternalID(t *testing.T) {
	repo := NewMockExternalStudentRepository()
	uc := NewStudentUseCase(repo)

	ctx := context.Background()

	// Create student
	createTestStudent(repo, "ext1", testFirstNameJohn, testLastNameDoe, "CS-101", "Computer Science")

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

func TestStudentUseCase_GetByExternalID_NotFound(t *testing.T) {
	repo := NewMockExternalStudentRepository()
	uc := NewStudentUseCase(repo)

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

func TestStudentUseCase_LinkToLocalUser(t *testing.T) {
	repo := NewMockExternalStudentRepository()
	uc := NewStudentUseCase(repo)

	ctx := context.Background()

	// Create student
	student := createTestStudent(repo, "ext1", testFirstNameJohn, testLastNameDoe, "CS-101", "Computer Science")

	// Link to local user
	err := uc.LinkToLocalUser(ctx, student.ID, 42)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify linked
	result, _ := repo.GetByID(ctx, student.ID)
	if !result.IsLinked() {
		t.Error("expected student to be linked")
	}
	if *result.LocalUserID != 42 {
		t.Errorf("expected local user ID 42, got %d", *result.LocalUserID)
	}
}

func TestStudentUseCase_LinkToLocalUser_NotFound(t *testing.T) {
	repo := NewMockExternalStudentRepository()
	uc := NewStudentUseCase(repo)

	ctx := context.Background()

	// Try to link non-existent student
	err := uc.LinkToLocalUser(ctx, 999, 42)
	if err == nil {
		t.Error("expected error for non-existent student")
	}
}

func TestStudentUseCase_LinkToLocalUser_AlreadyLinked(t *testing.T) {
	repo := NewMockExternalStudentRepository()
	uc := NewStudentUseCase(repo)

	ctx := context.Background()

	// Create and link student
	student := createTestStudent(repo, "ext1", testFirstNameJohn, testLastNameDoe, "CS-101", "Computer Science")
	_ = repo.LinkToLocalUser(ctx, student.ID, 42)

	// Try to link again
	err := uc.LinkToLocalUser(ctx, student.ID, 43)
	if err == nil {
		t.Error("expected error for already linked student")
	}
}

func TestStudentUseCase_LinkToLocalUser_LocalUserAlreadyLinked(t *testing.T) {
	repo := NewMockExternalStudentRepository()
	uc := NewStudentUseCase(repo)

	ctx := context.Background()

	// Create and link first student
	student1 := createTestStudent(repo, "ext1", testFirstNameJohn, testLastNameDoe, "CS-101", "Computer Science")
	_ = repo.LinkToLocalUser(ctx, student1.ID, 42)

	// Create second student
	student2 := createTestStudent(repo, "ext2", "Jane", "Smith", "CS-102", "Computer Science")

	// Try to link second student to same local user
	err := uc.LinkToLocalUser(ctx, student2.ID, 42)
	if err == nil {
		t.Error("expected error when local user is already linked to another student")
	}
}

func TestStudentUseCase_Unlink(t *testing.T) {
	repo := NewMockExternalStudentRepository()
	uc := NewStudentUseCase(repo)

	ctx := context.Background()

	// Create and link student
	student := createTestStudent(repo, "ext1", testFirstNameJohn, testLastNameDoe, "CS-101", "Computer Science")
	_ = repo.LinkToLocalUser(ctx, student.ID, 42)

	// Unlink
	err := uc.Unlink(ctx, student.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify unlinked
	result, _ := repo.GetByID(ctx, student.ID)
	if result.IsLinked() {
		t.Error("expected student to be unlinked")
	}
}

func TestStudentUseCase_Unlink_NotFound(t *testing.T) {
	repo := NewMockExternalStudentRepository()
	uc := NewStudentUseCase(repo)

	ctx := context.Background()

	// Try to unlink non-existent student
	err := uc.Unlink(ctx, 999)
	if err == nil {
		t.Error("expected error for non-existent student")
	}
}

func TestStudentUseCase_Unlink_NotLinked(t *testing.T) {
	repo := NewMockExternalStudentRepository()
	uc := NewStudentUseCase(repo)

	ctx := context.Background()

	// Create unlinked student
	student := createTestStudent(repo, "ext1", testFirstNameJohn, testLastNameDoe, "CS-101", "Computer Science")

	// Try to unlink
	err := uc.Unlink(ctx, student.ID)
	if err == nil {
		t.Error("expected error for unlinked student")
	}
}

func TestStudentUseCase_GetUnlinked(t *testing.T) {
	repo := NewMockExternalStudentRepository()
	uc := NewStudentUseCase(repo)

	ctx := context.Background()

	// Create students
	student1 := createTestStudent(repo, "ext1", testFirstNameJohn, testLastNameDoe, "CS-101", "Computer Science")
	createTestStudent(repo, "ext2", "Jane", "Smith", "CS-102", "Computer Science")
	createTestStudent(repo, "ext3", "Bob", "Johnson", "MATH-101", "Mathematics")

	// Link first student
	_ = repo.LinkToLocalUser(ctx, student1.ID, 42)

	// Get unlinked
	result, err := uc.GetUnlinked(ctx, 10, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.Total != 2 {
		t.Errorf("expected total 2 unlinked, got %d", result.Total)
	}
}

func TestStudentUseCase_GetByGroup(t *testing.T) {
	repo := NewMockExternalStudentRepository()
	uc := NewStudentUseCase(repo)

	ctx := context.Background()

	// Create students
	createTestStudent(repo, "ext1", testFirstNameJohn, testLastNameDoe, "CS-101", "Computer Science")
	createTestStudent(repo, "ext2", "Jane", "Smith", "CS-101", "Computer Science")
	createTestStudent(repo, "ext3", "Bob", "Johnson", "MATH-101", "Mathematics")

	// Get by group
	result, err := uc.GetByGroup(ctx, "CS-101")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 students, got %d", len(result))
	}
}

func TestStudentUseCase_GetByFaculty(t *testing.T) {
	repo := NewMockExternalStudentRepository()
	uc := NewStudentUseCase(repo)

	ctx := context.Background()

	// Create students
	createTestStudent(repo, "ext1", testFirstNameJohn, testLastNameDoe, "CS-101", "Computer Science")
	createTestStudent(repo, "ext2", "Jane", "Smith", "CS-102", "Computer Science")
	createTestStudent(repo, "ext3", "Bob", "Johnson", "MATH-101", "Mathematics")

	// Get by faculty
	result, err := uc.GetByFaculty(ctx, "Computer Science")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 students, got %d", len(result))
	}
}

func TestStudentUseCase_GetGroups(t *testing.T) {
	repo := NewMockExternalStudentRepository()
	uc := NewStudentUseCase(repo)

	ctx := context.Background()

	// Create students
	createTestStudent(repo, "ext1", testFirstNameJohn, testLastNameDoe, "CS-101", "Computer Science")
	createTestStudent(repo, "ext2", "Jane", "Smith", "CS-101", "Computer Science")
	createTestStudent(repo, "ext3", "Bob", "Johnson", "MATH-101", "Mathematics")

	// Get groups
	result, err := uc.GetGroups(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(result.Groups) != 2 {
		t.Errorf("expected 2 distinct groups, got %d", len(result.Groups))
	}
}

func TestStudentUseCase_GetFaculties(t *testing.T) {
	repo := NewMockExternalStudentRepository()
	uc := NewStudentUseCase(repo)

	ctx := context.Background()

	// Create students
	createTestStudent(repo, "ext1", testFirstNameJohn, testLastNameDoe, "CS-101", "Computer Science")
	createTestStudent(repo, "ext2", "Jane", "Smith", "CS-102", "Computer Science")
	createTestStudent(repo, "ext3", "Bob", "Johnson", "MATH-101", "Mathematics")

	// Get faculties
	result, err := uc.GetFaculties(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(result.Faculties) != 2 {
		t.Errorf("expected 2 distinct faculties, got %d", len(result.Faculties))
	}
}

func TestStudentUseCase_Delete(t *testing.T) {
	repo := NewMockExternalStudentRepository()
	uc := NewStudentUseCase(repo)

	ctx := context.Background()

	// Create student
	student := createTestStudent(repo, "ext1", testFirstNameJohn, testLastNameDoe, "CS-101", "Computer Science")

	// Delete
	err := uc.Delete(ctx, student.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify deleted
	result, _ := repo.GetByID(ctx, student.ID)
	if result != nil {
		t.Error("expected student to be deleted")
	}
}

// --- Error-injecting mock for student repository ---

type errorStudentRepo struct {
	MockExternalStudentRepository
	listErr          bool
	getByIDErr       bool
	getByExternalErr bool
	getByLocalErr    bool
	getUnlinkedErr   bool
	getByGroupErr    bool
	getByFacultyErr  bool
	getGroupsErr     bool
	getFacultiesErr  bool
	linkErr          bool
	unlinkErr        bool
	deleteErr        bool
}

func newErrorStudentRepo() *errorStudentRepo {
	return &errorStudentRepo{
		MockExternalStudentRepository: *NewMockExternalStudentRepository(),
	}
}

func (m *errorStudentRepo) List(_ context.Context, _ entities.ExternalStudentFilter) ([]*entities.ExternalStudent, int64, error) {
	if m.listErr {
		return nil, 0, errors.New("list error")
	}
	return m.MockExternalStudentRepository.List(context.Background(), entities.ExternalStudentFilter{})
}

func (m *errorStudentRepo) GetByID(_ context.Context, id int64) (*entities.ExternalStudent, error) {
	if m.getByIDErr {
		return nil, errors.New("get by ID error")
	}
	return m.MockExternalStudentRepository.GetByID(context.Background(), id)
}

func (m *errorStudentRepo) GetByExternalID(_ context.Context, externalID string) (*entities.ExternalStudent, error) {
	if m.getByExternalErr {
		return nil, errors.New("get by external ID error")
	}
	return m.MockExternalStudentRepository.GetByExternalID(context.Background(), externalID)
}

func (m *errorStudentRepo) GetByLocalUserID(_ context.Context, localUserID int64) (*entities.ExternalStudent, error) {
	if m.getByLocalErr {
		return nil, errors.New("get by local user error")
	}
	return m.MockExternalStudentRepository.GetByLocalUserID(context.Background(), localUserID)
}

func (m *errorStudentRepo) GetUnlinked(_ context.Context, limit, offset int) ([]*entities.ExternalStudent, int64, error) {
	if m.getUnlinkedErr {
		return nil, 0, errors.New("get unlinked error")
	}
	return m.MockExternalStudentRepository.GetUnlinked(context.Background(), limit, offset)
}

func (m *errorStudentRepo) GetByGroup(_ context.Context, groupName string) ([]*entities.ExternalStudent, error) {
	if m.getByGroupErr {
		return nil, errors.New("get by group error")
	}
	return m.MockExternalStudentRepository.GetByGroup(context.Background(), groupName)
}

func (m *errorStudentRepo) GetByFaculty(_ context.Context, faculty string) ([]*entities.ExternalStudent, error) {
	if m.getByFacultyErr {
		return nil, errors.New("get by faculty error")
	}
	return m.MockExternalStudentRepository.GetByFaculty(context.Background(), faculty)
}

func (m *errorStudentRepo) GetGroups(_ context.Context) ([]string, error) {
	if m.getGroupsErr {
		return nil, errors.New("get groups error")
	}
	return m.MockExternalStudentRepository.GetGroups(context.Background())
}

func (m *errorStudentRepo) GetFaculties(_ context.Context) ([]string, error) {
	if m.getFacultiesErr {
		return nil, errors.New("get faculties error")
	}
	return m.MockExternalStudentRepository.GetFaculties(context.Background())
}

func (m *errorStudentRepo) LinkToLocalUser(_ context.Context, id int64, localUserID int64) error {
	if m.linkErr {
		return errors.New("link error")
	}
	return m.MockExternalStudentRepository.LinkToLocalUser(context.Background(), id, localUserID)
}

func (m *errorStudentRepo) Unlink(_ context.Context, id int64) error {
	if m.unlinkErr {
		return errors.New("unlink error")
	}
	return m.MockExternalStudentRepository.Unlink(context.Background(), id)
}

func (m *errorStudentRepo) Delete(_ context.Context, id int64) error {
	if m.deleteErr {
		return errors.New("delete error")
	}
	return m.MockExternalStudentRepository.Delete(context.Background(), id)
}

// --- Error path tests ---

func TestStudentUseCase_List_Error(t *testing.T) {
	repo := newErrorStudentRepo()
	repo.listErr = true
	uc := NewStudentUseCase(repo)

	req := &dto.ExternalStudentListRequest{Limit: 10}
	result, err := uc.List(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to list students")
}

func TestStudentUseCase_GetByID_Error(t *testing.T) {
	repo := newErrorStudentRepo()
	repo.getByIDErr = true
	uc := NewStudentUseCase(repo)

	result, err := uc.GetByID(context.Background(), 1)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get student")
}

func TestStudentUseCase_GetByExternalID_Error(t *testing.T) {
	repo := newErrorStudentRepo()
	repo.getByExternalErr = true
	uc := NewStudentUseCase(repo)

	result, err := uc.GetByExternalID(context.Background(), "ext1")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get student by external ID")
}

func TestStudentUseCase_LinkToLocalUser_GetByIDError(t *testing.T) {
	repo := newErrorStudentRepo()
	repo.getByIDErr = true
	uc := NewStudentUseCase(repo)

	err := uc.LinkToLocalUser(context.Background(), 1, 42)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get student")
}

func TestStudentUseCase_LinkToLocalUser_GetByLocalUserError(t *testing.T) {
	repo := newErrorStudentRepo()
	uc := NewStudentUseCase(repo)
	ctx := context.Background()

	student := entities.NewExternalStudent("ext1", "CODE-ext1")
	student.FirstName = testFirstNameJohn
	student.LastName = testLastNameDoe
	_ = repo.Create(ctx, student)

	repo.getByLocalErr = true

	err := uc.LinkToLocalUser(ctx, student.ID, 42)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check existing link")
}

func TestStudentUseCase_LinkToLocalUser_LinkRepoError(t *testing.T) {
	repo := newErrorStudentRepo()
	uc := NewStudentUseCase(repo)
	ctx := context.Background()

	student := entities.NewExternalStudent("ext1", "CODE-ext1")
	student.FirstName = testFirstNameJohn
	student.LastName = testLastNameDoe
	_ = repo.Create(ctx, student)

	repo.linkErr = true

	err := uc.LinkToLocalUser(ctx, student.ID, 42)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to link student")
}

func TestStudentUseCase_Unlink_GetByIDError(t *testing.T) {
	repo := newErrorStudentRepo()
	repo.getByIDErr = true
	uc := NewStudentUseCase(repo)

	err := uc.Unlink(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get student")
}

func TestStudentUseCase_Unlink_UnlinkRepoError(t *testing.T) {
	repo := newErrorStudentRepo()
	uc := NewStudentUseCase(repo)
	ctx := context.Background()

	student := entities.NewExternalStudent("ext1", "CODE-ext1")
	_ = repo.Create(ctx, student)
	_ = repo.LinkToLocalUser(ctx, student.ID, 42)

	repo.unlinkErr = true

	err := uc.Unlink(ctx, student.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unlink student")
}

func TestStudentUseCase_GetUnlinked_Error(t *testing.T) {
	repo := newErrorStudentRepo()
	repo.getUnlinkedErr = true
	uc := NewStudentUseCase(repo)

	result, err := uc.GetUnlinked(context.Background(), 10, 0)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get unlinked students")
}

func TestStudentUseCase_GetByGroup_Error(t *testing.T) {
	repo := newErrorStudentRepo()
	repo.getByGroupErr = true
	uc := NewStudentUseCase(repo)

	result, err := uc.GetByGroup(context.Background(), "CS-101")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get students by group")
}

func TestStudentUseCase_GetByFaculty_Error(t *testing.T) {
	repo := newErrorStudentRepo()
	repo.getByFacultyErr = true
	uc := NewStudentUseCase(repo)

	result, err := uc.GetByFaculty(context.Background(), "CS")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get students by faculty")
}

func TestStudentUseCase_GetGroups_Error(t *testing.T) {
	repo := newErrorStudentRepo()
	repo.getGroupsErr = true
	uc := NewStudentUseCase(repo)

	result, err := uc.GetGroups(context.Background())
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get groups")
}

func TestStudentUseCase_GetFaculties_Error(t *testing.T) {
	repo := newErrorStudentRepo()
	repo.getFacultiesErr = true
	uc := NewStudentUseCase(repo)

	result, err := uc.GetFaculties(context.Background())
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get faculties")
}

func TestStudentUseCase_Delete_Error(t *testing.T) {
	repo := newErrorStudentRepo()
	repo.deleteErr = true
	uc := NewStudentUseCase(repo)

	err := uc.Delete(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete student")
}
