package solver

import (
	"reflect"
	"testing"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain"
)

// noPairwiseConflict asserts that no two placed assignments violate a hard rule.
func noPairwiseConflict(t *testing.T, got []Assignment) {
	t.Helper()
	for i := range len(got) {
		for j := i + 1; j < len(got); j++ {
			if assignmentsConflict(got[i], got[j]) {
				t.Errorf("hard-constraint conflict between placed assignments %+v and %+v", got[i], got[j])
			}
		}
	}
}

func TestSolve_FullSolutionFeasible(t *testing.T) {
	in := Input{
		Variables: []Variable{
			{ID: 1, GroupID: 1, TeacherID: 1, GroupSize: 20, WeekType: domain.WeekTypeAll},
			{ID: 2, GroupID: 2, TeacherID: 2, GroupSize: 20, WeekType: domain.WeekTypeAll},
		},
		Days:    []domain.DayOfWeek{domain.Monday, domain.Tuesday},
		Slots:   []int{1, 2},
		Rooms:   []Room{{ID: 10, Capacity: 30, Type: "lecture", Available: true}},
		Weights: NewDefaultWeights(),
	}

	res := Solve(in)

	if len(res.Assignments) != 2 {
		t.Fatalf("expected 2 placed assignments, got %d (%+v)", len(res.Assignments), res.Assignments)
	}
	if len(res.Unplaced) != 0 {
		t.Errorf("expected no unplaced variables, got %+v", res.Unplaced)
	}
	noPairwiseConflict(t, res.Assignments)
}

func TestSolve_BestEffortPartialOnOverload(t *testing.T) {
	// Three lessons of the SAME group must share the only day/slot/room — at most
	// one can be placed; the engine must not fail and must account for all three.
	in := Input{
		Variables: []Variable{
			{ID: 1, GroupID: 1, TeacherID: 1, GroupSize: 20, WeekType: domain.WeekTypeAll},
			{ID: 2, GroupID: 1, TeacherID: 2, GroupSize: 20, WeekType: domain.WeekTypeAll},
			{ID: 3, GroupID: 1, TeacherID: 3, GroupSize: 20, WeekType: domain.WeekTypeAll},
		},
		Days:    []domain.DayOfWeek{domain.Monday},
		Slots:   []int{1},
		Rooms:   []Room{{ID: 10, Capacity: 30, Type: "lecture", Available: true}},
		Weights: NewDefaultWeights(),
	}

	res := Solve(in)

	if got := len(res.Assignments) + len(res.Unplaced); got != 3 {
		t.Fatalf("every variable must be placed or unplaced: placed=%d unplaced=%d",
			len(res.Assignments), len(res.Unplaced))
	}
	if len(res.Assignments) != 1 {
		t.Errorf("only one same-group lesson fits the single slot, got %d placed", len(res.Assignments))
	}
	noPairwiseConflict(t, res.Assignments)
}

func TestSolve_EmptyDomainGoesUnplaced(t *testing.T) {
	in := Input{
		Variables: []Variable{
			{ID: 1, GroupID: 1, TeacherID: 1, GroupSize: 20, WeekType: domain.WeekTypeAll},
			{ID: 2, GroupID: 2, TeacherID: 2, GroupSize: 500, WeekType: domain.WeekTypeAll}, // no room fits
		},
		Days:    []domain.DayOfWeek{domain.Monday, domain.Tuesday},
		Slots:   []int{1, 2},
		Rooms:   []Room{{ID: 10, Capacity: 30, Type: "lecture", Available: true}},
		Weights: NewDefaultWeights(),
	}

	res := Solve(in)

	if len(res.Unplaced) != 1 || res.Unplaced[0].ID != 2 {
		t.Fatalf("variable 2 (no fitting room) must be unplaced, got unplaced=%+v", res.Unplaced)
	}
	if len(res.Assignments) != 1 || res.Assignments[0].Variable.ID != 1 {
		t.Errorf("variable 1 must still be placed, got %+v", res.Assignments)
	}
}

func TestSolve_PrefersEarlierSlot(t *testing.T) {
	in := Input{
		Variables: []Variable{
			{ID: 1, GroupID: 1, TeacherID: 1, GroupSize: 20, WeekType: domain.WeekTypeAll},
		},
		Days:    []domain.DayOfWeek{domain.Monday},
		Slots:   []int{1, 2, 3},
		Rooms:   []Room{{ID: 10, Capacity: 30, Type: "lecture", Available: true}},
		Weights: SoftWeights{EarlySlot: 1},
	}

	res := Solve(in)

	if len(res.Assignments) != 1 {
		t.Fatalf("expected the single variable to be placed, got %+v", res.Assignments)
	}
	if got := res.Assignments[0].Value.Slot; got != 1 {
		t.Errorf("EarlySlot preference must pick slot 1, got slot %d", got)
	}
}

func TestSolve_ExactDeterministicPlacement(t *testing.T) {
	// Two lessons of the same group must take distinct slots; with an EarlySlot
	// preference and lowest-index tiebreak the placement is fully pinned:
	// variable 1 -> slot 1, variable 2 -> slot 2, both in the only room.
	in := Input{
		Variables: []Variable{
			{ID: 1, GroupID: 1, TeacherID: 1, GroupSize: 20, WeekType: domain.WeekTypeAll},
			{ID: 2, GroupID: 1, TeacherID: 2, GroupSize: 20, WeekType: domain.WeekTypeAll},
		},
		Days:    []domain.DayOfWeek{domain.Monday},
		Slots:   []int{1, 2},
		Rooms:   []Room{{ID: 10, Capacity: 30, Type: "lecture", Available: true}},
		Weights: SoftWeights{EarlySlot: 1},
	}

	want := []Assignment{
		{Variable: in.Variables[0], Value: Value{Day: domain.Monday, Slot: 1, RoomID: 10}},
		{Variable: in.Variables[1], Value: Value{Day: domain.Monday, Slot: 2, RoomID: 10}},
	}

	res := Solve(in)
	if !reflect.DeepEqual(res.Assignments, want) {
		t.Errorf("unexpected placement:\n got %+v\nwant %+v", res.Assignments, want)
	}
	if len(res.Unplaced) != 0 {
		t.Errorf("expected nothing unplaced, got %+v", res.Unplaced)
	}
}

func TestSolve_Deterministic(t *testing.T) {
	// A non-trivial instance (mixed groups, parity, an oversized group) must
	// yield byte-for-byte identical results across repeated runs.
	in := Input{
		Variables: []Variable{
			{ID: 1, GroupID: 1, TeacherID: 1, GroupSize: 20, WeekType: domain.WeekTypeAll},
			{ID: 2, GroupID: 1, TeacherID: 2, GroupSize: 20, WeekType: domain.WeekTypeOdd},
			{ID: 3, GroupID: 2, TeacherID: 1, GroupSize: 20, WeekType: domain.WeekTypeEven},
			{ID: 4, GroupID: 2, TeacherID: 3, GroupSize: 500, WeekType: domain.WeekTypeAll}, // unplaceable
		},
		Days:    []domain.DayOfWeek{domain.Monday, domain.Tuesday},
		Slots:   []int{1, 2, 3},
		Rooms:   []Room{{ID: 10, Capacity: 30, Type: "lecture", Available: true}, {ID: 11, Capacity: 25, Type: "lecture", Available: true}},
		Weights: NewDefaultWeights(),
	}

	first := Solve(in)
	for range 5 {
		if got := Solve(in); !reflect.DeepEqual(got, first) {
			t.Fatalf("Solve is not deterministic:\nfirst %+v\ngot   %+v", first, got)
		}
	}
	noPairwiseConflict(t, first.Assignments)
}

func TestSolve_ParityDisjointSharesRoom(t *testing.T) {
	// An odd-week and an even-week lesson may share the same day/slot/room.
	in := Input{
		Variables: []Variable{
			{ID: 1, GroupID: 1, TeacherID: 1, GroupSize: 20, WeekType: domain.WeekTypeOdd},
			{ID: 2, GroupID: 2, TeacherID: 2, GroupSize: 20, WeekType: domain.WeekTypeEven},
		},
		Days:    []domain.DayOfWeek{domain.Monday},
		Slots:   []int{1},
		Rooms:   []Room{{ID: 10, Capacity: 30, Type: "lecture", Available: true}},
		Weights: NewDefaultWeights(),
	}

	res := Solve(in)

	if len(res.Assignments) != 2 {
		t.Fatalf("odd and even lessons should both fit the shared slot, got %d placed (%+v)",
			len(res.Assignments), res.Assignments)
	}
	if len(res.Unplaced) != 0 {
		t.Errorf("nothing should be unplaced, got %+v", res.Unplaced)
	}
}
