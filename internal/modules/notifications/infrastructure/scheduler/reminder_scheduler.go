// Package scheduler contains background job scheduling for notifications.
package scheduler

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/go-co-op/gocron/v2"

	notifEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/services"
	scheduleEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
	scheduleRepos "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/repositories"
)

// ReminderScheduler processes pending event reminders
type ReminderScheduler struct {
	scheduler        gocron.Scheduler
	db               *sql.DB
	reminderRepo     scheduleRepos.EventReminderRepository
	eventRepo        scheduleRepos.EventRepository
	notificationRepo repositories.NotificationRepository
	preferencesRepo  repositories.PreferencesRepository
	emailService     services.EmailService
	checkInterval    time.Duration
	batchSize        int
}

// ReminderSchedulerConfig contains configuration for the scheduler
type ReminderSchedulerConfig struct {
	CheckInterval time.Duration
	BatchSize     int
}

// DefaultConfig returns default scheduler configuration
func DefaultConfig() *ReminderSchedulerConfig {
	return &ReminderSchedulerConfig{
		CheckInterval: 1 * time.Minute,
		BatchSize:     100,
	}
}

// NewReminderScheduler creates a new reminder scheduler
func NewReminderScheduler(
	db *sql.DB,
	reminderRepo scheduleRepos.EventReminderRepository,
	eventRepo scheduleRepos.EventRepository,
	notificationRepo repositories.NotificationRepository,
	preferencesRepo repositories.PreferencesRepository,
	emailService services.EmailService,
	config *ReminderSchedulerConfig,
) (*ReminderScheduler, error) {
	if config == nil {
		config = DefaultConfig()
	}

	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduler: %w", err)
	}

	return &ReminderScheduler{
		scheduler:        scheduler,
		db:               db,
		reminderRepo:     reminderRepo,
		eventRepo:        eventRepo,
		notificationRepo: notificationRepo,
		preferencesRepo:  preferencesRepo,
		emailService:     emailService,
		checkInterval:    config.CheckInterval,
		batchSize:        config.BatchSize,
	}, nil
}

// Start starts the reminder scheduler
func (s *ReminderScheduler) Start() error {
	// Job for processing pending reminders
	_, err := s.scheduler.NewJob(
		gocron.DurationJob(s.checkInterval),
		gocron.NewTask(s.processPendingReminders),
		gocron.WithName("process_pending_reminders"),
	)
	if err != nil {
		return fmt.Errorf("failed to create reminder job: %w", err)
	}

	// Job for cleaning up expired notifications (run daily at 3 AM)
	_, err = s.scheduler.NewJob(
		gocron.CronJob("0 3 * * *", false),
		gocron.NewTask(s.cleanupExpiredNotifications),
		gocron.WithName("cleanup_expired_notifications"),
	)
	if err != nil {
		return fmt.Errorf("failed to create cleanup job: %w", err)
	}

	s.scheduler.Start()
	log.Println("Reminder scheduler started")

	return nil
}

// Stop stops the reminder scheduler
func (s *ReminderScheduler) Stop() error {
	if err := s.scheduler.Shutdown(); err != nil {
		return fmt.Errorf("failed to stop scheduler: %w", err)
	}

	log.Println("Reminder scheduler stopped")
	return nil
}

// processPendingReminders processes all pending event reminders
func (s *ReminderScheduler) processPendingReminders() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Get reminders that should be sent now
	reminders, err := s.reminderRepo.GetPendingReminders(ctx, time.Now())
	if err != nil {
		log.Printf("Error getting pending reminders: %v", err)
		return
	}

	if len(reminders) == 0 {
		return
	}

	log.Printf("Processing %d pending reminders", len(reminders))

	var processedIDs []int64

	for _, reminder := range reminders {
		if err := s.processReminder(ctx, reminder); err != nil {
			log.Printf("Error processing reminder %d: %v", reminder.ID, err)
			continue
		}

		processedIDs = append(processedIDs, reminder.ID)
	}

	// Mark processed reminders as sent
	if len(processedIDs) > 0 {
		if err := s.reminderRepo.MarkMultipleAsSent(ctx, processedIDs); err != nil {
			log.Printf("Error marking reminders as sent: %v", err)
		} else {
			log.Printf("Marked %d reminders as sent", len(processedIDs))
		}
	}
}

// processReminder processes a single reminder
func (s *ReminderScheduler) processReminder(ctx context.Context, reminder *scheduleEntities.EventReminder) error {
	// Get event details
	event, err := s.eventRepo.GetByID(ctx, reminder.EventID)
	if err != nil {
		return fmt.Errorf("failed to get event: %w", err)
	}

	if event == nil {
		return fmt.Errorf("event not found: %d", reminder.EventID)
	}

	// Get user preferences
	prefs, err := s.preferencesRepo.GetOrCreate(ctx, reminder.UserID)
	if err != nil {
		return fmt.Errorf("failed to get preferences: %w", err)
	}

	// Check quiet hours
	if prefs.IsWithinQuietHours(time.Now()) {
		log.Printf("Skipping reminder %d: user %d is in quiet hours", reminder.ID, reminder.UserID)
		return nil
	}

	// Process based on reminder type
	switch reminder.ReminderType {
	case scheduleEntities.ReminderTypeEmail:
		if prefs.EmailEnabled {
			return s.sendEmailReminder(ctx, reminder, event)
		}
	case scheduleEntities.ReminderTypeInApp:
		if prefs.InAppEnabled {
			return s.sendInAppReminder(ctx, reminder, event)
		}
	case scheduleEntities.ReminderTypePush:
		if prefs.PushEnabled {
			return s.sendPushReminder(ctx, reminder, event)
		}
	case scheduleEntities.ReminderTypeTelegram:
		if prefs.TelegramEnabled {
			return s.sendTelegramReminder(ctx, reminder, event)
		}
	default:
		// Send in-app notification as fallback
		if prefs.InAppEnabled {
			return s.sendInAppReminder(ctx, reminder, event)
		}
	}

	return nil
}

// sendEmailReminder sends an email reminder
func (s *ReminderScheduler) sendEmailReminder(ctx context.Context, reminder *scheduleEntities.EventReminder, event *scheduleEntities.Event) error {
	// Get user email from database
	var userEmail string
	err := s.db.QueryRowContext(ctx, "SELECT email FROM users WHERE id = $1", reminder.UserID).Scan(&userEmail)
	if err != nil {
		return fmt.Errorf("failed to get user email: %w", err)
	}

	subject := fmt.Sprintf("Напоминание: %s", event.Title)
	body := s.formatEventReminderEmail(event, reminder.MinutesBefore)

	return s.emailService.SendNotification(ctx, userEmail, subject, body)
}

// sendInAppReminder creates an in-app notification
func (s *ReminderScheduler) sendInAppReminder(ctx context.Context, reminder *scheduleEntities.EventReminder, event *scheduleEntities.Event) error {
	now := time.Now()

	notification := &notifEntities.Notification{
		UserID:    reminder.UserID,
		Type:      notifEntities.NotificationTypeReminder,
		Priority:  notifEntities.PriorityHigh,
		Title:     "Напоминание о событии",
		Message:   s.formatEventReminderMessage(event, reminder.MinutesBefore),
		Link:      fmt.Sprintf("/schedule/events/%d", event.ID),
		IsRead:    false,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return s.notificationRepo.Create(ctx, notification)
}

// sendPushReminder sends a push notification.
// Currently falls back to in-app notification; Firebase FCM integration planned for future.
func (s *ReminderScheduler) sendPushReminder(ctx context.Context, reminder *scheduleEntities.EventReminder, event *scheduleEntities.Event) error {
	return s.sendInAppReminder(ctx, reminder, event)
}

// sendTelegramReminder sends a Telegram notification.
// Currently falls back to in-app notification; Telegram bot integration planned for future.
func (s *ReminderScheduler) sendTelegramReminder(ctx context.Context, reminder *scheduleEntities.EventReminder, event *scheduleEntities.Event) error {
	return s.sendInAppReminder(ctx, reminder, event)
}

// formatEventReminderEmail formats the email body for event reminder
func (s *ReminderScheduler) formatEventReminderEmail(event *scheduleEntities.Event, minutesBefore int) string {
	timeText := s.formatMinutesText(minutesBefore)

	return fmt.Sprintf(`
		<html>
			<body>
				<h2>Напоминание о событии</h2>
				<p>Через %s начнётся событие:</p>
				<h3>%s</h3>
				<p><strong>Время:</strong> %s</p>
				%s
				%s
				<br>
				<p>С уважением,<br>Система Секретарь-Методист</p>
			</body>
		</html>
	`,
		timeText,
		event.Title,
		event.StartTime.Format("02.01.2006 15:04"),
		s.formatOptional("Место", event.Location),
		s.formatOptional("Описание", event.Description),
	)
}

// formatEventReminderMessage formats the message for in-app notification
func (s *ReminderScheduler) formatEventReminderMessage(event *scheduleEntities.Event, minutesBefore int) string {
	timeText := s.formatMinutesText(minutesBefore)
	return fmt.Sprintf("Через %s: %s (%s)", timeText, event.Title, event.StartTime.Format("15:04"))
}

// formatMinutesText formats minutes before as human-readable text
func (s *ReminderScheduler) formatMinutesText(minutes int) string {
	switch {
	case minutes < 60:
		return fmt.Sprintf("%d минут", minutes)
	case minutes == 60:
		return "1 час"
	case minutes < 1440:
		hours := minutes / 60
		return fmt.Sprintf("%d %s", hours, s.pluralHours(hours))
	default:
		days := minutes / 1440
		return fmt.Sprintf("%d %s", days, s.pluralDays(days))
	}
}

func (s *ReminderScheduler) pluralHours(n int) string {
	if n%10 == 1 && n%100 != 11 {
		return "час"
	}
	if n%10 >= 2 && n%10 <= 4 && (n%100 < 10 || n%100 >= 20) {
		return "часа"
	}
	return "часов"
}

func (s *ReminderScheduler) pluralDays(n int) string {
	if n%10 == 1 && n%100 != 11 {
		return "день"
	}
	if n%10 >= 2 && n%10 <= 4 && (n%100 < 10 || n%100 >= 20) {
		return "дня"
	}
	return "дней"
}

func (s *ReminderScheduler) formatOptional(label string, value *string) string {
	if value == nil || *value == "" {
		return ""
	}
	return fmt.Sprintf("<p><strong>%s:</strong> %s</p>", label, *value)
}

// cleanupExpiredNotifications removes expired notifications
func (s *ReminderScheduler) cleanupExpiredNotifications() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	count, err := s.notificationRepo.DeleteExpired(ctx)
	if err != nil {
		log.Printf("Error cleaning up expired notifications: %v", err)
		return
	}

	if count > 0 {
		log.Printf("Cleaned up %d expired notifications", count)
	}
}
