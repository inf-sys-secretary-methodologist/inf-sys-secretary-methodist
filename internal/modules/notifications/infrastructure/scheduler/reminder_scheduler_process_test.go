package scheduler

// v0.153.11 Phase 6 #196 backfill — covers processPendingReminders + sendEmailReminder
// + the email branch of processReminder. Mirrors к telegram/push integration test
// pattern from v0.138.0/v0.147.0 (separate dispatch unit + full processReminder walk).

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	notifEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/entities"
	notifServices "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/services"
	scheduleEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
	scheduleRepos "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/repositories"
)

// ----- fakes -----

// fakeEventReminderRepo — narrow EventReminderRepository surface
// для processPendingReminders. Only GetPendingReminders + MarkMultipleAsSent
// matter; rest is panic-on-call via embedded interface (same pattern as
// fakeEventRepo from reminder_scheduler_telegram_test.go).
type fakeEventReminderRepo struct {
	scheduleRepos.EventReminderRepository
	pending    []*scheduleEntities.EventReminder
	pendingErr error
	markedIDs  []int64
	markErr    error
}

func (r *fakeEventReminderRepo) GetPendingReminders(_ context.Context, _ time.Time) ([]*scheduleEntities.EventReminder, error) {
	return r.pending, r.pendingErr
}

func (r *fakeEventReminderRepo) MarkMultipleAsSent(_ context.Context, ids []int64) error {
	r.markedIDs = append(r.markedIDs, ids...)
	return r.markErr
}

// fakeEventLookup — per-ID event lookup so processReminder branches
// (success / not-found / GetByID error) can be exercised independently.
// Distinct from fakeEventRepo (single-event helper) used by the
// telegram/push integration tests.
type fakeEventLookup struct {
	scheduleRepos.EventRepository
	events map[int64]*scheduleEntities.Event
	errs   map[int64]error
}

func (r *fakeEventLookup) GetByID(_ context.Context, id int64) (*scheduleEntities.Event, error) {
	if err, ok := r.errs[id]; ok {
		return nil, err
	}
	return r.events[id], nil
}

// emailCall captures SendNotification invocations.
type emailCall struct {
	To, Subject, Body string
}

// fakeEmailService — captures SendNotification calls; returns sendErr
// when set. Other methods are inert no-ops because они не exercised
// в the reminder dispatch path.
type fakeEmailService struct {
	calls   []emailCall
	sendErr error
}

func (s *fakeEmailService) SendEmail(_ context.Context, _ *notifServices.SendEmailRequest) error {
	return nil
}

func (s *fakeEmailService) SendWelcomeEmail(_ context.Context, _, _ string) error {
	return nil
}

func (s *fakeEmailService) SendPasswordResetEmail(_ context.Context, _, _ string) error {
	return nil
}

func (s *fakeEmailService) SendNotification(_ context.Context, to, subject, body string) error {
	s.calls = append(s.calls, emailCall{To: to, Subject: subject, Body: body})
	return s.sendErr
}

// ----- processPendingReminders -----

func TestReminderScheduler_ProcessPendingReminders_NoPending_NoMarkCall(t *testing.T) {
	remRepo := &fakeEventReminderRepo{pending: nil}
	notifRepo := &fakeNotificationRepo{}
	prefsRepo := &fakePreferencesRepo{prefs: enabledPrefs()}
	eventRepo := &fakeEventLookup{events: map[int64]*scheduleEntities.Event{}}

	s, err := NewReminderScheduler(nil, remRepo, eventRepo, notifRepo, prefsRepo, nil, nil)
	require.NoError(t, err)

	s.processPendingReminders()

	assert.Empty(t, remRepo.markedIDs, "empty pending list — MarkMultipleAsSent must not be called")
	assert.Empty(t, notifRepo.created)
}

func TestReminderScheduler_ProcessPendingReminders_RepoError_EarlyReturn(t *testing.T) {
	remRepo := &fakeEventReminderRepo{pendingErr: errors.New("db down")}
	notifRepo := &fakeNotificationRepo{}
	prefsRepo := &fakePreferencesRepo{prefs: enabledPrefs()}
	eventRepo := &fakeEventLookup{events: map[int64]*scheduleEntities.Event{}}

	s, err := NewReminderScheduler(nil, remRepo, eventRepo, notifRepo, prefsRepo, nil, nil)
	require.NoError(t, err)

	s.processPendingReminders()

	assert.Empty(t, remRepo.markedIDs, "GetPendingReminders error — no MarkMultipleAsSent")
	assert.Empty(t, notifRepo.created)
}

func TestReminderScheduler_ProcessPendingReminders_HappyPath_MarksAll(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	event1 := &scheduleEntities.Event{ID: 10, Title: "Лекция", StartTime: now.Add(30 * time.Minute)}
	event2 := &scheduleEntities.Event{ID: 20, Title: "Семинар", StartTime: now.Add(60 * time.Minute)}
	rem1 := scheduleEntities.NewEventReminder(event1.ID, 7, scheduleEntities.ReminderTypeInApp, 30)
	rem1.ID = 101
	rem2 := scheduleEntities.NewEventReminder(event2.ID, 7, scheduleEntities.ReminderTypeInApp, 60)
	rem2.ID = 202

	remRepo := &fakeEventReminderRepo{pending: []*scheduleEntities.EventReminder{rem1, rem2}}
	notifRepo := &fakeNotificationRepo{}
	prefsRepo := &fakePreferencesRepo{prefs: enabledPrefs()}
	eventRepo := &fakeEventLookup{events: map[int64]*scheduleEntities.Event{
		event1.ID: event1,
		event2.ID: event2,
	}}

	s, err := NewReminderScheduler(nil, remRepo, eventRepo, notifRepo, prefsRepo, nil, nil)
	require.NoError(t, err)

	s.processPendingReminders()

	assert.ElementsMatch(t, []int64{101, 202}, remRepo.markedIDs, "all successfully processed reminders marked sent")
	assert.Len(t, notifRepo.created, 2, "both in-app notifications created")
}

func TestReminderScheduler_ProcessPendingReminders_PartialError_SkipsFailing(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	event2 := &scheduleEntities.Event{ID: 20, Title: "Семинар", StartTime: now.Add(60 * time.Minute)}
	rem1 := scheduleEntities.NewEventReminder(10, 7, scheduleEntities.ReminderTypeInApp, 30)
	rem1.ID = 101
	rem2 := scheduleEntities.NewEventReminder(event2.ID, 7, scheduleEntities.ReminderTypeInApp, 60)
	rem2.ID = 202

	remRepo := &fakeEventReminderRepo{pending: []*scheduleEntities.EventReminder{rem1, rem2}}
	notifRepo := &fakeNotificationRepo{}
	prefsRepo := &fakePreferencesRepo{prefs: enabledPrefs()}
	// rem1 → event 10 → returns error; rem2 → event 20 → success.
	eventRepo := &fakeEventLookup{
		events: map[int64]*scheduleEntities.Event{event2.ID: event2},
		errs:   map[int64]error{10: errors.New("event lookup failed")},
	}

	s, err := NewReminderScheduler(nil, remRepo, eventRepo, notifRepo, prefsRepo, nil, nil)
	require.NoError(t, err)

	s.processPendingReminders()

	assert.Equal(t, []int64{202}, remRepo.markedIDs, "only successful reminder marked sent")
	assert.Len(t, notifRepo.created, 1)
}

func TestReminderScheduler_ProcessPendingReminders_MarkError_NoPanic(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	event := &scheduleEntities.Event{ID: 10, Title: "Лекция", StartTime: now.Add(30 * time.Minute)}
	rem := scheduleEntities.NewEventReminder(event.ID, 7, scheduleEntities.ReminderTypeInApp, 30)
	rem.ID = 101

	remRepo := &fakeEventReminderRepo{
		pending: []*scheduleEntities.EventReminder{rem},
		markErr: errors.New("mark sent failed"),
	}
	notifRepo := &fakeNotificationRepo{}
	prefsRepo := &fakePreferencesRepo{prefs: enabledPrefs()}
	eventRepo := &fakeEventLookup{events: map[int64]*scheduleEntities.Event{event.ID: event}}

	s, err := NewReminderScheduler(nil, remRepo, eventRepo, notifRepo, prefsRepo, nil, nil)
	require.NoError(t, err)

	// Должно не паниковать — ошибка MarkMultipleAsSent логируется.
	assert.NotPanics(t, s.processPendingReminders)
	assert.Equal(t, []int64{101}, remRepo.markedIDs, "ids passed to MarkMultipleAsSent even if it errors")
	assert.Len(t, notifRepo.created, 1)
}

// ----- sendEmailReminder -----

func TestReminderScheduler_SendEmailReminder_HappyPath(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	emailSvc := &fakeEmailService{}
	notifRepo := &fakeNotificationRepo{}
	prefsRepo := &fakePreferencesRepo{prefs: enabledPrefs()}
	eventRepo := &fakeEventLookup{events: map[int64]*scheduleEntities.Event{}}
	remRepo := &fakeEventReminderRepo{}

	s, err := NewReminderScheduler(db, remRepo, eventRepo, notifRepo, prefsRepo, emailSvc, nil)
	require.NoError(t, err)

	mock.ExpectQuery("SELECT email FROM users WHERE id").
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"email"}).AddRow("teacher@example.com"))

	event := &scheduleEntities.Event{
		ID:        10,
		Title:     "Защита диплома",
		StartTime: time.Date(2026, 5, 19, 14, 30, 0, 0, time.UTC),
	}
	reminder := scheduleEntities.NewEventReminder(event.ID, 7, scheduleEntities.ReminderTypeEmail, 60)

	err = s.sendEmailReminder(context.Background(), reminder, event)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
	require.Len(t, emailSvc.calls, 1)
	assert.Equal(t, "teacher@example.com", emailSvc.calls[0].To)
	assert.Contains(t, emailSvc.calls[0].Subject, "Напоминание")
	assert.Contains(t, emailSvc.calls[0].Subject, "Защита диплома")
	assert.Contains(t, emailSvc.calls[0].Body, "1 час")
	assert.Contains(t, emailSvc.calls[0].Body, "Защита диплома")
}

func TestReminderScheduler_SendEmailReminder_UserLookupFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	emailSvc := &fakeEmailService{}
	notifRepo := &fakeNotificationRepo{}
	prefsRepo := &fakePreferencesRepo{prefs: enabledPrefs()}
	eventRepo := &fakeEventLookup{events: map[int64]*scheduleEntities.Event{}}
	remRepo := &fakeEventReminderRepo{}

	s, err := NewReminderScheduler(db, remRepo, eventRepo, notifRepo, prefsRepo, emailSvc, nil)
	require.NoError(t, err)

	mock.ExpectQuery("SELECT email FROM users WHERE id").
		WithArgs(int64(7)).
		WillReturnError(errors.New("db connection lost"))

	event := &scheduleEntities.Event{ID: 10, Title: "T", StartTime: time.Now()}
	reminder := scheduleEntities.NewEventReminder(event.ID, 7, scheduleEntities.ReminderTypeEmail, 60)

	err = s.sendEmailReminder(context.Background(), reminder, event)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get user email")
	assert.Empty(t, emailSvc.calls, "email service must not be called when user lookup fails")
}

func TestReminderScheduler_SendEmailReminder_EmailServiceFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	emailSvc := &fakeEmailService{sendErr: errors.New("smtp 5xx")}
	notifRepo := &fakeNotificationRepo{}
	prefsRepo := &fakePreferencesRepo{prefs: enabledPrefs()}
	eventRepo := &fakeEventLookup{events: map[int64]*scheduleEntities.Event{}}
	remRepo := &fakeEventReminderRepo{}

	s, err := NewReminderScheduler(db, remRepo, eventRepo, notifRepo, prefsRepo, emailSvc, nil)
	require.NoError(t, err)

	mock.ExpectQuery("SELECT email FROM users WHERE id").
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"email"}).AddRow("teacher@example.com"))

	event := &scheduleEntities.Event{ID: 10, Title: "T", StartTime: time.Now()}
	reminder := scheduleEntities.NewEventReminder(event.ID, 7, scheduleEntities.ReminderTypeEmail, 60)

	err = s.sendEmailReminder(context.Background(), reminder, event)
	require.Error(t, err)
	assert.Equal(t, "smtp 5xx", err.Error())
	require.Len(t, emailSvc.calls, 1, "send was attempted before failing")
}

// ----- processReminder branch coverage -----

func TestReminderScheduler_ProcessReminder_EmailIntegration(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	event := &scheduleEntities.Event{ID: 10, Title: "Лекция", StartTime: now.Add(30 * time.Minute)}
	reminder := scheduleEntities.NewEventReminder(event.ID, 7, scheduleEntities.ReminderTypeEmail, 30)

	emailSvc := &fakeEmailService{}
	notifRepo := &fakeNotificationRepo{}
	prefsRepo := &fakePreferencesRepo{prefs: enabledPrefs()}
	eventRepo := &fakeEventLookup{events: map[int64]*scheduleEntities.Event{event.ID: event}}
	remRepo := &fakeEventReminderRepo{}

	s, err := NewReminderScheduler(db, remRepo, eventRepo, notifRepo, prefsRepo, emailSvc, nil)
	require.NoError(t, err)

	mock.ExpectQuery("SELECT email FROM users WHERE id").
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"email"}).AddRow("user@example.com"))

	err = s.processReminder(context.Background(), reminder)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
	require.Len(t, emailSvc.calls, 1, "email dispatched from processReminder")
	assert.Empty(t, notifRepo.created, "email-type reminder must not fall through к in-app")
}

func TestReminderScheduler_ProcessReminder_EmailDisabled_NoDispatch(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	event := &scheduleEntities.Event{ID: 10, Title: "Лекция", StartTime: now.Add(30 * time.Minute)}
	reminder := scheduleEntities.NewEventReminder(event.ID, 7, scheduleEntities.ReminderTypeEmail, 30)

	emailSvc := &fakeEmailService{}
	notifRepo := &fakeNotificationRepo{}
	prefs := enabledPrefs()
	prefs.EmailEnabled = false
	prefsRepo := &fakePreferencesRepo{prefs: prefs}
	eventRepo := &fakeEventLookup{events: map[int64]*scheduleEntities.Event{event.ID: event}}
	remRepo := &fakeEventReminderRepo{}

	s, err := NewReminderScheduler(nil, remRepo, eventRepo, notifRepo, prefsRepo, emailSvc, nil)
	require.NoError(t, err)

	err = s.processReminder(context.Background(), reminder)
	require.NoError(t, err)
	assert.Empty(t, emailSvc.calls, "email disabled in prefs — no dispatch")
	assert.Empty(t, notifRepo.created, "email branch returns nil without falling through к in-app")
}

func TestReminderScheduler_ProcessReminder_UnknownType_FallsBackToInApp(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	event := &scheduleEntities.Event{ID: 10, Title: "Лекция", StartTime: now.Add(30 * time.Minute)}
	reminder := scheduleEntities.NewEventReminder(event.ID, 7, scheduleEntities.ReminderType("unknown"), 30)

	notifRepo := &fakeNotificationRepo{}
	prefsRepo := &fakePreferencesRepo{prefs: enabledPrefs()}
	eventRepo := &fakeEventLookup{events: map[int64]*scheduleEntities.Event{event.ID: event}}
	remRepo := &fakeEventReminderRepo{}

	s, err := NewReminderScheduler(nil, remRepo, eventRepo, notifRepo, prefsRepo, nil, nil)
	require.NoError(t, err)

	err = s.processReminder(context.Background(), reminder)
	require.NoError(t, err)
	require.Len(t, notifRepo.created, 1, "default branch falls back к in-app")
}

func TestReminderScheduler_ProcessReminder_QuietHours_Skipped(t *testing.T) {
	quietNow := time.Now().UTC()
	startHour := quietNow.Add(-time.Hour).Format("15:04")
	endHour := quietNow.Add(time.Hour).Format("15:04")

	event := &scheduleEntities.Event{ID: 10, Title: "Лекция", StartTime: quietNow.Add(30 * time.Minute)}
	reminder := scheduleEntities.NewEventReminder(event.ID, 7, scheduleEntities.ReminderTypeInApp, 30)

	notifRepo := &fakeNotificationRepo{}
	prefs := enabledPrefs()
	prefs.QuietHoursEnabled = true
	prefs.QuietHoursStart = startHour
	prefs.QuietHoursEnd = endHour
	prefs.Timezone = "UTC"
	prefsRepo := &fakePreferencesRepo{prefs: prefs}
	eventRepo := &fakeEventLookup{events: map[int64]*scheduleEntities.Event{event.ID: event}}
	remRepo := &fakeEventReminderRepo{}

	s, err := NewReminderScheduler(nil, remRepo, eventRepo, notifRepo, prefsRepo, nil, nil)
	require.NoError(t, err)

	err = s.processReminder(context.Background(), reminder)
	require.NoError(t, err)
	assert.Empty(t, notifRepo.created, "quiet hours — no in-app dispatch")
}

func TestReminderScheduler_ProcessReminder_EventNotFound_ReturnsError(t *testing.T) {
	reminder := scheduleEntities.NewEventReminder(999, 7, scheduleEntities.ReminderTypeInApp, 30)

	notifRepo := &fakeNotificationRepo{}
	prefsRepo := &fakePreferencesRepo{prefs: enabledPrefs()}
	eventRepo := &fakeEventLookup{events: map[int64]*scheduleEntities.Event{}} // empty map → returns nil event
	remRepo := &fakeEventReminderRepo{}

	s, err := NewReminderScheduler(nil, remRepo, eventRepo, notifRepo, prefsRepo, nil, nil)
	require.NoError(t, err)

	err = s.processReminder(context.Background(), reminder)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "event not found")
	assert.Empty(t, notifRepo.created)
}

func TestReminderScheduler_ProcessReminder_PrefsLookupFails(t *testing.T) {
	event := &scheduleEntities.Event{ID: 10, Title: "T", StartTime: time.Now()}
	reminder := scheduleEntities.NewEventReminder(event.ID, 7, scheduleEntities.ReminderTypeInApp, 30)

	// fakePreferencesRepo с nil prefs всё ещё возвращает (nil, nil) per
	// текущей impl, не error. Чтобы вызвать error-branch, используем
	// inline fake who возвращает explicit error.
	prefsRepo := &errPrefsRepo{err: errors.New("prefs db down")}
	eventRepo := &fakeEventLookup{events: map[int64]*scheduleEntities.Event{event.ID: event}}
	notifRepo := &fakeNotificationRepo{}
	remRepo := &fakeEventReminderRepo{}

	s, err := NewReminderScheduler(nil, remRepo, eventRepo, notifRepo, prefsRepo, nil, nil)
	require.NoError(t, err)

	err = s.processReminder(context.Background(), reminder)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get preferences")
}

// errPrefsRepo — minimal PreferencesRepository emitting error on GetOrCreate.
// Distinct от fakePreferencesRepo (которая возвращает stored prefs).
type errPrefsRepo struct {
	fakePreferencesRepo
	err error
}

func (r *errPrefsRepo) GetOrCreate(_ context.Context, _ int64) (*notifEntities.UserNotificationPreferences, error) {
	return nil, r.err
}
