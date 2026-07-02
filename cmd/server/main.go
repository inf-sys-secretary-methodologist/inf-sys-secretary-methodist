// Package main provides the entry point for the Information System Secretary-Methodologist server.
//
// @title           Inf-Sys Secretary-Methodist API
// @version         0.222.0
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
// @tag.name assignments
// @tag.description Учебные задания и выставление оценок (academic Tasks Context)
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
	"errors"

	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/XSAM/otelsql"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"

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
	analyticsEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/analytics/domain/entities"
	analyticsPersistence "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/analytics/infrastructure/persistence"
	analyticsScheduler "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/analytics/infrastructure/scheduler"
	analyticsHandler "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/analytics/interfaces/http/handlers"
	announcementUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/application/usecases"
	announcementsDomain "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/domain"
	announcementPersistence "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/infrastructure/persistence"
	announcementHandler "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/interfaces/http/handlers"
	announcementRoutes "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/announcements/interfaces/http/routes"
	assignUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/application/usecases"
	assignPersistence "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/infrastructure/persistence"
	assignHandler "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/assignments/interfaces/http/handlers"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/application/usecases"
	authDomain "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/domain"
	persistence "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/infrastructure/persistence"
	authHandler "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/interfaces/http/handlers"
	authMiddleware "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/auth/interfaces/http/middleware"
	brandingUseCases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/branding/application/usecases"
	brandingPersistence "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/branding/infrastructure/persistence"
	brandingHandlers "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/branding/interfaces/http/handlers"
	brandingRoutes "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/branding/interfaces/http/routes"
	curUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/application/usecases"
	curPersistence "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/infrastructure/persistence"
	curHandler "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/curriculum/interfaces/http/handlers"
	dashboardUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/dashboard/application/usecases"
	dashboardPersistence "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/dashboard/infrastructure/persistence"
	dashboardHandler "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/dashboard/interfaces/http/handlers"
	docUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/application/usecases"
	docEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/domain/entities"
	docPersistence "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/infrastructure/persistence"
	docSigning "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/infrastructure/signing"
	docHandler "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/documents/interfaces/http/handlers"
	extUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/extracurricular/application/usecases"
	extPersistence "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/extracurricular/infrastructure/persistence"
	extHandler "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/extracurricular/interfaces/http/handlers"
	filesUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/files/application/usecases"
	filesPersistence "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/files/infrastructure/persistence"
	filesHandler "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/files/interfaces/http/handlers"
	integration "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/integration"
	messagingUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/messaging/application/usecases"
	messagingPersistence "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/messaging/infrastructure/persistence"
	messagingWebsocket "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/messaging/infrastructure/websocket"
	messagingHandler "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/messaging/interfaces/http"
	messagingMessages "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/messaging/interfaces/http/messages"
	notifDTO "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/application/dto"
	notifServices "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/application/services"
	notifUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/application/usecases"
	notifEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/entities"
	emailDomain "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/services"
	notifPersistence "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/infrastructure/persistence"
	notifScheduler "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/infrastructure/scheduler"
	notifHttp "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/interfaces/http"
	emailHandlers "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/interfaces/http/handlers"
	reportUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/application/usecases"
	reportPersistence "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/infrastructure/persistence"
	reportQuery "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/infrastructure/query"
	reportHandler "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reporting/interfaces/http/handlers"
	annualUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reports/annual/application/usecases"
	annualDocxgen "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reports/annual/infrastructure/docxgen"
	annualHandler "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/reports/annual/interfaces/http/handlers"
	scheduleUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/application/usecases"
	schedulePersistence "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/infrastructure/persistence"
	scheduleHandler "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/interfaces/http/handlers"
	sdUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/application/usecases"
	sdExcel "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/infrastructure/excel"
	sdPersistence "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/infrastructure/persistence"
	sdHandler "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/student_debts/interfaces/http/handlers"
	taskDto "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/application/dto"
	taskUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/application/usecases"
	taskPersistence "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/infrastructure/persistence"
	taskHandler "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/interfaces/http/handlers"
	taskRoutes "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/interfaces/http/routes"
	usersUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/application/usecases"
	usersRepositories "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/domain/repositories"
	usersPersistence "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/infrastructure/persistence"
	usersHandler "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/interfaces/http/handlers"
	usersRoutes "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/users/interfaces/http/routes"
	wpUsecases "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/application/usecases"
	wpLLM "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/infrastructure/llm"
	wpPersistence "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/infrastructure/persistence"
	wpRateLimit "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/infrastructure/ratelimit"
	wpHandler "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/interfaces/http/handlers"
	adminAuditLog "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/admin/auditlog"
	adminBackups "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/admin/backups"
	adminComposio "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/admin/composio"
	adminIntegrations "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/admin/integrations"
	adminSentry "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/admin/sentry"
	appMiddleware "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/application/middleware"
	sharedDomainErrors "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/domain/errors"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/cache"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/config"
	authCrypto "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/crypto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/health"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/metrics"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/middleware"
	n8ninfra "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/n8n"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/storage"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/telegram"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/tracing"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/validation"
)

// versionString is the single runtime source for the --version banner.
// It is updated atomically by _tools/bump_version.sh alongside VERSION
// and the rest of the version-carrying files.
const versionString = "0.222.0"

// errorKey is the field name used in gin.H and logger context maps for
// error payloads. Extracted to satisfy goconst.
const errorKey = "error"

// handleVersionFlag prints the version banner when --version is passed and reports whether main should exit.
func handleVersionFlag() bool {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println("inf-sys-secretary-methodist v" + versionString)
		return true
	}
	return false
}

// handleHealthCheckFlag handles the `-healthcheck` subcommand invoked by the
// container/compose healthcheck. The runtime image is built FROM scratch and
// has no shell/wget/curl, so the binary probes its own /health endpoint and
// terminates the process with the probe's exit code (0 healthy, 1 unhealthy).
// It calls os.Exit directly because a healthcheck invocation never boots the
// server; it returns only when the flag is absent.
func handleHealthCheckFlag() {
	if len(os.Args) > 1 && os.Args[1] == "-healthcheck" {
		port := os.Getenv("SERVER_PORT")
		if port == "" {
			port = "8080"
		}
		os.Exit(health.Probe("http://localhost:"+port+"/health", 3*time.Second))
	}
}

// initSentry wires Sentry error tracking when SENTRY_DSN is set. Failures are logged, never fatal.
func initSentry(cfg *config.Config) {
	dsn := os.Getenv("SENTRY_DSN")
	if dsn == "" {
		return
	}
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		Environment:      cfg.Environment,
		Release:          cfg.Version,
		TracesSampleRate: 0.1,
		EnableTracing:    true,
	}); err != nil {
		log.Printf("Sentry initialization failed: %v", err)
		return
	}
	log.Println("Sentry initialized successfully")
}

// messagingAccessCheckerFunc adapts a closure to the messaging
// ConversationAccessChecker port. v0.162.0 ADR-1 (#297) wiring.
type messagingAccessCheckerFunc func(ctx context.Context, userID, conversationID int64) (bool, error)

func (f messagingAccessCheckerFunc) IsParticipant(ctx context.Context, userID, conversationID int64) (bool, error) {
	return f(ctx, userID, conversationID)
}

// messagingUserExistenceFunc adapts a closure to the messaging
// UserExistenceChecker port. v0.162.0 ADR-3 (#297) wiring.
type messagingUserExistenceFunc func(ctx context.Context, userID int64) (bool, error)

func (f messagingUserExistenceFunc) UserExists(ctx context.Context, userID int64) (bool, error) {
	return f(ctx, userID)
}

// newMessagingAccessChecker builds the ADR-1 conversation participant
// gate. Extracted from main() to keep gocyclo below the project gate.
func newMessagingAccessChecker(conversationRepo messagingUsecases.ConversationRepository) messagingAccessCheckerFunc {
	return func(ctx context.Context, userID, conversationID int64) (bool, error) {
		conv, err := conversationRepo.GetByID(ctx, conversationID)
		if err != nil {
			return false, err
		}
		return conv.HasParticipant(userID), nil
	}
}

// newMessagingUserExistenceChecker builds the ADR-3 recipient existence
// gate. Returns (false, nil) for missing users (oracle closure), and
// (false, err) for transport-level repo failures.
func newMessagingUserExistenceChecker(userRepo usecases.UserRepository) messagingUserExistenceFunc {
	return func(ctx context.Context, userID int64) (bool, error) {
		_, err := userRepo.GetByID(ctx, userID)
		if err != nil {
			if errors.Is(err, sharedDomainErrors.ErrNotFound) {
				return false, nil
			}
			return false, err
		}
		return true, nil
	}
}

// announcementUserIDsProvider adapts the users module's
// UserProfileRepository к the announcements module's UserIDsProvider
// port. v0.163.1 polish: wires the previously-nil broadcast fan-out
// with audience-scoped recipient lookup so a student-targeted
// announcement only pushes к students (no cross-audience push leak,
// closes v0.163.0 Tier 1 audit #7).
//
// Cross-module DI adapter pattern from v0.155.1 — announcements/
// application stays free of direct users/application imports.
type announcementUserIDsProvider struct {
	repo usersUsecases.UserProfileRepository
}

// announcementFanOutPageLimit caps the per-role page so a single
// announcement broadcast doesn't fan-out к more than this many users
// per matching role. Polish-grade default; scale-out via a dedicated
// "active IDs by role" repo method when fan-out exceeds this floor.
const announcementFanOutPageLimit = 500

// GetUserIDsForAudience returns the active user IDs whose role matches
// the announcement's target_audience. The audience-to-roles map mirrors
// the announcements/domain.VisibleAudiences matrix in reverse: which
// roles can see audience X.
func (p *announcementUserIDsProvider) GetUserIDsForAudience(ctx context.Context, audience announcementsDomain.TargetAudience) ([]int64, error) {
	roles := rolesForAudience(audience)
	if roles == nil {
		return nil, nil
	}
	seen := make(map[int64]struct{})
	out := make([]int64, 0, announcementFanOutPageLimit)
	for _, role := range roles {
		filter := &usersRepositories.UserFilter{
			Role:   role,
			Status: "active",
		}
		users, err := p.repo.ListUsersWithOrg(ctx, filter, announcementFanOutPageLimit, 0)
		if err != nil {
			return nil, err
		}
		for _, u := range users {
			if _, dup := seen[u.ID]; dup {
				continue
			}
			seen[u.ID] = struct{}{}
			out = append(out, u.ID)
		}
	}
	return out, nil
}

// messagingNotificationNotifier adapts the notifications module's
// NotificationUseCase к the messaging module's MessageNotifier port.
// v0.162.1 polish Item 3: the adapter previously lived в
// internal/modules/messaging/application/services/notifier.go and
// imported internal/modules/notifications/... directly — a
// cross-module-impl violation. Moving к main.go (the composition
// root) keeps the messaging module free of the notifications import
// while preserving the original adapter behavior. Mirror к the
// announcementUserIDsProvider DI seam above.
type messagingNotificationNotifier struct {
	notificationUseCase *notifUsecases.NotificationUseCase
}

// NotifyNewMessage sends a notification about a new message to the
// specified user. Nil notificationUseCase is treated as a successful
// no-op so the messaging module stays usable when notifications wiring
// is absent (e.g. dev/test profiles, mirror к the original adapter's
// short-circuit branch).
func (n *messagingNotificationNotifier) NotifyNewMessage(ctx context.Context, userID int64, senderName, content string, conversationID, messageID int64) error {
	if n.notificationUseCase == nil {
		return nil
	}
	input := &notifDTO.CreateNotificationInput{
		UserID:   userID,
		Type:     notifEntities.NotificationTypeInfo,
		Priority: notifEntities.PriorityNormal,
		Title:    fmt.Sprintf("Новое сообщение от %s", senderName),
		Message:  content,
		Link:     fmt.Sprintf("/messages/%d", conversationID),
		Metadata: map[string]any{
			"conversation_id": conversationID,
			"message_id":      messageID,
			"sender_name":     senderName,
		},
	}
	_, err := n.notificationUseCase.Create(ctx, input)
	return err
}

// signerNameResolverAdapter resolves a signer's display name from the auth
// users repository for the document e-signature handler (#140). Declared at the
// composition root so the documents module stays free of the auth import.
type signerNameResolverAdapter struct {
	userRepo usecases.UserRepository
}

// FullName returns the user's display name, or an error when the lookup fails.
func (a signerNameResolverAdapter) FullName(ctx context.Context, userID int64) (string, error) {
	u, err := a.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", err
	}
	return u.Name, nil
}

// studentDebtResitNotifier adapts the notifications module's
// NotificationUseCase to the student_debts DebtNotifier port (the
// composition root keeps student_debts free of the notifications import).
// The use case calls this synchronously on the request path, so the
// adapter detaches the request context (context.WithoutCancel) and hands
// off to a background goroutine — a fire-and-forget notification must
// neither block the HTTP response nor be canceled when the request ends.
type studentDebtResitNotifier struct {
	notificationUseCase *notifUsecases.NotificationUseCase
}

// NotifyResitScheduled notifies the student that a resit was scheduled. A
// nil use case is a silent no-op (dev/test profiles).
func (n *studentDebtResitNotifier) NotifyResitScheduled(ctx context.Context, studentUserID, debtID int64, disciplineName string, scheduledDate time.Time) {
	if n.notificationUseCase == nil {
		return
	}
	bgCtx := context.WithoutCancel(ctx)
	input := &notifDTO.CreateNotificationInput{
		UserID:   studentUserID,
		Type:     notifEntities.NotificationTypeInfo,
		Priority: notifEntities.PriorityNormal,
		Title:    "Назначена пересдача",
		Message:  fmt.Sprintf("По дисциплине «%s» назначена пересдача на %s", disciplineName, scheduledDate.Format("02.01.2006 15:04")),
		Link:     fmt.Sprintf("/student-debts/%d", debtID),
		Metadata: map[string]any{
			"debt_id":    debtID,
			"discipline": disciplineName,
		},
	}
	go func() {
		_, _ = n.notificationUseCase.Create(bgCtx, input)
	}()
}

// rolesForAudience inverts the announcements visibility matrix: which
// roles are recipients of an announcement addressed к the given
// audience. Empty role ("") means "any" — used for the broadcast
// audience.
func rolesForAudience(audience announcementsDomain.TargetAudience) []string {
	switch audience {
	case announcementsDomain.TargetAudienceAll:
		// Empty role filter passes every active user through.
		return []string{""}
	case announcementsDomain.TargetAudienceStudents:
		return []string{"student"}
	case announcementsDomain.TargetAudienceTeachers:
		return []string{"teacher"}
	case announcementsDomain.TargetAudienceStaff:
		return []string{"methodist", "academic_secretary"}
	case announcementsDomain.TargetAudienceAdmins:
		return []string{"system_admin"}
	default:
		return nil
	}
}

func main() {
	if handleVersionFlag() {
		return
	}
	handleHealthCheckFlag() // exits the process when -healthcheck is passed

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	initSentry(cfg)

	// serverCtx is the application lifecycle context. Background goroutines
	// and schedulers derive their tick contexts from it so SIGTERM cancels
	// in-flight work instead of letting it run to completion on
	// context.Background(). Canceled explicitly in the graceful-shutdown
	// branch before any scheduler.Stop() calls (see issue #263 ADR-4); no
	// defer here because log.Fatalf paths above can exit before any work
	// using the ctx kicks off.
	serverCtx, cancelServerCtx := context.WithCancel(context.Background())

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
				errorKey: err.Error(),
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
				errorKey: err.Error(),
			})
		}
	}()

	logger.Info("Database connected successfully", map[string]interface{}{
		"max_open_conns": cfg.Database.MaxOpenConns,
		"max_idle_conns": cfg.Database.MaxIdleConns,
	})

	// Wire audit_logs persistence (v0.130.0). AuditLogger continues to
	// emit structured stdout events for every LogAuditEvent call; with
	// the repository attached it ALSO inserts a row into audit_logs on
	// its own *sql.DB handle — independent of any business transaction
	// per plan ADR-2, so failed/denied business ops still get audited.
	// Write failure is logged at error level and not propagated to the
	// caller per ADR-3.
	auditLogger = auditLogger.WithRepository(logging.NewAuditLogRepositoryPG(db))

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
			errorKey: err.Error(),
		})
	}
	if redisCache != nil {
		defer func() {
			if err := redisCache.Close(); err != nil {
				logger.Error("Failed to close Redis connection", map[string]interface{}{
					errorKey: err.Error(),
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
				errorKey: err.Error(),
			})
		} else {
			// Ensure bucket exists
			if err := s3Client.EnsureBucket(context.Background()); err != nil {
				logger.Warn("Failed to ensure S3 bucket exists", map[string]interface{}{
					errorKey: err.Error(),
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
	// Workflow use cases (v0.148.0 #227). Declared at outer scope so
	// route registration далее в main() can reference them once the
	// optional documents block has wired them up.
	var submitDocUseCase *docUsecases.SubmitDocumentUseCase
	var approveDocUseCase *docUsecases.ApproveDocumentUseCase
	var rejectDocUseCase *docUsecases.RejectDocumentUseCase
	var registerDocUseCase *docUsecases.RegisterDocumentUseCase
	var startRoutingDocUseCase *docUsecases.StartRoutingUseCase
	var signVisaDocUseCase *docUsecases.SignVisaUseCase
	var assignExecutorDocUseCase *docUsecases.AssignExecutorUseCase
	var markExecutedDocUseCase *docUsecases.MarkExecutedUseCase
	var archiveDocUseCase *docUsecases.ArchiveDocumentUseCase
	var resubmitDocUseCase *docUsecases.ResubmitDocumentUseCase
	if s3Client != nil {
		docRepo := docPersistence.NewDocumentRepositoryPG(db)
		docTypeRepo := docPersistence.NewDocumentTypeRepositoryPG(db)
		docCategoryRepo := docPersistence.NewDocumentCategoryRepositoryPG(db)
		permissionRepo := docPersistence.NewPermissionRepositoryPG(db)
		publicLinkRepo := docPersistence.NewPublicLinkRepositoryPG(db)
		docTagRepo := docPersistence.NewDocumentTagRepositoryPG(db)
		docUseCase = docUsecases.NewDocumentUseCase(docRepo, docTypeRepo, docCategoryRepo, s3Client, auditLogger)
		sharingUseCase = docUsecases.NewSharingUseCase(docRepo, permissionRepo, publicLinkRepo, auditLogger, cfg.Server.BaseURL).
			WithShareNotifier(&documentsShareNotifier{notif: notificationUseCase})
		docVersionUseCase = docUsecases.NewDocumentVersionUseCase(docRepo, s3Client, auditLogger)
		tagUseCase = docUsecases.NewTagUseCase(docTagRepo, docRepo, auditLogger)
		templateRepo := docPersistence.NewTemplateRepositoryAdapter(docTypeRepo)
		templateUseCase = docUsecases.NewTemplateUseCase(templateRepo, docRepo, auditLogger)
		// v0.148.0 — workflow use cases (issue #227). The PG repo
		// returns a generic fmt.Errorf("document not found") which the
		// usecases compare via errors.Is(ErrDocumentNotFound) — wrap
		// it in workflowDocRepoAdapter so the sentinel matches без
		// touching existing PG repo consumers.
		workflowRepoAdapter := &workflowDocRepoAdapter{inner: docRepo}
		submitDocUseCase = docUsecases.NewSubmitDocumentUseCase(workflowRepoAdapter, auditLogger, nil)
		approveDocUseCase = docUsecases.NewApproveDocumentUseCase(workflowRepoAdapter, auditLogger, nil)
		rejectDocUseCase = docUsecases.NewRejectDocumentUseCase(workflowRepoAdapter, auditLogger, nil)
		registerDocUseCase = docUsecases.NewRegisterDocumentUseCase(workflowRepoAdapter, auditLogger, nil)
		// v0.150.0 Phase 3 — routing use cases (#231). registered →
		// routing → execution через single-step visa.
		startRoutingDocUseCase = docUsecases.NewStartRoutingUseCase(workflowRepoAdapter, auditLogger, nil)
		signVisaDocUseCase = docUsecases.NewSignVisaUseCase(workflowRepoAdapter, auditLogger, nil)
		// v0.151.0 Phase 4 — execution use cases (#232). execution →
		// executed via AssignExecutor (shape-only) + MarkExecuted.
		assignExecutorDocUseCase = docUsecases.NewAssignExecutorUseCase(workflowRepoAdapter, auditLogger, nil)
		markExecutedDocUseCase = docUsecases.NewMarkExecutedUseCase(workflowRepoAdapter, auditLogger, nil)
		// v0.152.0 Phase 5 — archive + resubmit use cases (#233, final
		// phase). Closes 5-phase pack #227. Archive admin-only;
		// Resubmit author OR edit-role per ADR-2.
		archiveDocUseCase = docUsecases.NewArchiveDocumentUseCase(workflowRepoAdapter, auditLogger, nil)
		resubmitDocUseCase = docUsecases.NewResubmitDocumentUseCase(workflowRepoAdapter, auditLogger, nil)
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

	// Task reminders (v0.138.0 — Phase 5 #5 final SetReminder).
	// Repo + 3 use cases + handler composition extracted к
	// initTaskReminderModule helper so main() cyclomatic complexity
	// stays под the golangci threshold. The cross-module
	// TaskReminderScheduler (wired below) reuses the existing
	// notifications module telegramRepo + ComposioTelegramService
	// for delivery.
	taskReminderRepo, taskReminderHandlerInstance := initTaskReminderModule(db, auditLogger)

	// Initialize assignments module — academic Tasks Context (separate
	// bounded context from project-management tasks). v0.109.0 ships the
	// SaveGrade flow only; submission upload + rubric land in later
	// releases.
	assignmentRepo := assignPersistence.NewAssignmentRepositoryPG(db)
	submissionRepo := assignPersistence.NewSubmissionRepositoryPG(db)

	// Notifier adapter: a struct local to main that translates the
	// assignments-module narrow port into a NotificationUseCase.Create
	// call. Keeping the adapter at the DI seam (here, not inside the
	// assignments module) avoids a cross-module Go import in
	// domain/usecase code.
	gradeNotifier := &assignmentsGradeNotifier{
		notif: notificationUseCase,
	}
	saveGradeUseCase := assignUsecases.NewSaveGradeUseCase(
		assignmentRepo, submissionRepo, gradeNotifier, auditLogger, nil,
	)
	returnNotifier := &assignmentsReturnNotifier{
		notif: notificationUseCase,
	}
	returnSubmissionUseCase := assignUsecases.NewReturnSubmissionUseCase(
		assignmentRepo, submissionRepo, returnNotifier, auditLogger, nil,
	)
	resubmitNotifier := &assignmentsResubmitNotifier{
		notif: notificationUseCase,
	}
	resubmitSubmissionUseCase := assignUsecases.NewResubmitSubmissionUseCase(
		assignmentRepo, submissionRepo, resubmitNotifier, auditLogger, nil,
	)
	listAssignmentsUseCase := assignUsecases.NewListAssignmentsUseCase(assignmentRepo)
	getAssignmentUseCase := assignUsecases.NewGetAssignmentUseCase(assignmentRepo)
	listSubmissionsUseCase := assignUsecases.NewListSubmissionsUseCase(assignmentRepo, submissionRepo)
	// Student-facing read use cases (v0.113.0). The list use case takes
	// the narrow MyAssignmentsRepository port — submissionRepo satisfies
	// it because ListByStudent is on the SubmissionRepository surface.
	listMyAssignmentsUseCase := assignUsecases.NewListMyAssignmentsUseCase(submissionRepo)
	getMyAssignmentDetailUseCase := assignUsecases.NewGetMyAssignmentDetailUseCase(assignmentRepo, submissionRepo)
	logger.Info("Assignments module initialized", nil)

	// Initialize curriculum module — academic curriculum (учебный план)
	// bounded context. v0.116.0 ships basic CRUD (create / read / list /
	// update); the approval workflow + Discipline child entity land in
	// v0.117.0. *logging.AuditLogger satisfies the curriculum-side
	// AuditSink interface structurally (same shape as the assignments
	// module's port — single concrete logger covers both modules
	// without a cross-module Go import).
	curriculumRepo := curPersistence.NewCurriculumRepositoryPG(db)
	createCurriculumUseCase := curUsecases.NewCreateCurriculumUseCase(curriculumRepo, auditLogger, nil)
	getCurriculumUseCase := curUsecases.NewGetCurriculumUseCase(curriculumRepo)
	listCurriculaUseCase := curUsecases.NewListCurriculaUseCase(curriculumRepo)
	updateCurriculumUseCase := curUsecases.NewUpdateCurriculumUseCase(curriculumRepo, auditLogger, nil)
	// v0.117.0 lifecycle transitions: Submit (methodist or admin) +
	// Approve / Reject (admin-only). Audit-symmetric with the four
	// CRUD use cases via the shared emitAudit helper.
	submitCurriculumUseCase := curUsecases.NewSubmitForApprovalUseCase(curriculumRepo, auditLogger, nil)
	approveCurriculumUseCase := curUsecases.NewApproveCurriculumUseCase(curriculumRepo, auditLogger, nil)
	rejectCurriculumUseCase := curUsecases.NewRejectCurriculumUseCase(curriculumRepo, auditLogger, nil)
	// v0.128.0 Section aggregate (раздел учебного плана) — child of
	// Curriculum per ADR-1 Beta (independent AR with FK). 5 CRUD usecases
	// all share the curriculumRepo as cross-aggregate lookup port. The
	// concrete *CurriculumRepositoryPG satisfies each narrow port
	// structurally (each usecase declares its own GetByID-shaped
	// interface for testability).
	sectionRepo := curPersistence.NewSectionRepositoryPG(db)
	createSectionUseCase := curUsecases.NewCreateSectionUseCase(sectionRepo, curriculumRepo, auditLogger, nil)
	getSectionUseCase := curUsecases.NewGetSectionUseCase(sectionRepo)
	listSectionsUseCase := curUsecases.NewListSectionsByCurriculumUseCase(sectionRepo)
	updateSectionUseCase := curUsecases.NewUpdateSectionUseCase(sectionRepo, curriculumRepo, auditLogger, nil)
	deleteSectionUseCase := curUsecases.NewDeleteSectionUseCase(sectionRepo, curriculumRepo, auditLogger)
	// v0.128.1 DisciplineItem aggregate (Layer 2 of B1a hierarchy) —
	// child of Section per ADR-1 Beta. 5 CRUD usecases share two-level
	// cross-aggregate lookup (sectionRepo + curriculumRepo) для
	// AuthorizeDisciplineItemEdit primitives.
	disciplineItemRepo := curPersistence.NewDisciplineItemRepositoryPG(db)
	createDisciplineItemUseCase := curUsecases.NewCreateDisciplineItemUseCase(disciplineItemRepo, sectionRepo, curriculumRepo, auditLogger, nil)
	getDisciplineItemUseCase := curUsecases.NewGetDisciplineItemUseCase(disciplineItemRepo)
	listDisciplineItemsUseCase := curUsecases.NewListDisciplineItemsBySectionUseCase(disciplineItemRepo, sectionRepo)
	updateDisciplineItemUseCase := curUsecases.NewUpdateDisciplineItemUseCase(disciplineItemRepo, sectionRepo, curriculumRepo, auditLogger, nil)
	deleteDisciplineItemUseCase := curUsecases.NewDeleteDisciplineItemUseCase(disciplineItemRepo, sectionRepo, curriculumRepo, auditLogger)
	// v0.128.3 Bulk-edit transactional endpoint (B1a Layer 3) per ADR-10:
	// dedicated UoW wraps *sql.DB и produces tx-bound repos via DBTX
	// reuse, supporting commit-or-rollback semantic для combined
	// creates+updates+deletes operations.
	bulkDisciplineItemsUoW := curPersistence.NewBulkDisciplineItemsUnitOfWorkPG(db)
	bulkEditDisciplineItemsUseCase := curUsecases.NewBulkEditDisciplineItemsUseCase(bulkDisciplineItemsUoW, auditLogger, nil)
	logger.Info("Curriculum module initialized", nil)

	// Annual methodist report (B4, v0.129.0+) — cross-module read-only
	// orchestrator. Calendar-year aggregation per ADR-4. The documents
	// dependency is the narrow DocumentActivityReaderPG (v0.129.1),
	// independent from the full DocumentRepositoryPG construction so
	// the annual orchestrator does not couple to that repo's invariants.
	annualReportUseCase := annualUsecases.NewAnnualReportUseCase(
		curriculumRepo,
		assignmentRepo,
		disciplineItemRepo,
		docPersistence.NewDocumentActivityReaderPG(db),
		annualDocxgen.NewRenderer(),
		auditLogger,
	)
	logger.Info("Annual report module initialized", nil)

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
	lessonSlotRepo := schedulePersistence.NewLessonSlotRepositoryPG(db)
	lessonSlotUseCase := scheduleUsecases.NewLessonSlotUseCase(lessonSlotRepo)
	teachingLoadRepo := schedulePersistence.NewTeachingLoadRepositoryPG(db)
	teachingLoadUseCase := scheduleUsecases.NewTeachingLoadUseCase(teachingLoadRepo)
	generateScheduleUseCase := scheduleUsecases.NewGenerateScheduleUseCase(
		teachingLoadRepo, lessonSlotRepo, classroomRepo,
		scheduleUsecases.WithApplyWriter(lessonRepo),
		scheduleUsecases.WithSemesters(referenceRepo),
	)
	logger.Info("Schedule lessons module initialized", nil)

	// Initialize calendar feed (external calendar subscription, issue #40)
	feedTokenRepo := schedulePersistence.NewCalendarFeedTokenRepositoryPG(db)
	studentGroupResolver := schedulePersistence.NewStudentGroupResolverPG(db)
	calendarFeedUseCase := scheduleUsecases.NewCalendarFeedUseCase(
		feedTokenRepo, lessonRepo, eventRepo, studentGroupResolver,
		scheduleUsecases.CalendarFeedConfig{
			ProdID:       "-//Secretary Methodist//Calendar Feed//EN",
			CalendarName: "Расписание",
			TZID:         "Europe/Moscow",
			UIDDomain:    "secretary-methodist",
			PastWindow:   30 * 24 * time.Hour,
			FutureWindow: 365 * 24 * time.Hour,
		},
	)
	logger.Info("Calendar feed module initialized", nil)

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
			errorKey: err.Error(),
		})
	} else {
		// v0.138.1 — wire telegram dispatch onto the existing event
		// reminder scheduler so event_reminders also light up the
		// Composio path (was a carry-forward gap closed alongside the
		// Phase 5 #5 final task_reminders telegram dispatch).
		wireEventReminderDispatch(reminderScheduler, telegramRepo, telegramService, webpushRepo, webpushService)
		if err := reminderScheduler.Start(); err != nil {
			logger.Error("Failed to start reminder scheduler", map[string]interface{}{
				errorKey: err.Error(),
			})
		} else {
			logger.Info("Reminder scheduler started", nil)
		}
	}

	// Task reminder scheduler (v0.138.0 — Phase 5 #5 final). Mirror
	// к the existing reminderScheduler shape но drives
	// task_reminders. Telegram dispatch uses the existing
	// ComposioTelegramService — Phase 5 #5 final closes by lighting
	// up the long-dormant fallback path. Extracted к initTaskReminder
	// Scheduler so main() cyclomatic complexity stays under the
	// golangci threshold.
	taskReminderScheduler := initTaskReminderScheduler(
		logger,
		taskReminderRepo,
		taskRepo,
		db,
		telegramRepo,
		telegramService,
		notificationRepo,
		preferencesRepo,
		notifEmailService,
		webpushRepo,
		webpushService,
	)

	// Initialize announcements module — repo only. The usecase wires
	// after the users module is up so the broadcast fan-out can use
	// the user profile repo via a cross-module DI adapter.
	announcementRepo := announcementPersistence.NewAnnouncementRepositoryPG(db)
	var announcementUseCase *announcementUsecases.AnnouncementUseCase
	_ = announcementRepo
	logger.Info("Announcements module initialized (usecase deferred)", nil)

	// Initialize dashboard module
	dashboardRepo := dashboardPersistence.NewDashboardRepositoryPG(db)
	dashboardUseCase := dashboardUsecases.NewDashboardUseCase(dashboardRepo)
	logger.Info("Dashboard module initialized", nil)

	// Initialize analytics module (predictive analytics for student risk assessment)
	analyticsRepo := analyticsPersistence.NewAnalyticsRepositoryPG(db)
	attendanceRepo := analyticsPersistence.NewAttendanceRepositoryPG(db)
	gradeRepo := analyticsPersistence.NewGradeRepositoryPG(db)
	// teacherScopeRepo resolves the teacher → group whitelist used by the
	// /api/analytics/* scope filter (v0.108.3). Reads schedule_lessons +
	// student_groups via a cross-table SQL JOIN.
	teacherScopeRepo := analyticsPersistence.NewTeacherScopeRepositoryPG(db)
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
		logger.Warn("Failed to initialize risk recalculation scheduler", map[string]any{errorKey: err.Error()})
	} else {
		riskScheduler.Start()
		defer func() { _ = riskScheduler.Stop() }()
	}

	logger.Info("Analytics module initialized", nil)

	// Initialize users module
	departmentRepo := usersPersistence.NewDepartmentRepositoryPG(db)
	positionRepo := usersPersistence.NewPositionRepositoryPG(db)
	userProfileRepo := usersPersistence.NewUserProfileRepositoryPG(db)
	// userRepo is the auth-side narrow port; the users module needs
	// the wider UserAccountRepository (includes CountByRole for the
	// #283 ADR-4 last-admin guard). Both concrete implementations
	// (*UserRepositoryPG, *CachedUserRepository) satisfy the wider
	// port — type-assert at the seam rather than coupling auth's
	// port to a users-only concern.
	usersUserAccountRepo, ok := userRepo.(usersUsecases.UserAccountRepository)
	if !ok {
		// Panic instead of log.Fatalf — log.Fatalf calls os.Exit which
		// skips deferred db.Close() etc. The wider port is a static DI
		// contract, never a runtime branch in production, so a panic
		// here is a boot-time invariant violation, not a user-facing
		// error. gocritic: exitAfterDefer.
		panic("auth user repository does not satisfy users.UserAccountRepository — CountByRole required for #283 ADR-4 last-admin guard")
	}
	// v0.160.1 polish Item 3: NotificationUseCase satisfies the
	// users.SystemNotifier narrow port structurally (single method
	// SendSystemNotification). The compile-time assertion below pins
	// the contract — drift в either signature fails build, not request.
	var _ usersUsecases.SystemNotifier = (*notifUsecases.NotificationUseCase)(nil)
	userUseCase := usersUsecases.NewUserUseCase(usersUserAccountRepo, userProfileRepo, departmentRepo, positionRepo, auditLogger, notificationUseCase).
		WithLifecycleContext(serverCtx)
	departmentUseCase := usersUsecases.NewDepartmentUseCase(departmentRepo, auditLogger)
	positionUseCase := usersUsecases.NewPositionUseCase(positionRepo, auditLogger)
	logger.Info("Users module initialized", nil)

	// v0.163.1 polish: wire announcement broadcast fan-out provider via
	// cross-module DI adapter (no direct users/application import from
	// announcements/application — mirror к v0.155.1 ai pattern).
	// announcementUserIDsAdapter maps announcement target_audience к
	// role(s) and pages users via UserProfileRepository. The deferred
	// usecase init avoids upward dependency on the users module from
	// the announcements repo block above.
	announcementUserIDsAdapter := &announcementUserIDsProvider{repo: userProfileRepo}
	announcementUseCase = announcementUsecases.
		NewAnnouncementUseCase(announcementRepo, auditLogger, notificationUseCase, announcementUserIDsAdapter).
		WithLifecycleContext(serverCtx)
	if s3Client != nil {
		announcementUseCase.SetAttachmentStorage(s3Client)
		logger.Info("Announcement attachment storage wired (S3)", nil)
	}
	logger.Info("Announcements broadcast fan-out wired (audience-scoped + lifecycle ctx)", nil)

	// Initialize messaging module.
	// v0.162.0 ADR-1/ADR-3 (#297): participant gate for WS subscribe/typing
	// and recipient existence gate for CreateDirect/Group are wired here.
	conversationRepo := messagingPersistence.NewConversationRepositoryPG(db)
	messageRepo := messagingPersistence.NewMessageRepositoryPG(db)
	messagingHub := messagingWebsocket.NewHub(logger).
		WithAccessChecker(newMessagingAccessChecker(conversationRepo))
	go messagingHub.Run() // Start WebSocket hub in background
	messageNotifier := &messagingNotificationNotifier{notificationUseCase: notificationUseCase}
	messagingUseCase := messagingUsecases.NewMessagingUseCase(conversationRepo, messageRepo, messagingHub, logger, messageNotifier, s3Client).
		WithAuditSink(auditLogger).
		WithUserExistenceChecker(newMessagingUserExistenceChecker(userRepo)).
		WithLifecycleContext(serverCtx).
		WithSystemMessageTexts(messagingUsecases.SystemMessageTexts{
			GroupCreated: messagingMessages.SystemGroupCreated,
			UserJoined:   messagingMessages.SystemUserJoined,
			UserLeft:     messagingMessages.SystemUserLeft,
		})
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
		taskReminderHandlerInstance,
		saveGradeUseCase,
		returnSubmissionUseCase,
		resubmitSubmissionUseCase,
		listAssignmentsUseCase,
		getAssignmentUseCase,
		listSubmissionsUseCase,
		listMyAssignmentsUseCase,
		getMyAssignmentDetailUseCase,
		createCurriculumUseCase,
		getCurriculumUseCase,
		listCurriculaUseCase,
		updateCurriculumUseCase,
		submitCurriculumUseCase,
		approveCurriculumUseCase,
		rejectCurriculumUseCase,
		createSectionUseCase,
		getSectionUseCase,
		listSectionsUseCase,
		updateSectionUseCase,
		deleteSectionUseCase,
		createDisciplineItemUseCase,
		getDisciplineItemUseCase,
		listDisciplineItemsUseCase,
		updateDisciplineItemUseCase,
		deleteDisciplineItemUseCase,
		bulkEditDisciplineItemsUseCase,
		annualReportUseCase,
		eventUseCase,
		lessonUseCase,
		lessonSlotUseCase,
		teachingLoadUseCase,
		generateScheduleUseCase,
		calendarFeedUseCase,
		announcementUseCase,
		dashboardUseCase,
		analyticsUseCase,
		teacherScopeRepo,
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
		submitDocUseCase,
		approveDocUseCase,
		rejectDocUseCase,
		registerDocUseCase,
		startRoutingDocUseCase,
		signVisaDocUseCase,
		assignExecutorDocUseCase,
		markExecutedDocUseCase,
		archiveDocUseCase,
		resubmitDocUseCase,
	)

	// Initialize integration module (1C synchronization)
	var integrationModule *integration.Module
	integrationModule, err = integration.NewModule(db, &cfg.Integration, logger)
	if err != nil {
		logger.Error("Failed to initialize integration module", map[string]interface{}{
			errorKey: err.Error(),
		})
	} else if integrationModule.IsEnabled() {
		integrationModule.WithAuditSink(auditLogger)
		// Register routes under protected API group with admin guard.
		// Only system_admin role may invoke 1C sync, browse external entities
		// or resolve conflicts — see AUDIT_REPORT critical item #3.
		apiGroup := router.Group("/api")
		apiGroup.Use(authMiddleware.JWTMiddleware(authUseCase))
		integrationModule.RegisterRoutes(apiGroup, authMiddleware.RequireRole(string(authDomain.RoleSystemAdmin)))

		// Start scheduler for periodic sync
		if err := integrationModule.StartScheduler(context.Background()); err != nil {
			logger.Error("Failed to start integration scheduler", map[string]interface{}{
				errorKey: err.Error(),
			})
		}

		// 1С debt import (PR7, #431): bridge the integration OData client into
		// the student_debts registry. Registered here — not in setupRoutes —
		// because the integration module is created after the router; reuses
		// apiGroup (JWT) since debt-manager gating lives in the use case, not
		// an admin guard.
		sdDebtRepo := sdPersistence.NewStudentDebtRepositoryPG(db)
		import1CUC := sdUsecases.NewImport1CDebtsUseCase(
			sdDebtRepo, debt1CSource{catalog: integrationModule.ODataClient()}, auditLogger)
		debt1CHandler := sdHandler.NewStudentDebt1CImportHandler(import1CUC)
		sdHandler.RegisterStudentDebt1CImportRoutes(apiGroup, debt1CHandler)

		logger.Info("Integration module initialized", nil)
	}

	// Initialize AI module (RAG/Chat functionality)
	// AI schedulers hoisted к outer scope so the graceful-shutdown branch
	// can call Stop() on them (issue #263 ADR-4 — closes goroutine leak via
	// gocron.Shutdown on SIGTERM). Both stay nil when cfg.AI.Enabled is false.
	var aiFactScheduler *aiScheduler.FactScheduler
	var aiIndexingScheduler *aiScheduler.IndexingScheduler

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
		// LLM reranking stage for modified RAG (search_mode="hybrid_rerank")
		aiEmbeddingUseCase.WithReranker(llmProvider)

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
			logger.Warn("Failed to seed fun facts", map[string]interface{}{errorKey: err.Error()})
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
				serverCtx,
				funFactUseCase,
				moodUseCase,
				telegramPersonalityService,
				telegramRepo,
				slog.Default(),
			)
			if err != nil {
				logger.Warn("Failed to create fact scheduler", map[string]interface{}{errorKey: err.Error()})
			} else {
				if err := factScheduler.Start(); err != nil {
					logger.Warn("Failed to start fact scheduler", map[string]interface{}{errorKey: err.Error()})
				} else {
					logger.Info("Fact scheduler started", nil)
					aiFactScheduler = factScheduler
				}
			}
		}

		// Initialize indexing scheduler for automatic document indexing
		indexingScheduler, err := aiScheduler.NewIndexingScheduler(
			serverCtx,
			aiEmbeddingUseCase,
			10,
			slog.Default(),
		)
		if err != nil {
			logger.Warn("Failed to create indexing scheduler", map[string]interface{}{errorKey: err.Error()})
		} else {
			if err := indexingScheduler.Start(); err != nil {
				logger.Warn("Failed to start indexing scheduler", map[string]interface{}{errorKey: err.Error()})
			} else {
				logger.Info("Indexing scheduler started", nil)
				aiIndexingScheduler = indexingScheduler
			}
		}

		// Initialize AI handler
		aiHandlerInstance := aiHandler.NewAIHandler(aiChatUseCase, aiEmbeddingUseCase, moodUseCase, funFactUseCase, auditLogger)

		// Register AI routes under protected API group. Per-user rate limit
		// applied after JWT so the bucket key is the authenticated identity
		// (issue #263 ADR-3 — closes token-cost flood DoS surface; the
		// outbound Anthropic sliding-window limiter in anthropic_provider.go
		// is vendor-side and gated all users together).
		aiAPIGroup := router.Group("/api")
		aiAPIGroup.Use(authMiddleware.JWTMiddleware(authUseCase))
		mountPerUserRateLimit(aiAPIGroup, redisCache)
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
				errorKey: err.Error(),
			})
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Server shutting down...", nil)

	// Cancel the lifecycle ctx first so scheduler tick bodies that have
	// captured it short-circuit on the next ctx.Err() check (issue #263 ADR-4).
	cancelServerCtx()

	// Stop AI schedulers (no-op when nil — cfg.AI.Enabled was false).
	stopAIScheduler(aiFactScheduler, "fact", logger)
	stopAIScheduler(aiIndexingScheduler, "indexing", logger)

	// Stop Telegram polling if running
	if telegramPollingService != nil {
		telegramPollingService.Stop()
		logger.Info("Telegram polling service stopped", nil)
	}

	// Stop reminder scheduler
	stopTaskReminderScheduler(taskReminderScheduler, logger)
	if reminderScheduler != nil {
		if err := reminderScheduler.Stop(); err != nil {
			logger.Error("Failed to stop reminder scheduler", map[string]interface{}{
				errorKey: err.Error(),
			})
		} else {
			logger.Info("Reminder scheduler stopped", nil)
		}
	}

	// Stop integration scheduler
	if integrationModule != nil && integrationModule.IsEnabled() {
		if err := integrationModule.StopScheduler(); err != nil {
			logger.Error("Failed to stop integration scheduler", map[string]interface{}{
				errorKey: err.Error(),
			})
		} else {
			logger.Info("Integration scheduler stopped", nil)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", map[string]interface{}{
			errorKey: err.Error(),
		})
	}

	if tracer != nil {
		if err := tracer.Shutdown(context.Background()); err != nil {
			logger.Error("Failed to shutdown tracer", map[string]interface{}{
				errorKey: err.Error(),
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
) (*usecases.AuthUseCase, usecases.UserRepository) {
	// JWT secrets from config
	jwtSecret := []byte(cfg.JWT.AccessSecret)
	refreshSecret := []byte(cfg.JWT.RefreshSecret)
	mfaIntermediateSecret := []byte(cfg.JWT.MFAIntermediateSecret)

	// Initialize base user repository.
	// v0.159.0 ADR-4: when MFA_SECRET_ENC_KEY is configured (64 hex chars
	// = 32-byte AES-256 KEK), wire the KEK so users.mfa_secret is
	// AES-GCM-encrypted at rest. Missing / malformed KEK degrades to the
	// legacy plaintext path with a warning — production deployments
	// MUST set the env var.
	baseUserRepo := persistence.NewUserRepositoryPG(db)
	if kek, err := authCrypto.ParseKEKHex(os.Getenv("MFA_SECRET_ENC_KEY")); err == nil {
		baseUserRepo = baseUserRepo.WithMFASecretKEK(kek)
		log.Println("[auth] MFA secret KEK configured — at-rest encryption ENABLED")
	} else if !errors.Is(err, authCrypto.ErrEmptyKEK) {
		log.Printf("[auth] MFA_SECRET_ENC_KEY malformed — at-rest encryption DISABLED (deploy fix required): %v", err)
	} else {
		log.Println("[auth] MFA_SECRET_ENC_KEY not configured — at-rest encryption DISABLED (legacy plaintext path)")
	}

	// Wrap with caching if Redis is available
	var userRepo interface{} = baseUserRepo
	if redisCache != nil {
		userCache := cache.NewUserCache(redisCache, 5*time.Minute)
		userRepo = persistence.NewCachedUserRepository(baseUserRepo, userCache, perfLog)
	}

	// Initialize use case with full logging and session repository
	authUseCase := usecases.NewAuthUseCase(
		userRepo.(usecases.UserRepository),
		jwtSecret,
		refreshSecret,
		mfaIntermediateSecret,
		securityLog,
		auditLog,
		notificationUseCase,
	)

	// v0.159.0 ADR-3: wire the per-account brute-force lockout when
	// Redis is available. Threshold + window are env-driven so ops can
	// tune without a redeploy. Without Redis the legacy IP-keyed rate
	// limiter remains the only floor (insufficient for production —
	// the deploy must include Redis).
	if redisCache != nil {
		threshold := envInt("LOGIN_LOCKOUT_THRESHOLD", 5)
		window := envDuration("LOGIN_LOCKOUT_WINDOW", 15*time.Minute)
		tracker := persistence.NewRedisLoginAttemptTracker(redisCache.Client(), threshold, window)
		authUseCase = authUseCase.WithLoginAttemptTracking(tracker)
		log.Printf("[auth] Login lockout wired: %d failures / %s window", threshold, window)
	} else {
		log.Println("[auth] Redis unavailable — per-account lockout DISABLED (legacy IP-only rate limit)")
	}

	return authUseCase, userRepo.(usecases.UserRepository)
}

// envInt reads an integer from env with a fallback default — used by
// the lockout / encryption wiring for ops-tunable knobs.
func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

// envDuration reads a Go-syntax duration (e.g. "15m", "1h") from env
// with a fallback default. Invalid input falls back rather than aborts
// startup — diploma scope prefers liveness over strict validation.
func envDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
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
	taskReminderHandler *taskHandler.TaskReminderHandler,
	saveGradeUseCase *assignUsecases.SaveGradeUseCase,
	returnSubmissionUseCase *assignUsecases.ReturnSubmissionUseCase,
	resubmitSubmissionUseCase *assignUsecases.ResubmitSubmissionUseCase,
	listAssignmentsUseCase *assignUsecases.ListAssignmentsUseCase,
	getAssignmentUseCase *assignUsecases.GetAssignmentUseCase,
	listSubmissionsUseCase *assignUsecases.ListSubmissionsUseCase,
	listMyAssignmentsUseCase *assignUsecases.ListMyAssignmentsUseCase,
	getMyAssignmentDetailUseCase *assignUsecases.GetMyAssignmentDetailUseCase,
	createCurriculumUseCase *curUsecases.CreateCurriculumUseCase,
	getCurriculumUseCase *curUsecases.GetCurriculumUseCase,
	listCurriculaUseCase *curUsecases.ListCurriculaUseCase,
	updateCurriculumUseCase *curUsecases.UpdateCurriculumUseCase,
	submitCurriculumUseCase *curUsecases.SubmitForApprovalUseCase,
	approveCurriculumUseCase *curUsecases.ApproveCurriculumUseCase,
	rejectCurriculumUseCase *curUsecases.RejectCurriculumUseCase,
	createSectionUseCase *curUsecases.CreateSectionUseCase,
	getSectionUseCase *curUsecases.GetSectionUseCase,
	listSectionsUseCase *curUsecases.ListSectionsByCurriculumUseCase,
	updateSectionUseCase *curUsecases.UpdateSectionUseCase,
	deleteSectionUseCase *curUsecases.DeleteSectionUseCase,
	createDisciplineItemUseCase *curUsecases.CreateDisciplineItemUseCase,
	getDisciplineItemUseCase *curUsecases.GetDisciplineItemUseCase,
	listDisciplineItemsUseCase *curUsecases.ListDisciplineItemsBySectionUseCase,
	updateDisciplineItemUseCase *curUsecases.UpdateDisciplineItemUseCase,
	deleteDisciplineItemUseCase *curUsecases.DeleteDisciplineItemUseCase,
	bulkEditDisciplineItemsUseCase *curUsecases.BulkEditDisciplineItemsUseCase,
	annualReportUseCase *annualUsecases.AnnualReportUseCase,
	eventUseCase *scheduleUsecases.EventUseCase,
	lessonUseCase *scheduleUsecases.LessonUseCase,
	lessonSlotUseCase *scheduleUsecases.LessonSlotUseCase,
	teachingLoadUseCase *scheduleUsecases.TeachingLoadUseCase,
	generateScheduleUseCase *scheduleUsecases.GenerateScheduleUseCase,
	calendarFeedUseCase *scheduleUsecases.CalendarFeedUseCase,
	announcementUseCase *announcementUsecases.AnnouncementUseCase,
	dashboardUseCase *dashboardUsecases.DashboardUseCase,
	analyticsUseCase *analyticsUsecases.AnalyticsUseCase,
	teacherScopeRepo analyticsUsecases.TeacherScopeRepository,
	userUseCase *usersUsecases.UserUseCase,
	departmentUseCase *usersUsecases.DepartmentUseCase,
	positionUseCase *usersUsecases.PositionUseCase,
	fileUseCase *filesUsecases.FileUseCase,
	versionUseCase *filesUsecases.VersionUseCase,
	notificationUseCase *notifUsecases.NotificationUseCase,
	preferencesUseCase *notifUsecases.PreferencesUseCase,
	telegramVerificationService *notifServices.TelegramVerificationService,
	telegramService emailDomain.TelegramService,
	webpushRepo notifUsecases.WebPushRepository,
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
	userRepo usecases.UserRepository,
	userProfileRepo usersUsecases.UserProfileRepository,
	validator *validation.Validator,
	jwtSecret []byte,
	submitDocUseCase *docUsecases.SubmitDocumentUseCase,
	approveDocUseCase *docUsecases.ApproveDocumentUseCase,
	rejectDocUseCase *docUsecases.RejectDocumentUseCase,
	registerDocUseCase *docUsecases.RegisterDocumentUseCase,
	startRoutingDocUseCase *docUsecases.StartRoutingUseCase,
	signVisaDocUseCase *docUsecases.SignVisaUseCase,
	assignExecutorDocUseCase *docUsecases.AssignExecutorUseCase,
	markExecutedDocUseCase *docUsecases.MarkExecutedUseCase,
	archiveDocUseCase *docUsecases.ArchiveDocumentUseCase,
	resubmitDocUseCase *docUsecases.ResubmitDocumentUseCase,
) (*gin.Engine, *telegram.PollingService) {
	router := gin.New()
	var telegramPollingService *telegram.PollingService

	// Token revocation infrastructure (logout endpoint, AUDIT_REPORT item #4).
	// If Redis is unavailable, revokedTokenRepo stays nil and JWTMiddlewareWithRevocation
	// gracefully degrades to plain validation without blacklist lookup.
	var revokedTokenRepo usecases.RevokedTokenRepository
	var logoutUseCase *usecases.LogoutUseCase
	if redisCache != nil {
		revokedTokenRepo = persistence.NewRedisRevokedTokenRepository(redisCache.Client())
		logoutUseCase = usecases.NewLogoutUseCase(revokedTokenRepo, jwtSecret)
		// Two independent concerns share the same revoked-token store:
		//   - MFA verify-login replay guard (±1 step ≈ 30 s drift)
		//   - Refresh-token rotation + RFC 6749 §10.4 reuse-detection
		//     cascade (v0.159.0 ADR-2)
		// Setters are commutative; either alone enables the receiver-
		// side field, both make the wiring intent explicit.
		authUseCase.WithMFAVerification(revokedTokenRepo, 1, time.Now)
		authUseCase.WithRefreshRotation(revokedTokenRepo)
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
		c.JSON(http.StatusNotFound, gin.H{errorKey: "route not found"})
	})

	// Health check endpoints for Kubernetes probes
	router.GET("/health", healthCheckHandler(db, redisCache))
	router.GET("/live", livenessHandler())
	router.GET("/ready", readinessHandler(db, redisCache))

	// Prometheus metrics endpoint
	router.GET("/metrics", metrics.Handler())

	// Swagger documentation endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Загрузка конфигурации rate limiting + trusted-proxy CIDR
	// allowlist (v0.159.0 ADR-3b). With no TRUSTED_PROXY_CIDRS env
	// configured the limiter ignores X-Forwarded-For entirely — secure
	// default for direct-internet deployments. Behind a reverse proxy
	// set TRUSTED_PROXY_CIDRS to the proxy's egress CIDR(s) so XFF
	// from those hops is honored.
	rateLimitConfig := middleware.LoadRateLimitConfig()
	trustedProxyCIDRs := middleware.ParseTrustedProxyCIDRs(os.Getenv("TRUSTED_PROXY_CIDRS"))
	if len(trustedProxyCIDRs) == 0 {
		logger.Info("Trusted-proxy CIDRs not configured — X-Forwarded-For will be ignored (secure default for direct-internet deployments)", nil)
	} else {
		logger.Info("Trusted-proxy CIDRs configured", map[string]interface{}{
			"count": len(trustedProxyCIDRs),
		})
	}

	var publicRateLimiter, authRateLimiter *middleware.RateLimiter
	if redisCache != nil {
		publicRateLimiter = rateLimitConfig.GetPublicRateLimiter(redisCache.Client()).WithTrustedProxies(trustedProxyCIDRs)
		authRateLimiter = rateLimitConfig.GetAuthRateLimiter(redisCache.Client()).WithTrustedProxies(trustedProxyCIDRs)
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

	// Public auth routes with rate limiting (configured via
	// RATE_LIMIT_PUBLIC_RPM / RATE_LIMIT_PUBLIC_BURST env vars; defaults
	// 300 RPM + 100 burst — see middleware.LoadRateLimitConfig). Per-
	// account brute-force lockout is layered on top inside AuthUseCase
	// (v0.159.0 ADR-3).
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
		// MFA verify-login: completes the second factor of an MFA-gated
		// login. The intermediate token IS the authorisation, so this
		// route stays under the same public-rate-limit group as /login —
		// no JWT middleware, no role gate.
		authGroup.POST("/mfa/verify-login", authHandlerInstance.VerifyMFALogin)
		authGroup.OPTIONS("/mfa/verify-login", func(c *gin.Context) { c.Status(http.StatusNoContent) })

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

		// MFA enrollment (v0.124.0). system_admin only — visible defense
		// hardening that demonstrates the auth surface during the diploma
		// review. Login flow MFA gating is deferred to a follow-up release.
		mfaUseCase := usecases.NewMFAUseCase(userRepo, auditLogger, "inf-sys-secretary-methodist")
		mfaHandlerInstance := authHandler.NewMFAHandler(mfaUseCase)
		authGroup.POST("/mfa/begin",
			authMiddleware.JWTMiddleware(authUseCase),
			authMiddleware.RequireRole("system_admin"),
			mfaHandlerInstance.Begin,
		)
		authGroup.POST("/mfa/confirm",
			authMiddleware.JWTMiddleware(authUseCase),
			authMiddleware.RequireRole("system_admin"),
			mfaHandlerInstance.Confirm,
		)
		authGroup.POST("/mfa/disable",
			authMiddleware.JWTMiddleware(authUseCase),
			authMiddleware.RequireRole("system_admin"),
			mfaHandlerInstance.Disable,
		)
		authGroup.OPTIONS("/mfa/begin", func(c *gin.Context) { c.Status(http.StatusNoContent) })
		authGroup.OPTIONS("/mfa/confirm", func(c *gin.Context) { c.Status(http.StatusNoContent) })
		authGroup.OPTIONS("/mfa/disable", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	}

	// Branding module wiring (v0.136.0). Singleton DB-backed
	// settings used by both /admin/branding (system_admin write) and
	// /public/branding (unauth read for the login page). Use cases
	// composed here once and shared by both handler groups.
	brandingRepo := brandingPersistence.NewBrandSettingsRepositoryPG(db)
	brandingGetUC := brandingUseCases.NewGetBrandingUseCase(brandingRepo)
	brandingUpdateUC := brandingUseCases.NewUpdateBrandingUseCase(
		brandingRepo,
		brandingUseCases.SystemClock{},
		auditLogger,
	)
	adminBrandingHandler := brandingHandlers.NewAdminBrandingHandler(brandingGetUC, brandingUpdateUC)
	publicBrandingHandler := brandingHandlers.NewPublicBrandingHandler(brandingGetUC)

	// Public branding route — always available (login page consumes
	// this before the user authenticates). Mounted independently of
	// the sharing-feature block below so a sharing-disabled deployment
	// still surfaces the branded login chrome. Admin branding routes
	// are mounted further down in the adminGroup block; both calls
	// share the same handler and use case instances.
	brandingPublicGroup := router.Group("/api/public")
	if publicRateLimiter != nil {
		brandingPublicGroup.Use(publicRateLimiter.RateLimitMiddleware())
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

	// Public calendar feed (no authentication — external calendar clients
	// subscribe with a secret token in the URL). Issue #40.
	if calendarFeedUseCase != nil {
		calendarFeedHandler := scheduleHandler.NewCalendarFeedHandler(calendarFeedUseCase, cfg.Server.BaseURL)
		calendarPublicGroup := router.Group("/api/public")
		if publicRateLimiter != nil {
			calendarPublicGroup.Use(publicRateLimiter.RateLimitMiddleware())
		}
		calendarPublicGroup.GET("/calendar/:token/feed.ics", calendarFeedHandler.ServeFeed)
		logger.Info("Public calendar feed route registered", nil)
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
				logger.Warn("Failed to start Telegram polling", map[string]interface{}{errorKey: err.Error()})
			} else {
				logger.Info("Telegram polling mode started (no webhook URL configured)", nil)
			}
		}
	}

	// Protected routes (require JWT) with auth rate limiting
	// (configured via RATE_LIMIT_AUTH_RPM / RATE_LIMIT_AUTH_BURST env
	// vars; defaults 1000 RPM + 200 burst — see LoadRateLimitConfig).
	protectedGroup := router.Group("/api")
	protectedGroup.Use(authMiddleware.JWTMiddlewareWithRevocation(authUseCase, revokedTokenRepo))
	if authRateLimiter != nil {
		protectedGroup.Use(authRateLimiter.RateLimitMiddleware())
	}
	{
		protectedGroup.GET("/me", func(c *gin.Context) {
			userID, exists := c.Get("user_id")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{errorKey: "user not authenticated"})
				return
			}

			// Get full user data with profile from database
			user, err := userProfileRepo.GetProfileByID(c.Request.Context(), userID.(int64))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{errorKey: "failed to get user data"})
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
						errorKey:  err.Error(),
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

			// v0.148.0 — workflow gates (issue #227). Submit endpoint
			// stays на /documents (non-student gate), approve/reject
			// move к /admin/documents с secretary+admin role guard.
			if submitDocUseCase != nil && approveDocUseCase != nil && rejectDocUseCase != nil {
				workflowHandler := docHandler.NewWorkflowHandler(submitDocUseCase, approveDocUseCase, rejectDocUseCase, registerDocUseCase, startRoutingDocUseCase, signVisaDocUseCase, assignExecutorDocUseCase, markExecutedDocUseCase, archiveDocUseCase, resubmitDocUseCase)
				docSubmitGroup := protectedGroup.Group("/documents")
				docSubmitGroup.Use(authMiddleware.RequireNonStudent())
				docHandler.RegisterSubmitRoute(docSubmitGroup, workflowHandler)
				docSubmitGroup.OPTIONS("/:id/submit", func(c *gin.Context) { c.Status(http.StatusNoContent) })

				adminDocsGroup := protectedGroup.Group("/admin/documents")
				adminDocsGroup.Use(authMiddleware.RequireRole(
					string(authDomain.RoleAcademicSecretary),
					string(authDomain.RoleSystemAdmin),
				))
				docHandler.RegisterAdminWorkflowRoutes(adminDocsGroup, workflowHandler)
				adminDocsGroup.OPTIONS("/:id/approve", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				adminDocsGroup.OPTIONS("/:id/reject", func(c *gin.Context) { c.Status(http.StatusNoContent) })

				logger.Info("Documents workflow routes registered (v0.148.0 #227)", nil)
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
			// Students have no business with custom reports — block the
			// entire group at the route layer. Closes the privilege-escalation
			// surface flagged in #260; mirrors reportsGroup precedent (line 2012).
			customReportsGroup.Use(authMiddleware.RequireNonStudent())
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

			// Task reminder routes (v0.138.0 — Phase 5 #5 final).
			// Mounted under the same /tasks/:id path tree but in a
			// dedicated registrar so the surface stays modular.
			// No adminMW parameter — reminder endpoints are
			// user-self-scoped (per-user privacy at use-case layer).
			taskRoutes.RegisterTaskReminderRoutes(protectedGroup, taskReminderHandler)
			logger.Info("Task reminder routes registered", nil)
		}

		// Assignments module routes — academic grading flow.
		// Behind RequireNonStudent because students must not see (or
		// even probe) grading endpoints for their peers.
		if saveGradeUseCase != nil {
			gradeHandlerInstance := assignHandler.NewGradeHandler(saveGradeUseCase)
			returnHandlerInstance := assignHandler.NewReturnHandler(returnSubmissionUseCase)
			assignmentsHandler := assignHandler.NewAssignmentsHandler(
				listAssignmentsUseCase, getAssignmentUseCase, listSubmissionsUseCase,
			)

			assignmentsGroup := protectedGroup.Group("/assignments")
			assignmentsGroup.Use(authMiddleware.RequireNonStudent())
			{
				assignmentsGroup.GET("", assignmentsHandler.ListAssignments)
				assignmentsGroup.GET("/:id", assignmentsHandler.GetAssignment)
				assignmentsGroup.GET("/:id/submissions", assignmentsHandler.ListSubmissions)
				assignmentsGroup.POST("/:id/grades", gradeHandlerInstance.SaveGrade)
				assignmentsGroup.OPTIONS("/:id/grades", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				assignmentsGroup.POST("/:id/returns", returnHandlerInstance.Return)
				assignmentsGroup.OPTIONS("/:id/returns", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			}

			// Student-only routes. Lives in a sibling group rather than
			// under assignmentsGroup because that group is gated by
			// RequireNonStudent — exactly the inverse of what student
			// endpoints require. The dedicated RequireRole("student")
			// middleware here plus the handler-level studentIDFromContext
			// whitelist give defense in depth: removing either one alone
			// still rejects every non-student request.
			resubmitHandlerInstance := assignHandler.NewResubmitHandler(resubmitSubmissionUseCase)
			myAssignmentsHandler := assignHandler.NewMyAssignmentsHandler(
				listMyAssignmentsUseCase, getMyAssignmentDetailUseCase,
			)
			studentAssignmentsGroup := protectedGroup.Group("/assignments")
			studentAssignmentsGroup.Use(authMiddleware.RequireRole("student"))
			{
				studentAssignmentsGroup.GET("/my", myAssignmentsHandler.List)
				studentAssignmentsGroup.GET("/:id/my", myAssignmentsHandler.Detail)
				studentAssignmentsGroup.POST("/:id/resubmit", resubmitHandlerInstance.Resubmit)
				studentAssignmentsGroup.OPTIONS("/:id/resubmit", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			}

			logger.Info("Assignments module routes registered", nil)
		}

		// Curriculum module routes. Behind RequireNonStudent because
		// the read scope is the four non-student roles; student read
		// with specialty filtering is a future scope.
		if createCurriculumUseCase != nil {
			curriculumHandler := curHandler.NewCurriculumHandler(
				createCurriculumUseCase, getCurriculumUseCase, listCurriculaUseCase, updateCurriculumUseCase,
				submitCurriculumUseCase, approveCurriculumUseCase, rejectCurriculumUseCase,
			)
			curriculumGroup := protectedGroup.Group("/curriculum")
			curriculumGroup.Use(authMiddleware.RequireNonStudent())
			{
				curriculumGroup.POST("", curriculumHandler.Create)
				curriculumGroup.OPTIONS("", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				curriculumGroup.GET("", curriculumHandler.List)
				curriculumGroup.GET("/:id", curriculumHandler.Get)
				curriculumGroup.PUT("/:id", curriculumHandler.Update)
				curriculumGroup.OPTIONS("/:id", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				// v0.117.0 — Submit (methodist or admin). Lives under
				// the non-student group because it's a write that the
				// methodist (the curriculum's author) initiates.
				curriculumGroup.POST("/:id/submit", curriculumHandler.Submit)
				curriculumGroup.OPTIONS("/:id/submit", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			}

			// Methodist + admin sibling group for Approve / Reject.
			// Lives in a parallel group rather than under
			// curriculumGroup because that group is gated by
			// RequireNonStudent — the inverse of what these
			// approval endpoints require. Per the diploma's role
			// matrix, the methodist is the approver of curricula
			// authored by the academic secretary; system_admin
			// retains an emergency override (v0.158.0).
			//
			// Mirrors the assignments v0.112.0 student-sibling pattern:
			// when a subset of routes needs an inverse middleware to
			// its sibling, register a parallel group instead of
			// special-casing one. The handler-level canApprove
			// whitelist is defense in depth on top of RequireRole.
			approverCurriculumGroup := protectedGroup.Group("/curriculum")
			approverCurriculumGroup.Use(authMiddleware.RequireRole(
				string(authDomain.RoleMethodist),
				string(authDomain.RoleSystemAdmin),
			))
			{
				approverCurriculumGroup.POST("/:id/approve", curriculumHandler.Approve)
				approverCurriculumGroup.OPTIONS("/:id/approve", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				approverCurriculumGroup.POST("/:id/reject", curriculumHandler.Reject)
				approverCurriculumGroup.OPTIONS("/:id/reject", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			}

			logger.Info("Curriculum module routes registered", nil)
		}

		// Annual methodist report — read-only DOCX download за календарный
		// год; methodist + system_admin only (academic_secretary excluded
		// per ADR-6 — observer не decision-maker). v0.129.0 B4.
		{
			annualReportHandlerInstance := annualHandler.NewAnnualReportHandler(annualReportUseCase)

			annualReportGroup := protectedGroup.Group("/reports/annual")
			annualReportGroup.Use(authMiddleware.RequireRole(
				string(authDomain.RoleMethodist),
				string(authDomain.RoleSystemAdmin),
			))
			{
				annualReportGroup.GET("", annualReportHandlerInstance.Generate)
				annualReportGroup.OPTIONS("", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			}

			logger.Info("Annual report module routes registered", nil)
		}

		// Section module routes (v0.128.0). RequireNonStudent — same scope
		// gate as parent Curriculum routes. Authorization (author methodist
		// vs admin override) is enforced at the use-case layer via
		// AuthorizeSectionEdit; handler does only role-class whitelisting
		// for write methods (canWrite — methodist + system_admin).
		//
		// Routes intentionally split between curriculum-scoped (Create /
		// List, where the curriculum id is a path component) and
		// section-scoped (Get / Update / Delete, addressed by section id
		// directly). This mirrors the standard REST sub-resource +
		// individual-resource convention.
		if createSectionUseCase != nil {
			sectionHandler := curHandler.NewSectionHandler(
				createSectionUseCase, getSectionUseCase, listSectionsUseCase,
				updateSectionUseCase, deleteSectionUseCase,
			)
			sectionGroup := protectedGroup.Group("")
			sectionGroup.Use(authMiddleware.RequireNonStudent())
			{
				sectionGroup.POST("/curricula/:curriculumID/sections", sectionHandler.Create)
				sectionGroup.OPTIONS("/curricula/:curriculumID/sections", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				sectionGroup.GET("/curricula/:curriculumID/sections", sectionHandler.List)
				sectionGroup.GET("/sections/:sectionID", sectionHandler.Get)
				sectionGroup.PUT("/sections/:sectionID", sectionHandler.Update)
				sectionGroup.DELETE("/sections/:sectionID", sectionHandler.Delete)
				sectionGroup.OPTIONS("/sections/:sectionID", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			}
			logger.Info("Section module routes registered", nil)
		}

		// DisciplineItem module routes (v0.128.1). RequireNonStudent — same
		// scope gate as Section. Authorization (author methodist vs admin
		// override) enforced в use-case layer via AuthorizeDisciplineItemEdit;
		// handler does only role-class whitelisting (canWrite for write
		// methods, canRead for read methods).
		//
		// Routes split: section-scoped (Create / List, where section id is
		// path component) и item-scoped (Get / Update / Delete, addressed
		// by item id directly). Mirror к Section routing convention.
		if createDisciplineItemUseCase != nil {
			disciplineItemHandler := curHandler.NewDisciplineItemHandler(
				createDisciplineItemUseCase, getDisciplineItemUseCase, listDisciplineItemsUseCase,
				updateDisciplineItemUseCase, deleteDisciplineItemUseCase,
			)
			disciplineItemGroup := protectedGroup.Group("")
			disciplineItemGroup.Use(authMiddleware.RequireNonStudent())
			{
				disciplineItemGroup.POST("/sections/:sectionID/items", disciplineItemHandler.Create)
				disciplineItemGroup.OPTIONS("/sections/:sectionID/items", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				disciplineItemGroup.GET("/sections/:sectionID/items", disciplineItemHandler.List)
				disciplineItemGroup.GET("/items/:id", disciplineItemHandler.Get)
				disciplineItemGroup.PUT("/items/:id", disciplineItemHandler.Update)
				disciplineItemGroup.DELETE("/items/:id", disciplineItemHandler.Delete)
				disciplineItemGroup.OPTIONS("/items/:id", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			}
			logger.Info("DisciplineItem module routes registered", nil)
		}

		// v0.128.3 Bulk-edit transactional endpoint (B1a Layer 3) per ADRs
		// 10-13. Separate handler + route from per-item DisciplineItem
		// endpoints; commit-or-rollback semantic atomically applies
		// combined creates+updates+deletes batch.
		if bulkEditDisciplineItemsUseCase != nil {
			bulkDisciplineItemsHandler := curHandler.NewBulkDisciplineItemsHandler(bulkEditDisciplineItemsUseCase)
			bulkDisciplineItemsGroup := protectedGroup.Group("")
			bulkDisciplineItemsGroup.Use(authMiddleware.RequireNonStudent())
			{
				bulkDisciplineItemsGroup.POST("/sections/:sectionID/items/bulk", bulkDisciplineItemsHandler.BulkEdit)
				bulkDisciplineItemsGroup.OPTIONS("/sections/:sectionID/items/bulk", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			}
			logger.Info("Bulk DisciplineItem route registered", nil)
		}

		// Extracurricular module routes (B3, v0.164.0) — внеучебные мероприятия.
		// Greenfield bounded context per plan docs/plans/2026-05-24-b3-extracurricular.md.
		// Audience-aware list/get + organizer-only edit/delete + self-register.
		// v0.165.0: EventNotifier wired through main.go adapter (ADR-7 closure).
		{
			extEventRepo := extPersistence.NewEventRepositoryPG(db)
			extNotifier := &extracurricularNotificationNotifier{notif: notificationUseCase}
			createEventUC := extUsecases.NewCreateEventUseCase(extEventRepo, auditLogger, nil)
			updateEventUC := extUsecases.NewUpdateEventUseCase(extEventRepo, auditLogger, extNotifier, nil)
			deleteEventUC := extUsecases.NewDeleteEventUseCase(extEventRepo, auditLogger)
			getEventUC := extUsecases.NewGetEventUseCase(extEventRepo)
			listEventsUC := extUsecases.NewListEventsUseCase(extEventRepo)
			registerParticipantUC := extUsecases.NewRegisterParticipantUseCase(extEventRepo, auditLogger, nil)
			unregisterParticipantUC := extUsecases.NewUnregisterParticipantUseCase(extEventRepo, auditLogger)

			extracurricularHandler := extHandler.NewEventHandler(
				createEventUC, updateEventUC, deleteEventUC, getEventUC,
				listEventsUC, registerParticipantUC, unregisterParticipantUC,
			)
			// RegisterExtracurricularRoutes contract (per its docstring) mounts
			// endpoints under /api/v1/extracurricular — wrap the protected
			// group in /v1 so the call site honors that promise. Without this
			// wrap the routes land on /api/extracurricular and the B3
			// frontend slice (which calls /api/v1/...) receives 404.
			v1Group := protectedGroup.Group("/v1")
			extHandler.RegisterExtracurricularRoutes(v1Group, extracurricularHandler)
			logger.Info("Extracurricular module routes registered", nil)
		}

		// Work program (РПД) module routes (PR 4a, v0.180.0) — рабочая
		// программа дисциплины. Greenfield bounded context per plan
		// docs/plans/2026-05-27-work-program-initiative.md. PR 4a mounts
		// read + create (POST / GET list / GET by id); transition
		// endpoints (submit/approve/reject/discard) land in PR 4b.
		{
			wpRepo := wpPersistence.NewWorkProgramRepositoryPG(db)
			createWPUC := wpUsecases.NewCreateWorkProgramUseCase(wpRepo, auditLogger)
			getWPUC := wpUsecases.NewGetWorkProgramUseCase(wpRepo, auditLogger)
			listWPUC := wpUsecases.NewListWorkProgramsUseCase(wpRepo, auditLogger)
			submitWPUC := wpUsecases.NewSubmitWorkProgramUseCase(wpRepo, auditLogger)
			approveWPUC := wpUsecases.NewApproveWorkProgramUseCase(wpRepo, auditLogger)
			rejectWPUC := wpUsecases.NewRejectWorkProgramUseCase(wpRepo, auditLogger)
			discardWPUC := wpUsecases.NewDiscardDraftWorkProgramUseCase(wpRepo, auditLogger)

			// LLM draft generation (PR 5b / v0.189.0): OpenAI-compatible
			// generator (OpenRouter by default) + curriculum discipline
			// enrichment + Redis rate limit. Generation degrades to
			// allow-all when Redis is unavailable so the feature still
			// works (the cost guard is best-effort, not a hard dependency).
			wpDraftGen := wpLLM.NewGenerator(wpLLM.Config{
				BaseURL:     cfg.AI.GenerationBaseURL,
				APIKey:      cfg.AI.GenerationAPIKey,
				Model:       cfg.AI.GenerationModel,
				Timeout:     cfg.AI.Timeout,
				Temperature: cfg.AI.GenerationTemperature,
				MaxTokens:   cfg.AI.GenerationMaxTokens,
			})
			wpDisciplineInfo := newDisciplineInfoAdapter(curPersistence.NewDisciplineItemRepositoryPG(db))
			var wpGenLimiter wpUsecases.GenerationRateLimiter = allowAllGenerationLimiter{}
			if redisCache != nil {
				wpGenLimiter = wpRateLimit.NewGenerationLimiter(redisCache.Client(), 5, time.Hour)
			} else {
				logger.Warn("РПД draft generation rate limiting disabled — Redis unavailable; LLM spend is uncapped", nil)
			}
			generateWPUC := wpUsecases.NewGenerateDraftUseCase(wpRepo, wpDraftGen, wpDisciplineInfo, wpGenLimiter, auditLogger)

			workProgramHandler := wpHandler.NewWorkProgramHandler(
				createWPUC, getWPUC, listWPUC,
				submitWPUC, approveWPUC, rejectWPUC, discardWPUC,
				generateWPUC,
			)
			// Routes mount under /api/v1/work-programs — wrap the
			// protected group in /v1 (mirror extracurricular) so the
			// path matches the documented /api/v1 contract.
			wpV1Group := protectedGroup.Group("/v1")
			wpHandler.RegisterWorkProgramRoutes(wpV1Group, workProgramHandler)
			logger.Info("Work program (РПД) module routes registered", nil)

			// Revision (лист актуализации) write-workflow (PR 1-3,
			// v0.197.0-v0.199.0) — create/submit (author) + approve/reject
			// (methodist) nested under /work-programs/:id/revisions. All
			// reuse the same wpRepo (load-mutate-persist through the
			// aggregate root) + audit sink.
			createRevUC := wpUsecases.NewCreateRevisionUseCase(wpRepo, auditLogger)
			submitRevUC := wpUsecases.NewSubmitRevisionUseCase(wpRepo, auditLogger)
			approveRevUC := wpUsecases.NewApproveRevisionUseCase(wpRepo, auditLogger)
			rejectRevUC := wpUsecases.NewRejectRevisionUseCase(wpRepo, auditLogger)
			revisionHandler := wpHandler.NewRevisionHandler(createRevUC, submitRevUC, approveRevUC, rejectRevUC)
			wpHandler.RegisterRevisionRoutes(wpV1Group, revisionHandler)
			logger.Info("Revision (лист актуализации) routes registered", nil)

			// Manual collection edit (slice 12b-1, v0.210.0) — methodist or
			// РПД author hand-edits goals / competences / topics nested
			// under /work-programs/:id/{goals,competences,topics}. Reuses the
			// same wpRepo (load-mutate-persist through the aggregate root) +
			// audit sink; author-scoped + status-gated in the use case/domain.
			contentUC := wpUsecases.NewWorkProgramContentUseCase(wpRepo, auditLogger)
			contentHandler := wpHandler.NewWorkProgramContentHandler(contentUC)
			wpHandler.RegisterWorkProgramContentRoutes(wpV1Group, contentHandler)
			logger.Info("Work program content (manual collection edit) routes registered", nil)

			// Минобрнауки order register (PR 6b-2, v0.194.0) — ADR-11.
			// Record (staff only) / Get / List on the same /api/v1 group.
			moRepo := wpPersistence.NewMinobrnaukiOrderRepositoryPG(db)
			recordMOUC := wpUsecases.NewRecordMinobrnaukiOrderUseCase(moRepo, auditLogger)
			// Trigger-revision (PR 6c, v0.195.0): on record, drive affected
			// approved РПД into needs_revision and delegate a revision task
			// to each program's author via the tasks module. Wired only when
			// the tasks use case is available; absence degrades gracefully to
			// standalone order recording.
			if taskUseCase != nil {
				revisionDelegator := minobrnaukiRevisionTaskDelegator{tasks: taskUseCase}
				triggerRevisionsUC := wpUsecases.NewTriggerOrderRevisionsUseCase(wpRepo, revisionDelegator, auditLogger)
				recordMOUC = recordMOUC.WithRevisionTrigger(triggerRevisionsUC)
				logger.Info("Минобрнауки revision trigger wired (needs_revision + teacher delegation)", nil)
			}
			getMOUC := wpUsecases.NewGetMinobrnaukiOrderUseCase(moRepo)
			listMOUC := wpUsecases.NewListMinobrnaukiOrdersUseCase(moRepo)
			moHandler := wpHandler.NewMinobrnaukiOrderHandler(recordMOUC, getMOUC, listMOUC)
			wpHandler.RegisterMinobrnaukiOrderRoutes(wpV1Group, moHandler)
			logger.Info("Минобрнауки order routes registered", nil)

			// AI bulk-revision (PR 11, ADR-12): a methodist triggers LLM
			// generation of a draft лист актуализации for every РПД affected
			// by an order. Reuses the order repo (source + FindAffected), the
			// РПД repo (load-mutate-persist), the same OpenRouter generator as
			// draft generation (it also implements RevisionDraftGenerator), and
			// the shared generation rate limiter. The drafts land for the РПД
			// author to submit and the methodist to approve via the revision
			// flow — never silently applied.
			// Slice 7: when an order has an attached document (PDF/DOCX), feed
			// its extracted text to the LLM so the revision is grounded on the
			// real приказ, not just the manual summary. Bridge to the documents +
			// text-extraction infrastructure (own DocumentAdapter instance — the
			// AI module's is scoped to its own block); best-effort, so a missing
			// document or extraction failure leaves generation working from the
			// summary alone.
			orderDocText := orderDocumentTextAdapter{docs: aiAdapters.NewDocumentAdapter(
				docPersistence.NewDocumentRepositoryPG(db),
				s3Client,
				aiServices.NewTextExtractionService(),
				db,
				slog.Default(),
			)}
			generateOrderRevsUC := wpUsecases.NewGenerateOrderRevisionsUseCase(
				moRepo, wpRepo, wpDraftGen, wpGenLimiter, auditLogger,
			).WithDocumentText(orderDocText)
			genRevHandler := wpHandler.NewGenerateOrderRevisionsHandler(generateOrderRevsUC)
			wpHandler.RegisterGenerateOrderRevisionsRoutes(wpV1Group, genRevHandler)
			logger.Info("AI bulk-revision (generate-revisions) route registered", nil)
		}

		// Student debts (Долги студентов) module — read endpoints (PR5b).
		// The teacher scope resolves a teacher's disciplines from the
		// schedule (schedule_lessons → discipline_id); migration 051
		// realigned student_debts.discipline_id onto disciplines(id) so both
		// share one id space. auditLogger satisfies the AuditSink port
		// structurally; the read use cases need no notifier.
		{
			sdRepo := sdPersistence.NewStudentDebtRepositoryPG(db)
			sdTeacherScope := sdPersistence.NewTeacherScopeResolverPG(db)

			getDebtUC := sdUsecases.NewGetDebtUseCase(sdRepo, sdTeacherScope, auditLogger)
			listDebtsUC := sdUsecases.NewListDebtsUseCase(sdRepo, sdTeacherScope, auditLogger)
			listMyDebtsUC := sdUsecases.NewListMyDebtsUseCase(sdRepo)
			debtStatsUC := sdUsecases.NewGetDebtStatsUseCase(sdRepo, sdTeacherScope, auditLogger)

			debtHandler := sdHandler.NewStudentDebtHandler(getDebtUC, listDebtsUC, listMyDebtsUC, debtStatsUC)
			sdHandler.RegisterStudentDebtRoutes(protectedGroup, debtHandler)
			logger.Info("Student debts module read routes registered", nil)

			// Write endpoints (resit lifecycle, PR5c). The resit notifier
			// detaches the request context (see studentDebtResitNotifier).
			// attemptsBeforeCommission is the policy N: regular failed
			// attempts before the debt escalates to a commission resit.
			const debtAttemptsBeforeCommission = 2
			sdResitNotifier := &studentDebtResitNotifier{notificationUseCase: notificationUseCase}
			scheduleResitUC := sdUsecases.NewScheduleResitUseCase(sdRepo, sdResitNotifier, auditLogger, time.Now)
			recordResitUC := sdUsecases.NewRecordResitResultUseCase(sdRepo, auditLogger, time.Now, debtAttemptsBeforeCommission)

			debtWriteHandler := sdHandler.NewStudentDebtWriteHandler(scheduleResitUC, recordResitUC)
			sdHandler.RegisterStudentDebtWriteRoutes(protectedGroup, debtWriteHandler)
			logger.Info("Student debts module write routes registered", nil)

			// Bulk transfer endpoints (import/export, PR5d). The Excel
			// adapter implements both DebtImporter and DebtExporter ports;
			// export reuses the same teacher scope as the read path.
			sdImporter := sdExcel.NewDebtImporter()
			sdExporter := sdExcel.NewDebtExporter()
			importUC := sdUsecases.NewImportDebtsUseCase(sdRepo, sdImporter, auditLogger)
			exportUC := sdUsecases.NewExportDebtsUseCase(sdRepo, sdTeacherScope, sdExporter, auditLogger)

			debtTransferHandler := sdHandler.NewStudentDebtTransferHandler(importUC, exportUC)
			sdHandler.RegisterStudentDebtTransferRoutes(protectedGroup, debtTransferHandler)
			logger.Info("Student debts module transfer routes registered", nil)
		}

		// Document e-signatures (#140) — cryptographic sign/list/verify.
		// Requires object storage (to hash file bodies) and DOC_SIGNING_ENC_KEY
		// (32-byte AES-256 KEK, 64 hex chars) to encrypt per-user private keys
		// at rest. Absent/malformed key => feature disabled (graceful degrade).
		if s3Client != nil {
			if sigKEK, kekErr := authCrypto.ParseKEKHex(os.Getenv("DOC_SIGNING_ENC_KEY")); kekErr == nil {
				sigDocRepo := docPersistence.NewDocumentRepositoryPG(db)
				sigRepo := docPersistence.NewSignatureRepositoryPG(db)
				sigEngine := docSigning.NewService(db, sigKEK, time.Now)
				sigView := docSigning.NewDocumentView(sigDocRepo, s3Client)

				var _ docUsecases.SignatureEngine = sigEngine // compile-time port check

				signUC := docUsecases.NewSignDocumentUseCase(sigRepo, sigView, sigEngine, auditLogger, time.Now)
				listSigUC := docUsecases.NewListSignaturesUseCase(sigRepo)
				verifySigUC := docUsecases.NewVerifySignatureUseCase(sigRepo, sigView, sigEngine)
				nameResolver := signerNameResolverAdapter{userRepo: userRepo}

				sigHandler := docHandler.NewSignatureHandler(signUC, listSigUC, verifySigUC, nameResolver)
				docHandler.RegisterSignatureRoutes(protectedGroup, sigHandler)
				logger.Info("Document e-signature routes registered", nil)
			} else if !errors.Is(kekErr, authCrypto.ErrEmptyKEK) {
				log.Printf("[documents] DOC_SIGNING_ENC_KEY malformed — e-signature DISABLED (deploy fix required): %v", kekErr)
			} else {
				log.Println("[documents] DOC_SIGNING_ENC_KEY not configured — e-signature DISABLED")
			}
		}

		// Calendar feed subscription management (authenticated). Issue #40.
		if calendarFeedUseCase != nil {
			calendarFeedHandlerInstance := scheduleHandler.NewCalendarFeedHandler(calendarFeedUseCase, cfg.Server.BaseURL)
			subGroup := protectedGroup.Group("/schedule/calendar-subscription")
			{
				subGroup.GET("", calendarFeedHandlerInstance.GetSubscription)
				subGroup.POST("", calendarFeedHandlerInstance.CreateSubscription)
				subGroup.POST("/rotate", calendarFeedHandlerInstance.RotateSubscription)
				subGroup.DELETE("", calendarFeedHandlerInstance.DeleteSubscription)
			}
			logger.Info("Calendar subscription routes registered", nil)
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

				// Bell-schedule catalog (lesson slots) — issue #139, Slice 1.
				slotHandlerInstance := scheduleHandler.NewLessonSlotHandler(lessonSlotUseCase)
				scheduleGroup.GET("/slots", slotHandlerInstance.List)
				scheduleGroup.POST("/slots", slotHandlerInstance.Create)
				scheduleGroup.PUT("/slots/:id", slotHandlerInstance.Update)
				scheduleGroup.DELETE("/slots/:id", slotHandlerInstance.Delete)

				// Planned teaching load — issue #139, Slice 2.
				loadHandlerInstance := scheduleHandler.NewTeachingLoadHandler(teachingLoadUseCase)
				scheduleGroup.GET("/teaching-load", loadHandlerInstance.List)
				scheduleGroup.POST("/teaching-load", loadHandlerInstance.Create)
				scheduleGroup.PUT("/teaching-load/:id", loadHandlerInstance.Update)
				scheduleGroup.DELETE("/teaching-load/:id", loadHandlerInstance.Delete)

				// Automatic schedule generation (issue #139): preview a draft, then apply it.
				generateHandlerInstance := scheduleHandler.NewGenerateScheduleHandler(generateScheduleUseCase)
				scheduleGroup.POST("/generate", generateHandlerInstance.Preview)
				scheduleGroup.POST("/generate/apply", generateHandlerInstance.Apply)

				scheduleGroup.OPTIONS("/lessons", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				scheduleGroup.OPTIONS("/lessons/timetable", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				scheduleGroup.OPTIONS("/lessons/:id", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				scheduleGroup.OPTIONS("/changes", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				scheduleGroup.OPTIONS("/slots", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				scheduleGroup.OPTIONS("/slots/:id", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				scheduleGroup.OPTIONS("/teaching-load", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				scheduleGroup.OPTIONS("/teaching-load/:id", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				scheduleGroup.OPTIONS("/generate", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				scheduleGroup.OPTIONS("/generate/apply", func(c *gin.Context) { c.Status(http.StatusNoContent) })
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
			// v0.163.0 ADR-1 (#303 TIER 0): mutation routes wrapped в
			// RequireNonStudent; reads stay on the parent group.
			announcementRoutes.RegisterAnnouncementRoutes(
				announcementsGroup,
				authMiddleware.RequireNonStudent(),
				announcementHandlerInstance,
			)
			{
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
			analyticsHandlerInstance := analyticsHandler.NewAnalyticsHandler(analyticsUseCase, teacherScopeRepo)
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

			// Users management routes — split into read-only and admin-write
			// subgroups by usersRoutes.RegisterUserRoutes. Closes the TIER 0
			// privilege-escalation gap (any authenticated user could
			// PUT /role, PUT /status, DELETE /:id, POST /bulk/* prior to
			// v0.133.0). Read-only endpoints (List, GetByID, by-department,
			// by-position, GET avatar) stay permissive for cross-module
			// consumers; write endpoints require system_admin.
			usersGroup := protectedGroup.Group("/users")
			usersAdminMW := authMiddleware.RequireRole(string(authDomain.RoleSystemAdmin))
			usersRoutes.RegisterUserRoutes(usersGroup, usersAdminMW, userHandlerInstance, avatarHandlerInstance)
			{
				// CORS preflight handlers — must not be admin-gated; the
				// browser preflight has no credentials/role context.
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

			// Departments routes — RegisterDepartmentRoutes gates
			// writes (POST/PUT/DELETE) behind usersAdminMW
			// (RequireRole(system_admin)) to close #283 ADR-2 TIER 0
			// (pre-fix, v0.133.0 admin-gate split skipped this group).
			// Read endpoints stay permissive for cross-module
			// resolvers and frontend dropdowns.
			departmentsGroup := protectedGroup.Group("/departments")
			usersRoutes.RegisterDepartmentRoutes(departmentsGroup, usersAdminMW, departmentHandlerInstance)
			// CORS preflight handlers — registered on the parent
			// group so they sit outside the admin gate.
			departmentsGroup.OPTIONS("", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			departmentsGroup.OPTIONS("/:id", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			departmentsGroup.OPTIONS("/:id/children", func(c *gin.Context) { c.Status(http.StatusNoContent) })

			// Positions routes — same shape as departments.
			positionsGroup := protectedGroup.Group("/positions")
			usersRoutes.RegisterPositionRoutes(positionsGroup, usersAdminMW, positionHandlerInstance)
			positionsGroup.OPTIONS("", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			positionsGroup.OPTIONS("/:id", func(c *gin.Context) { c.Status(http.StatusNoContent) })

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

				// Cleanup route — gated to system_admin only. Closes #290
				// ADR-5: previously was security-by-comment ("Доступ
				// должен быть ограничен администраторам" в handler) but
				// route had no middleware gate, so any authenticated
				// user could trigger MinIO mass-delete + DB cleanup.
				filesCleanupGroup := filesGroup.Group("")
				filesCleanupGroup.Use(authMiddleware.RequireRole(string(authDomain.RoleSystemAdmin)))
				filesCleanupGroup.POST("/cleanup", fileHandlerInstance.CleanupExpired)

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
			// Admin notification routes (create and broadcast notifications)
			if notificationUseCase != nil {
				notificationHandler := notifHttp.NewNotificationHandler(notificationUseCase)
				adminGroup.POST("/notifications", notificationHandler.Create)
				adminGroup.POST("/notifications/bulk", notificationHandler.CreateBulk)
				adminGroup.OPTIONS("/notifications", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				adminGroup.OPTIONS("/notifications/bulk", func(c *gin.Context) { c.Status(http.StatusNoContent) })
				logger.Info("Admin notification routes registered", nil)
			}

			// Admin audit-log read API. Both writer and reader wrap
			// the same *sql.DB pool, so freshly written events are
			// immediately visible to the read API — no replication
			// lag or coherency concern.
			adminAuditLogReader := logging.NewAuditLogRepositoryPG(db)
			adminAuditLogUseCase := adminAuditLog.NewAdminAuditLogUseCase(adminAuditLogReader)
			adminAuditLogHandler := adminAuditLog.NewAdminAuditLogHandler(adminAuditLogUseCase)
			adminGroup.GET("/audit-logs", adminAuditLogHandler.List)
			adminGroup.OPTIONS("/audit-logs", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			logger.Info("Admin audit-log read API registered", nil)

			// Admin backup observability (v0.132.0). Read-only surface
			// over the /backup sidecar's shared volumes — backend MUST
			// have backup_data and backup_metrics mounted RO. The
			// auditLogger satisfies AuditSink structurally; downloads
			// emit `backup.downloaded` with the actor user id.
			backupFileReader, err := adminBackups.NewFileReader(cfg.Backup.FilesDir)
			if err != nil {
				logger.Warn("Admin backups: file reader disabled", map[string]interface{}{"error": err.Error()})
			} else {
				backupMetricsReader, err := adminBackups.NewMetricsReader(cfg.Backup.MetricsDir + "/backup_metrics.prom")
				if err != nil {
					logger.Warn("Admin backups: metrics reader disabled", map[string]interface{}{"error": err.Error()})
				} else {
					adminBackupUseCase := adminBackups.
						NewAdminBackupUseCase(backupFileReader, backupMetricsReader, cfg.Backup.FilesDir).
						WithAuditSink(auditLogger)
					adminBackupHandler := adminBackups.NewAdminBackupHandler(adminBackupUseCase)
					adminGroup.GET("/backups", adminBackupHandler.List)
					adminGroup.GET("/backups/:type/:name/download", adminBackupHandler.Download)
					adminGroup.OPTIONS("/backups", func(c *gin.Context) { c.Status(http.StatusNoContent) })
					adminGroup.OPTIONS("/backups/:type/:name/download", func(c *gin.Context) { c.Status(http.StatusNoContent) })
					logger.Info("Admin backup observability registered", nil)
				}
			}

			// Admin Sentry config observability (v0.133.0). Read-only
			// view of the runtime Sentry integration: DSN configured
			// (boolean only — never the value), environment, release,
			// traces sample rate, tracing enabled. Mirrors initSentry's
			// configuration so admins can confirm error tracking is
			// wired without reading server logs.
			adminSentryUseCase := adminSentry.NewAdminSentryUseCase(
				adminSentry.EnvDSNProbe,
				cfg.Environment,
				cfg.Version,
			)
			adminSentryHandler := adminSentry.NewAdminSentryHandler(adminSentryUseCase)
			adminGroup.GET("/sentry/config", adminSentryHandler.GetConfig)
			adminGroup.OPTIONS("/sentry/config", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			logger.Info("Admin Sentry config view registered", nil)

			// Admin integrations config view (v0.134.0). Read-only
			// projection of the runtime WebPush (VAPID) + n8n
			// configuration. VAPID private key never exposed —
			// boolean presence only; public key surfaces (browser
			// receives it via /push/public-key anyway). n8n WebhookURL
			// is non-secret operational URL.
			adminIntegrationsUseCase := adminIntegrations.NewAdminIntegrationsUseCase(
				adminIntegrations.EnvVAPIDProbe,
				cfg.WebPush.VAPIDPublicKey,
				cfg.WebPush.VAPIDSubject,
				cfg.N8N.Enabled,
				cfg.N8N.WebhookURL,
			)
			adminIntegrationsHandler := adminIntegrations.NewAdminIntegrationsHandler(adminIntegrationsUseCase)
			adminGroup.GET("/integrations/config", adminIntegrationsHandler.GetConfig)
			adminGroup.OPTIONS("/integrations/config", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			logger.Info("Admin integrations config view registered", nil)

			// Admin Composio config view (v0.135.0). Read-only
			// projection of the runtime Composio integration state
			// (API key + entity ID + MCP config ID). Only per-field
			// booleans surface — the API key is a signing secret;
			// entity ID and MCP config ID are opaque platform
			// identifiers that do not serve admin observability
			// beyond presence. Env-direct probe so the projection
			// matches what consuming services in
			// modules/notifications/application/services see.
			adminComposioUseCase := adminComposio.NewAdminComposioUseCase(
				adminComposio.EnvComposioProbe,
			)
			adminComposioHandler := adminComposio.NewAdminComposioHandler(adminComposioUseCase)
			adminGroup.GET("/composio/config", adminComposioHandler.GetConfig)
			adminGroup.OPTIONS("/composio/config", func(c *gin.Context) { c.Status(http.StatusNoContent) })
			logger.Info("Admin Composio config view registered", nil)

			// Branding module routes (v0.137.1 — ADR-7 deviation
			// closure). RegisterBrandingRoutes owns both admin
			// (GET/PUT/OPTIONS /branding) and public (GET/OPTIONS
			// /branding) mounts. Domain invariants live в
			// modules/branding/domain/entities/brand_settings.go
			// (hex color + URL scheme whitelist + length bounds);
			// validation errors map к 422 in the handler. The
			// brandingPublicGroup was assembled earlier (with
			// optional rate-limit middleware) so that the public
			// surface stays reachable even on a sharing-disabled
			// deployment.
			brandingRoutes.RegisterBrandingRoutes(
				adminGroup,
				brandingPublicGroup,
				adminBrandingHandler,
				publicBrandingHandler,
			)
			logger.Info("Branding module routes registered", nil)
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

// assignmentsGradeNotifier adapts the platform NotificationUseCase to
// the assignments module's narrow SaveGradeNotifier port. It lives in
// main.go (the DI seam) so the assignments domain/usecase packages
// stay free of cross-module Go imports.
type assignmentsGradeNotifier struct {
	notif *notifUsecases.NotificationUseCase
}

// NotifyGraded creates a "task"-typed notification for the student.
// Returns nil silently when the notification subsystem is disabled
// (notif == nil) so wiring without notifications still grades cleanly.
func (n *assignmentsGradeNotifier) NotifyGraded(ctx context.Context, studentID, assignmentID int64, score, maxScore int) error {
	if n == nil || n.notif == nil {
		return nil
	}
	_, err := n.notif.Create(ctx, &notifDTO.CreateNotificationInput{
		UserID:   studentID,
		Type:     notifEntities.NotificationTypeTask,
		Priority: notifEntities.PriorityNormal,
		Title:    "Получена оценка",
		Message:  fmt.Sprintf("За задание выставлена оценка: %d из %d", score, maxScore),
		Link:     fmt.Sprintf("/assignments/%d", assignmentID),
	})
	return err
}

// assignmentsReturnNotifier adapts the platform NotificationUseCase to
// the assignments module's narrow ReturnSubmissionNotifier port. Same
// DI-seam pattern as assignmentsGradeNotifier — the assignments module
// itself stays free of cross-module Go imports.
type assignmentsReturnNotifier struct {
	notif *notifUsecases.NotificationUseCase
}

// NotifyReturned creates a "task"-typed notification telling the
// student that their submission was returned for revision. Returns nil
// silently when the notification subsystem is disabled (notif == nil).
// The reason text is intentionally truncated in the message body to
// fit a notification card; the full reason lives on the submission row
// so the frontend can render it on /assignments/:id.
func (n *assignmentsReturnNotifier) NotifyReturned(ctx context.Context, studentID, assignmentID int64, reason string) error {
	if n == nil || n.notif == nil {
		return nil
	}
	_, err := n.notif.Create(ctx, &notifDTO.CreateNotificationInput{
		UserID:   studentID,
		Type:     notifEntities.NotificationTypeTask,
		Priority: notifEntities.PriorityNormal,
		Title:    "Работа возвращена на доработку",
		Message:  fmt.Sprintf("Преподаватель просит исправить работу: %s", truncateForNotification(reason, 200)),
		Link:     fmt.Sprintf("/assignments/%d", assignmentID),
	})
	return err
}

// assignmentsResubmitNotifier adapts the platform NotificationUseCase to
// the assignments module's narrow ResubmitSubmissionNotifier port. Same
// DI-seam pattern as assignmentsGradeNotifier / assignmentsReturnNotifier
// — the assignments module itself stays free of cross-module Go imports.
type assignmentsResubmitNotifier struct {
	notif *notifUsecases.NotificationUseCase
}

// NotifyResubmitted creates a "task"-typed notification telling the
// teacher (assignment author) that their student has resubmitted the
// returned work and a fresh grading attempt awaits. Returns nil silently
// when the notification subsystem is disabled (notif == nil) so wiring
// without notifications still resubmits cleanly.
func (n *assignmentsResubmitNotifier) NotifyResubmitted(ctx context.Context, teacherID, assignmentID int64) error {
	if n == nil || n.notif == nil {
		return nil
	}
	_, err := n.notif.Create(ctx, &notifDTO.CreateNotificationInput{
		UserID:   teacherID,
		Type:     notifEntities.NotificationTypeTask,
		Priority: notifEntities.PriorityNormal,
		Title:    "Студент пересдал работу",
		Message:  "Студент отправил исправленную работу — её нужно проверить заново.",
		Link:     fmt.Sprintf("/assignments/%d", assignmentID),
	})
	return err
}

// truncateForNotification trims a string to maxLen runes and appends
// an ellipsis when truncation occurs. Inline because it's used only by
// the return-notifier adapter; promote to a shared helper if a third
// caller appears.
func truncateForNotification(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "…"
}

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
				errorKey: err.Error(),
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
					errorKey: err.Error(),
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
				errorKey: err.Error(),
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
					errorKey: err.Error(),
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

// taskDispatchLookup adapts the existing tasks TaskRepository to the
// narrow TaskLookup port consumed by TaskReminderScheduler. Lives in
// main.go (the DI seam) so the notifications module stays free of
// cross-module Go imports back into tasks.
type taskDispatchLookup struct {
	repo taskUsecases.TaskRepository
}

// GetByID loads the task and projects к the read-only DispatchView.
// Returns nil view (no error) when the task is absent so the
// scheduler can skip gracefully without retrying.
func (a *taskDispatchLookup) GetByID(ctx context.Context, id int64) (*notifScheduler.TaskDispatchView, error) {
	if a == nil || a.repo == nil {
		return nil, nil
	}
	task, err := a.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, nil
	}
	return &notifScheduler.TaskDispatchView{
		Title:   task.Title,
		DueDate: task.DueDate,
	}, nil
}

// userEmailFromDB resolves a user's email by id via the shared
// *sql.DB. Mirror к the existing ReminderScheduler.sendEmailReminder
// pattern (also queries users table directly) but expressed as a
// narrow interface so the scheduler tests can substitute a stub.
type userEmailFromDB struct {
	db *sql.DB
}

// GetEmailByID returns the user's email address or an empty string
// when not found. Errors propagate for the scheduler к log.
func (a *userEmailFromDB) GetEmailByID(ctx context.Context, userID int64) (string, error) {
	if a == nil || a.db == nil {
		return "", nil
	}
	var email string
	err := a.db.QueryRowContext(ctx, "SELECT email FROM users WHERE id = $1", userID).Scan(&email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return email, nil
}

// initTaskReminderScheduler builds the TaskReminderScheduler с
// the supplied deps + logs failure paths. Extracted из main() to
// keep its cyclomatic complexity under the golangci threshold (the
// v0.138.0 Pair 5 wiring would otherwise push main() over the
// limit). Returns nil on construction failure so the caller can
// skip the Stop() path в graceful shutdown.
func initTaskReminderScheduler(
	logger *logging.Logger,
	taskReminderRepo taskUsecases.TaskReminderRepository,
	taskRepo taskUsecases.TaskRepository,
	db *sql.DB,
	telegramRepo notifUsecases.TelegramRepository,
	telegramService emailDomain.TelegramService,
	notificationRepo notifUsecases.NotificationRepository,
	preferencesRepo notifUsecases.PreferencesRepository,
	notifEmailService emailDomain.EmailService,
	webpushRepo notifUsecases.WebPushRepository,
	webpushService emailDomain.WebPushService,
) *notifScheduler.TaskReminderScheduler {
	scheduler, err := notifScheduler.NewTaskReminderScheduler(
		taskReminderRepo,
		&taskDispatchLookup{repo: taskRepo},
		telegramRepo,
		telegramService,
		notificationRepo,
		preferencesRepo,
		notifEmailService,
		&userEmailFromDB{db: db},
		nil,
	)
	if err != nil {
		logger.Error("Failed to initialize task reminder scheduler", map[string]interface{}{
			errorKey: err.Error(),
		})
		return nil
	}
	// v0.147.0 — wire WebPush dispatch so the push switch-case lights up
	// the real WebPushService path instead of the silent in-app fallback
	// (issue #226). Scheduler's own gates handle un-configured runtime.
	if webpushService != nil {
		scheduler.WithWebPushDispatch(webpushRepo, webpushService)
	}
	if err := scheduler.Start(); err != nil {
		logger.Error("Failed to start task reminder scheduler", map[string]interface{}{
			errorKey: err.Error(),
		})
		return scheduler
	}
	logger.Info("Task reminder scheduler started", nil)
	return scheduler
}

// initTaskReminderModule composes the v0.138.0 task reminder
// pipeline: PG repo + 3 use cases (Set / List / Delete) + handler.
// Extracted из main() to keep its cyclomatic complexity under the
// golangci threshold (Pair 5 wiring would push main() over the
// limit otherwise).
func initTaskReminderModule(
	db *sql.DB,
	auditLogger *logging.AuditLogger,
) (taskUsecases.TaskReminderRepository, *taskHandler.TaskReminderHandler) {
	repo := taskPersistence.NewTaskReminderRepositoryPG(db)
	setUC := taskUsecases.NewSetReminderUseCase(repo, taskUsecases.SystemClock{}, auditLogger)
	listUC := taskUsecases.NewListTaskRemindersUseCase(repo)
	delUC := taskUsecases.NewDeleteReminderUseCase(repo, auditLogger)
	return repo, taskHandler.NewTaskReminderHandler(setUC, listUC, delUC)
}

// workflowDocRepoAdapter wraps DocumentRepositoryPG so the v0.148.0
// workflow use cases see the sentinel ErrDocumentNotFound on lookup
// failure instead of the legacy fmt.Errorf("document not found")
// string. Adapter pattern keeps the PG repo's existing consumers
// untouched while the workflow path gets the matchable error.
//
// Issue: #227
// revisionTaskTitleFmt is the title of the РПД-revision task delegated to
// a discipline's teacher when a приказ Минобрнауки marks the program for
// revision (ADR-11). Kept at the DI seam, not in the use case, so
// user-facing wording stays out of the application layer.
const revisionTaskTitleFmt = "Актуализировать РПД по приказу Минобрнауки %s"

// minobrnaukiRevisionTaskDelegator adapts the tasks module to the
// work_program RevisionTaskDelegator port (ADR-11 delegation). Keeping the
// cross-module wiring here at the DI seam avoids a direct import between
// the work_program and tasks modules.
type minobrnaukiRevisionTaskDelegator struct {
	tasks *taskUsecases.TaskUseCase
}

func (d minobrnaukiRevisionTaskDelegator) DelegateRevision(ctx context.Context, in wpUsecases.RevisionDelegation) error {
	assignee := in.TeacherID
	_, err := d.tasks.Create(ctx, in.CreatorID, taskDto.CreateTaskInput{
		Title:      fmt.Sprintf(revisionTaskTitleFmt, in.OrderNumber),
		AssigneeID: &assignee,
		Metadata: map[string]any{
			"work_program_id":      in.WorkProgramID,
			"minobrnauki_order_id": in.MinobrnaukiOrderID,
		},
	})
	return err
}

// orderDocumentTextAdapter bridges work_program's OrderDocumentTextProvider
// port (slice 7) to the documents + text-extraction infrastructure, reusing
// the AI module's DocumentAdapter (download from S3 → extract PDF/DOCX text →
// cache). Cross-module wiring kept at the composition root. It exposes only
// the extracted text; the document title the adapter also returns is unused
// by bulk-revision.
type orderDocumentTextAdapter struct {
	docs *aiAdapters.DocumentAdapter
}

func (a orderDocumentTextAdapter) GetDocumentText(ctx context.Context, documentID int64) (string, error) {
	text, _, err := a.docs.GetDocumentContent(ctx, documentID)
	return text, err
}

type workflowDocRepoAdapter struct {
	inner *docPersistence.DocumentRepositoryPG
}

func (a *workflowDocRepoAdapter) GetByID(ctx context.Context, id int64) (*docEntities.Document, error) {
	d, err := a.inner.GetByID(ctx, id)
	if err != nil {
		if strings.Contains(err.Error(), "document not found") {
			return nil, docUsecases.ErrDocumentNotFound
		}
		return nil, err
	}
	return d, nil
}

func (a *workflowDocRepoAdapter) Update(ctx context.Context, d *docEntities.Document) error {
	return a.inner.Update(ctx, d)
}

// wireEventReminderDispatch threads optional telegram + webpush
// dispatch hooks onto an event ReminderScheduler. Extracted из main()
// so adding dispatch channels (v0.147.0 push) does not push main()
// over the gocyclo threshold. Nil deps are no-op — scheduler's own
// gates handle un-configured runtime.
//
// Issue: #226
func wireEventReminderDispatch(
	scheduler *notifScheduler.ReminderScheduler,
	telegramRepo notifUsecases.TelegramRepository,
	telegramService emailDomain.TelegramService,
	webpushRepo notifUsecases.WebPushRepository,
	webpushService emailDomain.WebPushService,
) {
	if telegramService != nil {
		scheduler.WithTelegramDispatch(telegramRepo, telegramService)
	}
	if webpushService != nil {
		scheduler.WithWebPushDispatch(webpushRepo, webpushService)
	}
}

// stopTaskReminderScheduler is a small helper that wraps the nil-
// safe Stop() call + error log path. Extracted из main() so the
// shutdown branch keeps cyclomatic complexity in check.
func stopTaskReminderScheduler(scheduler *notifScheduler.TaskReminderScheduler, logger *logging.Logger) {
	if scheduler == nil {
		return
	}
	if err := scheduler.Stop(); err != nil {
		logger.Error("Failed to stop task reminder scheduler", map[string]interface{}{
			errorKey: err.Error(),
		})
	}
}

// aiSchedulerStopper is the narrow interface satisfied by both AI schedulers
// (FactScheduler + IndexingScheduler). Lets stopAIScheduler accept either
// concrete type without reflection or per-type branches.
type aiSchedulerStopper interface {
	Stop() error
}

// stopAIScheduler wraps the nil-safe Stop() call + error log used by both
// AI schedulers (fact + indexing). kind is a free-form label used in the
// log message for forensics. Issue #263 ADR-4.
func stopAIScheduler(s aiSchedulerStopper, kind string, logger *logging.Logger) {
	if s == nil {
		return
	}
	if err := s.Stop(); err != nil {
		logger.Error("Failed to stop AI scheduler", map[string]interface{}{
			"kind":   kind,
			errorKey: err.Error(),
		})
	} else {
		logger.Info("AI scheduler stopped", map[string]interface{}{"kind": kind})
	}
}

// mountPerUserRateLimit attaches a Redis-backed per-user rate limiter to
// the given group when redisCache is wired. The bucket key is the
// authenticated user_id ctx value (set by JWT middleware), so NAT'd
// students do not share a bucket — important for dollar-cost endpoints
// like /api/ai/chat where outbound LLM-provider throttling alone is
// insufficient. No-op when Redis is unavailable (graceful degradation).
// Must be mounted AFTER JWTMiddleware so user_id ctx value is set.
// Issue #263 ADR-3.
func mountPerUserRateLimit(group *gin.RouterGroup, redisCache *cache.RedisCache) {
	if redisCache == nil {
		return
	}
	limiter := middleware.LoadRateLimitConfig().GetAuthRateLimiter(redisCache.Client())
	if limiter == nil {
		return
	}
	group.Use(limiter.RateLimitByUserMiddleware())
}

// documentsShareNotifier adapts the platform NotificationUseCase к
// the documents module's narrow ShareNotifier port. Same DI-seam
// pattern as assignmentsGradeNotifier / assignmentsReturnNotifier — the
// documents module itself stays free of cross-module Go imports.
//
// v0.156.0 ADR-5 (#266): closes the cross-module concrete import +
// fire-and-forget goroutine с context.Background() + Russian UI strings
// previously inlined в sharing_usecase.go.
type documentsShareNotifier struct {
	notif *notifUsecases.NotificationUseCase
}

// NotifyDocumentShared creates a "info"-typed notification for the
// recipient telling them a document was shared. Returns nil silently
// when the notification subsystem is disabled (notif == nil) so sharing
// still works without notifications.
func (n *documentsShareNotifier) NotifyDocumentShared(ctx context.Context, recipientID int64, documentID int64, documentTitle string) error {
	if n == nil || n.notif == nil {
		return nil
	}
	link := fmt.Sprintf("/documents/%d", documentID)
	_, err := n.notif.Create(ctx, &notifDTO.CreateNotificationInput{
		UserID:   recipientID,
		Type:     notifEntities.NotificationTypeInfo,
		Priority: notifEntities.PriorityNormal,
		Title:    "Документ открыт для вас",
		Message:  fmt.Sprintf("Вам предоставлен доступ к документу «%s»", documentTitle),
		Link:     link,
	})
	return err
}

// extracurricularNotificationNotifier adapts the platform
// NotificationUseCase to the extracurricular module's narrow
// EventNotifier port. Same DI-seam pattern as other module notifiers —
// the extracurricular module itself stays free of cross-module Go imports.
//
// v0.165.0 ADR-7 closure: backend slice (v0.164.0) defined the port;
// this commit lands the production wiring slot. Audience-cohort
// resolution ("all" / "students" / "teachers" / "staff" → user IDs)
// is a follow-up — current methods log audit context and dispatch
// nothing until the cohort resolver lands. Wiring the slot now keeps
// the use case off the noop fallback so a future change is one-line.
type extracurricularNotificationNotifier struct {
	notif *notifUsecases.NotificationUseCase
}

func (n *extracurricularNotificationNotifier) NotifyEventPublished(ctx context.Context, eventID int64, title, audience string) {
	if n == nil || n.notif == nil {
		return
	}
	_ = audience
	_ = n.notif.BroadcastSystemNotification(
		ctx,
		[]int64{},
		fmt.Sprintf("Новое мероприятие: %s", title),
		fmt.Sprintf("Подробнее: /extracurricular/%d", eventID),
	)
}

func (n *extracurricularNotificationNotifier) NotifyEventCancelled(ctx context.Context, eventID int64, title, audience string) {
	if n == nil || n.notif == nil {
		return
	}
	_ = audience
	_ = n.notif.BroadcastSystemNotification(
		ctx,
		[]int64{},
		fmt.Sprintf("Мероприятие отменено: %s", title),
		fmt.Sprintf("Подробнее: /extracurricular/%d", eventID),
	)
}

func (n *extracurricularNotificationNotifier) NotifyEventUpdated(ctx context.Context, eventID int64, title, audience string) {
	if n == nil || n.notif == nil {
		return
	}
	_ = audience
	_ = n.notif.BroadcastSystemNotification(
		ctx,
		[]int64{},
		fmt.Sprintf("Мероприятие обновлено: %s", title),
		fmt.Sprintf("Подробнее: /extracurricular/%d", eventID),
	)
}
