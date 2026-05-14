package scheduler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/go-co-op/gocron/v2"

	notifRepositories "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/repositories"
	notifServices "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/services"
	tasksRepositories "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/repositories"
)

// TaskReminderScheduler processes pending task reminders. Mirror к
// the existing ReminderScheduler shape (gocron job, 1-min check
// interval, batched MarkSent) but reads from task_reminders + tasks
// JOIN instead of event_reminders + events. Greenfield в v0.138.0.
//
// Dispatch fans out by ReminderType. Telegram path uses the
// existing ComposioTelegramService — the v0.138.0 release closes
// Phase 5 #5 final, surfacing per-task deadline reminders to
// users' Telegram chats via Composio TELEGRAM_SEND_MESSAGE action.
type TaskReminderScheduler struct {
	scheduler        gocron.Scheduler
	reminderRepo     tasksRepositories.TaskReminderRepository
	taskLookup       TaskLookup
	telegramRepo     notifRepositories.TelegramRepository
	telegramService  notifServices.TelegramService
	notificationRepo notifRepositories.NotificationRepository
	preferencesRepo  notifRepositories.PreferencesRepository
	emailService     notifServices.EmailService
	userEmailLookup  UserEmailLookup
	checkInterval    time.Duration
}

// TaskLookup is the narrow port providing per-task data the
// scheduler needs for dispatch (title + due_date).
type TaskLookup interface {
	GetByID(ctx context.Context, id int64) (*TaskDispatchView, error)
}

// TaskDispatchView is the read-side projection the scheduler uses.
type TaskDispatchView struct {
	Title   string
	DueDate *time.Time
}

// UserEmailLookup is the narrow port for resolving a user's email
// address by id.
type UserEmailLookup interface {
	GetEmailByID(ctx context.Context, userID int64) (string, error)
}

// TaskReminderSchedulerConfig holds tunable parameters.
type TaskReminderSchedulerConfig struct {
	CheckInterval time.Duration
}

// DefaultTaskReminderConfig returns the production defaults.
func DefaultTaskReminderConfig() *TaskReminderSchedulerConfig {
	return &TaskReminderSchedulerConfig{CheckInterval: 1 * time.Minute}
}

// NewTaskReminderScheduler constructs the scheduler. Returns an
// error when a required dep is nil so DI failure surfaces at boot.
// Optional deps (telegramRepo / telegramService / emailService /
// userEmailLookup) may be nil — the corresponding dispatch path
// falls back к in-app notification when missing.
func NewTaskReminderScheduler(
	reminderRepo tasksRepositories.TaskReminderRepository,
	taskLookup TaskLookup,
	telegramRepo notifRepositories.TelegramRepository,
	telegramService notifServices.TelegramService,
	notificationRepo notifRepositories.NotificationRepository,
	preferencesRepo notifRepositories.PreferencesRepository,
	emailService notifServices.EmailService,
	userEmailLookup UserEmailLookup,
	config *TaskReminderSchedulerConfig,
) (*TaskReminderScheduler, error) {
	if reminderRepo == nil {
		return nil, errors.New("task_reminder_scheduler: nil TaskReminderRepository")
	}
	if taskLookup == nil {
		return nil, errors.New("task_reminder_scheduler: nil TaskLookup")
	}
	if notificationRepo == nil {
		return nil, errors.New("task_reminder_scheduler: nil NotificationRepository")
	}
	if preferencesRepo == nil {
		return nil, errors.New("task_reminder_scheduler: nil PreferencesRepository")
	}
	if config == nil {
		config = DefaultTaskReminderConfig()
	}
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, fmt.Errorf("task_reminder_scheduler: failed to build gocron: %w", err)
	}
	return &TaskReminderScheduler{
		scheduler:        s,
		reminderRepo:     reminderRepo,
		taskLookup:       taskLookup,
		telegramRepo:     telegramRepo,
		telegramService:  telegramService,
		notificationRepo: notificationRepo,
		preferencesRepo:  preferencesRepo,
		emailService:     emailService,
		userEmailLookup:  userEmailLookup,
		checkInterval:    config.CheckInterval,
	}, nil
}

// Start mounts the periodic processPendingReminders job and starts
// the underlying gocron scheduler. Stop must be called by main.go
// on shutdown.
func (s *TaskReminderScheduler) Start() error {
	_, err := s.scheduler.NewJob(
		gocron.DurationJob(s.checkInterval),
		gocron.NewTask(s.processPendingReminders),
		gocron.WithName("process_pending_task_reminders"),
	)
	if err != nil {
		return fmt.Errorf("task_reminder_scheduler: register job: %w", err)
	}
	s.scheduler.Start()
	log.Println("Task reminder scheduler started")
	return nil
}

// Stop halts the gocron loop.
func (s *TaskReminderScheduler) Stop() error {
	if err := s.scheduler.Shutdown(); err != nil {
		return fmt.Errorf("task_reminder_scheduler: shutdown: %w", err)
	}
	log.Println("Task reminder scheduler stopped")
	return nil
}

// ProcessOnce runs the dispatch pass synchronously. Exported for
// tests so a deterministic single-tick exercise replaces gocron's
// implicit timing.
//
// Stub for RED — GREEN replaces the body with the real dispatch
// loop.
func (s *TaskReminderScheduler) ProcessOnce(ctx context.Context, now time.Time) {
	_ = ctx
	_ = now
	log.Println("task_reminder_scheduler: ProcessOnce stub — GREEN replaces this")
}

// processPendingReminders is the gocron callback. Stub.
func (s *TaskReminderScheduler) processPendingReminders() {
	log.Println("task_reminder_scheduler: processPendingReminders stub")
}
