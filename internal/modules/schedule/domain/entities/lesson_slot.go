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

// NewLessonSlot constructs a validated LessonSlot. STUB — see GREEN commit.
func NewLessonSlot(number int, timeStart, timeEnd string, now time.Time) (*LessonSlot, error) {
	return &LessonSlot{
		Number:    number,
		TimeStart: timeStart,
		TimeEnd:   timeEnd,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}
