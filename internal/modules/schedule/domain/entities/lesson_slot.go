package entities

import (
	"errors"
	"time"
)

// Domain validation errors returned by NewLessonSlot.
var (
	ErrInvalidSlotNumber     = errors.New("slot number must be positive")
	ErrInvalidSlotTimeFormat = errors.New("slot time must be in HH:MM format")
	ErrInvalidSlotTimeRange  = errors.New("slot end time must be after start time")
)

// LessonSlot is one bell-schedule slot (пара): an ordered pair number with
// fixed start/end times shared institution-wide. It provides the discrete time
// domain the auto-scheduler places lessons into.
type LessonSlot struct {
	ID        int64     `json:"id"`
	Number    int       `json:"number"`
	TimeStart string    `json:"time_start"`
	TimeEnd   string    `json:"time_end"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewLessonSlot constructs a validated LessonSlot: number must be positive,
// times must be HH:MM and end must be strictly after start.
func NewLessonSlot(number int, timeStart, timeEnd string, now time.Time) (*LessonSlot, error) {
	if number <= 0 {
		return nil, ErrInvalidSlotNumber
	}
	start, err := parseSlotTime(timeStart)
	if err != nil {
		return nil, err
	}
	end, err := parseSlotTime(timeEnd)
	if err != nil {
		return nil, err
	}
	if !end.After(start) {
		return nil, ErrInvalidSlotTimeRange
	}
	return &LessonSlot{
		Number:    number,
		TimeStart: timeStart,
		TimeEnd:   timeEnd,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// parseSlotTime parses a strict zero-padded HH:MM value, returning
// ErrInvalidSlotTimeFormat on any malformed input. The round-trip check
// rejects non-canonical forms Go's parser would otherwise accept (e.g. "8:30").
func parseSlotTime(v string) (time.Time, error) {
	t, err := time.Parse("15:04", v)
	if err != nil || t.Format("15:04") != v {
		return time.Time{}, ErrInvalidSlotTimeFormat
	}
	return t, nil
}
