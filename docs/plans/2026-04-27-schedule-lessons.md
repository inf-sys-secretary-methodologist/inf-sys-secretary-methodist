# Schedule Lessons (GH #201) Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build full-stack schedule lessons module — backend CRUD + frontend timetable grid with role-based access (secretary/admin edit, others view-only).

**Architecture:** Add lesson entities to existing `internal/modules/schedule/` module. Backend: domain entities → repository interfaces → PG persistence → DTOs → usecases → HTTP handlers. Frontend: types → API client → SWR hooks → components → page. Reference data (classrooms, groups, disciplines, semesters, lesson_types) exposed as read-only endpoints.

**Tech Stack:** Go 1.25 + Gin + PostgreSQL (tables exist in migration 004), Next.js 15 + TypeScript + SWR + Tailwind, Jest for frontend tests, testify for Go tests.

**Module path:** `github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist`

---

## Phase 1: Backend — Domain Layer

### Task 1: Domain types (enums)

**Files:**
- Create: `internal/modules/schedule/domain/lesson_types.go`
- Test: `internal/modules/schedule/domain/lesson_types_test.go`

**Step 1: Write failing test**

```go
// lesson_types_test.go
package domain

import "testing"

func TestDayOfWeek_IsValid(t *testing.T) {
	tests := []struct {
		day  DayOfWeek
		want bool
	}{
		{Monday, true},
		{Sunday, true},
		{DayOfWeek(0), false},
		{DayOfWeek(8), false},
	}
	for _, tt := range tests {
		if got := tt.day.IsValid(); got != tt.want {
			t.Errorf("DayOfWeek(%d).IsValid() = %v, want %v", tt.day, got, tt.want)
		}
	}
}

func TestWeekType_IsValid(t *testing.T) {
	tests := []struct {
		wt   WeekType
		want bool
	}{
		{WeekTypeAll, true},
		{WeekTypeOdd, true},
		{WeekTypeEven, true},
		{WeekType("invalid"), false},
	}
	for _, tt := range tests {
		if got := tt.wt.IsValid(); got != tt.want {
			t.Errorf("WeekType(%s).IsValid() = %v, want %v", tt.wt, got, tt.want)
		}
	}
}

func TestChangeType_IsValid(t *testing.T) {
	tests := []struct {
		ct   ChangeType
		want bool
	}{
		{ChangeTypeCancelled, true},
		{ChangeTypeMoved, true},
		{ChangeTypeReplacedTeacher, true},
		{ChangeTypeReplacedClassroom, true},
		{ChangeType("invalid"), false},
	}
	for _, tt := range tests {
		if got := tt.ct.IsValid(); got != tt.want {
			t.Errorf("ChangeType(%s).IsValid() = %v, want %v", tt.ct, got, tt.want)
		}
	}
}
```

**Step 2: Run test — expect FAIL**

```bash
cd internal/modules/schedule && go test ./domain/ -run "TestDayOfWeek|TestWeekType|TestChangeType" -v
```

**Step 3: Implement**

```go
// lesson_types.go
package domain

// DayOfWeek represents day of week (1=Monday, 7=Sunday).
type DayOfWeek int

const (
	Monday    DayOfWeek = 1
	Tuesday   DayOfWeek = 2
	Wednesday DayOfWeek = 3
	Thursday  DayOfWeek = 4
	Friday    DayOfWeek = 5
	Saturday  DayOfWeek = 6
	Sunday    DayOfWeek = 7
)

func (d DayOfWeek) IsValid() bool {
	return d >= Monday && d <= Sunday
}

// WeekType represents which weeks the lesson occurs on.
type WeekType string

const (
	WeekTypeAll  WeekType = "all"
	WeekTypeOdd  WeekType = "odd"
	WeekTypeEven WeekType = "even"
)

func (w WeekType) IsValid() bool {
	switch w {
	case WeekTypeAll, WeekTypeOdd, WeekTypeEven:
		return true
	}
	return false
}

// ChangeType represents type of schedule change.
type ChangeType string

const (
	ChangeTypeCancelled         ChangeType = "cancelled"
	ChangeTypeMoved             ChangeType = "moved"
	ChangeTypeReplacedTeacher   ChangeType = "replaced_teacher"
	ChangeTypeReplacedClassroom ChangeType = "replaced_classroom"
)

func (c ChangeType) IsValid() bool {
	switch c {
	case ChangeTypeCancelled, ChangeTypeMoved, ChangeTypeReplacedTeacher, ChangeTypeReplacedClassroom:
		return true
	}
	return false
}
```

**Step 4: Run test — expect PASS**

**Step 5: Commit**

```bash
git add internal/modules/schedule/domain/lesson_types.go internal/modules/schedule/domain/lesson_types_test.go
# RED commit first (test only), then GREEN commit (impl only)
```

---

### Task 2: Domain entities (Lesson, Classroom, ScheduleChange)

**Files:**
- Create: `internal/modules/schedule/domain/entities/lesson.go`
- Create: `internal/modules/schedule/domain/entities/classroom.go`
- Create: `internal/modules/schedule/domain/entities/schedule_change.go`
- Create: `internal/modules/schedule/domain/entities/reference.go` (StudentGroup, Discipline, Semester, LessonType)
- Test: `internal/modules/schedule/domain/entities/lesson_test.go`

**Step 1: Write failing test**

```go
// lesson_test.go
package entities

import (
	"testing"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain"
)

func TestNewLesson(t *testing.T) {
	lesson := NewLesson(1, 1, 1, 1, 1, 1, domain.Monday, "09:00", "10:30", domain.WeekTypeAll)
	if lesson == nil {
		t.Fatal("NewLesson returned nil")
	}
	if lesson.DayOfWeek != domain.Monday {
		t.Errorf("DayOfWeek = %v, want %v", lesson.DayOfWeek, domain.Monday)
	}
	if lesson.TimeStart != "09:00" {
		t.Errorf("TimeStart = %v, want 09:00", lesson.TimeStart)
	}
	if lesson.IsCancelled {
		t.Error("new lesson should not be cancelled")
	}
}

func TestLesson_Validate(t *testing.T) {
	tests := []struct {
		name    string
		lesson  *Lesson
		wantErr bool
	}{
		{
			name:    "valid lesson",
			lesson:  NewLesson(1, 1, 1, 1, 1, 1, domain.Monday, "09:00", "10:30", domain.WeekTypeAll),
			wantErr: false,
		},
		{
			name: "invalid day of week",
			lesson: &Lesson{
				DayOfWeek: domain.DayOfWeek(0),
				TimeStart: "09:00",
				TimeEnd:   "10:30",
				WeekType:  domain.WeekTypeAll,
				DateStart: time.Now(),
				DateEnd:   time.Now().Add(24 * time.Hour),
			},
			wantErr: true,
		},
		{
			name: "end time before start time",
			lesson: &Lesson{
				DayOfWeek: domain.Monday,
				TimeStart: "10:30",
				TimeEnd:   "09:00",
				WeekType:  domain.WeekTypeAll,
				DateStart: time.Now(),
				DateEnd:   time.Now().Add(24 * time.Hour),
			},
			wantErr: true,
		},
		{
			name: "date_end before date_start",
			lesson: &Lesson{
				DayOfWeek: domain.Monday,
				TimeStart: "09:00",
				TimeEnd:   "10:30",
				WeekType:  domain.WeekTypeAll,
				DateStart: time.Now().Add(48 * time.Hour),
				DateEnd:   time.Now(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.lesson.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
```

**Step 2: Run test — expect FAIL**

**Step 3: Implement entities**

```go
// lesson.go
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
	ID           int64           `json:"id"`
	SemesterID   int64           `json:"semester_id"`
	DisciplineID int64           `json:"discipline_id"`
	LessonTypeID int64           `json:"lesson_type_id"`
	TeacherID    int64           `json:"teacher_id"`
	GroupID      int64           `json:"group_id"`
	ClassroomID  int64           `json:"classroom_id"`
	DayOfWeek    domain.DayOfWeek `json:"day_of_week"`
	TimeStart    string          `json:"time_start"`
	TimeEnd      string          `json:"time_end"`
	WeekType     domain.WeekType `json:"week_type"`
	DateStart    time.Time       `json:"date_start"`
	DateEnd      time.Time       `json:"date_end"`
	Notes        *string         `json:"notes,omitempty"`
	IsCancelled  bool            `json:"is_cancelled"`
	CancelReason *string         `json:"cancellation_reason,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`

	// Associations (loaded separately)
	Discipline *Discipline  `json:"discipline,omitempty"`
	LessonType *LessonType  `json:"lesson_type,omitempty"`
	Classroom  *Classroom   `json:"classroom,omitempty"`
	Group      *StudentGroup `json:"group,omitempty"`
	Teacher    *TeacherInfo `json:"teacher,omitempty"`
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
```

```go
// classroom.go
package entities

import "time"

type Classroom struct {
	ID          int64            `json:"id"`
	Building    string           `json:"building"`
	Number      string           `json:"number"`
	Name        *string          `json:"name,omitempty"`
	Capacity    int              `json:"capacity"`
	Type        *string          `json:"type,omitempty"`
	Equipment   map[string]any   `json:"equipment,omitempty"`
	IsAvailable bool             `json:"is_available"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}
```

```go
// schedule_change.go
package entities

import (
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain"
)

type ScheduleChange struct {
	ID             int64            `json:"id"`
	LessonID       int64            `json:"lesson_id"`
	ChangeType     domain.ChangeType `json:"change_type"`
	OriginalDate   time.Time        `json:"original_date"`
	NewDate        *time.Time       `json:"new_date,omitempty"`
	NewClassroomID *int64           `json:"new_classroom_id,omitempty"`
	NewTeacherID   *int64           `json:"new_teacher_id,omitempty"`
	Reason         *string          `json:"reason,omitempty"`
	CreatedBy      int64            `json:"created_by"`
	CreatedAt      time.Time        `json:"created_at"`
}
```

```go
// reference.go
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
```

**Step 4: Run tests — PASS**

**Step 5: Commits (RED then GREEN)**

---

### Task 3: Repository interfaces

**Files:**
- Create: `internal/modules/schedule/domain/repositories/lesson_repository.go`
- Create: `internal/modules/schedule/domain/repositories/classroom_repository.go`
- Create: `internal/modules/schedule/domain/repositories/reference_repository.go`
- Create: `internal/modules/schedule/domain/repositories/schedule_change_repository.go`

**Implementation (no test needed — interfaces only):**

```go
// lesson_repository.go
package repositories

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
)

type LessonFilter struct {
	SemesterID   *int64
	GroupID      *int64
	TeacherID    *int64
	ClassroomID  *int64
	DisciplineID *int64
	DayOfWeek    *domain.DayOfWeek
	WeekType     *domain.WeekType
}

type LessonRepository interface {
	Create(ctx context.Context, lesson *entities.Lesson) error
	Save(ctx context.Context, lesson *entities.Lesson) error
	GetByID(ctx context.Context, id int64) (*entities.Lesson, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, filter LessonFilter, limit, offset int) ([]*entities.Lesson, error)
	Count(ctx context.Context, filter LessonFilter) (int64, error)
	GetTimetable(ctx context.Context, filter LessonFilter) ([]*entities.Lesson, error)
}
```

```go
// classroom_repository.go
package repositories

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
)

type ClassroomFilter struct {
	Building    *string
	Type        *string
	MinCapacity *int
	IsAvailable *bool
}

type ClassroomRepository interface {
	GetByID(ctx context.Context, id int64) (*entities.Classroom, error)
	List(ctx context.Context, filter ClassroomFilter, limit, offset int) ([]*entities.Classroom, error)
	Count(ctx context.Context, filter ClassroomFilter) (int64, error)
}
```

```go
// reference_repository.go
package repositories

import (
	"context"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
)

type ReferenceRepository interface {
	ListStudentGroups(ctx context.Context, limit, offset int) ([]*entities.StudentGroup, error)
	ListDisciplines(ctx context.Context, limit, offset int) ([]*entities.Discipline, error)
	ListSemesters(ctx context.Context, activeOnly bool) ([]*entities.Semester, error)
	ListLessonTypes(ctx context.Context) ([]*entities.LessonType, error)
	GetActiveSemester(ctx context.Context) (*entities.Semester, error)
}
```

```go
// schedule_change_repository.go
package repositories

import (
	"context"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
)

type ScheduleChangeRepository interface {
	Create(ctx context.Context, change *entities.ScheduleChange) error
	GetByLessonID(ctx context.Context, lessonID int64) ([]*entities.ScheduleChange, error)
	GetByDateRange(ctx context.Context, start, end time.Time) ([]*entities.ScheduleChange, error)
}
```

**Commit:** `feat(schedule): add repository interfaces for lessons, classrooms, references, changes`

---

## Phase 2: Backend — Application Layer (DTOs + UseCases)

### Task 4: DTOs

**Files:**
- Create: `internal/modules/schedule/application/dto/lesson_dto.go`
- Test: `internal/modules/schedule/application/dto/lesson_dto_test.go`

**Key DTOs (implement with TDD):**

```go
// lesson_dto.go
package dto

import (
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/repositories"
)

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
	DateStart    string `json:"date_start" validate:"required"`
	DateEnd      string `json:"date_end" validate:"required"`
	Notes        *string `json:"notes,omitempty"`
}

type UpdateLessonInput struct {
	ClassroomID  *int64  `json:"classroom_id,omitempty"`
	TeacherID    *int64  `json:"teacher_id,omitempty"`
	DayOfWeek    *int    `json:"day_of_week,omitempty" validate:"omitempty,min=1,max=7"`
	TimeStart    *string `json:"time_start,omitempty"`
	TimeEnd      *string `json:"time_end,omitempty"`
	WeekType     *string `json:"week_type,omitempty" validate:"omitempty,oneof=all odd even"`
	Notes        *string `json:"notes,omitempty"`
}

type LessonFilterInput struct {
	SemesterID   *int64 `form:"semester_id"`
	GroupID      *int64 `form:"group_id"`
	TeacherID    *int64 `form:"teacher_id"`
	ClassroomID  *int64 `form:"classroom_id"`
	DisciplineID *int64 `form:"discipline_id"`
	DayOfWeek    *int   `form:"day_of_week"`
	Limit        int    `form:"limit,default=100"`
	Offset       int    `form:"offset,default=0"`
}

func (f *LessonFilterInput) ToFilter() repositories.LessonFilter {
	filter := repositories.LessonFilter{
		SemesterID:   f.SemesterID,
		GroupID:      f.GroupID,
		TeacherID:    f.TeacherID,
		ClassroomID:  f.ClassroomID,
		DisciplineID: f.DisciplineID,
	}
	if f.DayOfWeek != nil {
		day := domain.DayOfWeek(*f.DayOfWeek)
		filter.DayOfWeek = &day
	}
	return filter
}

type LessonOutput struct {
	ID           int64                `json:"id"`
	SemesterID   int64                `json:"semester_id"`
	DisciplineID int64                `json:"discipline_id"`
	LessonTypeID int64                `json:"lesson_type_id"`
	TeacherID    int64                `json:"teacher_id"`
	GroupID      int64                `json:"group_id"`
	ClassroomID  int64                `json:"classroom_id"`
	DayOfWeek    int                  `json:"day_of_week"`
	TimeStart    string               `json:"time_start"`
	TimeEnd      string               `json:"time_end"`
	WeekType     string               `json:"week_type"`
	DateStart    string               `json:"date_start"`
	DateEnd      string               `json:"date_end"`
	Notes        *string              `json:"notes,omitempty"`
	IsCancelled  bool                 `json:"is_cancelled"`
	CreatedAt    time.Time            `json:"created_at"`
	UpdatedAt    time.Time            `json:"updated_at"`
	Discipline   *DisciplineOutput    `json:"discipline,omitempty"`
	LessonType   *LessonTypeOutput    `json:"lesson_type,omitempty"`
	Classroom    *ClassroomOutput     `json:"classroom,omitempty"`
	Group        *StudentGroupOutput  `json:"group,omitempty"`
	Teacher      *TeacherOutput       `json:"teacher,omitempty"`
}

type ClassroomOutput struct {
	ID          int64  `json:"id"`
	Building    string `json:"building"`
	Number      string `json:"number"`
	Name        *string `json:"name,omitempty"`
	Capacity    int    `json:"capacity"`
	Type        *string `json:"type,omitempty"`
	IsAvailable bool   `json:"is_available"`
}

type StudentGroupOutput struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Course int    `json:"course"`
}

type DisciplineOutput struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Code *string `json:"code,omitempty"`
}

type LessonTypeOutput struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	ShortName string  `json:"short_name"`
	Color     *string `json:"color,omitempty"`
}

type TeacherOutput struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type SemesterOutput struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Number    int    `json:"number"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	IsActive  bool   `json:"is_active"`
}

type CreateChangeInput struct {
	LessonID       int64   `json:"lesson_id" validate:"required"`
	ChangeType     string  `json:"change_type" validate:"required,oneof=cancelled moved replaced_teacher replaced_classroom"`
	OriginalDate   string  `json:"original_date" validate:"required"`
	NewDate        *string `json:"new_date,omitempty"`
	NewClassroomID *int64  `json:"new_classroom_id,omitempty"`
	NewTeacherID   *int64  `json:"new_teacher_id,omitempty"`
	Reason         *string `json:"reason,omitempty"`
}

type LessonListOutput struct {
	Lessons []LessonOutput `json:"lessons"`
	Total   int64          `json:"total"`
	Limit   int            `json:"limit"`
	Offset  int            `json:"offset"`
}

func ToLessonOutput(l *entities.Lesson) LessonOutput {
	out := LessonOutput{
		ID:           l.ID,
		SemesterID:   l.SemesterID,
		DisciplineID: l.DisciplineID,
		LessonTypeID: l.LessonTypeID,
		TeacherID:    l.TeacherID,
		GroupID:      l.GroupID,
		ClassroomID:  l.ClassroomID,
		DayOfWeek:    int(l.DayOfWeek),
		TimeStart:    l.TimeStart,
		TimeEnd:      l.TimeEnd,
		WeekType:     string(l.WeekType),
		DateStart:    l.DateStart.Format("2006-01-02"),
		DateEnd:      l.DateEnd.Format("2006-01-02"),
		Notes:        l.Notes,
		IsCancelled:  l.IsCancelled,
		CreatedAt:    l.CreatedAt,
		UpdatedAt:    l.UpdatedAt,
	}
	if l.Discipline != nil {
		out.Discipline = &DisciplineOutput{ID: l.Discipline.ID, Name: l.Discipline.Name, Code: l.Discipline.Code}
	}
	if l.LessonType != nil {
		out.LessonType = &LessonTypeOutput{ID: l.LessonType.ID, Name: l.LessonType.Name, ShortName: l.LessonType.ShortName, Color: l.LessonType.Color}
	}
	if l.Classroom != nil {
		out.Classroom = &ClassroomOutput{ID: l.Classroom.ID, Building: l.Classroom.Building, Number: l.Classroom.Number, Name: l.Classroom.Name, Capacity: l.Classroom.Capacity, Type: l.Classroom.Type, IsAvailable: l.Classroom.IsAvailable}
	}
	if l.Group != nil {
		out.Group = &StudentGroupOutput{ID: l.Group.ID, Name: l.Group.Name, Course: l.Group.Course}
	}
	if l.Teacher != nil {
		out.Teacher = &TeacherOutput{ID: l.Teacher.ID, Name: l.Teacher.Name, Email: l.Teacher.Email}
	}
	return out
}
```

**Test:** write test for `ToLessonOutput` and `ToFilter` conversion.

---

### Task 5: Lesson UseCase

**Files:**
- Create: `internal/modules/schedule/application/usecases/lesson_usecase.go`
- Create: `internal/modules/schedule/application/usecases/lesson_usecase_test.go`
- Create: `internal/modules/schedule/application/usecases/mock_lesson_repos_test.go`

**Key structure:**

```go
// lesson_usecase.go
package usecases

import (
	"context"
	"errors"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/application/dto"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain/repositories"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/logging"
)

var (
	ErrLessonNotFound = errors.New("lesson not found")
	ErrForbidden      = errors.New("forbidden")
	ErrInvalidInput   = errors.New("invalid input")
)

type LessonUseCase struct {
	lessonRepo    repositories.LessonRepository
	classroomRepo repositories.ClassroomRepository
	referenceRepo repositories.ReferenceRepository
	changeRepo    repositories.ScheduleChangeRepository
	auditLogger   *logging.AuditLogger
}

func NewLessonUseCase(
	lessonRepo repositories.LessonRepository,
	classroomRepo repositories.ClassroomRepository,
	referenceRepo repositories.ReferenceRepository,
	changeRepo repositories.ScheduleChangeRepository,
	auditLogger *logging.AuditLogger,
) *LessonUseCase {
	return &LessonUseCase{
		lessonRepo:    lessonRepo,
		classroomRepo: classroomRepo,
		referenceRepo: referenceRepo,
		changeRepo:    changeRepo,
		auditLogger:   auditLogger,
	}
}

func (uc *LessonUseCase) Create(ctx context.Context, userID int64, input dto.CreateLessonInput) (*entities.Lesson, error) {
	dateStart, err := time.Parse("2006-01-02", input.DateStart)
	if err != nil {
		return nil, ErrInvalidInput
	}
	dateEnd, err := time.Parse("2006-01-02", input.DateEnd)
	if err != nil {
		return nil, ErrInvalidInput
	}

	lesson := entities.NewLesson(
		input.SemesterID, input.DisciplineID, input.LessonTypeID,
		input.TeacherID, input.GroupID, input.ClassroomID,
		domain.DayOfWeek(input.DayOfWeek), input.TimeStart, input.TimeEnd,
		domain.WeekType(input.WeekType),
	)
	lesson.DateStart = dateStart
	lesson.DateEnd = dateEnd
	lesson.Notes = input.Notes

	if err := lesson.Validate(); err != nil {
		return nil, err
	}

	if err := uc.lessonRepo.Create(ctx, lesson); err != nil {
		return nil, err
	}

	uc.auditLogger.LogAction(ctx, userID, "lesson_created", lesson.ID)
	return lesson, nil
}

func (uc *LessonUseCase) GetByID(ctx context.Context, id int64) (*entities.Lesson, error) {
	lesson, err := uc.lessonRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if lesson == nil {
		return nil, ErrLessonNotFound
	}
	return lesson, nil
}

func (uc *LessonUseCase) Update(ctx context.Context, userID, lessonID int64, input dto.UpdateLessonInput) (*entities.Lesson, error) {
	lesson, err := uc.lessonRepo.GetByID(ctx, lessonID)
	if err != nil {
		return nil, err
	}
	if lesson == nil {
		return nil, ErrLessonNotFound
	}

	if input.ClassroomID != nil {
		lesson.ClassroomID = *input.ClassroomID
	}
	if input.TeacherID != nil {
		lesson.TeacherID = *input.TeacherID
	}
	if input.DayOfWeek != nil {
		lesson.DayOfWeek = domain.DayOfWeek(*input.DayOfWeek)
	}
	if input.TimeStart != nil {
		lesson.TimeStart = *input.TimeStart
	}
	if input.TimeEnd != nil {
		lesson.TimeEnd = *input.TimeEnd
	}
	if input.WeekType != nil {
		lesson.WeekType = domain.WeekType(*input.WeekType)
	}
	if input.Notes != nil {
		lesson.Notes = input.Notes
	}
	lesson.UpdatedAt = time.Now()

	if err := lesson.Validate(); err != nil {
		return nil, err
	}

	if err := uc.lessonRepo.Save(ctx, lesson); err != nil {
		return nil, err
	}

	uc.auditLogger.LogAction(ctx, userID, "lesson_updated", lesson.ID)
	return lesson, nil
}

func (uc *LessonUseCase) Delete(ctx context.Context, userID, lessonID int64) error {
	lesson, err := uc.lessonRepo.GetByID(ctx, lessonID)
	if err != nil {
		return err
	}
	if lesson == nil {
		return ErrLessonNotFound
	}

	if err := uc.lessonRepo.Delete(ctx, lessonID); err != nil {
		return err
	}

	uc.auditLogger.LogAction(ctx, userID, "lesson_deleted", lessonID)
	return nil
}

func (uc *LessonUseCase) List(ctx context.Context, filter repositories.LessonFilter, limit, offset int) ([]*entities.Lesson, error) {
	return uc.lessonRepo.List(ctx, filter, limit, offset)
}

func (uc *LessonUseCase) Count(ctx context.Context, filter repositories.LessonFilter) (int64, error) {
	return uc.lessonRepo.Count(ctx, filter)
}

func (uc *LessonUseCase) GetTimetable(ctx context.Context, filter repositories.LessonFilter) ([]*entities.Lesson, error) {
	return uc.lessonRepo.GetTimetable(ctx, filter)
}

func (uc *LessonUseCase) CreateChange(ctx context.Context, userID int64, input dto.CreateChangeInput) (*entities.ScheduleChange, error) {
	originalDate, err := time.Parse("2006-01-02", input.OriginalDate)
	if err != nil {
		return nil, ErrInvalidInput
	}

	change := &entities.ScheduleChange{
		LessonID:       input.LessonID,
		ChangeType:     domain.ChangeType(input.ChangeType),
		OriginalDate:   originalDate,
		NewClassroomID: input.NewClassroomID,
		NewTeacherID:   input.NewTeacherID,
		Reason:         input.Reason,
		CreatedBy:      userID,
		CreatedAt:      time.Now(),
	}

	if input.NewDate != nil {
		nd, err := time.Parse("2006-01-02", *input.NewDate)
		if err != nil {
			return nil, ErrInvalidInput
		}
		change.NewDate = &nd
	}

	if err := uc.changeRepo.Create(ctx, change); err != nil {
		return nil, err
	}

	uc.auditLogger.LogAction(ctx, userID, "schedule_change_created", change.ID)
	return change, nil
}

// Reference data methods
func (uc *LessonUseCase) ListClassrooms(ctx context.Context, filter repositories.ClassroomFilter, limit, offset int) ([]*entities.Classroom, error) {
	return uc.classroomRepo.List(ctx, filter, limit, offset)
}

func (uc *LessonUseCase) ListStudentGroups(ctx context.Context, limit, offset int) ([]*entities.StudentGroup, error) {
	return uc.referenceRepo.ListStudentGroups(ctx, limit, offset)
}

func (uc *LessonUseCase) ListDisciplines(ctx context.Context, limit, offset int) ([]*entities.Discipline, error) {
	return uc.referenceRepo.ListDisciplines(ctx, limit, offset)
}

func (uc *LessonUseCase) ListSemesters(ctx context.Context, activeOnly bool) ([]*entities.Semester, error) {
	return uc.referenceRepo.ListSemesters(ctx, activeOnly)
}

func (uc *LessonUseCase) ListLessonTypes(ctx context.Context) ([]*entities.LessonType, error) {
	return uc.referenceRepo.ListLessonTypes(ctx)
}
```

**Test TDD:** mock repositories, test Create with validation, test Update partial fields, test Delete not-found case.

---

## Phase 3: Backend — Infrastructure + Handlers

### Task 6: PostgreSQL Repository Implementations

**Files:**
- Create: `internal/modules/schedule/infrastructure/persistence/lesson_repository_pg.go`
- Create: `internal/modules/schedule/infrastructure/persistence/classroom_repository_pg.go`
- Create: `internal/modules/schedule/infrastructure/persistence/reference_repository_pg.go`
- Create: `internal/modules/schedule/infrastructure/persistence/schedule_change_repository_pg.go`

These implement the interfaces from Task 3. Use `database/sql` with raw queries (project pattern — no ORM). `GetTimetable` does a JOIN to load associations.

---

### Task 7: HTTP Handler

**Files:**
- Create: `internal/modules/schedule/interfaces/http/handlers/lesson_handler.go`
- Test: `internal/modules/schedule/interfaces/http/handlers/lesson_handler_test.go`

**Routes to register in `cmd/server/main.go`:**

```go
// Schedule lessons routes
scheduleGroup := protectedGroup.Group("/schedule")
{
    scheduleGroup.POST("/lessons", lessonHandlerInstance.Create)
    scheduleGroup.GET("/lessons", lessonHandlerInstance.List)
    scheduleGroup.GET("/lessons/timetable", lessonHandlerInstance.GetTimetable)
    scheduleGroup.GET("/lessons/:id", lessonHandlerInstance.GetByID)
    scheduleGroup.PUT("/lessons/:id", lessonHandlerInstance.Update)
    scheduleGroup.DELETE("/lessons/:id", lessonHandlerInstance.Delete)
    scheduleGroup.POST("/changes", lessonHandlerInstance.CreateChange)
    scheduleGroup.GET("/changes", lessonHandlerInstance.ListChanges)
}

// Reference data (read-only)
classroomsGroup := protectedGroup.Group("/classrooms")
{
    classroomsGroup.GET("", lessonHandlerInstance.ListClassrooms)
}
studentGroupsGroup := protectedGroup.Group("/student-groups")
{
    studentGroupsGroup.GET("", lessonHandlerInstance.ListStudentGroups)
}
disciplinesGroup := protectedGroup.Group("/disciplines")
{
    disciplinesGroup.GET("", lessonHandlerInstance.ListDisciplines)
}
semestersGroup := protectedGroup.Group("/semesters")
{
    semestersGroup.GET("", lessonHandlerInstance.ListSemesters)
}
lessonTypesGroup := protectedGroup.Group("/lesson-types")
{
    lessonTypesGroup.GET("", lessonHandlerInstance.ListLessonTypes)
}
```

---

## Phase 4: Frontend — Types + API + Hooks

### Task 8: TypeScript types

**File:** `frontend/src/types/schedule.ts`

```typescript
export interface Lesson {
  id: number
  semester_id: number
  discipline_id: number
  lesson_type_id: number
  teacher_id: number
  group_id: number
  classroom_id: number
  day_of_week: number
  time_start: string
  time_end: string
  week_type: 'all' | 'odd' | 'even'
  date_start: string
  date_end: string
  notes?: string
  is_cancelled: boolean
  created_at: string
  updated_at: string
  discipline?: Discipline
  lesson_type?: LessonTypeInfo
  classroom?: Classroom
  group?: StudentGroup
  teacher?: TeacherInfo
}

export interface Classroom {
  id: number
  building: string
  number: string
  name?: string
  capacity: number
  type?: string
  is_available: boolean
}

export interface StudentGroup {
  id: number
  name: string
  course: number
}

export interface Discipline {
  id: number
  name: string
  code?: string
}

export interface LessonTypeInfo {
  id: number
  name: string
  short_name: string
  color?: string
}

export interface Semester {
  id: number
  name: string
  number: number
  start_date: string
  end_date: string
  is_active: boolean
}

export interface TeacherInfo {
  id: number
  name: string
  email: string
}

export interface ScheduleChange {
  id: number
  lesson_id: number
  change_type: 'cancelled' | 'moved' | 'replaced_teacher' | 'replaced_classroom'
  original_date: string
  new_date?: string
  new_classroom_id?: number
  new_teacher_id?: number
  reason?: string
  created_by: number
  created_at: string
}

export interface CreateLessonInput {
  semester_id: number
  discipline_id: number
  lesson_type_id: number
  teacher_id: number
  group_id: number
  classroom_id: number
  day_of_week: number
  time_start: string
  time_end: string
  week_type: 'all' | 'odd' | 'even'
  date_start: string
  date_end: string
  notes?: string
}

export interface LessonFilterParams {
  semester_id?: number
  group_id?: number
  teacher_id?: number
  classroom_id?: number
  discipline_id?: number
  day_of_week?: number
}
```

---

### Task 9: API client + SWR hooks

**Files:**
- Create: `frontend/src/lib/api/schedule.ts`
- Create: `frontend/src/hooks/useSchedule.ts`

Standard SWR pattern: `useSWR` for GET, mutation functions for POST/PUT/DELETE.

---

## Phase 5: Frontend — Components + Page + i18n

### Task 10: Components (TimetableGrid, LessonCard, ScheduleFilters)

**Files:**
- Create: `frontend/src/components/schedule/TimetableGrid.tsx`
- Create: `frontend/src/components/schedule/LessonCard.tsx`
- Create: `frontend/src/components/schedule/ScheduleFilters.tsx`
- Create: `frontend/src/components/schedule/ScheduleChangeForm.tsx`
- Create: `frontend/src/components/schedule/index.ts`
- Tests: `frontend/src/components/schedule/__tests__/`

**TimetableGrid:** 6 columns (Mon-Sat) × N time slots (rows). Each cell contains LessonCards for that day+time.

**LessonCard:** colored by lesson_type.color, shows discipline short name, teacher, classroom, group.

**ScheduleFilters:** semester select, group select, teacher select, classroom select.

---

### Task 11: Page `/schedule` + i18n ×4

**Files:**
- Modify: `frontend/src/app/schedule/page.tsx` (replace placeholder)
- Modify: `frontend/src/messages/ru.json` (add `schedule` namespace)
- Modify: `frontend/src/messages/en.json`
- Modify: `frontend/src/messages/fr.json`
- Modify: `frontend/src/messages/ar.json`

**Page logic:**
- Use `can(role, Resource.SCHEDULE, Action.CREATE)` for edit mode
- Default filter: active semester + user's group (student) or teacher (teacher)
- Grid view with ScheduleFilters on top
- Create/Edit dialog for secretary/admin
- ScheduleChangeForm for adding cancellations/moves

---

## Phase 6: Release

### Task 12: Version bump + docs update + release

- Bump version 0.104.0 → 0.105.0 in 8 files
- CHANGELOG entry
- Update `docs/roles-and-flows.md`: schedule module status → "работает полностью"
- Tag + GH Release
- Close GH #201

---

## Summary

| Phase | Tasks | Scope |
|-------|-------|-------|
| 1 | 1-3 | Domain: types, entities, repository interfaces |
| 2 | 4-5 | Application: DTOs, usecases |
| 3 | 6-7 | Infrastructure: PG repos, HTTP handlers, DI wiring |
| 4 | 8-9 | Frontend: types, API, hooks |
| 5 | 10-11 | Frontend: components, page, i18n |
| 6 | 12 | Release |

**Estimated:** ~12 TDD commit pairs (RED+GREEN) + integration commits. Large task, 2-3 hours.
