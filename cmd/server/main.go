// Package main provides the entry point for the Information System Secretary-Methodologist server.
//
// @title           Inf-Sys Secretary-Methodist API
// @version         0.108.0
// @description     API для информационной системы академического секретаря/методиста.
// @description     Включает управление документами, расписанием, задачами, уведомлениями и мессенджером.
//
// @contact.name    API Support
// @contact.url     https://github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist
//
// @license.name    MIT
// @license.url     https://opensource.org/licenses/MIT
//
// @host            localhost:8080
// @BasePath        /api
// @schemes         http https
//
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description JWT Bearer token. Format: "Bearer {token}"
//
// @tag.name auth
// @tag.description Аутентификация и авторизация
// @tag.name documents
// @tag.description Управление документами
// @tag.name tasks
// @tag.description Управление задачами и проектами
// @tag.name schedule
// @tag.description Управление расписанием и событиями
// @tag.name reports
// @tag.description Отчёты и аналитика
// @tag.name notifications
// @tag.description Уведомления и настройки
// @tag.name users
// @tag.description Управление пользователями
// @tag.name messaging
// @tag.description Внутренний мессенджер
// @tag.name files
// @tag.description Загрузка и управление файлами
// @tag.name announcements
// @tag.description Объявления
// @tag.name dashboard
// @tag.description Дашборд и статистика
// @tag.name AI
// @tag.description AI-ассистент и семантический поиск
package main

import (
	"context"
	"database/sql"

	"github.com/XSAM/otelsql"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/docs/swagger"

	aiServices "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/application/services"
	aiUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/ports"
	aiAdapters "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/infrastructure/adapters"
	aiPersistence "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/infrastructure/persistence"
	aiPrompts "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/infrastructure/prompts"
	aiProviders "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/infrastructure/providers"
	aiScheduler "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/infrastructure/scheduler"
	aiHandler "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/interfaces/http/handlers"
	analyticsUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/analytics/application/usecases"
	analyticsPersistence "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/analytics/infrastructure/persistence"
	analyticsEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/analytics/domain/entities"
	analyticsScheduler "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/analytics/infrastructure/scheduler"
	analyticsHandler "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/analytics/interfaces/http/handlers"
	announcementUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/application/usecases"
	announcementPersistence "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/infrastructure/persistence"
	announcementHandler "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/interfaces/http/handlers"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/application/usecases"
	authDomain "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain/repositories"
	persistence "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/infrastructure/persistence"
	authHandler "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/interfaces/http/handlers"
	authMiddleware "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/interfaces/http/middleware"
	dashboardUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/dashboard/application/usecases"
	dashboardPersistence "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/dashboard/infrastructure/persistence"
	dashboardHandler "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/dashboard/interfaces/http/handlers"
	docUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/application/usecases"
	docPersistence "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/infrastructure/persistence"
	docHandler "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/interfaces/http/handlers"
	filesUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/files/application/usecases"
	filesPersistence "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/files/infrastructure/persistence"
	filesHandler "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/files/interfaces/http/handlers"
	integration "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration"
	messagingServices "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/messaging/application/services"
	messagingUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/messaging/application/usecases"
	messagingPersistence "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/messaging/infrastructure/persistence"
	messagingWebsocket "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/messaging/infrastructure/websocket"
	messagingHandler "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/messaging/interfaces/http"
	notifDTO "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/application/dto"
	notifServices "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/application/services"
	notifUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/application/usecases"
	notifRepositories "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/repositories"
	emailDomain "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/services"
	notifPersistence "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/infrastructure/persistence"
	notifScheduler "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/infrastructure/scheduler"
	notifHttp "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/interfaces/http"
	emailHandlers "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/interfaces/http/handlers"
	reportUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/application/usecases"
	reportPersistence "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/infrastructure/persistence"
	reportQuery "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/infrastructure/query"
	reportHandler "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/interfaces/http/handlers"
	scheduleUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/application/usecases"
	schedulePersistence "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/infrastructure/persistence"
	scheduleHandler "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/interfaces/http/handlers"
	taskUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/application/usecases"
	taskPersistence "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/infrastructure/persistence"
	taskHandler "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/interfaces/http/handlers"
	usersUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/application/usecases"
	usersRepositories "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/domain/repositories"
	usersPersistence "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/infrastructure/persistence"
	usersHandler "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/interfaces/http/handlers"
	appMiddleware "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/application/middleware"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/cache"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/config"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/metrics"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/middleware"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/storage"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/telegram"
	n8ninfra "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/n8n"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/tracing"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/validation"
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

	// Initialize Sentry for error tracking
	sentryDSN := os.Getenv("SENTRY_DSN")
	if sentryDSN != "" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:              sentryDSN,
			Environment:      cfg.Environment,
			Release:          cfg.Version,
			TracesSampleRate: 0.1, // 10% трассировки в production
			EnableTracing:    true,
		})
		if err != nil {
			log.Printf("Sentry initialization failed: %v", err)
		} else {
			log.Println("Sentry initialized successfully")
		}
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

	// Initialize OpenTelemetry tracing if enabled
	var tracer *tracing.Tracer
	if cfg.Tracing.Enabled {
		tracerCfg := tracing.TracerConfig{
			ServiceName:  cfg.Tracing.ServiceName,
			Version:      cfg.Version,
			Environment:  cfg.Environment,
			OTLPEndpoint: cfg.Tracing.OTLPEndpoint,
			SamplingRate: cfg.Tracing.SamplingRate,
		}
		tracer, err = tracing.InitTracer(context.Background(), tracerCfg)
		if err != nil {
			logger.Warn("Failed to initialize tracing, running without distributed tracing", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			logger.Info("Distributed tracing initialized", map[string]interface{}{
				"endpoint":      cfg.Tracing.OTLPEndpoint,
				"sampling_rate": cfg.Tracing.SamplingRate,
			})
		}
	}

	// Initialize database
	db, err := initDatabase(cfg, logger, cfg.Tracing.Enabled)
	if err != nil {
		if tracer != nil {
			_ = tracer.Shutdown(context.Background())
		}
		sentry.Flush(2 * time.Second)
		log.Fatalf("Failed to initialize database: %v", err)
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

	// Initialize n8n webhook client for workflow automation
	n8nClient := n8ninfra.NewClient(n8ninfra.Config{
		WebhookURL: cfg.N8N.WebhookURL,
		Enabled:    cfg.N8N.Enabled,
	}, logger)
	if n8nClient.IsEnabled() {
		logger.Info("n8n webhook integration enabled", map[string]any{
			"webhook_url": cfg.N8N.WebhookURL,
		})
	}

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

	// Initialize notifications module FIRST (needed by other modules)
	notificationRepo := notifPersistence.NewNotificationRepositoryPG(db)
	preferencesRepo := notifPersistence.NewPreferencesRepositoryPG(db)

	// Initialize email service for notifications (will be used if Composio is configured)
	composioAPIKey := cfg.Composio.APIKey
	composioEntityID := cfg.Composio.EntityID
	var notifEmailService emailDomain.EmailService
	if composioAPIKey != "" && composioEntityID != "" {
		notifEmailService = notifServices.NewComposioEmailService(composioAPIKey, composioEntityID, auditLogger)
	}

	// Initialize Telegram service (Composio) for sending messages
	telegramRepo := notifPersistence.NewTelegramRepositoryPG(db)
	var telegramService emailDomain.TelegramService
	var telegramVerificationService *notifServices.TelegramVerificationService
	if composioAPIKey != "" && composioEntityID != "" {
		telegramService = notifServices.NewComposioTelegramService(composioAPIKey, composioEntityID, auditLogger)
		telegramVerificationService = notifServices.NewTelegramVerificationService(
			telegramRepo,
			preferencesRepo,
			telegramService,
			auditLogger,
			cfg.Telegram.BotUsername,
		)
		logger.Info("Telegram service initialized via Composio", map[string]interface{}{
			"bot_username": cfg.Telegram.BotUsername,
		})
	} else {
		logger.Warn("Telegram service not configured - COMPOSIO_API_KEY or COMPOSIO_ENTITY_ID not set", nil)
	}

	// Initialize Web Push service
	webpushRepo := notifPersistence.NewWebPushRepositoryPG(db)
	var webpushService emailDomain.WebPushService
	if cfg.WebPush.VAPIDPublicKey != "" && cfg.WebPush.VAPIDPrivateKey != "" {
		webpushService = notifServices.NewWebPushService(
			webpushRepo,
			cfg.WebPush.VAPIDPublicKey,
			cfg.WebPush.VAPIDPrivateKey,
			cfg.WebPush.VAPIDSubject,
			auditLogger,
		)
		logger.Info("Web Push service initialized", nil)
	} else {
		logger.Warn("Web Push service not configured - VAPID keys not set", nil)
	}

	notificationUseCase := notifUsecases.NewNotificationUseCase(notificationRepo, preferencesRepo, telegramRepo, notifEmailService, telegramService, webpushService)
	preferencesUseCase := notifUsecases.NewPreferencesUseCase(preferencesRepo)
	logger.Info("Notifications module initialized", nil)

	// Initialize auth module with all optimizations
	authUseCase, userRepo := initAuthModule(
		db,
		redisCache,
		securityLogger,
		auditLogger,
		perfLogger,
		cfg,
		notificationUseCase,
	)

	// Initialize documents module
	var docUseCase *docUsecases.DocumentUseCase
	var sharingUseCase *docUsecases.SharingUseCase
	var docVersionUseCase *docUsecases.DocumentVersionUseCase
	var tagUseCase *docUsecases.TagUseCase
	var templateUseCase *docUsecases.TemplateUseCase
	if s3Client != nil {
		docRepo := docPersistence.NewDocumentRepositoryPG(db)
		docTypeRepo := docPersistence.NewDocumentTypeRepositoryPG(db)
		docCategoryRepo := docPersistence.NewDocumentCategoryRepositoryPG(db)
		permissionRepo := docPersistence.NewPermissionRepositoryPG(db)
		publicLinkRepo := docPersistence.NewPublicLinkRepositoryPG(db)
		docTagRepo := docPersistence.NewDocumentTagRepositoryPG(db)
		docUseCase = docUsecases.NewDocumentUseCase(docRepo, docTypeRepo, docCategoryRepo, s3Client, auditLogger)
		sharingUseCase = docUsecases.NewSharingUseCase(docRepo, permissionRepo, publicLinkRepo, auditLogger, cfg.Server.BaseURL, notificationUseCase)
		docVersionUseCase = docUsecases.NewDocumentVersionUseCase(docRepo, s3Client, auditLogger)
		tagUseCase = docUsecases.NewTagUseCase(docTagRepo, docRepo, auditLogger)
		templateRepo := docPersistence.NewTemplateRepositoryAdapter(docTypeRepo)
		templateUseCase = docUsecases.NewTemplateUseCase(templateRepo, docRepo, auditLogger)
		logger.Info("Documents module initialized", nil)
	} else {
		logger.Warn("Documents module not initialized - S3 storage not available", nil)
	}

	// Initialize reporting module
	reportRepo := reportPersistence.NewReportRepositoryPG(db)
	reportTypeRepo := reportPersistence.NewReportTypeRepositoryPG(db)
	reportUseCase := reportUsecases.NewReportUseCase(reportRepo, reportTypeRepo, s3Client, auditLogger, notificationUseCase)
	logger.Info("Reporting module initialized", nil)

	// Initialize custom report module
	customReportRepo := reportPersistence.NewCustomReportRepositoryPG(db)
	customQueryBuilder := reportQuery.NewDynamicQueryBuilder(db)
	customReportUseCase := reportUsecases.NewCustomReportUseCase(customReportRepo, customQueryBuilder)
	logger.Info("Custom reports module initialized", nil)

	// Initialize tasks module
	taskRepo := taskPersistence.NewTaskRepositoryPG(db)
	projectRepo := taskPersistence.NewProjectRepositoryPG(db)
	taskUseCase := taskUsecases.NewTaskUseCase(taskRepo, projectRepo, auditLogger, notificationUseCase)
	projectUseCase := taskUsecases.NewProjectUseCase(projectRepo, auditLogger)
	logger.Info("Tasks module initialized", nil)

	// Initialize schedule module
	eventRepo := schedulePersistence.NewEventRepositoryPG(db)
	participantRepo := schedulePersistence.NewEventParticipantRepositoryPG(db)
	reminderRepo := schedulePersistence.NewEventReminderRepositoryPG(db)
	eventUseCase := scheduleUsecases.NewEventUseCase(eventRepo, participantRepo, reminderRepo, auditLogger, notificationUseCase)
	logger.Info("Schedule module initialized", nil)

	// Initialize schedule lessons
	lessonRepo := schedulePersistence.NewLessonRepositoryPG(db)
	classroomRepo := schedulePersistence.NewClassroomRepositoryPG(db)
	referenceRepo := schedulePersistence.NewReferenceRepositoryPG(db)
	changeRepo := schedulePersistence.NewScheduleChangeRepositoryPG(db)
	lessonUseCase := scheduleUsecases.NewLessonUseCase(lessonRepo, classroomRepo, referenceRepo, changeRepo, auditLogger)
	logger.Info("Schedule lessons module initialized", nil)

	// Initialize reminder scheduler
	var reminderScheduler *notifScheduler.ReminderScheduler
	reminderScheduler, err = notifScheduler.NewReminderScheduler(
		db,
		reminderRepo,
		eventRepo,
		notificationRepo,
		preferencesRepo,
		notifEmailService,
		nil, // Use default config
	)
	if err != nil {
		logger.Error("Failed to initialize reminder scheduler", map[string]interface{}{
			"error": err.Error(),
		})
	} else {
		if err := reminderScheduler.Start(); err != nil {
			logger.Error("Failed to start reminder scheduler", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			logger.Info("Reminder scheduler started", nil)
		}
	}

	// Initialize announcements module
	announcementRepo := announcementPersistence.NewAnnouncementRepositoryPG(db)
	announcementUseCase := announcementUsecases.NewAnnouncementUseCase(announcementRepo, auditLogger, notificationUseCase, nil)
	if s3Client != nil {
		announcementUseCase.SetAttachmentStorage(s3Client)
		logger.Info("Announcement attachment storage wired (S3)", nil)
	}
	logger.Info("Announcements module initialized", nil)

	// Initialize dashboard module
	dashboardRepo := dashboardPersistence.NewDashboardRepositoryPG(db)
	dashboardUseCase := dashboardUsecases.NewDashboardUseCase(dashboardRepo)
	logger.Info("Dashboard module initialized", nil)

	// Initialize analytics module (predictive analytics for student risk assessment)
	analyticsRepo := analyticsPersistence.NewAnalyticsRepositoryPG(db)
	attendanceRepo := analyticsPersistence.NewAttendanceRepositoryPG(db)
	gradeRepo := analyticsPersistence.NewGradeRepositoryPG(db)
	analyticsUseCase := analyticsUsecases.NewAnalyticsUseCase(analyticsRepo, attendanceRepo, gradeRepo, auditLogger)

	// Start risk recalculation scheduler (daily at 3:00 AM)
	// Alert curators when student risk > 70
	riskAlertFunc := func(ctx context.Context, student analyticsEntities.StudentRiskScore) {
		groupName := ""
		if student.GroupName != nil {
			groupName = *student.GroupName
		}

		// In-system curator notification (best-effort).
		if notificationUseCase != nil {
			_, _ = notificationUseCase.Create(ctx, &notifDTO.CreateNotificationInput{
				UserID:   student.StudentID,
				Type:     "warning",
				Priority: "high",
				Title:    "Студент в зоне риска",
				Message:  fmt.Sprintf("Студент %s (группа %s) имеет risk score %.0f/100 (уровень: %s)", student.StudentName, groupName, student.RiskScore, student.RiskLevel),
				Link:     fmt.Sprintf("/analytics?student=%d", student.StudentID),
			})
		}

		// Fan out to n8n so external workflows (Telegram broadcasts,
		// curator email digests, etc.) react in seconds instead of
		// waiting for the hourly schedule trigger in the
		// absence-alert workflow. Async; the scheduler's batch loop
		// must not block on webhook latency.
		//
		// PII surface: payload includes student name + numeric risk
		// score going to the operator-controlled n8n instance. Curator
		// notifications already carry the same fields, so this opens
		// no new data-classification gap — but any future addition
		// (email, phone, grades) must be reviewed against the
		// roles-and-flows.md data-handling section first.
		auditLogger.LogAuditEvent(ctx, "risk_alert_dispatched_to_n8n", "analytics", map[string]any{
			"student_id": student.StudentID,
			"risk_score": student.RiskScore,
			"risk_level": student.RiskLevel,
		})
		n8nClient.TriggerAsync(n8ninfra.PathRiskAlertDetected, map[string]any{
			"event_type":   n8ninfra.EventTypeRiskAlertDetected,
			"student_id":   student.StudentID,
			"student_name": student.StudentName,
			"group_name":   groupName,
			"risk_score":   student.RiskScore,
			"risk_level":   student.RiskLevel,
			"occurred_at":  time.Now().UTC().Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	riskScheduler, err := analyticsScheduler.NewRiskRecalcScheduler(analyticsRepo, logger, riskAlertFunc)
	if err != nil {
		logger.Warn("Failed to initialize risk recalculation scheduler", map[string]any{"error": err.Error()})
	} else {
		riskScheduler.Start()
		defer func() { _ = riskScheduler.Stop() }()
	}

	logger.Info("Analytics module initialized", nil)

	// Initialize users module
	departmentRepo := usersPersistence.NewDepartmentRepositoryPG(db)
	positionRepo := usersPersistence.NewPositionRepositoryPG(db)
	userProfileRepo := usersPersistence.NewUserProfileRepositoryPG(db)
	userUseCase := usersUsecases.NewUserUseCase(userRepo, userProfileRepo, departmentRepo, positionRepo, auditLogger, notificationUseCase)
	departmentUseCase := usersUsecases.NewDepartmentUseCase(departmentRepo, auditLogger)
	positionUseCase := usersUsecases.NewPositionUseCase(positionRepo, auditLogger)
	logger.Info("Users module initialized", nil)

	// Initialize messaging module
	conversationRepo := messagingPersistence.NewConversationRepositoryPG(db)
	messageRepo := messagingPersistence.NewMessageRepositoryPG(db)
	messagingHub := messagingWebsocket.NewHub(logger)
	go messagingHub.Run() // Start WebSocket hub in background
	// Create message notifier for sending notifications about new messages
	messageNotifier := messagingServices.NewNotificationNotifier(notificationUseCase)
	messagingUseCase := messagingUsecases.NewMessagingUseCase(conversationRepo, messageRepo, messagingHub, logger, messageNotifier, s3Client)
	logger.Info("Messaging module initialized", nil)

	// Initialize files module
	var fileUseCase *filesUsecases.FileUseCase
	var versionUseCase *filesUsecases.VersionUseCase
	if s3Client != nil {
		fileMetadataRepo := filesPersistence.NewFileMetadataRepositoryPG(db)
		fileVersionRepo := filesPersistence.NewFileVersionRepositoryPG(db)
		// Настройки валидатора файлов по умолчанию
		fileValidatorConfig := storage.FileValidatorConfig{
			MaxFileSize: 100 * 1024 * 1024, // 100 MB
			AllowedMimeTypes: []string{
				"application/pdf",
				"application/msword",
				"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
				"application/vnd.ms-excel",
				"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
				"application/vnd.ms-powerpoint",
				"application/vnd.openxmlformats-officedocument.presentationml.presentation",
				"text/plain",
				"text/csv",
				"image/jpeg",
				"image/png",
				"image/gif",
				"image/webp",
				"application/zip",
				"application/x-rar-compressed",
			},
			AllowedExtensions: []string{
				".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx",
				".txt", ".csv", ".jpg", ".jpeg", ".png", ".gif", ".webp",
				".zip", ".rar",
			},
		}
		fileValidator := storage.NewFileValidator(fileValidatorConfig)
		fileUseCase = filesUsecases.NewFileUseCase(fileMetadataRepo, fileVersionRepo, s3Client, fileValidator, auditLogger)
		versionUseCase = filesUsecases.NewVersionUseCase(fileMetadataRepo, fileVersionRepo, s3Client, auditLogger)
		logger.Info("Files module initialized", nil)
	} else {
		logger.Warn("Files module not initialized - S3 storage not available", nil)
	}

	// Initialize shared middleware
	corsMiddleware := appMiddleware.NewCORSMiddleware(
		cfg.CORS.AllowedOrigins,
		cfg.CORS.AllowedMethods,
		cfg.CORS.AllowedHeaders,
	)
	loggingMiddleware := appMiddleware.NewLoggingMiddleware(logger)

	// Initialize validator
	validator := validation.NewValidator()

	// JWT secret reused for token revocation use case (setupRoutes builds
	// LogoutUseCase with the same secret used to sign access tokens).
	jwtSecret := []byte(cfg.JWT.AccessSecret)

	// Setup router with all middleware
	router, telegramPollingService := setupRoutes(
		authUseCase,
		docUseCase,
		sharingUseCase,
		docVersionUseCase,
		tagUseCase,
		templateUseCase,
		reportUseCase,
		customReportUseCase,
		taskUseCase,
		projectUseCase,
		eventUseCase,
		lessonUseCase,
		announcementUseCase,
		dashboardUseCase,
		analyticsUseCase,
		userUseCase,
		departmentUseCase,
		positionUseCase,
		fileUseCase,
		versionUseCase,
		notificationUseCase,
		preferencesUseCase,
		telegramVerificationService,
		telegramService,
		webpushRepo,
		webpushService,
		messagingUseCase,
		messagingHub,
		s3Client,
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
		userProfileRepo,
		validator,
		jwtSecret,
	)

	// Initialize integration module (1C synchronization)
	var integrationModule *integration.Module
	integrationModule, err = integration.NewModule(db, &cfg.Integration, logger)
	if err != nil {
		logger.Error("Failed to initialize integration module", map[string]interface{}{
			"error": err.Error(),
		})
	} else if integrationModule.IsEnabled() {
		// Register routes under protected API group with admin guard.
		// Only system_admin role may invoke 1C sync, browse external entities
		// or resolve conflicts — see AUDIT_REPORT critical item #3.
		apiGroup := router.Group("/api")
		apiGroup.Use(authMiddleware.JWTMiddleware(authUseCase))
		integrationModule.RegisterRoutes(apiGroup, authMiddleware.RequireRole(string(authDomain.RoleSystemAdmin)))

		// Start scheduler for periodic sync
		if err := integrationModule.StartScheduler(context.Background()); err != nil {
			logger.Error("Failed to start integration scheduler", map[string]interface{}{
				"error": err.Error(),
			})
		}
		logger.Info("Integration module initialized", nil)
	}

	// Initialize AI module (RAG/Chat functionality)
	if cfg.AI.Enabled {
		// Initialize AI repositories
		aiEmbeddingRepo := aiPersistence.NewEmbeddingRepositoryPg(db)
		aiConversationRepo := aiPersistence.NewConversationRepositoryPg(db)
		aiMessageRepo := aiPersistence.NewMessageRepositoryPg(db)

		// Initialize document adapter for AI (uses documents module)
		aiDocRepo := docPersistence.NewDocumentRepositoryPG(db)
		textExtractionService := aiServices.NewTextExtractionService()
		documentAdapter := aiAdapters.NewDocumentAdapter(aiDocRepo, s3Client, textExtractionService, db, slog.Default())

		// Initialize AI providers based on configuration

		// Embeddings provider
		var embeddingProvider ports.EmbeddingProvider
		if cfg.AI.EmbeddingProvider == "gemini" {
			embeddingProvider = aiProviders.NewGeminiEmbeddingProvider(aiProviders.GeminiEmbeddingConfig{
				APIKey:               cfg.AI.EmbeddingAPIKey,
				Model:                cfg.AI.EmbeddingModel,
				OutputDimensionality: cfg.AI.EmbeddingDimensionality,
				Timeout:              cfg.AI.Timeout,
			})
		} else {
			embeddingProvider = aiProviders.NewOpenAIProvider(aiProviders.OpenAIConfig{
				APIKey:         cfg.AI.OpenAIAPIKey,
				BaseURL:        cfg.AI.OpenAIBaseURL,
				EmbeddingModel: cfg.AI.EmbeddingModel,
				Timeout:        cfg.AI.Timeout,
			})
		}

		// Wrap embedding provider with fallback if configured
		if cfg.AI.FallbackEmbeddingProvider != "" && cfg.AI.FallbackEmbeddingAPIKey != "" {
			fallbackDim := cfg.AI.FallbackEmbeddingDimensionality
			if fallbackDim == 0 {
				fallbackDim = cfg.AI.EmbeddingDimensionality
			}
			if fallbackDim != cfg.AI.EmbeddingDimensionality {
				slog.Warn("fallback embedding dimensionality differs from primary — similarity search may produce incorrect results",
					"primary", cfg.AI.EmbeddingDimensionality,
					"fallback", fallbackDim)
			}

			var fallbackEmbedding ports.EmbeddingProvider
			if cfg.AI.FallbackEmbeddingProvider == "gemini" {
				fallbackEmbedding = aiProviders.NewGeminiEmbeddingProvider(aiProviders.GeminiEmbeddingConfig{
					APIKey:               cfg.AI.FallbackEmbeddingAPIKey,
					Model:                cfg.AI.FallbackEmbeddingModel,
					OutputDimensionality: fallbackDim,
					Timeout:              cfg.AI.Timeout,
				})
			} else {
				fallbackEmbedding = aiProviders.NewOpenAIProvider(aiProviders.OpenAIConfig{
					APIKey:         cfg.AI.FallbackEmbeddingAPIKey,
					BaseURL:        cfg.AI.FallbackEmbeddingBaseURL,
					EmbeddingModel: cfg.AI.FallbackEmbeddingModel,
					Timeout:        cfg.AI.Timeout,
				})
			}
			embeddingProvider = aiProviders.NewFallbackEmbeddingProvider(embeddingProvider, fallbackEmbedding, slog.Default())
			slog.Info("AI fallback embedding provider configured",
				"provider", cfg.AI.FallbackEmbeddingProvider,
				"model", cfg.AI.FallbackEmbeddingModel)
		} else if cfg.AI.FallbackEmbeddingProvider != "" && cfg.AI.FallbackEmbeddingAPIKey == "" {
			slog.Warn("AI_FALLBACK_EMBEDDING_PROVIDER is set but AI_FALLBACK_EMBEDDING_API_KEY is empty, skipping embedding fallback")
		}

		// Chat (LLM) provider
		var llmProvider ports.LLMProvider
		switch {
		case cfg.AI.Provider == "anthropic":
			llmProvider = aiProviders.NewAnthropicProvider(aiProviders.AnthropicConfig{
				APIKey:      cfg.AI.AnthropicAPIKey,
				BaseURL:     cfg.AI.AnthropicBaseURL,
				ChatModel:   cfg.AI.ChatModel,
				MaxTokens:   cfg.AI.MaxTokens,
				Temperature: cfg.AI.Temperature,
				Timeout:     cfg.AI.Timeout,
			})
		case cfg.AI.ChatAPIKey != "":
			// Separate OpenAI-compatible provider for chat (e.g. Gemini, Groq)
			llmProvider = aiProviders.NewOpenAIProvider(aiProviders.OpenAIConfig{
				APIKey:      cfg.AI.ChatAPIKey,
				BaseURL:     cfg.AI.ChatBaseURL,
				ChatModel:   cfg.AI.ChatModel,
				MaxTokens:   cfg.AI.MaxTokens,
				Temperature: cfg.AI.Temperature,
				Timeout:     cfg.AI.Timeout,
			})
		default:
			llmProvider = aiProviders.NewOpenAIProvider(aiProviders.OpenAIConfig{
				APIKey:      cfg.AI.OpenAIAPIKey,
				BaseURL:     cfg.AI.OpenAIBaseURL,
				ChatModel:   cfg.AI.ChatModel,
				MaxTokens:   cfg.AI.MaxTokens,
				Temperature: cfg.AI.Temperature,
				Timeout:     cfg.AI.Timeout,
			})
		}

		// Wrap LLM with fallback provider if configured
		if cfg.AI.FallbackProvider != "" && cfg.AI.FallbackAPIKey != "" {
			var fallbackLLM ports.LLMProvider
			if cfg.AI.FallbackProvider == "anthropic" {
				fallbackLLM = aiProviders.NewAnthropicProvider(aiProviders.AnthropicConfig{
					APIKey:      cfg.AI.FallbackAPIKey,
					BaseURL:     cfg.AI.FallbackBaseURL,
					ChatModel:   cfg.AI.FallbackChatModel,
					MaxTokens:   cfg.AI.MaxTokens,
					Temperature: cfg.AI.Temperature,
					Timeout:     cfg.AI.Timeout,
				})
			} else {
				// OpenAI-compatible: groq, openai, etc.
				fallbackLLM = aiProviders.NewOpenAIProvider(aiProviders.OpenAIConfig{
					APIKey:      cfg.AI.FallbackAPIKey,
					BaseURL:     cfg.AI.FallbackBaseURL,
					ChatModel:   cfg.AI.FallbackChatModel,
					MaxTokens:   cfg.AI.MaxTokens,
					Temperature: cfg.AI.Temperature,
					Timeout:     cfg.AI.Timeout,
				})
			}
			llmProvider = aiProviders.NewFallbackLLMProvider(llmProvider, fallbackLLM, slog.Default())
			slog.Info("AI fallback LLM provider configured",
				"provider", cfg.AI.FallbackProvider,
				"model", cfg.AI.FallbackChatModel)
		} else if cfg.AI.FallbackProvider != "" && cfg.AI.FallbackAPIKey == "" {
			slog.Warn("AI_FALLBACK_PROVIDER is set but AI_FALLBACK_API_KEY is empty, skipping LLM fallback")
		}

		// Initialize AI use cases
		chunkSize := cfg.AI.ChunkSize
		if chunkSize <= 0 {
			chunkSize = 512
		}
		chunkOverlap := cfg.AI.ChunkOverlap
		if chunkOverlap <= 0 {
			chunkOverlap = 102
		}
		aiEmbeddingUseCase := aiUsecases.NewEmbeddingUseCase(
			aiEmbeddingRepo,
			embeddingProvider,
			documentAdapter,
			auditLogger,
			cfg.AI.EmbeddingModel,
			aiServices.ChunkingConfig{
				MaxTokens:    chunkSize,
				OverlapRatio: float64(chunkOverlap) / float64(chunkSize),
			},
		)
		if redisCache != nil {
			aiEmbeddingUseCase.SetCache(redisCache)
		}

		// Initialize Metodych personality (prompt provider)
		promptProvider := aiPrompts.NewPromptProvider()

		// Initialize Mood Engine (before ChatUseCase so it can be wired in)
		moodUseCase := aiUsecases.NewMoodUseCase(
			dashboardRepo,
			analyticsRepo,
			redisCache,
			promptProvider,
		)

		aiChatUseCase := aiUsecases.NewChatUseCase(
			aiConversationRepo,
			aiMessageRepo,
			aiEmbeddingRepo,
			aiEmbeddingUseCase,
			llmProvider,
			promptProvider,
			auditLogger,
			aiUsecases.ChatUseCaseOptions{
				ModelName:       cfg.AI.ChatModel,
				SearchTopK:      cfg.AI.SearchTopK,
				SearchThreshold: cfg.AI.SearchThreshold,
				MoodUseCase:     moodUseCase,
			},
		)

		// Initialize Fun Facts
		funFactRepo := aiPersistence.NewFunFactRepositoryPg(db)
		funFactUseCase := aiUsecases.NewFunFactUseCase(funFactRepo, promptProvider)

		// Seed fun facts if table is empty
		funFactSeeder := aiAdapters.NewFunFactSeeder(funFactRepo, slog.Default())
		if err := funFactSeeder.SeedIfEmpty(context.Background()); err != nil {
			logger.Warn("Failed to seed fun facts", map[string]interface{}{"error": err.Error()})
		}

		// Set personality on notification use case
		notificationUseCase.SetPersonalityProvider(promptProvider)

		// Initialize Telegram personality service (decorator)
		var telegramPersonalityService *aiServices.TelegramPersonalityService
		if telegramService != nil {
			telegramPersonalityService = aiServices.NewTelegramPersonalityService(
				telegramService,
				promptProvider,
			)
		}

		// Initialize fact scheduler for daily fact delivery
		if telegramPersonalityService != nil {
			factScheduler, err := aiScheduler.NewFactScheduler(
				funFactUseCase,
				moodUseCase,
				telegramPersonalityService,
				telegramRepo,
				slog.Default(),
			)
			if err != nil {
				logger.Warn("Failed to create fact scheduler", map[string]interface{}{"error": err.Error()})
			} else {
				if err := factScheduler.Start(); err != nil {
					logger.Warn("Failed to start fact scheduler", map[string]interface{}{"error": err.Error()})
				} else {
					logger.Info("Fact scheduler started", nil)
				}
			}
		}

		// Initialize indexing scheduler for automatic document indexing
		indexingScheduler, err := aiScheduler.NewIndexingScheduler(
			aiEmbeddingUseCase,
			10,
			slog.Default(),
		)
		if err != nil {
			logger.Warn("Failed to create indexing scheduler", map[string]interface{}{"error": err.Error()})
		} else {
			if err := indexingScheduler.Start(); err != nil {
				logger.Warn("Failed to start indexing scheduler", map[string]interface{}{"error": err.Error()})
			} else {
				logger.Info("Indexing scheduler started", nil)
			}
		}

		// Initialize AI handler
		aiHandlerInstance := aiHandler.NewAIHandler(aiChatUseCase, aiEmbeddingUseCase, moodUseCase, funFactUseCase, auditLogger)

		// Register AI routes under protected API group
		aiAPIGroup := router.Group("/api")
		aiAPIGroup.Use(authMiddleware.JWTMiddleware(authUseCase))
		aiHandlerInstance.RegisterRoutes(aiAPIGroup)

		logger.Info("AI module initialized", map[string]interface{}{
			"provider":        cfg.AI.Provider,
			"embedding_model": cfg.AI.EmbeddingModel,
			"chat_model":      cfg.AI.ChatModel,
			"personality":     "Metodych",
		})
	}

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

	// Stop Telegram polling if running
	if telegramPollingService != nil {
		telegramPollingService.Stop()
		logger.Info("Telegram polling service stopped", nil)
	}

	// Stop reminder scheduler
	if reminderScheduler != nil {
		if err := reminderScheduler.Stop(); err != nil {
			logger.Error("Failed to stop reminder scheduler", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			logger.Info("Reminder scheduler stopped", nil)
		}
	}

	// Stop integration scheduler
	if integrationModule != nil && integrationModule.IsEnabled() {
		if err := integrationModule.StopScheduler(); err != nil {
			logger.Error("Failed to stop integration scheduler", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			logger.Info("Integration scheduler stopped", nil)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", map[string]interface{}{
			"error": err.Error(),
		})
	}

	if tracer != nil {
		if err := tracer.Shutdown(context.Background()); err != nil {
			logger.Error("Failed to shutdown tracer", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}
	sentry.Flush(2 * time.Second)
	logger.Info("Server stopped", nil)
}

func initDatabase(cfg *config.Config, _ *logging.Logger, tracingEnabled bool) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Database,
	)

	var db *sql.DB
	var err error

	if tracingEnabled {
		// Use otelsql for database tracing — all queries appear as spans in traces
		db, err = otelsql.Open("postgres", dsn,
			otelsql.WithAttributes(semconv.DBSystemPostgreSQL),
		)
	} else {
		db, err = sql.Open("postgres", dsn)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Report DB connection pool metrics to OpenTelemetry
	if tracingEnabled {
		_, _ = otelsql.RegisterDBStatsMetrics(db, otelsql.WithAttributes(semconv.DBSystemPostgreSQL))
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
	// Enable Redis tracing if distributed tracing is enabled
	redisCache, err := cache.NewRedisCacheWithTracing(redisAddr, cfg.Redis.Password, cfg.Redis.DB, cfg.Tracing.Enabled)
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
	notificationUseCase *notifUsecases.NotificationUseCase,
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
		notificationUseCase,
	)

	return authUseCase, userRepo.(repositories.UserRepository)
}

func setupRoutes(
	authUseCase *usecases.AuthUseCase,
	docUseCase *docUsecases.DocumentUseCase,
	sharingUseCase *docUsecases.SharingUseCase,
	docVersionUseCase *docUsecases.DocumentVersionUseCase,
	tagUseCase *docUsecases.TagUseCase,
	templateUseCase *docUsecases.TemplateUseCase,
	reportUseCase *reportUsecases.ReportUseCase,
	customReportUseCase *reportUsecases.CustomReportUseCase,
	taskUseCase *taskUsecases.TaskUseCase,
	projectUseCase *taskUsecases.ProjectUseCase,
	eventUseCase *scheduleUsecases.EventUseCase,
	lessonUseCase *scheduleUsecases.LessonUseCase,
	announcementUseCase *announcementUsecases.AnnouncementUseCase,
	dashboardUseCase *dashboardUsecases.DashboardUseCase,
	analyticsUseCase *analyticsUsecases.AnalyticsUseCase,
	userUseCase *usersUsecases.UserUseCase,
	departmentUseCase *usersUsecases.DepartmentUseCase,
	positionUseCase *usersUsecases.PositionUseCase,
	fileUseCase *filesUsecases.FileUseCase,
	versionUseCase *filesUsecases.VersionUseCase,
	notificationUseCase *notifUsecases.NotificationUseCase,
	preferencesUseCase *notifUsecases.PreferencesUseCase,
	telegramVerificationService *notifServices.TelegramVerificationService,
	telegramService emailDomain.TelegramService,
	webpushRepo notifRepositories.WebPushRepository,
	webpushService emailDomain.WebPushService,
	messagingUseCase *messagingUsecases.MessagingUseCase,
	messagingHub *messagingWebsocket.Hub,
	s3Client *storage.S3Client,
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
	userProfileRepo usersRepositories.UserProfileRepository,
	validator *validation.Validator,
	jwtSecret []byte,
) (*gin.Engine, *telegram.PollingService) {
	router := gin.New()
	var telegramPollingService *telegram.PollingService

	// Token revocation infrastructure (logout endpoint, AUDIT_REPORT item #4).
	// If Redis is unavailable, revokedTokenRepo stays nil and JWTMiddlewareWithRevocation
	// gracefully degrades to plain validation without blacklist lookup.
	var revokedTokenRepo repositories.RevokedTokenRepository
	var logoutUseCase *usecases.LogoutUseCase
	if redisCache != nil {
		revokedTokenRepo = persistence.NewRedisRevokedTokenRepository(redisCache.Client())
		logoutUseCase = usecases.NewLogoutUseCase(revokedTokenRepo, jwtSecret)
	}

	// Global middleware stack (order matters!)
	router.Use(gin.Recovery())
	// Sentry middleware для отслеживания ошибок
	if os.Getenv("SENTRY_DSN") != "" {
		router.Use(sentrygin.New(sentrygin.Options{
			Repanic: true, // Перевыбрасывать панику после обработки
		}))
	}
	router.Use(corsMiddleware.Handler())         // CORS должен быть первым для обработки OPTIONS
	router.Use(middleware.RequestIDMiddleware()) // Request ID для трейсинга
	// OpenTelemetry tracing middleware (if enabled)
	if cfg.Tracing.Enabled {
		router.Use(middleware.TracingMiddleware(cfg.Tracing.ServiceName))
	}
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
	router.GET("/metrics", metrics.Handler())

	// Swagger documentation endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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
		emailService = notifServices.NewComposioEmailService(composioAPIKey, composioEntityID, auditLogger)
		logger.Info("Email service initialized", nil)
	}

	// Password recovery flow (v0.108.0). Requires Redis (token store)
	// and the email service (delivery). Either dependency missing skips
	// route registration; the frontend forgot-password page then degrades
	// to "service temporarily unavailable" via the absent 4xx route.
	var passwordResetUseCase *usecases.PasswordResetUseCase
	if redisCache != nil && emailService != nil {
		passwordResetTokenRepo := persistence.NewRedisPasswordResetTokenRepository(redisCache.Client())
		passwordResetUseCase = usecases.NewPasswordResetUseCase(userRepo, passwordResetTokenRepo, emailService)
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

		// Logout: requires JWT auth (no revocation check — otherwise a
		// already-revoked token couldn't call logout idempotently).
		// Endpoint blacklists the access token's JTI in Redis until exp.
		if logoutUseCase != nil {
			logoutHandlerInstance := authHandler.NewLogoutHandler(logoutUseCase)
			authGroup.POST("/logout",
				authMiddleware.JWTMiddleware(authUseCase),
				logoutHandlerInstance.Logout,
			)
			authGroup.OPTIONS("/logout", func(c *gin.Context) { c.Status(http.StatusNoContent) })
		}

		// Password recovery (v0.108.0). All three endpoints stay public
		// (a user who forgot their password has no token); rate-limited
		// via the same authGroup limiter so an attacker cannot spam the
		// request endpoint to enumerate emails by side effects.
		if passwordResetUseCase != nil {
			pwResetHandler := authHandler.NewPasswordResetHandler(passwordResetUseCase)
			authGroup.POST("/password-reset/request", pwResetHandler.RequestReset)
			authGroup.GET("/password-reset/verify/:token", pwResetHandler.VerifyResetToken)
			authGroup.POST("/password-reset/confirm", pwResetHandler.ConfirmReset)

			authGroup.OPTIONS("/password-reset/request", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			authGroup.OPTIONS("/password-reset/verify/:token", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			authGroup.OPTIONS("/password-reset/confirm", func(c *gin.Context) { c.Status(http.StatusNoContent) })
		}
	}

	// Public document access routes (no authentication required)
	if sharingUseCase != nil {
		publicSharingHandler := docHandler.NewSharingHandler(sharingUseCase, validator)
		publicGroup := router.Group("/api/public")
		if publicRateLimiter != nil {
			publicGroup.Use(publicRateLimiter.RateLimitMiddleware())
		}
		{
			publicGroup.POST("/documents/:token", publicSharingHandler.AccessPublicDocument)
			publicGroup.OPTIONS("/documents/:token", func(c *gin.Context) { c.Status(http.StatusNoContent) })
		}
		logger.Info("Public document access routes registered", nil)
	}

	// Telegram webhook (public - receives updates from Telegram servers)
	if telegramVerificationService != nil && telegramService != nil {
		webhookHandler := notifHttp.NewTelegramWebhookHandler(
			telegramVerificationService,
			telegramService,
			cfg.Telegram.WebhookSecret,
			slog.Default(),
			nil, // telegramPersonalityService - set later if AI module is enabled
			nil, // moodUseCase - set later if AI module is enabled
		)
		router.POST("/api/telegram/webhook", webhookHandler.HandleWebhook)
		logger.Info("Telegram webhook route registered", nil)

		// If no webhook URL is configured, use polling mode for local development
		if cfg.Telegram.WebhookURL == "" && cfg.Telegram.BotToken != "" {
			telegramPollingService = telegram.NewPollingService(cfg.Telegram.BotToken, slog.Default())
			telegramPollingService.SetHandler(webhookHandler.ProcessUpdate)
			if err := telegramPollingService.Start(context.Background()); err != nil {
				logger.Warn("Failed to start Telegram polling", map[string]interface{}{"error": err.Error()})
			} else {
				logger.Info("Telegram polling mode started (no webhook URL configured)", nil)
			}
		}
	}

	// Protected routes (require JWT) with auth rate limiting (60 req/min + burst 10)
	protectedGroup := router.Group("/api")
	protectedGroup.Use(authMiddleware.JWTMiddlewareWithRevocation(authUseCase, revokedTokenRepo))
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

			// Get full user data with profile from database
			user, err := userProfileRepo.GetProfileByID(c.Request.Context(), userID.(int64))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user data"})
				return
			}

			// Generate presigned URL for avatar if it exists
			avatarURL := ""
			if user.Avatar != "" && s3Client != nil {
				var err error
				avatarURL, err = s3Client.GetPresignedURL(c.Request.Context(), user.Avatar, 7*24*time.Hour)
				if err != nil {
					logger.Warn("Failed to generate presigned URL for avatar", map[string]interface{}{
						"user_id": user.ID,
						"avatar":  user.Avatar,
						"error":   err.Error(),
					})
				} else {
					logger.Info("Generated avatar URL", map[string]interface{}{
						"user_id":    user.ID,
						"avatar_key": user.Avatar,
						"avatar_url": avatarURL,
					})
				}
			}

			c.JSON(http.StatusOK, gin.H{
				"id":         user.ID,
				"email":      user.Email,
				"name":       user.Name,
				"role":       user.Role,
				"phone":      user.Phone,
				"bio":        user.Bio,
				"avatar":     avatarURL,
				"created_at": user.CreatedAt,
				"updated_at": user.UpdatedAt,
			})
		})
		protectedGroup.OPTIONS("/me", func(c *gin.Context) { c.Status(http.StatusNoContent) })

		// Notifications module routes
		if notificationUseCase != nil {
			notificationHandler := notifHttp.NewNotificationHandler(notificationUseCase)
			preferencesHandler := notifHttp.NewPreferencesHandler(preferencesUseCase)

			notificationsGroup := protectedGroup.Group("/notifications")
			{
				// Notification CRUD and management
				notificationsGroup.GET("", notificationHandler.List)
				notificationsGroup.GET("/:id", notificationHandler.GetByID)
				notificationsGroup.PUT("/:id/read", notificationHandler.MarkAsRead)
				notificationsGroup.PUT("/read-all", notificationHandler.MarkAllAsRead)
				notificationsGroup.DELETE("/:id", notificationHandler.Delete)
				notificationsGroup.DELETE("", notificationHandler.DeleteAll)
				notificationsGroup.GET("/unread-count", notificationHandler.GetUnreadCount)
				notificationsGroup.GET("/stats", notificationHandler.GetStats)

				// Preferences routes
				notificationsGroup.GET("/preferences", preferencesHandler.Get)
				notificationsGroup.PUT("/preferences", preferencesHandler.Update)
				notificationsGroup.PUT("/preferences/channel", preferencesHandler.ToggleChannel)
				notificationsGroup.PUT("/preferences/quiet-hours", preferencesHandler.UpdateQuietHours)
				notificationsGroup.POST("/preferences/reset", preferencesHandler.Reset)
				notificationsGroup.GET("/timezones", preferencesHandler.GetTimezones)

				// CORS preflight handlers
				notificationsGroup.OPTIONS("", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				notificationsGroup.OPTIONS("/:id", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				notificationsGroup.OPTIONS("/:id/read", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				notificationsGroup.OPTIONS("/read-all", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				notificationsGroup.OPTIONS("/unread-count", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				notificationsGroup.OPTIONS("/stats", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				notificationsGroup.OPTIONS("/preferences", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				notificationsGroup.OPTIONS("/preferences/channel", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				notificationsGroup.OPTIONS("/preferences/quiet-hours", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				notificationsGroup.OPTIONS("/preferences/reset", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				notificationsGroup.OPTIONS("/timezones", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			}
			logger.Info("Notifications module routes registered", nil)

			// Telegram routes (protected - for linking accounts)
			if telegramVerificationService != nil {
				telegramHandler := notifHttp.NewTelegramHandler(telegramVerificationService)
				telegramGroup := protectedGroup.Group("/telegram")
				{
					telegramGroup.POST("/verification-code", telegramHandler.GenerateVerificationCode)
					telegramGroup.GET("/status", telegramHandler.GetConnectionStatus)
					telegramGroup.POST("/disconnect", telegramHandler.DisconnectTelegram)

					// CORS preflight handlers
					telegramGroup.OPTIONS("/verification-code", func(c *gin.Context) { c.Status(http.StatusNoContent) })
					telegramGroup.OPTIONS("/status", func(c *gin.Context) { c.Status(http.StatusNoContent) })
					telegramGroup.OPTIONS("/disconnect", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				}
				logger.Info("Telegram API routes registered", nil)
			}

			// Web Push routes (protected - for browser push notifications)
			if webpushService != nil && webpushService.IsConfigured() {
				webpushHandler := notifHttp.NewWebPushHandler(webpushRepo, webpushService)
				pushGroup := notificationsGroup.Group("/push")
				{
					pushGroup.GET("/vapid-key", webpushHandler.GetVAPIDKey)
					pushGroup.POST("/subscribe", webpushHandler.Subscribe)
					pushGroup.POST("/unsubscribe", webpushHandler.Unsubscribe)
					pushGroup.GET("/status", webpushHandler.GetStatus)
					pushGroup.DELETE("/subscriptions/:id", webpushHandler.DeleteSubscription)
					pushGroup.POST("/test", webpushHandler.TestPush)

					// CORS preflight handlers
					pushGroup.OPTIONS("/vapid-key", func(c *gin.Context) { c.Status(http.StatusNoContent) })
					pushGroup.OPTIONS("/subscribe", func(c *gin.Context) { c.Status(http.StatusNoContent) })
					pushGroup.OPTIONS("/unsubscribe", func(c *gin.Context) { c.Status(http.StatusNoContent) })
					pushGroup.OPTIONS("/status", func(c *gin.Context) { c.Status(http.StatusNoContent) })
					pushGroup.OPTIONS("/subscriptions/:id", func(c *gin.Context) { c.Status(http.StatusNoContent) })
					pushGroup.OPTIONS("/test", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				}
				logger.Info("Web Push API routes registered", nil)
			}
		}

		// Email notification routes (legacy - for direct email sending)
		if emailService != nil {
			emailHandler := emailHandlers.NewEmailHandler(emailService)

			emailGroup := protectedGroup.Group("/email")
			{
				emailGroup.POST("/send", emailHandler.SendEmail)
				emailGroup.POST("/send-welcome", emailHandler.SendWelcomeEmail)
				emailGroup.OPTIONS("/send", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				emailGroup.OPTIONS("/send-welcome", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			}
			logger.Info("Email notification routes registered", nil)
		} else {
			logger.Warn("Email notification routes not registered - Composio credentials not configured", nil)
		}

		// Messaging module routes (internal messaging system)
		if messagingUseCase != nil {
			msgHandler := messagingHandler.NewMessagingHandler(messagingUseCase, messagingHub, logger, validator)
			msgHandler.RegisterRoutes(protectedGroup, authMiddleware.JWTMiddleware(authUseCase))
			logger.Info("Messaging module routes registered", nil)
		}

		// Document management routes
		if docUseCase != nil {
			docHandlerInstance := docHandler.NewDocumentHandler(docUseCase)

			documentsGroup := protectedGroup.Group("/documents")
			{
				// Students must not create documents (AUDIT_REPORT item #1).
				// Read paths stay open: students see what is shared with them via ACL.
				documentsGroup.POST("", authMiddleware.RequireNonStudent(), docHandlerInstance.Create)
				documentsGroup.GET("", docHandlerInstance.List)
				// Search route must be before /:id to avoid route conflict
				documentsGroup.GET("/search", docHandlerInstance.Search)
				documentsGroup.GET("/:id", docHandlerInstance.GetByID)
				documentsGroup.PUT("/:id", docHandlerInstance.Update)
				documentsGroup.DELETE("/:id", docHandlerInstance.Delete)
				documentsGroup.POST("/:id/file", docHandlerInstance.UploadFile)
				documentsGroup.GET("/:id/file", docHandlerInstance.DownloadFile)
				documentsGroup.DELETE("/:id/file", docHandlerInstance.DeleteFile)

				// CORS preflight handlers
				documentsGroup.OPTIONS("", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				documentsGroup.OPTIONS("/search", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				documentsGroup.OPTIONS("/:id", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				documentsGroup.OPTIONS("/:id/file", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			}

			// Document sharing routes
			if sharingUseCase != nil {
				sharingHandlerInstance := docHandler.NewSharingHandler(sharingUseCase, validator)

				// Shared documents list
				protectedGroup.GET("/documents/shared", sharingHandlerInstance.GetSharedDocuments)
				protectedGroup.OPTIONS("/documents/shared", func(c *gin.Context) { c.Status(http.StatusNoContent) })

				// My shared documents (documents I shared with others)
				protectedGroup.GET("/documents/my-shared", sharingHandlerInstance.GetMySharedDocuments)
				protectedGroup.OPTIONS("/documents/my-shared", func(c *gin.Context) { c.Status(http.StatusNoContent) })

				// Document permissions routes
				documentsGroup.POST("/:id/share", sharingHandlerInstance.ShareDocument)
				documentsGroup.GET("/:id/permissions", sharingHandlerInstance.GetDocumentPermissions)
				documentsGroup.DELETE("/:id/permissions/:permissionId", sharingHandlerInstance.RevokePermission)

				// Public links routes
				documentsGroup.POST("/:id/public-links", sharingHandlerInstance.CreatePublicLink)
				documentsGroup.GET("/:id/public-links", sharingHandlerInstance.GetDocumentPublicLinks)
				documentsGroup.POST("/:id/public-links/:linkId/deactivate", sharingHandlerInstance.DeactivatePublicLink)
				documentsGroup.DELETE("/:id/public-links/:linkId", sharingHandlerInstance.DeletePublicLink)

				// CORS preflight handlers for sharing routes
				protectedGroup.OPTIONS("/documents/:id/share", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				protectedGroup.OPTIONS("/documents/:id/permissions", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				protectedGroup.OPTIONS("/documents/:id/permissions/:permissionId", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				protectedGroup.OPTIONS("/documents/:id/public-links", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				protectedGroup.OPTIONS("/documents/:id/public-links/:linkId", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				protectedGroup.OPTIONS("/documents/:id/public-links/:linkId/deactivate", func(c *gin.Context) { c.Status(http.StatusNoContent) })

				logger.Info("Document sharing routes registered", nil)
			}

			// Document version control routes
			if docVersionUseCase != nil {
				versionHandlerInstance := docHandler.NewVersionHandler(docVersionUseCase)

				// Version routes - must be before /:id to avoid route conflicts
				documentsGroup.GET("/:id/versions/compare", versionHandlerInstance.CompareVersions)
				documentsGroup.GET("/:id/versions", versionHandlerInstance.GetVersions)
				documentsGroup.POST("/:id/versions", versionHandlerInstance.CreateVersion)
				documentsGroup.GET("/:id/versions/:version", versionHandlerInstance.GetVersion)
				documentsGroup.DELETE("/:id/versions/:version", versionHandlerInstance.DeleteVersion)
				documentsGroup.POST("/:id/versions/:version/restore", versionHandlerInstance.RestoreVersion)
				documentsGroup.GET("/:id/versions/:version/file", versionHandlerInstance.GetVersionFile)

				// CORS preflight handlers for version routes
				documentsGroup.OPTIONS("/:id/versions", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				documentsGroup.OPTIONS("/:id/versions/compare", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				documentsGroup.OPTIONS("/:id/versions/:version", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				documentsGroup.OPTIONS("/:id/versions/:version/restore", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				documentsGroup.OPTIONS("/:id/versions/:version/file", func(c *gin.Context) { c.Status(http.StatusNoContent) })

				logger.Info("Document version control routes registered", nil)
			}

			// Document tags routes
			if tagUseCase != nil {
				tagHandlerInstance := docHandler.NewTagHandler(tagUseCase)

				// Tags CRUD routes
				tagsGroup := protectedGroup.Group("/tags")
				{
					tagsGroup.POST("", tagHandlerInstance.Create)
					tagsGroup.GET("", tagHandlerInstance.GetAll)
					tagsGroup.GET("/search", tagHandlerInstance.Search)
					tagsGroup.GET("/:id", tagHandlerInstance.GetByID)
					tagsGroup.PUT("/:id", tagHandlerInstance.Update)
					tagsGroup.DELETE("/:id", tagHandlerInstance.Delete)
					tagsGroup.GET("/:id/documents", tagHandlerInstance.GetDocumentsByTag)

					// CORS preflight handlers
					tagsGroup.OPTIONS("", func(c *gin.Context) { c.Status(http.StatusNoContent) })
					tagsGroup.OPTIONS("/search", func(c *gin.Context) { c.Status(http.StatusNoContent) })
					tagsGroup.OPTIONS("/:id", func(c *gin.Context) { c.Status(http.StatusNoContent) })
					tagsGroup.OPTIONS("/:id/documents", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				}

				// Document-specific tag routes
				documentsGroup.GET("/:id/tags", tagHandlerInstance.GetDocumentTags)
				documentsGroup.PUT("/:id/tags", tagHandlerInstance.SetDocumentTags)
				documentsGroup.POST("/:id/tags/:tag_id", tagHandlerInstance.AddTagToDocument)
				documentsGroup.DELETE("/:id/tags/:tag_id", tagHandlerInstance.RemoveTagFromDocument)

				// CORS preflight handlers for document tag routes
				documentsGroup.OPTIONS("/:id/tags", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				documentsGroup.OPTIONS("/:id/tags/:tag_id", func(c *gin.Context) { c.Status(http.StatusNoContent) })

				logger.Info("Document tags routes registered", nil)
			}

			// Document types and categories (reference data)
			protectedGroup.GET("/document-types", docHandlerInstance.GetDocumentTypes)
			protectedGroup.GET("/document-categories", docHandlerInstance.GetCategories)
			protectedGroup.OPTIONS("/document-types", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			protectedGroup.OPTIONS("/document-categories", func(c *gin.Context) { c.Status(http.StatusNoContent) })

			// Document templates routes
			if templateUseCase != nil {
				templateHandlerInstance := docHandler.NewTemplateHandler(templateUseCase, validator)

				templatesGroup := protectedGroup.Group("/templates")
				{
					templatesGroup.GET("", templateHandlerInstance.GetTemplates)
					templatesGroup.GET("/:id", templateHandlerInstance.GetTemplate)
					templatesGroup.POST("/:id/preview", templateHandlerInstance.PreviewTemplate)
					templatesGroup.POST("/:id/create", templateHandlerInstance.CreateDocumentFromTemplate)
					templatesGroup.PUT("/:id", templateHandlerInstance.UpdateTemplate)

					// CORS preflight handlers
					templatesGroup.OPTIONS("", func(c *gin.Context) { c.Status(http.StatusNoContent) })
					templatesGroup.OPTIONS("/:id", func(c *gin.Context) { c.Status(http.StatusNoContent) })
					templatesGroup.OPTIONS("/:id/preview", func(c *gin.Context) { c.Status(http.StatusNoContent) })
					templatesGroup.OPTIONS("/:id/create", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				}
				logger.Info("Document templates routes registered", nil)
			}

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
			// Students have no business with reports — block the entire group
			// (AUDIT_REPORT item #1).
			reportsGroup.Use(authMiddleware.RequireNonStudent())
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

		// Custom reports module routes
		if customReportUseCase != nil {
			customReportHandlerInstance := reportHandler.NewCustomReportHandler(customReportUseCase)

			customReportsGroup := protectedGroup.Group("/custom-reports")
			{
				// CRUD operations
				customReportsGroup.POST("", customReportHandlerInstance.Create)
				customReportsGroup.GET("", customReportHandlerInstance.List)
				customReportsGroup.GET("/:id", customReportHandlerInstance.GetByID)
				customReportsGroup.PUT("/:id", customReportHandlerInstance.Update)
				customReportsGroup.DELETE("/:id", customReportHandlerInstance.Delete)

				// Report execution and export
				customReportsGroup.POST("/:id/execute", customReportHandlerInstance.Execute)
				customReportsGroup.POST("/:id/export", customReportHandlerInstance.Export)

				// Special queries
				customReportsGroup.GET("/my", customReportHandlerInstance.GetMyReports)
				customReportsGroup.GET("/public", customReportHandlerInstance.GetPublicReports)
				customReportsGroup.GET("/available-fields", customReportHandlerInstance.GetAvailableFields)

				// CORS preflight handlers
				customReportsGroup.OPTIONS("", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				customReportsGroup.OPTIONS("/:id", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				customReportsGroup.OPTIONS("/:id/execute", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				customReportsGroup.OPTIONS("/:id/export", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				customReportsGroup.OPTIONS("/my", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				customReportsGroup.OPTIONS("/public", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				customReportsGroup.OPTIONS("/available-fields", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			}

			logger.Info("Custom reports module routes registered", nil)
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

		// Schedule lessons module routes
		if lessonUseCase != nil {
			lessonHandlerInstance := scheduleHandler.NewLessonHandler(lessonUseCase)

			scheduleGroup := protectedGroup.Group("/schedule")
			{
				scheduleGroup.POST("/lessons", lessonHandlerInstance.Create)
				scheduleGroup.GET("/lessons", lessonHandlerInstance.List)
				scheduleGroup.GET("/lessons/timetable", lessonHandlerInstance.GetTimetable)
				scheduleGroup.GET("/lessons/:id", lessonHandlerInstance.GetByID)
				scheduleGroup.PUT("/lessons/:id", lessonHandlerInstance.Update)
				scheduleGroup.DELETE("/lessons/:id", lessonHandlerInstance.Delete)
				scheduleGroup.POST("/changes", lessonHandlerInstance.CreateChange)
				scheduleGroup.GET("/changes", lessonHandlerInstance.ListChanges)

				scheduleGroup.OPTIONS("/lessons", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				scheduleGroup.OPTIONS("/lessons/timetable", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				scheduleGroup.OPTIONS("/lessons/:id", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				scheduleGroup.OPTIONS("/changes", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			}

			classroomsGroup := protectedGroup.Group("/classrooms")
			{
				classroomsGroup.GET("", lessonHandlerInstance.ListClassrooms)
				classroomsGroup.OPTIONS("", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			}

			studentGroupsGroup := protectedGroup.Group("/student-groups")
			{
				studentGroupsGroup.GET("", lessonHandlerInstance.ListStudentGroups)
				studentGroupsGroup.OPTIONS("", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			}

			disciplinesGroup := protectedGroup.Group("/disciplines")
			{
				disciplinesGroup.GET("", lessonHandlerInstance.ListDisciplines)
				disciplinesGroup.OPTIONS("", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			}

			semestersGroup := protectedGroup.Group("/semesters")
			{
				semestersGroup.GET("", lessonHandlerInstance.ListSemesters)
				semestersGroup.OPTIONS("", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			}

			lessonTypesGroup := protectedGroup.Group("/lesson-types")
			{
				lessonTypesGroup.GET("", lessonHandlerInstance.ListLessonTypes)
				lessonTypesGroup.OPTIONS("", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			}

			logger.Info("Schedule lessons routes registered", nil)
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

				// Attachments
				announcementsGroup.POST("/:id/attachments", announcementHandlerInstance.UploadAttachment)
				announcementsGroup.DELETE("/:id/attachments/:attachmentID", announcementHandlerInstance.DeleteAttachment)

				// CORS preflight handlers
				announcementsGroup.OPTIONS("", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				announcementsGroup.OPTIONS("/:id", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				announcementsGroup.OPTIONS("/published", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				announcementsGroup.OPTIONS("/pinned", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				announcementsGroup.OPTIONS("/recent", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				announcementsGroup.OPTIONS("/:id/publish", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				announcementsGroup.OPTIONS("/:id/unpublish", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				announcementsGroup.OPTIONS("/:id/archive", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				announcementsGroup.OPTIONS("/:id/attachments", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				announcementsGroup.OPTIONS("/:id/attachments/:attachmentID", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			}

			logger.Info("Announcements module routes registered", nil)
		}

		// Dashboard module routes
		if dashboardUseCase != nil {
			dashboardHandlerInstance := dashboardHandler.NewDashboardHandler(dashboardUseCase)

			dashboardGroup := protectedGroup.Group("/dashboard")
			{
				dashboardGroup.GET("/stats", dashboardHandlerInstance.GetStats)
				dashboardGroup.GET("/trends", dashboardHandlerInstance.GetTrends)
				dashboardGroup.GET("/activity", dashboardHandlerInstance.GetActivity)
				dashboardGroup.POST("/export", dashboardHandlerInstance.Export)

				// CORS preflight handlers
				dashboardGroup.OPTIONS("/stats", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				dashboardGroup.OPTIONS("/trends", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				dashboardGroup.OPTIONS("/activity", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				dashboardGroup.OPTIONS("/export", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			}

			logger.Info("Dashboard module routes registered", nil)
		}

		// Analytics module routes (predictive analytics for students)
		if analyticsUseCase != nil {
			analyticsHandlerInstance := analyticsHandler.NewAnalyticsHandler(analyticsUseCase)
			attendanceHandlerInstance := analyticsHandler.NewAttendanceHandler(analyticsUseCase)

			// Analytics routes — students must not see at-risk lists or any
			// aggregated analytics about themselves or their peers
			// (AUDIT_REPORT item #1).
			analyticsGroup := protectedGroup.Group("/analytics")
			analyticsGroup.Use(authMiddleware.RequireNonStudent())
			{
				analyticsGroup.GET("/at-risk-students", analyticsHandlerInstance.GetAtRiskStudents)
				analyticsGroup.GET("/students/:id/risk", analyticsHandlerInstance.GetStudentRisk)
				analyticsGroup.GET("/groups/summary", analyticsHandlerInstance.GetAllGroupsSummary)
				analyticsGroup.GET("/groups/:name/summary", analyticsHandlerInstance.GetGroupSummary)
				analyticsGroup.GET("/risk-level/:level", analyticsHandlerInstance.GetStudentsByRiskLevel)
				analyticsGroup.GET("/attendance-trend", analyticsHandlerInstance.GetAttendanceTrend)
				analyticsGroup.GET("/students/:id/risk-history", analyticsHandlerInstance.GetStudentRiskHistory)
				analyticsGroup.GET("/export", analyticsHandlerInstance.ExportAtRiskStudents)
				analyticsGroup.GET("/config/weights", analyticsHandlerInstance.GetRiskWeightConfig)
				analyticsGroup.PUT("/config/weights", analyticsHandlerInstance.UpdateRiskWeightConfig)

				// CORS preflight handlers
				analyticsGroup.OPTIONS("/at-risk-students", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				analyticsGroup.OPTIONS("/students/:id/risk", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				analyticsGroup.OPTIONS("/groups/summary", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				analyticsGroup.OPTIONS("/groups/:name/summary", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				analyticsGroup.OPTIONS("/risk-level/:level", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				analyticsGroup.OPTIONS("/attendance-trend", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			}

			// Attendance routes
			attendanceGroup := protectedGroup.Group("/attendance")
			{
				attendanceGroup.POST("/mark", attendanceHandlerInstance.MarkAttendance)
				attendanceGroup.POST("/bulk", attendanceHandlerInstance.BulkMarkAttendance)
				attendanceGroup.GET("/lesson/:id/date/:date", attendanceHandlerInstance.GetLessonAttendance)
				attendanceGroup.POST("/lessons", attendanceHandlerInstance.CreateLesson)

				// CORS preflight handlers
				attendanceGroup.OPTIONS("/mark", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				attendanceGroup.OPTIONS("/bulk", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				attendanceGroup.OPTIONS("/lesson/:id/date/:date", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				attendanceGroup.OPTIONS("/lessons", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			}

			logger.Info("Analytics module routes registered", nil)
		}

		// Users module routes (system admin only for management)
		if userUseCase != nil {
			userHandlerInstance := usersHandler.NewUserHandler(userUseCase)
			departmentHandlerInstance := usersHandler.NewDepartmentHandler(departmentUseCase)
			positionHandlerInstance := usersHandler.NewPositionHandler(positionUseCase)
			avatarHandlerInstance := usersHandler.NewAvatarHandler(userUseCase, s3Client)

			// Users management routes
			usersGroup := protectedGroup.Group("/users")
			{
				usersGroup.GET("", userHandlerInstance.List)
				usersGroup.GET("/:id", userHandlerInstance.GetByID)
				usersGroup.PUT("/:id/profile", userHandlerInstance.UpdateProfile)
				usersGroup.PUT("/:id/role", userHandlerInstance.UpdateRole)
				usersGroup.PUT("/:id/status", userHandlerInstance.UpdateStatus)
				usersGroup.DELETE("/:id", userHandlerInstance.Delete)
				usersGroup.POST("/bulk/department", userHandlerInstance.BulkUpdateDepartment)
				usersGroup.POST("/bulk/position", userHandlerInstance.BulkUpdatePosition)
				usersGroup.GET("/by-department/:id", userHandlerInstance.GetByDepartment)
				usersGroup.GET("/by-position/:id", userHandlerInstance.GetByPosition)

				// Avatar routes
				usersGroup.POST("/:id/avatar", avatarHandlerInstance.Upload)
				usersGroup.DELETE("/:id/avatar", avatarHandlerInstance.Delete)
				usersGroup.GET("/:id/avatar", avatarHandlerInstance.GetAvatarURL)

				// CORS preflight handlers
				usersGroup.OPTIONS("", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				usersGroup.OPTIONS("/:id", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				usersGroup.OPTIONS("/:id/profile", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				usersGroup.OPTIONS("/:id/role", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				usersGroup.OPTIONS("/:id/status", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				usersGroup.OPTIONS("/:id/avatar", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				usersGroup.OPTIONS("/bulk/department", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				usersGroup.OPTIONS("/bulk/position", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				usersGroup.OPTIONS("/by-department/:id", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				usersGroup.OPTIONS("/by-position/:id", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			}

			// Departments routes
			departmentsGroup := protectedGroup.Group("/departments")
			{
				departmentsGroup.POST("", departmentHandlerInstance.Create)
				departmentsGroup.GET("", departmentHandlerInstance.List)
				departmentsGroup.GET("/:id", departmentHandlerInstance.GetByID)
				departmentsGroup.PUT("/:id", departmentHandlerInstance.Update)
				departmentsGroup.DELETE("/:id", departmentHandlerInstance.Delete)
				departmentsGroup.GET("/:id/children", departmentHandlerInstance.GetChildren)

				// CORS preflight handlers
				departmentsGroup.OPTIONS("", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				departmentsGroup.OPTIONS("/:id", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				departmentsGroup.OPTIONS("/:id/children", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			}

			// Positions routes
			positionsGroup := protectedGroup.Group("/positions")
			{
				positionsGroup.POST("", positionHandlerInstance.Create)
				positionsGroup.GET("", positionHandlerInstance.List)
				positionsGroup.GET("/:id", positionHandlerInstance.GetByID)
				positionsGroup.PUT("/:id", positionHandlerInstance.Update)
				positionsGroup.DELETE("/:id", positionHandlerInstance.Delete)

				// CORS preflight handlers
				positionsGroup.OPTIONS("", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				positionsGroup.OPTIONS("/:id", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			}

			logger.Info("Users module routes registered", nil)
		}

		// Files module routes
		if fileUseCase != nil && versionUseCase != nil {
			fileHandlerInstance := filesHandler.NewFileHandler(fileUseCase, versionUseCase)

			// Files management routes
			filesGroup := protectedGroup.Group("/files")
			{
				filesGroup.POST("/upload", fileHandlerInstance.Upload)
				filesGroup.GET("", fileHandlerInstance.List)
				filesGroup.GET("/:id", fileHandlerInstance.GetByID)
				filesGroup.GET("/:id/download", fileHandlerInstance.Download)
				filesGroup.POST("/:id/attach", fileHandlerInstance.Attach)
				filesGroup.DELETE("/:id", fileHandlerInstance.Delete)

				// Versioning routes
				filesGroup.POST("/:id/versions", fileHandlerInstance.CreateVersion)
				filesGroup.GET("/:id/versions", fileHandlerInstance.GetVersions)
				filesGroup.GET("/:id/versions/:version", fileHandlerInstance.DownloadVersion)

				// Cleanup route (admin only)
				filesGroup.POST("/cleanup", fileHandlerInstance.CleanupExpired)

				// CORS preflight handlers
				filesGroup.OPTIONS("/upload", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				filesGroup.OPTIONS("", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				filesGroup.OPTIONS("/:id", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				filesGroup.OPTIONS("/:id/download", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				filesGroup.OPTIONS("/:id/attach", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				filesGroup.OPTIONS("/:id/versions", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				filesGroup.OPTIONS("/:id/versions/:version", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				filesGroup.OPTIONS("/cleanup", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			}

			// Routes for getting files by entity (under /api/files to avoid route conflicts)
			filesGroup.GET("/by-document/:document_id", fileHandlerInstance.GetByDocument)
			filesGroup.GET("/by-task/:task_id", fileHandlerInstance.GetByTask)
			filesGroup.GET("/by-announcement/:announcement_id", fileHandlerInstance.GetByAnnouncement)
			filesGroup.OPTIONS("/by-document/:document_id", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			filesGroup.OPTIONS("/by-task/:task_id", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			filesGroup.OPTIONS("/by-announcement/:announcement_id", func(c *gin.Context) { c.Status(http.StatusNoContent) })

			logger.Info("Files module routes registered", nil)
		} else {
			logger.Warn("Files module routes not registered - S3 storage not available", nil)
		}

		// Admin only routes — DB stores role as "system_admin" (see migration 001),
		// so RequireRole must use the same value. The previous "admin" string
		// silently failed to match anyone (AUDIT_REPORT, system_admin section).
		adminGroup := protectedGroup.Group("/admin")
		adminGroup.Use(authMiddleware.RequireRole(string(authDomain.RoleSystemAdmin)))
		{
			adminGroup.GET("/users", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "Admin users list"})
			})
			adminGroup.OPTIONS("/users", func(c *gin.Context) { c.Status(http.StatusNoContent) })

			// Admin notification routes (create and broadcast notifications)
			if notificationUseCase != nil {
				notificationHandler := notifHttp.NewNotificationHandler(notificationUseCase)
				adminGroup.POST("/notifications", notificationHandler.Create)
				adminGroup.POST("/notifications/bulk", notificationHandler.CreateBulk)
				adminGroup.OPTIONS("/notifications", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				adminGroup.OPTIONS("/notifications/bulk", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				logger.Info("Admin notification routes registered", nil)
			}
		}
	}

	return router, telegramPollingService
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
