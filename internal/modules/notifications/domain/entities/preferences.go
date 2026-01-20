// Package entities contains domain entities for the notifications module.
package entities

import (
	"time"
)

// QuietHours represents the time range when notifications should not be sent
type QuietHours struct {
	Enabled   bool   `json:"enabled"`
	StartTime string `json:"start_time"` // HH:MM format, e.g., "22:00"
	EndTime   string `json:"end_time"`   // HH:MM format, e.g., "07:00"
	Timezone  string `json:"timezone"`   // e.g., "Europe/Moscow"
}

// ChannelPreferences represents preferences for a specific channel
type ChannelPreferences struct {
	Enabled      bool     `json:"enabled"`
	Types        []string `json:"types,omitempty"`         // notification types to receive on this channel
	ExcludeTypes []string `json:"exclude_types,omitempty"` // notification types to exclude
}

// UserNotificationPreferences represents user notification preferences
type UserNotificationPreferences struct {
	ID     int64 `json:"id"`
	UserID int64 `json:"user_id"`

	// Channel preferences
	EmailEnabled    bool `json:"email_enabled"`
	PushEnabled     bool `json:"push_enabled"`
	InAppEnabled    bool `json:"in_app_enabled"`
	TelegramEnabled bool `json:"telegram_enabled"`
	SlackEnabled    bool `json:"slack_enabled"`

	// Quiet hours
	QuietHoursEnabled bool   `json:"quiet_hours_enabled"`
	QuietHoursStart   string `json:"quiet_hours_start"` // HH:MM format
	QuietHoursEnd     string `json:"quiet_hours_end"`   // HH:MM format
	Timezone          string `json:"timezone"`

	// Digest settings
	DigestEnabled   bool   `json:"digest_enabled"`
	DigestFrequency string `json:"digest_frequency"` // "daily", "weekly"
	DigestTime      string `json:"digest_time"`      // HH:MM format

	// Notification type preferences (stored as JSONB)
	TypePreferences map[NotificationType]TypePreference `json:"type_preferences,omitempty"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TypePreference represents preferences for a specific notification type
type TypePreference struct {
	Enabled  bool     `json:"enabled"`
	Channels []string `json:"channels"` // which channels to use for this type
	Priority string   `json:"priority"` // minimum priority to notify
}

// NewUserNotificationPreferences creates default preferences for a user
func NewUserNotificationPreferences(userID int64) *UserNotificationPreferences {
	now := time.Now()
	return &UserNotificationPreferences{
		UserID:            userID,
		EmailEnabled:      true,
		PushEnabled:       true,
		InAppEnabled:      true,
		TelegramEnabled:   false,
		SlackEnabled:      false,
		QuietHoursEnabled: false,
		QuietHoursStart:   "22:00",
		QuietHoursEnd:     "07:00",
		Timezone:          "Europe/Moscow",
		DigestEnabled:     false,
		DigestFrequency:   "daily",
		DigestTime:        "09:00",
		TypePreferences:   make(map[NotificationType]TypePreference),
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

// IsChannelEnabled checks if a specific channel is enabled
func (p *UserNotificationPreferences) IsChannelEnabled(channel NotificationChannel) bool {
	switch channel {
	case ChannelEmail:
		return p.EmailEnabled
	case ChannelPush:
		return p.PushEnabled
	case ChannelInApp:
		return p.InAppEnabled
	case ChannelTelegram:
		return p.TelegramEnabled
	case ChannelSlack:
		return p.SlackEnabled
	default:
		return false
	}
}

// IsWithinQuietHours checks if the current time is within quiet hours
func (p *UserNotificationPreferences) IsWithinQuietHours(currentTime time.Time) bool {
	if !p.QuietHoursEnabled {
		return false
	}

	// Parse quiet hours
	loc, err := time.LoadLocation(p.Timezone)
	if err != nil {
		loc = time.UTC
	}
	localTime := currentTime.In(loc)

	// Parse start and end times
	startHour, startMin := parseTime(p.QuietHoursStart)
	endHour, endMin := parseTime(p.QuietHoursEnd)

	currentHour := localTime.Hour()
	currentMin := localTime.Minute()

	currentMinutes := currentHour*60 + currentMin
	startMinutes := startHour*60 + startMin
	endMinutes := endHour*60 + endMin

	// Handle overnight quiet hours (e.g., 22:00 - 07:00)
	if startMinutes > endMinutes {
		return currentMinutes >= startMinutes || currentMinutes < endMinutes
	}

	return currentMinutes >= startMinutes && currentMinutes < endMinutes
}

// ShouldNotify checks if a notification should be sent based on preferences
func (p *UserNotificationPreferences) ShouldNotify(channel NotificationChannel, notificationType NotificationType, currentTime time.Time) bool {
	// Check if channel is enabled
	if !p.IsChannelEnabled(channel) {
		return false
	}

	// Check quiet hours (except for urgent notifications)
	if p.IsWithinQuietHours(currentTime) {
		return false
	}

	// Check type-specific preferences
	if typePref, ok := p.TypePreferences[notificationType]; ok {
		if !typePref.Enabled {
			return false
		}
		// Check if this channel is allowed for this type
		channelAllowed := false
		for _, ch := range typePref.Channels {
			if ch == string(channel) {
				channelAllowed = true
				break
			}
		}
		if len(typePref.Channels) > 0 && !channelAllowed {
			return false
		}
	}

	return true
}

// parseTime parses a time string in HH:MM format
func parseTime(timeStr string) (hour, min int) {
	var h, m int
	n, _ := time.Parse("15:04", timeStr)
	h = n.Hour()
	m = n.Minute()
	return h, m
}
