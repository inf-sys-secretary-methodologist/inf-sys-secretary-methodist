package scheduler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/go-co-op/gocron/v2"

	notifEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/entities"
	notifRepositories "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/repositories"
	notifServices "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/services"
	tasksEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/entities"
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
//
// taskLookup is the cross-module narrow port — only DueDate +
// Title are read so the dispatch can compose a sensible reminder
// message. Defined as an interface here so unit tests can substitute
// a fake без dragging the full TaskRepository surface.
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
// scheduler needs for dispatch (title + due_date). Mirror к
// scheduleRepos.EventRepository.GetByID shape but typed для tasks.
type TaskLookup interface {
	GetByID(ctx context.Context, id int64) (*TaskDispatchView, error)
}

// TaskDispatchView is the read-side projection the scheduler uses.
// Kept narrow на purpose — domain code surfaces что dispatch needs,
// no more.
type TaskDispatchView struct {
	Title   string
	DueDate *time.Time
}

// UserEmailLookup is the narrow port for resolving a user's email
// address by id. The existing ReminderScheduler reads from `users`
// table directly via *sql.DB; for the new scheduler we depend on
// an interface so tests can stub it.
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

// NewTaskReminderScheduler constructs the scheduler. Panics on a
// nil required dep so misconfigured DI fails at boot. Optional
// deps (telegramRepo / telegramService / emailService /
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

// Stop halts the gocron loop. Idempotent — repeated calls return
// the gocron shutdown error без panic.
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
func (s *TaskReminderScheduler) ProcessOnce(ctx context.Context, now time.Time) {
	s.processPendingRemindersAt(ctx, now)
}

func (s *TaskReminderScheduler) processPendingReminders() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	s.processPendingRemindersAt(ctx, time.Now())
}

func (s *TaskReminderScheduler) processPendingRemindersAt(ctx context.Context, now time.Time) {
	reminders, err := s.reminderRepo.GetPendingReminders(ctx, now)
	if err != nil {
		log.Printf("task_reminder_scheduler: get pending: %v", err)
		return
	}
	if len(reminders) == 0 {
		return
	}
	log.Printf("task_reminder_scheduler: processing %d reminders", len(reminders))

	processedIDs := make([]int64, 0, len(reminders))
	for _, reminder := range reminders {
		if err := s.processReminder(ctx, reminder); err != nil {
			log.Printf("task_reminder_scheduler: reminder %d: %v", reminder.ID(), err)
			continue
		}
		processedIDs = append(processedIDs, reminder.ID())
	}
	if len(processedIDs) > 0 {
		if err := s.reminderRepo.MarkSentBatch(ctx, processedIDs, now); err != nil {
			log.Printf("task_reminder_scheduler: mark sent batch: %v", err)
		}
	}
}

// processReminder fans out by ReminderType. Quiet hours respected.
// Failures during dispatch are logged but the reminder still flips
// к sent (avoids retry storms на flaky Composio / SMTP).
func (s *TaskReminderScheduler) processReminder(ctx context.Context, reminder *tasksEntities.TaskReminder) error {
	view, err := s.taskLookup.GetByID(ctx, reminder.TaskID())
	if err != nil {
		return fmt.Errorf("task lookup: %w", err)
	}
	if view == nil {
		return fmt.Errorf("task %d not found", reminder.TaskID())
	}
	prefs, err := s.preferencesRepo.GetOrCreate(ctx, reminder.UserID())
	if err != nil {
		return fmt.Errorf("preferences: %w", err)
	}
	if prefs.IsWithinQuietHours(time.Now()) {
		log.Printf("task_reminder_scheduler: skip reminder %d — user %d in quiet hours", reminder.ID(), reminder.UserID())
		return nil
	}
	switch reminder.ReminderType() {
	case tasksEntities.ReminderTypeTelegram:
		if prefs.TelegramEnabled {
			return s.sendTelegram(ctx, reminder, view)
		}
	case tasksEntities.ReminderTypeEmail:
		if prefs.EmailEnabled {
			return s.sendEmail(ctx, reminder, view)
		}
	case tasksEntities.ReminderTypePush:
		if prefs.PushEnabled {
			return s.sendInApp(ctx, reminder, view)
		}
	case tasksEntities.ReminderTypeInApp:
		if prefs.InAppEnabled {
			return s.sendInApp(ctx, reminder, view)
		}
	}
	// Channel disabled in preferences или unknown reminder type —
	// fall back к in-app so the reminder is not silently lost.
	return s.sendInApp(ctx, reminder, view)
}

// sendTelegram dispatches via the injected ComposioTelegramService.
// Phase 5 #5 final: the route lights up the existing wiring from
// v0.134.0 + v0.135.0 admin observability surfaces.
func (s *TaskReminderScheduler) sendTelegram(ctx context.Context, reminder *tasksEntities.TaskReminder, view *TaskDispatchView) error {
	if s.telegramRepo == nil || s.telegramService == nil {
		// Telegram dispatch was opted out at construction time.
		return s.sendInApp(ctx, reminder, view)
	}
	conn, err := s.telegramRepo.GetConnectionByUserID(ctx, reminder.UserID())
	if err != nil || conn == nil || !conn.IsActive {
		// User has no active Telegram connection — graceful
		// fallback к in-app keeps the reminder reachable.
		return s.sendInApp(ctx, reminder, view)
	}
	chatID := strconv.FormatInt(conn.TelegramChatID, 10)
	title := "Напоминание о задаче"
	message := formatTaskReminderMessage(view, reminder.MinutesBefore())
	if err := s.telegramService.SendNotification(ctx, chatID, title, message, "high"); err != nil {
		log.Printf("task_reminder_scheduler: telegram dispatch failed: %v — falling back к in-app", err)
		return s.sendInApp(ctx, reminder, view)
	}
	return nil
}

// sendEmail composes a simple HTML body matching the existing
// event reminder style and dispatches via the injected
// EmailService.
func (s *TaskReminderScheduler) sendEmail(ctx context.Context, reminder *tasksEntities.TaskReminder, view *TaskDispatchView) error {
	if s.emailService == nil || s.userEmailLookup == nil {
		return s.sendInApp(ctx, reminder, view)
	}
	email, err := s.userEmailLookup.GetEmailByID(ctx, reminder.UserID())
	if err != nil || email == "" {
		return s.sendInApp(ctx, reminder, view)
	}
	subject := "Напоминание о задаче: " + view.Title
	body := formatTaskReminderEmail(view, reminder.MinutesBefore())
	if err := s.emailService.SendNotification(ctx, email, subject, body); err != nil {
		log.Printf("task_reminder_scheduler: email dispatch failed: %v — falling back к in-app", err)
		return s.sendInApp(ctx, reminder, view)
	}
	return nil
}

// sendInApp creates an in-app notification — the dispatch
// last-resort that is always reachable.
func (s *TaskReminderScheduler) sendInApp(ctx context.Context, reminder *tasksEntities.TaskReminder, view *TaskDispatchView) error {
	now := time.Now()
	n := &notifEntities.Notification{
		UserID:    reminder.UserID(),
		Type:      notifEntities.NotificationTypeReminder,
		Priority:  notifEntities.PriorityHigh,
		Title:     "Напоминание о задаче",
		Message:   formatTaskReminderMessage(view, reminder.MinutesBefore()),
		Link:      fmt.Sprintf("/tasks/%d", reminder.TaskID()),
		IsRead:    false,
		CreatedAt: now,
		UpdatedAt: now,
	}
	return s.notificationRepo.Create(ctx, n)
}

// formatTaskReminderMessage renders the one-line message body
// reused by Telegram + in-app dispatch.
func formatTaskReminderMessage(view *TaskDispatchView, minutesBefore int) string {
	if view.DueDate == nil {
		return fmt.Sprintf("Напоминание: «%s» — крайний срок не задан", view.Title)
	}
	return fmt.Sprintf("Через %s — «%s» (дедлайн %s)",
		formatMinutesText(minutesBefore),
		view.Title,
		view.DueDate.Format("02.01.2006 15:04"),
	)
}

// formatTaskReminderEmail renders a tiny HTML body for the email
// channel — mirror к existing ReminderScheduler.formatEventReminderEmail
// shape but simplified to fit the task domain (no location).
func formatTaskReminderEmail(view *TaskDispatchView, minutesBefore int) string {
	dueLine := ""
	if view.DueDate != nil {
		dueLine = fmt.Sprintf("<p><strong>Крайний срок:</strong> %s</p>", view.DueDate.Format("02.01.2006 15:04"))
	}
	return fmt.Sprintf(`<html><body><h2>Напоминание о задаче</h2><p>Через %s наступит крайний срок задачи:</p><h3>%s</h3>%s<br><p>С уважением,<br>Система Секретарь-Методист</p></body></html>`,
		formatMinutesText(minutesBefore),
		view.Title,
		dueLine,
	)
}

// formatMinutesText converts minutes к human-readable Russian text.
// Mirror к the existing ReminderScheduler.formatMinutesText logic;
// duplicated rather than exported because plural forms are
// scheduler-internal copy.
func formatMinutesText(minutes int) string {
	switch {
	case minutes < 60:
		return fmt.Sprintf("%d минут", minutes)
	case minutes == 60:
		return "1 час"
	case minutes < 1440:
		hours := minutes / 60
		return fmt.Sprintf("%d %s", hours, pluralHours(hours))
	default:
		days := minutes / 1440
		return fmt.Sprintf("%d %s", days, pluralDays(days))
	}
}

func pluralHours(n int) string {
	if n%10 == 1 && n%100 != 11 {
		return "час"
	}
	if n%10 >= 2 && n%10 <= 4 && (n%100 < 10 || n%100 >= 20) {
		return "часа"
	}
	return "часов"
}

func pluralDays(n int) string {
	if n%10 == 1 && n%100 != 11 {
		return "день"
	}
	if n%10 >= 2 && n%10 <= 4 && (n%100 < 10 || n%100 >= 20) {
		return "дня"
	}
	return "дней"
}
