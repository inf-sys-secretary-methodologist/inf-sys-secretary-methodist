package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/domain/entities"
)

// MockSyncConflictRepository implements SyncConflictRepository for testing.
type MockSyncConflictRepository struct {
	conflicts map[int64]*entities.SyncConflict
	nextID    int64
}

func NewMockSyncConflictRepository() *MockSyncConflictRepository {
	return &MockSyncConflictRepository{
		conflicts: make(map[int64]*entities.SyncConflict),
		nextID:    1,
	}
}

func (m *MockSyncConflictRepository) Create(_ context.Context, conflict *entities.SyncConflict) error {
	conflict.ID = m.nextID
	m.nextID++
	m.conflicts[conflict.ID] = conflict
	return nil
}

func (m *MockSyncConflictRepository) Update(_ context.Context, conflict *entities.SyncConflict) error {
	if _, exists := m.conflicts[conflict.ID]; !exists {
		return errors.New("conflict not found")
	}
	m.conflicts[conflict.ID] = conflict
	return nil
}

func (m *MockSyncConflictRepository) GetByID(_ context.Context, id int64) (*entities.SyncConflict, error) {
	if conflict, exists := m.conflicts[id]; exists {
		copiedConflict := *conflict
		return &copiedConflict, nil
	}
	return nil, nil
}

func (m *MockSyncConflictRepository) List(_ context.Context, filter entities.SyncConflictFilter) ([]*entities.SyncConflict, int64, error) {
	var result []*entities.SyncConflict
	for _, conflict := range m.conflicts {
		// Apply filters
		if filter.SyncLogID != nil && conflict.SyncLogID != *filter.SyncLogID {
			continue
		}
		if filter.EntityType != nil && conflict.EntityType != *filter.EntityType {
			continue
		}
		if filter.Resolution != nil && conflict.Resolution != *filter.Resolution {
			continue
		}
		result = append(result, conflict)
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

func (m *MockSyncConflictRepository) GetBySyncLogID(_ context.Context, syncLogID int64) ([]*entities.SyncConflict, error) {
	var result []*entities.SyncConflict
	for _, conflict := range m.conflicts {
		if conflict.SyncLogID == syncLogID {
			result = append(result, conflict)
		}
	}
	return result, nil
}

func (m *MockSyncConflictRepository) GetPending(_ context.Context, limit, offset int) ([]*entities.SyncConflict, int64, error) {
	var result []*entities.SyncConflict
	for _, conflict := range m.conflicts {
		if conflict.IsPending() {
			result = append(result, conflict)
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

func (m *MockSyncConflictRepository) GetPendingByEntityType(_ context.Context, entityType entities.SyncEntityType) ([]*entities.SyncConflict, error) {
	var result []*entities.SyncConflict
	for _, conflict := range m.conflicts {
		if conflict.IsPending() && conflict.EntityType == entityType {
			result = append(result, conflict)
		}
	}
	return result, nil
}

func (m *MockSyncConflictRepository) Resolve(_ context.Context, id int64, resolution entities.ConflictResolution, userID int64, resolvedData string) error {
	if conflict, exists := m.conflicts[id]; exists {
		conflict.Resolve(resolution, userID, resolvedData)
		return nil
	}
	return errors.New("conflict not found")
}

func (m *MockSyncConflictRepository) BulkResolve(_ context.Context, ids []int64, resolution entities.ConflictResolution, userID int64) error {
	for _, id := range ids {
		if conflict, exists := m.conflicts[id]; exists {
			conflict.Resolve(resolution, userID, "")
		}
	}
	return nil
}

func (m *MockSyncConflictRepository) Delete(_ context.Context, id int64) error {
	delete(m.conflicts, id)
	return nil
}

func (m *MockSyncConflictRepository) DeleteBySyncLogID(_ context.Context, syncLogID int64) error {
	for id, conflict := range m.conflicts {
		if conflict.SyncLogID == syncLogID {
			delete(m.conflicts, id)
		}
	}
	return nil
}

func (m *MockSyncConflictRepository) GetStats(_ context.Context) (*entities.ConflictStats, error) {
	stats := &entities.ConflictStats{
		ByEntityType: make(map[entities.SyncEntityType]int64),
	}

	for _, conflict := range m.conflicts {
		stats.TotalConflicts++
		if conflict.IsPending() {
			stats.PendingConflicts++
		} else {
			stats.ResolvedConflicts++
		}
		stats.ByEntityType[conflict.EntityType]++
	}

	return stats, nil
}

// Helper to create test conflict
func createTestConflict(repo *MockSyncConflictRepository, syncLogID int64, entityType entities.SyncEntityType, entityID string) *entities.SyncConflict {
	conflict := entities.NewSyncConflict(syncLogID, entityType, entityID)
	_ = repo.Create(context.Background(), conflict)
	return conflict
}

// Tests

func TestConflictUseCase_List(t *testing.T) {
	repo := NewMockSyncConflictRepository()
	uc := NewConflictUseCase(repo)

	ctx := context.Background()

	// Create conflicts
	createTestConflict(repo, 1, entities.SyncEntityEmployee, "emp-1")
	createTestConflict(repo, 1, entities.SyncEntityEmployee, "emp-2")
	createTestConflict(repo, 2, entities.SyncEntityStudent, "stud-1")

	// List all
	req := &dto.ConflictListRequest{
		Limit: 10,
	}
	result, err := uc.List(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.Total != 3 {
		t.Errorf("expected total 3, got %d", result.Total)
	}
}

func TestConflictUseCase_List_WithFilter(t *testing.T) {
	repo := NewMockSyncConflictRepository()
	uc := NewConflictUseCase(repo)

	ctx := context.Background()

	// Create conflicts
	createTestConflict(repo, 1, entities.SyncEntityEmployee, "emp-1")
	createTestConflict(repo, 1, entities.SyncEntityEmployee, "emp-2")
	createTestConflict(repo, 2, entities.SyncEntityStudent, "stud-1")

	// Filter by sync log ID
	syncLogID := int64(1)
	req := &dto.ConflictListRequest{
		SyncLogID: &syncLogID,
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

func TestConflictUseCase_GetByID(t *testing.T) {
	repo := NewMockSyncConflictRepository()
	uc := NewConflictUseCase(repo)

	ctx := context.Background()

	// Create conflict
	conflict := createTestConflict(repo, 1, entities.SyncEntityEmployee, "emp-1")

	// Get by ID
	result, err := uc.GetByID(ctx, conflict.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if result.EntityID != "emp-1" {
		t.Errorf("expected entity ID 'emp-1', got '%s'", result.EntityID)
	}
}

func TestConflictUseCase_GetByID_NotFound(t *testing.T) {
	repo := NewMockSyncConflictRepository()
	uc := NewConflictUseCase(repo)

	ctx := context.Background()

	// Get non-existent
	result, err := uc.GetByID(ctx, 999)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result != nil {
		t.Error("expected nil result for non-existent conflict")
	}
}

func TestConflictUseCase_GetPending(t *testing.T) {
	repo := NewMockSyncConflictRepository()
	uc := NewConflictUseCase(repo)

	ctx := context.Background()

	// Create conflicts
	createTestConflict(repo, 1, entities.SyncEntityEmployee, "emp-1")
	conflict2 := createTestConflict(repo, 1, entities.SyncEntityEmployee, "emp-2")
	createTestConflict(repo, 2, entities.SyncEntityStudent, "stud-1")

	// Resolve one
	_ = repo.Resolve(ctx, conflict2.ID, entities.ConflictResolutionUseLocal, 1, "{}")

	// Get pending
	result, err := uc.GetPending(ctx, 10, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.Total != 2 {
		t.Errorf("expected total 2 pending, got %d", result.Total)
	}
}

func TestConflictUseCase_GetBySyncLogID(t *testing.T) {
	repo := NewMockSyncConflictRepository()
	uc := NewConflictUseCase(repo)

	ctx := context.Background()

	// Create conflicts
	createTestConflict(repo, 1, entities.SyncEntityEmployee, "emp-1")
	createTestConflict(repo, 1, entities.SyncEntityEmployee, "emp-2")
	createTestConflict(repo, 2, entities.SyncEntityStudent, "stud-1")

	// Get by sync log ID
	result, err := uc.GetBySyncLogID(ctx, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 conflicts, got %d", len(result))
	}
}

func TestConflictUseCase_Resolve(t *testing.T) {
	repo := NewMockSyncConflictRepository()
	uc := NewConflictUseCase(repo)

	ctx := context.Background()

	// Create conflict
	conflict := createTestConflict(repo, 1, entities.SyncEntityEmployee, "emp-1")

	// Resolve
	req := &dto.ResolveConflictRequest{
		Resolution:   entities.ConflictResolutionUseLocal,
		ResolvedData: "{}",
	}
	err := uc.Resolve(ctx, conflict.ID, 42, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify resolved
	resolved, _ := repo.GetByID(ctx, conflict.ID)
	if resolved.IsPending() {
		t.Error("expected conflict to be resolved")
	}
	if resolved.Resolution != entities.ConflictResolutionUseLocal {
		t.Errorf("expected resolution 'use_local', got '%s'", resolved.Resolution)
	}
}

func TestConflictUseCase_Resolve_NotFound(t *testing.T) {
	repo := NewMockSyncConflictRepository()
	uc := NewConflictUseCase(repo)

	ctx := context.Background()

	// Try to resolve non-existent
	req := &dto.ResolveConflictRequest{
		Resolution: entities.ConflictResolutionUseLocal,
	}
	err := uc.Resolve(ctx, 999, 42, req)
	if err == nil {
		t.Error("expected error for non-existent conflict")
	}
}

func TestConflictUseCase_Resolve_AlreadyResolved(t *testing.T) {
	repo := NewMockSyncConflictRepository()
	uc := NewConflictUseCase(repo)

	ctx := context.Background()

	// Create and resolve conflict
	conflict := createTestConflict(repo, 1, entities.SyncEntityEmployee, "emp-1")
	_ = repo.Resolve(ctx, conflict.ID, entities.ConflictResolutionUseLocal, 1, "{}")

	// Try to resolve again
	req := &dto.ResolveConflictRequest{
		Resolution: entities.ConflictResolutionUseExternal,
	}
	err := uc.Resolve(ctx, conflict.ID, 42, req)
	if err == nil {
		t.Error("expected error for already resolved conflict")
	}
}

func TestConflictUseCase_Resolve_WithNotes(t *testing.T) {
	repo := NewMockSyncConflictRepository()
	uc := NewConflictUseCase(repo)

	ctx := context.Background()

	// Create conflict
	conflict := createTestConflict(repo, 1, entities.SyncEntityEmployee, "emp-1")

	// Resolve with notes
	req := &dto.ResolveConflictRequest{
		Resolution:   entities.ConflictResolutionMerge,
		ResolvedData: "{}",
		Notes:        "Merged manually",
	}
	err := uc.Resolve(ctx, conflict.ID, 42, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify notes
	resolved, _ := repo.GetByID(ctx, conflict.ID)
	if resolved.Notes != "Merged manually" {
		t.Errorf("expected notes 'Merged manually', got '%s'", resolved.Notes)
	}
}

func TestConflictUseCase_BulkResolve(t *testing.T) {
	repo := NewMockSyncConflictRepository()
	uc := NewConflictUseCase(repo)

	ctx := context.Background()

	// Create conflicts
	conflict1 := createTestConflict(repo, 1, entities.SyncEntityEmployee, "emp-1")
	conflict2 := createTestConflict(repo, 1, entities.SyncEntityEmployee, "emp-2")
	conflict3 := createTestConflict(repo, 1, entities.SyncEntityEmployee, "emp-3")

	// Bulk resolve
	req := &dto.BulkResolveRequest{
		IDs:        []int64{conflict1.ID, conflict2.ID, conflict3.ID},
		Resolution: entities.ConflictResolutionSkip,
	}
	err := uc.BulkResolve(ctx, 42, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify all resolved
	for _, id := range []int64{conflict1.ID, conflict2.ID, conflict3.ID} {
		resolved, _ := repo.GetByID(ctx, id)
		if resolved.IsPending() {
			t.Errorf("expected conflict %d to be resolved", id)
		}
	}
}

func TestConflictUseCase_BulkResolve_EmptyIDs(t *testing.T) {
	repo := NewMockSyncConflictRepository()
	uc := NewConflictUseCase(repo)

	ctx := context.Background()

	// Try to bulk resolve with empty IDs
	req := &dto.BulkResolveRequest{
		IDs:        []int64{},
		Resolution: entities.ConflictResolutionSkip,
	}
	err := uc.BulkResolve(ctx, 42, req)
	if err == nil {
		t.Error("expected error for empty IDs")
	}
}

func TestConflictUseCase_GetStats(t *testing.T) {
	repo := NewMockSyncConflictRepository()
	uc := NewConflictUseCase(repo)

	ctx := context.Background()

	// Create conflicts
	createTestConflict(repo, 1, entities.SyncEntityEmployee, "emp-1")
	conflict2 := createTestConflict(repo, 1, entities.SyncEntityEmployee, "emp-2")
	createTestConflict(repo, 2, entities.SyncEntityStudent, "stud-1")

	// Resolve one
	_ = repo.Resolve(ctx, conflict2.ID, entities.ConflictResolutionUseLocal, 1, "{}")

	// Get stats
	result, err := uc.GetStats(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.TotalConflicts != 3 {
		t.Errorf("expected total 3, got %d", result.TotalConflicts)
	}
	if result.PendingConflicts != 2 {
		t.Errorf("expected pending 2, got %d", result.PendingConflicts)
	}
	if result.ResolvedConflicts != 1 {
		t.Errorf("expected resolved 1, got %d", result.ResolvedConflicts)
	}
}

func TestConflictUseCase_Delete(t *testing.T) {
	repo := NewMockSyncConflictRepository()
	uc := NewConflictUseCase(repo)

	ctx := context.Background()

	// Create conflict
	conflict := createTestConflict(repo, 1, entities.SyncEntityEmployee, "emp-1")

	// Delete
	err := uc.Delete(ctx, conflict.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify deleted
	result, _ := repo.GetByID(ctx, conflict.ID)
	if result != nil {
		t.Error("expected conflict to be deleted")
	}
}

// --- Error-injecting mock for conflict repository ---

type errorConflictRepo struct {
	MockSyncConflictRepository
	listErr       bool
	getByIDErr    bool
	getPendingErr bool
	getBySyncErr  bool
	resolveErr    bool
	bulkErr       bool
	getStatsErr   bool
	deleteErr     bool
	updateErr     bool
}

func newErrorConflictRepo() *errorConflictRepo {
	return &errorConflictRepo{
		MockSyncConflictRepository: *NewMockSyncConflictRepository(),
	}
}

func (m *errorConflictRepo) List(_ context.Context, _ entities.SyncConflictFilter) ([]*entities.SyncConflict, int64, error) {
	if m.listErr {
		return nil, 0, errors.New("list error")
	}
	return m.MockSyncConflictRepository.List(context.Background(), entities.SyncConflictFilter{})
}

func (m *errorConflictRepo) GetByID(_ context.Context, id int64) (*entities.SyncConflict, error) {
	if m.getByIDErr {
		return nil, errors.New("get by ID error")
	}
	return m.MockSyncConflictRepository.GetByID(context.Background(), id)
}

func (m *errorConflictRepo) GetPending(_ context.Context, limit, offset int) ([]*entities.SyncConflict, int64, error) {
	if m.getPendingErr {
		return nil, 0, errors.New("get pending error")
	}
	return m.MockSyncConflictRepository.GetPending(context.Background(), limit, offset)
}

func (m *errorConflictRepo) GetBySyncLogID(_ context.Context, syncLogID int64) ([]*entities.SyncConflict, error) {
	if m.getBySyncErr {
		return nil, errors.New("get by sync log error")
	}
	return m.MockSyncConflictRepository.GetBySyncLogID(context.Background(), syncLogID)
}

func (m *errorConflictRepo) Resolve(_ context.Context, id int64, resolution entities.ConflictResolution, userID int64, resolvedData string) error {
	if m.resolveErr {
		return errors.New("resolve error")
	}
	return m.MockSyncConflictRepository.Resolve(context.Background(), id, resolution, userID, resolvedData)
}

func (m *errorConflictRepo) BulkResolve(_ context.Context, ids []int64, resolution entities.ConflictResolution, userID int64) error {
	if m.bulkErr {
		return errors.New("bulk resolve error")
	}
	return m.MockSyncConflictRepository.BulkResolve(context.Background(), ids, resolution, userID)
}

func (m *errorConflictRepo) GetStats(_ context.Context) (*entities.ConflictStats, error) {
	if m.getStatsErr {
		return nil, errors.New("get stats error")
	}
	return m.MockSyncConflictRepository.GetStats(context.Background())
}

func (m *errorConflictRepo) Delete(_ context.Context, id int64) error {
	if m.deleteErr {
		return errors.New("delete error")
	}
	return m.MockSyncConflictRepository.Delete(context.Background(), id)
}

func (m *errorConflictRepo) Update(_ context.Context, conflict *entities.SyncConflict) error {
	if m.updateErr {
		return errors.New("update error")
	}
	return m.MockSyncConflictRepository.Update(context.Background(), conflict)
}

// --- Error path tests ---

func TestConflictUseCase_List_Error(t *testing.T) {
	repo := newErrorConflictRepo()
	repo.listErr = true
	uc := NewConflictUseCase(repo)

	req := &dto.ConflictListRequest{Limit: 10}
	result, err := uc.List(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to list conflicts")
}

func TestConflictUseCase_GetByID_Error(t *testing.T) {
	repo := newErrorConflictRepo()
	repo.getByIDErr = true
	uc := NewConflictUseCase(repo)

	result, err := uc.GetByID(context.Background(), 1)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get conflict")
}

func TestConflictUseCase_GetPending_Error(t *testing.T) {
	repo := newErrorConflictRepo()
	repo.getPendingErr = true
	uc := NewConflictUseCase(repo)

	result, err := uc.GetPending(context.Background(), 10, 0)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get pending conflicts")
}

func TestConflictUseCase_GetBySyncLogID_Error(t *testing.T) {
	repo := newErrorConflictRepo()
	repo.getBySyncErr = true
	uc := NewConflictUseCase(repo)

	result, err := uc.GetBySyncLogID(context.Background(), 1)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get conflicts by sync log")
}

func TestConflictUseCase_Resolve_GetByIDError(t *testing.T) {
	repo := newErrorConflictRepo()
	repo.getByIDErr = true
	uc := NewConflictUseCase(repo)

	req := &dto.ResolveConflictRequest{Resolution: entities.ConflictResolutionUseLocal}
	err := uc.Resolve(context.Background(), 1, 42, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get conflict")
}

func TestConflictUseCase_Resolve_ResolveRepoError(t *testing.T) {
	repo := newErrorConflictRepo()
	uc := NewConflictUseCase(repo)
	ctx := context.Background()

	// Create a conflict
	conflict := entities.NewSyncConflict(1, entities.SyncEntityEmployee, "emp-1")
	_ = repo.MockSyncConflictRepository.Create(ctx, conflict)

	repo.resolveErr = true
	req := &dto.ResolveConflictRequest{
		Resolution:   entities.ConflictResolutionUseLocal,
		ResolvedData: "{}",
	}
	err := uc.Resolve(ctx, conflict.ID, 42, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to resolve conflict")
}

func TestConflictUseCase_Resolve_UpdateNotesError(t *testing.T) {
	repo := newErrorConflictRepo()
	uc := NewConflictUseCase(repo)
	ctx := context.Background()

	conflict := entities.NewSyncConflict(1, entities.SyncEntityEmployee, "emp-1")
	_ = repo.MockSyncConflictRepository.Create(ctx, conflict)

	repo.updateErr = true
	req := &dto.ResolveConflictRequest{
		Resolution:   entities.ConflictResolutionUseLocal,
		ResolvedData: "{}",
		Notes:        "some notes",
	}
	err := uc.Resolve(ctx, conflict.ID, 42, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update conflict notes")
}

func TestConflictUseCase_BulkResolve_Error(t *testing.T) {
	repo := newErrorConflictRepo()
	repo.bulkErr = true
	uc := NewConflictUseCase(repo)

	req := &dto.BulkResolveRequest{
		IDs:        []int64{1, 2},
		Resolution: entities.ConflictResolutionSkip,
	}
	err := uc.BulkResolve(context.Background(), 42, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to bulk resolve conflicts")
}

func TestConflictUseCase_GetStats_Error(t *testing.T) {
	repo := newErrorConflictRepo()
	repo.getStatsErr = true
	uc := NewConflictUseCase(repo)

	result, err := uc.GetStats(context.Background())
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get conflict stats")
}

func TestConflictUseCase_Delete_Error(t *testing.T) {
	repo := newErrorConflictRepo()
	repo.deleteErr = true
	uc := NewConflictUseCase(repo)

	err := uc.Delete(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete conflict")
}

func TestConflictUseCase_Resolve_NoNotes(t *testing.T) {
	repo := NewMockSyncConflictRepository()
	uc := NewConflictUseCase(repo)
	ctx := context.Background()

	conflict := createTestConflict(repo, 1, entities.SyncEntityEmployee, "emp-1")
	req := &dto.ResolveConflictRequest{
		Resolution:   entities.ConflictResolutionUseExternal,
		ResolvedData: "{}",
		Notes:        "",
	}
	err := uc.Resolve(ctx, conflict.ID, 42, req)
	require.NoError(t, err)

	resolved, _ := repo.GetByID(ctx, conflict.ID)
	assert.Equal(t, entities.ConflictResolutionUseExternal, resolved.Resolution)
}
