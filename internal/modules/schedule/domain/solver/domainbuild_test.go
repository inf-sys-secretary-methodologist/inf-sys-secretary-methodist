package solver

import (
	"testing"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/schedule/domain"
)

func TestRoomTypeOK(t *testing.T) {
	tests := []struct {
		name     string
		allowed  []string
		roomType string
		want     bool
	}{
		{"empty allow-list accepts any type", nil, "lecture", true},
		{"empty slice accepts any type", []string{}, "lab", true},
		{"matching type accepted", []string{"lecture", "seminar"}, "seminar", true},
		{"non-matching type rejected", []string{"lecture"}, "lab", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := roomTypeOK(tt.allowed, tt.roomType); got != tt.want {
				t.Errorf("roomTypeOK(%v, %q) = %v, want %v", tt.allowed, tt.roomType, got, tt.want)
			}
		})
	}
}

func TestBuildDomain_FiltersRooms(t *testing.T) {
	in := Input{
		Days:  []domain.DayOfWeek{domain.Monday, domain.Tuesday},
		Slots: []int{1, 2},
		Rooms: []Room{
			{ID: 1, Capacity: 30, Type: "lecture", Available: true},  // ok
			{ID: 2, Capacity: 10, Type: "lecture", Available: true},  // too small
			{ID: 3, Capacity: 30, Type: "lab", Available: true},      // wrong type
			{ID: 4, Capacity: 30, Type: "lecture", Available: false}, // unavailable
		},
	}
	v := Variable{GroupSize: 25, AllowedRoomTypes: []string{"lecture"}}

	got := buildDomain(v, in)

	// Only room 1 qualifies -> 2 days * 2 slots * 1 room = 4 values.
	if len(got) != 4 {
		t.Fatalf("buildDomain len = %d, want 4; got %+v", len(got), got)
	}
	for _, val := range got {
		if val.RoomID != 1 {
			t.Errorf("unexpected room %d in domain, only room 1 should qualify", val.RoomID)
		}
	}
}

func TestBuildDomain_EmptyWhenNoRoomFits(t *testing.T) {
	in := Input{
		Days:  []domain.DayOfWeek{domain.Monday},
		Slots: []int{1},
		Rooms: []Room{{ID: 1, Capacity: 10, Type: "lecture", Available: true}},
	}
	v := Variable{GroupSize: 40, AllowedRoomTypes: []string{"lecture"}}

	if got := buildDomain(v, in); len(got) != 0 {
		t.Errorf("expected empty domain for oversized group, got %+v", got)
	}
}

func TestBuildDomain_AnyRoomTypeWhenUnrestricted(t *testing.T) {
	in := Input{
		Days:  []domain.DayOfWeek{domain.Monday},
		Slots: []int{1},
		Rooms: []Room{
			{ID: 1, Capacity: 30, Type: "lecture", Available: true},
			{ID: 2, Capacity: 30, Type: "lab", Available: true},
		},
	}
	v := Variable{GroupSize: 20, AllowedRoomTypes: nil}

	if got := buildDomain(v, in); len(got) != 2 {
		t.Errorf("unrestricted variable should reach both rooms, got %d values", len(got))
	}
}
