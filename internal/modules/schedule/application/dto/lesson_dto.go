package dto

import (
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/repositories"
)

// CreateLessonInput represents input for creating a new lesson.
type CreateLessonInput struct {
	SemesterID   int64  `json:"semester_id" validate:"required"`
	DisciplineID int64  `json:"discipline_id" validate:"required"`
	LessonTypeID int64  `json:"lesson_type_id" validate:"required"`
	TeacherID    int64  `json:"teacher_id" validate:"required"`
	GroupID      int64  `json:"group_id" validate:"required"`
	ClassroomID  int64  `json:"classroom_id" validate:"required"`
	DayOfWeek    int    `json:"day_of_week" validate:"required,min=1,max=7"`
	TimeStart    string `json:"time_start" validate:"required"`
	TimeEnd      string `json:"time_end" validate:"required"`
	WeekType     string `json:"week_type" validate:"required,oneof=all odd even"`
	Notes        *string `json:"notes,omitempty"`
}

// UpdateLessonInput represents input for updating a lesson.
type UpdateLessonInput struct {
	SemesterID   *int64  `json:"semester_id,omitempty"`
	DisciplineID *int64  `json:"discipline_id,omitempty"`
	LessonTypeID *int64  `json:"lesson_type_id,omitempty"`
	TeacherID    *int64  `json:"teacher_id,omitempty"`
	GroupID      *int64  `json:"group_id,omitempty"`
	ClassroomID  *int64  `json:"classroom_id,omitempty"`
	DayOfWeek    *int    `json:"day_of_week,omitempty" validate:"omitempty,min=1,max=7"`
	TimeStart    *string `json:"time_start,omitempty"`
	TimeEnd      *string `json:"time_end,omitempty"`
	WeekType     *string `json:"week_type,omitempty" validate:"omitempty,oneof=all odd even"`
	Notes        *string `json:"notes,omitempty"`
}

// LessonFilterInput represents input for filtering lessons.
type LessonFilterInput struct {
	SemesterID   *int64  `form:"semester_id"`
	GroupID      *int64  `form:"group_id"`
	TeacherID    *int64  `form:"teacher_id"`
	ClassroomID  *int64  `form:"classroom_id"`
	DisciplineID *int64  `form:"discipline_id"`
	DayOfWeek    *int    `form:"day_of_week"`
	WeekType     *string `form:"week_type"`
	Limit        int     `form:"limit,default=100"`
	Offset       int     `form:"offset,default=0"`
}

// ToFilter converts LessonFilterInput to domain LessonFilter.
func (f *LessonFilterInput) ToFilter() repositories.LessonFilter {
	filter := repositories.LessonFilter{
		SemesterID:   f.SemesterID,
		GroupID:      f.GroupID,
		TeacherID:    f.TeacherID,
		ClassroomID:  f.ClassroomID,
		DisciplineID: f.DisciplineID,
	}
	if f.DayOfWeek != nil {
		dow := domain.DayOfWeek(*f.DayOfWeek)
		filter.DayOfWeek = &dow
	}
	if f.WeekType != nil {
		wt := domain.WeekType(*f.WeekType)
		filter.WeekType = &wt
	}
	return filter
}

// CreateChangeInput represents input for creating a schedule change.
type CreateChangeInput struct {
	LessonID       int64      `json:"lesson_id" validate:"required"`
	ChangeType     string     `json:"change_type" validate:"required,oneof=cancelled moved replaced_teacher replaced_classroom"`
	OriginalDate   time.Time  `json:"original_date" validate:"required"`
	NewDate        *time.Time `json:"new_date,omitempty"`
	NewClassroomID *int64     `json:"new_classroom_id,omitempty"`
	NewTeacherID   *int64     `json:"new_teacher_id,omitempty"`
	Reason         *string    `json:"reason,omitempty"`
}

// LessonOutput represents the output for a lesson.
type LessonOutput struct {
	ID           int64              `json:"id"`
	SemesterID   int64              `json:"semester_id"`
	DisciplineID int64              `json:"discipline_id"`
	LessonTypeID int64              `json:"lesson_type_id"`
	TeacherID    int64              `json:"teacher_id"`
	GroupID      int64              `json:"group_id"`
	ClassroomID  int64              `json:"classroom_id"`
	DayOfWeek    int                `json:"day_of_week"`
	TimeStart    string             `json:"time_start"`
	TimeEnd      string             `json:"time_end"`
	WeekType     string             `json:"week_type"`
	DateStart    time.Time          `json:"date_start"`
	DateEnd      time.Time          `json:"date_end"`
	Notes        *string            `json:"notes,omitempty"`
	IsCancelled  bool               `json:"is_cancelled"`
	CancelReason *string            `json:"cancellation_reason,omitempty"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
	Discipline   *DisciplineOutput  `json:"discipline,omitempty"`
	LessonType   *LessonTypeOutput  `json:"lesson_type,omitempty"`
	Classroom    *ClassroomOutput   `json:"classroom,omitempty"`
	Group        *StudentGroupOutput `json:"group,omitempty"`
	Teacher      *TeacherOutput     `json:"teacher,omitempty"`
}

// LessonListOutput represents the output for a list of lessons.
type LessonListOutput struct {
	Lessons []LessonOutput `json:"lessons"`
	Total   int64          `json:"total"`
	Limit   int            `json:"limit"`
	Offset  int            `json:"offset"`
}

// ClassroomOutput represents the output for a classroom.
type ClassroomOutput struct {
	ID          int64          `json:"id"`
	Building    string         `json:"building"`
	Number      string         `json:"number"`
	Name        *string        `json:"name,omitempty"`
	Capacity    int            `json:"capacity"`
	Type        *string        `json:"type,omitempty"`
	Equipment   map[string]any `json:"equipment,omitempty"`
	IsAvailable bool           `json:"is_available"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// StudentGroupOutput represents the output for a student group.
type StudentGroupOutput struct {
	ID          int64  `json:"id"`
	SpecialtyID int64  `json:"specialty_id"`
	Name        string `json:"name"`
	Course      int    `json:"course"`
	CuratorID   *int64 `json:"curator_id,omitempty"`
	Capacity    int    `json:"capacity"`
}

// DisciplineOutput represents the output for a discipline.
type DisciplineOutput struct {
	ID            int64   `json:"id"`
	Name          string  `json:"name"`
	Code          *string `json:"code,omitempty"`
	DepartmentID  *int64  `json:"department_id,omitempty"`
	Credits       *int    `json:"credits,omitempty"`
	HoursTotal    *int    `json:"hours_total,omitempty"`
	HoursLectures *int    `json:"hours_lectures,omitempty"`
	HoursPractice *int    `json:"hours_practice,omitempty"`
	HoursLabs     *int    `json:"hours_labs,omitempty"`
}

// LessonTypeOutput represents the output for a lesson type.
type LessonTypeOutput struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	ShortName string  `json:"short_name"`
	Color     *string `json:"color,omitempty"`
}

// TeacherOutput represents the output for a teacher.
type TeacherOutput struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// SemesterOutput represents the output for a semester.
type SemesterOutput struct {
	ID             int64     `json:"id"`
	AcademicYearID int64     `json:"academic_year_id"`
	Name           string    `json:"name"`
	Number         int       `json:"number"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	IsActive       bool      `json:"is_active"`
}

// ScheduleChangeOutput represents the output for a schedule change.
type ScheduleChangeOutput struct {
	ID             int64      `json:"id"`
	LessonID       int64      `json:"lesson_id"`
	ChangeType     string     `json:"change_type"`
	OriginalDate   time.Time  `json:"original_date"`
	NewDate        *time.Time `json:"new_date,omitempty"`
	NewClassroomID *int64     `json:"new_classroom_id,omitempty"`
	NewTeacherID   *int64     `json:"new_teacher_id,omitempty"`
	Reason         *string    `json:"reason,omitempty"`
	CreatedBy      int64      `json:"created_by"`
	CreatedAt      time.Time  `json:"created_at"`
}

// ToLessonOutput converts a Lesson entity to LessonOutput.
func ToLessonOutput(lesson *entities.Lesson) LessonOutput {
	output := LessonOutput{
		ID:           lesson.ID,
		SemesterID:   lesson.SemesterID,
		DisciplineID: lesson.DisciplineID,
		LessonTypeID: lesson.LessonTypeID,
		TeacherID:    lesson.TeacherID,
		GroupID:      lesson.GroupID,
		ClassroomID:  lesson.ClassroomID,
		DayOfWeek:    int(lesson.DayOfWeek),
		TimeStart:    lesson.TimeStart,
		TimeEnd:      lesson.TimeEnd,
		WeekType:     string(lesson.WeekType),
		DateStart:    lesson.DateStart,
		DateEnd:      lesson.DateEnd,
		Notes:        lesson.Notes,
		IsCancelled:  lesson.IsCancelled,
		CancelReason: lesson.CancelReason,
		CreatedAt:    lesson.CreatedAt,
		UpdatedAt:    lesson.UpdatedAt,
	}

	if lesson.Discipline != nil {
		d := ToDisciplineOutput(lesson.Discipline)
		output.Discipline = &d
	}

	if lesson.LessonType != nil {
		lt := ToLessonTypeOutput(lesson.LessonType)
		output.LessonType = &lt
	}

	if lesson.Classroom != nil {
		c := ToClassroomOutput(lesson.Classroom)
		output.Classroom = &c
	}

	if lesson.Group != nil {
		g := ToStudentGroupOutput(lesson.Group)
		output.Group = &g
	}

	if lesson.Teacher != nil {
		t := TeacherOutput{
			ID:    lesson.Teacher.ID,
			Name:  lesson.Teacher.Name,
			Email: lesson.Teacher.Email,
		}
		output.Teacher = &t
	}

	return output
}

// ToClassroomOutput converts a Classroom entity to ClassroomOutput.
func ToClassroomOutput(classroom *entities.Classroom) ClassroomOutput {
	return ClassroomOutput{
		ID:          classroom.ID,
		Building:    classroom.Building,
		Number:      classroom.Number,
		Name:        classroom.Name,
		Capacity:    classroom.Capacity,
		Type:        classroom.Type,
		Equipment:   classroom.Equipment,
		IsAvailable: classroom.IsAvailable,
		CreatedAt:   classroom.CreatedAt,
		UpdatedAt:   classroom.UpdatedAt,
	}
}

// ToStudentGroupOutput converts a StudentGroup entity to StudentGroupOutput.
func ToStudentGroupOutput(group *entities.StudentGroup) StudentGroupOutput {
	return StudentGroupOutput{
		ID:          group.ID,
		SpecialtyID: group.SpecialtyID,
		Name:        group.Name,
		Course:      group.Course,
		CuratorID:   group.CuratorID,
		Capacity:    group.Capacity,
	}
}

// ToDisciplineOutput converts a Discipline entity to DisciplineOutput.
func ToDisciplineOutput(discipline *entities.Discipline) DisciplineOutput {
	return DisciplineOutput{
		ID:            discipline.ID,
		Name:          discipline.Name,
		Code:          discipline.Code,
		DepartmentID:  discipline.DepartmentID,
		Credits:       discipline.Credits,
		HoursTotal:    discipline.HoursTotal,
		HoursLectures: discipline.HoursLectures,
		HoursPractice: discipline.HoursPractice,
		HoursLabs:     discipline.HoursLabs,
	}
}

// ToLessonTypeOutput converts a LessonType entity to LessonTypeOutput.
func ToLessonTypeOutput(lt *entities.LessonType) LessonTypeOutput {
	return LessonTypeOutput{
		ID:        lt.ID,
		Name:      lt.Name,
		ShortName: lt.ShortName,
		Color:     lt.Color,
	}
}

// ToSemesterOutput converts a Semester entity to SemesterOutput.
func ToSemesterOutput(semester *entities.Semester) SemesterOutput {
	return SemesterOutput{
		ID:             semester.ID,
		AcademicYearID: semester.AcademicYearID,
		Name:           semester.Name,
		Number:         semester.Number,
		StartDate:      semester.StartDate,
		EndDate:        semester.EndDate,
		IsActive:       semester.IsActive,
	}
}

// ToScheduleChangeOutput converts a ScheduleChange entity to ScheduleChangeOutput.
func ToScheduleChangeOutput(change *entities.ScheduleChange) ScheduleChangeOutput {
	return ScheduleChangeOutput{
		ID:             change.ID,
		LessonID:       change.LessonID,
		ChangeType:     string(change.ChangeType),
		OriginalDate:   change.OriginalDate,
		NewDate:        change.NewDate,
		NewClassroomID: change.NewClassroomID,
		NewTeacherID:   change.NewTeacherID,
		Reason:         change.Reason,
		CreatedBy:      change.CreatedBy,
		CreatedAt:      change.CreatedAt,
	}
}
