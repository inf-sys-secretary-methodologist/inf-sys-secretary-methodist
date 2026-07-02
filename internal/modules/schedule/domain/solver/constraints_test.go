package solver

import (
	"testing"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain"
)

func TestParityConflicts(t *testing.T) {
	tests := []struct {
		name string
		a, b domain.WeekType
		want bool
	}{
		{"all vs all", domain.WeekTypeAll, domain.WeekTypeAll, true},
		{"all vs odd", domain.WeekTypeAll, domain.WeekTypeOdd, true},
		{"all vs even", domain.WeekTypeAll, domain.WeekTypeEven, true},
		{"odd vs all", domain.WeekTypeOdd, domain.WeekTypeAll, true},
		{"even vs all", domain.WeekTypeEven, domain.WeekTypeAll, true},
		{"odd vs odd", domain.WeekTypeOdd, domain.WeekTypeOdd, true},
		{"even vs even", domain.WeekTypeEven, domain.WeekTypeEven, true},
		{"odd vs even", domain.WeekTypeOdd, domain.WeekTypeEven, false},
		{"even vs odd", domain.WeekTypeEven, domain.WeekTypeOdd, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parityConflicts(tt.a, tt.b); got != tt.want {
				t.Errorf("parityConflicts(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestAssignmentsConflict(t *testing.T) {
	// base assignment: group 10, teacher 20, room 30, Monday slot 1, every week.
	base := Assignment{
		Variable: Variable{GroupID: 10, TeacherID: 20, WeekType: domain.WeekTypeAll},
		Value:    Value{Day: domain.Monday, Slot: 1, RoomID: 30},
	}

	tests := []struct {
		name  string
		other Assignment
		want  bool
	}{
		{
			name: "different day never conflicts",
			other: Assignment{
				Variable: Variable{GroupID: 10, TeacherID: 20, WeekType: domain.WeekTypeAll},
				Value:    Value{Day: domain.Tuesday, Slot: 1, RoomID: 30},
			},
			want: false,
		},
		{
			name: "different slot never conflicts",
			other: Assignment{
				Variable: Variable{GroupID: 10, TeacherID: 20, WeekType: domain.WeekTypeAll},
				Value:    Value{Day: domain.Monday, Slot: 2, RoomID: 30},
			},
			want: false,
		},
		{
			name: "same slot, shared teacher conflicts",
			other: Assignment{
				Variable: Variable{GroupID: 99, TeacherID: 20, WeekType: domain.WeekTypeAll},
				Value:    Value{Day: domain.Monday, Slot: 1, RoomID: 77},
			},
			want: true,
		},
		{
			name: "same slot, shared group conflicts",
			other: Assignment{
				Variable: Variable{GroupID: 10, TeacherID: 88, WeekType: domain.WeekTypeAll},
				Value:    Value{Day: domain.Monday, Slot: 1, RoomID: 77},
			},
			want: true,
		},
		{
			name: "same slot, shared room conflicts",
			other: Assignment{
				Variable: Variable{GroupID: 99, TeacherID: 88, WeekType: domain.WeekTypeAll},
				Value:    Value{Day: domain.Monday, Slot: 1, RoomID: 30},
			},
			want: true,
		},
		{
			name: "same slot, no shared resource does not conflict",
			other: Assignment{
				Variable: Variable{GroupID: 99, TeacherID: 88, WeekType: domain.WeekTypeAll},
				Value:    Value{Day: domain.Monday, Slot: 1, RoomID: 77},
			},
			want: false,
		},
		{
			name: "same slot + shared room, base is all-weeks, overlaps even-week lesson",
			other: Assignment{
				Variable: Variable{GroupID: 99, TeacherID: 88, WeekType: domain.WeekTypeEven},
				Value:    Value{Day: domain.Monday, Slot: 1, RoomID: 30},
			},
			want: true, // "all" overlaps every week-type; shared room -> conflict
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := assignmentsConflict(base, tt.other); got != tt.want {
				t.Errorf("assignmentsConflict() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAssignmentsConflict_ParityDisjoint(t *testing.T) {
	// Two lessons sharing a room on the same day+slot but on odd vs even weeks
	// never physically collide.
	odd := Assignment{
		Variable: Variable{GroupID: 1, TeacherID: 1, WeekType: domain.WeekTypeOdd},
		Value:    Value{Day: domain.Monday, Slot: 1, RoomID: 5},
	}
	even := Assignment{
		Variable: Variable{GroupID: 2, TeacherID: 2, WeekType: domain.WeekTypeEven},
		Value:    Value{Day: domain.Monday, Slot: 1, RoomID: 5},
	}
	if assignmentsConflict(odd, even) {
		t.Error("odd-week and even-week lessons sharing a room must not conflict")
	}

	// Same room, same odd week -> conflict.
	odd2 := Assignment{
		Variable: Variable{GroupID: 2, TeacherID: 2, WeekType: domain.WeekTypeOdd},
		Value:    Value{Day: domain.Monday, Slot: 1, RoomID: 5},
	}
	if !assignmentsConflict(odd, odd2) {
		t.Error("two odd-week lessons sharing a room must conflict")
	}
}
