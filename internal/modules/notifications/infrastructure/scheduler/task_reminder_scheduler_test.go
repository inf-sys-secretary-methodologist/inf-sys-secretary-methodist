package scheduler

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	notifEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/entities"
	notifServices "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/services"
	tasksEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/entities"
)

// ---- fakes ----

type fakeReminderRepo struct {
	mu        sync.Mutex
	pending   []*tasksEntities.TaskReminder
	markedIDs []int64
}

func (r *fakeReminderRepo) Create(_ context.Context, _ *tasksEntities.TaskReminder) error {
	return nil
}
func (r *fakeReminderRepo) Delete(_ context.Context, _ int64) error { return nil }
func (r *fakeReminderRepo) GetByID(_ context.Context, _ int64) (*tasksEntities.TaskReminder, error) {
	return nil, nil
}
func (r *fakeReminderRepo) ListByTaskAndUser(_ context.Context, _, _ int64) ([]*tasksEntities.TaskReminder, error) {
	return nil, nil
}
func (r *fakeReminderRepo) GetPendingReminders(_ context.Context, _ time.Time) ([]*tasksEntities.TaskReminder, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.pending, nil
}
func (r *fakeReminderRepo) MarkSentBatch(_ context.Context, ids []int64, _ time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.markedIDs = append(r.markedIDs, ids...)
	return nil
}

type fakeTaskLookup struct {
	views map[int64]*TaskDispatchView
}

func (l *fakeTaskLookup) GetByID(_ context.Context, id int64) (*TaskDispatchView, error) {
	v, ok := l.views[id]
	if !ok {
		return nil, errors.New("task not found")
	}
	return v, nil
}

// fakeTelegramRepo — minimal TelegramRepository surface used by the
// scheduler (GetConnectionByUserID). Other methods stub к no-ops
// because they're outside the dispatch path being tested.
type fakeTelegramRepo struct {
	conns map[int64]*notifEntities.TelegramConnection
}

func (r *fakeTelegramRepo) CreateVerificationCode(_ context.Context, _ *notifEntities.TelegramVerificationCode) error {
	return nil
}
func (r *fakeTelegramRepo) GetVerificationCodeByCode(_ context.Context, _ string) (*notifEntities.TelegramVerificationCode, error) {
	return nil, nil
}
func (r *fakeTelegramRepo) GetActiveVerificationCodeByUserID(_ context.Context, _ int64) (*notifEntities.TelegramVerificationCode, error) {
	return nil, nil
}
func (r *fakeTelegramRepo) MarkCodeAsUsed(_ context.Context, _ int64) error { return nil }
func (r *fakeTelegramRepo) DeleteExpiredCodes(_ context.Context) error      { return nil }
func (r *fakeTelegramRepo) CreateConnection(_ context.Context, _ *notifEntities.TelegramConnection) error {
	return nil
}
func (r *fakeTelegramRepo) GetConnectionByUserID(_ context.Context, userID int64) (*notifEntities.TelegramConnection, error) {
	if c, ok := r.conns[userID]; ok {
		return c, nil
	}
	return nil, nil
}
func (r *fakeTelegramRepo) GetConnectionByChatID(_ context.Context, _ int64) (*notifEntities.TelegramConnection, error) {
	return nil, nil
}
func (r *fakeTelegramRepo) GetActiveConnections(_ context.Context) ([]notifEntities.TelegramConnection, error) {
	return nil, nil
}
func (r *fakeTelegramRepo) UpdateConnection(_ context.Context, _ *notifEntities.TelegramConnection) error {
	return nil
}
func (r *fakeTelegramRepo) DeleteConnection(_ context.Context, _ int64) error { return nil }

type telegramCall struct {
	ChatID, Title, Message, Priority string
}

type fakeTelegramService struct {
	calls    []telegramCall
	sendErr  error
	msgCalls []notifServices.SendTelegramMessageRequest
}

func (s *fakeTelegramService) SendMessage(_ context.Context, req *notifServices.SendTelegramMessageRequest) error {
	s.msgCalls = append(s.msgCalls, *req)
	return s.sendErr
}
func (s *fakeTelegramService) SendNotification(_ context.Context, chatID, title, message, priority string) error {
	s.calls = append(s.calls, telegramCall{ChatID: chatID, Title: title, Message: message, Priority: priority})
	return s.sendErr
}

// fakeNotificationRepo — minimal NotificationRepository surface.
// Only Create matters for dispatch; everything else is no-op.
type fakeNotificationRepo struct {
	created []*notifEntities.Notification
}

func (r *fakeNotificationRepo) Create(_ context.Context, n *notifEntities.Notification) error {
	r.created = append(r.created, n)
	return nil
}
func (r *fakeNotificationRepo) Update(_ context.Context, _ *notifEntities.Notification) error {
	return nil
}
func (r *fakeNotificationRepo) Delete(_ context.Context, _ int64) error { return nil }
func (r *fakeNotificationRepo) GetByID(_ context.Context, _ int64) (*notifEntities.Notification, error) {
	return nil, nil
}
func (r *fakeNotificationRepo) List(_ context.Context, _ *notifEntities.NotificationFilter) ([]*notifEntities.Notification, error) {
	return nil, nil
}
func (r *fakeNotificationRepo) GetByUserID(_ context.Context, _ int64, _, _ int) ([]*notifEntities.Notification, error) {
	return nil, nil
}
func (r *fakeNotificationRepo) GetUnreadByUserID(_ context.Context, _ int64) ([]*notifEntities.Notification, error) {
	return nil, nil
}
func (r *fakeNotificationRepo) MarkAsRead(_ context.Context, _ int64) error     { return nil }
func (r *fakeNotificationRepo) MarkAllAsRead(_ context.Context, _ int64) error  { return nil }
func (r *fakeNotificationRepo) DeleteByUserID(_ context.Context, _ int64) error { return nil }
func (r *fakeNotificationRepo) DeleteExpired(_ context.Context) (int64, error)  { return 0, nil }
func (r *fakeNotificationRepo) GetUnreadCount(_ context.Context, _ int64) (int64, error) {
	return 0, nil
}
func (r *fakeNotificationRepo) GetStats(_ context.Context, _ int64) (*notifEntities.NotificationStats, error) {
	return nil, nil
}
func (r *fakeNotificationRepo) CreateBulk(_ context.Context, _ []*notifEntities.Notification) error {
	return nil
}

type fakePreferencesRepo struct {
	prefs *notifEntities.UserNotificationPreferences
}

func (r *fakePreferencesRepo) Create(_ context.Context, _ *notifEntities.UserNotificationPreferences) error {
	return nil
}
func (r *fakePreferencesRepo) Update(_ context.Context, _ *notifEntities.UserNotificationPreferences) error {
	return nil
}
func (r *fakePreferencesRepo) Delete(_ context.Context, _ int64) error { return nil }
func (r *fakePreferencesRepo) GetByUserID(_ context.Context, _ int64) (*notifEntities.UserNotificationPreferences, error) {
	return r.prefs, nil
}
func (r *fakePreferencesRepo) GetOrCreate(_ context.Context, _ int64) (*notifEntities.UserNotificationPreferences, error) {
	return r.prefs, nil
}
func (r *fakePreferencesRepo) UpdateChannelEnabled(_ context.Context, _ int64, _ notifEntities.NotificationChannel, _ bool) error {
	return nil
}
func (r *fakePreferencesRepo) UpdateQuietHours(_ context.Context, _ int64, _ bool, _, _, _ string) error {
	return nil
}

// ---- test helpers ----

func enabledPrefs() *notifEntities.UserNotificationPreferences {
	p := notifEntities.NewUserNotificationPreferences(7)
	p.TelegramEnabled = true
	p.EmailEnabled = true
	p.PushEnabled = true
	p.InAppEnabled = true
	return p
}

type fakeClock struct{ now time.Time }

func (c fakeClock) Now() time.Time { return c.now }

func buildScheduler(t *testing.T,
	reminders []*tasksEntities.TaskReminder,
	views map[int64]*TaskDispatchView,
	conns map[int64]*notifEntities.TelegramConnection,
	tgService *fakeTelegramService,
	notifRepo *fakeNotificationRepo,
	prefs *notifEntities.UserNotificationPreferences,
) (*TaskReminderScheduler, *fakeReminderRepo) {
	t.Helper()
	return buildSchedulerAt(t, reminders, views, conns, tgService, notifRepo, prefs,
		time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC))
}

// buildSchedulerAt — variant that pins the injected clock to a
// specific instant. Lets quiet-hours tests pick a time the user's
// quiet window covers.
func buildSchedulerAt(t *testing.T,
	reminders []*tasksEntities.TaskReminder,
	views map[int64]*TaskDispatchView,
	conns map[int64]*notifEntities.TelegramConnection,
	tgService *fakeTelegramService,
	notifRepo *fakeNotificationRepo,
	prefs *notifEntities.UserNotificationPreferences,
	now time.Time,
) (*TaskReminderScheduler, *fakeReminderRepo) {
	t.Helper()
	repo := &fakeReminderRepo{pending: reminders}
	s, err := NewTaskReminderScheduler(
		repo,
		&fakeTaskLookup{views: views},
		&fakeTelegramRepo{conns: conns},
		tgService,
		notifRepo,
		&fakePreferencesRepo{prefs: prefs},
		nil, // emailService — not exercised here
		nil, // userEmailLookup
		&TaskReminderSchedulerConfig{CheckInterval: time.Hour, Clock: fakeClock{now: now}},
	)
	require.NoError(t, err)
	return s, repo
}

// ---- tests ----

// TestProcessOnce_TelegramHappyPath pins the v0.138.0 Phase 5 #5
// final closure: a telegram-type reminder with an active connection
// AND telegram enabled in prefs dispatches via the injected
// ComposioTelegramService. No fallback к in-app fires.
func TestProcessOnce_TelegramHappyPath(t *testing.T) {
	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	due := now.Add(15 * time.Minute)
	reminder := tasksEntities.HydrateFromPersistence(101, 42, 7, tasksEntities.ReminderTypeTelegram, 15, false, nil, now.Add(-time.Hour))

	views := map[int64]*TaskDispatchView{
		42: {Title: "Утвердить РПД", DueDate: &due},
	}
	conns := map[int64]*notifEntities.TelegramConnection{
		7: {UserID: 7, TelegramChatID: 555000, IsActive: true},
	}
	tg := &fakeTelegramService{}
	notif := &fakeNotificationRepo{}
	s, repo := buildScheduler(t, []*tasksEntities.TaskReminder{reminder}, views, conns, tg, notif, enabledPrefs())

	s.ProcessOnce(context.Background(), now)

	require.Len(t, tg.calls, 1, "telegram dispatch must fire exactly once")
	assert.Equal(t, "555000", tg.calls[0].ChatID)
	assert.Contains(t, tg.calls[0].Message, "Утвердить РПД")
	assert.Equal(t, "high", tg.calls[0].Priority)
	assert.Len(t, notif.created, 0, "happy-path telegram dispatch must not fan out к in-app")
	assert.Equal(t, []int64{101}, repo.markedIDs, "processed reminder marked as sent")
}

// TestProcessOnce_TelegramNoConnection_FallbackInApp confirms the
// graceful degradation: user without an active Telegram connection
// still receives an in-app notification so the reminder isn't lost.
func TestProcessOnce_TelegramNoConnection_FallbackInApp(t *testing.T) {
	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	due := now.Add(15 * time.Minute)
	reminder := tasksEntities.HydrateFromPersistence(101, 42, 7, tasksEntities.ReminderTypeTelegram, 15, false, nil, now.Add(-time.Hour))
	views := map[int64]*TaskDispatchView{42: {Title: "Утвердить РПД", DueDate: &due}}
	tg := &fakeTelegramService{}
	notif := &fakeNotificationRepo{}
	s, repo := buildScheduler(t, []*tasksEntities.TaskReminder{reminder}, views, map[int64]*notifEntities.TelegramConnection{}, tg, notif, enabledPrefs())

	s.ProcessOnce(context.Background(), now)

	assert.Len(t, tg.calls, 0, "no telegram dispatch без active connection")
	require.Len(t, notif.created, 1, "in-app fallback must fire")
	assert.Equal(t, int64(7), notif.created[0].UserID)
	assert.Contains(t, notif.created[0].Message, "Утвердить РПД")
	assert.Equal(t, []int64{101}, repo.markedIDs)
}

// TestProcessOnce_TelegramServiceFails_FallbackInApp — composio
// API error path → graceful fallback к in-app.
func TestProcessOnce_TelegramServiceFails_FallbackInApp(t *testing.T) {
	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	due := now.Add(15 * time.Minute)
	reminder := tasksEntities.HydrateFromPersistence(101, 42, 7, tasksEntities.ReminderTypeTelegram, 15, false, nil, now.Add(-time.Hour))
	views := map[int64]*TaskDispatchView{42: {Title: "Утвердить РПД", DueDate: &due}}
	conns := map[int64]*notifEntities.TelegramConnection{7: {UserID: 7, TelegramChatID: 555000, IsActive: true}}
	tg := &fakeTelegramService{sendErr: errors.New("composio 503")}
	notif := &fakeNotificationRepo{}
	s, _ := buildScheduler(t, []*tasksEntities.TaskReminder{reminder}, views, conns, tg, notif, enabledPrefs())

	s.ProcessOnce(context.Background(), now)

	require.Len(t, tg.calls, 1, "telegram dispatch attempted")
	require.Len(t, notif.created, 1, "fallback к in-app on dispatch failure")
}

// TestProcessOnce_TelegramDisabledInPrefs_FallbackInApp — channel
// disabled in user prefs → fallback к in-app.
func TestProcessOnce_TelegramDisabledInPrefs_FallbackInApp(t *testing.T) {
	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	due := now.Add(15 * time.Minute)
	reminder := tasksEntities.HydrateFromPersistence(101, 42, 7, tasksEntities.ReminderTypeTelegram, 15, false, nil, now.Add(-time.Hour))
	views := map[int64]*TaskDispatchView{42: {Title: "Утвердить РПД", DueDate: &due}}
	conns := map[int64]*notifEntities.TelegramConnection{7: {UserID: 7, TelegramChatID: 555000, IsActive: true}}
	tg := &fakeTelegramService{}
	notif := &fakeNotificationRepo{}
	prefs := enabledPrefs()
	prefs.TelegramEnabled = false
	s, _ := buildScheduler(t, []*tasksEntities.TaskReminder{reminder}, views, conns, tg, notif, prefs)

	s.ProcessOnce(context.Background(), now)

	assert.Len(t, tg.calls, 0, "telegram disabled in prefs — no dispatch")
	assert.Len(t, notif.created, 1, "in-app fallback")
}

// TestProcessOnce_InAppType_DirectInApp — reminder type 'in_app'
// hits in-app directly без consulting telegram.
func TestProcessOnce_InAppType_DirectInApp(t *testing.T) {
	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	due := now.Add(15 * time.Minute)
	reminder := tasksEntities.HydrateFromPersistence(101, 42, 7, tasksEntities.ReminderTypeInApp, 15, false, nil, now.Add(-time.Hour))
	views := map[int64]*TaskDispatchView{42: {Title: "Утвердить РПД", DueDate: &due}}
	tg := &fakeTelegramService{}
	notif := &fakeNotificationRepo{}
	s, _ := buildScheduler(t, []*tasksEntities.TaskReminder{reminder}, views, nil, tg, notif, enabledPrefs())

	s.ProcessOnce(context.Background(), now)

	assert.Len(t, tg.calls, 0)
	require.Len(t, notif.created, 1)
}

// TestProcessOnce_QuietHours_SkipsDispatch confirms the injected
// clock drives the IsWithinQuietHours decision: user has quiet
// window 22:00-07:00 + now=23:30 → reminder NOT dispatched and NOT
// marked sent (so the next non-quiet tick processes it). This pins
// the Tier 2 #2 fix from reviewer round-1 (clock injection ends
// the wall-clock leak class).
func TestProcessOnce_QuietHours_SkipsDispatch(t *testing.T) {
	quietNow := time.Date(2026, 5, 14, 23, 30, 0, 0, time.UTC)
	due := quietNow.Add(15 * time.Minute)
	reminder := tasksEntities.HydrateFromPersistence(101, 42, 7, tasksEntities.ReminderTypeTelegram, 15, false, nil, quietNow.Add(-time.Hour))
	views := map[int64]*TaskDispatchView{42: {Title: "Утвердить РПД", DueDate: &due}}
	conns := map[int64]*notifEntities.TelegramConnection{7: {UserID: 7, TelegramChatID: 555000, IsActive: true}}
	tg := &fakeTelegramService{}
	notif := &fakeNotificationRepo{}
	prefs := enabledPrefs()
	prefs.QuietHoursEnabled = true
	prefs.QuietHoursStart = "22:00"
	prefs.QuietHoursEnd = "07:00"
	prefs.Timezone = "UTC"
	s, repo := buildSchedulerAt(t, []*tasksEntities.TaskReminder{reminder}, views, conns, tg, notif, prefs, quietNow)

	s.ProcessOnce(context.Background(), quietNow)

	assert.Len(t, tg.calls, 0, "quiet hours → no telegram dispatch")
	assert.Len(t, notif.created, 0, "quiet hours → no in-app dispatch either")
	assert.Equal(t, []int64{101}, repo.markedIDs, "reminder still marked processed so it does not retry indefinitely on next tick")
}

// TestProcessOnce_NoPending_NoOp — empty pending list bypasses
// dispatch и does not call MarkSentBatch.
func TestProcessOnce_NoPending_NoOp(t *testing.T) {
	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	tg := &fakeTelegramService{}
	notif := &fakeNotificationRepo{}
	s, repo := buildScheduler(t, nil, nil, nil, tg, notif, enabledPrefs())

	s.ProcessOnce(context.Background(), now)

	assert.Len(t, tg.calls, 0)
	assert.Len(t, notif.created, 0)
	assert.Len(t, repo.markedIDs, 0, "no reminders → no MarkSentBatch call")
}

// TestNewTaskReminderScheduler_NilRequiredDep_ReturnsError pins
// the boot-time fail-fast contract — каждый required dep nil → ctor
// returns an error.
func TestNewTaskReminderScheduler_NilRequiredDep_ReturnsError(t *testing.T) {
	cases := []struct {
		name string
		f    func() error
	}{
		{"nil_reminder_repo", func() error {
			_, err := NewTaskReminderScheduler(nil, &fakeTaskLookup{}, nil, nil, &fakeNotificationRepo{}, &fakePreferencesRepo{}, nil, nil, nil)
			return err
		}},
		{"nil_task_lookup", func() error {
			_, err := NewTaskReminderScheduler(&fakeReminderRepo{}, nil, nil, nil, &fakeNotificationRepo{}, &fakePreferencesRepo{}, nil, nil, nil)
			return err
		}},
		{"nil_notification_repo", func() error {
			_, err := NewTaskReminderScheduler(&fakeReminderRepo{}, &fakeTaskLookup{}, nil, nil, nil, &fakePreferencesRepo{}, nil, nil, nil)
			return err
		}},
		{"nil_preferences_repo", func() error {
			_, err := NewTaskReminderScheduler(&fakeReminderRepo{}, &fakeTaskLookup{}, nil, nil, &fakeNotificationRepo{}, nil, nil, nil, nil)
			return err
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.f()
			require.Error(t, err)
		})
	}
}
