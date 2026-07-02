package solver

// roomTypeOK reports whether roomType satisfies a variable's allowed room types.
// An empty allow-list means the lesson type imposes no room-type requirement.
func roomTypeOK(allowed []string, roomType string) bool {
	panic("not implemented")
}

// buildDomain enumerates every legal (day, slot, room) placement for a single
// variable under the per-value hard constraint H4: the room must be available,
// have capacity for the whole group, and be of an allowed type. Cross-variable
// resource clashes (H1-H3) are handled later during search. A variable whose
// domain comes back empty is inherently unplaceable.
func buildDomain(v Variable, in Input) []Value {
	panic("not implemented")
}
