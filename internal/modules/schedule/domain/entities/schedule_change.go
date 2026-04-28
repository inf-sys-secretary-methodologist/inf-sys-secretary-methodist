package entities

import (
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain"
)

type ScheduleChange struct {
	ID             int64             `json:"id"`
	LessonID       int64             `json:"lesson_id"`
	ChangeType     domain.ChangeType `json:"change_type"`
	OriginalDate   time.Time         `json:"original_date"`
	NewDate        *time.Time        `json:"new_date,omitempty"`
	NewClassroomID *int64            `json:"new_classroom_id,omitempty"`
	NewTeacherID   *int64            `json:"new_teacher_id,omitempty"`
	Reason         *string           `json:"reason,omitempty"`
	CreatedBy      int64             `json:"created_by"`
	CreatedAt      time.Time         `json:"created_at"`
}

func NewScheduleChange(lessonID int64, changeType domain.ChangeType, originalDate time.Time, createdBy int64) *ScheduleChange {
	return &ScheduleChange{
		LessonID:     lessonID,
		ChangeType:   changeType,
		OriginalDate: originalDate,
		CreatedBy:    createdBy,
		CreatedAt:    time.Now(),
	}
}
