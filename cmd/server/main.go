// Package main provides the entry point for the Information System Secretary-Methodologist server.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/repositories"
	persistence "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/infrastructure/persistence"
	authHandler "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/interfaces/http/handlers"
	authMiddleware "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/interfaces/http/middleware"
	docUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/application/usecases"
	docPersistence "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/infrastructure/persistence"
	docHandler "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/interfaces/http/handlers"
	emailServices "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/application/services"
	emailDomain "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/services"
	emailHandlers "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/interfaces/http/handlers"
	reportUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/application/usecases"
	reportPersistence "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/infrastructure/persistence"
	reportHandler "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/interfaces/http/handlers"
	scheduleUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/application/usecases"
	schedulePersistence "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/infrastructure/persistence"
	scheduleHandler "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/interfaces/http/handlers"
	taskUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/application/usecases"
	taskPersistence "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/infrastructure/persistence"
	taskHandler "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/interfaces/http/handlers"
	announcementUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/application/usecases"
	announcementPersistence "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/infrastructure/persistence"
	announcementHandler "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/interfaces/http/handlers"
	appMiddleware "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/application/middleware"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/cache"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/config"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/metrics"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/middleware"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/storage"
)

func main() {
	// Handle version flag
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println("inf-sys-secretary-methodist v0.1.0")
		return
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger := logging.NewLogger(cfg.Log.Level)
	securityLogger := logging.NewSecurityLogger(logger)
	auditLogger := logging.NewAuditLogger(logger)
	perfLogger := logging.NewPerformanceLogger(logger)

	logger.Info("Starting application", map[string]interface{}{
		"environment": cfg.Environment,
		"version":     cfg.Version,
	})

	// Initialize database
	db, err := initDatabase(cfg, logger)
	if err != nil {
		logger.Error("Failed to initialize database", map[string]interface{}{
			"error": err.Error(),
		})
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("Failed to close database connection", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}()

	logger.Info("Database connected successfully", map[string]interface{}{
		"max_open_conns": cfg.Database.MaxOpenConns,
		"max_idle_conns": cfg.Database.MaxIdleConns,
	})

	// Initialize Redis cache
	redisCache, err := initRedisCache(cfg, logger)
	if err != nil {
		logger.Warn("Redis cache not available, running without cache", map[string]interface{}{
			"error": err.Error(),
		})
	}
	if redisCache != nil {
		defer func() {
			if err := redisCache.Close(); err != nil {
				logger.Error("Failed to close Redis connection", map[string]interface{}{
					"error": err.Error(),
				})
			}
		}()
		logger.Info("Redis cache connected successfully", nil)
	}

	// Initialize S3 client for document storage
	var s3Client *storage.S3Client
	if cfg.S3.Endpoint != "" {
		s3Client, err = storage.NewS3Client(cfg.S3)
		if err != nil {
			logger.Warn("S3 storage not available, document uploads disabled", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			// Ensure bucket exists
			if err := s3Client.EnsureBucket(context.Background()); err != nil {
				logger.Warn("Failed to ensure S3 bucket exists", map[string]interface{}{
					"error": err.Error(),
				})
			} else {
				logger.Info("S3 storage connected successfully", map[string]interface{}{
					"bucket": cfg.S3.BucketName,
				})
			}
		}
	}

	// Initialize auth module with all optimizations
	authUseCase, userRepo := initAuthModule(
		db,
		redisCache,
		securityLogger,
		auditLogger,
		perfLogger,
		cfg,
	)

	// Initialize documents module
	var docUseCase *docUsecases.DocumentUseCase
	if s3Client != nil {
		docRepo := docPersistence.NewDocumentRepositoryPG(db)
		docTypeRepo := docPersistence.NewDocumentTypeRepositoryPG(db)
		docCategoryRepo := docPersistence.NewDocumentCategoryRepositoryPG(db)
		docUseCase = docUsecases.NewDocumentUseCase(docRepo, docTypeRepo, docCategoryRepo, s3Client, auditLogger)
		logger.Info("Documents module initialized", nil)
	} else {
		logger.Warn("Documents module not initialized - S3 storage not available", nil)
	}

	// Initialize reporting module
	reportRepo := reportPersistence.NewReportRepositoryPG(db)
	reportTypeRepo := reportPersistence.NewReportTypeRepositoryPG(db)
	reportUseCase := reportUsecases.NewReportUseCase(reportRepo, reportTypeRepo, s3Client, auditLogger)
	logger.Info("Reporting module initialized", nil)

	// Initialize tasks module
	taskRepo := taskPersistence.NewTaskRepositoryPG(db)
	projectRepo := taskPersistence.NewProjectRepositoryPG(db)
	taskUseCase := taskUsecases.NewTaskUseCase(taskRepo, projectRepo, auditLogger)
	projectUseCase := taskUsecases.NewProjectUseCase(projectRepo, auditLogger)
	logger.Info("Tasks module initialized", nil)

	// Initialize schedule module
	eventRepo := schedulePersistence.NewEventRepositoryPG(db)
	participantRepo := schedulePersistence.NewEventParticipantRepositoryPG(db)
	reminderRepo := schedulePersistence.NewEventReminderRepositoryPG(db)
	eventUseCase := scheduleUsecases.NewEventUseCase(eventRepo, participantRepo, reminderRepo, auditLogger)
	logger.Info("Schedule module initialized", nil)

	// Initialize announcements module
	announcementRepo := announcementPersistence.NewAnnouncementRepositoryPG(db)
	announcementUseCase := announcementUsecases.NewAnnouncementUseCase(announcementRepo, auditLogger)
	logger.Info("Announcements module initialized", nil)

	// Initialize shared middleware
	corsMiddleware := appMiddleware.NewCORSMiddleware(
		cfg.CORS.AllowedOrigins,
		cfg.CORS.AllowedMethods,
		cfg.CORS.AllowedHeaders,
	)
	loggingMiddleware := appMiddleware.NewLoggingMiddleware(logger)

	// Setup router with all middleware
	router := setupRoutes(
		authUseCase,
		docUseCase,
		reportUseCase,
		taskUseCase,
		projectUseCase,
		eventUseCase,
		announcementUseCase,
		securityLogger,
		perfLogger,
		auditLogger,
		cfg,
		logger,
		corsMiddleware,
		loggingMiddleware,
		db,
		redisCache,
		userRepo,
	)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Server starting", map[string]interface{}{
			"port":        cfg.Server.Port,
			"environment": cfg.Environment,
		})
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", map[string]interface{}{
				"error": err.Error(),
			})
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Server shutting down...", nil)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", map[string]interface{}{
			"error": err.Error(),
		})
	}

	logger.Info("Server stopped", nil)
}

func initDatabase(cfg *config.Config, _ *logging.Logger) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Database,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool for optimal performance
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func initRedisCache(cfg *config.Config, _ *logging.Logger) (*cache.RedisCache, error) {
	redisAddr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)
	redisCache, err := cache.NewRedisCache(redisAddr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		return nil, err
	}
	return redisCache, nil
}

func initAuthModule(
	db *sql.DB,
	redisCache *cache.RedisCache,
	securityLog *logging.SecurityLogger,
	auditLog *logging.AuditLogger,
	perfLog *logging.PerformanceLogger,
	cfg *config.Config,
) (*usecases.AuthUseCase, repositories.UserRepository) {
	// JWT secrets from config
	jwtSecret := []byte(cfg.JWT.AccessSecret)
	refreshSecret := []byte(cfg.JWT.RefreshSecret)

	// Initialize base user repository
	baseUserRepo := persistence.NewUserRepositoryPG(db)

	// Initialize session repository (will be used in future for refresh token management)
	_ = persistence.NewSessionRepositoryPG(db)

	// Wrap with caching if Redis is available
	var userRepo interface{} = baseUserRepo
	if redisCache != nil {
		userCache := cache.NewUserCache(redisCache, 5*time.Minute)
		userRepo = persistence.NewCachedUserRepository(baseUserRepo, userCache, perfLog)
	}

	// Initialize use case with full logging and session repository
	authUseCase := usecases.NewAuthUseCase(
		userRepo.(repositories.UserRepository),
		jwtSecret,
		refreshSecret,
		securityLog,
		auditLog,
	)

	return authUseCase, userRepo.(repositories.UserRepository)
}

func setupRoutes(
	authUseCase *usecases.AuthUseCase,
	docUseCase *docUsecases.DocumentUseCase,
	reportUseCase *reportUsecases.ReportUseCase,
	taskUseCase *taskUsecases.TaskUseCase,
	projectUseCase *taskUsecases.ProjectUseCase,
	eventUseCase *scheduleUsecases.EventUseCase,
	announcementUseCase *announcementUsecases.AnnouncementUseCase,
	securityLog *logging.SecurityLogger,
	perfLog *logging.PerformanceLogger,
	auditLogger *logging.AuditLogger,
	cfg *config.Config,
	logger *logging.Logger,
	corsMiddleware *appMiddleware.CORSMiddleware,
	loggingMiddleware *appMiddleware.LoggingMiddleware,
	db *sql.DB,
	redisCache *cache.RedisCache,
	userRepo repositories.UserRepository,
) *gin.Engine {
	router := gin.New()

	// Global middleware stack (order matters!)
	router.Use(gin.Recovery())
	router.Use(corsMiddleware.Handler())         // CORS должен быть первым для обработки OPTIONS
	router.Use(middleware.RequestIDMiddleware()) // Request ID для трейсинга
	router.Use(middleware.RequestContextMiddleware())
	router.Use(authMiddleware.SecurityHeadersMiddleware())
	router.Use(metrics.PrometheusMiddleware()) // Prometheus метрики
	router.Use(loggingMiddleware.Handler())    // Логирование всех запросов
	router.Use(performanceMiddleware(perfLog))

	// Handle OPTIONS requests for routes that don't exist (CORS preflight)
	router.NoRoute(func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			c.Status(http.StatusNoContent)
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "route not found"})
	})

	// Health check endpoints for Kubernetes probes
	router.GET("/health", healthCheckHandler(db, redisCache))
	router.GET("/live", livenessHandler())
	router.GET("/ready", readinessHandler(db, redisCache))

	// Prometheus metrics endpoint
	router.GET("/metrics", metrics.MetricsHandler())

	// Загрузка конфигурации rate limiting
	rateLimitConfig := middleware.LoadRateLimitConfig()

	var publicRateLimiter, authRateLimiter *middleware.RateLimiter
	if redisCache != nil {
		publicRateLimiter = rateLimitConfig.GetPublicRateLimiter(redisCache.Client())
		authRateLimiter = rateLimitConfig.GetAuthRateLimiter(redisCache.Client())
	}

	// Initialize email service
	composioAPIKey := cfg.Composio.APIKey
	composioEntityID := cfg.Composio.EntityID
	var emailService emailDomain.EmailService
	if composioAPIKey != "" && composioEntityID != "" {
		emailService = emailServices.NewComposioEmailService(composioAPIKey, composioEntityID, auditLogger)
		logger.Info("Email service initialized", nil)
	}

	// Initialize auth handler with email service
	authHandlerInstance := authHandler.NewAuthHandler(authUseCase, emailService)

	// Public auth routes with rate limiting (10 req/min + burst 5)
	authGroup := router.Group("/api/auth")
	if publicRateLimiter != nil {
		authGroup.Use(publicRateLimiter.RateLimitMiddleware())
		authGroup.Use(rateLimitLogger(securityLog))
	}
	{
		// Register POST handlers
		authGroup.POST("/register", authHandlerInstance.Register)
		authGroup.POST("/login", authHandlerInstance.Login)
		authGroup.POST("/refresh", authHandlerInstance.RefreshToken)

		// Register OPTIONS handlers for CORS preflight
		authGroup.OPTIONS("/register", func(c *gin.Context) { c.Status(http.StatusNoContent) })
		authGroup.OPTIONS("/login", func(c *gin.Context) { c.Status(http.StatusNoContent) })
		authGroup.OPTIONS("/refresh", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	}

	// Protected routes (require JWT) with auth rate limiting (60 req/min + burst 10)
	protectedGroup := router.Group("/api")
	protectedGroup.Use(authMiddleware.JWTMiddleware(authUseCase))
	if authRateLimiter != nil {
		protectedGroup.Use(authRateLimiter.RateLimitMiddleware())
	}
	{
		protectedGroup.GET("/me", func(c *gin.Context) {
			userID, exists := c.Get("user_id")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
				return
			}

			// Get full user data from database
			user, err := userRepo.GetByID(c.Request.Context(), userID.(int64))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user data"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"id":        user.ID,
				"email":     user.Email,
				"name":      user.Name,
				"role":      user.Role,
				"created_at": user.CreatedAt,
				"updated_at": user.UpdatedAt,
			})
		})
		protectedGroup.OPTIONS("/me", func(c *gin.Context) { c.Status(http.StatusNoContent) })

		// Email notification routes
		if emailService != nil {
			emailHandler := emailHandlers.NewEmailHandler(emailService)

			notificationsGroup := protectedGroup.Group("/notifications")
			{
				notificationsGroup.POST("/send-email", emailHandler.SendEmail)
				notificationsGroup.POST("/send-welcome", emailHandler.SendWelcomeEmail)
				notificationsGroup.OPTIONS("/send-email", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				notificationsGroup.OPTIONS("/send-welcome", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			}
			logger.Info("Email notification routes registered", nil)
		} else {
			logger.Warn("Email notification routes not registered - Composio credentials not configured", nil)
		}

		// Document management routes
		if docUseCase != nil {
			docHandlerInstance := docHandler.NewDocumentHandler(docUseCase)

			documentsGroup := protectedGroup.Group("/documents")
			{
				documentsGroup.POST("", docHandlerInstance.Create)
				documentsGroup.GET("", docHandlerInstance.List)
				documentsGroup.GET("/:id", docHandlerInstance.GetByID)
				documentsGroup.PUT("/:id", docHandlerInstance.Update)
				documentsGroup.DELETE("/:id", docHandlerInstance.Delete)
				documentsGroup.POST("/:id/file", docHandlerInstance.UploadFile)
				documentsGroup.GET("/:id/file", docHandlerInstance.DownloadFile)
				documentsGroup.DELETE("/:id/file", docHandlerInstance.DeleteFile)

				// CORS preflight handlers
				documentsGroup.OPTIONS("", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				documentsGroup.OPTIONS("/:id", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				documentsGroup.OPTIONS("/:id/file", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			}

			// Document types and categories (reference data)
			protectedGroup.GET("/document-types", docHandlerInstance.GetDocumentTypes)
			protectedGroup.GET("/document-categories", docHandlerInstance.GetCategories)
			protectedGroup.OPTIONS("/document-types", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			protectedGroup.OPTIONS("/document-categories", func(c *gin.Context) { c.Status(http.StatusNoContent) })

			logger.Info("Documents module routes registered", nil)
		} else {
			logger.Warn("Documents module routes not registered - S3 storage not available", nil)
		}

		// Reporting module routes
		if reportUseCase != nil {
			reportHandlerInstance := reportHandler.NewReportHandler(reportUseCase)

			// Report types (reference data)
			protectedGroup.GET("/report-types", reportHandlerInstance.GetReportTypes)
			protectedGroup.GET("/report-types/:id", reportHandlerInstance.GetReportTypeByID)
			protectedGroup.OPTIONS("/report-types", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			protectedGroup.OPTIONS("/report-types/:id", func(c *gin.Context) { c.Status(http.StatusNoContent) })

			reportsGroup := protectedGroup.Group("/reports")
			{
				// CRUD operations
				reportsGroup.POST("", reportHandlerInstance.Create)
				reportsGroup.GET("", reportHandlerInstance.List)
				reportsGroup.GET("/:id", reportHandlerInstance.GetByID)
				reportsGroup.PUT("/:id", reportHandlerInstance.Update)
				reportsGroup.DELETE("/:id", reportHandlerInstance.Delete)

				// Report generation and workflow
				reportsGroup.POST("/:id/generate", reportHandlerInstance.Generate)
				reportsGroup.POST("/:id/submit", reportHandlerInstance.SubmitForReview)
				reportsGroup.POST("/:id/review", reportHandlerInstance.Review)
				reportsGroup.POST("/:id/publish", reportHandlerInstance.Publish)

				// Access management
				reportsGroup.GET("/:id/access", reportHandlerInstance.GetAccess)
				reportsGroup.POST("/:id/access", reportHandlerInstance.AddAccess)
				reportsGroup.DELETE("/:id/access/:access_id", reportHandlerInstance.RemoveAccess)

				// Comments
				reportsGroup.GET("/:id/comments", reportHandlerInstance.GetComments)
				reportsGroup.POST("/:id/comments", reportHandlerInstance.AddComment)

				// History
				reportsGroup.GET("/:id/history", reportHandlerInstance.GetHistory)

				// CORS preflight handlers
				reportsGroup.OPTIONS("", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				reportsGroup.OPTIONS("/:id", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				reportsGroup.OPTIONS("/:id/generate", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				reportsGroup.OPTIONS("/:id/submit", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				reportsGroup.OPTIONS("/:id/review", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				reportsGroup.OPTIONS("/:id/publish", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				reportsGroup.OPTIONS("/:id/access", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				reportsGroup.OPTIONS("/:id/access/:access_id", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				reportsGroup.OPTIONS("/:id/comments", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				reportsGroup.OPTIONS("/:id/history", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			}

			logger.Info("Reporting module routes registered", nil)
		}

		// Tasks module routes
		if taskUseCase != nil {
			taskHandlerInstance := taskHandler.NewTaskHandler(taskUseCase)
			projectHandlerInstance := taskHandler.NewProjectHandler(projectUseCase)

			// Projects routes
			projectsGroup := protectedGroup.Group("/projects")
			{
				projectsGroup.POST("", projectHandlerInstance.Create)
				projectsGroup.GET("", projectHandlerInstance.List)
				projectsGroup.GET("/:id", projectHandlerInstance.GetByID)
				projectsGroup.PUT("/:id", projectHandlerInstance.Update)
				projectsGroup.DELETE("/:id", projectHandlerInstance.Delete)
				projectsGroup.POST("/:id/activate", projectHandlerInstance.Activate)
				projectsGroup.POST("/:id/hold", projectHandlerInstance.PutOnHold)
				projectsGroup.POST("/:id/complete", projectHandlerInstance.Complete)
				projectsGroup.POST("/:id/cancel", projectHandlerInstance.Cancel)

				// CORS preflight handlers
				projectsGroup.OPTIONS("", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				projectsGroup.OPTIONS("/:id", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				projectsGroup.OPTIONS("/:id/activate", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				projectsGroup.OPTIONS("/:id/hold", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				projectsGroup.OPTIONS("/:id/complete", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				projectsGroup.OPTIONS("/:id/cancel", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			}

			// Tasks routes
			tasksGroup := protectedGroup.Group("/tasks")
			{
				// CRUD operations
				tasksGroup.POST("", taskHandlerInstance.Create)
				tasksGroup.GET("", taskHandlerInstance.List)
				tasksGroup.GET("/:id", taskHandlerInstance.GetByID)
				tasksGroup.PUT("/:id", taskHandlerInstance.Update)
				tasksGroup.DELETE("/:id", taskHandlerInstance.Delete)

				// Task workflow
				tasksGroup.POST("/:id/assign", taskHandlerInstance.Assign)
				tasksGroup.POST("/:id/unassign", taskHandlerInstance.Unassign)
				tasksGroup.POST("/:id/start", taskHandlerInstance.StartWork)
				tasksGroup.POST("/:id/review", taskHandlerInstance.SubmitForReview)
				tasksGroup.POST("/:id/complete", taskHandlerInstance.Complete)
				tasksGroup.POST("/:id/cancel", taskHandlerInstance.Cancel)
				tasksGroup.POST("/:id/reopen", taskHandlerInstance.Reopen)

				// Watchers
				tasksGroup.GET("/:id/watchers", taskHandlerInstance.GetWatchers)
				tasksGroup.POST("/:id/watchers", taskHandlerInstance.AddWatcher)
				tasksGroup.DELETE("/:id/watchers/:watcher_id", taskHandlerInstance.RemoveWatcher)

				// Comments
				tasksGroup.GET("/:id/comments", taskHandlerInstance.GetComments)
				tasksGroup.POST("/:id/comments", taskHandlerInstance.AddComment)
				tasksGroup.PUT("/comments/:comment_id", taskHandlerInstance.UpdateComment)
				tasksGroup.DELETE("/comments/:comment_id", taskHandlerInstance.DeleteComment)

				// Checklists
				tasksGroup.GET("/:id/checklists", taskHandlerInstance.GetChecklists)
				tasksGroup.POST("/:id/checklists", taskHandlerInstance.AddChecklist)
				tasksGroup.DELETE("/checklists/:checklist_id", taskHandlerInstance.DeleteChecklist)
				tasksGroup.POST("/checklists/:checklist_id/items", taskHandlerInstance.AddChecklistItem)
				tasksGroup.DELETE("/checklists/items/:item_id", taskHandlerInstance.DeleteChecklistItem)

				// History
				tasksGroup.GET("/:id/history", taskHandlerInstance.GetHistory)

				// CORS preflight handlers
				tasksGroup.OPTIONS("", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				tasksGroup.OPTIONS("/:id", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				tasksGroup.OPTIONS("/:id/assign", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				tasksGroup.OPTIONS("/:id/unassign", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				tasksGroup.OPTIONS("/:id/start", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				tasksGroup.OPTIONS("/:id/review", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				tasksGroup.OPTIONS("/:id/complete", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				tasksGroup.OPTIONS("/:id/cancel", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				tasksGroup.OPTIONS("/:id/reopen", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				tasksGroup.OPTIONS("/:id/watchers", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				tasksGroup.OPTIONS("/:id/watchers/:watcher_id", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				tasksGroup.OPTIONS("/:id/comments", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				tasksGroup.OPTIONS("/comments/:comment_id", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				tasksGroup.OPTIONS("/:id/checklists", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				tasksGroup.OPTIONS("/checklists/:checklist_id", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				tasksGroup.OPTIONS("/checklists/:checklist_id/items", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				tasksGroup.OPTIONS("/checklists/items/:item_id", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				tasksGroup.OPTIONS("/:id/history", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			}

			logger.Info("Tasks module routes registered", nil)
		}

		// Schedule/Events module routes
		if eventUseCase != nil {
			eventHandlerInstance := scheduleHandler.NewEventHandler(eventUseCase)

			eventsGroup := protectedGroup.Group("/events")
			{
				// CRUD operations
				eventsGroup.POST("", eventHandlerInstance.Create)
				eventsGroup.GET("", eventHandlerInstance.List)
				eventsGroup.GET("/:id", eventHandlerInstance.GetByID)
				eventsGroup.PUT("/:id", eventHandlerInstance.Update)
				eventsGroup.DELETE("/:id", eventHandlerInstance.Delete)

				// Special queries
				eventsGroup.GET("/range", eventHandlerInstance.GetByDateRange)
				eventsGroup.GET("/upcoming", eventHandlerInstance.GetUpcoming)
				eventsGroup.GET("/invitations", eventHandlerInstance.GetPendingInvitations)

				// Event actions
				eventsGroup.POST("/:id/cancel", eventHandlerInstance.Cancel)
				eventsGroup.POST("/:id/reschedule", eventHandlerInstance.Reschedule)

				// Participants
				eventsGroup.POST("/:id/participants", eventHandlerInstance.AddParticipants)
				eventsGroup.DELETE("/:id/participants/:user_id", eventHandlerInstance.RemoveParticipant)
				eventsGroup.POST("/:id/respond", eventHandlerInstance.UpdateParticipantStatus)

				// CORS preflight handlers
				eventsGroup.OPTIONS("", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				eventsGroup.OPTIONS("/:id", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				eventsGroup.OPTIONS("/range", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				eventsGroup.OPTIONS("/upcoming", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				eventsGroup.OPTIONS("/invitations", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				eventsGroup.OPTIONS("/:id/cancel", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				eventsGroup.OPTIONS("/:id/reschedule", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				eventsGroup.OPTIONS("/:id/participants", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				eventsGroup.OPTIONS("/:id/participants/:user_id", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				eventsGroup.OPTIONS("/:id/respond", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			}

			logger.Info("Schedule module routes registered", nil)
		}

		// Announcements module routes
		if announcementUseCase != nil {
			announcementHandlerInstance := announcementHandler.NewAnnouncementHandler(announcementUseCase)

			announcementsGroup := protectedGroup.Group("/announcements")
			{
				// CRUD operations
				announcementsGroup.POST("", announcementHandlerInstance.Create)
				announcementsGroup.GET("", announcementHandlerInstance.List)
				announcementsGroup.GET("/:id", announcementHandlerInstance.GetByID)
				announcementsGroup.PUT("/:id", announcementHandlerInstance.Update)
				announcementsGroup.DELETE("/:id", announcementHandlerInstance.Delete)

				// Special queries
				announcementsGroup.GET("/published", announcementHandlerInstance.GetPublished)
				announcementsGroup.GET("/pinned", announcementHandlerInstance.GetPinned)
				announcementsGroup.GET("/recent", announcementHandlerInstance.GetRecent)

				// Announcement actions
				announcementsGroup.POST("/:id/publish", announcementHandlerInstance.Publish)
				announcementsGroup.POST("/:id/unpublish", announcementHandlerInstance.Unpublish)
				announcementsGroup.POST("/:id/archive", announcementHandlerInstance.Archive)

				// CORS preflight handlers
				announcementsGroup.OPTIONS("", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				announcementsGroup.OPTIONS("/:id", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				announcementsGroup.OPTIONS("/published", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				announcementsGroup.OPTIONS("/pinned", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				announcementsGroup.OPTIONS("/recent", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				announcementsGroup.OPTIONS("/:id/publish", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				announcementsGroup.OPTIONS("/:id/unpublish", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				announcementsGroup.OPTIONS("/:id/archive", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			}

			logger.Info("Announcements module routes registered", nil)
		}

		// Admin only routes
		adminGroup := protectedGroup.Group("/admin")
		adminGroup.Use(authMiddleware.RequireRole("admin"))
		{
			adminGroup.GET("/users", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "Admin users list"})
			})
			adminGroup.OPTIONS("/users", func(c *gin.Context) { c.Status(http.StatusNoContent) })
		}
	}

	return router
}

// performanceMiddleware logs HTTP request performance
func performanceMiddleware(perfLog *logging.PerformanceLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		perfLog.LogHTTPRequest(
			c.Request.Context(),
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			duration,
		)
	}
}

// rateLimitLogger logs rate limit violations
func rateLimitLogger(securityLog *logging.SecurityLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if rate limited
		if c.Writer.Status() == http.StatusTooManyRequests {
			securityLog.LogRateLimitExceeded(c.Request.Context(), c.Request.URL.Path)
		}
	}
}

const (
	healthStatusOK       = "OK"
	healthStatusDegraded = "DEGRADED"
)

// livenessHandler returns a simple liveness probe endpoint for Kubernetes.
// It only checks if the application process is running.
func livenessHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "UP",
			"timestamp": time.Now().UTC(),
		})
	}
}

// readinessHandler returns a readiness probe endpoint for Kubernetes.
// It checks if all dependencies are available and the service is ready to accept traffic.
func readinessHandler(db *sql.DB, redisCache *cache.RedisCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		ready := true
		checks := make(map[string]any)

		// Check database - required dependency
		if err := db.PingContext(ctx); err != nil {
			ready = false
			checks["database"] = map[string]any{
				"status": "DOWN",
				"error":  err.Error(),
			}
		} else {
			checks["database"] = map[string]any{
				"status": "UP",
			}
		}

		// Check Redis if available (optional dependency)
		if redisCache != nil {
			if err := redisCache.Ping(ctx); err != nil {
				// Redis is optional, so we don't set ready=false
				checks["redis"] = map[string]any{
					"status": "DOWN",
					"error":  err.Error(),
				}
			} else {
				checks["redis"] = map[string]any{
					"status": "UP",
				}
			}
		} else {
			checks["redis"] = map[string]any{
				"status": "DISABLED",
			}
		}

		httpStatus := http.StatusOK
		if !ready {
			httpStatus = http.StatusServiceUnavailable
		}

		c.JSON(httpStatus, gin.H{
			"ready":     ready,
			"timestamp": time.Now().UTC(),
			"checks":    checks,
		})
	}
}

// healthCheckHandler returns a health check endpoint with dependency checks
func healthCheckHandler(db *sql.DB, redisCache *cache.RedisCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		status := healthStatusOK
		checks := make(map[string]interface{})

		// Check database
		if err := db.PingContext(ctx); err != nil {
			status = healthStatusDegraded
			checks["database"] = map[string]interface{}{
				"status": "DOWN",
				"error":  err.Error(),
			}
		} else {
			checks["database"] = map[string]interface{}{
				"status": "UP",
			}
		}

		// Check Redis if available
		if redisCache != nil {
			if err := redisCache.Ping(ctx); err != nil {
				status = healthStatusDegraded
				checks["redis"] = map[string]interface{}{
					"status": "DOWN",
					"error":  err.Error(),
				}
			} else {
				checks["redis"] = map[string]interface{}{
					"status": "UP",
				}
			}
		} else {
			checks["redis"] = map[string]interface{}{
				"status": "DISABLED",
			}
		}

		httpStatus := http.StatusOK
		if status == healthStatusDegraded {
			httpStatus = http.StatusServiceUnavailable
		}

		c.JSON(httpStatus, gin.H{
			"status":    status,
			"timestamp": time.Now().UTC(),
			"checks":    checks,
		})
	}
}
