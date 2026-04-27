package entities

import "time"

type StudentGroup struct {
	ID          int64  `json:"id"`
	SpecialtyID int64  `json:"specialty_id"`
	Name        string `json:"name"`
	Course      int    `json:"course"`
	CuratorID   *int64 `json:"curator_id,omitempty"`
	Capacity    int    `json:"capacity"`
}

type Discipline struct {
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

type Semester struct {
	ID             int64     `json:"id"`
	AcademicYearID int64     `json:"academic_year_id"`
	Name           string    `json:"name"`
	Number         int       `json:"number"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	IsActive       bool      `json:"is_active"`
}

type LessonType struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	ShortName string  `json:"short_name"`
	Color     *string `json:"color,omitempty"`
}
