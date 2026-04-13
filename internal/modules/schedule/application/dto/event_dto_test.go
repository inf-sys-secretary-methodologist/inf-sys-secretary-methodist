package dto

import (
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToEventOutput(t *testing.T) {
	now := time.Now()
	endTime := now.Add(time.Hour)
	desc := "A meeting"
	loc := "Room 101"
	color := "#FF0000"

	event := &entities.Event{
		ID:          1,
		Title:       "Team Meeting",
		Description: &desc,
		EventType:   entities.EventTypeMeeting,
		Status:      entities.EventStatusScheduled,
		StartTime:   now,
		EndTime:     &endTime,
		AllDay:      false,
		Timezone:    "UTC",
		Location:    &loc,
		OrganizerID: 42,
		IsRecurring: false,
		Color:       &color,
		Priority:    3,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	output := ToEventOutput(event)

	require.NotNil(t, output)
	assert.Equal(t, int64(1), output.ID)
	assert.Equal(t, "Team Meeting", output.Title)
	assert.Equal(t, &desc, output.Description)
	assert.Equal(t, "meeting", output.EventType)
	assert.Equal(t, "scheduled", output.Status)
	assert.Equal(t, now, output.StartTime)
	assert.Equal(t, &endTime, output.EndTime)
	assert.False(t, output.AllDay)
	assert.Equal(t, "UTC", output.Timezone)
	assert.Equal(t, &loc, output.Location)
	assert.Equal(t, int64(42), output.OrganizerID)
	assert.False(t, output.IsRecurring)
	assert.Equal(t, &color, output.Color)
	assert.Equal(t, 3, output.Priority)
	assert.Nil(t, output.RecurrenceRule)
}

func TestToEventOutput_WithRecurrenceRule(t *testing.T) {
	now := time.Now()
	count := 10
	event := &entities.Event{
		ID:          1,
		Title:       "Weekly Meeting",
		EventType:   entities.EventTypeMeeting,
		Status:      entities.EventStatusScheduled,
		StartTime:   now,
		Timezone:    "UTC",
		OrganizerID: 42,
		IsRecurring: true,
		RecurrenceRule: &entities.RecurrenceRule{
			Frequency: entities.FrequencyWeekly,
			Interval:  1,
			Count:     &count,
			ByWeekday: []entities.Weekday{entities.WeekdayMonday, entities.WeekdayFriday},
			WeekStart: entities.WeekdayMonday,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	output := ToEventOutput(event)

	assert.True(t, output.IsRecurring)
	require.NotNil(t, output.RecurrenceRule)
	assert.Equal(t, "weekly", output.RecurrenceRule.Frequency)
	assert.Equal(t, 1, output.RecurrenceRule.Interval)
	assert.Equal(t, &count, output.RecurrenceRule.Count)
	require.Len(t, output.RecurrenceRule.ByWeekday, 2)
	assert.Equal(t, "MO", output.RecurrenceRule.ByWeekday[0])
	assert.Equal(t, "FR", output.RecurrenceRule.ByWeekday[1])
}

func TestToRecurrenceRuleOutput(t *testing.T) {
	until := time.Now().Add(30 * 24 * time.Hour)
	rule := &entities.RecurrenceRule{
		Frequency:  entities.FrequencyMonthly,
		Interval:   2,
		Until:      &until,
		ByMonthDay: []int{1, 15},
		ByMonth:    []int{1, 6},
		WeekStart:  entities.WeekdayMonday,
	}

	output := ToRecurrenceRuleOutput(rule)

	assert.Equal(t, "monthly", output.Frequency)
	assert.Equal(t, 2, output.Interval)
	assert.Equal(t, &until, output.Until)
	assert.Equal(t, []int{1, 15}, output.ByMonthDay)
	assert.Equal(t, []int{1, 6}, output.ByMonth)
	assert.Equal(t, "MO", output.WeekStart)
}

func TestToParticipantOutput_Schedule(t *testing.T) {
	respondedAt := time.Now()
	p := &entities.EventParticipant{
		UserID:         42,
		ResponseStatus: entities.ParticipantStatusAccepted,
		Role:           entities.ParticipantRoleRequired,
		RespondedAt:    &respondedAt,
	}

	output := ToParticipantOutput(p)

	require.NotNil(t, output)
	assert.Equal(t, int64(42), output.UserID)
	assert.Equal(t, "accepted", output.ResponseStatus)
	assert.Equal(t, "required", output.Role)
	assert.Equal(t, &respondedAt, output.RespondedAt)
}

func TestToReminderOutput(t *testing.T) {
	sentAt := time.Now()
	r := &entities.EventReminder{
		ID:            1,
		ReminderType:  entities.ReminderTypeEmail,
		MinutesBefore: 30,
		IsSent:        true,
		SentAt:        &sentAt,
	}

	output := ToReminderOutput(r)

	require.NotNil(t, output)
	assert.Equal(t, int64(1), output.ID)
	assert.Equal(t, "email", output.ReminderType)
	assert.Equal(t, 30, output.MinutesBefore)
	assert.True(t, output.IsSent)
	assert.Equal(t, &sentAt, output.SentAt)
}

func TestToRecurrenceRule(t *testing.T) {
	count := 5
	input := &RecurrenceRuleInput{
		Frequency: "weekly",
		Interval:  2,
		Count:     &count,
		ByWeekday: []string{"MO", "WE", "FR"},
		WeekStart: "MO",
	}

	rule := ToRecurrenceRule(input)

	require.NotNil(t, rule)
	assert.Equal(t, entities.FrequencyWeekly, rule.Frequency)
	assert.Equal(t, 2, rule.Interval)
	assert.Equal(t, &count, rule.Count)
	require.Len(t, rule.ByWeekday, 3)
	assert.Equal(t, entities.WeekdayMonday, rule.ByWeekday[0])
	assert.Equal(t, entities.WeekdayMonday, rule.WeekStart)
}

func TestToRecurrenceRule_Nil(t *testing.T) {
	rule := ToRecurrenceRule(nil)
	assert.Nil(t, rule)
}

func TestToRecurrenceRule_Defaults(t *testing.T) {
	input := &RecurrenceRuleInput{
		Frequency: "daily",
	}

	rule := ToRecurrenceRule(input)

	require.NotNil(t, rule)
	assert.Equal(t, entities.WeekdayMonday, rule.WeekStart) // default
	assert.Equal(t, 1, rule.Interval)                        // default
}
