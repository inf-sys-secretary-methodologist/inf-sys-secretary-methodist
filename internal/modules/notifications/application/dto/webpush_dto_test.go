package dto

import (
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebPushSubscribeInput_ToEntity(t *testing.T) {
	input := &WebPushSubscribeInput{
		Endpoint:   "https://push.example.com/sub/123",
		P256dhKey:  "p256dh-key",
		AuthKey:    "auth-key",
		UserAgent:  "Mozilla/5.0",
		DeviceName: "Chrome on Linux",
	}

	entity := input.ToEntity(42)

	require.NotNil(t, entity)
	assert.Equal(t, int64(42), entity.UserID)
	assert.Equal(t, "https://push.example.com/sub/123", entity.Endpoint)
	assert.Equal(t, "p256dh-key", entity.P256dhKey)
	assert.Equal(t, "auth-key", entity.AuthKey)
	assert.Equal(t, "Mozilla/5.0", entity.UserAgent)
	assert.Equal(t, "Chrome on Linux", entity.DeviceName)
	assert.True(t, entity.IsActive)
}

func TestToSubscriptionOutput(t *testing.T) {
	now := time.Now()
	lastUsed := now.Add(-time.Hour)
	sub := &entities.WebPushSubscription{
		ID:         1,
		UserID:     42,
		DeviceName: "Firefox",
		UserAgent:  "Mozilla/5.0",
		IsActive:   true,
		LastUsedAt: &lastUsed,
		CreatedAt:  now,
	}

	output := ToSubscriptionOutput(sub)

	require.NotNil(t, output)
	assert.Equal(t, int64(1), output.ID)
	assert.Equal(t, "Firefox", output.DeviceName)
	assert.Equal(t, "Mozilla/5.0", output.UserAgent)
	assert.True(t, output.IsActive)
	assert.Equal(t, &lastUsed, output.LastUsedAt)
	assert.Equal(t, now, output.CreatedAt)
}

func TestToSubscriptionOutputList(t *testing.T) {
	now := time.Now()
	subs := []*entities.WebPushSubscription{
		{ID: 1, UserID: 42, IsActive: true, CreatedAt: now},
		{ID: 2, UserID: 42, IsActive: false, CreatedAt: now},
	}

	outputs := ToSubscriptionOutputList(subs)

	require.Len(t, outputs, 2)
	assert.True(t, outputs[0].IsActive)
	assert.False(t, outputs[1].IsActive)
}
