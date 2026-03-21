package usecases

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/infrastructure/odata"
)

const (
	testFirstNameJohn  = "John"
	testLastNameDoe    = "Doe"
	testOldHash        = "oldhash"
	testOldEmail       = "old@example.com"
	testFirstNameAlice = "Alice"
	testLastNameWonder = "Wonder"
)

// --- Mock SyncLogRepository ---

type mockSyncLogRepo struct {
	logs        map[int64]*entities.SyncLog
	nextID      int64
	createErr   bool
	updateErr   bool
	getByIDErr  bool
	listErr     bool
	getStatsErr bool
}

func newMockSyncLogRepo() *mockSyncLogRepo {
	return &mockSyncLogRepo{
		logs:   make(map[int64]*entities.SyncLog),
		nextID: 1,
	}
}

func (m *mockSyncLogRepo) Create(_ context.Context, log *entities.SyncLog) error {
	if m.createErr {
		return errors.New("create sync log error")
	}
	log.ID = m.nextID
	m.nextID++
	m.logs[log.ID] = log
	return nil
}

func (m *mockSyncLogRepo) Update(_ context.Context, log *entities.SyncLog) error {
	if m.updateErr {
		return errors.New("update sync log error")
	}
	m.logs[log.ID] = log
	return nil
}

func (m *mockSyncLogRepo) GetByID(_ context.Context, id int64) (*entities.SyncLog, error) {
	if m.getByIDErr {
		return nil, errors.New("get sync log error")
	}
	if log, ok := m.logs[id]; ok {
		copied := *log
		return &copied, nil
	}
	return nil, nil
}

func (m *mockSyncLogRepo) List(_ context.Context, filter entities.SyncLogFilter) ([]*entities.SyncLog, int64, error) {
	if m.listErr {
		return nil, 0, errors.New("list sync logs error")
	}
	var result []*entities.SyncLog
	for _, log := range m.logs {
		if filter.EntityType != nil && log.EntityType != *filter.EntityType {
			continue
		}
		if filter.Direction != nil && log.Direction != *filter.Direction {
			continue
		}
		if filter.Status != nil && log.Status != *filter.Status {
			continue
		}
		result = append(result, log)
	}
	total := int64(len(result))
	if filter.Offset > 0 && filter.Offset < len(result) {
		result = result[filter.Offset:]
	}
	if filter.Limit > 0 && filter.Limit < len(result) {
		result = result[:filter.Limit]
	}
	return result, total, nil
}

func (m *mockSyncLogRepo) GetLatest(_ context.Context, _ entities.SyncEntityType) (*entities.SyncLog, error) {
	return nil, nil
}

func (m *mockSyncLogRepo) GetRunning(_ context.Context) ([]*entities.SyncLog, error) {
	return nil, nil
}

func (m *mockSyncLogRepo) GetStats(_ context.Context, _ *entities.SyncEntityType) (*entities.SyncStats, error) {
	if m.getStatsErr {
		return nil, errors.New("get stats error")
	}
	return &entities.SyncStats{
		TotalSyncs:      5,
		SuccessfulSyncs: 3,
		FailedSyncs:     2,
		TotalRecords:    100,
		TotalConflicts:  10,
		LastSyncAt:      time.Now(),
	}, nil
}

func (m *mockSyncLogRepo) Delete(_ context.Context, id int64) error {
	delete(m.logs, id)
	return nil
}

func (m *mockSyncLogRepo) DeleteOlderThan(_ context.Context, _ int) (int64, error) {
	return 0, nil
}

// --- Error-injecting employee repo ---

type errMockEmployeeRepo struct {
	MockExternalEmployeeRepository
	getByExternalIDErr   bool
	createErr            bool
	updateErr            bool
	getAllExternalIDsErr bool
	markInactiveErr      bool
}

func newErrMockEmployeeRepo() *errMockEmployeeRepo {
	return &errMockEmployeeRepo{
		MockExternalEmployeeRepository: *NewMockExternalEmployeeRepository(),
	}
}

func (m *errMockEmployeeRepo) GetByExternalID(ctx context.Context, externalID string) (*entities.ExternalEmployee, error) {
	if m.getByExternalIDErr {
		return nil, errors.New("get by external ID error")
	}
	return m.MockExternalEmployeeRepository.GetByExternalID(ctx, externalID)
}

func (m *errMockEmployeeRepo) Create(ctx context.Context, emp *entities.ExternalEmployee) error {
	if m.createErr {
		return errors.New("create employee error")
	}
	return m.MockExternalEmployeeRepository.Create(ctx, emp)
}

func (m *errMockEmployeeRepo) Update(ctx context.Context, emp *entities.ExternalEmployee) error {
	if m.updateErr {
		return errors.New("update employee error")
	}
	return m.MockExternalEmployeeRepository.Update(ctx, emp)
}

func (m *errMockEmployeeRepo) GetAllExternalIDs(ctx context.Context) ([]string, error) {
	if m.getAllExternalIDsErr {
		return nil, errors.New("get all external IDs error")
	}
	return m.MockExternalEmployeeRepository.GetAllExternalIDs(ctx)
}

func (m *errMockEmployeeRepo) MarkInactiveExcept(ctx context.Context, ids []string) error {
	if m.markInactiveErr {
		return errors.New("mark inactive error")
	}
	return m.MockExternalEmployeeRepository.MarkInactiveExcept(ctx, ids)
}

// --- Error-injecting student repo ---

type errMockStudentRepo struct {
	MockExternalStudentRepository
	getByExternalIDErr   bool
	createErr            bool
	updateErr            bool
	getAllExternalIDsErr bool
	markInactiveErr      bool
}

func newErrMockStudentRepo() *errMockStudentRepo {
	return &errMockStudentRepo{
		MockExternalStudentRepository: *NewMockExternalStudentRepository(),
	}
}

func (m *errMockStudentRepo) GetByExternalID(ctx context.Context, externalID string) (*entities.ExternalStudent, error) {
	if m.getByExternalIDErr {
		return nil, errors.New("get by external ID error")
	}
	return m.MockExternalStudentRepository.GetByExternalID(ctx, externalID)
}

func (m *errMockStudentRepo) Create(ctx context.Context, s *entities.ExternalStudent) error {
	if m.createErr {
		return errors.New("create student error")
	}
	return m.MockExternalStudentRepository.Create(ctx, s)
}

func (m *errMockStudentRepo) Update(ctx context.Context, s *entities.ExternalStudent) error {
	if m.updateErr {
		return errors.New("update student error")
	}
	return m.MockExternalStudentRepository.Update(ctx, s)
}

func (m *errMockStudentRepo) GetAllExternalIDs(ctx context.Context) ([]string, error) {
	if m.getAllExternalIDsErr {
		return nil, errors.New("get all external IDs error")
	}
	return m.MockExternalStudentRepository.GetAllExternalIDs(ctx)
}

func (m *errMockStudentRepo) MarkInactiveExcept(ctx context.Context, ids []string) error {
	if m.markInactiveErr {
		return errors.New("mark inactive error")
	}
	return m.MockExternalStudentRepository.MarkInactiveExcept(ctx, ids)
}

// --- Error-injecting conflict repo ---

type errMockConflictRepo struct {
	MockSyncConflictRepository
	createErr bool
}

func newErrMockConflictRepo() *errMockConflictRepo {
	return &errMockConflictRepo{
		MockSyncConflictRepository: *NewMockSyncConflictRepository(),
	}
}

func (m *errMockConflictRepo) Create(ctx context.Context, c *entities.SyncConflict) error {
	if m.createErr {
		return errors.New("create conflict error")
	}
	return m.MockSyncConflictRepository.Create(ctx, c)
}

// --- Test helpers ---

func newTestODataServer(t *testing.T, employeesHandler, studentsHandler http.HandlerFunc) (*httptest.Server, *odata.Client) {
	t.Helper()
	mux := http.NewServeMux()
	if employeesHandler != nil {
		mux.HandleFunc("/odata/Catalog_Employees", employeesHandler)
	}
	if studentsHandler != nil {
		mux.HandleFunc("/odata/Catalog_Students", studentsHandler)
	}
	// Ping handler
	mux.HandleFunc("/odata/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	server := httptest.NewServer(mux)

	cfg := &odata.Config{
		BaseURL:          server.URL + "/odata",
		Username:         "test",
		Password:         "test",
		Timeout:          5 * time.Second,
		MaxRetries:       0,
		RetryDelay:       0,
		EmployeesCatalog: "Catalog_Employees",
		StudentsCatalog:  "Catalog_Students",
	}
	client := odata.NewClient(cfg)
	return server, client
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func newSyncUCWithOData(t *testing.T, odataClient *odata.Client) (*SyncUseCase, *mockSyncLogRepo, *errMockEmployeeRepo, *errMockStudentRepo, *errMockConflictRepo) {
	t.Helper()
	syncLogRepo := newMockSyncLogRepo()
	empRepo := newErrMockEmployeeRepo()
	studRepo := newErrMockStudentRepo()
	conflictRepo := newErrMockConflictRepo()
	uc := NewSyncUseCase(odataClient, syncLogRepo, empRepo, studRepo, conflictRepo, newTestLogger())
	return uc, syncLogRepo, empRepo, studRepo, conflictRepo
}

// --- Tests ---

func TestNewSyncUseCase(t *testing.T) {
	_, client := newTestODataServer(t, nil, nil)
	uc := NewSyncUseCase(client, newMockSyncLogRepo(), newErrMockEmployeeRepo(), newErrMockStudentRepo(), newErrMockConflictRepo(), newTestLogger())
	assert.NotNil(t, uc)
	assert.NotNil(t, uc.running)
}

func TestStartSync_AlreadyRunning(t *testing.T) {
	server, client := newTestODataServer(t, nil, nil)
	defer server.Close()
	uc, _, _, _, _ := newSyncUCWithOData(t, client)
	ctx := context.Background()

	// Manually mark as running
	uc.mu.Lock()
	uc.running[entities.SyncEntityEmployee] = true
	uc.mu.Unlock()

	req := &dto.StartSyncRequest{
		EntityType: entities.SyncEntityEmployee,
		Direction:  entities.SyncDirectionImport,
		Force:      false,
	}
	result, err := uc.StartSync(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "sync is already running")
}

func TestStartSync_ForceOverridesRunning(t *testing.T) {
	employees := []entities.ODataEmployee{
		{RefKey: "emp-1", Code: "C1", FirstName: "John", LastName: "Doe"},
	}
	server, client := newTestODataServer(t, func(w http.ResponseWriter, _ *http.Request) {
		resp := odata.Response[entities.ODataEmployee]{Value: employees}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}, nil)
	defer server.Close()

	uc, _, _, _, _ := newSyncUCWithOData(t, client)
	ctx := context.Background()

	// Manually mark as running
	uc.mu.Lock()
	uc.running[entities.SyncEntityEmployee] = true
	uc.mu.Unlock()

	req := &dto.StartSyncRequest{
		EntityType: entities.SyncEntityEmployee,
		Direction:  entities.SyncDirectionImport,
		Force:      true,
	}
	result, err := uc.StartSync(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Employee sync completed", result.Message)

	// Running flag should be cleared after sync
	assert.False(t, uc.IsSyncRunning(entities.SyncEntityEmployee))
}

func TestStartSync_CreateSyncLogError(t *testing.T) {
	server, client := newTestODataServer(t, nil, nil)
	defer server.Close()

	uc, syncLogRepo, _, _, _ := newSyncUCWithOData(t, client)
	syncLogRepo.createErr = true
	ctx := context.Background()

	req := &dto.StartSyncRequest{
		EntityType: entities.SyncEntityEmployee,
		Direction:  entities.SyncDirectionImport,
	}
	result, err := uc.StartSync(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to create sync log")
}

func TestStartSync_UnsupportedEntityType(t *testing.T) {
	server, client := newTestODataServer(t, nil, nil)
	defer server.Close()

	uc, _, _, _, _ := newSyncUCWithOData(t, client)
	ctx := context.Background()

	req := &dto.StartSyncRequest{
		EntityType: entities.SyncEntityFinance,
		Direction:  entities.SyncDirectionImport,
	}
	result, err := uc.StartSync(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unsupported entity type")
}

func TestStartSync_EmployeeSync_Success_NewEmployees(t *testing.T) {
	employees := []entities.ODataEmployee{
		{RefKey: "emp-1", Code: "C1", FirstName: "John", LastName: "Doe", Email: "john@example.com"},
		{RefKey: "emp-2", Code: "C2", FirstName: "Jane", LastName: "Smith", Email: "jane@example.com"},
	}
	server, client := newTestODataServer(t, func(w http.ResponseWriter, _ *http.Request) {
		resp := odata.Response[entities.ODataEmployee]{Value: employees}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}, nil)
	defer server.Close()

	uc, _, _, _, _ := newSyncUCWithOData(t, client)
	ctx := context.Background()

	req := &dto.StartSyncRequest{
		EntityType: entities.SyncEntityEmployee,
		Direction:  entities.SyncDirectionImport,
	}
	result, err := uc.StartSync(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 2, result.TotalRecords)
	assert.Equal(t, 2, result.SuccessCount)
	assert.Equal(t, 0, result.ErrorCount)
	assert.Equal(t, "Employee sync completed", result.Message)
}

func TestStartSync_EmployeeSync_ODataFetchError(t *testing.T) {
	server, client := newTestODataServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad request"))
	}, nil)
	defer server.Close()

	uc, _, _, _, _ := newSyncUCWithOData(t, client)
	ctx := context.Background()

	req := &dto.StartSyncRequest{
		EntityType: entities.SyncEntityEmployee,
		Direction:  entities.SyncDirectionImport,
	}
	result, err := uc.StartSync(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to fetch employees from 1C")
}

func TestStartSync_EmployeeSync_GetAllExternalIDsError(t *testing.T) {
	employees := []entities.ODataEmployee{
		{RefKey: "emp-1", Code: "C1", FirstName: "John", LastName: "Doe"},
	}
	server, client := newTestODataServer(t, func(w http.ResponseWriter, _ *http.Request) {
		resp := odata.Response[entities.ODataEmployee]{Value: employees}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}, nil)
	defer server.Close()

	uc, _, empRepo, _, _ := newSyncUCWithOData(t, client)
	empRepo.getAllExternalIDsErr = true
	ctx := context.Background()

	req := &dto.StartSyncRequest{
		EntityType: entities.SyncEntityEmployee,
		Direction:  entities.SyncDirectionImport,
	}
	result, err := uc.StartSync(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get existing employee IDs")
}

func TestStartSync_EmployeeSync_UpdateExisting(t *testing.T) {
	employees := []entities.ODataEmployee{
		{RefKey: "emp-1", Code: "C1", FirstName: "John-Updated", LastName: "Doe"},
	}
	server, client := newTestODataServer(t, func(w http.ResponseWriter, _ *http.Request) {
		resp := odata.Response[entities.ODataEmployee]{Value: employees}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}, nil)
	defer server.Close()

	uc, _, empRepo, _, _ := newSyncUCWithOData(t, client)
	ctx := context.Background()

	// Pre-create an existing employee with different data
	existing := entities.NewExternalEmployee("emp-1", "C1")
	existing.FirstName = testFirstNameJohn
	existing.LastName = testLastNameDoe
	existing.ExternalDataHash = testOldHash
	_ = empRepo.Create(ctx, existing)

	req := &dto.StartSyncRequest{
		EntityType: entities.SyncEntityEmployee,
		Direction:  entities.SyncDirectionImport,
	}
	result, err := uc.StartSync(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 1, result.SuccessCount)
}

func TestStartSync_EmployeeSync_UpdateExistingLinked_CreatesConflict(t *testing.T) {
	employees := []entities.ODataEmployee{
		{RefKey: "emp-1", Code: "C1", FirstName: "John-Updated", LastName: "Doe", Email: "new@example.com"},
	}
	server, client := newTestODataServer(t, func(w http.ResponseWriter, _ *http.Request) {
		resp := odata.Response[entities.ODataEmployee]{Value: employees}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}, nil)
	defer server.Close()

	uc, _, empRepo, _, conflictRepo := newSyncUCWithOData(t, client)
	ctx := context.Background()

	// Pre-create an existing linked employee with different data
	existing := entities.NewExternalEmployee("emp-1", "C1")
	existing.FirstName = testFirstNameJohn
	existing.LastName = testLastNameDoe
	existing.Email = testOldEmail
	existing.ExternalDataHash = testOldHash
	localUserID := int64(42)
	existing.LocalUserID = &localUserID
	_ = empRepo.Create(ctx, existing)

	req := &dto.StartSyncRequest{
		EntityType: entities.SyncEntityEmployee,
		Direction:  entities.SyncDirectionImport,
	}
	result, err := uc.StartSync(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 1, result.ConflictCount)

	// Verify conflict was created
	conflicts, _ := conflictRepo.GetBySyncLogID(ctx, result.SyncLogID)
	assert.Len(t, conflicts, 1)
}

func TestStartSync_EmployeeSync_ConflictCreateError(t *testing.T) {
	employees := []entities.ODataEmployee{
		{RefKey: "emp-1", Code: "C1", FirstName: "John-Updated", LastName: "Doe", Email: "new@example.com"},
	}
	server, client := newTestODataServer(t, func(w http.ResponseWriter, _ *http.Request) {
		resp := odata.Response[entities.ODataEmployee]{Value: employees}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}, nil)
	defer server.Close()

	uc, _, empRepo, _, conflictRepo := newSyncUCWithOData(t, client)
	conflictRepo.createErr = true
	ctx := context.Background()

	// Pre-create existing linked employee with different data
	existing := entities.NewExternalEmployee("emp-1", "C1")
	existing.FirstName = testFirstNameJohn
	existing.LastName = testLastNameDoe
	existing.Email = testOldEmail
	existing.ExternalDataHash = testOldHash
	localUserID := int64(42)
	existing.LocalUserID = &localUserID
	_ = empRepo.Create(ctx, existing)

	req := &dto.StartSyncRequest{
		EntityType: entities.SyncEntityEmployee,
		Direction:  entities.SyncDirectionImport,
	}
	// Should still succeed - conflict creation error is non-fatal
	result, err := uc.StartSync(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestStartSync_EmployeeSync_GetByExternalIDError(t *testing.T) {
	employees := []entities.ODataEmployee{
		{RefKey: "emp-1", Code: "C1", FirstName: "John", LastName: "Doe"},
	}
	server, client := newTestODataServer(t, func(w http.ResponseWriter, _ *http.Request) {
		resp := odata.Response[entities.ODataEmployee]{Value: employees}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}, nil)
	defer server.Close()

	uc, _, empRepo, _, _ := newSyncUCWithOData(t, client)
	empRepo.getByExternalIDErr = true
	ctx := context.Background()

	req := &dto.StartSyncRequest{
		EntityType: entities.SyncEntityEmployee,
		Direction:  entities.SyncDirectionImport,
	}
	result, err := uc.StartSync(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 1, result.ErrorCount)
	assert.Equal(t, 0, result.SuccessCount)
}

func TestStartSync_EmployeeSync_CreateEmployeeError(t *testing.T) {
	employees := []entities.ODataEmployee{
		{RefKey: "emp-new", Code: "CN", FirstName: "New", LastName: "Emp"},
	}
	server, client := newTestODataServer(t, func(w http.ResponseWriter, _ *http.Request) {
		resp := odata.Response[entities.ODataEmployee]{Value: employees}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}, nil)
	defer server.Close()

	uc, _, empRepo, _, _ := newSyncUCWithOData(t, client)
	empRepo.createErr = true
	ctx := context.Background()

	req := &dto.StartSyncRequest{
		EntityType: entities.SyncEntityEmployee,
		Direction:  entities.SyncDirectionImport,
	}
	result, err := uc.StartSync(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 1, result.ErrorCount)
	assert.Equal(t, 0, result.SuccessCount)
}

func TestStartSync_EmployeeSync_UpdateEmployeeError(t *testing.T) {
	employees := []entities.ODataEmployee{
		{RefKey: "emp-1", Code: "C1", FirstName: "Updated", LastName: "Doe"},
	}
	server, client := newTestODataServer(t, func(w http.ResponseWriter, _ *http.Request) {
		resp := odata.Response[entities.ODataEmployee]{Value: employees}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}, nil)
	defer server.Close()

	uc, _, empRepo, _, _ := newSyncUCWithOData(t, client)
	ctx := context.Background()

	// Pre-create existing with different hash
	existing := entities.NewExternalEmployee("emp-1", "C1")
	existing.FirstName = "Old"
	existing.LastName = testLastNameDoe
	existing.ExternalDataHash = testOldHash
	_ = empRepo.MockExternalEmployeeRepository.Create(ctx, existing)

	// Now enable update error
	empRepo.updateErr = true

	req := &dto.StartSyncRequest{
		EntityType: entities.SyncEntityEmployee,
		Direction:  entities.SyncDirectionImport,
	}
	result, err := uc.StartSync(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 1, result.ErrorCount)
}

func TestStartSync_EmployeeSync_MarkInactiveError(t *testing.T) {
	employees := []entities.ODataEmployee{
		{RefKey: "emp-1", Code: "C1", FirstName: "John", LastName: "Doe"},
	}
	server, client := newTestODataServer(t, func(w http.ResponseWriter, _ *http.Request) {
		resp := odata.Response[entities.ODataEmployee]{Value: employees}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}, nil)
	defer server.Close()

	uc, _, empRepo, _, _ := newSyncUCWithOData(t, client)
	empRepo.markInactiveErr = true
	ctx := context.Background()

	req := &dto.StartSyncRequest{
		EntityType: entities.SyncEntityEmployee,
		Direction:  entities.SyncDirectionImport,
	}
	// Should still succeed - mark inactive error is non-fatal
	result, err := uc.StartSync(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestStartSync_EmployeeSync_SameHash_NoUpdate(t *testing.T) {
	employees := []entities.ODataEmployee{
		{RefKey: "emp-1", Code: "C1", FirstName: "John", LastName: "Doe"},
	}
	server, client := newTestODataServer(t, func(w http.ResponseWriter, _ *http.Request) {
		resp := odata.Response[entities.ODataEmployee]{Value: employees}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}, nil)
	defer server.Close()

	uc, _, empRepo, _, _ := newSyncUCWithOData(t, client)
	ctx := context.Background()

	// Pre-create existing with same hash as what the sync will compute
	rawData, _ := json.Marshal(employees[0])
	hash := calculateHash(rawData)
	existing := entities.NewExternalEmployee("emp-1", "C1")
	existing.FirstName = testFirstNameJohn
	existing.LastName = testLastNameDoe
	existing.ExternalDataHash = hash
	_ = empRepo.Create(ctx, existing)

	req := &dto.StartSyncRequest{
		EntityType: entities.SyncEntityEmployee,
		Direction:  entities.SyncDirectionImport,
	}
	result, err := uc.StartSync(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, result)
	// Should count as success with no update needed
	assert.Equal(t, 1, result.SuccessCount)
}

func TestStartSync_StudentSync_Success_NewStudents(t *testing.T) {
	students := []entities.ODataStudent{
		{RefKey: "stud-1", Code: "S1", FirstName: "Alice", LastName: "Wonder", GroupName: "CS-101"},
		{RefKey: "stud-2", Code: "S2", FirstName: "Bob", LastName: "Builder", GroupName: "CS-102"},
	}
	server, client := newTestODataServer(t, nil, func(w http.ResponseWriter, _ *http.Request) {
		resp := odata.Response[entities.ODataStudent]{Value: students}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	uc, _, _, _, _ := newSyncUCWithOData(t, client)
	ctx := context.Background()

	req := &dto.StartSyncRequest{
		EntityType: entities.SyncEntityStudent,
		Direction:  entities.SyncDirectionImport,
	}
	result, err := uc.StartSync(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 2, result.TotalRecords)
	assert.Equal(t, 2, result.SuccessCount)
	assert.Equal(t, "Student sync completed", result.Message)
}

func TestStartSync_StudentSync_ODataFetchError(t *testing.T) {
	server, client := newTestODataServer(t, nil, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad request"))
	})
	defer server.Close()

	uc, _, _, _, _ := newSyncUCWithOData(t, client)
	ctx := context.Background()

	req := &dto.StartSyncRequest{
		EntityType: entities.SyncEntityStudent,
		Direction:  entities.SyncDirectionImport,
	}
	result, err := uc.StartSync(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to fetch students from 1C")
}

func TestStartSync_StudentSync_GetAllExternalIDsError(t *testing.T) {
	students := []entities.ODataStudent{
		{RefKey: "stud-1", Code: "S1", FirstName: "Alice", LastName: "Wonder"},
	}
	server, client := newTestODataServer(t, nil, func(w http.ResponseWriter, _ *http.Request) {
		resp := odata.Response[entities.ODataStudent]{Value: students}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	uc, _, _, studRepo, _ := newSyncUCWithOData(t, client)
	studRepo.getAllExternalIDsErr = true
	ctx := context.Background()

	req := &dto.StartSyncRequest{
		EntityType: entities.SyncEntityStudent,
		Direction:  entities.SyncDirectionImport,
	}
	result, err := uc.StartSync(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get existing student IDs")
}

func TestStartSync_StudentSync_UpdateExistingLinked_CreatesConflict(t *testing.T) {
	students := []entities.ODataStudent{
		{RefKey: "stud-1", Code: "S1", FirstName: "Alice-Updated", LastName: "Wonder", GroupName: "CS-999", Email: "new@example.com", Status: "graduated"},
	}
	server, client := newTestODataServer(t, nil, func(w http.ResponseWriter, _ *http.Request) {
		resp := odata.Response[entities.ODataStudent]{Value: students}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	uc, _, _, studRepo, conflictRepo := newSyncUCWithOData(t, client)
	ctx := context.Background()

	// Pre-create existing linked student with different data
	existing := entities.NewExternalStudent("stud-1", "S1")
	existing.FirstName = testFirstNameAlice
	existing.LastName = testLastNameWonder
	existing.GroupName = "CS-101"
	existing.Email = testOldEmail
	existing.Status = "enrolled"
	existing.ExternalDataHash = testOldHash
	localUserID := int64(42)
	existing.LocalUserID = &localUserID
	_ = studRepo.Create(ctx, existing)

	req := &dto.StartSyncRequest{
		EntityType: entities.SyncEntityStudent,
		Direction:  entities.SyncDirectionImport,
	}
	result, err := uc.StartSync(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 1, result.ConflictCount)

	// Verify conflict was created with fields
	conflicts, _ := conflictRepo.GetBySyncLogID(ctx, result.SyncLogID)
	assert.Len(t, conflicts, 1)
}

func TestStartSync_StudentSync_GetByExternalIDError(t *testing.T) {
	students := []entities.ODataStudent{
		{RefKey: "stud-1", Code: "S1", FirstName: "Alice", LastName: "Wonder"},
	}
	server, client := newTestODataServer(t, nil, func(w http.ResponseWriter, _ *http.Request) {
		resp := odata.Response[entities.ODataStudent]{Value: students}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	uc, _, _, studRepo, _ := newSyncUCWithOData(t, client)
	studRepo.getByExternalIDErr = true
	ctx := context.Background()

	req := &dto.StartSyncRequest{
		EntityType: entities.SyncEntityStudent,
		Direction:  entities.SyncDirectionImport,
	}
	result, err := uc.StartSync(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 1, result.ErrorCount)
}

func TestStartSync_StudentSync_CreateStudentError(t *testing.T) {
	students := []entities.ODataStudent{
		{RefKey: "stud-new", Code: "SN", FirstName: "New", LastName: "Student"},
	}
	server, client := newTestODataServer(t, nil, func(w http.ResponseWriter, _ *http.Request) {
		resp := odata.Response[entities.ODataStudent]{Value: students}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	uc, _, _, studRepo, _ := newSyncUCWithOData(t, client)
	studRepo.createErr = true
	ctx := context.Background()

	req := &dto.StartSyncRequest{
		EntityType: entities.SyncEntityStudent,
		Direction:  entities.SyncDirectionImport,
	}
	result, err := uc.StartSync(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 1, result.ErrorCount)
}

func TestStartSync_StudentSync_UpdateStudentError(t *testing.T) {
	students := []entities.ODataStudent{
		{RefKey: "stud-1", Code: "S1", FirstName: "Updated", LastName: "Student"},
	}
	server, client := newTestODataServer(t, nil, func(w http.ResponseWriter, _ *http.Request) {
		resp := odata.Response[entities.ODataStudent]{Value: students}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	uc, _, _, studRepo, _ := newSyncUCWithOData(t, client)
	ctx := context.Background()

	// Pre-create existing with different hash
	existing := entities.NewExternalStudent("stud-1", "S1")
	existing.FirstName = "Old"
	existing.LastName = "Student"
	existing.ExternalDataHash = testOldHash
	_ = studRepo.MockExternalStudentRepository.Create(ctx, existing)

	studRepo.updateErr = true

	req := &dto.StartSyncRequest{
		EntityType: entities.SyncEntityStudent,
		Direction:  entities.SyncDirectionImport,
	}
	result, err := uc.StartSync(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 1, result.ErrorCount)
}

func TestStartSync_StudentSync_MarkInactiveError(t *testing.T) {
	students := []entities.ODataStudent{
		{RefKey: "stud-1", Code: "S1", FirstName: "Alice", LastName: "Wonder"},
	}
	server, client := newTestODataServer(t, nil, func(w http.ResponseWriter, _ *http.Request) {
		resp := odata.Response[entities.ODataStudent]{Value: students}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	uc, _, _, studRepo, _ := newSyncUCWithOData(t, client)
	studRepo.markInactiveErr = true
	ctx := context.Background()

	req := &dto.StartSyncRequest{
		EntityType: entities.SyncEntityStudent,
		Direction:  entities.SyncDirectionImport,
	}
	// Non-fatal error
	result, err := uc.StartSync(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestStartSync_StudentSync_ConflictCreateError(t *testing.T) {
	students := []entities.ODataStudent{
		{RefKey: "stud-1", Code: "S1", FirstName: "Updated", LastName: "Wonder", Email: "new@x.com"},
	}
	server, client := newTestODataServer(t, nil, func(w http.ResponseWriter, _ *http.Request) {
		resp := odata.Response[entities.ODataStudent]{Value: students}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	uc, _, _, studRepo, conflictRepo := newSyncUCWithOData(t, client)
	conflictRepo.createErr = true
	ctx := context.Background()

	existing := entities.NewExternalStudent("stud-1", "S1")
	existing.FirstName = testFirstNameAlice
	existing.LastName = testLastNameWonder
	existing.Email = "old@x.com"
	existing.ExternalDataHash = testOldHash
	localUserID := int64(42)
	existing.LocalUserID = &localUserID
	_ = studRepo.Create(ctx, existing)

	req := &dto.StartSyncRequest{
		EntityType: entities.SyncEntityStudent,
		Direction:  entities.SyncDirectionImport,
	}
	// Non-fatal error
	result, err := uc.StartSync(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestStartSync_StudentSync_SameHash_NoUpdate(t *testing.T) {
	students := []entities.ODataStudent{
		{RefKey: "stud-1", Code: "S1", FirstName: "Alice", LastName: "Wonder"},
	}
	server, client := newTestODataServer(t, nil, func(w http.ResponseWriter, _ *http.Request) {
		resp := odata.Response[entities.ODataStudent]{Value: students}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	uc, _, _, studRepo, _ := newSyncUCWithOData(t, client)
	ctx := context.Background()

	rawData, _ := json.Marshal(students[0])
	hash := calculateHash(rawData)
	existing := entities.NewExternalStudent("stud-1", "S1")
	existing.FirstName = testFirstNameAlice
	existing.LastName = testLastNameWonder
	existing.ExternalDataHash = hash
	_ = studRepo.Create(ctx, existing)

	req := &dto.StartSyncRequest{
		EntityType: entities.SyncEntityStudent,
		Direction:  entities.SyncDirectionImport,
	}
	result, err := uc.StartSync(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 1, result.SuccessCount)
}

func TestStartSync_UpdateSyncLogError(t *testing.T) {
	employees := []entities.ODataEmployee{
		{RefKey: "emp-1", Code: "C1", FirstName: "John", LastName: "Doe"},
	}
	server, client := newTestODataServer(t, func(w http.ResponseWriter, _ *http.Request) {
		resp := odata.Response[entities.ODataEmployee]{Value: employees}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}, nil)
	defer server.Close()

	uc, syncLogRepo, _, _, _ := newSyncUCWithOData(t, client)
	ctx := context.Background()

	req := &dto.StartSyncRequest{
		EntityType: entities.SyncEntityEmployee,
		Direction:  entities.SyncDirectionImport,
	}

	// Start sync first to create the log, then enable update error
	// We can't easily test this since update is called mid-sync;
	// just set updateErr after the Create succeeds
	syncLogRepo.updateErr = true

	result, err := uc.StartSync(ctx, req)
	// The sync itself should still succeed even if sync log update fails
	require.NoError(t, err)
	require.NotNil(t, result)
}

// --- GetSyncLogs ---

func TestGetSyncLogs_Success(t *testing.T) {
	server, client := newTestODataServer(t, nil, nil)
	defer server.Close()
	uc, syncLogRepo, _, _, _ := newSyncUCWithOData(t, client)
	ctx := context.Background()

	// Create sync logs
	log1 := entities.NewSyncLog(entities.SyncEntityEmployee, entities.SyncDirectionImport)
	log1.Start()
	_ = syncLogRepo.Create(ctx, log1)

	log2 := entities.NewSyncLog(entities.SyncEntityStudent, entities.SyncDirectionImport)
	log2.Start()
	_ = syncLogRepo.Create(ctx, log2)

	req := &dto.SyncListRequest{Limit: 10}
	result, err := uc.GetSyncLogs(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, int64(2), result.Total)
	assert.Len(t, result.Items, 2)
}

func TestGetSyncLogs_WithFilter(t *testing.T) {
	server, client := newTestODataServer(t, nil, nil)
	defer server.Close()
	uc, syncLogRepo, _, _, _ := newSyncUCWithOData(t, client)
	ctx := context.Background()

	log1 := entities.NewSyncLog(entities.SyncEntityEmployee, entities.SyncDirectionImport)
	_ = syncLogRepo.Create(ctx, log1)

	log2 := entities.NewSyncLog(entities.SyncEntityStudent, entities.SyncDirectionImport)
	_ = syncLogRepo.Create(ctx, log2)

	entityType := entities.SyncEntityEmployee
	req := &dto.SyncListRequest{EntityType: &entityType, Limit: 10}
	result, err := uc.GetSyncLogs(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.Total)
}

func TestGetSyncLogs_Error(t *testing.T) {
	server, client := newTestODataServer(t, nil, nil)
	defer server.Close()
	uc, syncLogRepo, _, _, _ := newSyncUCWithOData(t, client)
	syncLogRepo.listErr = true

	req := &dto.SyncListRequest{Limit: 10}
	result, err := uc.GetSyncLogs(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to list sync logs")
}

// --- GetSyncLog ---

func TestGetSyncLog_Success(t *testing.T) {
	server, client := newTestODataServer(t, nil, nil)
	defer server.Close()
	uc, syncLogRepo, _, _, _ := newSyncUCWithOData(t, client)
	ctx := context.Background()

	log := entities.NewSyncLog(entities.SyncEntityEmployee, entities.SyncDirectionImport)
	_ = syncLogRepo.Create(ctx, log)

	result, err := uc.GetSyncLog(ctx, log.ID)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, log.ID, result.ID)
}

func TestGetSyncLog_NotFound(t *testing.T) {
	server, client := newTestODataServer(t, nil, nil)
	defer server.Close()
	uc, _, _, _, _ := newSyncUCWithOData(t, client)

	result, err := uc.GetSyncLog(context.Background(), 999)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestGetSyncLog_Error(t *testing.T) {
	server, client := newTestODataServer(t, nil, nil)
	defer server.Close()
	uc, syncLogRepo, _, _, _ := newSyncUCWithOData(t, client)
	syncLogRepo.getByIDErr = true

	result, err := uc.GetSyncLog(context.Background(), 1)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get sync log")
}

// --- GetSyncStats ---

func TestGetSyncStats_Success(t *testing.T) {
	server, client := newTestODataServer(t, nil, nil)
	defer server.Close()
	uc, _, _, _, _ := newSyncUCWithOData(t, client)

	result, err := uc.GetSyncStats(context.Background(), nil)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, int64(5), result.TotalSyncs)
	assert.Equal(t, int64(3), result.SuccessfulSyncs)
}

func TestGetSyncStats_WithEntityType(t *testing.T) {
	server, client := newTestODataServer(t, nil, nil)
	defer server.Close()
	uc, _, _, _, _ := newSyncUCWithOData(t, client)

	entityType := entities.SyncEntityEmployee
	result, err := uc.GetSyncStats(context.Background(), &entityType)
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestGetSyncStats_Error(t *testing.T) {
	server, client := newTestODataServer(t, nil, nil)
	defer server.Close()
	uc, syncLogRepo, _, _, _ := newSyncUCWithOData(t, client)
	syncLogRepo.getStatsErr = true

	result, err := uc.GetSyncStats(context.Background(), nil)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get sync stats")
}

// --- IsSyncRunning ---

func TestIsSyncRunning(t *testing.T) {
	server, client := newTestODataServer(t, nil, nil)
	defer server.Close()
	uc, _, _, _, _ := newSyncUCWithOData(t, client)

	assert.False(t, uc.IsSyncRunning(entities.SyncEntityEmployee))

	uc.mu.Lock()
	uc.running[entities.SyncEntityEmployee] = true
	uc.mu.Unlock()

	assert.True(t, uc.IsSyncRunning(entities.SyncEntityEmployee))
	assert.False(t, uc.IsSyncRunning(entities.SyncEntityStudent))
}

// --- CancelSync ---

func TestCancelSync_Success(t *testing.T) {
	server, client := newTestODataServer(t, nil, nil)
	defer server.Close()
	uc, syncLogRepo, _, _, _ := newSyncUCWithOData(t, client)
	ctx := context.Background()

	log := entities.NewSyncLog(entities.SyncEntityEmployee, entities.SyncDirectionImport)
	log.Start() // Set to in_progress
	_ = syncLogRepo.Create(ctx, log)

	err := uc.CancelSync(ctx, log.ID)
	require.NoError(t, err)

	// Verify canceled
	canceled, _ := syncLogRepo.GetByID(ctx, log.ID)
	assert.Equal(t, entities.SyncStatusCancelled, canceled.Status)
}

func TestCancelSync_NotFound(t *testing.T) {
	server, client := newTestODataServer(t, nil, nil)
	defer server.Close()
	uc, _, _, _, _ := newSyncUCWithOData(t, client)

	err := uc.CancelSync(context.Background(), 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sync log not found")
}

func TestCancelSync_NotRunning(t *testing.T) {
	server, client := newTestODataServer(t, nil, nil)
	defer server.Close()
	uc, syncLogRepo, _, _, _ := newSyncUCWithOData(t, client)
	ctx := context.Background()

	log := entities.NewSyncLog(entities.SyncEntityEmployee, entities.SyncDirectionImport)
	log.Complete() // Set to completed, not running
	_ = syncLogRepo.Create(ctx, log)

	err := uc.CancelSync(ctx, log.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sync is not running")
}

func TestCancelSync_GetByIDError(t *testing.T) {
	server, client := newTestODataServer(t, nil, nil)
	defer server.Close()
	uc, syncLogRepo, _, _, _ := newSyncUCWithOData(t, client)
	syncLogRepo.getByIDErr = true

	err := uc.CancelSync(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get sync log")
}

func TestCancelSync_UpdateError(t *testing.T) {
	server, client := newTestODataServer(t, nil, nil)
	defer server.Close()
	uc, syncLogRepo, _, _, _ := newSyncUCWithOData(t, client)
	ctx := context.Background()

	log := entities.NewSyncLog(entities.SyncEntityEmployee, entities.SyncDirectionImport)
	log.Start()
	_ = syncLogRepo.Create(ctx, log)

	syncLogRepo.updateErr = true

	err := uc.CancelSync(ctx, log.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update sync log")
}

// --- Ping ---

func TestPing_Success(t *testing.T) {
	server, client := newTestODataServer(t, nil, nil)
	defer server.Close()
	uc, _, _, _, _ := newSyncUCWithOData(t, client)

	err := uc.Ping(context.Background())
	assert.NoError(t, err)
}

func TestPing_Error(t *testing.T) {
	// Use a server that is already closed
	server, client := newTestODataServer(t, nil, nil)
	server.Close()
	uc, _, _, _, _ := newSyncUCWithOData(t, client)

	err := uc.Ping(context.Background())
	assert.Error(t, err)
}

// --- calculateHash ---

func TestCalculateHash(t *testing.T) {
	data := []byte("test data")
	hash1 := calculateHash(data)
	hash2 := calculateHash(data)
	assert.Equal(t, hash1, hash2)
	assert.Len(t, hash1, 64) // SHA256 hex

	differentHash := calculateHash([]byte("different data"))
	assert.NotEqual(t, hash1, differentHash)
}

// --- detectEmployeeConflict ---

func TestDetectEmployeeConflict_WithFields(t *testing.T) {
	server, client := newTestODataServer(t, nil, nil)
	defer server.Close()
	uc, _, _, _, _ := newSyncUCWithOData(t, client)

	existing := &entities.ExternalEmployee{
		ExternalID: "emp-1",
		FirstName:  "John",
		LastName:   "Doe",
		Email:      "john@old.com",
		Position:   "Developer",
		Department: "Engineering",
	}
	localUserID := int64(42)
	existing.LocalUserID = &localUserID

	updated := &entities.ExternalEmployee{
		ExternalID: "emp-1",
		FirstName:  "Johnny",
		LastName:   "Doe-Updated",
		Email:      "john@new.com",
		Position:   "Senior Developer",
		Department: "R&D",
	}

	conflict := uc.detectEmployeeConflict(existing, updated, 1)
	require.NotNil(t, conflict)
	assert.Equal(t, entities.ConflictTypeUpdate, conflict.ConflictType)
	assert.Contains(t, conflict.ConflictFields, "first_name")
	assert.Contains(t, conflict.ConflictFields, "last_name")
	assert.Contains(t, conflict.ConflictFields, "email")
	assert.Contains(t, conflict.ConflictFields, "position")
	assert.Contains(t, conflict.ConflictFields, "department")
	assert.Len(t, conflict.ConflictFields, 5)
	assert.NotEmpty(t, conflict.LocalData)
	assert.NotEmpty(t, conflict.ExternalData)
}

func TestDetectEmployeeConflict_NoConflict(t *testing.T) {
	server, client := newTestODataServer(t, nil, nil)
	defer server.Close()
	uc, _, _, _, _ := newSyncUCWithOData(t, client)

	existing := &entities.ExternalEmployee{
		ExternalID: "emp-1",
		FirstName:  "John",
		LastName:   "Doe",
		Email:      "john@example.com",
		Position:   "Developer",
		Department: "Engineering",
	}
	updated := &entities.ExternalEmployee{
		ExternalID: "emp-1",
		FirstName:  "John",
		LastName:   "Doe",
		Email:      "john@example.com",
		Position:   "Developer",
		Department: "Engineering",
	}

	conflict := uc.detectEmployeeConflict(existing, updated, 1)
	assert.Nil(t, conflict)
}

func TestDetectEmployeeConflict_PartialFields(t *testing.T) {
	server, client := newTestODataServer(t, nil, nil)
	defer server.Close()
	uc, _, _, _, _ := newSyncUCWithOData(t, client)

	existing := &entities.ExternalEmployee{
		ExternalID: "emp-1",
		FirstName:  "John",
		LastName:   "Doe",
		Email:      "john@old.com",
		Position:   "Developer",
		Department: "Engineering",
	}
	updated := &entities.ExternalEmployee{
		ExternalID: "emp-1",
		FirstName:  "John",
		LastName:   "Doe",
		Email:      "john@new.com",
		Position:   "Developer",
		Department: "Engineering",
	}

	conflict := uc.detectEmployeeConflict(existing, updated, 1)
	require.NotNil(t, conflict)
	assert.Equal(t, []string{"email"}, conflict.ConflictFields)
}

// --- detectStudentConflict ---

func TestDetectStudentConflict_WithFields(t *testing.T) {
	server, client := newTestODataServer(t, nil, nil)
	defer server.Close()
	uc, _, _, _, _ := newSyncUCWithOData(t, client)

	existing := &entities.ExternalStudent{
		ExternalID: "stud-1",
		FirstName:  "Alice",
		LastName:   "Wonder",
		Email:      "alice@old.com",
		GroupName:  "CS-101",
		Status:     "enrolled",
	}
	updated := &entities.ExternalStudent{
		ExternalID: "stud-1",
		FirstName:  "Alice-Updated",
		LastName:   "Wonderland",
		Email:      "alice@new.com",
		GroupName:  "CS-202",
		Status:     "graduated",
	}

	conflict := uc.detectStudentConflict(existing, updated, 1)
	require.NotNil(t, conflict)
	assert.Equal(t, entities.ConflictTypeUpdate, conflict.ConflictType)
	assert.Contains(t, conflict.ConflictFields, "first_name")
	assert.Contains(t, conflict.ConflictFields, "last_name")
	assert.Contains(t, conflict.ConflictFields, "email")
	assert.Contains(t, conflict.ConflictFields, "group_name")
	assert.Contains(t, conflict.ConflictFields, "status")
	assert.Len(t, conflict.ConflictFields, 5)
}

func TestDetectStudentConflict_NoConflict(t *testing.T) {
	server, client := newTestODataServer(t, nil, nil)
	defer server.Close()
	uc, _, _, _, _ := newSyncUCWithOData(t, client)

	existing := &entities.ExternalStudent{
		ExternalID: "stud-1",
		FirstName:  "Alice",
		LastName:   "Wonder",
		Email:      "alice@example.com",
		GroupName:  "CS-101",
		Status:     "enrolled",
	}
	updated := &entities.ExternalStudent{
		ExternalID: "stud-1",
		FirstName:  "Alice",
		LastName:   "Wonder",
		Email:      "alice@example.com",
		GroupName:  "CS-101",
		Status:     "enrolled",
	}

	conflict := uc.detectStudentConflict(existing, updated, 1)
	assert.Nil(t, conflict)
}

func TestDetectStudentConflict_PartialFields(t *testing.T) {
	server, client := newTestODataServer(t, nil, nil)
	defer server.Close()
	uc, _, _, _, _ := newSyncUCWithOData(t, client)

	existing := &entities.ExternalStudent{
		ExternalID: "stud-1",
		FirstName:  "Alice",
		LastName:   "Wonder",
		Email:      "alice@example.com",
		GroupName:  "CS-101",
		Status:     "enrolled",
	}
	updated := &entities.ExternalStudent{
		ExternalID: "stud-1",
		FirstName:  "Alice",
		LastName:   "Wonder",
		Email:      "alice@example.com",
		GroupName:  "CS-202",
		Status:     "enrolled",
	}

	conflict := uc.detectStudentConflict(existing, updated, 1)
	require.NotNil(t, conflict)
	assert.Equal(t, []string{"group_name"}, conflict.ConflictFields)
}

// --- Test UpdateSyncLog error after failed sync ---

func TestStartSync_FailedSync_UpdateLogError(t *testing.T) {
	// OData server returns error to trigger sync failure
	server, client := newTestODataServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprint(w, "error")
	}, nil)
	defer server.Close()

	uc, syncLogRepo, _, _, _ := newSyncUCWithOData(t, client)
	// After the sync log is created, enable update error
	syncLogRepo.updateErr = true

	ctx := context.Background()
	req := &dto.StartSyncRequest{
		EntityType: entities.SyncEntityEmployee,
		Direction:  entities.SyncDirectionImport,
	}
	result, err := uc.StartSync(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, result)
}

// --- Employee sync with existing and same hash (no conflict, no update) ---

func TestStartSync_EmployeeSync_ExistingLinkedSameHash(t *testing.T) {
	employees := []entities.ODataEmployee{
		{RefKey: "emp-1", Code: "C1", FirstName: "John", LastName: "Doe"},
	}
	server, client := newTestODataServer(t, func(w http.ResponseWriter, _ *http.Request) {
		resp := odata.Response[entities.ODataEmployee]{Value: employees}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}, nil)
	defer server.Close()

	uc, _, empRepo, _, conflictRepo := newSyncUCWithOData(t, client)
	ctx := context.Background()

	rawData, _ := json.Marshal(employees[0])
	hash := calculateHash(rawData)
	existing := entities.NewExternalEmployee("emp-1", "C1")
	existing.FirstName = testFirstNameJohn
	existing.LastName = testLastNameDoe
	existing.ExternalDataHash = hash
	localUserID := int64(42)
	existing.LocalUserID = &localUserID
	_ = empRepo.Create(ctx, existing)

	req := &dto.StartSyncRequest{
		EntityType: entities.SyncEntityEmployee,
		Direction:  entities.SyncDirectionImport,
	}
	result, err := uc.StartSync(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 0, result.ConflictCount)

	conflicts, _ := conflictRepo.GetBySyncLogID(ctx, result.SyncLogID)
	assert.Len(t, conflicts, 0)
}
