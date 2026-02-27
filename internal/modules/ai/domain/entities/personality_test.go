package entities

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTimeOfDay(t *testing.T) {
	tests := []struct {
		name     string
		hour     int
		expected string
	}{
		{"midnight", 0, "night"},
		{"early night", 5, "night"},
		{"morning start", 6, "morning"},
		{"late morning", 11, "morning"},
		{"afternoon start", 12, "afternoon"},
		{"late afternoon", 16, "afternoon"},
		{"evening start", 17, "evening"},
		{"late evening", 21, "evening"},
		{"night start", 22, "night"},
		{"before midnight", 23, "night"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetTimeOfDay(tt.hour)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMoodStateConstants(t *testing.T) {
	assert.Equal(t, MoodState("happy"), MoodHappy)
	assert.Equal(t, MoodState("content"), MoodContent)
	assert.Equal(t, MoodState("worried"), MoodWorried)
	assert.Equal(t, MoodState("stressed"), MoodStressed)
	assert.Equal(t, MoodState("panicking"), MoodPanicking)
	assert.Equal(t, MoodState("relaxed"), MoodRelaxed)
	assert.Equal(t, MoodState("inspired"), MoodInspired)
}
