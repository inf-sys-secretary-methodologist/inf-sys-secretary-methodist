package dto

import (
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToPreferencesOutput(t *testing.T) {
	now := time.Now()
	prefs := &entities.UserNotificationPreferences{
		ID:                1,
		UserID:            42,
		EmailEnabled:      true,
		PushEnabled:       false,
		InAppEnabled:      true,
		TelegramEnabled:   false,
		SlackEnabled:      false,
		QuietHoursEnabled: true,
		QuietHoursStart:   "22:00",
		QuietHoursEnd:     "07:00",
		Timezone:          "Europe/Moscow",
		DigestEnabled:     true,
		DigestFrequency:   "daily",
		DigestTime:        "09:00",
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	output := ToPreferencesOutput(prefs)

	require.NotNil(t, output)
	assert.Equal(t, int64(1), output.ID)
	assert.Equal(t, int64(42), output.UserID)
	assert.True(t, output.EmailEnabled)
	assert.False(t, output.PushEnabled)
	assert.True(t, output.InAppEnabled)
	assert.True(t, output.QuietHoursEnabled)
	assert.Equal(t, "22:00", output.QuietHoursStart)
	assert.Equal(t, "07:00", output.QuietHoursEnd)
	assert.Equal(t, "Europe/Moscow", output.Timezone)
	assert.True(t, output.DigestEnabled)
	assert.Equal(t, "daily", output.DigestFrequency)
	assert.Equal(t, "09:00", output.DigestTime)
}

func TestPreferencesInput_ApplyToEntity(t *testing.T) {
	p := &entities.UserNotificationPreferences{
		EmailEnabled: true,
		PushEnabled:  true,
		InAppEnabled: true,
		Timezone:     "UTC",
	}

	emailOff := false
	pushOn := true
	input := &PreferencesInput{
		EmailEnabled:    &emailOff,
		PushEnabled:     &pushOn,
		QuietHoursStart: "23:00",
		QuietHoursEnd:   "06:00",
		Timezone:        "Europe/Berlin",
		DigestFrequency: "weekly",
		DigestTime:      "08:00",
	}

	input.ApplyToEntity(p)

	assert.False(t, p.EmailEnabled)
	assert.True(t, p.PushEnabled)
	assert.True(t, p.InAppEnabled) // unchanged
	assert.Equal(t, "23:00", p.QuietHoursStart)
	assert.Equal(t, "06:00", p.QuietHoursEnd)
	assert.Equal(t, "Europe/Berlin", p.Timezone)
	assert.Equal(t, "weekly", p.DigestFrequency)
	assert.Equal(t, "08:00", p.DigestTime)
}

func TestPreferencesInput_ApplyToEntity_PartialUpdate(t *testing.T) {
	p := &entities.UserNotificationPreferences{
		EmailEnabled: true,
		PushEnabled:  true,
		Timezone:     "UTC",
	}

	// Only update telegram
	telegramOn := true
	input := &PreferencesInput{
		TelegramEnabled: &telegramOn,
	}

	input.ApplyToEntity(p)

	assert.True(t, p.EmailEnabled)  // unchanged
	assert.True(t, p.PushEnabled)   // unchanged
	assert.True(t, p.TelegramEnabled)
	assert.Equal(t, "UTC", p.Timezone) // unchanged
}

func TestPreferencesInput_ApplyToEntity_TypePreferences(t *testing.T) {
	p := &entities.UserNotificationPreferences{}

	input := &PreferencesInput{
		TypePreferences: map[entities.NotificationType]entities.TypePreference{
			entities.NotificationTypeTask: {
				Enabled:  true,
				Channels: []string{"email", "push"},
				Priority: "high",
			},
		},
	}

	input.ApplyToEntity(p)

	require.NotNil(t, p.TypePreferences)
	pref, ok := p.TypePreferences[entities.NotificationTypeTask]
	require.True(t, ok)
	assert.True(t, pref.Enabled)
	assert.Equal(t, []string{"email", "push"}, pref.Channels)
}
