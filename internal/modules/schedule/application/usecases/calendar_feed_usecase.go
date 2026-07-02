package usecases

import (
	"context"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
)

// CalendarFeedConfig holds the presentation settings for a rendered feed.
type CalendarFeedConfig struct {
	ProdID       string        // PRODID identifier
	CalendarName string        // X-WR-CALNAME display name
	TZID         string        // IANA time zone for wall-clock rendering
	UIDDomain    string        // right-hand side of generated VEVENT UIDs
	PastWindow   time.Duration // how far back one-off events are included
	FutureWindow time.Duration // how far ahead one-off events are included
}

// CalendarFeedUseCase manages per-user feed tokens and renders the iCalendar
// subscription document from the user's lessons and events.
type CalendarFeedUseCase struct {
	tokens   CalendarFeedTokenRepository
	lessons  LessonRepository
	events   EventRepository
	groups   StudentGroupResolver
	cfg      CalendarFeedConfig
	now      func() time.Time
	generate func() (string, error)
	loc      *time.Location
}

// CalendarFeedOption overrides an optional dependency (used in tests).
type CalendarFeedOption func(*CalendarFeedUseCase)

// WithClock overrides the time source.
func WithClock(fn func() time.Time) CalendarFeedOption {
	return func(uc *CalendarFeedUseCase) { uc.now = fn }
}

// WithTokenGenerator overrides the token generator.
func WithTokenGenerator(fn func() (string, error)) CalendarFeedOption {
	return func(uc *CalendarFeedUseCase) { uc.generate = fn }
}

// NewCalendarFeedUseCase wires the use case with its dependencies.
func NewCalendarFeedUseCase(
	tokens CalendarFeedTokenRepository,
	lessons LessonRepository,
	events EventRepository,
	groups StudentGroupResolver,
	cfg CalendarFeedConfig,
	opts ...CalendarFeedOption,
) *CalendarFeedUseCase {
	uc := &CalendarFeedUseCase{
		tokens:   tokens,
		lessons:  lessons,
		events:   events,
		groups:   groups,
		cfg:      cfg,
		now:      time.Now,
		generate: entities.GenerateCalendarFeedToken,
		loc:      time.UTC,
	}
	for _, opt := range opts {
		opt(uc)
	}
	return uc
}

// EnsureToken returns the user's feed token, creating one if none exists.
func (uc *CalendarFeedUseCase) EnsureToken(ctx context.Context, userID int64) (*entities.CalendarFeedToken, error) {
	return nil, nil
}

// RotateToken issues a fresh token for the user, invalidating the previous URL.
func (uc *CalendarFeedUseCase) RotateToken(ctx context.Context, userID int64) (*entities.CalendarFeedToken, error) {
	return nil, nil
}

// GetToken returns the user's current token, or ErrCalendarFeedTokenNotFound.
func (uc *CalendarFeedUseCase) GetToken(ctx context.Context, userID int64) (*entities.CalendarFeedToken, error) {
	return nil, nil
}

// DeleteToken disables the user's feed.
func (uc *CalendarFeedUseCase) DeleteToken(ctx context.Context, userID int64) error {
	return nil
}

// RenderFeed resolves an opaque token to its owner and renders their iCalendar
// document. Returns ErrCalendarFeedTokenNotFound for an unknown token.
func (uc *CalendarFeedUseCase) RenderFeed(ctx context.Context, token string) (string, error) {
	return "", nil
}
