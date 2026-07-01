package scheduler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	notifEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/entities"
	scheduleRepos "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/application/usecases"
	scheduleEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
)

// fakeEventRepo satisfies the broad scheduleRepos.EventRepository
// surface через embedding the interface — only GetByID is overridden;
// any unused method would panic on call (nil interface). Sufficient
// для the processReminder integration path tested here.
type fakeEventRepo struct {
	scheduleRepos.EventRepository
	event *scheduleEntities.Event
}

func (r *fakeEventRepo) GetByID(_ context.Context, _ int64) (*scheduleEntities.Event, error) {
	return r.event, nil
}

// reminder_scheduler_telegram_test.go — v0.138.1 carry-forward fix
// covers the existing event reminder ReminderScheduler.sendTelegramReminder
// dispatch path. Mirror к TaskReminderScheduler.sendTelegram contract.
//
// Calling sendTelegramReminder directly is safe (same-package test);
// processReminder + processPendingReminders branches are not exercised
// here — those are integration-shaped paths already covered indirectly
// through manual smoke + production wiring.

func buildEventScheduler(
	t *testing.T,
	notifRepo *fakeNotificationRepo,
	prefsRepo *fakePreferencesRepo,
	tgRepo *fakeTelegramRepo,
	tgService *fakeTelegramService,
) *ReminderScheduler {
	t.Helper()
	return buildEventSchedulerWithEvents(t, notifRepo, prefsRepo, tgRepo, tgService, nil)
}

// buildEventSchedulerWithEvents extends buildEventScheduler с an
// optional eventRepo for integration-shaped tests that exercise the
// full processReminder switch path.
func buildEventSchedulerWithEvents(
	t *testing.T,
	notifRepo *fakeNotificationRepo,
	prefsRepo *fakePreferencesRepo,
	tgRepo *fakeTelegramRepo,
	tgService *fakeTelegramService,
	eventRepo scheduleRepos.EventRepository,
) *ReminderScheduler {
	t.Helper()

	s, err := NewReminderScheduler(
		nil,
		nil,
		eventRepo,
		notifRepo,
		prefsRepo,
		nil,
		nil,
	)
	require.NoError(t, err)
	if tgRepo != nil || tgService != nil {
		s.WithTelegramDispatch(tgRepo, tgService)
	}
	return s
}

func TestReminderSchedulerSendTelegram(t *testing.T) {
	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	event := &scheduleEntities.Event{ID: 99, Title: "Защита диплома", StartTime: now.Add(30 * time.Minute)}
	reminder := scheduleEntities.NewEventReminder(event.ID, 7, scheduleEntities.ReminderTypeTelegram, 30)

	cases := []struct {
		name        string
		wired       bool
		connActive  bool
		dispatchErr error
		wantTGCalls int
		wantInApp   int
	}{
		{
			name:        "wired + active connection — telegram dispatched",
			wired:       true,
			connActive:  true,
			wantTGCalls: 1,
			wantInApp:   0,
		},
		{
			name:        "wired + no connection — falls back к in-app",
			wired:       true,
			connActive:  false,
			wantTGCalls: 0,
			wantInApp:   1,
		},
		{
			name:        "wired + dispatch failure — falls back к in-app after telegram attempt",
			wired:       true,
			connActive:  true,
			dispatchErr: errors.New("composio rate limit"),
			wantTGCalls: 1,
			wantInApp:   1,
		},
		{
			name:        "un-wired — falls back к in-app without telegram attempt",
			wired:       false,
			connActive:  false,
			wantTGCalls: 0,
			wantInApp:   1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			notifRepo := &fakeNotificationRepo{}
			var tgRepo *fakeTelegramRepo
			var tgService *fakeTelegramService
			if tc.wired {
				tgRepo = &fakeTelegramRepo{conns: map[int64]*notifEntities.TelegramConnection{}}
				if tc.connActive {
					tgRepo.conns[reminder.UserID] = &notifEntities.TelegramConnection{
						UserID:         reminder.UserID,
						TelegramChatID: 555111000,
						IsActive:       true,
					}
				}
				tgService = &fakeTelegramService{sendErr: tc.dispatchErr}
			}

			s := buildEventScheduler(t, notifRepo, nil, tgRepo, tgService)

			err := s.sendTelegramReminder(context.Background(), reminder, event)
			require.NoError(t, err)

			if tc.wired {
				assert.Len(t, tgService.calls, tc.wantTGCalls, "telegram dispatch count")
				if tc.wantTGCalls > 0 {
					assert.Equal(t, "555111000", tgService.calls[0].ChatID)
					assert.Contains(t, tgService.calls[0].Title, "Напоминание")
				}
			}
			assert.Len(t, notifRepo.created, tc.wantInApp, "in-app fallback count")
		})
	}
}

// TestReminderSchedulerProcessReminderTelegramIntegration walks the full
// processReminder switch path для the ReminderTypeTelegram branch so a
// regression of the `case scheduleEntities.ReminderTypeTelegram:` arm
// в reminder_scheduler.go would surface here rather than only in the
// direct sendTelegramReminder unit. Closes the reviewer Tier 1 gap.
func TestReminderSchedulerProcessReminderTelegramIntegration(t *testing.T) {
	now := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	event := &scheduleEntities.Event{ID: 99, Title: "Защита диплома", StartTime: now.Add(30 * time.Minute)}
	reminder := scheduleEntities.NewEventReminder(event.ID, 7, scheduleEntities.ReminderTypeTelegram, 30)

	notifRepo := &fakeNotificationRepo{}
	prefs := enabledPrefs()
	prefsRepo := &fakePreferencesRepo{prefs: prefs}
	eventRepo := &fakeEventRepo{event: event}
	tgRepo := &fakeTelegramRepo{conns: map[int64]*notifEntities.TelegramConnection{
		reminder.UserID: {UserID: reminder.UserID, TelegramChatID: 555111000, IsActive: true},
	}}
	tgService := &fakeTelegramService{}

	s := buildEventSchedulerWithEvents(t, notifRepo, prefsRepo, tgRepo, tgService, eventRepo)

	err := s.processReminder(context.Background(), reminder)
	require.NoError(t, err)

	assert.Len(t, tgService.calls, 1, "telegram dispatch from processReminder")
	assert.Equal(t, "555111000", tgService.calls[0].ChatID)
	assert.Empty(t, notifRepo.created, "no in-app fallback when telegram dispatch succeeds")
}
