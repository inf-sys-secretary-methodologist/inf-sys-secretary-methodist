package solver

// NewDefaultWeights returns the default soft-preference weights. Hard rules are
// absolute; these only order otherwise-legal choices, so the values are small
// and relative: compactness (no gaps) matters most, an even spread across days
// next, and a mild nudge toward earlier slots last.
func NewDefaultWeights() SoftWeights {
	return SoftWeights{
		GroupGap:   1.0,
		TeacherGap: 1.0,
		DaySpread:  0.5,
		EarlySlot:  0.1,
	}
}

// gapCount returns the number of empty slots strictly between the earliest and
// latest occupied slot in the list — i.e. the "windows" in a day. A day with no
// lesson or a single lesson has no gaps.
func gapCount(slots []int) int {
	if len(slots) <= 1 {
		return 0
	}
	mn, mx := slots[0], slots[0]
	for _, s := range slots[1:] {
		if s < mn {
			mn = s
		}
		if s > mx {
			mx = s
		}
	}
	// Span covers (mx-mn+1) slot positions; len(slots) of them are occupied.
	// The remainder are windows. Occupied slots are distinct in a valid day, so
	// this is never negative in practice; clamp defensively all the same.
	if gaps := (mx - mn + 1) - len(slots); gaps > 0 {
		return gaps
	}
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
	groupSlots := []int{candidate.Value.Slot}
	teacherSlots := []int{candidate.Value.Slot}
	groupSameDay := 0

	for _, a := range current {
		if a.Value.Day != candidate.Value.Day {
			continue
		}
		if a.Variable.GroupID == candidate.Variable.GroupID {
			groupSlots = append(groupSlots, a.Value.Slot)
			groupSameDay++
		}
		if a.Variable.TeacherID == candidate.Variable.TeacherID {
			teacherSlots = append(teacherSlots, a.Value.Slot)
		}
	}

	p := w.EarlySlot * float64(candidate.Value.Slot)
	p += w.GroupGap * float64(gapCount(groupSlots))
	p += w.TeacherGap * float64(gapCount(teacherSlots))
	p += w.DaySpread * float64(groupSameDay)
	return p
}
