package scheduler

// v0.153.8 Phase 6 backfill — pure formatters across reminder_scheduler.go
// + task_reminder_scheduler.go. Closes plural*/format* helpers + Default*
// + systemClock.Now + structural Start/Stop (mirror к v0.153.6 schedulers).
// No production change.

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	scheduleEntities "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
)

// ===== ReminderScheduler.pluralHours / pluralDays (russian plural forms) =====

func TestReminderScheduler_PluralHours(t *testing.T) {
	s := &ReminderScheduler{}
	cases := []struct {
		n    int
		want string
	}{
		{1, "час"},  // 1 → singular
		{2, "часа"}, // 2-4 → genitive singular
		{3, "часа"},
		{4, "часа"},
		{5, "часов"},  // 5-20 → genitive plural
		{11, "часов"}, // 11 special-case
		{21, "час"},   // 21 → 1-like ending
		{22, "часа"},  // 22 → 2-like ending
		{25, "часов"},
		{100, "часов"},
		{101, "час"},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, s.pluralHours(tc.n), "n=%d", tc.n)
	}
}

func TestReminderScheduler_PluralDays(t *testing.T) {
	s := &ReminderScheduler{}
	cases := []struct {
		n    int
		want string
	}{
		{1, "день"},
		{2, "дня"},
		{4, "дня"},
		{5, "дней"},
		{11, "дней"},
		{21, "день"},
		{22, "дня"},
		{25, "дней"},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, s.pluralDays(tc.n), "n=%d", tc.n)
	}
}

func TestReminderScheduler_FormatMinutesText(t *testing.T) {
	s := &ReminderScheduler{}
	cases := []struct {
		minutes int
		want    string
	}{
		{1, "1 минут"}, // < 60 → minutes branch
		{59, "59 минут"},
		{60, "1 час"},   // exactly 60 → special-case
		{120, "2 часа"}, // 61-1439 → hours
		{180, "3 часа"},
		{300, "5 часов"},
		{1440, "1 день"}, // ≥ 1440 → days
		{2880, "2 дня"},
		{7200, "5 дней"},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, s.formatMinutesText(tc.minutes), "minutes=%d", tc.minutes)
	}
}

// ===== ReminderScheduler.formatOptional =====

func TestReminderScheduler_FormatOptional(t *testing.T) {
	s := &ReminderScheduler{}
	t.Run("nil pointer returns empty string", func(t *testing.T) {
		assert.Equal(t, "", s.formatOptional("Место", nil))
	})
	t.Run("empty string pointer returns empty string", func(t *testing.T) {
		empty := ""
		assert.Equal(t, "", s.formatOptional("Место", &empty))
	})
	t.Run("non-empty value wraps в strong tag", func(t *testing.T) {
		v := "Аудитория 305"
		got := s.formatOptional("Место", &v)
		assert.Contains(t, got, "<strong>Место:</strong>")
		assert.Contains(t, got, "Аудитория 305")
	})
}

// ===== ReminderScheduler.formatEventReminderEmail / Message =====

func TestReminderScheduler_FormatEventReminderEmail(t *testing.T) {
	s := &ReminderScheduler{}
	location := "Аудитория 305"
	description := "Защита диплома"
	event := &scheduleEntities.Event{
		Title:       "Семинар",
		StartTime:   time.Date(2026, 5, 20, 10, 30, 0, 0, time.UTC),
		Location:    &location,
		Description: &description,
	}

	got := s.formatEventReminderEmail(event, 60)
	assert.Contains(t, got, "Напоминание о событии")
	assert.Contains(t, got, "Семинар")
	assert.Contains(t, got, "20.05.2026 10:30")
	assert.Contains(t, got, "Аудитория 305")
	assert.Contains(t, got, "Защита диплома")
	assert.Contains(t, got, "1 час")
}

func TestReminderScheduler_FormatEventReminderEmail_NilLocationAndDescription(t *testing.T) {
	// Covers formatOptional nil-path inside formatEventReminderEmail.
	s := &ReminderScheduler{}
	event := &scheduleEntities.Event{
		Title:     "Без места",
		StartTime: time.Date(2026, 5, 20, 9, 0, 0, 0, time.UTC),
	}
	got := s.formatEventReminderEmail(event, 15)
	assert.Contains(t, got, "Без места")
	assert.Contains(t, got, "15 минут")
	assert.NotContains(t, got, "<strong>Место:")
	assert.NotContains(t, got, "<strong>Описание:")
}

func TestReminderScheduler_FormatEventReminderMessage(t *testing.T) {
	s := &ReminderScheduler{}
	event := &scheduleEntities.Event{
		Title:     "Лекция",
		StartTime: time.Date(2026, 5, 20, 14, 0, 0, 0, time.UTC),
	}
	got := s.formatEventReminderMessage(event, 60)
	assert.Equal(t, "Через 1 час: Лекция (14:00)", got)
}

// ===== ReminderScheduler.DefaultConfig =====

func TestReminderScheduler_DefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	require.NotNil(t, cfg)
	assert.Equal(t, time.Minute, cfg.CheckInterval)
	assert.Equal(t, 100, cfg.BatchSize)
}

// ===== task_reminder_scheduler free-func plural*/formatMinutesText =====

func TestTaskRemFreeFuncs_PluralHours(t *testing.T) {
	cases := []struct {
		n    int
		want string
	}{
		{1, "час"},
		{2, "часа"},
		{5, "часов"},
		{11, "часов"},
		{21, "час"},
		{22, "часа"},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, pluralHours(tc.n), "n=%d", tc.n)
	}
}

func TestTaskRemFreeFuncs_PluralDays(t *testing.T) {
	cases := []struct {
		n    int
		want string
	}{
		{1, "день"},
		{3, "дня"},
		{7, "дней"},
		{11, "дней"},
		{21, "день"},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, pluralDays(tc.n), "n=%d", tc.n)
	}
}

func TestTaskRemFreeFuncs_FormatMinutesText(t *testing.T) {
	cases := []struct {
		minutes int
		want    string
	}{
		{1, "1 минут"},
		{59, "59 минут"},
		{60, "1 час"},
		{120, "2 часа"},
		{300, "5 часов"},
		{1440, "1 день"},
		{7200, "5 дней"},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, formatMinutesText(tc.minutes), "minutes=%d", tc.minutes)
	}
}

// ===== formatTaskReminderEmail =====

func TestFormatTaskReminderEmail_WithDueDate(t *testing.T) {
	due := time.Date(2026, 5, 25, 18, 0, 0, 0, time.UTC)
	view := &TaskDispatchView{
		Title:   "Проверить ВКР",
		DueDate: &due,
	}
	got := formatTaskReminderEmail(view, 1440)
	assert.Contains(t, got, "Напоминание о задаче")
	assert.Contains(t, got, "Проверить ВКР")
	assert.Contains(t, got, "Крайний срок:")
	assert.Contains(t, got, "25.05.2026 18:00")
	assert.Contains(t, got, "1 день")
}

func TestFormatTaskReminderEmail_NilDueDate(t *testing.T) {
	view := &TaskDispatchView{Title: "Без срока"}
	got := formatTaskReminderEmail(view, 60)
	assert.Contains(t, got, "Без срока")
	assert.Contains(t, got, "1 час")
	assert.NotContains(t, got, "Крайний срок:")
}

// ===== DefaultTaskReminderConfig + systemClock.Now =====

func TestDefaultTaskReminderConfig(t *testing.T) {
	cfg := DefaultTaskReminderConfig()
	require.NotNil(t, cfg)
	assert.Equal(t, time.Minute, cfg.CheckInterval)
	require.NotNil(t, cfg.Clock)
	// systemClock.Now() returns time.Now() — non-zero in any case.
	got := cfg.Clock.Now()
	assert.False(t, got.IsZero())
}

func TestSystemClock_Now(t *testing.T) {
	c := systemClock{}
	before := time.Now().Add(-time.Second)
	got := c.Now()
	after := time.Now().Add(time.Second)
	assert.True(t, got.After(before) && got.Before(after),
		"systemClock.Now() must return value within [before-1s, after+1s]")
}

// ===== Structural Start / Stop coverage (mirror v0.153.6 schedulers) =====

func TestReminderScheduler_StartStop(t *testing.T) {
	// Constructor accepts nil deps for happy-path (deps are dereferenced
	// only inside processPendingReminders / cleanupExpiredNotifications
	// — neither fires during the brief Start→Stop test window because
	// checkInterval=1m and cleanup is cron at 03:00).
	s, err := NewReminderScheduler(nil, nil, nil, &fakeNotificationRepo{}, nil, nil, nil)
	require.NoError(t, err)
	require.NoError(t, s.Start())
	require.NoError(t, s.Stop())
}

func TestReminderScheduler_CleanupExpiredNotifications(t *testing.T) {
	// Covers cleanupExpiredNotifications happy path (count=0, no logs).
	s, err := NewReminderScheduler(nil, nil, nil, &fakeNotificationRepo{}, nil, nil, nil)
	require.NoError(t, err)
	s.cleanupExpiredNotifications()
}

func TestTaskReminderScheduler_StartStop(t *testing.T) {
	// All required deps must be non-nil per constructor guards
	// (reminderRepo / taskLookup / notificationRepo / preferencesRepo).
	prefs := enabledPrefs()
	s, err := NewTaskReminderScheduler(
		&fakeReminderRepo{},
		&fakeTaskLookup{},
		nil, nil,
		&fakeNotificationRepo{},
		&fakePreferencesRepo{prefs: prefs},
		nil, nil, nil,
	)
	require.NoError(t, err)
	require.NoError(t, s.Start())
	require.NoError(t, s.Stop())
}
