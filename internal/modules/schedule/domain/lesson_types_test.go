package domain

import "testing"

func TestDayOfWeek_IsValid(t *testing.T) {
	tests := []struct {
		day  DayOfWeek
		want bool
	}{
		{Monday, true},
		{Tuesday, true},
		{Wednesday, true},
		{Thursday, true},
		{Friday, true},
		{Saturday, true},
		{Sunday, true},
		{DayOfWeek(0), false},
		{DayOfWeek(8), false},
		{DayOfWeek(-1), false},
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
		{WeekType(""), false},
		{WeekType("invalid"), false},
		{WeekType("ALL"), false},
	}
	for _, tt := range tests {
		if got := tt.wt.IsValid(); got != tt.want {
			t.Errorf("WeekType(%q).IsValid() = %v, want %v", tt.wt, got, tt.want)
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
		{ChangeType(""), false},
		{ChangeType("invalid"), false},
	}
	for _, tt := range tests {
		if got := tt.ct.IsValid(); got != tt.want {
			t.Errorf("ChangeType(%q).IsValid() = %v, want %v", tt.ct, got, tt.want)
		}
	}
}
