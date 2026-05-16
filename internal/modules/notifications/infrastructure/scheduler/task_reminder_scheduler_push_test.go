package scheduler

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	notifEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/entities"
	tasksEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/entities"
)

// buildSchedulerWithPush mirrors buildScheduler but threads
// WithWebPushDispatch onto the scheduler. Push repo + service may be
// nil — that exercises the un-wired fallback. Reuses the broad fakes
// from task_reminder_scheduler_test.go (fakeReminderRepo, fakeTaskLookup,
// fakeNotificationRepo, fakePreferencesRepo) and the push fakes from
// reminder_scheduler_push_test.go (fakeWebPushRepo, fakeWebPushService).
func buildSchedulerWithPush(
	t *testing.T,
	reminders []*tasksEntities.TaskReminder,
	views map[int64]*TaskDispatchView,
	pushRepo *fakeWebPushRepo,
	pushService *fakeWebPushService,
	notifRepo *fakeNotificationRepo,
	prefs *notifEntities.UserNotificationPreferences,
	now time.Time,
) (*TaskReminderScheduler, *fakeReminderRepo) {
	t.Helper()
	repo := &fakeReminderRepo{pending: reminders}
	s, err := NewTaskReminderScheduler(
		repo,
		&fakeTaskLookup{views: views},
		&fakeTelegramRepo{},
		&fakeTelegramService{},
		notifRepo,
		&fakePreferencesRepo{prefs: prefs},
		nil,
		nil,
		&TaskReminderSchedulerConfig{CheckInterval: time.Hour, Clock: fakeClock{now: now}},
	)
	require.NoError(t, err)
	if pushRepo != nil || pushService != nil {
		s.WithWebPushDispatch(pushRepo, pushService)
	}
	return s, repo
}

// TestProcessOnce_PushHappyPath pins the v0.147.0 task-side push
// dispatch contract: a push-type reminder, configured WebPushService,
// active subscription AND push enabled in prefs dispatches via
// WebPushService.SendToUser. No fallback к in-app.
//
// Issue: #226
func TestProcessOnce_PushHappyPath(t *testing.T) {
	now := time.Date(2026, 5, 16, 12, 0, 0, 0, time.UTC)
	due := now.Add(15 * time.Minute)
	reminder := tasksEntities.HydrateFromPersistence(101, 42, 7, tasksEntities.ReminderTypePush, 15, false, nil, now.Add(-time.Hour))
	views := map[int64]*TaskDispatchView{42: {Title: "Утвердить РПД", DueDate: &due}}
	pushRepo := &fakeWebPushRepo{subs: map[int64][]*notifEntities.WebPushSubscription{
		7: {{ID: 1, UserID: 7, IsActive: true}},
	}}
	pushService := &fakeWebPushService{configured: true}
	notif := &fakeNotificationRepo{}

	s, repo := buildSchedulerWithPush(t, []*tasksEntities.TaskReminder{reminder}, views, pushRepo, pushService, notif, enabledPrefs(), now)

	s.ProcessOnce(context.Background(), now)

	require.Len(t, pushService.calls, 1, "push dispatch must fire exactly once")
	assert.Equal(t, int64(7), pushService.calls[0].UserID)
	assert.Contains(t, pushService.calls[0].Payload.Title, "Напоминание")
	assert.Contains(t, pushService.calls[0].Payload.Body, "Утвердить РПД")
	// Deep-link URL pins the defense-critical "click reminder → open task"
	// navigation contract. A regression that silently drops the URL or Tag
	// would surface here, not only in manual smoke.
	assert.Equal(t, fmt.Sprintf("/tasks/%d", reminder.TaskID()), pushService.calls[0].Payload.URL)
	assert.Equal(t, fmt.Sprintf("task-reminder-%d", reminder.ID()), pushService.calls[0].Payload.Tag)
	assert.Empty(t, notif.created, "happy-path push dispatch must not fan out к in-app")
	assert.Equal(t, []int64{101}, repo.markedIDs, "processed reminder marked as sent")
}

// TestProcessOnce_PushNotConfigured_FallbackInApp — VAPID keys missing
// → graceful fallback к in-app.
func TestProcessOnce_PushNotConfigured_FallbackInApp(t *testing.T) {
	now := time.Date(2026, 5, 16, 12, 0, 0, 0, time.UTC)
	due := now.Add(15 * time.Minute)
	reminder := tasksEntities.HydrateFromPersistence(101, 42, 7, tasksEntities.ReminderTypePush, 15, false, nil, now.Add(-time.Hour))
	views := map[int64]*TaskDispatchView{42: {Title: "Утвердить РПД", DueDate: &due}}
	pushRepo := &fakeWebPushRepo{subs: map[int64][]*notifEntities.WebPushSubscription{
		7: {{ID: 1, UserID: 7, IsActive: true}},
	}}
	pushService := &fakeWebPushService{configured: false}
	notif := &fakeNotificationRepo{}

	s, _ := buildSchedulerWithPush(t, []*tasksEntities.TaskReminder{reminder}, views, pushRepo, pushService, notif, enabledPrefs(), now)

	s.ProcessOnce(context.Background(), now)

	assert.Empty(t, pushService.calls, "no push dispatch when service unconfigured")
	require.Len(t, notif.created, 1, "in-app fallback must fire")
}

// TestProcessOnce_PushNoSubscriptions_FallbackInApp — user has no
// active webpush subscription → in-app fallback so reminder isn't lost.
func TestProcessOnce_PushNoSubscriptions_FallbackInApp(t *testing.T) {
	now := time.Date(2026, 5, 16, 12, 0, 0, 0, time.UTC)
	due := now.Add(15 * time.Minute)
	reminder := tasksEntities.HydrateFromPersistence(101, 42, 7, tasksEntities.ReminderTypePush, 15, false, nil, now.Add(-time.Hour))
	views := map[int64]*TaskDispatchView{42: {Title: "Утвердить РПД", DueDate: &due}}
	pushRepo := &fakeWebPushRepo{subs: map[int64][]*notifEntities.WebPushSubscription{}}
	pushService := &fakeWebPushService{configured: true}
	notif := &fakeNotificationRepo{}

	s, _ := buildSchedulerWithPush(t, []*tasksEntities.TaskReminder{reminder}, views, pushRepo, pushService, notif, enabledPrefs(), now)

	s.ProcessOnce(context.Background(), now)

	assert.Empty(t, pushService.calls, "no push dispatch when user has no subscriptions")
	require.Len(t, notif.created, 1, "in-app fallback must fire")
}

// TestProcessOnce_PushDispatchError_FallbackInApp — WebPush 5xx →
// fallback к in-app after attempting dispatch.
func TestProcessOnce_PushDispatchError_FallbackInApp(t *testing.T) {
	now := time.Date(2026, 5, 16, 12, 0, 0, 0, time.UTC)
	due := now.Add(15 * time.Minute)
	reminder := tasksEntities.HydrateFromPersistence(101, 42, 7, tasksEntities.ReminderTypePush, 15, false, nil, now.Add(-time.Hour))
	views := map[int64]*TaskDispatchView{42: {Title: "Утвердить РПД", DueDate: &due}}
	pushRepo := &fakeWebPushRepo{subs: map[int64][]*notifEntities.WebPushSubscription{
		7: {{ID: 1, UserID: 7, IsActive: true}},
	}}
	pushService := &fakeWebPushService{configured: true, sendErr: errors.New("webpush 502")}
	notif := &fakeNotificationRepo{}

	s, _ := buildSchedulerWithPush(t, []*tasksEntities.TaskReminder{reminder}, views, pushRepo, pushService, notif, enabledPrefs(), now)

	s.ProcessOnce(context.Background(), now)

	require.Len(t, pushService.calls, 1, "push dispatch attempted")
	require.Len(t, notif.created, 1, "in-app fallback fires on dispatch failure")
}

// TestProcessOnce_PushUnwired_FallbackInApp — push deps not wired
// onto the scheduler → in-app fallback без push attempt.
func TestProcessOnce_PushUnwired_FallbackInApp(t *testing.T) {
	now := time.Date(2026, 5, 16, 12, 0, 0, 0, time.UTC)
	due := now.Add(15 * time.Minute)
	reminder := tasksEntities.HydrateFromPersistence(101, 42, 7, tasksEntities.ReminderTypePush, 15, false, nil, now.Add(-time.Hour))
	views := map[int64]*TaskDispatchView{42: {Title: "Утвердить РПД", DueDate: &due}}
	notif := &fakeNotificationRepo{}

	s, _ := buildSchedulerWithPush(t, []*tasksEntities.TaskReminder{reminder}, views, nil, nil, notif, enabledPrefs(), now)

	s.ProcessOnce(context.Background(), now)

	require.Len(t, notif.created, 1, "in-app fallback fires when push un-wired")
}

// TestProcessOnce_PushDisabledInPrefs_FallbackInApp — channel toggle off
// in user prefs → fallback к in-app без push attempt (existing catch-all
// behavior at the bottom of processReminder).
func TestProcessOnce_PushDisabledInPrefs_FallbackInApp(t *testing.T) {
	now := time.Date(2026, 5, 16, 12, 0, 0, 0, time.UTC)
	due := now.Add(15 * time.Minute)
	reminder := tasksEntities.HydrateFromPersistence(101, 42, 7, tasksEntities.ReminderTypePush, 15, false, nil, now.Add(-time.Hour))
	views := map[int64]*TaskDispatchView{42: {Title: "Утвердить РПД", DueDate: &due}}
	pushRepo := &fakeWebPushRepo{subs: map[int64][]*notifEntities.WebPushSubscription{
		7: {{ID: 1, UserID: 7, IsActive: true}},
	}}
	pushService := &fakeWebPushService{configured: true}
	notif := &fakeNotificationRepo{}
	prefs := enabledPrefs()
	prefs.PushEnabled = false

	s, _ := buildSchedulerWithPush(t, []*tasksEntities.TaskReminder{reminder}, views, pushRepo, pushService, notif, prefs, now)

	s.ProcessOnce(context.Background(), now)

	assert.Empty(t, pushService.calls, "push disabled in prefs — no dispatch attempt")
	require.Len(t, notif.created, 1, "in-app fallback fires when push disabled in prefs")
}
