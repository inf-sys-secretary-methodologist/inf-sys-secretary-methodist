package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/domain/entities"
)

// MockPositionRepository implements PositionRepository for testing.
type MockPositionRepository struct {
	positions map[int64]*entities.Position
	nextID    int64
	createErr error
	updateErr error
	deleteErr error
	listErr   error
	countErr  error
}

func NewMockPositionRepository() *MockPositionRepository {
	return &MockPositionRepository{
		positions: make(map[int64]*entities.Position),
		nextID:    1,
	}
}

func (m *MockPositionRepository) Create(_ context.Context, position *entities.Position) error {
	if m.createErr != nil {
		return m.createErr
	}
	position.ID = m.nextID
	m.nextID++
	m.positions[position.ID] = position
	return nil
}

func (m *MockPositionRepository) GetByID(_ context.Context, id int64) (*entities.Position, error) {
	pos, exists := m.positions[id]
	if !exists {
		return nil, errors.New("position not found")
	}
	return pos, nil
}

func (m *MockPositionRepository) GetByCode(_ context.Context, code string) (*entities.Position, error) {
	for _, pos := range m.positions {
		if pos.Code == code {
			return pos, nil
		}
	}
	return nil, errors.New("position not found")
}

func (m *MockPositionRepository) Update(_ context.Context, position *entities.Position) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	m.positions[position.ID] = position
	return nil
}

func (m *MockPositionRepository) Delete(_ context.Context, id int64) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	delete(m.positions, id)
	return nil
}

func (m *MockPositionRepository) List(_ context.Context, limit, offset int, _ bool) ([]*entities.Position, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var positions []*entities.Position
	i := 0
	for _, pos := range m.positions {
		if i >= offset && len(positions) < limit {
			positions = append(positions, pos)
		}
		i++
	}
	return positions, nil
}

func (m *MockPositionRepository) Count(_ context.Context, _ bool) (int64, error) {
	if m.countErr != nil {
		return 0, m.countErr
	}
	return int64(len(m.positions)), nil
}

// Tests

func TestPositionUseCase_CreatePosition(t *testing.T) {
	repo := NewMockPositionRepository()
	uc := NewPositionUseCase(repo, nil)

	ctx := context.Background()
	input := &dto.CreatePositionInput{
		Name:        "Senior Developer",
		Code:        "SENDEV",
		Description: "Senior Software Developer",
		Level:       3,
	}

	pos, err := uc.CreatePosition(ctx, input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if pos.ID == 0 {
		t.Error("expected position ID to be set")
	}

	if pos.Name != "Senior Developer" {
		t.Errorf("expected name 'Senior Developer', got '%s'", pos.Name)
	}

	if pos.Code != "SENDEV" {
		t.Errorf("expected code 'SENDEV', got '%s'", pos.Code)
	}

	if pos.Level != 3 {
		t.Errorf("expected level 3, got %d", pos.Level)
	}

	if !pos.IsActive {
		t.Error("expected new position to be active")
	}
}

func TestPositionUseCase_CreatePosition_WithAuditLogger(t *testing.T) {
	repo := NewMockPositionRepository()
	uc := NewPositionUseCase(repo, testAuditLogger())

	ctx := context.Background()
	input := &dto.CreatePositionInput{
		Name: "Developer",
		Code: "DEV",
	}

	pos, err := uc.CreatePosition(ctx, input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if pos.ID == 0 {
		t.Error("expected position ID to be set")
	}
}

func TestPositionUseCase_CreatePosition_CreateError(t *testing.T) {
	repo := NewMockPositionRepository()
	repo.createErr = errors.New("create error")
	uc := NewPositionUseCase(repo, nil)

	ctx := context.Background()
	input := &dto.CreatePositionInput{
		Name: "Developer",
		Code: "DEV",
	}

	_, err := uc.CreatePosition(ctx, input)
	if err == nil {
		t.Error("expected error from create")
	}
}

func TestPositionUseCase_GetPosition(t *testing.T) {
	repo := NewMockPositionRepository()
	uc := NewPositionUseCase(repo, nil)

	ctx := context.Background()

	// Create position
	created, _ := uc.CreatePosition(ctx, &dto.CreatePositionInput{
		Name: "Developer",
		Code: "DEV",
	})

	// Get position
	pos, err := uc.GetPosition(ctx, created.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if pos.ID != created.ID {
		t.Errorf("expected ID %d, got %d", created.ID, pos.ID)
	}
}

func TestPositionUseCase_GetPosition_NotFound(t *testing.T) {
	repo := NewMockPositionRepository()
	uc := NewPositionUseCase(repo, nil)

	ctx := context.Background()

	_, err := uc.GetPosition(ctx, 999)
	if err == nil {
		t.Error("expected error for non-existent position")
	}
}

func TestPositionUseCase_GetPositionByCode(t *testing.T) {
	repo := NewMockPositionRepository()
	uc := NewPositionUseCase(repo, nil)

	ctx := context.Background()

	// Create position
	_, _ = uc.CreatePosition(ctx, &dto.CreatePositionInput{
		Name: "Developer",
		Code: "DEV",
	})

	// Get by code
	pos, err := uc.GetPositionByCode(ctx, "DEV")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if pos.Code != "DEV" {
		t.Errorf("expected code 'DEV', got '%s'", pos.Code)
	}
}

func TestPositionUseCase_GetPositionByCode_NotFound(t *testing.T) {
	repo := NewMockPositionRepository()
	uc := NewPositionUseCase(repo, nil)

	ctx := context.Background()

	_, err := uc.GetPositionByCode(ctx, "NONEXISTENT")
	if err == nil {
		t.Error("expected error for non-existent code")
	}
}

func TestPositionUseCase_ListPositions(t *testing.T) {
	repo := NewMockPositionRepository()
	uc := NewPositionUseCase(repo, nil)

	ctx := context.Background()

	// Create positions
	_, _ = uc.CreatePosition(ctx, &dto.CreatePositionInput{Name: "Pos 1", Code: "P1"})
	_, _ = uc.CreatePosition(ctx, &dto.CreatePositionInput{Name: "Pos 2", Code: "P2"})
	_, _ = uc.CreatePosition(ctx, &dto.CreatePositionInput{Name: "Pos 3", Code: "P3"})

	// List
	result, err := uc.ListPositions(ctx, 1, 10, true)
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

func TestPositionUseCase_ListPositions_Pagination(t *testing.T) {
	repo := NewMockPositionRepository()
	uc := NewPositionUseCase(repo, nil)

	ctx := context.Background()

	// Default page
	result, _ := uc.ListPositions(ctx, 0, 10, true)
	if result.Page != 1 {
		t.Errorf("expected default page 1, got %d", result.Page)
	}

	// Default limit
	result, _ = uc.ListPositions(ctx, 1, 0, true)
	if result.Limit != 10 {
		t.Errorf("expected default limit 10, got %d", result.Limit)
	}

	// Max limit
	result, _ = uc.ListPositions(ctx, 1, 200, true)
	if result.Limit != 100 {
		t.Errorf("expected max limit 100, got %d", result.Limit)
	}
}

func TestPositionUseCase_ListPositions_ListError(t *testing.T) {
	repo := NewMockPositionRepository()
	repo.listErr = errors.New("list error")
	uc := NewPositionUseCase(repo, nil)

	_, err := uc.ListPositions(context.Background(), 1, 10, true)
	if err == nil {
		t.Error("expected error from list")
	}
}

func TestPositionUseCase_ListPositions_CountError(t *testing.T) {
	repo := NewMockPositionRepository()
	repo.countErr = errors.New("count error")
	uc := NewPositionUseCase(repo, nil)

	_, err := uc.ListPositions(context.Background(), 1, 10, true)
	if err == nil {
		t.Error("expected error from count")
	}
}

func TestPositionUseCase_ListPositions_TotalPages(t *testing.T) {
	repo := NewMockPositionRepository()
	uc := NewPositionUseCase(repo, nil)

	ctx := context.Background()

	// Create 5 positions
	for i := 1; i <= 5; i++ {
		_, _ = uc.CreatePosition(ctx, &dto.CreatePositionInput{
			Name: "Position",
			Code: "P",
		})
	}

	// List with limit 2 (should be 3 pages)
	result, _ := uc.ListPositions(ctx, 1, 2, true)
	if result.TotalPages != 3 {
		t.Errorf("expected 3 total pages, got %d", result.TotalPages)
	}
}

func TestPositionUseCase_UpdatePosition(t *testing.T) {
	repo := NewMockPositionRepository()
	uc := NewPositionUseCase(repo, nil)

	ctx := context.Background()

	// Create position
	created, _ := uc.CreatePosition(ctx, &dto.CreatePositionInput{
		Name:  "Developer",
		Code:  "DEV",
		Level: 1,
	})

	// Update
	isActive := false
	input := &dto.UpdatePositionInput{
		Name:        "Senior Developer",
		Code:        "SENDEV",
		Description: "Updated description",
		Level:       3,
		IsActive:    &isActive,
	}

	updated, err := uc.UpdatePosition(ctx, created.ID, input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if updated.Name != "Senior Developer" {
		t.Errorf("expected name 'Senior Developer', got '%s'", updated.Name)
	}

	if updated.Code != "SENDEV" {
		t.Errorf("expected code 'SENDEV', got '%s'", updated.Code)
	}

	if updated.Level != 3 {
		t.Errorf("expected level 3, got %d", updated.Level)
	}

	if updated.IsActive {
		t.Error("expected IsActive to be false")
	}
}

func TestPositionUseCase_UpdatePosition_WithAuditLogger(t *testing.T) {
	repo := NewMockPositionRepository()
	uc := NewPositionUseCase(repo, testAuditLogger())

	ctx := context.Background()

	created, _ := uc.CreatePosition(ctx, &dto.CreatePositionInput{
		Name:  "Developer",
		Code:  "DEV",
		Level: 1,
	})

	input := &dto.UpdatePositionInput{
		Name:  "Senior Developer",
		Code:  "SENDEV",
		Level: 3,
	}

	_, err := uc.UpdatePosition(ctx, created.ID, input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestPositionUseCase_UpdatePosition_UpdateError(t *testing.T) {
	repo := NewMockPositionRepository()
	uc := NewPositionUseCase(repo, nil)

	ctx := context.Background()

	created, _ := uc.CreatePosition(ctx, &dto.CreatePositionInput{
		Name: "Developer",
		Code: "DEV",
	})

	repo.updateErr = errors.New("update error")

	input := &dto.UpdatePositionInput{
		Name: "Updated",
		Code: "UPD",
	}

	_, err := uc.UpdatePosition(ctx, created.ID, input)
	if err == nil {
		t.Error("expected error from update")
	}
}

func TestPositionUseCase_UpdatePosition_NotFound(t *testing.T) {
	repo := NewMockPositionRepository()
	uc := NewPositionUseCase(repo, nil)

	ctx := context.Background()

	_, err := uc.UpdatePosition(ctx, 999, &dto.UpdatePositionInput{
		Name: "Test",
		Code: "TEST",
	})

	if err == nil {
		t.Error("expected error for non-existent position")
	}
}

func TestPositionUseCase_UpdatePosition_PartialUpdate(t *testing.T) {
	repo := NewMockPositionRepository()
	uc := NewPositionUseCase(repo, nil)

	ctx := context.Background()

	// Create position
	created, _ := uc.CreatePosition(ctx, &dto.CreatePositionInput{
		Name:  "Developer",
		Code:  "DEV",
		Level: 1,
	})

	// Update only some fields (IsActive nil means no change)
	input := &dto.UpdatePositionInput{
		Name:  "Developer Updated",
		Code:  "DEV",
		Level: 2,
	}

	updated, err := uc.UpdatePosition(ctx, created.ID, input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// IsActive should remain true (default)
	if !updated.IsActive {
		t.Error("expected IsActive to remain true")
	}
}

func TestPositionUseCase_DeletePosition(t *testing.T) {
	repo := NewMockPositionRepository()
	uc := NewPositionUseCase(repo, nil)

	ctx := context.Background()

	// Create position
	created, _ := uc.CreatePosition(ctx, &dto.CreatePositionInput{
		Name: "Developer",
		Code: "DEV",
	})

	// Delete
	err := uc.DeletePosition(ctx, created.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify deleted
	_, err = uc.GetPosition(ctx, created.ID)
	if err == nil {
		t.Error("expected error for deleted position")
	}
}

func TestPositionUseCase_DeletePosition_WithAuditLogger(t *testing.T) {
	repo := NewMockPositionRepository()
	uc := NewPositionUseCase(repo, testAuditLogger())

	ctx := context.Background()

	created, _ := uc.CreatePosition(ctx, &dto.CreatePositionInput{
		Name: "Developer",
		Code: "DEV",
	})

	err := uc.DeletePosition(ctx, created.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestPositionUseCase_DeletePosition_DeleteError(t *testing.T) {
	repo := NewMockPositionRepository()
	uc := NewPositionUseCase(repo, nil)

	ctx := context.Background()

	created, _ := uc.CreatePosition(ctx, &dto.CreatePositionInput{
		Name: "Developer",
		Code: "DEV",
	})

	repo.deleteErr = errors.New("delete error")

	err := uc.DeletePosition(ctx, created.ID)
	if err == nil {
		t.Error("expected error from delete")
	}
}

func TestPositionUseCase_DeletePosition_NotFound(t *testing.T) {
	repo := NewMockPositionRepository()
	uc := NewPositionUseCase(repo, nil)

	ctx := context.Background()

	err := uc.DeletePosition(ctx, 999)
	if err == nil {
		t.Error("expected error for non-existent position")
	}
}
