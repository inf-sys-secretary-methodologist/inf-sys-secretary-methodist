package usecases

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/infrastructure/odata"
)

// SyncUseCase handles synchronization operations with 1C
type SyncUseCase struct {
	odataClient  *odata.Client
	syncLogRepo  repositories.SyncLogRepository
	employeeRepo repositories.ExternalEmployeeRepository
	studentRepo  repositories.ExternalStudentRepository
	conflictRepo repositories.SyncConflictRepository
	logger       *slog.Logger
	mu           sync.Mutex
	running      map[entities.SyncEntityType]bool
}

// NewSyncUseCase creates a new sync use case
func NewSyncUseCase(
	odataClient *odata.Client,
	syncLogRepo repositories.SyncLogRepository,
	employeeRepo repositories.ExternalEmployeeRepository,
	studentRepo repositories.ExternalStudentRepository,
	conflictRepo repositories.SyncConflictRepository,
	logger *slog.Logger,
) *SyncUseCase {
	return &SyncUseCase{
		odataClient:  odataClient,
		syncLogRepo:  syncLogRepo,
		employeeRepo: employeeRepo,
		studentRepo:  studentRepo,
		conflictRepo: conflictRepo,
		logger:       logger,
		running:      make(map[entities.SyncEntityType]bool),
	}
}

// StartSync starts a synchronization operation
func (uc *SyncUseCase) StartSync(ctx context.Context, req *dto.StartSyncRequest) (*dto.SyncResultDTO, error) {
	// Check if sync is already running for this entity type
	uc.mu.Lock()
	if uc.running[req.EntityType] && !req.Force {
		uc.mu.Unlock()
		return nil, fmt.Errorf("sync is already running for %s", req.EntityType)
	}
	uc.running[req.EntityType] = true
	uc.mu.Unlock()

	defer func() {
		uc.mu.Lock()
		uc.running[req.EntityType] = false
		uc.mu.Unlock()
	}()

	// Create sync log
	syncLog := entities.NewSyncLog(req.EntityType, req.Direction)
	syncLog.Start()

	if err := uc.syncLogRepo.Create(ctx, syncLog); err != nil {
		return nil, fmt.Errorf("failed to create sync log: %w", err)
	}

	uc.logger.Info("Starting sync",
		slog.Int64("sync_log_id", syncLog.ID),
		slog.String("entity_type", string(req.EntityType)),
		slog.String("direction", string(req.Direction)))

	var result *dto.SyncResultDTO
	var err error

	switch req.EntityType {
	case entities.SyncEntityEmployee:
		result, err = uc.syncEmployees(ctx, syncLog)
	case entities.SyncEntityStudent:
		result, err = uc.syncStudents(ctx, syncLog)
	default:
		err = fmt.Errorf("unsupported entity type: %s", req.EntityType)
	}

	// Update sync log with final status
	if err != nil {
		syncLog.Fail(err.Error())
		uc.logger.Error("Sync failed",
			slog.Int64("sync_log_id", syncLog.ID),
			slog.String("error", err.Error()))
	} else {
		syncLog.Complete()
		uc.logger.Info("Sync completed",
			slog.Int64("sync_log_id", syncLog.ID),
			slog.Int("total", syncLog.TotalRecords),
			slog.Int("success", syncLog.SuccessCount),
			slog.Int("errors", syncLog.ErrorCount),
			slog.Int("conflicts", syncLog.ConflictCount))
	}

	if updateErr := uc.syncLogRepo.Update(ctx, syncLog); updateErr != nil {
		uc.logger.Error("Failed to update sync log", slog.String("error", updateErr.Error()))
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

// syncEmployees synchronizes employee data from 1C
func (uc *SyncUseCase) syncEmployees(ctx context.Context, syncLog *entities.SyncLog) (*dto.SyncResultDTO, error) {
	// Fetch employees from 1C
	odataEmployees, err := uc.odataClient.GetAllEmployees(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch employees from 1C: %w", err)
	}

	syncLog.SetTotalRecords(len(odataEmployees))
	_ = uc.syncLogRepo.Update(ctx, syncLog)

	// Get existing external IDs
	existingIDs, err := uc.employeeRepo.GetAllExternalIDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing employee IDs: %w", err)
	}
	existingMap := make(map[string]bool)
	for _, id := range existingIDs {
		existingMap[id] = true
	}

	// Track active external IDs
	activeExternalIDs := make([]string, 0, len(odataEmployees))

	// Process each employee
	for _, odataEmp := range odataEmployees {
		syncLog.IncrementProcessed()

		// Convert to entity
		emp := odataEmp.ToExternalEmployee()
		emp.LastSyncAt = time.Now()

		// Calculate hash for change detection
		rawData, _ := json.Marshal(odataEmp)
		emp.RawData = string(rawData)
		emp.ExternalDataHash = calculateHash(rawData)

		// Check if exists
		existing, err := uc.employeeRepo.GetByExternalID(ctx, emp.ExternalID)
		if err != nil {
			uc.logger.Error("Failed to check existing employee",
				slog.String("external_id", emp.ExternalID),
				slog.String("error", err.Error()))
			syncLog.IncrementError()
			continue
		}

		if existing != nil {
			// Check for changes
			if existing.ExternalDataHash != emp.ExternalDataHash {
				// Check for conflicts (if linked to local user and data differs)
				if existing.IsLinked() {
					conflict := uc.detectEmployeeConflict(existing, emp, syncLog.ID)
					if conflict != nil {
						if err := uc.conflictRepo.Create(ctx, conflict); err != nil {
							uc.logger.Error("Failed to create conflict",
								slog.String("error", err.Error()))
						}
						syncLog.IncrementConflict()
					}
				}

				// Update employee
				emp.ID = existing.ID
				emp.LocalUserID = existing.LocalUserID
				if err := uc.employeeRepo.Update(ctx, emp); err != nil {
					uc.logger.Error("Failed to update employee",
						slog.String("external_id", emp.ExternalID),
						slog.String("error", err.Error()))
					syncLog.IncrementError()
					continue
				}
			}
		} else {
			// Create new employee
			if err := uc.employeeRepo.Create(ctx, emp); err != nil {
				uc.logger.Error("Failed to create employee",
					slog.String("external_id", emp.ExternalID),
					slog.String("error", err.Error()))
				syncLog.IncrementError()
				continue
			}
		}

		activeExternalIDs = append(activeExternalIDs, emp.ExternalID)
		syncLog.IncrementSuccess()
	}

	// Mark employees not in current sync as inactive
	if err := uc.employeeRepo.MarkInactiveExcept(ctx, activeExternalIDs); err != nil {
		uc.logger.Error("Failed to mark inactive employees", slog.String("error", err.Error()))
	}

	return &dto.SyncResultDTO{
		SyncLogID:      syncLog.ID,
		Status:         string(syncLog.Status),
		TotalRecords:   syncLog.TotalRecords,
		ProcessedCount: syncLog.ProcessedCount,
		SuccessCount:   syncLog.SuccessCount,
		ErrorCount:     syncLog.ErrorCount,
		ConflictCount:  syncLog.ConflictCount,
		Message:        "Employee sync completed",
	}, nil
}

// syncStudents synchronizes student data from 1C
func (uc *SyncUseCase) syncStudents(ctx context.Context, syncLog *entities.SyncLog) (*dto.SyncResultDTO, error) {
	// Fetch students from 1C
	odataStudents, err := uc.odataClient.GetAllStudents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch students from 1C: %w", err)
	}

	syncLog.SetTotalRecords(len(odataStudents))
	_ = uc.syncLogRepo.Update(ctx, syncLog)

	// Get existing external IDs
	existingIDs, err := uc.studentRepo.GetAllExternalIDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing student IDs: %w", err)
	}
	existingMap := make(map[string]bool)
	for _, id := range existingIDs {
		existingMap[id] = true
	}

	// Track active external IDs
	activeExternalIDs := make([]string, 0, len(odataStudents))

	// Process each student
	for _, odataStudent := range odataStudents {
		syncLog.IncrementProcessed()

		// Convert to entity
		student := odataStudent.ToExternalStudent()
		student.LastSyncAt = time.Now()

		// Calculate hash for change detection
		rawData, _ := json.Marshal(odataStudent)
		student.RawData = string(rawData)
		student.ExternalDataHash = calculateHash(rawData)

		// Check if exists
		existing, err := uc.studentRepo.GetByExternalID(ctx, student.ExternalID)
		if err != nil {
			uc.logger.Error("Failed to check existing student",
				slog.String("external_id", student.ExternalID),
				slog.String("error", err.Error()))
			syncLog.IncrementError()
			continue
		}

		if existing != nil {
			// Check for changes
			if existing.ExternalDataHash != student.ExternalDataHash {
				// Check for conflicts (if linked to local user and data differs)
				if existing.IsLinked() {
					conflict := uc.detectStudentConflict(existing, student, syncLog.ID)
					if conflict != nil {
						if err := uc.conflictRepo.Create(ctx, conflict); err != nil {
							uc.logger.Error("Failed to create conflict",
								slog.String("error", err.Error()))
						}
						syncLog.IncrementConflict()
					}
				}

				// Update student
				student.ID = existing.ID
				student.LocalUserID = existing.LocalUserID
				if err := uc.studentRepo.Update(ctx, student); err != nil {
					uc.logger.Error("Failed to update student",
						slog.String("external_id", student.ExternalID),
						slog.String("error", err.Error()))
					syncLog.IncrementError()
					continue
				}
			}
		} else {
			// Create new student
			if err := uc.studentRepo.Create(ctx, student); err != nil {
				uc.logger.Error("Failed to create student",
					slog.String("external_id", student.ExternalID),
					slog.String("error", err.Error()))
				syncLog.IncrementError()
				continue
			}
		}

		activeExternalIDs = append(activeExternalIDs, student.ExternalID)
		syncLog.IncrementSuccess()
	}

	// Mark students not in current sync as inactive
	if err := uc.studentRepo.MarkInactiveExcept(ctx, activeExternalIDs); err != nil {
		uc.logger.Error("Failed to mark inactive students", slog.String("error", err.Error()))
	}

	return &dto.SyncResultDTO{
		SyncLogID:      syncLog.ID,
		Status:         string(syncLog.Status),
		TotalRecords:   syncLog.TotalRecords,
		ProcessedCount: syncLog.ProcessedCount,
		SuccessCount:   syncLog.SuccessCount,
		ErrorCount:     syncLog.ErrorCount,
		ConflictCount:  syncLog.ConflictCount,
		Message:        "Student sync completed",
	}, nil
}

// detectEmployeeConflict detects conflicts between existing and new employee data
func (uc *SyncUseCase) detectEmployeeConflict(existing, new *entities.ExternalEmployee, syncLogID int64) *entities.SyncConflict {
	var conflictFields []string

	if existing.FirstName != new.FirstName {
		conflictFields = append(conflictFields, "first_name")
	}
	if existing.LastName != new.LastName {
		conflictFields = append(conflictFields, "last_name")
	}
	if existing.Email != new.Email {
		conflictFields = append(conflictFields, "email")
	}
	if existing.Position != new.Position {
		conflictFields = append(conflictFields, "position")
	}
	if existing.Department != new.Department {
		conflictFields = append(conflictFields, "department")
	}

	if len(conflictFields) == 0 {
		return nil
	}

	conflict := entities.NewSyncConflict(syncLogID, entities.SyncEntityEmployee, existing.ExternalID)
	conflict.ConflictType = entities.ConflictTypeUpdate
	conflict.SetConflictFields(conflictFields)

	localData, _ := json.Marshal(existing)
	externalData, _ := json.Marshal(new)
	conflict.SetLocalData(string(localData))
	conflict.SetExternalData(string(externalData))

	return conflict
}

// detectStudentConflict detects conflicts between existing and new student data
func (uc *SyncUseCase) detectStudentConflict(existing, new *entities.ExternalStudent, syncLogID int64) *entities.SyncConflict {
	var conflictFields []string

	if existing.FirstName != new.FirstName {
		conflictFields = append(conflictFields, "first_name")
	}
	if existing.LastName != new.LastName {
		conflictFields = append(conflictFields, "last_name")
	}
	if existing.Email != new.Email {
		conflictFields = append(conflictFields, "email")
	}
	if existing.GroupName != new.GroupName {
		conflictFields = append(conflictFields, "group_name")
	}
	if existing.Status != new.Status {
		conflictFields = append(conflictFields, "status")
	}

	if len(conflictFields) == 0 {
		return nil
	}

	conflict := entities.NewSyncConflict(syncLogID, entities.SyncEntityStudent, existing.ExternalID)
	conflict.ConflictType = entities.ConflictTypeUpdate
	conflict.SetConflictFields(conflictFields)

	localData, _ := json.Marshal(existing)
	externalData, _ := json.Marshal(new)
	conflict.SetLocalData(string(localData))
	conflict.SetExternalData(string(externalData))

	return conflict
}

// GetSyncLogs retrieves sync logs with filtering
func (uc *SyncUseCase) GetSyncLogs(ctx context.Context, req *dto.SyncListRequest) (*dto.SyncListResponse, error) {
	filter := entities.SyncLogFilter{
		EntityType: req.EntityType,
		Direction:  req.Direction,
		Status:     req.Status,
		Limit:      req.Limit,
		Offset:     req.Offset,
	}

	logs, total, err := uc.syncLogRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list sync logs: %w", err)
	}

	items := make([]*dto.SyncLogDTO, len(logs))
	for i, log := range logs {
		items[i] = dto.FromSyncLog(log)
	}

	return &dto.SyncListResponse{
		Items: items,
		Total: total,
	}, nil
}

// GetSyncLog retrieves a single sync log by ID
func (uc *SyncUseCase) GetSyncLog(ctx context.Context, id int64) (*dto.SyncLogDTO, error) {
	log, err := uc.syncLogRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get sync log: %w", err)
	}
	if log == nil {
		return nil, nil
	}
	return dto.FromSyncLog(log), nil
}

// GetSyncStats retrieves sync statistics
func (uc *SyncUseCase) GetSyncStats(ctx context.Context, entityType *entities.SyncEntityType) (*dto.SyncStatsDTO, error) {
	stats, err := uc.syncLogRepo.GetStats(ctx, entityType)
	if err != nil {
		return nil, fmt.Errorf("failed to get sync stats: %w", err)
	}
	return dto.FromSyncStats(stats), nil
}

// IsSyncRunning checks if a sync is currently running
func (uc *SyncUseCase) IsSyncRunning(entityType entities.SyncEntityType) bool {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	return uc.running[entityType]
}

// CancelSync cancels a running sync (best effort)
func (uc *SyncUseCase) CancelSync(ctx context.Context, syncLogID int64) error {
	log, err := uc.syncLogRepo.GetByID(ctx, syncLogID)
	if err != nil {
		return fmt.Errorf("failed to get sync log: %w", err)
	}
	if log == nil {
		return fmt.Errorf("sync log not found")
	}

	if !log.IsRunning() {
		return fmt.Errorf("sync is not running")
	}

	log.Cancel()
	if err := uc.syncLogRepo.Update(ctx, log); err != nil {
		return fmt.Errorf("failed to update sync log: %w", err)
	}

	return nil
}

// Ping checks 1C connection
func (uc *SyncUseCase) Ping(ctx context.Context) error {
	return uc.odataClient.Ping(ctx)
}

// calculateHash calculates SHA256 hash of data
func calculateHash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
