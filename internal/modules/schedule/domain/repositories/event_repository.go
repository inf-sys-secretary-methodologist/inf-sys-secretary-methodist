// Package repositories defines interfaces for schedule/calendar persistence.
package repositories

import (
	"context"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
)

// EventRepository defines the interface for event persistence
type EventRepository interface {
	// CRUD operations
	Create(ctx context.Context, event *entities.Event) error
	Update(ctx context.Context, event *entities.Event) error
	Delete(ctx context.Context, id int64) error
	SoftDelete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*entities.Event, error)

	// Query operations
	List(ctx context.Context, filter EventFilter) ([]*entities.Event, int64, error)
	GetByDateRange(ctx context.Context, start, end time.Time, userID *int64) ([]*entities.Event, error)
	GetByOrganizer(ctx context.Context, organizerID int64, limit, offset int) ([]*entities.Event, error)
	GetByParticipant(ctx context.Context, userID int64, limit, offset int) ([]*entities.Event, error)
	GetUpcoming(ctx context.Context, userID int64, limit int) ([]*entities.Event, error)

	// Recurrence operations
	GetRecurringEvents(ctx context.Context) ([]*entities.Event, error)
	GetRecurrenceInstances(ctx context.Context, parentEventID int64, start, end time.Time) ([]*entities.Event, error)
	CreateRecurrenceInstance(ctx context.Context, event *entities.Event) error
	GetRecurrenceExceptions(ctx context.Context, parentEventID int64) ([]time.Time, error)
	AddRecurrenceException(ctx context.Context, parentEventID int64, exceptionDate time.Time) error
}

// EventFilter contains filter options for listing events
type EventFilter struct {
	OrganizerID   *int64
	ParticipantID *int64
	EventType     *entities.EventType
	Status        *entities.EventStatus
	StartFrom     *time.Time
	StartTo       *time.Time
	SearchQuery   *string // search in title and description
	IsRecurring   *bool
	IncludeDeleted bool
	Limit         int
	Offset        int
	OrderBy       string // e.g., "start_time ASC"
}

// EventParticipantRepository defines the interface for event participant persistence
type EventParticipantRepository interface {
	// CRUD operations
	Create(ctx context.Context, participant *entities.EventParticipant) error
	Update(ctx context.Context, participant *entities.EventParticipant) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*entities.EventParticipant, error)

	// Query operations
	GetByEventID(ctx context.Context, eventID int64) ([]*entities.EventParticipant, error)
	GetByUserID(ctx context.Context, userID int64, limit, offset int) ([]*entities.EventParticipant, error)
	GetByEventAndUser(ctx context.Context, eventID, userID int64) (*entities.EventParticipant, error)

	// Bulk operations
	AddParticipants(ctx context.Context, eventID int64, userIDs []int64, role entities.ParticipantRole) error
	RemoveParticipants(ctx context.Context, eventID int64, userIDs []int64) error
	RemoveAllParticipants(ctx context.Context, eventID int64) error

	// Status operations
	UpdateStatus(ctx context.Context, eventID, userID int64, status entities.ParticipantStatus) error
	GetPendingInvitations(ctx context.Context, userID int64) ([]*entities.EventParticipant, error)
}

// EventReminderRepository defines the interface for event reminder persistence
type EventReminderRepository interface {
	// CRUD operations
	Create(ctx context.Context, reminder *entities.EventReminder) error
	Update(ctx context.Context, reminder *entities.EventReminder) error
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*entities.EventReminder, error)

	// Query operations
	GetByEventID(ctx context.Context, eventID int64) ([]*entities.EventReminder, error)
	GetByUserID(ctx context.Context, userID int64) ([]*entities.EventReminder, error)
	GetByEventAndUser(ctx context.Context, eventID, userID int64) ([]*entities.EventReminder, error)

	// Reminder processing
	GetPendingReminders(ctx context.Context, beforeTime time.Time) ([]*entities.EventReminder, error)
	MarkAsSent(ctx context.Context, id int64) error
	MarkMultipleAsSent(ctx context.Context, ids []int64) error

	// Bulk operations
	DeleteByEventID(ctx context.Context, eventID int64) error
	CreateDefault(ctx context.Context, eventID, userID int64) error // creates default reminders
}
