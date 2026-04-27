package domain

// DayOfWeek represents day of week (1=Monday, 7=Sunday) matching PostgreSQL convention.
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
