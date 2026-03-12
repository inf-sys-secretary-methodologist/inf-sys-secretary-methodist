package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	authDomain "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
	authEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/domain/repositories"
)

const testEmail = "test@example.com"

// MockUserRepository implements authRepos.UserRepository for testing.
type MockUserRepository struct {
	users  map[int64]*authEntities.User
	nextID int64
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:  make(map[int64]*authEntities.User),
		nextID: 1,
	}
}

func (m *MockUserRepository) Create(_ context.Context, user *authEntities.User) error {
	user.ID = m.nextID
	m.nextID++
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) Save(_ context.Context, user *authEntities.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) GetByID(_ context.Context, id int64) (*authEntities.User, error) {
	user, exists := m.users[id]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (m *MockUserRepository) GetByEmail(_ context.Context, email string) (*authEntities.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, errors.New("user not found")
}

func (m *MockUserRepository) GetByEmailForAuth(_ context.Context, email string) (*authEntities.User, error) {
	return m.GetByEmail(context.Background(), email)
}

func (m *MockUserRepository) Delete(_ context.Context, id int64) error {
	delete(m.users, id)
	return nil
}

func (m *MockUserRepository) List(_ context.Context, limit, offset int) ([]*authEntities.User, error) {
	var users []*authEntities.User
	i := 0
	for _, user := range m.users {
		if i >= offset && len(users) < limit {
			users = append(users, user)
		}
		i++
	}
	return users, nil
}

// MockUserProfileRepository implements UserProfileRepository for testing.
type MockUserProfileRepository struct {
	profiles map[int64]*entities.UserWithOrg
	nextID   int64
}

func NewMockUserProfileRepository() *MockUserProfileRepository {
	return &MockUserProfileRepository{
		profiles: make(map[int64]*entities.UserWithOrg),
		nextID:   1,
	}
}

func (m *MockUserProfileRepository) GetProfileByID(_ context.Context, userID int64) (*entities.UserWithOrg, error) {
	profile, exists := m.profiles[userID]
	if !exists {
		return nil, errors.New("profile not found")
	}
	return profile, nil
}

func (m *MockUserProfileRepository) UpdateProfile(_ context.Context, userID int64, departmentID, positionID *int64, phone, avatar, bio string) error {
	profile, exists := m.profiles[userID]
	if !exists {
		// Create new profile
		profile = &entities.UserWithOrg{ID: userID}
		m.profiles[userID] = profile
	}
	profile.DepartmentID = departmentID
	profile.PositionID = positionID
	profile.Phone = phone
	profile.Avatar = avatar
	profile.Bio = bio
	return nil
}

func (m *MockUserProfileRepository) ListUsersWithOrg(_ context.Context, _ *repositories.UserFilter, limit, offset int) ([]*entities.UserWithOrg, error) {
	var profiles []*entities.UserWithOrg
	i := 0
	for _, profile := range m.profiles {
		if i >= offset && len(profiles) < limit {
			profiles = append(profiles, profile)
		}
		i++
	}
	return profiles, nil
}

func (m *MockUserProfileRepository) CountUsers(_ context.Context, _ *repositories.UserFilter) (int64, error) {
	return int64(len(m.profiles)), nil
}

func (m *MockUserProfileRepository) GetUsersByDepartment(_ context.Context, departmentID int64) ([]*entities.UserWithOrg, error) {
	var users []*entities.UserWithOrg
	for _, profile := range m.profiles {
		if profile.DepartmentID != nil && *profile.DepartmentID == departmentID {
			users = append(users, profile)
		}
	}
	return users, nil
}

func (m *MockUserProfileRepository) GetUsersByPosition(_ context.Context, positionID int64) ([]*entities.UserWithOrg, error) {
	var users []*entities.UserWithOrg
	for _, profile := range m.profiles {
		if profile.PositionID != nil && *profile.PositionID == positionID {
			users = append(users, profile)
		}
	}
	return users, nil
}

func (m *MockUserProfileRepository) BulkUpdateDepartment(_ context.Context, userIDs []int64, departmentID *int64) error {
	for _, id := range userIDs {
		if profile, exists := m.profiles[id]; exists {
			profile.DepartmentID = departmentID
		}
	}
	return nil
}

func (m *MockUserProfileRepository) BulkUpdatePosition(_ context.Context, userIDs []int64, positionID *int64) error {
	for _, id := range userIDs {
		if profile, exists := m.profiles[id]; exists {
			profile.PositionID = positionID
		}
	}
	return nil
}

// AddProfile helper for tests
func (m *MockUserProfileRepository) AddProfile(profile *entities.UserWithOrg) {
	m.profiles[profile.ID] = profile
}

// Tests

func TestUserUseCase_GetUser(t *testing.T) {
	userRepo := NewMockUserRepository()
	profileRepo := NewMockUserProfileRepository()
	deptRepo := NewMockDepartmentRepository()
	posRepo := NewMockPositionRepository()
	uc := NewUserUseCase(userRepo, profileRepo, deptRepo, posRepo, nil, nil)

	ctx := context.Background()

	// Add profile
	profileRepo.AddProfile(&entities.UserWithOrg{
		ID:    1,
		Email: testEmail,
		Name:  "Test User",
	})

	// Get user
	user, err := uc.GetUser(ctx, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if user.Email != testEmail {
		t.Errorf("expected email 'test@example.com', got '%s'", user.Email)
	}
}

func TestUserUseCase_GetUser_NotFound(t *testing.T) {
	userRepo := NewMockUserRepository()
	profileRepo := NewMockUserProfileRepository()
	deptRepo := NewMockDepartmentRepository()
	posRepo := NewMockPositionRepository()
	uc := NewUserUseCase(userRepo, profileRepo, deptRepo, posRepo, nil, nil)

	ctx := context.Background()

	_, err := uc.GetUser(ctx, 999)
	if err == nil {
		t.Error("expected error for non-existent user")
	}
}

func TestUserUseCase_ListUsers(t *testing.T) {
	userRepo := NewMockUserRepository()
	profileRepo := NewMockUserProfileRepository()
	deptRepo := NewMockDepartmentRepository()
	posRepo := NewMockPositionRepository()
	uc := NewUserUseCase(userRepo, profileRepo, deptRepo, posRepo, nil, nil)

	ctx := context.Background()

	// Add profiles
	profileRepo.AddProfile(&entities.UserWithOrg{ID: 1, Name: "User 1"})
	profileRepo.AddProfile(&entities.UserWithOrg{ID: 2, Name: "User 2"})
	profileRepo.AddProfile(&entities.UserWithOrg{ID: 3, Name: "User 3"})

	// List users
	filter := &dto.UserListFilter{Page: 1, Limit: 10}
	result, err := uc.ListUsers(ctx, filter)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.Total != 3 {
		t.Errorf("expected total 3, got %d", result.Total)
	}
}

func TestUserUseCase_ListUsers_Pagination(t *testing.T) {
	userRepo := NewMockUserRepository()
	profileRepo := NewMockUserProfileRepository()
	deptRepo := NewMockDepartmentRepository()
	posRepo := NewMockPositionRepository()
	uc := NewUserUseCase(userRepo, profileRepo, deptRepo, posRepo, nil, nil)

	ctx := context.Background()

	// Default page
	filter := &dto.UserListFilter{Page: 0, Limit: 10}
	result, _ := uc.ListUsers(ctx, filter)
	if result.Page != 1 {
		t.Errorf("expected default page 1, got %d", result.Page)
	}

	// Default limit
	filter = &dto.UserListFilter{Page: 1, Limit: 0}
	result, _ = uc.ListUsers(ctx, filter)
	if result.Limit != 10 {
		t.Errorf("expected default limit 10, got %d", result.Limit)
	}

	// Max limit
	filter = &dto.UserListFilter{Page: 1, Limit: 200}
	result, _ = uc.ListUsers(ctx, filter)
	if result.Limit != 100 {
		t.Errorf("expected max limit 100, got %d", result.Limit)
	}
}

func TestUserUseCase_ListUsers_TotalPages(t *testing.T) {
	userRepo := NewMockUserRepository()
	profileRepo := NewMockUserProfileRepository()
	deptRepo := NewMockDepartmentRepository()
	posRepo := NewMockPositionRepository()
	uc := NewUserUseCase(userRepo, profileRepo, deptRepo, posRepo, nil, nil)

	ctx := context.Background()

	// Add 5 profiles
	for i := int64(1); i <= 5; i++ {
		profileRepo.AddProfile(&entities.UserWithOrg{ID: i, Name: "User"})
	}

	// List with limit 2 (should be 3 pages)
	filter := &dto.UserListFilter{Page: 1, Limit: 2}
	result, _ := uc.ListUsers(ctx, filter)
	if result.TotalPages != 3 {
		t.Errorf("expected 3 total pages, got %d", result.TotalPages)
	}
}

func TestUserUseCase_UpdateUserProfile(t *testing.T) {
	userRepo := NewMockUserRepository()
	profileRepo := NewMockUserProfileRepository()
	deptRepo := NewMockDepartmentRepository()
	posRepo := NewMockPositionRepository()
	uc := NewUserUseCase(userRepo, profileRepo, deptRepo, posRepo, nil, nil)

	ctx := context.Background()

	// Create user
	user := authEntities.NewUser(testEmail, "password", "Test User", authDomain.RoleStudent)
	_ = userRepo.Create(ctx, user)

	// Create department and position
	_ = deptRepo.Create(ctx, entities.NewDepartment("IT", "IT", "", nil))
	_ = posRepo.Create(ctx, entities.NewPosition("Dev", "DEV", "", 1))

	// Get IDs (from mock implementation)
	var deptID, posID int64 = 1, 1 // Mock starts at 1

	// Update profile
	input := &dto.UpdateUserProfileInput{
		DepartmentID: &deptID,
		PositionID:   &posID,
		Phone:        "+1234567890",
		Bio:          "Test bio",
	}

	err := uc.UpdateUserProfile(ctx, user.ID, input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestUserUseCase_UpdateUserProfile_UserNotFound(t *testing.T) {
	userRepo := NewMockUserRepository()
	profileRepo := NewMockUserProfileRepository()
	deptRepo := NewMockDepartmentRepository()
	posRepo := NewMockPositionRepository()
	uc := NewUserUseCase(userRepo, profileRepo, deptRepo, posRepo, nil, nil)

	ctx := context.Background()

	input := &dto.UpdateUserProfileInput{Phone: "123"}
	err := uc.UpdateUserProfile(ctx, 999, input)
	if err == nil {
		t.Error("expected error for non-existent user")
	}
}

func TestUserUseCase_UpdateUserProfile_DepartmentNotFound(t *testing.T) {
	userRepo := NewMockUserRepository()
	profileRepo := NewMockUserProfileRepository()
	deptRepo := NewMockDepartmentRepository()
	posRepo := NewMockPositionRepository()
	uc := NewUserUseCase(userRepo, profileRepo, deptRepo, posRepo, nil, nil)

	ctx := context.Background()

	// Create user
	user := authEntities.NewUser(testEmail, "password", "Test User", authDomain.RoleStudent)
	_ = userRepo.Create(ctx, user)

	// Try to update with non-existent department
	deptID := int64(999)
	input := &dto.UpdateUserProfileInput{DepartmentID: &deptID}

	err := uc.UpdateUserProfile(ctx, user.ID, input)
	if err == nil {
		t.Error("expected error for non-existent department")
	}
}

func TestUserUseCase_UpdateUserProfile_PositionNotFound(t *testing.T) {
	userRepo := NewMockUserRepository()
	profileRepo := NewMockUserProfileRepository()
	deptRepo := NewMockDepartmentRepository()
	posRepo := NewMockPositionRepository()
	uc := NewUserUseCase(userRepo, profileRepo, deptRepo, posRepo, nil, nil)

	ctx := context.Background()

	// Create user
	user := authEntities.NewUser(testEmail, "password", "Test User", authDomain.RoleStudent)
	_ = userRepo.Create(ctx, user)

	// Try to update with non-existent position
	posID := int64(999)
	input := &dto.UpdateUserProfileInput{PositionID: &posID}

	err := uc.UpdateUserProfile(ctx, user.ID, input)
	if err == nil {
		t.Error("expected error for non-existent position")
	}
}

func TestUserUseCase_UpdateUserRole(t *testing.T) {
	userRepo := NewMockUserRepository()
	profileRepo := NewMockUserProfileRepository()
	deptRepo := NewMockDepartmentRepository()
	posRepo := NewMockPositionRepository()
	uc := NewUserUseCase(userRepo, profileRepo, deptRepo, posRepo, nil, nil)

	ctx := context.Background()

	// Create user
	user := authEntities.NewUser(testEmail, "password", "Test User", authDomain.RoleStudent)
	user.Role = authDomain.RoleStudent
	_ = userRepo.Create(ctx, user)

	// Update role
	input := &dto.UpdateUserRoleInput{Role: "teacher"}
	err := uc.UpdateUserRole(ctx, user.ID, input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify role updated
	updated, _ := userRepo.GetByID(ctx, user.ID)
	if updated.Role != authDomain.RoleTeacher {
		t.Errorf("expected role 'teacher', got '%s'", updated.Role)
	}
}

func TestUserUseCase_UpdateUserRole_NotFound(t *testing.T) {
	userRepo := NewMockUserRepository()
	profileRepo := NewMockUserProfileRepository()
	deptRepo := NewMockDepartmentRepository()
	posRepo := NewMockPositionRepository()
	uc := NewUserUseCase(userRepo, profileRepo, deptRepo, posRepo, nil, nil)

	ctx := context.Background()

	input := &dto.UpdateUserRoleInput{Role: "teacher"}
	err := uc.UpdateUserRole(ctx, 999, input)
	if err == nil {
		t.Error("expected error for non-existent user")
	}
}

func TestUserUseCase_UpdateUserStatus(t *testing.T) {
	userRepo := NewMockUserRepository()
	profileRepo := NewMockUserProfileRepository()
	deptRepo := NewMockDepartmentRepository()
	posRepo := NewMockPositionRepository()
	uc := NewUserUseCase(userRepo, profileRepo, deptRepo, posRepo, nil, nil)

	ctx := context.Background()

	// Create user
	user := authEntities.NewUser(testEmail, "password", "Test User", authDomain.RoleStudent)
	_ = userRepo.Create(ctx, user)

	// Test activate
	err := uc.UpdateUserStatus(ctx, user.ID, &dto.UpdateUserStatusInput{Status: "active"})
	if err != nil {
		t.Fatalf("activate error: %v", err)
	}

	updated, _ := userRepo.GetByID(ctx, user.ID)
	if updated.Status != authEntities.UserStatusActive {
		t.Errorf("expected status 'active', got '%s'", updated.Status)
	}

	// Test deactivate
	err = uc.UpdateUserStatus(ctx, user.ID, &dto.UpdateUserStatusInput{Status: "inactive"})
	if err != nil {
		t.Fatalf("deactivate error: %v", err)
	}

	updated, _ = userRepo.GetByID(ctx, user.ID)
	if updated.Status != authEntities.UserStatusInactive {
		t.Errorf("expected status 'inactive', got '%s'", updated.Status)
	}

	// Test block
	err = uc.UpdateUserStatus(ctx, user.ID, &dto.UpdateUserStatusInput{Status: "blocked"})
	if err != nil {
		t.Fatalf("block error: %v", err)
	}

	updated, _ = userRepo.GetByID(ctx, user.ID)
	if updated.Status != authEntities.UserStatusBlocked {
		t.Errorf("expected status 'blocked', got '%s'", updated.Status)
	}
}

func TestUserUseCase_UpdateUserStatus_NotFound(t *testing.T) {
	userRepo := NewMockUserRepository()
	profileRepo := NewMockUserProfileRepository()
	deptRepo := NewMockDepartmentRepository()
	posRepo := NewMockPositionRepository()
	uc := NewUserUseCase(userRepo, profileRepo, deptRepo, posRepo, nil, nil)

	ctx := context.Background()

	err := uc.UpdateUserStatus(ctx, 999, &dto.UpdateUserStatusInput{Status: "active"})
	if err == nil {
		t.Error("expected error for non-existent user")
	}
}

func TestUserUseCase_DeleteUser(t *testing.T) {
	userRepo := NewMockUserRepository()
	profileRepo := NewMockUserProfileRepository()
	deptRepo := NewMockDepartmentRepository()
	posRepo := NewMockPositionRepository()
	uc := NewUserUseCase(userRepo, profileRepo, deptRepo, posRepo, nil, nil)

	ctx := context.Background()

	// Create user
	user := authEntities.NewUser(testEmail, "password", "Test User", authDomain.RoleStudent)
	_ = userRepo.Create(ctx, user)

	// Delete
	err := uc.DeleteUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify deleted
	_, err = userRepo.GetByID(ctx, user.ID)
	if err == nil {
		t.Error("expected error for deleted user")
	}
}

func TestUserUseCase_DeleteUser_NotFound(t *testing.T) {
	userRepo := NewMockUserRepository()
	profileRepo := NewMockUserProfileRepository()
	deptRepo := NewMockDepartmentRepository()
	posRepo := NewMockPositionRepository()
	uc := NewUserUseCase(userRepo, profileRepo, deptRepo, posRepo, nil, nil)

	ctx := context.Background()

	err := uc.DeleteUser(ctx, 999)
	if err == nil {
		t.Error("expected error for non-existent user")
	}
}

func TestUserUseCase_BulkUpdateDepartment(t *testing.T) {
	userRepo := NewMockUserRepository()
	profileRepo := NewMockUserProfileRepository()
	deptRepo := NewMockDepartmentRepository()
	posRepo := NewMockPositionRepository()
	uc := NewUserUseCase(userRepo, profileRepo, deptRepo, posRepo, nil, nil)

	ctx := context.Background()

	// Create department
	_ = deptRepo.Create(ctx, entities.NewDepartment("IT", "IT", "", nil))

	// Add profiles
	profileRepo.AddProfile(&entities.UserWithOrg{ID: 1, Name: "User 1"})
	profileRepo.AddProfile(&entities.UserWithOrg{ID: 2, Name: "User 2"})

	// Bulk update
	deptID := int64(1)
	input := &dto.BulkUpdateDepartmentInput{
		UserIDs:      []int64{1, 2},
		DepartmentID: &deptID,
	}

	err := uc.BulkUpdateDepartment(ctx, input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify updates
	users, _ := uc.GetUsersByDepartment(ctx, 1)
	if len(users) != 2 {
		t.Errorf("expected 2 users in department, got %d", len(users))
	}
}

func TestUserUseCase_BulkUpdateDepartment_NotFound(t *testing.T) {
	userRepo := NewMockUserRepository()
	profileRepo := NewMockUserProfileRepository()
	deptRepo := NewMockDepartmentRepository()
	posRepo := NewMockPositionRepository()
	uc := NewUserUseCase(userRepo, profileRepo, deptRepo, posRepo, nil, nil)

	ctx := context.Background()

	deptID := int64(999)
	input := &dto.BulkUpdateDepartmentInput{
		UserIDs:      []int64{1, 2},
		DepartmentID: &deptID,
	}

	err := uc.BulkUpdateDepartment(ctx, input)
	if err == nil {
		t.Error("expected error for non-existent department")
	}
}

func TestUserUseCase_BulkUpdateDepartment_NilDepartment(t *testing.T) {
	userRepo := NewMockUserRepository()
	profileRepo := NewMockUserProfileRepository()
	deptRepo := NewMockDepartmentRepository()
	posRepo := NewMockPositionRepository()
	uc := NewUserUseCase(userRepo, profileRepo, deptRepo, posRepo, nil, nil)

	ctx := context.Background()

	// Add profiles with department
	deptID := int64(1)
	profileRepo.AddProfile(&entities.UserWithOrg{ID: 1, Name: "User 1", DepartmentID: &deptID})

	// Remove department (nil)
	input := &dto.BulkUpdateDepartmentInput{
		UserIDs:      []int64{1},
		DepartmentID: nil,
	}

	err := uc.BulkUpdateDepartment(ctx, input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify removed
	profile, _ := profileRepo.GetProfileByID(ctx, 1)
	if profile.DepartmentID != nil {
		t.Error("expected department ID to be nil")
	}
}

func TestUserUseCase_BulkUpdatePosition(t *testing.T) {
	userRepo := NewMockUserRepository()
	profileRepo := NewMockUserProfileRepository()
	deptRepo := NewMockDepartmentRepository()
	posRepo := NewMockPositionRepository()
	uc := NewUserUseCase(userRepo, profileRepo, deptRepo, posRepo, nil, nil)

	ctx := context.Background()

	// Create position
	_ = posRepo.Create(ctx, entities.NewPosition("Dev", "DEV", "", 1))

	// Add profiles
	profileRepo.AddProfile(&entities.UserWithOrg{ID: 1, Name: "User 1"})
	profileRepo.AddProfile(&entities.UserWithOrg{ID: 2, Name: "User 2"})

	// Bulk update
	posID := int64(1)
	input := &dto.BulkUpdatePositionInput{
		UserIDs:    []int64{1, 2},
		PositionID: &posID,
	}

	err := uc.BulkUpdatePosition(ctx, input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify updates
	users, _ := uc.GetUsersByPosition(ctx, 1)
	if len(users) != 2 {
		t.Errorf("expected 2 users in position, got %d", len(users))
	}
}

func TestUserUseCase_BulkUpdatePosition_NotFound(t *testing.T) {
	userRepo := NewMockUserRepository()
	profileRepo := NewMockUserProfileRepository()
	deptRepo := NewMockDepartmentRepository()
	posRepo := NewMockPositionRepository()
	uc := NewUserUseCase(userRepo, profileRepo, deptRepo, posRepo, nil, nil)

	ctx := context.Background()

	posID := int64(999)
	input := &dto.BulkUpdatePositionInput{
		UserIDs:    []int64{1, 2},
		PositionID: &posID,
	}

	err := uc.BulkUpdatePosition(ctx, input)
	if err == nil {
		t.Error("expected error for non-existent position")
	}
}

func TestUserUseCase_GetUsersByDepartment(t *testing.T) {
	userRepo := NewMockUserRepository()
	profileRepo := NewMockUserProfileRepository()
	deptRepo := NewMockDepartmentRepository()
	posRepo := NewMockPositionRepository()
	uc := NewUserUseCase(userRepo, profileRepo, deptRepo, posRepo, nil, nil)

	ctx := context.Background()

	// Add profiles
	deptID := int64(1)
	profileRepo.AddProfile(&entities.UserWithOrg{ID: 1, Name: "User 1", DepartmentID: &deptID})
	profileRepo.AddProfile(&entities.UserWithOrg{ID: 2, Name: "User 2", DepartmentID: &deptID})
	profileRepo.AddProfile(&entities.UserWithOrg{ID: 3, Name: "User 3"}) // No department

	// Get by department
	users, err := uc.GetUsersByDepartment(ctx, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(users) != 2 {
		t.Errorf("expected 2 users, got %d", len(users))
	}
}

func TestUserUseCase_GetUsersByPosition(t *testing.T) {
	userRepo := NewMockUserRepository()
	profileRepo := NewMockUserProfileRepository()
	deptRepo := NewMockDepartmentRepository()
	posRepo := NewMockPositionRepository()
	uc := NewUserUseCase(userRepo, profileRepo, deptRepo, posRepo, nil, nil)

	ctx := context.Background()

	// Add profiles
	posID := int64(1)
	profileRepo.AddProfile(&entities.UserWithOrg{ID: 1, Name: "User 1", PositionID: &posID})
	profileRepo.AddProfile(&entities.UserWithOrg{ID: 2, Name: "User 2"}) // No position

	// Get by position
	users, err := uc.GetUsersByPosition(ctx, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(users) != 1 {
		t.Errorf("expected 1 user, got %d", len(users))
	}
}

func TestUserUseCase_GetBaseUser(t *testing.T) {
	userRepo := NewMockUserRepository()
	profileRepo := NewMockUserProfileRepository()
	deptRepo := NewMockDepartmentRepository()
	posRepo := NewMockPositionRepository()
	uc := NewUserUseCase(userRepo, profileRepo, deptRepo, posRepo, nil, nil)

	ctx := context.Background()

	// Create user
	user := authEntities.NewUser(testEmail, "password", "Test User", authDomain.RoleStudent)
	_ = userRepo.Create(ctx, user)

	// Get base user
	result, err := uc.GetBaseUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.Email != testEmail {
		t.Errorf("expected email 'test@example.com', got '%s'", result.Email)
	}
}

func TestUserUseCase_GetBaseUser_NotFound(t *testing.T) {
	userRepo := NewMockUserRepository()
	profileRepo := NewMockUserProfileRepository()
	deptRepo := NewMockDepartmentRepository()
	posRepo := NewMockPositionRepository()
	uc := NewUserUseCase(userRepo, profileRepo, deptRepo, posRepo, nil, nil)

	ctx := context.Background()

	_, err := uc.GetBaseUser(ctx, 999)
	if err == nil {
		t.Error("expected error for non-existent user")
	}
}

// Entity tests

func TestNewDepartment(t *testing.T) {
	parentID := int64(1)
	dept := entities.NewDepartment("IT Department", "IT", "Information Technology", &parentID)

	if dept.Name != "IT Department" {
		t.Errorf("expected name 'IT Department', got '%s'", dept.Name)
	}

	if dept.Code != "IT" {
		t.Errorf("expected code 'IT', got '%s'", dept.Code)
	}

	if dept.Description != "Information Technology" {
		t.Errorf("expected description 'Information Technology', got '%s'", dept.Description)
	}

	if dept.ParentID == nil || *dept.ParentID != 1 {
		t.Errorf("expected parent ID 1, got %v", dept.ParentID)
	}

	if !dept.IsActive {
		t.Error("expected new department to be active")
	}

	if dept.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}

	if dept.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestNewPosition(t *testing.T) {
	pos := entities.NewPosition("Developer", "DEV", "Software Developer", 2)

	if pos.Name != "Developer" {
		t.Errorf("expected name 'Developer', got '%s'", pos.Name)
	}

	if pos.Code != "DEV" {
		t.Errorf("expected code 'DEV', got '%s'", pos.Code)
	}

	if pos.Description != "Software Developer" {
		t.Errorf("expected description 'Software Developer', got '%s'", pos.Description)
	}

	if pos.Level != 2 {
		t.Errorf("expected level 2, got %d", pos.Level)
	}

	if !pos.IsActive {
		t.Error("expected new position to be active")
	}

	if pos.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestUserWithOrg_Fields(t *testing.T) {
	deptID := int64(1)
	posID := int64(2)
	now := time.Now()

	user := &entities.UserWithOrg{
		ID:             1,
		Email:          testEmail,
		Name:           "Test User",
		Role:           "teacher",
		Status:         "active",
		Phone:          "+1234567890",
		Avatar:         "avatar.png",
		Bio:            "Test bio",
		DepartmentID:   &deptID,
		DepartmentName: "IT",
		PositionID:     &posID,
		PositionName:   "Developer",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if user.ID != 1 {
		t.Errorf("expected ID 1, got %d", user.ID)
	}

	if user.Email != testEmail {
		t.Errorf("expected email 'test@example.com', got '%s'", user.Email)
	}

	if *user.DepartmentID != 1 {
		t.Errorf("expected department ID 1, got %d", *user.DepartmentID)
	}

	if *user.PositionID != 2 {
		t.Errorf("expected position ID 2, got %d", *user.PositionID)
	}
}
