package solver

import (
	"testing"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain"
)

func TestNewDefaultWeights_AllPositive(t *testing.T) {
	w := NewDefaultWeights()
	if w.GroupGap <= 0 || w.TeacherGap <= 0 || w.DaySpread <= 0 || w.EarlySlot <= 0 {
		t.Errorf("all default weights must be positive, got %+v", w)
	}
}

func TestGapCount(t *testing.T) {
	tests := []struct {
		name  string
		slots []int
		want  int
	}{
		{"empty", nil, 0},
		{"single", []int{3}, 0},
		{"contiguous", []int{1, 2, 3}, 0},
		{"one window", []int{1, 3}, 1},
		{"two windows", []int{2, 5}, 2},
		{"unordered contiguous", []int{3, 1, 2}, 0},
		{"unordered with window", []int{4, 1}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := gapCount(tt.slots); got != tt.want {
				t.Errorf("gapCount(%v) = %d, want %d", tt.slots, got, tt.want)
			}
		})
	}
}

func TestPenalty_EarlySlotPreferred(t *testing.T) {
	w := SoftWeights{EarlySlot: 1} // isolate the early-slot term
	v := Variable{GroupID: 1, TeacherID: 1}
	early := Assignment{Variable: v, Value: Value{Day: domain.Monday, Slot: 1}}
	late := Assignment{Variable: v, Value: Value{Day: domain.Monday, Slot: 4}}

	if penalty(early, nil, w) >= penalty(late, nil, w) {
		t.Error("earlier slot must have strictly lower penalty when EarlySlot dominates")
	}
}

func TestPenalty_GroupGapPreferred(t *testing.T) {
	w := SoftWeights{GroupGap: 1} // isolate the group-gap term
	v := Variable{GroupID: 1, TeacherID: 1}
	// Group already has a lesson at Monday slot 1.
	current := []Assignment{{Variable: v, Value: Value{Day: domain.Monday, Slot: 1}}}

	adjacent := Assignment{Variable: v, Value: Value{Day: domain.Monday, Slot: 2}} // no window
	gapped := Assignment{Variable: v, Value: Value{Day: domain.Monday, Slot: 4}}   // windows at 2,3

	if penalty(adjacent, current, w) >= penalty(gapped, current, w) {
		t.Error("adjacent placement must beat a placement that opens a window")
	}
}

func TestPenalty_TeacherGapCounted(t *testing.T) {
	w := SoftWeights{TeacherGap: 1}
	// Same teacher, different groups: a shared teacher already teaches at slot 1.
	current := []Assignment{{
		Variable: Variable{GroupID: 1, TeacherID: 7},
		Value:    Value{Day: domain.Monday, Slot: 1},
	}}
	cand := Variable{GroupID: 2, TeacherID: 7}
	adjacent := Assignment{Variable: cand, Value: Value{Day: domain.Monday, Slot: 2}}
	gapped := Assignment{Variable: cand, Value: Value{Day: domain.Monday, Slot: 5}}

	if penalty(adjacent, current, w) >= penalty(gapped, current, w) {
		t.Error("teacher-gap term must penalize windows in the teacher's day")
	}
}

func TestPenalty_ParityDisjointNoPhantomGap(t *testing.T) {
	// An odd-week lesson and an even-week lesson never share a physical week, so
	// placing an even-week lesson at slot 3 while an odd-week lesson sits at slot 1
	// must NOT be scored as a window at slot 2 — the two are on different weeks.
	w := SoftWeights{GroupGap: 1, TeacherGap: 1, DaySpread: 1}
	oddCurrent := []Assignment{{
		Variable: Variable{GroupID: 1, TeacherID: 1, WeekType: domain.WeekTypeOdd},
		Value:    Value{Day: domain.Monday, Slot: 1},
	}}
	evenCandidate := Assignment{
		Variable: Variable{GroupID: 1, TeacherID: 1, WeekType: domain.WeekTypeEven},
		Value:    Value{Day: domain.Monday, Slot: 3},
	}

	if got := penalty(evenCandidate, oddCurrent, w); got != 0 {
		t.Errorf("parity-disjoint lessons must not incur gap/spread penalty, got %v", got)
	}
}

func TestPenalty_DaySpreadCounted(t *testing.T) {
	w := SoftWeights{DaySpread: 1}
	v := Variable{GroupID: 1, TeacherID: 1}
	// Group already has two lessons on Monday, none on Tuesday.
	current := []Assignment{
		{Variable: v, Value: Value{Day: domain.Monday, Slot: 1}},
		{Variable: v, Value: Value{Day: domain.Monday, Slot: 2}},
	}
	sameDay := Assignment{Variable: v, Value: Value{Day: domain.Monday, Slot: 3}}
	otherDay := Assignment{Variable: v, Value: Value{Day: domain.Tuesday, Slot: 1}}

	if penalty(otherDay, current, w) >= penalty(sameDay, current, w) {
		t.Error("spreading to a lighter day must beat piling onto a loaded day")
	}
}
