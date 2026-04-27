package entities

import (
	"errors"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain"
)

var (
	ErrInvalidDayOfWeek = errors.New("invalid day of week")
	ErrInvalidTimeRange = errors.New("end time must be after start time")
	ErrInvalidDateRange = errors.New("end date must be after start date")
	ErrInvalidWeekType  = errors.New("invalid week type")
)

type Lesson struct {
	ID           int64            `json:"id"`
	SemesterID   int64            `json:"semester_id"`
	DisciplineID int64            `json:"discipline_id"`
	LessonTypeID int64            `json:"lesson_type_id"`
	TeacherID    int64            `json:"teacher_id"`
	GroupID      int64            `json:"group_id"`
	ClassroomID  int64            `json:"classroom_id"`
	DayOfWeek    domain.DayOfWeek `json:"day_of_week"`
	TimeStart    string           `json:"time_start"`
	TimeEnd      string           `json:"time_end"`
	WeekType     domain.WeekType  `json:"week_type"`
	DateStart    time.Time        `json:"date_start"`
	DateEnd      time.Time        `json:"date_end"`
	Notes        *string          `json:"notes,omitempty"`
	IsCancelled  bool             `json:"is_cancelled"`
	CancelReason *string          `json:"cancellation_reason,omitempty"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`

	Discipline *Discipline   `json:"discipline,omitempty"`
	LessonType *LessonType   `json:"lesson_type,omitempty"`
	Classroom  *Classroom    `json:"classroom,omitempty"`
	Group      *StudentGroup `json:"group,omitempty"`
	Teacher    *TeacherInfo  `json:"teacher,omitempty"`
}

type TeacherInfo struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func NewLesson(semesterID, disciplineID, lessonTypeID, teacherID, groupID, classroomID int64, day domain.DayOfWeek, timeStart, timeEnd string, weekType domain.WeekType) *Lesson {
	now := time.Now()
	return &Lesson{
		SemesterID:   semesterID,
		DisciplineID: disciplineID,
		LessonTypeID: lessonTypeID,
		TeacherID:    teacherID,
		GroupID:      groupID,
		ClassroomID:  classroomID,
		DayOfWeek:    day,
		TimeStart:    timeStart,
		TimeEnd:      timeEnd,
		WeekType:     weekType,
		DateStart:    now,
		DateEnd:      now.Add(120 * 24 * time.Hour),
		IsCancelled:  false,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func (l *Lesson) Validate() error {
	if !l.DayOfWeek.IsValid() {
		return ErrInvalidDayOfWeek
	}
	if !l.WeekType.IsValid() {
		return ErrInvalidWeekType
	}
	if l.TimeStart >= l.TimeEnd {
		return ErrInvalidTimeRange
	}
	if l.DateEnd.Before(l.DateStart) {
		return ErrInvalidDateRange
	}
	return nil
}
