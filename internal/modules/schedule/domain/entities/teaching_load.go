package entities

import (
	"errors"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain"
)

// Domain validation errors returned by NewTeachingLoad.
var (
	ErrInvalidLoadReference = errors.New("teaching load references must be positive")
	ErrInvalidLoadPairs     = errors.New("pairs per week must be positive")
	ErrInvalidLoadWeekType  = errors.New("invalid week type")
)

// Repository-level sentinels for TeachingLoad lookups and constraints.
var (
	ErrTeachingLoadNotFound  = errors.New("teaching load not found")
	ErrTeachingLoadDuplicate = errors.New("teaching load for this group/discipline/type already exists")
)

// TeachingLoad is one planned teaching assignment: a group studies a discipline
// with a teacher for N pairs per week (on all/odd/even weeks). It is the source
// of truth the auto-scheduler expands into schedulable lessons.
type TeachingLoad struct {
	ID           int64           `json:"id"`
	SemesterID   int64           `json:"semester_id"`
	GroupID      int64           `json:"group_id"`
	DisciplineID int64           `json:"discipline_id"`
	TeacherID    int64           `json:"teacher_id"`
	LessonTypeID int64           `json:"lesson_type_id"`
	PairsPerWeek int             `json:"pairs_per_week"`
	WeekType     domain.WeekType `json:"week_type"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`

	// Hydrated associations for read models (optional).
	Group      *StudentGroup `json:"group,omitempty"`
	Discipline *Discipline   `json:"discipline,omitempty"`
	LessonType *LessonType   `json:"lesson_type,omitempty"`
	Teacher    *TeacherInfo  `json:"teacher,omitempty"`
}

// NewTeachingLoad constructs a validated TeachingLoad. STUB — see GREEN commit.
func NewTeachingLoad(semesterID, groupID, disciplineID, teacherID, lessonTypeID int64, pairsPerWeek int, weekType domain.WeekType, now time.Time) (*TeachingLoad, error) {
	return &TeachingLoad{
		SemesterID:   semesterID,
		GroupID:      groupID,
		DisciplineID: disciplineID,
		TeacherID:    teacherID,
		LessonTypeID: lessonTypeID,
		PairsPerWeek: pairsPerWeek,
		WeekType:     weekType,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}
