package usecases

import (
	"context"
	"errors"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/ical"
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
		loc:      ical.Location(cfg.TZID),
	}
	for _, opt := range opts {
		opt(uc)
	}
	return uc
}

// EnsureToken returns the user's feed token, creating one if none exists.
func (uc *CalendarFeedUseCase) EnsureToken(ctx context.Context, userID int64) (*entities.CalendarFeedToken, error) {
	existing, err := uc.tokens.GetByUserID(ctx, userID)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, entities.ErrCalendarFeedTokenNotFound) {
		return nil, err
	}
	return uc.issueToken(ctx, userID)
}

// RotateToken issues a fresh token for the user, invalidating the previous URL.
func (uc *CalendarFeedUseCase) RotateToken(ctx context.Context, userID int64) (*entities.CalendarFeedToken, error) {
	return uc.issueToken(ctx, userID)
}

// issueToken generates and persists a new token (upsert replaces any existing).
func (uc *CalendarFeedUseCase) issueToken(ctx context.Context, userID int64) (*entities.CalendarFeedToken, error) {
	raw, err := uc.generate()
	if err != nil {
		return nil, err
	}
	tok, err := entities.NewCalendarFeedToken(userID, raw, uc.now())
	if err != nil {
		return nil, err
	}
	if err := uc.tokens.Save(ctx, tok); err != nil {
		return nil, err
	}
	return tok, nil
}

// GetToken returns the user's current token, or ErrCalendarFeedTokenNotFound.
func (uc *CalendarFeedUseCase) GetToken(ctx context.Context, userID int64) (*entities.CalendarFeedToken, error) {
	return uc.tokens.GetByUserID(ctx, userID)
}

// DeleteToken disables the user's feed.
func (uc *CalendarFeedUseCase) DeleteToken(ctx context.Context, userID int64) error {
	return uc.tokens.DeleteByUserID(ctx, userID)
}

// RenderFeed resolves an opaque token to its owner and renders their iCalendar
// document. Returns ErrCalendarFeedTokenNotFound for an unknown token.
func (uc *CalendarFeedUseCase) RenderFeed(ctx context.Context, token string) (string, error) {
	tok, err := uc.tokens.GetByToken(ctx, token)
	if err != nil {
		return "", err
	}

	cal := ical.Calendar{
		ProdID: uc.cfg.ProdID,
		Name:   uc.cfg.CalendarName,
		TZID:   uc.cfg.TZID,
	}

	lessons, err := uc.gatherLessons(ctx, tok.UserID)
	if err != nil {
		return "", err
	}
	for _, l := range lessons {
		if ev, ok := lessonToICalEvent(l, uc.cfg.UIDDomain, uc.loc); ok {
			cal.Events = append(cal.Events, ev)
		}
	}

	now := uc.now()
	from := now.Add(-uc.cfg.PastWindow)
	to := now.Add(uc.cfg.FutureWindow)
	// GetByDateRange scopes events to those the user is involved in (organizer
	// or participant) within the window.
	events, err := uc.events.GetByDateRange(ctx, from, to, &tok.UserID)
	if err != nil {
		return "", err
	}
	for _, e := range events {
		if ev, ok := eventToICalEvent(e, uc.cfg.UIDDomain, uc.loc); ok {
			cal.Events = append(cal.Events, ev)
		}
	}

	return cal.Render(), nil
}

// gatherLessons collects the lessons a user should see: the lessons they teach,
// plus their student group's lessons if they belong to one. Results are
// de-duplicated by lesson id.
func (uc *CalendarFeedUseCase) gatherLessons(ctx context.Context, userID int64) ([]*entities.Lesson, error) {
	seen := make(map[int64]bool)
	var out []*entities.Lesson

	add := func(ls []*entities.Lesson) {
		for _, l := range ls {
			if !seen[l.ID] {
				seen[l.ID] = true
				out = append(out, l)
			}
		}
	}

	taught, err := uc.lessons.GetTimetable(ctx, LessonFilter{TeacherID: &userID})
	if err != nil {
		return nil, err
	}
	add(taught)

	groupID, err := uc.groups.ResolveGroupID(ctx, userID)
	switch {
	case err == nil:
		grouped, gerr := uc.lessons.GetTimetable(ctx, LessonFilter{GroupID: &groupID})
		if gerr != nil {
			return nil, gerr
		}
		add(grouped)
	case errors.Is(err, ErrStudentGroupNotFound):
		// user is not a student in any group — teacher/staff feed only
	default:
		return nil, err
	}

	return out, nil
}
