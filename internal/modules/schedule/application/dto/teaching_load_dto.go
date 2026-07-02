package dto

import "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"

// TeachingLoadInput is the request body for creating or updating a load line.
type TeachingLoadInput struct {
	SemesterID   int64  `json:"semester_id" binding:"required"`
	GroupID      int64  `json:"group_id" binding:"required"`
	DisciplineID int64  `json:"discipline_id" binding:"required"`
	TeacherID    int64  `json:"teacher_id" binding:"required"`
	LessonTypeID int64  `json:"lesson_type_id" binding:"required"`
	PairsPerWeek int    `json:"pairs_per_week" binding:"required"`
	WeekType     string `json:"week_type" binding:"required"`
}

// TeachingLoadOutput is the API representation of a load line with hydrated names.
type TeachingLoadOutput struct {
	ID             int64  `json:"id"`
	SemesterID     int64  `json:"semester_id"`
	GroupID        int64  `json:"group_id"`
	GroupName      string `json:"group_name"`
	DisciplineID   int64  `json:"discipline_id"`
	DisciplineName string `json:"discipline_name"`
	TeacherID      int64  `json:"teacher_id"`
	TeacherName    string `json:"teacher_name"`
	LessonTypeID   int64  `json:"lesson_type_id"`
	LessonTypeName string `json:"lesson_type_name"`
	PairsPerWeek   int    `json:"pairs_per_week"`
	WeekType       string `json:"week_type"`
}

// ToTeachingLoadOutput maps a TeachingLoad entity to its API representation,
// tolerating missing (un-hydrated) associations.
func ToTeachingLoadOutput(l *entities.TeachingLoad) TeachingLoadOutput {
	out := TeachingLoadOutput{
		ID:           l.ID,
		SemesterID:   l.SemesterID,
		GroupID:      l.GroupID,
		DisciplineID: l.DisciplineID,
		TeacherID:    l.TeacherID,
		LessonTypeID: l.LessonTypeID,
		PairsPerWeek: l.PairsPerWeek,
		WeekType:     string(l.WeekType),
	}
	if l.Group != nil {
		out.GroupName = l.Group.Name
	}
	if l.Discipline != nil {
		out.DisciplineName = l.Discipline.Name
	}
	if l.Teacher != nil {
		out.TeacherName = l.Teacher.Name
	}
	if l.LessonType != nil {
		out.LessonTypeName = l.LessonType.Name
	}
	return out
}
