package solver

import "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain"

// parityConflicts reports whether two week-types can ever collide on the same
// day+slot. An "all" (every-week) lesson overlaps any other week-type; two
// odd-week lessons overlap, two even-week lessons overlap, but an odd-week and
// an even-week lesson never share a physical week and so never conflict.
func parityConflicts(a, b domain.WeekType) bool {
	panic("not implemented")
}

// assignmentsConflict reports whether two assignments violate a hard resource
// constraint (H1-H3): they land on the same day and slot, their weeks overlap
// (parityConflicts), and they share the same teacher, group, or room.
func assignmentsConflict(a1, a2 Assignment) bool {
	panic("not implemented")
}
