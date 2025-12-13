// Package dto contains data transfer objects for the notifications module.
package dto

import (
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/entities"
)

// PreferencesInput represents input for updating notification preferences
type PreferencesInput struct {
	EmailEnabled      *bool                                                 `json:"email_enabled,omitempty"`
	PushEnabled       *bool                                                 `json:"push_enabled,omitempty"`
	InAppEnabled      *bool                                                 `json:"in_app_enabled,omitempty"`
	TelegramEnabled   *bool                                                 `json:"telegram_enabled,omitempty"`
	SlackEnabled      *bool                                                 `json:"slack_enabled,omitempty"`
	QuietHoursEnabled *bool                                                 `json:"quiet_hours_enabled,omitempty"`
	QuietHoursStart   string                                                `json:"quiet_hours_start,omitempty" validate:"omitempty,len=5"`
	QuietHoursEnd     string                                                `json:"quiet_hours_end,omitempty" validate:"omitempty,len=5"`
	Timezone          string                                                `json:"timezone,omitempty" validate:"omitempty,max=50"`
	DigestEnabled     *bool                                                 `json:"digest_enabled,omitempty"`
	DigestFrequency   string                                                `json:"digest_frequency,omitempty" validate:"omitempty,oneof=daily weekly"`
	DigestTime        string                                                `json:"digest_time,omitempty" validate:"omitempty,len=5"`
	TypePreferences   map[entities.NotificationType]entities.TypePreference `json:"type_preferences,omitempty"`
}

// PreferencesOutput represents notification preferences in API response
type PreferencesOutput struct {
	ID                int64                                                 `json:"id"`
	UserID            int64                                                 `json:"user_id"`
	EmailEnabled      bool                                                  `json:"email_enabled"`
	PushEnabled       bool                                                  `json:"push_enabled"`
	InAppEnabled      bool                                                  `json:"in_app_enabled"`
	TelegramEnabled   bool                                                  `json:"telegram_enabled"`
	SlackEnabled      bool                                                  `json:"slack_enabled"`
	QuietHoursEnabled bool                                                  `json:"quiet_hours_enabled"`
	QuietHoursStart   string                                                `json:"quiet_hours_start"`
	QuietHoursEnd     string                                                `json:"quiet_hours_end"`
	Timezone          string                                                `json:"timezone"`
	DigestEnabled     bool                                                  `json:"digest_enabled"`
	DigestFrequency   string                                                `json:"digest_frequency"`
	DigestTime        string                                                `json:"digest_time"`
	TypePreferences   map[entities.NotificationType]entities.TypePreference `json:"type_preferences,omitempty"`
	TelegramConnected bool                                                  `json:"telegram_connected"`
	SlackConnected    bool                                                  `json:"slack_connected"`
	CreatedAt         time.Time                                             `json:"created_at"`
	UpdatedAt         time.Time                                             `json:"updated_at"`
}

// ChannelToggleInput represents input for toggling a channel
type ChannelToggleInput struct {
	Channel string `json:"channel" validate:"required,oneof=email push in_app telegram slack"`
	Enabled bool   `json:"enabled"`
}

// QuietHoursInput represents input for updating quiet hours
type QuietHoursInput struct {
	Enabled   bool   `json:"enabled"`
	StartTime string `json:"start_time" validate:"required,len=5"`
	EndTime   string `json:"end_time" validate:"required,len=5"`
	Timezone  string `json:"timezone" validate:"required,max=50"`
}

// TelegramConnectionOutput represents Telegram connection info
type TelegramConnectionOutput struct {
	Connected        bool       `json:"connected"`
	TelegramUsername string     `json:"telegram_username,omitempty"`
	TelegramChatID   int64      `json:"telegram_chat_id,omitempty"`
	ConnectedAt      *time.Time `json:"connected_at,omitempty"`
}

// SlackConnectionOutput represents Slack connection info
type SlackConnectionOutput struct {
	Connected        bool       `json:"connected"`
	SlackUsername    string     `json:"slack_username,omitempty"`
	SlackChannelID   string     `json:"slack_channel_id,omitempty"`
	SlackWorkspaceID string     `json:"slack_workspace_id,omitempty"`
	ConnectedAt      *time.Time `json:"connected_at,omitempty"`
}

// ToPreferencesOutput converts entity to output DTO
func ToPreferencesOutput(p *entities.UserNotificationPreferences) *PreferencesOutput {
	return &PreferencesOutput{
		ID:                p.ID,
		UserID:            p.UserID,
		EmailEnabled:      p.EmailEnabled,
		PushEnabled:       p.PushEnabled,
		InAppEnabled:      p.InAppEnabled,
		TelegramEnabled:   p.TelegramEnabled,
		SlackEnabled:      p.SlackEnabled,
		QuietHoursEnabled: p.QuietHoursEnabled,
		QuietHoursStart:   p.QuietHoursStart,
		QuietHoursEnd:     p.QuietHoursEnd,
		Timezone:          p.Timezone,
		DigestEnabled:     p.DigestEnabled,
		DigestFrequency:   p.DigestFrequency,
		DigestTime:        p.DigestTime,
		TypePreferences:   p.TypePreferences,
		CreatedAt:         p.CreatedAt,
		UpdatedAt:         p.UpdatedAt,
	}
}

// ApplyToEntity applies input to existing preferences entity
func (input *PreferencesInput) ApplyToEntity(p *entities.UserNotificationPreferences) {
	if input.EmailEnabled != nil {
		p.EmailEnabled = *input.EmailEnabled
	}
	if input.PushEnabled != nil {
		p.PushEnabled = *input.PushEnabled
	}
	if input.InAppEnabled != nil {
		p.InAppEnabled = *input.InAppEnabled
	}
	if input.TelegramEnabled != nil {
		p.TelegramEnabled = *input.TelegramEnabled
	}
	if input.SlackEnabled != nil {
		p.SlackEnabled = *input.SlackEnabled
	}
	if input.QuietHoursEnabled != nil {
		p.QuietHoursEnabled = *input.QuietHoursEnabled
	}
	if input.QuietHoursStart != "" {
		p.QuietHoursStart = input.QuietHoursStart
	}
	if input.QuietHoursEnd != "" {
		p.QuietHoursEnd = input.QuietHoursEnd
	}
	if input.Timezone != "" {
		p.Timezone = input.Timezone
	}
	if input.DigestEnabled != nil {
		p.DigestEnabled = *input.DigestEnabled
	}
	if input.DigestFrequency != "" {
		p.DigestFrequency = input.DigestFrequency
	}
	if input.DigestTime != "" {
		p.DigestTime = input.DigestTime
	}
	if input.TypePreferences != nil {
		p.TypePreferences = input.TypePreferences
	}
}
