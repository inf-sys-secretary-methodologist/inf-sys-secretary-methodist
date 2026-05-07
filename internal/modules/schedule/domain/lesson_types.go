package domain

// DayOfWeek represents day of week (1=Monday, 7=Sunday) matching PostgreSQL convention.
type DayOfWeek int

// DayOfWeek values matching the PostgreSQL ISO convention (Monday=1 … Sunday=7).
const (
	Monday    DayOfWeek = 1
	Tuesday   DayOfWeek = 2
	Wednesday DayOfWeek = 3
	Thursday  DayOfWeek = 4
	Friday    DayOfWeek = 5
	Saturday  DayOfWeek = 6
	Sunday    DayOfWeek = 7
)

// IsValid reports whether d is one of the seven defined weekdays.
func (d DayOfWeek) IsValid() bool {
	return d >= Monday && d <= Sunday
}

// WeekType represents which weeks the lesson occurs on.
type WeekType string

// WeekType values: every week, odd weeks only, or even weeks only.
const (
	WeekTypeAll  WeekType = "all"
	WeekTypeOdd  WeekType = "odd"
	WeekTypeEven WeekType = "even"
)

// IsValid reports whether w is a recognized week-type value.
func (w WeekType) IsValid() bool {
	switch w {
	case WeekTypeAll, WeekTypeOdd, WeekTypeEven:
		return true
	}
	return false
}

// ChangeType represents type of schedule change.
type ChangeType string

// ChangeType values describing a one-off modification to a recurring lesson.
const (
	ChangeTypeCancelled         ChangeType = "canceled"
	ChangeTypeMoved             ChangeType = "moved"
	ChangeTypeReplacedTeacher   ChangeType = "replaced_teacher"
	ChangeTypeReplacedClassroom ChangeType = "replaced_classroom"
)

// IsValid reports whether c is a recognized change-type value.
func (c ChangeType) IsValid() bool {
	switch c {
	case ChangeTypeCancelled, ChangeTypeMoved, ChangeTypeReplacedTeacher, ChangeTypeReplacedClassroom:
		return true
	}
	return false
}
