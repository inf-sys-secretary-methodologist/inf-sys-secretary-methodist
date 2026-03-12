package entities

import (
	"testing"
	"time"
)

func TestNewEvent(t *testing.T) {
	title := "Test Meeting"
	eventType := EventTypeMeeting
	startTime := time.Now().Add(24 * time.Hour)
	organizerID := int64(42)

	event := NewEvent(title, eventType, startTime, organizerID)

	if event.Title != title {
		t.Errorf("expected title %q, got %q", title, event.Title)
	}
	if event.EventType != eventType {
		t.Errorf("expected event type %q, got %q", eventType, event.EventType)
	}
	if event.Status != EventStatusScheduled {
		t.Errorf("expected status %q, got %q", EventStatusScheduled, event.Status)
	}
	if !event.StartTime.Equal(startTime) {
		t.Errorf("expected start time %v, got %v", startTime, event.StartTime)
	}
	if event.OrganizerID != organizerID {
		t.Errorf("expected organizer ID %d, got %d", organizerID, event.OrganizerID)
	}
	if event.AllDay {
		t.Error("expected AllDay to be false")
	}
	if event.IsRecurring {
		t.Error("expected IsRecurring to be false")
	}
	if event.Priority != 3 {
		t.Errorf("expected priority 3, got %d", event.Priority)
	}
	if event.Timezone != "Europe/Moscow" {
		t.Errorf("expected timezone %q, got %q", "Europe/Moscow", event.Timezone)
	}
}

func TestEvent_SetRecurrence(t *testing.T) {
	event := NewEvent("Meeting", EventTypeMeeting, time.Now(), 1)
	rule := &RecurrenceRule{
		Frequency: FrequencyWeekly,
		Interval:  1,
	}

	event.SetRecurrence(rule)

	if !event.IsRecurring {
		t.Error("expected IsRecurring to be true")
	}
	if event.RecurrenceRule == nil {
		t.Error("expected RecurrenceRule to be set")
	}
}

func TestEvent_SetRecurrence_Nil(t *testing.T) {
	event := NewEvent("Meeting", EventTypeMeeting, time.Now(), 1)
	event.IsRecurring = true
	event.RecurrenceRule = &RecurrenceRule{}

	event.SetRecurrence(nil)

	if event.IsRecurring {
		t.Error("expected IsRecurring to be false after setting nil rule")
	}
}

func TestEvent_Cancel(t *testing.T) {
	event := NewEvent("Meeting", EventTypeMeeting, time.Now(), 1)

	event.Cancel()

	if event.Status != EventStatusCancelled {
		t.Errorf("expected status %q, got %q", EventStatusCancelled, event.Status)
	}
}

func TestEvent_Complete(t *testing.T) {
	event := NewEvent("Meeting", EventTypeMeeting, time.Now(), 1)

	event.Complete()

	if event.Status != EventStatusCompleted {
		t.Errorf("expected status %q, got %q", EventStatusCompleted, event.Status)
	}
}

func TestEvent_Postpone(t *testing.T) {
	event := NewEvent("Meeting", EventTypeMeeting, time.Now(), 1)

	event.Postpone()

	if event.Status != EventStatusPostponed {
		t.Errorf("expected status %q, got %q", EventStatusPostponed, event.Status)
	}
}

func TestEvent_Reschedule(t *testing.T) {
	event := NewEvent("Meeting", EventTypeMeeting, time.Now(), 1)
	event.Status = EventStatusCancelled
	newStart := time.Now().Add(48 * time.Hour)
	newEnd := newStart.Add(1 * time.Hour)

	event.Reschedule(newStart, &newEnd)

	if !event.StartTime.Equal(newStart) {
		t.Errorf("expected start time %v, got %v", newStart, event.StartTime)
	}
	if event.EndTime == nil || !event.EndTime.Equal(newEnd) {
		t.Errorf("expected end time %v, got %v", newEnd, event.EndTime)
	}
	if event.Status != EventStatusScheduled {
		t.Errorf("expected status %q, got %q", EventStatusScheduled, event.Status)
	}
}

func TestEvent_IsDeleted(t *testing.T) {
	event := NewEvent("Meeting", EventTypeMeeting, time.Now(), 1)

	if event.IsDeleted() {
		t.Error("expected new event to not be deleted")
	}

	event.SoftDelete()

	if !event.IsDeleted() {
		t.Error("expected soft-deleted event to be deleted")
	}
}

func TestEvent_SoftDelete(t *testing.T) {
	event := NewEvent("Meeting", EventTypeMeeting, time.Now(), 1)

	event.SoftDelete()

	if event.DeletedAt == nil {
		t.Error("expected DeletedAt to be set")
	}
}

func TestEvent_Restore(t *testing.T) {
	event := NewEvent("Meeting", EventTypeMeeting, time.Now(), 1)
	event.SoftDelete()

	event.Restore()

	if event.DeletedAt != nil {
		t.Error("expected DeletedAt to be nil after restore")
	}
}

func TestEvent_IsActive(t *testing.T) {
	tests := []struct {
		name   string
		status EventStatus
		want   bool
	}{
		{"scheduled is active", EventStatusScheduled, true},
		{"ongoing is active", EventStatusOngoing, true},
		{"completed is not active", EventStatusCompleted, false},
		{"canceled is not active", EventStatusCancelled, false},
		{"postponed is not active", EventStatusPostponed, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := NewEvent("Meeting", EventTypeMeeting, time.Now(), 1)
			event.Status = tt.status

			got := event.IsActive()
			if got != tt.want {
				t.Errorf("IsActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEvent_IsPast(t *testing.T) {
	pastTime := time.Now().Add(-1 * time.Hour)
	futureTime := time.Now().Add(1 * time.Hour)

	// Event without end time, past start
	event1 := NewEvent("Past", EventTypeMeeting, pastTime, 1)
	if !event1.IsPast() {
		t.Error("expected event with past start time to be past")
	}

	// Event without end time, future start
	event2 := NewEvent("Future", EventTypeMeeting, futureTime, 1)
	if event2.IsPast() {
		t.Error("expected event with future start time to not be past")
	}

	// Event with end time, past end
	event3 := NewEvent("Past With End", EventTypeMeeting, pastTime.Add(-1*time.Hour), 1)
	event3.EndTime = &pastTime
	if !event3.IsPast() {
		t.Error("expected event with past end time to be past")
	}

	// Event with end time, future end
	event4 := NewEvent("Future With End", EventTypeMeeting, time.Now(), 1)
	event4.EndTime = &futureTime
	if event4.IsPast() {
		t.Error("expected event with future end time to not be past")
	}
}

func TestEvent_Duration(t *testing.T) {
	event := NewEvent("Meeting", EventTypeMeeting, time.Now(), 1)

	// No end time
	if event.Duration() != 0 {
		t.Errorf("expected duration 0, got %v", event.Duration())
	}

	// With end time
	endTime := event.StartTime.Add(2 * time.Hour)
	event.EndTime = &endTime

	if event.Duration() != 2*time.Hour {
		t.Errorf("expected duration 2h, got %v", event.Duration())
	}
}

func TestNewEventReminder(t *testing.T) {
	eventID := int64(1)
	userID := int64(42)
	reminderType := ReminderTypeEmail
	minutesBefore := 30

	reminder := NewEventReminder(eventID, userID, reminderType, minutesBefore)

	if reminder.EventID != eventID {
		t.Errorf("expected event ID %d, got %d", eventID, reminder.EventID)
	}
	if reminder.UserID != userID {
		t.Errorf("expected user ID %d, got %d", userID, reminder.UserID)
	}
	if reminder.ReminderType != reminderType {
		t.Errorf("expected reminder type %q, got %q", reminderType, reminder.ReminderType)
	}
	if reminder.MinutesBefore != minutesBefore {
		t.Errorf("expected minutes before %d, got %d", minutesBefore, reminder.MinutesBefore)
	}
	if reminder.IsSent {
		t.Error("expected IsSent to be false")
	}
}

func TestEventReminder_MarkSent(t *testing.T) {
	reminder := NewEventReminder(1, 1, ReminderTypeEmail, 30)

	reminder.MarkSent()

	if !reminder.IsSent {
		t.Error("expected IsSent to be true")
	}
	if reminder.SentAt == nil {
		t.Error("expected SentAt to be set")
	}
}

func TestEventTypeConstants(t *testing.T) {
	tests := []struct {
		name      string
		eventType EventType
		expected  string
	}{
		{"meeting", EventTypeMeeting, "meeting"},
		{"deadline", EventTypeDeadline, "deadline"},
		{"task", EventTypeTask, "task"},
		{"reminder", EventTypeReminder, "reminder"},
		{"holiday", EventTypeHoliday, "holiday"},
		{"personal", EventTypePersonal, "personal"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.eventType) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.eventType)
			}
		})
	}
}

func TestEventStatusConstants(t *testing.T) {
	tests := []struct {
		name     string
		status   EventStatus
		expected string
	}{
		{"scheduled", EventStatusScheduled, "scheduled"},
		{"ongoing", EventStatusOngoing, "ongoing"},
		{"completed", EventStatusCompleted, "completed"},
		{"canceled", EventStatusCancelled, "canceled"},
		{"postponed", EventStatusPostponed, "postponed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.status)
			}
		})
	}
}
