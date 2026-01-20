// Package entities contains domain entities for the schedule module.
package entities

import (
	"time"
)

// EventType represents the type of calendar event
type EventType string

const (
	EventTypeMeeting  EventType = "meeting"  // Встреча
	EventTypeDeadline EventType = "deadline" // Дедлайн
	EventTypeTask     EventType = "task"     // Задача
	EventTypeReminder EventType = "reminder" // Напоминание
	EventTypeHoliday  EventType = "holiday"  // Праздник/выходной
	EventTypePersonal EventType = "personal" // Личное событие
)

// EventStatus represents the status of an event
type EventStatus string

const (
	EventStatusScheduled EventStatus = "scheduled" // Запланировано
	EventStatusOngoing   EventStatus = "ongoing"   // В процессе
	EventStatusCompleted EventStatus = "completed" // Завершено
	EventStatusCancelled EventStatus = "cancelled" // Отменено
	EventStatusPostponed EventStatus = "postponed" // Отложено
)

// RecurrenceFrequency represents the frequency of recurring events
type RecurrenceFrequency string

const (
	FrequencyDaily   RecurrenceFrequency = "daily"
	FrequencyWeekly  RecurrenceFrequency = "weekly"
	FrequencyMonthly RecurrenceFrequency = "monthly"
	FrequencyYearly  RecurrenceFrequency = "yearly"
)

// Weekday represents days of the week for recurrence rules
type Weekday string

const (
	WeekdayMonday    Weekday = "MO"
	WeekdayTuesday   Weekday = "TU"
	WeekdayWednesday Weekday = "WE"
	WeekdayThursday  Weekday = "TH"
	WeekdayFriday    Weekday = "FR"
	WeekdaySaturday  Weekday = "SA"
	WeekdaySunday    Weekday = "SU"
)

// RecurrenceRule represents recurrence pattern for events (RFC 5545 RRULE)
type RecurrenceRule struct {
	Frequency  RecurrenceFrequency `json:"frequency"`               // FREQ: daily, weekly, monthly, yearly
	Interval   int                 `json:"interval"`                // INTERVAL: every N frequency units
	Count      *int                `json:"count,omitempty"`         // COUNT: number of occurrences
	Until      *time.Time          `json:"until,omitempty"`         // UNTIL: end date for recurrence
	ByWeekday  []Weekday           `json:"by_weekday,omitempty"`    // BYDAY: specific days of week
	ByMonthDay []int               `json:"by_monthday,omitempty"`   // BYMONTHDAY: specific days of month
	ByMonth    []int               `json:"by_month,omitempty"`      // BYMONTH: specific months
	WeekStart  Weekday             `json:"week_start"`              // WKST: week start day
}

// Event represents a calendar event entity
type Event struct {
	ID          int64       `json:"id"`
	Title       string      `json:"title"`
	Description *string     `json:"description,omitempty"`
	EventType   EventType   `json:"event_type"`
	Status      EventStatus `json:"status"`

	// Time fields
	StartTime time.Time  `json:"start_time"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	AllDay    bool       `json:"all_day"`
	Timezone  string     `json:"timezone"`

	// Location
	Location *string `json:"location,omitempty"`

	// Participants
	OrganizerID    int64   `json:"organizer_id"`
	ParticipantIDs []int64 `json:"participant_ids,omitempty"` // loaded separately

	// Recurrence
	IsRecurring    bool            `json:"is_recurring"`
	RecurrenceRule *RecurrenceRule `json:"recurrence_rule,omitempty"`
	ParentEventID  *int64          `json:"parent_event_id,omitempty"` // for recurring instances
	RecurrenceID   *time.Time      `json:"recurrence_id,omitempty"`   // original start time of instance

	// Categorization
	Color    *string `json:"color,omitempty"` // hex color for UI
	Priority int     `json:"priority"`        // 1-5, higher = more important

	// Metadata
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	ExternalID *string                `json:"external_id,omitempty"` // for external calendar sync

	// Audit
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// NewEvent creates a new event with default values
func NewEvent(title string, eventType EventType, startTime time.Time, organizerID int64) *Event {
	now := time.Now()
	return &Event{
		Title:       title,
		EventType:   eventType,
		Status:      EventStatusScheduled,
		StartTime:   startTime,
		AllDay:      false,
		Timezone:    "Europe/Moscow",
		OrganizerID: organizerID,
		IsRecurring: false,
		Priority:    3, // normal priority
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// SetRecurrence sets recurrence rule for the event
func (e *Event) SetRecurrence(rule *RecurrenceRule) {
	e.IsRecurring = rule != nil
	e.RecurrenceRule = rule
	e.UpdatedAt = time.Now()
}

// Cancel cancels the event
func (e *Event) Cancel() {
	e.Status = EventStatusCancelled
	e.UpdatedAt = time.Now()
}

// Complete marks the event as completed
func (e *Event) Complete() {
	e.Status = EventStatusCompleted
	e.UpdatedAt = time.Now()
}

// Postpone postpones the event
func (e *Event) Postpone() {
	e.Status = EventStatusPostponed
	e.UpdatedAt = time.Now()
}

// Reschedule changes the event time
func (e *Event) Reschedule(newStart time.Time, newEnd *time.Time) {
	e.StartTime = newStart
	e.EndTime = newEnd
	e.Status = EventStatusScheduled
	e.UpdatedAt = time.Now()
}

// IsDeleted checks if event is soft-deleted
func (e *Event) IsDeleted() bool {
	return e.DeletedAt != nil
}

// SoftDelete marks the event as deleted
func (e *Event) SoftDelete() {
	now := time.Now()
	e.DeletedAt = &now
	e.UpdatedAt = now
}

// Restore restores a soft-deleted event
func (e *Event) Restore() {
	e.DeletedAt = nil
	e.UpdatedAt = time.Now()
}

// IsActive checks if event is in active status
func (e *Event) IsActive() bool {
	return e.Status == EventStatusScheduled || e.Status == EventStatusOngoing
}

// IsPast checks if event has already passed
func (e *Event) IsPast() bool {
	if e.EndTime != nil {
		return e.EndTime.Before(time.Now())
	}
	return e.StartTime.Before(time.Now())
}

// Duration returns the duration of the event
func (e *Event) Duration() time.Duration {
	if e.EndTime == nil {
		return 0
	}
	return e.EndTime.Sub(e.StartTime)
}

// EventParticipant represents a participant in an event
type EventParticipant struct {
	ID             int64             `json:"id"`
	EventID        int64             `json:"event_id"`
	UserID         int64             `json:"user_id"`
	ResponseStatus ParticipantStatus `json:"response_status"`
	Role           ParticipantRole   `json:"role"`
	NotifiedAt     *time.Time        `json:"notified_at,omitempty"`
	RespondedAt    *time.Time        `json:"responded_at,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
}

// ParticipantStatus represents the response status of a participant
type ParticipantStatus string

const (
	ParticipantStatusPending   ParticipantStatus = "pending"   // Ожидает ответа
	ParticipantStatusAccepted  ParticipantStatus = "accepted"  // Принял
	ParticipantStatusDeclined  ParticipantStatus = "declined"  // Отклонил
	ParticipantStatusTentative ParticipantStatus = "tentative" // Возможно
)

// ParticipantRole represents the role of a participant
type ParticipantRole string

const (
	ParticipantRoleRequired ParticipantRole = "required" // Обязательный
	ParticipantRoleOptional ParticipantRole = "optional" // Необязательный
	ParticipantRoleResource ParticipantRole = "resource" // Ресурс (переговорная и т.д.)
)

// EventReminder represents a reminder for an event
type EventReminder struct {
	ID            int64        `json:"id"`
	EventID       int64        `json:"event_id"`
	UserID        int64        `json:"user_id"`
	ReminderType  ReminderType `json:"reminder_type"`
	MinutesBefore int          `json:"minutes_before"` // minutes before event
	IsSent        bool         `json:"is_sent"`
	SentAt        *time.Time   `json:"sent_at,omitempty"`
	CreatedAt     time.Time    `json:"created_at"`
}

// ReminderType represents the type of reminder notification
type ReminderType string

const (
	ReminderTypeEmail    ReminderType = "email"
	ReminderTypePush     ReminderType = "push"
	ReminderTypeInApp    ReminderType = "in_app"
	ReminderTypeTelegram ReminderType = "telegram"
)

// NewEventReminder creates a new reminder
func NewEventReminder(eventID, userID int64, reminderType ReminderType, minutesBefore int) *EventReminder {
	return &EventReminder{
		EventID:       eventID,
		UserID:        userID,
		ReminderType:  reminderType,
		MinutesBefore: minutesBefore,
		IsSent:        false,
		CreatedAt:     time.Now(),
	}
}

// MarkSent marks the reminder as sent
func (r *EventReminder) MarkSent() {
	now := time.Now()
	r.IsSent = true
	r.SentAt = &now
}
