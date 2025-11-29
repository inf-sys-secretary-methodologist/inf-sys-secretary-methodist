// Package dto contains Data Transfer Objects for the schedule module.
package dto

import (
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
)

// CreateEventInput represents input for creating a new event
type CreateEventInput struct {
	Title          string     `json:"title" validate:"required,min=1,max=500"`
	Description    *string    `json:"description,omitempty" validate:"omitempty,max=5000"`
	EventType      string     `json:"event_type" validate:"required,oneof=meeting deadline task reminder holiday personal"`
	StartTime      time.Time  `json:"start_time" validate:"required"`
	EndTime        *time.Time `json:"end_time,omitempty"`
	AllDay         bool       `json:"all_day"`
	Timezone       string     `json:"timezone" validate:"omitempty,max=50"`
	Location       *string    `json:"location,omitempty" validate:"omitempty,max=500"`
	ParticipantIDs []int64    `json:"participant_ids,omitempty"`
	Color          *string    `json:"color,omitempty" validate:"omitempty,hexcolor"`
	Priority       *int       `json:"priority,omitempty" validate:"omitempty,min=1,max=5"`

	// Recurrence
	IsRecurring    bool                `json:"is_recurring"`
	RecurrenceRule *RecurrenceRuleInput `json:"recurrence_rule,omitempty"`

	// Reminders
	Reminders []ReminderInput `json:"reminders,omitempty"`
}

// UpdateEventInput represents input for updating an event
type UpdateEventInput struct {
	Title          *string    `json:"title,omitempty" validate:"omitempty,min=1,max=500"`
	Description    *string    `json:"description,omitempty" validate:"omitempty,max=5000"`
	EventType      *string    `json:"event_type,omitempty" validate:"omitempty,oneof=meeting deadline task reminder holiday personal"`
	Status         *string    `json:"status,omitempty" validate:"omitempty,oneof=scheduled ongoing completed cancelled postponed"`
	StartTime      *time.Time `json:"start_time,omitempty"`
	EndTime        *time.Time `json:"end_time,omitempty"`
	AllDay         *bool      `json:"all_day,omitempty"`
	Timezone       *string    `json:"timezone,omitempty" validate:"omitempty,max=50"`
	Location       *string    `json:"location,omitempty" validate:"omitempty,max=500"`
	Color          *string    `json:"color,omitempty" validate:"omitempty,hexcolor"`
	Priority       *int       `json:"priority,omitempty" validate:"omitempty,min=1,max=5"`

	// Recurrence update
	IsRecurring    *bool                `json:"is_recurring,omitempty"`
	RecurrenceRule *RecurrenceRuleInput `json:"recurrence_rule,omitempty"`
}

// RecurrenceRuleInput represents input for recurrence rule
type RecurrenceRuleInput struct {
	Frequency  string     `json:"frequency" validate:"required,oneof=daily weekly monthly yearly"`
	Interval   int        `json:"interval" validate:"min=1,max=365"`
	Count      *int       `json:"count,omitempty" validate:"omitempty,min=1,max=999"`
	Until      *time.Time `json:"until,omitempty"`
	ByWeekday  []string   `json:"by_weekday,omitempty" validate:"omitempty,dive,oneof=MO TU WE TH FR SA SU"`
	ByMonthDay []int      `json:"by_monthday,omitempty" validate:"omitempty,dive,min=1,max=31"`
	ByMonth    []int      `json:"by_month,omitempty" validate:"omitempty,dive,min=1,max=12"`
	WeekStart  string     `json:"week_start" validate:"omitempty,oneof=MO TU WE TH FR SA SU"`
}

// ReminderInput represents input for creating a reminder
type ReminderInput struct {
	ReminderType  string `json:"reminder_type" validate:"required,oneof=email push in_app telegram"`
	MinutesBefore int    `json:"minutes_before" validate:"required,min=0,max=40320"` // max 4 weeks
}

// EventOutput represents output for a single event
type EventOutput struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	EventType   string  `json:"event_type"`
	Status      string  `json:"status"`

	// Time
	StartTime time.Time  `json:"start_time"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	AllDay    bool       `json:"all_day"`
	Timezone  string     `json:"timezone"`

	// Location
	Location *string `json:"location,omitempty"`

	// Organizer
	OrganizerID   int64  `json:"organizer_id"`
	OrganizerName string `json:"organizer_name,omitempty"`

	// Participants
	Participants []ParticipantOutput `json:"participants,omitempty"`

	// Recurrence
	IsRecurring    bool                  `json:"is_recurring"`
	RecurrenceRule *RecurrenceRuleOutput `json:"recurrence_rule,omitempty"`
	ParentEventID  *int64                `json:"parent_event_id,omitempty"`

	// Display
	Color    *string `json:"color,omitempty"`
	Priority int     `json:"priority"`

	// Reminders
	Reminders []ReminderOutput `json:"reminders,omitempty"`

	// Audit
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RecurrenceRuleOutput represents output for recurrence rule
type RecurrenceRuleOutput struct {
	Frequency  string     `json:"frequency"`
	Interval   int        `json:"interval"`
	Count      *int       `json:"count,omitempty"`
	Until      *time.Time `json:"until,omitempty"`
	ByWeekday  []string   `json:"by_weekday,omitempty"`
	ByMonthDay []int      `json:"by_monthday,omitempty"`
	ByMonth    []int      `json:"by_month,omitempty"`
	WeekStart  string     `json:"week_start"`
}

// ParticipantOutput represents output for a participant
type ParticipantOutput struct {
	UserID         int64      `json:"user_id"`
	UserName       string     `json:"user_name,omitempty"`
	Email          string     `json:"email,omitempty"`
	ResponseStatus string     `json:"response_status"`
	Role           string     `json:"role"`
	RespondedAt    *time.Time `json:"responded_at,omitempty"`
}

// ReminderOutput represents output for a reminder
type ReminderOutput struct {
	ID            int64      `json:"id"`
	ReminderType  string     `json:"reminder_type"`
	MinutesBefore int        `json:"minutes_before"`
	IsSent        bool       `json:"is_sent"`
	SentAt        *time.Time `json:"sent_at,omitempty"`
}

// EventListOutput represents paginated list of events
type EventListOutput struct {
	Events     []*EventOutput `json:"events"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

// EventFilterInput represents filter options for listing events
type EventFilterInput struct {
	OrganizerID   *int64  `form:"organizer_id"`
	ParticipantID *int64  `form:"participant_id"`
	EventType     *string `form:"event_type"`
	Status        *string `form:"status"`
	StartFrom     *string `form:"start_from"` // RFC3339 format
	StartTo       *string `form:"start_to"`   // RFC3339 format
	Search        *string `form:"search"`
	IsRecurring   *bool   `form:"is_recurring"`
	Page          int     `form:"page,default=1"`
	PageSize      int     `form:"page_size,default=20"`
	OrderBy       *string `form:"order_by"`
}

// DateRangeInput represents input for date range queries
type DateRangeInput struct {
	Start time.Time `form:"start" binding:"required"`
	End   time.Time `form:"end" binding:"required"`
}

// UpdateParticipantStatusInput represents input for updating participant status
type UpdateParticipantStatusInput struct {
	Status string `json:"status" validate:"required,oneof=accepted declined tentative"`
}

// AddParticipantsInput represents input for adding participants
type AddParticipantsInput struct {
	UserIDs []int64 `json:"user_ids" validate:"required,min=1"`
	Role    string  `json:"role" validate:"required,oneof=required optional resource"`
}

// ToEventOutput converts entity to output DTO
func ToEventOutput(event *entities.Event) *EventOutput {
	output := &EventOutput{
		ID:            event.ID,
		Title:         event.Title,
		Description:   event.Description,
		EventType:     string(event.EventType),
		Status:        string(event.Status),
		StartTime:     event.StartTime,
		EndTime:       event.EndTime,
		AllDay:        event.AllDay,
		Timezone:      event.Timezone,
		Location:      event.Location,
		OrganizerID:   event.OrganizerID,
		IsRecurring:   event.IsRecurring,
		ParentEventID: event.ParentEventID,
		Color:         event.Color,
		Priority:      event.Priority,
		CreatedAt:     event.CreatedAt,
		UpdatedAt:     event.UpdatedAt,
	}

	if event.RecurrenceRule != nil {
		output.RecurrenceRule = ToRecurrenceRuleOutput(event.RecurrenceRule)
	}

	return output
}

// ToRecurrenceRuleOutput converts recurrence rule to output
func ToRecurrenceRuleOutput(rule *entities.RecurrenceRule) *RecurrenceRuleOutput {
	output := &RecurrenceRuleOutput{
		Frequency: string(rule.Frequency),
		Interval:  rule.Interval,
		Count:     rule.Count,
		Until:     rule.Until,
		WeekStart: string(rule.WeekStart),
	}

	if len(rule.ByWeekday) > 0 {
		output.ByWeekday = make([]string, len(rule.ByWeekday))
		for i, w := range rule.ByWeekday {
			output.ByWeekday[i] = string(w)
		}
	}

	output.ByMonthDay = rule.ByMonthDay
	output.ByMonth = rule.ByMonth

	return output
}

// ToParticipantOutput converts participant entity to output
func ToParticipantOutput(p *entities.EventParticipant) *ParticipantOutput {
	return &ParticipantOutput{
		UserID:         p.UserID,
		ResponseStatus: string(p.ResponseStatus),
		Role:           string(p.Role),
		RespondedAt:    p.RespondedAt,
	}
}

// ToReminderOutput converts reminder entity to output
func ToReminderOutput(r *entities.EventReminder) *ReminderOutput {
	return &ReminderOutput{
		ID:            r.ID,
		ReminderType:  string(r.ReminderType),
		MinutesBefore: r.MinutesBefore,
		IsSent:        r.IsSent,
		SentAt:        r.SentAt,
	}
}

// ToRecurrenceRule converts input to domain entity
func ToRecurrenceRule(input *RecurrenceRuleInput) *entities.RecurrenceRule {
	if input == nil {
		return nil
	}

	rule := &entities.RecurrenceRule{
		Frequency:  entities.RecurrenceFrequency(input.Frequency),
		Interval:   input.Interval,
		Count:      input.Count,
		Until:      input.Until,
		ByMonthDay: input.ByMonthDay,
		ByMonth:    input.ByMonth,
		WeekStart:  entities.Weekday(input.WeekStart),
	}

	if len(input.ByWeekday) > 0 {
		rule.ByWeekday = make([]entities.Weekday, len(input.ByWeekday))
		for i, w := range input.ByWeekday {
			rule.ByWeekday[i] = entities.Weekday(w)
		}
	}

	if rule.WeekStart == "" {
		rule.WeekStart = entities.WeekdayMonday
	}

	if rule.Interval == 0 {
		rule.Interval = 1
	}

	return rule
}
