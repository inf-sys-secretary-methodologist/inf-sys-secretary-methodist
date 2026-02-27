// Package dto contains data transfer objects for the AI module.
package dto

import (
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/entities"
)

// MoodResponse represents the mood API response
type MoodResponse struct {
	State            string    `json:"state"`
	Intensity        float64   `json:"intensity"`
	Reason           string    `json:"reason"`
	Message          string    `json:"message"`
	Greeting         string    `json:"greeting"`
	FunFact          *string   `json:"fun_fact,omitempty"`
	OverdueDocuments int       `json:"overdue_documents"`
	AtRiskStudents   int       `json:"at_risk_students"`
	ComputedAt       time.Time `json:"computed_at"`
}

// ToMoodResponse converts a MoodContext to a MoodResponse
func ToMoodResponse(mood *entities.MoodContext, message, greeting string) *MoodResponse {
	return &MoodResponse{
		State:            string(mood.State),
		Intensity:        mood.Intensity,
		Reason:           mood.Reason,
		Message:          message,
		Greeting:         greeting,
		OverdueDocuments: mood.OverdueDocuments,
		AtRiskStudents:   mood.AtRiskStudents,
		ComputedAt:       mood.ComputedAt,
	}
}
