package prompts

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/entities"
)

func TestNewPromptProvider(t *testing.T) {
	// Should not panic.
	p := NewPromptProvider()
	assert.NotNil(t, p)
}

func TestBuildSystemPrompt(t *testing.T) {
	p := NewPromptProvider()

	tests := []struct {
		name             string
		mood             entities.MoodContext
		expectedContains []string
	}{
		{
			name: "happy mood",
			mood: entities.MoodContext{State: entities.MoodHappy},
			expectedContains: []string{
				"Методыч",
				"отличное настроение",
			},
		},
		{
			name: "content mood",
			mood: entities.MoodContext{State: entities.MoodContent},
			expectedContains: []string{
				"Методыч",
				"спокойное",
			},
		},
		{
			name: "worried mood",
			mood: entities.MoodContext{State: entities.MoodWorried},
			expectedContains: []string{
				"Методыч",
				"переживаешь",
			},
		},
		{
			name: "stressed mood",
			mood: entities.MoodContext{State: entities.MoodStressed},
			expectedContains: []string{
				"Методыч",
				"стрессе",
			},
		},
		{
			name: "panicking mood",
			mood: entities.MoodContext{State: entities.MoodPanicking},
			expectedContains: []string{
				"Методыч",
				"ПАНИКА",
			},
		},
		{
			name: "relaxed mood",
			mood: entities.MoodContext{State: entities.MoodRelaxed},
			expectedContains: []string{
				"Методыч",
				"Расслабленное",
			},
		},
		{
			name: "inspired mood",
			mood: entities.MoodContext{State: entities.MoodInspired},
			expectedContains: []string{
				"Методыч",
				"вдохновлён",
			},
		},
		{
			name: "with overdue documents",
			mood: entities.MoodContext{
				State:            entities.MoodStressed,
				OverdueDocuments: 5,
			},
			expectedContains: []string{
				"5 новых документов",
			},
		},
		{
			name: "with at-risk students",
			mood: entities.MoodContext{
				State:          entities.MoodWorried,
				AtRiskStudents: 3,
			},
			expectedContains: []string{
				"3 студентов в зоне риска",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := p.BuildSystemPrompt(tt.mood)
			for _, substr := range tt.expectedContains {
				assert.Contains(t, result, substr)
			}
		})
	}
}

func TestFormatRAGContext(t *testing.T) {
	p := NewPromptProvider()

	t.Run("empty sources returns empty string", func(t *testing.T) {
		result := p.FormatRAGContext(nil)
		assert.Empty(t, result)
	})

	t.Run("formats sources correctly", func(t *testing.T) {
		sources := []entities.ChunkWithScore{
			{
				Chunk: &entities.DocumentChunk{
					ChunkText: "Содержимое документа",
				},
				DocumentTitle:   "Приказ №1",
				SimilarityScore: 0.95,
			},
			{
				Chunk: &entities.DocumentChunk{
					ChunkText: "Другой фрагмент",
				},
				DocumentTitle:   "Приказ №2",
				SimilarityScore: 0.80,
			},
		}

		result := p.FormatRAGContext(sources)
		assert.Contains(t, result, "КОНТЕКСТ ИЗ ДОКУМЕНТОВ")
		assert.Contains(t, result, "Приказ №1")
		assert.Contains(t, result, "Приказ №2")
		assert.Contains(t, result, "95%")
		assert.Contains(t, result, "80%")
		assert.Contains(t, result, "Содержимое документа")
		assert.Contains(t, result, "<documents>")
		assert.Contains(t, result, "</documents>")
		assert.Contains(t, result, "[1]")
		assert.Contains(t, result, "[2]")
	})
}

func TestGetGreeting(t *testing.T) {
	p := NewPromptProvider()

	tests := []struct {
		name      string
		timeOfDay string
	}{
		{"morning", "morning"},
		{"afternoon", "afternoon"},
		{"evening", "evening"},
		{"night", "night"},
		{"unknown falls back", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := p.GetGreeting(tt.timeOfDay)
			assert.NotEmpty(t, result)
		})
	}
}

func TestGetMoodComment(t *testing.T) {
	p := NewPromptProvider()

	moods := []entities.MoodState{
		entities.MoodHappy,
		entities.MoodContent,
		entities.MoodWorried,
		entities.MoodStressed,
		entities.MoodPanicking,
		entities.MoodRelaxed,
		entities.MoodInspired,
	}

	for _, mood := range moods {
		t.Run(string(mood), func(t *testing.T) {
			result := p.GetMoodComment(entities.MoodContext{State: mood})
			assert.NotEmpty(t, result)
		})
	}

	t.Run("unknown mood falls back to content", func(t *testing.T) {
		result := p.GetMoodComment(entities.MoodContext{State: "unknown"})
		assert.NotEmpty(t, result)
	})
}

func TestFormatNotification(t *testing.T) {
	p := NewPromptProvider()

	mood := entities.MoodContext{State: entities.MoodContent}

	tests := []struct {
		name             string
		notifType        string
		title            string
		message          string
		expectedContains []string
	}{
		{
			name:      "default template",
			notifType: "default",
			title:     "Тест",
			message:   "Сообщение",
			expectedContains: []string{
				"Тест",
				"Сообщение",
				"Ваш Методыч",
			},
		},
		{
			name:      "document template",
			notifType: "document",
			title:     "Новый документ",
			message:   "Документ создан",
			expectedContains: []string{
				"Новый документ",
				"Документ создан",
				"документы любят порядок",
			},
		},
		{
			name:      "reminder template",
			notifType: "reminder",
			title:     "Напоминание",
			message:   "Скоро дедлайн",
			expectedContains: []string{
				"Напоминание",
				"Скоро дедлайн",
				"не откладывай на завтра",
			},
		},
		{
			name:      "task template",
			notifType: "task",
			title:     "Задача",
			message:   "Новая задача",
			expectedContains: []string{
				"Задача",
				"Новая задача",
				"задачи не решаются сами",
			},
		},
		{
			name:      "system template",
			notifType: "system",
			title:     "Система",
			message:   "Обновление",
			expectedContains: []string{
				"Система",
				"Обновление",
				"Методыч в курсе",
			},
		},
		{
			name:      "unknown type falls back to default",
			notifType: "nonexistent",
			title:     "Тест",
			message:   "Сообщение",
			expectedContains: []string{
				"Тест",
				"Сообщение",
				"Ваш Методыч",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := p.FormatNotification(tt.notifType, tt.title, tt.message, mood)
			for _, substr := range tt.expectedContains {
				assert.Contains(t, result, substr)
			}
		})
	}
}
