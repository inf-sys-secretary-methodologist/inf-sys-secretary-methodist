// Package usecases contains application use cases for the notifications module.
package usecases

import (
	"context"
	"fmt"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/repositories"
)

// PreferencesUseCase handles notification preferences operations
type PreferencesUseCase struct {
	preferencesRepo repositories.PreferencesRepository
}

// NewPreferencesUseCase creates a new preferences use case
func NewPreferencesUseCase(preferencesRepo repositories.PreferencesRepository) *PreferencesUseCase {
	return &PreferencesUseCase{
		preferencesRepo: preferencesRepo,
	}
}

// Get retrieves notification preferences for a user
func (uc *PreferencesUseCase) Get(ctx context.Context, userID int64) (*dto.PreferencesOutput, error) {
	preferences, err := uc.preferencesRepo.GetOrCreate(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get preferences: %w", err)
	}

	return dto.ToPreferencesOutput(preferences), nil
}

// Update updates notification preferences
func (uc *PreferencesUseCase) Update(ctx context.Context, userID int64, input *dto.PreferencesInput) (*dto.PreferencesOutput, error) {
	preferences, err := uc.preferencesRepo.GetOrCreate(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get preferences: %w", err)
	}

	// Apply input to existing preferences
	input.ApplyToEntity(preferences)

	if err := uc.preferencesRepo.Update(ctx, preferences); err != nil {
		return nil, fmt.Errorf("failed to update preferences: %w", err)
	}

	return dto.ToPreferencesOutput(preferences), nil
}

// ToggleChannel enables or disables a notification channel
func (uc *PreferencesUseCase) ToggleChannel(ctx context.Context, userID int64, input *dto.ChannelToggleInput) (*dto.PreferencesOutput, error) {
	channel := entities.NotificationChannel(input.Channel)

	if err := uc.preferencesRepo.UpdateChannelEnabled(ctx, userID, channel, input.Enabled); err != nil {
		return nil, fmt.Errorf("failed to toggle channel: %w", err)
	}

	return uc.Get(ctx, userID)
}

// UpdateQuietHours updates quiet hours settings
func (uc *PreferencesUseCase) UpdateQuietHours(ctx context.Context, userID int64, input *dto.QuietHoursInput) (*dto.PreferencesOutput, error) {
	if err := uc.preferencesRepo.UpdateQuietHours(ctx, userID, input.Enabled, input.StartTime, input.EndTime, input.Timezone); err != nil {
		return nil, fmt.Errorf("failed to update quiet hours: %w", err)
	}

	return uc.Get(ctx, userID)
}

// Reset resets preferences to defaults
func (uc *PreferencesUseCase) Reset(ctx context.Context, userID int64) (*dto.PreferencesOutput, error) {
	// Delete existing preferences
	_ = uc.preferencesRepo.Delete(ctx, userID)

	// Get or create will create new default preferences
	preferences, err := uc.preferencesRepo.GetOrCreate(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to reset preferences: %w", err)
	}

	return dto.ToPreferencesOutput(preferences), nil
}

// EnableChannel enables a specific notification channel
func (uc *PreferencesUseCase) EnableChannel(ctx context.Context, userID int64, channel string) error {
	ch := entities.NotificationChannel(channel)
	return uc.preferencesRepo.UpdateChannelEnabled(ctx, userID, ch, true)
}

// DisableChannel disables a specific notification channel
func (uc *PreferencesUseCase) DisableChannel(ctx context.Context, userID int64, channel string) error {
	ch := entities.NotificationChannel(channel)
	return uc.preferencesRepo.UpdateChannelEnabled(ctx, userID, ch, false)
}

// GetAvailableTimezones returns a list of available timezones
func (uc *PreferencesUseCase) GetAvailableTimezones() []string {
	return []string{
		"Europe/Moscow",
		"Europe/Kaliningrad",
		"Europe/Samara",
		"Asia/Yekaterinburg",
		"Asia/Omsk",
		"Asia/Krasnoyarsk",
		"Asia/Irkutsk",
		"Asia/Yakutsk",
		"Asia/Vladivostok",
		"Asia/Magadan",
		"Asia/Kamchatka",
		"UTC",
	}
}
