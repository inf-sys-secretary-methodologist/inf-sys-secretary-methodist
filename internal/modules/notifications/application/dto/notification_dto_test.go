package dto

import (
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/notifications/domain/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToOutput(t *testing.T) {
	now := time.Now()
	readAt := now.Add(-time.Hour)
	expiresAt := now.Add(24 * time.Hour)
	n := &entities.Notification{
		ID:        1,
		UserID:    42,
		Type:      entities.NotificationTypeInfo,
		Priority:  entities.PriorityHigh,
		Title:     "Test",
		Message:   "Hello",
		Link:      "https://example.com",
		ImageURL:  "https://example.com/img.png",
		IsRead:    true,
		ReadAt:    &readAt,
		ExpiresAt: &expiresAt,
		Metadata:  map[string]any{"key": "value"},
		CreatedAt: now,
	}

	output := ToOutput(n)

	require.NotNil(t, output)
	assert.Equal(t, int64(1), output.ID)
	assert.Equal(t, int64(42), output.UserID)
	assert.Equal(t, entities.NotificationTypeInfo, output.Type)
	assert.Equal(t, entities.PriorityHigh, output.Priority)
	assert.Equal(t, "Test", output.Title)
	assert.Equal(t, "Hello", output.Message)
	assert.Equal(t, "https://example.com", output.Link)
	assert.Equal(t, "https://example.com/img.png", output.ImageURL)
	assert.True(t, output.IsRead)
	assert.Equal(t, &readAt, output.ReadAt)
	assert.Equal(t, &expiresAt, output.ExpiresAt)
	assert.Equal(t, map[string]any{"key": "value"}, output.Metadata)
	assert.NotEmpty(t, output.TimeAgo)
}

func TestToOutputList(t *testing.T) {
	now := time.Now()
	notifications := []*entities.Notification{
		{ID: 1, UserID: 42, Type: entities.NotificationTypeInfo, Priority: entities.PriorityNormal,
			Title: "A", Message: "M", CreatedAt: now},
		{ID: 2, UserID: 42, Type: entities.NotificationTypeWarning, Priority: entities.PriorityHigh,
			Title: "B", Message: "N", CreatedAt: now},
	}

	outputs := ToOutputList(notifications)

	require.Len(t, outputs, 2)
	assert.Equal(t, "A", outputs[0].Title)
	assert.Equal(t, "B", outputs[1].Title)
}

func TestCreateNotificationInput_ToEntity(t *testing.T) {
	input := &CreateNotificationInput{
		UserID:   42,
		Type:     entities.NotificationTypeTask,
		Priority: entities.PriorityUrgent,
		Title:    "Deadline",
		Message:  "Task is due",
		Link:     "https://example.com/task/1",
		ImageURL: "https://example.com/img.png",
		Metadata: map[string]any{"task_id": 1},
	}

	entity := input.ToEntity()

	require.NotNil(t, entity)
	assert.Equal(t, int64(42), entity.UserID)
	assert.Equal(t, entities.NotificationTypeTask, entity.Type)
	assert.Equal(t, entities.PriorityUrgent, entity.Priority)
	assert.Equal(t, "Deadline", entity.Title)
	assert.Equal(t, "Task is due", entity.Message)
	assert.Equal(t, "https://example.com/task/1", entity.Link)
	assert.Equal(t, "https://example.com/img.png", entity.ImageURL)
	assert.False(t, entity.IsRead)
}

func TestCreateNotificationInput_ToEntity_DefaultPriority(t *testing.T) {
	input := &CreateNotificationInput{
		UserID:  42,
		Type:    entities.NotificationTypeInfo,
		Title:   "Info",
		Message: "Some info",
	}

	entity := input.ToEntity()
	assert.Equal(t, entities.PriorityNormal, entity.Priority)
}

func TestFormatTimeAgo_JustNow(t *testing.T) {
	result := formatTimeAgo(time.Now().Add(-10 * time.Second))
	assert.Equal(t, "только что", result)
}

func TestFormatTimeAgo_Minutes(t *testing.T) {
	result := formatTimeAgo(time.Now().Add(-5 * time.Minute))
	assert.Contains(t, result, "минут")
}

func TestFormatTimeAgo_Hours(t *testing.T) {
	result := formatTimeAgo(time.Now().Add(-3 * time.Hour))
	assert.Contains(t, result, "час")
}

func TestFormatTimeAgo_Days(t *testing.T) {
	result := formatTimeAgo(time.Now().Add(-2 * 24 * time.Hour))
	assert.Contains(t, result, "дн")
}

func TestFormatTimeAgo_OldDate(t *testing.T) {
	result := formatTimeAgo(time.Now().Add(-60 * 24 * time.Hour))
	assert.Contains(t, result, ".")
}

func TestFormatRussianPlural(t *testing.T) {
	tests := []struct {
		n        int
		expected string
	}{
		{1, "минуту"},
		{2, "минуты"},
		{5, "минут"},
		{11, "минут"},
		{21, "минуту"},
		{22, "минуты"},
		{25, "минут"},
	}

	for _, tt := range tests {
		result := formatRussianPlural(tt.n, "минуту", "минуты", "минут")
		assert.Equal(t, tt.expected, result, "for n=%d", tt.n)
	}
}
