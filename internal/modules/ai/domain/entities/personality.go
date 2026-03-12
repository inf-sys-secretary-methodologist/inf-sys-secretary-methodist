// Package entities contains domain entities for the AI module.
package entities

import "time"

// MoodState represents the emotional state of the Metodych character
type MoodState string

// MoodState constants define the possible emotional states.
const (
	MoodHappy     MoodState = "happy"
	MoodContent   MoodState = "content"
	MoodWorried   MoodState = "worried"
	MoodStressed  MoodState = "stressed"
	MoodPanicking MoodState = "panicking"
	MoodRelaxed   MoodState = "relaxed"
	MoodInspired  MoodState = "inspired"
)

// MoodContext contains the context data used to compute mood
type MoodContext struct {
	State             MoodState `json:"state"`
	Intensity         float64   `json:"intensity"` // 0.0-1.0
	Reason            string    `json:"reason"`
	OverdueDocuments  int       `json:"overdue_documents"`
	AtRiskStudents    int       `json:"at_risk_students"`
	UpcomingDeadlines int       `json:"upcoming_deadlines"`
	TimeOfDay         string    `json:"time_of_day"` // morning, afternoon, evening, night
	DayOfWeek         string    `json:"day_of_week"`
	AttendanceTrend   string    `json:"attendance_trend"` // improving, stable, declining
	ComputedAt        time.Time `json:"computed_at"`
}

// GetTimeOfDay returns the time-of-day category for a given hour
func GetTimeOfDay(hour int) string {
	switch {
	case hour >= 6 && hour < 12:
		return "morning"
	case hour >= 12 && hour < 17:
		return "afternoon"
	case hour >= 17 && hour < 22:
		return "evening"
	default:
		return "night"
	}
}
