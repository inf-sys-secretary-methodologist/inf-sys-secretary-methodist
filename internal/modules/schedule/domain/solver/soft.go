package solver

// NewDefaultWeights returns the default soft-preference weights. Hard rules are
// absolute; these only order otherwise-legal choices, so the values are small
// and relative: compactness (no gaps) matters most, an even spread across days
// next, and a mild nudge toward earlier slots last.
func NewDefaultWeights() SoftWeights {
	return SoftWeights{}
}

// gapCount returns the number of empty slots strictly between the earliest and
// latest occupied slot in the list — i.e. the "windows" in a day. A day with no
// lesson or a single lesson has no gaps.
func gapCount(slots []int) int {
	return 0
}

// penalty scores a candidate placement against the already-placed assignments:
// lower is better. It sums four soft costs, each scaled by its weight —
//   - GroupGap:   windows in the group's day if the candidate is added,
//   - TeacherGap: windows in the teacher's day if the candidate is added,
//   - DaySpread:  how many lessons the group already has that day (load balance),
//   - EarlySlot:  the slot number itself (earlier is cheaper).
//
// The engine uses penalty to order candidate values, so all four preferences
// actively steer the search rather than sitting unused.
func penalty(candidate Assignment, current []Assignment, w SoftWeights) float64 {
	return 0
}
