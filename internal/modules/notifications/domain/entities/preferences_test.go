package entities

import (
	"testing"
	"time"
)

const (
	testQuietStart  = "22:00"
	testQuietEnd    = "07:00"
	testTimezoneUTC = "UTC"
)

func TestNewUserNotificationPreferences(t *testing.T) {
	userID := int64(42)

	p := NewUserNotificationPreferences(userID)

	if p.UserID != userID {
		t.Errorf("expected user ID %d, got %d", userID, p.UserID)
	}
	if !p.EmailEnabled {
		t.Error("expected EmailEnabled to be true")
	}
	if !p.PushEnabled {
		t.Error("expected PushEnabled to be true")
	}
	if !p.InAppEnabled {
		t.Error("expected InAppEnabled to be true")
	}
	if p.TelegramEnabled {
		t.Error("expected TelegramEnabled to be false")
	}
	if p.SlackEnabled {
		t.Error("expected SlackEnabled to be false")
	}
	if p.QuietHoursEnabled {
		t.Error("expected QuietHoursEnabled to be false")
	}
	if p.QuietHoursStart != testQuietStart {
		t.Errorf("expected QuietHoursStart '22:00', got '%s'", p.QuietHoursStart)
	}
	if p.QuietHoursEnd != testQuietEnd {
		t.Errorf("expected QuietHoursEnd '07:00', got '%s'", p.QuietHoursEnd)
	}
	if p.Timezone != "Europe/Moscow" {
		t.Errorf("expected Timezone 'Europe/Moscow', got '%s'", p.Timezone)
	}
	if p.DigestEnabled {
		t.Error("expected DigestEnabled to be false")
	}
	if p.DigestFrequency != "daily" {
		t.Errorf("expected DigestFrequency 'daily', got '%s'", p.DigestFrequency)
	}
	if p.DigestTime != "09:00" {
		t.Errorf("expected DigestTime '09:00', got '%s'", p.DigestTime)
	}
	if p.TypePreferences == nil {
		t.Error("expected TypePreferences to be initialized")
	}
	if p.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
	if p.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestUserNotificationPreferences_IsChannelEnabled(t *testing.T) {
	p := NewUserNotificationPreferences(1)

	tests := []struct {
		name     string
		channel  NotificationChannel
		expected bool
	}{
		{"email enabled", ChannelEmail, true},
		{"push enabled", ChannelPush, true},
		{"in_app enabled", ChannelInApp, true},
		{"telegram disabled", ChannelTelegram, false},
		{"slack disabled", ChannelSlack, false},
		{"unknown channel", NotificationChannel("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := p.IsChannelEnabled(tt.channel); got != tt.expected {
				t.Errorf("IsChannelEnabled(%s) = %v, want %v", tt.channel, got, tt.expected)
			}
		})
	}
}

func TestUserNotificationPreferences_IsChannelEnabled_Modified(t *testing.T) {
	p := NewUserNotificationPreferences(1)
	p.EmailEnabled = false
	p.TelegramEnabled = true

	if p.IsChannelEnabled(ChannelEmail) {
		t.Error("expected email to be disabled")
	}
	if !p.IsChannelEnabled(ChannelTelegram) {
		t.Error("expected telegram to be enabled")
	}
}

func TestUserNotificationPreferences_IsWithinQuietHours_Disabled(t *testing.T) {
	p := NewUserNotificationPreferences(1)
	p.QuietHoursEnabled = false

	currentTime := time.Date(2024, 1, 1, 23, 0, 0, 0, time.UTC)
	if p.IsWithinQuietHours(currentTime) {
		t.Error("expected quiet hours to not apply when disabled")
	}
}

func TestUserNotificationPreferences_IsWithinQuietHours_OvernightWithinStart(t *testing.T) {
	p := NewUserNotificationPreferences(1)
	p.QuietHoursEnabled = true
	p.QuietHoursStart = testQuietStart
	p.QuietHoursEnd = testQuietEnd
	p.Timezone = testTimezoneUTC

	// 23:00 should be within quiet hours (after start)
	currentTime := time.Date(2024, 1, 1, 23, 0, 0, 0, time.UTC)
	if !p.IsWithinQuietHours(currentTime) {
		t.Error("expected 23:00 to be within quiet hours (22:00-07:00)")
	}
}

func TestUserNotificationPreferences_IsWithinQuietHours_OvernightWithinEnd(t *testing.T) {
	p := NewUserNotificationPreferences(1)
	p.QuietHoursEnabled = true
	p.QuietHoursStart = testQuietStart
	p.QuietHoursEnd = testQuietEnd
	p.Timezone = testTimezoneUTC

	// 06:00 should be within quiet hours (before end)
	currentTime := time.Date(2024, 1, 1, 6, 0, 0, 0, time.UTC)
	if !p.IsWithinQuietHours(currentTime) {
		t.Error("expected 06:00 to be within quiet hours (22:00-07:00)")
	}
}

func TestUserNotificationPreferences_IsWithinQuietHours_OvernightOutside(t *testing.T) {
	p := NewUserNotificationPreferences(1)
	p.QuietHoursEnabled = true
	p.QuietHoursStart = testQuietStart
	p.QuietHoursEnd = testQuietEnd
	p.Timezone = testTimezoneUTC

	// 12:00 should be outside quiet hours
	currentTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	if p.IsWithinQuietHours(currentTime) {
		t.Error("expected 12:00 to be outside quiet hours (22:00-07:00)")
	}
}

func TestUserNotificationPreferences_IsWithinQuietHours_SameDay(t *testing.T) {
	p := NewUserNotificationPreferences(1)
	p.QuietHoursEnabled = true
	p.QuietHoursStart = "09:00"
	p.QuietHoursEnd = "17:00"
	p.Timezone = testTimezoneUTC

	tests := []struct {
		name     string
		hour     int
		expected bool
	}{
		{"08:00 outside", 8, false},
		{"09:00 within", 9, true},
		{"12:00 within", 12, true},
		{"16:59 within", 16, true},
		{"17:00 outside", 17, false},
		{"18:00 outside", 18, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			currentTime := time.Date(2024, 1, 1, tt.hour, 0, 0, 0, time.UTC)
			if got := p.IsWithinQuietHours(currentTime); got != tt.expected {
				t.Errorf("IsWithinQuietHours at %02d:00 = %v, want %v", tt.hour, got, tt.expected)
			}
		})
	}
}

func TestUserNotificationPreferences_IsWithinQuietHours_InvalidTimezone(t *testing.T) {
	p := NewUserNotificationPreferences(1)
	p.QuietHoursEnabled = true
	p.QuietHoursStart = testQuietStart
	p.QuietHoursEnd = "23:00"
	p.Timezone = "Invalid/Timezone"

	// Should fall back to UTC
	currentTime := time.Date(2024, 1, 1, 22, 30, 0, 0, time.UTC)
	if !p.IsWithinQuietHours(currentTime) {
		t.Error("expected quiet hours to work with invalid timezone (fallback to UTC)")
	}
}

func TestUserNotificationPreferences_ShouldNotify_ChannelDisabled(t *testing.T) {
	p := NewUserNotificationPreferences(1)
	p.TelegramEnabled = false

	currentTime := time.Now()
	if p.ShouldNotify(ChannelTelegram, NotificationTypeInfo, currentTime) {
		t.Error("should not notify when channel is disabled")
	}
}

func TestUserNotificationPreferences_ShouldNotify_WithinQuietHours(t *testing.T) {
	p := NewUserNotificationPreferences(1)
	p.QuietHoursEnabled = true
	p.QuietHoursStart = testQuietStart
	p.QuietHoursEnd = testQuietEnd
	p.Timezone = testTimezoneUTC

	// 23:00 is within quiet hours
	currentTime := time.Date(2024, 1, 1, 23, 0, 0, 0, time.UTC)
	if p.ShouldNotify(ChannelEmail, NotificationTypeInfo, currentTime) {
		t.Error("should not notify during quiet hours")
	}
}

func TestUserNotificationPreferences_ShouldNotify_TypeDisabled(t *testing.T) {
	p := NewUserNotificationPreferences(1)
	p.TypePreferences[NotificationTypeInfo] = TypePreference{
		Enabled:  false,
		Channels: []string{},
	}

	currentTime := time.Now()
	if p.ShouldNotify(ChannelEmail, NotificationTypeInfo, currentTime) {
		t.Error("should not notify when type is disabled")
	}
}

func TestUserNotificationPreferences_ShouldNotify_ChannelNotAllowedForType(t *testing.T) {
	p := NewUserNotificationPreferences(1)
	p.TypePreferences[NotificationTypeInfo] = TypePreference{
		Enabled:  true,
		Channels: []string{string(ChannelPush)}, // Only push allowed
	}

	currentTime := time.Now()
	if p.ShouldNotify(ChannelEmail, NotificationTypeInfo, currentTime) {
		t.Error("should not notify via email when only push is allowed for type")
	}
}

func TestUserNotificationPreferences_ShouldNotify_Success(t *testing.T) {
	p := NewUserNotificationPreferences(1)
	p.TypePreferences[NotificationTypeInfo] = TypePreference{
		Enabled:  true,
		Channels: []string{string(ChannelEmail)},
	}

	currentTime := time.Now()
	if !p.ShouldNotify(ChannelEmail, NotificationTypeInfo, currentTime) {
		t.Error("should notify when all conditions are met")
	}
}

func TestUserNotificationPreferences_ShouldNotify_NoTypePreference(t *testing.T) {
	p := NewUserNotificationPreferences(1)
	// No type preference set

	currentTime := time.Now()
	if !p.ShouldNotify(ChannelEmail, NotificationTypeInfo, currentTime) {
		t.Error("should notify when no type preference is set")
	}
}

func TestQuietHoursStruct(t *testing.T) {
	qh := QuietHours{
		Enabled:   true,
		StartTime: testQuietStart,
		EndTime:   testQuietEnd,
		Timezone:  "Europe/Moscow",
	}

	if !qh.Enabled {
		t.Error("expected enabled to be true")
	}
	if qh.StartTime != testQuietStart {
		t.Errorf("expected start time '22:00', got '%s'", qh.StartTime)
	}
	if qh.EndTime != testQuietEnd {
		t.Errorf("expected end time '07:00', got '%s'", qh.EndTime)
	}
	if qh.Timezone != "Europe/Moscow" {
		t.Errorf("expected timezone 'Europe/Moscow', got '%s'", qh.Timezone)
	}
}

func TestChannelPreferencesStruct(t *testing.T) {
	cp := ChannelPreferences{
		Enabled:      true,
		Types:        []string{"info", "warning"},
		ExcludeTypes: []string{"system"},
	}

	if !cp.Enabled {
		t.Error("expected enabled to be true")
	}
	if len(cp.Types) != 2 {
		t.Errorf("expected 2 types, got %d", len(cp.Types))
	}
	if len(cp.ExcludeTypes) != 1 {
		t.Errorf("expected 1 exclude type, got %d", len(cp.ExcludeTypes))
	}
}

func TestTypePreferenceStruct(t *testing.T) {
	tp := TypePreference{
		Enabled:  true,
		Channels: []string{"email", "push"},
		Priority: "high",
	}

	if !tp.Enabled {
		t.Error("expected enabled to be true")
	}
	if len(tp.Channels) != 2 {
		t.Errorf("expected 2 channels, got %d", len(tp.Channels))
	}
	if tp.Priority != "high" {
		t.Errorf("expected priority 'high', got '%s'", tp.Priority)
	}
}
