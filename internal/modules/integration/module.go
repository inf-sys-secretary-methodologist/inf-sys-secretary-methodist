// Package integration provides 1C enterprise system integration module
package integration

import (
	"context"
	"database/sql"
	"log/slog"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/infrastructure/odata"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/infrastructure/persistence"
	integrationHttp "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration/interfaces/http"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/config"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

// Module represents the integration module with all its dependencies
type Module struct {
	config *config.IntegrationConfig
	logger *logging.Logger

	// Use cases
	syncUseCase     *usecases.SyncUseCase
	employeeUseCase *usecases.EmployeeUseCase
	studentUseCase  *usecases.StudentUseCase
	conflictUseCase *usecases.ConflictUseCase

	// Handlers
	syncHandler     *integrationHttp.SyncHandler
	employeeHandler *integrationHttp.EmployeeHandler
	studentHandler  *integrationHttp.StudentHandler
	conflictHandler *integrationHttp.ConflictHandler

	// Scheduler
	scheduler *cron.Cron
	mu        sync.RWMutex
}

// NewModule creates a new integration module
func NewModule(db *sql.DB, cfg *config.IntegrationConfig, logger *logging.Logger) (*Module, error) {
	if !cfg.Enabled {
		logger.Info("Integration module is disabled", nil)
		return &Module{config: cfg, logger: logger}, nil
	}

	// Initialize OData client
	odataConfig := &odata.Config{
		BaseURL:    cfg.BaseURL,
		Username:   cfg.Username,
		Password:   cfg.Password,
		Timeout:    cfg.Timeout,
		MaxRetries: cfg.MaxRetries,
		RetryDelay: cfg.RetryDelay,
	}
	odataClient := odata.NewClient(odataConfig)

	// Initialize repositories
	syncLogRepo := persistence.NewSyncLogRepositoryPg(db)
	employeeRepo := persistence.NewExternalEmployeeRepositoryPg(db)
	studentRepo := persistence.NewExternalStudentRepositoryPg(db)
	conflictRepo := persistence.NewSyncConflictRepositoryPg(db)

	// Initialize use cases
	slogLogger := slog.Default()

	syncUseCase := usecases.NewSyncUseCase(
		odataClient,
		syncLogRepo,
		employeeRepo,
		studentRepo,
		conflictRepo,
		slogLogger,
	)

	employeeUseCase := usecases.NewEmployeeUseCase(employeeRepo)
	studentUseCase := usecases.NewStudentUseCase(studentRepo)
	conflictUseCase := usecases.NewConflictUseCase(conflictRepo)

	// Initialize handlers
	syncHandler := integrationHttp.NewSyncHandler(syncUseCase)
	employeeHandler := integrationHttp.NewEmployeeHandler(employeeUseCase)
	studentHandler := integrationHttp.NewStudentHandler(studentUseCase)
	conflictHandler := integrationHttp.NewConflictHandler(conflictUseCase)

	module := &Module{
		config:          cfg,
		logger:          logger,
		syncUseCase:     syncUseCase,
		employeeUseCase: employeeUseCase,
		studentUseCase:  studentUseCase,
		conflictUseCase: conflictUseCase,
		syncHandler:     syncHandler,
		employeeHandler: employeeHandler,
		studentHandler:  studentHandler,
		conflictHandler: conflictHandler,
	}

	return module, nil
}

// RegisterRoutes registers all integration module routes
func (m *Module) RegisterRoutes(router *gin.RouterGroup) {
	if !m.config.Enabled {
		m.logger.Warn("Integration module is disabled, routes not registered", nil)
		return
	}

	integrationGroup := router.Group("/integration")
	{
		// Sync routes
		m.syncHandler.RegisterRoutes(integrationGroup)

		// Employee routes
		m.employeeHandler.RegisterRoutes(integrationGroup)

		// Student routes
		m.studentHandler.RegisterRoutes(integrationGroup)

		// Conflict routes
		m.conflictHandler.RegisterRoutes(integrationGroup)
	}

	m.logger.Info("Integration module routes registered", nil)
}

// StartScheduler starts the sync scheduler
func (m *Module) StartScheduler(ctx context.Context) error {
	if !m.config.Enabled {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.scheduler = cron.New(cron.WithSeconds())

	// Schedule employee sync
	if m.config.SyncCronEmployee != "" {
		_, err := m.scheduler.AddFunc(m.config.SyncCronEmployee, func() {
			m.runEmployeeSync(ctx)
		})
		if err != nil {
			m.logger.Error("Failed to schedule employee sync", map[string]interface{}{
				"cron":  m.config.SyncCronEmployee,
				"error": err.Error(),
			})
		} else {
			m.logger.Info("Scheduled employee sync", map[string]interface{}{
				"cron": m.config.SyncCronEmployee,
			})
		}
	}

	// Schedule student sync
	if m.config.SyncCronStudent != "" {
		_, err := m.scheduler.AddFunc(m.config.SyncCronStudent, func() {
			m.runStudentSync(ctx)
		})
		if err != nil {
			m.logger.Error("Failed to schedule student sync", map[string]interface{}{
				"cron":  m.config.SyncCronStudent,
				"error": err.Error(),
			})
		} else {
			m.logger.Info("Scheduled student sync", map[string]interface{}{
				"cron": m.config.SyncCronStudent,
			})
		}
	}

	m.scheduler.Start()
	m.logger.Info("Integration scheduler started", nil)

	return nil
}

// StopScheduler stops the sync scheduler
func (m *Module) StopScheduler() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.scheduler != nil {
		ctx := m.scheduler.Stop()
		<-ctx.Done()
		m.logger.Info("Integration scheduler stopped", nil)
	}

	return nil
}

// runEmployeeSync runs employee synchronization using the use case's StartSync method
func (m *Module) runEmployeeSync(ctx context.Context) {
	m.logger.Info("Starting scheduled employee sync", nil)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	req := &dto.StartSyncRequest{
		EntityType: entities.SyncEntityEmployee,
		Direction:  entities.SyncDirectionImport,
		Force:      false,
	}

	result, err := m.syncUseCase.StartSync(ctx, req)
	if err != nil {
		m.logger.Error("Scheduled employee sync failed", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	m.logger.Info("Scheduled employee sync completed", map[string]interface{}{
		"success_count":  result.SuccessCount,
		"error_count":    result.ErrorCount,
		"conflict_count": result.ConflictCount,
	})
}

// runStudentSync runs student synchronization using the use case's StartSync method
func (m *Module) runStudentSync(ctx context.Context) {
	m.logger.Info("Starting scheduled student sync", nil)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	req := &dto.StartSyncRequest{
		EntityType: entities.SyncEntityStudent,
		Direction:  entities.SyncDirectionImport,
		Force:      false,
	}

	result, err := m.syncUseCase.StartSync(ctx, req)
	if err != nil {
		m.logger.Error("Scheduled student sync failed", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	m.logger.Info("Scheduled student sync completed", map[string]interface{}{
		"success_count":  result.SuccessCount,
		"error_count":    result.ErrorCount,
		"conflict_count": result.ConflictCount,
	})
}

// IsEnabled returns whether the module is enabled
func (m *Module) IsEnabled() bool {
	return m.config.Enabled
}

// GetSyncUseCase returns the sync use case
func (m *Module) GetSyncUseCase() *usecases.SyncUseCase {
	return m.syncUseCase
}

// GetEmployeeUseCase returns the employee use case
func (m *Module) GetEmployeeUseCase() *usecases.EmployeeUseCase {
	return m.employeeUseCase
}

// GetStudentUseCase returns the student use case
func (m *Module) GetStudentUseCase() *usecases.StudentUseCase {
	return m.studentUseCase
}

// GetConflictUseCase returns the conflict use case
func (m *Module) GetConflictUseCase() *usecases.ConflictUseCase {
	return m.conflictUseCase
}
