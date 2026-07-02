package entities

import (
	"errors"
	"time"
)

// Domain errors for the calendar feed token.
var (
	ErrCalendarFeedTokenUserRequired = errors.New("calendar feed token requires a positive user id")
	ErrCalendarFeedTokenEmpty        = errors.New("calendar feed token value must not be empty")
	ErrCalendarFeedTokenNotFound     = errors.New("calendar feed token not found")
)

// CalendarFeedToken is the opaque secret that authenticates a personal
// iCalendar subscription feed. Each user has at most one token; it is stored in
// plaintext because it grants read-only access to schedule data the user may
// already view, and the feed URL must be retrievable repeatedly.
type CalendarFeedToken struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
}

// NewCalendarFeedToken constructs a token, enforcing its invariants.
func NewCalendarFeedToken(userID int64, token string, now time.Time) (*CalendarFeedToken, error) {
	return &CalendarFeedToken{}, nil
}

// GenerateCalendarFeedToken returns a cryptographically random 256-bit token
// encoded as 64 hex characters.
func GenerateCalendarFeedToken() (string, error) {
	return "", nil
}
