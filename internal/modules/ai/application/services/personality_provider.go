// Package services contains application services for the AI module.
package services

import "github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/entities"

// PersonalityProvider defines the interface for Metodych personality features.
// It abstracts away prompt content, greetings, mood comments, and notification formatting.
type PersonalityProvider interface {
	// BuildSystemPrompt builds a system prompt for the LLM incorporating mood context.
	BuildSystemPrompt(mood entities.MoodContext) string

	// FormatRAGContext formats retrieved document chunks into a context string for RAG.
	// Returns empty string if sources is empty.
	FormatRAGContext(sources []entities.ChunkWithScore) string

	// GetGreeting returns a greeting appropriate for the time of day.
	GetGreeting(timeOfDay string) string

	// GetMoodComment returns a comment based on the current mood.
	GetMoodComment(mood entities.MoodContext) string

	// FormatNotification formats a notification with personality (no LLM, instant).
	FormatNotification(notifType, title, message string, mood entities.MoodContext) string
}
