package usecases

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/entities"
)

func TestPreferencesUseCase_Get(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully gets preferences", func(t *testing.T) {
		mockPrefsRepo := new(MockPreferencesRepository)

		uc := NewPreferencesUseCase(mockPrefsRepo)

		now := time.Now()
		prefs := &entities.UserNotificationPreferences{
			ID:              1,
			UserID:          1,
			EmailEnabled:    true,
			PushEnabled:     true,
			InAppEnabled:    true,
			TelegramEnabled: false,
			SlackEnabled:    false,
			Timezone:        "Europe/Moscow",
			CreatedAt:       now,
			UpdatedAt:       now,
		}

		mockPrefsRepo.On("GetOrCreate", ctx, int64(1)).Return(prefs, nil)

		output, err := uc.Get(ctx, 1)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, int64(1), output.UserID)
		assert.True(t, output.EmailEnabled)
		assert.Equal(t, "Europe/Moscow", output.Timezone)
		mockPrefsRepo.AssertExpectations(t)
	})

	t.Run("returns error when repository fails", func(t *testing.T) {
		mockPrefsRepo := new(MockPreferencesRepository)

		uc := NewPreferencesUseCase(mockPrefsRepo)

		mockPrefsRepo.On("GetOrCreate", ctx, int64(1)).Return(nil, assert.AnError)

		output, err := uc.Get(ctx, 1)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get preferences")
		assert.Nil(t, output)
		mockPrefsRepo.AssertExpectations(t)
	})
}

func TestPreferencesUseCase_Update(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully updates preferences", func(t *testing.T) {
		mockPrefsRepo := new(MockPreferencesRepository)

		uc := NewPreferencesUseCase(mockPrefsRepo)

		now := time.Now()
		existingPrefs := &entities.UserNotificationPreferences{
			ID:           1,
			UserID:       1,
			EmailEnabled: true,
			PushEnabled:  false,
			InAppEnabled: true,
			Timezone:     "UTC",
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		emailEnabled := false
		pushEnabled := true
		input := &dto.PreferencesInput{
			EmailEnabled: &emailEnabled,
			PushEnabled:  &pushEnabled,
			Timezone:     "Europe/Moscow",
		}

		mockPrefsRepo.On("GetOrCreate", ctx, int64(1)).Return(existingPrefs, nil)
		mockPrefsRepo.On("Update", ctx, mock.AnythingOfType("*entities.UserNotificationPreferences")).Return(nil)

		output, err := uc.Update(ctx, 1, input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.False(t, output.EmailEnabled)
		assert.True(t, output.PushEnabled)
		assert.Equal(t, "Europe/Moscow", output.Timezone)
		mockPrefsRepo.AssertExpectations(t)
	})

	t.Run("returns error when get fails", func(t *testing.T) {
		mockPrefsRepo := new(MockPreferencesRepository)

		uc := NewPreferencesUseCase(mockPrefsRepo)

		input := &dto.PreferencesInput{}

		mockPrefsRepo.On("GetOrCreate", ctx, int64(1)).Return(nil, assert.AnError)

		output, err := uc.Update(ctx, 1, input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get preferences")
		assert.Nil(t, output)
		mockPrefsRepo.AssertExpectations(t)
	})

	t.Run("returns error when update fails", func(t *testing.T) {
		mockPrefsRepo := new(MockPreferencesRepository)

		uc := NewPreferencesUseCase(mockPrefsRepo)

		now := time.Now()
		existingPrefs := &entities.UserNotificationPreferences{
			ID:        1,
			UserID:    1,
			CreatedAt: now,
			UpdatedAt: now,
		}

		input := &dto.PreferencesInput{}

		mockPrefsRepo.On("GetOrCreate", ctx, int64(1)).Return(existingPrefs, nil)
		mockPrefsRepo.On("Update", ctx, mock.AnythingOfType("*entities.UserNotificationPreferences")).Return(assert.AnError)

		output, err := uc.Update(ctx, 1, input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update preferences")
		assert.Nil(t, output)
		mockPrefsRepo.AssertExpectations(t)
	})
}

func TestPreferencesUseCase_ToggleChannel(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully toggles channel on", func(t *testing.T) {
		mockPrefsRepo := new(MockPreferencesRepository)

		uc := NewPreferencesUseCase(mockPrefsRepo)

		input := &dto.ChannelToggleInput{
			Channel: "telegram",
			Enabled: true,
		}

		now := time.Now()
		updatedPrefs := &entities.UserNotificationPreferences{
			ID:              1,
			UserID:          1,
			TelegramEnabled: true,
			CreatedAt:       now,
			UpdatedAt:       now,
		}

		mockPrefsRepo.On("UpdateChannelEnabled", ctx, int64(1), entities.NotificationChannel("telegram"), true).Return(nil)
		mockPrefsRepo.On("GetOrCreate", ctx, int64(1)).Return(updatedPrefs, nil)

		output, err := uc.ToggleChannel(ctx, 1, input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.True(t, output.TelegramEnabled)
		mockPrefsRepo.AssertExpectations(t)
	})

	t.Run("successfully toggles channel off", func(t *testing.T) {
		mockPrefsRepo := new(MockPreferencesRepository)

		uc := NewPreferencesUseCase(mockPrefsRepo)

		input := &dto.ChannelToggleInput{
			Channel: "email",
			Enabled: false,
		}

		now := time.Now()
		updatedPrefs := &entities.UserNotificationPreferences{
			ID:           1,
			UserID:       1,
			EmailEnabled: false,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		mockPrefsRepo.On("UpdateChannelEnabled", ctx, int64(1), entities.NotificationChannel("email"), false).Return(nil)
		mockPrefsRepo.On("GetOrCreate", ctx, int64(1)).Return(updatedPrefs, nil)

		output, err := uc.ToggleChannel(ctx, 1, input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.False(t, output.EmailEnabled)
		mockPrefsRepo.AssertExpectations(t)
	})

	t.Run("returns error when toggle fails", func(t *testing.T) {
		mockPrefsRepo := new(MockPreferencesRepository)

		uc := NewPreferencesUseCase(mockPrefsRepo)

		input := &dto.ChannelToggleInput{
			Channel: "push",
			Enabled: true,
		}

		mockPrefsRepo.On("UpdateChannelEnabled", ctx, int64(1), entities.NotificationChannel("push"), true).Return(assert.AnError)

		output, err := uc.ToggleChannel(ctx, 1, input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to toggle channel")
		assert.Nil(t, output)
		mockPrefsRepo.AssertExpectations(t)
	})
}

func TestPreferencesUseCase_UpdateQuietHours(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully updates quiet hours", func(t *testing.T) {
		mockPrefsRepo := new(MockPreferencesRepository)

		uc := NewPreferencesUseCase(mockPrefsRepo)

		input := &dto.QuietHoursInput{
			Enabled:   true,
			StartTime: "22:00",
			EndTime:   "08:00",
			Timezone:  "Europe/Moscow",
		}

		now := time.Now()
		updatedPrefs := &entities.UserNotificationPreferences{
			ID:                1,
			UserID:            1,
			QuietHoursEnabled: true,
			QuietHoursStart:   "22:00",
			QuietHoursEnd:     "08:00",
			Timezone:          "Europe/Moscow",
			CreatedAt:         now,
			UpdatedAt:         now,
		}

		mockPrefsRepo.On("UpdateQuietHours", ctx, int64(1), true, "22:00", "08:00", "Europe/Moscow").Return(nil)
		mockPrefsRepo.On("GetOrCreate", ctx, int64(1)).Return(updatedPrefs, nil)

		output, err := uc.UpdateQuietHours(ctx, 1, input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.True(t, output.QuietHoursEnabled)
		assert.Equal(t, "22:00", output.QuietHoursStart)
		assert.Equal(t, "08:00", output.QuietHoursEnd)
		assert.Equal(t, "Europe/Moscow", output.Timezone)
		mockPrefsRepo.AssertExpectations(t)
	})

	t.Run("returns error when update fails", func(t *testing.T) {
		mockPrefsRepo := new(MockPreferencesRepository)

		uc := NewPreferencesUseCase(mockPrefsRepo)

		input := &dto.QuietHoursInput{
			Enabled:   true,
			StartTime: "22:00",
			EndTime:   "08:00",
			Timezone:  "Europe/Moscow",
		}

		mockPrefsRepo.On("UpdateQuietHours", ctx, int64(1), true, "22:00", "08:00", "Europe/Moscow").Return(assert.AnError)

		output, err := uc.UpdateQuietHours(ctx, 1, input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update quiet hours")
		assert.Nil(t, output)
		mockPrefsRepo.AssertExpectations(t)
	})
}

func TestPreferencesUseCase_Reset(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully resets preferences", func(t *testing.T) {
		mockPrefsRepo := new(MockPreferencesRepository)

		uc := NewPreferencesUseCase(mockPrefsRepo)

		now := time.Now()
		defaultPrefs := &entities.UserNotificationPreferences{
			ID:           1,
			UserID:       1,
			EmailEnabled: true,
			PushEnabled:  true,
			InAppEnabled: true,
			Timezone:     "UTC",
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		mockPrefsRepo.On("Delete", ctx, int64(1)).Return(nil)
		mockPrefsRepo.On("GetOrCreate", ctx, int64(1)).Return(defaultPrefs, nil)

		output, err := uc.Reset(ctx, 1)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.True(t, output.EmailEnabled)
		assert.True(t, output.InAppEnabled)
		mockPrefsRepo.AssertExpectations(t)
	})

	t.Run("resets even when delete fails", func(t *testing.T) {
		mockPrefsRepo := new(MockPreferencesRepository)

		uc := NewPreferencesUseCase(mockPrefsRepo)

		now := time.Now()
		defaultPrefs := &entities.UserNotificationPreferences{
			ID:        1,
			UserID:    1,
			CreatedAt: now,
			UpdatedAt: now,
		}

		// Delete fails but GetOrCreate should still be called
		mockPrefsRepo.On("Delete", ctx, int64(1)).Return(assert.AnError)
		mockPrefsRepo.On("GetOrCreate", ctx, int64(1)).Return(defaultPrefs, nil)

		output, err := uc.Reset(ctx, 1)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		mockPrefsRepo.AssertExpectations(t)
	})

	t.Run("returns error when get fails", func(t *testing.T) {
		mockPrefsRepo := new(MockPreferencesRepository)

		uc := NewPreferencesUseCase(mockPrefsRepo)

		mockPrefsRepo.On("Delete", ctx, int64(1)).Return(nil)
		mockPrefsRepo.On("GetOrCreate", ctx, int64(1)).Return(nil, assert.AnError)

		output, err := uc.Reset(ctx, 1)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to reset preferences")
		assert.Nil(t, output)
		mockPrefsRepo.AssertExpectations(t)
	})
}

func TestPreferencesUseCase_EnableChannel(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully enables channel", func(t *testing.T) {
		mockPrefsRepo := new(MockPreferencesRepository)

		uc := NewPreferencesUseCase(mockPrefsRepo)

		mockPrefsRepo.On("UpdateChannelEnabled", ctx, int64(1), entities.NotificationChannel("email"), true).Return(nil)

		err := uc.EnableChannel(ctx, 1, "email")

		assert.NoError(t, err)
		mockPrefsRepo.AssertExpectations(t)
	})

	t.Run("returns error when enable fails", func(t *testing.T) {
		mockPrefsRepo := new(MockPreferencesRepository)

		uc := NewPreferencesUseCase(mockPrefsRepo)

		mockPrefsRepo.On("UpdateChannelEnabled", ctx, int64(1), entities.NotificationChannel("telegram"), true).Return(assert.AnError)

		err := uc.EnableChannel(ctx, 1, "telegram")

		assert.Error(t, err)
		mockPrefsRepo.AssertExpectations(t)
	})
}

func TestPreferencesUseCase_DisableChannel(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully disables channel", func(t *testing.T) {
		mockPrefsRepo := new(MockPreferencesRepository)

		uc := NewPreferencesUseCase(mockPrefsRepo)

		mockPrefsRepo.On("UpdateChannelEnabled", ctx, int64(1), entities.NotificationChannel("push"), false).Return(nil)

		err := uc.DisableChannel(ctx, 1, "push")

		assert.NoError(t, err)
		mockPrefsRepo.AssertExpectations(t)
	})

	t.Run("returns error when disable fails", func(t *testing.T) {
		mockPrefsRepo := new(MockPreferencesRepository)

		uc := NewPreferencesUseCase(mockPrefsRepo)

		mockPrefsRepo.On("UpdateChannelEnabled", ctx, int64(1), entities.NotificationChannel("slack"), false).Return(assert.AnError)

		err := uc.DisableChannel(ctx, 1, "slack")

		assert.Error(t, err)
		mockPrefsRepo.AssertExpectations(t)
	})
}

func TestPreferencesUseCase_GetAvailableTimezones(t *testing.T) {
	t.Run("returns available timezones", func(t *testing.T) {
		mockPrefsRepo := new(MockPreferencesRepository)

		uc := NewPreferencesUseCase(mockPrefsRepo)

		timezones := uc.GetAvailableTimezones()

		assert.NotEmpty(t, timezones)
		assert.Contains(t, timezones, "Europe/Moscow")
		assert.Contains(t, timezones, "UTC")
		assert.Contains(t, timezones, "Asia/Vladivostok")
		assert.Len(t, timezones, 12)
	})
}
