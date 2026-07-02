// Package solver is a pure, I/O-free constraint-satisfaction engine that places
// teaching-load items into a weekly timetable. It knows nothing about databases,
// HTTP, or persistence: callers unfold their domain data into Variables/Rooms,
// run Solve, and map the Result back. All hard rules (H1-H4) and soft preferences
// are enforced here so the engine stays fully unit-testable.
package solver

import "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain"

// Variable is a single lesson that must be placed on the timetable: one weekly
// occurrence of a group studying a discipline with a teacher, for a lesson type.
// AllowedRoomTypes lists the room types compatible with the lesson type (empty
// means any room type); it is computed by the caller (Slice 4), the solver only
// checks it. The engine treats every Variable as needing exactly one Value.
type Variable struct {
	ID               int
	LoadID           int64
	GroupID          int64
	TeacherID        int64
	DisciplineID     int64
	LessonTypeID     int64
	GroupSize        int
	AllowedRoomTypes []string
	WeekType         domain.WeekType
}

// Value is a concrete timetable placement: which day, which lesson slot (bell),
// and which room.
type Value struct {
	Day    domain.DayOfWeek
	Slot   int
	RoomID int64
}

// Room is an available teaching space with a capacity and a type. Unavailable
// rooms are excluded from every variable's domain.
type Room struct {
	ID        int64
	Capacity  int
	Type      string
	Available bool
}

// SoftWeights tunes the relative importance of the four soft preferences.
type SoftWeights struct {
	GroupGap   float64
	TeacherGap float64
	DaySpread  float64
	EarlySlot  float64
}

// Input is the full problem statement handed to Solve.
type Input struct {
	Variables []Variable
	Days      []domain.DayOfWeek
	Slots     []int
	Rooms     []Room
	Weights   SoftWeights
}

// Assignment binds a Variable to the Value chosen for it.
type Assignment struct {
	Variable Variable
	Value    Value
}

// Result is the outcome of Solve: the placed assignments plus any variables the
// engine could not place (best-effort, never a hard failure).
type Result struct {
	Assignments []Assignment
	Unplaced    []Variable
}
