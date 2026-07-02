package solver

import "sort"

// maxBacktrackSteps caps the exhaustive search so an infeasible or huge instance
// degrades to best-effort greedy placement instead of running unbounded. It is a
// fixed value so results stay deterministic across runs.
const maxBacktrackSteps = 1_000_000

// Solve places every Variable in the Input onto the weekly timetable, honoring
// the hard constraints H1-H4 and steering by the soft preferences. It is
// best-effort: on a feasible instance it returns a complete, conflict-free
// assignment honoring every hard rule; on an over-constrained one it places as
// many variables as it can and reports the rest in Unplaced, never failing
// outright.
func Solve(in Input) Result {
	n := len(in.Variables)
	domains := make([][]Value, n)
	for i := range in.Variables {
		domains[i] = buildDomain(in.Variables[i], in)
	}

	// Phase 1: try for a complete, conflict-free assignment of every variable.
	if assigned, ok := solveComplete(in, domains); ok {
		res := Result{Assignments: make([]Assignment, 0, n)}
		for i := range in.Variables {
			res.Assignments = append(res.Assignments, Assignment{Variable: in.Variables[i], Value: assigned[i]})
		}
		return res
	}

	// Phase 2: best-effort — place as many variables as possible, rest Unplaced.
	return solveGreedy(in, domains)
}

// solveComplete runs a backtracking search (MRV variable ordering, soft-guided
// value ordering, forward-checking via on-demand consistency counts) that tries
// to assign all variables. It returns (assignments, true) on a complete solution
// found within the step budget, else (nil, false).
func solveComplete(in Input, domains [][]Value) ([]Value, bool) {
	n := len(in.Variables)
	assigned := make([]Value, n)
	set := make([]bool, n)
	steps := 0

	var backtrack func() bool
	backtrack = func() bool {
		idx := selectMRV(in, domains, assigned, set)
		if idx == -1 {
			return true // every variable is assigned
		}
		cands := consistentValues(in, domains[idx], assigned, set, idx)
		sortValues(cands, in.Variables[idx], currentAssignments(in, assigned, set), in.Weights)
		for _, val := range cands {
			steps++
			if steps > maxBacktrackSteps {
				return false
			}
			assigned[idx] = val
			set[idx] = true
			if backtrack() {
				return true
			}
			set[idx] = false
		}
		return false
	}

	if backtrack() {
		return assigned, true
	}
	return nil, false
}

// solveGreedy places variables one by one, most-constrained (smallest domain)
// first, taking each variable's cheapest consistent value or leaving it unplaced.
// It always terminates and never produces a hard conflict among placed lessons.
func solveGreedy(in Input, domains [][]Value) Result {
	n := len(in.Variables)
	assigned := make([]Value, n)
	set := make([]bool, n)

	order := make([]int, n)
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool {
		return len(domains[order[a]]) < len(domains[order[b]])
	})

	var unplaced []Variable
	for _, idx := range order {
		cands := consistentValues(in, domains[idx], assigned, set, idx)
		if len(cands) == 0 {
			unplaced = append(unplaced, in.Variables[idx])
			continue
		}
		sortValues(cands, in.Variables[idx], currentAssignments(in, assigned, set), in.Weights)
		assigned[idx] = cands[0]
		set[idx] = true
	}

	res := Result{Unplaced: unplaced}
	for i := range in.Variables {
		if set[i] {
			res.Assignments = append(res.Assignments, Assignment{Variable: in.Variables[i], Value: assigned[i]})
		}
	}
	return res
}

// selectMRV returns the index of the unassigned variable with the fewest
// remaining consistent values (Minimum Remaining Values). A variable whose
// domain has been wiped out (count 0) is selected first so the search fails fast
// — this is the forward-checking effect. Returns -1 when all variables are set.
// Ties break on the lowest index for determinism.
func selectMRV(in Input, domains [][]Value, assigned []Value, set []bool) int {
	best := -1
	bestCount := 0
	for i := range in.Variables {
		if set[i] {
			continue
		}
		c := len(consistentValues(in, domains[i], assigned, set, i))
		if best == -1 || c < bestCount {
			best = i
			bestCount = c
		}
	}
	return best
}

// consistentValues filters a variable's domain to the values that clash with no
// currently-placed assignment.
func consistentValues(in Input, dom []Value, assigned []Value, set []bool, idx int) []Value {
	var out []Value
	for _, val := range dom {
		if consistent(in, assigned, set, idx, val) {
			out = append(out, val)
		}
	}
	return out
}

// consistent reports whether assigning val to variable idx conflicts with any
// already-placed assignment under the hard rules.
func consistent(in Input, assigned []Value, set []bool, idx int, val Value) bool {
	cand := Assignment{Variable: in.Variables[idx], Value: val}
	for j := range in.Variables {
		if j == idx || !set[j] {
			continue
		}
		if assignmentsConflict(cand, Assignment{Variable: in.Variables[j], Value: assigned[j]}) {
			return false
		}
	}
	return true
}

// currentAssignments collects the placed assignments, in variable order, for
// soft-penalty scoring.
func currentAssignments(in Input, assigned []Value, set []bool) []Assignment {
	var cur []Assignment
	for i := range in.Variables {
		if set[i] {
			cur = append(cur, Assignment{Variable: in.Variables[i], Value: assigned[i]})
		}
	}
	return cur
}

// sortValues orders candidate values by ascending soft penalty (cheapest first),
// with a canonical (day, slot, room) tiebreak so the search is deterministic.
func sortValues(values []Value, v Variable, current []Assignment, w SoftWeights) {
	sort.SliceStable(values, func(i, j int) bool {
		pi := penalty(Assignment{Variable: v, Value: values[i]}, current, w)
		pj := penalty(Assignment{Variable: v, Value: values[j]}, current, w)
		if pi != pj {
			return pi < pj
		}
		if values[i].Day != values[j].Day {
			return values[i].Day < values[j].Day
		}
		if values[i].Slot != values[j].Slot {
			return values[i].Slot < values[j].Slot
		}
		return values[i].RoomID < values[j].RoomID
	})
}
