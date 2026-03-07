package content

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/agentsim/agent"
)

// LLMGenerator generates content using an LLM API.
type LLMGenerator struct {
	provider string
	model    string
	apiKey   string
	client   *http.Client
	fallback *TemplateGenerator
}

// NewLLMGenerator creates an LLM-based content generator.
func NewLLMGenerator(provider, model, apiKey string) *LLMGenerator {
	return &LLMGenerator{
		provider: provider,
		model:    model,
		apiKey:   apiKey,
		client:   &http.Client{Timeout: 30 * time.Second},
		fallback: NewTemplateGenerator(),
	}
}

func (g *LLMGenerator) generate(systemPrompt, userPrompt string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	var result string
	var err error

	switch g.provider {
	case "anthropic":
		result, err = g.callAnthropic(ctx, systemPrompt, userPrompt)
	case "openai":
		result, err = g.callOpenAI(ctx, systemPrompt, userPrompt)
	default:
		result, err = g.callAnthropic(ctx, systemPrompt, userPrompt)
	}

	if err != nil {
		// Fall back to template on LLM failure
		return ""
	}
	return result
}

func (g *LLMGenerator) callAnthropic(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	body := map[string]any{
		"model":      g.model,
		"max_tokens": 500,
		"system":     systemPrompt,
		"messages": []map[string]string{
			{"role": "user", "content": userPrompt},
		},
	}

	data, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", g.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := g.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("anthropic API error %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}

	if len(result.Content) > 0 {
		return result.Content[0].Text, nil
	}
	return "", fmt.Errorf("empty response from Anthropic")
}

func (g *LLMGenerator) callOpenAI(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	body := map[string]any{
		"model":      g.model,
		"max_tokens": 500,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
	}

	data, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+g.apiKey)

	resp, err := g.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("openai API error %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}

	if len(result.Choices) > 0 {
		return result.Choices[0].Message.Content, nil
	}
	return "", fmt.Errorf("empty response from OpenAI")
}

const llmSystemPrompt = `Ты генерируешь реалистичный текст для симуляции работы информационной системы вуза.
Пиши кратко, по-деловому, на русском языке. Не используй markdown-разметку.
Отвечай ТОЛЬКО запрошенным текстом, без пояснений и комментариев.`

func (g *LLMGenerator) DocumentTitle(docType, ctx string) string {
	result := g.generate(llmSystemPrompt,
		fmt.Sprintf("Сгенерируй заголовок документа типа '%s'. Контекст: %s. Только заголовок, одно предложение.", docType, ctx))
	if result == "" {
		return g.fallback.DocumentTitle(docType, ctx)
	}
	return result
}

func (g *LLMGenerator) DocumentContent(docType, title, ctx string) string {
	result := g.generate(llmSystemPrompt,
		fmt.Sprintf("Сгенерируй текст документа типа '%s' с заголовком '%s'. Контекст: %s. 2-4 абзаца делового текста.", docType, title, ctx))
	if result == "" {
		return g.fallback.DocumentContent(docType, title, ctx)
	}
	return result
}

func (g *LLMGenerator) ChatMessage(from *agent.Agent, to, topic string) string {
	result := g.generate(llmSystemPrompt,
		fmt.Sprintf("Ты — %s (%s, %s). Напиши короткое сообщение коллеге на тему: %s. 1-2 предложения.",
			from.Name, from.Role, from.Personality, topic))
	if result == "" {
		return g.fallback.ChatMessage(from, to, topic)
	}
	return result
}

func (g *LLMGenerator) TaskTitle(subject string) string {
	if subject != "" {
		return subject
	}
	result := g.generate(llmSystemPrompt,
		"Сгенерируй название задачи для сотрудника вуза. Только название, одно предложение.")
	if result == "" {
		return g.fallback.TaskTitle(subject)
	}
	return result
}

func (g *LLMGenerator) TaskDescription(title, ctx string) string {
	result := g.generate(llmSystemPrompt,
		fmt.Sprintf("Сгенерируй описание задачи '%s'. Контекст: %s. 2-3 предложения.", title, ctx))
	if result == "" {
		return g.fallback.TaskDescription(title, ctx)
	}
	return result
}

func (g *LLMGenerator) EventTitle(eventType string) string {
	result := g.generate(llmSystemPrompt,
		fmt.Sprintf("Сгенерируй название мероприятия типа '%s' в вузе. Только название.", eventType))
	if result == "" {
		return g.fallback.EventTitle(eventType)
	}
	return result
}

func (g *LLMGenerator) Comment(about string) string {
	result := g.generate(llmSystemPrompt,
		fmt.Sprintf("Напиши краткий комментарий на тему: %s. 1-2 предложения, деловой стиль.", about))
	if result == "" {
		return g.fallback.Comment(about)
	}
	return result
}

func (g *LLMGenerator) ReportTitle(reportType string) string {
	result := g.generate(llmSystemPrompt,
		fmt.Sprintf("Сгенерируй название отчёта типа '%s'. Только название.", reportType))
	if result == "" {
		return g.fallback.ReportTitle(reportType)
	}
	return result
}

func (g *LLMGenerator) ReportDescription(title, ctx string) string {
	result := g.generate(llmSystemPrompt,
		fmt.Sprintf("Напиши описание отчёта '%s'. Контекст: %s. 2-3 предложения.", title, ctx))
	if result == "" {
		return g.fallback.ReportDescription(title, ctx)
	}
	return result
}
