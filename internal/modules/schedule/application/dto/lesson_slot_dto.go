package dto

import "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"

// LessonSlotInput is the request body for creating or updating a bell-schedule slot.
type LessonSlotInput struct {
	Number    int    `json:"number" binding:"required"`
	TimeStart string `json:"time_start" binding:"required"`
	TimeEnd   string `json:"time_end" binding:"required"`
}

// LessonSlotOutput is the API representation of a bell-schedule slot.
type LessonSlotOutput struct {
	ID        int64  `json:"id"`
	Number    int    `json:"number"`
	TimeStart string `json:"time_start"`
	TimeEnd   string `json:"time_end"`
}

// ToLessonSlotOutput maps a LessonSlot entity to its API representation.
func ToLessonSlotOutput(s *entities.LessonSlot) LessonSlotOutput {
	return LessonSlotOutput{
		ID:        s.ID,
		Number:    s.Number,
		TimeStart: s.TimeStart,
		TimeEnd:   s.TimeEnd,
	}
}
