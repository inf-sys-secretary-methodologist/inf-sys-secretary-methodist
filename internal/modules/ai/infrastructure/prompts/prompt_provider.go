// Package prompts contains prompt templates and personality data for Metodych.
package prompts

import (
	"bytes"
	"embed"
	"fmt"
	"math/rand"
	"text/template"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/entities"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

// templateFuncs contains custom functions available in templates.
var templateFuncs = template.FuncMap{
	"inc": func(i int) int { return i + 1 },
	"mul": func(a, b float64) float64 { return a * b },
}

// PromptProvider implements services.PersonalityProvider using embedded templates.
type PromptProvider struct {
	systemPromptTmpl *template.Template
	ragContextTmpl   *template.Template
	notifTemplates   map[string]*template.Template
}

// NewPromptProvider creates a new PromptProvider, parsing all embedded templates.
// Panics if any template is invalid (fail-fast at startup).
func NewPromptProvider() *PromptProvider {
	p := &PromptProvider{
		notifTemplates: make(map[string]*template.Template),
	}

	p.systemPromptTmpl = template.Must(
		template.New("system_prompt.tmpl").Funcs(templateFuncs).ParseFS(templateFS, "templates/system_prompt.tmpl"),
	)

	p.ragContextTmpl = template.Must(
		template.New("rag_context.tmpl").Funcs(templateFuncs).ParseFS(templateFS, "templates/rag_context.tmpl"),
	)

	notifTypes := []string{"default", "document", "reminder", "task", "system"}
	for _, nt := range notifTypes {
		filename := fmt.Sprintf("templates/notification_%s.tmpl", nt)
		tmpl := template.Must(
			template.New(fmt.Sprintf("notification_%s.tmpl", nt)).Funcs(templateFuncs).ParseFS(templateFS, filename),
		)
		p.notifTemplates[nt] = tmpl
	}

	return p
}

// systemPromptData holds template data for the system prompt.
type systemPromptData struct {
	MoodInstruction  string
	OverdueDocuments int
	AtRiskStudents   int
}

// BuildSystemPrompt builds a system prompt for the LLM incorporating mood context.
func (p *PromptProvider) BuildSystemPrompt(mood entities.MoodContext) string {
	instruction := MoodInstructions[mood.State]

	data := systemPromptData{
		MoodInstruction:  instruction,
		OverdueDocuments: mood.OverdueDocuments,
		AtRiskStudents:   mood.AtRiskStudents,
	}

	var buf bytes.Buffer
	if err := p.systemPromptTmpl.Execute(&buf, data); err != nil {
		// Should never happen with valid templates, but provide a minimal fallback.
		return "Ты — Методыч, ветеран-методист. Отвечай на русском языке."
	}
	return buf.String()
}

// ragContextData holds template data for RAG context formatting.
type ragContextData struct {
	Sources []entities.ChunkWithScore
}

// FormatRAGContext formats retrieved document chunks into a context string for RAG.
func (p *PromptProvider) FormatRAGContext(sources []entities.ChunkWithScore) string {
	if len(sources) == 0 {
		return ""
	}

	data := ragContextData{Sources: sources}

	var buf bytes.Buffer
	if err := p.ragContextTmpl.Execute(&buf, data); err != nil {
		return ""
	}
	return buf.String()
}

// GetGreeting returns a greeting appropriate for the time of day.
func (p *PromptProvider) GetGreeting(timeOfDay string) string {
	greetings, ok := Greetings[timeOfDay]
	if !ok {
		greetings = Greetings["morning"]
	}
	return greetings[rand.Intn(len(greetings))] // #nosec G404 -- weak RNG is fine for random greeting selection
}

// GetMoodComment returns a comment based on the current mood.
func (p *PromptProvider) GetMoodComment(mood entities.MoodContext) string {
	comments, ok := MoodComments[mood.State]
	if !ok {
		comments = MoodComments[entities.MoodContent]
	}
	return comments[rand.Intn(len(comments))] // #nosec G404 -- weak RNG is fine for random mood comment selection
}

// notificationData holds template data for notification formatting.
type notificationData struct {
	Title   string
	Message string
	Mood    entities.MoodState
	Emoji   string
}

// FormatNotification formats a notification with personality (no LLM, instant).
func (p *PromptProvider) FormatNotification(notifType, title, message string, mood entities.MoodContext) string {
	tmpl, ok := p.notifTemplates[notifType]
	if !ok {
		tmpl = p.notifTemplates["default"]
	}

	emoji, ok := MoodEmojis[mood.State]
	if !ok {
		emoji = DefaultMoodEmoji
	}

	data := notificationData{
		Title:   title,
		Message: message,
		Mood:    mood.State,
		Emoji:   emoji,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Sprintf("%s\n%s", title, message)
	}
	return buf.String()
}
