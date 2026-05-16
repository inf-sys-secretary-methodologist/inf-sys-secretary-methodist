package scheduler

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	notifEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/entities"
	scheduleEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
)

// fakeWebPushRepo — minimal WebPushRepository surface для dispatch path.
// Только GetActiveByUserID exercised; остальные methods panic-on-call
// через nil if accidentally invoked (embedded interface pattern would be
// equivalent — chose explicit no-op stubs для readability).
type fakeWebPushRepo struct {
	mu   sync.Mutex
	subs map[int64][]*notifEntities.WebPushSubscription
}

func (r *fakeWebPushRepo) GetActiveByUserID(_ context.Context, userID int64) ([]*notifEntities.WebPushSubscription, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.subs[userID], nil
}

func (r *fakeWebPushRepo) Create(_ context.Context, _ *notifEntities.WebPushSubscription) error {
	return nil
}
func (r *fakeWebPushRepo) GetByID(_ context.Context, _ int64) (*notifEntities.WebPushSubscription, error) {
	return nil, nil
}
func (r *fakeWebPushRepo) GetByEndpoint(_ context.Context, _ string) (*notifEntities.WebPushSubscription, error) {
	return nil, nil
}
func (r *fakeWebPushRepo) GetByUserID(_ context.Context, _ int64) ([]*notifEntities.WebPushSubscription, error) {
	return nil, nil
}
func (r *fakeWebPushRepo) Update(_ context.Context, _ *notifEntities.WebPushSubscription) error {
	return nil
}
func (r *fakeWebPushRepo) UpdateLastUsed(_ context.Context, _ int64) error    { return nil }
func (r *fakeWebPushRepo) Deactivate(_ context.Context, _ int64) error        { return nil }
func (r *fakeWebPushRepo) Delete(_ context.Context, _ int64) error            { return nil }
func (r *fakeWebPushRepo) DeleteByEndpoint(_ context.Context, _ string) error { return nil }
func (r *fakeWebPushRepo) DeleteByUserID(_ context.Context, _ int64) error    { return nil }
func (r *fakeWebPushRepo) CountByUserID(_ context.Context, _ int64) (int64, error) {
	return 0, nil
}

type webPushCall struct {
	UserID  int64
	Payload *notifEntities.WebPushPayload
}

// fakeWebPushService captures dispatch invocations. configured controls
// the IsConfigured() gate; sendErr seeds dispatch failure.
type fakeWebPushService struct {
	calls      []webPushCall
	configured bool
	sendErr    error
}

func (s *fakeWebPushService) SendNotification(_ context.Context, _ *notifEntities.WebPushSubscription, _ *notifEntities.WebPushPayload) error {
	return nil
}

func (s *fakeWebPushService) SendToUser(_ context.Context, userID int64, payload *notifEntities.WebPushPayload) error {
	s.calls = append(s.calls, webPushCall{UserID: userID, Payload: payload})
	return s.sendErr
}

func (s *fakeWebPushService) GetVAPIDPublicKey() string { return "" }
func (s *fakeWebPushService) IsConfigured() bool        { return s.configured }

// TestReminderSchedulerSendPushReminder pins the v0.147.0 push dispatch
// contract on the event-side scheduler. Mirror к the telegram dispatch
// table-test shape: 4 cases × 3 fallback gates (nil deps / not-configured
// / dispatch-error) + happy path.
//
// Issue: github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist#226
func TestReminderSchedulerSendPushReminder(t *testing.T) {
	now := time.Date(2026, 5, 16, 12, 0, 0, 0, time.UTC)
	event := &scheduleEntities.Event{ID: 99, Title: "Защита диплома", StartTime: now.Add(30 * time.Minute)}
	reminder := scheduleEntities.NewEventReminder(event.ID, 7, scheduleEntities.ReminderTypePush, 30)

	cases := []struct {
		name         string
		wired        bool
		configured   bool
		hasSub       bool
		dispatchErr  error
		wantPushCall int
		wantInApp    int
	}{
		{
			name:         "wired + configured + active sub — push dispatched",
			wired:        true,
			configured:   true,
			hasSub:       true,
			wantPushCall: 1,
			wantInApp:    0,
		},
		{
			name:         "wired + not configured — falls back к in-app",
			wired:        true,
			configured:   false,
			hasSub:       true,
			wantPushCall: 0,
			wantInApp:    1,
		},
		{
			name:         "wired + no active subs — falls back к in-app",
			wired:        true,
			configured:   true,
			hasSub:       false,
			wantPushCall: 0,
			wantInApp:    1,
		},
		{
			name:         "wired + dispatch error — falls back к in-app after push attempt",
			wired:        true,
			configured:   true,
			hasSub:       true,
			dispatchErr:  errors.New("webpush 502"),
			wantPushCall: 1,
			wantInApp:    1,
		},
		{
			name:         "un-wired — falls back к in-app without push attempt",
			wired:        false,
			wantPushCall: 0,
			wantInApp:    1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			notifRepo := &fakeNotificationRepo{}
			var pushRepo *fakeWebPushRepo
			var pushService *fakeWebPushService
			if tc.wired {
				pushRepo = &fakeWebPushRepo{subs: map[int64][]*notifEntities.WebPushSubscription{}}
				if tc.hasSub {
					pushRepo.subs[reminder.UserID] = []*notifEntities.WebPushSubscription{
						{ID: 1, UserID: reminder.UserID, IsActive: true},
					}
				}
				pushService = &fakeWebPushService{configured: tc.configured, sendErr: tc.dispatchErr}
			}

			s := buildEventScheduler(t, notifRepo, nil, nil, nil)
			if tc.wired {
				s.WithWebPushDispatch(pushRepo, pushService)
			}

			err := s.sendPushReminder(context.Background(), reminder, event)
			require.NoError(t, err)

			if tc.wired {
				assert.Len(t, pushService.calls, tc.wantPushCall, "push dispatch count")
				if tc.wantPushCall > 0 {
					assert.Equal(t, reminder.UserID, pushService.calls[0].UserID)
					assert.NotNil(t, pushService.calls[0].Payload)
					assert.Contains(t, pushService.calls[0].Payload.Title, "Напоминание")
					// Deep-link URL pins the defense-critical "click reminder
					// → open event" navigation contract. A regression that
					// silently drops the URL would surface here, not only in
					// the manual smoke test.
					assert.Equal(t, fmt.Sprintf("/schedule/events/%d", event.ID), pushService.calls[0].Payload.URL)
					assert.Equal(t, fmt.Sprintf("event-reminder-%d", reminder.ID), pushService.calls[0].Payload.Tag)
				}
			}
			assert.Len(t, notifRepo.created, tc.wantInApp, "in-app fallback count")
		})
	}
}

// TestReminderSchedulerProcessReminderPushIntegration walks the full
// processReminder switch path для the ReminderTypePush branch so a
// regression of the `case scheduleEntities.ReminderTypePush:` arm
// в reminder_scheduler.go surfaces here rather than only in the direct
// sendPushReminder unit. Mirror к the telegram integration test pattern.
func TestReminderSchedulerProcessReminderPushIntegration(t *testing.T) {
	now := time.Date(2026, 5, 16, 12, 0, 0, 0, time.UTC)
	event := &scheduleEntities.Event{ID: 99, Title: "Защита диплома", StartTime: now.Add(30 * time.Minute)}
	reminder := scheduleEntities.NewEventReminder(event.ID, 7, scheduleEntities.ReminderTypePush, 30)

	notifRepo := &fakeNotificationRepo{}
	prefs := enabledPrefs()
	prefsRepo := &fakePreferencesRepo{prefs: prefs}
	eventRepo := &fakeEventRepo{event: event}
	pushRepo := &fakeWebPushRepo{subs: map[int64][]*notifEntities.WebPushSubscription{
		reminder.UserID: {{ID: 1, UserID: reminder.UserID, IsActive: true}},
	}}
	pushService := &fakeWebPushService{configured: true}

	s := buildEventSchedulerWithEvents(t, notifRepo, prefsRepo, nil, nil, eventRepo)
	s.WithWebPushDispatch(pushRepo, pushService)

	err := s.processReminder(context.Background(), reminder)
	require.NoError(t, err)

	assert.Len(t, pushService.calls, 1, "push dispatch from processReminder")
	assert.Equal(t, reminder.UserID, pushService.calls[0].UserID)
	assert.Empty(t, notifRepo.created, "no in-app fallback when push dispatch succeeds")
}
