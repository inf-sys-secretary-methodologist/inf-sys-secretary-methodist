package entities

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrInvalidEvent signals a violation of one of the
// ExtracurricularEvent construction or update invariants (empty title,
// oversize fields, non-chronological time range, invalid VO, negative
// capacity, non-positive organizer_id). Handlers map this sentinel to
// HTTP 422.
var ErrInvalidEvent = errors.New("extracurricular_event: invalid event")

// ErrEventScopeForbidden indicates that a caller is not authorized to
// mutate or view the event — typically because the caller is not the
// organizer and not an admin. Handlers map this sentinel to HTTP 403.
var ErrEventScopeForbidden = errors.New("extracurricular_event: caller cannot operate on this event")

// ErrCannotEditEvent indicates the event is in a terminal status
// (canceled or completed) that does not permit edits per ADR-2.
// Handlers map this sentinel to HTTP 422.
var ErrCannotEditEvent = errors.New("extracurricular_event: cannot edit in current status")

// ErrParticipantExists indicates a double-registration attempt — the
// user already has a participant row for this event. Handlers map this
// sentinel to HTTP 409.
var ErrParticipantExists = errors.New("extracurricular_event: participant already registered")

// ErrParticipantNotFound indicates Unregister was called for a user
// that was not previously registered. Handlers map this sentinel to
// HTTP 404.
var ErrParticipantNotFound = errors.New("extracurricular_event: participant not registered")

// ErrEventFull indicates a registration attempt against an event that
// has reached its max_capacity. Handlers map this sentinel to HTTP 409.
var ErrEventFull = errors.New("extracurricular_event: event at full capacity")

// ErrEventNotOpenForRegistration indicates registration was attempted
// against an event that is not in the `published` status (draft event
// is not yet visible; canceled/completed are terminal). Handlers map
// this sentinel to HTTP 422.
var ErrEventNotOpenForRegistration = errors.New("extracurricular_event: event not open for registration")

// ExtracurricularEvent is the aggregate root for внеучебных
// мероприятий per ADR-1 (plan 2026-05-24-b3-extracurricular.md).
// Participants are inner entities — accessed only through aggregate
// methods (Register / Unregister / HasParticipant) so the
// capacity invariant `len(participants) <= maxCapacity` стой within
// the transactional boundary.
//
// Lifecycle (ADR-2): status drives editability and registration
// eligibility. Optimistic locking (ADR-5): version bumps on every
// successful write; UPDATE WHERE version=? → 409 on race.
type ExtracurricularEvent struct {
	ID             int64
	title          string
	description    string
	category       Category
	targetAudience TargetAudience
	status         Status
	location       string
	startAt        time.Time
	endAt          time.Time
	maxCapacity    *int
	organizerID    int64
	participants   []Participant
	version        int
	createdAt      time.Time
	updatedAt      time.Time
}

// Participant is an inner entity of the ExtracurricularEvent aggregate
// representing one user's registration for the event. Persisted to a
// separate table (extracurricular_participants) but loaded together
// with the parent event on aggregate fetch.
type Participant struct {
	UserID       int64
	RegisteredAt time.Time
}

// NewExtracurricularEventParams bundles the constructor inputs to keep
// call sites readable as optional fields accumulate (mirror к
// NewSectionParams pattern).
type NewExtracurricularEventParams struct {
	Title          string
	Description    string
	Category       Category
	TargetAudience TargetAudience
	Location       string
	StartAt        time.Time
	EndAt          time.Time
	MaxCapacity    *int // nil = unlimited
	OrganizerID    int64
	Now            time.Time
}

// Event text-field bounds — mirrored by SQL CHECK constraints в
// migration 046.
const (
	maxEventTitleLen       = 255
	maxEventDescriptionLen = 4096
	maxEventLocationLen    = 255
)

// NewExtracurricularEvent validates invariants and returns a fresh
// event in draft status at version 0 (optimistic-locking baseline).
//
// Invariants (mirroring SQL CHECK constraints в migration 046):
//   - organizer_id > 0
//   - title trimmed-non-empty, ≤ 255 runes
//   - description ≤ 4096 runes (blank OK)
//   - location ≤ 255 runes (blank OK)
//   - category.IsValid()
//   - target_audience.IsValid()
//   - start_at < end_at
//   - max_capacity nil OR ≥ 0
//
// Each violation wraps ErrInvalidEvent с the offending field so
// errors.Is resolves the sentinel for the 422 mapping в handlers.
func NewExtracurricularEvent(p NewExtracurricularEventParams) (*ExtracurricularEvent, error) {
	if p.OrganizerID <= 0 {
		return nil, fmt.Errorf("%w: organizer_id must be positive, got %d",
			ErrInvalidEvent, p.OrganizerID)
	}
	title := strings.TrimSpace(p.Title)
	if title == "" {
		return nil, fmt.Errorf("%w: title must not be empty", ErrInvalidEvent)
	}
	if len([]rune(title)) > maxEventTitleLen {
		return nil, fmt.Errorf("%w: title length %d exceeds max %d",
			ErrInvalidEvent, len([]rune(title)), maxEventTitleLen)
	}
	description := strings.TrimSpace(p.Description)
	if len([]rune(description)) > maxEventDescriptionLen {
		return nil, fmt.Errorf("%w: description length %d exceeds max %d",
			ErrInvalidEvent, len([]rune(description)), maxEventDescriptionLen)
	}
	location := strings.TrimSpace(p.Location)
	if len([]rune(location)) > maxEventLocationLen {
		return nil, fmt.Errorf("%w: location length %d exceeds max %d",
			ErrInvalidEvent, len([]rune(location)), maxEventLocationLen)
	}
	if !p.Category.IsValid() {
		return nil, fmt.Errorf("%w: invalid category %q", ErrInvalidEvent, p.Category)
	}
	if !p.TargetAudience.IsValid() {
		return nil, fmt.Errorf("%w: invalid target_audience %q", ErrInvalidEvent, p.TargetAudience)
	}
	if !p.StartAt.Before(p.EndAt) {
		return nil, fmt.Errorf("%w: start_at (%v) must be before end_at (%v)",
			ErrInvalidEvent, p.StartAt, p.EndAt)
	}
	if p.MaxCapacity != nil && *p.MaxCapacity < 0 {
		return nil, fmt.Errorf("%w: max_capacity must be non-negative, got %d",
			ErrInvalidEvent, *p.MaxCapacity)
	}
	return &ExtracurricularEvent{
		title:          title,
		description:    description,
		category:       p.Category,
		targetAudience: p.TargetAudience,
		status:         StatusDraft,
		location:       location,
		startAt:        p.StartAt,
		endAt:          p.EndAt,
		maxCapacity:    p.MaxCapacity,
		organizerID:    p.OrganizerID,
		participants:   nil,
		version:        0,
		createdAt:      p.Now,
		updatedAt:      p.Now,
	}, nil
}

// ReconstituteExtracurricularEvent rebuilds an event from authoritative
// storage без re-validating invariants — DB CHECKs are the canonical
// gate at write time. Used exclusively by repository implementations.
func ReconstituteExtracurricularEvent(
	id int64,
	title, description string,
	category Category,
	targetAudience TargetAudience,
	status Status,
	location string,
	startAt, endAt time.Time,
	maxCapacity *int,
	organizerID int64,
	participants []Participant,
	version int,
	createdAt, updatedAt time.Time,
) *ExtracurricularEvent {
	return &ExtracurricularEvent{
		ID:             id,
		title:          title,
		description:    description,
		category:       category,
		targetAudience: targetAudience,
		status:         status,
		location:       location,
		startAt:        startAt,
		endAt:          endAt,
		maxCapacity:    maxCapacity,
		organizerID:    organizerID,
		participants:   participants,
		version:        version,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
	}
}

// Title returns the human-readable event title.
func (e *ExtracurricularEvent) Title() string { return e.title }

// Description returns the optional free-form description.
func (e *ExtracurricularEvent) Description() string { return e.description }

// Category returns the event classification (cultural/sports/...).
func (e *ExtracurricularEvent) Category() Category { return e.category }

// TargetAudience returns the role-cohort eligible to see + register.
func (e *ExtracurricularEvent) TargetAudience() TargetAudience { return e.targetAudience }

// Status returns the lifecycle state.
func (e *ExtracurricularEvent) Status() Status { return e.status }

// Location returns the venue (optional).
func (e *ExtracurricularEvent) Location() string { return e.location }

// StartAt returns the scheduled start timestamp.
func (e *ExtracurricularEvent) StartAt() time.Time { return e.startAt }

// EndAt returns the scheduled end timestamp.
func (e *ExtracurricularEvent) EndAt() time.Time { return e.endAt }

// MaxCapacity returns the participant cap; nil means unlimited.
func (e *ExtracurricularEvent) MaxCapacity() *int { return e.maxCapacity }

// OrganizerID returns the FK to users.id of the organizer.
func (e *ExtracurricularEvent) OrganizerID() int64 { return e.organizerID }

// Participants returns a defensive copy of the participants slice.
func (e *ExtracurricularEvent) Participants() []Participant {
	out := make([]Participant, len(e.participants))
	copy(out, e.participants)
	return out
}

// Version returns the optimistic-locking version.
func (e *ExtracurricularEvent) Version() int { return e.version }

// CreatedAt returns the creation timestamp.
func (e *ExtracurricularEvent) CreatedAt() time.Time { return e.createdAt }

// UpdatedAt returns the last-mutation timestamp.
func (e *ExtracurricularEvent) UpdatedAt() time.Time { return e.updatedAt }

// Register adds a participant to the event. Invariants:
//   - userID > 0
//   - event.Status().CanRegister() (only published)
//   - userID not already registered (ErrParticipantExists)
//   - len(participants) < maxCapacity если cap is set (ErrEventFull)
//
// On success appends new Participant{UserID, RegisteredAt: at}.
func (e *ExtracurricularEvent) Register(userID int64, at time.Time) error {
	if userID <= 0 {
		return fmt.Errorf("%w: user_id must be positive, got %d", ErrInvalidEvent, userID)
	}
	if !e.status.CanRegister() {
		return fmt.Errorf("%w: status=%q", ErrEventNotOpenForRegistration, e.status)
	}
	for _, p := range e.participants {
		if p.UserID == userID {
			return fmt.Errorf("%w: user_id=%d", ErrParticipantExists, userID)
		}
	}
	if e.maxCapacity != nil && len(e.participants) >= *e.maxCapacity {
		return fmt.Errorf("%w: capacity=%d", ErrEventFull, *e.maxCapacity)
	}
	e.participants = append(e.participants, Participant{UserID: userID, RegisteredAt: at})
	return nil
}

// Unregister removes the participant entry for userID. Returns
// ErrParticipantNotFound if no such registration exists.
func (e *ExtracurricularEvent) Unregister(userID int64) error {
	for i, p := range e.participants {
		if p.UserID == userID {
			e.participants = append(e.participants[:i], e.participants[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("%w: user_id=%d", ErrParticipantNotFound, userID)
}

// HasParticipant reports whether userID is currently registered.
func (e *ExtracurricularEvent) HasParticipant(userID int64) bool {
	for _, p := range e.participants {
		if p.UserID == userID {
			return true
		}
	}
	return false
}

// Publish transitions draft → published. Only legal transition source
// is draft; calling on any other status returns ErrCannotEditEvent.
func (e *ExtracurricularEvent) Publish(now time.Time) error {
	if e.status != StatusDraft {
		return fmt.Errorf("%w: cannot publish from status=%q", ErrCannotEditEvent, e.status)
	}
	e.status = StatusPublished
	e.updatedAt = now
	return nil
}

// Cancel transitions draft|published → canceled. Terminal statuses
// (canceled, completed) return ErrCannotEditEvent.
func (e *ExtracurricularEvent) Cancel(now time.Time) error {
	if e.status != StatusDraft && e.status != StatusPublished {
		return fmt.Errorf("%w: cannot cancel from status=%q", ErrCannotEditEvent, e.status)
	}
	e.status = StatusCanceled
	e.updatedAt = now
	return nil
}

// Complete transitions published → completed. Only published may be
// completed (draft must be published first; canceled is terminal).
func (e *ExtracurricularEvent) Complete(now time.Time) error {
	if e.status != StatusPublished {
		return fmt.Errorf("%w: cannot complete from status=%q", ErrCannotEditEvent, e.status)
	}
	e.status = StatusCompleted
	e.updatedAt = now
	return nil
}
