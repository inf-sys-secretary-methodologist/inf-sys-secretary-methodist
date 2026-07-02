package domain

import (
	"slices"
	"testing"
)

func TestAllowedRoomTypesForLesson(t *testing.T) {
	tests := []struct {
		name      string
		shortName string
		want      []string
	}{
		{"lecture short", "Лек", []string{"lecture"}},
		{"lecture full", "Лекция", []string{"lecture"}},
		{"practice", "Практ", []string{"practice"}},
		{"seminar", "Сем", []string{"practice"}},
		{"lab maps to lab and computer", "Лаб", []string{"lab", "computer"}},
		{"consultation is unrestricted", "Конс", nil},
		{"exam is unrestricted", "Экз", nil},
		{"unknown is unrestricted", "Хз", nil},
		{"empty is unrestricted", "", nil},
		{"case and spaces are normalized", "  лаб  ", []string{"lab", "computer"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AllowedRoomTypesForLesson(tt.shortName)
			if !slices.Equal(got, tt.want) {
				t.Errorf("AllowedRoomTypesForLesson(%q) = %v, want %v", tt.shortName, got, tt.want)
			}
		})
	}
}
