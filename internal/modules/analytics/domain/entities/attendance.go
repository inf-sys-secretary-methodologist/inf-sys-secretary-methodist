// Package entities contains domain entities for the analytics module.
package entities

import "time"

// AttendanceStatus represents the status of attendance
type AttendanceStatus string

// AttendanceStatus values.
const (
	AttendanceStatusPresent AttendanceStatus = "present"
	AttendanceStatusAbsent  AttendanceStatus = "absent"
	AttendanceStatusLate    AttendanceStatus = "late"
	AttendanceStatusExcused AttendanceStatus = "excused"
)

// LessonType represents the type of lesson
type LessonType string

// LessonType values.
const (
	LessonTypeLecture  LessonType = "lecture"
	LessonTypePractice LessonType = "practice"
	LessonTypeLab      LessonType = "lab"
	LessonTypeSeminar  LessonType = "seminar"
	LessonTypeExam     LessonType = "exam"
)

// Lesson represents a class/lesson that can be attended
type Lesson struct {
	ID         int64      `json:"id"`
	Name       string     `json:"name"`
	Subject    string     `json:"subject"`
	TeacherID  *int64     `json:"teacher_id,omitempty"`
	GroupName  *string    `json:"group_name,omitempty"`
	LessonType LessonType `json:"lesson_type"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// AttendanceRecord represents a single attendance record
type AttendanceRecord struct {
	ID         int64            `json:"id"`
	StudentID  int64            `json:"student_id"`
	LessonID   int64            `json:"lesson_id"`
	LessonDate time.Time        `json:"lesson_date"`
	Status     AttendanceStatus `json:"status"`
	MarkedBy   *int64           `json:"marked_by,omitempty"`
	Notes      *string          `json:"notes,omitempty"`
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at"`
}

// GradeType represents the type of grade
type GradeType string

// GradeType values.
const (
	GradeTypeCurrent  GradeType = "current"
	GradeTypeMidterm  GradeType = "midterm"
	GradeTypeFinal    GradeType = "final"
	GradeTypeTest     GradeType = "test"
	GradeTypeHomework GradeType = "homework"
)

// Grade represents a student's grade
type Grade struct {
	ID         int64     `json:"id"`
	StudentID  int64     `json:"student_id"`
	Subject    string    `json:"subject"`
	GradeType  GradeType `json:"grade_type"`
	GradeValue float64   `json:"grade_value"`
	MaxValue   float64   `json:"max_value"`
	Weight     float64   `json:"weight"`
	GradedBy   *int64    `json:"graded_by,omitempty"`
	GradeDate  time.Time `json:"grade_date"`
	Notes      *string   `json:"notes,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
