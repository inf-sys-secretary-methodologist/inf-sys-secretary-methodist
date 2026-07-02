package solver

import "slices"

// roomTypeOK reports whether roomType satisfies a variable's allowed room types.
// An empty allow-list means the lesson type imposes no room-type requirement.
func roomTypeOK(allowed []string, roomType string) bool {
	if len(allowed) == 0 {
		return true
	}
	return slices.Contains(allowed, roomType)
}

// buildDomain enumerates every legal (day, slot, room) placement for a single
// variable under the per-value hard constraint H4: the room must be available,
// have capacity for the whole group, and be of an allowed type. Cross-variable
// resource clashes (H1-H3) are handled later during search. A variable whose
// domain comes back empty is inherently unplaceable.
func buildDomain(v Variable, in Input) []Value {
	values := make([]Value, 0, len(in.Days)*len(in.Slots)*len(in.Rooms))
	for _, day := range in.Days {
		for _, slot := range in.Slots {
			for _, room := range in.Rooms {
				if !room.Available {
					continue
				}
				if room.Capacity < v.GroupSize {
					continue
				}
				if !roomTypeOK(v.AllowedRoomTypes, room.Type) {
					continue
				}
				values = append(values, Value{Day: day, Slot: slot, RoomID: room.ID})
			}
		}
	}
	return values
}
